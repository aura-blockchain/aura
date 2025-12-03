# AURA Production Roadmap

**Status:** 85% Complete | **Chain:** Cosmos SDK 0.53.4 + CometBFT | **Build:** ✅ Passing | **Local Testnet:** ✅ Running (Block 4900+)

---

## Existing Components (DO NOT DUPLICATE)

### Core Chain (`/chain/`)
- ✅ App wiring: `/chain/app/app.go`
- ✅ CLI daemon: `/chain/cmd/aurad/`
- ✅ Makefile: `/chain/Makefile`, `/chain/Makefile.security`

### Custom Modules (27 total in `/chain/x/`)
All modules have keepers, protos, and tests:
- **Identity:** identity, vcregistry, identitychange, inclusionroutines, confidencescore
- **Privacy:** privacy, cryptography, networksecurity, validatorsecurity, walletsecurity, incidentresponse, security
- **Economics:** economics, economicsecurity, governance, dex
- **Infrastructure:** bridge, dataregistry, monitoring, prevalidation, compliance
- **AI/Contracts:** aiassistant, wasm, aura-bindings, contractregistry, auth

### Smart Contracts (`/contracts/`)
- ✅ aura-bindings: `/contracts/packages/aura-bindings/`
- ✅ binding-tester: `/contracts/binding-tester/`
- ✅ vc-issuer: `/contracts/vc-issuer/`
- ⚠️ Need deployment testing and on-chain instantiation

### SDKs (`/sdk/`)
- ✅ Go: `/sdk/go/`
- ✅ JavaScript: `/sdk/javascript/`
- ✅ Python: `/sdk/python/`

### Wallets (`/wallet/`)
- ✅ Desktop: `/wallet/desktop/`
- ✅ Mobile: `/wallet/mobile/`
- ✅ Browser Extension: `/wallet/browser-extension/`
- ✅ Web: `/wallet/web/`

### Infrastructure
- ✅ Docker: `/docker-compose.yml`, `/docker-compose.secure.yml`, `/docker-compose.testnet.yml`
- ✅ Kubernetes: `/k8s/base/`, `/k8s/overlays/` (dev, staging, production)
- ✅ Monitoring: `/prometheus/`, `/grafana/dashboards/`
- ✅ Deployment scripts: `/deployment-security/scripts/`
- ✅ Testnet scripts: `/scripts/testnet-init.sh`, `/scripts/testnet-manage.sh`

### Testing (376+ test files in `/chain/testing/`)
- ✅ Unit, integration, e2e, chaos, benchmark tests

### Documentation (`/docs/`)
- ✅ RFCs: `/docs/rfcs/`
- ✅ Architecture: `/docs/architecture/`
- ✅ Economics: `/docs/economics/`
- ✅ Compliance: `/docs/compliance/`
- ✅ Runbooks: `/docs/runbooks/`

---

## Critical Gaps

| Gap | Priority | Location |
|-----|----------|----------|
| Production genesis file | ✅ Done | `/networks/mainnet/genesis.json` |
| External security audit | 🔴 Critical | Not started |
| Active testnet | 🟢 Local running | `aura-local-1` producing blocks |
| IBC channels | 🔴 Critical | Not established |
| Block explorer | 🟡 High | Needs deployment |
| Faucet service | 🟡 High | Needs deployment |
| Public RPC/API | 🟡 High | Not provisioned |

---

## Phase 0: Pre-Deployment (2-3 weeks)

### Genesis Configuration
- [x] Create production genesis with all 27 modules → `/networks/mainnet/genesis.json`
- [x] Configure: 100 initial validators, 21-day unbonding, 0.025uaura min gas
- [x] Load inclusion routines from `/data/inclusion_routines/ir_genesis_300.json`
- [ ] Create genesis accounts per `/docs/economics/founder-wallets.md` (pending ops approval)
- [x] Validate: `aurad start` - genesis loads successfully, all 27 modules initialize

