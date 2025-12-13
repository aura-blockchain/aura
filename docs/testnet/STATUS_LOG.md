# Cloud Testnet Status Log

Time-ordered log of actions, issues, and milestones for `aura-testnet-1`. Update this file whenever infrastructure, configuration, or operational status changes. The goal is to keep Phase 2 progress transparent and to give handoffs enough context to continue without meetings.

## Logging Guidelines

- Prefer short entries with timestamp (UTC), author/agent, summary, and next actions.
- Reference relevant files/PRs/scripts when noting work (e.g., `infra/terraform/cloud-testnet`, `scripts/testnet-init.sh`).
- Document outages, restarts, Hermes operations, validator joins/leaves, config drift, or security events immediately.
- When closing an action item, link back to the earlier log entry for traceability.

## Template

```
### 2025-01-18 @ 14:32 UTC — Agent Name
- Action: terraform apply (us-east-1/us-west-2)
- Result: Provisioned validator-us-east-1, sentry-us-east-1, validator-us-west-2
- Follow-up: Run scripts/testnet-init.sh against new hosts, update docs/testnet/INVENTORY.md with IPs
```

## Entries

### 2025-01-18 @ 16:05 UTC — Automation Agent
- Action: Deprecated the remote cloud Terraform workflow and pivoted Phase 2 toward local racks + Docker/Kubernetes deployments.
- Result: Removed `infra/terraform/cloud-testnet`, rewrote docs to emphasize local inventory/logging.
- Follow-up: Provision additional local hosts/VMs per `docs/testnet/LOCAL_TESTNET_PLAN.md` and keep `docs/testnet/INVENTORY.md` updated manually.

### 2025-01-18 @ 18:02 UTC — Automation Agent
- Action: Full `go test ./...` run from `chain/` to validate module/unit suites locally.
- Result: Initial run failed due to unused import in `proto/aura/wasm/v1beta1/tx.pb.go`; removed the unused `types` alias and reran successfully (all Go tests now passing locally).
- Follow-up: Update `docs/testnet/LOCAL_VALIDATION_MATRIX.md` statuses for modules covered by the Go suites; proceed with service-level validation (validators, observers, faucet, explorer, Hermes).

### 2025-01-18 @ 19:05 UTC — Automation Agent
- Action: Ran `./scripts/testnet-init.sh` to bootstrap the 4-validator local testnet data (addresses, node IDs, genesis params).
- Result: Initial runs failed because security module params were zero (gentx validation complained about `max_requests_per_second`); added jq overrides in the script to set sane local defaults, piped keyring password into `gentx`, and successfully generated volumes + `populate-volumes.sh`. Inventory updated with validator addresses/IPs.
- Follow-up: Run `cd testnet-data && ./populate-volumes.sh` followed by `docker-compose -f docker-compose.testnet.yml up -d` to start the validators, then update `docs/testnet/LOCAL_VALIDATION_MATRIX.md` infrastructure section with runtime results.

### 2025-01-18 @ 19:30 UTC — Automation Agent
- Action: Populated docker volumes (`testnet-data/populate-volumes.sh`), added chown/permission fixes, updated configs to bind RPC/REST to `0.0.0.0`, and started the stack via `docker compose -f docker-compose.testnet.yml up -d`.
- Result: Four validators + Prometheus/Grafana are running locally; RPC verified at `curl http://127.0.0.1:27657/status` (height 7, not catching up). Previous restart loop (`client.toml` not found) resolved by ensuring permissions and regenerated configs.
- Follow-up: Validate Prometheus/Grafana dashboards, observer endpoints, and continue module-level CLI checks per `docs/testnet/LOCAL_VALIDATION_MATRIX.md`.

### 2025-01-18 @ 19:45 UTC — Automation Agent
- Action: Imported all Grafana dashboards via `scripts/import_grafana_dashboards.sh`, verified Prometheus ready endpoint (9094) and Grafana dashboards (API search), ran `scripts/testnet-monitor.sh quick` and `performance` to validate validator health and block times.
- Result: Monitoring stack confirmed healthy; metrics accessible, dashboards registered, validators producing blocks (~3.3s). `docs/testnet/LOCAL_VALIDATION_MATRIX.md` updated (monitoring ✅).
- Follow-up: Bring up observer/faucet/explorer services next and continue module-level validations.

### 2025-12-13 @ 00:55 UTC — Automation Agent
- Action: Finished Ping.pub-style frontend (`explorer/ping-pub-explorer`) with local RPC polling, rebuilt the explorer Docker image, and launched it via `docker compose -f docker-compose.explorer.yml up -d`.
- Result: `http://localhost:8088` now serves the local-only explorer with nginx proxying `/rpc`/`/api` directly to `aura-validator-1` on the compose network; `curl http://localhost:8088/rpc/status` reports live block heights (no cloud services involved). Updated validation matrix + inventory to reflect explorer availability.
- Follow-up: Document faucet bring-up procedure next and continue expanding local validation/test coverage before considering any remote environments.

