# Validator Security - Quick Reference Guide

## Quick Start

### 1. Register Your Validator

```bash
aurad tx validatorsecurity register-validator \
  --validator-address=auravaloper1... \
  --hot-key="your_hot_key_pub" \
  --cold-key="your_cold_key_pub" \
  --region="us-west-2" \
  --country-code="US" \
  --latitude=37.7749 \
  --longitude=-122.4194 \
  --backup-validators="auravaloper2...,auravaloper3..." \
  --from=your_key \
  --chain-id=aura
```

### 2. Register Sentry Nodes (Minimum 2 Required)

```bash
# First sentry node
aurad tx validatorsecurity register-sentry-node \
  --validator-address=auravaloper1... \
  --sentry-address=aurasentry1... \
  --ip-address="192.168.1.100" \
  --port=26656 \
  --from=your_key

# Second sentry node
aurad tx validatorsecurity register-sentry-node \
  --validator-address=auravaloper1... \
  --sentry-address=aurasentry2... \
  --ip-address="192.168.1.101" \
  --port=26656 \
  --from=your_key
```

### 3. Monitor Your Validator

```bash
# Check validator status
aurad query validatorsecurity validator auravaloper1...

# Check alerts
aurad query validatorsecurity alerts auravaloper1...

# Check sentry nodes
aurad query validatorsecurity sentry-nodes auravaloper1...
```

## Security Requirements Checklist

### ✅ Before Going Live

- [ ] Hot and cold keys are different
- [ ] At least 2 sentry nodes registered
- [ ] Sentry nodes are online and responding
- [ ] Minimum stake requirement met (1000 tokens)
- [ ] Geographic location properly configured
- [ ] Backup validators configured
- [ ] Monitoring system set up
- [ ] Alert notification configured

## Critical Thresholds

| Metric | Threshold | Action |
|--------|-----------|--------|
| **Missed Blocks** | 5000 out of 10,000 | Automatic jailing (24h) + 0.01% slash |
| **Double Signing** | Any occurrence | Permanent tombstone + 5% slash |
| **Minimum Stake** | Below 1000 tokens | Warning alert, cannot unjail |
| **Sentry Nodes** | Below 2 active | Critical alert, possible failover |
| **Jail Duration** | 24 hours | Can unjail after period expires |

## Alert Severity Levels

### 🟢 INFO
- Failover activated/restored
- Normal operational events
- **Action:** Monitor

### 🟡 WARNING
- Downtime detected
- Stake below minimum
- Geographic violations
- **Action:** Investigate within hours

### 🔴 CRITICAL
- Double-signing detected
- Sentry nodes offline
- Key compromise suspected
- **Action:** Immediate response required

## Common Operations

### Check if Jailed

```bash
aurad query validatorsecurity jailed
```

### Unjail Your Validator

```bash
# Requirements:
# 1. Jail period expired (24h)
# 2. Minimum stake maintained
# 3. At least 2 active sentry nodes

aurad tx validatorsecurity unjail \
  --validator-address=auravaloper1... \
  --from=your_key
```

### Update Security Info

```bash
aurad tx validatorsecurity update-security-info \
  --validator-address=auravaloper1... \
  --hot-key="new_hot_key" \
  --cold-key="new_cold_key" \
  --region="eu-west-1" \
  --backup-validators="new_backup1,new_backup2" \
  --from=your_key
```

### Acknowledge Alert

```bash
aurad tx validatorsecurity acknowledge-alert \
  --alert-id="alert-123" \
  --acknowledger-address=aura1... \
  --from=your_key
```

## Penalties Overview

### Double Signing
- **Slash:** 5% of bonded tokens
- **Jail:** Permanent (tombstoned)
- **Recovery:** IMPOSSIBLE
- **Prevention:** Never run two validator instances with same key

### Downtime
- **Slash:** 0.01% of bonded tokens
- **Jail:** 24 hours
- **Recovery:** Can unjail after period
- **Prevention:** Maintain high availability, use sentry nodes

## Best Practices

### 🔐 Security

1. **Key Management**
   - Store cold keys in hardware wallet
   - Never expose cold keys on internet-connected systems
   - Rotate hot keys periodically
   - Use different keys for signing and staking

2. **Sentry Node Architecture**
   ```
   Internet → Sentry Node 1 ─┐
                              ├→ Validator Node (Private)
   Internet → Sentry Node 2 ─┘
   ```
   - Minimum 2 sentry nodes
   - Different geographic locations
   - DDoS protection enabled
   - Firewall configured

3. **Monitoring**
   - Set up automated alerts
   - Monitor missed blocks daily
   - Check sentry node health
   - Review alerts promptly