### Smart Contracts
- [x] Install Rust toolchain: `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh`
- [x] Compile contracts: `cargo build --release --target wasm32-unknown-unknown`
- [x] WASM output: `binding_tester.wasm` (227KB), `vc_issuer.wasm` (328KB)
- [x] Optimize: `make optimize-wasm` → `/contracts/artifacts/`
- [ ] Test deployment on local testnet
- [x] Create deployment scripts: `/scripts/deploy-contracts.sh`

### Security
- [x] HSM integration guide → `/docs/security/HSM_INTEGRATION.md`
- [x] Run secret management scripts: `/deployment-security/scripts/generate-secrets.sh`
- [x] TLS setup: `/deployment-security/scripts/tls-setup.sh` (self-signed for dev)
- [x] Review node configs: `/networks/mainnet/config.toml` - added seed TODOs, double_sign_check

### Testing
- [x] Run full test suite: `make test` (59/101 packages passing - mock infrastructure needs fixes)
- [x] Run chaos tests: `/chain/testing/chaos/` (passing)
- [x] Run benchmarks: `/chain/testing/benchmark/` (passing)
- [x] Fix test mock infrastructure (ante_test.go, integration tests) - tests now compile
- [ ] Target: >80% coverage

### Documentation
- [x] Create: `/docs/ops/PRODUCTION_DEPLOYMENT.md`
- [x] Create: `/docs/validators/ONBOARDING.md`
- [x] Create: `/docs/ops/runbooks/` (upgrade, incident, backup procedures)

---

## Phase 1: Local Testnet (1-2 weeks)

### Single Node
- [x] `cd /chain && ./aurad init local-validator --chain-id aura-local-1`
- [x] Configure genesis with single validator (900,000 voting power)
- [x] Start: `aurad start --home ~/.aura --grpc.enable=false`
- [x] Verify: RPC `localhost:26657` ✅, API `localhost:1317` ✅
- [x] Node producing blocks (confirmed at height 19+)

### Module Testing
- [x] Identity/VC: Query commands functional (vcregistry, identitychange)
- [x] Inclusion Routines: Query commands functional
- [x] Governance: Query commands functional
- [x] Bridge: Simulate cross-chain transfer, test Merkle proofs (`chain/x/bridge/keeper/cross_chain_flow_test.go` covers lock flow + Merkle verification)
- [ ] DEX: Create pool, execute swaps, verify AMM calculations
- [ ] Compliance: Configure AML rules, test transaction screening

### Smart Contracts
- [x] WASM CLI commands implemented: `aurad tx aura_wasm_security [store|instantiate|execute|migrate]`
- [ ] [BLOCKED] Deploy vc-issuer: `aurad tx aura_wasm_security store contracts/artifacts/vc_issuer.wasm`
  - **Blocker:** tx signing currently fails with `signature verification failed; ... SIGN_MODE_DIRECT ... unauthorized` even when `--sign-mode legacy-amino-json` is passed. Keyring account/sequence and CLI tx plumbing need fixing.
  - **Automation ready:** `scripts/test-vc-issuer-e2e.sh` spins up an ephemeral node, stores, instantiates, registers issuer, requests, fulfills, and queries. Broadcast now uses `sync`, gas bumped to 5,000,000, and sign-mode flag set, but the CLI still produces SIGN_MODE_DIRECT; TxConfig now defaults to LEGACY_AMINO_JSON and the Tx/Query CLI has been re-wired to the SDK—re-run after keyring/sequence fixes.
- [ ] [BLOCKED] Instantiate and test execute/query
  - **Blocker:** same signing/keyring issue; resolve store path first.
- [ ] Restore full genesis CLI (add-genesis-account/gentx/collect-gentxs) or ship a scripted genesis injector for local dev/testnet parity
- [ ] Benchmark gas consumption (store/instantiate/execute) once signing unblocked

