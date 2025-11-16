package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// VALIDATOR SET ROTATION
// ============================================================================

// InitiateValidatorRotation starts a new validator set rotation
func (k Keeper) InitiateValidatorRotation(
	ctx sdk.Context,
	newValidators []*types.BridgeValidator,
) (*types.ValidatorSetRotation, error) {
	params := k.GetParams(ctx)

	// Get current validators
	currentValidators := k.GetAllBridgeValidators(ctx)

	// Generate rotation ID
	rotationId := fmt.Sprintf("rotation-%d", ctx.BlockHeight())

	// Create rotation record
	rotation := &types.ValidatorSetRotation{
		RotationId:         rotationId,
		PreviousValidators: currentValidators,
		NewValidators:      newValidators,
		RotationTime:       ctx.BlockTime(),
		EffectiveHeight:    uint64(ctx.BlockHeight()) + 100, // Effective after 100 blocks
		Status:             types.RotationStatus_ROTATION_PENDING,
		Approvals:          []*types.ValidatorSignature{},
	}

	// Store rotation
	k.SetValidatorRotation(ctx, rotation)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_rotation_initiated",
			sdk.NewAttribute("rotation_id", rotationId),
			sdk.NewAttribute("previous_count", fmt.Sprintf("%d", len(currentValidators))),
			sdk.NewAttribute("new_count", fmt.Sprintf("%d", len(newValidators))),
			sdk.NewAttribute("effective_height", fmt.Sprintf("%d", rotation.EffectiveHeight)),
		),
	)

	return rotation, nil
}

// ApproveValidatorRotation allows validators to approve a rotation
func (k Keeper) ApproveValidatorRotation(
	ctx sdk.Context,
	rotationId string,
	validatorAddr string,
	signature []byte,
) error {
	rotation := k.GetValidatorRotation(ctx, rotationId)
	if rotation == nil {
		return fmt.Errorf("rotation not found: %s", rotationId)
	}

	if rotation.Status != types.RotationStatus_ROTATION_PENDING {
		return fmt.Errorf("rotation is not pending: %s", rotation.Status.String())
	}

	// Verify validator is current validator
	validator := k.GetBridgeValidator(ctx, validatorAddr)
	if validator == nil {
		return fmt.Errorf("validator not found: %s", validatorAddr)
	}

	if !validator.Active {
		return fmt.Errorf("validator is not active: %s", validatorAddr)
	}

	// Check if already approved
	for _, approval := range rotation.Approvals {
		if approval.ValidatorAddress == validatorAddr {
			return fmt.Errorf("validator already approved: %s", validatorAddr)
		}
	}

	// Add approval
	approval := &types.ValidatorSignature{
		ValidatorAddress: validatorAddr,
		Signature:        signature,
		Timestamp:        ctx.BlockTime(),
	}
	rotation.Approvals = append(rotation.Approvals, approval)

	// Check if we have enough approvals (2/3+)
	totalPower := uint64(0)
	approvalPower := uint64(0)

	for _, v := range rotation.PreviousValidators {
		if v.Active {
			totalPower += v.Power
		}
	}

	for _, approval := range rotation.Approvals {
		for _, v := range rotation.PreviousValidators {
			if v.Address == approval.ValidatorAddress && v.Active {
				approvalPower += v.Power
				break
			}
		}
	}

	requiredPower := (totalPower * 2) / 3
	if approvalPower >= requiredPower {
		rotation.Status = types.RotationStatus_ROTATION_APPROVED
	}

	k.SetValidatorRotation(ctx, rotation)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_rotation_approved",
			sdk.NewAttribute("rotation_id", rotationId),
			sdk.NewAttribute("validator", validatorAddr),
			sdk.NewAttribute("approval_power", fmt.Sprintf("%d", approvalPower)),
			sdk.NewAttribute("required_power", fmt.Sprintf("%d", requiredPower)),
			sdk.NewAttribute("status", rotation.Status.String()),
		),
	)

	return nil
}

