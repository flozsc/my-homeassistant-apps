#!/bin/bash

# MyGit Automated API Test Script
# Prerequisites: Go 1.21+, git, curl

set -e

# Configuration
HTTP_PORT=${HTTP_PORT:-3456}
ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
ADMIN_PASSWORD=${ADMIN_PASSWORD:-secret123}
REPO_STORAGE=${REPO_STORAGE:-./test-repos-auto}
BINARY_NAME="mygit-test"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

cleanup() {
    echo
    echo "Cleaning up..."
    if [ -n "$APP_PID" ] && kill -0 "$APP_PID" 2>/dev/null; then
        kill "$APP_PID" 2>/dev/null || true
        wait "$APP_PID" 2>/dev/null || true
    fi
    rm -rf "$REPO_STORAGE"
    rm -f "$BINARY_NAME"
}

trap cleanup EXIT

# Check prerequisites
check_prereqs() {
    echo "Checking prerequisites..."

    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi

    if ! command -v curl &> /dev/null; then
        echo -e "${RED}Error: curl is not installed${NC}"
        exit 1
    fi

    if ! command -v git &> /dev/null; then
        echo -e "${RED}Error: git is not installed${NC}"
        exit 1
    fi

    echo -e "${GREEN}Prerequisites OK${NC}"
}

# Build the application
build() {
    echo
    echo "Building MyGit..."
    go build -o "$BINARY_NAME" .
    echo -e "${GREEN}Build successful${NC}"
}

# Start the server
start_server() {
    echo
    echo "Starting server on port $HTTP_PORT..."

    # Clean up any existing test repos
    rm -rf "$REPO_STORAGE"
    mkdir -p "$REPO_STORAGE"

    HTTP_PORT="$HTTP_PORT" \
    ADMIN_USERNAME="$ADMIN_USERNAME" \
    ADMIN_PASSWORD="$ADMIN_PASSWORD" \
    REPO_STORAGE="$REPO_STORAGE" \
    ./"$BINARY_NAME" &

    APP_PID=$!
    echo "Server PID: $APP_PID"

    # Wait for server to start
    sleep 2

    # Verify server is running
    if ! kill -0 "$APP_PID" 2>/dev/null; then
        echo -e "${RED}Server failed to start${NC}"
        exit 1
    fi

    echo -e "${GREEN}Server started${NC}"
}

# Test helper function
run_test() {
    local test_name="$1"
    local expected_status="$2"
    local url="$3"
    local extra_opts="$4"

    local response
    response=$(curl -s -w "\n%{http_code}" $extra_opts "$url" 2>/dev/null)
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    if [ "$status_code" = "$expected_status" ]; then
        echo -e "  ${GREEN}✓${NC} $test_name (status: $status_code)"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "  ${RED}✗${NC} $test_name (expected: $expected_status, got: $status_code)"
        echo "    Response: $body"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# Test health endpoint
test_health() {
    echo
    echo "Testing health endpoint..."
    run_test "Health check (no auth)" "200" "http://localhost:$HTTP_PORT/api/v1/health"
}

# Test UI endpoints
test_ui() {
    echo
    echo "Testing UI endpoints..."

    # Test index.html
    run_test "UI index.html (root)" "200" "http://localhost:$HTTP_PORT/"

    # Test static assets
    run_test "UI styles.css" "200" "http://localhost:$HTTP_PORT/ui/styles.css"
    run_test "UI app.js" "200" "http://localhost:$HTTP_PORT/ui/app.js"

    # Test SPA fallback (non-existent route should return index.html)
    run_test "UI SPA fallback" "200" "http://localhost:$HTTP_PORT/repos"
}

# Test authentication with Basic Auth
test_auth() {
    echo
    echo "Testing authentication..."

    # Test 401 without auth
    run_test "401 without auth" "401" "http://localhost:$HTTP_PORT/api/v1/repos"

    # Test with Basic Auth - list repos
    run_test "List repos with Basic Auth" "200" "http://localhost:$HTTP_PORT/api/v1/repos" "-u $ADMIN_USERNAME:$ADMIN_PASSWORD"
}

# Test repos API
test_repos() {
    echo
    echo "Testing repos API..."

    local auth_opts="-u $ADMIN_USERNAME:$ADMIN_PASSWORD"

    # Test list repos (empty)
    run_test "List repos (empty)" "200" "http://localhost:$HTTP_PORT/api/v1/repos" "$auth_opts"

    # Create a test repo
    local create_resp
    create_resp=$(curl -s -w "\n%{http_code}" $auth_opts -X POST "http://localhost:$HTTP_PORT/api/v1/repos" \
        -H "Content-Type: application/json" \
        -d '{"name":"test-repo","description":"Test repository"}')
    local create_status=$(echo "$create_resp" | tail -n1)

    if [ "$create_status" = "201" ] || [ "$create_status" = "200" ]; then
        echo -e "  ${GREEN}✓${NC} Create repo (status: $create_status)"
        PASSED=$((PASSED + 1))
    else
        echo -e "  ${RED}✗${NC} Create repo (expected: 201, got: $create_status)"
        FAILED=$((FAILED + 1))
    fi

    # Verify repo was created by listing
    run_test "List repos (with repo)" "200" "http://localhost:$HTTP_PORT/api/v1/repos" "$auth_opts"

    # Get repo details
    run_test "Get repo details" "200" "http://localhost:$HTTP_PORT/api/v1/repos/test-repo" "$auth_opts"
}

# Test 404 for non-existent endpoints
test_404() {
    echo
    echo "Testing 404 handling..."

    # These need auth - expect 401, not 404 (auth blocks before 404)
    run_test "401 for non-existent endpoint (auth required)" "401" "http://localhost:$HTTP_PORT/api/v1/nonexistent"
    run_test "401 for non-existent repo (auth required)" "401" "http://localhost:$HTTP_PORT/api/v1/repos/nonexistent-repo"
}

# Main
main() {
    echo "============================================"
    echo "MyGit Automated API Test Suite"
    echo "============================================"

    check_prereqs
    build
    start_server

    test_health
    test_ui
    test_auth
    test_repos
    test_404

    echo
    echo "============================================"
    echo "Test Results: ${PASSED} passed, ${FAILED} failed"
    echo "============================================"

    if [ $FAILED -gt 0 ]; then
        echo -e "${RED}Some tests failed${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

main "$@"