### Multi-Node (4 validators)
- [x] Deploy with Docker Compose → `/docker-compose.testnet.yml`
- [x] Create initialization script → `/scripts/testnet-init.sh`
- [x] Create management script → `/scripts/testnet-manage.sh`
- [x] Create Prometheus config → `/prometheus/prometheus-testnet.yml`
- [x] Create testnet documentation → `/TESTNET_SETUP.md`, `/TESTNET_QUICKSTART.md`
- [x] Document Docker runbook for local testnet → `/docs/runbooks/LOCAL_TESTNET_DOCKER.md`
- [ ] Run initialization and start testnet
- [ ] Test Byzantine fault tolerance (stop 1 validator = consensus continues)
- [ ] Test state synchronization

### Monitoring
- [x] Deploy: `docker-compose -f docker-compose.monitoring.yml up -d` (config ready)
- [x] Import dashboards from `/grafana/dashboards/` → `/docker/monitoring/grafana/dashboards/`
- [x] Configure alerts for chain halt, high resource usage, low peer count → `/docker/monitoring/prometheus/rules/aura-alerts.yml`

### Module Architecture Alignment (Cosmos SDK + Native Parity)
- [x] Add lightweight `IsAppModule` tags/adapters for many Aura modules; scaffold adapter wrapper (`chain/app/module_adapters.go`).
- [x] Add adapters for remaining Aura modules and wire all into an SDK `module.Manager`.
- [x] Build SDK `module.Manager` with core modules (auth, bank, staking, slashing, distribution, params, consensus, genutil, wasm) and Aura adapters; set InitGenesis/BeginBlock/EndBlock orderings; register services via configurator.
- [x] Fix adapter wiring to execute module `InitGenesis`/`BeginBlock`/`EndBlock` with `codec.JSONCodec` and add regression tests; migrate `x/walletsecurity` to a full `AppModule`/`AppModuleBasic` with genesis import/export.
- [x] Hook BaseApp pre-block, InitChainer, BeginBlocker, EndBlocker, prepare-check-state, and precommit to the module manager with JSON app state handling.
- [x] Make `start` honor chain-id from genesis and app.toml/flags for RPC/API/GRPC ports; expose flags to override gRPC/API ports for e2e harnesses.
- [ ] Run `scripts/test-vc-issuer-e2e.sh` after refactor to confirm tx execution works; current failure is SIGN_MODE_DIRECT auth rejection in wasm store tx despite amino flag. Instrument BaseApp/store version logging if signing fix does not unblock tx execution.
- [x] Align DEX invariants, msg server guardrails, keeper harness, and positive msg server paths with current proto types; `go test ./chain/x/dex/...` now passes.
- [x] Seeded KV stores at InitGenesis to ensure all mounted stores persist version 1 (prevents IAVL version lookup failures in tx/query paths).
- [ ] Update DEX/Compliance tests and coverage after tx path is restored.
- [ ] **Module Migration to Unified AppModule Pattern**
  - Converted governance, compliance, identitychange, dataregistry, inclusionroutines, and aiassistant to native `AppModule`/`AppModuleBasic` implementations; each now registers its proto Msg/MsgResponse interfaces, gRPC services via `module.Configurator`, and default genesis/validation logic.
  - Next up: finish migrating the remaining Aura modules (bridge, dex, economicsecurity, vcregistry, cryptography, wasm/security, monitoring, etc.), then delete the adapter layer so the ModuleManager wires real `AppModule`s, and rerun `go test ./chain/...` to clear the outstanding suite failures before finishing the migration.

