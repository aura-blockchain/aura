package keeper

import (
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// ============================================================================
// DID Key Rotation Management
// ============================================================================

// RotateDIDKey initiates a DID key rotation with grace period
func (k *Keeper) RotateDIDKey(ctx sdk.Context, did, initiator, newVerificationMethod, reason string) (*types.DIDKeyRotation, error) {
	// Get metrics instance
	metrics := GetIdentityMetrics()

	// Validate inputs
	if did == "" {
		metrics.DIDKeyRotationFails.WithLabelValues("invalid_did").Inc()
		return nil, types.ErrInvalidDID.Wrap("DID cannot be empty")
	}
	if newVerificationMethod == "" {
		metrics.DIDKeyRotationFails.WithLabelValues("invalid_verification_method").Inc()
		return nil, types.ErrInvalidVerificationMethod.Wrap("verification method cannot be empty")
	}

	// Get identity record
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		metrics.DIDKeyRotationFails.WithLabelValues("identity_not_found").Inc()
		return nil, types.ErrIdentityNotFound.Wrapf("identity not found: %s", did)
	}

	// Check authorization: initiator must be DID owner or have permission
	if record.Address != initiator {
		if err := k.RequirePermission(ctx, initiator, types.PermissionManageIdentity); err != nil {
			metrics.DIDKeyRotationFails.WithLabelValues("unauthorized").Inc()
			return nil, types.ErrUnauthorized.Wrapf("initiator %s not authorized to rotate keys for %s", initiator, did)
		}
	}

	// Check if identity is erased
	if record.Erased {
		metrics.DIDKeyRotationFails.WithLabelValues("identity_erased").Inc()
		return nil, types.ErrIdentityErased.Wrapf("cannot rotate keys for erased identity %s", did)
	}

	// Check if rotation already in progress
	existingRotation, err := k.GetDIDKeyRotation(ctx, did)
	if err == nil && existingRotation.Status == types.DIDKeyRotationStatusPending {
		metrics.DIDKeyRotationFails.WithLabelValues("rotation_in_progress").Inc()
		return nil, types.ErrDIDKeyRotationInProgress.Wrapf("rotation already in progress for %s", did)
	}

	// Get current verification method (first one if multiple)
	var oldVerificationMethod string
	if len(record.VerificationMethods) > 0 {
		oldVerificationMethod = record.VerificationMethods[0]
	} else {
		// No existing verification method, treat as first key
		oldVerificationMethod = ""
	}

	// Get grace period from params
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	now := ctx.BlockTime()
	gracePeriodDuration := time.Duration(params.Change.KeyRotationGracePeriodSeconds) * time.Second
	gracePeriodEnd := now.Add(gracePeriodDuration)

	// Create rotation record
	rotation := &types.DIDKeyRotation{
		Did:                    did,
		OldVerificationMethod:  oldVerificationMethod,
		NewVerificationMethod:  newVerificationMethod,
		RotationTime:           timestamppb.New(now),
		InitiatedBy:            initiator,
		Reason:                 reason,
		GracePeriodEnd:         timestamppb.New(gracePeriodEnd),
		Status:                 types.DIDKeyRotationStatusPending,
	}

	// Save rotation
	if err := k.SetDIDKeyRotation(ctx, rotation); err != nil {
		return nil, fmt.Errorf("failed to save DID key rotation: %w", err)
	}

	// Add old key to history if it exists
	if oldVerificationMethod != "" {
		if err := k.AddKeyToHistory(ctx, did, oldVerificationMethod, now, &gracePeriodEnd, initiator, reason, false); err != nil {
			return nil, fmt.Errorf("failed to add key to history: %w", err)
		}
	}

	// Update identity record with new verification method
	// Keep old verification method in the list during grace period
	newMethods := []string{newVerificationMethod}
	if oldVerificationMethod != "" {
		newMethods = append(newMethods, oldVerificationMethod)
	}
	record.VerificationMethods = newMethods
	record.UpdatedAt = timestamppb.New(now)

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to update identity record: %w", err)
	}

	// Log audit trail
	k.LogAudit(ctx, initiator, "rotate_did_key", did, "success", map[string]string{
		"old_method":       oldVerificationMethod,
		"new_method":       newVerificationMethod,
		"grace_period_end": gracePeriodEnd.Format(time.RFC3339),
		"reason":           reason,
	}, "")

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "rotate_did_key"),
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("old_method", oldVerificationMethod),
			sdk.NewAttribute("new_method", newVerificationMethod),
			sdk.NewAttribute("grace_period_end", gracePeriodEnd.Format(time.RFC3339)),
		),
	)

	// Record successful key rotation metric
	// Determine key type from verification method (simplified - could be enhanced)
	keyType := "unknown"
	if len(newVerificationMethod) > 0 {
		if newVerificationMethod[0:2] == "Ed" {
			keyType = "ed25519"
		} else if newVerificationMethod[0:3] == "Sec" {
			keyType = "secp256k1"
		}
	}
	metrics.DIDKeyRotations.WithLabelValues(did, keyType).Inc()

	return rotation, nil
}

