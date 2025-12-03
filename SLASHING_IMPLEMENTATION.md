# Bridge Validator Slashing Implementation

## Overview

This document describes the comprehensive validator slashing mechanism implemented for the Aura bridge module to address **TODO 072: Bridge No Validator Slashing**.

## Problem Statement

**Before Implementation:**
- Malicious validators faced no punishment for signing fraudulent transfers
- No economic deterrent against validator collusion
- Bridge security relied solely on validator honesty
- Validators could sign fake deposits without consequences

**Impact:**
- High risk of validator fraud
- No accountability for malicious behavior
- Weak security model for cross-chain transfers

## Solution

A comprehensive validator slashing system with:
1. **Multiple slashing conditions** (fraud, double-signing, downtime)
2. **Configurable slashing amounts** per offense type
3. **Evidence submission mechanism** for fraud detection
4. **Automatic integration** with fraud proof resolution
5. **Validator jailing** to remove bad actors from the active set

---

## Implementation Details

### 1. Slashing Conditions

Four types of slashable offenses:

| Slash Reason | Description | Default Slash | Jailed? |
|--------------|-------------|---------------|---------|
| **SLASH_FRAUD_ATTEMPT** | Signing fraudulent transfers (invalid Merkle proofs, fake deposits) | 50% of stake | Yes |
| **SLASH_DOUBLE_SIGN** | Signing conflicting messages for the same transfer | 100% of stake (tombstone) | Yes (permanent) |
| **SLASH_DOWNTIME** | Failing to maintain minimum uptime | 1% of stake | No |
| **SLASH_UNAUTHORIZED_MINT** | Attempting unauthorized token minting | 50% of stake | Yes |

### 2. Parameters

Slashing parameters are configurable via governance:

```go
type Params struct {
    // Slashing fractions (decimal strings: "0.50" = 50%)
    SlashFraudSignature string // Default: "0.50"
    SlashDoubleSigning  string // Default: "1.00"
    SlashOffline        string // Default: "0.01"

    // Liveness tracking for downtime slashing
    MinSigningWindow    int64  // Default: 10000 blocks (~18 hours)
    MinSignedPerWindow  string // Default: "0.50" (50% uptime required)
}
```

**Parameter Validation:**
- Slash fractions must be between 0.0 and 1.0
- Signing window must be between 100 and 100,000 blocks
- All parameters are validated on update

### 3. Evidence Submission

Anyone can submit evidence of validator misbehavior:

```go
func (k Keeper) SubmitSlashingEvidence(
    ctx sdk.Context,
    validatorAddress string,
    reason SlashReason,
    transferId string,
    evidenceHash []byte,
    submitter string,
) (*SlashingEvent, error)
```

**Evidence Requirements:**
- Non-empty evidence data
- Valid validator address (must exist in bridge validator set)
- Appropriate evidence format for the slash reason
- Evidence proves the claimed infraction

**Security Checks:**
- Validator exists and is active
- Evidence is properly formatted
- No double-jeopardy (prevent slashing twice for same infraction)
- Transfer exists if referenced

### 4. Fraud Proof Integration

**Automatic Slashing on Fraud Proof Validation:**

When a fraud proof is validated as correct, the system automatically:

1. **Retrieves all validators** who signed the fraudulent transfer
2. **Slashes each validator** using `SlashFraudSignature` fraction
3. **Jails all validators** to remove them from active set
4. **Records slashing events** for audit trail
5. **Pays reward** to fraud proof challenger

```go
// In ResolveFraudProof when valid=true:
if err := k.slashValidatorsForFraudulentTransfer(ctx, transferID, proof.ProofId); err != nil {
    // Log error but continue - fraud proof is still valid
}
```

### 5. Double-Signing Detection

**Automatic detection and slashing:**

```go
func (k Keeper) DetectDoubleSigning(
    ctx sdk.Context,
    transferId string,
    newSignature []byte,
    validatorAddress string,
) (bool, error)
```

**Detection Logic:**
- Checks if validator has previously signed the same transfer
- Compares new signature with existing signature
- If signatures differ → **DOUBLE-SIGNING DETECTED**
- Automatically slashes the validator with 100% fraction
- Permanently jails the validator (tombstone)

**Integration Point:**
Call before accepting any validator signature:

