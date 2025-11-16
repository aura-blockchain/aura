package validatorsecurity

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Track signing for all validators
	// This would typically be called for each validator based on their signing status
	// In production, this should integrate with Tendermint's vote information

	return nil
}

// EndBlocker is called at the end of every block
func EndBlocker(ctx context.Context, k keeper.Keeper) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(ctx)

	// Run monitoring checks at intervals
	if shouldRunMonitoring(sdkCtx, params.MonitoringInterval) {
		k.MonitorAllValidators(ctx)
	}

	// Auto-unjail validators whose jail period has expired
	jailedValidators := k.GetJailedValidators(ctx)
	for _, val := range jailedValidators {
		if val.JailedUntil != nil && sdkCtx.BlockTime().After(*val.JailedUntil) {
			// Don't auto-unjail, just log that they can unjail themselves
			k.Logger(ctx).Info("validator can now unjail",
				"validator", val.ValidatorAddress,
				"jailed_until", val.JailedUntil,
			)
		}
	}

	return nil
}

func shouldRunMonitoring(ctx sdk.Context, interval time.Duration) bool {
	// Run monitoring every N blocks based on interval
	// Convert interval to approximate block count (assuming ~6 second blocks)
	blocksPerInterval := int64(interval.Seconds() / 6)
	if blocksPerInterval == 0 {
		blocksPerInterval = 1
	}

	return ctx.BlockHeight()%blocksPerInterval == 0
}
