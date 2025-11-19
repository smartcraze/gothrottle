#!/bin/bash

# Kill any existing proxy
pkill -f "go run cmd/proxy" 2>/dev/null
sleep 1

# Start proxy in background
cd /home/suraj/projects/gothrottle
go run cmd/proxy/main.go > /tmp/proxy.log 2>&1 &
PROXY_PID=$!
sleep 2

echo "=========================================="
echo "Rate Limit Test: 20 req/min, burst: 4"
echo "Expected: 0.333 req/sec = 1 req every 3 sec"
echo "=========================================="
echo ""

# Make rapid requests
for i in {1..8}; do
  response=$(curl -s -w "\n%{http_code}" http://localhost:8080/ping 2>&1)
  code=$(echo "$response" | tail -1)
  timestamp=$(date +%H:%M:%S.%N | cut -b1-12)
  
  if [ "$code" = "200" ]; then
    echo "[$timestamp] Request $i: ✓ ALLOWED (HTTP $code)"
  else
    echo "[$timestamp] Request $i: ✗ BLOCKED (HTTP $code)"
  fi
  sleep 0.3
done

echo ""
echo "Waiting 4 seconds for token refill (should get 1 token)..."
sleep 4
echo ""

# Test after waiting
for i in {9..11}; do
  response=$(curl -s -w "\n%{http_code}" http://localhost:8080/ping 2>&1)
  code=$(echo "$response" | tail -1)
  timestamp=$(date +%H:%M:%S.%N | cut -b1-12)
  
  if [ "$code" = "200" ]; then
    echo "[$timestamp] Request $i: ✓ ALLOWED (HTTP $code)"
  else
    echo "[$timestamp] Request $i: ✗ BLOCKED (HTTP $code)"
  fi
  sleep 0.3
done

# Cleanup
kill $PROXY_PID 2>/dev/null
echo ""
echo "Test complete!"
