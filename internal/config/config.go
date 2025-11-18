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
Supports either requests_per_second OR requests_per_minute (not both).
Burst defines the maximum tokens available for handling traffic spikes.
*/
type RateLimit struct {
	RequestsPerSecond int `yaml:"requests_per_second"`
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

/*
GetRequestsPerSecond converts the rate limit to requests per second.
If requests_per_minute is specified, it converts to seconds.
*/
func (r *RateLimit) GetRequestsPerSecond() float64 {
	if r.RequestsPerSecond > 0 {
		return float64(r.RequestsPerSecond)
	}
	if r.RequestsPerMinute > 0 {
		return float64(r.RequestsPerMinute) / 60.0
	}
	return 1.0
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

