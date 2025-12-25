# Cryptography Module

The Cryptography module provides comprehensive cryptographic security features for the Aura blockchain, implementing production-grade security mechanisms for key management, encryption, and secure communications.

## Features

### 1. Key Rotation Mechanism with Automated Schedules

Automated key rotation helps maintain security by regularly updating cryptographic keys.

**Implementation Files:**
- `keeper/key_rotation.go` (lines 1-201)

**Key Functions:**
- `CreateKeyRotationSchedule`: Creates automated rotation schedules
- `RotateKey`: Performs manual key rotation
- `ProcessScheduledRotations`: Automatically processes due rotations
- `GetKeyRotationSchedule`: Retrieves rotation schedule details

**Features:**
- Configurable rotation intervals (minimum 1 hour)
- Automatic rotation based on schedules
- Rotation policies with age limits and warnings
- Maximum rotation attempt limits
- Notification support for key stakeholders
- Enable/disable schedules dynamically

### 2. Hierarchical Deterministic Key Derivation (BIP32/44)

Implements BIP32/BIP44 standard for hierarchical deterministic wallet key derivation.

**Implementation Files:**
- `keeper/hd_keys.go` (lines 1-247)

**Key Functions:**
- `DeriveHDKey`: Derives keys using BIP32/44 paths
- `DeriveChildKey`: Derives child keys from parent keys
- `ValidateBIP44Path`: Validates BIP44 derivation paths
- `GetHDKeyDerivation`: Retrieves HD key information

