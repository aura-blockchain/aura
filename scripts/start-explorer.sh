#!/bin/bash
# ============================================================================
# Aura Block Explorer - Quick Start Script
# ============================================================================
# Deploys Ping.pub block explorer for Aura testnet
#
# Usage:
#   ./start-explorer.sh [options]
#
# Options:
#   --build    Force rebuild of the image
#   --logs     Follow logs after starting
#   --help     Show this help message
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Function to print colored messages
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_header() {
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

# Parse arguments
BUILD_FLAG=""
FOLLOW_LOGS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --build)
            BUILD_FLAG="--build"
            shift
            ;;
        --logs)
            FOLLOW_LOGS=true
            shift
            ;;
        --help)
            sed -n '/^# =/,/^# =/p' "$0" | grep -v '^# ='
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# ============================================================================
# Pre-flight Checks
# ============================================================================

print_header "Aura Block Explorer - Deployment"

print_info "Running pre-flight checks..."

# Check if testnet is running
if ! docker ps | grep -q "aura-validator"; then
    print_error "Aura testnet is not running!"
    echo ""
    echo "Start the testnet first with:"
    echo "  docker compose -f docker-compose.testnet.yml up -d"
    echo ""
    exit 1
fi

print_success "Testnet is running"

# Check if network exists
if ! docker network ls | grep -q "aura_aura-testnet"; then
    print_error "Testnet network not found!"
    echo ""
    echo "The testnet docker network 'aura_aura-testnet' does not exist."
    echo "Please ensure the testnet is running properly."
    echo ""
    exit 1
fi

print_success "Testnet network exists"

# Check validator connectivity
print_info "Checking validator connectivity..."

if docker exec aura-validator-1 curl -s --max-time 5 http://localhost:26657/status > /dev/null 2>&1; then
    print_success "Validator-1 RPC is accessible"
else
    print_warning "Validator-1 RPC check failed (may still work)"
fi

# ============================================================================
# Deploy Explorer
# ============================================================================

print_header "Deploying Block Explorer"

print_info "Configuration:"
echo "  - Chain ID: aura-local-4"
echo "  - Primary RPC: http://172.26.0.10:26657 (validator-1)"
echo "  - Primary API: http://172.26.0.10:1317 (validator-1)"
echo "  - Explorer Port: 8088"
echo ""

# Check if explorer is already running
if docker ps | grep -q "aura-block-explorer"; then
    print_info "Explorer is already running. Stopping..."
    docker compose -f docker-compose.explorer.yml down
fi

# Start the explorer
print_info "Starting explorer (this may take 3-5 minutes on first build)..."

if [ -n "$BUILD_FLAG" ]; then
    print_info "Force rebuilding image..."
fi

docker compose -f docker-compose.explorer.yml up -d $BUILD_FLAG

if [ $? -ne 0 ]; then
    print_error "Failed to start explorer!"
    echo ""
    echo "Check logs with:"
    echo "  docker compose -f docker-compose.explorer.yml logs"
    echo ""
    exit 1
fi

print_success "Explorer container started"

# ============================================================================
# Wait for Health Check
# ============================================================================

print_header "Waiting for Explorer to be Ready"

print_info "Waiting for container to be healthy (max 60 seconds)..."

TIMEOUT=60
ELAPSED=0
HEALTHY=false

while [ $ELAPSED -lt $TIMEOUT ]; do
    if docker ps --filter "name=aura-block-explorer" --format "{{.Status}}" | grep -q "healthy"; then
        HEALTHY=true
        break
    fi

    if docker ps --filter "name=aura-block-explorer" --format "{{.Status}}" | grep -q "unhealthy"; then
        print_warning "Container is unhealthy. Checking logs..."
        docker logs aura-block-explorer --tail 20
        break
    fi

    echo -n "."
    sleep 2
    ELAPSED=$((ELAPSED + 2))
done

echo ""

if [ "$HEALTHY" = true ]; then
    print_success "Explorer is healthy and ready!"
else
    print_warning "Health check timeout or failed (explorer may still work)"
fi

# ============================================================================
# Verification
# ============================================================================

print_header "Deployment Complete"

print_info "Testing HTTP endpoint..."
if curl -s --max-time 5 http://localhost:8088/ > /dev/null 2>&1; then
    print_success "Explorer is responding on http://localhost:8088"
else
    print_warning "Explorer HTTP check failed (may need more time to start)"
fi

# Show container status
echo ""
print_info "Container Status:"
docker ps --filter "name=aura-block-explorer" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

# Success message
echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    DEPLOYMENT SUCCESSFUL                       ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "🌐 Access the explorer at: ${BLUE}http://localhost:8088${NC}"
echo ""
echo "📖 Documentation: EXPLORER_DEPLOYMENT_GUIDE.md"
echo ""
echo "📊 Useful Commands:"
echo "  - View logs:         docker logs aura-block-explorer -f"
echo "  - Restart:           docker compose -f docker-compose.explorer.yml restart"
echo "  - Stop:              docker compose -f docker-compose.explorer.yml down"
echo "  - Check status:      docker ps | grep explorer"
echo ""

# Follow logs if requested
if [ "$FOLLOW_LOGS" = true ]; then
    print_info "Following logs (Ctrl+C to exit)..."
    docker logs aura-block-explorer -f
fi
