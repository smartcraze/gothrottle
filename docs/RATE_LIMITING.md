# Rate Limiting Configuration

GoThrottle supports flexible rate limiting configuration with two time-based options.

## Configuration Options

You can configure rate limits using **either** `requests_per_second` **OR** `requests_per_minute` (not both).

### Option 1: Requests Per Second

For high-frequency rate limits, use `requests_per_second`:

```yaml
rate_limit:
  requests_per_second: 10  # Allow 10 requests per second
  burst: 20                # Allow 20 instant requests before limiting
```

**Use cases:**
- High-traffic APIs
- Services requiring sub-second precision
- Performance testing with high request rates

### Option 2: Requests Per Minute

For more restrictive or demonstration-friendly rate limits, use `requests_per_minute`:

```yaml
rate_limit:
  requests_per_minute: 6   # Allow 6 requests per minute (1 every 10 seconds)
  burst: 3                 # Allow 3 instant requests before limiting
```

**Use cases:**
- Strict rate limiting for resource-intensive operations
- Demo/testing scenarios where slower rates are easier to observe
- API endpoints with natural request intervals (e.g., every 10-30 seconds)

## How It Works

### Conversion

Internally, `requests_per_minute` is automatically converted to `requests_per_second`:

- `6 requests/minute` = `0.10 requests/second`
- `60 requests/minute` = `1.0 requests/second`

### Burst Capacity

The `burst` parameter defines how many tokens are initially available:

```yaml
rate_limit:
  requests_per_minute: 6
  burst: 3
```

With this configuration:
1. **First 3 requests**: Instant success (using burst tokens)
2. **Request 4+**: Limited to 1 request every 10 seconds (0.10 req/sec)

### Example Behavior

```bash
# Burst of 3, then rate limited
curl http://localhost:8080/api  # ✓ 200 OK (burst token 1)
curl http://localhost:8080/api  # ✓ 200 OK (burst token 2)
curl http://localhost:8080/api  # ✓ 200 OK (burst token 3)
curl http://localhost:8080/api  # ✗ 429 Too Many Requests (no tokens available)

# Wait 10 seconds for token refill
sleep 10
curl http://localhost:8080/api  # ✓ 200 OK (refilled 1 token)
```

## Validation Rules

1. **Must specify exactly one time-based option**:
   - ✓ `requests_per_second: 10`
   - ✓ `requests_per_minute: 60`
   - ✗ Both specified (validation error)
   - ✗ Neither specified (validation error)

2. **Values must be positive integers**:
   - ✓ `requests_per_second: 1`
   - ✗ `requests_per_second: 0` (validation error)
   - ✗ `requests_per_second: -5` (validation error)

3. **Burst must be positive**:
   - ✓ `burst: 1`
   - ✗ `burst: 0` (validation error)

## Complete Configuration Examples

### High-Traffic API

```yaml
server:
  port: 8080

rate_limit:
  requests_per_second: 100
  burst: 200

routes:
  - path: "/api"
    target: "http://backend:8000"
```

### Demonstration/Testing

```yaml
server:
  port: 8080

rate_limit:
  requests_per_minute: 6   # Easy to observe in demos
  burst: 3

routes:
  - path: "/api"
    target: "http://localhost:8000"
  - path: "/auth"
    target: "http://localhost:9000"
```

### Strict Resource Protection

```yaml
server:
  port: 8080

rate_limit:
  requests_per_minute: 30  # 1 request every 2 seconds
  burst: 5                 # Small burst allowance

routes:
  - path: "/expensive-operation"
    target: "http://compute-service:8000"
```

## Rate Limit Response

When rate limit is exceeded, clients receive:

**HTTP Status:** `429 Too Many Requests`

**Headers:**
```
Retry-After: 1
```

**Response Body:**
```json
{
  "error": "rate limit exceeded",
  "message": "too many requests, please try again later"
}
```

## Per-Client Rate Limiting

Rate limits are enforced **per client IP address**:

- Each client has an independent token bucket
- Client A's rate limit doesn't affect Client B
- Tokens refill continuously at the configured rate
- Useful for multi-tenant scenarios

## Internal Implementation

The rate limiter uses the **Token Bucket Algorithm**:

1. Each client gets a bucket with `burst` tokens initially
2. Tokens refill at `requests_per_second` rate
3. Each request consumes 1 token
4. Requests are denied when tokens = 0
5. Maximum tokens = `burst` (excess tokens are discarded)

This ensures:
- ✓ Smooth rate limiting
- ✓ Burst handling for traffic spikes
- ✓ Fair resource allocation
- ✓ Thread-safe concurrent access
