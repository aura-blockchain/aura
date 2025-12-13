#!/bin/bash
# Phase 7.2: Resource Constraint Test
# Tests node performance and stability under heavy resource constraints

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RESULTS_FILE="$SCRIPT_DIR/test_7.2_results.txt"
CONTAINER_NAME="aura-resource-test"

echo "==================================================================="
echo "Phase 7.2: Resource Constraint Test"
echo "==================================================================="
echo ""

# Initialize results file
cat > "$RESULTS_FILE" << EOF
=================================================================
Phase 7.2: Resource Constraint Test Results
=================================================================
Timestamp: $(date)
Container: $CONTAINER_NAME

EOF

# Function to log results
log_result() {
    echo "$1" | tee -a "$RESULTS_FILE"
}

# Function to log section header
log_section() {
    echo ""
    echo "-----------------------------------------------------------------"
    echo "$1"
    echo "-----------------------------------------------------------------"
    echo "" | tee -a "$RESULTS_FILE"
    echo "$1" >> "$RESULTS_FILE"
    echo "-----------------------------------------------------------------" >> "$RESULTS_FILE"
}

# Function to cleanup
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"

    # Stop and remove container
    docker stop "$CONTAINER_NAME" 2>/dev/null || true
    docker rm "$CONTAINER_NAME" 2>/dev/null || true

    echo "Cleanup complete"
}

trap cleanup EXIT

# Check Docker is running
log_section "Checking prerequisites"
if ! docker info &> /dev/null; then
    log_result "❌ Docker is not running"
    exit 1
fi
log_result "✅ Docker is running"

# Build the binary first
log_section "Building aurad binary"
cd "$PROJECT_ROOT/chain"

if go build -o "$PROJECT_ROOT/chain/aurad" ./cmd/aurad 2>&1 | tee -a "$RESULTS_FILE"; then
    log_result "✅ Binary built successfully"
else
    log_result "❌ Binary build failed"
    exit 1
fi

# Build the Docker image
log_section "Building Docker image"
cd "$PROJECT_ROOT/chain"

# Create a simple Dockerfile for testing
cat > "$PROJECT_ROOT/chain/Dockerfile.resource-test" << 'DOCKEREOF'
FROM alpine:latest

RUN apk add --no-cache ca-certificates jq bash curl bc

COPY aurad /usr/local/bin/

EXPOSE 26656 26657 9090 1317

CMD ["aurad"]
DOCKEREOF

log_result "Building Docker image..."
if docker build -f Dockerfile.resource-test -t aura-resource-test:latest . &>> "$RESULTS_FILE"; then
    log_result "✅ Docker image built successfully"
else
    log_result "❌ Docker image build failed"
    cat "$RESULTS_FILE" | tail -20
    exit 1
fi

# Test 1: Baseline (no resource constraints)
log_section "Test 1: Baseline Performance (No Constraints)"

log_result "Starting container with no resource limits..."
docker run -d \
    --name "${CONTAINER_NAME}-baseline" \
    aura-resource-test:latest \
    sh -c "aurad init test-node --chain-id resource-test && aurad start" \
    &>> "$RESULTS_FILE"

sleep 10

# Check if container is running
if docker ps | grep -q "${CONTAINER_NAME}-baseline"; then
    log_result "✅ Baseline container is running"

    # Check resource usage
    BASELINE_STATS=$(docker stats --no-stream --format "{{.MemUsage}} | {{.CPUPerc}}" "${CONTAINER_NAME}-baseline")
    log_result "Baseline resource usage: $BASELINE_STATS"

    # Check logs for block production
    BASELINE_LOGS=$(docker logs "${CONTAINER_NAME}-baseline" 2>&1 | grep "committed state" | tail -5)
    if [ -n "$BASELINE_LOGS" ]; then
        BASELINE_HEIGHT=$(echo "$BASELINE_LOGS" | tail -1 | grep -oP 'height=\K[0-9]+' || echo "0")
        log_result "✅ Baseline producing blocks (height: $BASELINE_HEIGHT)"
    else
        log_result "⚠️  No blocks detected in baseline"
    fi
else
    log_result "❌ Baseline container failed to start"
fi

# Stop baseline
docker stop "${CONTAINER_NAME}-baseline" &>> "$RESULTS_FILE"
docker rm "${CONTAINER_NAME}-baseline" &>> "$RESULTS_FILE"
sleep 2

# Test 2: Heavy resource constraints (512MB RAM, 0.5 CPU)
log_section "Test 2: Constrained Resources (512MB RAM, 0.5 CPU)"

log_result "Starting container with resource limits..."
docker run -d \
    --name "$CONTAINER_NAME" \
    --memory="512m" \
    --memory-swap="512m" \
    --cpus="0.5" \
    aura-resource-test:latest \
    sh -c "aurad init constrained-node --chain-id resource-test && aurad start" \
    &>> "$RESULTS_FILE"

sleep 15

# Monitor for 60 seconds
log_result "Monitoring constrained node for 60 seconds..."
for i in {1..12}; do
    sleep 5

    if ! docker ps | grep -q "$CONTAINER_NAME"; then
        log_result "❌ Container stopped unexpectedly at $(($i * 5))s"
        docker logs "$CONTAINER_NAME" 2>&1 | tail -50 >> "$RESULTS_FILE"
        break
    fi

    # Check resource usage
    STATS=$(docker stats --no-stream --format "{{.MemUsage}} | {{.CPUPerc}}" "$CONTAINER_NAME")
    log_result "[$((i * 5))s] Resource usage: $STATS"

    # Check for OOM killer
    if docker inspect "$CONTAINER_NAME" 2>&1 | grep -q "OOMKilled.*true"; then
        log_result "❌ Container killed by OOM killer"
        break
    fi
