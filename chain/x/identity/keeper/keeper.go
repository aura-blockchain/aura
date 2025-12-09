package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// Keeper manages the identity module state
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	authority    string
	logger       log.Logger
}

// NewKeeper creates a new identity keeper
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	authority string,
	logger log.Logger,
) *Keeper {
	return &Keeper{
		storeService: storeService,
		cdc:          cdc,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the module authority address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns the module logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ============================================================================
// Genesis Methods
// ============================================================================

// Genesis methods are implemented in genesis.go

// ============================================================================
// Params Methods
// ============================================================================

// GetParams retrieves the module parameters
func (k *Keeper) GetParams(ctx sdk.Context) (*types.Params, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}

	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return nil, err
	}
	return &params, nil
}

// SetParams sets the module parameters
func (k *Keeper) SetParams(ctx sdk.Context, params *types.Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}

// ============================================================================
// Helper Methods
// ============================================================================

// initializeDefaultRoles creates the default system roles
func (k *Keeper) initializeDefaultRoles(ctx sdk.Context) error {
	now := ctx.BlockTime()

	// Admin role with all permissions
	adminRole := &types.Role{
		Name: types.RoleAdmin,
		Permissions: []string{
			types.PermissionAdmin,
			types.PermissionCreateRole,
			types.PermissionAssignRole,
			types.PermissionRevokeRole,
			types.PermissionManageMultisig,
			types.PermissionManageTimeLock,
			types.PermissionManageEmergency,
			types.PermissionRotateValidatorKey,
			types.PermissionManageSession,
			types.PermissionViewAuditLogs,
			types.PermissionManageIdentity,
			types.PermissionVerifyIdentity,
			types.PermissionApproveChangeRequest,
		},
		Description: "Full administrative access",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	if err := k.SetRole(ctx, adminRole); err != nil {
		return err
	}

	// Moderator role
	moderatorRole := &types.Role{
		Name: types.RoleModerator,
		Permissions: []string{
			types.PermissionAssignRole,
			types.PermissionRevokeRole,
			types.PermissionManageSession,
			types.PermissionViewAuditLogs,
			types.PermissionVerifyIdentity,
		},
		Description: "Moderate user permissions and verify identities",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	if err := k.SetRole(ctx, moderatorRole); err != nil {
		return err
	}

	// Validator role
	validatorRole := &types.Role{
		Name: types.RoleValidator,
		Permissions: []string{
			types.PermissionRotateValidatorKey,
		},
		Description: "Validator-specific permissions",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	if err := k.SetRole(ctx, validatorRole); err != nil {
		return err
	}

	// User role (basic)
	userRole := &types.Role{
		Name:        types.RoleUser,
		Permissions: []string{},
		Description: "Basic user permissions",
		CreatedAt: timestamppb.New(now),
		CreatedBy:   "",
		IsSystemRole: true,
		UpdatedAt: timestamppb.New(now),
	}
	return k.SetRole(ctx, userRole)
}

// GetAllDataForPrefix retrieves all items with a given prefix
func (k *Keeper) GetAllDataForPrefix(ctx sdk.Context, prefix []byte) ([][]byte, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var items [][]byte
	for ; iterator.Valid(); iterator.Next() {
		items = append(items, iterator.Value())
	}
	return items, nil
}
