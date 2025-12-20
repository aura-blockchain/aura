---
sidebar_position: 3
---

# Chain Upgrades

Handle chain upgrades safely and efficiently with Cosmovisor.

## Upgrade Process

### 1. Monitor Governance

```bash
# List active proposals
aurad query gov proposals --status voting_period

# Check upgrade proposal details
aurad query gov proposal <proposal-id>

# Vote on upgrade proposal
aurad tx gov vote <proposal-id> yes \
  --from validator \
  --chain-id aura-mainnet-1 \
  --gas auto
```

### 2. Prepare for Upgrade

When an upgrade is approved:

```bash
# Note upgrade height from proposal
UPGRADE_HEIGHT=1234567
UPGRADE_NAME="v2.0.0"

# Build new binary
git clone https://github.com/aura-blockchain/aura.git
cd aura
git checkout $UPGRADE_NAME
cd chain
make install

# Verify new version
aurad version
```

### 3. Manual Upgrade

If not using Cosmovisor:

```bash
# Stop node at upgrade height
sudo systemctl stop aurad

# Backup state
cp -r ~/.aura/data ~/.aura/data-backup-$UPGRADE_HEIGHT

# Install new binary
make install

# Run migration (if required)
aurad migrate $UPGRADE_NAME

# Start node
sudo systemctl start aurad
```

### 4. Automated Upgrade with Cosmovisor

```bash
# Create upgrade directory
mkdir -p ~/.aura/cosmovisor/upgrades/$UPGRADE_NAME/bin

# Copy new binary
cp $(which aurad) ~/.aura/cosmovisor/upgrades/$UPGRADE_NAME/bin/

# Cosmovisor will automatically upgrade at specified height
# No manual intervention needed!
```

## Rollback Procedure

If upgrade fails:

```bash
# Stop node
sudo systemctl stop aurad

# Restore backup
rm -rf ~/.aura/data
cp -r ~/.aura/data-backup-$UPGRADE_HEIGHT ~/.aura/data

# Revert to old binary
git checkout <previous-version>
make install

# Start node
sudo systemctl start aurad
```

## Testing Upgrades

### Test on Testnet First

```bash
# Connect to testnet
aurad init test-validator --chain-id aura-testnet-1

# Download testnet genesis
curl https://testnet.aura.network/genesis.json > ~/.aura/config/genesis.json

# Test upgrade process
```

## Best Practices

- Always test upgrades on testnet first
- Backup node state before upgrades
- Monitor upgrade proposals closely
- Have rollback plan ready
- Coordinate with other validators
- Join validator Discord for upgrade coordination

## Emergency Procedures

### Halt Chain

In case of critical issues:

```bash
# Validators can halt chain via governance emergency proposal
aurad tx gov submit-proposal software-upgrade emergency-halt \
  --title "Emergency Halt" \
  --description "Critical bug detected" \
  --upgrade-height <current-height+100> \
  --from validator
```

## Resources

- [Upgrade Documentation](https://github.com/aura-blockchain/aura/blob/main/docs/ops/UPGRADE_PROCEDURES.md)
- [Cosmovisor Guide](https://docs.cosmos.network/main/tooling/cosmovisor)
- [Upgrade Coordination Discord](https://discord.gg/aura)
