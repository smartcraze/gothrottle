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
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	requestsPerSec := cfg.RateLimit.GetRequestsPerSecond()

	log.Printf("Loaded configuration:")
	log.Printf("  Server port: %d", cfg.Server.Port)
	if cfg.RateLimit.RequestsPerSecond > 0 {
		log.Printf("  Rate limit: %d req/sec, burst: %d", cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	} else {
		log.Printf("  Rate limit: %d req/min (%.2f req/sec), burst: %d", 
			cfg.RateLimit.RequestsPerMinute, requestsPerSec, cfg.RateLimit.Burst)
	}
	for i, route := range cfg.Routes {
		log.Printf("  Route %d: %s -> %s", i+1, route.Path, route.Target)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	rateLimiter := middleware.NewRateLimiter(requestsPerSec, cfg.RateLimit.Burst)
	r.Use(rateLimiter.Limit())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "healthy",
		})
	})

	proxyHandler, err := proxy.NewHandler(cfg.Routes)
	if err != nil {
		log.Fatalf("Failed to create proxy handler: %v", err)
	}

	r.NoRoute(proxyHandler.Handle)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Starting reverse proxy server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}



