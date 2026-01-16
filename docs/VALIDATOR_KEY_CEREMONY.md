# AURA MVP Validator Key Ceremony

**Version**: 1.0
**Last Updated**: 2026-01-13
**Chain ID**: aura-mvp-1

## Overview

This document describes the secure process for generating and distributing validator keys for the AURA MVP testnet launch. The ceremony ensures cryptographic integrity and secure key handling.

## Prerequisites

- [ ] All validator operators identified and confirmed
- [ ] Secure communication channel established (Signal/encrypted)
- [ ] Each operator has secure, airgapped machine available
- [ ] Genesis coordinator designated
- [ ] Backup procedures reviewed

## Validator Allocation

| Validator | Operator | Server | Voting Power |
|-----------|----------|--------|--------------|
| val1 | AURA Foundation | aura-testnet | 25% |
| val2 | AURA Foundation | aura-testnet | 25% |
| val3 | AURA Foundation | services-testnet | 25% |
| val4 | AURA Foundation | services-testnet | 25% |

---

## Phase 1: Key Generation (Airgapped)

Each validator operator performs these steps on an **airgapped machine** (no network connection).

### 1.1 Prepare Airgapped Environment

```bash
# On airgapped machine
# Verify no network
ip addr show  # Should show only loopback

# Create secure workspace
mkdir -p ~/aura-keygen
cd ~/aura-keygen

# Copy aurad binary via USB (pre-verified checksum)
# Verify binary
sha256sum aurad
# Compare with published checksum
```

### 1.2 Generate Validator Key

```bash
# Initialize temporary node to generate keys
./aurad init val-keygen --chain-id aura-mvp-1 --home ./keygen

# The following files are created:
# ./keygen/config/priv_validator_key.json  <- CRITICAL: Validator signing key
# ./keygen/config/node_key.json            <- P2P identity key
# ./keygen/data/priv_validator_state.json  <- Signing state (starts at height 0)
```

### 1.3 Extract Public Information

```bash
# Get validator public key (safe to share)
cat ./keygen/config/priv_validator_key.json | jq -r '.pub_key'

# Get node ID (safe to share)
./aurad tendermint show-node-id --home ./keygen

# Get validator address (safe to share)
./aurad tendermint show-address --home ./keygen

# Record these values for genesis coordinator
```

### 1.4 Secure Key Backup

```bash
# Create encrypted backup of private keys
tar -czf validator-keys-val[N].tar.gz ./keygen/config/priv_validator_key.json ./keygen/config/node_key.json

# Encrypt with GPG (use operator's key)
gpg --symmetric --cipher-algo AES256 validator-keys-val[N].tar.gz

# Store encrypted backup in multiple secure locations:
# - Hardware security module (if available)
# - Encrypted USB drive in secure safe
# - Paper backup of mnemonic (if using mnemonic-based keys)

# Securely delete unencrypted files
shred -u validator-keys-val[N].tar.gz
```

### 1.5 Transfer to Online Machine

```bash
# Copy ONLY the files needed for genesis to USB:
# - priv_validator_key.json (for gentx signing)
# - Public key info recorded above

# Transfer via USB to online machine
# NEVER transfer the encrypted backup to a networked machine
```

---

## Phase 2: Genesis Transaction Creation

Each operator creates their genesis transaction (gentx) on an **online machine**.

### 2.1 Create Operator Account

```bash
# On online machine with aurad binary
./aurad keys add val[N]-operator --keyring-backend file --home ~/.aura-val[N]

# Record the mnemonic securely (paper backup)
# Record the address for genesis account allocation
```

### 2.2 Initialize Validator Home

```bash
# Initialize with chain-id
./aurad init val[N] --chain-id aura-mvp-1 --home ~/.aura-val[N]

# Copy validator key from USB
cp /usb/priv_validator_key.json ~/.aura-val[N]/config/
```

### 2.3 Create Genesis Transaction

```bash
# Create gentx (will be collected by genesis coordinator)
./aurad genesis gentx val[N]-operator 250000000000uaura \
  --chain-id aura-mvp-1 \
  --moniker "Validator [N]" \
  --commission-rate 0.10 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --keyring-backend file \
  --home ~/.aura-val[N]

# The gentx is created at:
# ~/.aura-val[N]/config/gentx/gentx-*.json
```

### 2.4 Submit Gentx to Coordinator

```bash
# Send gentx file to genesis coordinator via secure channel
# Include:
# - gentx-*.json file
# - Operator address
# - Node ID
# - P2P endpoint (ip:port)
```

---

## Phase 3: Genesis Assembly (Coordinator Only)

The genesis coordinator collects all gentxs and assembles the final genesis.

### 3.1 Collect All Gentxs

```bash
# Create genesis workspace
mkdir -p ~/genesis-assembly
cd ~/genesis-assembly

# Initialize base genesis
./aurad init genesis-coordinator --chain-id aura-mvp-1 --home ./genesis

# Collect gentx files from all validators
mkdir -p ./genesis/config/gentx
cp gentx-val1.json ./genesis/config/gentx/
cp gentx-val2.json ./genesis/config/gentx/
cp gentx-val3.json ./genesis/config/gentx/
cp gentx-val4.json ./genesis/config/gentx/
```

