# Bridge Module Disaster Recovery

**Module:** `x/bridge` | **Version:** 1.0 | **Last Updated:** 2025-12-23

---

## Emergency Response Quick Reference

### Immediate Actions (< 5 min)

```bash
# 1. Emergency pause bridge (globally)
aurad tx gov submit-proposal param-change bridge-pause.json --from governance
# Or via authorized address:
aurad tx bridge emergency-pause --from $EMERGENCY_ADDRESS

# 2. Check current bridge status
aurad query bridge params | grep -E "paused|enabled"

# 3. Verify invariants
aurad query crisis invariants bridge
```

### Pause Specific Chain

```bash
# Pause specific connected chain
aurad tx gov submit-proposal param-change \
  --title "Pause PAW bridge" \
  --description "Security incident on PAW chain" \
  --changes '[{"subspace":"bridge","key":"PausedChains","value":["paw"]}]' \
  --from governance
```

---

## Incident Types & Procedures

### 1. Fraudulent Transfer Detected

**Symptoms:** Invalid merkle proof, fake deposit claims, unauthorized minting

**Response:**
```bash
# 1. Pause bridge immediately
aurad tx bridge emergency-pause --from $EMERGENCY_ADDRESS

# 2. Check fraud proof window status
aurad query bridge params | grep fraud_proof_window

# 3. Submit fraud proof (if within window)
aurad tx bridge submit-fraud-proof \
  --transfer-id $TRANSFER_ID \
  --proof-data $PROOF_BYTES \
  --from $CHALLENGER

# 4. Review affected transfers
aurad query bridge transfer $TRANSFER_ID
aurad query bridge all-transfers --status PENDING

# 5. Slash malicious validators (automatic on fraud proof acceptance)
# Check slashing events:
aurad query bridge slashing-events
```

**Validators who signed fraudulent transfers are automatically slashed 50% and jailed.**

### 2. Validator Double-Signing

**Symptoms:** Same validator signs conflicting messages for same transfer

**Response:**
```bash
# Detection is automatic. Check slashing events:
aurad query bridge slashing-events --validator $VAL_ADDR

# Manual verification:
aurad query bridge validator $VAL_ADDR
aurad query bridge transfer $TRANSFER_ID --show-signatures
```

**Double-signing results in 100% stake slash (tombstoned) and permanent jail.**

### 3. Supply Cap Breach

**Symptoms:** Wrapped token supply exceeds configured cap

**Response:**
```bash
# 1. Check current supply vs caps
aurad query bridge stats
aurad query bridge params | grep supply_caps

# 2. Bridge auto-pauses if hourly threshold exceeded
# Check auto-pause status:
aurad query bridge params | grep auto_pause

# 3. Adjust caps via governance if legitimate
aurad tx gov submit-proposal param-change \
  --changes '[{"subspace":"bridge","key":"SupplyCaps","value":{"wrapped_eth":"10000000000"}}]'
```

### 4. Validator Set Degradation

**Symptoms:** Insufficient active validators, transfers stuck

**Response:**
```bash
# 1. Check validator status
aurad query bridge validators
aurad query bridge params | grep validator_threshold

# 2. Run invariant check
aurad query crisis invariants bridge

# 3. Unjail validators (if eligible)
aurad tx slashing unjail --from $VALIDATOR_KEY

# 4. Add emergency validators via governance
aurad tx gov submit-proposal add-bridge-validator \
  --validator-address $NEW_VALIDATOR \
  --from governance
```

### 5. State Corruption / Data Integrity Failure

**Symptoms:** Invariant failures, unmarshal errors, mismatched balances

**Response:**
```bash
# 1. Identify corruption
aurad query crisis invariants bridge

# 2. Export current state
aurad export --modules bridge > bridge_state_backup.json

# 3. Stop node and backup
sudo systemctl stop aurad
cp -r ~/.aura/data ~/.aura/data.corrupt.$(date +%Y%m%d)

# 4. Restore from last known good snapshot
# See "State Recovery" section below
```

---

## State Recovery Procedures

### Export Bridge State

```bash
# Full module export
aurad export --modules bridge --height $HEIGHT > bridge_export.json

# Verify export integrity
jq '.app_state.bridge.transfers | length' bridge_export.json
jq '.app_state.bridge.validators | length' bridge_export.json
```

### Import Bridge State

```bash
# Validate genesis before import
aurad validate-genesis genesis_with_bridge.json

# CRITICAL: Check for duplicate transfer IDs (causes collision)
jq -r '.app_state.bridge.transfers[].transfer_id' genesis.json | sort | uniq -d

# Import (requires chain restart)
aurad init-genesis genesis_with_bridge.json --home ~/.aura
```

### Rollback Transfer

```bash
# For stuck/failed transfers, mark as failed via governance:
aurad tx gov submit-proposal cancel-transfer \
  --transfer-id $TRANSFER_ID \
  --reason "Rollback due to source chain reorg" \
  --from governance

# Refund escrowed funds (automatic on transfer cancellation)
```

---

## Circuit Breaker Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `Paused` | `false` | Global bridge pause |
| `PausedChains` | `[]` | Per-chain pause list |
| `AutoPauseEnabled` | `false` | Auto-pause on anomaly |
| `AutoPauseThreshold` | `5B/hour` | Mint rate trigger |
| `EmergencyPauseAddresses` | `[]` | Authorized pause signers |
| `FraudProofWindow` | `7 days` | Challenge period |

### Modify via Governance

```bash
# Enable auto-pause with threshold
aurad tx gov submit-proposal param-change \
  --changes '[
    {"subspace":"bridge","key":"AutoPauseEnabled","value":true},
    {"subspace":"bridge","key":"AutoPauseThreshold","value":"1000000000"}
  ]'
```

---

## Slashing Parameters

| Offense | Slash % | Jail |
|---------|---------|------|
| Fraud signature | 50% | Yes |
| Double-signing | 100% | Permanent |
| Downtime | 1% | No |

**Liveness:** Validators must sign 50% of blocks in 10,000 block window.

---

## Invariant Checks

Run these after any incident:

```bash
# All bridge invariants
aurad query crisis invariants bridge

# Specific checks:
# - transfer-balance: Module balance >= locked transfers
# - merkle-proof-validity: All proofs have valid structure
# - validator-set-validity: Active validators >= minimum
# - security-parameters: Params within valid ranges
# - transfer-limits: No transfer exceeds max
# - channel-state: Chain configs properly configured
# - transfer-chain-integrity: All transfers reference valid chains
```

---

## Monitoring Alerts

Set up alerts for these events:

| Event | Severity | Action |
|-------|----------|--------|
| `bridge_auto_paused` | CRITICAL | Investigate immediately |
| `validator_slashed` | HIGH | Review and respond |
| `transfer_failed` | MEDIUM | Check and retry |
| `fraud_proof_submitted` | HIGH | Validate proof |

---

## Contacts & Escalation

See `docs/runbooks/EMERGENCY_PROCEDURES.md` for:
- Emergency hotline numbers
- PagerDuty integration
- Validator coordination channels
- Multi-sig key holders

---

## Recovery Checklist

```
[ ] Bridge paused / threat contained
[ ] Affected transfers identified
[ ] Malicious validators slashed
[ ] Invariants pass
[ ] State exported and backed up
[ ] Root cause identified
[ ] Fix deployed (if code issue)
[ ] Governance proposal for resume
[ ] Bridge resumed
[ ] Post-incident review scheduled
```
