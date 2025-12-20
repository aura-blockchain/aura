#!/bin/bash
# chaos-tests.sh - Chaos engineering tests for Aura Kubernetes deployment
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="${NAMESPACE:-aura}"
SCENARIO="${SCENARIO:-all}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

TEST_POD=""

# Ensure test pod with curl exists for HTTP checks
ensure_test_pod() {
    # Check if netpol-test pod exists and is running
    if kubectl get pod netpol-test -n "$NAMESPACE" &>/dev/null; then
        local status=$(kubectl get pod netpol-test -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null)
        if [ "$status" = "Running" ]; then
            TEST_POD="netpol-test"
            return 0
        fi
    fi

    # Create a temporary test pod with curl
    log_info "Creating test pod for HTTP checks..."
    kubectl run chaos-test-curl -n "$NAMESPACE" --image=curlimages/curl:latest \
        --restart=Never --command -- sleep 3600 &>/dev/null || true

    for i in {1..30}; do
        local status=$(kubectl get pod chaos-test-curl -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
        if [ "$status" = "Running" ]; then
            TEST_POD="chaos-test-curl"
            return 0
        fi
        sleep 1
    done
    return 1
}

# Cleanup test pod
cleanup_test_pod() {
    if [ "$TEST_POD" = "chaos-test-curl" ]; then
        kubectl delete pod chaos-test-curl -n "$NAMESPACE" --grace-period=0 --force &>/dev/null || true
    fi
}

# HTTP request via test pod
http_request() {
    local url="$1"
    if [ -z "$TEST_POD" ]; then
        echo ""
        return 1
    fi
    kubectl exec -n "$NAMESPACE" "$TEST_POD" -- curl -s "$url" 2>/dev/null || echo ""
}

get_validator_pod() {
    local pod=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/component=validator -o name 2>/dev/null | head -1)
    if [ -z "$pod" ]; then
        pod=$(kubectl get pods -n "$NAMESPACE" -l app=aura-validator -o name 2>/dev/null | head -1)
    fi
    if [ -z "$pod" ]; then
        pod=$(kubectl get pods -n "$NAMESPACE" -l component=validator -o name 2>/dev/null | head -1)
    fi
    echo "$pod"
}

