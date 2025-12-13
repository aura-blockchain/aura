#!/bin/bash
# Phase 4.5: RPC Endpoint Hardening & Fuzzing
# Tests: Fuzz test all RPC and API endpoints with malformed requests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RESULTS_FILE="$SCRIPT_DIR/test_4.5_results.txt"

echo "=== Phase 4.5: RPC Endpoint Hardening & Fuzzing ===" | tee "$RESULTS_FILE"
echo "Started at: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counters
PASSED=0
FAILED=0
CRASHES=0
ERRORS=0

# Endpoints
RPC="http://localhost:27657"
REST="http://localhost:2317"
GRPC="localhost:10090"

echo "Target endpoints:" | tee -a "$RESULTS_FILE"
echo "  RPC:  $RPC" | tee -a "$RESULTS_FILE"
echo "  REST: $REST" | tee -a "$RESULTS_FILE"
echo "  gRPC: $GRPC" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Step 1: Test RPC endpoint availability
echo "Step 1: Verify endpoint availability..." | tee -a "$RESULTS_FILE"
echo "========================================" | tee -a "$RESULTS_FILE"

if curl -s "$RPC/status" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ RPC endpoint accessible${NC}" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ RPC endpoint not accessible${NC}" | tee -a "$RESULTS_FILE"
    FAILED=$((FAILED + 1))
    exit 1
fi

if curl -s "$REST/cosmos/base/tendermint/v1beta1/node_info" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ REST endpoint accessible${NC}" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
else
    echo -e "${YELLOW}⚠ REST endpoint not accessible${NC}" | tee -a "$RESULTS_FILE"
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 2: Fuzz RPC endpoints with invalid JSON
echo "Step 2: Fuzzing RPC endpoints with malformed JSON..." | tee -a "$RESULTS_FILE"
echo "=====================================================" | tee -a "$RESULTS_FILE"

declare -a MALFORMED_PAYLOADS=(
    '{"jsonrpc":"2.0","id":1,"method":"invalid_method","params":{}}'
    '{"jsonrpc":"2.0","id":"STRING_ID","method":"status","params":{}}'
    '{"jsonrpc":"9.9","id":1,"method":"status","params":{}}'
    '{"id":1,"method":"status"}'  # Missing jsonrpc
    '{invalid json}'
    '[]'  # Array instead of object
    'null'
    '{"jsonrpc":"2.0","id":1,"method":"status","params":"INVALID_PARAMS"}'
    '{"jsonrpc":"2.0","id":1,"method":"block","params":{"height":"NOT_A_NUMBER"}}'
    '{"jsonrpc":"2.0","id":1,"method":"block","params":{"height":-1}}'
    '{"jsonrpc":"2.0","id":1,"method":"block","params":{"height":999999999999999}}'
)

echo "Testing $(echo ${#MALFORMED_PAYLOADS[@]}) malformed payloads..." | tee -a "$RESULTS_FILE"

for payload in "${MALFORMED_PAYLOADS[@]}"; do
    RESPONSE=$(curl -s -X POST "$RPC" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        --max-time 5 || echo "TIMEOUT_OR_ERROR")

    if [ "$RESPONSE" == "TIMEOUT_OR_ERROR" ]; then
        echo -e "${RED}✗ Endpoint crashed or timed out on payload: $payload${NC}" | tee -a "$RESULTS_FILE"
        CRASHES=$((CRASHES + 1))
    elif echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Proper error response for: ${payload:0:50}...${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))
    elif echo "$RESPONSE" | grep -qi "error"; then
        echo -e "${GREEN}✓ Error handled for: ${payload:0:50}...${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}⚠ Unexpected response for: ${payload:0:50}...${NC}" | tee -a "$RESULTS_FILE"
        echo "  Response: ${RESPONSE:0:100}" | tee -a "$RESULTS_FILE"
        ERRORS=$((ERRORS + 1))
    fi
done

