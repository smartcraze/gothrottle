package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Log incoming request
        log.Printf("[API-SERVER] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
        log.Printf("[API-SERVER] X-Forwarded-For: %s", r.Header.Get("X-Forwarded-For"))

        // Simulate processing time
        time.Sleep(100 * time.Millisecond)

        // Response
        response := map[string]interface{}{
            "server":    "API Backend",
            "path":      r.URL.Path,
            "timestamp": time.Now().Format(time.RFC3339),
            "message":   "Successfully proxied to API server",
            "data": map[string]interface{}{
                "users": []string{"Alice", "Bob", "Charlie"},
                "count": 3,
            },
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    })

    port := 8000
    log.Printf("🚀 API Backend Server starting on port %d", port)
    log.Printf("📍 Try: curl http://localhost:%d/api/users", port)
    
    if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}