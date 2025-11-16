# Privacy Module Quick Reference

## File Locations & Line Numbers

### Core Implementation Files

| Feature | File Path | Lines | Key Functions |
|---------|-----------|-------|---------------|
| **Zero-Knowledge Proofs** | `chain/x/privacy/zkproof.go` | 1-417 | NewZKProofSystem (24-43)<br>GenerateProof (45-60)<br>VerifyProof (62-77)<br>GenerateRangeProof (180-208) |
| **Stealth Addresses** | `chain/x/privacy/stealth.go` | 1-377 | NewStealthAddressScheme (18-22)<br>GenerateStealthKeys (35-66)<br>GenerateStealthAddress (77-118)<br>ScanForStealthPayments (120-162) |
| **Ring Signatures** | `chain/x/privacy/ringsig.go` | 1-459 | NewRingSigner (28-32)<br>Sign (43-136)<br>Verify (138-188)<br>SignMLSAG (295-357) |
| **Confidential Transactions** | `chain/x/privacy/confidential.go` | 1-420 | NewConfidentialTransactionSystem (31-62)<br>CreateCommitment (73-103)<br>GenerateBulletproof (117-187)<br>CreateRingCT (230-282) |
| **Network Privacy** | `chain/x/privacy/network.go` | 1-363 | NewNetworkPrivacyManager (33-65)<br>CreateCircuit (246-285)<br>RotateCircuits (307-339) |
| **Coin Mixing** | `chain/x/privacy/mixing.go` | 1-372 | NewMixingService (63-70)<br>CreatePool (79-125)<br>ExecuteMixing (187-232) |
| **Encrypted Memos** | `chain/x/privacy/encryption.go` | 1-335 | NewMemoEncryptor (27-31)<br>Encrypt (39-55)<br>Decrypt (57-73) |
| **View Keys** | `chain/x/privacy/encryption.go` | 336-469 | NewViewKeyManager (352-357)<br>GenerateViewKey (364-399)<br>RevokeViewKey (419-430) |

### Supporting Files

| File | Path | Purpose |
|------|------|---------|
| **Keeper** | `chain/x/privacy/keeper.go` | State management, module integration |
| **Module** | `chain/x/privacy/module.go` | Cosmos SDK module definition |
| **Proto - Privacy** | `proto/aura/privacy/v1beta1/privacy.proto` | Privacy message definitions |
| **Proto - TX** | `proto/aura/privacy/v1beta1/tx.proto` | Transaction messages |
| **Proto - Query** | `proto/aura/privacy/v1beta1/query.proto` | Query messages |
| **Tests - ZK** | `chain/x/privacy/zkproof_test.go` | ZK proof tests |
| **Tests - Stealth** | `chain/x/privacy/stealth_test.go` | Stealth address tests |
| **Tests - Main** | `chain/x/privacy/privacy_test.go` | Integration tests |
| **Documentation** | `chain/x/privacy/README.md` | Comprehensive documentation |

## Quick Usage Examples

### 1. Zero-Knowledge Proofs

```go
// Create ZK proof system
zkSystem, _ := NewZKProofSystem(ZKProofTypeGroth16, "transfer_circuit")

// Generate proof
witness := []byte("secret_data")
publicInputs := [][]byte{[]byte("public_input")}
proof, _ := zkSystem.GenerateProof(witness, publicInputs)

// Verify proof
valid, _ := zkSystem.VerifyProof(proof, publicInputs)
```

**File**: `chain/x/privacy/zkproof.go`, Lines 24-77

### 2. Stealth Addresses

```go
// Create stealth address scheme
scheme := NewStealthAddressScheme()

// Generate recipient keys
keys, _ := scheme.GenerateStealthKeys()

// Generate one-time address
stealthAddr, _ := scheme.GenerateStealthAddress(
    keys.SpendKeyPair.PublicKey,
    keys.ViewKeyPair.PublicKey,
)

// Scan for payments
isForMe, _ := scheme.ScanForStealthPayments(
    stealthAddr.TxPublicKey,
    stealthAddr.OneTimePublicKey,
    keys.ViewKeyPair.PrivateKey,
    keys.SpendKeyPair.PublicKey,
)
```

**File**: `chain/x/privacy/stealth.go`, Lines 18-162

