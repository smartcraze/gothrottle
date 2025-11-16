#!/bin/bash

# Integration test script for GoThrottle

set -e

echo "🚀 Starting GoThrottle Integration Tests"
echo "========================================"

# Build the application
echo ""
echo "📦 Building application..."
go build -o bin/proxy ./cmd/proxy/

# Run unit tests
echo ""
echo "🧪 Running unit tests..."
go test ./... -v

echo ""
echo "✅ All tests passed!"
echo ""
echo "📊 Test Coverage:"
go test ./... -cover

echo ""
echo "========================================"
echo "✨ Integration tests complete!"
echo ""
echo "To run the proxy server:"
echo "  ./bin/proxy"
echo ""
echo "To test manually:"
echo "  # Health check"
echo "  curl http://localhost:8080/ping"
echo ""
echo "  # Test rate limiting (make 15 rapid requests)"
echo "  for i in {1..15}; do echo \"Request \$i:\"; curl -s http://localhost:8080/api/test; echo; done"
