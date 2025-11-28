# AURA Troubleshooting Guide

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** Node Operators, Validators, Support Engineers

---

## Table of Contents

1. [Quick Diagnostics](#quick-diagnostics)
2. [Node Won't Start](#node-wont-start)
3. [Sync Issues](#sync-issues)
4. [Performance Problems](#performance-problems)
5. [Network Connectivity](#network-connectivity)
6. [Validator Issues](#validator-issues)
7. [RPC/API Problems](#rpcapi-problems)
8. [Database Issues](#database-issues)
9. [Module-Specific Errors](#module-specific-errors)
10. [Emergency Procedures](#emergency-procedures)

---

## Quick Diagnostics

### Essential Commands

```bash
# Node status
aurad status 2>&1 | jq

# Check if running
ps aux | grep aurad
systemctl status aurad

# Recent logs
journalctl -u aurad -n 100 --no-pager

# Real-time logs
journalctl -u aurad -f

# Check version
aurad version

# Check sync status
aurad status 2>&1 | jq .SyncInfo

# Check peers
curl -s localhost:26657/net_info | jq '.result.n_peers'

# Check disk space
df -h

# Check memory
free -h

# Check port bindings
sudo netstat -tulpn | grep aurad
```

### Health Check Script

```bash
#!/bin/bash
# quick-diagnostic.sh

echo "=== AURA Node Diagnostics ==="
echo

# 1. Process check
echo "1. Process Status:"
if pgrep -x aurad > /dev/null; then
    echo "  ✓ aurad is running"
else
    echo "  ✗ aurad is NOT running"
fi
echo

# 2. Sync status
echo "2. Sync Status:"
SYNC_INFO=$(aurad status 2>&1 | jq -r .SyncInfo)
echo "  Latest Height: $(echo $SYNC_INFO | jq -r .latest_block_height)"
echo "  Catching Up: $(echo $SYNC_INFO | jq -r .catching_up)"
echo

# 3. Peers
echo "3. Network:"
PEERS=$(curl -s localhost:26657/net_info | jq -r '.result.n_peers')
echo "  Peers: $PEERS"
echo

# 4. Resources
echo "4. Resources:"
echo "  Disk: $(df -h ~/.aura | tail -1 | awk '{print $5}') used"
echo "  Memory: $(free -h | grep Mem | awk '{print $3 "/" $2}')"
echo

# 5. Recent errors
echo "5. Recent Errors:"
journalctl -u aurad --since "5 minutes ago" | grep -i error | tail -5
```

---

## Node Won't Start

### Problem: Service fails to start

**Check Logs:**
```bash
journalctl -u aurad -n 50 --no-pager
```

### Common Causes & Solutions

#### 1. Port Already in Use

**Symptom:**
```
panic: listen tcp 0.0.0.0:26656: bind: address already in use
```

**Solution:**
```bash
# Find process using port
sudo lsof -i :26656
sudo lsof -i :26657

# Kill old process
sudo kill -9 <PID>

# Or change port in config.toml
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:26756"/' ~/.aura/config/config.toml
```

#### 2. Corrupted Database

**Symptom:**
```
panic: database corruption detected
FATAL: failed to start node
```

**Solution:**
```bash
# Stop node
sudo systemctl stop aurad

# Backup current data
mv ~/.aura/data ~/.aura/data.backup

# Reset database
aurad unsafe-reset-all

# Restore from snapshot or resync
wget https://snapshots.aura.network/latest.tar.gz
tar -xzf latest.tar.gz -C ~/.aura/

# Start node
sudo systemctl start aurad
```

#### 3. Genesis File Mismatch

**Symptom:**
```
Error: genesis file hash mismatch
```

**Solution:**
```bash
# Download correct genesis
wget -O ~/.aura/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# Verify checksum
sha256sum ~/.aura/config/genesis.json
# Compare with official: <expected_hash>

# Reset and restart
aurad unsafe-reset-all
sudo systemctl restart aurad
```

#### 4. Insufficient Disk Space

**Symptom:**
```
fatal error: out of memory
write: no space left on device
```

**Solution:**
```bash
# Check disk space
df -h

# Clean up logs
sudo journalctl --vacuum-size=1G

# Prune blockchain data
# Edit app.toml:
pruning = "everything"  # More aggressive pruning

# Or expand storage
```

#### 5. Permission Issues

**Symptom:**
```
permission denied: /home/aura/.aura/data
```

**Solution:**
```bash
# Fix ownership
sudo chown -R aura:aura /home/aura/.aura

# Fix permissions
chmod 755 /home/aura/.aura
chmod 700 /home/aura/.aura/config
chmod 600 /home/aura/.aura/config/*
```

---

## Sync Issues

### Problem: Node won't sync or is stuck

#### Symptom 1: Catching Up = false but behind network

**Diagnosis:**
```bash
# Local height
LOCAL_HEIGHT=$(aurad status 2>&1 | jq -r .SyncInfo.latest_block_height)

# Network height
NETWORK_HEIGHT=$(curl -s https://rpc.aura.network:26657/status | jq -r .result.sync_info.latest_block_height)

echo "Local: $LOCAL_HEIGHT"
echo "Network: $NETWORK_HEIGHT"
echo "Difference: $((NETWORK_HEIGHT - LOCAL_HEIGHT))"
```

**Solution:**
```bash
# 1. Restart node
sudo systemctl restart aurad

# 2. If still stuck, reset peer connections
rm ~/.aura/config/addrbook.json
sudo systemctl restart aurad

# 3. Add known peers
# Edit config.toml
persistent_peers = "peer1@ip1:26656,peer2@ip2:26656"

# 4. Try state sync
# See STATE_SYNC section in NODE_OPERATOR_GUIDE.md
```

#### Symptom 2: Very slow sync

**Diagnosis:**
```bash
# Monitor block progression
watch -n 5 'aurad status 2>&1 | jq -r .SyncInfo.latest_block_height'

# Check peer quality
curl -s localhost:26657/net_info | jq '.result.peers[] | {moniker, remote_ip}'
```

**Solution:**
```bash
# 1. Use state sync or snapshot (fastest)
# Download snapshot
wget https://snapshots.aura.network/latest.tar.gz

# 2. Increase peer connections
# Edit config.toml
max_num_outbound_peers = 50

# 3. Upgrade hardware (especially disk I/O)

# 4. Use faster database backend
# Rebuild with rocksdb support
```

#### Symptom 3: Repeated blockchain resets

**Diagnosis:**
```bash
# Look for consensus failures
journalctl -u aurad | grep -i "consensus"
journalctl -u aurad | grep -i "evidence"
```

**Solution:**
```bash
# Likely wrong genesis or corrupted state
# 1. Verify genesis matches network
sha256sum ~/.aura/config/genesis.json

# 2. Complete reset
sudo systemctl stop aurad
rm -rf ~/.aura/data
aurad unsafe-reset-all
# Download correct genesis
sudo systemctl start aurad
```

---

## Performance Problems

### Problem: High CPU usage

**Diagnosis:**
```bash
top -p $(pgrep aurad)
htop  # Filter for aurad
```

**Causes & Solutions:**

1. **Normal during sync** - CPU-intensive during initial sync
   - Solution: Wait for sync to complete

2. **Transaction flood** - High transaction volume
   ```bash
   # Check mempool
   curl -s localhost:26657/num_unconfirmed_txs

   # Reduce mempool size in config.toml
   [mempool]
   size = 2000
   ```

3. **Peer overload** - Too many connections
   ```bash
   # Reduce peers in config.toml
   max_num_inbound_peers = 40
   max_num_outbound_peers = 10
   ```

### Problem: High memory usage

**Diagnosis:**
```bash
# Check memory
free -h

# Process memory
ps aux | grep aurad | awk '{print $6}'

# OOM killer events
dmesg | grep -i "out of memory"
journalctl -k | grep -i "killed process"
```

**Solutions:**

```bash
# 1. Reduce cache sizes in app.toml
[state-sync]
snapshot-interval = 0  # Disable if not serving snapshots

# 2. More aggressive pruning
pruning = "everything"

# 3. Reduce mempool
[mempool]
size = 1000
cache_size = 1000

# 4. Restart service periodically (temporary fix)
# Weekly restart during low usage
0 2 * * 0 systemctl restart aurad

# 5. Add swap (temporary)
sudo fallocate -l 8G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# 6. Upgrade RAM (permanent)
```

### Problem: High disk I/O

**Diagnosis:**
```bash
# Monitor I/O
iostat -x 1

# Check aurad I/O
sudo iotop -p $(pgrep aurad)

# Disk usage
du -sh ~/.aura/data/*
```

**Solutions:**
```bash
# 1. Use faster storage (NVMe SSD)

# 2. Reduce database writes
# Increase commit interval (config.toml)
[consensus]
timeout_commit = "10s"  # Slower but less I/O

# 3. Separate WAL to different disk
# (Requires rebuild)

# 4. Enable database caching
# Use RocksDB with tuned cache
```

---

## Network Connectivity

### Problem: No peers connecting

**Diagnosis:**
```bash
# Check peer count
curl -s localhost:26657/net_info | jq '.result.n_peers'

# Check P2P port
sudo netstat -tulpn | grep 26656

# Test external connectivity
telnet $(curl -s ifconfig.me) 26656
```

**Solutions:**

```bash
# 1. Check firewall
sudo ufw status | grep 26656
# Allow P2P port
sudo ufw allow 26656/tcp

# 2. Check external_address in config.toml
external_address = "tcp://YOUR_PUBLIC_IP:26656"

# 3. Add seed nodes
seeds = "seed1@seed1.aura.network:26656,seed2@seed2.aura.network:26656"

# 4. Reset address book
rm ~/.aura/config/addrbook.json
sudo systemctl restart aurad

# 5. Check if IP is banned
# Review logs for "banned peer"
journalctl -u aurad | grep -i "banned"
```

### Problem: Frequent peer disconnections

**Diagnosis:**
```bash
# Monitor connection changes
watch -n 2 'curl -s localhost:26657/net_info | jq ".result.n_peers"'

# Check for errors
journalctl -u aurad | grep -i "connection"
```

**Solutions:**
```bash
# 1. Network instability - use persistent peers
persistent_peers = "stable_peer1@ip1:26656,stable_peer2@ip2:26656"

# 2. Increase timeout values in config.toml
[p2p]
dial_timeout = "3s"
handshake_timeout = "20s"

# 3. Disable peer exchange if using private network
pex = false

# 4. Check for DDoS/rate limiting
# Review iptables/ufw rules
```

---

## Validator Issues

### Problem: Validator not signing blocks

**Diagnosis:**
```bash
# Check validator status
aurad query staking validator $(aurad keys show validator --bech val -a)

# Check signing info
aurad query slashing signing-info $(aurad tendermint show-validator)

# Recent logs
journalctl -u aurad | grep -i "signed\|missing"
```

**Causes & Solutions:**

#### 1. Node not synced
```bash
aurad status 2>&1 | jq .SyncInfo.catching_up
# If true, wait for sync to complete
```

#### 2. Validator jailed
```bash
# Check if jailed
aurad query staking validator $(aurad keys show validator --bech val -a) | grep jailed

# Unjail if downtime period expired
aurad tx slashing unjail \
  --from validator \
  --chain-id aura-mainnet-1 \
  --fees 5000uaura
```

#### 3. Insufficient voting power
```bash
# Check voting power
curl localhost:26657/validators | jq '.result.validators[] | select(.address=="YOUR_VAL_ADDRESS")'

# May need more stake
```

#### 4. Wrong validator key
```bash
# Verify key matches
aurad tendermint show-validator
# Should match registered validator pubkey
```

### Problem: Double sign risk

**Prevention:**
```bash
# NEVER run same validator key on multiple machines

# Check priv_validator_state.json
cat ~/.aura/data/priv_validator_state.json

# If setting up backup:
# 1. Stop primary completely
sudo systemctl stop aurad
sleep 30

# 2. Verify stopped
pgrep aurad || echo "Stopped successfully"

# 3. Copy state to backup
scp ~/.aura/data/priv_validator_state.json backup-server:~/.aura/data/

# 4. Start backup only
```

---

## RPC/API Problems

### Problem: RPC not responding

**Diagnosis:**
```bash
# Test RPC
curl localhost:26657/status

# Check if listening
sudo netstat -tulpn | grep 26657

# Check rate limits
journalctl -u nginx | grep -i limit
```

**Solutions:**
```bash
# 1. Verify RPC enabled in config.toml
[rpc]
laddr = "tcp://0.0.0.0:26657"

# 2. Increase connection limits
max_open_connections = 1000

# 3. If using nginx, check configuration
sudo nginx -t
sudo systemctl status nginx

# 4. Check for DDoS/attack
tail -f /var/log/nginx/access.log
```

### Problem: API returning errors

**Diagnosis:**
```bash
# Test API
curl localhost:1317/cosmos/base/tendermint/v1beta1/node_info

# Check logs
journalctl -u aurad | grep -i "api"
```

**Solutions:**
```bash
# Enable API in app.toml
[api]
enable = true
address = "tcp://0.0.0.0:1317"

# Check CORS if browser requests
enabled-unsafe-cors = false  # Or true if needed

# Restart node
sudo systemctl restart aurad
```

---

## Database Issues

### Problem: Database corruption

**Symptoms:**
```
panic: database corruption
leveldb: corruption detected
```

**Recovery:**
```bash
# 1. Stop node
sudo systemctl stop aurad

# 2. Backup corrupted data
mv ~/.aura/data ~/.aura/data.corrupted

# 3. Option A: Restore from backup
tar -xzf backup.tar.gz -C ~/.aura/

# 4. Option B: Resync from scratch
aurad unsafe-reset-all
# Download snapshot or use state sync

# 5. Restart
sudo systemctl start aurad
```

### Problem: Database growing too large

**Solutions:**
```bash
# 1. Enable pruning in app.toml
pruning = "default"  # Or "everything" for minimal size

# 2. Disable tx indexing if not needed
# config.toml
[tx_index]
indexer = "null"  # Disables tx indexing

# 3. Compact database
sudo systemctl stop aurad
aurad compact
sudo systemctl start aurad

# 4. Periodic snapshots and resync
# Monthly: Reset and resync from snapshot
```

---

## Module-Specific Errors

### Bridge Module Errors

```bash
# Error: Invalid Merkle proof
# Check bridge parameters
aurad query bridge params

# Verify chain is connected
aurad query bridge status

# Check recent bridge transactions
aurad query txs --events 'message.module=bridge'
```

### DEX Module Errors

```bash
# Error: Insufficient liquidity
# Check pool status
aurad query dex pool <pool-id>

# Error: Slippage too high
# Adjust slippage tolerance in transaction

# Check DEX parameters
aurad query dex params
```

### VC Registry Errors

```bash
# Error: Invalid credential
# Verify credential format
aurad query vcregistry credential <cred-id>

# Check issuer status
aurad query vcregistry issuer <issuer-address>
```

---

## Emergency Procedures

### Network Halt Recovery

```bash
# If network halted:

# 1. DO NOT restart independently
# 2. Join validator chat for coordination
# 3. Wait for official guidance
# 4. Coordinate restart at specific height
# 5. May require state export/import

# Emergency contacts:
# Discord: #validators-emergency
# Email: security@aura.network
```

### Ransomware/Security Breach

```bash
# If system compromised:

# 1. Immediately isolate system
sudo ufw deny all
sudo systemctl stop aurad

# 2. Do NOT pay ransom
# 3. Contact security team
# 4. Activate backup validator (if available)
# 5. Forensic investigation
# 6. Rebuild from clean backup
# 7. Rotate all keys
```

---

## Getting Help

**Before asking for help, collect:**
```bash
# System info
uname -a
cat /etc/os-release

# AURA version
aurad version --long

# Node status
aurad status 2>&1 | jq

# Recent logs
journalctl -u aurad -n 200 > logs.txt

# Hardware info
free -h
df -h
lscpu | grep -E '^Model name|^CPU\\(s\\)'
```

**Support Channels:**
- Discord: #node-support
- GitHub Issues: https://github.com/aequitas/aura/issues
- Validator Chat: #validators (for validators)
- Email: support@aura.network

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25
