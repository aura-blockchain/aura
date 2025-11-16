# Aura Blockchain Testnet Configuration Guide

## Overview

This document provides comprehensive instructions for configuring and deploying Aura blockchain testnets for various testing scenarios.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Network Topologies](#network-topologies)
3. [Configuration Files](#configuration-files)
4. [Validator Setup](#validator-setup)
5. [Genesis Configuration](#genesis-configuration)
6. [Module Configuration](#module-configuration)
7. [Security Settings](#security-settings)
8. [Monitoring and Observability](#monitoring-and-observability)
9. [Troubleshooting](#troubleshooting)

## Quick Start

### Single Node Testnet

```bash
# Initialize node
aurad init mynode --chain-id aura-testnet-1

# Create validator key
aurad keys add validator

# Add genesis account
aurad genesis add-genesis-account validator 100000000000stake

# Create genesis transaction
aurad genesis gentx validator 1000000000stake --chain-id aura-testnet-1

# Collect genesis transactions
aurad genesis collect-gentxs

# Start node
aurad start
```

### Multi-Node Testnet

```bash
# On each node, repeat the initialization
for i in {1..4}; do
    aurad init node$i --chain-id aura-testnet-1 --home ~/.aura-node$i
done

# Configure persistent peers
# Edit config.toml on each node to add peer addresses
```

## Network Topologies

### Development Testnet (Single Node)

**Purpose**: Local development and unit testing

**Configuration**:
- 1 validator node
- Fast block times (1-2 seconds)
- Minimal security requirements
- No external connectivity

**Use Cases**:
- Module development
- Smart contract testing
- Quick iterations

### Integration Testnet (3-4 Nodes)

**Purpose**: Integration testing and module interaction verification

**Configuration**:
- 3-4 validator nodes
- Moderate block times (5 seconds)
- Basic security enabled
- Local network or VPN

**Use Cases**:
- Cross-module testing
- Consensus verification
- State synchronization testing

### Staging Testnet (10+ Nodes)

**Purpose**: Pre-production testing and load testing

**Configuration**:
- 10+ validator nodes
- Production-like block times (6 seconds)
- Full security enabled
- Public internet connectivity
- Geographic distribution

**Use Cases**:
- Load testing
- Security audits
- Public testing
- Partner integration

### Chaos Testnet (Variable)

**Purpose**: Chaos engineering and resilience testing

**Configuration**:
- Variable number of nodes
- Random failure injection
- Network partitioning
- Byzantine behavior simulation

**Use Cases**:
- Fault tolerance testing
- Recovery procedures
- Edge case discovery

## Configuration Files

### app.toml

```toml
# API Configuration
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

# gRPC Configuration
[grpc]
enable = true
address = "0.0.0.0:9090"

# State Sync Configuration
[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 2

# Telemetry Configuration
[telemetry]
enabled = true
prometheus-retention-time = 60
```

### config.toml

```toml
# Consensus Configuration
[consensus]
timeout_propose = "3s"
timeout_propose_delta = "500ms"
timeout_prevote = "1s"
timeout_prevote_delta = "500ms"
timeout_precommit = "1s"
timeout_precommit_delta = "500ms"
timeout_commit = "5s"

# P2P Configuration
[p2p]
laddr = "tcp://0.0.0.0:26656"
external_address = ""
persistent_peers = ""
max_num_inbound_peers = 40
max_num_outbound_peers = 10

# Mempool Configuration
[mempool]
size = 5000
cache_size = 10000
max_txs_bytes = 1073741824

# RPC Configuration
[rpc]
laddr = "tcp://127.0.0.1:26657"
cors_allowed_origins = ["*"]
max_open_connections = 900
```

## Validator Setup

### Validator Key Generation

```bash
# Create validator operator key
aurad keys add validator-operator

# Create consensus key (automatically generated during init)
# Location: ~/.aura/config/priv_validator_key.json

# Backup keys securely
cp ~/.aura/config/priv_validator_key.json ~/secure-backup/
aurad keys export validator-operator > ~/secure-backup/validator-operator.key
```

### Validator Registration

```bash
# Create validator
aurad tx staking create-validator \
  --amount=1000000stake \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="MyValidator" \
  --chain-id=aura-testnet-1 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from=validator-operator
```

### Validator Security

1. **Key Management**
   - Use HSM for production validators
   - Implement key rotation policies
   - Maintain secure backups

2. **Sentry Node Architecture**
   ```
   [Public Network] <-> [Sentry Nodes] <-> [Validator Node (Private)]
   ```

3. **DDoS Protection**
   - Rate limiting
   - IP whitelisting
   - CloudFlare or similar CDN

## Genesis Configuration

### genesis.json Structure

```json
{
  "genesis_time": "2025-01-01T00:00:00.000000Z",
  "chain_id": "aura-testnet-1",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "100000000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000"
    },
    "validator": {
      "pub_key_types": ["ed25519"]
    }
  },
  "app_state": {
    "auth": {
      "params": {
        "max_memo_characters": "256",
        "tx_sig_limit": "7",
        "tx_size_cost_per_byte": "10"
      }
    },
    "bank": {
      "params": {
        "send_enabled": [],
        "default_send_enabled": true
      }
    },
    "staking": {
      "params": {
        "unbonding_time": "1814400s",
        "max_validators": 100,
        "max_entries": 7,
        "historical_entries": 10000,
        "bond_denom": "stake"
      }
    }
  }
}
```

### Custom Module Parameters

```json
{
  "vcregistry": {
    "params": {
      "max_vc_size": "1048576",
      "revocation_grace_period": "86400s"
    }
  },
  "inclusionroutines": {
    "params": {
      "min_confidence_score": "75",
      "max_routine_size": "10485760"
    }
  },
  "dex": {
    "params": {
      "min_liquidity": "1000",
      "swap_fee_rate": "0.003",
      "max_slippage": "0.05"
    }
  }
}
```

## Module Configuration

### Identity Change Module

```toml
[identitychange]
enabled = true
max_changes_per_day = 5
verification_required = true
```

### VC Registry Module

```toml
[vcregistry]
enabled = true
max_vc_per_identity = 100
revocation_enabled = true
selective_disclosure = true
```

### Data Registry Module

```toml
[dataregistry]
enabled = true
ipfs_endpoint = "http://localhost:5001"
max_data_size = "10485760"
encryption_required = false
```

### DEX Module

```toml
[dex]
enabled = true
min_liquidity = "1000"
swap_fee = "0.003"
max_price_impact = "0.10"
```

### Bridge Module

```toml
[bridge]
enabled = true
supported_chains = ["ethereum", "polygon", "bsc"]
relay_nodes = ["http://relay1:8545", "http://relay2:8545"]
security_threshold = 3
```

## Security Settings

### Firewall Rules

```bash
# Allow P2P
sudo ufw allow 26656/tcp

# Allow RPC (restrict to trusted IPs in production)
sudo ufw allow from 10.0.0.0/8 to any port 26657

# Allow API (restrict to trusted IPs in production)
sudo ufw allow from 10.0.0.0/8 to any port 1317

# Allow gRPC (restrict to trusted IPs in production)
sudo ufw allow from 10.0.0.0/8 to any port 9090

# Enable firewall
sudo ufw enable
```

### TLS Configuration

```bash
# Generate TLS certificates
openssl req -newkey rsa:4096 -nodes -keyout key.pem -x509 -days 365 -out cert.pem

# Update config.toml
[rpc]
tls_cert_file = "/path/to/cert.pem"
tls_key_file = "/path/to/key.pem"
```

### Rate Limiting

```toml
[rpc]
max_open_connections = 900
max_subscription_clients = 100
max_subscriptions_per_client = 5

[api]
max_open_connections = 1000
rpc_read_timeout = "10s"
rpc_write_timeout = "10s"
```

## Monitoring and Observability

### Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'aura-node'
    static_configs:
      - targets: ['localhost:26660']
```

### Grafana Dashboard

Key metrics to monitor:
- Block height
- Block time
- Transaction throughput
- Peer count
- Validator uptime
- Memory usage
- CPU usage
- Disk I/O

### Logging Configuration

```toml
# config.toml
log_level = "info"
log_format = "json"

[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

### Alerting Rules

```yaml
groups:
  - name: aura_alerts
    rules:
      - alert: NodeDown
        expr: up{job="aura-node"} == 0
        for: 5m
        annotations:
          summary: "Node {{ $labels.instance }} is down"

      - alert: BlockProductionSlow
        expr: rate(tendermint_consensus_height[5m]) < 0.1
        for: 10m
        annotations:
          summary: "Block production is slow"

      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes > 8e9
        for: 5m
        annotations:
          summary: "High memory usage detected"
```

## Troubleshooting

### Common Issues

#### 1. Node Not Syncing

**Symptoms**: Block height not increasing

**Solutions**:
```bash
# Check peer connections
aurad status | jq .SyncInfo

# Reset node (WARNING: This will delete all data)
aurad tendermint unsafe-reset-all

# Try state sync
# Edit config.toml to enable state sync
```

#### 2. Out of Memory

**Symptoms**: Node crashes, OOM errors in logs

**Solutions**:
```bash
# Increase system swap
sudo fallocate -l 8G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Adjust pruning settings in app.toml
pruning = "custom"
pruning-keep-recent = "100"
pruning-interval = "10"
```

#### 3. Network Partition

**Symptoms**: Node isolated, no peers

**Solutions**:
```bash
# Check network connectivity
nc -zv <peer-ip> 26656

# Update persistent peers
aurad config set config persistent_peers "peer1@ip1:26656,peer2@ip2:26656"

# Restart node
sudo systemctl restart aurad
```

#### 4. Database Corruption

**Symptoms**: Panic on startup, "database is locked"

**Solutions**:
```bash
# Stop node
sudo systemctl stop aurad

# Check database integrity
cd ~/.aura/data
ls -lh

# Restore from snapshot or resync
aurad tendermint unsafe-reset-all
# Then restore from backup or state sync
```

### Debug Commands

```bash
# Check node status
aurad status

# Query node info
aurad query node-info

# Check validator status
aurad query staking validator <validator-address>

# View logs
journalctl -u aurad -f

# Check peers
curl http://localhost:26657/net_info

# Check consensus state
curl http://localhost:26657/consensus_state
```

### Performance Tuning

```toml
# app.toml optimizations
[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 2

[store]
pruning = "custom"
pruning-keep-recent = "100"
pruning-interval = "10"

[grpc]
max-recv-msg-size = "10485760"
max-send-msg-size = "10485760"
```

## Best Practices

1. **Regular Backups**
   - Daily backup of validator keys
   - Weekly backup of blockchain data
   - Store backups in multiple locations

2. **Monitoring**
   - Set up 24/7 monitoring
   - Configure alerts for critical metrics
   - Maintain runbooks for common issues

3. **Security**
   - Keep software up to date
   - Use firewall rules
   - Implement DDoS protection
   - Regular security audits

4. **Disaster Recovery**
   - Document recovery procedures
   - Test recovery regularly
   - Maintain emergency contacts

5. **Capacity Planning**
   - Monitor disk usage
   - Plan for data growth
   - Scale infrastructure proactively

## Additional Resources

- [Cosmos SDK Documentation](https://docs.cosmos.network)
- [Tendermint Documentation](https://docs.tendermint.com)
- [Aura GitHub Repository](https://github.com/aequitas/aura)
- [Community Discord](https://discord.gg/aura)

## Support

For technical support:
- GitHub Issues: https://github.com/aequitas/aura/issues
- Discord: #testnet-support
- Email: support@aura-blockchain.io