### ⚡ Performance

1. **Uptime**
   - Target: >95% block signing
   - Use redundant infrastructure
   - Configure automatic failover
   - Register backup validators

2. **Sentry Nodes**
   - Keep heartbeat regular (every 5 min)
   - Monitor blocked DDoS requests
   - Scale horizontally if needed
   - Use load balancers

3. **Geographic Distribution**
   - Choose uncongested regions
   - Respect regional limits (max 10/region)
   - Consider latency to other validators
   - Backup in different region

## Troubleshooting

### Problem: Validator Jailed for Downtime

**Check:**
```bash
aurad query validatorsecurity validator auravaloper1...
```

**Solution:**
1. Wait for jail period (24h)
2. Ensure minimum stake maintained
3. Verify sentry nodes active
4. Run unjail command
5. Monitor closely after unjailing

### Problem: Sentry Node Offline

**Check:**
```bash
aurad query validatorsecurity sentry-nodes auravaloper1...
```

**Solution:**
1. Restart sentry node
2. Check network connectivity
3. Verify firewall rules
4. Update heartbeat
5. Add additional sentry if needed

### Problem: Below Minimum Stake

**Check:**
```bash
aurad query staking validator auravaloper1...
```

**Solution:**
1. Delegate more tokens
2. Cannot unjail until above minimum
3. Alert will clear automatically when above threshold

### Problem: Region Capacity Exceeded

**Check:**
```bash
aurad query validatorsecurity params
```

**Solution:**
1. Update validator region
2. Choose less congested region
3. Contact governance if legitimate need

## Emergency Procedures

### If Double-Signing Detected

1. **IMMEDIATELY STOP ALL VALIDATOR INSTANCES**
2. Investigate cause (likely key duplication)
3. Validator is permanently tombstoned
4. Must set up new validator with new keys
5. Cannot recover existing validator

### If Extended Downtime

1. Check validator process status
2. Check sentry node connectivity
3. Review system resources (CPU, memory, disk)
4. Check network connectivity
5. Failover should activate automatically if configured
6. Restore primary when issue resolved

### If Alerts Not Received

1. Check alert monitoring system
2. Query alerts manually via CLI
3. Verify alert configuration
4. Set up redundant alert channels
5. Review alert acknowledgment history

## Query Commands Reference

```bash
# Get module parameters
aurad query validatorsecurity params

# Get validator security info
aurad query validatorsecurity validator <validator-address>

# List all validators
aurad query validatorsecurity validators

# List jailed validators
aurad query validatorsecurity jailed

# List tombstoned validators
aurad query validatorsecurity tombstoned

# Get double-sign evidences
aurad query validatorsecurity evidences

# Get validator alerts
aurad query validatorsecurity alerts <validator-address>

# Get sentry nodes
aurad query validatorsecurity sentry-nodes <validator-address>
```

## Configuration Parameters

```yaml
double_sign_slash_fraction: "0.05"      # 5%
downtime_slash_fraction: "0.0001"       # 0.01%
signed_blocks_window: 10000             # blocks
min_signed_per_window: "0.5"            # 50%
downtime_jail_duration: "24h"
minimum_stake_amount: "1000000000000"   # 1000 tokens
enable_geo_distribution: true
max_validators_per_region: 10
require_sentry_nodes: true
min_sentry_nodes: 2
monitoring_interval: "5m"
enable_auto_failover: true
failover_timeout: "10m"
```

## Monitoring Metrics to Track

1. **Missed Blocks** - Should be < 500 in 10,000 block window
2. **Last Seen** - Should be within last few minutes
3. **Sentry Heartbeat** - All nodes should heartbeat every 5 min
4. **Active Sentry Nodes** - Should always be ≥ 2
5. **Unacknowledged Alerts** - Should be 0
6. **Failover Status** - Should be inactive
7. **Jail Status** - Should be false
8. **Tombstone Status** - Should be false

## Support

For issues or questions:
- Review logs: `journalctl -u aurad -f`
- Check documentation: `/x/validatorsecurity/README.md`
- Review alerts: `aurad query validatorsecurity alerts`
- Community support: [Aura Discord/Forum]

## Important Notes

⚠️ **NEVER:**
- Run multiple validator instances with same key
- Expose cold keys to internet
- Ignore critical alerts
- Run without sentry nodes
- Skip monitoring setup

✅ **ALWAYS:**
- Maintain key separation
- Keep sentry nodes active
- Monitor alerts regularly
- Maintain minimum stake
- Have backup validators configured
- Test failover periodically
- Keep infrastructure updated
