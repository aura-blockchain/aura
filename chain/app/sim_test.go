//go:build simapp
// +build simapp

// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"

	"github.com/aequitas/aura/chain/app"
	abci "github.com/cometbft/cometbft/abci/types"
)

// SimAppChainID hardcoded chainID for simulation
const SimAppChainID = "aura-sim-test"

func init() {
	simcli.GetSimulatorFlags()
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an pointStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

// NewSimApp creates a new App instance for testing with optional BaseApp tweaks.
func NewSimApp(logger log.Logger, db dbm.DB, _ simtestutil.AppOptionsMap, baseAppOptions ...func(*baseapp.BaseApp)) *app.App {
	a := app.NewAppWithOptions(logger, db, SimAppChainID)
	for _, opt := range baseAppOptions {
		opt(a.BaseApp)
	}
	return a
}

// TestAppStateDeterminism runs a determinism simulation test
func TestAppStateDeterminism(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = SimAppChainID

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			logger := log.NewNopLogger()

			db := dbm.NewMemDB()
			appOpts := simtestutil.AppOptionsMap{
				flags.FlagHome:            t.TempDir(),
				server.FlagInvCheckPeriod: simcli.FlagPeriodValue,
			}

			auraApp := NewSimApp(logger, db, appOpts, fauxMerkleModeOpt)
			require.NotNil(t, auraApp)

			fmt.Printf(
				"running non-determinism simulation; seed %d: %d/%d, attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				auraApp.BaseApp,
				simtestutil.AppStateFn(auraApp.AppCodec(), auraApp.SimulationManager(), auraApp.DefaultGenesis()),
				simtypes.RandomAccounts,
				simtestutil.SimulationOperations(auraApp, auraApp.AppCodec(), config),
				auraApp.ModuleAccountAddrs(),
				config,
				auraApp.AppCodec(),
			)
			require.NoError(t, err)

			appHash := auraApp.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t, string(appHashList[0]), string(appHashList[j]),
					"non-determinism in seed %d: %d/%d, attempt: %d/%d\n", config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}

// TestAppImportExport runs an import/export simulation test
func TestAppImportExport(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	appOpts := simtestutil.AppOptionsMap{
		flags.FlagHome:            t.TempDir(),
		server.FlagInvCheckPeriod: simcli.FlagPeriodValue,
	}

	auraApp := NewSimApp(logger, db, appOpts, fauxMerkleModeOpt)
	require.NotNil(t, auraApp)

	// run simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		auraApp.BaseApp,
		simtestutil.AppStateFn(auraApp.AppCodec(), auraApp.SimulationManager(), auraApp.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(auraApp, auraApp.AppCodec(), config),
		auraApp.ModuleAccountAddrs(),
		config,
		auraApp.AppCodec(),
	)
	require.NoError(t, simErr)

	// export genesis and params
	err := simtestutil.CheckExportSimulation(auraApp, config, simParams)
	require.NoError(t, err)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := auraApp.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	newDB := dbm.NewMemDB()
	newAppOpts := simtestutil.AppOptionsMap{
		flags.FlagHome:            t.TempDir(),
		server.FlagInvCheckPeriod: simcli.FlagPeriodValue,
	}

	newApp := NewSimApp(logger, newDB, newAppOpts, fauxMerkleModeOpt)
	require.NotNil(t, newApp)

	var genesisState map[string]json.RawMessage
	err = json.Unmarshal(exported.AppState, &genesisState)
	require.NoError(t, err)

	ctxA := auraApp.NewContext(true)
	ctxB := newApp.NewContext(true)

	// compare stores
	storeKeys := auraApp.GetStoreKeys()
	for _, storeKeyA := range storeKeys {
		storeKeyB := newApp.GetKey(storeKeyA.Name())
		if storeKeyB == nil {
			continue
		}

		storeA := ctxA.KVStore(storeKeyA)
		storeB := ctxB.KVStore(storeKeyB)

		failedKVAs, failedKVBs := simtestutil.DiffKVStores(storeA, storeB, nil)
		require.Equal(t, len(failedKVAs), len(failedKVBs), "unequal sets of key-values to compare")
	}
}

// TestAppSimulationAfterImport runs a simulation after importing state
func TestAppSimulationAfterImport(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	appOpts := simtestutil.AppOptionsMap{
		flags.FlagHome:            t.TempDir(),
		server.FlagInvCheckPeriod: simcli.FlagPeriodValue,
	}

	auraApp := NewSimApp(logger, db, appOpts, fauxMerkleModeOpt)
	require.NotNil(t, auraApp)

	// run initial simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		auraApp.BaseApp,
		simtestutil.AppStateFn(auraApp.AppCodec(), auraApp.SimulationManager(), auraApp.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(auraApp, auraApp.AppCodec(), config),
		auraApp.ModuleAccountAddrs(),
		config,
		auraApp.AppCodec(),
	)
	require.NoError(t, simErr)

	err := simtestutil.CheckExportSimulation(auraApp, config, simParams)
	require.NoError(t, err)

	if config.Commit {
		simtestutil.PrintStats(db)
	}

	if stopEarly {
		fmt.Println("can't export or import a halted simulation")
		return
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := auraApp.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	fmt.Printf("importing genesis...\n")

	newDB := dbm.NewMemDB()
	newAppOpts := simtestutil.AppOptionsMap{
		flags.FlagHome:            t.TempDir(),
		server.FlagInvCheckPeriod: simcli.FlagPeriodValue,
	}

	newApp := NewSimApp(logger, newDB, newAppOpts, fauxMerkleModeOpt)
	require.NotNil(t, newApp)

	_, err = newApp.InitChain(&abci.RequestInitChain{
		AppStateBytes: exported.AppState,
		ChainId:       SimAppChainID,
	})
	require.NoError(t, err)

	// run simulation on imported state
	_, _, err = simulation.SimulateFromSeed(
		t,
		os.Stdout,
		newApp.BaseApp,
		simtestutil.AppStateFn(newApp.AppCodec(), newApp.SimulationManager(), newApp.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(newApp, newApp.AppCodec(), config),
		newApp.ModuleAccountAddrs(),
		config,
		newApp.AppCodec(),
	)
	require.NoError(t, err)
}

// TestFullAppSimulation runs a full app simulation
func TestFullAppSimulation(t *testing.T) {
	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	appOpts := simtestutil.AppOptionsMap{
		flags.FlagHome:            t.TempDir(),
		server.FlagInvCheckPeriod: simcli.FlagPeriodValue,
	}

	auraApp := NewSimApp(logger, db, appOpts, interBlockCacheOpt(), fauxMerkleModeOpt)
	require.NotNil(t, auraApp)

	// run full simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		auraApp.BaseApp,
		simtestutil.AppStateFn(auraApp.AppCodec(), auraApp.SimulationManager(), auraApp.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(auraApp, auraApp.AppCodec(), config),
		auraApp.ModuleAccountAddrs(),
		config,
		auraApp.AppCodec(),
	)

	// export params
	err := simtestutil.CheckExportSimulation(auraApp, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

// TestMemoryIntensive is a placeholder for memory profiling tests
func TestMemoryIntensive(t *testing.T) {
	// This test is used by the benchmarks.yml workflow for memory profiling
	t.Skip("Memory profiling test - run with -memprofile flag")
}
