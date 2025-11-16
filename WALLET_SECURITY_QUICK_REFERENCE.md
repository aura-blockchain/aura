# Wallet Security - Quick Reference Guide

## File Locations

### Core Implementation
```
chain/x/walletsecurity/
├── keeper/
│   ├── keeper.go                 (Base keeper - Lines 1-283)
│   ├── hardware_wallet.go        (Hardware wallet - Lines 1-273)
│   ├── multisig.go              (Multi-sig - Lines 1-344)
│   ├── social_recovery.go       (Social recovery - Lines 1-344)
│   ├── security_features.go     (Other features - Lines 1-686)
│   ├── keeper_test.go           (Unit tests - Lines 1-444)
│   └── integration_test.go      (Integration tests - Lines 1-397)
├── types/
│   ├── keys.go                  (Store keys - Lines 1-106)
│   └── errors.go                (Errors - Lines 1-90)
└── module.go                    (Module definition - Lines 1-445)
```

### Protocol Buffers
```
proto/aura/walletsecurity/v1beta1/
├── wallet_security.proto        (Core types - Lines 1-322)
├── tx.proto                     (Msg service - Lines 1-234)
└── query.proto                  (Query service - Lines 1-93)
```

## Quick API Reference

### Hardware Wallet

**Register Device**
```go
RegisterHardwareWallet(
    address string,
    hwType HardwareWalletType,
    deviceID string,
    firmwareVersion string,
    derivationPath string,
    signature []byte,
) (*HardwareWalletConfig, error)
```

**Supported Types**
- LEDGER
- TREZOR
- KEEPKEY
- COLDCARD

### Multi-Signature

**Create Wallet**
```go
CreateMultiSigWallet(
    creator string,
    signers []string,
    threshold int32,
    signerWeights map[string]int32,  // optional
    weightThreshold int32,            // optional
    timeLock *durationpb.Duration,   // optional
) (*MultiSigWallet, error)
```

**Sign Transaction**
```go
SignMultiSigTransaction(
    txID string,
    signer string,
    signature []byte,
) (bool, error)  // returns true when ready to execute
```

### Social Recovery

**Configure Recovery**
```go
ConfigureSocialRecovery(
    walletID string,
    guardians []*Guardian,
    recoveryThreshold int32,
    recoveryDelay *durationpb.Duration,
) (*SocialRecoveryConfig, error)
```

**Initiate Recovery**
```go
InitiateRecovery(
    walletID string,
    newAddress string,
    initiator string,  // must be confirmed guardian
) (*RecoveryRequest, error)
```

### Transaction Simulation

```go
SimulateTransaction(
    txData []byte,
    sender string,
) (*TransactionSimulation, error)
```

**Risk Levels**
- LOW - Normal transaction
- MEDIUM - Moderate complexity
- HIGH - Complex or unusual
- CRITICAL - Dangerous transaction

### Spending Limits

**Set Limits**
```go
SetSpendingLimit(
    walletID string,
    denom string,
    dailyLimit string,
    weeklyLimit string,
    monthlyLimit string,
) (*SpendingLimit, error)
```

**Check Limit**
```go
CheckSpendingLimit(
    walletID string,
    denom string,
    amount string,
) error  // returns error if exceeded
```

### Session Management

**Configure Session**
```go
ConfigureSession(
    walletID string,
    timeoutDuration *durationpb.Duration,
    autoLockEnabled bool,
    inactivityThreshold int32,
) (*SessionConfig, error)
```

**Lock/Unlock**
```go
LockSession(sessionID string) error
UnlockSession(sessionID string, authProof []byte) error
```

### Biometric Authentication

**Enroll**
```go
EnrollBiometric(
    walletID string,
    biometricType BiometricType,
    enrollmentData []byte,
) (*BiometricAuth, error)
```

**Authenticate**
```go
AuthenticateBiometric(
    walletID string,
    biometricProof []byte,
) (bool, error)
```

**Types**
- FINGERPRINT
- FACE
- IRIS
- VOICE

### Secure Enclave

```go
StoreInSecureEnclave(
    walletID string,
    enclaveType EnclaveType,
    encryptedKeyMaterial []byte,
    attestationCert string,
) (*SecureEnclaveConfig, error)
```

**Enclave Types**
- TEE (Trusted Execution Environment)
- SGX (Intel SGX)
- KEYCHAIN (iOS)
- KEYSTORE (Android)
- TPM (Trusted Platform Module)

### Encrypted Backup

```go
CreateEncryptedBackup(
    walletID string,
    encryptedSeed []byte,
    encryptionAlgo string,
    kdf string,
    salt []byte,
    iterations int32,
    location BackupLocation,
) (*EncryptedBackup, error)
```

**Recommended Settings**
- Algorithm: "AES-256-GCM"
- KDF: "PBKDF2-SHA256" or "Argon2id"
- Iterations: 100,000+ (PBKDF2) or 3+ (Argon2)

### Dust Filter

**Configure**
```go
ConfigureDustFilter(
    walletID string,
    enabled bool,
    minimumAmount string,
    maxDustTxPerBlock int32,
    suspiciousThreshold int32,
) (*DustAttackFilter, error)
```

