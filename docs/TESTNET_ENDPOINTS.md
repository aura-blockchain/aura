# AURA Public Testnet Endpoints

## Chain ID

`aura-mvp-1`

## Network Status

- **Status**: Active
- **Last Updated**: 2026-01-20
- **Validators**: 4 (val1, val2, val3, val4)
- **Sentries**: 2 (sentry1, sentry2)

---

## Chain Registry

```json
{
  "$schema": "../chain.schema.json",
  "chain_name": "aura",
  "chain_id": "aura-mvp-1",
  "pretty_name": "AURA Testnet",
  "network_type": "testnet",
  "status": "live",
  "bech32_prefix": "aura",
  "daemon_name": "aurad",
  "node_home": "$HOME/.aura",
  "key_algos": ["secp256k1"],
  "slip44": 118,
  "apis": {
    "rpc": [
      { "address": "https://testnet-rpc.aurablockchain.org", "provider": "AURA Foundation" }
    ],
    "rest": [
      { "address": "https://testnet-api.aurablockchain.org", "provider": "AURA Foundation" }
    ],
    "grpc": [
      { "address": "testnet-grpc.aurablockchain.org:443", "provider": "AURA Foundation" }
    ]
  },
  "explorers": [
    { "kind": "AURA Explorer", "url": "https://testnet-explorer.aurablockchain.org", "tx_page": "https://testnet-explorer.aurablockchain.org/tx/${txHash}" }
  ],
  "codebase": {
    "git_repo": "https://github.com/aura-blockchain/aura"
  }
}
```

---

## Public Endpoints

| Service | URL |
|---------|-----|
| RPC | https://testnet-rpc.aurablockchain.org |
| REST API | https://testnet-api.aurablockchain.org |
| gRPC | testnet-grpc.aurablockchain.org:443 |
| WebSocket | wss://testnet-ws.aurablockchain.org |
| GraphQL | https://testnet-graphql.aurablockchain.org/graphql |

---

## Resources

| Resource | URL |
|----------|-----|
| Explorer | https://testnet-explorer.aurablockchain.org |
| Faucet | https://testnet-faucet.aurablockchain.org |
| Documentation | https://testnet-docs.aurablockchain.org |
| Snapshots | https://snapshots.aurablockchain.org |
| Artifacts | https://artifacts.aurablockchain.org |

---

## Monitoring

| Dashboard | URL |
|-----------|-----|
| Grafana Home | https://monitoring.aurablockchain.org |
| Comprehensive | https://monitoring.aurablockchain.org/d/aura-comprehensive-v1 |
| Node Stats | https://monitoring.aurablockchain.org/d/aura-testnet-node |
| Validator | https://monitoring.aurablockchain.org/d/dtfjRVM7z |
| Cosmos Stats | https://monitoring.aurablockchain.org/d/9nXvbXO7z |
| Node Exporter | https://monitoring.aurablockchain.org/d/rYdddlPWk |

---

## Artifacts

Download testnet configuration files from https://artifacts.aurablockchain.org:

| File | Description |
|------|-------------|
| [genesis.json](https://artifacts.aurablockchain.org/genesis.json) | Genesis file (required) |
| [peers.txt](https://artifacts.aurablockchain.org/peers.txt) | Persistent peer list |
| [seeds.txt](https://artifacts.aurablockchain.org/seeds.txt) | Seed nodes |
| [addrbook.json](https://artifacts.aurablockchain.org/addrbook.json) | Address book |
| [chain.json](https://artifacts.aurablockchain.org/chain.json) | Chain registry metadata |

---

## Get Test Tokens

1. Create a wallet:
   ```bash
   aurad keys add mykey --home ~/.aura
   ```

2. Request tokens from the faucet:
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

---

## Internal Node Architecture

The testnet runs on two OVH servers with a validator/sentry topology.

### Server: aura-testnet (10.10.0.1)

| Node | Type | RPC | P2P | Home Directory |
|------|------|-----|-----|----------------|
| val1 | Validator | 127.0.0.1:26657 | 26656 | ~/.aura-mvp-val1 |
| val2 | Validator | 127.0.0.1:26757 | 26756 | ~/.aura-mvp-val2 |
| sentry1 | Sentry | 0.0.0.0:26680 | 26681 | ~/.aura-mvp-sentry1 |

Sentry1 additional ports: API 1319, gRPC 9092

### Server: services-testnet (10.10.0.4)

| Node | Type | RPC | P2P | Home Directory |
|------|------|-----|-----|----------------|
| val3 | Validator | 127.0.0.1:26657 | 26656 | ~/.aura-mvp-val3 |
| val4 | Validator | 127.0.0.1:26757 | 26756 | ~/.aura-mvp-val4 |
| sentry2 | Sentry | 0.0.0.0:26680 | 26681 | ~/.aura-mvp-sentry2 |

Sentry2 additional ports: API 1319, gRPC 9092

### Systemd Services

```bash
# Check node status
sudo systemctl status aurad-mvp-val1
sudo systemctl status aurad-mvp-sentry1

# View logs
journalctl -u aurad-mvp-val1 -f --no-hostname

# Restart node
sudo systemctl restart aurad-mvp-val1
```

---

## Security Baseline

See `docs/PUBLIC_TESTNET_SECURITY.md` for expected security controls and
`scripts/public-testnet-health-check.sh` for automated checks of public endpoints.

---

## Operator Notes

### Adding New Nodes

**Historical Note**: A non-determinism bug in the economics module (fixed 2026-01-20) prevented
nodes from replaying from genesis. Nodes that joined before this fix must sync state from an
existing healthy node. New nodes joining after deployment of the fix should be able to replay
from genesis normally.

For legacy nodes or quick sync, copy state from an existing node:

```bash
# Stop the new node
sudo systemctl stop aurad-mvp-newnode

# Copy data from a healthy node (e.g., val1)
rsync -av --delete ubuntu@10.10.0.1:~/.aura-mvp-val1/data/ ~/.aura-mvp-newnode/data/

# Start the new node
sudo systemctl start aurad-mvp-newnode
```

See `ARCHITECTURE.md` for the full node topology diagram.
