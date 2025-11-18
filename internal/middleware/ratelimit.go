package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcraze/gothrottle/internal/ratelimit"
)

type RateLimiter struct {
	storage *ratelimit.Storage
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	return &RateLimiter{
		storage: ratelimit.NewStorage(requestsPerSecond, burst),
	}
}

/*
Limit returns a Gin middleware function that enforces per-client rate limiting.
Requests exceeding the limit receive HTTP 429 status with a Retry-After header.
*/
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()

		if !rl.storage.Allow(clientID) {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) Storage() *ratelimit.Storage {
	return rl.storage
}