### Next 20 Tasks (handoff for next agent)
1. Fix wasm tx signing path end-to-end: configure TxConfig to enable LEGACY_AMINO_JSON as default, ensure CLI uses it, and set correct chain-id/account-number/sequence so wasm store succeeds.
2. Regenerate or align keyring keys with genesis; add deterministic seeds and explicit keyring-backend in `scripts/test-vc-issuer-e2e.sh` with automated account/sequence fetch.
3. Implement real CLI query/tx subcommands (bank, tx, account, block) using SDK client contexts (not stubs) and add regression tests.
4. Re-run `scripts/test-vc-issuer-e2e.sh` after signing fix; capture logs, keep temp home path, and baseline gas per step (store/instantiate/execute).
5. Add offline sign fallback (`--generate-only` + `tx sign --sign-mode LEGACY_AMINO_JSON`) in the contract flow to avoid SIGN_MODE_DIRECT regressions.
6. Harden e2e script preflight: assert account-number/sequence before sending, check increments afterward, and fail fast on mismatches.
7. Add BaseApp diagnostics: log store name/height on `CacheMultiStoreWithVersion` failures and emit a post-InitGenesis sanity check that every mounted KV store has a persisted version.
8. Bake store-init marker verification into multi-node bootstrap/testnet scripts and document the behavior; confirm AppHash consistency across start→stop→start.
9. Finish migrating remaining Aura modules (bridge, dex, economicsecurity, vcregistry, cryptography, wasm/security, monitoring, aurabindings) to native `AppModule`s and delete the adapter layer.
10. Rerun `go test ./chain/...` after migrations; raise coverage toward 80%+ and fix any failing suites.
11. ~~Add wasm/security invariants and tests (code size/upload limits, admin enforcement, gas caps, event emission).~~ ✅ **COMPLETE** - Added 4 new invariants (CodeSizeLimits, UploadAuthEnforcement, GasCaps, AdminEnforcement) with 10 comprehensive security tests. All security tests passing.
12. Expand DEX/compliance scenario tests: pool create/swap math, fee/gas guardrails, AML screening edge cases.
13. Benchmark wasm store/instantiate/execute gas on local node and tune app.toml defaults; record baselines in docs.
14. Restore full genesis CLI helpers (`add-genesis-account`, `gentx`, `collect-gentxs`) wired to Aura modules; add tests.
15. Harden key management: deterministic key material, backend selection, env/secret handling, and per-script validation.
16. Add monitoring probes and alerts for wasm tx failures, state load errors, and signature mismatch rates; surface in Grafana/Prometheus.
17. Integrate explorer/faucet deployment steps into Phase 1 docs and ensure API/RPC endpoints are exposed for local testnet users.
18. Prepare cloud testnet automation (parameterized K8s overlays, DNS entries, validator configs) and run a dry-run rollout.
19. Verify replay on a seeded DB via start→stop→start integration test; ensure AppHash stability and no “version does not exist” errors.
20. Document the signing/keyring remediation, store seeding guarantees, and new gas baselines in `/docs/runbooks/LOCAL_TESTNET_DOCKER.md` and contract deployment docs.

---

## Phase 2: Cloud Testnet (3-4 weeks)

### Infrastructure
- [ ] Provision cloud (GCP/AWS): 4+ validators across US, EU, Asia
- [ ] Deploy K8s cluster using `/k8s/overlays/staging/`
- [ ] Configure DNS: rpc.testnet.aura.network, api.testnet.aura.network
- [ ] Setup CDN/DDoS protection (Cloudflare)

### Genesis Coordination
- [ ] Create testnet genesis: chain-id `aura-testnet-1`
- [ ] Collect gentx from 10+ validators
- [ ] Coordinate genesis time, distribute final genesis.json

### Public Services
- [ ] Deploy faucet → faucet.testnet.aura.network
- [ ] Deploy block explorer → explorer.testnet.aura.network
- [ ] Customize explorer for AURA modules (VC Registry, Inclusion Routines, AI Assistant)

### IBC Integration
- [ ] Deploy Hermes relayer
- [ ] Establish channel to Cosmos Hub testnet
- [ ] Test cross-chain token transfers

