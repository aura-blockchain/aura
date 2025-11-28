# AURA Validator Onboarding Guide

Welcome to the AURA validator community! This guide will help you set up and operate a validator node on the AURA blockchain.

## Overview

Validators are responsible for:
- Proposing and validating blocks
- Participating in consensus
- Securing the network through staked AURA tokens
- Running AI assistants for Proof-of-Identity verification (optional)

## Requirements

### Minimum Hardware

| Component | Specification |
|-----------|---------------|
| CPU | 4 cores, 2.5 GHz+ |
| RAM | 16 GB |
| Storage | 500 GB NVMe SSD |
| Network | 100 Mbps, static IP |

### Recommended Hardware

| Component | Specification |
|-----------|---------------|
| CPU | 8+ cores, 3.0 GHz+ |
| RAM | 32 GB |
| Storage | 1 TB NVMe SSD |
| Network | 1 Gbps, static IP |
| HSM | YubiHSM 2 or equivalent |

### Staking Requirements

| Parameter | Value |
|-----------|-------|
| Minimum self-delegation | 1,000,000 AURA |
| Recommended self-delegation | 10,000,000+ AURA |
| Minimum commission | 5% |
| Unbonding period | 21 days |

## Quick Start

### Step 1: Prepare Server

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y build-essential git curl jq lz4

# Install Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Step 2: Build AURA

```bash
# Clone repository
git clone https://github.com/aequitas/aura.git
cd aura/chain

# Build binary
go build -o aurad ./cmd/aurad

# Install globally
sudo mv aurad /usr/local/bin/

# Verify
aurad version
```

### Step 3: Initialize Node

```bash
# Set variables
export CHAIN_ID="aura-mainnet-1"
export MONIKER="your-validator-name"

# Initialize
aurad init $MONIKER --chain-id $CHAIN_ID

# Download genesis
curl -o ~/.aura/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# Validate genesis
aurad validate-genesis
```

### Step 4: Configure Node

Edit `~/.aura/config/config.toml`:

```toml
[p2p]
seeds = "seed1.aura.network:26656,seed2.aura.network:26656,seed3.aura.network:26656"
```

Edit `~/.aura/config/app.toml`:

```toml
minimum-gas-prices = "0.025uaura"
```

### Step 5: Start Node

```bash
# Create service file
sudo tee /etc/systemd/system/aurad.service > /dev/null << EOF
[Unit]
Description=AURA Node
After=network-online.target

[Service]
User=$USER
ExecStart=/usr/local/bin/aurad start
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl daemon-reload
sudo systemctl enable aurad
sudo systemctl start aurad

# Check logs
journalctl -u aurad -f
```

### Step 6: Wait for Sync

Monitor sync progress:

```bash
# Check current height
aurad status | jq '.SyncInfo.latest_block_height'

# Check if catching up
aurad status | jq '.SyncInfo.catching_up'
# Should be "false" when synced
```

### Step 7: Create Validator

Once synced, create your validator:

```bash
# Create key (save mnemonic securely!)
aurad keys add validator --keyring-backend file

# Get your address
aurad keys show validator -a

# Fund your address with AURA tokens
# (transfer from exchange or other wallet)

# Check balance
aurad query bank balances $(aurad keys show validator -a)

# Create validator
aurad tx staking create-validator \
  --amount=1000000000000uaura \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="$MONIKER" \
  --website="https://your-website.com" \
  --identity="YOUR_KEYBASE_ID" \
  --details="Your validator description" \
  --security-contact="security@your-domain.com" \
  --chain-id=$CHAIN_ID \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000000000" \
  --gas="auto" \
  --gas-adjustment="1.5" \
  --gas-prices="0.025uaura" \
  --from=validator \
  --keyring-backend=file
```

### Step 8: Verify Validator

```bash
# Check validator status
aurad query staking validator $(aurad keys show validator --bech val -a)

# View in explorer
echo "https://explorer.aura.network/validators/$(aurad keys show validator --bech val -a)"
```

## Validator Operations

### Edit Validator Info

```bash
aurad tx staking edit-validator \
  --moniker="New Name" \
  --website="https://new-website.com" \
  --details="Updated description" \
  --commission-rate="0.12" \
  --from=validator \
  --chain-id=$CHAIN_ID
```

### Delegate More Tokens

```bash
aurad tx staking delegate \
  $(aurad keys show validator --bech val -a) \
  1000000000000uaura \
  --from=validator \
  --chain-id=$CHAIN_ID
```

### Withdraw Rewards

```bash
# Withdraw commission + rewards
aurad tx distribution withdraw-rewards \
  $(aurad keys show validator --bech val -a) \
  --commission \
  --from=validator \
  --chain-id=$CHAIN_ID
```

### Unjail Validator

If jailed for downtime:

```bash
aurad tx slashing unjail \
  --from=validator \
  --chain-id=$CHAIN_ID
```

## Monitoring

### Essential Checks

