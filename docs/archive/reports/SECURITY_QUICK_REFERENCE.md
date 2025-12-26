# Security Quick Reference - Aura Blockchain
**For Developers & Auditors**

## Critical Security Implementations by Module

### 🔒 Bridge Module - Cross-Chain Security

**Location:** `/home/hudson/blockchain-projects/aura/chain/x/bridge/`

**Key Security Features:**

1. **Multi-Validator Signature Verification**
   - File: `keeper/msg_server.go` (lines 39-153)
   - Function: `verifyRawValidatorSignatures()`
   - Enforces minimum 2 validators (MinAllowedConfirmations)
   - Active validator set checking at current block height
   - Prevents duplicate signature counting

2. **Replay Attack Prevention**
   - File: `keeper/msg_server.go` (lines 346-373, 437-443)
   - Source hash tracking: `IsSourceHashProcessed()`
   - Signature set hashing: `computeSignatureSetHash()`
   - Used in UnlockTokens message handler

3. **Fraud Proof System**
   - File: `keeper/msg_server.go` (lines 1070-1171)
   - Pending transfer escrow during fraud window
   - Challenge submission: `SubmitFraudProof()`
   - Finalization: `FinalizeTransfer()` (lines 948-1046)

4. **Emergency Controls**
   - File: `keeper/msg_server.go` (lines 1195-1280)
   - Emergency pause: `EmergencyPause()` with ACL
   - Auto-pause on mint threshold breach
   - Governance-only unpause

5. **Rate Limiting & Circuit Breakers**
   - File: `keeper/msg_server.go` (lines 552-594)
   - Hourly mint limit check
   - Daily mint limit check
   - Per-transfer maximum (MaxTransferAmount)
   - Supply caps per token

**Security Tests:**
- `chain/x/bridge/types/params_security_test.go` - Parameter validation
- `chain/x/common/testing/intermodule_security_test.go` - Inter-module security

---

### 🆔 Identity Module - Access Control & Auth

**Location:** `/home/hudson/blockchain-projects/aura/chain/x/identity/`

**Key Security Features:**

1. **RBAC (Role-Based Access Control)**
   - File: `keeper/msg_server.go` (lines 152-209)
   - Functions: `CreateRole()`, `AssignRole()`, `RevokeRole()`
   - Role expiry support
   - Permission requirements: `RequirePermission()`

2. **Multisig Wallet Support**
   - File: `keeper/msg_server.go` (lines 211-256)
   - Threshold signature validation (M-of-N)
   - Proposal system with expiry
   - Functions: `CreateMultisigWallet()`, `SignMultisigProposal()`

3. **Time-Locked Actions**
   - File: `keeper/msg_server.go` (lines 415-535)
   - Delay enforcement before execution
   - Cancellation support for pending actions
   - Functions: `ProposeTimeLockedAction()`, `ExecuteTimeLockedAction()`

4. **Emergency Admin Activation**
   - File: `keeper/msg_server.go` (lines 537-602)
   - Temporary privilege grants
   - Time-bound activation
   - Functions: `ActivateEmergencyAdmin()`, `DeactivateEmergencyAdmin()`

5. **GDPR Compliance**
   - File: `keeper/msg_server.go` (lines 711-731)
   - Right to erasure implementation
   - Function: `EraseIdentity()`

**Security Tests:**
- Multiple test suites in `keeper/*_test.go`

---

### 📜 Wasm Module - Smart Contract Security

**Location:** `/home/hudson/blockchain-projects/aura/chain/x/wasm/`

**Key Security Features:**

1. **Contract Pause/Unpause**
   - File: `keeper/security_methods.go` (lines 85-120)
   - Functions: `IsContractPaused()`, `PauseContract()`, `UnpauseContract()`
   - Prevents execution of paused contracts

2. **Upload Authorization**
   - File: `keeper/security_methods.go` (lines 59-82)
   - Whitelist-based upload control
   - Functions: `IsAuthorizedUploader()`, `AuthorizeUploader()`

