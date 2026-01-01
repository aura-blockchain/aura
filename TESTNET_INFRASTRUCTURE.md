# AURA Blockchain Testnet Infrastructure

> **Last Updated**: 2026-01-01
> **Maintainer**: Jeff DeCristofaro <info@aurablockchain.org>
> **License**: Apache 2.0

## Quick Reference

| Item | Value |
|------|-------|
| **Server IP** | 158.69.119.76 |
| **SSH Access** | `ssh aura-testnet` (from WSL2) |
| **VPN IP** | 10.10.0.1 |
| **Chain ID** | aura-testnet-1 |
| **Denom** | uaura |
| **Binary** | `~/.aura/cosmovisor/genesis/bin/aurad` |
| **Home Dir** | `~/.aura` |
| **RPC Port** | 26657 (localhost) |
| **P2P Port** | 26656 |
| **gRPC Port** | 9090 |
| **API Port** | 1317 |

---

## 1. Server Access

### SSH Configuration (from WSL2)
```bash
# Direct access
ssh aura-testnet

# Or explicitly
ssh ubuntu@158.69.119.76
```

SSH is configured in `~/.ssh/config` on WSL2 with key-based authentication.

### WireGuard VPN
- **Interface**: wg0
- **Private IP**: 10.10.0.1/24
- **Config**: `/etc/wireguard/wg0.conf`
- **Peers**: PAW (10.10.0.2), XAI (10.10.0.3)

```bash
# Check VPN status
sudo wg show

# Ping other nodes
ping 10.10.0.2  # PAW
ping 10.10.0.3  # XAI
```

---

## 2. Directory Structure

```
~/.aura/
├── config/
│   ├── app.toml          # Application configuration
│   ├── client.toml       # Client configuration
│   ├── config.toml       # CometBFT configuration
│   ├── genesis.json      # Chain genesis
│   ├── node_key.json     # Node identity key
│   └── priv_validator_key.json  # Validator signing key
├── cosmovisor/
│   ├── current -> genesis
│   └── genesis/
│       └── bin/
│           └── aurad     # Main binary
├── data/
│   ├── application.db/   # Application state (IAVL)
│   ├── blockstore.db/    # Block storage
│   ├── state.db/         # CometBFT state
│   ├── tx_index.db/      # Transaction index
│   ├── evidence.db/      # Evidence storage
│   ├── snapshots/        # State sync snapshots
│   └── priv_validator_state.json
└── logs/
    └── node.log          # Node output log

~/aura/                   # Source code repository
```

---

## 3. Binary & CLI

### Location
```bash
# Primary binary (via cosmovisor)
~/.aura/cosmovisor/genesis/bin/aurad

# Alias configured in ~/.bashrc
alias aurad="~/.aura/cosmovisor/genesis/bin/aurad"
```

### Common Commands
```bash
# Node status
aurad status --home ~/.aura

# Query modules
aurad query bank balances <address> --home ~/.aura
aurad query staking validators --home ~/.aura
aurad query gov proposals --home ~/.aura

# Keys management
aurad keys list --home ~/.aura
aurad keys add <name> --home ~/.aura
aurad keys show <name> --bech val --home ~/.aura

# Transaction commands
aurad tx bank send <from> <to> <amount> --home ~/.aura --chain-id aura-testnet-1
aurad tx staking delegate <validator> <amount> --home ~/.aura --chain-id aura-testnet-1
```

---

## 4. Configuration Files

### app.toml (Key Settings)
```toml
# Location: ~/.aura/config/app.toml

minimum-gas-prices = "0.001uaura"
pruning = "default"
iavl-disable-fastnode = true

[api]
enable = true
address = "tcp://localhost:1317"

[grpc]
enable = true
address = "localhost:9090"

[state-sync]
snapshot-interval = 2000
snapshot-keep-recent = 5
```

### config.toml (Key Settings)
```toml
# Location: ~/.aura/config/config.toml

moniker = "aura-testnet"

[rpc]
laddr = "tcp://127.0.0.1:26657"

[p2p]
laddr = "tcp://0.0.0.0:26656"
persistent_peers = ""
```

---

## 5. Service Management

### Systemd Service
```bash
# Service name
sudo systemctl status cosmovisor-aura
sudo systemctl start cosmovisor-aura
sudo systemctl stop cosmovisor-aura
sudo systemctl restart cosmovisor-aura

# View logs
sudo journalctl -u cosmovisor-aura -f
```

### Manual Start (if no systemd)
```bash
# Start node
nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &

# Stop node
pkill -f aurad

# View logs
tail -f ~/.aura/logs/node.log
```

---

## 6. Chain Operations

### Reset Chain (Full Reset)
```bash
# Stop node first
pkill -f aurad

# Reset all data
aurad tendermint unsafe-reset-all --home ~/.aura

# Start fresh
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &
```

### Export State
```bash
aurad export --home ~/.aura > genesis_export.json
```

