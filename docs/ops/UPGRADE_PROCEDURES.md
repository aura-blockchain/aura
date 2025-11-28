# AURA Upgrade Procedures

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** Validators, Node Operators, DevOps

---

## Table of Contents

1. [Overview](#overview)
2. [Upgrade Types](#upgrade-types)
3. [Pre-Upgrade Preparation](#pre-upgrade-preparation)
4. [Governance Upgrades](#governance-upgrades)
5. [Manual Upgrades](#manual-upgrades)
6. [Cosmovisor Setup](#cosmovisor-setup)
7. [Rollback Procedures](#rollback-procedures)
8. [Testing on Testnet](#testing-on-testnet)
9. [Emergency Procedures](#emergency-procedures)

---

## Overview

AURA supports two upgrade mechanisms:
- **Governance Upgrades**: Coordinated via on-chain governance
- **Manual Upgrades**: Coordinated off-chain (emergency fixes)

All production upgrades should be tested on testnet first.

---

## Upgrade Types

### Coordinated Upgrades

**Characteristics:**
- Planned via governance proposal
- Specific block height trigger
- Network-wide coordination
- Automated with Cosmovisor

**Example Scenarios:**
- Protocol upgrades
- New module additions
- Breaking API changes
- Security patches (non-urgent)

### Emergency Upgrades

**Characteristics:**
- Urgent security fixes
- Minimal coordination time
- Manual intervention required
- May require network halt

**Example Scenarios:**
- Critical security vulnerabilities
- Consensus bugs
- Network halt recovery

---

## Pre-Upgrade Preparation

### Validator Checklist

```bash
# ✓ Backup current state
tar -czf ~/aura-backup-$(date +%Y%m%d).tar.gz ~/.aura/config ~/.aura/data/priv_validator_state.json

# ✓ Verify current version
aurad version --long

# ✓ Download new binary
wget https://github.com/aequitas/aura/releases/download/v2.0.0/aurad-v2.0.0-linux-amd64
chmod +x aurad-v2.0.0-linux-amd64

# ✓ Verify checksum
sha256sum aurad-v2.0.0-linux-amd64
# Compare with official release notes

# ✓ Test on testnet
# (See Testing on Testnet section)

# ✓ Verify upgrade height
aurad query gov proposal <upgrade-proposal-id>

# ✓ Monitor communication channels
# - Discord #validators
# - Governance forum
# - GitHub releases

# ✓ Prepare rollback plan
# (See Rollback Procedures section)

# ✓ Alert delegators
# Post upgrade notice 48-72 hours in advance

# ✓ Schedule maintenance window
# Coordinate with team for monitoring
```

### Infrastructure Preparation

```bash
# Ensure sufficient disk space
df -h ~/.aura
# Need at least 50GB free

# Check system resources
free -h
htop

# Verify backup system operational
# Test restore from backup

# Update monitoring alerts
# Disable non-critical alerts during upgrade window

# Prepare emergency contacts
# Ensure all team members available
```

---

## Governance Upgrades

### Step 1: Monitor Governance

```bash
# List active proposals
aurad query gov proposals --status voting_period

# Get upgrade proposal details
aurad query gov proposal <proposal-id>

# Vote on proposal (validators)
aurad tx gov vote <proposal-id> yes \
  --from validator-operator \
  --chain-id aura-mainnet-1 \
  --fees 5000uaura
```

### Step 2: Prepare for Upgrade

```bash
# Note upgrade height from proposal
UPGRADE_HEIGHT=12345678

# Download new binary
wget https://github.com/aequitas/aura/releases/download/v2.0.0/aurad
chmod +x aurad

# Verify binary
./aurad version
sha256sum aurad

# Place binary in path
sudo mv aurad /usr/local/bin/aurad-v2.0.0
```

### Step 3: Upgrade Execution

**Option A: Cosmovisor (Recommended)**

```bash
# Cosmovisor automatically swaps binaries at upgrade height
# See Cosmovisor Setup section for configuration

# Monitor upgrade
journalctl -u cosmovisor -f

# Cosmovisor will:
# 1. Detect upgrade height
# 2. Stop current binary
# 3. Swap to new binary
# 4. Restart node
```

**Option B: Manual Upgrade**

```bash
# Monitor block height
watch -n 1 'aurad status 2>&1 | jq -r .SyncInfo.latest_block_height'

# When approaching upgrade height (e.g., 100 blocks before):
# Stop node
sudo systemctl stop aurad

# Backup current binary
sudo cp /usr/local/bin/aurad /usr/local/bin/aurad-v1.0.0

# Install new binary
sudo cp /usr/local/bin/aurad-v2.0.0 /usr/local/bin/aurad

# Start node with new binary
sudo systemctl start aurad

# Monitor logs
journalctl -u aurad -f

# Verify upgrade
aurad version
aurad status
```

### Step 4: Post-Upgrade Verification

```bash
# Check node is running new version
aurad version --long

# Verify sync status
aurad status 2>&1 | jq .SyncInfo

# Check validator is signing (for validators)
aurad query slashing signing-info $(aurad tendermint show-validator)

# Monitor consensus
curl http://localhost:26657/consensus_state | jq

# Check for errors in logs
journalctl -u aurad --since "10 minutes ago" | grep -i error

# Verify API endpoints
curl http://localhost:26657/status
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info

# Monitor metrics
curl http://localhost:26660/metrics | grep aura
```

---

## Manual Upgrades

### Coordinated Manual Upgrade

```bash
# 1. Core team announces upgrade
#    - Specific UTC time
#    - Block height target
#    - Binary download location

# 2. Operators prepare
#    - Download binary
#    - Verify checksum
#    - Test on devnet

# 3. At coordinated time:
# Stop node
sudo systemctl stop aurad

# Backup
cp -r ~/.aura/data/priv_validator_state.json ~/backup/

# Install new binary
sudo mv aurad-new /usr/local/bin/aurad

# Start node
sudo systemctl start aurad

# 4. Monitor network consensus
# Network resumes when >67% voting power online
```

### Emergency Security Patch

```bash
# URGENT: Critical vulnerability patch

# 1. Stop node immediately
sudo systemctl stop aurad

# 2. Apply patch
wget https://github.com/aequitas/aura/releases/download/v1.0.1-hotfix/aurad
chmod +x aurad
sha256sum aurad  # Verify with security advisory

# 3. Install patched binary
sudo mv aurad /usr/local/bin/

# 4. Restart
sudo systemctl start aurad

# 5. Verify patch applied
aurad version
journalctl -u aurad -f

# 6. Monitor for attack signatures
# Review logs for exploitation attempts
grep -i "CVE-XXXX" /var/log/aurad.log
```

---

## Cosmovisor Setup

Cosmovisor automates binary swaps during upgrades.

### Installation

```bash
# Install Cosmovisor
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Verify installation
cosmovisor version
```

### Directory Structure

```bash
# Create Cosmovisor directory structure
mkdir -p ~/.aura/cosmovisor/genesis/bin
mkdir -p ~/.aura/cosmovisor/upgrades

# Copy current binary to genesis
cp $(which aurad) ~/.aura/cosmovisor/genesis/bin/

# Verify
~/.aura/cosmovisor/genesis/bin/aurad version
```

### Configuration

```bash
# Configure Cosmovisor environment
sudo tee /etc/systemd/system/cosmovisor.service > /dev/null <<EOF
[Unit]
Description=Cosmovisor AURA
After=network-online.target

[Service]
User=aura
ExecStart=$(which cosmovisor) run start
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

Environment="DAEMON_NAME=aurad"
Environment="DAEMON_HOME=/home/aura/.aura"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="UNSAFE_SKIP_BACKUP=false"
Environment="DAEMON_DATA_BACKUP_DIR=/home/aura/aura-backups"

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable cosmovisor
sudo systemctl start cosmovisor

# Monitor
journalctl -u cosmovisor -f
```

### Preparing Upgrade Binaries

```bash
# When upgrade proposal passes with name "v2.0.0"
UPGRADE_NAME="v2.0.0"

# Create upgrade directory
mkdir -p ~/.aura/cosmovisor/upgrades/$UPGRADE_NAME/bin

# Download and place new binary
wget https://github.com/aequitas/aura/releases/download/v2.0.0/aurad
chmod +x aurad
mv aurad ~/.aura/cosmovisor/upgrades/$UPGRADE_NAME/bin/

# Verify
~/.aura/cosmovisor/upgrades/$UPGRADE_NAME/bin/aurad version

# Cosmovisor will automatically swap at upgrade height
```

---

## Rollback Procedures

### When to Rollback

- Critical bug discovered post-upgrade
- Network fails to reach consensus
- Data corruption detected
- Security vulnerability in new version

### Rollback Steps

```bash
# 1. STOP NODE IMMEDIATELY
sudo systemctl stop aurad  # or cosmovisor

# 2. Assess situation
# - Is network halted?
# - How many validators rolled back?
# - Coordination with other validators needed?

# 3. Restore previous binary
sudo cp /usr/local/bin/aurad-v1.0.0 /usr/local/bin/aurad

# 4. Restore state (if database corrupted)
sudo systemctl stop aurad
rm -rf ~/.aura/data
tar -xzf ~/aura-backup-YYYYMMDD.tar.gz -C ~/

# 5. Restart with old binary
sudo systemctl start aurad

# 6. Monitor sync
journalctl -u aurad -f
aurad status 2>&1 | jq .SyncInfo

# 7. Verify consensus
curl http://localhost:26657/consensus_state

# 8. Check validator signing (validators only)
aurad query slashing signing-info $(aurad tendermint show-validator)
```

### Network-Wide Rollback

```bash
# Coordinated rollback requires >67% voting power

# 1. Core team coordination
#    - Announce rollback on validator channels
#    - Agree on rollback height
#    - Coordinate restart time

# 2. Export state at rollback height
aurad export --height <rollback-height> > genesis_rollback.json

# 3. Reset node
aurad unsafe-reset-all

# 4. Use rollback genesis
cp genesis_rollback.json ~/.aura/config/genesis.json

# 5. Install previous binary
sudo cp /usr/local/bin/aurad-v1.0.0 /usr/local/bin/aurad

# 6. Coordinated restart
# At agreed time:
sudo systemctl start aurad

# Network resumes from rollback height
```

---

## Testing on Testnet

### Testnet Upgrade Process

```bash
# 1. Setup testnet node
aurad init testnet-node --chain-id aura-testnet-1

# 2. Download testnet genesis
wget -O ~/.aura/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/testnet/genesis.json

# 3. Sync testnet
# Use state sync or snapshot for faster sync

# 4. Monitor testnet upgrade proposal
aurad query gov proposals --node https://rpc-testnet.aura.network:26657

# 5. Prepare upgrade binary (same as mainnet)
wget https://github.com/aequitas/aura/releases/download/v2.0.0/aurad-testnet

# 6. Perform upgrade
# Follow same procedure as mainnet

# 7. Document issues
# - Upgrade duration
# - Errors encountered
# - Resource usage
# - Breaking changes

# 8. Report to core team
# Share findings in #testnet channel
```

### Upgrade Testing Checklist

- [ ] Binary builds and runs
- [ ] Genesis migration successful
- [ ] State sync works
- [ ] API endpoints functional
- [ ] All modules operational
- [ ] Custom module compatibility
- [ ] Smart contract execution
- [ ] Transaction submission
- [ ] Query functionality
- [ ] Performance acceptable
- [ ] No memory leaks
- [ ] Prometheus metrics working

---

## Emergency Procedures

### Network Halt Recovery

```bash
# If network halts (consensus failure)

# 1. DO NOT restart node
# Wait for coordination

# 2. Join validator chat
# Discord #validators-emergency

# 3. Core team assessment
# - Identify halt cause
# - Determine recovery plan
# - Coordinate restart

# 4. Potential recovery methods:

# Method A: Simple restart (minor halt)
# Coordinated restart at specific time

# Method B: State export/import
aurad export --height <halt-height> > export.json
# Fix export.json if needed
cp export.json ~/.aura/config/genesis.json
aurad unsafe-reset-all
# Coordinated restart

# Method C: Emergency patch
# Apply hotfix binary
# Coordinated restart

# 5. Post-recovery
# - Monitor consensus
# - Verify all validators signing
# - Check for repeated halts
```

### Validator Tombstoned

```bash
# If validator tombstoned during upgrade:

# 1. Confirm tombstone status
aurad query slashing signing-info $(aurad tendermint show-validator)

# 2. Tombstoning is PERMANENT
# Cannot unjail this validator

# 3. Create new validator
# - Generate new validator key
# - Create new validator with different address
# - Communicate with delegators about migration

# 4. Prevent in future
# - Use Cosmovisor
# - Maintain hot standby
# - Test on testnet first
```

---

## Best Practices

### Before Every Upgrade

1. **Test on testnet** - No exceptions
2. **Backup everything** - Keys, config, state
3. **Read release notes** - Breaking changes, migration guides
4. **Verify checksums** - Don't run untrusted binaries
5. **Coordinate with team** - 24/7 coverage during upgrade
6. **Monitor channels** - Stay connected for coordination
7. **Prepare rollback** - Know how to revert quickly

### During Upgrade

1. **Monitor closely** - Watch logs continuously
2. **Check consensus** - Ensure network progressing
3. **Verify signing** - Validators should be active
4. **Test endpoints** - Confirm APIs working
5. **Watch alerts** - Monitor for anomalies

### After Upgrade

1. **Verify version** - Confirm running correct binary
2. **Test functionality** - Submit test transactions
3. **Monitor performance** - Check for degradation
4. **Review logs** - Look for errors or warnings
5. **Update documentation** - Record any issues/solutions
6. **Communicate status** - Inform delegators of success

---

## Upgrade Communication Template

```markdown
# AURA Mainnet Upgrade Notice

**Upgrade Name:** v2.0.0
**Upgrade Height:** 12,345,678
**Estimated Time:** 2025-12-01 14:00:00 UTC
**Expected Downtime:** 10-15 minutes

## What's New

- New module: Cross-chain bridge v2
- Performance improvements: 2x transaction throughput
- Bug fixes: See release notes

## Action Required

**For Validators:**
- Download binary: https://github.com/aequitas/aura/releases/tag/v2.0.0
- Verify checksum: sha256:abc123...
- Test on testnet first
- Prepare upgrade by height 12,345,000

**For Delegators:**
- No action required
- Brief downtime expected
- Rewards unaffected

## Resources

- Release Notes: https://github.com/aequitas/aura/releases/tag/v2.0.0
- Upgrade Guide: https://docs.aura.network/upgrades/v2.0.0
- Support: Discord #validators

---
Validator Team
security@yourvalidator.com
```

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25
