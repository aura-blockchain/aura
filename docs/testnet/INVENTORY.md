# Local Testnet Inventory Template

Source of truth for **Phase 2** infrastructure when running validators, sentries, observers, and supporting services on local/colo hardware. Update this document any time a host is added, re-imaged, or repurposed.

## Usage

1. Assign each rack/region a short code (e.g., `lab-east`, `lab-west`, `eu-colo`, `apac-colo`).
2. Record hostnames/IPs, WireGuard keys, and roles in the tables below. No external cloud automation is assumed — populate by hand from the monitoring dashboards or inventory spreadsheets.
3. Track genesis hashes, commit IDs, and service endpoints so validators can verify they are on the correct build.
4. Log operational events in `docs/testnet/STATUS_LOG.md` whenever inventory changes.

> **Reminder:** Never store private keys or secrets in this document. Reference their vault locations instead.

---

## Validator / Sentry / Observer Nodes

| Rack/Region | Hostname | Role | Host Type (bare metal / VM / container) | Public IP / DNS (if any) | Private IP / WireGuard IP | SSH Key Name | Notes |
|-------------|----------|------|----------------------------------------|--------------------------|--------------------------|--------------|-------|
| local-dev | validator-1 | Validator | Docker volume (compose) | - | 172.26.0.10 | aura-lab-testnet | Addr `aura1aq0t2wmf8w34tmxyfm34wwu4jw3rkkrskudrmd`, Node `47cdfa3ea6…` |
| local-dev | validator-2 | Validator | Docker volume (compose) | - | 172.26.0.11 | aura-lab-testnet | Addr `aura1p9ncgl7xn3t8qeknfasx28s5h8d93xwpxl97he`, Node `223ae06d83…` |
| local-dev | validator-3 | Validator | Docker volume (compose) | - | 172.26.0.12 | aura-lab-testnet | Addr `aura1q6jgc8qa6d64xt9kxa3l0supder88yy67zc22e`, Node `c1077a0e61…` |
| local-dev | validator-4 | Validator | Docker volume (compose) | - | 172.26.0.13 | aura-lab-testnet | Addr `aura1swjmu3hfhd8amjvvk5prggc2np7sxfanx3edld`, Node `c6fb375517…` |
| local-dev | sentry-1 | Sentry | Docker container (`docker-compose.testnet.yml`) | - | 172.26.0.20 | aura-lab-testnet | Peers with all validators only; exposes RPC `http://localhost:28659` for diagnostics |
| local-dev | sentry-2 | Sentry | Docker container (`docker-compose.testnet.yml`) | - | 172.26.0.21 | aura-lab-testnet | Backup sentry; RPC `http://localhost:28663` |
| local-dev | observer-1 | Observer/RPC | Docker container (`docker-compose.observer.yml`) | rpc.lab.aura.local (mapped to `localhost:8080` HTTP / `localhost:12091` gRPC) | 172.26.0.50 | aura-lab-testnet | Dedicated RPC/API/gRPC node; exposes direct host ports (RPC `28657`, REST `2318`, gRPC `12090`, metrics `28660`) and feeds the nginx proxy |
| local-dev | aura-counter | Counterparty | Docker container (`docker-compose.counterparty.yml`) | - | aura-counter-net | aura-lab-testnet | Standalone single-validator chain (`aura-counter-1`); RPC `http://localhost:29657`, gRPC `http://localhost:12092` |
| local-dev | monitoring-1 | Monitoring/Faucet | (pending) | status.lab.aura.local | | aura-lab-testnet | Prometheus/Grafana/Hermes |

*(Add/adjust rows for the actual hardware set; the entries above are placeholders.)*

---

## Genesis / Chain Metadata

| Field | Value |
|-------|-------|
| Chain ID | `aura-local-4` |
| Genesis SHA256 | |
| Genesis Source Commit | |
| Genesis File Location | `/networks/testnet/genesis.json` |
| App Commit Hash | |
| aurad Version | |
| Config Template | `config/templates/local-testnet/` |

---

## Services & Endpoints

| Service | Endpoint / Host | Backing Nodes | Notes |
|---------|-----------------|---------------|-------|
| RPC | `http://rpc.lab.aura.local/rpc` (`http://localhost:8080/rpc`) | observer-1.lab via `docker-compose.proxy.yml` | Nginx proxy fronts the observer; Apache on host already binds :80 so proxy listens on :8080 |
| REST API | `http://api.lab.aura.local/api` (`http://localhost:8080/api`) | observer-1.lab | REST handler currently returns status stub; future releases will expose full gRPC-gateway |
| gRPC | `grpc.lab.aura.local:12091` (`localhost:12091`) | observer-1.lab proxying to observer-1 gRPC (`12090`) | Use this port for wallet/faucet clients; validator-1 backup listener kept internal |
| Explorer | `http://localhost:8088` | monitoring-1.lab (docker compose) | Local Ping.pub-style UI served by nginx; proxies stay on aura_aura-testnet (no AWS) |
| Faucet | `http://localhost:8081` | monitoring-1.lab (docker compose) | Local Postgres/Redis/Go stack (no cloud); `.env.faucet` drives config; last tx `A6FE8C...724B` height 2837 |
| Monitoring | `status.lab.aura.local` | monitoring-1.lab | Grafana + Alertmanager |
| Hermes Relayer | monitoring-1.lab | Hermes binary | Use `scripts/hermes-bootstrap.sh` + proxy endpoints (`http://localhost:8080/rpc`, `grpc localhost:12091`); currently blocked on missing `cosmos.tx.v1beta1.Service/Simulate` gRPC support (Hermes cannot estimate gas) |

---

## Hermes / IBC Metadata

| Field | Value |
|-------|-------|
| Hermes Version | v1.13.2+bab3b80 (local CLI) |
| Aura Client ID | |
| Cosmos Hub Client ID | |
| Connection ID | |
| Channel ID (ICS20) | |
| Relayer Keys | `relayer-aura`, `relayer-counter` (stored in Vault) |
| Funding Status | Aura: offline-signed `aurad tx` (tx `947F37E039D053875173F7B4BD643651D9B700829B3CED6EFEEB92404C439552` at height TBD); Counterparty: offline-signed tx `A4C93ECA2820D5C0465B67C76C28F63EC87CB21AF3CA62E1A6E827AC2143BEB5` after running `scripts/counterparty-init.sh` |

---

## Contacts & Escalation

| Area | Owner | Contact |
|------|-------|---------|
| Hardware / Racks | | |
| Docker / Kubernetes | | |
| Monitoring / Alerts | | |
| Hermes / IBC | | |
| Validators / Genesis | | |

---

Keep this document updated before and after every infrastructure change to maintain situational awareness during the Phase 2 rollout. Local resources should be exhausted before considering any cloud services.
