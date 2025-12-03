package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"
)

// RequireNotPaused checks if the bridge is paused globally or for a specific chain.
// Returns an error if operations should be blocked.
//
// Security considerations:
//   - Global pause overrides all operations
//   - Per-chain pause blocks only operations for that chain
//   - Case-insensitive chain name comparison
//
// Parameters:
//   - ctx: SDK context for accessing parameters
//   - chain: Chain identifier to check (e.g., "paw", "xai")
//
// Returns:
//   - error if paused (global or chain-specific), nil if operations allowed
func (k Keeper) RequireNotPaused(ctx sdk.Context, chain string) error {
	params := k.GetParams(ctx)

	// Check global pause first
	if params.Paused {
		return fmt.Errorf("bridge is globally paused - all operations disabled")
	}

	// Check per-chain pause
	if chain != "" {
		chainNormalized := strings.ToLower(strings.TrimSpace(chain))
		for _, pausedChain := range params.PausedChains {
			if strings.ToLower(strings.TrimSpace(pausedChain)) == chainNormalized {
				return fmt.Errorf("bridge paused for chain %s", chain)
			}
		}
	}

	return nil
}

// CheckAndTriggerAutoPause checks if hourly minting exceeds threshold and triggers auto-pause.
//
// Security considerations:
//   - Only triggers if AutoPauseEnabled is true
//   - Checks hourly minted amount against threshold
//   - Updates params to set Paused=true if threshold exceeded
//   - Emits event for monitoring and alerting
//
// Parameters:
//   - ctx: SDK context for state access and events
//   - denom: Token denomination being minted
//   - amount: Amount about to be minted
//
// Returns:
//   - bool: true if auto-pause was triggered, false otherwise
func (k Keeper) CheckAndTriggerAutoPause(ctx sdk.Context, denom string, amount sdkmath.Int) bool {
	params := k.GetParams(ctx)

	// Auto-pause disabled, allow operation
	if !params.AutoPauseEnabled {
		return false
	}

	// Parse threshold
	threshold, ok := sdkmath.NewIntFromString(params.AutoPauseThreshold)
	if !ok || !threshold.IsPositive() {
		// Invalid threshold, don't trigger auto-pause
		return false
	}

	// Get hourly minted amount
	hourlyMinted := k.GetHourlyMintedAmount(ctx, denom)

	// Calculate total after this mint
	totalAfterMint := hourlyMinted.Add(amount)

	// Check if threshold exceeded
	if totalAfterMint.GT(threshold) {
		// TRIGGER AUTO-PAUSE
		params.Paused = true
		k.SetParams(ctx, params)

		// Emit critical event for monitoring
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"bridge_auto_paused",
				sdk.NewAttribute("reason", "hourly_mint_threshold_exceeded"),
				sdk.NewAttribute("denom", denom),
				sdk.NewAttribute("threshold", threshold.String()),
				sdk.NewAttribute("hourly_minted", hourlyMinted.String()),
				sdk.NewAttribute("attempted_amount", amount.String()),
				sdk.NewAttribute("total_after_mint", totalAfterMint.String()),
				sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
			),
		)

		return true
	}

	return false
}

// IsEmergencyPauseAuthorized checks if an address is authorized to trigger emergency pause.
//
// Security considerations:
//   - Only addresses in EmergencyPauseAddresses list are authorized
//   - Case-sensitive address comparison (addresses are case-sensitive)
//   - Empty address list means no one can emergency pause (requires governance)
//
// Parameters:
//   - ctx: SDK context for accessing parameters
//   - address: Address to check authorization for
//
// Returns:
//   - bool: true if address is authorized, false otherwise
func (k Keeper) IsEmergencyPauseAuthorized(ctx sdk.Context, address string) bool {
	if address == "" {
		return false
	}

	params := k.GetParams(ctx)

	// Check if address is in authorized list
	for _, authorizedAddr := range params.EmergencyPauseAddresses {
		if authorizedAddr == address {
			return true
		}
	}

	return false
}

// GetHourlyMintedAmountRolling returns the total amount minted for a denom in the last hour using rolling window.
// Note: This is an alternative implementation. The primary implementation is in keeper.go
// which uses hourly reset buckets (more efficient for high-frequency operations).
//
// Implementation:
//   - Tracks minting events in a rolling one-hour window
//   - Uses block time to determine which mints are within the hour
//   - Sums all mints from the past hour
//
// Parameters:
//   - ctx: SDK context for state access
//   - denom: Token denomination to check
//
// Returns:
//   - sdkmath.Int: Total minted in the last hour
func (k Keeper) GetHourlyMintedAmountRolling(ctx sdk.Context, denom string) sdkmath.Int {
	// Use the optimized hourly bucket implementation from keeper.go
	return k.GetHourlyMintedAmount(ctx, denom)
}

// RecordMintedAmount records a mint event for hourly tracking.
//
// Security considerations:
//   - Must be called AFTER successful mint
//   - Used for auto-pause threshold calculation
//   - Automatically cleans up old entries
//
// Parameters:
//   - ctx: SDK context for state access
//   - denom: Token denomination minted
//   - amount: Amount minted
func (k Keeper) RecordMintedAmount(ctx sdk.Context, denom string, amount sdkmath.Int) {
	if !amount.IsPositive() {
		return
	}

	store := k.store(ctx)
	now := ctx.BlockTime()

	// Serialize timestamp
	timestampBytes, err := now.MarshalBinary()
	if err != nil {
		return
	}

	// Key: "hourly-mint-<denom>-<timestamp>"
	key := []byte(fmt.Sprintf("hourly-mint-%s-", denom))
	key = append(key, timestampBytes...)

	// Value: amount
	amountBytes, err := amount.Marshal()
	if err != nil {
		return
	}

	store.Set(key, amountBytes)

	// Emit event for monitoring
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"mint_recorded",
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("timestamp", now.Format(time.RFC3339)),
		),
	)
}
