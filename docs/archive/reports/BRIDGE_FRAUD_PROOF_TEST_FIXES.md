# Bridge Fraud Proof Test Fixes

## Summary

Fixed multiple test failures in the bridge keeper fraud proof tests related to missing pending transfers and invalid parameter validation.

## Issues Fixed

### 1. Fraud Proof Tests Failing with "fraud proof has expired"

**Problem:**
Tests were failing with error `fraud proof has expired` in:
- `TestSubmitFraudProof`
- `TestSubmitFraudProofDuplicate`
- `TestResolveFraudProofValid`
- `TestResolveFraudProofInvalid`
- `TestResolveFraudProofExpired`

**Root Cause:**
The fraud proof submission logic checks for the existence of a `PendingTransfer` with a valid `UnlockTime`. The tests were only creating `CrossChainTransfer` objects but not the corresponding `PendingTransfer` objects required by the fraud proof system.

The fraud proof security architecture requires:
1. A `CrossChainTransfer` record (the main transfer record)
2. A `PendingTransfer` record with an `UnlockTime` (holds transfer in escrow during fraud proof window)

When `SubmitFraudProof` is called, it:
1. Checks if a `PendingTransfer` exists
2. Checks if the current time is before the `UnlockTime` (fraud proof window not expired)
3. If either check fails, returns `ErrFraudProofExpired`

**Solution:**

Created a new test helper function `seedBridgeTransferWithPending` that creates both:
- The `CrossChainTransfer` record
- A `PendingTransfer` record with `UnlockTime` set to `BlockTime + FraudProofWindow` (7 days in future)

Updated all fraud proof tests to use the new helper instead of `seedBridgeTransfer`.

### 2. Supply Caps Tests Failing with "fraud proof window must be at least 1 hour"

**Problem:**
Tests in `supply_caps_simple_test.go` were failing validation with:
```
fraud proof window must be at least 1 hour (3600 seconds), got 0
```

**Root Cause:**
The test params were missing the new security parameters added to `types.Params`:
- `FraudProofWindow` - must be >= 3600 seconds (1 hour)
- `SlashFraudSignature` - decimal string for slash fraction
- `SlashDoubleSigning` - decimal string for slash fraction
- `SlashOffline` - decimal string for slash fraction
- `MinSigningWindow` - integer for liveness tracking
- `MinSignedPerWindow` - decimal string for minimum signatures

The `Params.Validate()` method enforces these constraints but the test cases weren't providing valid values.

**Solution:**

Added all required security parameters to each test case with valid values:
```go
FraudProofWindow:        3600,   // 1 hour minimum
SlashFraudSignature:     "0.50", // 50%
SlashDoubleSigning:      "1.00", // 100%
SlashOffline:            "0.01", // 1%
MinSigningWindow:        10000,  // blocks
MinSignedPerWindow:      "0.50", // 50%
```

## Files Modified

### 1. `/home/decri/blockchain-projects/aura/chain/x/bridge/keeper/test_helpers_test.go`

**Added:**
- Import `cosmossdk.io/math` as `sdkmath`
- New function `seedBridgeTransferWithPending` that creates both transfer and pending transfer

**Key Implementation Details:**
```go
// Creates pending transfer with unlock time in the future
unlockTime := input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow)

// PendingTransfer.Amount is math.Int, not string
amountInt, ok := sdkmath.NewIntFromString(amount)
if !ok {
    t.Fatalf("invalid amount string: %s", amount)
}

pending := &types.PendingTransfer{
    TransferId:   transferID,
    Recipient:    keepertest.GenTestAddr().String(),
    Amount:       amountInt,  // math.Int type
    Denom:        "uaura",
    SourceChain:  "paw",
    SourceTxHash: "0xabcd1234",
    CreatedAt:    timestamppb.New(input.Ctx.BlockTime()),
    UnlockTime:   timestamppb.New(unlockTime),  // Future time
    Challenged:   false,
    FraudProofId: "",
}
```

### 2. `/home/decri/blockchain-projects/aura/chain/x/bridge/keeper/keeper_extended_test.go`

**Changed:**
All fraud proof tests now use `seedBridgeTransferWithPending` instead of `seedBridgeTransfer`:
- `TestSubmitFraudProof`
- `TestSubmitFraudProofDuplicate`
- `TestSubmitFraudProofWindowExpired`
- `TestResolveFraudProofValid`
- `TestResolveFraudProofInvalid`
- `TestResolveFraudProofExpired`
- `TestFraudProofWindow`

### 3. `/home/decri/blockchain-projects/aura/chain/x/bridge/keeper/supply_caps_simple_test.go`

**Changed:**
Added complete security parameter set to all test cases in `TestSupplyCaps_ValidateParams`:
- "valid supply caps" test case
- "invalid supply cap value" test case
- "invalid daily limit" test case
- "invalid hourly limit" test case
- "empty supply caps map is valid" test case

## Security Architecture Validated

These fixes validate the fraud proof security architecture:

### Fraud Proof Window
- **Purpose:** Provides time for community to submit fraud proofs before finalizing transfers
- **Duration:** Configurable, default 7 days (604,800 seconds)
- **Minimum:** 1 hour (3,600 seconds) - enforced by validation
- **Maximum:** 30 days (2,592,000 seconds) - enforced by validation

### Pending Transfers
- **Purpose:** Holds transfers in escrow during fraud proof window
- **UnlockTime:** Set to creation time + fraud proof window
- **Challenged Flag:** Prevents finalization when fraud proof submitted
- **State:** Separate from main transfer record for security isolation

### Validator Slashing
- **FraudSignature:** 50% slash for signing fraudulent transfers
- **DoubleSigning:** 100% slash (tombstoning) for double-signing
- **Offline:** 1% slash for failing liveness checks
- **Window:** 10,000 blocks for liveness tracking
- **Threshold:** Must sign 50% of blocks in window

## Testing Best Practices Established

1. **Always create PendingTransfer for fraud proof tests** - Use `seedBridgeTransferWithPending`
2. **Set UnlockTime in the future** - Add fraud proof window to current block time
3. **Use proper types** - PendingTransfer.Amount is `math.Int`, not `string`
4. **Include all security params** - Tests must set all required security parameters with valid values
5. **Validate timing** - Tests that advance time must account for fraud proof window

## Expected Test Results

After these fixes:
- ✅ All fraud proof submission tests should pass
- ✅ All fraud proof resolution tests should pass
- ✅ All fraud proof window expiration tests should pass
- ✅ All supply caps validation tests should pass
- ✅ Parameter validation enforces security constraints

## Related Files

- `/home/decri/blockchain-projects/aura/chain/x/bridge/keeper/keeper.go` - Fraud proof logic
- `/home/decri/blockchain-projects/aura/chain/x/bridge/types/params.go` - Parameter validation
- `/home/decri/blockchain-projects/aura/proto/aura/bridge/v1beta1/security.proto` - PendingTransfer definition

## Verification

To verify the fixes work correctly:

```bash
cd /home/decri/blockchain-projects/aura/chain

# Run all bridge keeper tests
go test ./x/bridge/keeper/... -v

# Run specific fraud proof tests
go test ./x/bridge/keeper/... -run TestSubmitFraudProof -v
go test ./x/bridge/keeper/... -run TestResolveFraudProof -v

# Run supply caps tests
go test ./x/bridge/keeper/... -run TestSupplyCaps -v
```

All tests should now pass with proper fraud proof window enforcement and pending transfer creation.
