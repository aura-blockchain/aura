# AURA Public Testnet Endpoints

## Chain ID

- `aura-testnet-1`

---

## Public Endpoints (Updated 2026-01-03)

| Service | URL | Status |
|---------|-----|--------|
| **RPC** | https://testnet-rpc.aurablockchain.org | OK |
| **REST API** | https://testnet-api.aurablockchain.org | OK |
| **gRPC** | https://testnet-grpc.aurablockchain.org | OK |
| **WebSocket** | wss://testnet-ws.aurablockchain.org | OK |
| **GraphQL** | https://testnet-graphql.aurablockchain.org/graphql | OK |
| **Explorer** | https://testnet-explorer.aurablockchain.org | OK |
| **Faucet** | https://testnet-faucet.aurablockchain.org | OK |
| **Archive RPC** | https://testnet-archive.aurablockchain.org | OK |
| **Docs** | https://testnet-docs.aurablockchain.org | OK |
| **Monitoring** | https://monitoring.aurablockchain.org | OK |
| **Snapshots** | https://snapshots.aurablockchain.org | OK |
| **Artifacts** | https://artifacts.aurablockchain.org | OK |

---

## Public Artifacts

Download testnet configuration files from https://artifacts.aurablockchain.org:

| File | URL | Description |
|------|-----|-------------|
| genesis.json | [Download](https://artifacts.aurablockchain.org/genesis.json) | Genesis file (required) |
| peers.txt | [Download](https://artifacts.aurablockchain.org/peers.txt) | Persistent peer list |
| seeds.txt | [Download](https://artifacts.aurablockchain.org/seeds.txt) | Seed nodes |
| addrbook.json | [Download](https://artifacts.aurablockchain.org/addrbook.json) | Address book |
| chain.json | [Download](https://artifacts.aurablockchain.org/chain.json) | Chain registry metadata |
| state_sync.md | [View](https://artifacts.aurablockchain.org/state_sync.md) | State sync guide |

---

## Direct Server Access (Operators)

| Service | Address |
|---------|---------|
| Server IP | 158.69.119.76 |
| VPN IP | 10.10.0.1 |
| RPC | http://127.0.0.1:10657 |
| REST API | http://127.0.0.1:1317 |
| gRPC | 127.0.0.1:9190 |
| P2P | 0.0.0.0:26656 |

---

## Get Test Tokens

1. Create a wallet:
   ```bash
   aurad keys add mykey --home ~/.aura
   ```

2. Request tokens from the faucet:
   - Visit https://testnet-faucet.aurablockchain.org
   - Or use the API:
     ```bash
     curl -X POST https://testnet-faucet.aurablockchain.org/faucet \
       -H "Content-Type: application/json" \
       -d '{"address": "aura1..."}'
     ```

3. Check your balance:
   ```bash
   aurad query bank balances $(aurad keys show mykey -a --home ~/.aura) --home ~/.aura
   ```

---

## Quick Commands

```bash
# Check node status
curl -s https://testnet-rpc.aurablockchain.org/status | jq '.result.sync_info'

# Query via REST API
curl -s https://testnet-api.aurablockchain.org/cosmos/auth/v1beta1/params | jq '.params'

# Query identity module
aurad query identity params --home ~/.aura
```
