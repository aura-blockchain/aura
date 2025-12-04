#!/bin/bash
# Stop Aura Block Explorer
# This script stops the block explorer Docker containers

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

EXPLORER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/explorer"

echo -e "${YELLOW}=== Stopping Aura Block Explorer ===${NC}"
echo ""

# Determine which docker compose command to use
if docker compose version &> /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

cd "$EXPLORER_DIR"

# Stop and remove containers
$DOCKER_COMPOSE down

echo ""
echo -e "${GREEN}Block Explorer stopped successfully${NC}"
echo ""
