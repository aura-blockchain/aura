package aiassistant

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/chain/x/aiassistant/client/cli"
	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// ModuleServices describes the registration hooks exposed by the app module manager.
type ModuleServices interface {
	RegisterMsgServer(types.MsgServer)
	RegisterQueryServer(types.QueryServer)
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

func (m AppModule) Name() string { return types.ModuleName }

func (m AppModule) RegisterServices(ms ModuleServices) {
	if ms == nil {
		panic(fmt.Sprintf("%s: nil module services", types.ModuleName))
	}
	ms.RegisterMsgServer(keeper.NewMsgServer(m.keeper))
	ms.RegisterQueryServer(keeper.NewQueryServer(m.keeper))
}

func (m AppModule) BeginBlock(context.Context) {}
func (m AppModule) EndBlock(context.Context)   {}

func (m AppModule) InitGenesis(ctx sdk.Context, state types.GenesisState) error {
	return m.keeper.InitGenesis(ctx, state)
}

func (m AppModule) ExportGenesis(ctx sdk.Context) types.GenesisState {
	return m.keeper.ExportGenesis(ctx)
}

func (m AppModule) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

func (m AppModule) GetQueryCmd() *cobra.Command {
	return cli.NewQueryCmd()
}

// RegisterInvariants registers the aiassistant module invariants
func (m AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, *m.keeper)
}
