#!/bin/bash

# MyGit Local Development Test Script
# Prerequisites: Go 1.25+, git

set -e

echo "MyGit Local Development Test"
echo "============================="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.25+ first."
    echo "Download from: https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oP '\d+\.\d+' | head -1)
echo "Go version: $GO_VERSION"
echo

# Build the application
echo "Building MyGit..."
go build -o mygit ./src/main.go
echo "Build successful"
echo

# Set up directories
TEST_DIR="./test-repos"
mkdir -p "$TEST_DIR"

# Copy web templates to expected location
mkdir -p /data/web/templates
mkdir -p /data/web/static
cp -r web/templates/* /data/web/templates/ 2>/dev/null || true
cp -r web/static/* /data/web/static/ 2>/dev/null || true

# Configuration
HTTP_PORT=${HTTP_PORT:-3000}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-admin}
REPO_STORAGE=${REPO_STORAGE:-$TEST_DIR}

echo "Configuration:"
echo "  Port: $HTTP_PORT"
echo "  User: $ADMIN_USERNAME"
echo "  Password: $ADMIN_PASSWORD"
echo "  Storage: $(realpath $REPO_STORAGE)"
echo

# Export for the app
export HTTP_PORT ADMIN_USERNAME ADMIN_PASSWORD REPO_STORAGE

# Start the application
echo "Starting MyGit..."
./mygit &
APP_PID=$!
echo "PID: $APP_PID"

# Wait for startup
sleep 2

# Test basic connectivity
echo
echo "Running tests..."

# Test 1: 401 without auth
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$HTTP_PORT/)
if [ "$RESPONSE" = "401" ]; then
    echo "  Test 1: 401 Unauthorized - expected"
else
    echo "  Test 1: Unexpected response: $RESPONSE"
fi

# Test 2: Authenticated
RESPONSE=$(curl -s -u "$ADMIN_USERNAME:$ADMIN_PASSWORD" -w "%{http_code}" http://localhost:$HTTP_PORT/)
if [ "$RESPONSE" = "200" ]; then
    echo "  Test 2: Authenticated - OK"
else
    echo "  Test 2: Failed: $RESPONSE"
fi

# Test 3: Repos endpoint
RESPONSE=$(curl -s -u "$ADMIN_USERNAME:$ADMIN_PASSWORD" -w "%{http_code}" http://localhost:$HTTP_PORT/repos)
if [ "$RESPONSE" = "200" ]; then
    echo "  Test 3: Repos endpoint - OK"
else
    echo "  Test 3: Failed: $RESPONSE"
fi

echo
echo "Server running at http://localhost:$HTTP_PORT"
echo "Press Ctrl+C to stop"

# Wait for interrupt
trap "kill $APP_PID 2>/dev/null; rm -f mygit; exit" EXIT INT TERM

wait $APP_PID
