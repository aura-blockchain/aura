# Bridge Security Features - Quick Reference Guide

## File Locations

### Core Implementation Files
| Feature | File Path | Lines |
|---------|-----------|-------|
| Merkle Proofs | `chain/x/bridge/keeper/security_merkle.go` | 1-219 |
| TSS Signatures | `chain/x/bridge/keeper/security_tss.go` | 1-203 |
| Validators & Slashing | `chain/x/bridge/keeper/security_validators.go` | 1-285 |
| Fraud Proofs | `chain/x/bridge/keeper/security_fraud.go` | 1-273 |
| Limits & Circuit Breaker | `chain/x/bridge/keeper/security_limits.go` | 1-375 |
| Access Control | `chain/x/bridge/keeper/security_access.go` | 1-243 |
| Fees & Insurance | `chain/x/bridge/keeper/security_fees.go` | 1-273 |

### Configuration Files
| Type | File Path |
|------|-----------|
| Security Parameters | `chain/x/bridge/types/params_security.go` |
| Storage Keys | `chain/x/bridge/types/keys_security.go` |
| Error Definitions | `chain/x/bridge/types/errors_security.go` |
| Proto Definitions | `proto/aura/bridge/v1beta1/security.proto` |

### Test Files
| Type | File Path |
|------|-----------|
| Merkle Tests | `chain/x/bridge/keeper/security_merkle_test.go` |
| Integration Tests | `chain/x/bridge/keeper/security_integration_test.go` |

---

## Default Security Parameters

```go
EmergencyPaused:              false
MinTransferAmount:            1 token
MaxTransferAmount:            100,000 tokens
TimeLockDuration:             24 hours
TimeLockThreshold:            10,000 tokens
DailyWithdrawalLimit:         50,000 tokens
CircuitBreakerEnabled:        true
MaxHourlyVolume:              500,000 tokens/hour
MaxFailedTransfersPerHour:    10
MinValidatorSignatures:       3
ValidatorRotationPeriod:      30 days
SlashFractionInvalidProof:    5%
SlashFractionDoubleSign:      10%
SlashFractionDowntime:        1%
FraudProofReward:             1,000 tokens
FraudProofWindowDuration:     7 days
FixedTransferFee:             0.1 token
PercentageFeeBPS:             10 (0.1%)
InsuranceFundContributionBPS: 2000 (20%)
WhitelistEnabled:             false
```

---

## Key Functions by Feature

### 1. Merkle Proof Verification
```go
VerifyMerkleProof(ctx, proof, transferData) error
StoreMerkleRoot(ctx, chainId, blockHeight, root) error
GetMerkleRoot(ctx, chainId, blockHeight) []byte
VerifyTransferWithProof(ctx, transfer, proof) error
ComputeMerkleRoot(transfers) []byte
GenerateMerkleProof(transfers, leafIndex) (*MerkleProof, error)
```

### 2. TSS Signatures
```go
VerifyTSSSignature(ctx, message, signature) error
CreateTSSSignature(ctx, message, signers) (*TSSSignature, error)
GetTSSNonce(ctx) uint64
SetTSSNonce(ctx, nonce)
VerifyValidatorSignature(ctx, message, validatorAddr, signature) error
AggregateValidatorSignatures(ctx, transferId, signatures) (*TSSSignature, error)
```

### 3. Validator Management
```go
InitiateValidatorRotation(ctx, newValidators) (*ValidatorSetRotation, error)
ApproveValidatorRotation(ctx, rotationId, validatorAddr, signature) error
ExecuteValidatorRotation(ctx, rotationId) error
CheckAndExecutePendingRotations(ctx)
```

### 4. Slashing
```go
SlashValidatorForInvalidProof(ctx, validatorAddr, transferId) error
SlashValidatorForDoubleSign(ctx, validatorAddr, evidence) error
```

### 5. Fraud Proofs
```go
SubmitFraudProof(ctx, challenger, transferId, fraudType, evidence) (*FraudProof, error)
InvestigateFraudProof(ctx, proofId, investigator) error
ResolveFraudProof(ctx, proofId, isValid, resolution) error
CheckExpiredFraudProofs(ctx)
```