### 3. Ring Signatures

```go
// Create ring signer
signer := NewRingSigner()

// Sign message
privateKey := big.NewInt(12345)
publicKeys := [][]byte{key1, key2, key3}
message := []byte("transaction data")

signature, _ := signer.Sign(0, privateKey, publicKeys, message)

// Verify signature
valid, _ := signer.Verify(signature, message)
```

**File**: `chain/x/privacy/ringsig.go`, Lines 28-188

### 4. Confidential Transactions

```go
// Create CT system
ct, _ := NewConfidentialTransactionSystem(32)

// Create commitment
value := big.NewInt(1000)
commitment, _ := ct.CreateCommitment(value)

// Generate range proof
rangeProof, _ := ct.GenerateBulletproof(
    value,
    commitment.BlindingFactor,
)

// Verify range proof
valid, _ := ct.VerifyBulletproof(rangeProof)
```

**File**: `chain/x/privacy/confidential.go`, Lines 31-219

### 5. Network Privacy (Tor/I2P)

```go
// Configure network privacy
config := &NetworkPrivacyConfig{
    NetworkType:     NetworkTypeTor,
    TorProxyAddr:    "127.0.0.1:9050",
    CircuitLifetime: 10 * time.Minute,
    StreamIsolation: true,
}

// Create privacy manager
npm, _ := NewNetworkPrivacyManager(config)

// Create circuit
circuit, _ := npm.CreateCircuit()

// Get stats
stats := npm.GetNetworkPrivacyStats()
```

**File**: `chain/x/privacy/network.go`, Lines 33-339

### 6. Coin Mixing

```go
// Create mixing service
mixingService := NewMixingService(5)

// Create mixing pool
pool, _ := mixingService.CreatePool(
    big.NewInt(1000), // denomination
    5,                // min participants
    20,               // max participants
    3,                // mixing rounds
    1 * time.Hour,    // deadline
    big.NewInt(10),   // fee
)

// Join pool
commitment := []byte("commitment_hash")
err := mixingService.JoinPool(
    pool.ID,
    "participant_address",
    commitment,
    outputAddress,
    blindingFactor,
)

// Execute mixing
result, _ := mixingService.ExecuteMixing(pool.ID)
```

**File**: `chain/x/privacy/mixing.go`, Lines 63-232

### 7. Encrypted Memos

```go
// Create memo encryptor
encryptor := NewMemoEncryptor(AlgorithmChaCha20Poly1305)

// Encrypt memo
memo := []byte("Private transaction note")
recipientPubKey := []byte("recipient_public_key")

encrypted, _ := encryptor.Encrypt(memo, recipientPubKey)

// Decrypt memo
privateKey := []byte("private_key")
decrypted, _ := encryptor.Decrypt(encrypted, privateKey)
```

**File**: `chain/x/privacy/encryption.go`, Lines 27-305

### 8. View Keys

```go
// Create view key manager
vkm := NewViewKeyManager()

// Generate view key
address := []byte("wallet_address")
permissions := []string{"view_incoming", "view_balance"}
expiresAt := time.Now().Add(24 * time.Hour)

viewKey, _ := vkm.GenerateViewKey(
    ViewKeyTypeIncoming,
    address,
    permissions,
    &expiresAt,
)

// Verify permission
hasPermission, _ := vkm.VerifyPermission(
    viewKey.PublicKey,
    "view_incoming",
)

// Revoke key
vkm.RevokeViewKey(viewKey.PublicKey)
```

**File**: `chain/x/privacy/encryption.go`, Lines 352-469

## Testing Commands

```bash
# Run all privacy tests
cd chain/x/privacy
go test -v

# Run specific test
go test -v -run TestZKProofSystem_GenerateAndVerify

# Run with coverage
go test -cover

# Run benchmarks
go test -bench=.
```

## Common Constants & Types

### Zero-Knowledge Proof Types
```go
ZKProofTypeGroth16      // Lines: zkproof.go:13
ZKProofTypePlonk        // Lines: zkproof.go:15
ZKProofTypeBulletproofs // Lines: zkproof.go:17
ZKProofTypeSTARK        // Lines: zkproof.go:19
```

