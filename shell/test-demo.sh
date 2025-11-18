#!/bin/bash

echo "🎬 GoThrottle Demo Test Script"
echo "=============================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "📍 Demo 1: Health Check"
echo "----------------------"
curl -s http://localhost:8080/ping | jq '.'
echo ""
sleep 2

echo "📍 Demo 2: Route to API Backend (/api)"
echo "---------------------------------------"
curl -s http://localhost:8080/api/users | jq '.'
echo ""
sleep 2

echo "📍 Demo 3: Route to Auth Service (/auth)"
echo "-----------------------------------------"
curl -s http://localhost:8080/auth/login | jq '.'
echo ""
sleep 2

echo "📍 Demo 4: Rate Limiting Test (15 rapid requests)"
echo "--------------------------------------------------"
echo "Sending 15 requests in rapid succession..."
echo ""

for i in {1..15}; do
    response=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/test)
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$status_code" = "200" ]; then
        echo -e "${GREEN}Request $i: ✅ Status $status_code${NC}"
    else
        echo -e "${RED}Request $i: ❌ Status $status_code - Rate Limited!${NC}"
        echo "  Response: $body"
    fi
    
    sleep 0.1
done

echo ""
echo "📍 Demo 5: Different Client IPs (Isolation Test)"
echo "-------------------------------------------------"
echo "Client A exhausts limit..."
for i in {1..12}; do
    curl -s http://localhost:8080/api/client-a > /dev/null
done

echo "Client A (exhausted):"
response_a=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/client-a)
echo "$response_a" | head -n-1 | jq -r '.error // .message'

echo ""
echo "Client B (fresh IP):"
response_b=$(curl -s -H "X-Forwarded-For: 192.168.1.100" http://localhost:8080/api/client-b)
echo "$response_b" | jq -r '.message'

echo ""
echo "📍 Demo 6: Token Refill Test"
echo "----------------------------"
echo "Exhausting tokens..."
for i in {1..12}; do
    curl -s http://localhost:8080/api/refill-test > /dev/null
done

echo "Immediately after (should fail):"
curl -s http://localhost:8080/api/refill-test | jq -r '.error // .message'

echo ""
echo "Waiting 1 second for token refill..."
sleep 1

echo "After 1 second (tokens refilled):"
curl -s http://localhost:8080/api/refill-test | jq -r '.message'

echo ""
echo "✅ Demo Complete!"