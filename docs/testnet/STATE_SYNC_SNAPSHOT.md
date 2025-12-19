# State Sync & Snapshots (Aura Testnet)

Guide for bringing new validators or light nodes online quickly using state sync and optional snapshots. Defaults align with `aura-testnet-1` and the hardened RPC proxy (`https://rpc.aura-testnet.com`).

## Fast Path (one command)
```bash
# Optional overrides:
# MODE=light|full, AURA_HOME=~/.aura-testnet, RPC_ENDPOINT=https://rpc.aura-testnet.com
MODE=light ./scripts/join-aura-testnet.sh
```
The script:
- Inits the home directory and pulls genesis from the primary RPC
- Builds `persistent_peers` automatically from `/net_info` (limit 8, overridable via `PERSISTENT_PEERS`)
- Sets `minimum-gas-prices=0.025uaura`
- Applies pruning for light nodes (`keep-recent=1000`, `interval=50`)
- Enables state sync with a trusted header (`trust_height = latest - 200`)
- Writes seeds/persistent peers to `config.toml`

Start the node:
```bash
aurad start --home ~/.aura
```

## Manual State Sync (config.toml)
1. Capture a trusted header:
   ```bash
   RPC=https://rpc.aura-testnet.com
   LATEST=$(curl -s ${RPC}/block | jq -r '.result.block.header.height')
   TRUST_HEIGHT=$((LATEST-200))
   TRUST_HASH=$(curl -s "${RPC}/block?height=${TRUST_HEIGHT}" | jq -r '.result.block_id.hash')
   ```
2. Set in `~/.aura/config/config.toml`:
   ```toml
   [statesync]
   enable = true
   rpc_servers = "https://rpc.aura-testnet.com,https://rpc.aura-testnet.com"
   trust_height = TRUST_HEIGHT
   trust_hash = "TRUST_HASH"
   trust_period = "168h0m0s"
   ```
3. Peers (automatic query):
   ```bash
   curl -s ${RPC}/net_info | jq -r \
     '.result.peers[] | "\(.node_info.id)@\(.remote_ip // (.node_info.listen_addr|sub("tcp://";"")|split(":")[0])):\(.node_info.listen_addr|sub("tcp://";"")|split(":")[1])"' \
     | head -n 8 | paste -sd, -
   ```
   Add the resulting string to `persistent_peers` in `config.toml`.

## Snapshots (optional)
- If a snapshot URL is provided (LZ4-compressed tar), run:
  ```bash
  SNAPSHOT_URL=https://snapshots.aura-testnet.com/latest.tar.lz4 \
    ./scripts/join-aura-testnet.sh
  ```
- Manual restore:
  ```bash
  curl -L "$SNAPSHOT_URL" | lz4 -d | tar -C ~/.aura -xf -
  ```
- After restoring, keep state sync **disabled** to avoid conflicting resets.

## Seeds, Gas Prices, Pruning
- Seeds: set `SEEDS` env var or edit `config.toml` (`seeds = "id@host:26656,..."`).
- Persistent peers: automatically sourced from `/net_info`; override via `PERSISTENT_PEERS`.
- Minimum gas price: `0.025uaura` (from `docs/chain-registry/aura.json`), set in `app.toml`.
- Pruning presets:
  - Light: `keep-recent=1000`, `interval=50`
  - Full: `pruning="default"`

## Verification
- Sync progress: `curl -s localhost:26657/status | jq '.result.sync_info'`
- State sync completion: `catching_up` becomes `false` and `latest_block_height` matches peers.
- Peer set: `curl -s localhost:26657/net_info | jq '.result.n_peers'`

## References
- Chain registry: `docs/chain-registry/aura.json`
- Hardened proxy: `docker-compose.proxy.yml` (exposes `http://localhost:8080/rpc`)
- Light client validation: `docs/testnet/LIGHT_CLIENT_GUIDE.md`
