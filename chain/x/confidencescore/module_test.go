package confidencescore

import (
	"testing"

	"github.com/aequitas/aura/chain/x/confidencescore/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/params"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestAppModule_Name(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()), "aura10d07y265gmmuvt4z0w9aw880jnsr700jqhc7kh")
	module := NewAppModule(k)

	if module.Name() != types.ModuleName {
		t.Errorf("expected module name %s, got %s", types.ModuleName, module.Name())
	}
}

func TestAppModule_InitExportGenesis(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()), "aura10d07y265gmmuvt4z0w9aw880jnsr700jqhc7kh")
	module := NewAppModule(k)

	// Create test genesis state
	defaultParams := types.DefaultParams()
	genesis := types.GenesisState{
		Params:       &defaultParams,
		UserRecords:  []*types.UserConfidenceRecord{},
		SlashRecords: []*types.SlashRecord{},
	}

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
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()), "aura10d07y265gmmuvt4z0w9aw880jnsr700jqhc7kh")
	module := NewAppModule(k)

	// Should not panic
	module.BeginBlock()
}

func TestAppModule_EndBlock(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(types.DefaultParams()), "aura10d07y265gmmuvt4z0w9aw880jnsr700jqhc7kh")
	module := NewAppModule(k)

	// Should not panic - EndBlock currently does nothing
	module.EndBlock()

	// Verify it can be called multiple times
	module.EndBlock()
	module.EndBlock()
}

func TestAppModuleBasic_Name(t *testing.T) {
	basic := AppModuleBasic{}

	if basic.Name() != types.ModuleName {
		t.Errorf("expected module name %s, got %s", types.ModuleName, basic.Name())
	}
}
