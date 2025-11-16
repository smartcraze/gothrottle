# GoThrottle - Rate Limiting Reverse Proxy

A high-performance reverse proxy with token bucket rate limiting built in Go using the Gin framework.

## Features

- **Path-Based Routing**: Route requests to different upstream backends based on URL path prefixes
- **Token Bucket Rate Limiting**: Per-client IP rate limiting with configurable burst capacity
- **Longest Prefix Matching**: Intelligent route matching for nested paths
- **Graceful Error Handling**: Proper HTTP status codes and error messages
- **Concurrent Safe**: Thread-safe rate limiting with efficient locking
- **Easy Configuration**: YAML-based configuration

## Quick Start

### Build

```bash
go build -o bin/proxy ./cmd/proxy/
```

### Run

```bash
./bin/proxy
```

The proxy will start on port 8080 (or the port specified in `configs/config.yaml`).

## Configuration

Edit `configs/config.yaml`:

```yaml
server:
  port: 8080

rate_limit:
  requests_per_second: 5    # Token refill rate
  burst: 10                 # Maximum burst capacity

routes:
  - path: "/api"
    target: "https://backend.com"
  - path: "/auth"
    target: "https://auth.service.com"
```

### Configuration Options

#### Server
- `port`: Server port (default: 8080)

#### Rate Limit
- `requests_per_second`: Number of tokens added per second (request rate)
- `burst`: Maximum number of tokens in the bucket (burst capacity)

#### Routes
- `path`: URL path prefix to match (e.g., `/api` matches `/api/users`, `/api/v1/data`, etc.)
- `target`: Upstream URL to proxy requests to

## How It Works

### Token Bucket Algorithm

Each client IP gets its own token bucket:

1. **Initial State**: Bucket starts with `burst` tokens
2. **Token Refill**: Tokens are added at `requests_per_second` rate
3. **Request Handling**: Each request consumes 1 token
4. **Rate Limiting**: If no tokens available, request is rejected with HTTP 429

**Example**: With `requests_per_second: 5` and `burst: 10`:
- Client can make 10 requests immediately (burst)
- After burst, client can make 5 requests per second
- Tokens refill continuously over time

### Path Matching

Routes are matched using **longest prefix matching**:

```yaml
routes:
  - path: "/api"
    target: "http://localhost:8000"
  - path: "/api/v2"
    target: "http://localhost:9000"
```

- Request to `/api/v2/users` → routes to `http://localhost:9000`
- Request to `/api/users` → routes to `http://localhost:8000`
- Request to `/unknown` → returns 404

## API Endpoints

### Health Check

```bash
curl http://localhost:8080/ping
```

Response:
```json
{
  "message": "pong",
  "status": "healthy"
}
```

### Proxied Requests

All other requests are proxied based on configured routes:

```bash
curl http://localhost:8080/api/users
# Proxied to: https://backend.com/api/users
```

### Rate Limit Response

When rate limit is exceeded:

```bash
curl http://localhost:8080/api/users
```

Response (HTTP 429):
```json
{
  "error": "rate limit exceeded",
  "message": "too many requests, please try again later"
}
```

Headers:
- `Retry-After: 1` (retry after 1 second)

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Run specific package tests:

```bash
go test -v ./internal/ratelimit/
go test -v ./internal/proxy/
go test -v ./internal/middleware/
```

## Project Structure

```
gothrottle/
├── cmd/
│   └── proxy/
│       └── main.go              # Application entry point
├── configs/
│   └── config.yaml              # Configuration file
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration structs
│   │   ├── loader.go            # YAML loader
│   │   └── loader_test.go       # Config tests
│   ├── middleware/
│   │   ├── logging.go           # Request logging
│   │   ├── ratelimit.go         # Rate limit middleware
│   │   └── ratelimit_test.go    # Middleware tests
│   ├── proxy/
│   │   ├── balancer.go          # Load balancer (round-robin)
│   │   ├── proxy.go             # Reverse proxy handler
│   │   └── proxy_test.go        # Proxy tests
│   └── ratelimit/
│       ├── limiter.go           # Token bucket implementation
│       ├── limiter_test.go      # Limiter tests
│       ├── storage.go           # Per-client storage
│       └── storage_test.go      # Storage tests
├── go.mod
├── go.sum
└── README.md
```

## Examples

### Example 1: Testing Rate Limits

```bash
# Make rapid requests to test rate limiting
for i in {1..15}; do
  echo "Request $i:"
  curl -s http://localhost:8080/api/test | jq
done
```

First 10 requests succeed (burst capacity), then rate limited.

### Example 2: Different Clients

```bash
# Client 1 - exhaust rate limit
for i in {1..15}; do
  curl -s http://localhost:8080/api/test
done

# Client 2 - still allowed (separate bucket)
curl -s -H "X-Forwarded-For: 192.168.1.2" http://localhost:8080/api/test
```

Each client IP has its own rate limit bucket.

### Example 3: Custom Configuration

Create a custom config for local testing:

```yaml
server:
  port: 3000

rate_limit:
  requests_per_second: 100
  burst: 200

routes:
  - path: "/api"
    target: "http://localhost:8000"
  - path: "/static"
    target: "http://localhost:8080"
```

## Performance Considerations

- **Concurrent Safe**: Uses `sync.RWMutex` for efficient read/write locking
- **Memory Efficient**: Only stores buckets for active clients
- **No External Dependencies**: In-memory storage (extensible to Redis)
- **Efficient Matching**: O(n) route matching where n is number of routes

## Future Enhancements

- [ ] Redis backend for distributed rate limiting
- [ ] Per-route rate limits (different limits for different paths)
- [ ] Health checks for upstream servers
- [ ] Metrics and monitoring (Prometheus)
- [ ] Circuit breaker pattern
- [ ] Request/response transformation
- [ ] Authentication and authorization
- [ ] WebSocket support

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
