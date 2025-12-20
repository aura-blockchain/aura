#!/bin/bash
# smoke-tests.sh - Quick validation of Aura Kubernetes deployment
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="${NAMESPACE:-aura}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0
WARNINGS=0
TEST_POD=""

log_test() { echo -e "${BLUE}[TEST]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; PASSED=$((PASSED+1)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; FAILED=$((FAILED+1)); }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; WARNINGS=$((WARNINGS+1)); }

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
    log_test "Creating test pod for HTTP checks..."
    kubectl run smoke-test-curl -n "$NAMESPACE" --image=curlimages/curl:latest \
        --restart=Never --command -- sleep 3600 &>/dev/null || true

    # Wait for pod to be ready
    for i in {1..30}; do
        local status=$(kubectl get pod smoke-test-curl -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null || echo "")
        if [ "$status" = "Running" ]; then
            TEST_POD="smoke-test-curl"
            return 0
        fi
        sleep 1
    done

    log_warn "Could not create test pod for HTTP checks"
    return 1
}

# HTTP request via test pod
http_request() {
    local url="$1"
    local output_mode="${2:-status}"  # "status" for HTTP code, "body" for response body

    if [ -z "$TEST_POD" ]; then
        echo "000"
        return 1
    fi

    if [ "$output_mode" = "status" ]; then
        kubectl exec -n "$NAMESPACE" "$TEST_POD" -- curl -s -o /dev/null -w '%{http_code}' "$url" 2>/dev/null || echo "000"
    else
        kubectl exec -n "$NAMESPACE" "$TEST_POD" -- curl -s "$url" 2>/dev/null || echo ""
    fi
}

# Cleanup test pod on exit
cleanup_test_pod() {
    if [ "$TEST_POD" = "smoke-test-curl" ]; then
        kubectl delete pod smoke-test-curl -n "$NAMESPACE" --grace-period=0 --force &>/dev/null || true
    fi
}

test_namespace_exists() {
    log_test "Checking namespace exists..."

    if kubectl get namespace "$NAMESPACE" &>/dev/null; then
        log_pass "Namespace $NAMESPACE exists"
    else
        log_fail "Namespace $NAMESPACE does not exist"
        return 1
    fi
}

