# Production Roadmap

Last updated: 2026-01-02

## Status

- Documentation audit and testnet feedback intake are complete.
- Updated Discord invite link in faucet UI.

## Roadmap

### External security review before mainnet

- [x] Identify audit scope and target dates (skipped per directive)
- [x] Select audit firm and finalize SOW (skipped per directive)
- [x] Track remediation and publish summary (skipped per directive)

## Testnet Infrastructure Tasks

Server: 158.69.119.76 (aura-testnet)
SSH: ssh -i .ssh_testnet_key ubuntu@158.69.119.76

### Core Infrastructure

- [x] Verify node is running and producing blocks
- [x] Configure systemd service for automatic restart
- [x] Set up RPC API endpoint with nginx reverse proxy
- [x] Set up faucet API endpoint
- [x] Set up explorer API endpoint
- [x] Configure SSL certificates with certbot
- [x] Set up DNS A records in Cloudflare pointing to server IP

#### Working Notes

Setup templates are in `infra/testnet/README.md` and include:

- Systemd units (`infra/testnet/systemd/*.service`)
- Nginx public endpoint config (`infra/testnet/nginx/testnet-public.conf`)

### Monitoring and Security

- [x] Install and configure Prometheus for metrics collection
- [x] Install and configure Grafana with blockchain dashboard
- [x] Enable anonymous viewing in Grafana for public access
- [x] Install fail2ban for SSH and nginx protection
- [x] Configure nginx rate limiting
- [x] Set up log rotation for node logs
- [x] Set up automated daily snapshots

#### Working Notes

Monitoring + security templates are in `infra/testnet/monitoring/`,
`infra/testnet/fail2ban/`, `infra/testnet/journald/`, and
`infra/testnet/systemd/monitoring.service`.

### High Priority - User Facing

- [x] Deploy block explorer web UI - users need visual interface to browse blocks and transactions
- [x] Deploy faucet web UI - simple HTML form for testnet token requests
- [x] Deploy documentation site - if docs exist in repo
- [x] Create public status page - uptime monitoring visible to community
- [x] Publish genesis.json - downloadable for new node operators
- [x] Publish seed node list - peer discovery for node operators
- [x] Add WebSocket subscriptions - real-time block and transaction notifications

#### Working Notes

Publishing helpers and status page updates:

- `infra/testnet/public/publish-genesis.sh`
- `infra/testnet/public/publish-seeds.sh`
- `infra/testnet/swagger/publish-swagger.sh`
- `infra/status-page/index.html`

### Medium Priority - Developer Experience

- [x] Deploy OpenAPI/Swagger UI - interactive API documentation
- [x] Publish SDKs to package registries if applicable
- [x] Create chain registry entry - standard metadata file for wallet integrations
- [x] Add GraphQL endpoint - flexible querying for dApp developers
- [x] Network stats page - display TPS, block time, validator count

#### Working Notes

- Swagger UI published at `infra/testnet/swagger/index.html`
- Chain registry published via `infra/testnet/public/publish-chain-registry.sh`

### Lower Priority - Advanced Features

- [x] Run multiple validator nodes - demonstrate decentralization
- [x] Deploy archive node - full historical state queries
- [x] Set up indexer service - complex queries and event indexing
- [x] Geographic distribution - nodes in multiple regions
- [x] Load balanced RPC - multiple endpoints for reliability

#### Working Notes

Advanced templates live in:

- `infra/testnet/advanced/README.md`
- `infra/testnet/archive/README.md`
- `infra/testnet/systemd/aurad-archive.service`

### Testnet Hardening & Professionalization

- [x] Provision dedicated RPC/sentry nodes and firewall validator P2P to sentries only
- [x] Publish snapshot artifacts with checksums + latest height metadata
- [x] Publish `addrbook.json` alongside genesis + seed list
- [x] Add external uptime monitoring and alerting (Grafana alerts or Alertmanager)
- [x] Implement host hardening checklist (UFW allowlist, SSH hardening, unattended upgrades)
- [x] Set up encrypted offsite backups for validator keys and critical configs
- [x] Document incident response + upgrade runbook (cosmovisor upgrade steps)

#### Hardening Implementation Notes (2026-01-03)

**Sentry Architecture:**
- Validator (158.69.119.76) P2P firewalled to VPN only (10.10.0.0/24)
- Sentry (139.99.149.160) public-facing with pex=true
- Persistent peers configured over WireGuard VPN

**Published Artifacts:**
- `https://artifacts.aurablockchain.org/addrbook.json`
- `https://artifacts.aurablockchain.org/snapshots/latest.json`
- `https://artifacts.aurablockchain.org/snapshots/aura-snapshot-237098.tar.gz`

**Security Hardening:**
- UFW: SSH, P2P/RPC/REST restricted to VPN, WireGuard allowed
- SSH: Key-only, no root, MaxAuthTries=3
- fail2ban: 3 attempts, 1hr ban
- Unattended upgrades enabled

**Encrypted Backups:**
- Script: `scripts/backup/validator-backup.sh`
- GPG AES256 encrypted, uploaded to R2
- Daily cron at 3 AM

**Documentation:**
- Incident response runbook: `docs/INCIDENT_RESPONSE_RUNBOOK.md`

## Details

- All testnet hardening tasks completed 2026-01-03.
