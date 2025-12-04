#!/bin/bash
# Test script for Aura RPC endpoints
# This script validates that all RPC endpoints are properly configured and accessible

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
RPC_HOST="${RPC_HOST:-localhost}"
RPC_PORT="${RPC_PORT:-26657}"
API_PORT="${API_PORT:-1317}"
GRPC_PORT="${GRPC_PORT:-9090}"
USE_HTTPS="${USE_HTTPS:-false}"

if [ "$USE_HTTPS" = "true" ]; then
    PROTOCOL="https"
    CURL_OPTS="-k" # Allow self-signed certs
else
    PROTOCOL="http"
    CURL_OPTS=""
fi

RPC_URL="${PROTOCOL}://${RPC_HOST}:${RPC_PORT}"
API_URL="${PROTOCOL}://${RPC_HOST}:${API_PORT}"

echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}Aura RPC Endpoint Test Suite${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""
echo "Configuration:"
echo "  RPC URL: $RPC_URL"
echo "  API URL: $API_URL"
echo "  gRPC: ${RPC_HOST}:${GRPC_PORT}"
echo ""

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test endpoint
test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_code="${3:-200}"

    echo -n "Testing $name... "

    response=$(curl -s -w "\n%{http_code}" $CURL_OPTS "$url" 2>/dev/null || echo "000")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}PASSED${NC} (HTTP $http_code)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}FAILED${NC} (Expected HTTP $expected_code, got $http_code)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Helper function to test JSON response
test_json_endpoint() {
    local name="$1"
    local url="$2"
    local jq_filter="${3:-.}"

    echo -n "Testing $name... "

    response=$(curl -s $CURL_OPTS "$url" 2>/dev/null)

    if echo "$response" | jq -e "$jq_filter" >/dev/null 2>&1; then
        echo -e "${GREEN}PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}FAILED${NC}"
        echo "  Response: $response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}1. Testing Tendermint RPC Endpoints${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""

# Note: These tests will only work if a node is actually running
# For now, we'll create the test framework

test_endpoint "RPC Health Check" "$RPC_URL/health" "200" || true
test_json_endpoint "RPC Status" "$RPC_URL/status" ".result.node_info.network" || true
test_json_endpoint "RPC Latest Block" "$RPC_URL/block" ".result.block.header.height" || true
test_json_endpoint "RPC ABCI Info" "$RPC_URL/abci_info" ".result.response.version" || true
test_json_endpoint "RPC Net Info" "$RPC_URL/net_info" ".result.n_peers" || true
test_json_endpoint "RPC Genesis" "$RPC_URL/genesis" ".result.genesis.chain_id" || true

echo ""
echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}2. Testing Cosmos REST API Endpoints${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""

test_json_endpoint "API Node Info" "$API_URL/cosmos/base/tendermint/v1beta1/node_info" ".default_node_info.network" || true
test_json_endpoint "API Latest Block" "$API_URL/cosmos/base/tendermint/v1beta1/blocks/latest" ".block.header.height" || true
test_json_endpoint "API Syncing" "$API_URL/cosmos/base/tendermint/v1beta1/syncing" ".syncing" || true
test_json_endpoint "API Staking Params" "$API_URL/cosmos/staking/v1beta1/params" ".params.unbonding_time" || true

echo ""
echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}3. Testing CORS Headers${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""

echo -n "Testing CORS preflight (OPTIONS)... "
cors_response=$(curl -s -X OPTIONS -H "Origin: http://example.com" -H "Access-Control-Request-Method: GET" -I $CURL_OPTS "$RPC_URL/status" 2>/dev/null || echo "")

if echo "$cors_response" | grep -qi "access-control-allow-origin"; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC} (CORS headers not found)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo -n "Testing CORS headers on GET... "
cors_get=$(curl -s -H "Origin: http://example.com" -I $CURL_OPTS "$RPC_URL/status" 2>/dev/null || echo "")

if echo "$cors_get" | grep -qi "access-control-allow-origin"; then
    echo -e "${GREEN}PASSED${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC} (CORS headers not found)"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi

echo ""
echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}4. Testing Rate Limiting${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""

echo "Sending rapid requests to test rate limiting..."
echo -n "Rate limit test (50 requests in quick succession)... "

rate_limit_hit=false
for i in {1..50}; do
    status_code=$(curl -s -o /dev/null -w "%{http_code}" $CURL_OPTS "$RPC_URL/health" 2>/dev/null || echo "000")
    if [ "$status_code" = "429" ] || [ "$status_code" = "503" ]; then
        rate_limit_hit=true
        break
    fi
    sleep 0.01
done

if [ "$rate_limit_hit" = true ]; then
    echo -e "${GREEN}PASSED${NC} (Rate limiting is active)"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${YELLOW}WARNING${NC} (Rate limit not triggered - may not be configured or limit is very high)"
    # Don't count as failure since node might not be running with nginx
fi

echo ""
echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}5. Testing SSL/TLS (if HTTPS)${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""

if [ "$USE_HTTPS" = "true" ]; then
    echo -n "Testing SSL certificate... "

    cert_info=$(openssl s_client -connect ${RPC_HOST}:443 -servername ${RPC_HOST} </dev/null 2>/dev/null | openssl x509 -noout -subject 2>/dev/null || echo "")

    if [ -n "$cert_info" ]; then
        echo -e "${GREEN}PASSED${NC}"
        echo "  Certificate: $cert_info"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC} (Could not retrieve certificate)"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo "Skipping SSL tests (USE_HTTPS=false)"
fi

echo ""
echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}6. Testing gRPC Endpoint${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""

if command -v grpcurl >/dev/null 2>&1; then
    echo -n "Testing gRPC service list... "

    grpc_services=$(grpcurl -plaintext ${RPC_HOST}:${GRPC_PORT} list 2>/dev/null || echo "")

    if echo "$grpc_services" | grep -q "cosmos"; then
        echo -e "${GREEN}PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}FAILED${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
else
    echo -e "${YELLOW}WARNING${NC} grpcurl not installed, skipping gRPC tests"
    echo "  Install with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi

echo ""
echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""
echo "Total tests run: $((TESTS_PASSED + TESTS_FAILED))"
echo -e "${GREEN}Tests passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Check configuration.${NC}"
    echo ""
    echo "Note: Many tests require a running Aura node."
    echo "Start a node with: docker-compose -f docker/docker-compose.rpc.yml up -d"
    exit 1
fi
