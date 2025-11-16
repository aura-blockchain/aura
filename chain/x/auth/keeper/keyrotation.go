package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// InitiateValidatorKeyRotation initiates a validator key rotation
func (k *Keeper) InitiateValidatorKeyRotation(ctx context.Context, initiator, validatorAddress, newConsensusPubkey string) (*authproto.ValidatorKeyRotation, error) {
	// Validate initiator has permission
	if err := k.RequirePermission(ctx, initiator, types.PermissionRotateValidatorKey); err != nil {
		return nil, err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if rotation is already in progress
	if existing, ok := k.validatorRotations[validatorAddress]; ok {
		if existing.RotationStatus == authproto.RotationStatus_ROTATION_STATUS_PENDING {
			k.LogAudit(ctx, initiator, "initiate_key_rotation", validatorAddress, "failed", nil, "rotation already in progress")
			return nil, types.ErrRotationInProgress
		}
	}

	// Get current consensus pubkey (in real implementation, query from staking module)
	oldConsensusPubkey := k.getValidatorConsensusPubkey(validatorAddress)

	now := time.Now()
	rotation := &authproto.ValidatorKeyRotation{
		ValidatorAddress:   validatorAddress,
		OldConsensusPubkey: oldConsensusPubkey,
		NewConsensusPubkey: newConsensusPubkey,
		RotationTime:       &now,
		InitiatedBy:        initiator,
		RotationStatus:     authproto.RotationStatus_ROTATION_STATUS_PENDING,
	}

	k.validatorRotations[validatorAddress] = rotation
	k.LogAudit(ctx, initiator, "initiate_key_rotation", validatorAddress, "success", map[string]string{
		"old_pubkey": oldConsensusPubkey,
		"new_pubkey": newConsensusPubkey,
	}, "")

	return rotation, nil
}

// CompleteValidatorKeyRotation completes a pending validator key rotation
func (k *Keeper) CompleteValidatorKeyRotation(ctx context.Context, completer, validatorAddress string) error {
	// Validate completer has permission
	if err := k.RequirePermission(ctx, completer, types.PermissionRotateValidatorKey); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	rotation, ok := k.validatorRotations[validatorAddress]
	if !ok {
		k.LogAudit(ctx, completer, "complete_key_rotation", validatorAddress, "failed", nil, "rotation not found")
		return types.ErrRotationNotFound
	}

	if rotation.RotationStatus != authproto.RotationStatus_ROTATION_STATUS_PENDING {
		k.LogAudit(ctx, completer, "complete_key_rotation", validatorAddress, "failed", nil, "rotation not pending")
		return fmt.Errorf("rotation not in pending state")
	}

	// In a real implementation, this would:
	// 1. Update the validator's consensus pubkey in the staking module
	// 2. Notify tendermint/cometbft of the key change
	// 3. Update validator signing info
	// 4. Handle any necessary slashing protection

	rotation.RotationStatus = authproto.RotationStatus_ROTATION_STATUS_COMPLETED

	k.LogAudit(ctx, completer, "complete_key_rotation", validatorAddress, "success", map[string]string{
		"old_pubkey": rotation.OldConsensusPubkey,
		"new_pubkey": rotation.NewConsensusPubkey,
	}, "")

	return nil
}

// FailValidatorKeyRotation marks a key rotation as failed
func (k *Keeper) FailValidatorKeyRotation(ctx context.Context, failer, validatorAddress string, reason string) error {
	// Validate failer has permission
	if err := k.RequirePermission(ctx, failer, types.PermissionRotateValidatorKey); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	rotation, ok := k.validatorRotations[validatorAddress]
	if !ok {
		k.LogAudit(ctx, failer, "fail_key_rotation", validatorAddress, "failed", nil, "rotation not found")
		return types.ErrRotationNotFound
	}

	rotation.RotationStatus = authproto.RotationStatus_ROTATION_STATUS_FAILED

	k.LogAudit(ctx, failer, "fail_key_rotation", validatorAddress, "success", map[string]string{
		"reason": reason,
	}, "")

	return nil
}

// GetValidatorKeyRotation retrieves key rotation status for a validator
func (k *Keeper) GetValidatorKeyRotation(validatorAddress string) (*authproto.ValidatorKeyRotation, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	rotation, ok := k.validatorRotations[validatorAddress]
	if !ok {
		return nil, types.ErrRotationNotFound
	}

	return rotation, nil
}

// ListValidatorKeyRotations returns all key rotations
func (k *Keeper) ListValidatorKeyRotations() []*authproto.ValidatorKeyRotation {
	k.mu.RLock()
	defer k.mu.RUnlock()

	rotations := make([]*authproto.ValidatorKeyRotation, 0, len(k.validatorRotations))
	for _, rotation := range k.validatorRotations {
		rotations = append(rotations, rotation)
	}

	return rotations
}

// getValidatorConsensusPubkey retrieves the current consensus pubkey for a validator
func (k *Keeper) getValidatorConsensusPubkey(validatorAddress string) string {
	// In a real implementation, this would query the staking module
	// For now, return a placeholder
	return "current_pubkey_" + validatorAddress
}

// RotateOperatorKey rotates an operator key (not consensus key)
func (k *Keeper) RotateOperatorKey(ctx context.Context, operator, validatorAddress, newOperatorKey string) error {
	// Validate operator has permission
	if err := k.RequirePermission(ctx, operator, types.PermissionRotateValidatorKey); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// In a real implementation, this would update the operator key
	// which is used for signing validator operations (not blocks)

	k.LogAudit(ctx, operator, "rotate_operator_key", validatorAddress, "success", map[string]string{
		"new_operator_key": newOperatorKey,
	}, "")

	return nil
}

// ScheduleKeyRotation schedules a key rotation for a future time
func (k *Keeper) ScheduleKeyRotation(ctx context.Context, scheduler, validatorAddress, newConsensusPubkey string, scheduleTime time.Time) error {
	// Validate scheduler has permission
	if err := k.RequirePermission(ctx, scheduler, types.PermissionRotateValidatorKey); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// In a real implementation, this would create a scheduled task
	// to execute the key rotation at the specified time

	k.LogAudit(ctx, scheduler, "schedule_key_rotation", validatorAddress, "success", map[string]string{
		"new_pubkey":    newConsensusPubkey,
		"schedule_time": scheduleTime.Format(time.RFC3339),
	}, "")

	return nil
}

// ValidateKeyRotation performs security checks before key rotation
func (k *Keeper) ValidateKeyRotation(validatorAddress, newConsensusPubkey string) error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Check if validator exists
	// In real implementation, query staking module

	// Check if new key is different from current key
	currentPubkey := k.getValidatorConsensusPubkey(validatorAddress)
	if currentPubkey == newConsensusPubkey {
		return fmt.Errorf("new pubkey must be different from current pubkey")
	}

	// Check if new key is not already in use by another validator
	for addr, rotation := range k.validatorRotations {
		if addr != validatorAddress {
			if rotation.NewConsensusPubkey == newConsensusPubkey || rotation.OldConsensusPubkey == newConsensusPubkey {
				return fmt.Errorf("pubkey already in use by validator %s", addr)
			}
		}
	}

	return nil
}
