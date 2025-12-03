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
// Credential Revocation Management
// ============================================================================

// IsCredentialRevoked checks if a credential has been revoked
// This is the primary verification check that MUST be called before accepting any credential
func (k *Keeper) IsCredentialRevoked(ctx sdk.Context, credentialID string) bool {
	if credentialID == "" {
		return false
	}

	store := k.storeService.OpenKVStore(ctx)
	key := types.GetCredentialRevocationKey(credentialID)
	has, err := store.Has(key)
	if err != nil {
		return false
	}
	return has
}

// GetCredentialRevocation retrieves a credential revocation record
func (k *Keeper) GetCredentialRevocation(ctx sdk.Context, credentialID string) (*types.CredentialRevocation, error) {
	if credentialID == "" {
		return nil, types.ErrInvalidCredentialID.Wrap("credential ID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	key := types.GetCredentialRevocationKey(credentialID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrCredentialNotFound.Wrapf("credential revocation not found: %s", credentialID)
	}

	var revocation types.CredentialRevocation
	if err := k.cdc.Unmarshal(bz, &revocation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential revocation: %w", err)
	}
	return &revocation, nil
}

// RevokeCredential revokes a credential with full audit trail
// This function enforces authorization and creates an indexed revocation record
func (k *Keeper) RevokeCredential(ctx sdk.Context, credentialID, did, revoker, reason string, metadata map[string]string) error {
	// Validate inputs
	if credentialID == "" {
		return types.ErrInvalidCredentialID.Wrap("credential ID cannot be empty")
	}
	if did == "" {
		return types.ErrInvalidDID.Wrap("DID cannot be empty")
	}
	if revoker == "" {
		return types.ErrInvalidAddress.Wrap("revoker address cannot be empty")
	}

	// Check if already revoked
	if k.IsCredentialRevoked(ctx, credentialID) {
		return types.ErrCredentialAlreadyRevoked.Wrapf("credential %s already revoked", credentialID)
	}

	// Verify the identity exists
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return types.ErrIdentityNotFound.Wrapf("identity %s not found", did)
	}

	// Check authorization: revoker must be owner or have VerifyIdentity permission
	if record.Address != revoker {
		if err := k.RequirePermission(ctx, revoker, types.PermissionVerifyIdentity); err != nil {
			return types.ErrUnauthorized.Wrapf("revoker %s not authorized to revoke credentials for %s", revoker, did)
		}
	}

	// Create revocation record
	now := ctx.BlockTime()
	revocation := &types.CredentialRevocation{
		CredentialId: credentialID,
		Did:          did,
		RevokedAt:    timestamppb.New(now),
		RevokedBy:    revoker,
		Reason:       reason,
		Metadata:     metadata,
	}

	// Store revocation in indexed storage
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetCredentialRevocationKey(credentialID)
	bz, err := k.cdc.Marshal(revocation)
	if err != nil {
		return fmt.Errorf("failed to marshal credential revocation: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return fmt.Errorf("failed to store credential revocation: %w", err)
	}

	// Emit event for indexing and monitoring
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCredentialRevoked,
			sdk.NewAttribute(types.AttributeKeyCredentialId, credentialID),
			sdk.NewAttribute(types.AttributeKeyDID, did),
			sdk.NewAttribute(types.AttributeKeyRevokedBy, revoker),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyTimestamp, now.Format(time.RFC3339)),
		),
	)

	// Log audit trail
	auditMetadata := map[string]string{
		"credential_id": credentialID,
		"did":           did,
		"reason":        reason,
	}
	for k, v := range metadata {
		auditMetadata[k] = v
	}
	k.LogAudit(ctx, revoker, "revoke_credential", credentialID, "success", auditMetadata, "")

	return nil
}

// BatchRevokeCredentials revokes multiple credentials in a single transaction
// This is more efficient than individual revocations and ensures atomicity
func (k *Keeper) BatchRevokeCredentials(ctx sdk.Context, credentialIDs []string, did, revoker, reason string) error {
	if len(credentialIDs) == 0 {
		return types.ErrInvalidInput.Wrap("credential IDs list cannot be empty")
	}

	// Verify authorization once for the batch
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return types.ErrIdentityNotFound.Wrapf("identity %s not found", did)
	}

	if record.Address != revoker {
		if err := k.RequirePermission(ctx, revoker, types.PermissionVerifyIdentity); err != nil {
			return types.ErrUnauthorized.Wrapf("revoker %s not authorized", revoker)
		}
	}

	// Revoke each credential
	successCount := 0
	failedCredentials := make([]string, 0)

	for _, credentialID := range credentialIDs {
		if credentialID == "" {
			failedCredentials = append(failedCredentials, credentialID)
			continue
		}

		// Skip if already revoked
		if k.IsCredentialRevoked(ctx, credentialID) {
			continue
		}

		// Perform revocation
		metadata := map[string]string{
			"batch_revocation": "true",
			"batch_reason":     reason,
		}
		if err := k.RevokeCredential(ctx, credentialID, did, revoker, reason, metadata); err != nil {
			k.logger.Error("failed to revoke credential in batch",
				"credential_id", credentialID,
				"error", err.Error())
			failedCredentials = append(failedCredentials, credentialID)
			continue
		}
		successCount++
	}

	// Emit batch event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBatchCredentialRevocation,
			sdk.NewAttribute(types.AttributeKeyDID, did),
			sdk.NewAttribute(types.AttributeKeyRevokedBy, revoker),
			sdk.NewAttribute("total_count", fmt.Sprintf("%d", len(credentialIDs))),
			sdk.NewAttribute("success_count", fmt.Sprintf("%d", successCount)),
			sdk.NewAttribute("failed_count", fmt.Sprintf("%d", len(failedCredentials))),
		),
	)

	// Log batch audit
	k.LogAudit(ctx, revoker, "batch_revoke_credentials", did, "completed", map[string]string{
		"total":   fmt.Sprintf("%d", len(credentialIDs)),
		"success": fmt.Sprintf("%d", successCount),
		"failed":  fmt.Sprintf("%d", len(failedCredentials)),
		"reason":  reason,
	}, "")

	if len(failedCredentials) > 0 {
		return fmt.Errorf("batch revocation completed with %d failures out of %d", len(failedCredentials), len(credentialIDs))
	}

	return nil
}

