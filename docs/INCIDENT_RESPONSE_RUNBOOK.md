# AURA Testnet Incident Response & Upgrade Runbook

## Server Details
| Role | IP | VPN | SSH |
|------|-----|-----|-----|
| Primary | 158.69.119.76 | 10.10.0.1 | `ssh aura-testnet` |
| Secondary | 139.99.149.160 | - | `ssh services-testnet` |

**Chain ID:** `aura-testnet-1` | **Binary:** `~/.aura/cosmovisor/genesis/bin/aurad` | **Home:** `~/.aura`

---

## 1. Node Down Recovery

```bash
# Check status
ssh aura-testnet
systemctl status aurad || ps aux | grep aurad

# Check logs for errors
tail -100 ~/.aura/logs/node.log
journalctl -u aurad -n 100 --no-pager

# Restart node
pkill -f aurad
nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &

# Verify sync
curl -s localhost:26657/status | jq '.result.sync_info'
```

---

## 2. Missed Blocks Recovery

```bash
# Check validator status
aurad query staking validator $(aurad keys show validator --bech val -a) --home ~/.aura

# Check signing info
aurad query slashing signing-info $(aurad tendermint show-validator --home ~/.aura) --home ~/.aura

# If jailed, unjail after downtime window
aurad tx slashing unjail --from validator --chain-id aura-testnet-1 --home ~/.aura -y

# Increase priority (reduce missed blocks)
# Edit ~/.aura/config/config.toml:
# timeout_commit = "1s"  # Reduce if too slow
```

---

## 3. Chain Halt Procedures

```bash
# 1. Confirm halt (check multiple nodes)
curl -s http://158.69.119.76:26657/status | jq '.result.sync_info.latest_block_height'

# 2. Check consensus state
curl -s localhost:26657/dump_consensus_state | jq '.result.round_state.height_vote_set'

# 3. Coordinate with other validators (Discord/Telegram)
# 4. If >2/3 validators agree, export state and restart:
aurad export --home ~/.aura > genesis_export.json

# 5. Reset state (CAUTION: coordinate with all validators)
aurad tendermint unsafe-reset-all --home ~/.aura --keep-addr-book

# 6. Replace genesis and restart
cp genesis_export.json ~/.aura/config/genesis.json
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &
```

---

## 4. Cosmovisor Upgrade Steps

```bash
# 1. Build new binary locally (on bcpc)
cd ~/blockchain-projects/aura && make build

# 2. Create upgrade directory on server
ssh aura-testnet "mkdir -p ~/.aura/cosmovisor/upgrades/<upgrade-name>/bin"

# 3. Copy binary
scp build/aurad aura-testnet:~/.aura/cosmovisor/upgrades/<upgrade-name>/bin/

# 4. Verify binary
ssh aura-testnet "~/.aura/cosmovisor/upgrades/<upgrade-name>/bin/aurad version"

# 5. Submit upgrade proposal (if governance required)
aurad tx gov submit-proposal software-upgrade <upgrade-name> \
  --title "Upgrade to vX.Y.Z" \
  --description "Description" \
  --upgrade-height <HEIGHT> \
  --from validator \
  --chain-id aura-testnet-1 \
  --home ~/.aura -y

# Cosmovisor handles the rest automatically at upgrade height
```

---

## 5. Manual Binary Upgrade (Emergency)

```bash
# 1. Stop node
ssh aura-testnet
pkill -f aurad

# 2. Backup current binary
cp ~/.aura/cosmovisor/genesis/bin/aurad ~/.aura/cosmovisor/genesis/bin/aurad.bak

# 3. Copy new binary (from bcpc)
scp build/aurad aura-testnet:~/.aura/cosmovisor/genesis/bin/

# 4. Verify and restart
ssh aura-testnet
chmod +x ~/.aura/cosmovisor/genesis/bin/aurad
~/.aura/cosmovisor/genesis/bin/aurad version
nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &

# 5. Monitor logs
tail -f ~/.aura/logs/node.log
```

---

