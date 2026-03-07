#!/bin/bash

# MyGit Local Development Test Script
# Prerequisite: Go 1.25+ must be installed

echo "🧪 MyGit Local Development Test Script"
echo "===================================="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed. Please install Go 1.25+ first."
    echo "Download from: https://go.dev/dl/"
    exit 1
fi

# Get Go version
GO_VERSION=$(go version | grep -oP '\d+\.\d+')
if [[ "$GO_VERSION" < "1.25" ]]; then
    echo "⚠️  Warning: Go version $GO_VERSION detected. Recommend Go 1.25+ for best compatibility."
fi

echo "✅ Go $GO_VERSION detected"
echo

# Build the application
echo "🔨 Building MyGit..."
go build -o mygit ./src/main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed. Please check for compilation errors."
    exit 1
fi
echo "✅ Build successful"
echo

# Set up test directory
TEST_DIR="./test-repos"
mkdir -p "$TEST_DIR"
chmod 755 "$TEST_DIR"
echo "📁 Test repository directory: $(realpath $TEST_DIR)"
echo

# Configuration
HTTP_PORT=${HTTP_PORT:-3000}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-}

# Display configuration
echo "📋 Test Configuration:"
echo "  Port: $HTTP_PORT"
echo "  Admin User: $ADMIN_USERNAME"
echo "  Admin Password: ${ADMIN_PASSWORD:-[NOT SET - will use default]}"
echo "  Repository Storage: $(realpath $TEST_DIR)"
echo

# Start the application in background
echo "🚀 Starting MyGit..."
export HTTP_PORT
export ADMIN_USERNAME
if [ -n "$ADMIN_PASSWORD" ]; then
    export ADMIN_PASSWORD
    echo "  Authentication: Configured (using provided password)"
else
    echo "  Authentication: Default (admin/admin)"
fi
export REPO_STORAGE="$TEST_DIR"

./mygit &
APP_PID=$!
echo "  PID: $APP_PID"
echo

# Wait for startup
sleep 2

# Test basic connectivity
echo "🧪 Running Tests..."
echo

# Test 1: Basic connectivity (should get 401 without auth)
echo "Test 1: Basic connectivity"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$HTTP_PORT/)
if [ "$RESPONSE" = "401" ]; then
    echo "  ✅ Server responding (401 Unauthorized - expected)"
else
    echo "  ❌ Unexpected response: $RESPONSE"
fi

# Test 2: Authentication (using default admin/admin if no password set)
echo "Test 2: Authentication"
if [ -n "$ADMIN_PASSWORD" ]; then
    AUTH="admin:$ADMIN_PASSWORD"
else
    AUTH="admin:admin"
fi

RESPONSE=$(curl -s -u "$AUTH" -w "%{http_code}" http://localhost:$HTTP_PORT/)
if [ "$RESPONSE" = "200" ]; then
    echo "  ✅ Authentication successful"
else
    echo "  ❌ Authentication failed: $RESPONSE"
fi

# Test 3: Repos endpoint
echo "Test 3: Repos endpoint"
RESPONSE=$(curl -s -u "$AUTH" -w "%{http_code}" http://localhost:$HTTP_PORT/repos)
if [ "$RESPONSE" = "200" ]; then
    echo "  ✅ Repos endpoint working"
else
    echo "  ❌ Repos endpoint failed: $RESPONSE"
fi

echo
# Test 4: Git protocol (basic check)
echo "Test 4: Git protocol"
RESPONSE=$(curl -s -u "$AUTH" -w "%{http_code}" "http://localhost:$HTTP_PORT/test.git/info/refs?service=git-upload-pack")
if [ "$RESPONSE" = "200" ] || [ "$RESPONSE" = "404" ]; then
    echo "  ✅ Git protocol accessible"
else
    echo "  ❌ Git protocol issue: $RESPONSE"
fi

echo
echo "📊 Test Summary:"
echo "  All basic tests completed. Check output above for details."
echo

# Cleanup
echo "🧹 Cleaning up..."
kill $APP_PID 2>/dev/null
wait $APP_PID 2>/dev/null
rm -f mygit

echo "✅ Test script completed"
echo
echo "Next steps:"
echo "  • Run with password: ADMIN_PASSWORD=yourpass ./test-local.sh"
echo "  • Manual testing: ./mygit (after setting env vars)"
echo "  • See DEVELOPMENT.md for more options"