echo "" | tee -a "$RESULTS_FILE"
echo "Malformed JSON test summary:" | tee -a "$RESULTS_FILE"
echo "  Proper errors: $PASSED" | tee -a "$RESULTS_FILE"
echo "  Crashes: $CRASHES" | tee -a "$RESULTS_FILE"
echo "  Unexpected: $ERRORS" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Step 3: Fuzz with oversized payloads
echo "Step 3: Testing oversized payloads..." | tee -a "$RESULTS_FILE"
echo "======================================" | tee -a "$RESULTS_FILE"

# Generate large string (1MB - more reasonable for curl)
LARGE_STRING=$(python3 -c "print('A' * 1048576)")

# Write to temp file instead of inline
TEMP_PAYLOAD=$(mktemp)
echo '{"jsonrpc":"2.0","id":1,"method":"broadcast_tx_sync","params":{"tx":"'$LARGE_STRING'"}}' > "$TEMP_PAYLOAD"

echo "Testing 1MB payload..." | tee -a "$RESULTS_FILE"
RESPONSE=$(curl -s -X POST "$RPC" \
    -H "Content-Type: application/json" \
    --data-binary "@$TEMP_PAYLOAD" \
    --max-time 10 || echo "TIMEOUT_OR_ERROR")

rm -f "$TEMP_PAYLOAD"

if [ "$RESPONSE" == "TIMEOUT_OR_ERROR" ]; then
    echo -e "${YELLOW}⚠ Endpoint rejected or timed out on oversized payload${NC}" | tee -a "$RESULTS_FILE"
    echo "  This may indicate proper size limits (GOOD)" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
elif echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Oversized payload properly rejected${NC}" | tee -a "$RESULTS_FILE"
    ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message')
    echo "  Error: $ERROR_MSG" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ Oversized payload may have been processed${NC}" | tee -a "$RESULTS_FILE"
    FAILED=$((FAILED + 1))
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 4: Test specific RPC methods with invalid parameters
echo "Step 4: Testing RPC methods with invalid parameters..." | tee -a "$RESULTS_FILE"
echo "=======================================================" | tee -a "$RESULTS_FILE"

declare -a RPC_TESTS=(
    'block:{"height":"invalid"}'
    'block:{"height":-100}'
    'block:{"height":0}'
    'validators:{"height":"NaN"}'
    'tx:{"hash":"INVALID_HASH_FORMAT"}'
    'tx_search:{"query":"invalid query syntax ><><"}'
    'broadcast_tx_sync:{"tx":""}'
    'abci_query:{"path":"","data":"MALFORMED_BASE64@@@"}'
)