### State Sync (Join Existing Network)
```bash
# Configure in config.toml
[statesync]
enable = true
rpc_servers = "<rpc1>,<rpc2>"
trust_height = <height>
trust_hash = "<hash>"
```

---

## 7. Monitoring & Debugging

### Check Node Status
```bash
# Basic status
aurad status --home ~/.aura | jq '.sync_info'

# Detailed status
curl -s localhost:26657/status | jq

# Check if catching up
aurad status --home ~/.aura | jq '.sync_info.catching_up'
```

### Check Logs for Errors
```bash
# Recent errors
grep -i error ~/.aura/logs/node.log | tail -20

# Watch live
tail -f ~/.aura/logs/node.log | grep -E "ERR|error|panic"
```

### Check Peers
```bash
curl -s localhost:26657/net_info | jq '.result.peers | length'
```

---

## 8. Genesis & Validator Info

### Chain Parameters
- **Chain ID**: aura-testnet-1
- **Denom**: uaura (1 AURA = 1,000,000 uaura)
- **Block Time**: ~4 seconds
- **Unbonding Period**: 21 days (testnet may vary)

### Validator
```bash
# Get validator address
aurad keys show <key-name> --bech val --home ~/.aura

# Check validator status
aurad query staking validator <valoper-address> --home ~/.aura
```

---

## 9. Source Code Repository

### Location on Server
```bash
~/aura/
```

### GitHub
```
git@github.com:aura-blockchain/aura.git
```

### Build from Source
```bash
cd ~/aura
git pull origin main
make build
cp build/aurad ~/.aura/cosmovisor/genesis/bin/aurad
```

---

## 10. Network Topology

```
┌─────────────────────────────────────────────────────────────────┐
│                     OVHcloud KS-5 Servers                       │
├─────────────────┬─────────────────┬─────────────────────────────┤
│   AURA Testnet  │   PAW Testnet   │        XAI Testnet          │
│  158.69.119.76  │  54.39.103.49   │       54.39.129.11          │
│   10.10.0.1     │   10.10.0.2     │        10.10.0.3            │
│   (wg0)         │   (wg0)         │        (wg0)                │
└────────┬────────┴────────┬────────┴────────────┬────────────────┘
         │                 │                     │
         └─────────────────┼─────────────────────┘
                           │
                    WireGuard Mesh
                    (Port 51820)
```

---

## 11. Ports Summary

| Port | Service | Binding |
|------|---------|---------|
| 26656 | P2P | 0.0.0.0 |
| 26657 | RPC | localhost |
| 9090 | gRPC | localhost |
| 1317 | REST API | localhost |
| 26660 | Prometheus | localhost |
| 51820 | WireGuard | 0.0.0.0 |

---

## 12. Troubleshooting

### Node Won't Start
```bash
# Check for port conflicts
sudo lsof -i :26656
sudo lsof -i :26657

# Check disk space
df -h

# Check memory
free -h

# Review logs
tail -100 ~/.aura/logs/node.log
```

### Sync Issues
```bash
# Check peer count
curl -s localhost:26657/net_info | jq '.result.n_peers'

# Check catching up status
aurad status --home ~/.aura | jq '.sync_info'

# Reset and resync if needed
aurad tendermint unsafe-reset-all --home ~/.aura
```

### Query Errors
```bash
# If "version does not exist" error, check ConsensusVersion in modules
# Ensure modules start at version 1 for fresh genesis

# Check gRPC is enabled
grep -A2 '^\[grpc\]' ~/.aura/config/app.toml
```

---

## 13. Maintenance Tasks

### Daily
- Monitor disk usage
- Check node sync status
- Review error logs

### Weekly
- Pull latest code updates
- Check for security patches
- Verify backups

### Monthly
- Review and rotate logs
- Update dependencies
- Performance review

---

## 14. Emergency Procedures

### Node Crash Recovery
```bash
# 1. Check what happened
tail -200 ~/.aura/logs/node.log

# 2. Try restart
pkill -f aurad
sleep 5
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &

# 3. If persistent issues, reset
aurad tendermint unsafe-reset-all --home ~/.aura
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &
```

### Backup Validator Keys
```bash
# CRITICAL - backup these files securely
cp ~/.aura/config/priv_validator_key.json /secure/backup/
cp ~/.aura/config/node_key.json /secure/backup/
```

---

## 15. Contact & Support

| Resource | Link |
|----------|------|
| **Maintainer** | Jeff DeCristofaro |
| **Email** | info@aurablockchain.org |
| **Bug Reports** | GitHub Issues |
| **Security Issues** | See SECURITY.md |

---

## 16. Related Documentation

- [Cosmos SDK Docs](https://docs.cosmos.network)
- [CometBFT Docs](https://docs.cometbft.com)
- [CONTRIBUTING.md](./CONTRIBUTING.md) - Contribution guidelines
- [LICENSE](./LICENSE) - Apache 2.0 License
- [AUTHORS](./AUTHORS) - Project authors and contributors
- [SECURITY.md](./SECURITY.md) - Security policy and reporting
