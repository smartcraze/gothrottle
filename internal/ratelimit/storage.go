package ratelimit

import (
	"sync"
)

/*
Storage manages token buckets for multiple clients, providing thread-safe
access to per-client rate limiters. Each client is identified by their IP address.
*/
type Storage struct {
	mu             sync.RWMutex
	buckets        map[string]*TokenBucket
	requestsPerSec int
	burst          int
}

func NewStorage(requestsPerSecond, burst int) *Storage {
	return &Storage{
		buckets:        make(map[string]*TokenBucket),
		requestsPerSec: requestsPerSecond,
		burst:          burst,
	}
}

/*
GetBucket retrieves or creates a token bucket for a client using double-checked
locking pattern for optimal performance in concurrent environments.
*/
func (s *Storage) GetBucket(clientID string) *TokenBucket {
	s.mu.RLock()
	bucket, exists := s.buckets[clientID]
	s.mu.RUnlock()

	if exists {
		return bucket
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if bucket, exists := s.buckets[clientID]; exists {
		return bucket
	}

	bucket = NewTokenBucket(s.requestsPerSec, s.burst)
	s.buckets[clientID] = bucket
	return bucket
}

func (s *Storage) Allow(clientID string) bool {
	bucket := s.GetBucket(clientID)
	return bucket.Allow()
}

func (s *Storage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.buckets)
}

func (s *Storage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets = make(map[string]*TokenBucket)
}