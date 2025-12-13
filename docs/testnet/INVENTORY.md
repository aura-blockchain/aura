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
| local-dev | sentry-1 | Sentry | (pending) | | | aura-lab-testnet | |
| local-dev | sentry-2 | Sentry | (pending) | | | aura-lab-testnet | |
| local-dev | observer-1 | Observer/RPC | (pending) | rpc.lab.aura.local | | aura-lab-testnet | Public proxy host TBD |
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
| RPC | `rpc.lab.aura.local` | observer-1.lab (nginx proxy) | Cloudflare tunnel optional only if public exposure required |
| REST API | `api.lab.aura.local` | observer-1.lab |  |
| gRPC | `grpc.lab.aura.local` | observer-1.lab | validator-1 gRPC exposed on `0.0.0.0:10090` (compose port `9090`); keep open for faucet/relayer |
| Explorer | `http://localhost:8088` | monitoring-1.lab (docker compose) | Local Ping.pub-style UI served by nginx; proxies stay on aura_aura-testnet (no AWS) |
| Faucet | `http://localhost:8081` | monitoring-1.lab (docker compose) | Local Postgres/Redis/Go stack (no cloud); `.env.faucet` drives config; last tx `A6FE8C...724B` height 2837 |
| Monitoring | `status.lab.aura.local` | monitoring-1.lab | Grafana + Alertmanager |
| Hermes Relayer | monitoring-1.lab | Hermes binary | Track client/connection/channel IDs below |

---

## Hermes / IBC Metadata

| Field | Value |
|-------|-------|
| Hermes Version | |
| Aura Client ID | |
| Cosmos Hub Client ID | |
| Connection ID | |
| Channel ID (ICS20) | |
| Relayer Keys | `relayer-aura`, `relayer-gaia` (stored in Vault) |
| Funding Status | Pending / Funded |

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
