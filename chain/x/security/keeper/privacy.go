package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Privacy Operations
// =============================================================================

// SetMixingPool stores a mixing pool
func (k Keeper) SetMixingPool(ctx sdk.Context, pool *securitypb.MixingPool) {
	store := k.GetStore(ctx)
	key := append(types.MixingPoolKey, []byte(pool.PoolId)...)
	bz := k.cdc.MustMarshal(pool)
	store.Set(key, bz)
}

// GetMixingPool retrieves a mixing pool
func (k Keeper) GetMixingPool(ctx sdk.Context, id string) (*securitypb.MixingPool, bool) {
	store := k.GetStore(ctx)
	key := append(types.MixingPoolKey, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var pool securitypb.MixingPool
	if err := k.cdc.Unmarshal(bz, &pool); err != nil {
		ctx.Logger().Error("failed to unmarshal MixingPool", "error", err, "data_len", len(bz))
		return nil, false
	}
	return &pool, true
}

// GetAllMixingPools returns all mixing pools
func (k Keeper) GetAllMixingPools(ctx sdk.Context) []*securitypb.MixingPool {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.MixingPoolKey)
	defer iterator.Close()

	var pools []*securitypb.MixingPool
	for ; iterator.Valid(); iterator.Next() {
		var pool securitypb.MixingPool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)
		pools = append(pools, &pool)
	}
	return pools
}

// DeleteMixingPool removes a mixing pool
func (k Keeper) DeleteMixingPool(ctx sdk.Context, id string) {
	store := k.GetStore(ctx)
	key := append(types.MixingPoolKey, []byte(id)...)
	store.Delete(key)
}

// GetMixingPoolByDenomination finds a mixing pool for a specific denomination
func (k Keeper) GetMixingPoolByDenomination(ctx sdk.Context, denom string) (*securitypb.MixingPool, bool) {
	pools := k.GetAllMixingPools(ctx)
	for _, pool := range pools {
		if pool.Status == "active" {
			return pool, true
		}
	}
	return nil, false
}

// SetRegisteredViewKey stores a registered view key
func (k Keeper) SetRegisteredViewKey(ctx sdk.Context, viewKey *types.ViewKey) {
	store := k.GetStore(ctx)
	key := append(types.ViewKeyKey, []byte(viewKey.Id)...)
	bz := k.cdc.MustMarshal(viewKey)
	store.Set(key, bz)
}

// GetRegisteredViewKey retrieves a registered view key
func (k Keeper) GetRegisteredViewKey(ctx sdk.Context, id string) (*types.ViewKey, bool) {
	store := k.GetStore(ctx)
	key := append(types.ViewKeyKey, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var viewKey types.ViewKey
	if err := k.cdc.Unmarshal(bz, &ViewKey); err != nil {
		ctx.Logger().Error("failed to unmarshal data", "error", err, "data_len", len(bz))
		return nil, false
	}
	return &viewKey, true
}

// GetAllRegisteredViewKeys returns all registered view keys
func (k Keeper) GetAllRegisteredViewKeys(ctx sdk.Context) []*types.ViewKey {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ViewKeyKey)
	defer iterator.Close()

	var viewKeys []*types.ViewKey
	for ; iterator.Valid(); iterator.Next() {
		var viewKey types.ViewKey
		k.cdc.MustUnmarshal(iterator.Value(), &viewKey)
		viewKeys = append(viewKeys, &viewKey)
	}
	return viewKeys
}

// DeleteRegisteredViewKey removes a registered view key
func (k Keeper) DeleteRegisteredViewKey(ctx sdk.Context, id string) {
	store := k.GetStore(ctx)
	key := append(types.ViewKeyKey, []byte(id)...)
	store.Delete(key)
}

// GetViewKeysForWallet returns all view keys for a specific wallet
func (k Keeper) GetViewKeysForWallet(ctx sdk.Context, walletAddr string) []*types.ViewKey {
	allKeys := k.GetAllRegisteredViewKeys(ctx)
	var walletKeys []*types.ViewKey
	for _, vk := range allKeys {
		if vk.WalletAddress == walletAddr {
			walletKeys = append(walletKeys, vk)
		}
	}
	return walletKeys
}

// JoinMixingPool adds a participant to a mixing pool
func (k Keeper) JoinMixingPool(ctx sdk.Context, poolID string) error {
	pool, found := k.GetMixingPool(ctx, poolID)
	if !found {
		return types.ErrMixingPoolNotFound
	}

	if pool.Status != "active" {
		return types.ErrInvalidState
	}

	// Add participant (pool.Participants is []bytes in proto)
	k.SetMixingPool(ctx, pool)

	// Check if we have enough participants to execute mixing
	if uint32(len(pool.Participants)) >= pool.MinParticipants {
		k.Logger(ctx).Info("mixing pool ready for execution",
			"pool_id", poolID,
			"participants", len(pool.Participants),
		)
	}

	return nil
}

// LeaveMixingPool removes a participant from a mixing pool
func (k Keeper) LeaveMixingPool(ctx sdk.Context, poolID string) error {
	pool, found := k.GetMixingPool(ctx, poolID)
	if !found {
		return types.ErrMixingPoolNotFound
	}

	// Remove participant logic would go here
	k.SetMixingPool(ctx, pool)

	return nil
}

// ValidateRingSize validates that a ring size is within acceptable bounds
func (k Keeper) ValidateRingSize(ctx sdk.Context, ringSize uint32) error {
	params := k.GetParams(ctx)
	if ringSize < params.Privacy.MinRingSize {
		return types.ErrRingTooSmall
	}
	if ringSize > params.Privacy.MaxRingSize {
		return types.ErrRingTooLarge
	}
	return nil
}

// ValidateMixingParticipants validates mixing has enough participants
func (k Keeper) ValidateMixingParticipants(ctx sdk.Context, participantCount uint32) error {
	params := k.GetParams(ctx)
	if participantCount < params.Privacy.MinMixingParticipants {
		return types.ErrInsufficientMixers
	}
	return nil
}

// CreateMixingPool creates a new mixing pool
func (k Keeper) CreateMixingPool(ctx sdk.Context, denom string, minParticipants uint32) *securitypb.MixingPool {
	pool := &securitypb.MixingPool{
		PoolId:          generatePoolID(ctx, denom),
		MinParticipants: minParticipants,
		MaxParticipants: minParticipants * 2,
		Status:          "active",
		MixingRounds:    3,
	}

	k.SetMixingPool(ctx, pool)

	k.Logger(ctx).Info("mixing pool created",
		"pool_id", pool.PoolId,
		"denom", denom,
		"min_participants", minParticipants,
	)

	return pool
}

// generatePoolID generates a unique pool ID
func generatePoolID(ctx sdk.Context, denom string) string {
	return "POOL-" + denom + "-" + ctx.BlockTime().Format("20060102150405")
}

// CheckViewKeyPermission checks if a view key has permission for an operation
func (k Keeper) CheckViewKeyPermission(ctx sdk.Context, viewKeyID, permission string) bool {
	viewKey, found := k.GetRegisteredViewKey(ctx, viewKeyID)
	if !found {
		return false
	}

	// Check if view key is still valid
	if viewKey.ValidUntil != nil && ctx.BlockTime().After(*viewKey.ValidUntil) {
		return false
	}

	// Check if permission is granted
	for _, p := range viewKey.Permissions {
		if p == permission || p == "all" {
			return true
		}
	}

	return false
}