// ExecuteValidatorRotation executes an approved rotation at the effective height
func (k Keeper) ExecuteValidatorRotation(ctx sdk.Context, rotationId string) error {
	rotation := k.GetValidatorRotation(ctx, rotationId)
	if rotation == nil {
		return fmt.Errorf("rotation not found: %s", rotationId)
	}

	if rotation.Status != types.RotationStatus_ROTATION_APPROVED {
		return fmt.Errorf("rotation is not approved: %s", rotation.Status.String())
	}

	if uint64(ctx.BlockHeight()) < rotation.EffectiveHeight {
		return fmt.Errorf(
			"rotation not yet effective: current height %d, effective height %d",
			ctx.BlockHeight(),
			rotation.EffectiveHeight,
		)
	}

	// Deactivate old validators
	for _, v := range rotation.PreviousValidators {
		v.Active = false
		k.SetBridgeValidator(ctx, v)
	}

	// Activate new validators
	for _, v := range rotation.NewValidators {
		v.Active = true
		k.SetBridgeValidator(ctx, v)
	}

	// Update rotation status
	rotation.Status = types.RotationStatus_ROTATION_ACTIVE
	k.SetValidatorRotation(ctx, rotation)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_rotation_executed",
			sdk.NewAttribute("rotation_id", rotationId),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return nil
}

// CheckAndExecutePendingRotations checks and executes any pending rotations
func (k Keeper) CheckAndExecutePendingRotations(ctx sdk.Context) {
	rotations := k.GetAllValidatorRotations(ctx)

	for _, rotation := range rotations {
		if rotation.Status == types.RotationStatus_ROTATION_APPROVED &&
			uint64(ctx.BlockHeight()) >= rotation.EffectiveHeight {
			if err := k.ExecuteValidatorRotation(ctx, rotation.RotationId); err != nil {
				ctx.Logger().Error("failed to execute rotation", "rotation_id", rotation.RotationId, "error", err)
			}
		}
	}
}

// ============================================================================
// SLASHING MECHANISM
// ============================================================================

// SlashValidatorForInvalidProof slashes a validator for submitting an invalid Merkle proof
func (k Keeper) SlashValidatorForInvalidProof(
	ctx sdk.Context,
	validatorAddr string,
	transferId string,
) error {
	params := k.GetParams(ctx)

	validator := k.GetBridgeValidator(ctx, validatorAddr)
	if validator == nil {
		return fmt.Errorf("validator not found: %s", validatorAddr)
	}

	// Calculate slash amount
	// In production, this would slash from their staked tokens
	slashAmount := sdk.NewInt(1000000000) // Placeholder: 1000 tokens

	// Create slashing event
	event := &types.SlashingEvent{
		EventId:          fmt.Sprintf("slash-%d", ctx.BlockHeight()),
		ValidatorAddress: validatorAddr,
		Reason:           types.SlashReason_SLASH_INVALID_PROOF,
		SlashAmount:      slashAmount,
		EvidenceHash:     []byte(transferId),
		InfractionHeight: uint64(ctx.BlockHeight()),
		Timestamp:        ctx.BlockTime(),
		Jailed:           false,
	}

	// Store slashing event
	k.SetSlashingEvent(ctx, event)

	// Deactivate validator if slash is severe
	if params.SlashFractionInvalidProof.GTE(sdk.NewDecWithPrec(5, 2)) { // >= 5%
		validator.Active = false
		event.Jailed = true
		k.SetBridgeValidator(ctx, validator)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_slashed",
			sdk.NewAttribute("validator", validatorAddr),
			sdk.NewAttribute("reason", event.Reason.String()),
			sdk.NewAttribute("amount", slashAmount.String()),
			sdk.NewAttribute("jailed", fmt.Sprintf("%t", event.Jailed)),
		),
	)

	return nil
}