### 6. Time-Locks
```go
CreateTimeLock(ctx, transferId, amount, denom, recipient) (*TimeLock, error)
ProcessTimeLocks(ctx)
ChallengeTimeLock(ctx, lockId, challenger, evidence) error
```

### 7. Withdrawal Limits
```go
CheckWithdrawalLimit(ctx, address, amount) error
UpdateWithdrawalLimit(ctx, address, newLimit, tier) error
```

### 8. Circuit Breaker
```go
UpdateCircuitBreaker(ctx, transferAmount, transferFailed) error
TripCircuitBreaker(ctx, breaker) error
ResetCircuitBreaker(ctx) error
CheckAutoResetCircuitBreaker(ctx)
```

### 9. Nonce Management
```go
GetNonce(ctx, address, chainId) uint64
IncrementNonce(ctx, address, chainId) uint64
VerifyNonce(ctx, address, chainId, nonce) error
VerifyAndIncrementNonce(ctx, address, chainId, nonce) error
```

### 10. Emergency Pause
```go
PauseBridge(ctx, reason, pauser) error
UnpauseBridge(ctx, unpauser) error
IsBridgePaused(ctx) bool
```

### 11. Permissions
```go
AddToWhitelist(ctx, address, reason, addedBy) error
AddToBlacklist(ctx, address, reason, addedBy, expiresAt) error
RemoveFromPermissionList(ctx, address) error
CheckAddressPermission(ctx, address) error
CleanupExpiredPermissions(ctx)
```

### 12. Fees
```go
CalculateTransferFee(ctx, amount, feeType) (sdk.Int, error)
CollectTransferFee(ctx, sender, amount, feeType) (sdk.Int, error)
DistributeFees(ctx, totalFee)
```

### 13. Insurance Fund
```go
SubmitInsuranceClaim(ctx, claimant, transferId, claimAmount, reason, evidence) (*InsuranceClaim, error)
ApproveInsuranceClaim(ctx, claimId, approvedAmount) error
RejectInsuranceClaim(ctx, claimId, reason) error
```

---

## Security Event Triggers

### Circuit Breaker Trips When:
- Single transfer > MaxTransferAmount
- Hourly volume > MaxHourlyVolume (500,000 tokens)
- Failed transfers > MaxFailedTransfersPerHour (10)

### Time-Lock Required When:
- Transfer amount >= TimeLockThreshold (10,000 tokens)

### Slashing Occurs When:
- Invalid Merkle proof submitted (5% slash)
- Double-signing detected (10% slash, jail)
- Fraud proof validated (10% slash, jail)
- Excessive downtime (1% slash)

### Fraud Proof Can Be Submitted:
- Within 7 days of transfer
- By any address
- With supporting evidence
- Rewards 1,000 tokens if valid

---

## Storage Key Prefixes

```go
MerkleRootPrefix         = 0x10
TSSNoncePrefix           = 0x11
BridgeValidatorPrefix    = 0x12
ValidatorRotationPrefix  = 0x13
SlashingEventPrefix      = 0x14
FraudProofPrefix         = 0x15
TimeLockPrefix           = 0x16
WithdrawalLimitPrefix    = 0x17
CircuitBreakerPrefix     = 0x18
NonceTrackerPrefix       = 0x19
AddressPermissionPrefix  = 0x1a
BridgeFeePrefix          = 0x1b
InsuranceFundPrefix      = 0x1c
InsuranceClaimPrefix     = 0x1d
```

---

## Common Error Types

```go
ErrBridgePaused
ErrInvalidMerkleProof
ErrInvalidTSSSignature
ErrInsufficientSignatures
ErrAmountExceedsMaximum
ErrDailyLimitExceeded
ErrCircuitBreakerOpen
ErrAddressBlacklisted
ErrFraudProofExpired
ErrInsufficientInsuranceFund
```

---

## Proto Message Types

### Security Messages
```protobuf
MerkleProof
TSSSignature
ValidatorSetRotation
SlashingEvent
FraudProof
TimeLock
WithdrawalLimit
CircuitBreaker
NonceTracker
AddressPermission
BridgeFee
InsuranceFund
InsuranceClaim
```