// ValidateDIDKey checks if a verification method is valid for a DID
// Considers both current key and keys in grace period
func (k *Keeper) ValidateDIDKey(ctx sdk.Context, did, verificationMethod string) error {
	// Get identity record
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return types.ErrIdentityNotFound.Wrapf("identity not found: %s", did)
	}

	// Check if it's the current (first) key
	if len(record.VerificationMethods) > 0 && record.VerificationMethods[0] == verificationMethod {
		return nil
	}

	// Check if key is in grace period
	rotation, err := k.GetDIDKeyRotation(ctx, did)
	if err == nil && rotation.Status == types.DIDKeyRotationStatusPending {
		// Check if old key matches and grace period hasn't expired
		if rotation.OldVerificationMethod == verificationMethod {
			now := ctx.BlockTime()
			if now.Before(rotation.GracePeriodEnd.AsTime()) {
				// Old key is still valid in grace period
				return nil
			}
		}
	}

	// Check key history for any keys still in their grace periods
	history, err := k.GetDIDKeyHistory(ctx, did)
	if err == nil {
		now := ctx.BlockTime()
		for _, entry := range history.Entries {
			if entry.VerificationMethod == verificationMethod {
				// Check if this historical key is still in grace period
				if entry.ActiveUntil != nil && now.Before(entry.ActiveUntil.AsTime()) {
					return nil
				}
			}
		}
	}

	return types.ErrKeyNotValid.Wrapf("verification method %s not valid for DID %s", verificationMethod, did)
}

// CompleteKeyRotation marks a key rotation as complete after grace period
// This should be called via a cron job or when processing blocks after grace period
func (k *Keeper) CompleteKeyRotation(ctx sdk.Context, did string) error {
	rotation, err := k.GetDIDKeyRotation(ctx, did)
	if err != nil {
		return types.ErrDIDKeyRotationNotFound.Wrapf("rotation not found for %s", did)
	}

	if rotation.Status != types.DIDKeyRotationStatusPending {
		return types.ErrInvalidAction.Wrapf("rotation not in pending status for %s", did)
	}

	now := ctx.BlockTime()
	if now.Before(rotation.GracePeriodEnd.AsTime()) {
		return types.ErrKeyInGracePeriod.Wrapf("grace period not yet ended for %s", did)
	}

	// Update rotation status
	rotation.Status = types.DIDKeyRotationStatusCompleted
	if err := k.SetDIDKeyRotation(ctx, rotation); err != nil {
		return fmt.Errorf("failed to update rotation: %w", err)
	}

	// Remove old verification method from identity record
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return types.ErrIdentityNotFound.Wrapf("identity not found: %s", did)
	}

	// Filter out the old verification method
	newMethods := []string{}
	for _, method := range record.VerificationMethods {
		if method != rotation.OldVerificationMethod {
			newMethods = append(newMethods, method)
		}
	}
	record.VerificationMethods = newMethods
	record.UpdatedAt = timestamppb.New(now)

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to update identity record: %w", err)
	}

	// Update key history to mark old key as no longer valid
	if err := k.UpdateKeyHistoryActiveUntil(ctx, did, rotation.OldVerificationMethod, &now); err != nil {
		return fmt.Errorf("failed to update key history: %w", err)
	}

	// Log audit trail
	k.LogAudit(ctx, "system", "complete_key_rotation", did, "success", map[string]string{
		"old_method": rotation.OldVerificationMethod,
		"new_method": rotation.NewVerificationMethod,
	}, "")

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.ModuleName,
			sdk.NewAttribute("action", "complete_key_rotation"),
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("old_method", rotation.OldVerificationMethod),
		),
	)

	return nil
}

