package compliance

import (
	"github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

// AppModule represents the compliance module
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new compliance module
func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// Name returns the module name
func (am AppModule) Name() string {
	return "compliance"
}

// GetKeeper returns the module keeper
func (am AppModule) GetKeeper() *keeper.Keeper {
	return am.keeper
}

// DefaultGenesis returns default genesis state
func (am AppModule) DefaultGenesis() types.ComplianceParams {
	return types.DefaultParams()
}

// ValidateGenesis validates genesis state
func (am AppModule) ValidateGenesis(params types.ComplianceParams) error {
	return params.Validate()
}
