#!/bin/bash
# Slashing Behavior Detection for Aura Validators

NAMESPACE="aura"
KUBECTL="kubectl"

# Check if using k3s
if [ -f /etc/rancher/k3s/k3s.yaml ]; then
    KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
    KUBECTL="sudo kubectl --kubeconfig=$KUBECONFIG"
fi

get_validator_pods() {
    local pods=$($KUBECTL get pods -n "$NAMESPACE" -l app=aura-validator -o name 2>/dev/null)
    if [ -z "$pods" ]; then
        pods=$($KUBECTL get pods -n "$NAMESPACE" -l app.kubernetes.io/component=validator -o name 2>/dev/null)
    fi
    echo "$pods"
}

echo "=== Aura Slashing Detection ==="

# Check for double-signing indicators
echo "[1] Checking for double-signing patterns..."
$KUBECTL logs -n $NAMESPACE -l app=aura-validator --tail=1000 2>/dev/null | grep -i "double.*sign\|duplicate.*vote\|conflicting.*block" || \
$KUBECTL logs -n $NAMESPACE -l app.kubernetes.io/component=validator --tail=1000 2>/dev/null | grep -i "double.*sign\|duplicate.*vote\|conflicting.*block" || \
echo "No double-signing detected"

# Check for downtime/missed blocks
echo ""
echo "[2] Checking validator uptime..."
for pod in $(get_validator_pods); do
    echo "Pod: $pod"
    RESTARTS=$($KUBECTL get $pod -n $NAMESPACE -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || echo "N/A")
    READY=$($KUBECTL get $pod -n $NAMESPACE -o jsonpath='{.status.containerStatuses[0].ready}' 2>/dev/null || echo "N/A")
    echo "  Restarts: $RESTARTS"
    echo "  Ready: $READY"

    $KUBECTL logs -n $NAMESPACE $pod --tail=100 2>/dev/null | grep -i "missed.*block\|timeout.*proposal\|validator.*offline" || echo "  No missed blocks detected"
done

# Check for byzantine behavior
echo ""
echo "[3] Checking for byzantine behavior..."
$KUBECTL logs -n $NAMESPACE -l app=aura-validator --tail=1000 2>/dev/null | grep -i "byzantine\|malicious\|invalid.*signature\|equivocation" || \
$KUBECTL logs -n $NAMESPACE -l app.kubernetes.io/component=validator --tail=1000 2>/dev/null | grep -i "byzantine\|malicious\|invalid.*signature\|equivocation" || \
echo "No byzantine behavior detected"

# Network partition detection
echo ""
echo "[4] Checking for network partitions..."
$KUBECTL logs -n $NAMESPACE -l app=aura-validator --tail=500 2>/dev/null | grep -i "partition\|split.*brain\|isolated\|peer.*disconnect" || \
$KUBECTL logs -n $NAMESPACE -l app.kubernetes.io/component=validator --tail=500 2>/dev/null | grep -i "partition\|split.*brain\|isolated\|peer.*disconnect" || \
echo "No partition detected"

echo ""
echo "=== Slashable Conditions ==="
echo "Monitor for these patterns that trigger slashing:"
echo "  - Double signing: Signing two different blocks at same height"
echo "  - Downtime: Missing >500 consecutive blocks (configurable)"
echo "  - Byzantine behavior: Invalid signatures, malformed proposals"
echo ""
echo "To simulate (DANGEROUS - testnet only):"
echo "  - Double-sign: Run same validator with same key on 2 pods"
echo "  - Downtime: kubectl scale statefulset aura-validator --replicas=0"
