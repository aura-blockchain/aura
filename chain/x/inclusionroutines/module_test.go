package inclusionroutines

import (
	"testing"

	"github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

func TestAppModuleBasic(t *testing.T) {
	basic := AppModuleBasic{}

	if basic.Name() != types.ModuleName {
		t.Errorf("expected module name %s, got %s", types.ModuleName, basic.Name())
	}
}

func TestNewAppModule(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(store, "authority")
	module := NewAppModule(k)

	if module.Name() != types.ModuleName {
		t.Errorf("expected module name %s, got %s", types.ModuleName, module.Name())
	}

	if module.keeper == nil {
		t.Error("expected keeper to be non-nil")
	}
}

func TestRegisterServicesPanic(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(store, "authority")
	module := NewAppModule(k)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when RegisterServices called with nil config")
		}
	}()

	module.RegisterServices(nil)
}

func TestBeginBlock(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(store, "authority")
	module := NewAppModule(k)

	// Should not panic
	module.BeginBlock()
}

func TestEndBlock(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := keeper.NewKeeper(store, "authority")
	module := NewAppModule(k)

	// Should not panic
	module.EndBlock()
}
