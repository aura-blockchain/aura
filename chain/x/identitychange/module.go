package identitychange

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

type ModuleServices interface {
	RegisterMsgServer(MsgServer)
	RegisterQueryServer(QueryServer)
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
	config.RegisterMsgServer(NewMsgServer(m.keeper))
	config.RegisterQueryServer(NewQueryServer(m.keeper))
}

func (m AppModule) BeginBlock() {}

func (m AppModule) EndBlock() {}