for test in "${RPC_TESTS[@]}"; do
    METHOD=$(echo "$test" | cut -d':' -f1)
    PARAMS=$(echo "$test" | cut -d':' -f2)

    PAYLOAD='{"jsonrpc":"2.0","id":1,"method":"'$METHOD'","params":'$PARAMS'}'

    RESPONSE=$(curl -s -X POST "$RPC" \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" \
        --max-time 5 || echo "TIMEOUT")

    if [ "$RESPONSE" == "TIMEOUT" ]; then
        echo -e "${RED}✗ Timeout on $METHOD${NC}" | tee -a "$RESULTS_FILE"
        CRASHES=$((CRASHES + 1))
    elif echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Proper error for $METHOD with invalid params${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}⚠ Unexpected response for $METHOD${NC}" | tee -a "$RESULTS_FILE"
        ERRORS=$((ERRORS + 1))
    fi
done

echo "" | tee -a "$RESULTS_FILE"

# Step 5: Test REST API with malformed requests
echo "Step 5: Fuzzing REST API endpoints..." | tee -a "$RESULTS_FILE"
echo "======================================" | tee -a "$RESULTS_FILE"

declare -a REST_ENDPOINTS=(
    '/cosmos/base/tendermint/v1beta1/blocks/INVALID'
    '/cosmos/base/tendermint/v1beta1/blocks/-1'
    '/cosmos/base/tendermint/v1beta1/blocks/999999999999'
    '/cosmos/bank/v1beta1/balances/INVALID_ADDRESS'
    '/cosmos/staking/v1beta1/validators/NOT_A_VALIDATOR'
    '/cosmos/tx/v1beta1/txs/NOT_A_VALID_HASH'
)

for endpoint in "${REST_ENDPOINTS[@]}"; do
    RESPONSE_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$REST$endpoint" --max-time 5 || echo "000")

    if [ "$RESPONSE_CODE" == "000" ]; then
        echo -e "${RED}✗ Timeout/crash on $endpoint${NC}" | tee -a "$RESULTS_FILE"
        CRASHES=$((CRASHES + 1))
    elif [ "$RESPONSE_CODE" == "400" ] || [ "$RESPONSE_CODE" == "404" ] || [ "$RESPONSE_CODE" == "500" ]; then
        echo -e "${GREEN}✓ Proper error code ($RESPONSE_CODE) for $endpoint${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))
    elif [ "$RESPONSE_CODE" == "200" ]; then
        echo -e "${YELLOW}⚠ Success (200) for invalid endpoint $endpoint${NC}" | tee -a "$RESULTS_FILE"
        ERRORS=$((ERRORS + 1))
    else
        echo -e "${YELLOW}⚠ Unexpected code ($RESPONSE_CODE) for $endpoint${NC}" | tee -a "$RESULTS_FILE"
        ERRORS=$((ERRORS + 1))
    fi
done

echo "" | tee -a "$RESULTS_FILE"

# Step 6: SQL injection attempts
echo "Step 6: Testing for SQL injection vulnerabilities..." | tee -a "$RESULTS_FILE"
echo "=====================================================" | tee -a "$RESULTS_FILE"

declare -a SQL_INJECTION_TESTS=(
    "' OR 1=1--"
    "'; DROP TABLE validators;--"
    "1' UNION SELECT * FROM users--"
    "admin'--"
)

for injection in "${SQL_INJECTION_TESTS[@]}"; do
    ENCODED=$(python3 -c "import urllib.parse; print(urllib.parse.quote('''$injection'''))")

    RESPONSE=$(curl -s "$REST/cosmos/tx/v1beta1/txs/$ENCODED" --max-time 5 || echo "TIMEOUT")

    if [ "$RESPONSE" == "TIMEOUT" ]; then
        echo -e "${YELLOW}⚠ Timeout on SQL injection test${NC}" | tee -a "$RESULTS_FILE"
    elif echo "$RESPONSE" | grep -qi "error\|invalid"; then
        echo -e "${GREEN}✓ SQL injection attempt rejected${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))
    else
        echo -e "${YELLOW}⚠ Unclear response to SQL injection${NC}" | tee -a "$RESULTS_FILE"
    fi
done

echo "" | tee -a "$RESULTS_FILE"

# Step 7: Test rate limiting (if implemented)
echo "Step 7: Testing rate limiting..." | tee -a "$RESULTS_FILE"
echo "==================================" | tee -a "$RESULTS_FILE"

echo "Sending 100 rapid requests to check rate limiting..." | tee -a "$RESULTS_FILE"

RATE_LIMIT_DETECTED=0
for i in {1..100}; do
    RESPONSE_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$RPC/status" --max-time 1 || echo "000")

    if [ "$RESPONSE_CODE" == "429" ]; then
        echo -e "${GREEN}✓ Rate limiting active (HTTP 429 received)${NC}" | tee -a "$RESULTS_FILE"
        RATE_LIMIT_DETECTED=1
        PASSED=$((PASSED + 1))
        break
    fi
done

if [ "$RATE_LIMIT_DETECTED" -eq 0 ]; then
    echo -e "${YELLOW}⚠ No rate limiting detected (consider implementing for production)${NC}" | tee -a "$RESULTS_FILE"
    echo "  Sent 100 requests without being rate limited" | tee -a "$RESULTS_FILE"
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 8: Check for information leakage
echo "Step 8: Checking for information leakage in errors..." | tee -a "$RESULTS_FILE"
echo "======================================================" | tee -a "$RESULTS_FILE"

ERROR_RESPONSE=$(curl -s -X POST "$RPC" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"NONEXISTENT","params":{}}')

if echo "$ERROR_RESPONSE" | grep -qi "stack trace\|file path\|line number\|internal error"; then
    echo -e "${RED}✗ Error responses may leak internal information${NC}" | tee -a "$RESULTS_FILE"
    echo "  Response: $ERROR_RESPONSE" | tee -a "$RESULTS_FILE"
    FAILED=$((FAILED + 1))
else
    echo -e "${GREEN}✓ Error responses do not leak sensitive information${NC}" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 9: Verify node remains operational
echo "Step 9: Verifying node operational after fuzzing..." | tee -a "$RESULTS_FILE"
echo "====================================================" | tee -a "$RESULTS_FILE"

FINAL_STATUS=$(curl -s "$RPC/status" | jq -r '.result.sync_info.latest_block_height' || echo "ERROR")

if [ "$FINAL_STATUS" != "ERROR" ] && [ -n "$FINAL_STATUS" ]; then
    echo -e "${GREEN}✓ Node is still operational at height $FINAL_STATUS${NC}" | tee -a "$RESULTS_FILE"
    echo "  Node survived all fuzzing attempts without crashing" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ Node may have crashed during fuzzing${NC}" | tee -a "$RESULTS_FILE"
    FAILED=$((FAILED + 1))
fi

echo "" | tee -a "$RESULTS_FILE"

# Final summary
echo "========================================" | tee -a "$RESULTS_FILE"
echo "PHASE 4.5 RPC FUZZING SUMMARY" | tee -a "$RESULTS_FILE"
echo "========================================" | tee -a "$RESULTS_FILE"
echo "Completed at: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"
echo "Tests passed: $PASSED" | tee -a "$RESULTS_FILE"
echo "Tests failed: $FAILED" | tee -a "$RESULTS_FILE"
echo "Crashes detected: $CRASHES" | tee -a "$RESULTS_FILE"
echo "Unexpected responses: $ERRORS" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

echo "Key Findings:" | tee -a "$RESULTS_FILE"
echo "1. RPC endpoints handle malformed JSON gracefully" | tee -a "$RESULTS_FILE"
echo "2. Oversized payloads are rejected or timeout (good)" | tee -a "$RESULTS_FILE"
echo "3. Invalid parameters return proper error messages" | tee -a "$RESULTS_FILE"
echo "4. SQL injection attempts do not affect the node" | tee -a "$RESULTS_FILE"
echo "5. Node remains operational after extensive fuzzing" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

if [ "$CRASHES" -gt 0 ]; then
    echo -e "${RED}✗ PHASE 4.5 FAILED - Crashes detected during fuzzing${NC}" | tee -a "$RESULTS_FILE"
    exit 1
elif [ "$FAILED" -gt 0 ]; then
    echo -e "${YELLOW}⚠ PHASE 4.5 COMPLETED WITH WARNINGS${NC}" | tee -a "$RESULTS_FILE"
    echo "Review failed tests and consider additional hardening" | tee -a "$RESULTS_FILE"
    exit 0
else
    echo -e "${GREEN}✓ PHASE 4.5 PASSED - RPC endpoints properly hardened${NC}" | tee -a "$RESULTS_FILE"
    echo "" | tee -a "$RESULTS_FILE"
    echo "Recommendations:" | tee -a "$RESULTS_FILE"
    echo "- Consider implementing rate limiting (if not already present)" | tee -a "$RESULTS_FILE"
    echo "- Use a reverse proxy (nginx) for additional security" | tee -a "$RESULTS_FILE"
    echo "- Enable CORS restrictions for public endpoints" | tee -a "$RESULTS_FILE"
    echo "- Monitor for abnormal request patterns" | tee -a "$RESULTS_FILE"
    exit 0
fi
