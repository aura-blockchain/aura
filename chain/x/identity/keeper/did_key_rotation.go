// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
		Did:                   did,
		OldVerificationMethod: oldVerificationMethod,
		NewVerificationMethod: newVerificationMethod,
		RotationTime:          now,
		InitiatedBy:           initiator,
		Reason:                reason,
		GracePeriodEnd:        gracePeriodEnd,
		Status:                types.DIDKeyRotationStatusPending,
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
	updatedAt := now
	record.UpdatedAt = &updatedAt

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
			if now.Before(rotation.GracePeriodEnd) {
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
				if entry.ActiveUntil != nil && now.Before(*entry.ActiveUntil) {
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
	if now.Before(rotation.GracePeriodEnd) {
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
	updatedAt2 := now
	record.UpdatedAt = &updatedAt2

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

	rotations := make([]*types.DIDKeyRotation, 0, 64)
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
		ActiveFrom:         activeFrom,
		RotatedBy:          rotatedBy,
		RotationReason:     reason,
		IsCurrent:          isCurrent,
		ActiveUntil:        activeUntil,
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
			history.Entries[i].ActiveUntil = activeUntil
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

	histories := make([]*types.DIDKeyHistory, 0, 64)
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
	return now.Before(rotation.GracePeriodEnd)
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
		return fmt.Errorf("failed to get for GetAllDIDKeyRotations: %w", err)
	}

	now := ctx.BlockTime()
	for _, rotation := range rotations {
		if rotation.Status == types.DIDKeyRotationStatusPending {
			if !now.Before(rotation.GracePeriodEnd) {
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
// This function performs cryptographic signature verification using the public key
// associated with the DID's verification method. It supports both secp256k1 and
// ed25519 signature schemes.
//
// Security considerations:
//   - Revocation check: Verification method must not be revoked
//   - Key validity: Verification method must be current or in grace period
//   - Cryptographic verification: Uses battle-tested Cosmos SDK crypto libraries
//   - Hash verification: Message is hashed with SHA-256 before verification
//
// Parameters:
//   - ctx: SDK context for state access
//   - did: Decentralized identifier
//   - verificationMethod: Public key identifier (hex or base64 encoded)
//   - message: Original message bytes to verify
//   - signature: Signature bytes to verify against message
//
// Returns:
//   - error: ErrCredentialRevoked, ErrKeyNotValid, ErrInvalidSignature, or parsing errors
func (k *Keeper) VerifySignatureWithKey(ctx sdk.Context, did, verificationMethod string, message, signature []byte) error {
	// CRITICAL: Check if verification method (credential) has been revoked
	// This is the primary security check and must be performed first
	// Revoked credentials cannot be used for any authentication or signing
	if k.IsCredentialRevoked(ctx, verificationMethod) {
		return types.ErrCredentialRevoked.Wrapf(
			"verification method %s has been revoked and cannot be used for authentication",
			verificationMethod,
		)
	}

	// Validate that the verification method is currently valid for this DID
	// This checks:
	// - Method is the current active key, OR
	// - Method is an old key still within grace period, OR
	// - Method is in key history and still within its active period
	if err := k.ValidateDIDKey(ctx, did, verificationMethod); err != nil {
		return types.ErrKeyNotValid.Wrapf(
			"verification method %s not valid for DID %s: %v",
			verificationMethod, did, err,
		)
	}

	// Validate inputs
	if len(message) == 0 {
		return types.ErrInvalidSignature.Wrap("message cannot be empty")
	}
	if len(signature) == 0 {
		return types.ErrInvalidSignature.Wrap("signature cannot be empty")
	}

	// Parse the verification method to extract the public key
	// Verification methods are stored as encoded public keys (hex or base64)
	pubKey, err := k.parseVerificationMethod(verificationMethod)
	if err != nil {
		return types.ErrInvalidVerificationMethod.Wrapf(
			"failed to parse verification method %s: %v",
			verificationMethod, err,
		)
	}

	// Hash the message using SHA-256
	// This is the standard approach for signature verification in Cosmos SDK
	messageHash := sha256.Sum256(message)

	// Verify the signature using the parsed public key
	// This supports both secp256k1 (ECDSA) and ed25519 signature schemes
	if !k.verifySignature(pubKey, messageHash[:], signature) {
		return types.ErrInvalidSignature.Wrapf(
			"signature verification failed for verification method %s",
			verificationMethod,
		)
	}

	// Signature verification successful
	return nil
}

// parseVerificationMethod parses a verification method string to extract the public key
// Supports both hex-encoded and base64-encoded public keys
// Returns the appropriate cryptotypes.PubKey based on key length and format
func (k *Keeper) parseVerificationMethod(verificationMethod string) (cryptotypes.PubKey, error) {
	if verificationMethod == "" {
		return nil, fmt.Errorf("verification method cannot be empty")
	}

	// Try hex decoding first (common format for public keys)
	pubKeyBytes, err := hex.DecodeString(verificationMethod)
	if err != nil {
		// If hex fails, try base64 decoding
		pubKeyBytes, err = base64.StdEncoding.DecodeString(verificationMethod)
		if err != nil {
			return nil, fmt.Errorf("verification method must be hex or base64 encoded: %w", err)
		}
	}

	// Validate we have key data
	if len(pubKeyBytes) == 0 {
		return nil, fmt.Errorf("decoded verification method is empty")
	}

	// Determine key type based on length and construct appropriate public key
	// secp256k1: 33 bytes (compressed) or 65 bytes (uncompressed)
	// ed25519: 32 bytes
	switch len(pubKeyBytes) {
	case 33:
		// secp256k1 compressed public key (most common in Cosmos SDK)
		return &secp256k1.PubKey{Key: pubKeyBytes}, nil

	case 65:
		// secp256k1 uncompressed public key
		// Convert to compressed format for consistency
		if pubKeyBytes[0] != 0x04 {
			return nil, fmt.Errorf("invalid uncompressed secp256k1 public key format")
		}
		return &secp256k1.PubKey{Key: pubKeyBytes}, nil

	case 32:
		// ed25519 public key
		return &ed25519.PubKey{Key: pubKeyBytes}, nil

	default:
		return nil, fmt.Errorf(
			"unsupported public key length: %d bytes (expected 32 for ed25519, 33 or 65 for secp256k1)",
			len(pubKeyBytes),
		)
	}
}

// verifySignature performs the actual cryptographic signature verification
// Supports both secp256k1 (ECDSA) and ed25519 signature schemes
// Returns true if signature is valid, false otherwise
func (k *Keeper) verifySignature(pubKey cryptotypes.PubKey, messageHash, signature []byte) bool {
	if pubKey == nil {
		k.logger.Error("nil public key provided for signature verification")
		return false
	}

	// Use type assertion to call the appropriate verification method
	// This allows us to handle different key types correctly
	switch pk := pubKey.(type) {
	case *secp256k1.PubKey:
		// Verify secp256k1 (ECDSA) signature
		// This is the most common signature type in Cosmos SDK blockchains
		return pk.VerifySignature(messageHash, signature)

	case *ed25519.PubKey:
		// Verify ed25519 signature
		// Used for high-performance or specific security requirements
		return pk.VerifySignature(messageHash, signature)

	default:
		// Fallback: use the generic VerifySignature interface method
		// This handles any other cryptotypes.PubKey implementations
		k.logger.Warn("using fallback signature verification for unknown key type")
		return pubKey.VerifySignature(messageHash, signature)
	}
}
