# Consensus Bug Fix Examples

## Quick Reference: Before & After

### Example 1: CheckDustTransaction (walletsecurity/keeper/keeper.go)

**BEFORE (CONSENSUS BREAKING):**
```go
func (k Keeper) CheckDustTransaction(ctx context.Context, walletID, txHash, fromAddress, toAddress, amount, denom string) (bool, error) {
    // ... validation code ...

    if isDust {
        if txHash == "" {
            txHash = fmt.Sprintf("dust_%s_%d", walletID, time.Now().UnixNano())  // ❌ NON-DETERMINISTIC
        }
        record := &wsproto.DustTransaction{
            TxHash:       txHash,
            FromAddress:  fromAddress,
            ToAddress:    toAddress,
            Amount:       amount,
            Denom:        denom,
            DetectedAt:   timestamppb.Now(),  // ❌ NON-DETERMINISTIC
            Blocked:      true,
            Reason:       reason,
        }
    }
}
```

**AFTER (CONSENSUS SAFE):**
```go
func (k Keeper) CheckDustTransaction(ctx context.Context, walletID, txHash, fromAddress, toAddress, amount, denom string) (bool, error) {
    // ... validation code ...

    if isDust {
        if txHash == "" {
            txHash = fmt.Sprintf("dust_%s_%d", walletID, sdk.UnwrapSDKContext(ctx).BlockTime().UnixNano())  // ✅ DETERMINISTIC
        }
        record := &wsproto.DustTransaction{
            TxHash:       txHash,
            FromAddress:  fromAddress,
            ToAddress:    toAddress,
            Amount:       amount,
            Denom:        denom,
            DetectedAt:   timestamppb.New(sdk.UnwrapSDKContext(ctx).BlockTime()),  // ✅ DETERMINISTIC
            Blocked:      true,
            Reason:       reason,
        }
    }
}
```

---

### Example 2: SetSpendingLimit (walletsecurity/keeper/keeper.go)

**BEFORE (CONSENSUS BREAKING):**
```go
func (k Keeper) SetSpendingLimit(ctx context.Context, walletID string, denom string, dailyLimit string, weeklyLimit string, monthlyLimit string) (*wsproto.SpendingLimit, error) {
    limit := &wsproto.SpendingLimit{
        WalletId:            walletID,
        Denom:               denom,
        DailyLimit:          dailyLimit,
        WeeklyLimit:         weeklyLimit,
        MonthlyLimit:        monthlyLimit,
        CurrentDailySpent:   "0",
        CurrentWeeklySpent:  "0",
        CurrentMonthlySpent: "0",
        Enabled:             true,
        DailyResetAt:        timestamppb.New(time.Now().Add(24 * time.Hour)),      // ❌ NON-DETERMINISTIC
        WeeklyResetAt:       timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),  // ❌ NON-DETERMINISTIC
        MonthlyResetAt:      timestamppb.New(time.Now().Add(30 * 24 * time.Hour)), // ❌ NON-DETERMINISTIC
    }
}
```

**AFTER (CONSENSUS SAFE):**
```go
func (k Keeper) SetSpendingLimit(ctx context.Context, walletID string, denom string, dailyLimit string, weeklyLimit string, monthlyLimit string) (*wsproto.SpendingLimit, error) {
    limit := &wsproto.SpendingLimit{
        WalletId:            walletID,
        Denom:               denom,
        DailyLimit:          dailyLimit,
        WeeklyLimit:         weeklyLimit,
        MonthlyLimit:        monthlyLimit,
        CurrentDailySpent:   "0",
        CurrentWeeklySpent:  "0",
        CurrentMonthlySpent: "0",
        Enabled:             true,
        DailyResetAt:        timestamppb.New(sdk.UnwrapSDKContext(ctx).BlockTime().Add(24 * time.Hour)),      // ✅ DETERMINISTIC
        WeeklyResetAt:       timestamppb.New(sdk.UnwrapSDKContext(ctx).BlockTime().Add(7 * 24 * time.Hour)),  // ✅ DETERMINISTIC
        MonthlyResetAt:      timestamppb.New(sdk.UnwrapSDKContext(ctx).BlockTime().Add(30 * 24 * time.Hour)), // ✅ DETERMINISTIC
    }
}
```

