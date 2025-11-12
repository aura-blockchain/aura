package identitychange_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

type testServices struct {
	receivedMsg   identitychange.MsgServer
	receivedQuery identitychange.QueryServer
}

func (s *testServices) RegisterMsgServer(server identitychange.MsgServer) {
	s.receivedMsg = server
}

func (s *testServices) RegisterQueryServer(server identitychange.QueryServer) {
	s.receivedQuery = server
}

func TestAppModuleRegisterServices(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(store)
	module := identitychange.NewAppModule(k)
	services := &testServices{}
	module.RegisterServices(services)
	if services.receivedMsg == nil || services.receivedQuery == nil {
		t.Fatalf("services not registered")
	}
}