// VerifyCredential verifies a credential with comprehensive checks
// This is the main entry point for credential verification
func (k *Keeper) VerifyCredential(ctx sdk.Context, credentialID, did string) error {
	// Validate inputs
	if credentialID == "" {
		return types.ErrInvalidCredentialID.Wrap("credential ID cannot be empty")
	}
	if did == "" {
		return types.ErrInvalidDID.Wrap("DID cannot be empty")
	}

	// CHECK REVOCATION FIRST - this is the critical security check
	if k.IsCredentialRevoked(ctx, credentialID) {
		// Get revocation details for audit
		revocation, err := k.GetCredentialRevocation(ctx, credentialID)
		if err == nil {
			return types.ErrCredentialRevoked.Wrapf(
				"credential %s was revoked at %s by %s, reason: %s",
				credentialID,
				revocation.RevokedAt.AsTime().Format(time.RFC3339),
				revocation.RevokedBy,
				revocation.Reason,
			)
		}
		return types.ErrCredentialRevoked.Wrapf("credential %s has been revoked", credentialID)
	}

	// Verify identity exists and is active
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return types.ErrIdentityNotFound.Wrapf("identity %s not found", did)
	}

	// Check if identity is erased (GDPR compliance)
	if record.Erased {
		return types.ErrIdentityErased.Wrapf("identity %s has been erased", did)
	}

	// Check identity status
	if record.Status != types.IdentityStatusActive {
		return types.ErrIdentityNotFound.Wrapf("identity %s is not active (status: %s)", did, record.Status.String())
	}

	// Additional credential verification logic can be added here
	// For example: expiration checks, signature verification, etc.

	return nil
}

// GetAllCredentialRevocations retrieves all credential revocations
func (k *Keeper) GetAllCredentialRevocations(ctx sdk.Context) ([]*types.CredentialRevocation, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.CredentialRevocationPrefix, storetypes.PrefixEndBytes(types.CredentialRevocationPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var revocations []*types.CredentialRevocation
	for ; iterator.Valid(); iterator.Next() {
		var revocation types.CredentialRevocation
		if err := k.cdc.Unmarshal(iterator.Value(), &revocation); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credential revocation: %w", err)
		}
		revocations = append(revocations, &revocation)
	}
	return revocations, nil
}

// GetCredentialRevocationsByDID retrieves all revocations for a specific DID
func (k *Keeper) GetCredentialRevocationsByDID(ctx sdk.Context, did string) ([]*types.CredentialRevocation, error) {
	allRevocations, err := k.GetAllCredentialRevocations(ctx)
	if err != nil {
		return nil, err
	}

	var didRevocations []*types.CredentialRevocation
	for _, revocation := range allRevocations {
		if revocation.Did == did {
			didRevocations = append(didRevocations, revocation)
		}
	}
	return didRevocations, nil
}

// RestoreCredential removes a revocation (for administrative corrections only)
// This should be used sparingly and only with proper authorization
func (k *Keeper) RestoreCredential(ctx sdk.Context, credentialID, admin, reason string) error {
	// Validate inputs
	if credentialID == "" {
		return types.ErrInvalidCredentialID.Wrap("credential ID cannot be empty")
	}

	// Require admin permission for restoration
	if err := k.RequirePermission(ctx, admin, types.PermissionAdmin); err != nil {
		return types.ErrUnauthorized.Wrapf("admin %s not authorized to restore credentials", admin)
	}

	// Check if credential is actually revoked
	if !k.IsCredentialRevoked(ctx, credentialID) {
		return types.ErrCredentialNotFound.Wrapf("credential %s is not revoked", credentialID)
	}

	// Get revocation details for audit
	revocation, _ := k.GetCredentialRevocation(ctx, credentialID)

	// Remove revocation
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetCredentialRevocationKey(credentialID)
	if err := store.Delete(key); err != nil {
		return fmt.Errorf("failed to delete credential revocation: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCredentialRestored,
			sdk.NewAttribute(types.AttributeKeyCredentialId, credentialID),
			sdk.NewAttribute(types.AttributeKeyRestoredBy, admin),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	// Log audit trail
	auditMetadata := map[string]string{
		"credential_id": credentialID,
		"reason":        reason,
	}
	if revocation != nil {
		auditMetadata["previous_revoked_by"] = revocation.RevokedBy
		auditMetadata["previous_revoked_at"] = revocation.RevokedAt.AsTime().Format(time.RFC3339)
		auditMetadata["previous_reason"] = revocation.Reason
	}
	k.LogAudit(ctx, admin, "restore_credential", credentialID, "success", auditMetadata, "")

	return nil
}
