# Security Module - Quick Reference Card

## Files Overview

| File | Lines | Purpose |
|------|-------|---------|
| `security.proto` | 963 | Core types and parameters |
| `query.proto` | 593 | Query service (40+ methods) |
| `tx.proto` | 761 | Transaction service (50+ messages) |
| `genesis.proto` | 133 | Genesis state definition |

## Module Statistics

- **Total Message Types**: 85+
- **Total Enum Types**: 15+
- **Query Methods**: 40+
- **Transaction Types**: 50+
- **Consolidated Modules**: 6

## Domain Breakdown

### Network Security
- **Config Types**: 7 (RateLimit, Connection, Mempool, Reputation, Gossip, Fork, Partition)
- **State Types**: 7 (PeerInfo, RateLimitEntry, NodeReputation, TrustedPeer, ForkAlert, PartitionAlert, MempoolStats)
- **Queries**: 9
- **Transactions**: 7

### Validator Security
- **Types**: 6 (ValidatorSecurityInfo, DoubleSignEvidence, DowntimeInfraction, ValidatorAlert, SentryNodeInfo)
- **Enums**: 2 (Severity, AlertType)
- **Queries**: 6
- **Transactions**: 7

### Wallet Security
- **Types**: 11 (HardwareWallet, MultiSig, SocialRecovery, Guardian, Recovery, Simulation, SpendingLimit, Biometric)
- **Enums**: 4 (HardwareWalletType, RecoveryStatus, SimulationRisk, BiometricType)
- **Queries**: 8
- **Transactions**: 10

### Incident Response
- **Types**: 3 (Incident, ResponseAction, AuditLogEntry)
- **Enums**: 2 (IncidentSeverity, IncidentStatus)
- **Queries**: 4
- **Transactions**: 5

### Cryptography
- **Types**: 5 (KeyRotation, ThresholdScheme, ZKProof, QuantumKey)
- **Enums**: 4 (ThresholdSchemeType, ThresholdSchemeStatus, ZKProofType, QuantumAlgorithm)
- **Queries**: 5
- **Transactions**: 7

### Privacy
- **Types**: 4 (StealthAddress, RingSignature, ConfidentialTx, MixingPool)
- **Queries**: 4
- **Transactions**: 6

## Quick Commands

### Generate Code
```bash
cd proto && buf generate
```

### Query Examples
```bash
# General
aurad query security params
aurad query security status

# Network
aurad query security network peer <peer-id>
aurad query security network health

# Validator
aurad query security validator <address>
aurad query security validator <address> alerts

# Wallet
aurad query security wallet <wallet-id>
aurad query security wallet <wallet-id> multisig

# Incident
aurad query security incident <id>
aurad query security incidents

# Crypto
aurad query security crypto key-rotation-schedule <id>
aurad query security crypto verify-zk-proof

# Privacy
aurad query security privacy mixing-pool <pool-id>
```

### Transaction Examples
```bash
# Network
aurad tx security add-trusted-peer <peer-id> <address> <pubkey>
aurad tx security ban-peer <peer-id> --duration=24h

# Validator
aurad tx security register-validator-security <hot-key> <cold-key>
aurad tx security register-sentry-node <address> <ip> <port>

# Wallet
aurad tx security register-hardware-wallet <type> <device-id>
aurad tx security create-multisig <signers> <threshold>

# Incident
aurad tx security create-incident <title> <desc> <severity>

# Crypto
aurad tx security create-key-rotation-schedule <key-id> <interval>
aurad tx security rotate-key <key-id> <new-pubkey>

# Privacy
aurad tx security create-mixing-pool <participants> <denomination>
```

## REST API Endpoints

### Query Endpoints (GET)
```
/aura/security/v1beta1/params
/aura/security/v1beta1/status
/aura/security/v1beta1/network/peer/{peer_id}
/aura/security/v1beta1/network/health
/aura/security/v1beta1/validator/{validator_address}
/aura/security/v1beta1/wallet/{wallet_id}
/aura/security/v1beta1/incident/{incident_id}
/aura/security/v1beta1/crypto/key_rotation_schedule/{id}
/aura/security/v1beta1/privacy/mixing_pool/{pool_id}
```

### Transaction Endpoints (POST)
```
/aura/security/v1beta1/tx/add_trusted_peer
/aura/security/v1beta1/tx/register_validator_security
/aura/security/v1beta1/tx/create_multisig_wallet
/aura/security/v1beta1/tx/create_incident
/aura/security/v1beta1/tx/rotate_key
/aura/security/v1beta1/tx/create_mixing_pool
```

## Store Key Prefixes

```go
// Network Security (0x01-0x05)
TrustedPeerPrefix      = []byte{0x01}
NodeReputationPrefix   = []byte{0x02}
RateLimitPrefix        = []byte{0x03}
ForkAlertPrefix        = []byte{0x04}
PartitionAlertPrefix   = []byte{0x05}

// Validator Security (0x10-0x14)
ValidatorSecurityPrefix     = []byte{0x10}
DoubleSignEvidencePrefix    = []byte{0x11}
DowntimeInfractionPrefix    = []byte{0x12}
ValidatorAlertPrefix        = []byte{0x13}
SentryNodePrefix            = []byte{0x14}

// Wallet Security (0x20-0x26)
HardwareWalletPrefix        = []byte{0x20}
MultiSigWalletPrefix        = []byte{0x21}
PendingMultiSigTxPrefix     = []byte{0x22}
SocialRecoveryPrefix        = []byte{0x23}
RecoveryRequestPrefix       = []byte{0x24}
SpendingLimitPrefix         = []byte{0x25}
BiometricAuthPrefix         = []byte{0x26}

// Incident Response (0x30-0x32)
IncidentPrefix              = []byte{0x30}
AuditLogPrefix              = []byte{0x31}
ResponseActionPrefix        = []byte{0x32}

// Cryptography (0x40-0x43)
KeyRotationSchedulePrefix   = []byte{0x40}
ThresholdSchemePrefix       = []byte{0x41}
ZKProofConfigPrefix         = []byte{0x42}
QuantumResistantKeyPrefix   = []byte{0x43}

// Privacy (0x50-0x53)
MixingPoolPrefix            = []byte{0x50}
StealthAddressPrefix        = []byte{0x51}
RingSignaturePrefix         = []byte{0x52}
ConfidentialTxPrefix        = []byte{0x53}
```