---

### Example 3: CheckSpendingLimit (walletsecurity/keeper/keeper.go)

**BEFORE (CONSENSUS BREAKING):**
```go
func (k Keeper) CheckSpendingLimit(ctx context.Context, walletID, denom, amount string) error {
    // ... parsing code ...

    now := time.Now()  // ❌ NON-DETERMINISTIC
    if limit.DailyResetAt == nil || now.After(limit.DailyResetAt.AsTime()) {
        limit.CurrentDailySpent = "0"
        limit.DailyResetAt = timestamppb.New(now.Add(24 * time.Hour))
    }
    if limit.WeeklyResetAt == nil || now.After(limit.WeeklyResetAt.AsTime()) {
        limit.CurrentWeeklySpent = "0"
        limit.WeeklyResetAt = timestamppb.New(now.Add(7 * 24 * time.Hour))
    }
}
```

**AFTER (CONSENSUS SAFE):**
```go
func (k Keeper) CheckSpendingLimit(ctx context.Context, walletID, denom, amount string) error {
    // ... parsing code ...

    now := sdk.UnwrapSDKContext(ctx).BlockTime()  // ✅ DETERMINISTIC
    if limit.DailyResetAt == nil || now.After(limit.DailyResetAt.AsTime()) {
        limit.CurrentDailySpent = "0"
        limit.DailyResetAt = timestamppb.New(now.Add(24 * time.Hour))
    }
    if limit.WeeklyResetAt == nil || now.After(limit.WeeklyResetAt.AsTime()) {
        limit.CurrentWeeklySpent = "0"
        limit.WeeklyResetAt = timestamppb.New(now.Add(7 * 24 * time.Hour))
    }
}
```

---

### Example 4: RecordIRCompletion (confidencescore/keeper/ir_completion.go)

**BEFORE (CONSENSUS BREAKING):**
```go
func (k *Keeper) validateIRCompletionInputs(walletAddr, irID string, proofHash, verifierHash []byte, timestamp int64) error {
    // ... validation code ...

    // Check timestamp freshness (within 5 minutes)
    if time.Now().Unix()-timestamp > 300 {  // ❌ NON-DETERMINISTIC
        return types.ErrStaleAttestation
    }

    return nil
}
```

**AFTER (CONSENSUS SAFE):**
```go
func (k *Keeper) validateIRCompletionInputs(ctx sdk.Context, walletAddr, irID string, proofHash, verifierHash []byte, timestamp int64) error {
    // ... validation code ...

    // Check timestamp freshness (within 5 minutes)
    if ctx.BlockTime().Unix()-timestamp > 300 {  // ✅ DETERMINISTIC
        return types.ErrStaleAttestation
    }

    return nil
}
```

---

### Example 5: Rate Limiting (confidencescore/keeper/ir_completion.go)

**BEFORE (CONSENSUS BREAKING):**
```go
func (k *Keeper) getRateLimitKey(walletAddr, window string) string {
    now := time.Now()  // ❌ NON-DETERMINISTIC

    switch window {
    case "hour":
        return fmt.Sprintf("hour_%d_%d", now.Year(), now.YearDay()*24+now.Hour())
    case "day":
        return fmt.Sprintf("day_%d_%d", now.Year(), now.YearDay())
    default:
        return ""
    }
}
```

**AFTER (CONSENSUS SAFE):**
```go
func (k *Keeper) getRateLimitKey(ctx sdk.Context, walletAddr, window string) string {
    now := ctx.BlockTime()  // ✅ DETERMINISTIC

    switch window {
    case "hour":
        return fmt.Sprintf("hour_%d_%d", now.Year(), now.YearDay()*24+now.Hour())
    case "day":
        return fmt.Sprintf("day_%d_%d", now.Year(), now.YearDay())
    default:
        return ""
    }
}
```

---

### Example 6: Session Management (walletsecurity/keeper/session_biometric.go)