### 2025-12-13 @ 01:02 UTC — Automation Agent
- Action: Reworked `scripts/faucet-setup.sh` to rely solely on local RPC checks (no AWS) and to fund the faucet wallet via validator-1, then built the Go faucet backend image (bumped to Go 1.25), and started the full docker-compose stack with Postgres/Redis/backend on the `aura_aura-testnet` network.
- Result: `.env.faucet` now captures the funded mnemonic/address (100k AURA allocation). `docker compose -f docker-compose.faucet.yml --env-file .env.faucet up -d` brings up the local-only faucet at `http://localhost:8081` with health + info endpoints responding; backend logs show DB/Redis connections and node polling. Balance queries still rely on placeholder REST responses, so follow-up work is required to wire Cosmos RPC/gRPC directly.
- Follow-up: Replace the faucet balance/tx logic with proper Cosmos SDK signing + Tendermint RPC queries so requests actually broadcast txs, then mark the faucet row in the validation matrix as fully ✅ once token transfers are confirmed.

### 2025-12-13 @ 03:20 UTC — Automation Agent
- Action: Implemented full Cosmos SDK signing/broadcast in `faucet-service/backend` using gRPC (`NODE_GRPC`) with tx factory + keyring, REST balance checks via `NODE_API`, and bech32 validation. Added deterministic sonic stub to work with Go 1.24 toolchain and updated env/docs/docker-compose defaults for RPC/API/GRPC endpoints.
- Result: `go test ./...` passes for the backend; API/E2E/integration tests skip cleanly without infra. Code now produces real MsgSend tx bytes instead of mock hashes.
- Follow-up: Run faucet against the local validator network to confirm live transfers, then flip the faucet row in `docs/testnet/LOCAL_VALIDATION_MATRIX.md` to ✅ and document the manual steps here (include tx hash examples). Update `docs/testnet/INVENTORY.md` with faucet endpoints once validated.

### 2025-12-13 @ 03:51 UTC — Automation Agent
- Action: Exposed validator gRPC on `0.0.0.0:9090`, switched faucet broadcast to Tendermint RPC (`broadcast_tx_sync`) with gRPC balance/account queries, and reran the faucet stack via `docker-compose.faucet.yml` + `.env.faucet`.
- Result: Faucet request succeeded end-to-end; tx hash `A6FE8C6599FFCFCFFA0FEFE5A196AD4E7580B4856CE7D7BA73BD200F7B9E724B` (height 2837) delivering `100000000uaura` to `aura1rzmhdpng8c6fykyn6rty4jdv9l5cp0345prwku`. Balance endpoint now reports funded faucet wallet.
- Follow-up: Update `docs/testnet/INVENTORY.md` with faucet endpoint (`http://localhost:8081`) and note gRPC exposure requirement (app.toml `grpc.address=0.0.0.0:9090`). Keep validator gRPC open for future relayer work.

### 2025-12-13 @ 03:26 UTC — Automation Agent
- Action: Added a dedicated observer node via `docker-compose.observer.yml` (new `testnet-data/observer-1` home, persistent peers, static IP `172.26.0.50`) and fronted it with the hardened nginx proxy (`docker-compose.proxy.yml`) so wallets/explorer stop hitting validator RPC directly.
- Result: `aura-observer-1` is syncing in lock-step with the validators; direct ports exposed on host (`28657` RPC, `2318` REST stub, `12090` gRPC) plus proxy endpoints at `http://localhost:8080/rpc` and `grpc localhost:12091`. Inventory + validation matrix updated to mark the observer row ✅.
- Follow-up: Point faucet/explorer env vars at the proxy host, then continue with sentry/Hermes validation before promoting anything beyond the lab network.

### 2025-12-13 @ 03:40 UTC — Automation Agent
- Action: Re-pointed faucet (backend/env/scripts) and explorer compose files to consume RPC/REST/gRPC from `aura-observer-1`, updated docs/env samples, and documented the hardened proxy ports (`8080`, `12091`) for local clients.
- Result: All local client stacks now depend on the observer instead of validator-1; faucet/explorer docs instruct using the proxy/observer endpoints, preventing wallets from ever talking to validator containers directly.
- Follow-up: Next infra milestone is sentry nodes + Hermes bootstrap so relayers can run against the same hardened RPC surface.

### 2025-12-13 @ 03:44 UTC — Automation Agent
- Action: Initialized two sentry nodes (`testnet-data/sentry-1`, `sentry-2`), wired persistent peers to all validators, assigned static compose IPs/ports, and added `sentry-{1,2}` services to `docker-compose.testnet.yml`.
- Result: `aura-sentry-1` (`d6eb…@172.26.0.20`) and `aura-sentry-2` (`baa8…@172.26.0.21`) are running and accessible via host RPC ports `28659/28663`; inventory + validation matrix updated to reflect active sentries.
- Follow-up: Bootstrap Hermes against the observer/sentry surface (update `config/hermes/config.toml` to use proxy endpoints) and document client/connection/channel progress.

