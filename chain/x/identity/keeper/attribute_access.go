package keeper

import (
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// ============================================================================
// Attribute Permission Management
// ============================================================================

// GrantAttributeAccess grants access to an attribute with specified level
//
// Security considerations:
//   - Only owner can grant access (enforced by transaction signer)
//   - Expiry time is optional (zero time means no expiry)
//   - Records consent for GDPR compliance
//   - Emits event for audit trail
//
// Parameters:
//   - ctx: SDK context for state access
//   - owner: Address of attribute owner (must be signer)
//   - attribute: Name of attribute to grant access to
//   - grantee: Address to grant access to ("*" for public)
//   - level: Access level (NONE, VERIFY_ONLY, READ)
//   - expiry: Optional expiry time (zero time for no expiry)
//   - purpose: Purpose for data sharing (GDPR requirement)
//
// Returns:
//   - error: ErrInvalidInput, ErrInvalidAccessLevel, or storage error
//
// Events emitted:
//   - EventAttributeAccessGranted with owner, attribute, grantee, and level
func (k *Keeper) GrantAttributeAccess(ctx sdk.Context, owner, attribute, grantee string, level identitypb.AccessLevel, expiry time.Time, purpose string) error {
	// Validate inputs
	if owner == "" {
		return types.ErrInvalidInput.Wrap("owner cannot be empty")
	}
	if attribute == "" {
		return types.ErrInvalidInput.Wrap("attribute name cannot be empty")
	}
	if grantee == "" {
		return types.ErrInvalidInput.Wrap("grantee cannot be empty")
	}
	if level == identitypb.AccessLevel_ACCESS_LEVEL_UNSPECIFIED {
		return types.ErrInvalidAccessLevel.Wrap("access level must be specified")
	}

	now := ctx.BlockTime()

	// Create permission record
	// GrantedAt is time.Time, ExpiresAt is *time.Time (nullable)
	permission := &identitypb.AttributePermission{
		AttributeName: attribute,
		GrantedTo:     grantee,
		GrantedAt:     now,
		ExpiresAt:     &expiry,
		AccessLevel:   level,
		GrantedBy:     owner,
		Metadata:      purpose,
	}

	// Store permission
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAttributePermissionKey(owner, attribute, grantee)
	bz, err := k.cdc.Marshal(permission)
	if err != nil {
		return fmt.Errorf("failed to marshal permission: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return fmt.Errorf("failed to store permission: %w", err)
	}

	// Record consent
	// ConsentedAt is time.Time, ExpiresAt is *time.Time (nullable)
	consent := &identitypb.AttributeConsentRecord{
		Did:           owner,
		AttributeName: attribute,
		Grantee:       grantee,
		Purpose:       purpose,
		ConsentedAt:   now,
		ExpiresAt:     &expiry,
		Revoked:       false,
		AccessLevel:   level,
	}
	consentKey := types.GetAttributeConsentKey(owner, attribute, grantee)
	consentBz, err := k.cdc.Marshal(consent)
	if err != nil {
		return fmt.Errorf("failed to marshal consent: %w", err)
	}
	if err := store.Set(consentKey, consentBz); err != nil {
		return fmt.Errorf("failed to store consent: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"attribute_access_granted",
			sdk.NewAttribute("owner", owner),
			sdk.NewAttribute("attribute", attribute),
			sdk.NewAttribute("grantee", grantee),
			sdk.NewAttribute("access_level", level.String()),
			sdk.NewAttribute("expires_at", expiry.String()),
			sdk.NewAttribute("purpose", purpose),
		),
	)

	return nil
}

// RevokeAttributeAccess revokes access to an attribute
//
// Security considerations:
//   - Only owner can revoke access
//   - Records revocation in consent record for audit
//   - Emits event for transparency
//
// Parameters:
//   - ctx: SDK context for state access
//   - owner: Address of attribute owner
//   - attribute: Name of attribute
//   - grantee: Address to revoke access from
//   - reason: Reason for revocation (audit trail)
//
// Returns:
//   - error: ErrPermissionNotFound or storage error
func (k *Keeper) RevokeAttributeAccess(ctx sdk.Context, owner, attribute, grantee, reason string) error {
	// Validate inputs
	if owner == "" {
		return types.ErrInvalidInput.Wrap("owner cannot be empty")
	}
	if attribute == "" {
		return types.ErrInvalidInput.Wrap("attribute name cannot be empty")
	}
	if grantee == "" {
		return types.ErrInvalidInput.Wrap("grantee cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	now := ctx.BlockTime()

	// Delete permission
	key := types.GetAttributePermissionKey(owner, attribute, grantee)
	if err := store.Delete(key); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	// Update consent record to mark as revoked
	consentKey := types.GetAttributeConsentKey(owner, attribute, grantee)
	consentBz, err := store.Get(consentKey)
	if err == nil && consentBz != nil {
		var consent identitypb.AttributeConsentRecord
		if err := k.cdc.Unmarshal(consentBz, &consent); err == nil {
			consent.Revoked = true
			consent.RevokedAt = &now
			consent.RevocationReason = reason

			updatedBz, err := k.cdc.Marshal(&consent)
			if err == nil {
				_ = store.Set(consentKey, updatedBz)
			}
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"attribute_access_revoked",
			sdk.NewAttribute("owner", owner),
			sdk.NewAttribute("attribute", attribute),
			sdk.NewAttribute("grantee", grantee),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// CanAccessAttribute checks if requester can access attribute and returns access level
//
// Security considerations:
//   - Owner always has full READ access
//   - Checks specific grant first, then public grant
//   - Validates expiry time
//   - Returns appropriate access level
//
// Parameters:
//   - ctx: SDK context for state access
//   - owner: Address of attribute owner
//   - attribute: Name of attribute
//   - requester: Address requesting access
//
// Returns:
//   - AccessLevel: The level of access granted
//   - error: ErrAccessDenied, ErrAccessExpired, or storage error
func (k *Keeper) CanAccessAttribute(ctx sdk.Context, owner, attribute, requester string) (identitypb.AccessLevel, error) {
	// Owner always has full access
	if owner == requester {
		return identitypb.AccessLevel_ACCESS_LEVEL_READ, nil
	}

	store := k.storeService.OpenKVStore(ctx)
	now := ctx.BlockTime()

	// Check specific grant
	key := types.GetAttributePermissionKey(owner, attribute, requester)
	bz, err := store.Get(key)
	if err == nil && bz != nil {
		var permission identitypb.AttributePermission
		if err := k.cdc.Unmarshal(bz, &permission); err == nil {
			// Check expiry - ExpiresAt is *time.Time (nullable pointer)
			if permission.ExpiresAt != nil && !permission.ExpiresAt.IsZero() {
				if now.After(*permission.ExpiresAt) {
					return identitypb.AccessLevel_ACCESS_LEVEL_NONE, types.ErrAccessExpired.Wrapf(
						"permission expired at %s", *permission.ExpiresAt)
				}
			}
			return permission.AccessLevel, nil
		}
	}

	// Check public grant
	publicKey := types.GetAttributePermissionKey(owner, attribute, "*")
	publicBz, err := store.Get(publicKey)
	if err == nil && publicBz != nil {
		var permission identitypb.AttributePermission
		if err := k.cdc.Unmarshal(publicBz, &permission); err == nil {
			// Check expiry - ExpiresAt is *time.Time (nullable pointer)
			if permission.ExpiresAt != nil && !permission.ExpiresAt.IsZero() {
				if now.After(*permission.ExpiresAt) {
					return identitypb.AccessLevel_ACCESS_LEVEL_NONE, types.ErrAccessExpired.Wrapf(
						"public permission expired at %s", *permission.ExpiresAt)
				}
			}
			return permission.AccessLevel, nil
		}
	}

	return identitypb.AccessLevel_ACCESS_LEVEL_NONE, types.ErrAccessDenied.Wrapf(
		"requester %s does not have access to attribute %s of owner %s", requester, attribute, owner)
}

// GetAttributeWithAccessControl retrieves an attribute with access control enforcement
//
// Security considerations:
//   - Enforces access control via CanAccessAttribute
//   - Returns only commitment for VERIFY_ONLY access
//   - Returns full value for READ access
//   - Logs access attempt for audit
//
// Parameters:
//   - ctx: SDK context for state access
//   - owner: Address of attribute owner
//   - attribute: Name of attribute
//   - requester: Address requesting access
//
// Returns:
//   - value: Attribute value (commitment for VERIFY_ONLY, full value for READ)
//   - error: ErrAccessDenied, ErrAttributeNotFound, or storage error
//
// Note: This is a placeholder - actual attribute storage needs to be implemented
// based on how attributes are stored in IdentityRecord or separate storage
func (k *Keeper) GetAttributeWithAccessControl(ctx sdk.Context, owner, attribute, requester string) ([]byte, error) {
	// Check access permission
	level, err := k.CanAccessAttribute(ctx, owner, attribute, requester)
	if err != nil {
		// Log failed access attempt
		_ = k.logAttributeAccess(ctx, owner, attribute, requester, level, false, err.Error())
		return nil, err
	}

	if level == identitypb.AccessLevel_ACCESS_LEVEL_NONE {
		_ = k.logAttributeAccess(ctx, owner, attribute, requester, level, false, "access denied")
		return nil, types.ErrAccessDenied.Wrapf("no access to attribute %s", attribute)
	}

	// Get identity record
	record, err := k.GetIdentityRecord(ctx, owner)
	if err != nil {
		_ = k.logAttributeAccess(ctx, owner, attribute, requester, level, false, "identity not found")
		return nil, err
	}

	// For GDPR-compliant storage, we only have PII commitment
	// Actual attributes would need to be retrieved from off-chain storage
	// This implementation returns the commitment or a reference
	var value []byte
	if level == identitypb.AccessLevel_ACCESS_LEVEL_VERIFY_ONLY {
		// Return only commitment for verification
		value = record.PiiCommitment
	} else if level == identitypb.AccessLevel_ACCESS_LEVEL_READ {
		// In production, this would fetch from off-chain storage using off_chain_data_ref
		// For now, return metadata hash as a placeholder
		value = []byte(record.MetadataHash)
	}

	// Log successful access
	_ = k.logAttributeAccess(ctx, owner, attribute, requester, level, true, "")

	return value, nil
}

// GetAttributePermissions retrieves all permissions for an attribute
func (k *Keeper) GetAttributePermissions(ctx sdk.Context, owner, attribute string) ([]*identitypb.AttributePermission, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetAttributePermissionsByAttributePrefix(owner, attribute)

	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var permissions []*identitypb.AttributePermission
	for ; iterator.Valid(); iterator.Next() {
		var permission identitypb.AttributePermission
		if err := k.cdc.Unmarshal(iterator.Value(), &permission); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permission: %w", err)
		}
		permissions = append(permissions, &permission)
	}

	return permissions, nil
}

// GetAllAttributePermissions retrieves all permissions for an owner
func (k *Keeper) GetAllAttributePermissions(ctx sdk.Context, owner string) ([]*identitypb.AttributePermission, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetAttributePermissionsByOwnerPrefix(owner)

	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var permissions []*identitypb.AttributePermission
	for ; iterator.Valid(); iterator.Next() {
		var permission identitypb.AttributePermission
		if err := k.cdc.Unmarshal(iterator.Value(), &permission); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permission: %w", err)
		}
		permissions = append(permissions, &permission)
	}

	return permissions, nil
}

// ============================================================================
// Attribute Access Logging
// ============================================================================

// logAttributeAccess records an attribute access event for audit trail
func (k *Keeper) logAttributeAccess(ctx sdk.Context, owner, attribute, requester string, level identitypb.AccessLevel, success bool, errorMsg string) error {
	store := k.storeService.OpenKVStore(ctx)

	// Get next log ID
	counterBz, err := store.Get(types.AttributeAccessLogCounterPrefix)
	var logID uint64 = 1
	if err == nil && counterBz != nil {
		logID = sdk.BigEndianToUint64(counterBz)
	}

	now := ctx.BlockTime()
	log := &identitypb.AttributeAccessLog{
		Id:            fmt.Sprintf("access-%d", logID),
		Owner:         owner,
		AttributeName: attribute,
		Requester:     requester,
		AccessLevel:   level,
		AccessedAt:    now,
		Success:       success,
		ErrorMessage:  errorMsg,
		BlockHeight:   ctx.BlockHeight(),
	}

	// Store log
	key := types.GetAttributeAccessLogKey(logID)
	bz, err := k.cdc.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal access log: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return fmt.Errorf("failed to store access log: %w", err)
	}

	// Increment counter
	if err := store.Set(types.AttributeAccessLogCounterPrefix, sdk.Uint64ToBigEndian(logID+1)); err != nil {
		return fmt.Errorf("failed to update log counter: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"attribute_accessed",
			sdk.NewAttribute("owner", owner),
			sdk.NewAttribute("attribute", attribute),
			sdk.NewAttribute("requester", requester),
			sdk.NewAttribute("access_level", level.String()),
			sdk.NewAttribute("success", fmt.Sprintf("%t", success)),
		),
	)

	return nil
}

// GetAttributeAccessLogs retrieves access logs for an owner's attributes
func (k *Keeper) GetAttributeAccessLogs(ctx sdk.Context, owner string, limit, offset uint64) ([]*identitypb.AttributeAccessLog, uint64, error) {
	store := k.storeService.OpenKVStore(ctx)

	// Iterate all logs and filter by owner
	// In production, this would use a secondary index for efficiency
	iterator, err := store.Iterator(types.AttributeAccessLogPrefix, storetypes.PrefixEndBytes(types.AttributeAccessLogPrefix))
	if err != nil {
		return nil, 0, err
	}
	defer iterator.Close()

	var logs []*identitypb.AttributeAccessLog
	var total uint64
	var skipped uint64

	for ; iterator.Valid(); iterator.Next() {
		var log identitypb.AttributeAccessLog
		if err := k.cdc.Unmarshal(iterator.Value(), &log); err != nil {
			continue
		}

		// Filter by owner
		if log.Owner != owner {
			continue
		}

		total++

		// Apply offset
		if skipped < offset {
			skipped++
			continue
		}

		// Apply limit
		if limit > 0 && uint64(len(logs)) >= limit {
			continue
		}

		logs = append(logs, &log)
	}

	return logs, total, nil
}

// GetAttributeConsent retrieves consent record for an attribute
func (k *Keeper) GetAttributeConsent(ctx sdk.Context, did, attribute, grantee string) (*identitypb.AttributeConsentRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAttributeConsentKey(did, attribute, grantee)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrPermissionNotFound.Wrapf("consent not found for attribute %s", attribute)
	}

	var consent identitypb.AttributeConsentRecord
	if err := k.cdc.Unmarshal(bz, &consent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consent: %w", err)
	}

	return &consent, nil
}

// GetAllAttributeConsents retrieves all consent records for a DID
func (k *Keeper) GetAllAttributeConsents(ctx sdk.Context, did string) ([]*identitypb.AttributeConsentRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := types.GetAttributeConsentByDIDPrefix(did)

	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var consents []*identitypb.AttributeConsentRecord
	for ; iterator.Valid(); iterator.Next() {
		var consent identitypb.AttributeConsentRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &consent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal consent: %w", err)
		}
		consents = append(consents, &consent)
	}

	return consents, nil
}
