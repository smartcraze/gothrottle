package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10.0, 50)
	
	if rl.storage == nil {
		t.Error("Expected storage to be initialized")
	}
	
	if rl.storage.Count() != 0 {
		t.Errorf("Expected 0 clients initially, got %d", rl.storage.Count())
	}
}

func TestRateLimiterAllow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(10, 3)
	
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	
	// First 3 requests should be allowed (burst capacity)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}
	
	// 4th request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Request 4: expected status 429, got %d", w.Code)
	}
	
	// Check headers
	if w.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header")
	}
}

func TestRateLimiterMultipleClients(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(10, 2)
	
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	
	// Client 1: exhaust tokens
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
	
	// Client 1: should be rate limited
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	if w1.Code != http.StatusTooManyRequests {
		t.Errorf("Client 1 should be rate limited, got status %d", w1.Code)
	}
	
	// Client 2: should still be allowed (separate bucket)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	if w2.Code != http.StatusOK {
		t.Errorf("Client 2 should be allowed, got status %d", w2.Code)
	}
}

func TestRateLimiterAbort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(10, 1)
	
	handlerCalled := false
	router := gin.New()
	router.Use(rl.Limit())
	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	
	// First request: allowed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	if !handlerCalled {
		t.Error("Handler should be called for allowed request")
	}
	
	// Second request: should be blocked before handler
	handlerCalled = false
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	if handlerCalled {
		t.Error("Handler should NOT be called for rate limited request")
	}
	
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w2.Code)
	}
}
