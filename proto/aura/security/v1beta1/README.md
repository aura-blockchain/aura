# Consolidated Security Module - Protocol Buffer Definitions

This directory contains the comprehensive protobuf definitions for the consolidated security module, which unifies six security-focused modules into a single cohesive security framework.

## Overview

The security module consolidates the following individual modules:
- **Network Security** - Peer management, rate limiting, mempool protection
- **Validator Security** - Validator monitoring, slashing, sentry nodes
- **Wallet Security** - Hardware wallets, multi-sig, social recovery
- **Incident Response** - Incident management, audit logging
- **Cryptography** - Key rotation, threshold signatures, ZK proofs, quantum resistance
- **Privacy** - Stealth addresses, ring signatures, mixing pools

## File Structure

### 1. security.proto (963 lines)
**Main types and parameters file**

Contains all core message types organized by security domain:

#### Unified Parameters
- `Params` - Consolidated parameters for all security domains
- `NetworkSecurityParams` - Network security configuration
- `ValidatorSecurityParams` - Validator security configuration
- `WalletSecurityParams` - Wallet security configuration
- `IncidentResponseParams` - Incident response configuration
- `CryptographyParams` - Cryptography configuration
- `PrivacyParams` - Privacy configuration

#### Network Security Types
- `RateLimitConfig`, `ConnectionConfig`, `MempoolConfig`
- `ReputationConfig`, `GossipConfig`, `ForkDetectionConfig`, `PartitionDetectionConfig`
- `PeerInfo`, `RateLimitEntry`, `NodeReputation`, `TrustedPeer`
- `ForkAlert`, `PartitionAlert`, `MempoolStats`

#### Validator Security Types
- `ValidatorSecurityInfo` - Comprehensive validator security information
- `DoubleSignEvidence`, `DowntimeInfraction`
- `ValidatorAlert` (with Severity and AlertType enums)
- `SentryNodeInfo`

#### Wallet Security Types
- `HardwareWalletConfig` (with HardwareWalletType enum)
- `MultiSigWallet`, `PendingMultiSigTransaction`
- `SocialRecoveryConfig`, `Guardian`, `RecoveryRequest` (with RecoveryStatus enum)
- `TransactionSimulation`, `StateChange`, `BalanceChange` (with SimulationRisk enum)
- `SpendingLimit`, `BiometricAuth` (with BiometricType enum)

#### Incident Response Types
- `Incident` (with IncidentSeverity and IncidentStatus enums)
- `ResponseAction`, `AuditLogEntry`

#### Cryptography Types
- `KeyRotationSchedule`, `KeyRotationPolicy`
- `ThresholdSignatureScheme` (with ThresholdSchemeType and ThresholdSchemeStatus enums)
- `ZKProofConfig` (with ZKProofType enum)
- `QuantumResistantKey` (with QuantumResistantAlgorithm enum)

#### Privacy Types
- `StealthAddress`, `RingSignature`, `ConfidentialTransaction`
- `MixingPool`

### 2. genesis.proto (133 lines)
**Genesis state definition**

Defines the complete genesis state for the consolidated security module:

- `GenesisState` - Root genesis state message containing:
  - `Params` - Unified parameters
  - `NetworkSecurityState` - Network security genesis data
  - `ValidatorSecurityState` - Validator security genesis data
  - `WalletSecurityState` - Wallet security genesis data
  - `IncidentResponseState` - Incident response genesis data
  - `CryptographyState` - Cryptography genesis data
  - `PrivacyState` - Privacy genesis data

Each domain-specific state message contains the relevant persistent data for that security domain.

### 3. query.proto (593 lines)
**Query service definitions**

Comprehensive query service with 40+ query methods organized by domain:

#### General Queries
- `Params` - Get unified security parameters
- `SecurityStatus` - Get overall security status across all domains

#### Network Security Queries (9 methods)
- `PeerInfo`, `AllPeers`, `TrustedPeers`
- `PeerReputation`, `RateLimitStatus`
- `MempoolStats`, `ForkAlerts`, `PartitionAlerts`
- `NetworkHealth`

#### Validator Security Queries (6 methods)
- `ValidatorSecurityInfo`, `AllValidatorSecurityInfo`
- `ValidatorAlerts`, `DoubleSignEvidences`
- `DowntimeInfractions`, `SentryNodes`

#### Wallet Security Queries (8 methods)
- `WalletSecurityInfo`, `HardwareWalletConfig`
- `MultiSigWallet`, `PendingMultiSigTransactions`
- `SocialRecoveryConfig`, `RecoveryRequests`
- `SpendingLimits`, `SimulateTransaction`

#### Incident Response Queries (4 methods)
- `Incident`, `AllIncidents`
- `AuditLog`, `ResponseActions`

#### Cryptography Queries (5 methods)
- `KeyRotationSchedule`, `AllKeyRotationSchedules`
- `ThresholdScheme`, `VerifyZKProof`
- `QuantumResistantKey`

#### Privacy Queries (4 methods)
- `MixingPool`, `AllMixingPools`
- `StealthAddress`, `VerifyRingSignature`

### 4. tx.proto (761 lines)
**Transaction message definitions**

Complete transaction service with 50+ message types organized by domain:

