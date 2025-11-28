# AURA Validator Setup Guide

**Version:** 1.0
**Last Updated:** 2025-11-25
**Target Audience:** Validators, Staking Service Providers

---

## Table of Contents

1. [Introduction](#introduction)
2. [Validator Requirements](#validator-requirements)
3. [Key Generation and Security](#key-generation-and-security)
4. [Sentry Node Architecture](#sentry-node-architecture)
5. [Validator Setup](#validator-setup)
6. [Double-Sign Protection](#double-sign-protection)
7. [Monitoring and Alerting](#monitoring-and-alerting)
8. [Slashing Conditions](#slashing-conditions)
9. [Upgrade Procedures](#upgrade-procedures)
10. [Emergency Procedures](#emergency-procedures)
11. [Validator Operations](#validator-operations)

---

## Introduction

Running a validator on AURA is a critical responsibility that requires:
- **High Availability**: 99.9%+ uptime required to avoid slashing
- **Security**: Protection against key compromise and double-signing
- **Technical Expertise**: Understanding of blockchain operations
- **Financial Commitment**: Minimum stake + operational costs

This guide covers setting up and operating a production validator on the AURA blockchain.

### Validator Responsibilities

1. **Block Production**: Sign and propose blocks in consensus rotation
2. **Network Security**: Maintain network integrity and Byzantine Fault Tolerance
3. **Governance**: Participate in on-chain governance votes
4. **Upgrades**: Coordinate network upgrades
5. **Community**: Engage with delegators and community

### Validator Economics

- **Minimum Self-Delegation**: 1 AURA (subject to governance)
- **Commission**: Set your own (recommended 5-10%)
- **Rewards**: Block rewards + transaction fees distributed proportionally
- **Slashing**: 5% for double-sign, 0.01% for downtime

---

## Validator Requirements

### Hardware Requirements

**Minimum Production Spec:**
- **CPU**: 16 cores @ 3.5+ GHz (AMD EPYC 7002/7003 or Intel Xeon Gold)
- **RAM**: 64 GB ECC DDR4
- **Storage**: 2 TB NVMe SSD (enterprise grade, RAID 1)
- **Network**: 1 Gbps dedicated, < 50ms latency to majority of validators
- **Power**: Redundant PSU, UPS backup
- **Cooling**: Enterprise cooling solution

**Recommended Production Spec:**
- **CPU**: 32 cores @ 4.0+ GHz
- **RAM**: 128 GB ECC DDR4
- **Storage**: 4 TB NVMe SSD RAID 10
- **Network**: 10 Gbps with failover
- **Location**: Tier 3+ datacenter, geographically distributed sentries

### Infrastructure Requirements

1. **Primary Validator Node**: Private, highly secured
2. **Sentry Nodes**: Minimum 2, recommended 3+ in different locations
3. **Backup Validator**: Hot standby (critical for large validators)
4. **Monitoring**: Prometheus + Grafana + alerting
5. **Backup Power**: UPS + generator for extended outages

### Operational Requirements

- **24/7 Monitoring**: On-call support for critical alerts
- **Incident Response**: Documented procedures for common failures
- **Upgrade Testing**: Test all upgrades on testnet first
- **Security Audits**: Regular security reviews and penetration testing
- **Backup Procedures**: Automated daily backups with off-site storage

### Staking Requirements

```bash
# Minimum stake for validator creation
minimum_stake = 1000000 uaura  # 1 AURA

# Recommended stake for competitiveness
recommended_stake = 100000000000 uaura  # 100,000 AURA
```

### KYC/AML Requirements

**Note**: AURA includes a compliance module. Depending on network configuration:

- Validators may need to complete KYC verification
- AML compliance may be required for large validators
- Jurisdiction-specific regulations apply

Check current network parameters:
```bash
aurad query compliance params
```

---

## Key Generation and Security

### Generate Validator Keys (Air-Gapped Method)

**Step 1: Prepare Air-Gapped Machine**

```bash
# Use a clean, never-networked machine
# Install minimal OS (Ubuntu Server)
# Transfer aurad binary via USB

# Verify binary integrity
sha256sum aurad
# Compare with official release hash
```

**Step 2: Generate Keys**

```bash
# Initialize validator on air-gapped machine
aurad init validator-name --chain-id aura-mainnet-1

# The validator key is created at:
# ~/.aura/config/priv_validator_key.json

# Generate operator account
aurad keys add validator-operator --keyring-backend file

# CRITICAL: Write down mnemonic on paper
# Store in multiple secure locations (fireproof safe, bank vault)
```

**Step 3: Backup Keys**

```bash
# Encrypt validator key
gpg --symmetric --cipher-algo AES256 \
  ~/.aura/config/priv_validator_key.json

# Create multiple encrypted copies
# Store in geographically distributed locations:
# - Primary secure facility
# - Secondary datacenter
# - Bank safe deposit box

# Create recovery procedure document
```

**Step 4: Transfer Public Key**

```bash
# Export public key (safe to transfer)
aurad tendermint show-validator

# Transfer this to your production validator node
# via USB or secure channel
```

### Hardware Security Module (HSM) Setup

For enterprise validators, use HSM to protect validator keys:

**Supported HSMs:**
- YubiHSM 2
- Ledger Nano
- AWS CloudHSM
- Tendermint KMS with HSM backend

**Setup with Tendermint KMS:**

```bash
# Install Tendermint KMS
cargo install tmkms

# Initialize KMS configuration
tmkms init /etc/tmkms

# Configure for YubiHSM
# Edit /etc/tmkms/tmkms.toml
cat > /etc/tmkms/tmkms.toml <<EOF
[[chain]]
id = "aura-mainnet-1"
key_format = { type = "bech32", account_key_prefix = "aurapub", consensus_key_prefix = "auravalconspub" }

[[validator]]
addr = "tcp://10.0.1.1:26658"  # Your validator private IP
chain_id = "aura-mainnet-1"
reconnect = true
secret_key = "/etc/tmkms/secrets/kms-identity.key"

[[providers.yubihsm]]
adapter = { type = "usb" }
auth = { key = 1, password_file = "/etc/tmkms/secrets/password" }
keys = [
    { chain_ids = ["aura-mainnet-1"], key = 1 }
]
serial_number = "1234567890"
EOF

# Generate or import key to YubiHSM
tmkms yubihsm keys import -i 1 /path/to/priv_validator_key.json

# Start KMS
tmkms start -c /etc/tmkms/tmkms.toml
```

**Configure Validator for KMS:**

```toml
# In validator's config.toml
[priv_validator_laddr]
priv_validator_laddr = "tcp://0.0.0.0:26658"

# Remove or backup priv_validator_key.json
# Validator will now use KMS for signing
```

### Key Rotation Procedures

```bash
# Generate new validator key (on air-gapped machine)
# This requires governance proposal or validator update transaction

# 1. Generate new key
aurad init new-validator --chain-id aura-mainnet-1

# 2. Submit validator edit transaction
aurad tx staking edit-validator \
  --new-pubkey=$(aurad tendermint show-validator) \
  --from=validator-operator \
  --chain-id=aura-mainnet-1 \
  --gas=auto \
  --fees=10000uaura

# 3. Update all monitoring and backup procedures
# 4. Securely destroy old key after verification period
```

---

## Sentry Node Architecture

### Why Sentry Nodes?

Sentry nodes protect validators from:
- **DDoS attacks**: Absorb attack traffic
- **Connection exhaustion**: Filter connections
- **IP exposure**: Hide validator IP
- **Network partitions**: Maintain connectivity

### Recommended Architecture

```
                     Internet
                        │
        ┌───────────────┼───────────────┐
        │               │               │
   ┌────▼────┐    ┌────▼────┐    ┌────▼────┐
   │ Sentry  │    │ Sentry  │    │ Sentry  │
   │ US-East │    │ EU-West │    │ AP-East │
   │ Public  │    │ Public  │    │ Public  │
   └────┬────┘    └────┬────┘    └────┬────┘
        │              │              │
        │   Private VPN Network       │
        │              │              │
        └──────────────┼──────────────┘
                       │
                  ┌────▼────┐
                  │Validator│
                  │ Private │
                  │ (Hidden)│
                  └─────────┘
```

### Setup Sentry Node 1

```bash
# On sentry machine
aurad init sentry-us-east --chain-id aura-mainnet-1

# Download genesis
wget -O ~/.aura/config/genesis.json \
  https://raw.githubusercontent.com/aequitas/aura/main/networks/mainnet/genesis.json

# Configure config.toml
cat >> ~/.aura/config/config.toml <<EOF
[p2p]
# External address for peers to connect
external_address = "tcp://sentry1.yourdomain.com:26656"

# Enable peer exchange
pex = true

# Connect to validator (use private IP/VPN)
persistent_peers = "validator_node_id@10.0.1.1:26656"

# Protect validator identity
private_peer_ids = "validator_node_id"

# Unconditionally maintain validator connection
unconditional_peer_ids = "validator_node_id"

# Allow more connections for public node
max_num_inbound_peers = 100
max_num_outbound_peers = 50

# Seed mode (optional, for network bootstrapping)
seed_mode = false

# Address book
addr_book_strict = false
EOF

# Start sentry
sudo systemctl start aurad
```

### Setup Additional Sentries

```bash
# Repeat for each sentry in different geographic locations
# Update persistent_peers to include validator + other sentries

# Example for 3 sentries:
persistent_peers = "validator@10.0.1.1:26656,sentry1@10.0.2.1:26656,sentry2@10.0.3.1:26656"
```

### Validator Configuration for Sentries

```bash
# On validator node, configure to ONLY connect to sentries
# Edit ~/.aura/config/config.toml

[p2p]
# Listen on private interface
laddr = "tcp://10.0.1.1:26656"

# Disable peer exchange (don't discover peers)
pex = false

# ONLY connect to your sentry nodes
persistent_peers = "sentry1_id@10.0.2.1:26656,sentry2_id@10.0.3.1:26656,sentry3_id@10.0.4.1:26656"

# No private peers (sentries are trusted)
private_peer_ids = ""

# Don't advertise to peer exchange
addr_book_strict = true

# Minimal connections
max_num_inbound_peers = 10
max_num_outbound_peers = 10

# Disable RPC/API on validator (use sentries for RPC)
[rpc]
laddr = "tcp://127.0.0.1:26657"  # Localhost only

[api]
enable = false  # Or localhost only
```

### Firewall Configuration

**Validator Node:**
```bash
# ONLY allow connections from sentries and management
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow sentry connections
sudo ufw allow from 10.0.2.1 to any port 26656 proto tcp
sudo ufw allow from 10.0.3.1 to any port 26656 proto tcp
sudo ufw allow from 10.0.4.1 to any port 26656 proto tcp

# Management SSH (from specific IPs)
sudo ufw allow from YOUR_MGMT_IP to any port 22 proto tcp

sudo ufw enable
sudo ufw status numbered
```

**Sentry Nodes:**
```bash
# Allow P2P from anywhere
sudo ufw allow 26656/tcp

# Allow RPC/API if serving public requests
sudo ufw allow 26657/tcp
sudo ufw allow 1317/tcp
sudo ufw allow 9090/tcp

# Rate limit SSH
sudo ufw limit 22/tcp

sudo ufw enable
```

---

## Validator Setup

### Prerequisites Checklist

- [ ] Hardware meets requirements
- [ ] Sentry nodes deployed and synced
- [ ] Validator keys generated securely
- [ ] Operator account funded (for gas fees)
- [ ] Stake amount prepared
- [ ] Monitoring configured
- [ ] Backup procedures tested

### Step-by-Step Validator Creation

**Step 1: Verify Node is Synced**

```bash
# Check sync status
aurad status 2>&1 | jq .SyncInfo.catching_up

# Should return: false
# If true, wait for full sync
```

**Step 2: Verify Account Balance**

```bash
# Check operator account has sufficient funds
aurad query bank balances $(aurad keys show validator-operator -a)

# Need: stake amount + transaction fees
# Recommended: stake + 100 AURA for fees
```

**Step 3: Create Validator**

```bash
# Prepare validator metadata
export MONIKER="Your Validator Name"
export WEBSITE="https://yourvalidator.com"
export IDENTITY="keybase_identity"  # 16-digit keybase.io PGP fingerprint
export DETAILS="Professional validator service with 99.9% uptime"
export SECURITY_CONTACT="security@yourvalidator.com"
export COMMISSION_RATE="0.10"  # 10%
export COMMISSION_MAX_RATE="0.20"  # 20% max
export COMMISSION_MAX_CHANGE="0.01"  # 1% max change per day
export MIN_SELF_DELEGATION="1000000"  # 1 AURA minimum

# Create validator transaction
aurad tx staking create-validator \
  --amount=100000000000uaura \
  --pubkey=$(aurad tendermint show-validator) \
  --moniker="$MONIKER" \
  --website="$WEBSITE" \
  --identity="$IDENTITY" \
  --details="$DETAILS" \
  --security-contact="$SECURITY_CONTACT" \
  --chain-id="aura-mainnet-1" \
  --commission-rate="$COMMISSION_RATE" \
  --commission-max-rate="$COMMISSION_MAX_RATE" \
  --commission-max-change-rate="$COMMISSION_MAX_CHANGE" \
  --min-self-delegation="$MIN_SELF_DELEGATION" \
  --from=validator-operator \
  --gas=auto \
  --gas-adjustment=1.5 \
  --fees=10000uaura \
  --keyring-backend=file

# Transaction will be broadcast and confirmed
```

**Step 4: Verify Validator Creation**

```bash
# Get validator address
VALIDATOR_ADDR=$(aurad keys show validator-operator --bech val -a)

# Query validator info
aurad query staking validator $VALIDATOR_ADDR

# Check validator is in active set
aurad query staking validators | jq '.validators[] | select(.operator_address=="'$VALIDATOR_ADDR'")'

# Check validator signing info
aurad query slashing signing-info $(aurad tendermint show-validator)
```

**Step 5: Update Validator Profile**

Add validator logo and extended metadata:

```bash
# Upload logo to keybase.io
# Logo should be 256x256 PNG

# Update identity in validator metadata
aurad tx staking edit-validator \
  --identity="YOUR_KEYBASE_ID" \
  --from=validator-operator \
  --chain-id=aura-mainnet-1 \
  --fees=5000uaura
```

---

## Double-Sign Protection

Double-signing is the most severe validator offense, resulting in **5% stake slashing** and **tombstoning** (permanent jailing).

### What is Double-Signing?

Signing two different blocks at the same height, which can occur from:
- Running validator on multiple machines simultaneously
- Failed failover with both instances running
- Restored from old backup with outdated validator state

### Prevention Mechanisms

**1. Validator State File Protection**

```bash
# The file ~/.aura/data/priv_validator_state.json tracks:
# - height: last signed block height
# - round: last signed round
# - step: last signed step

# NEVER restore from old backup without clearing this
# NEVER run multiple instances with same key
```

**2. Sentinel Monitoring**

```bash
# Monitor for double-sign evidence
# Add alert to Prometheus:

- alert: DoubleSignDetected
  expr: increase(aura_validator_double_sign_total[5m]) > 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "DOUBLE SIGN DETECTED - IMMEDIATE ACTION REQUIRED"
```

**3. Atomic Failover Procedures**

```bash
#!/bin/bash
# safe-failover.sh

PRIMARY_NODE="10.0.1.1"
STANDBY_NODE="10.0.1.2"

# 1. Stop primary validator (ensure it's fully stopped)
ssh $PRIMARY_NODE "sudo systemctl stop aurad && sleep 10"

# 2. Verify primary is stopped
ssh $PRIMARY_NODE "pgrep aurad" && echo "ERROR: Primary still running!" && exit 1

# 3. Copy latest validator state from primary
scp $PRIMARY_NODE:~/.aura/data/priv_validator_state.json \
    $STANDBY_NODE:~/.aura/data/priv_validator_state.json

# 4. Start standby
ssh $STANDBY_NODE "sudo systemctl start aurad"

# 5. Verify standby is signing
ssh $STANDBY_NODE "journalctl -u aurad -f | grep 'Signed block'"
```

**4. Tendermint KMS with HSM**

Using KMS prevents double-signing through cryptographic enforcement:

```bash
# KMS maintains state and rejects duplicate signing requests
# Even if multiple validators connect, only one can sign
```

### Recovery from Double-Sign

If double-sign occurs:

```bash
# 1. IMMEDIATELY stop all validator instances
sudo systemctl stop aurad

# 2. Check if validator is tombstoned
aurad query slashing signing-info $(aurad tendermint show-validator)

# 3. If tombstoned, validator cannot be unjailed
# Must create new validator with different key and address

# 4. Communicate with delegators about migration plan
# 5. Review incident and implement additional safeguards
```

---

## Monitoring and Alerting

### Critical Metrics to Monitor

**1. Validator Health**
```bash
# Missed blocks
aura_validatorsecurity_missed_blocks

# Signing percentage
aura_validatorsecurity_signing_percentage

# Validator jailed status
aura_staking_validator_jailed
```

**2. Node Health**
```bash
# Block sync status
aura_latest_block_height - aura_network_height < 10

# Peer connections
aura_p2p_peers >= 10

# Memory usage
process_resident_memory_bytes < threshold
```

**3. Security Events**
```bash
# Double sign attempts
aura_validator_double_sign_total

# Failed authentication
failed_ssh_attempts

# Unusual activity
aura_monitoring_anomaly_detections_total
```

### Prometheus Alert Rules

```yaml
# /etc/prometheus/validator-alerts.yml

groups:
  - name: validator_critical
    interval: 10s
    rules:
      - alert: ValidatorNotSigning
        expr: increase(aura_validatorsecurity_missed_blocks[5m]) > 10
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Validator missing blocks - URGENT"
          description: "Missed {{ $value }} blocks in 5 minutes"

      - alert: ValidatorJailed
        expr: aura_staking_validator_jailed == 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "VALIDATOR JAILED"
          description: "Validator has been jailed and is not signing"

      - alert: ValidatorLowSigningPercentage
        expr: aura_validatorsecurity_signing_percentage < 95
        for: 10m
        labels:
          severity: high
        annotations:
          summary: "Low signing percentage"
          description: "Signing percentage: {{ $value }}%"

      - alert: LowPeerCount
        expr: aura_p2p_peers < 5
        for: 5m
        labels:
          severity: high
        annotations:
          summary: "Low peer connections"
          description: "Only {{ $value }} peers connected"

      - alert: BlockHeightStalled
        expr: rate(aura_latest_block_height[5m]) == 0
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "Block height not increasing"
          description: "Node may be stuck or disconnected"
```

### Grafana Dashboard

Import the validator dashboard from `/grafana/dashboards/validator-monitoring.json`:

**Key Panels:**
1. Validator uptime percentage
2. Missed blocks (last 24h)
3. Signing percentage trend
4. Voting power chart
5. Commission earnings
6. Delegator count
7. Network rank

### Alert Escalation

**Level 1: Warning** (5min response)
- Low peer count
- High memory usage
- Moderate missed blocks (< 5)

**Level 2: High** (2min response)
- Low signing percentage
- Validator close to jailing
- Sentry node down

**Level 3: Critical** (Immediate)
- Validator missing blocks
- Validator jailed
- Double-sign detected
- All sentries down

### Notification Channels

```bash
# Configure Alertmanager
cat > /etc/prometheus/alertmanager.yml <<EOF
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'validator-team'
  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'
      continue: true

receivers:
  - name: 'validator-team'
    email_configs:
      - to: 'validators@yourcompany.com'
    slack_configs:
      - channel: '#validator-alerts'
        api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK'

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_KEY'
EOF
```

---

## Slashing Conditions

### Downtime Slashing

**Conditions:**
- Missing blocks for extended period
- Threshold: Missing > 50% of blocks in signed_blocks_window
- Default window: 10,000 blocks

**Penalties:**
- Slashing: 0.01% of staked amount
- Jailing: Removed from active validator set
- Unjailing: Manual unjail transaction required

**Check Parameters:**
```bash
aurad query slashing params

# Returns:
# signed_blocks_window: 10000
# min_signed_per_window: 0.5
# downtime_jail_duration: 600s
# slash_fraction_downtime: 0.0001
```

**Monitoring:**
```bash
# Check current signing window status
aurad query slashing signing-info $(aurad tendermint show-validator)

# Returns:
# - index_offset: current position in window
# - missed_blocks_counter: missed blocks count
# - jailed_until: jail expiration (if jailed)
```

**Unjailing Procedure:**
```bash
# After fixing downtime issue:

# 1. Verify node is synced and healthy
aurad status

# 2. Wait for jail period to expire
aurad query slashing signing-info $(aurad tendermint show-validator) | grep jailed_until

# 3. Submit unjail transaction
aurad tx slashing unjail \
  --from=validator-operator \
  --chain-id=aura-mainnet-1 \
  --gas=auto \
  --fees=5000uaura

# 4. Verify validator is active
aurad query staking validator $VALIDATOR_ADDR
```

### Double-Sign Slashing

**Conditions:**
- Signing two different blocks at same height
- Detected by network consensus

**Penalties:**
- Slashing: 5% of staked amount
- Tombstoning: Permanent jailing (cannot unjail)
- Reputation: Severe damage to validator reputation

**Prevention:**
- Never run same validator key on multiple machines
- Implement atomic failover procedures
- Use Tendermint KMS with HSM
- Maintain validator state backups carefully

**Parameters:**
```bash
aurad query slashing params | grep double

# slash_fraction_double_sign: 0.05  # 5%
```

### Slashing Protection Checklist

- [ ] Monitoring alerts for missed blocks configured
- [ ] Redundant infrastructure (sentries, backup validator)
- [ ] Automated failover tested
- [ ] Validator state backup procedures documented
- [ ] HSM or KMS implemented for key security
- [ ] 24/7 on-call rotation established
- [ ] Incident response playbook created

---

## Upgrade Procedures

See [UPGRADE_PROCEDURES.md](./UPGRADE_PROCEDURES.md) for detailed upgrade instructions.

### Pre-Upgrade Checklist

- [ ] Upgrade tested on testnet
- [ ] Backup created (keys, config, state)
- [ ] Upgrade block height confirmed
- [ ] Binary downloaded and verified
- [ ] Monitoring alerts configured
- [ ] Communication plan (delegators, team)
- [ ] Rollback plan prepared

### Coordinated Upgrade Process

```bash
# 1. Download new binary
wget https://github.com/aequitas/aura/releases/download/v2.0.0/aurad
chmod +x aurad
sudo mv aurad /usr/local/bin/

# 2. Verify binary
aurad version
sha256sum /usr/local/bin/aurad

# 3. Stop node at upgrade height
# Option A: Manual stop
sudo systemctl stop aurad

# Option B: Cosmovisor (automated)
# Cosmovisor will automatically swap binaries at upgrade height

# 4. Start with new binary
sudo systemctl start aurad

# 5. Monitor upgrade
sudo journalctl -u aurad -f

# 6. Verify validator signing
aurad query slashing signing-info $(aurad tendermint show-validator)
```

---

## Emergency Procedures

### Emergency Contact List

Maintain updated contact list:
```
- Primary Operator: [Name, Phone, Email]
- Secondary Operator: [Name, Phone, Email]
- On-Call Rotation: [Schedule]
- Datacenter Support: [Phone, Ticket System]
- AURA Core Team: security@aura.network
- Validator Chat: Discord #validators
```

### Common Emergencies

#### Validator Down / Not Signing

```bash
# Quick diagnostic
ssh validator "aurad status && systemctl status aurad"

# Check logs
ssh validator "journalctl -u aurad -n 100 --no-pager"

# Common fixes:
# 1. Restart service
sudo systemctl restart aurad

# 2. Check disk space
df -h

# 3. Check memory
free -h

# 4. Check peer connections
curl localhost:26657/net_info | jq .result.n_peers

# 5. If stuck, try unsafe reset (WARNING: resyncs from network)
# Only if you have < 30min before jailing
sudo systemctl stop aurad
aurad unsafe-reset-all
# Download latest snapshot
# Restart
```

#### Sentry Node Compromised

```bash
# 1. Immediately disconnect sentry from validator
ssh validator "aurad tendermint unsafe-reset-all --keep-addr-book=false"
# Remove compromised sentry from persistent_peers

# 2. Investigate compromise
# - Check for unauthorized access
# - Review logs
# - Scan for malware

# 3. Rebuild sentry from clean image
# 4. Restore from verified backup
# 5. Re-add to validator after security review
```

#### Validator Key Potentially Compromised

```bash
# CRITICAL - IMMEDIATE ACTION

# 1. Stop validator immediately
sudo systemctl stop aurad

# 2. Notify AURA security team
# Email: security@aura.network

# 3. Generate new validator key (air-gapped)

# 4. Submit validator update (if not tombstoned)
aurad tx staking edit-validator \
  --new-pubkey=$(aurad tendermint show-validator) \
  --from=validator-operator

# 5. Forensic investigation
# 6. Security audit before re-activation
```

#### Network Halt / Fork Detection

```bash
# Network consensus halt detected

# 1. DO NOT restart node
# 2. Join validator coordination channel
# 3. Wait for official guidance from core team
# 4. Prepare for coordinated restart at specific height
# 5. May require genesis export and network relaunch
```

### Emergency Runbook

Document your specific procedures:

```markdown
# Validator Emergency Runbook

## Infrastructure
- Validator IP: 10.0.1.1
- Sentry IPs: 10.0.2.1, 10.0.3.1, 10.0.4.1
- Backup Validator IP: 10.0.5.1

## Access
- SSH Keys: [Location]
- Sudo Password: [Secure Location]
- HSM Access: [Procedure]

## Emergency Procedures
1. Missed blocks > 10: [Steps]
2. All sentries down: [Steps]
3. Validator hardware failure: [Steps]
4. Network upgrade emergency: [Steps]

## Escalation
1. Primary on-call: [Contact]
2. Secondary on-call: [Contact]
3. Datacenter support: [Contact]
```

---

## Validator Operations

### Daily Operations

```bash
# Morning checks (automated via monitoring)
aurad status
aurad query staking validator $VALIDATOR_ADDR
aurad query slashing signing-info $(aurad tendermint show-validator)

# Check for missed blocks
curl localhost:26657/consensus_state | jq

# Review alerts (Grafana)
# Check disk space growth
# Verify backups completed
```

### Weekly Operations

- Review validator performance metrics
- Check for upcoming network upgrades
- Test backup restore procedure
- Security log review
- Update dependencies (OS patches)
- Delegator communication

### Monthly Operations

- Full security audit
- Disaster recovery drill
- Performance optimization review
- Hardware health check
- Update documentation
- Financial review (rewards, costs)

### Quarterly Operations

- Major security audit
- Infrastructure capacity planning
- Delegator survey
- Marketing/branding update
- Contract renewal (datacenter, services)
- Long-term strategy review

---

## Best Practices Summary

1. **Security First**: Use HSM, multi-sig, geographic distribution
2. **Monitoring**: 24/7 alerting on critical metrics
3. **Documentation**: Maintain runbooks and procedures
4. **Testing**: Test all procedures on testnet first
5. **Communication**: Keep delegators informed
6. **Redundancy**: Multiple sentries, backup validator
7. **Compliance**: Follow network governance decisions
8. **Community**: Participate in validator discussions
9. **Economics**: Balance commission vs. competitiveness
10. **Continuous Improvement**: Regular reviews and updates

---

**Document Status**: Production Ready
**Review Cycle**: Quarterly
**Next Review**: 2026-02-25
