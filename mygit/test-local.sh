#!/bin/bash

# MyGit Local Development Test Script
# Prerequisites: Go 1.21+, git

set -e

echo "MyGit Local Development Test"
echo "============================="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.21+ first."
    echo "Download from: https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oP '\d+\.\d+' | head -1)
echo "Go version: $GO_VERSION"
echo

# Build the application
echo "Building MyGit..."
go build -o mygit .
echo "Build successful"
echo

# Set up directories
TEST_DIR="./test-repos"
mkdir -p "$TEST_DIR"

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

# Test 1: UI should be public (200)
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$HTTP_PORT/)
if [ "$RESPONSE" = "200" ]; then
    echo "  Test 1: UI served - OK"
else
    echo "  Test 1: Unexpected response: $RESPONSE"
fi

# Test 2: UI assets should be public (200)
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$HTTP_PORT/ui/styles.css)
if [ "$RESPONSE" = "200" ]; then
    echo "  Test 2: UI styles.css - OK"
else
    echo "  Test 2: Failed: $RESPONSE"
fi

# Test 3: API requires auth (401)
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$HTTP_PORT/api/v1/repos)
if [ "$RESPONSE" = "401" ]; then
    echo "  Test 3: API auth required - OK"
else
    echo "  Test 3: Unexpected response: $RESPONSE"
fi

# Test 4: Authenticated API access
RESPONSE=$(curl -s -w "\n%{http_code}" -u "$ADMIN_USERNAME:$ADMIN_PASSWORD" http://localhost:$HTTP_PORT/api/v1/repos)
STATUS=$(echo "$RESPONSE" | tail -n1)
if [ "$STATUS" = "200" ]; then
    echo "  Test 4: Authenticated API - OK"
else
    echo "  Test 4: Failed: $STATUS"
fi

echo
echo "Server running at http://localhost:$HTTP_PORT"
echo "Press Ctrl+C to stop"

# Wait for interrupt
trap "kill $APP_PID 2>/dev/null; rm -f mygit; exit" EXIT INT TERM

wait $APP_PID
