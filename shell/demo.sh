#!/bin/bash

# Demo script to show GoThrottle in action
# This script starts mock backend servers and demonstrates the proxy

echo "🎬 GoThrottle Demo"
echo "=================="
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    kill $BACKEND1_PID $BACKEND2_PID $PROXY_PID 2>/dev/null
    exit
}

trap cleanup EXIT INT TERM

# Start mock backend server 1 (simulates API backend)
echo "🔧 Starting mock backend server 1 on port 8001..."
python3 -m http.server 8001 > /dev/null 2>&1 &
BACKEND1_PID=$!

# Start mock backend server 2 (simulates Auth backend)
echo "🔧 Starting mock backend server 2 on port 8002..."
python3 -m http.server 8002 > /dev/null 2>&1 &
BACKEND2_PID=$!

# Wait for backends to start
sleep 1

# Update config to point to local backends
echo "📝 Creating demo config..."
cat > configs/config-demo.yaml << EOF
server:
  port: 8080

rate_limit:
  requests_per_second: 2
  burst: 5

routes:
  - path: "/api"
    target: "http://localhost:8001"
  - path: "/auth"
    target: "http://localhost:8002"
EOF

# Build the proxy
echo "📦 Building proxy..."
go build -o bin/proxy ./cmd/proxy/ > /dev/null 2>&1

# Start the proxy (temporarily use demo config)
echo "🚀 Starting GoThrottle proxy on port 8080..."
cp configs/config.yaml configs/config.yaml.bak
cp configs/config-demo.yaml configs/config.yaml
./bin/proxy > /tmp/proxy.log 2>&1 &
PROXY_PID=$!

# Wait for proxy to start
sleep 2

echo ""
echo "✅ All services running!"
echo ""
echo "================================"
echo "Demo 1: Health Check"
echo "================================"
echo "$ curl http://localhost:8080/ping"
curl -s http://localhost:8080/ping | jq
echo ""

echo "================================"
echo "Demo 2: Proxy to Backend 1 (/api)"
echo "================================"
echo "$ curl http://localhost:8080/api/"
echo "Response:"
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" http://localhost:8080/api/
echo ""

echo "================================"
echo "Demo 3: Rate Limiting"
echo "================================"
echo "Making 8 rapid requests (burst=5, rate=2/sec)..."
echo ""

for i in {1..8}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/)
    if [ "$STATUS" == "200" ] || [ "$STATUS" == "404" ]; then
        echo "Request $i: ✅ Allowed (HTTP $STATUS)"
    else
        echo "Request $i: ❌ Rate Limited (HTTP $STATUS)"
    fi
done

echo ""
echo "================================"
echo "Demo 4: Per-Client Rate Limiting"
echo "================================"
echo "Making requests from different IPs..."
echo ""

# Client 1 (default)
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/)
echo "Client 1 (127.0.0.1): HTTP $STATUS"

# Client 2 (simulated with X-Forwarded-For)
STATUS=$(curl -s -H "X-Forwarded-For: 192.168.1.100" -o /dev/null -w "%{http_code}" http://localhost:8080/api/)
echo "Client 2 (192.168.1.100): HTTP $STATUS (separate rate limit bucket)"

echo ""
echo "================================"
echo "💡 Tip: Check proxy logs"
echo "================================"
echo "$ tail -20 /tmp/proxy.log"
echo ""
tail -20 /tmp/proxy.log | grep -E "(Loaded|Route|Starting|GET|POST)" || echo "(no logs yet)"

echo ""
echo "================================"
echo "Press Ctrl+C to stop demo"
echo "================================"

# Keep running
wait