// ============================================================================
// DID Key Rotation Storage
// ============================================================================

// SetDIDKeyRotation stores a DID key rotation
func (k *Keeper) SetDIDKeyRotation(ctx sdk.Context, rotation *types.DIDKeyRotation) error {
	if rotation.Did == "" {
		return types.ErrInvalidDID.Wrap("DID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(rotation)
	if err != nil {
		return fmt.Errorf("failed to marshal DID key rotation: %w", err)
	}

	key := types.GetDIDKeyRotationKey(rotation.Did)
	return store.Set(key, bz)
}

// GetDIDKeyRotation retrieves a DID key rotation
func (k *Keeper) GetDIDKeyRotation(ctx sdk.Context, did string) (*types.DIDKeyRotation, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetDIDKeyRotationKey(did)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrDIDKeyRotationNotFound.Wrapf("rotation not found for %s", did)
	}

	var rotation types.DIDKeyRotation
	if err := k.cdc.Unmarshal(bz, &rotation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DID key rotation: %w", err)
	}
	return &rotation, nil
}

// GetAllDIDKeyRotations retrieves all DID key rotations
func (k *Keeper) GetAllDIDKeyRotations(ctx sdk.Context) ([]*types.DIDKeyRotation, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.DIDKeyRotationPrefix, storetypes.PrefixEndBytes(types.DIDKeyRotationPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var rotations []*types.DIDKeyRotation
	for ; iterator.Valid(); iterator.Next() {
		var rotation types.DIDKeyRotation
		if err := k.cdc.Unmarshal(iterator.Value(), &rotation); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DID key rotation: %w", err)
		}
		rotations = append(rotations, &rotation)
	}
	return rotations, nil
}

// DeleteDIDKeyRotation removes a DID key rotation
func (k *Keeper) DeleteDIDKeyRotation(ctx sdk.Context, did string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetDIDKeyRotationKey(did)
	return store.Delete(key)
}

// ============================================================================
// DID Key History Management
// ============================================================================

// AddKeyToHistory adds a key to the DID's key history
func (k *Keeper) AddKeyToHistory(ctx sdk.Context, did, verificationMethod string, activeFrom time.Time, activeUntil *time.Time, rotatedBy, reason string, isCurrent bool) error {
	history, err := k.GetDIDKeyHistory(ctx, did)
	if err != nil {
		// Create new history if not found
		history = &types.DIDKeyHistory{
			Did:     did,
			Entries: []*types.DIDKeyHistoryEntry{},
		}
	}

	// Create history entry
	entry := &types.DIDKeyHistoryEntry{
		VerificationMethod: verificationMethod,
		ActiveFrom:         timestamppb.New(activeFrom),
		RotatedBy:          rotatedBy,
		RotationReason:     reason,
		IsCurrent:          isCurrent,
	}

	if activeUntil != nil {
		entry.ActiveUntil = timestamppb.New(*activeUntil)
	}

	// Add to history
	history.Entries = append(history.Entries, entry)

	return k.SetDIDKeyHistory(ctx, history)
}

// UpdateKeyHistoryActiveUntil updates the active_until time for a key in history
func (k *Keeper) UpdateKeyHistoryActiveUntil(ctx sdk.Context, did, verificationMethod string, activeUntil *time.Time) error {
	history, err := k.GetDIDKeyHistory(ctx, did)
	if err != nil {
		return fmt.Errorf("key history not found for %s: %w", did, err)
	}

	// Find and update the entry
	found := false
	for i, entry := range history.Entries {
		if entry.VerificationMethod == verificationMethod {
			if activeUntil != nil {
				history.Entries[i].ActiveUntil = timestamppb.New(*activeUntil)
			}
			history.Entries[i].IsCurrent = false
			found = true
			break
		}
	}

	if !found {
		return types.ErrKeyNotValid.Wrapf("key %s not found in history for %s", verificationMethod, did)
	}

	return k.SetDIDKeyHistory(ctx, history)
}

// SetDIDKeyHistory stores a DID key history
func (k *Keeper) SetDIDKeyHistory(ctx sdk.Context, history *types.DIDKeyHistory) error {
	if history.Did == "" {
		return types.ErrInvalidDID.Wrap("DID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal DID key history: %w", err)
	}

	key := types.GetDIDKeyHistoryKey(history.Did)
	return store.Set(key, bz)
}

// GetDIDKeyHistory retrieves a DID key history
func (k *Keeper) GetDIDKeyHistory(ctx sdk.Context, did string) (*types.DIDKeyHistory, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetDIDKeyHistoryKey(did)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrDIDKeyRotationNotFound.Wrapf("key history not found for %s", did)
	}

	var history types.DIDKeyHistory
	if err := k.cdc.Unmarshal(bz, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DID key history: %w", err)
	}
	return &history, nil
}

