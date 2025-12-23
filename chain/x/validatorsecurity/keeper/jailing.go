package keeper

import (
	"context"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// JailValidator jails a validator for a specified duration
func (k Keeper) JailValidator(ctx context.Context, validatorAddr string, duration time.Duration) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator info
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for validatorsecurity: %w", err)
	}

	// Cannot jail tombstoned validators
	if info.IsTombstoned {
		return types.ErrValidatorTombstoned
	}

	// Already jailed
	if info.IsJailed {
		return nil
	}

	// Convert to consensus address for jailing
	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to ValAddressFromBech32 for validatorAddr: %w", err)
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("failed to Validator for validatorAddr: %w", err)
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Jail the validator
	if err := k.stakingKeeper.Jail(ctx, consAddr); err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Set jail time
	jailUntil := sdkCtx.BlockTime().Add(duration)

	// Update validator info
	info.IsJailed = true
	info.JailedUntil = &jailUntil // JailedUntil is *time.Time with stdtime=true
	k.SetValidatorSecurityInfo(ctx, info)

	k.Logger(sdkCtx).Info("validator jailed",
		"validator", validatorAddr,
		"duration", duration,
		"jailed_until", jailUntil,
	)

	return nil
}

// UnjailValidator unjails a validator
func (k Keeper) UnjailValidator(ctx context.Context, validatorAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator info
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for UnjailValidator: %w", err)
	}

	// Cannot unjail tombstoned validators
	if info.IsTombstoned {
		return types.ErrValidatorTombstoned
	}

	// Not jailed
	if !info.IsJailed {
		return types.ErrCannotUnjail
	}

	// Check if jail period has passed
	// JailedUntil is *time.Time with stdtime=true
	if info.JailedUntil != nil && sdkCtx.BlockTime().Before(*info.JailedUntil) {
		return fmt.Errorf("cannot unjail before %s", *info.JailedUntil)
	}

	// Verify minimum stake requirement
	if err := k.ValidateMinimumStake(ctx, validatorAddr); err != nil {
		return fmt.Errorf("error in UnjailValidator for ValidateMinimumStake: %w", err)
	}

	// Verify sentry node requirements
	params := k.GetParams(ctx)
	if params.RequireSentryNodes {
		sentryNodes := k.GetValidatorSentryNodes(ctx, validatorAddr)
		if int32(len(sentryNodes)) < params.MinSentryNodes {
			return types.ErrInsufficientSentryNodes
		}
	}

	// Convert to consensus address for unjailing
	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to ValAddressFromBech32 for GetValidatorSentryNodes: %w", err)
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("failed to Validator for validatorAddr: %w", err)
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Unjail the validator
	if err := k.stakingKeeper.Unjail(ctx, consAddr); err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Reset missed blocks counter
	info.IsJailed = false
	info.JailedUntil = nil
	info.MissedBlocksCounter = 0
	k.SetValidatorSecurityInfo(ctx, info)

	k.Logger(sdkCtx).Info("validator unjailed", "validator", validatorAddr)

	return nil
}

// TombstoneValidator permanently bans a validator
func (k Keeper) TombstoneValidator(ctx context.Context, validatorAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator info
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for TombstoneValidator: %w", err)
	}

	// Already tombstoned
	if info.IsTombstoned {
		return nil
	}

	// Convert to consensus address
	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to ValAddressFromBech32 for validatorAddr: %w", err)
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("failed to Validator for validatorAddr: %w", err)
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Jail the validator first
	if !info.IsJailed {
		if err := k.stakingKeeper.Jail(ctx, consAddr); err != nil {
			return fmt.Errorf("failed to get for validator: %w", err)
		}
	}

	// Tombstone via slashing keeper
	if err := k.slashingKeeper.Tombstone(ctx, consAddr); err != nil {
		return fmt.Errorf("error in TombstoneValidator for validator: %w", err)
	}

	// Update validator info
	tombstonedAt := sdkCtx.BlockTime()
	info.IsTombstoned = true
	info.TombstonedAt = &tombstonedAt // TombstonedAt is *time.Time with stdtime=true
	info.IsJailed = true
	k.SetValidatorSecurityInfo(ctx, info)

	// Decrement region count if geo distribution is enabled
	params := k.GetParams(ctx)
	if params.EnableGeoDistribution && info.Region != "" {
		k.decrementRegionCount(ctx, info.Region)
	}

	k.Logger(sdkCtx).Error("validator tombstoned (permanent ban)",
		"validator", validatorAddr,
		"tombstoned_at", tombstonedAt,
	)

	return nil
}

// GetJailedValidators returns all jailed validators in deterministic order.
// Results are ordered lexicographically by validator address to ensure
// consensus determinism across all nodes.
func (k Keeper) GetJailedValidators(ctx context.Context) []types.ValidatorSecurityInfo {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.JailedValidatorsKey)
	defer iterator.Close()

	validators := make([]types.ValidatorSecurityInfo, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		// Extract validator address from key
		validatorAddr := string(iterator.Key()[len(types.JailedValidatorsKey):])
		if info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr); err == nil {
			if info.IsJailed {
				validators = append(validators, info)
			}
		}
	}

	// KVStorePrefixIterator already returns keys in lexicographic order,
	// which provides deterministic ordering by validator address.
	// No additional sorting needed as the iteration order IS the sorted order.
	return validators
}

// GetTombstonedValidators returns all tombstoned validators in deterministic order.
// Results are ordered lexicographically by validator address to ensure
// consensus determinism across all nodes.
func (k Keeper) GetTombstonedValidators(ctx context.Context) []types.ValidatorSecurityInfo {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.TombstonedValidatorsKey)
	defer iterator.Close()

	validators := make([]types.ValidatorSecurityInfo, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		// Extract validator address from key
		validatorAddr := string(iterator.Key()[len(types.TombstonedValidatorsKey):])
		if info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr); err == nil {
			if info.IsTombstoned {
				validators = append(validators, info)
			}
		}
	}

	// KVStorePrefixIterator already returns keys in lexicographic order,
	// which provides deterministic ordering by validator address.
	// No additional sorting needed as the iteration order IS the sorted order.
	return validators
}

// IsValidatorJailed checks if a validator is jailed
func (k Keeper) IsValidatorJailed(ctx context.Context, validatorAddr string) bool {
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return false
	}
	return info.IsJailed
}

// IsValidatorTombstoned checks if a validator is tombstoned
func (k Keeper) IsValidatorTombstoned(ctx context.Context, validatorAddr string) bool {
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return false
	}
	return info.IsTombstoned
}

