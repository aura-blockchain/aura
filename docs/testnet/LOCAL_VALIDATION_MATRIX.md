# Local Validation Matrix

All functionality must be verified locally (homelab/colo hardware, Docker/K8s) before any cloud testing. Use this matrix to track readiness, follow test procedures, and record blockers. Update status after each run and log details in `docs/testnet/STATUS_LOG.md`.

## Legend
- **Status**: `❌` not started, `🟡` in progress, `✅` validated locally
- **Owner**: person/role responsible for current step (optional)
- **Latest automated snapshot:** `2025-01-18` — `go test ./...` (all module/unit suites passing locally). Each row still needs to be updated to reflect manual validations as they are performed.

## 1. Infrastructure & Core Services

| Component | Validation Steps (Local) | Status | Owner | Notes |
|-----------|--------------------------|--------|-------|-------|
| Validators (4x) | Use `scripts/testnet-init.sh` to seed volumes, start via `docker compose -f docker-compose.testnet.yml up -d validator-{1..4}`, confirm block production `curl localhost:26657/status` on each host. | ✅ | | Running via docker compose with height >7 verified (2025-01-18) |
| Sentries (2x) | Bring up sentry stack on separate hosts, ensure persistent peers configured in `config.toml`, verify inbound/outbound peers via `aurad status`. | ✅ | | Running via compose; `curl http://localhost:28659/net_info` + `28663` show 4 peers each, `catching_up=false` |
| Observer RPC/API nodes | Deploy observer compose stack with Nginx proxy, confirm REST/gRPC endpoints locally and via LAN clients. | ✅ | | `docker compose -f docker-compose.observer.yml up -d` + `docker compose -f docker-compose.proxy.yml up -d` — RPC served at `http://localhost:8080/rpc`, REST stub `/api`, gRPC `localhost:12091` |
| WireGuard/VLAN networking | Run `wg show` across racks, verify restricted P2P ports (26656) and management access. | ❌ | | |
| Monitoring (Prometheus/Grafana/Alertmanager) | `docker compose -f docker-compose.testnet.yml` brings up Prometheus/Grafana; import dashboards via `scripts/import_grafana_dashboards.sh`, verify scrape targets and dashboard availability. | ✅ | | Prometheus ready + Grafana dashboards imported; `testnet-monitor.sh quick/performance` run (2025-01-18) |
| Explorer + Indexer | Follow `explorer/DEPLOYMENT_GUIDE.md` locally, point to observer RPC, validate search/blocks/tx views. | ✅ | | Ping.pub React/Vite frontend built + `docker compose -f docker-compose.explorer.yml up -d`; `http://localhost:8088` confirmed pulling data from validator-1 |
| Faucet | Deploy via `faucet-service/docker-compose.yml`, use local RPC, confirm rate limiting + hCaptcha (dev bypass), send test tx. | ✅ | | Verified end-to-end: funded faucet wallet, gRPC balance checks, RPC broadcast via validator-1. Sample tx: `A6FE8C6599FFCFCFFA0FEFE5A196AD4E7580B4856CE7D7BA73BD200F7B9E724B` (height 2837) |
| Hermes Relayer | Configure `config/hermes/config.toml` for local endpoints, run `scripts/hermes-bootstrap.sh`, ensure client/connection/channel created. | ❌ | | Counterparty chain + relayer keys funded (tx `947F37E...` / `A4C93E...`); bootstrap now blocked because gRPC lacks `cosmos.tx.v1beta1.Service/Simulate`, so Hermes cannot estimate gas |
| Wallets (desktop/mobile/web/extension) | Connect to local RPC endpoints, perform send/receive, staking, gov vote flows per `wallet/README.md`. | ❌ | | |
| Analytics Dashboards | Serve dashboards via local HTTP/Nginx, validate data sources (REST/RPC) and confirm charts load. | ❌ | | |

## 2. Module Validation Checklist

For each module, run the specified local tests (unit/integration/fuzz) and manual flows where applicable.

