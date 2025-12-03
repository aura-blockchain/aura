---
id: "035"
title: "Bridge Single Validator Can Drain Funds"
status: ready
priority: p1
category: security
module: bridge
severity: CRITICAL
cvss: 9.8
source: bridge-security-matrix
---

# Bridge Single Validator Can Drain Funds

## Problem

`MinConfirmations` defaults to 1, meaning a single compromised validator can approve any bridge transfer.

## Affected Files

- `chain/x/bridge/types/params.go`
- `chain/x/bridge/keeper/msg_server.go`

## Vulnerability

```go
// Default params
func DefaultParams() Params {
    return Params{
        MinConfirmations: 1,  // CRITICAL: Only 1 validator needed
        // ...
    }
}
```

## Impact

- **Max Loss: Entire bridge balance**
- Single point of failure
- Compromised validator = complete bridge drain

## Required Fix

```go
const (
    DefaultMinConfirmations = 3  // Minimum 3 validators required
    MinAllowedConfirmations = 2  // Never allow less than 2
)

func DefaultParams() Params {
    return Params{
        MinConfirmations:    DefaultMinConfirmations,
        MinValidatorPower:   1000,
        MaxTransferAmount:   "1000000000000", // 1M tokens per transfer
        // ...
    }
}

// Validate params
func (p Params) Validate() error {
    if p.MinConfirmations < MinAllowedConfirmations {
        return fmt.Errorf("MinConfirmations must be >= %d, got %d",
            MinAllowedConfirmations, p.MinConfirmations)
    }

    // Ensure MinConfirmations is less than total validators
    // to prevent deadlock
    // ...

    return nil
}

// In UnlockTokens
func (ms msgServer) UnlockTokens(...) {
    params := ms.Keeper.GetParams(ctx)

    // Enforce minimum even if params were misconfigured
    required := params.MinConfirmations
    if required < MinAllowedConfirmations {
        required = MinAllowedConfirmations
    }

    // Count unique validator signatures
    uniqueValidators := make(map[string]bool)
    for _, sig := range msg.ValidatorSignatures {
        validator := extractValidatorAddress(sig)
        if k.IsActiveValidator(ctx, validator) {
            uniqueValidators[validator] = true
        }
    }

    if len(uniqueValidators) < int(required) {
        return nil, status.Errorf(codes.PermissionDenied,
            "insufficient validators: %d < %d required",
            len(uniqueValidators), required)
    }

    // ...
}
```

## Acceptance Criteria

- [ ] MinConfirmations default increased to 3
- [ ] Minimum of 2 confirmations enforced at protocol level
- [ ] Unique validator counting (not just signature count)
- [ ] Active validator verification
- [ ] Tests for insufficient validator rejection
- [ ] Tests for duplicate validator signature rejection
