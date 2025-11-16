# Wallet Security Guide - Hot and Cold Storage

## Table of Contents
1. [Overview](#overview)
2. [Hot Wallet Security](#hot-wallet-security)
3. [Cold Storage Security](#cold-storage-security)
4. [Key Management](#key-management)
5. [Security Monitoring](#security-monitoring)
6. [Emergency Procedures](#emergency-procedures)

## Overview

This guide defines security protocols for managing hot wallets and cold storage on the Aura blockchain network. Proper implementation of these security measures is critical to protecting user funds and treasury assets.

### Security Architecture

```
┌─────────────────────────────────────────────────────┐
│              Total Treasury: 100B AURA               │
└─────────────────────────────────────────────────────┘
                         │
                         ├─── 90% (90B) ──────► Deep Cold Storage
                         │                      - 5-of-7 Multi-sig
                         │                      - Air-gapped
                         │                      - Physical vault
                         │                      - In-person signing
                         │
                         ├─── 9% (9B) ────────► Cold Storage
                         │                      - 3-of-5 Multi-sig
                         │                      - Hardware wallets
                         │                      - 24-hour timelock
                         │                      - Geographic distribution
                         │
                         ├─── 0.9% (900M) ────► Warm Storage
                         │                      - 2-of-3 Multi-sig
                         │                      - HSM protected
                         │                      - 1-hour timelock
                         │                      - Operational reserves
                         │
                         └─── 0.1% (100M) ────► Hot Wallets
                                               - Single-sig with limits
                                               - Daily operations
                                               - Real-time monitoring
                                               - Automated alerts
```

## Hot Wallet Security

### Configuration

#### Balance Limits
```bash
# Global hot wallet limits (set via governance)
aurad tx incidentresponse set-wallet-limits \
  $HOT_WALLET_ADDRESS \
  --max-balance=10000000000 \      # 10B tokens maximum
  --max-transaction=1000000000 \    # 1B tokens per transaction
  --daily-limit=5000000000 \        # 5B tokens per day
  --from=security-admin

# Verify limits
aurad query incidentresponse wallet-limits $HOT_WALLET_ADDRESS
```

#### Monitoring Configuration
```yaml
# config/hot-wallet-monitoring.yaml
hot_wallet:
  address: "aura1..."
  limits:
    max_balance: "10000000000"
    max_transaction_size: "1000000000"
    daily_limit: "5000000000"

  alerts:
    balance_threshold: 0.8  # Alert at 80% of max
    transaction_threshold: 0.7  # Alert at 70% of max
    daily_usage_threshold: 0.75  # Alert at 75% of daily limit

  notifications:
    - type: email
      recipients: ["security@aura.io"]
    - type: sms
      recipients: ["+1-XXX-XXX-XXXX"]
    - type: pagerduty
      service_key: "..."
```

### Access Controls

#### Authentication Requirements
```bash
# Enable multi-factor authentication
aurad keys add hot-wallet --multisig=hot-wallet-mfa

# Configure MFA providers
aurad config set mfa.providers "totp,yubikey"
aurad config set mfa.required true

# Set session timeouts
aurad config set session.timeout "15m"
aurad config set session.max_idle "5m"
```

#### IP Whitelisting
```bash
# Restrict hot wallet access to specific IPs
aurad config set hot-wallet.ip-whitelist "10.0.1.0/24,10.0.2.0/24"

# Configure firewall rules
iptables -A INPUT -p tcp --dport 26657 -s 10.0.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 26657 -j DROP
```

### Security Best Practices

#### 1. Key Rotation Schedule
```bash
# Rotate hot wallet keys monthly
# Generate new wallet
aurad keys add hot-wallet-new --keyring-backend file

# Transfer funds to new wallet
aurad tx bank send $OLD_HOT_WALLET \
  $(aurad keys show hot-wallet-new -a) \
  [amount]uaura \
  --gas=auto

# Update configuration
aurad config set hot-wallet.address $(aurad keys show hot-wallet-new -a)

# Archive old wallet
aurad keys rename $OLD_HOT_WALLET hot-wallet-archived-$(date +%Y%m)
```

#### 2. Transaction Review
```bash
# Require secondary approval for large transactions
aurad config set hot-wallet.approval-threshold "1000000000"
aurad config set hot-wallet.approvers "admin1,admin2"

# Review pending transactions
aurad query txs --events 'tx.from=$HOT_WALLET&tx.status=pending'
```

#### 3. Audit Logging
```yaml
# config/audit-logging.yaml
audit:
  enabled: true
  hot_wallet_transactions: true

  log_events:
    - transaction_submitted
    - transaction_signed
    - transaction_broadcast
    - balance_changed
    - limit_exceeded
    - access_denied

  log_destination:
    - type: file
      path: "/var/log/aura/hot-wallet-audit.log"
    - type: syslog
      server: "syslog.aura.io:514"
    - type: s3
      bucket: "aura-audit-logs"
      prefix: "hot-wallet/"
```

### Monitoring Alerts

#### Real-time Monitoring
```bash
#!/bin/bash
# hot-wallet-monitor.sh

HOT_WALLET="aura1..."
MAX_BALANCE="10000000000"
ALERT_THRESHOLD="8000000000"

while true; do
  # Check current balance
  BALANCE=$(aurad query bank balances $HOT_WALLET -o json | \
    jq -r '.balances[] | select(.denom=="uaura") | .amount')

  # Check if approaching limit
  if [ "$BALANCE" -gt "$ALERT_THRESHOLD" ]; then
    echo "WARNING: Hot wallet balance ($BALANCE) approaching limit"
    # Send alert
    curl -X POST https://api.pagerduty.com/incidents \
      -H "Authorization: Token token=$PAGERDUTY_KEY" \
      -d '{"incident":{"type":"incident","title":"Hot wallet balance high"}}'
  fi

  # Check recent transactions
  RECENT_TXS=$(aurad query txs --events "transfer.sender=$HOT_WALLET" \
    --limit 10 -o json)

  # Analyze for suspicious patterns
  # - Unusual amounts
  # - Unknown recipients
  # - High frequency

  sleep 60
done
```

#### Daily Report
```bash
#!/bin/bash
# generate-daily-report.sh

cat > hot-wallet-report-$(date +%Y%m%d).md <<EOF
# Hot Wallet Daily Report - $(date +%Y-%m-%d)

## Balance Status
- Current Balance: $(aurad query bank balances $HOT_WALLET)
- Limit Utilization: XX%
- Daily Transfer Amount: XX AURA
- Daily Limit Utilization: XX%

## Transaction Summary
- Total Transactions: XX
- Total Sent: XX AURA
- Total Received: XX AURA
- Average Transaction Size: XX AURA
- Largest Transaction: XX AURA

## Security Events
$(grep "hot-wallet" /var/log/aura/security.log | tail -20)

## Recommendations
- Transfer excess funds to cold storage: [YES/NO]
- Key rotation due: [DATE]
- Unusual activity detected: [NONE/DETAILS]
EOF

# Email report
mail -s "Hot Wallet Daily Report" security@aura.io < hot-wallet-report-$(date +%Y%m%d).md
```

## Cold Storage Security

### Setup Procedures

#### Deep Cold Storage (5-of-7 Multi-sig)
```bash
# Initialize air-gapped machine
# 1. Use clean, never-networked laptop
# 2. Install OS from verified media
# 3. Never connect to any network
# 4. Verify all software signatures

# Generate keys (on air-gapped machine)
for i in {1..7}; do
  aurad keys add cold-storage-key-$i \
    --keyring-backend file \
    --output json > key-$i.json
done

# Create multi-sig address
aurad keys add cold-storage-multisig \
  --multisig=cold-storage-key-1,cold-storage-key-2,cold-storage-key-3,cold-storage-key-4,cold-storage-key-5,cold-storage-key-6,cold-storage-key-7 \
  --multisig-threshold=5

# Print addresses and QR codes
for i in {1..7}; do
  aurad keys show cold-storage-key-$i -a
  aurad keys show cold-storage-key-$i -a | qrencode -o qr-key-$i.png
done

# Backup keys using Shamir Secret Sharing
# Split each key into 5 shares, requiring 3 to reconstruct
for i in {1..7}; do
  ssss-split -t 3 -n 5 -w key$i < key-$i.json
done

# Store shares in secure locations:
# - Bank safety deposit boxes (3 different banks)
# - Physical vault (company facility)
# - Encrypted hardware wallets (geographic distribution)
```

#### Standard Cold Storage (3-of-5 Multi-sig)
```bash
# Use hardware wallets (Ledger/Trezor)
# Generate keys on each device

# Connect hardware wallet 1
aurad keys add cold-key-1 --ledger --account 0

# Connect hardware wallet 2
aurad keys add cold-key-2 --ledger --account 0

# Connect hardware wallet 3
aurad keys add cold-key-3 --ledger --account 0

# Connect hardware wallet 4
aurad keys add cold-key-4 --ledger --account 0

# Connect hardware wallet 5
aurad keys add cold-key-5 --ledger --account 0

# Create multi-sig
aurad keys add cold-storage \
  --multisig=cold-key-1,cold-key-2,cold-key-3,cold-key-4,cold-key-5 \
  --multisig-threshold=3

# Distribute hardware wallets to key holders:
# - Geographic distribution (different countries)
# - Secure storage facilities
# - Trusted individuals with legal agreements
```

### Transaction Procedures

#### Creating Cold Storage Transaction
```bash
# Step 1: Create unsigned transaction (on online machine)
aurad tx bank send \
  $(aurad keys show cold-storage -a) \
  $DESTINATION_ADDRESS \
  1000000000uaura \
  --generate-only \
  > unsigned-tx.json

# Step 2: Copy to USB drive
cp unsigned-tx.json /media/usb/

# Step 3: Scan USB for malware
clamscan -r /media/usb/

# Step 4: Transfer to air-gapped signing stations
```

#### Signing Process (Air-gapped machines)
```bash
# Key Holder 1 (Location: New York)
aurad tx sign /media/usb/unsigned-tx.json \
  --from=cold-storage-key-1 \
  --multisig=$(aurad keys show cold-storage -a) \
  --chain-id=aura-mainnet \
  --output-document=/media/usb/signed-1.json

# Key Holder 2 (Location: London)
aurad tx sign /media/usb/unsigned-tx.json \
  --from=cold-storage-key-2 \
  --multisig=$(aurad keys show cold-storage -a) \
  --chain-id=aura-mainnet \
  --output-document=/media/usb/signed-2.json

# Key Holder 3 (Location: Singapore)
aurad tx sign /media/usb/unsigned-tx.json \
  --from=cold-storage-key-3 \
  --multisig=$(aurad keys show cold-storage -a) \
  --chain-id=aura-mainnet \
  --output-document=/media/usb/signed-3.json
```

#### Combining and Broadcasting
```bash
# Combine signatures (on air-gapped machine)
aurad tx multisign /media/usb/unsigned-tx.json \
  cold-storage \
  /media/usb/signed-1.json \
  /media/usb/signed-2.json \
  /media/usb/signed-3.json \
  --output-document=/media/usb/final-tx.json

# Verify transaction
aurad tx validate-signatures /media/usb/final-tx.json

# Transfer to online machine via USB
# Scan again for malware

# Broadcast (on online machine)
aurad tx broadcast /media/usb/final-tx.json

# Verify on blockchain
aurad query tx [TXHASH]
```

### Timelock Implementation

#### 24-hour Timelock for Cold Storage
```bash
# Create timelock transaction
aurad tx bank send \
  $COLD_STORAGE \
  $DESTINATION \
  1000000000uaura \
  --timelock=24h \
  --generate-only > timelock-tx.json

# Sign with required keys
# Transaction won't execute until 24 hours after submission

# During timelock period, can be cancelled by security team
aurad tx incidentresponse cancel-timelock [tx-hash] --from=security-admin
```

## Key Management

### Key Generation

#### High-Security Key Generation
```bash
# Use hardware random number generator
# Verify entropy quality

# Check entropy
cat /proc/sys/kernel/random/entropy_avail
# Should be > 3000

# Generate key with high entropy
aurad keys add secure-wallet \
  --keyring-backend file \
  --algo secp256k1

# Backup mnemonic to multiple secure locations
# - Print on paper, store in vault
# - Engrave on metal plate (fireproof)
# - Split using Shamir Secret Sharing
```

#### Key Derivation
```bash
# Use BIP39/BIP44 standard derivation
# Allows recreation of keys from mnemonic

# Create multiple accounts from same mnemonic
aurad keys add account-0 --recover --account 0
aurad keys add account-1 --recover --account 1
aurad keys add account-2 --recover --account 2
```

### Key Storage

#### Physical Security
```yaml
key_storage:
  deep_cold_storage:
    location: "Multiple bank vaults"
    format: "Paper + Metal plate"
    encryption: "BIP39 passphrase"
    access: "Dual-control, biometric"

  cold_storage:
    location: "Hardware wallets in safes"
    format: "Ledger Nano X"
    encryption: "PIN + passphrase"
    access: "Key holder + backup"

  warm_storage:
    location: "HSM in datacenter"
    format: "FIPS 140-2 Level 3 HSM"
    encryption: "Hardware-backed"
    access: "2-factor auth"

  hot_storage:
    location: "Encrypted filesystem"
    format: "Keyring file"
    encryption: "AES-256"
    access: "Application service account"
```

#### Digital Security
```bash
# Encrypt keyring file
gpg --symmetric --cipher-algo AES256 ~/.aura/keyring-file

# Store encrypted backup offsite
aws s3 cp ~/.aura/keyring-file.gpg \
  s3://aura-key-backups/keyring-$(date +%Y%m%d).gpg \
  --sse AES256

# Set restrictive permissions
chmod 600 ~/.aura/keyring-file
chattr +i ~/.aura/keyring-file  # Make immutable
```

### Key Rotation

#### Rotation Schedule
```yaml
rotation_schedule:
  hot_wallets: "Monthly"
  warm_storage: "Quarterly"
  cold_storage: "Annually"
  deep_cold_storage: "Every 2 years or on compromise"

  validator_keys: "After any security incident"
  admin_keys: "Every 6 months"
  api_keys: "Every 90 days"
```

#### Rotation Procedure
```bash
#!/bin/bash
# key-rotation.sh

OLD_KEY="old-wallet"
NEW_KEY="new-wallet"

echo "Starting key rotation..."

# 1. Generate new key
aurad keys add $NEW_KEY --keyring-backend file

# 2. Transfer all funds
BALANCE=$(aurad query bank balances $(aurad keys show $OLD_KEY -a) -o json | \
  jq -r '.balances[] | select(.denom=="uaura") | .amount')

aurad tx bank send \
  $(aurad keys show $OLD_KEY -a) \
  $(aurad keys show $NEW_KEY -a) \
  ${BALANCE}uaura \
  --gas=auto \
  --yes

# 3. Wait for confirmation
sleep 30

# 4. Verify transfer
NEW_BALANCE=$(aurad query bank balances $(aurad keys show $NEW_KEY -a) -o json | \
  jq -r '.balances[] | select(.denom=="uaura") | .amount')

if [ "$NEW_BALANCE" == "$BALANCE" ]; then
  echo "Transfer successful"

  # 5. Update configuration
  aurad config set wallet.address $(aurad keys show $NEW_KEY -a)

  # 6. Archive old key
  aurad keys rename $OLD_KEY $OLD_KEY-archived-$(date +%Y%m%d)

  # 7. Notify team
  echo "Key rotation completed successfully" | \
    mail -s "Key Rotation Complete" security@aura.io
else
  echo "ERROR: Transfer failed"
  exit 1
fi
```

## Security Monitoring

### Real-time Monitoring

#### Transaction Monitoring
```python
#!/usr/bin/env python3
# monitor-transactions.py

import requests
import time
from datetime import datetime

HOT_WALLET = "aura1..."
ALERT_AMOUNT = 1000000000  # 1B tokens

def check_transactions():
    url = f"https://api.aura.io/cosmos/tx/v1beta1/txs?events=transfer.sender={HOT_WALLET}"
    response = requests.get(url)
    txs = response.json()

    for tx in txs.get('tx_responses', []):
        # Parse transaction
        amount = parse_amount(tx)
        recipient = parse_recipient(tx)
        timestamp = tx.get('timestamp')

        # Check for suspicious activity
        if amount > ALERT_AMOUNT:
            send_alert(f"Large transaction detected: {amount} to {recipient}")

        if recipient in BLACKLIST:
            send_alert(f"Transaction to blacklisted address: {recipient}")

        if is_unusual_pattern(tx):
            send_alert(f"Unusual transaction pattern detected")

def send_alert(message):
    # Send to multiple channels
    send_email(message)
    send_sms(message)
    send_pagerduty(message)
    log_alert(message)

while True:
    check_transactions()
    time.sleep(10)
```

#### Balance Monitoring
```bash
#!/bin/bash
# monitor-balances.sh

declare -A WALLETS=(
  ["hot-wallet-1"]="aura1..."
  ["hot-wallet-2"]="aura1..."
  ["warm-storage"]="aura1..."
)

while true; do
  for NAME in "${!WALLETS[@]}"; do
    ADDRESS="${WALLETS[$NAME]}"

    # Get balance
    BALANCE=$(aurad query bank balances $ADDRESS -o json | \
      jq -r '.balances[] | select(.denom=="uaura") | .amount')

    # Get limit
    LIMIT=$(aurad query incidentresponse wallet-limits $ADDRESS -o json | \
      jq -r '.max_balance')

    # Calculate utilization
    UTIL=$(echo "scale=2; $BALANCE / $LIMIT * 100" | bc)

    echo "$(date) - $NAME: $BALANCE / $LIMIT ($UTIL%)"

    # Alert if high utilization
    if (( $(echo "$UTIL > 80" | bc -l) )); then
      echo "WARNING: $NAME balance high: $UTIL%"
      # Send alert
    fi
  done

  sleep 300  # Check every 5 minutes
done
```

### Anomaly Detection

#### Machine Learning Model
```python
# anomaly_detection.py

import pandas as pd
from sklearn.ensemble import IsolationForest

class TransactionAnomalyDetector:
    def __init__(self):
        self.model = IsolationForest(contamination=0.01)

    def train(self, historical_data):
        features = self.extract_features(historical_data)
        self.model.fit(features)

    def predict(self, transaction):
        features = self.extract_features([transaction])
        score = self.model.predict(features)
        return score[0] == -1  # -1 indicates anomaly

    def extract_features(self, transactions):
        return pd.DataFrame([{
            'amount': tx['amount'],
            'hour_of_day': tx['timestamp'].hour,
            'day_of_week': tx['timestamp'].dayofweek,
            'recipient_new': tx['recipient'] not in self.known_recipients,
            'frequency': self.calculate_frequency(tx),
        } for tx in transactions])
```

## Emergency Procedures

### Wallet Compromise Response

#### Immediate Actions (< 5 minutes)
```bash
#!/bin/bash
# emergency-wallet-response.sh

COMPROMISED_WALLET=$1

echo "=== EMERGENCY: WALLET COMPROMISE ==="
echo "Compromised wallet: $COMPROMISED_WALLET"
echo "Time: $(date)"

# 1. Generate new emergency wallet
NEW_WALLET=$(aurad keys add emergency-$(date +%s) --keyring-backend file --output json | jq -r .address)
echo "New wallet: $NEW_WALLET"

# 2. Transfer all funds immediately
echo "Transferring funds..."
aurad tx bank send $COMPROMISED_WALLET $NEW_WALLET \
  $(aurad query bank balances $COMPROMISED_WALLET -o json | \
    jq -r '.balances[] | select(.denom=="uaura") | .amount')uaura \
  --gas=auto \
  --fees=100000uaura \
  --broadcast-mode=block \
  --yes

# 3. Report incident
aurad tx incidentresponse report-incident \
  "Wallet compromise" \
  "Wallet $COMPROMISED_WALLET compromised, funds moved to $NEW_WALLET" \
  critical \
  "wallet" \
  --from=security-admin \
  --yes

# 4. Revoke old wallet
aurad keys delete $COMPROMISED_WALLET --yes

# 5. Alert all stakeholders
curl -X POST https://hooks.slack.com/... \
  -d '{"text":"🚨 WALLET COMPROMISED: Funds secured in emergency wallet"}'

echo "Emergency response complete"
echo "New wallet address: $NEW_WALLET"
```

### Key Loss Recovery

#### Recovery from Backups
```bash
# Recover from mnemonic backup
aurad keys add recovered-wallet --recover

# Verify recovery
EXPECTED_ADDRESS="aura1..."
RECOVERED_ADDRESS=$(aurad keys show recovered-wallet -a)

if [ "$EXPECTED_ADDRESS" == "$RECOVERED_ADDRESS" ]; then
  echo "Recovery successful"
else
  echo "ERROR: Recovery failed - address mismatch"
  exit 1
fi
```

#### Shamir Secret Reconstruction
```bash
# Reconstruct key from 3 of 5 shares
ssss-combine -t 3 -n 5

# Enter shares from 3 different locations
# Share 1: [ENTER]
# Share 2: [ENTER]
# Share 3: [ENTER]

# Output: reconstructed key
# Import reconstructed key
aurad keys import recovered-key keyfile.json
```

---

**Document Classification**: Internal - Highly Restricted
**Review Schedule**: Quarterly
**Next Review**: 2025-04-13
**Document Owner**: Chief Security Officer