## 6. Database Corruption Recovery

```bash
# Option A: Reset and sync from genesis
aurad tendermint unsafe-reset-all --home ~/.aura --keep-addr-book
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &

# Option B: Restore from snapshot
pkill -f aurad
rm -rf ~/.aura/data
# Download snapshot
curl -L https://artifacts.aurablockchain.org/snapshots/latest.tar.gz | tar -xzf - -C ~/.aura/
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &

# Option C: State sync from peer
# Edit ~/.aura/config/config.toml:
# [statesync]
# enable = true
# rpc_servers = "peer1:26657,peer2:26657"
# trust_height = <RECENT_HEIGHT>
# trust_hash = "<BLOCK_HASH>"
aurad tendermint unsafe-reset-all --home ~/.aura --keep-addr-book
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &
```

---

## 7. Key Compromise Response

```bash
# IMMEDIATE ACTIONS (within minutes):
# 1. Unbond all tokens
aurad tx staking unbond $(aurad keys show validator --bech val -a) <AMOUNT>uaura \
  --from validator --chain-id aura-testnet-1 --home ~/.aura -y

# 2. Stop compromised node
ssh aura-testnet "pkill -f aurad"

# 3. Secure/rotate keys
# Backup old keys (for investigation)
cp -r ~/.aura/config/priv_validator_key.json ~/.aura/config/priv_validator_key.json.compromised
cp -r ~/.aura/keyring-file ~/.aura/keyring-file.compromised

# 4. Generate new validator key
aurad init temp --chain-id aura-testnet-1 --home /tmp/newkeys
cp /tmp/newkeys/config/priv_validator_key.json ~/.aura/config/

# 5. Create new operator account
aurad keys add validator-new --home ~/.aura

# 6. After unbonding period, redelegate with new keys
# 7. Investigate breach, update access controls, rotate SSH keys
```

---

## 8. Network Split/Fork Recovery

```bash
# 1. Identify correct chain (check with trusted peers)
curl -s http://TRUSTED_PEER:26657/status | jq '.result.sync_info'

# 2. Compare block hashes
aurad query block <HEIGHT> --home ~/.aura | jq '.block_id.hash'

# 3. If on wrong fork, reset and sync
pkill -f aurad
aurad tendermint unsafe-reset-all --home ~/.aura --keep-addr-book

# 4. Ensure correct peers in config.toml
# persistent_peers = "node_id@ip:26656,..."

# 5. Restart and monitor
nohup aurad start --home ~/.aura > ~/.aura/logs/node.log 2>&1 &
curl -s localhost:26657/status | jq '.result.sync_info'
```

---

## 9. Emergency Contacts

| Role | Contact | Method |
|------|---------|--------|
| Lead Dev | TBD | Discord/Telegram |
| Validator Ops | TBD | Phone |
| Security | TBD | Signal |
| Hosting Provider | OVH | support.ovh.com |

**Discord:** #aura-validators | **Telegram:** @aura_validators

---

## 10. Monitoring Dashboard Links

| Service | URL |
|---------|-----|
| Node Status | `curl -s http://158.69.119.76:26657/status` |
| Netdata (bcpc) | http://192.168.100.2:19999 |
| Block Explorer | http://139.99.149.160:4000 (if running) |
| Prometheus | http://158.69.119.76:26660/metrics |
| Grafana | TBD |

**Quick Health Check:**
```bash
# One-liner status check
ssh aura-testnet 'curl -s localhost:26657/status | jq "{catching_up: .result.sync_info.catching_up, latest_height: .result.sync_info.latest_block_height, peers: .result.n_peers}"'
```

---

## Quick Reference Commands

```bash
# View logs
tail -f ~/.aura/logs/node.log

# Check peers
curl -s localhost:26657/net_info | jq '.result.n_peers'

# Get validator address
aurad tendermint show-address --home ~/.aura

# Check disk usage
du -sh ~/.aura/data

# Prune old blocks (if disk full)
aurad tendermint unsafe-reset-all --home ~/.aura --keep-addr-book
```
