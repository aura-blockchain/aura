# Local Multi-Region Testnet Plan (Phase 2)

Replaces the previous cloud-focused instructions with a **local-first strategy** using Docker/Kubernetes across homelab hardware, bare-metal servers, or on-prem virtual machines. Cloud services are only considered when there is absolutely no local equivalent.

## Target Topology

- **Validator nodes:** 4 primary validators + 2 sentry nodes distributed across lab racks or co-located machines (US-East, US-West, EU, APAC racks).
- **Observer nodes:** 2 RPC/API hosts behind a self-managed Nginx reverse proxy (can be containers or small bare-metal boxes).
- **Supporting services:** Prometheus, Grafana, Loki, Hermes relayer, faucet, Ping.pub explorer, Nginx proxy — all run via Docker Compose or Kubernetes on local equipment.
- **Networking:** WireGuard mesh between racks, VLAN segmentation on the switch, and optional Cloudflare tunnels only if a specific public endpoint must be exposed.

## Step-by-Step

### 1. Provision Hardware / VMs
1. Allocate Ubuntu 22.04 hosts (physical or Proxmox/VMware/KVM VMs) with 4 vCPU, 16GB RAM, 1TB NVMe for validators.
2. Label racks/regions in `docs/testnet/INVENTORY.md` (e.g., `lab-east`, `lab-west`, `eu-colo`, `apac-colo`).
3. Generate SSH keys `~/.ssh/aura-lab-testnet` for automation; share public keys with the ops team managing the racks.

### 2. Base Image Prep
1. Use the existing Docker images (`docker-compose.testnet.yml`, `docker-compose.explorer.yml`, etc.) or build a local OCI image with `make docker-build`.
2. For bare-metal deployments, install Go + aurad once, then use systemd units defined in `scripts/systemd/` (create if missing) so everything is managed locally.

### 3. Configuration Management
1. Maintain `inventory/local-testnet.ini` for Ansible (or simple SSH scripts) with hostnames/IPs for each rack.
2. Use `scripts/testnet-init.sh --home <remote path>` via SSH to push genesis configs to each host.
3. Templates for `config.toml`, `app.toml`, and `client.toml` live in `config/templates/`; populate them with per-rack seeds/persistent peers.

### 4. Secure Networking
1. Mesh the racks with WireGuard using `scripts/wireguard/generate-peers.sh`; store the peer map in `docs/testnet/INVENTORY.md`.
2. Restrict Tendermint P2P (26656) to the WireGuard/VLAN subnets. Keep RPC/API/GRPC ports bound to management networks or behind Nginx.
3. If public endpoints are required, use self-hosted reverse proxies (nginx/haproxy) and optional Cloudflare tunnels; no direct dependency on managed cloud load balancers.

### 5. Genesis Coordination
1. Use `scripts/testnet-collect-gentx.sh` and local file sharing (rsync, NFS, git-annex) to gather gentxs.
2. Validate `networks/testnet/genesis.json` locally with `./aurad genesis validate-genesis`.
3. Distribute the final genesis over SSH to each validator’s `~/.aura/config/`.

### 6. Service Bring-Up
1. Start validators sequentially using `docker compose -f docker-compose.testnet.yml up -d validator-{n}` or systemd units; verify with `curl localhost:26657/status`. Ensure `config/app.toml` binds gRPC to `0.0.0.0:9090` (handled by `scripts/testnet-init.sh`) so faucet/relayer can reach the node.
2. Launch sentries/observer nodes using the same compose stack, pointing RPC/API traffic through the local proxy host.
3. Deploy Ping.pub, faucet, and relayer containers on dedicated monitoring boxes via `docker-compose.explorer.yml`, `docker-compose.faucet.yml`, and `scripts/hermes-bootstrap.sh`.

### 7. Monitoring & Alerts
1. Run `docker-compose -f docker-compose.monitoring.yml up -d` on the monitoring host. All exporters connect over the WireGuard/VLAN network.
2. Import dashboards from `/grafana/dashboards/` and configure alert webhooks to internal channels (no third-party SaaS required).
3. Verify metrics ingestion per node and log findings in `docs/testnet/STATUS_LOG.md`.

### 8. Smoke Tests
1. Run `scripts/testnet-monitor.sh --hosts lab-east,lab-west,...` to check health across racks.
2. Execute `scripts/test-vc-issuer-e2e.sh` against the observer RPC endpoint to confirm wasm tx flow.
3. Run Hermes sanity transfer using local endpoints defined in `config/hermes/config.toml`.

### 9. Documentation Updates
1. Keep `docs/testnet/INVENTORY.md` updated with rack names, peer IDs, WireGuard keys, and service endpoints.
2. Record every rack bring-up, failure, or topology change in `docs/testnet/STATUS_LOG.md`.
3. Update `ROADMAP_PRODUCTION.md` once each rack/region milestone completes.

## Blocking Dependencies
- Physical lab hardware or virtualization capacity.
- Local networking (VLANs/WireGuard) with firewall rules managed in-house.
- On-call coverage for monitoring stack hosted locally.

## Automation Hooks
- Docker Compose stacks under the repo’s root (no Terraform).
- Optional Ansible scripts (`scripts/ansible/` when added) for pushing configs.
- CI tasks remain local-first; no cloud dependencies.
