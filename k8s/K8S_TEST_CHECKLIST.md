# Aura Kubernetes Test Checklist

Comprehensive test suite consolidated from PAW and XAI projects. Run all tests before cloud deployment.

## Prerequisites

- [ ] KinD cluster running (create with `k8s/kind/dev-cluster.yaml`)
- [ ] Namespace `aura` exists
- [ ] Validators deployed and running
- [ ] Network policies applied

---

## 1. Smoke Tests (Quick Validation)

Basic deployment validation - run first.

| # | Test | Command/Check | Expected |
|---|------|---------------|----------|
| 1.1 | Namespace exists | `kubectl get ns aura` | Exists |
| 1.2 | Pods running | `kubectl get pods -n aura` | All Running/Completed |
| 1.3 | Validators ready | `kubectl get sts aura-validator -n aura` | All replicas ready |
| 1.4 | Services have endpoints | `kubectl get endpoints -n aura` | All have IPs |
| 1.5 | PVCs bound | `kubectl get pvc -n aura` | All Bound |
| 1.6 | Secrets exist | `kubectl get secrets -n aura` | validator-keys, genesis |
| 1.7 | Network policies | `kubectl get networkpolicies -n aura` | >=1 policy |
| 1.8 | Resource quotas | `kubectl get resourcequota -n aura` | Quotas defined |
| 1.9 | RPC connectivity | `curl localhost:26657/status` (in pod) | HTTP 200 |
| 1.10 | Health endpoint | `curl localhost:26657/health` (in pod) | HTTP 200 |
| 1.11 | Block production | Check `latest_block_height` advancing | Height > 0 |

---

## 2. Integration Tests (E2E Functionality)

Full functionality validation.

| # | Test | Description | Expected |
|---|------|-------------|----------|
| 2.1 | Transaction submission | Create test account with `aurad keys add` | Account created |
| 2.2 | Block finality | Monitor height increases over 10s | Height advances |
| 2.3 | RPC API | `GET /status` | HTTP 200 |
| 2.4 | REST API | `GET /cosmos/base/tendermint/v1beta1/node_info` | HTTP 200 |
| 2.5 | gRPC API | `grpcurl localhost:9090 list` | Services listed |
| 2.6 | Consensus participation | `GET /validators` | Validators > 0 |
| 2.7 | Peer connectivity | `GET /net_info` | n_peers > 0 |
| 2.8 | Metrics export | `GET localhost:26660/metrics` | Prometheus format |
| 2.9 | Service discovery | `nslookup aura-validator-headless.aura.svc.cluster.local` | Resolves |
| 2.10 | Persistent storage | Verify PVC data persists across restarts | Data intact |

---

## 3. Chaos Engineering Tests (Resilience)

Test fault tolerance and recovery.

| # | Scenario | Procedure | Expected Recovery |
|---|----------|-----------|-------------------|
| 3.1 | Pod failure | `kubectl delete pod aura-validator-0 -n aura --force` | New pod starts, rejoins consensus |
| 3.2 | Network partition | Apply isolation NetworkPolicy to one pod | 2/3 validators maintain consensus |
| 3.3 | High latency | Inject 200ms delay via `tc qdisc` | Slower but stable consensus |
| 3.4 | Memory pressure | Run `stress-ng --vm 1 --vm-bytes 512M` | Pod OOMKilled or recovers |
| 3.5 | Rolling restart | `kubectl rollout restart sts/aura-validator` | Zero-downtime restart |
| 3.6 | Storage stress | Write 100MB file to `/data` | Chain continues |
| 3.7 | DNS failure | Block DNS egress temporarily | Graceful degradation |
| 3.8 | Leader election | Kill leader pod | New leader elected |

---

## 4. Security Tests (Hardening Validation)

Security posture verification.

| # | Test | Check | Expected |
|---|------|-------|----------|
| 4.1 | Pod security context | `runAsNonRoot`, no root UID | Non-root enforced |
| 4.2 | Privilege escalation | `allowPrivilegeEscalation: false` | Blocked |
| 4.3 | Capabilities | No added capabilities | All dropped |
| 4.4 | Network policies | Default deny policy exists | Egress blocked by default |
| 4.5 | RBAC | Custom service accounts used | Not using `default` SA |
| 4.6 | Secret management | Secrets mounted as volumes (not env) | Volume mounts |
| 4.7 | External secrets | Vault integration configured | ExternalSecret CRDs |
| 4.8 | Resource limits | CPU/memory limits on all containers | Limits set |
| 4.9 | PDB | Pod Disruption Budget exists | minAvailable >= 2 |
| 4.10 | Image tags | No `:latest` tags | Explicit versions |
| 4.11 | Service mesh | Linkerd injection enabled | mTLS active |
| 4.12 | Audit logging | API server audit enabled | Logs captured |
| 4.13 | Sensitive data | No secrets in ConfigMaps | Clean ConfigMaps |
| 4.14 | Read-only filesystem | `readOnlyRootFilesystem: true` | Enabled where possible |

