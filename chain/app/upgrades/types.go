package upgrades

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// BaseAppParamManager defines an interface that BaseApp is expected to fulfill
// that allows upgrade handlers to modify BaseApp parameters.
type BaseAppParamManager interface {
	GetConsensusParams(ctx sdk.Context) (*cmtproto.ConsensusParams, error)
	StoreConsensusParams(ctx sdk.Context, cp *cmtproto.ConsensusParams) error
}

// Upgrade defines a struct containing necessary fields that an upgrade needs
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. "v2"
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler
	CreateUpgradeHandler func(*module.Manager, module.Configurator) upgradetypes.UpgradeHandler

	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades storetypes.StoreUpgrades
}

// UpgradeHandler is a function that performs chain upgrades
type UpgradeHandler func(
	ctx context.Context,
	plan upgradetypes.Plan,
	fromVM module.VersionMap,
) (module.VersionMap, error)
