#!/bin/bash
# Run all Aura K8s tests with proper kubeconfig setup
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
NAMESPACE="${NAMESPACE:-aura}"
KUBECONFIG="${KUBECONFIG:-/tmp/kind-aura.kubeconfig}"

export KUBECONFIG

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=============================================="
echo -e "${BLUE}Aura Kubernetes Test Suite${NC}"
echo "=============================================="
echo "Namespace: $NAMESPACE"
echo "Kubeconfig: $KUBECONFIG"
echo "Time: $(date)"
echo ""

# Verify cluster access
if ! kubectl get nodes &>/dev/null; then
    echo -e "${RED}Cannot connect to cluster. Check KUBECONFIG.${NC}"
    exit 1
fi

echo -e "${GREEN}Cluster connection verified${NC}"
echo ""

# Run tests in order
FAILED=0

echo -e "${BLUE}=== Running Smoke Tests ===${NC}"
if "$SCRIPT_DIR/smoke-tests.sh" -n "$NAMESPACE"; then
    echo -e "${GREEN}Smoke tests PASSED${NC}"
else
    echo -e "${RED}Smoke tests FAILED${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${BLUE}=== Running Integration Tests ===${NC}"
if "$SCRIPT_DIR/integration-tests.sh" -n "$NAMESPACE"; then
    echo -e "${GREEN}Integration tests PASSED${NC}"
else
    echo -e "${RED}Integration tests FAILED${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${BLUE}=== Running Security Tests ===${NC}"
if "$SCRIPT_DIR/security-tests.sh" -n "$NAMESPACE"; then
    echo -e "${GREEN}Security tests PASSED${NC}"
else
    echo -e "${RED}Security tests FAILED${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${BLUE}=== Running Network Policy Tests ===${NC}"
if "$SCRIPT_DIR/../network-policies/test-network-policies.sh" -n "$NAMESPACE"; then
    echo -e "${GREEN}Network policy tests PASSED${NC}"
else
    echo -e "${RED}Network policy tests FAILED${NC}"
    ((FAILED++))
fi
echo ""

echo -e "${BLUE}=== Running Blockchain Tests ===${NC}"
if "$PROJECT_ROOT/scripts/k8s-blockchain-test-suite.sh"; then
    echo -e "${GREEN}Blockchain tests PASSED${NC}"
else
    echo -e "${RED}Blockchain tests FAILED${NC}"
    ((FAILED++))
fi
echo ""

# Chaos tests are optional and destructive
if [[ "${RUN_CHAOS:-false}" == "true" ]]; then
    echo -e "${BLUE}=== Running Chaos Tests ===${NC}"
    if "$SCRIPT_DIR/chaos-tests.sh" -n "$NAMESPACE" -s all; then
        echo -e "${GREEN}Chaos tests PASSED${NC}"
    else
        echo -e "${RED}Chaos tests FAILED${NC}"
        ((FAILED++))
    fi
fi

echo ""
echo "=============================================="
echo "FINAL RESULTS"
echo "=============================================="
if [ "$FAILED" -eq 0 ]; then
    echo -e "${GREEN}ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "${RED}$FAILED test suite(s) FAILED${NC}"
    exit 1
fi
