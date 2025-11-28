# Emergency Procedures - Quick Reference Guide

## Critical Incidents Response

### 🚨 EMERGENCY CHAIN PAUSE

**When to use**: Active exploit, consensus failure, or critical security breach

```bash
# Step 1: Initiate pause (requires 3 of 5 signatures)
aurad tx incidentresponse request-pause \
  --requester=$YOUR_KEY \
  --level=full \
  --reason="[CRITICAL REASON]" \
  --incident-id=[INC-ID] \
  --duration=2h \
  --gas=auto \
  --yes

# Step 2: Other authorized signers approve
aurad tx incidentresponse approve-pause \
  --pause-id=[PAUSE-ID] \
  --approver=$YOUR_KEY \
  --gas=auto \
  --yes

# Step 3: Verify pause status
aurad query incidentresponse pause-state

# Step 4: After fix, resume chain
aurad tx incidentresponse resume \
  --resumed-by=$YOUR_KEY \
  --reason="[RESOLUTION]" \
  --gas=auto \
  --yes
```

**Authorized Key Holders**:
1. CSO: `aura1...` (Slot 1)
2. CTO: `aura1...` (Slot 2)
3. Lead Engineer: `aura1...` (Slot 3)
4. Validator Rep 1: `aura1...` (Slot 4)
5. Validator Rep 2: `aura1...` (Slot 5)

**Need 3 signatures to pause**

---

### 💰 HOT WALLET COMPROMISE

**Symptoms**: Unauthorized transactions, suspicious activity, key exposure

```bash
# IMMEDIATE ACTIONS (< 5 minutes)

# 1. Check wallet balance
aurad query bank balances $COMPROMISED_WALLET

# 2. Generate new wallet IMMEDIATELY
aurad keys add emergency-wallet --keyring-backend file

# 3. Transfer remaining funds to new wallet
aurad tx bank send $COMPROMISED_WALLET \
  $(aurad keys show emergency-wallet -a) \
  [amount]uaura \
  --gas=auto \
  --fees=1000uaura \
  --broadcast-mode=block \
  --yes

# 4. Report incident
aurad tx incidentresponse report-incident \
  --title="Hot wallet compromise" \
  --description="Wallet $COMPROMISED_WALLET compromised" \
  --severity=critical \
  --reported-by=$YOUR_KEY \
  --affected-systems="wallet,treasury" \
  --yes

# 5. Pause affected services
systemctl stop hot-wallet-service

# 6. Alert team
curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
  -H 'Content-Type: application/json' \
  -d '{"text":"🚨 HOT WALLET COMPROMISED: '$COMPROMISED_WALLET'"}'

# 7. Forensics (preserve logs)
cp -r /var/log/aurad /tmp/incident-logs-$(date +%Y%m%d-%H%M%S)/
tar -czf incident-evidence.tar.gz /tmp/incident-logs-*
```

---

### 🔐 COLD STORAGE EMERGENCY WITHDRAWAL

**When to use**: Security threat to cold storage, emergency fund needs

```bash
# REQUIRES IN-PERSON KEY HOLDER COORDINATION

# Step 1: Convene minimum required signers (3 of 5)
# - Verify identities with multi-factor auth
# - Use secure, swept meeting location
# - Verify hardware wallet firmware integrity

# Step 2: Connect hardware wallets to air-gapped machine
# Verify: No network connectivity, clean OS install

# Step 3: Create unsigned transaction
aurad tx bank send \
  $COLD_STORAGE_ADDRESS \
  $DESTINATION_ADDRESS \
  [amount]uaura \
  --generate-only \
  > /media/usb/unsigned-tx.json

# Step 4: Each key holder signs (on air-gapped machine)
# Key Holder 1:
aurad tx sign /media/usb/unsigned-tx.json \
  --from=cold-storage-key-1 \
  --multisig=$COLD_STORAGE_MULTISIG \
  --output-document=/media/usb/signed-1.json

# Key Holder 2:
aurad tx sign /media/usb/unsigned-tx.json \
  --from=cold-storage-key-2 \
  --multisig=$COLD_STORAGE_MULTISIG \
  --output-document=/media/usb/signed-2.json

# Key Holder 3:
aurad tx sign /media/usb/unsigned-tx.json \
  --from=cold-storage-key-3 \
  --multisig=$COLD_STORAGE_MULTISIG \
  --output-document=/media/usb/signed-3.json

# Step 5: Combine signatures (still air-gapped)
aurad tx multisign /media/usb/unsigned-tx.json \
  $COLD_STORAGE_MULTISIG \
  /media/usb/signed-1.json \
  /media/usb/signed-2.json \
  /media/usb/signed-3.json \
  --output-document=/media/usb/final-tx.json

# Step 6: Transfer to online machine via USB
# Scan USB for malware before connecting

# Step 7: Broadcast from online machine
aurad tx broadcast /media/usb/final-tx.json

# Step 8: Verify transaction
aurad query tx [TXHASH]
aurad query bank balances $DESTINATION_ADDRESS
```

