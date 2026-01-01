# Public Testnet

## Chain ID

- `aura-testnet-1`

## Default endpoints

These defaults match `scripts/join-aura-testnet.sh` and can be overridden.

- RPC: `https://rpc.aura-testnet.com`
- REST: `https://api.aura-testnet.com`

## Join with the helper script

```bash
scripts/join-aura-testnet.sh
```

### Common overrides

```bash
CHAIN_ID=aura-testnet-1 \
RPC_ENDPOINT=https://rpc.aura-testnet.com \
REST_ENDPOINT=https://api.aura-testnet.com \
AURA_HOME=$HOME/.aura \
scripts/join-aura-testnet.sh
```

## Keys and funding

Create keys with `aurad keys add` and request test tokens from the faucet service (see `faucet/README.md`).