**BEFORE (CONSENSUS BREAKING):**
```go
func (k Keeper) ConfigureSession(ctx context.Context, walletID string, timeoutDuration *durationpb.Duration, autoLockEnabled bool, inactivityThresholdSeconds int32) (*wsproto.SessionConfig, error) {
    config := &wsproto.SessionConfig{
        SessionId:                  fmt.Sprintf("session_%s_%d", walletID, time.Now().Unix()),  // ❌ NON-DETERMINISTIC
        WalletId:                   walletID,
        TimeoutDuration:            timeoutDuration,
        AutoLockEnabled:            autoLockEnabled,
        InactivityThresholdSeconds: inactivityThresholdSeconds,
        StartedAt:                  timestamppb.Now(),  // ❌ NON-DETERMINISTIC
        Locked:                     false,
    }
}
```

**AFTER (CONSENSUS SAFE):**
```go
func (k Keeper) ConfigureSession(ctx context.Context, walletID string, timeoutDuration *durationpb.Duration, autoLockEnabled bool, inactivityThresholdSeconds int32) (*wsproto.SessionConfig, error) {
    config := &wsproto.SessionConfig{
        SessionId:                  fmt.Sprintf("session_%s_%d", walletID, sdk.UnwrapSDKContext(ctx).BlockTime().Unix()),  // ✅ DETERMINISTIC
        WalletId:                   walletID,
        TimeoutDuration:            timeoutDuration,
        AutoLockEnabled:            autoLockEnabled,
        InactivityThresholdSeconds: inactivityThresholdSeconds,
        StartedAt:                  timestamppb.New(sdk.UnwrapSDKContext(ctx).BlockTime()),  // ✅ DETERMINISTIC
        Locked:                     false,
    }
}
```

---

### Example 7: DEX Swap Recording (dex/keeper/liquidity_pool.go)

**BEFORE (CONSENSUS BREAKING):**
```go
func (k Keeper) SwapExactIn(...) (sdkmath.Int, sdkmath.LegacyDec, sdkmath.LegacyDec, error) {
    // ... swap logic ...

    // Record swap stats for price tracking
    k.RecordSwapStats(ctx, poolID, coinIn.Amount, amountOut, time.Now())  // ❌ NON-DETERMINISTIC
}
```

**AFTER (CONSENSUS SAFE):**
```go
func (k Keeper) SwapExactIn(...) (sdkmath.Int, sdkmath.LegacyDec, sdkmath.LegacyDec, error) {
    // ... swap logic ...

    // Record swap stats for price tracking
    k.RecordSwapStats(ctx, poolID, coinIn.Amount, amountOut, ctx.BlockTime())  // ✅ DETERMINISTIC
}
```

---

## Key Pattern Summary

| Before (Wrong) | After (Correct) |
|----------------|----------------|
| `time.Now()` | `ctx.BlockTime()` or `sdk.UnwrapSDKContext(ctx).BlockTime()` |
| `time.Now().Unix()` | `ctx.BlockTime().Unix()` |
| `time.Now().UnixNano()` | `ctx.BlockTime().UnixNano()` |
| `timestamppb.Now()` | `timestamppb.New(ctx.BlockTime())` |
| `time.Now().Add(duration)` | `ctx.BlockTime().Add(duration)` |

## Why This Matters

**Scenario:** Two validators process the same block at block height 1000:

### Before Fix (BROKEN):
```
Validator A processes at 2025-11-26 10:00:00.123
Validator B processes at 2025-11-26 10:00:00.456  (333ms difference)

Result:
  - Different state hashes
  - Consensus failure
  - Chain halt
```

### After Fix (WORKING):
```
Validator A uses block 1000 timestamp: 2025-11-26 10:00:00.000
Validator B uses block 1000 timestamp: 2025-11-26 10:00:00.000  (IDENTICAL)

Result:
  - Same state hash
  - Consensus achieved
  - Chain continues
```

---

## Testing Guidelines

1. **Multi-Validator Test:**
   - Spin up 3+ validators
   - Execute transactions that trigger time-dependent code
   - Verify all validators produce identical state hashes

2. **Time Sensitivity Test:**
   - Test operations that depend on time windows (rate limits, expiry, etc.)
   - Verify consistent behavior across all validators

3. **State Sync Test:**
   - Sync a new node from scratch
   - Verify state hash matches existing validators

---

**Remember:** In blockchain, determinism is EVERYTHING. All validators must see the same input and produce the same output. Using `time.Now()` breaks this fundamental rule.