```go
isValid, err := k.CheckAndSlashDoubleSigning(ctx, transferID, signature, validatorAddr)
if !isValid {
    return err // Signature rejected, validator slashed
}
```

### 6. Liveness Tracking

**Downtime Slashing:**

The system tracks whether each validator signs transfers over a configurable window:

```go
// Record signing at each block
func (k Keeper) RecordValidatorSigning(ctx sdk.Context, validatorAddress string, signed bool)

// Check liveness
func (k Keeper) CheckValidatorLiveness(ctx sdk.Context, validatorAddress string) (bool, error)

// Slash for downtime
func (k Keeper) SlashForDowntime(ctx sdk.Context, validatorAddress string) error
```

**How It Works:**
1. Every block, record whether validator signed
2. Maintain rolling window of last N blocks (configurable)
3. Calculate percentage of blocks signed
4. If below threshold, slash for downtime
5. Can be triggered manually or automatically in EndBlocker

### 7. Validator Jailing

**Jailing Process:**

```go
func (k Keeper) jailValidator(ctx sdk.Context, validatorAddress string) error
```

**Effects:**
1. **Bridge Module:** Sets validator's `Active` flag to `false`
2. **Staking Module:** Calls `stakingKeeper.Jail()` to jail in consensus
3. **Removal:** Validator can no longer sign bridge operations
4. **Recovery:** Requires governance action to unjail

**Unjailing:**

```go
func (k Keeper) UnjailValidator(ctx sdk.Context, validatorAddr string, authorizedBy string) error
```

Only authorized addresses (governance) can unjail validators.

---

## API Reference

### Core Functions

#### SubmitSlashingEvidence
Submit evidence of validator misbehavior.

**Parameters:**
- `validatorAddress`: Address of validator to slash
- `reason`: SlashReason enum (FRAUD, DOUBLE_SIGN, DOWNTIME, etc.)
- `transferId`: Related transfer ID (if applicable)
- `evidenceHash`: Hash of evidence data
- `submitter`: Address submitting evidence

**Returns:**
- `*SlashingEvent`: Created slashing event
- `error`: If validation fails

#### slashValidatorsForFraudulentTransfer
Internal function called when fraud proof is validated.

**Parameters:**
- `transferID`: The fraudulent transfer
- `fraudProofID`: The validated fraud proof

**Returns:**
- `error`: If slashing fails

#### DetectDoubleSigning
Check if validator is attempting to double-sign.

**Parameters:**
- `transferId`: Transfer being signed
- `newSignature`: Signature being submitted
- `validatorAddress`: Validator submitting signature

**Returns:**
- `bool`: true if double-signing detected
- `error`: If validation fails

#### CheckAndSlashDoubleSigning
Detect and automatically slash double-signing.

**Parameters:**
- `transferId`: Transfer being signed
- `signature`: Signature being submitted
- `validatorAddress`: Validator submitting signature

**Returns:**
- `bool`: true if signature is valid (not double-sign)
- `error`: If double-signing detected or validation fails

### Query Functions

#### GetSlashingEvent
Retrieve a slashing event by ID.

```go
func (k Keeper) GetSlashingEvent(ctx sdk.Context, eventId string) (*SlashingEvent, bool)
```

#### GetAllSlashingEvents
Retrieve all slashing events.

```go
func (k Keeper) GetAllSlashingEvents(ctx sdk.Context) []*SlashingEvent
```

#### GetValidatorSlashingHistory
Retrieve all slashing events for a specific validator.

```go
func (k Keeper) GetValidatorSlashingHistory(ctx sdk.Context, validatorAddress string) []*SlashingEvent
```

#### GetValidatorSigningInfo
Retrieve liveness information for a validator.

```go
func (k Keeper) GetValidatorSigningInfo(ctx sdk.Context, validatorAddress string) map[int64]bool
```

---

## Events

### EventTypeValidatorSlashed

Emitted when a validator is slashed.

**Attributes:**
- `event_id`: Unique slashing event ID
- `validator`: Validator address
- `reason`: Slash reason (fraud, double-sign, downtime)
- `slash_amount`: Amount slashed (in base units)
- `jailed`: Whether validator was jailed (true/false)
- `transfer_id`: Related transfer (if applicable)
- `submitter`: Who submitted the evidence

