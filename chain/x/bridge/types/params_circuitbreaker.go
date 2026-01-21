// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
//   - k: Keeper for accessing params
//   - chain: Chain identifier to check (e.g., "paw", "xai")
//
// Returns:
//   - error if paused (global or chain-specific), nil if operations allowed
func RequireNotPaused(ctx sdk.Context, params Params, chain string) error {
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
//   - params: Current bridge parameters
//   - denom: Token denomination being minted
//   - amount: Amount about to be minted
//   - hourlyMinted: Total amount minted in the last hour for this denom
//
// Returns:
//   - bool: true if auto-pause was triggered, false otherwise
//   - Params: Updated params (with Paused=true if triggered)
func CheckAndTriggerAutoPause(ctx sdk.Context, params Params, denom string, amount sdkmath.Int, hourlyMinted sdkmath.Int) (bool, Params) {
	// Auto-pause disabled, allow operation
	if !params.AutoPauseEnabled {
		return false, params
	}

	// Parse threshold
	threshold, ok := sdkmath.NewIntFromString(params.AutoPauseThreshold)
	if !ok || !threshold.IsPositive() {
		// Invalid threshold, don't trigger auto-pause
		return false, params
	}

	// Calculate total after this mint
	totalAfterMint := hourlyMinted.Add(amount)

	// Check if threshold exceeded
	if totalAfterMint.GT(threshold) {
		// TRIGGER AUTO-PAUSE
		params.Paused = true

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

		return true, params
	}

	return false, params
}

// IsEmergencyPauseAuthorized checks if an address is authorized to trigger emergency pause.
//
// Security considerations:
//   - Only addresses in EmergencyPauseAddresses list are authorized
//   - Case-sensitive address comparison (addresses are case-sensitive)
//   - Empty address list means no one can emergency pause (requires governance)
//
// Parameters:
//   - params: Current bridge parameters
//   - address: Address to check authorization for
//
// Returns:
//   - bool: true if address is authorized, false otherwise
func IsEmergencyPauseAuthorized(params Params, address string) bool {
	if address == "" {
		return false
	}

	// Check if address is in authorized list
	for _, authorizedAddr := range params.EmergencyPauseAddresses {
		if authorizedAddr == address {
			return true
		}
	}

	return false
}