**Features:**
- Full BIP32 key derivation support
- BIP44 coin-type compliance (118 for Cosmos)
- Hardened and non-hardened derivation
- Chain code generation
- Path validation (e.g., m/44'/118'/0'/0/0)
- Metadata extraction (purpose, coin type, account, change, address index)

### 3. Threshold Signature Implementation

Supports threshold signature schemes requiring m-of-n signatures.

**Implementation Files:**
- `keeper/threshold_signatures.go` (lines 1-314)

**Key Functions:**
- `CreateThresholdScheme`: Creates threshold signature schemes
- `SubmitThresholdSignatureShare`: Submits signature shares
- `VerifyThresholdSignature`: Verifies combined signatures
- `RevokeThresholdScheme`: Revokes a scheme

**Supported Schemes:**
- ECDSA threshold signatures
- EdDSA threshold signatures
- BLS threshold signatures
- Schnorr threshold signatures

**Features:**
- Configurable threshold (t) and participant count (n)
- Signature share aggregation
- Automatic threshold detection
- Participant validation
- Scheme status management (active, suspended, revoked)

### 4. Zero-Knowledge Proof Integration

Provides ZK proof generation and verification infrastructure.

**Implementation Files:**
- `keeper/zk_proofs.go` (lines 1-223)

**Key Functions:**
- `RegisterZKProofCircuit`: Registers ZK proof circuits
- `SubmitZKProof`: Submits proofs for verification
- `VerifyZKProof`: Verifies zero-knowledge proofs
- `BatchVerifyZKProofs`: Efficient batch verification

**Supported Proof Systems:**
- Groth16
- PLONK
- Bulletproofs
- STARK

**Features:**
- Circuit registration with verification keys
- Public parameter management
- Proof verification with public inputs
- Batch verification for efficiency
- Multiple proof system support

### 5. Secure Enclave Support for Key Storage

Integrates with hardware security modules and secure enclaves.

**Implementation Files:**
- `keeper/secure_enclave.go` (lines 1-227)

**Key Functions:**
- `RegisterSecureEnclave`: Registers secure enclaves
- `SealDataToEnclave`: Seals data to enclave
- `UnsealDataFromEnclave`: Unseals data from enclave
- `RemoteAttestEnclave`: Performs remote attestation

**Supported Enclave Types:**
- Intel SGX (Software Guard Extensions)
- AMD SEV (Secure Encrypted Virtualization)
- TPM (Trusted Platform Module)
- HSM (Hardware Security Module)
- System Keychain

**Features:**
- Attestation verification
- Data sealing/unsealing
- Remote attestation
- Enclave status management
- Metadata storage

### 6. Quantum-Resistant Cryptographic Algorithms

Implements post-quantum cryptographic algorithms.

**Implementation Files:**
- `keeper/quantum_resistant.go` (lines 1-237)

**Key Functions:**
- `GenerateQuantumResistantKey`: Generates PQC keys
- `GetQuantumResistantKey`: Retrieves quantum-resistant keys
- `ValidateQuantumResistantKey`: Validates PQC keys
- `RotateQuantumResistantKey`: Rotates PQC keys

**Supported Algorithms (NIST PQC Standards):**
- CRYSTALS-Dilithium (Digital Signatures)
- CRYSTALS-Kyber (Key Encapsulation)
- Falcon (Digital Signatures)
- SPHINCS+ (Digital Signatures)
- NTRU (Key Encapsulation)

**Features:**
- NIST-standardized algorithms
- Key expiration management
- Algorithm-specific key lengths
- Metadata storage
- Key rotation support

### 7. Cryptographically Secure Random Number Generation

Provides CSPRNG (Cryptographically Secure Pseudo-Random Number Generator) functionality.

**Implementation Files:**
- `keeper/random.go` (lines 1-240)

**Key Functions:**
- `InitializeRandomSource`: Initializes random sources
- `GenerateSecureRandomBytes`: Generates random bytes
- `GenerateRandomInRange`: Generates random numbers in range
- `ReseedRandomSource`: Reseeds entropy pool
- `CheckEntropyHealth`: Monitors entropy health

**Random Source Types:**
- System random (crypto/rand)
- Hardware random
- Quantum random
- Combined sources

**Features:**
- Minimum entropy requirements (256 bits default)
- Entropy pool management
- Automatic health monitoring
- Reseeding support
- Statistics tracking

### 8. Salt for All Cryptographic Hashes

Implements salted hashing for all cryptographic operations.

**Implementation Files:**
- `keeper/hashing.go` (lines 1-260)

**Key Functions:**
- `CreateSaltedHash`: Creates hash with random salt
- `VerifySaltedHash`: Verifies data against hash
- `HashWithCustomSalt`: Uses custom salt
- `GenerateSalt`: Generates secure salt
- `BatchHashWithSalt`: Batch hashing

**Supported Hash Algorithms:**
- SHA-256
- SHA-512
- SHA3-256
- SHA3-512
- BLAKE2b
- BLAKE3

**Features:**
- Automatic salt generation (minimum 16 bytes)
- Iterated hashing support
- Constant-time comparison
- Batch operations
- Custom salt support

### 9. Key Stretching with PBKDF2 and Argon2

Password-based key derivation with industry-standard algorithms.

**Implementation Files:**
- `keeper/key_stretching.go` (lines 1-276)

**Key Functions:**
- `CreateKeyStretchingConfig`: Creates stretching configuration
- `StretchKey`: Performs key stretching
- `GetRecommendedStretchingConfig`: Returns recommended settings
- `VerifyStretchedKey`: Verifies stretched keys

**Supported Algorithms:**
- PBKDF2-SHA256 (100,000 iterations minimum)
- PBKDF2-SHA512 (100,000 iterations minimum)
- Argon2i (side-channel resistant)
- Argon2d (GPU attack resistant)
- Argon2id (hybrid, recommended)
- scrypt

**Features:**
- Configurable iterations, memory, and parallelism
- OWASP-compliant defaults
- Argon2 memory hardness
- Algorithm-specific validation
- Recommended configuration generator

### 10. Certificate Pinning for Network Communications

Implements certificate pinning to prevent MITM attacks.

**Implementation Files:**
- `keeper/cert_pinning.go` (lines 1-313)

**Key Functions:**
- `AddCertificatePin`: Pins certificates for hostnames
- `VerifyCertificatePin`: Verifies certificates
- `UpdateCertificatePin`: Updates pinned certificates
- `RotateCertificatePin`: Rotates certificate pins
- `CleanupExpiredPins`: Removes expired pins

**Pin Types:**
- SPKI (Subject Public Key Info) - Recommended
- Full Certificate
- Intermediate Certificate

**Features:**
- Multiple certificate hashes per hostname
- Certificate expiration tracking
- Pin rotation support
- Enable/disable pins
- Automatic cleanup of expired pins
- SPKI hash extraction

## Module Structure

```
chain/x/cryptography/
├── keeper/
│   ├── keeper.go              # Main keeper with core functionality
│   ├── key_rotation.go        # Key rotation implementation
│   ├── hd_keys.go            # HD key derivation (BIP32/44)
│   ├── threshold_signatures.go # Threshold signatures
│   ├── zk_proofs.go          # Zero-knowledge proofs
│   ├── secure_enclave.go     # Secure enclave support
│   ├── quantum_resistant.go  # Post-quantum cryptography
│   ├── random.go             # Secure random generation
│   ├── hashing.go            # Salted hashing
│   ├── key_stretching.go     # PBKDF2/Argon2
│   ├── cert_pinning.go       # Certificate pinning
│   └── keeper_test.go        # Comprehensive tests
├── types/
│   ├── keys.go               # Store keys
│   ├── errors.go             # Error definitions
│   ├── params.go             # Module parameters
│   └── genesis.go            # Genesis state
├── module.go                 # Module definition
└── README.md                 # This file
```

## Proto Definitions

```
proto/aura/cryptography/v1beta1/
├── cryptography.proto        # Core types and enums
├── genesis.proto            # Genesis state
├── query.proto              # Query service
└── tx.proto                 # Transaction messages
```

## Security Parameters

Default module parameters ensure production-grade security:

```go
DefaultRotationIntervalDays: 90
EnableAutoRotation: true
MinThresholdParticipants: 2
MaxThresholdParticipants: 100
MinEntropyBits: 256
MinPbkdf2Iterations: 100000
MinArgon2MemoryKb: 65536 (64 MB)
MinArgon2Iterations: 3
EnforceCertificatePinning: true
CertificatePinValidityDays: 365
MinSaltLengthBytes: 16
MinKeyLengthBits: 256
```

## Testing

Comprehensive test suite covering:
- All 10 cryptographic features
- Edge cases and error conditions
- Integration scenarios
- Performance benchmarks

Run tests:
```bash
cd chain/x/cryptography
go test -v ./...
```

## Usage Examples

### Key Rotation
```go
scheduleID, err := keeper.CreateKeyRotationSchedule(
    ctx,
    creator,
    keyID,
    86400, // 24 hours
    policy,
)
```

### HD Key Derivation
```go
hdKey, err := keeper.DeriveHDKey(
    ctx,
    masterKeyID,
    seedHash,
    "m/44'/118'/0'/0/0",
)
```

### Threshold Signatures
```go
schemeID, publicKey, err := keeper.CreateThresholdScheme(
    ctx,
    creator,
    3, // threshold
    5, // total participants
    participantIDs,
    ThresholdSchemeType_ECDSA,
)
```

### Quantum-Resistant Keys
```go
keyID, publicKey, err := keeper.GenerateQuantumResistantKey(
    ctx,
    creator,
    QuantumResistantAlgorithm_CRYSTALS_DILITHIUM,
    &expiresAt,
)
```

### Key Stretching
```go
configID, salt, err := keeper.CreateKeyStretchingConfig(
    ctx,
    KeyStretchingAlgorithm_ARGON2ID,
    3,     // iterations
    65536, // memory (64MB)
    4,     // parallelism
    32,    // key length
)

key, err := keeper.StretchKey(ctx, configID, password)
```

### Certificate Pinning
```go
pinID, err := keeper.AddCertificatePin(
    ctx,
    creator,
    "api.example.com",
    certificateHashes,
    CertificatePinType_SPKI,
    &expiresAt,
)

valid, err := keeper.VerifyCertificatePin(
    ctx,
    "api.example.com",
    certificate,
)
```

## Security Best Practices

1. **Key Rotation**: Enable automatic rotation for all long-lived keys
2. **Entropy**: Ensure minimum 256 bits of entropy for all random operations
3. **Salting**: Always use unique salts for each hash operation
4. **Key Stretching**: Use Argon2id with recommended parameters for passwords
5. **Certificate Pinning**: Pin SPKI hashes for critical endpoints
6. **Quantum Readiness**: Consider CRYSTALS-Dilithium for future-proof signatures
7. **Threshold Signatures**: Use t >= ⌈n/2⌉ + 1 for Byzantine fault tolerance
8. **Secure Storage**: Use secure enclaves for sensitive key material
9. **ZK Proofs**: Verify all proofs against registered circuits
10. **Monitoring**: Regularly check entropy health and rotation status

## Events

### EventKeyGenerated
Emitted when cryptographic key is generated.

**Attributes**: `key_type`, `key_id`, `algorithm`

### EventKeyRotated
Emitted when key is rotated.

**Attributes**: `key_id`, `old_key_hash`, `new_key_hash`

### EventSignatureCreated
Emitted when digital signature is created.

**Attributes**: `signer`, `message_hash`, `signature_type`

### EventSignatureVerified
Emitted when signature verification completes.

**Attributes**: `message_hash`, `valid`, `signer`

### EventEncryptionPerformed
Emitted when data is encrypted.

**Attributes**: `algorithm`, `key_id`, `data_size`

### EventDecryptionPerformed
Emitted when data is decrypted.

**Attributes**: `algorithm`, `key_id`

## Dependencies

- `golang.org/x/crypto`: Core cryptographic primitives
- `github.com/cosmos/cosmos-sdk`: Cosmos SDK framework
- Standard library: crypto/rand, crypto/sha256, crypto/sha512, crypto/x509

## License

Same as Aura blockchain project license.
