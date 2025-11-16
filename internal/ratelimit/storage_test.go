package ratelimit

import (
	"testing"
)

func TestNewStorage(t *testing.T) {
	storage := NewStorage(10, 50)
	
	if storage.requestsPerSec != 10 {
		t.Errorf("Expected requestsPerSec 10, got %d", storage.requestsPerSec)
	}
	
	if storage.burst != 50 {
		t.Errorf("Expected burst 50, got %d", storage.burst)
	}
	
	if storage.Count() != 0 {
		t.Errorf("Expected 0 clients initially, got %d", storage.Count())
	}
}

func TestStorageGetBucket(t *testing.T) {
	storage := NewStorage(10, 50)
	
	// Get bucket for client1
	bucket1 := storage.GetBucket("client1")
	if bucket1 == nil {
		t.Fatal("Expected bucket for client1")
	}
	
	// Should have 1 client now
	if storage.Count() != 1 {
		t.Errorf("Expected 1 client, got %d", storage.Count())
	}
	
	// Get same bucket again
	bucket1Again := storage.GetBucket("client1")
	if bucket1 != bucket1Again {
		t.Error("Should return same bucket for same client")
	}
	
	// Still should have 1 client
	if storage.Count() != 1 {
		t.Errorf("Expected 1 client, got %d", storage.Count())
	}
	
	// Get bucket for different client
	bucket2 := storage.GetBucket("client2")
	if bucket2 == nil {
		t.Fatal("Expected bucket for client2")
	}
	
	if bucket1 == bucket2 {
		t.Error("Different clients should have different buckets")
	}
	
	// Should have 2 clients now
	if storage.Count() != 2 {
		t.Errorf("Expected 2 clients, got %d", storage.Count())
	}
}

func TestStorageAllow(t *testing.T) {
	storage := NewStorage(10, 5)
	
	clientID := "192.168.1.1"
	
	// Should allow first 5 requests
	for i := 0; i < 5; i++ {
		if !storage.Allow(clientID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
	
	// 6th request should be denied
	if storage.Allow(clientID) {
		t.Error("Request 6 should be denied")
	}
}

func TestStorageMultipleClients(t *testing.T) {
	storage := NewStorage(10, 5)
	
	client1 := "192.168.1.1"
	client2 := "192.168.1.2"
	
	// Exhaust client1's tokens
	for i := 0; i < 5; i++ {
		storage.Allow(client1)
	}
	
	// client1 should be denied
	if storage.Allow(client1) {
		t.Error("client1 should be rate limited")
	}
	
	// client2 should still be allowed (separate bucket)
	if !storage.Allow(client2) {
		t.Error("client2 should be allowed (separate bucket)")
	}
}

func TestStorageClear(t *testing.T) {
	storage := NewStorage(10, 50)
	
	// Create buckets for multiple clients
	storage.GetBucket("client1")
	storage.GetBucket("client2")
	storage.GetBucket("client3")
	
	if storage.Count() != 3 {
		t.Errorf("Expected 3 clients, got %d", storage.Count())
	}
	
	// Clear all buckets
	storage.Clear()
	
	if storage.Count() != 0 {
		t.Errorf("Expected 0 clients after clear, got %d", storage.Count())
	}
}

func TestStorageConcurrent(t *testing.T) {
	storage := NewStorage(100, 50)
	
	// Test concurrent access from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		clientID := string(rune('A' + i))
		go func(id string) {
			for j := 0; j < 20; j++ {
				storage.Allow(id)
			}
			done <- true
		}(clientID)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Should have 10 clients
	if storage.Count() != 10 {
		t.Errorf("Expected 10 clients, got %d", storage.Count())
	}
}