### Enums
```protobuf
RotationStatus (PENDING, APPROVED, ACTIVE, EXPIRED)
SlashReason (INVALID_PROOF, DOUBLE_SIGN, FRAUD, DOWNTIME)
FraudType (INVALID_PROOF, DOUBLE_SPEND, INVALID_SIG, AMOUNT_MISMATCH)
FraudProofStatus (PENDING, INVESTIGATING, VALID, INVALID, EXPIRED)
TimeLockStatus (LOCKED, UNLOCKED, CHALLENGED, CANCELLED)
CircuitBreakerStatus (CLOSED, OPEN, HALF_OPEN)
PermissionType (NONE, WHITELISTED, BLACKLISTED)
FeeType (TRANSFER, MINT_WRAPPED, BURN_WRAPPED, FAST_TRANSFER)
ClaimStatus (PENDING, INVESTIGATING, APPROVED, REJECTED, PAID)
```

---

## Fee Structure

### Transfer Fees
- **Fixed Fee:** 0.1 token
- **Percentage Fee:** 0.1% (10 bps)
- **Maximum Fee:** 5% of amount
- **Minimum Fee:** 0.1 token

### Fee Distribution
- **Insurance Fund:** 20% of fees
- **Validators:** 80% of fees

---

## Monitoring Checklist

### Daily Checks
- [ ] Circuit breaker status
- [ ] Active fraud proofs
- [ ] Pending time-locks
- [ ] Insurance fund balance
- [ ] Failed transfer count

### Weekly Checks
- [ ] Validator rotation schedule
- [ ] Slashing events
- [ ] Insurance claims
- [ ] Withdrawal limit adjustments

### Monthly Checks
- [ ] Security parameter tuning
- [ ] Validator set health
- [ ] Insurance fund sustainability
- [ ] Fee revenue analysis

---

## Emergency Response

### If Circuit Breaker Trips:
1. Check logs for trigger reason
2. Investigate recent transfers
3. Assess if malicious or legitimate spike
4. Reset manually or wait for auto-reset

### If Fraud Proof Submitted:
1. Pause affected transfer
2. Assign investigators
3. Review evidence
4. Resolve within 7 days

### If Validator Slashed:
1. Review slashing event
2. Remove validator if jailed
3. Notify validator operator
4. Update validator set if needed

### Emergency Pause:
1. Call `PauseBridge(ctx, reason, admin)`
2. Investigate issue
3. Implement fix
4. Test thoroughly
5. Call `UnpauseBridge(ctx, admin)`

---

## Integration Points

### With Other Modules
- **Bank Module:** Token transfers
- **Staking Module:** Validator power
- **Governance Module:** Parameter updates
- **VC Registry Module:** Identity verification

### External Systems
- **PAW Chain:** Cross-chain transfers
- **XAI Chain:** Cross-chain transfers
- **Relayers:** Transfer monitoring
- **Block Explorers:** Event tracking

---

## Testing Commands

### Run Merkle Tests
```bash
go test -v ./chain/x/bridge/keeper/security_merkle_test.go
```

### Run Integration Tests
```bash
go test -v ./chain/x/bridge/keeper/security_integration_test.go
```

### Run All Bridge Tests
```bash
cd chain/x/bridge
go test -v ./...
```

---

## Quick Troubleshooting

| Issue | Likely Cause | Solution |
|-------|--------------|----------|
| Transfer rejected | Daily limit exceeded | Wait 24h or increase limit |
| Time-lock active | Amount > 10,000 tokens | Wait 24h for unlock |
| Circuit breaker open | Volume/failures too high | Wait 1h or manual reset |
| Invalid nonce | Replay or wrong sequence | Check nonce tracker |
| Validator slashed | Invalid proof/fraud | Review slashing event |
| Address blocked | Blacklisted | Review permissions |
| Bridge paused | Emergency or CB trip | Check pause reason |

---

## Additional Resources

- Full Implementation: `BRIDGE_SECURITY_IMPLEMENTATION.md`
- Proto Definitions: `proto/aura/bridge/v1beta1/security.proto`
- Test Files: `chain/x/bridge/keeper/security_*_test.go`
