# Aura Local Testnet (Docker)

This runbook brings up the 4-validator local testnet entirely in Docker using the prebuilt Compose stack.

## Prerequisites
- Docker + Docker Compose v2
- Rust/Go toolchains are **not** required on the host for runtime, but the image build compiles `aurad` from source inside the container.

## One-time initialization
Run from repo root:
```bash
# 1) Generate validator configs, keys, and genesis under ./testnet-data
./scripts/testnet-init.sh

# 2) Seed docker volumes with the generated configs
cd testnet-data && ./populate-volumes.sh && cd ..
```

## Start the cluster
```bash
docker compose -f docker-compose.testnet.yml up -d
```
Notes:
- First run builds the `aurad` image (Debian-based) and can take several minutes on slow networks while downloading Go modules.
- If the build is interrupted, simply rerun the command; it will reuse completed layers.

## Determinism check (auth/governance canonicalization)
Auth account ordering and governance param encoding are canonicalized in the app, but verify AppHashes across validators after startup:
```bash
# Expect identical hashes from all four validators after height > 10
for v in 1 2 3 4; do
  docker compose -f docker-compose.testnet.yml exec validator-$v aurad status | jq -r '.SyncInfo.latest_app_hash'
done
```
If hashes differ, capture per-validator `aurad status` and `logs -f validator-N` and re-run `./scripts/testnet-init.sh` to regenerate a clean genesis.

## Inspect / interact
- Logs: `docker compose -f docker-compose.testnet.yml logs -f validator-1`
- RPC: `http://localhost:27657`
- REST: `http://localhost:2317`
- gRPC: `localhost:10090`
- Metrics: `http://localhost:27660`

## Stop / clean
```bash
docker compose -f docker-compose.testnet.yml down
```
To reset state completely, delete `./testnet-data` and repeat the initialization steps.

## Contract deployment helper
With the cluster running, deploy CosmWasm artifacts:
```bash
./scripts/deploy-contracts.sh \
  --artifact contracts/artifacts/vc_issuer.wasm \
  --from validator \
  --chain-id aura-local-4 \
  --node http://localhost:27657 \
  --home ~/.aura
```
This script uploads the wasm, extracts the `code_id`, and instantiates by default (admin defaults to signer).
