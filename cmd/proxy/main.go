package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartcraze/gothrottle/internal/config"
	"github.com/smartcraze/gothrottle/internal/middleware"
	"github.com/smartcraze/gothrottle/internal/proxy"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Loaded configuration:")
	log.Printf("  Server port: %d", cfg.Server.Port)
	log.Printf("  Rate limit: %d req/sec, burst: %d", cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	for i, route := range cfg.Routes {
		log.Printf("  Route %d: %s -> %s", i+1, route.Path, route.Target)
	}

	// Initialize Gin router (without default middleware)
	r := gin.New()

	// Add custom middleware
	r.Use(gin.Recovery()) // Panic recovery
	r.Use(middleware.Logger()) // Custom logging

	// Initialize rate limiter middleware
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	r.Use(rateLimiter.Limit())

	// Health check endpoint (before rate limiting)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "healthy",
		})
	})

	// Initialize proxy handler
	proxyHandler, err := proxy.NewHandler(cfg.Routes)
	if err != nil {
		log.Fatalf("Failed to create proxy handler: %v", err)
	}

	// Route all other requests through the proxy
	r.NoRoute(proxyHandler.Handle)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting reverse proxy server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}