```bash
# Node status
aurad status | jq

# Validator signing info
aurad query slashing signing-info $(aurad tendermint show-validator)

# Check missed blocks
aurad query slashing signing-info $(aurad tendermint show-validator) | jq '.missed_blocks_counter'

# Peer count
curl -s localhost:26657/net_info | jq '.result.n_peers'
```

### Prometheus Metrics

Enable in `~/.aura/config/config.toml`:

```toml
[instrumentation]
prometheus = true
prometheus_listen_addr = ":26660"
```

Key metrics:
- `tendermint_consensus_height` - Current block height
- `tendermint_consensus_validators` - Total validators
- `tendermint_consensus_missing_validators` - Validators not signing
- `tendermint_p2p_peers` - Connected peers

### Recommended Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| Missed Blocks | > 50 in 1 hour | Warning |
| Missed Blocks | > 100 in 1 hour | Critical |
| Node Down | Unreachable > 1 min | Critical |
| Peer Count Low | < 5 peers | Warning |
| Disk Space | < 20% free | Warning |

## Security Best Practices

### 1. Key Management

- **NEVER** store validator keys on hot servers
- Use [HSM](/docs/security/HSM_INTEGRATION.md) for production
- Backup keys securely (encrypted, offline)
- Use separate keys for validator vs. operator accounts

### 2. Network Security

```bash
# Firewall rules
sudo ufw default deny incoming
sudo ufw allow ssh
sudo ufw allow 26656/tcp  # P2P
sudo ufw enable
```

- Run RPC/API on internal network only
- Use sentry nodes for DDoS protection
- Enable fail2ban for SSH

### 3. Server Security

- Disable root login
- Use SSH keys only
- Keep system updated
- Enable automatic security updates
- Run aurad as non-root user

### 4. Operational Security

- Monitor 24/7
- Have backup servers ready
- Document recovery procedures
- Test restores regularly
- Join validator communication channels

## Slashing Conditions

| Offense | Penalty | Jail Duration |
|---------|---------|---------------|
| Double signing | 5% slash, tombstone | Permanent |
| Downtime (95% missed in 10,000 blocks) | 0.01% slash | 10 minutes |

### Avoiding Slashing

1. **Never run two validators with same key**
2. Use HSM to prevent key extraction
3. Monitor uptime continuously
4. Have redundant infrastructure
5. Test failover procedures

## Governance Participation

Validators should actively participate in governance:

```bash
# View proposals
aurad query gov proposals

# Vote on proposal
aurad tx gov vote 1 yes \
  --from=validator \
  --chain-id=$CHAIN_ID
```

Vote options: `yes`, `no`, `abstain`, `no_with_veto`

## AI Assistant Integration (Optional)

Validators can run AI assistants for PoI verification to earn additional rewards.

### Requirements

- Additional compute for AI models
- GPU recommended (NVIDIA RTX 3080+)
- Minimum 50 GB additional storage

### Setup

See [AI Assistant Operator Guide](/docs/modules/aiassistant/OPERATOR_GUIDE.md).

## Upgrades

### Automatic (Cosmovisor)

```bash
# Install Cosmovisor
go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

# Setup
export DAEMON_NAME=aurad
export DAEMON_HOME=$HOME/.aura
cosmovisor init /usr/local/bin/aurad

# Run with Cosmovisor
cosmovisor run start
```

### Manual

1. Monitor governance for upgrade proposals
2. Download new binary before upgrade height
3. Stop node at upgrade height
4. Replace binary
5. Restart node

## Support & Community

### Resources

- **Documentation:** https://docs.aura.network
- **Explorer:** https://explorer.aura.network
- **GitHub:** https://github.com/aequitas/aura

### Communication

- **Discord:** discord.aura.network (validator channel)
- **Telegram:** t.me/auravalidators
- **Forum:** forum.aura.network

### Getting Help

1. Check documentation first
2. Search Discord/forum for similar issues
3. Ask in #validator-support channel
4. For security issues: security@aura.network

## Checklist

Before going live, ensure:

- [ ] Server meets hardware requirements
- [ ] Node fully synced
- [ ] Keys backed up securely
- [ ] HSM configured (production)
- [ ] Firewall configured
- [ ] Monitoring setup
- [ ] Alerts configured
- [ ] Backup procedures documented
- [ ] Recovery tested
- [ ] Joined validator channels
- [ ] Read slashing conditions
- [ ] Understand governance responsibilities

## FAQ

**Q: How long does initial sync take?**
A: Full sync: 2-7 days. State sync: 1-2 hours. Snapshot: 30-60 minutes.

**Q: What's the minimum stake?**
A: 1,000,000 AURA minimum self-delegation. To be in active set, you need enough total stake to be in top 100.

**Q: How often should I check my validator?**
A: Automated monitoring 24/7. Manual check at least daily.

**Q: Can I run multiple validators?**
A: Yes, but each needs separate keys and infrastructure. Never share keys.

**Q: What happens if I miss blocks?**
A: Missing 95%+ of blocks in 10,000 block window results in 0.01% slash and 10-minute jail.

**Q: How do I migrate to new hardware?**
A: Stop old node → Transfer priv_validator_state.json → Start new node. NEVER run both simultaneously.
