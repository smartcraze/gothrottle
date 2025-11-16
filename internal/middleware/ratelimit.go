package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcraze/gothrottle/internal/ratelimit"
)

// RateLimiter is a Gin middleware that enforces rate limits
type RateLimiter struct {
	storage *ratelimit.Storage
}

// NewRateLimiter creates a new rate limiting middleware
func NewRateLimiter(requestsPerSecond, burst int) *RateLimiter {
	return &RateLimiter{
		storage: ratelimit.NewStorage(requestsPerSecond, burst),
	}
}

// Limit returns a Gin middleware function that enforces rate limiting
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address)
		clientID := c.ClientIP()

		// Check if request is allowed
		if !rl.storage.Allow(clientID) {
			// Calculate retry after time (simplified: 1 second)
			c.Header("Retry-After", "1")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		// Request is allowed, continue to next handler
		c.Next()
	}
}

// Storage returns the underlying storage (for testing)
func (rl *RateLimiter) Storage() *ratelimit.Storage {
	return rl.storage
}
