package confidencescore

import (
	"testing"

	"github.com/aequitas/aura/chain/x/confidencescore/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/params"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestAppModule_Name(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()))
	module := NewAppModule(k)

	if module.Name() != types.ModuleName {
		t.Errorf("expected module name %s, got %s", types.ModuleName, module.Name())
	}
}

func TestAppModule_InitExportGenesis(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()))
	module := NewAppModule(k)

	genesis := types.TestGenesisState()

	// Init genesis
	if err := module.InitGenesis(genesis); err != nil {
		t.Fatalf("failed to init genesis: %v", err)
	}

	// Export genesis
	exported := module.ExportGenesis()

	if len(exported.UserRecords) != len(genesis.UserRecords) {
		t.Errorf("expected %d user records, got %d",
			len(genesis.UserRecords), len(exported.UserRecords))
	}
}

func TestAppModule_BeginBlock(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()))
	module := NewAppModule(k)

	// Should not panic
	module.BeginBlock()
}

func TestAppModule_EndBlock(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()))
	module := NewAppModule(k)

	// Should not panic
	module.EndBlock()
}

func TestAppModuleBasic_Name(t *testing.T) {
	basic := AppModuleBasic{}

	if basic.Name() != types.ModuleName {
		t.Errorf("expected module name %s, got %s", types.ModuleName, basic.Name())
	}
}