### 2025-12-13 @ 03:47 UTC — Automation Agent
- Action: Repointed `config/hermes/config.toml` to consume the nginx proxy (`http://localhost:8080/rpc`, `grpc localhost:12091`) and added `scripts/hermes-bootstrap.sh` to automate client/connection/channel creation with the hardened endpoints.
- Result: Hermes no longer touches validator RPC directly; operators can run the bootstrap script after importing relayer keys to stand up ICS20 primitives. Docs updated to describe the new workflow.
- Follow-up: Fund relayer keys and execute the bootstrap script once the counterparty endpoints are ready, then log the resulting client/connection/channel IDs.

### 2025-12-13 @ 03:56 UTC — Automation Agent
- Action: Funded the local relayer key from validator-1 and ran `./scripts/hermes-bootstrap.sh` against the proxy endpoints.
- Result: Aura-side funding succeeded (tx `63762913B96A4932DFAFBCBB842B3CC7B2F89B27DBF6DBB6D4FEA6CE44ACDB82`), but the bootstrap aborted because the Cosmos Hub theta testnet RPC (`https://rpc.sentry-01.theta-testnet.polypore.xyz:26657`) timed out before returning version info.
- Follow-up: Retry the bootstrap once the counterparty RPC stabilizes or replace it with a reachable endpoint; document client/connection/channel IDs after a successful run.

### 2025-12-13 @ 03:59 UTC — Automation Agent
- Action: Swapped Hermes counterparty endpoints to `https://rpc.testnet.cosmos.network:443` / `https://grpc.testnet.cosmos.network:443` and re-ran the bootstrap with `--config config/hermes/config.toml`.
- Result: Hermes now reads the repo config, but the Cosmos Hub testnet RPC rejected the TLS handshake (`HandshakeFailure`). Still blocked on a stable counterparty RPC despite the local observer/sentry stack being healthy.
- Follow-up: Identify a reliable theta-testnet RPC (or run a local Gaia node) before attempting another bootstrap; once successful, log the resulting client/connection/channel IDs.

### 2025-12-13 @ 03:56 UTC — Automation Agent
- Action: Updated `scripts/testnet-init.sh` to bind gRPC to `0.0.0.0:9090` on all validators during init (was matching an already-updated pattern and leaving localhost binding).
- Result: Fresh testnet setups will expose gRPC without manual edits; faucet/relayer workloads can reach validators on the compose network by default.
- Follow-up: Regenerate testnet data if gRPC access is needed on existing volumes, or patch app.toml manually to match.

### 2025-12-13 @ 04:10 UTC — Automation Agent
- Action: Added `scripts/hermes-fund-keys.sh` to look up the Hermes relayer address and fund it from `validator-1`, updated `docs/testnet/HERMES_SETUP.md`, `docs/testnet/LOCAL_VALIDATION_MATRIX.md`, and `docs/testnet/INVENTORY.md` with the new workflow/details.
- Result: Funding the Aura-side relayer key is now a single command that records tx JSON under `logs/`; Hermes inventory now captures local version and UAURA funding status.
- Follow-up: Still need a reliable theta-testnet RPC (or local Gaia node) plus counterparty funding before rerunning `scripts/hermes-bootstrap.sh` to record client/connection/channel IDs.

### 2025-12-13 @ 04:18 UTC — Automation Agent
- Action: Added `scripts/counterparty-init.sh` and repointed `config/hermes/config.toml`/docs to use the local `aura-counter-1` chain (docker-compose.counterparty), including funding instructions for the `relayer-counter` key.
- Result: Both sides of the Hermes setup can now be bootstrapped with repo scripts (init, fund, compose) without relying on external theta-testnet RPC endpoints; inventory + validation matrix reference the new workflow.
- Follow-up: Import/fund the `relayer-counter` key from Vault, run `scripts/hermes-bootstrap.sh`, and log the resulting client/connection/channel IDs once the counterparty relayer key is ready.

### 2025-12-13 @ 05:52 UTC — Automation Agent
- Action: Reinitialized the `aura-counter-1` chain via `scripts/counterparty-init.sh` (adds security param patches + jq requirement), brought `docker-compose.counterparty.yml` back up, created/imported the `relayer-counter` key, and manually funded both relayer keys using offline `aurad tx` signing (`txhash 947F37E039D053875173F7B4BD643651D9B700829B3CED6EFEEB92404C439552` on aura-local-4, `A4C93ECA2820D5C0465B67C76C28F63EC87CB21AF3CA62E1A6E827AC2143BEB5` on aura-counter-1). Updated `scripts/hermes-bootstrap.sh` defaults to point at the counterparty.
- Result: Hermes now sees funded keys on both chains, but `scripts/hermes-bootstrap.sh` fails because the chain binaries do not expose `cosmos.tx.v1beta1.Service/Simulate` over gRPC (Hermes tries to simulate and receives `unknown service cosmos.tx.v1beta1.Service`). Clients/connections/channels remain uncreated.
- Follow-up: Expose the Cosmos Tx gRPC service (or add a dry-run-less path) so Hermes can send transactions without simulation; once fixed, re-run the bootstrap script and capture the resulting client/connection/channel IDs in `docs/testnet/INVENTORY.md`.
