package identitychange_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

type testServices struct {
	receivedMsg   identitychangepb.MsgServer
	receivedQuery identitychangepb.QueryServer
}

func (s *testServices) RegisterMsgServer(server identitychangepb.MsgServer) {
	s.receivedMsg = server
}

func (s *testServices) RegisterQueryServer(server identitychangepb.QueryServer) {
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
