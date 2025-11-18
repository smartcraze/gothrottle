package ratelimit

import (
	"sync"
	"time"
)

/*
TokenBucket implements the token bucket algorithm for rate limiting.
Tokens are continuously added at a fixed rate and consumed by incoming requests.
When tokens are exhausted, requests are rejected until new tokens become available.
*/
type TokenBucket struct {
	mu             sync.Mutex
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
}

func NewTokenBucket(requestsPerSecond float64, burst int) *TokenBucket {
	return &TokenBucket{
		tokens:         float64(burst),
		maxTokens:      float64(burst),
		refillRate:     requestsPerSecond,
		lastRefillTime: time.Now(),
	}
}

/*
Allow checks if a request is allowed based on token availability.
It refills tokens based on elapsed time, then attempts to consume one token.
Returns true if the request is allowed, false if rate limit is exceeded.
*/
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefillTime = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

func (tb *TokenBucket) Tokens() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefillTime = now
	return tb.tokens
}

func (tb *TokenBucket) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.tokens = tb.maxTokens
	tb.lastRefillTime = time.Now()
}