package ratelimit

import (
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	tb := NewTokenBucket(10, 50)
	
	if tb.maxTokens != 50 {
		t.Errorf("Expected maxTokens 50, got %f", tb.maxTokens)
	}
	
	if tb.refillRate != 10 {
		t.Errorf("Expected refillRate 10, got %f", tb.refillRate)
	}
	
	if tb.tokens != 50 {
		t.Errorf("Expected initial tokens 50, got %f", tb.tokens)
	}
}

func TestTokenBucketAllow(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	
	// Should allow first 5 requests (burst capacity)
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
	
	// 6th request should be denied (no tokens left)
	if tb.Allow() {
		t.Error("Request 6 should be denied (no tokens)")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	
	// Exhaust all tokens
	for i := 0; i < 5; i++ {
		tb.Allow()
	}
	
	// Should be denied immediately
	if tb.Allow() {
		t.Error("Should be denied when no tokens available")
	}
	
	// Wait 200ms (should refill ~2 tokens at 10 tokens/sec)
	time.Sleep(200 * time.Millisecond)
	
	// Should allow 2 requests
	if !tb.Allow() {
		t.Error("Should allow request after refill")
	}
	if !tb.Allow() {
		t.Error("Should allow second request after refill")
	}
	
	// Third should be denied
	if tb.Allow() {
		t.Error("Third request should be denied")
	}
}

func TestTokenBucketMaxCapacity(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	
	// Wait for refill (tokens should cap at maxTokens)
	time.Sleep(2 * time.Second)
	
	tokens := tb.Tokens()
	if tokens > 5.0 {
		t.Errorf("Tokens should not exceed maxTokens (5), got %f", tokens)
	}
}

func TestTokenBucketReset(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	
	// Exhaust tokens
	for i := 0; i < 5; i++ {
		tb.Allow()
	}
	
	// Reset
	tb.Reset()
	
	// Should have full capacity again
	tokens := tb.Tokens()
	if tokens != 5.0 {
		t.Errorf("Expected 5 tokens after reset, got %f", tokens)
	}
}

func TestTokenBucketConcurrent(t *testing.T) {
	tb := NewTokenBucket(100, 50)
	
	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				tb.Allow()
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should have processed 100 requests without race conditions
	// (50 from burst + some refilled during execution)
	// Just verify no panic occurred
}