## Common Types

### Params
```protobuf
message Params {
  NetworkSecurityParams network = 1;
  ValidatorSecurityParams validator = 2;
  WalletSecurityParams wallet = 3;
  IncidentResponseParams incident = 4;
  CryptographyParams crypto = 5;
  PrivacyParams privacy = 6;
}
```

### GenesisState
```protobuf
message GenesisState {
  Params params = 1;
  NetworkSecurityState network_security = 2;
  ValidatorSecurityState validator_security = 3;
  WalletSecurityState wallet_security = 4;
  IncidentResponseState incident_response = 5;
  CryptographyState cryptography = 6;
  PrivacyState privacy = 7;
}
```

## Key Enums

### IncidentSeverity
```
LOW = 1
MEDIUM = 2
HIGH = 3
CRITICAL = 4
```

### ValidatorAlert.Severity
```
INFO = 0
WARNING = 1
CRITICAL = 2
```

### RecoveryStatus
```
PENDING = 1
APPROVED = 2
EXECUTED = 3
CANCELLED = 4
EXPIRED = 5
```

### ThresholdSchemeType
```
ECDSA = 1
EDDSA = 2
BLS = 3
SCHNORR = 4
```

### ZKProofType
```
GROTH16 = 1
PLONK = 2
BULLETPROOFS = 3
STARK = 4
```

## Implementation Checklist

- [ ] Generate Go code: `buf generate`
- [ ] Create module structure: `chain/x/security/`
- [ ] Implement keeper with CRUD operations
- [ ] Implement message server (50+ handlers)
- [ ] Implement query server (40+ handlers)
- [ ] Implement genesis init/export
- [ ] Define module in `module.go`
- [ ] Add to app.go
- [ ] Create CLI commands
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Add documentation
- [ ] Create migration scripts
- [ ] Deploy to testnet
- [ ] Security audit
- [ ] Deploy to mainnet

## Common Operations

### Keeper Operations
```go
// Store
k.SetTrustedPeer(ctx, peer)
k.SetValidatorSecurityInfo(ctx, info)
k.SetHardwareWalletConfig(ctx, config)

// Get
peer, found := k.GetTrustedPeer(ctx, peerID)
info, found := k.GetValidatorSecurityInfo(ctx, addr)
config, found := k.GetHardwareWalletConfig(ctx, walletID)

// List
peers := k.GetAllTrustedPeers(ctx)
validators := k.GetAllValidatorSecurityInfo(ctx)
wallets := k.GetAllHardwareWallets(ctx)

// Delete
k.RemoveTrustedPeer(ctx, peerID)
```

### Event Emission
```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        types.EventTypeAddTrustedPeer,
        sdk.NewAttribute(types.AttributeKeyPeerID, peerID),
        sdk.NewAttribute(types.AttributeKeyAddress, address),
    ),
)
```

## GoGo Proto Annotations

```protobuf
// Custom SDK types
string amount = 1 [
  (gogoproto.customtype) = "cosmossdk.io/math.Int",
  (gogoproto.nullable) = false
];

string fraction = 2 [
  (gogoproto.customtype) = "cosmossdk.io/math.LegacyDec",
  (gogoproto.nullable) = false
];

// Timestamps
google.protobuf.Timestamp time = 3 [
  (gogoproto.stdtime) = true,
  (gogoproto.nullable) = false
];

// Durations
google.protobuf.Duration duration = 4 [
  (gogoproto.stdduration) = true,
  (gogoproto.nullable) = false
];
```

## Error Handling

```go
// Common errors
ErrUnauthorized
ErrInvalidRequest
ErrNotFound
ErrAlreadyExists
ErrInvalidState
ErrThresholdNotMet
ErrExpired
```

## Testing Patterns

```go
// Unit test
func TestAddTrustedPeer(t *testing.T) {
    k, ctx := setupKeeper(t)
    peer := &types.TrustedPeer{...}
    err := k.SetTrustedPeer(ctx, peer)
    require.NoError(t, err)

    got, found := k.GetTrustedPeer(ctx, peer.PeerId)
    require.True(t, found)
    require.Equal(t, peer, got)
}

// Integration test
func TestMultiSigWorkflow(t *testing.T) {
    // Create wallet
    // Propose transaction
    // Sign transaction
    // Execute transaction
    // Verify outcome
}
```

## Resources

- **Proto Definitions**: `/home/decri/blockchain-projects/aura/proto/aura/security/v1beta1/`
- **Full Documentation**: `README.md` in proto directory
- **Implementation Guide**: `IMPLEMENTATION_GUIDE.md` in proto directory
- **Summary**: `/home/decri/blockchain-projects/aura/SECURITY_MODULE_PROTOBUF_SUMMARY.md`

## Support Contacts

For questions or issues:
1. Review the proto definitions
2. Check the implementation guide
3. Consult Cosmos SDK documentation
4. Review existing module implementations
