package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implements the token bucket algorithm for rate limiting
type TokenBucket struct {
	mu              sync.Mutex
	tokens          float64   // Current number of tokens
	maxTokens       float64   // Maximum tokens (burst capacity)
	refillRate      float64   // Tokens added per second
	lastRefillTime  time.Time // Last time tokens were refilled
}

// NewTokenBucket creates a new token bucket with the specified rate and burst capacity
func NewTokenBucket(requestsPerSecond int, burst int) *TokenBucket {
	return &TokenBucket{
		tokens:         float64(burst),
		maxTokens:      float64(burst),
		refillRate:     float64(requestsPerSecond),
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if available
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tb.tokens += elapsed * tb.refillRate

	// Cap tokens at maximum
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefillTime = now

	// Check if we have tokens available
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

// Tokens returns the current number of tokens (for testing/monitoring)
func (tb *TokenBucket) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tb.tokens += elapsed * tb.refillRate

	// Cap tokens at maximum
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefillTime = now
	return tb.tokens
}

// Reset resets the bucket to full capacity (for testing)
func (tb *TokenBucket) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.tokens = tb.maxTokens
	tb.lastRefillTime = time.Now()
}