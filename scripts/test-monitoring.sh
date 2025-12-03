#!/bin/bash
# ============================================================================
# Test script for WASM & Bridge Security Monitoring (Task #16)
# ============================================================================
# Validates that monitoring infrastructure is correctly configured

set -e

echo "============================================"
echo "WASM & Bridge Monitoring Validation"
echo "============================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
test_result() {
    local test_name="$1"
    local result="$2"

    if [ "$result" = "0" ]; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        ((TESTS_FAILED++))
    fi
}

# ============================================================================
# Test 1: Verify alert rules file exists
# ============================================================================
echo "Test 1: Checking alert rules file..."
if [ -f "/home/decri/blockchain-projects/aura/prometheus/rules/wasm-bridge-alerts.yml" ]; then
    test_result "Alert rules file exists" 0
else
    test_result "Alert rules file exists" 1
fi

# ============================================================================
# Test 2: Verify Grafana dashboard exists
# ============================================================================
echo "Test 2: Checking Grafana dashboard..."
if [ -f "/home/decri/blockchain-projects/aura/grafana/dashboards/wasm-bridge-security.json" ]; then
    test_result "Grafana dashboard file exists" 0
else
    test_result "Grafana dashboard file exists" 1
fi

# ============================================================================
# Test 3: Verify telemetry files exist
# ============================================================================
echo "Test 3: Checking telemetry implementation files..."
if [ -f "/home/decri/blockchain-projects/aura/chain/x/wasm/keeper/telemetry.go" ]; then
    test_result "WASM telemetry file exists" 0
else
    test_result "WASM telemetry file exists" 1
fi

if [ -f "/home/decri/blockchain-projects/aura/chain/x/bridge/keeper/telemetry.go" ]; then
    test_result "Bridge telemetry file exists" 0
else
    test_result "Bridge telemetry file exists" 1
fi

# ============================================================================
# Test 4: Verify documentation exists
# ============================================================================
echo "Test 4: Checking documentation..."
if [ -f "/home/decri/blockchain-projects/aura/docs/monitoring/wasm-bridge-security-monitoring.md" ]; then
    test_result "Monitoring documentation exists" 0
else
    test_result "Monitoring documentation exists" 1
fi

# ============================================================================
# Test 5: Check YAML syntax of alert rules
# ============================================================================
echo "Test 5: Validating alert rules YAML syntax..."
if command -v yamllint &> /dev/null; then
    if yamllint -d relaxed /home/decri/blockchain-projects/aura/prometheus/rules/wasm-bridge-alerts.yml &> /dev/null; then
        test_result "Alert rules YAML syntax valid" 0
    else
        test_result "Alert rules YAML syntax valid" 1
    fi
else
    echo -e "${YELLOW}⚠${NC} yamllint not installed - skipping YAML validation"
fi

# ============================================================================
# Test 6: Check JSON syntax of Grafana dashboard
# ============================================================================
echo "Test 6: Validating Grafana dashboard JSON syntax..."
if command -v jq &> /dev/null; then
    if jq empty /home/decri/blockchain-projects/aura/grafana/dashboards/wasm-bridge-security.json &> /dev/null; then
        test_result "Grafana dashboard JSON syntax valid" 0
    else
        test_result "Grafana dashboard JSON syntax valid" 1
    fi
else
    echo -e "${YELLOW}⚠${NC} jq not installed - skipping JSON validation"
fi

# ============================================================================
# Test 7: Verify key alert rules are defined
# ============================================================================
echo "Test 7: Checking for required alert rules..."

REQUIRED_ALERTS=(
    "WasmTxFailureRateHigh"
    "StateLoadErrorsDetected"
    "SignatureMismatchRateHigh"
    "WasmCircuitBreakerOpen"
    "SignatureFailuresByChain"
    "UnmarshalFailuresDetected"
)

for alert in "${REQUIRED_ALERTS[@]}"; do
    if grep -q "$alert" /home/decri/blockchain-projects/aura/prometheus/rules/wasm-bridge-alerts.yml; then
        test_result "Alert rule '$alert' defined" 0
    else
        test_result "Alert rule '$alert' defined" 1
    fi
done

# ============================================================================
# Test 8: Verify Prometheus metrics are defined in code
# ============================================================================
echo "Test 8: Checking for Prometheus metric definitions..."

REQUIRED_WASM_METRICS=(
    "wasmTxFailuresTotal"
    "wasmCircuitBreakerState"
    "wasmValidationCacheHitsTotal"
)

for metric in "${REQUIRED_WASM_METRICS[@]}"; do
    if grep -q "$metric" /home/decri/blockchain-projects/aura/chain/x/wasm/keeper/telemetry.go; then
        test_result "WASM metric '$metric' defined" 0
    else
        test_result "WASM metric '$metric' defined" 1
    fi
done

REQUIRED_BRIDGE_METRICS=(
    "bridgeSignatureMismatchesTotal"
    "bridgeInvalidRecoveryIDTotal"
    "stateLoadErrorsTotal"
)

for metric in "${REQUIRED_BRIDGE_METRICS[@]}"; do
    if grep -q "$metric" /home/decri/blockchain-projects/aura/chain/x/bridge/keeper/telemetry.go; then
        test_result "Bridge metric '$metric' defined" 0
    else
        test_result "Bridge metric '$metric' defined" 1
    fi
done

# ============================================================================
# Test 9: Verify telemetry is integrated into keeper methods
# ============================================================================
echo "Test 9: Checking telemetry integration..."

if grep -q "recordSignatureVerification" /home/decri/blockchain-projects/aura/chain/x/bridge/keeper/keeper.go; then
    test_result "Signature verification telemetry integrated" 0
else
    test_result "Signature verification telemetry integrated" 1
fi

if grep -q "recordSignatureMismatch" /home/decri/blockchain-projects/aura/chain/x/bridge/keeper/keeper.go; then
    test_result "Signature mismatch telemetry integrated" 0
else
    test_result "Signature mismatch telemetry integrated" 1
fi

# ============================================================================
# Test 10: Check Go code compiles
# ============================================================================
echo "Test 10: Verifying Go code compilation..."

cd /home/decri/blockchain-projects/aura/chain

if go build ./x/bridge/keeper/... &> /dev/null; then
    test_result "Bridge keeper compiles" 0
else
    test_result "Bridge keeper compiles" 1
fi

if go build ./x/wasm/keeper/... &> /dev/null; then
    test_result "WASM keeper compiles" 0
else
    test_result "WASM keeper compiles" 1
fi

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "============================================"
echo "Test Results Summary"
echo "============================================"
echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
echo -e "${RED}Failed:${NC} $TESTS_FAILED"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "Monitoring infrastructure is correctly configured."
    echo ""
    echo "Next steps:"
    echo "  1. Start Prometheus: prometheus --config.file=/home/decri/blockchain-projects/aura/prometheus/prometheus.yml"
    echo "  2. Start Grafana and import dashboard: /home/decri/blockchain-projects/aura/grafana/dashboards/wasm-bridge-security.json"
    echo "  3. Run Aura node and verify metrics are being collected"
    echo "  4. Review documentation: docs/monitoring/wasm-bridge-security-monitoring.md"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo "Please review failed tests above and fix issues."
    exit 1
fi
