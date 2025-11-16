package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// NONCE MANAGEMENT (Replay Attack Prevention)
// ============================================================================

// GetNonce retrieves the current nonce for an address
func (k Keeper) GetNonce(ctx sdk.Context, address string, chainId string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.NonceKey(address, chainId)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementNonce increments the nonce for an address
func (k Keeper) IncrementNonce(ctx sdk.Context, address string, chainId string) uint64 {
	currentNonce := k.GetNonce(ctx, address, chainId)
	newNonce := currentNonce + 1

	store := ctx.KVStore(k.storeKey)
	key := types.NonceKey(address, chainId)
	bz := sdk.Uint64ToBigEndian(newNonce)
	store.Set(key, bz)

	// Update tracker
	tracker := &types.NonceTracker{
		Address:     address,
		Nonce:       newNonce,
		ChainId:     chainId,
		LastUpdated: ctx.BlockTime(),
	}
	k.SetNonceTracker(ctx, tracker)

	return newNonce
}

// VerifyNonce verifies that a nonce is valid (prevents replay attacks)
func (k Keeper) VerifyNonce(ctx sdk.Context, address string, chainId string, nonce uint64) error {
	expectedNonce := k.GetNonce(ctx, address, chainId)

	if nonce != expectedNonce {
		return fmt.Errorf(
			"invalid nonce for %s on %s: got %d, expected %d",
			address,
			chainId,
			nonce,
			expectedNonce,
		)
	}

	return nil
}

// VerifyAndIncrementNonce verifies and increments nonce atomically
func (k Keeper) VerifyAndIncrementNonce(
	ctx sdk.Context,
	address string,
	chainId string,
	nonce uint64,
) error {
	if err := k.VerifyNonce(ctx, address, chainId, nonce); err != nil {
		return err
	}

	k.IncrementNonce(ctx, address, chainId)
	return nil
}

// ============================================================================
// EMERGENCY PAUSE MECHANISM
// ============================================================================

// PauseBridge pauses all bridge operations (emergency only)
func (k Keeper) PauseBridge(ctx sdk.Context, reason string, pauser string) error {
	params := k.GetParams(ctx)

	if params.EmergencyPaused {
		return fmt.Errorf("bridge is already paused")
	}

	params.EmergencyPaused = true
	k.SetParams(ctx, params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_paused",
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("paused_by", pauser),
			sdk.NewAttribute("timestamp", ctx.BlockTime().Format(time.RFC3339)),
		),
	)

	return nil
}

// UnpauseBridge resumes bridge operations
func (k Keeper) UnpauseBridge(ctx sdk.Context, unpauser string) error {
	params := k.GetParams(ctx)

	if !params.EmergencyPaused {
		return fmt.Errorf("bridge is not paused")
	}

	params.EmergencyPaused = false
	k.SetParams(ctx, params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_unpaused",
			sdk.NewAttribute("unpaused_by", unpauser),
			sdk.NewAttribute("timestamp", ctx.BlockTime().Format(time.RFC3339)),
		),
	)

	return nil
}

// IsBridgePaused checks if the bridge is currently paused
func (k Keeper) IsBridgePaused(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	return params.EmergencyPaused
}

// ============================================================================
// WHITELIST / BLACKLIST
// ============================================================================

// AddToWhitelist adds an address to the whitelist
func (k Keeper) AddToWhitelist(
	ctx sdk.Context,
	address string,
	reason string,
	addedBy string,
) error {
	permission := &types.AddressPermission{
		Address:        address,
		PermissionType: types.PermissionType_PERMISSION_WHITELISTED,
		Reason:         reason,
		AddedBy:        addedBy,
		AddedAt:        ctx.BlockTime(),
		ExpiresAt:      time.Time{}, // Never expires
	}

	k.SetAddressPermission(ctx, permission)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"address_whitelisted",
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("added_by", addedBy),
		),
	)

	return nil
}

// AddToBlacklist adds an address to the blacklist
func (k Keeper) AddToBlacklist(
	ctx sdk.Context,
	address string,
	reason string,
	addedBy string,
	expiresAt time.Time,
) error {
	permission := &types.AddressPermission{
		Address:        address,
		PermissionType: types.PermissionType_PERMISSION_BLACKLISTED,
		Reason:         reason,
		AddedBy:        addedBy,
		AddedAt:        ctx.BlockTime(),
		ExpiresAt:      expiresAt,
	}

	k.SetAddressPermission(ctx, permission)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"address_blacklisted",
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("added_by", addedBy),
		),
	)

	return nil
}