---

## 5. Network Policy Tests (Isolation)

Verify network segmentation.

| # | Test | Procedure | Expected |
|---|------|-----------|----------|
| 5.1 | DNS resolution | nslookup from pod | Works |
| 5.2 | External blocked | `curl google.com` from pod | Blocked/timeout |
| 5.3 | Monitoring access | Access Prometheus from pod | Allowed |
| 5.4 | Intra-namespace | Pod-to-pod on RPC port | Allowed |
| 5.5 | Unauthorized egress | `curl 8.8.8.8:80` | Blocked |
| 5.6 | P2P egress | Port 26656 to validators | Allowed |
| 5.7 | Vault access | Access Vault from pod | Allowed |
| 5.8 | Policy count | At least 3 NetworkPolicies | >= 3 |

---

## 6. Blockchain-Specific Tests

Cosmos SDK / Tendermint specific validation.

| # | Test | Description | Expected |
|---|------|-------------|----------|
| 6.1 | Validator key rotation | Hot-swap validator keys via Secret | Rolling restart, no slashing |
| 6.2 | Slashing detection | Check logs for double-sign patterns | None detected |
| 6.3 | Downtime detection | Check for missed blocks in logs | None detected |
| 6.4 | Byzantine detection | Check for malicious behavior logs | None detected |
| 6.5 | Finality verification | Blocks produced every 1-3s | Consistent timing |
| 6.6 | Consensus messages | prevote/precommit in logs | Present |
| 6.7 | Partition detection | Check for split-brain logs | None detected |
| 6.8 | Genesis validation | Genesis file matches expected | Correct chain-id |

---

## 7. Rate Limiting & DDoS Tests

Load testing and rate limit validation.

| # | Test | Procedure | Expected |
|---|------|-----------|----------|
| 7.1 | RPC rate limit | 100 requests/second to RPC | ~10 succeed, rest 429 |
| 7.2 | Connection limit | 50 concurrent connections | ~5 allowed, rest rejected |
| 7.3 | Load test | k6 or ab against ingress | Limits enforced |
| 7.4 | Burst handling | 20 requests in 1 second | Burst multiplier applied |

---

## 8. Observability Tests

Monitoring and alerting validation.

| # | Test | Check | Expected |
|---|------|-------|----------|
| 8.1 | Prometheus scraping | Targets in Prometheus | All UP |
| 8.2 | Grafana dashboards | Dashboards load | Data displayed |
| 8.3 | AlertManager | Test alert fires | Alert received |
| 8.4 | Log aggregation | Logs in Loki/ELK | Searchable |
| 8.5 | Tracing | Jaeger traces visible | Spans captured |

---

## Missing Infrastructure in Aura

Features present in PAW/XAI but missing in Aura:

### Must Add
- [ ] `k8s/kind/dev-cluster.yaml` - KinD cluster config
- [ ] `k8s/tests/smoke-tests.sh` - Smoke test suite
- [ ] `k8s/tests/integration-tests.sh` - Integration test suite
- [ ] `k8s/tests/chaos-tests.sh` - Chaos test suite
- [ ] `k8s/tests/security-tests.sh` - Security test suite
- [ ] `k8s/network-policies/test-network-policies.sh` - Network policy tests
- [ ] `k8s/testnet-deploy/` - Quick testnet setup
- [ ] `scripts/k8s-blockchain-test-suite.sh` - Blockchain-specific tests
- [ ] `scripts/k8s-validator-rotation.sh` - Key rotation automation
- [ ] `scripts/k8s-slashing-detection.sh` - Slashing detection
- [ ] `scripts/k8s-finality-check.sh` - Finality verification

### Nice to Have
- [ ] Helm chart (`helm/aura/`)
- [ ] Rate limit test deployment
- [ ] MEV protection network policy
- [ ] Multiple environment overlays (local, staging, prod)

---

## Running Tests

```bash
# After creating infrastructure, run in order:
./k8s/tests/smoke-tests.sh -n aura
./k8s/tests/integration-tests.sh -n aura
./k8s/tests/security-tests.sh -n aura
./k8s/network-policies/test-network-policies.sh -n aura
./k8s/tests/chaos-tests.sh -n aura -s all  # Run last (destructive)
./scripts/k8s-blockchain-test-suite.sh
```

---

## Test Results Tracking

| Date | Tester | Smoke | Integration | Security | Chaos | Blockchain | Notes |
|------|--------|-------|-------------|----------|-------|------------|-------|
| | | | | | | | |