**Cold Storage Multi-Sig**: `aura1...multisig...`
**Required Signatures**: 3 of 5
**Timelock**: 24 hours for non-emergency

---

### 📊 DATABASE CORRUPTION

**Symptoms**: Invalid state hash, consensus failures, query errors

```bash
# Step 1: Immediately pause chain
aurad tx incidentresponse request-pause \
  --requester=$YOUR_KEY \
  --level=full \
  --reason="Database corruption detected" \
  --duration=4h \
  --yes

# Step 2: Backup corrupted state for analysis
sudo systemctl stop aurad
cp -r ~/.aura/data ~/.aura/data.corrupt.$(date +%Y%m%d-%H%M%S)

# Step 3: Identify last known good height
aurad query block [height] --output json | jq .block.header.height

# Step 4: Download clean backup
BACKUP_HEIGHT=[LAST_GOOD_HEIGHT]
aws s3 cp s3://aura-backups-us/snapshot-${BACKUP_HEIGHT}.json.gz.gpg /tmp/

# Step 5: Decrypt and verify backup
gpg --decrypt /tmp/snapshot-${BACKUP_HEIGHT}.json.gz.gpg | \
  gunzip > /tmp/backup.json

jq empty /tmp/backup.json  # Validate JSON

# Step 6: Clear corrupted data
sudo rm -rf ~/.aura/data/*

# Step 7: Import backup
aurad import /tmp/backup.json --home ~/.aura

# Step 8: Restart and verify
sudo systemctl start aurad
sleep 30
aurad status | jq .sync_info

# Step 9: Coordinate with validators
# All validators must be on same state

# Step 10: Resume chain
aurad tx incidentresponse resume \
  --resumed-by=$YOUR_KEY \
  --reason="Corruption fixed, clean state restored" \
  --yes
```

---

### 🔥 VALIDATOR NODE FAILURE

**Symptoms**: Node not signing blocks, offline in validator set

```bash
# RAPID RESPONSE (< 10 minutes)

# Step 1: Check validator status
aurad query staking validator $VALIDATOR_OPERATOR_ADDR
aurad query slashing signing-info $VALIDATOR_CONS_ADDR

# Step 2: Activate backup validator
ssh backup-validator.aura.io

# Step 3: Sync backup validator (if not already synced)
sudo systemctl start aurad
aurad status | jq .sync_info.catching_up

# Step 4: Wait for full sync (or use state sync)
# State sync for faster recovery:
TRUST_HEIGHT=$(curl -s https://rpc.aura.io/block | jq -r .result.block.header.height)
TRUST_HASH=$(curl -s https://rpc.aura.io/block?height=$TRUST_HEIGHT | jq -r .result.block_id.hash)

# Update config.toml
sed -i.bak -e "s|^enable =.*|enable = true|" \
  -e "s|^rpc_servers =.*|rpc_servers = \"https://rpc.aura.io:26657,https://rpc2.aura.io:26657\"|" \
  -e "s|^trust_height =.*|trust_height = $TRUST_HEIGHT|" \
  -e "s|^trust_hash =.*|trust_hash = \"$TRUST_HASH\"|" \
  ~/.aura/config/config.toml

sudo systemctl restart aurad

# Step 5: Unjail validator (if jailed)
aurad tx slashing unjail \
  --from=$VALIDATOR_KEY \
  --gas=auto \
  --yes

# Step 6: Monitor validator signing
watch -n 5 'aurad query slashing signing-info $VALIDATOR_CONS_ADDR'

# Step 7: Update load balancer to use backup
aws elbv2 modify-target-group \
  --target-group-arn $TARGET_GROUP_ARN \
  --targets Id=backup-validator-1

# Step 8: Investigate failed primary
# Check logs, system resources, network connectivity
journalctl -u aurad -n 1000 --no-pager
```

