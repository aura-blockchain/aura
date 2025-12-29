---
sidebar_position: 5
---

# Security Modules

Comprehensive security controls and compliance.

## security

Core security controls and threat detection.

### Key Features
- Access control lists
- Rate limiting
- Threat detection
- Emergency shutdown

### Messages
| Message | Description |
|---------|-------------|
| `MsgUpdateSecurityPolicy` | Update security parameters |
| `MsgReportThreat` | Report security threat |
| `MsgEmergencyHalt` | Emergency chain halt (governance) |

### Queries
```bash
aurad query security policy
aurad query security threats --status active
aurad query security rate-limits <address>
```

---

## compliance

KYC/AML integration without storing PII.

### Key Features
- Compliance level verification
- Sanctions screening (hash-based)
- Transaction monitoring
- Jurisdiction rules

### Messages
| Message | Description |
|---------|-------------|
| `MsgSetComplianceLevel` | Set user compliance level |
| `MsgVerifyCompliance` | Verify transaction compliance |
| `MsgUpdateSanctionsList` | Update sanctions hashes |

### Compliance Levels
| Level | Description |
|-------|-------------|
| 0 | None (limited access) |
| 1 | Basic verification |
| 2 | Enhanced due diligence |
| 3 | Full verification |

---

## networksecurity

Network-level security protections.

### Key Features
- Peer reputation
- DoS protection
- P2P encryption
- Node banning

### Queries
```bash
aurad query networksecurity peer-reputation <peer-id>
aurad query networksecurity banned-peers
aurad query networksecurity network-health
```

---

## validatorsecurity

Validator-specific security policies.

### Key Features
- Slashing conditions
- Double-sign detection
- Uptime monitoring
- Key security

### Queries
```bash
aurad query validatorsecurity slashing-params
aurad query validatorsecurity validator-security-status <val-addr>
aurad query validatorsecurity signing-info <cons-addr>
```

---

## walletsecurity

Wallet protection features.

### Key Features
- Spending limits
- Multi-signature support
- Time-locked transactions
- Recovery mechanisms

### Messages
| Message | Description |
|---------|-------------|
| `MsgSetSpendingLimit` | Set daily spending limit |
| `MsgCreateTimelock` | Create time-locked transfer |
| `MsgSetupRecovery` | Configure recovery guardians |

---

## economicsecurity

Economic attack prevention.

### Key Features
- Flash loan protection
- Oracle manipulation detection
- MEV mitigation
- Liquidity monitoring

### Queries
```bash
aurad query economicsecurity flash-loan-params
aurad query economicsecurity oracle-health
aurad query economicsecurity mev-stats
```

---

## incidentresponse

Security incident handling.

### Key Features
- Incident reporting
- Automated response
- Post-incident analysis
- Recovery procedures

### Messages
| Message | Description |
|---------|-------------|
| `MsgReportIncident` | Report security incident |
| `MsgAcknowledgeIncident` | Acknowledge and assign |
| `MsgResolveIncident` | Mark incident resolved |
```
