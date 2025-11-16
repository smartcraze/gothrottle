package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smartcraze/gothrottle/internal/config"
)

// responseWriterWrapper wraps httptest.ResponseRecorder to implement http.CloseNotifier
type responseWriterWrapper struct {
	*httptest.ResponseRecorder
}

func (w *responseWriterWrapper) CloseNotify() <-chan bool {
	return make(chan bool)
}

func TestNewHandler(t *testing.T) {
	routes := []config.Route{
		{Path: "/api", Target: "http://localhost:8000"},
		{Path: "/auth", Target: "http://localhost:9000"},
	}

	handler, err := NewHandler(routes)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	if len(handler.proxies) != 2 {
		t.Errorf("Expected 2 proxies, got %d", len(handler.proxies))
	}

	if len(handler.routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(handler.routes))
	}
}

func TestNewHandlerInvalidURL(t *testing.T) {
	routes := []config.Route{
		{Path: "/api", Target: "://invalid-url"},
	}

	_, err := NewHandler(routes)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestHandleRouteMatching(t *testing.T) {
	// Create mock upstream servers
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"api","path":"` + r.URL.Path + `"}`))
	}))
	defer apiServer.Close()

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service":"auth","path":"` + r.URL.Path + `"}`))
	}))
	defer authServer.Close()

	routes := []config.Route{
		{Path: "/api", Target: apiServer.URL},
		{Path: "/auth", Target: authServer.URL},
	}

	handler, err := NewHandler(routes)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "route to api",
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"service":"api","path":"/api/users"}`,
		},
		{
			name:           "route to auth",
			path:           "/auth/login",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"service":"auth","path":"/auth/login"}`,
		},
		{
			name:           "no matching route",
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			rec := httptest.NewRecorder()
			w := &responseWriterWrapper{rec}
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tt.path, nil)

			handler.Handle(c)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" && rec.Body.String() != tt.expectedBody {
				t.Errorf("Expected body %s, got %s", tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestHandleLongestPrefixMatch(t *testing.T) {
	// Create mock servers
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api"))
	}))
	defer apiServer.Close()

	apiV2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api-v2"))
	}))
	defer apiV2Server.Close()

	routes := []config.Route{
		{Path: "/api", Target: apiServer.URL},
		{Path: "/api/v2", Target: apiV2Server.URL},
	}

	handler, err := NewHandler(routes)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	gin.SetMode(gin.TestMode)

	// Test that /api/v2/users goes to apiV2Server (longest match)
	rec := httptest.NewRecorder()
	w := &responseWriterWrapper{rec}
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v2/users", nil)
	handler.Handle(c)

	if rec.Body.String() != "api-v2" {
		t.Errorf("Expected 'api-v2', got %s (longest prefix match failed)", rec.Body.String())
	}

	// Test that /api/users goes to apiServer
	rec = httptest.NewRecorder()
	w = &responseWriterWrapper{rec}
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	handler.Handle(c)

	if rec.Body.String() != "api" {
		t.Errorf("Expected 'api', got %s", rec.Body.String())
	}
}