---

### 🌐 NETWORK PARTITION

**Symptoms**: Validators can't reach consensus, multiple competing chains

```bash
# Step 1: Identify partition
# Check if validators can communicate
for VALIDATOR in $(cat validator-list.txt); do
  echo "=== $VALIDATOR ==="
  curl -s https://$VALIDATOR:26657/status | jq .result.sync_info
done

# Step 2: Pause all chains
# Each partition must coordinate pause
aurad tx incidentresponse request-pause \
  --level=full \
  --reason="Network partition detected" \
  --duration=2h

# Step 3: Identify canonical chain
# Choose chain with most validators and highest height
# Compare state hashes across validators

# Step 4: Coordinate validator consensus
# All validators must agree on canonical chain
# Use out-of-band communication (phone, telegram)

# Step 5: Validators on wrong chain rollback
aurad rollback
# Or restore from backup of canonical chain

# Step 6: Fix network connectivity
# Resolve DNS issues, firewall rules, ISP problems

# Step 7: All validators restart on canonical chain
pdsh -w ^validator-list.txt "systemctl restart aurad"

# Step 8: Verify consensus achieved
for VALIDATOR in $(cat validator-list.txt); do
  HASH=$(curl -s https://$VALIDATOR:26657/status | \
    jq -r .result.sync_info.latest_app_hash)
  echo "$VALIDATOR: $HASH"
done

# Step 9: Resume unified chain
aurad tx incidentresponse resume \
  --reason="Network partition resolved, consensus restored"
```

---

### 💾 DISASTER RECOVERY - COMPLETE INFRASTRUCTURE LOSS

**When to use**: Data center destroyed, all nodes offline, catastrophic failure

```bash
# ACTIVATE DISASTER RECOVERY PROTOCOL

# Phase 1: Situation Assessment (0-15 min)
# - What infrastructure is lost?
# - What backups are available?
# - Which validators can come online?

# Phase 2: Emergency Communication (15-30 min)
# Alert all stakeholders
curl -X POST https://api.statuspage.io/v1/pages/$PAGE_ID/incidents \
  -H "Authorization: OAuth $STATUSPAGE_TOKEN" \
  -d '{
    "incident": {
      "name": "Complete Infrastructure Failure",
      "status": "investigating",
      "impact": "critical",
      "body": "We are experiencing complete infrastructure failure. Disaster recovery protocols activated."
    }
  }'

# Phase 3: Deploy Emergency Infrastructure (30 min - 2 hours)
cd /path/to/aura-infrastructure/disaster-recovery/

# Initialize Terraform
terraform init

# Deploy emergency infrastructure
terraform apply \
  -var="scenario=disaster-recovery" \
  -var="validator_count=5" \
  -var="api_replicas=3" \
  -auto-approve

# Phase 4: Restore from Backup (2-4 hours)

# Get latest backup from multiple sources
LATEST_BACKUP=$(aws s3 ls s3://aura-backups-us/ | \
  grep snapshot | tail -n 1 | awk '{print $4}')

# Download backup
aws s3 cp s3://aura-backups-us/$LATEST_BACKUP /tmp/

# Verify backup integrity
gpg --decrypt /tmp/$LATEST_BACKUP | gunzip | jq empty

# Distribute to all emergency validators
for VALIDATOR in $(terraform output -json validator_ips | jq -r .[]); do
  echo "Deploying to $VALIDATOR"
  scp /tmp/$LATEST_BACKUP $VALIDATOR:/tmp/
  ssh $VALIDATOR "
    gpg --decrypt /tmp/$LATEST_BACKUP | gunzip > /tmp/backup.json
    aurad import /tmp/backup.json --home ~/.aura
    systemctl start aurad
  "
done

# Phase 5: Achieve Consensus (4-6 hours)
# Monitor validator sync
watch -n 10 '
  for IP in $(terraform output -json validator_ips | jq -r .[]); do
    echo "=== $IP ==="
    curl -s http://$IP:26657/status | jq .result.sync_info
  done
'

# Phase 6: Restore Services (6-8 hours)
# Deploy API, explorers, monitoring
kubectl apply -f k8s/disaster-recovery/

# Update DNS
aws route53 change-resource-record-sets \
  --hosted-zone-id $HOSTED_ZONE_ID \
  --change-batch file://dns-disaster-recovery.json

# Phase 7: Validation and Communication (8-10 hours)
# Run comprehensive test suite
./scripts/validate-disaster-recovery.sh

# Update status page
curl -X PATCH https://api.statuspage.io/v1/pages/$PAGE_ID/incidents/$INCIDENT_ID \
  -H "Authorization: OAuth $STATUSPAGE_TOKEN" \
  -d '{
    "incident": {
      "status": "resolved",
      "body": "Infrastructure recovered. All services operational."
    }
  }'
```

