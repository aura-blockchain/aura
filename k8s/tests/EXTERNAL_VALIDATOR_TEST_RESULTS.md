# External Validator Test Results

**Date:** 2025-12-20
**Cluster:** bcpc-staging + wsl2-worker (multi-node K8s)

## Summary

All external validator tests now run successfully. Key fixes made to test scripts for Linkerd sidecar compatibility.

## Tests Executed

### 1. Network Policy P2P Egress Test ✓
- **Status:** PASS
- **External validator:** `external-validator-sim` deployed to `external-validators` namespace
- **Test:** Validated P2P connectivity from aura namespace to external validator on port 26656

### 2. Integration Tests ✓
- **All 9 tests PASS:**
  - Transaction submission (network info check)
  - Block finality (blocks 9142 → 9144)
  - RPC endpoint responsive
  - REST API responsive
  - 3 validators in consensus
  - 2 peers connected
  - Prometheus metrics exporting
  - Service discovery working
  - 4 PVCs bound

### 3. Chaos Tests ✓
- **Pod failure test:** Validator killed and recovered successfully
- **Rolling restart:** All 3 validators restarted, consensus maintained at height 9155
- **External validator on WSL2:** aura-validator-1 runs on `wsl2-worker` node

## Multi-Node Validator Distribution
```
aura-validator-0: bcpc-staging
aura-validator-1: wsl2-worker (external/WSL2)
aura-validator-2: bcpc-staging
```

## Fixes Applied
1. Added `-c rpc-probe` container spec for curl commands (aura container lacks curl)
2. Added `component=validator` label fallback for pod selection
3. Fixed label selectors across all test scripts

## Files Modified
- `k8s/network-policies/test-network-policies.sh`
- `k8s/tests/integration-tests.sh`
- `k8s/tests/chaos-tests.sh`
