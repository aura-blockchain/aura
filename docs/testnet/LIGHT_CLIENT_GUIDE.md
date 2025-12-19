# Aura Testnet Light Client (CometBFT)

Run a trustless RPC proxy that verifies headers and ABCI proofs with the CometBFT light client. The stack is configured for `aura-testnet-1` and uses the public RPC plus a local witness by default.

## Prerequisites
- Docker (Compose v2)
- `jq` (for grabbing trusted headers)
- Access to two RPC endpoints (primary + witness). Defaults: `https://rpc.aura-testnet.com` and `http://localhost:8080/rpc`.
- Optional: copy `.env.light-client.example` to `.env.light-client` and fill `TRUST_HEIGHT`/`TRUST_HASH`.

## Quick Start (public testnet defaults)
```bash
# 1) Capture a trusted header from the primary RPC
PRIMARY_RPC=${PRIMARY_RPC:-https://rpc.aura-testnet.com}   # set to http://localhost:8080/rpc if DNS/remote is unavailable
eval "$(PRIMARY_RPC=${PRIMARY_RPC} ./scripts/generate-trust-header.sh)"  # populates TRUST_HEIGHT/TRUST_HASH

# 2) Create the env file for docker-compose.light-client.yml
cat > .env.light-client <<EOF
CHAIN_ID=aura-testnet-1
PRIMARY_RPC=${PRIMARY_RPC}
WITNESS_RPC=http://localhost:8080/rpc
TRUST_LEVEL=1/3
TRUSTING_PERIOD=168h0m0s
TRUST_HEIGHT=${TRUST_HEIGHT}
TRUST_HASH=${TRUST_HASH}
LIGHT_CLIENT_PORT=8888
EOF

# 3) Start the light client proxy
docker compose -f docker-compose.light-client.yml up -d

# 4) Health check
curl -s http://localhost:8888/status | jq '.result.sync_info.latest_block_height'
```

The light client exposes the same RPC shape as a full node at `http://localhost:8888`. All proof-capable calls are verified before being returned.

**DNS fallback:** If `rpc.aura-testnet.com` is unreachable, set `PRIMARY_RPC`/`WITNESS_RPC` to a reachable endpoint (e.g., `http://localhost:8080/rpc` from `docker-compose.proxy.yml` or an IP-based RPC URL) before running the commands above.
**Bring up a local RPC if needed:** `docker compose -f docker-compose.observer.yml up -d && docker compose -f docker-compose.proxy.yml up -d` (or start the full testnet stack) to expose `http://localhost:8080/rpc` for trust header generation.

## Trustless Queries (examples)
- Latest block (validated):  
  `curl -s http://localhost:8888/block | jq .`
- Account balance with proof:  
  `curl -s "http://localhost:8888/abci_query?path=\"/cosmos.auth.v1beta1.Accounts\"&data=0x&prove=true" | jq .`
- Bank balance with store proof (replace address):  
  ```bash
  ADDR=aura1...
  KEY=$(aurad debug addr ${ADDR} | jq -r '.address_bytes')    # or use bech32 decode
  curl -s "http://localhost:8888/abci_query?path=\"/store/bank/key\"&data=0x${KEY}&prove=true" | jq .
  ```
- Event subscription (validated WS):  
  `wscat -c ws://localhost:8888/websocket` then `{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='NewBlock'"],"id":1}`
- Smoke test script: `LC_RPC=http://localhost:8888 ADDRESS=<optional_bech32> ./scripts/test-light-client.sh`

## Local Testnet (aura-local-4) Notes
- Use the local chain ID when your RPC comes from the dockerized validators (`CHAIN_ID=aura-local-4`, `PRIMARY_RPC=http://localhost:27657`, `WITNESS_RPC=http://localhost:27757`).
- Prefer host networking for the light client so it can reach the host-published validator ports (`network_mode: host` in `docker-compose.light-client.yml`).
- If you see `permission denied` creating `/cometbft/.light/light-client-db.db`, reset the volume with the correct owner (tmuser: `101:1000`):  
  ```bash
  docker rm -f aura-light-client || true
  docker volume rm aura-light-client || true
  docker volume create aura-light-client
  docker run --rm -v aura-light-client:/data alpine sh -c "chown -R 101:1000 /data"
  ```
- After starting, validate with `LC_RPC=http://localhost:8888 ./scripts/test-light-client.sh`.

### Validation Snapshot (local, 2025-12-18 16:32 UTC)
- Config: `CHAIN_ID=aura-local-4`, `PRIMARY_RPC=http://localhost:27657`, `WITNESS_RPC=http://localhost:27757`, `TRUST_HEIGHT=90251`, `TRUST_HASH=CA647F...90F5`, host networking.
- `./scripts/test-light-client.sh` output: height `90435`, `catching_up=false`, block hash `91D716E3ACE37780E55F86B39684E3A6EAD5F42EAAD5887E583C94A28C7BBEFE`.

## Operational Notes
- Data directory: `aura-light-client` Docker volume (`/cometbft/.light` inside the container).
- Change the listening port via `LIGHT_CLIENT_PORT` in `.env.light-client`.
- Multiple witnesses: set `WITNESS_RPC` to a comma-separated list (e.g., `http://localhost:8080/rpc,https://rpc.backup.aura-testnet.com`).
- If the trusting period elapses, regenerate `TRUST_HEIGHT`/`TRUST_HASH` from a fresh header and restart the container.

## Chain Registry & Endpoints
- Chain registry JSON: `docs/chain-registry/aura.json` (includes RPC/REST/gRPC endpoints for wallets and dApps).
- Public endpoints: `https://rpc.aura-testnet.com` (RPC/WebSocket), `https://api.aura-testnet.com` (REST), `grpc.aura-testnet.com:443` (gRPC).
- Local hardened proxy (observer stack): `http://localhost:8080/rpc` (via `docker-compose.proxy.yml`).

## Manual (no Docker) Command
```bash
cometbft light aura-testnet-1 \
  --primary https://rpc.aura-testnet.com \
  --witnesses http://localhost:8080/rpc \
  --trust-level 1/3 \
  --trusting-period 168h0m0s \
  --height ${TRUST_HEIGHT} \
  --hash ${TRUST_HASH} \
  --home-dir ~/.aura-light-client \
  --laddr tcp://0.0.0.0:8888
```

This replicates the compose setup for bare-metal hosts or systemd units.
