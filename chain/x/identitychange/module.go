package identitychange

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

type ModuleServices interface {
	RegisterMsgServer(identitychangepb.MsgServer)
	RegisterQueryServer(identitychangepb.QueryServer)
}

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterServices(ModuleServices) {}

type AppModule struct {
	keeper *keeper.Keeper
}

func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

func (AppModule) Name() string { return types.ModuleName }

func (m AppModule) RegisterServices(config ModuleServices) {
	if config == nil {
		panic(fmt.Sprintf("%s: nil module services", types.ModuleName))
	}
	config.RegisterMsgServer(keeper.NewMsgServer(m.keeper))
	config.RegisterQueryServer(keeper.NewQueryServer(m.keeper))
}

func (m AppModule) BeginBlock() {}

func (m AppModule) EndBlock() {}

// RegisterInvariants registers the module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// TODO: Uncomment when invariants.go is fixed
	// keeper.RegisterInvariants(ir, am.keeper)
}
