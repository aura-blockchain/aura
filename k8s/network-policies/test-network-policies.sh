#!/bin/bash
# Test script for Aura network policies
# Verifies egress restrictions and ingress allowances
set -euo pipefail

NAMESPACE="${NAMESPACE:-aura}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_test() { echo -e "${BLUE}[TEST]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; }
log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; }

PASSED=0
FAILED=0
SKIPPED=0

echo "==================================="
echo "Aura Network Policy Testing"
echo "==================================="
echo "Namespace: $NAMESPACE"
echo ""

ensure_test_pod() {
    if ! kubectl get pod netpol-test -n "$NAMESPACE" &>/dev/null; then
        echo "Creating network policy test pod..."
        kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: netpol-test
  namespace: $NAMESPACE
  labels:
    app: netpol-test
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: RuntimeDefault
  containers:
  - name: curl
    image: curlimages/curl:8.5.0
    command: ["sleep", "infinity"]
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
EOF
        echo "Waiting for test pod to be ready..."
        kubectl wait --for=condition=ready pod/netpol-test -n "$NAMESPACE" --timeout=60s
    fi
}

test_dns() {
    log_test "DNS resolution to kube-system"
    if kubectl exec -n "$NAMESPACE" netpol-test -c curl -- nslookup kubernetes.default.svc.cluster.local > /dev/null 2>&1; then
        log_pass "DNS resolution works"
        PASSED=$((PASSED+1))
    else
        log_fail "DNS resolution blocked"
        FAILED=$((FAILED+1))
    fi
}

test_external_blocked() {
    log_test "External network access (should be BLOCKED)"
    if timeout 5 kubectl exec -n "$NAMESPACE" netpol-test -c curl -- curl -s -m 3 http://www.google.com > /dev/null 2>&1; then
        log_fail "External network access allowed (security issue!)"
        FAILED=$((FAILED+1))
    else
        log_pass "External network access blocked as expected"
        PASSED=$((PASSED+1))
    fi
}

test_monitoring_access() {
    log_test "Access to monitoring namespace"
    PROM_IP=$(kubectl get pod -n monitoring -l app.kubernetes.io/name=prometheus -o jsonpath='{.items[0].status.podIP}' 2>/dev/null || echo "")
    if [ -n "$PROM_IP" ]; then
        if timeout 5 kubectl exec -n "$NAMESPACE" netpol-test -c curl -- curl -s -m 3 "http://$PROM_IP:9090/-/healthy" > /dev/null 2>&1; then
            log_pass "Monitoring namespace accessible"
            PASSED=$((PASSED+1))
        else
            log_fail "Monitoring namespace blocked"
            FAILED=$((FAILED+1))
        fi
    else
        log_skip "No Prometheus pod found in monitoring namespace"
        SKIPPED=$((SKIPPED+1))
    fi
}

test_intra_namespace() {
    log_test "Intra-namespace communication"
    # Try common validator label patterns
    VALIDATOR_IP=$(kubectl get pod -n "$NAMESPACE" -l app.kubernetes.io/component=validator -o jsonpath='{.items[0].status.podIP}' 2>/dev/null || echo "")
    if [ -z "$VALIDATOR_IP" ]; then
        VALIDATOR_IP=$(kubectl get pod -n "$NAMESPACE" -l app=aura-validator -o jsonpath='{.items[0].status.podIP}' 2>/dev/null || echo "")
    fi
    if [ -z "$VALIDATOR_IP" ]; then
        VALIDATOR_IP=$(kubectl get pod -n "$NAMESPACE" -l component=validator -o jsonpath='{.items[0].status.podIP}' 2>/dev/null || echo "")
    fi
    if [ -n "$VALIDATOR_IP" ]; then
        if timeout 5 kubectl exec -n "$NAMESPACE" netpol-test -c curl -- curl -s -m 3 "http://$VALIDATOR_IP:26657/health" > /dev/null 2>&1; then
            log_pass "Intra-namespace communication works"
            PASSED=$((PASSED+1))
        else
            log_fail "Intra-namespace communication blocked"
            FAILED=$((FAILED+1))
        fi
    else
        log_skip "No validator pod found"
        SKIPPED=$((SKIPPED+1))
    fi
}

test_unauthorized_port() {
    log_test "Access to unauthorized external port"
    if timeout 5 kubectl exec -n "$NAMESPACE" netpol-test -c curl -- curl -s -m 3 http://8.8.8.8:80 > /dev/null 2>&1; then
        log_fail "Unauthorized egress allowed (security issue!)"
        FAILED=$((FAILED+1))
    else
        log_pass "Unauthorized egress blocked as expected"
        PASSED=$((PASSED+1))
    fi
}