test_pods_running() {
    log_test "Checking pods are running..."

    local pod_count=$(kubectl get pods -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
    if [ "$pod_count" -eq 0 ]; then
        log_fail "No pods found in namespace $NAMESPACE"
        return 1
    fi

    local not_running
    not_running=$(kubectl get pods -n "$NAMESPACE" --no-headers 2>/dev/null | grep -cv "Running\|Completed" || true)
    if [ "$not_running" -gt 0 ]; then
        log_warn "$not_running pods not in Running state"
        kubectl get pods -n "$NAMESPACE" --no-headers | grep -v "Running\|Completed" || true
    else
        log_pass "All $pod_count pods are running"
    fi
}

test_validators_ready() {
    log_test "Checking validators are ready..."

    local ready=$(kubectl get statefulset aura-validator -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
    local desired=$(kubectl get statefulset aura-validator -n "$NAMESPACE" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")

    if [ "$ready" = "$desired" ] && [ "$desired" != "0" ]; then
        log_pass "All $ready/$desired validators are ready"
    else
        log_fail "Only $ready/$desired validators are ready"
        return 1
    fi
}

test_services_have_endpoints() {
    log_test "Checking services have endpoints..."

    local services=$(kubectl get svc -n "$NAMESPACE" -o name 2>/dev/null)
    local failed=0

    for svc in $services; do
        local svc_name=$(echo "$svc" | sed 's/service\///')
        local endpoints=$(kubectl get endpoints "$svc_name" -n "$NAMESPACE" -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null || true)

        if [ -z "$endpoints" ]; then
            log_warn "Service $svc_name has no endpoints"
            failed=$((failed+1))
        fi
    done

    if [ "$failed" -eq 0 ]; then
        log_pass "All services have endpoints"
    else
        log_warn "$failed services have no endpoints"
    fi
}

test_rpc_connectivity() {
    log_test "Checking RPC connectivity..."

    if [ -z "$TEST_POD" ]; then
        log_warn "No test pod available for RPC connectivity test"
        return 0
    fi

    # Use headless service DNS to reach validator-0
    local status=$(http_request "http://aura-validator-0.aura-validator-headless.${NAMESPACE}.svc.cluster.local:26657/status" "status")

    if [ "$status" = "200" ]; then
        log_pass "RPC endpoint is responsive (HTTP $status)"
    else
        log_fail "RPC endpoint not responsive (HTTP $status)"
        return 1
    fi
}

test_block_production() {
    log_test "Checking block production..."

    if [ -z "$TEST_POD" ]; then
        log_warn "No test pod available for block production test"
        return 0
    fi

    local response=$(http_request "http://aura-validator-0.aura-validator-headless.${NAMESPACE}.svc.cluster.local:26657/status" "body")
    local height=$(echo "$response" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0")

    if [ "$height" != "0" ] && [ "$height" != "null" ] && [ -n "$height" ]; then
        log_pass "Chain is producing blocks (height: $height)"
    else
        log_warn "Cannot verify block production (height: $height)"
    fi
}

test_health_endpoint() {
    log_test "Checking health endpoints..."

    if [ -z "$TEST_POD" ]; then
        log_warn "No test pod available for health endpoint test"
        return 0
    fi

    local health=$(http_request "http://aura-validator-0.aura-validator-headless.${NAMESPACE}.svc.cluster.local:26657/health" "status")

    if [ "$health" = "200" ]; then
        log_pass "Health endpoint is responsive"
    else
        log_fail "Health endpoint not responsive (HTTP $health)"
        return 1
    fi
}

test_pvcs_bound() {
    log_test "Checking PVCs are bound..."

    local total=$(kubectl get pvc -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
    local bound=$(kubectl get pvc -n "$NAMESPACE" --no-headers 2>/dev/null | grep "Bound" | wc -l)

    if [ "$total" -eq 0 ]; then
        log_warn "No PVCs found"
        return 0
    fi

    if [ "$bound" -eq "$total" ]; then
        log_pass "All $bound/$total PVCs are bound"
    else
        log_fail "Only $bound/$total PVCs are bound"
        kubectl get pvc -n "$NAMESPACE" --no-headers | grep -v "Bound"
        return 1
    fi
}

test_secrets_exist() {
    log_test "Checking secrets exist..."

    local expected_secrets=("aura-validator-keys" "aura-genesis" "aura-api-secrets")
    local missing=0

    for secret in "${expected_secrets[@]}"; do
        if ! kubectl get secret "$secret" -n "$NAMESPACE" &>/dev/null; then
            log_warn "Secret $secret not found"
            missing=$((missing+1))
        fi
    done

    if [ "$missing" -eq 0 ]; then
        log_pass "All expected secrets exist"
    else
        log_warn "$missing secrets are missing (may be using configmaps or external secrets)"
    fi
}

test_network_policies() {
    log_test "Checking network policies..."

    local policy_count=$(kubectl get networkpolicies -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)

    if [ "$policy_count" -gt 0 ]; then
        log_pass "$policy_count network policies configured"
    else
        log_warn "No network policies found"
    fi
}

test_resource_quotas() {
    log_test "Checking resource quotas..."

    local quota=$(kubectl get resourcequota -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)

    if [ "$quota" -gt 0 ]; then
        log_pass "Resource quotas configured"
    else
        log_warn "No resource quotas found"
    fi
}

print_summary() {
    echo ""
    echo "=============================================="
    echo "SMOKE TEST RESULTS"
    echo "=============================================="
    echo -e "Passed:   ${GREEN}$PASSED${NC}"
    echo -e "Failed:   ${RED}$FAILED${NC}"
    echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
    echo "=============================================="

    if [ "$FAILED" -gt 0 ]; then
        echo -e "${RED}SMOKE TESTS FAILED${NC}"
        exit 1
    elif [ "$WARNINGS" -gt 0 ]; then
        echo -e "${YELLOW}SMOKE TESTS PASSED WITH WARNINGS${NC}"
        exit 0
    else
        echo -e "${GREEN}ALL SMOKE TESTS PASSED${NC}"
        exit 0
    fi
}

main() {
    echo ""
    echo "=============================================="
    echo -e "${BLUE}Aura Kubernetes Smoke Tests${NC}"
    echo "=============================================="
    echo "Namespace: $NAMESPACE"
    echo "Time: $(date)"
    echo ""

    # Setup cleanup trap
    trap cleanup_test_pod EXIT

    test_namespace_exists
    test_pods_running
    test_validators_ready
    test_services_have_endpoints
    test_pvcs_bound
    test_secrets_exist
    test_network_policies
    test_resource_quotas

    # Ensure test pod exists for HTTP-based tests
    ensure_test_pod

    test_rpc_connectivity
    test_health_endpoint
    test_block_production

    print_summary
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --namespace|-n)
            NAMESPACE="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --namespace, -n NAME  Namespace to test (default: aura)"
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