// RemoveFromPermissionList removes an address from whitelist/blacklist
func (k Keeper) RemoveFromPermissionList(ctx sdk.Context, address string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.AddressPermissionKey(address)
	store.Delete(key)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"address_permission_removed",
			sdk.NewAttribute("address", address),
		),
	)

	return nil
}

// CheckAddressPermission checks if an address is allowed to use the bridge
func (k Keeper) CheckAddressPermission(ctx sdk.Context, address string) error {
	params := k.GetParams(ctx)

	permission := k.GetAddressPermission(ctx, address)
	if permission == nil {
		// No specific permission set
		if params.WhitelistEnabled {
			return fmt.Errorf("address %s is not whitelisted", address)
		}
		return nil
	}

	// Check if blacklisted
	if permission.PermissionType == types.PermissionType_PERMISSION_BLACKLISTED {
		// Check expiry
		if !permission.ExpiresAt.IsZero() && ctx.BlockTime().After(permission.ExpiresAt) {
			// Blacklist expired, remove it
			k.RemoveFromPermissionList(ctx, address)
			return nil
		}
		return fmt.Errorf("address %s is blacklisted: %s", address, permission.Reason)
	}

	// If whitelisted, always allow
	if permission.PermissionType == types.PermissionType_PERMISSION_WHITELISTED {
		return nil
	}

	// If whitelist is enabled but address is not whitelisted
	if params.WhitelistEnabled {
		return fmt.Errorf("address %s is not whitelisted", address)
	}

	return nil
}

// CleanupExpiredPermissions removes expired blacklist entries
func (k Keeper) CleanupExpiredPermissions(ctx sdk.Context) {
	permissions := k.GetAllAddressPermissions(ctx)

	for _, permission := range permissions {
		if permission.PermissionType == types.PermissionType_PERMISSION_BLACKLISTED &&
			!permission.ExpiresAt.IsZero() &&
			ctx.BlockTime().After(permission.ExpiresAt) {

			k.RemoveFromPermissionList(ctx, permission.Address)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"blacklist_expired",
					sdk.NewAttribute("address", permission.Address),
				),
			)
		}
	}
}

// ============================================================================
// STORAGE
// ============================================================================

// NonceTracker storage
func (k Keeper) GetNonceTracker(ctx sdk.Context, address string, chainId string) *types.NonceTracker {
	store := ctx.KVStore(k.storeKey)
	key := types.NonceTrackerKey(address, chainId)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var tracker types.NonceTracker
	k.cdc.MustUnmarshal(bz, &tracker)
	return &tracker
}

func (k Keeper) SetNonceTracker(ctx sdk.Context, tracker *types.NonceTracker) {
	store := ctx.KVStore(k.storeKey)
	key := types.NonceTrackerKey(tracker.Address, tracker.ChainId)

	bz := k.cdc.MustMarshal(tracker)
	store.Set(key, bz)
}

// AddressPermission storage
func (k Keeper) GetAddressPermission(ctx sdk.Context, address string) *types.AddressPermission {
	store := ctx.KVStore(k.storeKey)
	key := types.AddressPermissionKey(address)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var permission types.AddressPermission
	k.cdc.MustUnmarshal(bz, &permission)
	return &permission
}

func (k Keeper) SetAddressPermission(ctx sdk.Context, permission *types.AddressPermission) {
	store := ctx.KVStore(k.storeKey)
	key := types.AddressPermissionKey(permission.Address)

	bz := k.cdc.MustMarshal(permission)
	store.Set(key, bz)
}

func (k Keeper) GetAllAddressPermissions(ctx sdk.Context) []*types.AddressPermission {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AddressPermissionPrefix)
	defer iterator.Close()

	permissions := []*types.AddressPermission{}
	for ; iterator.Valid(); iterator.Next() {
		var permission types.AddressPermission
		k.cdc.MustUnmarshal(iterator.Value(), &permission)
		permissions = append(permissions, &permission)
	}

	return permissions
}
