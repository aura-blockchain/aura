# AURA Network Upgrade Procedures

This runbook covers planned software upgrades for the AURA blockchain network.

## Overview

AURA uses Cosmos SDK's governance module for coordinated network upgrades. All upgrades go through governance voting before execution.

## Upgrade Types

| Type | Description | Requires |
|------|-------------|----------|
| **Soft Fork** | Non-breaking changes, backward compatible | Governance vote |
| **Hard Fork** | Breaking changes, all validators must upgrade | Governance vote + coordination |
| **Emergency** | Security patches, rapid deployment | Emergency multisig (if enabled) |

## Pre-Upgrade Checklist

### 1. Announcement (T-14 days)
- [ ] Publish upgrade proposal RFC
- [ ] Announce in Discord #announcements
- [ ] Notify validator operators via email
- [ ] Update documentation with upgrade notes

### 2. Governance Proposal (T-14 days)
```bash
# Submit upgrade proposal
aurad tx gov submit-proposal software-upgrade v2.0.0 \
  --title "Upgrade to v2.0.0" \
  --description "Release notes: https://github.com/aequitas/aura/releases/tag/v2.0.0" \
  --upgrade-height 5000000 \
  --upgrade-info '{"binaries":{"linux/amd64":"https://github.com/aequitas/aura/releases/download/v2.0.0/aurad-v2.0.0-linux-amd64"}}' \
  --deposit 10000000000uaura \
  --from governance-multisig \
  --chain-id aura-mainnet-1
```

### 3. Voting Period (14 days)
- [ ] Monitor vote progress
- [ ] Address validator questions
- [ ] Ensure quorum (40%) and threshold (50%) met

### 4. Binary Distribution (T-3 days before upgrade height)
- [ ] Publish release with verified checksums
- [ ] Test binary on testnet
- [ ] Distribute to validators

## Upgrade Execution

### Using Cosmovisor (Recommended)

Cosmovisor handles automatic binary switching at upgrade height.

#### Setup Cosmovisor

```bash
# Install Cosmovisor
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Set environment variables
export DAEMON_NAME=aurad
export DAEMON_HOME=$HOME/.aura
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
export DAEMON_RESTART_AFTER_UPGRADE=true

# Initialize
cosmovisor init $(which aurad)

# Add upgrade binary
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v2.0.0/bin
cp aurad-v2.0.0 $DAEMON_HOME/cosmovisor/upgrades/v2.0.0/bin/aurad
chmod +x $DAEMON_HOME/cosmovisor/upgrades/v2.0.0/bin/aurad
```

#### Upgrade Process

1. **Before upgrade height:**
   - Cosmovisor detects upgrade plan from on-chain
   - Verifies new binary exists

2. **At upgrade height:**
   - Chain halts automatically
   - Cosmovisor switches to new binary
   - Node restarts automatically

3. **After upgrade:**
   - Verify node is running new version
   - Check consensus participation

### Manual Upgrade

For operators not using Cosmovisor:

```bash
# 1. Monitor for upgrade height
watch -n 1 'aurad status | jq ".SyncInfo.latest_block_height"'

# 2. At upgrade height, chain halts
# You'll see: "UPGRADE NEEDED at height: 5000000"

# 3. Stop the node
sudo systemctl stop aurad

# 4. Backup current binary
cp $(which aurad) ~/aurad-v1.x.x-backup

# 5. Install new binary
wget https://github.com/aequitas/aura/releases/download/v2.0.0/aurad-v2.0.0-linux-amd64
chmod +x aurad-v2.0.0-linux-amd64
sudo mv aurad-v2.0.0-linux-amd64 /usr/local/bin/aurad

# 6. Verify version
aurad version

# 7. Start node
sudo systemctl start aurad

# 8. Monitor logs
journalctl -u aurad -f
```

## Post-Upgrade Verification

### Immediate Checks

```bash
# Version check
aurad version
# Expected: v2.0.0

# Node status
aurad status | jq '.SyncInfo'
# Verify: catching_up = false, latest_block_height increasing

# Consensus participation (validators)
aurad query slashing signing-info $(aurad tendermint show-validator)
# Verify: not jailed, missed_blocks_counter low
```

### Network Health

