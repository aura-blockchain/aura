#!/bin/bash
# Start Aura Block Explorer (Ping.pub)
# This script starts the block explorer using Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

EXPLORER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/explorer"

echo -e "${GREEN}=== Aura Block Explorer Startup ===${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    echo "Please start Docker and try again"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null 2>&1; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    echo "Please install Docker Compose and try again"
    exit 1
fi

# Determine which docker compose command to use
if docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

# Check if Aura node is running
echo "Checking Aura node availability..."
if curl -s http://localhost:26657/status > /dev/null 2>&1; then
    echo -e "${GREEN} Aura RPC (26657) is accessible${NC}"
else
    echo -e "${YELLOW}Warning: Cannot reach Aura RPC on localhost:26657${NC}"
    echo "The explorer may not work properly until the Aura node is started"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborting..."
        exit 1
    fi
fi

if curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info > /dev/null 2>&1; then
    echo -e "${GREEN} Aura API (1317) is accessible${NC}"
else
    echo -e "${YELLOW}Warning: Cannot reach Aura API on localhost:1317${NC}"
fi

echo ""
echo "Starting Block Explorer..."
echo "Explorer directory: $EXPLORER_DIR"
cd "$EXPLORER_DIR"

# Build and start the explorer
$DOCKER_COMPOSE up -d --build

echo ""
echo -e "${GREEN}=== Block Explorer Started ===${NC}"
echo ""
echo "Access the explorer at: http://localhost:8088"
echo ""
echo "To view logs: $DOCKER_COMPOSE -f $EXPLORER_DIR/docker-compose.yml logs -f"
echo "To stop: $DOCKER_COMPOSE -f $EXPLORER_DIR/docker-compose.yml down"
echo ""
echo -e "${YELLOW}Note: Make sure your Aura node has the API enabled in app.toml${NC}"
echo ""