// SlashValidatorForDoubleSign slashes a validator for double-signing
func (k Keeper) SlashValidatorForDoubleSign(
	ctx sdk.Context,
	validatorAddr string,
	evidence []byte,
) error {
	params := k.GetParams(ctx)

	validator := k.GetBridgeValidator(ctx, validatorAddr)
	if validator == nil {
		return fmt.Errorf("validator not found: %s", validatorAddr)
	}

	slashAmount := sdk.NewInt(5000000000) // Severe: 5000 tokens

	event := &types.SlashingEvent{
		EventId:          fmt.Sprintf("slash-%d", ctx.BlockHeight()),
		ValidatorAddress: validatorAddr,
		Reason:           types.SlashReason_SLASH_DOUBLE_SIGN,
		SlashAmount:      slashAmount,
		EvidenceHash:     evidence,
		InfractionHeight: uint64(ctx.BlockHeight()),
		Timestamp:        ctx.BlockTime(),
		Jailed:           true, // Always jail for double-sign
	}

	k.SetSlashingEvent(ctx, event)

	// Permanently deactivate validator
	validator.Active = false
	k.SetBridgeValidator(ctx, validator)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_slashed",
			sdk.NewAttribute("validator", validatorAddr),
			sdk.NewAttribute("reason", event.Reason.String()),
			sdk.NewAttribute("amount", slashAmount.String()),
			sdk.NewAttribute("jailed", "true"),
		),
	)

	return nil
}

// ============================================================================
// STORAGE
// ============================================================================

// GetBridgeValidator retrieves a bridge validator
func (k Keeper) GetBridgeValidator(ctx sdk.Context, address string) *types.BridgeValidator {
	store := ctx.KVStore(k.storeKey)
	key := types.BridgeValidatorKey(address)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var validator types.BridgeValidator
	k.cdc.MustUnmarshal(bz, &validator)
	return &validator
}

// SetBridgeValidator stores a bridge validator
func (k Keeper) SetBridgeValidator(ctx sdk.Context, validator *types.BridgeValidator) {
	store := ctx.KVStore(k.storeKey)
	key := types.BridgeValidatorKey(validator.Address)

	bz := k.cdc.MustMarshal(validator)
	store.Set(key, bz)
}

// GetAllBridgeValidators retrieves all bridge validators
func (k Keeper) GetAllBridgeValidators(ctx sdk.Context) []*types.BridgeValidator {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.BridgeValidatorPrefix)
	defer iterator.Close()

	validators := []*types.BridgeValidator{}
	for ; iterator.Valid(); iterator.Next() {
		var validator types.BridgeValidator
		k.cdc.MustUnmarshal(iterator.Value(), &validator)
		validators = append(validators, &validator)
	}

	return validators
}

// GetValidatorRotation retrieves a validator rotation
func (k Keeper) GetValidatorRotation(ctx sdk.Context, rotationId string) *types.ValidatorSetRotation {
	store := ctx.KVStore(k.storeKey)
	key := types.ValidatorRotationKey(rotationId)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var rotation types.ValidatorSetRotation
	k.cdc.MustUnmarshal(bz, &rotation)
	return &rotation
}

// SetValidatorRotation stores a validator rotation
func (k Keeper) SetValidatorRotation(ctx sdk.Context, rotation *types.ValidatorSetRotation) {
	store := ctx.KVStore(k.storeKey)
	key := types.ValidatorRotationKey(rotation.RotationId)

	bz := k.cdc.MustMarshal(rotation)
	store.Set(key, bz)
}

// GetAllValidatorRotations retrieves all validator rotations
func (k Keeper) GetAllValidatorRotations(ctx sdk.Context) []*types.ValidatorSetRotation {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorRotationPrefix)
	defer iterator.Close()

	rotations := []*types.ValidatorSetRotation{}
	for ; iterator.Valid(); iterator.Next() {
		var rotation types.ValidatorSetRotation
		k.cdc.MustUnmarshal(iterator.Value(), &rotation)
		rotations = append(rotations, &rotation)
	}

	return rotations
}

// SetSlashingEvent stores a slashing event
func (k Keeper) SetSlashingEvent(ctx sdk.Context, event *types.SlashingEvent) {
	store := ctx.KVStore(k.storeKey)
	key := types.SlashingEventKey(event.EventId)

	bz := k.cdc.MustMarshal(event)
	store.Set(key, bz)
}

// GetSlashingEvent retrieves a slashing event
func (k Keeper) GetSlashingEvent(ctx sdk.Context, eventId string) *types.SlashingEvent {
	store := ctx.KVStore(k.storeKey)
	key := types.SlashingEventKey(eventId)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var event types.SlashingEvent
	k.cdc.MustUnmarshal(bz, &event)
	return &event
}
