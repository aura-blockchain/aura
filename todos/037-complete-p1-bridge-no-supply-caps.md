---
id: "037"
title: "Bridge No Supply Caps"
status: ready
priority: p1
category: security
module: bridge
severity: CRITICAL
cvss: 9.0
source: bridge-security-matrix
---

# Bridge No Supply Caps

## Problem

No limit on minted wrapped tokens. Even if source chain has limited supply, bridge can mint unlimited tokens.

## Affected Files

- `chain/x/bridge/keeper/msg_server.go`
- `chain/x/bridge/types/params.go`

## Vulnerability

```go
// Current: Mints any amount without checking caps
func (ms msgServer) UnlockTokens(...) {
    // Mint tokens to recipient
    coins := sdk.NewCoins(sdk.NewCoin(wrappedDenom, amount))
    if err := ms.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
        return nil, err
    }
    // NO CHECK: Does this exceed source chain total supply?
    // NO CHECK: Does this exceed per-token cap?
    // NO CHECK: Does this exceed daily mint limit?
}
```

## Impact

- Token inflation beyond source supply
- Economic manipulation
- Loss of peg to source asset

## Required Fix

```go
// Add to params
type Params struct {
    // Per-token supply caps (max wrapped tokens per denom)
    SupplyCaps map[string]string `json:"supply_caps"`

    // Per-transfer limits
    MaxTransferAmount string `json:"max_transfer_amount"`

    // Daily mint limits
    DailyMintLimit string `json:"daily_mint_limit"`

    // Hourly mint limits (rate limiting)
    HourlyMintLimit string `json:"hourly_mint_limit"`
}

func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    params := ms.Keeper.GetParams(ctx)

    amount, ok := sdkmath.NewIntFromString(msg.Amount)
    if !ok {
        return nil, status.Error(codes.InvalidArgument, "invalid amount")
    }

    wrappedDenom := "wrapped." + msg.SourceChain + "." + msg.Denom

    // 1. Check per-transfer limit
    maxTransfer, _ := sdkmath.NewIntFromString(params.MaxTransferAmount)
    if amount.GT(maxTransfer) {
        return nil, status.Errorf(codes.InvalidArgument,
            "amount %s exceeds max transfer %s", amount, maxTransfer)
    }

    // 2. Check supply cap
    if cap, exists := params.SupplyCaps[wrappedDenom]; exists {
        supplyCap, _ := sdkmath.NewIntFromString(cap)
        currentSupply := ms.bankKeeper.GetSupply(ctx, wrappedDenom).Amount
        newSupply := currentSupply.Add(amount)

        if newSupply.GT(supplyCap) {
            return nil, status.Errorf(codes.ResourceExhausted,
                "would exceed supply cap: %s + %s > %s",
                currentSupply, amount, supplyCap)
        }
    }

    // 3. Check daily mint limit
    dailyMinted := ms.Keeper.GetDailyMintedAmount(ctx, wrappedDenom)
    dailyLimit, _ := sdkmath.NewIntFromString(params.DailyMintLimit)
    if dailyMinted.Add(amount).GT(dailyLimit) {
        return nil, status.Errorf(codes.ResourceExhausted,
            "would exceed daily limit: %s + %s > %s",
            dailyMinted, amount, dailyLimit)
    }

    // 4. Check hourly rate limit
    hourlyMinted := ms.Keeper.GetHourlyMintedAmount(ctx, wrappedDenom)
    hourlyLimit, _ := sdkmath.NewIntFromString(params.HourlyMintLimit)
    if hourlyMinted.Add(amount).GT(hourlyLimit) {
        return nil, status.Errorf(codes.ResourceExhausted,
            "would exceed hourly limit: %s + %s > %s",
            hourlyMinted, amount, hourlyLimit)
    }

    // Track minted amounts
    ms.Keeper.AddDailyMintedAmount(ctx, wrappedDenom, amount)
    ms.Keeper.AddHourlyMintedAmount(ctx, wrappedDenom, amount)

    // Proceed with minting...
}
```

## Acceptance Criteria

- [ ] Per-token supply caps implemented
- [ ] Per-transfer max amount enforced
- [ ] Daily mint limits implemented
- [ ] Hourly rate limiting implemented
- [ ] Governance can update caps
- [ ] Tests for supply cap enforcement
- [ ] Tests for rate limiting
