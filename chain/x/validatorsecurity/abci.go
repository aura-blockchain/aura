// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package validatorsecurity

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx context.Context, k keeper.Keeper) (err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't halt the chain
			k.Logger(sdkCtx).Error("panic in BeginBlocker", "module", "validatorsecurity", "panic", r)
			err = nil // Don't propagate the error to prevent chain halt
		}
	}()

	// Track signing for all validators
	// This would typically be called for each validator based on their signing status
	// In production, this should integrate with Tendermint's vote information

	return nil
}

// EndBlocker is called at the end of every block
//
// CRITICAL CONSENSUS SAFETY: This function uses batched operations to prevent
// consensus failure. With 1000+ validators, unbounded monitoring causes block
// production to exceed timeout, halting the chain.
//
// Solution: Cap validators monitored per block using cursor-based pagination.
// All validators are eventually monitored across multiple blocks without risking
// consensus failure.
func EndBlocker(ctx context.Context, k keeper.Keeper) (err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't halt the chain
			k.Logger(sdkCtx).Error("panic in EndBlocker", "module", "validatorsecurity", "panic", r)
			err = nil // Don't propagate the error to prevent chain halt
		}
	}()

	params := k.GetParams(ctx)

	// Run monitoring checks at intervals using batched operations
	// Monitors up to MaxValidatorsPerMonitoringBatch validators per block
	monitoringInterval := params.MonitoringInterval
	if shouldRunMonitoring(sdkCtx, monitoringInterval) {
		processed := k.MonitorValidatorsBatched(ctx, 0) // 0 uses default limit
		if processed > 0 {
			k.Logger(sdkCtx).Debug("monitored validators batch",
				"count", processed,
				"block_height", sdkCtx.BlockHeight())
		}
	}

	// Auto-unjail validators whose jail period has expired (batched to prevent consensus timeout)
	// Process up to MaxValidatorsPerMonitoringBatch validators per block
	const jailedBatchLimit = 50 // Use same batch limit as MonitorValidatorsBatched
	jailedValidators := k.GetJailedValidators(ctx, jailedBatchLimit)
	for _, val := range jailedValidators {
		if val.JailedUntil != nil && sdkCtx.BlockTime().After(*val.JailedUntil) {
			// Don't auto-unjail, just log that they can unjail themselves
			k.Logger(sdkCtx).Info("validator can now unjail",
				"validator", val.ValidatorAddress,
				"jailed_until", *val.JailedUntil,
			)
		}
	}

	return nil
}

func shouldRunMonitoring(ctx sdk.Context, interval time.Duration) bool {
	// Run monitoring every N blocks based on interval
	// Convert interval to approximate block count (assuming ~6 second blocks)
	//
	// DETERMINISM: Use integer arithmetic (nanoseconds) instead of floating-point
	// to ensure identical results across all validator machines. Floating-point
	// division (interval.Seconds() / 6) can produce different results on different
	// CPU architectures, causing consensus divergence and chain halts.
	blocksPerInterval := interval.Nanoseconds() / 6_000_000_000
	if blocksPerInterval == 0 {
		blocksPerInterval = 1
	}

	return ctx.BlockHeight()%blocksPerInterval == 0
}
