# Aura Testnet Reference

This folder consolidates the documentation needed to stand up and operate the Aura testnet environments (`aura-testnet-1` and `aura-local-*`), covering multi-node setup, resilience validation, and public services.

## Multi-Node Local Testnet (Phase 1)

- Run `./scripts/testnet-init.sh` followed by `cd testnet-data && ./populate-volumes.sh && cd ..` to bootstrap four validators; the full sequence, networking notes, and monitoring pointers live in [`TESTNET_SETUP.md`](../TESTNET_SETUP.md).  
- Use `docker-compose -f docker-compose.testnet.yml up -d` to start validators + monitoring, and `./scripts/testnet-manage.sh help` for the available lifecycle commands (`start`, `status`, `logs`, `health`, `bft-test`, `clean`, etc.).
- Verify RPC/API/GRPC endpoints at the ports described in `TESTNET_QUICKSTART.md` and `TESTNET_SETUP.md:30-120`.
- Test Byzantine fault tolerance by running `./scripts/testnet-manage.sh bft-test` or manually stopping one validator and watching `curl http://localhost:26657/status`; the scripted flow is documented at `TESTNET_SETUP.md:178-208`.
- Validate state sync by cycling validator‑3 via `docker-compose -f docker-compose.testnet.yml stop validator-3`, waiting 30+ seconds, restarting it, and tailing `docker-compose -f docker-compose.testnet.yml logs -f validator-3` (`TESTNET_SETUP.md:194-208`).
- Exercise module transactions and queries (DEX pools, compliance screening, VC Registry) using the CLI and REST endpoints described in `TESTNET_SETUP.md:210-260`.

## Cloud Testnet (Phase 2)

- Apply the Kubernetes manifests under `k8s/base/` via `kubectl apply -k k8s/overlays/staging/` after populating the overlay (see `k8s/overlays/staging/kustomization.yaml`). The overlay now targets the `aura-testnet` namespace, uses `aequitas/aura:testnet`, and exposes `api.testnet.aura.network`, `rpc.testnet.aura.network`, and `grpc.testnet.aura.network` through the patched ingress.
- Populate the `config` `ConfigMap` and TLS secrets (`aura-testnet-tls`) with the production-like `genesis.json`, `config.toml`, and `app.toml` files; the init container copies those into each node before starting.
- Front the API/RPC endpoints with Cloudflare (rate-limit annotations in `k8s/base/ingress.yaml` already codify the behavior) and publish the DNS records once the stage cluster is reachable.
- Deploy the faucet (`faucet-service/`) and block explorer (`explorer/`) services with the RPC endpoints configured via `NODE_RPC_URL`/`NODE_API_URL`/`CHAIN_ID` so they can be reached at `faucet.testnet.aura.network` and `explorer.testnet.aura.network`.

## Genesis Coordination

- Follow the workflow in `docs/testing/TESTNET_CONFIGURATION.md:21-80` to init each node with `aurad init ... --chain-id aura-testnet-1`, add genesis accounts, run `aurad genesis gentx ...`, and run `aurad genesis collect-gentxs`. Distill the final `genesis.json` and distribute it to every `~/.aura/config/` directory before the agreed genesis time (see `docs/ops/UPGRADE_PROCEDURES.md:451-470` for download/publish steps).
- Log the genesis time, chain parameters, and diff in the release checklist so operators can verify the hash before starting their nodes.

## Public Services & Community

- Point the faucet at the testnet RPC and include your hCaptcha keys along with PostgreSQL/Redis connection strings as described in `faucet-service/README.md`.
- Deploy the block explorer container per `explorer/DEPLOYMENT_SUMMARY.md` (or via `docker-compose.yml`) so it exposes `/health`, `/api/search`, and WebSocket updates for blocks/txs; add module-specific queries (VC Registry, Inclusion Routines, AI Assistant) by reusing the JSON paths from `explorer/DEPLOYMENT_SUMMARY.md:20-50`.
- Publish the bug bounty program (`docs/testing/BUG_BOUNTY_PROGRAM.md`) alongside these docs to encourage responsible disclosure.
- Recruit validators by referencing `docs/ops/VALIDATOR_SETUP_GUIDE.md` (expected hardware, security, monitoring, and upgrade procedures) and by sharing status updates that point to the testnet dashboards/diffs.

## IBC / Hermes Integration

- The multi-chain e2e scaffolding in `chain/tests/e2e/README.md` and `chain/tests/e2e/chain.go:227-233` already demonstrates how to simulate IBC transfers between two Aura chains. Once the Cosmos Hub testnet channel exists, use that code as a template for Hermes-based relays and `suite.SimulateIBCTransfer(...)` to verify packets.
- Gear up Hermes by configuring it to talk to `rpc.testnet.aura.network:26657` and the Cosmos Hub testnet endpoint, then run `hermes hermes start`/`hermes hermes create channel` followed by a transfer; log the success in this folder for traceability.

## Compliance & Security

- Reference `docs/ops/compliance/minimal-aml-checklist.md` for transaction monitoring thresholds and escalation runbooks, `docs/security/HSM_INTEGRATION.md` for validator key protection, and `docs/compliance/SECURITIES_LAW_ANALYSIS.md` / `PRIVACY_POLICY.md` for legal guardrails.
- Add audit prep notes to this README so stakeholders can see which sections cover the Phase‑3 requirements (consensus/module audit, contracts, cryptography, P2P).

## Monitoring & Alerts

- Use the Grafana dashboards in `grafana/dashboards/` and alert rules in `docker/monitoring/prometheus/rules/aura-alerts.yml` to watch for chain halts, high resource usage, and peer count drops. Import the dashboards into the staging Grafana instance after connecting it to the Prometheus server running alongside the testnet nodes.

## Running Tests

- Execute `make test` and `go test ./chain/...` in the `chain/` directory to keep the suite green; aim for `aurad start` success on every new genesis drop.
- Keep the `scripts/testnet-manage.sh bft-test` and state sync steps documented in this README for easy reproduction.
