package config


type Route struct {
	Path   string `yaml:"path"`   
	Target string `yaml:"target"` 
}

type RateLimit struct {
	RequestsPerSecond int `yaml:"requests_per_second"` 
	Burst             int `yaml:"burst"`              
}

type ServerConfig struct {
	Port int `yaml:"port"` 
}

type Config struct {
	Routes     []Route       `yaml:"routes"`      
	RateLimit  RateLimit     `yaml:"rate_limit"`  
	Server     ServerConfig  `yaml:"server"`      
}

