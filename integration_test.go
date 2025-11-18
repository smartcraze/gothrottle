package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smartcraze/gothrottle/internal/config"
	"github.com/smartcraze/gothrottle/internal/middleware"
	"github.com/smartcraze/gothrottle/internal/proxy"
)

type responseWriterWrapper struct {
	*httptest.ResponseRecorder
}

func (w *responseWriterWrapper) CloseNotify() <-chan bool {
	return make(chan bool)
}

/*
TestEndToEndRateLimitingAndProxy validates the complete request flow
from rate limiting through to proxy forwarding using only public APIs.
*/
func TestEndToEndRateLimitingAndProxy(t *testing.T) {
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer mockBackend.Close()

	routes := []config.Route{
		{Path: "/api", Target: mockBackend.URL},
	}

	rateLimit := config.RateLimit{
		RequestsPerSecond: 10,
		Burst:             3,
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	rateLimiter := middleware.NewRateLimiter(float64(rateLimit.RequestsPerSecond), rateLimit.Burst)
	router.Use(rateLimiter.Limit())

	proxyHandler, err := proxy.NewHandler(routes)
	if err != nil {
		t.Fatalf("Failed to create proxy handler: %v", err)
	}
	router.NoRoute(proxyHandler.Handle)

	clientIP := "192.168.1.100"

	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		req.RemoteAddr = clientIP + ":1234"
		rec := httptest.NewRecorder()
		w := &responseWriterWrapper{rec}
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		router.ServeHTTP(w, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.RemoteAddr = clientIP + ":1234"
	rec := httptest.NewRecorder()
	w := &responseWriterWrapper{rec}
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	router.ServeHTTP(w, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Request 4 (rate limited): expected status 429, got %d", rec.Code)
	}

	if rec.Header().Get("Retry-After") != "1" {
		t.Errorf("Expected Retry-After header to be '1', got '%s'", rec.Header().Get("Retry-After"))
	}
}

/*
TestMultipleClientsIsolation verifies that rate limits are enforced
independently for different client IPs.
*/
func TestMultipleClientsIsolation(t *testing.T) {
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockBackend.Close()

	routes := []config.Route{
		{Path: "/api", Target: mockBackend.URL},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	rateLimiter := middleware.NewRateLimiter(5, 2)
	router.Use(rateLimiter.Limit())

	proxyHandler, err := proxy.NewHandler(routes)
	if err != nil {
		t.Fatalf("Failed to create proxy handler: %v", err)
	}
	router.NoRoute(proxyHandler.Handle)

	client1 := "192.168.1.100:1234"
	client2 := "192.168.1.200:5678"

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req.RemoteAddr = client1
		rec := httptest.NewRecorder()
		w := &responseWriterWrapper{rec}
		router.ServeHTTP(w, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.RemoteAddr = client1
	rec := httptest.NewRecorder()
	w := &responseWriterWrapper{rec}
	router.ServeHTTP(w, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Client 1 should be rate limited, got status %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.RemoteAddr = client2
	rec = httptest.NewRecorder()
	w = &responseWriterWrapper{rec}
	router.ServeHTTP(w, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Client 2 should be allowed (separate bucket), got status %d", rec.Code)
	}
}

/*
TestProxyRoutingWithRateLimit validates that path-based routing works
correctly when combined with rate limiting middleware.
*/
func TestProxyRoutingWithRateLimit(t *testing.T) {
	apiBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api-response"))
	}))
	defer apiBackend.Close()

	authBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("auth-response"))
	}))
	defer authBackend.Close()

	routes := []config.Route{
		{Path: "/api", Target: apiBackend.URL},
		{Path: "/auth", Target: authBackend.URL},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	rateLimiter := middleware.NewRateLimiter(100.0, 50)
	router.Use(rateLimiter.Limit())

	proxyHandler, err := proxy.NewHandler(routes)
	if err != nil {
		t.Fatalf("Failed to create proxy handler: %v", err)
	}
	router.NoRoute(proxyHandler.Handle)

	tests := []struct {
		path           string
		expectedBody   string
		expectedStatus int
	}{
		{"/api/users", "api-response", http.StatusOK},
		{"/auth/login", "auth-response", http.StatusOK},
		{"/unknown", "", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.RemoteAddr = "192.168.1.1:1234"
			rec := httptest.NewRecorder()
			w := &responseWriterWrapper{rec}

			router.ServeHTTP(w, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Path %s: expected status %d, got %d", tt.path, tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" && rec.Body.String() != tt.expectedBody {
				t.Errorf("Path %s: expected body %s, got %s", tt.path, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

/*
TestRateLimitRefill verifies that tokens are refilled over time,
allowing requests after waiting for the refill period.
*/
func TestRateLimitRefill(t *testing.T) {
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockBackend.Close()

	routes := []config.Route{
		{Path: "/api", Target: mockBackend.URL},
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	rateLimiter := middleware.NewRateLimiter(5, 2)
	router.Use(rateLimiter.Limit())

	proxyHandler, err := proxy.NewHandler(routes)
	if err != nil {
		t.Fatalf("Failed to create proxy handler: %v", err)
	}
	router.NoRoute(proxyHandler.Handle)

	clientIP := "192.168.1.1:1234"

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = clientIP
		rec := httptest.NewRecorder()
		w := &responseWriterWrapper{rec}
		router.ServeHTTP(w, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Initial request %d should be allowed", i+1)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = clientIP
	rec := httptest.NewRecorder()
	w := &responseWriterWrapper{rec}
	router.ServeHTTP(w, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Error("Request should be rate limited after burst")
	}

	time.Sleep(250 * time.Millisecond)

	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = clientIP
	rec = httptest.NewRecorder()
	w = &responseWriterWrapper{rec}
	router.ServeHTTP(w, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Request should be allowed after refill period, got status %d", rec.Code)
	}
}
