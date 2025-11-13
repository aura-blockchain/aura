package app

import (
	"testing"

	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type fakeRegistrar struct {
	services []string
}

func (f *fakeRegistrar) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	f.services = append(f.services, desc.ServiceName)
}

func TestRegisterGRPCServices(t *testing.T) {
	k := keeper.NewKeeper(nil)
	manager := NewModuleManager(identitychange.NewAppModule(k))

	registrar := &fakeRegistrar{}
	manager.RegisterGRPCServices(registrar)

	require.Condition(t, func() bool {
		for _, svc := range registrar.services {
			if svc == "aura.identitychange.v1beta1.Msg" {
				return true
			}
		}
		return false
	}, "Msg service not registered")

	require.Condition(t, func() bool {
		for _, svc := range registrar.services {
			if svc == "aura.identitychange.v1beta1.Query" {
				return true
			}
		}
		return false
	}, "Query service not registered")
}