### 3.2 Add Genesis Accounts

```bash
# Add each validator's operator account
./aurad genesis add-genesis-account [val1-addr] 1000000000000uaura --home ./genesis
./aurad genesis add-genesis-account [val2-addr] 1000000000000uaura --home ./genesis
./aurad genesis add-genesis-account [val3-addr] 1000000000000uaura --home ./genesis
./aurad genesis add-genesis-account [val4-addr] 1000000000000uaura --home ./genesis

# Add faucet account
./aurad genesis add-genesis-account [faucet-addr] 100000000000000uaura --home ./genesis
```

### 3.3 Collect Genesis Transactions

```bash
# Validate and collect all gentxs
./aurad genesis collect-gentxs --home ./genesis

# Validate final genesis
./aurad genesis validate ./genesis/config/genesis.json
```

### 3.4 Set Genesis Time

```bash
# Edit genesis.json to set coordinated start time
# Set genesis_time to agreed launch time (e.g., 15 minutes in future)
jq '.genesis_time = "2026-01-20T12:00:00Z"' ./genesis/config/genesis.json > genesis-final.json
mv genesis-final.json ./genesis/config/genesis.json

# Validate again
./aurad genesis validate ./genesis/config/genesis.json
```

### 3.5 Distribute Final Genesis

```bash
# Calculate checksum
sha256sum ./genesis/config/genesis.json

# Distribute to all validators via:
# 1. Secure channel (Signal/encrypted email)
# 2. Upload to R2 artifacts (after launch time is past)
# 3. Include in testnets repository
```

---

## Phase 4: Validator Setup

Each operator sets up their validator with the final genesis.

### 4.1 Deploy Genesis

```bash
# Copy final genesis to validator home
scp genesis.json [server]:~/.aura-val[N]/config/genesis.json

# Verify checksum matches
ssh [server] "sha256sum ~/.aura-val[N]/config/genesis.json"
```

### 4.2 Configure Peers

```bash
# Edit config.toml with peer information
ssh [server]
vim ~/.aura-val[N]/config/config.toml

# Set persistent_peers with all other validators
# persistent_peers = "node1_id@ip1:port,node2_id@ip2:port,..."
```

### 4.3 Verify Setup

```bash
# Verify validator key matches genesis
./aurad tendermint show-validator --home ~/.aura-val[N]
# Compare with validator entry in genesis.json

# Verify genesis hash matches coordinator's
./aurad genesis validate ~/.aura-val[N]/config/genesis.json
```

---

## Phase 5: Coordinated Launch

### 5.1 Pre-Launch Checklist

Each operator confirms:
- [ ] Genesis file in place with correct checksum
- [ ] Validator key in place
- [ ] Peers configured
- [ ] Ports open (P2P, RPC)
- [ ] Monitoring ready
- [ ] On standby at launch time

### 5.2 Launch Sequence

```bash
# At T-5 minutes: All operators confirm ready
# At T-0: Start validators in sequence

# Validator 1 starts first (seed node)
ssh aura-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val1 > ~/.aura-val1/node.log 2>&1 &"

# Wait 30 seconds, then validators 2-4 start
ssh aura-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val2 > ~/.aura-val2/node.log 2>&1 &"
ssh services-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val3 > ~/.aura-val3/node.log 2>&1 &"
ssh services-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val4 > ~/.aura-val4/node.log 2>&1 &"
```

### 5.3 Verify Consensus

```bash
# After genesis time passes, verify blocks are being produced
curl -s http://127.0.0.1:10657/status | jq '.result.sync_info.latest_block_height'

# Verify all validators signing
curl -s http://127.0.0.1:10657/validators | jq '.result.validators | length'
# Expected: 4
```

---

## Security Considerations

### Key Storage
- Private validator keys should NEVER be stored unencrypted on networked machines
- Use HSM or secure enclaves where available
- Maintain encrypted offline backups in geographically separate locations

### Key Rotation
- Rotate keys annually or upon suspected compromise
- Document key rotation procedure separately
- Maintain key history for audit purposes

### Access Control
- Limit SSH access to validator servers
- Use key-based authentication only
- Enable 2FA where possible
- Monitor access logs

### Incident Response
- If key compromise suspected, immediately:
  1. Unbond validator stake
  2. Stop validator
  3. Rotate keys following this ceremony
  4. Investigate breach
- See INCIDENT_RESPONSE_RUNBOOK.md for details

---

## Ceremony Checklist

### Pre-Ceremony
- [ ] All operators identified and verified
- [ ] Secure communication channel established
- [ ] Airgapped machines prepared
- [ ] Binary checksums verified
- [ ] Launch time agreed

### During Ceremony
- [ ] Phase 1: Keys generated on airgapped machines
- [ ] Phase 2: Gentxs created and submitted
- [ ] Phase 3: Genesis assembled and validated
- [ ] Phase 4: Validators configured
- [ ] Phase 5: Coordinated launch successful

### Post-Ceremony
- [ ] All keys securely stored
- [ ] Genesis checksum published
- [ ] Chain producing blocks
- [ ] All validators signing
- [ ] Monitoring active
