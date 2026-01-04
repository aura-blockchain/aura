# Aura Testnet Server (Current Deployment)

> **Last verified**: 2026-01-03
> **Server**: `aura-testnet` (158.69.119.76, VPN 10.10.0.1)

This folder contains templates and notes for the *current* Aura testnet
server deployment. Values below reflect the live configuration.

---

## 1) Live Port Map

| Component | Bind | Notes |
| --- | --- | --- |
| RPC | `0.0.0.0:10657` | CometBFT RPC (public endpoint should point here) |
| P2P | `0.0.0.0:26656` | Advertises VPN address |
| REST API | `127.0.0.1:1317` | REST API (currently empty replies) |
| gRPC | `127.0.0.1:9190` | App gRPC |
| GraphQL | `0.0.0.0:10400` | Internal only |
| WS proxy | `0.0.0.0:10082` | Internal only |
| Faucet API | `0.0.0.0:8081` | Faucet backend |
| Explorer API | `0.0.0.0:8082` | Explorer backend |
| Prometheus | `0.0.0.0:9090` | Monitoring |
| Grafana | `0.0.0.0:3000` | Monitoring UI |

---

## 2) Nginx Public Endpoints (Actual)

| Host | Backend | Notes |
| --- | --- | --- |
| testnet-rpc.aurablockchain.org | `aura_rpc_lb` | Points to 127.0.0.1:10657 |
| testnet-api.aurablockchain.org | `aura_api_lb` | Points to 127.0.0.1:1317 |
| testnet-grpc.aurablockchain.org | `grpc://127.0.0.1:9190` | gRPC upstream |
| testnet-graphql.aurablockchain.org | `http://127.0.0.1:10400` | GraphQL gateway |
| testnet-ws.aurablockchain.org | `http://127.0.0.1:10082` | WS proxy |
| testnet-archive.aurablockchain.org | `http://127.0.0.1:10657` | Proxies primary RPC |
| testnet-faucet.aurablockchain.org | `http://127.0.0.1:8081` | UI + API |
| testnet-explorer.aurablockchain.org | `http://127.0.0.1:8082` | Explorer UI/API |
| snapshots.aurablockchain.org | `/home/ubuntu/snapshots` | Static dir |
| monitoring.aurablockchain.org | `http://127.0.0.1:3000` | Grafana |

---

## 3) Systemd Services

```
- aurad.service
- aura-faucet.service
- aura-explorer.service
- aura-graphql.service
- aura-websocket-proxy.service
- prometheus.service
- grafana-server.service
```

---

## 4) Known Gaps

- REST API returns empty replies (investigate 1317).

---

## 5) Quick Verification (Local)

```bash
# RPC (local)
curl -s http://127.0.0.1:10657/status | jq '.result.sync_info'

# Faucet health
curl -s http://127.0.0.1:8081/health | jq

# Explorer
curl -s http://127.0.0.1:8082/ | head
```