**Check Transaction**
```go
CheckDustTransaction(
    walletID string,
    txHash string,
    fromAddress string,
    toAddress string,
    amount string,
    denom string,
) (bool, error)  // returns true if dust
```

### Address Checksum

```go
ValidateAddressChecksum(
    address string,
    algorithm ChecksumAlgorithm,
) (bool, string, error)
```

**Algorithms**
- EIP55 (Ethereum)
- BECH32 (Cosmos/Bitcoin)
- BASE58CHECK (Bitcoin)
- CRC32

### Phishing Protection

```go
VerifyDomain(
    domain string,
    certificateHash string,
    verifier string,
) (*DomainVerification, error)
```

## Common Patterns

### Complete Wallet Setup
```go
// 1. Register hardware wallet
hwConfig, _ := keeper.RegisterHardwareWallet(...)

// 2. Configure spending limits
keeper.SetSpendingLimit(walletID, "uatom", "1000000", "7000000", "30000000")

// 3. Set up social recovery
guardians := []*Guardian{...}
keeper.ConfigureSocialRecovery(walletID, guardians, 2, delay)

// 4. Enable biometric
keeper.EnrollBiometric(walletID, FINGERPRINT, data)

// 5. Create backup
keeper.CreateEncryptedBackup(walletID, encrypted, algo, kdf, salt, iterations, location)
```

### Secure Transaction Flow
```go
// 1. Simulate transaction
simulation, _ := keeper.SimulateTransaction(txData, sender)
if simulation.RiskLevel >= HIGH {
    // Warn user
}

// 2. Check spending limit
if err := keeper.CheckSpendingLimit(walletID, denom, amount); err != nil {
    // Transaction exceeds limit
}

// 3. Check dust filter
if isDust, _ := keeper.CheckDustTransaction(...); isDust {
    // Block transaction
}

// 4. Validate recipient address
valid, _, _ := keeper.ValidateAddressChecksum(recipient, EIP55)
if !valid {
    // Warn about invalid checksum
}

// 5. Execute transaction
```

### Recovery Workflow
```go
// 1. Configure recovery
keeper.ConfigureSocialRecovery(walletID, guardians, threshold, delay)

// 2. Guardians confirm
for _, guardian := range guardians {
    keeper.ConfirmGuardian(walletID, guardian.Address)
}

// 3. If keys lost, initiate recovery
request, _ := keeper.InitiateRecovery(walletID, newAddress, guardian1)

// 4. Other guardians approve
keeper.ApproveRecovery(requestID, guardian2, signature)

// 5. Wait for delay period

// 6. Execute recovery
keeper.ExecuteRecovery(requestID)
```

## Error Handling

### Common Errors
```go
ErrHardwareWalletNotFound
ErrInsufficientSignatures
ErrRecoveryNotEnabled
ErrSpendingLimitExceeded
ErrSessionExpired
ErrBiometricLockedOut
ErrDustTransactionBlocked
```

### Error Checking Pattern
```go
if err := keeper.SomeOperation(...); err != nil {
    switch err {
    case types.ErrSpendingLimitExceeded:
        // Handle limit exceeded
    case types.ErrSessionExpired:
        // Require re-authentication
    default:
        // Handle other errors
    }
}
```

## Testing

### Run All Tests
```bash
cd chain/x/walletsecurity
go test ./keeper/... -v
```

### Run Specific Test
```bash
go test ./keeper/... -run TestHardwareWallet -v
```

### Integration Tests
```bash
go test ./keeper/... -run TestComplete -v
```

## Security Best Practices

### 1. Hardware Wallets
- Always verify device signature
- Check firmware version
- Require PIN for transactions

### 2. Multi-Sig
- Minimum threshold of 2
- Use time locks for high-value wallets
- Regular key rotation

### 3. Social Recovery
- Minimum 48-hour delay
- Diverse guardian set
- Confirm all guardians

### 4. Spending Limits
- Set conservative daily limits
- Review and adjust monthly
- Monitor spending patterns

### 5. Session Management
- 15-30 minute timeout
- Auto-lock on inactivity
- Device fingerprinting

### 6. Biometric
- Never store raw biometric data
- Implement rate limiting
- 30-minute lockout after failures

### 7. Backups
- Use strong encryption (AES-256)
- High iteration count (100k+)
- Multiple backup locations

### 8. Dust Filter
- Set reasonable minimum (0.001 ATOM)
- Monitor blocked transactions
- Update blacklist regularly

## Performance Considerations

### Storage
- Total storage per wallet: ~5KB
- Indexed by wallet ID
- Efficient key-value lookups

### Gas Costs
- Hardware wallet registration: ~50k gas
- Multi-sig creation: ~100k gas
- Transaction simulation: ~30k gas
- Most operations: <50k gas

### Scalability
- All operations O(1) or O(n) where n is small
- No unbounded loops
- Efficient proto serialization

## Integration Checklist

- [ ] Add module to app.go
- [ ] Register store keys
- [ ] Generate proto files (buf generate)
- [ ] Run tests
- [ ] Update genesis
- [ ] Configure parameters
- [ ] Deploy and test
