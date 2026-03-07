#!/bin/bash
# Local testing script for Gitea addon
# Usage: ./test-gitea-local.sh [build|run|clean|logs]

set -e

REPO_DIR="/var/home/flozsc/Code/my-homeassistant-apps"
TEST_DIR="/tmp/gitea-test"
IMAGE_NAME="local-gitea-test"
CONTAINER_NAME="test-gitea"

# Create test directory structure
setup_test_dir() {
    mkdir -p "${TEST_DIR}/data"
    mkdir -p "${TEST_DIR}/git"
    mkdir -p "${TEST_DIR}/data/gitea"
    mkdir -p "${TEST_DIR}/data/git/repositories"
    mkdir -p "${TEST_DIR}/data/git/custom"
    chmod -R 775 "${TEST_DIR}/data"
    echo "Test directory setup at ${TEST_DIR}"
}

# Build the test image
build_image() {
    echo "Building test image..."
    # Use the actual Home Assistant base image for proper testing
    podman build -t "${IMAGE_NAME}" \
        --build-arg BUILD_FROM=ghcr.io/home-assistant/amd64-base:latest \
        --build-arg BUILD_ARCH=amd64 \
        -f "${REPO_DIR}/gitea/Dockerfile" "${REPO_DIR}/gitea/"
    echo "Image built: ${IMAGE_NAME}"
}

# Run the test container
run_container() {
    echo "Starting test container..."
    podman run -d --name "${CONTAINER_NAME}" \
        -v "${TEST_DIR}/data:/data" \
        -v "${TEST_DIR}/git:/data/git" \
        -p 3000:3000 \
        -p 2222:2222 \
        --user 1000:1000 \
        --security-opt label=disable \
        "${IMAGE_NAME}"
    
    echo "Container started: ${CONTAINER_NAME}"
    echo "Waiting for initialization..."
    sleep 5
}

# Show container logs
show_logs() {
    echo "=== Container Logs ==="
    podman logs "${CONTAINER_NAME}"
}

# Test container functionality
test_functionality() {
    echo "=== Testing HTTP Service ==="
    curl -I http://localhost:3000 2>/dev/null || echo "HTTP service not responding"
    
    echo "=== Testing SSH Service ==="
    timeout 2 bash -c "echo > /dev/tcp/localhost/2222" 2>/dev/null && echo "SSH port accessible" || echo "SSH port not accessible"
    
    echo "=== Testing Process Status ==="
    podman exec "${CONTAINER_NAME}" ps aux || echo "Could not check processes"
}

# Clean up test environment
cleanup() {
    echo "Cleaning up..."
    podman stop "${CONTAINER_NAME}" 2>/dev/null || true
    podman rm "${CONTAINER_NAME}" 2>/dev/null || true
    podman rmi "${IMAGE_NAME}" 2>/dev/null || true
    rm -rf "${TEST_DIR}"
    echo "Cleanup complete"
}

# Main script logic
case "$1" in
    build)
        build_image
        ;;
    run)
        setup_test_dir
        run_container
        ;;
    test)
        test_functionality
        ;;
    logs)
        show_logs
        ;;
    clean)
        cleanup
        ;;
    full)
        cleanup
        setup_test_dir
        build_image
        run_container
        sleep 3
        test_functionality
        show_logs
        ;;
    *)
        echo "Usage: $0 {build|run|test|logs|clean|full}"
        exit 1
        ;;
esac