test_p2p_external() {
    log_test "P2P port (26656) egress"

    # Check if external-validators namespace exists with simulation pod
    EXT_VAL_IP=$(kubectl get pod -n external-validators -l app=external-validator -o jsonpath='{.items[0].status.podIP}' 2>/dev/null || echo "")

    if [ -z "$EXT_VAL_IP" ]; then
        log_skip "P2P egress test requires external-validators namespace (run: kubectl apply -f k8s/external-validator-sim.yaml)"
        SKIPPED=$((SKIPPED+1))
        return
    fi

    # Test P2P connectivity from aura namespace to external validator
    # Use curl -v and check for "Connected to" in stderr to verify TCP connection works
    # P2P ports don't speak HTTP, so curl will error but connection proves network policy works
    # Note: -c curl required when Linkerd sidecar is injected
    CURL_OUTPUT=$(timeout 5 kubectl exec -n "$NAMESPACE" netpol-test -c curl -- curl -v --connect-timeout 3 "http://$EXT_VAL_IP:26656" 2>&1 || true)

    if echo "$CURL_OUTPUT" | grep -q "Connected to"; then
        log_pass "P2P egress to external validator works ($EXT_VAL_IP:26656)"
        PASSED=$((PASSED+1))
    elif echo "$CURL_OUTPUT" | grep -q "Connection timed out\|Connection refused"; then
        log_fail "P2P egress to external validator blocked"
        FAILED=$((FAILED+1))
    else
        log_fail "P2P egress test inconclusive"
        FAILED=$((FAILED+1))
    fi
}

test_policy_count() {
    log_test "Network policy count verification"
    POLICY_COUNT=$(kubectl get networkpolicies -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
    echo "   Total network policies in $NAMESPACE namespace: $POLICY_COUNT"
    if [ "$POLICY_COUNT" -ge 1 ]; then
        log_pass "Network policies present ($POLICY_COUNT policies)"
        PASSED=$((PASSED+1))
    else
        log_fail "No network policies found"
        FAILED=$((FAILED+1))
    fi
}

test_vault_access() {
    log_test "Access to Vault namespace"
    VAULT_IP=$(kubectl get pod -n vault -l app.kubernetes.io/name=vault -o jsonpath='{.items[0].status.podIP}' 2>/dev/null || echo "")
    if [ -n "$VAULT_IP" ]; then
        if timeout 5 kubectl exec -n "$NAMESPACE" netpol-test -c curl -- curl -s -m 3 "http://$VAULT_IP:8200/v1/sys/health" > /dev/null 2>&1; then
            log_pass "Vault namespace accessible"
            PASSED=$((PASSED+1))
        else
            log_fail "Vault namespace blocked"
            FAILED=$((FAILED+1))
        fi
    else
        log_skip "No Vault pod found"
        SKIPPED=$((SKIPPED+1))
    fi
}

cleanup() {
    read -p "Delete test pod? (y/N): " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        kubectl delete pod netpol-test -n "$NAMESPACE" --ignore-not-found
        echo "Test pod deleted"
    fi
}

main() {
    ensure_test_pod

    echo ""
    test_dns
    echo ""
    test_external_blocked
    echo ""
    test_monitoring_access
    echo ""
    test_intra_namespace
    echo ""
    test_unauthorized_port
    echo ""
    test_p2p_external
    echo ""
    test_policy_count
    echo ""
    test_vault_access

    echo ""
    echo "==================================="
    echo "Test Summary"
    echo "==================================="
    echo -e "Passed:  ${GREEN}$PASSED${NC}"
    echo -e "Failed:  ${RED}$FAILED${NC}"
    echo -e "Skipped: ${YELLOW}$SKIPPED${NC}"
    echo "==================================="

    if [ "$FAILED" -gt 0 ]; then
        echo -e "${RED}NETWORK POLICY TESTS FAILED${NC}"
        echo "Review network policies and fix security issues"
        exit 1
    else
        echo -e "${GREEN}NETWORK POLICY TESTS PASSED${NC}"
    fi
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace|-n)
            NAMESPACE="$2"
            shift 2
            ;;
        --cleanup)
            NAMESPACE="${2:-aura}"
            kubectl delete pod netpol-test -n "$NAMESPACE" --ignore-not-found
            exit 0
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --namespace, -n NAME  Namespace to test (default: aura)"
            echo "  --cleanup             Delete test pod and exit"
            echo "  --help, -h            Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

main
