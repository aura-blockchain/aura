# Aura Kubernetes Test Suite

Comprehensive testing framework for validating Aura k8s deployments before cloud migration.

## Quick Start

```bash
# Run all tests in sequence
./smoke-tests.sh -n aura
./integration-tests.sh -n aura
./security-tests.sh -n aura
../network-policies/test-network-policies.sh -n aura
./chaos-tests.sh -n aura -s all  # Run last (destructive)
```

## Test Suites

| Suite | Purpose | Duration |
|-------|---------|----------|
| smoke-tests.sh | Basic deployment validation | ~30s |
| integration-tests.sh | E2E functionality | ~2min |
| security-tests.sh | Security hardening checks | ~1min |
| chaos-tests.sh | Fault tolerance | ~10min |
| test-network-policies.sh | Network isolation | ~2min |

## Blockchain Tests

Additional blockchain-specific tests in `scripts/`:

```bash
../../scripts/k8s-blockchain-test-suite.sh  # All blockchain tests
../../scripts/k8s-slashing-detection.sh     # Slashing indicators
../../scripts/k8s-finality-check.sh         # Block finality
../../scripts/k8s-validator-rotation.sh     # Key rotation
```

## Options

All scripts support:
- `--namespace, -n NAME` - Target namespace (default: aura)
- `--help, -h` - Show help

Chaos tests additionally support:
- `--scenario, -s NAME` - Run specific scenario

## Exit Codes

- `0` - All tests passed
- `1` - One or more tests failed

## See Also

- `../K8S_TEST_CHECKLIST.md` - Full test checklist
- `../kind/dev-cluster.yaml` - Local cluster setup