done

# Final check
if docker ps | grep -q "$CONTAINER_NAME"; then
    log_result "✅ Container survived 60s under resource constraints"

    # Check block production
    CONSTRAINED_LOGS=$(docker logs "$CONTAINER_NAME" 2>&1 | grep "committed state" | tail -5)
    if [ -n "$CONSTRAINED_LOGS" ]; then
        CONSTRAINED_HEIGHT=$(echo "$CONSTRAINED_LOGS" | tail -1 | grep -oP 'height=\K[0-9]+' || echo "0")
        log_result "✅ Constrained node producing blocks (height: $CONSTRAINED_HEIGHT)"

        if [ "$CONSTRAINED_HEIGHT" -gt 0 ]; then
            # Calculate block production rate
            BLOCK_RATE=$(echo "scale=2; $CONSTRAINED_HEIGHT / 60" | bc)
            log_result "Block production rate: $BLOCK_RATE blocks/second"
        fi
    else
        log_result "⚠️  No blocks detected under constraints"
        log_result "Checking logs for errors..."
        docker logs "$CONTAINER_NAME" 2>&1 | grep -i "error\|panic\|fatal" | tail -10 >> "$RESULTS_FILE"
    fi

    # Get final stats
    FINAL_STATS=$(docker stats --no-stream --format "MEM: {{.MemUsage}} | CPU: {{.CPUPerc}}" "$CONTAINER_NAME")
    log_result "Final resource usage: $FINAL_STATS"
else
    log_result "❌ Container did not survive resource constraints"
    log_result "Last logs:"
    docker logs "$CONTAINER_NAME" 2>&1 | tail -30 >> "$RESULTS_FILE"
fi

# Test 3: Extreme constraints (256MB RAM, 0.25 CPU)
log_section "Test 3: Extreme Constraints (256MB RAM, 0.25 CPU)"

docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME" 2>/dev/null || true
sleep 2

log_result "Starting container with extreme resource limits..."
docker run -d \
    --name "${CONTAINER_NAME}-extreme" \
    --memory="256m" \
    --memory-swap="256m" \
    --cpus="0.25" \
    aura-resource-test:latest \
    sh -c "aurad init extreme-node --chain-id resource-test && aurad start" \
    &>> "$RESULTS_FILE"

sleep 15

# Monitor for 30 seconds
log_result "Monitoring extreme constraints for 30 seconds..."
for i in {1..6}; do
    sleep 5

    if ! docker ps | grep -q "${CONTAINER_NAME}-extreme"; then
        log_result "❌ Container stopped at $(($i * 5))s under extreme constraints"
        docker logs "${CONTAINER_NAME}-extreme" 2>&1 | tail -50 >> "$RESULTS_FILE"
        break
    fi

    STATS=$(docker stats --no-stream --format "{{.MemUsage}} | {{.CPUPerc}}" "${CONTAINER_NAME}-extreme")
    log_result "[$((i * 5))s] Resource usage: $STATS"

    # Check for OOM
    if docker inspect "${CONTAINER_NAME}-extreme" 2>&1 | grep -q "OOMKilled.*true"; then
        log_result "❌ Container killed by OOM killer (expected with 256MB)"
        break
    fi
done

# Final check
if docker ps | grep -q "${CONTAINER_NAME}-extreme"; then
    log_result "✅ Container survived 30s under extreme constraints"

    EXTREME_LOGS=$(docker logs "${CONTAINER_NAME}-extreme" 2>&1 | grep "committed state" | tail -5)
    if [ -n "$EXTREME_LOGS" ]; then
        EXTREME_HEIGHT=$(echo "$EXTREME_LOGS" | tail -1 | grep -oP 'height=\K[0-9]+' || echo "0")
        log_result "✅ Extreme node producing blocks (height: $EXTREME_HEIGHT)"
    else
        log_result "⚠️  No blocks under extreme constraints (may be too limited)"
    fi
else
    log_result "⚠️  Container could not run with 256MB RAM / 0.25 CPU"
    log_result "This defines the minimum system requirements"
fi

# Cleanup extreme container
docker stop "${CONTAINER_NAME}-extreme" 2>/dev/null || true
docker rm "${CONTAINER_NAME}-extreme" 2>/dev/null || true

# Summary
log_section "Test 7.2 Summary"

log_result ""
log_result "Minimum System Requirements Assessment:"
log_result "- Recommended: ≥ 512MB RAM, ≥ 0.5 CPU cores"
log_result "- Minimum (if blocks produced under extreme test): 256MB RAM, 0.25 CPU cores"
log_result "- Optimal: ≥ 2GB RAM, ≥ 2 CPU cores for production use"
log_result ""

echo ""
echo "=================================================================="
echo "Test 7.2 Complete - Results saved to:"
echo "$RESULTS_FILE"
echo "=================================================================="
echo ""

# Display summary
cat "$RESULTS_FILE" | grep -E "^(✅|❌|⚠️)" | tee -a "$RESULTS_FILE"

echo ""
log_result "Test execution completed at: $(date)"
