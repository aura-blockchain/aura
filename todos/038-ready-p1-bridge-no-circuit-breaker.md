---
id: "038"
title: "Bridge No Circuit Breaker"
status: ready
priority: p1
category: security
module: bridge
severity: HIGH
cvss: 8.5
source: bridge-security-matrix
---

# Bridge No Circuit Breaker / Emergency Pause

## Problem

No mechanism to pause bridge operations during an attack. If an exploit is discovered, no way to stop the bleeding.

## Affected Files

- `chain/x/bridge/keeper/msg_server.go`
- `chain/x/bridge/types/params.go`

## Current State

```go
// No pause check anywhere
func (ms msgServer) UnlockTokens(...) {
    // Proceeds even during active exploit
}

func (ms msgServer) LockTokens(...) {
    // Proceeds even during active exploit
}
```

## Impact

- Cannot stop ongoing attacks
- Time window to drain entire bridge
- No incident response capability

## Required Fix

```go
// Add circuit breaker to params
type Params struct {
    // Global pause
    Paused bool `json:"paused"`

    // Per-chain pause
    PausedChains []string `json:"paused_chains"`

    // Auto-pause triggers
    AutoPauseThreshold string `json:"auto_pause_threshold"` // Pause if > X minted in 1 hour
    AutoPauseEnabled   bool   `json:"auto_pause_enabled"`

    // Emergency pause addresses (multi-sig)
    EmergencyPauseAddresses []string `json:"emergency_pause_addresses"`
}

// Pause check helper
func (k Keeper) RequireNotPaused(ctx sdk.Context, chain string) error {
    params := k.GetParams(ctx)

    // Check global pause
    if params.Paused {
        return fmt.Errorf("bridge is paused")
    }

    // Check per-chain pause
    for _, pausedChain := range params.PausedChains {
        if pausedChain == chain {
            return fmt.Errorf("bridge paused for chain %s", chain)
        }
    }

    return nil
}

// Auto-pause check
func (k Keeper) CheckAutoPause(ctx sdk.Context, denom string, amount sdkmath.Int) {
    params := k.GetParams(ctx)
    if !params.AutoPauseEnabled {
        return
    }

    threshold, _ := sdkmath.NewIntFromString(params.AutoPauseThreshold)
    hourlyMinted := k.GetHourlyMintedAmount(ctx, denom)

    if hourlyMinted.Add(amount).GT(threshold) {
        // AUTO PAUSE
        params.Paused = true
        k.SetParams(ctx, params)

        ctx.EventManager().EmitEvent(
            sdk.NewEvent(
                "bridge_auto_paused",
                sdk.NewAttribute("reason", "threshold_exceeded"),
                sdk.NewAttribute("threshold", threshold.String()),
                sdk.NewAttribute("minted", hourlyMinted.Add(amount).String()),
            ),
        )
    }
}

// Emergency pause message
func (ms msgServer) EmergencyPause(goCtx context.Context, msg *bridgepb.MsgEmergencyPause) (*bridgepb.MsgEmergencyPauseResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    params := ms.Keeper.GetParams(ctx)

    // Verify sender is authorized
    signers := msg.GetSigners()
    senderStr := signers[0].String()

    authorized := false
    for _, addr := range params.EmergencyPauseAddresses {
        if addr == senderStr {
            authorized = true
            break
        }
    }

    if !authorized {
        return nil, status.Error(codes.PermissionDenied, "not authorized for emergency pause")
    }

    params.Paused = true
    ms.Keeper.SetParams(ctx, params)

    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeBridgePaused,
            sdk.NewAttribute("paused_by", senderStr),
            sdk.NewAttribute("reason", msg.Reason),
        ),
    )

    return &bridgepb.MsgEmergencyPauseResponse{}, nil
}

// Use in all operations
func (ms msgServer) UnlockTokens(...) {
    // First check pause status
    if err := ms.Keeper.RequireNotPaused(ctx, msg.SourceChain); err != nil {
        return nil, status.Error(codes.FailedPrecondition, err.Error())
    }

    // Check auto-pause
    ms.Keeper.CheckAutoPause(ctx, wrappedDenom, amount)

    // ...
}
```

## Acceptance Criteria

- [ ] Global pause mechanism implemented
- [ ] Per-chain pause capability
- [ ] Auto-pause on anomaly detection
- [ ] Emergency pause message handler
- [ ] Multi-sig authorization for pause
- [ ] Unpause with timelock/governance
- [ ] Tests for pause functionality
- [ ] Tests for auto-pause triggers
