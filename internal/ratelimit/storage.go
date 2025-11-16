package ratelimit

import (
	"sync"
)

// Storage manages rate limiters for multiple clients
type Storage struct {
	mu              sync.RWMutex
	buckets         map[string]*TokenBucket
	requestsPerSec  int
	burst           int
}

// NewStorage creates a new storage instance for rate limiters
func NewStorage(requestsPerSecond, burst int) *Storage {
	return &Storage{
		buckets:        make(map[string]*TokenBucket),
		requestsPerSec: requestsPerSecond,
		burst:          burst,
	}
}

// GetBucket retrieves or creates a token bucket for a client
func (s *Storage) GetBucket(clientID string) *TokenBucket {
	// Try read lock first for better performance
	s.mu.RLock()
	bucket, exists := s.buckets[clientID]
	s.mu.RUnlock()

	if exists {
		return bucket
	}

	// Need to create new bucket, acquire write lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check in case another goroutine created it
	if bucket, exists := s.buckets[clientID]; exists {
		return bucket
	}

	// Create new bucket
	bucket = NewTokenBucket(s.requestsPerSec, s.burst)
	s.buckets[clientID] = bucket
	return bucket
}

// Allow checks if a request from the client is allowed
func (s *Storage) Allow(clientID string) bool {
	bucket := s.GetBucket(clientID)
	return bucket.Allow()
}

// Count returns the number of tracked clients
func (s *Storage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.buckets)
}

// Clear removes all stored buckets (for testing)
func (s *Storage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets = make(map[string]*TokenBucket)
}