---

### 📞 EMERGENCY CONTACTS

```
PRIMARY ON-CALL
Security: +1-XXX-XXX-XXXX (Signal, WhatsApp)
DevOps:   +1-XXX-XXX-XXXX (Signal, WhatsApp)
CTO:      +1-XXX-XXX-XXXX (Signal, WhatsApp)

EMERGENCY HOTLINE: +1-XXX-XXX-XXXX
Available 24/7

PagerDuty: https://aura.pagerduty.com
Trigger incident via API or phone

COMMUNICATION CHANNELS
Telegram: @aura-emergency-team
Slack: #incident-response
Discord: #emergency-ops

STATUS PAGE
https://status.aura-network.io
Update via API or dashboard
```

---

### 🔧 QUICK DIAGNOSTICS

```bash
# Check chain status
aurad status | jq

# Check validator signing
aurad query slashing signing-info $(aurad tendermint show-validator)

# Check mempool
aurad query txs --events 'tx.height>0' --limit 1 --order_by desc

# Check pending transactions
curl -s http://localhost:26657/num_unconfirmed_txs

# Check peer connectivity
curl -s http://localhost:26657/net_info | jq .result.n_peers

# Check consensus state
curl -s http://localhost:26657/dump_consensus_state | jq

# Check validator set
aurad query staking validators --output json | \
  jq '.validators[] | {moniker: .description.moniker, status: .status}'

# Check system resources
df -h
free -h
top -b -n 1 | head -20
iostat -x 1 3
```

---

### 📋 INCIDENT CHECKLIST

```
Initial Response (0-15 min)
[ ] Incident detected and verified
[ ] Incident ID assigned: INC-___________
[ ] Severity assessed: [ ] Low [ ] Medium [ ] High [ ] Critical
[ ] Response team activated
[ ] Status page updated
[ ] Initial communication sent

Containment (15-60 min)
[ ] Attack vector identified
[ ] Affected systems isolated
[ ] Chain paused (if needed)
[ ] Evidence preserved
[ ] Monitoring enhanced
[ ] Stakeholders notified

Investigation (1-4 hours)
[ ] Root cause identified
[ ] Impact assessed
[ ] Attack timeline documented
[ ] Affected users identified
[ ] Financial impact calculated

Eradication (4-24 hours)
[ ] Fix developed and tested
[ ] Security updates deployed
[ ] Systems hardened
[ ] Compromised keys rotated
[ ] Chain resumed (if paused)

Recovery (24-72 hours)
[ ] All services restored
[ ] Operations normalized
[ ] Monitoring validated
[ ] Users notified of resolution

Post-Incident (3-7 days)
[ ] Post-mortem completed
[ ] Action items assigned
[ ] Lessons learned documented
[ ] Incident closed
```

---

### 🔒 SECURITY BEST PRACTICES

1. **Never share private keys via insecure channels**
2. **Always verify commands before execution**
3. **Use hardware wallets for signing critical transactions**
4. **Verify recipient addresses multiple times**
5. **Document all actions in incident timeline**
6. **Preserve evidence before remediation**
7. **Coordinate with all validators before chain-level actions**
8. **Test in isolated environment when possible**
9. **Have multiple authorized signers verify emergency actions**
10. **Maintain operational security during high-stress incidents**

---

**Document Version**: 1.0
**Last Updated**: 2025-01-13
**Review Schedule**: Monthly
**Distribution**: Emergency Response Team Only
