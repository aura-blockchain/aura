// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// HandleDoubleSign handles double signing evidence
func (k Keeper) HandleDoubleSign(
	ctx context.Context,
	validatorAddr string,
	height int64,
	voteA, voteB []byte,
) (math.Int, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator info
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Check if already tombstoned
	if info.IsTombstoned {
		return math.ZeroInt(), types.ErrValidatorTombstoned
	}

	params := k.GetParams(ctx)

	// Validate votes are different
	if string(voteA) == string(voteB) {
		return math.ZeroInt(), types.ErrInvalidDoubleSignEvidence
	}

	// Convert validator address to consensus address for slashing
	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return math.ZeroInt(), err
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return math.ZeroInt(), err
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return math.ZeroInt(), err
	}

	// Get validator power at infraction height
	tokens := validator.GetTokens()
	powerReduction := k.stakingKeeper.PowerReduction(ctx)
	power := tokens.Quo(powerReduction).Int64()

	// Get slash fraction (already a LegacyDec)
	slashFraction := params.DoubleSignSlashFraction

	// Slash the validator
	slashAmount, err := k.stakingKeeper.Slash(
		ctx,
		consAddr,
		height,
		power,
		slashFraction,
	)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Tombstone the validator (permanent ban)
	if err := k.TombstoneValidator(ctx, validatorAddr); err != nil {
		return slashAmount, err
	}

	// Store evidence
	blockTime := sdkCtx.BlockTime()
	evidence := types.DoubleSignEvidence{
		ValidatorAddress: validatorAddr,
		Height:           height,
		Time:             &blockTime,
		VoteA:            voteA,
		VoteB:            voteB,
		SlashFraction:    params.DoubleSignSlashFraction,
	}

	k.SetDoubleSignEvidence(ctx, evidence)

	// Create critical alert
	alertTime := sdkCtx.BlockTime()
	k.CreateAlert(ctx, types.ValidatorAlert{
		Id:               fmt.Sprintf("double-sign-%s-%d", validatorAddr, height),
		ValidatorAddress: validatorAddr,
		AlertType:        types.ValidatorAlert_DOUBLE_SIGN,
		Severity:         types.ValidatorAlert_CRITICAL,
		Message:          fmt.Sprintf("Double sign detected at height %d, validator tombstoned", height),
		Timestamp:        &alertTime,
		Acknowledged:     false,
	})

	k.Logger(sdkCtx).Error("validator double signed and tombstoned",
		"validator", validatorAddr,
		"height", height,
		"slash_amount", slashAmount,
	)

	return slashAmount, nil
}

// HandleDowntime handles validator downtime violations
func (k Keeper) HandleDowntime(ctx context.Context, validatorAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Don't slash if already jailed or tombstoned
	if info.IsJailed || info.IsTombstoned {
		return nil
	}

	params := k.GetParams(ctx)

	// Check if downtime threshold exceeded
	missedBlocks := info.MissedBlocksCounter
	// MinSignedPerWindow is already a LegacyDec
	minSignedDec := params.MinSignedPerWindow
	minSigned := minSignedDec.MulInt64(params.SignedBlocksWindow).TruncateInt64()
	maxMissed := params.SignedBlocksWindow - minSigned

	if missedBlocks < maxMissed {
		return nil
	}

	// Convert to validator address for slashing
	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to ValAddressFromBech32 for validator: %w", err)
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("failed to Validator for validator: %w", err)
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Get validator power
	tokens := validator.GetTokens()
	powerReduction := k.stakingKeeper.PowerReduction(ctx)
	power := tokens.Quo(powerReduction).Int64()

	// Get downtime slash fraction (already a LegacyDec)
	slashFraction := params.DowntimeSlashFraction

	// Slash for downtime
	slashAmount, err := k.stakingKeeper.Slash(
		ctx,
		consAddr,
		sdkCtx.BlockHeight(),
		power,
		slashFraction,
	)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Jail the validator (DowntimeJailDuration is already a time.Duration)
	if err := k.JailValidator(ctx, validatorAddr, params.DowntimeJailDuration); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Store infraction
	blockTime := sdkCtx.BlockTime()
	infraction := types.DowntimeInfraction{
		ValidatorAddress: validatorAddr,
		MissedBlocks:     missedBlocks,
		WindowSize:       params.SignedBlocksWindow,
		DetectedAt:       &blockTime,
		SlashFraction:    params.DowntimeSlashFraction,
	}

	k.SetDowntimeInfraction(ctx, infraction)

	// Create alert
	alertTime := sdkCtx.BlockTime()
	k.CreateAlert(ctx, types.ValidatorAlert{
		Id:               fmt.Sprintf("downtime-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
		ValidatorAddress: validatorAddr,
		AlertType:        types.ValidatorAlert_DOWNTIME,
		Severity:         types.ValidatorAlert_WARNING,
		Message:          fmt.Sprintf("Downtime violation: missed %d blocks, slashed %s", missedBlocks, slashAmount),
		Timestamp:        &alertTime,
		Acknowledged:     false,
	})

	k.Logger(sdkCtx).Warn("validator jailed for downtime",
		"validator", validatorAddr,
		"missed_blocks", missedBlocks,
		"slash_amount", slashAmount,
	)

	return nil
}

// ValidateMinimumStake checks if validator meets minimum staking requirements
func (k Keeper) ValidateMinimumStake(ctx context.Context, validatorAddr string) error {
	params := k.GetParams(ctx)

	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to ValAddressFromBech32 for ValidateMinimumStake: %w", err)
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return fmt.Errorf("failed to Validator for ValidateMinimumStake: %w", err)
	}

	tokens := validator.GetTokens()
	// MinimumStakeAmount is already a math.Int
	minStake := params.MinimumStakeAmount

	if tokens.LT(minStake) {
		// Create alert
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		blockTime := sdkCtx.BlockTime()
		k.CreateAlert(ctx, types.ValidatorAlert{
			Id:               fmt.Sprintf("low-stake-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
			ValidatorAddress: validatorAddr,
			AlertType:        types.ValidatorAlert_LOW_STAKE,
			Severity:         types.ValidatorAlert_WARNING,
			Message:          fmt.Sprintf("Stake %s below minimum requirement %s", tokens, params.MinimumStakeAmount),
			Timestamp:        &blockTime,
			Acknowledged:     false,
		})

		return types.ErrInsufficientStake
	}

	return nil
}

// Note: SetDoubleSignEvidence, GetDoubleSignEvidence, GetAllDoubleSignEvidences,
// SetDowntimeInfraction, and GetDowntimeInfraction are defined in genesis.go
// to avoid duplication.