3. **Contract Validation**
   - File: `keeper/security_methods.go` (lines 125-163)
   - Max code size enforcement
   - Empty code rejection
   - Execution validation before each call

4. **Reentrancy Protection**
   - File: `keeper/security_methods.go` (lines 404-441)
   - Execution context tracking
   - Max depth limit (10)
   - Functions: `getOrCreateExecutionContext()`, `IsContractExecuting()`

5. **Migration Security**
   - File: `keeper/security_methods.go` (lines 290-314)
   - Admin-only migration requirement
   - Function: `ensureMigrationAdmin()`

6. **Security Audit Logging**
   - File: `keeper/security_methods.go` (lines 447-472)
   - Function: `LogSecurityEvent()`
   - Emits security audit events

**Security Tests:**
- `keeper/security_test.go`
- `keeper/wasm_operations_test.go`

---

### 🔐 Cryptography Module

**Location:** `/home/hudson/blockchain-projects/aura/chain/x/cryptography/`

**Key Security Features:**

1. **Key Rotation Scheduling**
   - File: `keeper/msg_server.go` (lines 25-53)
   - Function: `CreateKeyRotationSchedule()`
   - Signer verification before rotation

2. **Key Rotation Authorization**
   - File: `keeper/msg_server.go` (lines 55-95)
   - Ownership verification of rotation schedules
   - Function: `RotateKey()`

**Random Number Generation:**
- ✅ Uses `crypto/rand` throughout (verified in all production code)
- ⚠️ Single usage of `math/rand` in test file only (acceptable)

---

## Security Patterns to Follow

### ✅ Input Validation Pattern

```go
// Always validate messages in ValidateBasic()
func (msg *MsgExample) ValidateBasic() error {
    if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
        return errors.New("amount must be positive")
    }
    if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
        return errors.New("invalid sender address")
    }
    if err := sdk.ValidateDenom(msg.Denom); err != nil {
        return errors.New("invalid denom")
    }
    return nil
}
```

**Where Used:**
- All 29 message types across modules
- Bridge: `types/messages.go`
- Identity: Message validation
- Wasm: Contract validation

---

### ✅ Authorization Pattern

```go
// Check permissions before sensitive operations
func (ms msgServer) SensitiveOperation(ctx context.Context, msg *MsgSensitive) (*MsgResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)

    // Permission check
    if err := ms.Keeper.RequirePermission(sdkCtx, msg.Sender, types.PermissionRequired); err != nil {
        return nil, status.Error(codes.PermissionDenied, err.Error())
    }

    // Proceed with operation
    // ...
}
```

**Where Used:**
- Identity module: 93 permission checks across all modules
- Bridge: Emergency pause authorization
- Wasm: Contract upload authorization

---

### ✅ Reentrancy Protection Pattern

```go
// Checks-Effects-Interactions pattern
func (ms msgServer) TransferFunds(ctx context.Context, msg *MsgTransfer) (*MsgResponse, error) {
    // 1. CHECKS - Validate inputs and state
    if err := validateTransfer(msg); err != nil {
        return nil, err
    }

    // 2. EFFECTS - Update state BEFORE external calls
    ms.Keeper.MarkTransferProcessed(ctx, msg.TransferId)
    ms.Keeper.UpdateBalance(ctx, msg.Sender, msg.Amount.Neg())

    // 3. INTERACTIONS - External calls AFTER state changes
    if err := ms.Keeper.SendCoins(ctx, msg.Sender, msg.Recipient, msg.Amount); err != nil {
        return nil, err
    }

    return &MsgResponse{Success: true}, nil
}
```

**Where Used:**
- Bridge: UnlockTokens (lines 545-652)
- Bridge: FinalizeTransfer (lines 998-1026)
- Wasm: Contract execution hooks

---

### ✅ SafeMath Pattern

```go
import sdkmath "cosmossdk.io/math"

// Always use sdkmath for arithmetic operations
func calculateTotal(amounts []sdkmath.Int) sdkmath.Int {
    total := sdkmath.ZeroInt()
    for _, amount := range amounts {
        total = total.Add(amount)  // Safe addition
    }
    return total
}

// NEVER use raw arithmetic like:
// total = total + amount  // ❌ UNSAFE - can overflow
```

