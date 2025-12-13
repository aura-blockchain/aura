Local 4-Validator + 2-Sentry Testnet
======================================

This doc captures the exact steps (and pitfalls) to spin up the working local testnet with four validators and two sentries using Docker Compose.

Prereqs
-------
- Docker + docker compose
- Go toolchain (for `scripts/testnet-init.sh` build step)
- From repo root (`/home/hudson/blockchain-projects/aura`)

One-shot bring-up
-----------------
1. Generate fresh testnet data and staking state:
   ```bash
   bash scripts/testnet-init.sh
   ```
   This builds `chain/aurad`, creates four validator homes under `testnet-data/validator-*`, fixes staking state, and writes the final `genesis.json`.

2. Populate the validator Docker volumes:
   ```bash
   cd testnet-data && ./populate-volumes.sh && cd ..
   ```

3. Make sure sentries share the same genesis and clean state (already done by the scripts below, but if reusing data, recopy genesis):
   ```bash
   cp testnet-data/validator-1/config/genesis.json testnet-data/sentry-1/config/genesis.json
   cp testnet-data/validator-1/config/genesis.json testnet-data/sentry-2/config/genesis.json
   rm -rf testnet-data/sentry-1/data testnet-data/sentry-2/data
   printf '{\n  "height": "0",\n  "round": 0,\n  "step": 0\n}\n' > testnet-data/sentry-1/data/priv_validator_state.json
   cp testnet-data/sentry-1/data/priv_validator_state.json testnet-data/sentry-2/data/priv_validator_state.json
   ```

4. Start the stack:
   ```bash
   docker compose -f docker-compose.testnet.yml up -d
   ```

5. Verify consensus (validators):
   ```bash
   for p in 27657 27757 27857 27957; do \
     echo "--- $p ---"; \
     curl -s http://localhost:$p/status | jq -r '.result.sync_info | "height=\(.latest_block_height) catching_up=\(.catching_up)"'; \
   done
   ```

6. Verify sentries:
   ```bash
   for p in 28659 28663; do \
     echo "--- $p ---"; \
     curl -s http://localhost:$p/status | jq -r '.result.sync_info | "height=\(.latest_block_height) catching_up=\(.catching_up)"'; \
   done
   ```

Wiring overview
---------------
- Validators peer **only** with sentries:
  - `persistent_peers = "b08caf40107965bb0f914bc16b9ad71f8754c9ba@172.26.0.20:26656,82101b73a7911812f8df27a78e7fa38472efac0b@172.26.0.21:26656"`
  - `pex = false`, `unconditional_peer_ids` set to both sentries.
- Sentries peer with all validators and each other; `pex = true`.
- Ports:
  - Validators RPC: 27657/27757/27857/27957
  - Validators P2P: 27656/27756/27856/27956
  - Sentries RPC: 28659/28663, P2P: 28658/28662, Metrics: 28661/28664

Pitfalls to avoid (what was fixed)
----------------------------------
- **Nil context panic in vcregistry InitGenesis**: fixed by passing `sdk.Context` (see `chain/x/vcregistry/module.go`).
- **Empty `delegator_address` in gentx**: use `--from` + `--keyring-dir` when creating gentx. Script now patches any blank fields.
- **Stray `stake` denom causing supply mismatch**: script strips `stake` from balances/supply and sets `bond_denom=uaura`.
- **Missing validators in staking state**: previously only gentx existed; staking validators/last powers were empty, causing “validator set is empty after InitGenesis”. Script now rebuilds `validators`, `last_validator_powers`, `last_total_power`, and `delegations` from gentx files.
- **Genesis divergence on sentries**: ensure sentries use the exact same `genesis.json` checksum as validators. If you see app-hash mismatch errors, recopy genesis and wipe `testnet-data/sentry-*/data`.
- **priv_validator_state.json missing**: sentries need an initial `data/priv_validator_state.json` (height 0) or they will crash on start.

Resetting cleanly
-----------------
If things get out of sync:
```bash
docker compose -f docker-compose.testnet.yml down -v
bash scripts/testnet-init.sh
cd testnet-data && ./populate-volumes.sh && cd ..
cp testnet-data/validator-1/config/genesis.json testnet-data/sentry-1/config/genesis.json
cp testnet-data/validator-1/config/genesis.json testnet-data/sentry-2/config/genesis.json
rm -rf testnet-data/sentry-1/data testnet-data/sentry-2/data
printf '{\n  "height": "0",\n  "round": 0,\n  "step": 0\n}\n' > testnet-data/sentry-1/data/priv_validator_state.json
cp testnet-data/sentry-1/data/priv_validator_state.json testnet-data/sentry-2/data/priv_validator_state.json
docker compose -f docker-compose.testnet.yml up -d
```

Quick health commands
---------------------
- Validator heights/hash:
  ```bash
  for p in 27657 27757 27857 27957; do curl -s http://localhost:$p/status | jq '.result.sync_info'; done
  ```
- Sentry heights:
  ```bash
  for p in 28659 28663; do curl -s http://localhost:$p/status | jq '.result.sync_info'; done
  ```
