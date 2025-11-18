package config

/*
Route represents a path-based routing rule that maps incoming request paths
to upstream backend targets.
*/
type Route struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
}

/*
RateLimit configures the token bucket rate limiting parameters.
RequestsPerSecond determines how fast tokens refill, while Burst
defines the maximum tokens available for handling traffic spikes.
*/
type RateLimit struct {
	RequestsPerSecond int `yaml:"requests_per_second"`
	Burst             int `yaml:"burst"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

/*
Config holds the complete application configuration including server settings,
rate limiting parameters, and routing rules.
*/
type Config struct {
	Routes    []Route      `yaml:"routes"`
	RateLimit RateLimit    `yaml:"rate_limit"`
	Server    ServerConfig `yaml:"server"`
}

