// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// RegisterInvariants registers all networksecurity module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "peer-reputation-consistency", PeerReputationConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rate-limit-validity", RateLimitValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "mempool-security-state", MempoolSecurityStateInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sybil-detection-integrity", SybilDetectionIntegrityInvariant(k))
}

// AllInvariants runs all invariants of the networksecurity module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			PeerReputationConsistencyInvariant(k),
			RateLimitValidityInvariant(k),
			MempoolSecurityStateInvariant(k),
			SybilDetectionIntegrityInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-get",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}
		// Sub-configs are value types (gogoproto.nullable = false), so they can't be nil
		// Validate params using the validation function instead
		if err := types.ValidateParams(&params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid params: %s", err.Error()),
			), true
		}
		return "", false
	}
}

// PeerReputationConsistencyInvariant checks peer reputation scores
func PeerReputationConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		reputations := k.GetAllReputations(ctx)

		for _, reputation := range reputations {
			// Check peer ID is not empty
			if reputation.PeerId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"peer-reputation-consistency",
					"peer reputation has empty peer ID",
				), true
			}

			// Check reputation score is in valid range (0-100)
			if reputation.Score < 0 || reputation.Score > 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"peer-reputation-consistency",
					fmt.Sprintf("peer %s has invalid reputation score: %d",
						reputation.PeerId, reputation.Score),
				), true
			}

			// Check last updated height
			if reputation.LastUpdatedHeight < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"peer-reputation-consistency",
					fmt.Sprintf("peer %s has negative last_updated_height", reputation.PeerId),
				), true
			}
		}

		return "", false
	}
}

// RateLimitValidityInvariant checks rate limit configurations
func RateLimitValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		rateLimits := k.GetAllRateLimits(ctx)

		for _, rateLimit := range rateLimits {
			// Check peer ID is not empty
			if rateLimit.PeerId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					"rate limit has empty peer ID",
				), true
			}

			// WindowStart is a value type (time.Time), check if it's zero instead
			if rateLimit.WindowStart.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					fmt.Sprintf("rate limit for peer %s has zero window_start", rateLimit.PeerId),
				), true
			}
		}

		return "", false
	}
}

// MempoolSecurityStateInvariant checks mempool security state
func MempoolSecurityStateInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		stats := k.GetMempoolStats(ctx)

		// Check that stats are reasonable (all uint64, can't be negative, but check for overflow)
		if stats.TxCount > 1<<60 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"mempool-security-state",
				fmt.Sprintf("mempool has unreasonable transaction count: %d", stats.TxCount),
			), true
		}

		if stats.SizeBytes > 1<<60 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"mempool-security-state",
				fmt.Sprintf("mempool has unreasonable size: %d", stats.SizeBytes),
			), true
		}

		return "", false
	}
}

// SybilDetectionIntegrityInvariant checks Sybil attack detection state
func SybilDetectionIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Simplified invariant - just check that we can get fork and partition alerts
		forkAlerts := k.GetAllForkAlerts(ctx, false)
		partitionAlerts := k.GetAllPartitionAlerts(ctx, false)

		// Validate fork alerts
		for _, alert := range forkAlerts {
			if alert.AlertId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sybil-detection-integrity",
					"fork alert has empty ID",
				), true
			}
			if alert.BlockHeight < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sybil-detection-integrity",
					fmt.Sprintf("fork alert %s has negative height %d", alert.AlertId, alert.BlockHeight),
				), true
			}
		}

		// Validate partition alerts
		for _, alert := range partitionAlerts {
			if alert.AlertId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sybil-detection-integrity",
					"partition alert has empty ID",
				), true
			}
		}

		return "", false
	}
}
