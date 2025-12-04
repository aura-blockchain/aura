#!/bin/bash
# Start Aura Block Explorer in Development Mode (without Docker)
# This script starts the Vite dev server for hot-reload development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

EXPLORER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/explorer/ping-pub-explorer"

echo -e "${GREEN}=== Aura Block Explorer (Development Mode) ===${NC}"
echo ""

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}Error: Node.js is not installed${NC}"
    echo "Please install Node.js (version 18 or higher) and try again"
    exit 1
fi

# Check if yarn is installed
if ! command -v yarn &> /dev/null; then
    echo -e "${RED}Error: Yarn is not installed${NC}"
    echo "Installing yarn..."
    npm install -g yarn
fi

# Check Node version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo -e "${YELLOW}Warning: Node.js version 18+ is recommended${NC}"
    echo "Current version: $(node -v)"
fi

# Check if Aura node is running
echo "Checking Aura node availability..."
if curl -s http://localhost:26657/status > /dev/null 2>&1; then
    echo -e "${GREEN} Aura RPC (26657) is accessible${NC}"
else
    echo -e "${YELLOW}Warning: Cannot reach Aura RPC on localhost:26657${NC}"
    echo "The explorer may not work properly until the Aura node is started"
fi

if curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info > /dev/null 2>&1; then
    echo -e "${GREEN} Aura API (1317) is accessible${NC}"
else
    echo -e "${YELLOW}Warning: Cannot reach Aura API on localhost:1317${NC}"
fi

echo ""
echo "Starting Explorer in Development Mode..."
cd "$EXPLORER_DIR"

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    yarn install --ignore-engines
fi

echo ""
echo -e "${GREEN}Starting Vite dev server...${NC}"
echo "This will open the explorer with hot-reload enabled"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Start dev server
yarn dev --host 0.0.0.0
