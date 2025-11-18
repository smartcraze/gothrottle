#!/bin/bash

echo "Starting GoThrottle Demo Environment"

# Build all servers
echo "Building servers..."
go build -o bin/proxy ./cmd/proxy/
go build -o bin/api-server ./examples/api-server
go build -o bin/auth-server ./examples/auth-server

# Start backend servers in background
echo "Starting API Backend Server (port 8000)..."
./bin/api-server &
API_PID=$!

sleep 1

echo "Starting Auth Service Server (port 9000)..."
./bin/auth-server &
AUTH_PID=$!

sleep 1

echo "Starting GoThrottle Reverse Proxy (port 8080)..."

./bin/proxy

cleanup() {
    echo ""
    echo " Shutting down servers..."
    kill $API_PID $AUTH_PID 2>/dev/null
    exit 0
}

trap cleanup INT TERM