// GetAllDIDKeyHistories retrieves all DID key histories
func (k *Keeper) GetAllDIDKeyHistories(ctx sdk.Context) ([]*types.DIDKeyHistory, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.DIDKeyHistoryPrefix, storetypes.PrefixEndBytes(types.DIDKeyHistoryPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var histories []*types.DIDKeyHistory
	for ; iterator.Valid(); iterator.Next() {
		var history types.DIDKeyHistory
		if err := k.cdc.Unmarshal(iterator.Value(), &history); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DID key history: %w", err)
		}
		histories = append(histories, &history)
	}
	return histories, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// IsKeyInGracePeriod checks if a key is currently in its grace period
func (k *Keeper) IsKeyInGracePeriod(ctx sdk.Context, did, verificationMethod string) bool {
	rotation, err := k.GetDIDKeyRotation(ctx, did)
	if err != nil {
		return false
	}

	if rotation.Status != types.DIDKeyRotationStatusPending {
		return false
	}

	if rotation.OldVerificationMethod != verificationMethod {
		return false
	}

	now := ctx.BlockTime()
	return now.Before(rotation.GracePeriodEnd.AsTime())
}

// GetCurrentVerificationMethod returns the current (primary) verification method for a DID
func (k *Keeper) GetCurrentVerificationMethod(ctx sdk.Context, did string) (string, error) {
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return "", types.ErrIdentityNotFound.Wrapf("identity not found: %s", did)
	}

	if len(record.VerificationMethods) == 0 {
		return "", types.ErrInvalidVerificationMethod.Wrap("no verification methods found")
	}

	return record.VerificationMethods[0], nil
}

// ProcessExpiredGracePeriods processes all rotations with expired grace periods
// This should be called in BeginBlocker or EndBlocker
func (k *Keeper) ProcessExpiredGracePeriods(ctx sdk.Context) error {
	rotations, err := k.GetAllDIDKeyRotations(ctx)
	if err != nil {
		return err
	}

	now := ctx.BlockTime()
	for _, rotation := range rotations {
		if rotation.Status == types.DIDKeyRotationStatusPending {
			if !now.Before(rotation.GracePeriodEnd.AsTime()) {
				// Grace period has expired, complete the rotation
				if err := k.CompleteKeyRotation(ctx, rotation.Did); err != nil {
					k.logger.Error("failed to complete key rotation", "did", rotation.Did, "error", err)
					// Continue processing other rotations
					continue
				}
			}
		}
	}

	return nil
}

// VerifySignatureWithKey verifies a signature using a verification method
// This is a placeholder for actual signature verification logic
// In production, this would integrate with cryptographic verification
func (k *Keeper) VerifySignatureWithKey(ctx sdk.Context, did, verificationMethod string, message, signature []byte) error {
	// CRITICAL: Check if verification method (credential) has been revoked
	// This is the primary security check and must be performed first
	if k.IsCredentialRevoked(ctx, verificationMethod) {
		return types.ErrCredentialRevoked.Wrapf(
			"verification method %s has been revoked and cannot be used for authentication",
			verificationMethod,
		)
	}

	// Check if the key is valid
	if err := k.ValidateDIDKey(ctx, did, verificationMethod); err != nil {
		return err
	}

	// TODO: Implement actual signature verification
	// This would involve:
	// 1. Parsing the verification method to get the public key
	// 2. Using the appropriate cryptographic algorithm (ECDSA, Ed25519, etc.)
	// 3. Verifying the signature against the message and public key

	// For now, just check if the key is valid
	// In production, this would call actual crypto verification functions

	// Compare signature bytes as placeholder
	if len(signature) == 0 {
		return fmt.Errorf("empty signature")
	}

	// Placeholder: actual verification would happen here
	return nil
}
