package keeper

import (
	"context"
	"fmt"
	"time"

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
	power := validator.GetConsensusPower(k.stakingKeeper.PowerReduction(ctx))

	// Slash the validator
	slashAmount, err := k.stakingKeeper.Slash(
		ctx,
		consAddr,
		height,
		power,
		params.DoubleSignSlashFraction,
	)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Tombstone the validator (permanent ban)
	if err := k.TombstoneValidator(ctx, validatorAddr); err != nil {
		return slashAmount, err
	}

	// Store evidence
	evidence := types.DoubleSignEvidence{
		ValidatorAddress: validatorAddr,
		Height:           height,
		Time:             sdkCtx.BlockTime(),
		VoteA:            voteA,
		VoteB:            voteB,
		SlashFraction:    params.DoubleSignSlashFraction,
	}

	k.SetDoubleSignEvidence(ctx, evidence)

	// Create critical alert
	k.CreateAlert(ctx, types.ValidatorAlert{
		Id:               fmt.Sprintf("double-sign-%s-%d", validatorAddr, height),
		ValidatorAddress: validatorAddr,
		AlertType:        types.ValidatorAlert_DOUBLE_SIGN,
		Severity:         types.ValidatorAlert_CRITICAL,
		Message:          fmt.Sprintf("Double sign detected at height %d, validator tombstoned", height),
		Timestamp:        sdkCtx.BlockTime(),
		Acknowledged:     false,
	})

	k.Logger(ctx).Error("validator double signed and tombstoned",
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
		return err
	}

	// Don't slash if already jailed or tombstoned
	if info.IsJailed || info.IsTombstoned {
		return nil
	}

	params := k.GetParams(ctx)

	// Check if downtime threshold exceeded
	missedBlocks := info.MissedBlocksCounter
	minSigned := params.MinSignedPerWindow.MulInt64(params.SignedBlocksWindow).TruncateInt64()
	maxMissed := params.SignedBlocksWindow - minSigned

	if missedBlocks < maxMissed {
		return nil
	}

	// Convert to validator address for slashing
	valAddr, err := sdk.ValAddressFromBech32(validatorAddr)
	if err != nil {
		return err
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}

	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return err
	}

	power := validator.GetConsensusPower(k.stakingKeeper.PowerReduction(ctx))

	// Slash for downtime
	slashAmount, err := k.stakingKeeper.Slash(
		ctx,
		consAddr,
		sdkCtx.BlockHeight(),
		power,
		params.DowntimeSlashFraction,
	)
	if err != nil {
		return err
	}

	// Jail the validator
	if err := k.JailValidator(ctx, validatorAddr, params.DowntimeJailDuration); err != nil {
		return err
	}

	// Store infraction
	infraction := types.DowntimeInfraction{
		ValidatorAddress: validatorAddr,
		MissedBlocks:     missedBlocks,
		WindowSize:       params.SignedBlocksWindow,
		DetectedAt:       sdkCtx.BlockTime(),
		SlashFraction:    params.DowntimeSlashFraction,
	}

	k.SetDowntimeInfraction(ctx, infraction)

	// Create alert
	k.CreateAlert(ctx, types.ValidatorAlert{
		Id:               fmt.Sprintf("downtime-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
		ValidatorAddress: validatorAddr,
		AlertType:        types.ValidatorAlert_DOWNTIME,
		Severity:         types.ValidatorAlert_WARNING,
		Message:          fmt.Sprintf("Downtime violation: missed %d blocks, slashed %s", missedBlocks, slashAmount),
		Timestamp:        sdkCtx.BlockTime(),
		Acknowledged:     false,
	})

	k.Logger(ctx).Warn("validator jailed for downtime",
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
		return err
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return err
	}

	tokens := validator.GetTokens()
	if tokens.LT(params.MinimumStakeAmount) {
		// Create alert
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		k.CreateAlert(ctx, types.ValidatorAlert{
			Id:               fmt.Sprintf("low-stake-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
			ValidatorAddress: validatorAddr,
			AlertType:        types.ValidatorAlert_LOW_STAKE,
			Severity:         types.ValidatorAlert_WARNING,
			Message:          fmt.Sprintf("Stake %s below minimum requirement %s", tokens, params.MinimumStakeAmount),
			Timestamp:        sdkCtx.BlockTime(),
			Acknowledged:     false,
		})

		return types.ErrInsufficientStake
	}

	return nil
}

// SetDoubleSignEvidence stores double sign evidence
func (k Keeper) SetDoubleSignEvidence(ctx context.Context, evidence types.DoubleSignEvidence) {
	store := k.getStore(ctx)
	key := types.GetDoubleSignEvidenceKey(evidence.ValidatorAddress, evidence.Height)
	bz := k.cdc.MustMarshal(&evidence)
	store.Set(key, bz)
}

// GetDoubleSignEvidence retrieves double sign evidence
func (k Keeper) GetDoubleSignEvidence(ctx context.Context, validatorAddr string, height int64) (types.DoubleSignEvidence, bool) {
	store := k.getStore(ctx)
	key := types.GetDoubleSignEvidenceKey(validatorAddr, height)
	bz := store.Get(key)
	if bz == nil {
		return types.DoubleSignEvidence{}, false
	}

	var evidence types.DoubleSignEvidence
	k.cdc.MustUnmarshal(bz, &evidence)
	return evidence, true
}

// GetAllDoubleSignEvidences returns all double sign evidences
func (k Keeper) GetAllDoubleSignEvidences(ctx context.Context) []types.DoubleSignEvidence {
	store := k.getStore(ctx)
	iterator := sdk.KVStorePrefixIterator(store, types.DoubleSignEvidenceKey)
	defer iterator.Close()

	var evidences []types.DoubleSignEvidence
	for ; iterator.Valid(); iterator.Next() {
		var evidence types.DoubleSignEvidence
		k.cdc.MustUnmarshal(iterator.Value(), &evidence)
		evidences = append(evidences, evidence)
	}

	return evidences
}

// SetDowntimeInfraction stores downtime infraction
func (k Keeper) SetDowntimeInfraction(ctx context.Context, infraction types.DowntimeInfraction) {
	store := k.getStore(ctx)
	key := types.GetDowntimeInfractionKey(infraction.ValidatorAddress)
	bz := k.cdc.MustMarshal(&infraction)
	store.Set(key, bz)
}

// GetDowntimeInfraction retrieves downtime infraction
func (k Keeper) GetDowntimeInfraction(ctx context.Context, validatorAddr string) (types.DowntimeInfraction, bool) {
	store := k.getStore(ctx)
	key := types.GetDowntimeInfractionKey(validatorAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.DowntimeInfraction{}, false
	}

	var infraction types.DowntimeInfraction
	k.cdc.MustUnmarshal(bz, &infraction)
	return infraction, true
}
