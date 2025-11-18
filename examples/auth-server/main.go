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
        log.Printf("[AUTH-SERVER] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
        log.Printf("[AUTH-SERVER] X-Forwarded-For: %s", r.Header.Get("X-Forwarded-For"))

        // Simulate processing time
        time.Sleep(150 * time.Millisecond)

        // Response
        response := map[string]interface{}{
            "server":    "Auth Service",
            "path":      r.URL.Path,
            "timestamp": time.Now().Format(time.RFC3339),
            "message":   "Successfully proxied to Auth server",
            "auth": map[string]interface{}{
                "authenticated": true,
                "user":          "demo_user",
                "token":         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            },
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    })

    port := 9000
    log.Printf("🔐 Auth Service Server starting on port %d", port)
    log.Printf("📍 Try: curl http://localhost:%d/auth/login", port)
    
    if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}