### Community
- [ ] Publish testnet docs: `/docs/testnet/`
- [ ] Launch bug bounty: `/docs/testing/BUG_BOUNTY_PROGRAM.md`
- [ ] Recruit testnet validators (target: 50+)

---

## Phase 3: Security Audit (6-8 weeks)

### Internal Review
- [ ] Audit consensus layer: `/chain/app/app.go`, BeginBlocker/EndBlocker hooks
- [ ] Audit all 27 module keepers for reentrancy, overflow, access control
- [ ] Audit smart contracts: vc-issuer, aura-bindings
- [ ] Audit cryptography: key generation, ZK proofs, signature verification
- [ ] Audit P2P layer: networksecurity module

### External Audit
- [ ] Select firm: Certik, Trail of Bits, Halborn, or Quantstamp
- [ ] Budget: $100,000-$200,000
- [ ] Scope: consensus, modules, contracts, IBC, P2P
- [ ] Remediate all critical/high findings
- [ ] Publish report → `/docs/security/AUDIT_REPORT_2025.pdf`

### Penetration Testing
- [ ] Test DDoS resilience, API exploits, consensus manipulation
- [ ] Document findings and mitigations

### Compliance
- [ ] Legal review per `/docs/compliance/SECURITIES_LAW_ANALYSIS.md`
- [ ] Privacy review per `/docs/compliance/PRIVACY_POLICY.md`
- [ ] Finalize AML checklist: `/docs/ops/compliance/minimal-aml-checklist.md`

---

## Phase 4: Mainnet Launch (4-6 weeks)

### Genesis
- [ ] Finalize mainnet genesis: chain-id `aura-mainnet-1`
- [ ] Coordinate with 100+ validators
- [ ] Collect and validate all gentx
- [ ] Set production parameters (14-day voting, 40% quorum)
- [ ] Publish genesis with SHA256 checksum

### Infrastructure
- [ ] Deploy production K8s: `/k8s/overlays/production/`
- [ ] Multi-region deployment (5+ regions)
- [ ] Configure production DNS: rpc.aura.network, api.aura.network, grpc.aura.network
- [ ] DDoS protection, rate limiting

### Launch
- [ ] All validators start at genesis time
- [ ] Monitor first 1000 blocks
- [ ] 24/7 monitoring for first 72 hours

### Ecosystem
- [ ] Deploy mainnet smart contracts
- [ ] Launch explorer → explorer.aura.network
- [ ] Enable IBC to Cosmos Hub
- [ ] Integrate with Keplr, Leap, Cosmostation wallets
- [ ] Launch developer portal → docs.aura.network

---

## Phase 5: Post-Launch (Ongoing)

- [ ] Regular network upgrades via governance (quarterly)
- [ ] 24/7 monitoring: status.aura.network
- [ ] Quarterly security audits
- [ ] Secret rotation per `/deployment-security/scripts/rotate-secrets.sh`
- [ ] Expand AI assistant network
- [ ] Grow VC ecosystem partnerships
- [ ] Performance optimization (target: 1000+ TPS)
- [ ] Research layer 2 scaling solutions

---

## Timeline Summary

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 0: Pre-Deployment | 2-3 weeks | 3 weeks |
| Phase 1: Local Testnet | 1-2 weeks | 5 weeks |
| Phase 2: Cloud Testnet | 3-4 weeks | 9 weeks |
| Phase 3: Security Audit | 6-8 weeks | 17 weeks |
| Phase 4: Mainnet Launch | 4-6 weeks | 23 weeks |

**Total: ~4-6 months to mainnet**

---

## Quick Commands

```bash
# Build
cd /home/decri/blockchain-projects/aura/chain
go build -o aurad ./cmd/aurad

# Test
make test
make test-coverage

# Local node
./aurad init local-validator --chain-id aura-local-1
./aurad start

# Contracts
cd /home/decri/blockchain-projects/aura/contracts/vc-issuer
cargo wasm

# Monitoring
docker-compose -f docker-compose.monitoring.yml up -d
```
