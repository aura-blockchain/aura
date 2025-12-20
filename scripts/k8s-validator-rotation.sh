#!/bin/bash
# Validator Key Rotation Script for K8s

set -e

NAMESPACE="aura"
PROJECT_ROOT="/home/hudson/blockchain-projects/aura"
KUBECTL="kubectl"

# Check if using k3s
if [ -f /etc/rancher/k3s/k3s.yaml ]; then
    KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
    KUBECTL="sudo kubectl --kubeconfig=$KUBECONFIG"
fi

echo "=== Aura Validator Key Rotation ==="

# Step 1: Backup current StatefulSet
echo "[1/5] Backing up current StatefulSet..."
$KUBECTL get statefulset aura-validator -n $NAMESPACE -o yaml > /tmp/aura-validator-backup.yaml 2>/dev/null || \
    echo "  StatefulSet not found, skipping backup"

# Step 2: Check for rotated keys secret
echo "[2/5] Checking for rotated keys secret..."
if $KUBECTL get secret validator-keys-rotated -n $NAMESPACE &>/dev/null; then
    echo "  Rotated keys secret exists"
else
    echo "  Creating placeholder rotated keys secret..."
    $KUBECTL apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: validator-keys-rotated
  namespace: $NAMESPACE
type: Opaque
data:
  # Replace with actual base64-encoded keys
  priv_validator_key.json: e30=  # Empty JSON placeholder
  node_key.json: e30=  # Empty JSON placeholder
EOF
    echo "  IMPORTANT: Update validator-keys-rotated with actual keys before rotating"
fi

# Step 3: Patch StatefulSet to use new secret
echo "[3/5] Preparing StatefulSet patch..."
echo "  To apply key rotation, patch the StatefulSet volumes:"
echo ""
echo "  $KUBECTL patch statefulset aura-validator -n $NAMESPACE --type json -p '["
echo "    {"
echo "      \"op\": \"replace\","
echo "      \"path\": \"/spec/template/spec/volumes/0/secret/secretName\","
echo "      \"value\": \"validator-keys-rotated\""
echo "    }"
echo "  ]'"
echo ""

# Step 4: Rolling restart
echo "[4/5] To perform rolling restart:"
echo "  $KUBECTL rollout restart statefulset/aura-validator -n $NAMESPACE"
echo ""

# Step 5: Verification
echo "[5/5] Verification commands:"
echo "  $KUBECTL rollout status statefulset/aura-validator -n $NAMESPACE --timeout=5m"
echo "  $KUBECTL get pods -n $NAMESPACE -l app=aura-validator"
echo ""

echo "=== Key Rotation Procedure ==="
echo ""
echo "1. Generate new validator keys on secure machine"
echo "2. Base64 encode and update validator-keys-rotated secret"
echo "3. Apply StatefulSet patch to reference new secret"
echo "4. Trigger rolling restart"
echo "5. Monitor consensus participation"
echo ""
echo "CAUTION: Incorrect key rotation can cause double-signing (slashing)!"
echo "Always test in testnet first."