**Example:**
```json
{
  "type": "validator_slashed",
  "attributes": [
    {"key": "event_id", "value": "slash-aura1val-12345"},
    {"key": "validator", "value": "aura1validator123"},
    {"key": "reason", "value": "SLASH_FRAUD_ATTEMPT"},
    {"key": "slash_amount", "value": "50000000"},
    {"key": "jailed", "value": "true"},
    {"key": "transfer_id", "value": "transfer-456"},
    {"key": "submitter", "value": "fraud-proof:proof-789"}
  ]
}
```

---

## Storage

### SlashingEvent

**Key:** `SlashingEventPrefix + eventId`

**Value:**
```protobuf
message SlashingEvent {
  string event_id = 1;
  string validator_address = 2;
  SlashReason reason = 3;
  string slash_amount = 4; // cosmos.base.v1beta1.Int
  bytes evidence_hash = 5;
  uint64 infraction_height = 6;
  google.protobuf.Timestamp timestamp = 7;
  bool jailed = 8;
}
```

### ValidatorSigningInfo

**Key:** `ValidatorSigningPrefix + validatorAddress + "-" + blockHeight`

**Value:** Single byte (1 = signed, 0 = missed)

Used for liveness tracking over the signing window.

---

## Testing

Comprehensive test coverage in `/chain/x/bridge/keeper/slashing_test.go`:

### Test Categories

1. **Evidence Submission Tests**
   - `TestSubmitSlashingEvidence_FraudSignature`
   - `TestSubmitSlashingEvidence_DoubleSigning`
   - `TestSubmitSlashingEvidence_Downtime`
   - `TestSubmitSlashingEvidence_ValidatorNotFound`
   - `TestSubmitSlashingEvidence_InvalidEvidence`
   - `TestSubmitSlashingEvidence_AlreadyJailed`

2. **Double-Signing Detection Tests**
   - `TestDetectDoubleSigning_SameSignature`
   - `TestDetectDoubleSigning_DifferentSignature`
   - `TestDetectDoubleSigning_FirstSignature`
   - `TestCheckAndSlashDoubleSigning_DetectsAndSlashes`

3. **Liveness Tracking Tests**
   - `TestRecordValidatorSigning_Signed`
   - `TestRecordValidatorSigning_Missed`
   - `TestCheckValidatorLiveness_MeetsRequirement`
   - `TestCheckValidatorLiveness_FailsRequirement`
   - `TestSlashForDowntime_ValidatorOnline`
   - `TestSlashForDowntime_ValidatorOffline`

4. **Fraud Proof Integration Tests**
   - `TestSlashValidatorsForFraudulentTransfer_Success`
   - `TestSlashValidatorsForFraudulentTransfer_NoSignatures`
   - `TestSlashValidatorsForFraudulentTransfer_TransferNotFound`
   - `TestSlashValidatorsForFraudulentTransfer_PartialFailure`
   - `TestResolveFraudProof_SlashesValidators`

5. **Query Tests**
   - `TestGetValidatorSlashingHistory`
   - `TestGetAllSlashingEvents`

6. **Parameter Tests**
   - `TestSlashingParams_ValidFractions`
   - `TestSlashingParams_SigningWindow`

---

## Security Considerations

### 1. Double-Jeopardy Prevention

Validators cannot be slashed twice for the same infraction:

```go
func (k Keeper) isAlreadySlashed(
    ctx sdk.Context,
    validatorAddr string,
    reason SlashReason,
    infractionHeight uint64,
) bool
```

### 2. Evidence Validation

All evidence is validated before slashing:
- Validator exists
- Evidence format is correct
- Infraction height is reasonable (not in future, not too old)
- Transfer exists if referenced

### 3. Partial Failure Handling

When slashing multiple validators (fraud proof resolution):
- Continue slashing other validators if one fails
- Log errors for failed slashes
- Succeed if at least one validator is slashed
- Return error only if ALL validators fail to slash

### 4. Staking Module Integration

Integration with Cosmos SDK staking module:
- Converts bridge validator addresses to staking validator addresses
- Retrieves validator power for slash calculations
- Executes slash via `stakingKeeper.Slash()`
- Jails validators via `stakingKeeper.Jail()`

**Fallback:** If staking keeper is not available (tests or separate validator sets), operations are no-ops but don't fail.

### 5. Access Control

**Governance-Only Operations:**
- Updating slashing parameters
- Unjailing validators

**Public Operations:**
- Submitting slashing evidence (permissionless)
- Viewing slashing events (queries)

---

