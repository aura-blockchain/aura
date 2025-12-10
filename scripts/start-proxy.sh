#!/bin/bash
# Aura Testnet Public Endpoint Proxy - Start Script
# This script starts the nginx reverse proxy for public testnet endpoints

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored messages
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Banner
echo -e "${BLUE}"
echo "========================================"
echo "  Aura Testnet Public Endpoint Proxy"
echo "========================================"
echo -e "${NC}"

# Check if running from project root
cd "$PROJECT_ROOT" || exit 1

# Step 1: Check if testnet is running
info "Checking if testnet is running..."
if ! docker network ls | grep -q "aura_aura-testnet"; then
    error "Testnet network not found!"
    error "Please start the testnet first:"
    echo ""
    echo "  docker-compose -f docker-compose.testnet.yml up -d"
    echo ""
    exit 1
fi
success "Testnet network found"

# Step 2: Check if validators are healthy
info "Checking validator health..."
VALIDATOR_COUNT=$(docker ps --filter "name=aura-validator-" --filter "health=healthy" --format "{{.Names}}" | wc -l)
if [ "$VALIDATOR_COUNT" -lt 1 ]; then
    warning "No healthy validators found!"
    warning "Validators may still be starting up..."
    echo ""
    echo "Current validator status:"
    docker-compose -f docker-compose.testnet.yml ps
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Aborted by user"
        exit 1
    fi
else
    success "Found $VALIDATOR_COUNT healthy validator(s)"
fi

# Step 3: Check if proxy is already running
info "Checking if proxy is already running..."
if docker ps --filter "name=aura-testnet-proxy" --format "{{.Names}}" | grep -q "aura-testnet-proxy"; then
    warning "Proxy is already running!"
    read -p "Restart the proxy? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        info "Stopping existing proxy..."
        docker-compose -f docker-compose.proxy.yml down
        success "Proxy stopped"
    else
        info "Keeping existing proxy running"
        echo ""
        echo "Proxy endpoints:"
        echo "  - RPC:     http://localhost/rpc"
        echo "  - API:     http://localhost/api"
        echo "  - gRPC:    localhost:9090"
        echo "  - Swagger: http://localhost/api/swagger/"
        echo "  - Health:  http://localhost/health"
        exit 0
    fi
fi

# Step 4: Start the proxy
info "Starting the proxy..."
docker-compose -f docker-compose.proxy.yml up -d

# Step 5: Wait for proxy to be healthy
info "Waiting for proxy to become healthy..."
sleep 5

MAX_RETRIES=30
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s -f http://localhost/health > /dev/null 2>&1; then
        success "Proxy is healthy!"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo -n "."
    sleep 1
done
echo ""

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    error "Proxy failed to become healthy after ${MAX_RETRIES} seconds"
    echo ""
    echo "Check logs with:"
    echo "  docker-compose -f docker-compose.proxy.yml logs"
    exit 1
fi

# Step 6: Test endpoints
echo ""
info "Testing endpoints..."

# Test health endpoint
if curl -s -f http://localhost/health > /dev/null 2>&1; then
    success "✓ Health endpoint OK"
else
    error "✗ Health endpoint failed"
fi

# Test RPC endpoint
if curl -s -f http://localhost/rpc/status > /dev/null 2>&1; then
    success "✓ RPC endpoint OK"
else
    warning "✗ RPC endpoint failed (validator may still be starting)"
fi

# Test API endpoint
if curl -s -f http://localhost/api/cosmos/base/tendermint/v1beta1/node_info > /dev/null 2>&1; then
    success "✓ API endpoint OK"
else
    warning "✗ API endpoint failed (validator may still be starting)"
fi

# Step 7: Display connection info
echo ""
success "Proxy started successfully!"
echo ""
echo -e "${GREEN}Available endpoints:${NC}"
echo "  RPC:     http://localhost/rpc"
echo "  API:     http://localhost/api"
echo "  gRPC:    localhost:9090"
echo "  Swagger: http://localhost/api/swagger/"
echo "  Health:  http://localhost/health"
echo ""
echo -e "${BLUE}Example commands:${NC}"
echo "  # Get node status"
echo "  curl http://localhost/rpc/status | jq"
echo ""
echo "  # Get latest block"
echo "  curl http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest | jq"
echo ""
echo "  # View proxy logs"
echo "  docker-compose -f docker-compose.proxy.yml logs -f"
echo ""
echo "  # Stop proxy"
echo "  docker-compose -f docker-compose.proxy.yml down"
echo ""
echo -e "${BLUE}Documentation:${NC}"
echo "  docs/TESTNET_PUBLIC_ENDPOINTS.md"
echo ""