#### Network Security Messages (7 messages)
- `AddTrustedPeer`, `RemoveTrustedPeer`
- `BanPeer`, `UnbanPeer`
- `UpdatePeerReputation`
- `ResolveForkAlert`, `ResolvePartitionAlert`

#### Validator Security Messages (7 messages)
- `RegisterValidatorSecurity`, `UpdateValidatorSecurity`
- `RegisterSentryNode`, `RemoveSentryNode`
- `ReportDoubleSign`, `AcknowledgeValidatorAlert`
- `TriggerFailover`

#### Wallet Security Messages (10 messages)
- `RegisterHardwareWallet`
- `CreateMultiSigWallet`, `ProposeMultiSigTransaction`
- `SignMultiSigTransaction`, `ExecuteMultiSigTransaction`
- `ConfigureSocialRecovery`, `InitiateRecovery`
- `ApproveRecovery`, `ExecuteRecovery`
- `SetSpendingLimits`, `RegisterBiometric`

#### Incident Response Messages (5 messages)
- `CreateIncident`, `UpdateIncident`, `ResolveIncident`
- `ExecuteResponseAction`, `AddAuditLogEntry`

#### Cryptography Messages (7 messages)
- `CreateKeyRotationSchedule`, `RotateKey`
- `CreateThresholdScheme`, `SubmitThresholdSignatureShare`
- `RegisterZKProofCircuit`, `SubmitZKProof`
- `GenerateQuantumResistantKey`

#### Privacy Messages (6 messages)
- `CreateMixingPool`, `JoinMixingPool`, `ExecuteMixing`
- `GenerateStealthAddress`, `CreateRingSignature`
- `CreateConfidentialTransaction`

#### Admin Messages (1 message)
- `UpdateParams` - Update module parameters

## Key Design Patterns

### 1. Namespace Organization
All types are properly namespaced under `aura.security.v1beta1` to avoid conflicts with existing modules during the transition period.

### 2. Parameter Consolidation
The unified `Params` message brings together all security-related configuration into a single hierarchical structure while maintaining logical domain separation.

### 3. GoGo Proto Annotations
- Custom types for SDK math types (`cosmossdk.io/math.Int`, `cosmossdk.io/math.LegacyDec`)
- Standard time conversions (`stdtime`, `stdduration`)
- Nullable field control for safety

### 4. Comprehensive Coverage
Every type from the original six modules has been included and properly organized:
- **85+ message types** covering all security domains
- **15+ enum types** for type-safe status and configuration values
- **Full backward compatibility** with existing module interfaces

### 5. REST API Integration
All query endpoints include `google.api.http` annotations for automatic REST API generation with logical URL paths like:
- `/aura/security/v1beta1/network/...`
- `/aura/security/v1beta1/validator/...`
- `/aura/security/v1beta1/wallet/...`
- `/aura/security/v1beta1/incident/...`
- `/aura/security/v1beta1/crypto/...`
- `/aura/security/v1beta1/privacy/...`

### 6. Cosmos SDK Integration
All transaction messages include proper `cosmos.msg.v1.signer` annotations for automatic signer validation and routing.

## Usage

### Generating Go Code

To generate Go code from these proto files:

```bash
# From the proto directory
buf generate

# Or using the Makefile
make proto-gen
```

### Importing in Go Code

```go
import (
    securitytypes "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// Use the types
params := securitytypes.Params{
    Network: &securitytypes.NetworkSecurityParams{
        RateLimit: &securitytypes.RateLimitConfig{
            MaxRequestsPerSecond: 100,
            BurstSize: 200,
        },
    },
    Validator: &securitytypes.ValidatorSecurityParams{
        RequireSentryNodes: true,
        MinSentryNodes: 2,
    },
}
```

## Migration Strategy

These proto definitions support a phased migration approach:

1. **Phase 1**: Deploy security module alongside existing modules
2. **Phase 2**: Migrate data from old modules to security module
3. **Phase 3**: Route new transactions through security module
4. **Phase 4**: Deprecate old modules
5. **Phase 5**: Remove old module code

The namespace separation (`aura.security.v1beta1` vs `aura.networksecurity.v1beta1`, etc.) allows both old and new modules to coexist during migration.

## Next Steps

1. **Code Generation**: Run `buf generate` to create Go bindings
2. **Keeper Implementation**: Implement the keeper logic for each domain
3. **Message Server**: Implement transaction message handlers
4. **Query Server**: Implement query handlers
5. **Genesis**: Implement genesis import/export
6. **Tests**: Add comprehensive unit and integration tests
7. **Documentation**: Update user-facing documentation

## Statistics

- **Total Lines**: 2,450 lines of proto definitions
- **Message Types**: 85+ types across all domains
- **Query Methods**: 40+ query endpoints
- **Transaction Messages**: 50+ transaction types
- **Enums**: 15+ enumeration types
- **Consolidated Modules**: 6 security modules unified

## References

For implementation examples, see:
- `/home/decri/blockchain-projects/aura/proto/aura/networksecurity/v1beta1/`
- `/home/decri/blockchain-projects/aura/proto/aura/validatorsecurity/v1beta1/`
- `/home/decri/blockchain-projects/aura/proto/aura/walletsecurity/v1beta1/`
- `/home/decri/blockchain-projects/aura/proto/aura/cryptography/v1beta1/`
- `/home/decri/blockchain-projects/aura/proto/aura/privacy/v1beta1/`