## Migration Guide

### Upgrading Existing Chains

1. **Add Parameters:**
   ```go
   params := k.GetParams(ctx)
   params.SlashFraudSignature = "0.50"
   params.SlashDoubleSigning = "1.00"
   params.SlashOffline = "0.01"
   params.MinSigningWindow = 10000
   params.MinSignedPerWindow = "0.50"
   k.SetParams(ctx, params)
   ```

2. **No State Migration Required:**
   - Slashing events are new (no existing data)
   - Validator signing info starts fresh
   - Existing validators are not affected

3. **Monitor Initial Period:**
   - Watch for false positives in liveness tracking
   - Tune `MinSignedPerWindow` if needed
   - Verify slash amounts are appropriate

---

## Operational Procedures

### Investigating Slashing Events

```bash
# Query all slashing events
aurad q bridge slashing-events

# Query specific validator's history
aurad q bridge validator-slashing-history <validator-address>

# Query specific slashing event
aurad q bridge slashing-event <event-id>
```

### Unjailing a Validator

**Via Governance Proposal:**
```bash
aurad tx gov submit-proposal unjail-validator \
  --validator=<validator-address> \
  --from=<proposer> \
  --deposit=10000000aura
```

**After Approval:**
The proposal execution will call `UnjailValidator` and restore the validator to active status.

### Adjusting Parameters

**Via Governance Proposal:**
```json
{
  "title": "Update Bridge Slashing Parameters",
  "description": "Adjust fraud slashing to 60%",
  "changes": [
    {
      "subspace": "bridge",
      "key": "SlashFraudSignature",
      "value": "\"0.60\""
    }
  ]
}
```

---

## Future Enhancements

Potential improvements for future versions:

1. **Graduated Penalties:**
   - First offense: Warning
   - Second offense: Small slash
   - Third offense: Large slash + jail

2. **Slashing Rewards:**
   - Reward evidence submitters
   - Create incentive for fraud detection

3. **Historical Power Tracking:**
   - Use validator power at infraction height
   - More accurate slash calculations

4. **Reputation System:**
   - Track validator reliability score
   - Weight attestations by reputation

5. **Automated Fraud Detection:**
   - Monitor for suspicious patterns
   - Auto-submit fraud proofs

6. **Insurance Fund Integration:**
   - Slashed funds go to insurance pool
   - Compensate users affected by fraud

---

## Acceptance Criteria

All requirements from TODO 072 have been met:

- [x] **Slashing conditions defined**
  - FRAUD_ATTEMPT, DOUBLE_SIGN, DOWNTIME, UNAUTHORIZED_MINT

- [x] **Slashing amounts configurable in params**
  - SlashFraudSignature, SlashDoubleSigning, SlashOffline
  - Validated on update

- [x] **Evidence submission mechanism**
  - `SubmitSlashingEvidence` function
  - Comprehensive evidence validation

- [x] **SlashValidator and JailValidator functions**
  - `slashValidator` via staking keeper
  - `jailValidator` removes from active set

- [x] **Tests for slashing scenarios**
  - 20+ test cases covering all scenarios
  - Unit tests and integration tests

- [x] **Fraud proof integration**
  - Automatic slashing on fraud proof validation
  - Slashes all validators who signed fraudulent transfers

- [x] **Economic deterrent achieved**
  - 50% slash for fraud (configurable)
  - 100% slash for double-signing
  - Validators face real economic consequences

---

## References

- **Implementation:** `/chain/x/bridge/keeper/slashing.go`
- **Tests:** `/chain/x/bridge/keeper/slashing_test.go`
- **Parameters:** `/chain/x/bridge/types/params.go`
- **Proto Definitions:** `/proto/aura/bridge/v1beta1/security.proto`
- **Events:** `/chain/x/bridge/types/events.go`

---

## Conclusion

The validator slashing implementation provides a robust economic deterrent against malicious behavior in the Aura bridge. Validators now face real consequences for fraud, double-signing, and downtime, significantly improving bridge security.

**Key Achievements:**
- ✅ Comprehensive slashing for multiple offense types
- ✅ Configurable parameters via governance
- ✅ Automatic integration with fraud proofs
- ✅ Liveness tracking for uptime enforcement
- ✅ Extensive test coverage
- ✅ Production-ready implementation

The bridge security model now relies on **cryptoeconomic security** rather than just validator honesty.
