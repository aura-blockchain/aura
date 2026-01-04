# Public Testnet

## Chain ID

- `aura-testnet-1`

---

## Public Endpoints (Status as of 2026-01-03)

| Service | URL | Status | Notes |
| --- | --- | --- | --- |
| RPC | https://testnet-rpc.aurablockchain.org | **OK** | RPC on 10657. |
| REST API | https://testnet-api.aurablockchain.org | **Degraded** | REST port responds with empty replies. |
| gRPC | https://testnet-grpc.aurablockchain.org | **OK** | gRPC on 9190. |
| WebSocket | https://testnet-ws.aurablockchain.org | **OK** | WS proxy on 10082. |
| GraphQL | https://testnet-graphql.aurablockchain.org | **OK** | GraphQL gateway on 10400. |
| Faucet | https://testnet-faucet.aurablockchain.org | **OK** | UI + API on 8081. |
| Explorer | https://testnet-explorer.aurablockchain.org | **OK** | Explorer API/UI on 8082. |
| Archive RPC | https://testnet-archive.aurablockchain.org | **OK** | Currently proxies primary RPC (not a true archive). |
| Docs | https://docs.aurablockchain.org | **OK** | No testnet-docs host configured. |
| Monitoring | https://monitoring.aurablockchain.org | **OK** | Grafana. |
| Snapshots | https://snapshots.aurablockchain.org | **OK** | Snapshot directory. |

---

## Direct Server Access (For Operators)

| Service | Address |
| --- | --- |
| Server IP | 158.69.119.76 |
| VPN IP | 10.10.0.1 |
| RPC | http://127.0.0.1:10657 |
| REST API | http://127.0.0.1:1317 |
| gRPC | 127.0.0.1:9190 |
| P2P | 0.0.0.0:26656 |

---

## Join the Testnet

### Using the helper script

```bash
scripts/join-aura-testnet.sh
```

### Manual configuration

```bash
CHAIN_ID=aura-testnet-1 \
RPC_ENDPOINT=https://testnet-rpc.aurablockchain.org \
REST_ENDPOINT=https://testnet-api.aurablockchain.org \
AURA_HOME=$HOME/.aura \
scripts/join-aura-testnet.sh
```

> Note: public RPC/REST endpoints are currently degraded. If you are on the VPN
> or operating from the server, use the direct localhost endpoints above.

---

## Get Test Tokens

1. Create a wallet:
   ```bash
   aurad keys add mykey --home ~/.aura
   ```

2. Request tokens from the faucet:
   - Visit https://testnet-faucet.aurablockchain.org
   - Or use the CLI: `aurad tx faucet request --from mykey --home ~/.aura`

3. Check your balance:
   ```bash
   aurad query bank balances $(aurad keys show mykey -a --home ~/.aura) --home ~/.aura
   ```

---

## Quick Commands

```bash
# Check node status
curl -s http://127.0.0.1:10657/status | jq '.result.sync_info'

# REST API (local)
curl -s http://127.0.0.1:1317/cosmos/base/tendermint/v1beta1/blocks/latest | jq '.block.header.height'
```

---

## Hermes IBC Config (Operator Use)

```toml
[[chains]]
id = 'aura-testnet-1'
rpc_addr = 'http://127.0.0.1:10657'
grpc_addr = 'http://127.0.0.1:9190'
```

---

## Feedback

Use `docs/testnet/FEEDBACK.md` for bug reports and testnet feedback.

## Status

- **Network**: Active (public endpoints partially degraded)
- **Last Updated**: 2026-01-03
