package compliance

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

// ModuleServices defines the interface for registering module services.
type ModuleServices interface {
	RegisterMsgServer(compliancepb.MsgServer)
	RegisterQueryServer(compliancepb.QueryServer)
}

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

// RegisterServices registers the module's gRPC Msg/Query servers.
func (am AppModule) RegisterServices(config ModuleServices) {
	if config == nil {
		panic(fmt.Sprintf("%s: nil module services", am.Name()))
	}
	config.RegisterMsgServer(keeper.NewMsgServer(am.keeper))
	config.RegisterQueryServer(keeper.NewQueryServer(am.keeper))
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
	return types.ValidateParams(params)
}
