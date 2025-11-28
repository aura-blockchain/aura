# AURA Production Roadmap

**Status:** 75% Complete | **Chain:** Cosmos SDK 0.53.4 + CometBFT | **Build:** ✅ Passing

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
- ✅ Docker: `/docker-compose.yml`, `/docker-compose.secure.yml`
- ✅ Kubernetes: `/k8s/base/`, `/k8s/overlays/` (dev, staging, production)
- ✅ Monitoring: `/prometheus/`, `/grafana/dashboards/`
- ✅ Deployment scripts: `/deployment-security/scripts/`

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
| Active testnet | 🔴 Critical | No deployment |
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
- [ ] Optimize: `make optimize-wasm` → `/contracts/artifacts/`
- [ ] Test deployment on local testnet
- [ ] Create deployment scripts: `/scripts/deploy-contracts.sh`

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
- [ ] `cd /chain && ./build/aurad init local-validator --chain-id aura-local-1`
- [ ] Configure genesis with single validator
- [ ] Start: `aurad start --home ~/.aura`
- [ ] Verify: RPC `localhost:26657`, API `localhost:1317`, metrics `localhost:26660`

### Module Testing
- [ ] Identity/VC: Create DID, issue credential, query vcregistry, verify confidence scores
- [ ] Inclusion Routines: Load definitions, submit proofs, verify PoI rewards
- [ ] Governance: Submit proposal, vote with ZK privacy, verify execution
- [ ] Bridge: Simulate cross-chain transfer, test Merkle proofs
- [ ] DEX: Create pool, execute swaps, verify AMM calculations
- [ ] Compliance: Configure AML rules, test transaction screening

### Smart Contracts
- [ ] Deploy vc-issuer: `aurad tx wasm store contracts/artifacts/vc_issuer.wasm`
- [ ] Instantiate and test execute/query
- [ ] Benchmark gas consumption

### Multi-Node (4 validators)
- [ ] Deploy with Docker Compose
- [ ] Test Byzantine fault tolerance (stop 1 validator = consensus continues)
- [ ] Test state synchronization

### Monitoring
- [ ] Deploy: `docker-compose -f docker-compose.monitoring.yml up -d`
- [ ] Import dashboards from `/grafana/dashboards/`
- [ ] Configure alerts for chain halt, high resource usage, low peer count

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