```bash
# Peer count
curl -s localhost:26657/net_info | jq '.result.n_peers'
# Expected: > 10

# Consensus state
curl -s localhost:26657/consensus_state | jq '.result.round_state.height_vote_set'
# Verify: validators signing
```

## Rollback Procedures

### If Upgrade Fails Before Consensus

```bash
# 1. Stop node
sudo systemctl stop aurad

# 2. Restore old binary
sudo mv ~/aurad-v1.x.x-backup /usr/local/bin/aurad

# 3. Clear upgrade info (if needed)
aurad rollback

# 4. Restart
sudo systemctl start aurad
```

### If Network Needs Emergency Rollback

Requires coordination with > 2/3 validators:

1. **Halt network** at agreed block height
2. **Export state** at last good height
3. **Create new genesis** from exported state
4. **Coordinate restart** with validators

## Emergency Upgrade

For critical security patches:

### Using Emergency Multisig (if enabled)

```bash
# Submit emergency upgrade (no voting period)
aurad tx gov submit-proposal software-upgrade emergency-v1.0.1 \
  --title "Emergency Security Patch" \
  --description "Critical vulnerability fix" \
  --upgrade-height $(aurad status | jq -r '.SyncInfo.latest_block_height' | awk '{print $1+100}') \
  --no-validate \
  --from emergency-multisig
```

### Manual Emergency Procedure

1. **Notify validators** via Discord/Telegram
2. **Publish patch** with clear instructions
3. **Coordinate halt** if needed
4. **Execute upgrade** as fast as possible
5. **Verify network** resumes

## Module-Specific Migrations

### State Migrations

New module versions may require state migrations:

```go
// Example migration in upgrade handler
app.UpgradeKeeper.SetUpgradeHandler(
    "v2.0.0",
    func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
        // Run module migrations
        return app.ModuleManager.RunMigrations(ctx, app.configurator, fromVM)
    },
)
```

### Store Migrations

Check for store key changes:

```bash
# Before upgrade, backup stores
cp -r ~/.aura/data ~/.aura/data-backup-v1

# If migration fails, restore
rm -rf ~/.aura/data
cp -r ~/.aura/data-backup-v1 ~/.aura/data
```

## Communication Templates

### Upgrade Announcement

```
📢 AURA Network Upgrade Announcement

Version: v2.0.0
Upgrade Height: 5,000,000 (estimated: 2025-03-15 14:00 UTC)

Changes:
- [Feature 1]
- [Feature 2]
- [Bug fix]

Action Required:
1. Download new binary before upgrade height
2. Use Cosmovisor for automatic upgrade OR
3. Manual upgrade at halt

Resources:
- Release: https://github.com/aequitas/aura/releases/tag/v2.0.0
- Docs: https://docs.aura.network/upgrades/v2.0.0

Questions? Ask in #validator-support
```

### Upgrade Success

```
✅ AURA Network Upgrade Complete

The network has successfully upgraded to v2.0.0.
- Block height resumed: 5,000,001
- Validators signing: 98/100
- Network healthy

Thank you to all validators for smooth coordination!
```

## Troubleshooting

### Node Won't Start After Upgrade

```bash
# Check logs
journalctl -u aurad -n 100 --no-pager

# Common fixes:
# 1. Wrong binary version
aurad version  # Verify correct version

# 2. State migration failed
aurad rollback  # Try rollback

# 3. Consensus mismatch
# Wait for other validators or check genesis
```

### Validator Jailed After Upgrade

```bash
# Check jail reason
aurad query slashing signing-info $(aurad tendermint show-validator)

# If jailed for downtime during upgrade, unjail
aurad tx slashing unjail --from validator --chain-id aura-mainnet-1

# Wait for unbonding before unjailing (usually immediate after upgrade)
```

### AppHash Mismatch

```bash
# CRITICAL: AppHash mismatch means state divergence
# DO NOT continue - coordinate with team

# 1. Stop node
sudo systemctl stop aurad

# 2. Report in #validator-support with:
#    - Block height
#    - App hash
#    - aurad version

# 3. Wait for instructions
```

## Contacts

- **Upgrade Coordinator:** upgrades@aura.network
- **Discord:** #validator-support
- **Emergency:** emergency@aura.network
