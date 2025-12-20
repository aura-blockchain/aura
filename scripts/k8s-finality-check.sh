#!/bin/bash
# Blockchain Finality Verification for Aura

NAMESPACE="aura"
KUBECTL="kubectl"

# Check if using k3s
if [ -f /etc/rancher/k3s/k3s.yaml ]; then
    KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
    KUBECTL="sudo kubectl --kubeconfig=$KUBECONFIG"
fi

get_validator_pod() {
    local pod=$($KUBECTL get pods -n "$NAMESPACE" -l app=aura-validator -o name 2>/dev/null | head -1)
    if [ -z "$pod" ]; then
        pod=$($KUBECTL get pods -n "$NAMESPACE" -l app.kubernetes.io/component=validator -o name 2>/dev/null | head -1)
    fi
    echo "$pod"
}

echo "=== Aura Finality Verification ==="

POD=$(get_validator_pod)

if [ -z "$POD" ]; then
    echo "[ERROR] No validator pod found"
    exit 1
fi

# Check block production rate
echo "[1] Checking block production..."
echo "Querying $POD for latest block..."

HEIGHT=$($KUBECTL exec -n $NAMESPACE $POD -- curl -s http://localhost:26657/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "N/A")
BLOCK_TIME=$($KUBECTL exec -n $NAMESPACE $POD -- curl -s http://localhost:26657/status 2>/dev/null | jq -r '.result.sync_info.latest_block_time' 2>/dev/null || echo "N/A")

echo "  Current block height: $HEIGHT"
echo "  Latest block time: $BLOCK_TIME"

# Check for finality delays
echo ""
echo "[2] Checking for finality delays..."
$KUBECTL logs -n $NAMESPACE $POD --tail=500 2>/dev/null | grep -i "finality.*delay\|slow.*commit\|timeout.*commit" || echo "No finality delays detected"

# Check consensus participation
echo ""
echo "[3] Checking consensus participation..."
for pod in $($KUBECTL get pods -n $NAMESPACE -l app=aura-validator -o name 2>/dev/null); do
    echo "$pod consensus activity:"
    $KUBECTL logs -n $NAMESPACE $pod --tail=100 2>/dev/null | grep -i "commit\|prevote\|precommit\|round" | tail -5 || echo "  No consensus messages found"
done

# Monitor block confirmations
echo ""
echo "[4] Block confirmation monitoring..."
echo "For production finality verification:"
echo "  - Query RPC: curl http://validator:26657/status | jq '.result.sync_info'"
echo "  - Check: latest_block_height vs latest_block_time"
echo "  - Finality threshold: Typically 2/3+ validators commit"
echo ""
echo "Prometheus metrics (if enabled):"
echo "  - tendermint_consensus_height"
echo "  - tendermint_consensus_validators"
echo "  - tendermint_consensus_missing_validators"

echo ""
echo "=== Finality Health Indicators ==="
echo "Good: Blocks produced every 1-3s, no timeout_commits"
echo "Warning: Blocks >5s apart, occasional timeouts"
echo "Critical: No new blocks >30s, persistent timeouts"