**Where Used:**
- Bridge: All amount calculations
- Wasm: Token supply tracking
- All modules: Financial operations

---

### ✅ Error Handling Pattern

```go
// Use typed errors with context
if err := validateInput(msg); err != nil {
    return nil, errorsmod.Wrap(types.ErrInvalidInput, err.Error())
}

// Emit telemetry for security events
if err != nil {
    telemetry.IncrCounterWithLabels(
        []string{"module", "operation_failed"},
        1,
        []metrics.Label{{Name: "error_type", Value: err.Error()}},
    )
}
```

**Where Used:**
- All modules: Consistent error wrapping
- Identity: Auth denial tracking
- Bridge: Validation failures

---

## Common Vulnerabilities - Prevention Guide

### ❌ Integer Overflow
**Prevention:** Use `sdkmath` instead of raw arithmetic
**Example:** Bridge module line 326 `token.TotalSupply = token.TotalSupply.Add(amount)`

### ❌ Replay Attacks
**Prevention:** Track transaction hashes, signature sets
**Example:** Bridge UnlockTokens lines 370-373, 437-443

### ❌ Reentrancy
**Prevention:** Checks-Effects-Interactions pattern
**Example:** Bridge UnlockTokens lines 545-652

### ❌ Unauthorized Access
**Prevention:** Permission checks, authority validation
**Example:** Identity module lines 225-228, Bridge EmergencyPause lines 1210-1213

### ❌ Input Validation Failures
**Prevention:** ValidateBasic() on all messages
**Example:** Bridge LockTokens lines 160-162

---

## Security Testing Locations

### Fuzz Tests
- `/chain/x/common/testing/intermodule_fuzz_test.go`

### Security-Specific Tests
- `/chain/x/bridge/types/params_security_test.go`
- `/chain/x/common/testing/intermodule_security_test.go`
- `/chain/x/wasm/keeper/security_test.go`
- `/chain/x/vcregistry/keeper/msg_server_security_test.go`

### Integration Tests
- `/chain/x/common/testing/intermodule_message_flow_test.go`
- `/chain/x/wasm/integration_test.go`

---

## Deployment Security

### Secret Management
- **Development:** `k8s/testnet-deploy/secret.yaml` (TEST ONLY - clearly marked)
- **Production:** Use `deployment-security/scripts/generate-secrets.sh`
- **Best Practice:** External secret management (Vault, AWS Secrets Manager)

### Network Policies
- **Location:** `deployment-security/network-policies.yaml`
- Pod-to-pod communication restrictions
- Ingress/egress controls

---

## Emergency Response

### Circuit Breaker Activation

**Bridge Emergency Pause:**
```bash
# Via CLI (requires authorized address)
aurad tx bridge emergency-pause \
  --chains ethereum,binance \
  --reason "Suspected exploit detected" \
  --from <authorized_address>
```

**Contract Pause:**
```bash
# Via governance or admin
aurad tx wasm pause-contract <contract_address> \
  --from <admin_address>
```

### Fraud Proof Submission

```bash
aurad tx bridge submit-fraud-proof \
  --transfer-id <transfer_id> \
  --fraud-type INVALID_MERKLE_PROOF \
  --evidence <evidence_data> \
  --challenger <address>
```

---

## Security Audit History

- **2025-12-22:** Comprehensive internal audit (this report)
  - Status: ✅ TESTNET READY
  - Critical Issues: 0
  - Risk Rating: LOW

---

## Contact for Security Issues

**For security vulnerabilities:**
- Email: security@aura-blockchain.org (when available)
- Current: File issue on private GitHub repo

**For audit inquiries:**
- See `SECURITY_AUDIT_REPORT.md` for full details
- See `SECURITY_FINDINGS_SUMMARY.md` for executive summary

---

**Document Version:** 1.0
**Last Updated:** 2025-12-22
**Maintainer:** Security Team