| Module | Local Validation Steps | Status | Owner | Notes |
|--------|------------------------|--------|-------|-------|
| app (wiring) | `cd chain && go test ./app/...`; start local node and run `scripts/test-vc-issuer-e2e.sh`. | 🟡 | | `go test` pass (2025-01-18); manual node/e2e validation pending |
| monitoring | `go test ./chain/x/monitoring/...`; run CLI queries (`aurad q monitoring ...`). | 🟡 | | `go test` pass (2025-01-18); CLI validation pending |
| privacy | `go test ./chain/x/privacy/...`; execute CLI commands for stealth/ringsig features. | 🟡 | | `go test` pass (2025-01-18); CLI validation pending |
| economicsecurity | `go test ./chain/x/economicsecurity/...`; verify invariants via `go test ./chain/x/economicsecurity/keeper -run TestInvariant`. | 🟡 | | `go test` pass (2025-01-18); invariant rerun pending |
| security | `go test ./chain/x/security/...`; run alert workflows via CLI. | 🟡 | | `go test` pass (2025-01-18); CLI workflows pending |
| prevalidation | `go test ./chain/x/prevalidation/...`; simulate tx queue using integration tests. | 🟡 | | `go test` pass (2025-01-18); queue simulation pending |
| incidentresponse | `go test ./chain/x/incidentresponse/...`; trigger mock incidents via CLI. | 🟡 | | `go test` pass (2025-01-18); CLI flows pending |
| contractregistry | `go test ./chain/x/contractregistry/...`; CLI store/query flows. | 🟡 | | `go test` pass (2025-01-18); CLI flows pending |
| governance | `go test ./chain/x/governance/...`; submit proposal via CLI, ensure local voting works. | 🟡 | | `go test` pass (2025-01-18); proposal flow pending |
| identity | `go test ./chain/x/identity/...`; run `scripts/test-vc-issuer-e2e.sh` to ensure VC flows. | 🟡 | | `go test` pass (2025-01-18); e2e script pending |
| identitychange | `go test ./chain/x/identitychange/...`; CLI update flows. | 🟡 | | `go test` pass (2025-01-18); CLI updates pending |
| networksecurity | `go test ./chain/x/networksecurity/...`; run simulated Sybil detection tests locally. | 🟡 | | `go test` pass (2025-01-18); scenario replay pending |
| walletsecurity | `go test ./chain/x/walletsecurity/...`; execute CLI for wallet policies. | 🟡 | | `go test` pass (2025-01-18); CLI policies pending |
| validatorsecurity | `go test ./chain/x/validatorsecurity/...`; confirm invariants. | 🟡 | | `go test` pass (2025-01-18); invariant script pending |
| bridge | `go test ./chain/x/bridge/...`; run `scripts/hermes-bootstrap.sh` for ICS20 transfer simulation. | 🟡 | | `go test` pass (2025-01-18); Hermes flow pending |
| compliance | `go test ./chain/x/compliance/...`; execute KYC/AML CLI flows with encryption enabled. | 🟡 | | `go test` pass (2025-01-18); CLI/encryption validation pending |
| confidencescore | `go test ./chain/x/confidencescore/...`; insert/test IR completions via CLI. | 🟡 | | `go test` pass (2025-01-18); CLI flows pending |
| wasm | `go test ./chain/x/wasm/...`; run `scripts/test-vc-issuer-e2e.sh` for store/instantiate/execute. | 🟡 | | `go test` pass (2025-01-18); wasm CLI/e2e pending |
| aura-bindings | `go test ./chain/x/aura-bindings/...`; run binding CLI sample. | 🟡 | | `go test` pass (2025-01-18); CLI sample pending |
| inclusionroutines | `go test ./chain/x/inclusionroutines/...`; load sample routine via CLI. | 🟡 | | `go test` pass (2025-01-18); CLI routine pending |
| auth | `go test ./chain/x/auth/...`; verify invariants with `go test ./chain/x/auth/... -run TestInvariants`. | 🟡 | | `go test` pass (2025-01-18); invariants pending |
| dex | `go test ./chain/x/dex/...`; run `chain/x/dex/keeper/msg_server_integration_test.go`. | 🟡 | | `go test` pass (2025-01-18); integration/CLI pending |
| cryptography | `go test ./chain/x/cryptography/...`; run zk proof tests + CLI queries. | 🟡 | | `go test` pass (2025-01-18); CLI tests pending |
| dataregistry | `go test ./chain/x/dataregistry/...`; perform CLI store/query. | 🟡 | | `go test` pass (2025-01-18); CLI store/query pending |
| vcregistry | `go test ./chain/x/vcregistry/...`; ensure new GetDisclosureRequest query works via CLI. | 🟡 | | `go test` pass (2025-01-18); CLI query pending |
| common/internal | `go test ./chain/x/common/...` and `./chain/x/internal/...`. | 🟡 | | `go test` pass (2025-01-18); runtime validation pending |
| economics | `go test ./chain/x/economics/...`; run msg/query server integration tests. | 🟡 | | `go test` pass (2025-01-18); integration tests pending |

(Add any new modules or components as they are introduced.)

## 3. Wallet & Client Applications

| Client | Validation Steps | Status | Owner | Notes |
|--------|------------------|--------|-------|-------|
| Desktop Wallet | Connect to local RPC, perform send/receive, staking, gov vote, key export/import. | ❌ | | |
| Mobile Wallet | Use local dev server or emulator pointing to local RPC, confirm basic flows. | ❌ | | |
| Browser Extension | Load extension in dev mode, connect to local RPC, sign transactions. | ❌ | | |
| Web Wallet | Serve locally, ensure CORS set for local RPC, run full transaction flow. | ❌ | | |

## 4. Analytics / Tooling

| Tool | Validation Steps | Status | Owner | Notes |
|------|------------------|--------|-------|-------|
| Grafana Dashboards | Import dashboards, verify data sources, ensure alerts fire using test rules. | ❌ | | |
| Prometheus Rules | Run `promtool check rules` locally; simulate alerts using `scripts/testnet-monitor.sh --simulate`. | ❌ | | |
| Logs (Loki/Promtail) | Ensure promtail collects logs from docker services, query sample logs in Grafana. | ❌ | | |

## 5. Tracking & Updates
- Update this file after each validation session.
- Log detailed outcomes (commands, heights, screenshots) in `docs/testnet/STATUS_LOG.md`.
- Blockers requiring code changes should open issues referencing the relevant module.
