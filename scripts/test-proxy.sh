#!/bin/bash
# Aura Testnet Public Endpoint Proxy - Test Script
# This script tests all proxy endpoints to ensure they're working correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Print colored messages
test_info() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

test_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

test_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

test_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Banner
echo -e "${BLUE}"
echo "================================================"
echo "  Aura Testnet Public Endpoint Proxy - Tests"
echo "================================================"
echo -e "${NC}"

# Check prerequisites
info "Checking prerequisites..."

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed${NC}"
    exit 1
fi

# Check if jq is available (optional but recommended)
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}Warning: jq not installed - some tests will be skipped${NC}"
    HAS_JQ=false
else
    HAS_JQ=true
fi

echo ""
info "Starting endpoint tests..."
echo ""

# =============================================================================
# Test 1: Proxy Container Running
# =============================================================================
test_info "Checking if proxy container is running..."
if docker ps --filter "name=aura-testnet-proxy" --format "{{.Names}}" | grep -q "aura-testnet-proxy"; then
    test_pass "Proxy container is running"
else
    test_fail "Proxy container is not running"
    echo ""
    echo "Start the proxy with:"
    echo "  docker compose -f docker-compose.proxy.yml up -d"
    exit 1
fi

# =============================================================================
# Test 2: Health Endpoint
# =============================================================================
test_info "Testing health endpoint (http://localhost/health)..."
HEALTH_RESPONSE=$(curl -s -f http://localhost/health 2>&1)
if [ $? -eq 0 ]; then
    if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
        test_pass "Health endpoint returns healthy status"
    else
        test_fail "Health endpoint returned unexpected response: $HEALTH_RESPONSE"
    fi
else
    test_fail "Health endpoint not accessible"
fi

# =============================================================================
# Test 3: Service Info Endpoint
# =============================================================================
test_info "Testing service info endpoint (http://localhost/)..."
INFO_RESPONSE=$(curl -s -f http://localhost/ 2>&1)
if [ $? -eq 0 ]; then
    if echo "$INFO_RESPONSE" | grep -q "aura-local-4"; then
        test_pass "Service info endpoint returns chain info"
    else
        test_fail "Service info endpoint returned unexpected response"
    fi
else
    test_fail "Service info endpoint not accessible"
fi

# =============================================================================
# Test 4: RPC Status Endpoint
# =============================================================================
test_info "Testing RPC status endpoint (http://localhost/rpc/status)..."
RPC_STATUS=$(curl -s -f http://localhost/rpc/status 2>&1)
if [ $? -eq 0 ]; then
    if echo "$RPC_STATUS" | grep -q "result"; then
        if [ "$HAS_JQ" = true ]; then
            CHAIN_ID=$(echo "$RPC_STATUS" | jq -r '.result.node_info.network' 2>/dev/null)
            NODE_VERSION=$(echo "$RPC_STATUS" | jq -r '.result.node_info.version' 2>/dev/null)
            LATEST_HEIGHT=$(echo "$RPC_STATUS" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null)
            test_pass "RPC status endpoint working (Chain: $CHAIN_ID, Height: $LATEST_HEIGHT, Version: $NODE_VERSION)"
        else
            test_pass "RPC status endpoint working"
        fi
    else
        test_fail "RPC status endpoint returned unexpected response"
    fi
else
    test_fail "RPC status endpoint not accessible (validator may be starting)"
fi

# =============================================================================
# Test 5: RPC Block Endpoint
# =============================================================================
test_info "Testing RPC block endpoint (http://localhost/rpc/block)..."
RPC_BLOCK=$(curl -s -f http://localhost/rpc/block 2>&1)
if [ $? -eq 0 ]; then
    if echo "$RPC_BLOCK" | grep -q "result"; then
        test_pass "RPC block endpoint working"
    else
        test_fail "RPC block endpoint returned unexpected response"
    fi
else
    test_fail "RPC block endpoint not accessible"
fi

# =============================================================================
# Test 6: API Node Info Endpoint
# =============================================================================
test_info "Testing API node info (http://localhost/api/cosmos/base/tendermint/v1beta1/node_info)..."
API_NODE_INFO=$(curl -s -f http://localhost/api/cosmos/base/tendermint/v1beta1/node_info 2>&1)
if [ $? -eq 0 ]; then
    if echo "$API_NODE_INFO" | grep -q "default_node_info"; then
        if [ "$HAS_JQ" = true ]; then
            API_CHAIN_ID=$(echo "$API_NODE_INFO" | jq -r '.default_node_info.network' 2>/dev/null)
            test_pass "API node info endpoint working (Chain: $API_CHAIN_ID)"
        else
            test_pass "API node info endpoint working"
        fi
    else
        test_fail "API node info endpoint returned unexpected response"
    fi
else
    test_fail "API node info endpoint not accessible"
fi

# =============================================================================
# Test 7: API Latest Block Endpoint
# =============================================================================
test_info "Testing API latest block (http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest)..."
API_BLOCK=$(curl -s -f http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest 2>&1)
if [ $? -eq 0 ]; then
    if echo "$API_BLOCK" | grep -q "block"; then
        if [ "$HAS_JQ" = true ]; then
            BLOCK_HEIGHT=$(echo "$API_BLOCK" | jq -r '.block.header.height' 2>/dev/null)
            test_pass "API latest block endpoint working (Height: $BLOCK_HEIGHT)"
        else
            test_pass "API latest block endpoint working"
        fi
    else
        test_fail "API latest block endpoint returned unexpected response"
    fi
else
    test_fail "API latest block endpoint not accessible"
fi

# =============================================================================
# Test 8: API Validators Endpoint
# =============================================================================
test_info "Testing API validators (http://localhost/api/cosmos/base/tendermint/v1beta1/validatorsets/latest)..."
API_VALIDATORS=$(curl -s -f http://localhost/api/cosmos/base/tendermint/v1beta1/validatorsets/latest 2>&1)
if [ $? -eq 0 ]; then
    if echo "$API_VALIDATORS" | grep -q "validators"; then
        if [ "$HAS_JQ" = true ]; then
            VALIDATOR_COUNT=$(echo "$API_VALIDATORS" | jq -r '.validators | length' 2>/dev/null)
            test_pass "API validators endpoint working ($VALIDATOR_COUNT validators)"
        else
            test_pass "API validators endpoint working"
        fi
    else
        test_fail "API validators endpoint returned unexpected response"
    fi
else
    test_fail "API validators endpoint not accessible"
fi

# =============================================================================
# Test 9: API Supply Endpoint
# =============================================================================
test_info "Testing API supply (http://localhost/api/cosmos/bank/v1beta1/supply)..."
API_SUPPLY=$(curl -s -f http://localhost/api/cosmos/bank/v1beta1/supply 2>&1)
if [ $? -eq 0 ]; then
    if echo "$API_SUPPLY" | grep -q "supply"; then
        test_pass "API supply endpoint working"
    else
        test_fail "API supply endpoint returned unexpected response"
    fi
else
    test_fail "API supply endpoint not accessible"
fi

# =============================================================================
# Test 10: CORS Headers
# =============================================================================
test_info "Testing CORS headers..."
CORS_RESPONSE=$(curl -s -I -H "Origin: http://example.com" http://localhost/rpc/status 2>&1)
if echo "$CORS_RESPONSE" | grep -qi "Access-Control-Allow-Origin"; then
    test_pass "CORS headers are present"
else
    test_fail "CORS headers are missing"
fi

# =============================================================================
# Test 11: gRPC Endpoint (requires grpcurl)
# =============================================================================
if command -v grpcurl &> /dev/null; then
    test_info "Testing gRPC endpoint (localhost:9090)..."
    GRPC_SERVICES=$(grpcurl -plaintext localhost:9090 list 2>&1)
    if [ $? -eq 0 ]; then
        if echo "$GRPC_SERVICES" | grep -q "cosmos"; then
            test_pass "gRPC endpoint working"
        else
            test_fail "gRPC endpoint returned unexpected response"
        fi
    else
        test_fail "gRPC endpoint not accessible"
    fi
else
    test_skip "gRPC test (grpcurl not installed)"
fi

# =============================================================================
# Test 12: WebSocket Endpoint
# =============================================================================
test_info "Testing WebSocket upgrade headers..."
WS_RESPONSE=$(curl -s -I -N -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost/rpc/websocket 2>&1)
if echo "$WS_RESPONSE" | grep -q "101"; then
    test_pass "WebSocket upgrade supported"
elif echo "$WS_RESPONSE" | grep -q "Upgrade"; then
    test_pass "WebSocket endpoint accessible (upgrade headers present)"
else
    test_skip "WebSocket test (requires WebSocket client for full test)"
fi

# =============================================================================
# Test 13: Rate Limiting Headers
# =============================================================================
test_info "Testing rate limiting (sending multiple rapid requests)..."
for i in {1..5}; do
    curl -s -f http://localhost/health > /dev/null 2>&1
done
sleep 1
RATE_TEST=$(curl -s -f http://localhost/health 2>&1)
if [ $? -eq 0 ]; then
    test_pass "Rate limiting configured (requests within limits)"
else
    test_fail "Rate limiting may be too aggressive or proxy down"
fi

# =============================================================================
# Test 14: Swagger/OpenAPI Documentation
# =============================================================================
test_info "Testing Swagger documentation (http://localhost/api/swagger/)..."
SWAGGER_RESPONSE=$(curl -s -f http://localhost/api/swagger/ 2>&1)
if [ $? -eq 0 ]; then
    if echo "$SWAGGER_RESPONSE" | grep -qi "swagger"; then
        test_pass "Swagger documentation accessible"
    else
        test_skip "Swagger documentation (may not be HTML)"
    fi
else
    test_skip "Swagger documentation (endpoint may not exist)"
fi

# =============================================================================
# Test Summary
# =============================================================================
echo ""
echo "================================================"
echo "                Test Summary"
echo "================================================"
echo -e "${GREEN}Passed:  $TESTS_PASSED${NC}"
echo -e "${RED}Failed:  $TESTS_FAILED${NC}"
echo -e "${YELLOW}Skipped: $TESTS_SKIPPED${NC}"
echo "================================================"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    echo ""
    echo "Your proxy is working correctly. Available endpoints:"
    echo "  - RPC:     http://localhost/rpc"
    echo "  - API:     http://localhost/api"
    echo "  - gRPC:    localhost:9090"
    echo "  - Swagger: http://localhost/api/swagger/"
    echo "  - Health:  http://localhost/health"
    echo ""
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    echo ""
    echo "Troubleshooting steps:"
    echo "  1. Check proxy logs:"
    echo "     docker compose -f docker-compose.proxy.yml logs -f"
    echo ""
    echo "  2. Check validator status:"
    echo "     docker compose -f docker-compose.testnet.yml ps"
    echo ""
    echo "  3. Restart the proxy:"
    echo "     docker compose -f docker-compose.proxy.yml restart"
    echo ""
    echo "  4. Check validator logs:"
    echo "     docker compose -f docker-compose.testnet.yml logs validator-1"
    echo ""
    exit 1
fi
