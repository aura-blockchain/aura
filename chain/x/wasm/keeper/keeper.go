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
	if k.wasmKeeper == nil {
		return fmt.Errorf("wasm keeper not configured")
	}

	wasmGenesis := wasmtypes.GenesisState{
		Params: wasmtypes.DefaultParams(),
	}
	_, err := wasmkeeper.InitGenesis(ctx, k.wasmKeeper, wasmGenesis)
	return err
}

// ExportGenesis exports the module's state to a genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	if k.wasmKeeper == nil {
		return types.NewGenesisState(
			types.DefaultParams(),
			[]*types.Code{},
			[]*types.Contract{},
			[]*types.Sequence{},
			[]string{},
			[]string{},
			types.DefaultSecurityStats(),
		)
	}

	_ = wasmkeeper.ExportGenesis(ctx, k.wasmKeeper) // rely on defaults for Aura module export
	return types.DefaultGenesisState()
}