### Network Privacy Types
```go
NetworkTypeTor   // Lines: network.go:17
NetworkTypeI2P   // Lines: network.go:18
NetworkTypeMixed // Lines: network.go:19
```

### View Key Types
```go
ViewKeyTypeIncoming // Lines: encryption.go:338
ViewKeyTypeOutgoing // Lines: encryption.go:339
ViewKeyTypeAudit    // Lines: encryption.go:340
ViewKeyTypeFull     // Lines: encryption.go:341
```

### Encryption Algorithms
```go
AlgorithmAES256GCM         // Lines: encryption.go:14
AlgorithmChaCha20Poly1305  // Lines: encryption.go:15
AlgorithmXChaCha20Poly1305 // Lines: encryption.go:16
```

### Mixing Pool Statuses
```go
PoolStatusPending   // Lines: mixing.go:13
PoolStatusActive    // Lines: mixing.go:14
PoolStatusMixing    // Lines: mixing.go:15
PoolStatusCompleted // Lines: mixing.go:16
PoolStatusCancelled // Lines: mixing.go:17
```

## Error Handling Patterns

All functions follow consistent error handling:

```go
// Input validation
if input == nil || len(input) == 0 {
    return nil, errors.New("input cannot be empty")
}

// Range checking
if value.Sign() < 0 || value.Cmp(maxValue) >= 0 {
    return nil, errors.New("value out of range")
}

// Wrapped errors
if err := operation(); err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
```

## Security Best Practices

1. **Always verify proofs before accepting transactions**
2. **Use minimum ring size of 11 for production**
3. **Rotate Tor circuits every 10 minutes**
4. **Never reuse nonces in encryption**
5. **Set expiration times on view keys**
6. **Use multiple mixing rounds (3+)**
7. **Verify range proofs to prevent negative amounts**
8. **Scan for stealth payments regularly**

## Performance Tips

1. **Batch ZK proof verification** when possible
2. **Use Curve25519** for better stealth address performance
3. **Cache verification keys** to avoid regeneration
4. **Use stream isolation** only when necessary
5. **Implement circuit pooling** for Tor/I2P
6. **Parallel mixing** for multiple pools
7. **Precompute** curve points for frequent operations

## Integration with Cosmos SDK

```go
// In app.go
import "github.com/aequitas/aura/chain/x/privacy"

// Add store key
keys := sdk.NewKVStoreKeys(
    // ... other keys
    privacy.StoreKey,
)

// Create keeper
privacyKeeper := privacy.NewKeeper(
    appCodec,
    keys[privacy.StoreKey],
)

// Register module
app.ModuleManager = module.NewManager(
    // ... other modules
    privacy.NewAppModule(appCodec, privacyKeeper),
)
```

## Module Parameters

```go
Params{
    EnableZKProofs:                 true,
    EnableStealthAddresses:         true,
    EnableRingSignatures:           true,
    EnableConfidentialTransactions: true,
    EnableNetworkPrivacy:           true,
    EnableMixing:                   true,
    MinRingSize:                    3,
    MaxRingSize:                    16,
    MinMixingParticipants:          5,
    MixingFee:                      "100",
    ZKProofVerificationCost:        1000,
}
```

**File**: `chain/x/privacy/module.go`, Lines 114-128

## Cryptographic Primitives Used

- **Elliptic Curves**: NIST P-256, secp256k1, Curve25519
- **Hash Functions**: SHA-256, SHA3-256, Keccak-256
- **Encryption**: AES-256-GCM, ChaCha20-Poly1305
- **Commitments**: Pedersen commitments
- **Signatures**: ECDSA, Ring signatures
- **Key Agreement**: ECDH
- **KDF**: HKDF-SHA256

## Support & Documentation

- **Full README**: `chain/x/privacy/README.md` (567 lines)
- **Implementation Summary**: `PRIVACY_MODULE_IMPLEMENTATION.md`
- **Proto Documentation**: `proto/aura/privacy/v1beta1/*.proto`
- **Test Examples**: `chain/x/privacy/*_test.go`

## Total Implementation Statistics

- **Core Files**: 9 Go files
- **Total Lines**: 3,708 lines of code
- **Functions**: 157 functions
- **Tests**: 569 lines, 41 test functions
- **Documentation**: 567 lines
- **Proto Definitions**: 444 lines
