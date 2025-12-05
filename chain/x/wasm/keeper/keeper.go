package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper wraps the wasmd keeper with AURA-specific security controls
type Keeper struct {
	cdc              codec.BinaryCodec
	storeKey         storetypes.StoreKey
	wasmKeeper       *wasmkeeper.Keeper
	authority        string // Address of the governance authority
	contractRegistry *contractregistrykeeper.Keeper // Contract registry for policy enforcement
}

// NewKeeper creates a new wasm Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	wasmKeeper *wasmkeeper.Keeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		wasmKeeper: wasmKeeper,
		authority:  authority,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetWasmKeeper returns the underlying wasmd keeper
func (k Keeper) GetWasmKeeper() *wasmkeeper.Keeper {
	return k.wasmKeeper
}

// GetStoreKey returns the store key
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// SetContractRegistry sets the contract registry keeper
func (k *Keeper) SetContractRegistry(registry *contractregistrykeeper.Keeper) {
	k.contractRegistry = registry
}

// InitGenesis initializes the module's state from a genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	// First validate the genesis state
	if err := types.ValidateGenesis(&data); err != nil {
		return err
	}

	// Set params if provided
	if data.Params != nil {
		if err := k.SetParams(ctx, *data.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Set authorized uploaders
	for _, uploader := range data.AuthorizedUploaders {
		if err := k.AuthorizeUploader(ctx, uploader); err != nil {
			return fmt.Errorf("failed to authorize uploader %s: %w", uploader, err)
		}
	}

	// Set paused contracts
	for _, contract := range data.PausedContracts {
		if err := k.PauseContract(ctx, contract); err != nil {
			return fmt.Errorf("failed to pause contract %s: %w", contract, err)
		}
	}

	// Set security stats if provided
	if data.SecurityStats != nil {
		k.SetSecurityStats(ctx, *data.SecurityStats)
	}

	// Only initialize wasmd keeper if configured (for full integration)
	// In unit tests without wasmd, we still want params/state to work
	if k.wasmKeeper != nil {
		wasmGenesis := wasmtypes.GenesisState{
			Params: wasmtypes.DefaultParams(),
		}
		_, err := wasmkeeper.InitGenesis(ctx, k.wasmKeeper, wasmGenesis)
		if err != nil {
			return fmt.Errorf("failed to init wasmd genesis: %w", err)
		}
	}

	return nil
}

// ExportGenesis exports the module's state to a genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Get current params
	params := k.GetParams(ctx)

	// Get authorized uploaders by iterating over the prefix
	var authorizedUploaders []string
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.AuthorizedUploaderPrefix, storetypes.PrefixEndBytes(types.AuthorizedUploaderPrefix))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		// Key is prefix + address string, extract the address
		key := iterator.Key()
		address := string(key[len(types.AuthorizedUploaderPrefix):])
		authorizedUploaders = append(authorizedUploaders, address)
	}

	// Get paused contracts by iterating over the prefix
	var pausedContracts []string
	pauseIterator := store.Iterator(types.PausedContractPrefix, storetypes.PrefixEndBytes(types.PausedContractPrefix))
	defer pauseIterator.Close()
	for ; pauseIterator.Valid(); pauseIterator.Next() {
		key := pauseIterator.Key()
		contractAddr := string(key[len(types.PausedContractPrefix):])
		pausedContracts = append(pausedContracts, contractAddr)
	}

	// Get security stats
	stats := k.GetSecurityStats(ctx)

	return types.NewGenesisState(
		&params,
		[]*types.Code{},
		[]*types.Contract{},
		[]*types.Sequence{},
		authorizedUploaders,
		pausedContracts,
		&stats,
	)
}

// ============================================================================
// CONTRACT ADMIN MANAGEMENT
// ============================================================================

// SetContractAdmin sets the admin for a contract
// This provides AURA-specific admin tracking in addition to wasmd's admin storage
func (k Keeper) SetContractAdmin(ctx sdk.Context, contractAddr sdk.AccAddress, admin sdk.AccAddress) error {
	if contractAddr.Empty() {
		return types.ErrInvalidContractAddress.Wrap("contract address cannot be empty")
	}
	if admin.Empty() {
		return types.ErrInvalidAdmin.Wrap("admin address cannot be empty")
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetContractAdminKey(contractAddr.String())
	store.Set(key, []byte(admin.String()))

	k.Logger(ctx).Debug("contract admin set",
		"contract", contractAddr.String(),
		"admin", admin.String())

	return nil
}

// GetContractAdmin retrieves the admin for a contract
// Returns empty address if no admin is set
func (k Keeper) GetContractAdmin(ctx sdk.Context, contractAddr sdk.AccAddress) (sdk.AccAddress, error) {
	if contractAddr.Empty() {
		return nil, types.ErrInvalidContractAddress.Wrap("contract address cannot be empty")
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetContractAdminKey(contractAddr.String())
	bz := store.Get(key)

	if bz == nil {
		// No admin set - check wasmd keeper as fallback
		if k.wasmKeeper != nil {
			info := k.wasmKeeper.GetContractInfo(ctx, contractAddr)
			if info != nil {
				adminAddr, err := sdk.AccAddressFromBech32(info.Admin)
				if err == nil && !adminAddr.Empty() {
					return adminAddr, nil
				}
			}
		}
		return sdk.AccAddress{}, nil
	}

	adminAddr, err := sdk.AccAddressFromBech32(string(bz))
	if err != nil {
		return nil, types.ErrInvalidAdmin.Wrapf("stored admin address is invalid: %s", err)
	}

	return adminAddr, nil
}

// DeleteContractAdmin removes the admin for a contract
func (k Keeper) DeleteContractAdmin(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	if contractAddr.Empty() {
		return types.ErrInvalidContractAddress.Wrap("contract address cannot be empty")
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetContractAdminKey(contractAddr.String())
	store.Delete(key)

	k.Logger(ctx).Debug("contract admin deleted",
		"contract", contractAddr.String())

	return nil
}

// HasContractAdmin checks if a contract has an admin set
func (k Keeper) HasContractAdmin(ctx sdk.Context, contractAddr sdk.AccAddress) bool {
	admin, err := k.GetContractAdmin(ctx, contractAddr)
	if err != nil {
		return false
	}
	return !admin.Empty()
}

// IsContractAdmin checks if the given address is the admin of the contract
func (k Keeper) IsContractAdmin(ctx sdk.Context, contractAddr, candidate sdk.AccAddress) (bool, error) {
	admin, err := k.GetContractAdmin(ctx, contractAddr)
	if err != nil {
		return false, err
	}
	if admin.Empty() {
		return false, nil
	}
	return admin.Equals(candidate), nil
}
