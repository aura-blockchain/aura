---
sidebar_position: 3
---

# Privacy Modules

Zero-knowledge proofs, encryption, and privacy-preserving features.

## privacy

Core privacy module with ZK proof verification and selective disclosure.

### Key Features
- Zero-knowledge proof verification
- Selective disclosure proofs
- Privacy-preserving credential presentation
- Nullifier management (prevent double-use)

### Messages
| Message | Description |
|---------|-------------|
| `MsgSubmitZKProof` | Submit ZK proof for verification |
| `MsgCreateSelectiveDisclosure` | Create selective disclosure proof |
| `MsgVerifyPresentation` | Verify privacy-preserving presentation |

### Queries
```bash
aurad query privacy proof <proof-id>
aurad query privacy nullifier-used <nullifier>
aurad query privacy supported-circuits
```

### Supported Circuits
- **Age verification**: Prove 18+ without revealing DOB
- **Residency**: Prove country without revealing address
- **Credential ownership**: Prove holding VC without revealing content

---

## cryptography

Advanced cryptographic primitives.

### Key Features
- Ring signatures (anonymous signing)
- Threshold signatures (multi-party)
- BLS signatures (aggregation)
- Pedersen commitments

### Messages
| Message | Description |
|---------|-------------|
| `MsgCreateRingSig` | Create ring signature |
| `MsgVerifyRingSig` | Verify ring signature |
| `MsgInitThresholdKey` | Initialize threshold key ceremony |
| `MsgSubmitKeyShare` | Submit threshold key share |

### Queries
```bash
aurad query cryptography ring-groups
aurad query cryptography threshold-keys
aurad query cryptography verify-signature <sig-data>
```

---

## prevalidation

Pre-transaction validation for privacy-preserving transactions.

### Key Features
- Validate transactions before broadcast
- Privacy checks without state modification
- Gas estimation for ZK proofs
- Circuit compatibility verification

### Queries
```bash
aurad query prevalidation validate-proof <proof-data>
aurad query prevalidation estimate-gas <tx-type>
```
