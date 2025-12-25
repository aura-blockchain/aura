#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=================================================="
echo "  Aura 4-Node Testnet Launcher"
echo "=================================================="
echo ""

# Check Docker is running
echo "Step 1/6: Checking Docker status..."
if ! docker ps >/dev/null 2>&1; then
    echo "❌ ERROR: Docker daemon is not running"
    echo ""
    echo "Please start Docker first:"
    echo "  sudo systemctl start docker"
    echo ""
    exit 1
fi
echo "✓ Docker is running"
echo ""

# Populate volumes
echo "Step 2/6: Populating testnet volumes..."
cd testnet-data
./populate-volumes.sh
cd ..
echo "✓ Volumes populated"
echo ""

# Start testnet
echo "Step 3/6: Starting 4-node testnet..."
./scripts/testnet-manage.sh start
echo "✓ Testnet started"
echo ""

# Wait for initialization
echo "Step 4/6: Waiting for node initialization (90 seconds)..."
for i in {90..1}; do
    printf "\rTime remaining: %2d seconds" $i
    sleep 1
done
echo ""
echo "✓ Initialization period complete"
echo ""

# Verify containers
echo "Step 5/6: Verifying container status..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(NAMES|aura-node)"
echo ""

# Health check
echo "Step 6/6: Health check on all nodes..."
for port in 27657 27757 27857 27957; do
    status=$(curl -s "http://localhost:$port/health" 2>&1 || echo "ERROR")
    if [[ "$status" == *"result"* ]] || [[ "$status" == "{}" ]]; then
        echo "✓ Port $port: Healthy"
    else
        echo "⚠ Port $port: $status"
    fi
done
echo ""

# Quick consensus check
echo "Consensus Status:"
./scripts/testnet-monitor.sh quick
echo ""

echo "=================================================="
echo "  Testnet Launch Complete!"
echo "=================================================="
echo ""
echo "RPC Endpoints:"
echo "  Node 1: http://localhost:27657"
echo "  Node 2: http://localhost:27757"
echo "  Node 3: http://localhost:27857"
echo "  Node 4: http://localhost:27957"
echo ""
echo "Management Commands:"
echo "  Monitor: ./scripts/testnet-monitor.sh"
echo "  Stop:    ./scripts/testnet-manage.sh stop"
echo "  Restart: ./scripts/testnet-manage.sh restart"
echo "  Logs:    docker logs -f aura-node1"
echo ""