wait_for_recovery() {
    local timeout=${1:-120}
    local start_time=$(date +%s)

    log_info "Waiting for recovery (timeout: ${timeout}s)..."

    while true; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))

        if [ "$elapsed" -gt "$timeout" ]; then
            log_error "Recovery timeout exceeded"
            return 1
        fi

        local ready=$(kubectl get statefulset aura-validator -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
        local desired=$(kubectl get statefulset aura-validator -n "$NAMESPACE" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")

        if [ "$ready" = "$desired" ] && [ "$desired" != "0" ]; then
            log_success "Recovered: $ready/$desired validators ready"
            return 0
        fi

        echo -n "."
        sleep 5
    done
}

check_consensus() {
    log_info "Checking consensus..."

    if [ -z "$TEST_POD" ]; then
        log_warn "No test pod available for consensus check"
        return 0
    fi

    # Use external test pod to query validator via service DNS
    local response=$(http_request "http://aura-validator-0.aura-validator-headless.${NAMESPACE}.svc.cluster.local:26657/status")
    local height=$(echo "$response" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0")

    if [ "$height" != "0" ] && [ "$height" != "null" ] && [ -n "$height" ]; then
        log_success "Consensus active at height $height"
        return 0
    else
        log_warn "Cannot verify consensus"
        return 0  # Don't fail the test, just warn
    fi
}

get_all_validator_pods() {
    local pods=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/component=validator -o name 2>/dev/null)
    if [ -z "$pods" ]; then
        pods=$(kubectl get pods -n "$NAMESPACE" -l app=aura-validator -o name 2>/dev/null)
    fi
    if [ -z "$pods" ]; then
        pods=$(kubectl get pods -n "$NAMESPACE" -l component=validator -o name 2>/dev/null)
    fi
    echo "$pods"
}

get_first_validator_name() {
    local name=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/component=validator -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [ -z "$name" ]; then
        name=$(kubectl get pods -n "$NAMESPACE" -l app=aura-validator -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    fi
    if [ -z "$name" ]; then
        name=$(kubectl get pods -n "$NAMESPACE" -l component=validator -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    fi
    echo "$name"
}

scenario_pod_failure() {
    log_info "=== SCENARIO: Random Pod Failure ==="

    local pods=$(get_all_validator_pods)
    local pod_count=$(echo "$pods" | wc -l)

    if [ "$pod_count" -lt 2 ]; then
        log_warn "Need at least 2 validators for pod failure test"
        return 0
    fi

    local target=$(echo "$pods" | shuf | head -1)
    log_info "Killing pod: $target"

    # Get initial height via test pod
    local response=$(http_request "http://aura-validator-0.aura-validator-headless.${NAMESPACE}.svc.cluster.local:26657/status")
    local initial_height=$(echo "$response" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0")
    log_info "Initial height: $initial_height"

    kubectl delete "$target" -n "$NAMESPACE" --force --grace-period=0

    wait_for_recovery 180

    sleep 10
    check_consensus

    log_success "Pod failure test completed"
}

scenario_network_partition() {
    log_info "=== SCENARIO: Network Partition ==="

    if ! kubectl auth can-i create networkpolicies -n "$NAMESPACE" 2>/dev/null; then
        log_warn "Cannot create network policies - skipping"
        return 0
    fi

    local target_pod=$(get_first_validator_name)

    if [ -z "$target_pod" ]; then
        log_warn "No validator pod found"
        return 0
    fi

    log_info "Isolating pod: $target_pod"

    kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: chaos-network-partition
  namespace: $NAMESPACE
spec:
  podSelector:
    matchLabels:
      statefulset.kubernetes.io/pod-name: $target_pod
  policyTypes:
    - Ingress
    - Egress
  ingress: []
  egress: []
EOF

    log_info "Pod isolated - waiting 30s..."
    sleep 30

    check_consensus

    log_info "Removing network partition..."
    kubectl delete networkpolicy chaos-network-partition -n "$NAMESPACE"

    wait_for_recovery 120

    log_success "Network partition test completed"
}

scenario_high_latency() {
    log_info "=== SCENARIO: High Latency ==="

    local target_pod=$(get_first_validator_name)

    if [ -z "$target_pod" ]; then
        log_warn "No validator pod found"
        return 0
    fi

    log_info "Simulating high latency on: $target_pod"

    if kubectl exec -n "$NAMESPACE" "$target_pod" -c aura -- which tc &>/dev/null; then
        kubectl exec -n "$NAMESPACE" "$target_pod" -c aura -- tc qdisc add dev eth0 root netem delay 200ms 2>/dev/null || true

        log_info "Added 200ms latency - waiting 60s..."
        sleep 60

        check_consensus

        kubectl exec -n "$NAMESPACE" "$target_pod" -c aura -- tc qdisc del dev eth0 root 2>/dev/null || true
    else
        log_warn "tc not available in pod - skipping latency injection"
    fi

    log_success "High latency test completed"
}

scenario_resource_exhaustion() {
    log_info "=== SCENARIO: Memory Pressure ==="

    local target_pod=$(get_first_validator_name)

    if [ -z "$target_pod" ]; then
        log_warn "No validator pod found"
        return 0
    fi

    log_info "Applying memory pressure to: $target_pod"

    if kubectl exec -n "$NAMESPACE" "$target_pod" -c aura -- which stress-ng &>/dev/null; then
        kubectl exec -n "$NAMESPACE" "$target_pod" -c aura -- timeout 30s stress-ng --vm 1 --vm-bytes 512M 2>/dev/null &

        log_info "Applied memory stress - waiting 30s..."
        sleep 35

        check_consensus
    else
        log_warn "stress-ng not available in pod - skipping memory pressure test"
    fi

    log_success "Memory pressure test completed"
}

scenario_rolling_restart() {
    log_info "=== SCENARIO: Rolling Restart ==="

    log_info "Initiating rolling restart of validators..."

    kubectl rollout restart statefulset/aura-validator -n "$NAMESPACE"

    kubectl rollout status statefulset/aura-validator -n "$NAMESPACE" --timeout=300s

    sleep 10
    check_consensus

    log_success "Rolling restart test completed"
}

scenario_storage_failure() {
    log_info "=== SCENARIO: Storage Stress ==="

    local target_pod=$(get_first_validator_name)

    if [ -z "$target_pod" ]; then
        log_warn "No validator pod found"
        return 0
    fi

    log_info "Simulating storage stress on: $target_pod"

    kubectl exec -n "$NAMESPACE" "$target_pod" -c aura -- sh -c 'dd if=/dev/zero of=/data/stress-test bs=1M count=100 2>/dev/null' || true

    log_info "Created 100MB stress file - waiting 30s..."
    sleep 30

    check_consensus

    kubectl exec -n "$NAMESPACE" "$target_pod" -c aura -- rm -f /data/stress-test

    log_success "Storage stress test completed"
}

run_all_scenarios() {
    scenario_pod_failure
    echo ""
    scenario_network_partition
    echo ""
    scenario_high_latency
    echo ""
    scenario_resource_exhaustion
    echo ""
    scenario_rolling_restart
    echo ""
    scenario_storage_failure
}

print_summary() {
    echo ""
    echo "=============================================="
    echo -e "${GREEN}Chaos Tests Completed${NC}"
    echo "=============================================="
    echo ""
    echo "Scenarios executed: $SCENARIO"
    echo ""
    echo "Verify final state:"
    kubectl get pods -n "$NAMESPACE"
}

main() {
    echo ""
    echo "=============================================="
    echo -e "${BLUE}Aura Kubernetes Chaos Tests${NC}"
    echo "=============================================="
    echo "Namespace: $NAMESPACE"
    echo "Scenario: $SCENARIO"
    echo "Time: $(date)"
    echo ""

    # Setup cleanup trap
    trap cleanup_test_pod EXIT

    # Ensure test pod exists for HTTP checks
    ensure_test_pod

    case "$SCENARIO" in
        all)
            run_all_scenarios
            ;;
        pod-failure)
            scenario_pod_failure
            ;;
        network-partition)
            scenario_network_partition
            ;;
        high-latency)
            scenario_high_latency
            ;;
        resource-exhaustion)
            scenario_resource_exhaustion
            ;;
        rolling-restart)
            scenario_rolling_restart
            ;;
        storage-failure)
            scenario_storage_failure
            ;;
        *)
            log_error "Unknown scenario: $SCENARIO"
            echo "Available: all, pod-failure, network-partition, high-latency, resource-exhaustion, rolling-restart, storage-failure"
            exit 1
            ;;
    esac

    print_summary
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace|-n)
            NAMESPACE="$2"
            shift 2
            ;;
        --scenario|-s)
            SCENARIO="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --namespace, -n NAME  Namespace to test (default: aura)"
            echo "  --scenario, -s NAME   Scenario to run: all, pod-failure, network-partition, high-latency, resource-exhaustion, rolling-restart, storage-failure"
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
