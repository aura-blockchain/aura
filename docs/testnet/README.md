# Public Testnet

## Chain ID

- `aura-testnet-1`

## Live Endpoints

| Service | URL |
|---------|-----|
| **RPC** | https://testnet-rpc.aurablockchain.org |
| **REST API** | https://testnet-api.aurablockchain.org |
| **gRPC** | testnet-rpc.aurablockchain.org:9090 |
| **Faucet** | https://testnet-faucet.aurablockchain.org |
| **Explorer** | https://testnet-explorer.aurablockchain.org |

### Direct Server Access (Development)

| Service | Address |
|---------|---------|
| Server IP | 158.69.119.76 |
| VPN IP | 10.10.0.1 |
| RPC | http://158.69.119.76:26657 |
| REST API | http://158.69.119.76:1317 |
| gRPC | 158.69.119.76:9090 |
| P2P | 158.69.119.76:26656 |

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

## Get Test Tokens

1. Create a wallet:
   ```bash
   aurad keys add mykey --home ~/.aura
   ```

<<<<<<< HEAD
2. Request tokens from the faucet:
   - Visit https://testnet-faucet.aurablockchain.org
   - Or use the CLI: `aurad tx faucet request --from mykey --home ~/.aura`

3. Check your balance:
   ```bash
   aurad query bank balances $(aurad keys show mykey -a --home ~/.aura) --home ~/.aura
   ```

## Feedback

Use `docs/testnet/FEEDBACK.md` for bug reports and testnet feedback.

## Quick Commands

```bash
# Check node status
aurad status --home ~/.aura

# Query latest block
curl -s https://testnet-rpc.aurablockchain.org/status | jq '.result.sync_info'

# Query via REST API
curl -s https://testnet-api.aurablockchain.org/cosmos/base/tendermint/v1beta1/blocks/latest | jq '.block.header.height'
```

## Hermes IBC Config

Example Hermes config is at `config/hermes/config.toml.example`. Update endpoints:

```toml
[[chains]]
id = 'aura-testnet-1'
rpc_addr = 'https://testnet-rpc.aurablockchain.org'
grpc_addr = 'https://testnet-rpc.aurablockchain.org:9090'
```

## Status

- **Network**: Active
- **Last Updated**: 2026-01-01
