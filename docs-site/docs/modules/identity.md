---
sidebar_position: 2
---

# Identity Modules

Core modules for decentralized identity and verifiable credentials.

## identity

Manages Decentralized Identifiers (DIDs) on Aura.

### Key Features
- W3C DID Core specification compliant
- DID Document management
- Verification methods and authentication

### Messages
| Message | Description |
|---------|-------------|
| `MsgCreateDID` | Create a new DID |
| `MsgUpdateDID` | Update DID Document |
| `MsgDeactivateDID` | Deactivate a DID |

### Queries
```bash
aurad query identity did <did-id>
aurad query identity did-by-address <address>
aurad query identity all-dids --limit 100
```

---

## vcregistry

Verifiable Credential issuance, verification, and lifecycle management.

### Key Features
- W3C VC Data Model compliant
- Credential schema registry
- Revocation lists
- Issuer management

### Messages
| Message | Description |
|---------|-------------|
| `MsgIssueCredential` | Issue a new VC |
| `MsgRevokeCredential` | Revoke a VC |
| `MsgSuspendCredential` | Temporarily suspend VC |
| `MsgRegisterIssuer` | Register as credential issuer |

### Queries
```bash
aurad query vcregistry credential <id>
aurad query vcregistry verify-credential <id>
aurad query vcregistry credentials-by-subject <did>
aurad query vcregistry issuers
```

---

## confidencescore

Privacy-preserving identity confidence calculation.

### Key Features
- Aggregated confidence scores
- No PII exposure
- Configurable thresholds
- Proof-of-identity integration

### Queries
```bash
aurad query confidencescore score <did>
aurad query confidencescore threshold <credential-type>
```

---

## inclusionroutines

Proof-of-identity verification routines and rewards.

### Key Features
- Verification challenges
- Reward distribution
- AI assistant integration
- Fraud detection

### Messages
| Message | Description |
|---------|-------------|
| `MsgStartRoutine` | Begin verification |
| `MsgSubmitProof` | Submit verification proof |
| `MsgClaimReward` | Claim completion reward |

---

## identitychange

Identity update and recovery mechanisms.

### Key Features
- Key rotation
- Recovery procedures
- Guardian system
- Change history

### Messages
| Message | Description |
|---------|-------------|
| `MsgRotateKey` | Rotate verification keys |
| `MsgInitiateRecovery` | Start recovery process |
| `MsgCompleteRecovery` | Complete recovery |
