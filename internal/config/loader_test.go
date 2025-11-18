package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test loading the actual config file
	configPath := filepath.Join("..", "..", "configs", "config.yaml")
	
	config, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify server config
	if config.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Server.Port)
	}

	// Verify rate limit config (now uses requests_per_minute)
	if config.RateLimit.RequestsPerMinute != 6 {
		t.Errorf("Expected requests_per_minute 6, got %d", config.RateLimit.RequestsPerMinute)
	}
	if config.RateLimit.Burst != 3 {
		t.Errorf("Expected burst 3, got %d", config.RateLimit.Burst)
	}

	// Verify routes
	if len(config.Routes) != 2 {
		t.Fatalf("Expected 2 routes, got %d", len(config.Routes))
	}

	// Check first route
	if config.Routes[0].Path != "/api" {
		t.Errorf("Expected path '/api', got '%s'", config.Routes[0].Path)
	}
	if config.Routes[0].Target != "http://localhost:8000" {
		t.Errorf("Expected target 'http://localhost:8000', got '%s'", config.Routes[0].Target)
	}

	// Check second route
	if config.Routes[1].Path != "/auth" {
		t.Errorf("Expected path '/auth', got '%s'", config.Routes[1].Path)
	}
	if config.Routes[1].Target != "http://localhost:9000" {
		t.Errorf("Expected target 'http://localhost:9000', got '%s'", config.Routes[1].Target)
	}

	t.Logf("✓ Config loaded successfully:")
	t.Logf("  Server Port: %d", config.Server.Port)
	if config.RateLimit.RequestsPerSecond > 0 {
		t.Logf("  Rate Limit: %d req/sec, burst: %d", config.RateLimit.RequestsPerSecond, config.RateLimit.Burst)
	} else {
		t.Logf("  Rate Limit: %d req/min, burst: %d", config.RateLimit.RequestsPerMinute, config.RateLimit.Burst)
	}
	for i, route := range config.Routes {
		t.Logf("  Route %d: %s -> %s", i+1, route.Path, route.Target)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError bool
	}{
		{
			name: "valid config",
			config: `
server:
  port: 8080
rate_limit:
  requests_per_second: 10
  burst: 50
routes:
  - path: "/api"
    target: "http://localhost:8000"
`,
			expectError: false,
		},
		{
			name: "missing routes",
			config: `
rate_limit:
  requests_per_second: 10
  burst: 50
`,
			expectError: true,
		},
		{
			name: "invalid requests_per_second",
			config: `
rate_limit:
  requests_per_second: 0
  burst: 50
routes:
  - path: "/api"
    target: "http://localhost:8000"
`,
			expectError: true,
		},
		{
			name: "empty path",
			config: `
rate_limit:
  requests_per_second: 10
  burst: 50
routes:
  - path: ""
    target: "http://localhost:8000"
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpFile, err := os.CreateTemp("", "config-*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.config); err != nil {
				t.Fatal(err)
			}
			tmpFile.Close()

			// Test loading
			_, err = Load(tmpFile.Name())
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	// Test that default port is set when not specified
	config := `
rate_limit:
  requests_per_second: 10
  burst: 50
routes:
  - path: "/api"
    target: "http://localhost:8000"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(config); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}
}
