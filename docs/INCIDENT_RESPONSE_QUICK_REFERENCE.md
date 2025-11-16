# Incident Response - Quick Reference Card

## 🚨 Emergency Contacts

**Security Hotline**: [REDACTED] (24/7)
**PagerDuty**: https://aura.pagerduty.com
**Telegram**: @aura-emergency-team
**Email**: security@aura-network.io

## Emergency Chain Pause

```bash
# Step 1: Request pause (First signer)
aurad tx incidentresponse request-pause \
  full "Critical issue" INC-001 2h \
  --from security-key-1 --yes

# Step 2: Approve pause (Second signer)
aurad tx incidentresponse approve-pause \
  pause-001 --from security-key-2 --yes

# Step 3: Approve pause (Third signer)
aurad tx incidentresponse approve-pause \
  pause-001 --from security-key-3 --yes

# Resume after fix
aurad tx incidentresponse resume \
  "Fixed and verified" --from security-key-1 --yes
```

**Required**: 3 of 5 authorized signers
**Max Duration**: 24 hours

## Hot Wallet Compromise

```bash
# 1. Generate new wallet
NEW=$(aurad keys add emergency-$(date +%s) -o json | jq -r .address)

# 2. Transfer all funds IMMEDIATELY
aurad tx bank send $COMPROMISED $NEW \
  $(aurad query bank balances $COMPROMISED -o json | \
    jq -r '.balances[0].amount')uaura \
  --gas=auto --fees=100000uaura --yes

# 3. Report incident
aurad tx incidentresponse report-incident \
  "Wallet compromise" "Funds secured" critical "wallet" \
  --from admin --yes
```

**Timeline**: < 5 minutes

## Report Incident

```bash
aurad tx incidentresponse report-incident \
  "[TITLE]" \
  "[DESCRIPTION]" \
  [low|medium|high|critical] \
  "system1,system2" \
  --from mykey --yes
```

## Query Status

```bash
# Check chain pause state
aurad query incidentresponse pause-state

# Get incident details
aurad query incidentresponse incident INC-001

# List all incidents
aurad query incidentresponse incidents

# Check wallet limits
aurad query incidentresponse wallet-limits aura1...
```

## Severity Levels

| Level | Response Time | Escalation |
|-------|--------------|------------|
| **Critical** | < 15 min | All team + CEO |
| **High** | < 1 hour | Response team + CTO |
| **Medium** | < 4 hours | Relevant teams |
| **Low** | < 24 hours | Security team |

## Pause Levels

- **transactions**: Block new transactions only
- **modules**: Disable specific modules
- **full**: Complete chain halt

## Key Limits

| Wallet Type | Max Balance | Max TX | Daily Limit |
|-------------|-------------|--------|-------------|
| Hot Wallet | 10B AURA | 1B AURA | 5B AURA |
| Warm Storage | 900M AURA | - | - |
| Cold Storage | 9B AURA | - | - |
| Deep Cold | 90B AURA | - | - |

## Response Phases

1. **Detection** (0-15 min): Identify and triage
2. **Containment** (15-60 min): Stop the bleeding
3. **Investigation** (1-4 hours): Root cause
4. **Eradication** (4-24 hours): Fix and deploy
5. **Recovery** (24-72 hours): Restore services
6. **Post-Mortem** (3-7 days): Document and improve

## Authorized Pause Keys

1. CSO: `aura1...` (Slot 1)
2. CTO: `aura1...` (Slot 2)
3. Lead Engineer: `aura1...` (Slot 3)
4. Validator Rep 1: `aura1...` (Slot 4)
5. Validator Rep 2: `aura1...` (Slot 5)

**Requirement**: 3 signatures to pause chain

## Communication Channels

- **Status Page**: https://status.aura-network.io
- **Twitter**: @AuraNetwork
- **Discord**: #announcements
- **Telegram**: @AuraOfficial
- **Email**: All registered users

## Quick Diagnostics

```bash
# Chain status
aurad status | jq .sync_info

# Validator signing
aurad query slashing signing-info $(aurad tendermint show-validator)

# Check peers
curl -s http://localhost:26657/net_info | jq .result.n_peers

# System health
df -h && free -h && uptime
```

## Backup Validation

```bash
# Check last backup
aws s3 ls s3://aura-backups-us/ | tail -1

# Validate backup
aws s3 cp s3://aura-backups-us/latest.json.gz.gpg /tmp/
gpg --decrypt /tmp/latest.json.gz.gpg | gunzip | jq empty
```

## Recovery Objectives

- **RTO** (Recovery Time): 2 hours
- **RPO** (Recovery Point): 15 minutes
- **MTD** (Max Tolerable Downtime): 6 hours

## Documentation

- Incident Response Plan: `docs/INCIDENT_RESPONSE_PLAN.md`
- Disaster Recovery: `docs/DISASTER_RECOVERY_PLAN.md`
- Emergency Procedures: `docs/runbooks/EMERGENCY_PROCEDURES.md`
- Wallet Security: `docs/WALLET_SECURITY_GUIDE.md`
- Communication Plan: `docs/COMMUNICATION_PLAN.md`

## Keep This Card Accessible

Print and store in:
- [ ] Operations war room
- [ ] Each validator location
- [ ] Executive offices
- [ ] Home offices (on-call team)
- [ ] Mobile device (screenshot)

---

**Last Updated**: 2025-01-13
**Version**: 1.0
**Review**: Quarterly
