#!/bin/bash
# Simple verification script for monitoring setup

echo "Verifying monitoring setup..."
echo ""

# Check files exist
echo "1. Alert rules: $([ -f prometheus/rules/wasm-bridge-alerts.yml ] && echo '✓' || echo '✗')"
echo "2. Grafana dashboard: $([ -f grafana/dashboards/wasm-bridge-security.json ] && echo '✓' || echo '✗')"
echo "3. WASM telemetry: $([ -f chain/x/wasm/keeper/telemetry.go ] && echo '✓' || echo '✗')"
echo "4. Bridge telemetry: $([ -f chain/x/bridge/keeper/telemetry.go ] && echo '✓' || echo '✗')"
echo "5. Documentation: $([ -f docs/monitoring/wasm-bridge-security-monitoring.md ] && echo '✓' || echo '✗')"
echo ""

# Count alert rules
ALERT_COUNT=$(grep -c "alert:" prometheus/rules/wasm-bridge-alerts.yml 2>/dev/null || echo "0")
echo "Alert rules defined: $ALERT_COUNT"
echo ""

# Check code compiles
echo "Testing compilation..."
cd chain
if go build ./x/bridge/keeper/... 2>&1 | grep -q "error"; then
    echo "Bridge keeper: ✗ (compilation errors)"
else
    echo "Bridge keeper: ✓"
fi

if go build ./x/wasm/keeper/... 2>&1 | grep -q "error"; then
    echo "WASM keeper: ✗ (compilation errors)"
else
    echo "WASM keeper: ✓"
fi

echo ""
echo "Verification complete!"
