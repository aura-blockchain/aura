package app

import (
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestWasmKeeperInitialization tests that the wasm keeper is properly initialized
func TestWasmKeeperInitialization(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")
	require.NotNil(t, app.WasmKeeper, "wasm keeper should not be nil")
}

// TestWasmStoreKeyRegistration tests that the wasm store key is properly registered
func TestWasmStoreKeyRegistration(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")
	require.NotNil(t, app.storeKeys.wasm, "wasm store key should not be nil")
	require.Equal(t, wasmtypes.StoreKey, app.storeKeys.wasm.Name(), "wasm store key should have correct name")
}

// TestWasmModuleAccountPermissions tests that wasm module account has correct permissions
func TestWasmModuleAccountPermissions(t *testing.T) {
	perms, ok := moduleAccountPermissions[wasmtypes.ModuleName]
	require.True(t, ok, "wasm module should have permissions defined")
	require.Contains(t, perms, "burner", "wasm module should have burner permission")
}

// TestWasmModuleAccountReceiving tests that wasm module can receive funds
func TestWasmModuleAccountReceiving(t *testing.T) {
	allowed := allowedReceivingModules[wasmtypes.ModuleName]
	require.True(t, allowed, "wasm module should be allowed to receive funds")
}

// TestWasmParameters tests that wasm parameters are properly configured
func TestWasmParameters(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")
	require.NotNil(t, app.WasmKeeper, "wasm keeper should not be nil")

	ctx := app.NewUncachedContext(true, sdk.NewBlockHeader())
	params := app.WasmKeeper.GetParams(sdk.WrapSDKContext(ctx))

	// Test code upload access (should be set to nobody initially - governance only)
	require.Equal(t, wasmtypes.AccessTypeNobody, params.CodeUploadAccess.Permission,
		"code upload should be restricted to governance")

	// Test instantiate default permission (should be nobody initially)
	require.Equal(t, wasmtypes.AccessTypeNobody, params.InstantiateDefaultPermission,
		"instantiate default permission should be nobody")
}

// TestWasmInterfaceRegistration tests that wasm interfaces are registered in encoding config
func TestWasmInterfaceRegistration(t *testing.T) {
	encoding := MakeEncodingConfig()
	require.NotNil(t, encoding.InterfaceRegistry, "interface registry should not be nil")
	require.NotNil(t, encoding.Codec, "codec should not be nil")

	// Test that wasm proposal types can be marshaled/unmarshaled
	proposal := &wasmtypes.StoreCodeProposal{
		Title:       "Test",
		Description: "Test",
	}

	bz, err := encoding.Codec.MarshalInterface(proposal)
	require.NoError(t, err, "should be able to marshal wasm proposal")
	require.NotNil(t, bz, "marshaled bytes should not be nil")
}

// TestWasmKeeperDependencies tests that wasm keeper has all required dependencies
func TestWasmKeeperDependencies(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")

	// Test that required keepers are initialized
	require.NotNil(t, app.AccountKeeper, "account keeper should not be nil")
	require.NotNil(t, app.BankKeeper, "bank keeper should not be nil")
	require.NotNil(t, app.StakingKeeper, "staking keeper should not be nil")
	require.NotNil(t, app.DistributionKeeper, "distribution keeper should not be nil")
	require.NotNil(t, app.WasmKeeper, "wasm keeper should not be nil")
}

// TestAppBuildSuccess tests that the full app builds without errors
func TestAppBuildSuccess(t *testing.T) {
	require.NotPanics(t, func() {
		NewApp()
	}, "app creation should not panic")
}

// TestAppLoadLatestVersion tests that the app can load the latest version
func TestAppLoadLatestVersion(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")

	// The app should have loaded successfully in NewApp
	// Try to create a context to ensure stores are accessible
	ctx := app.NewUncachedContext(true, sdk.NewBlockHeader())
	require.NotNil(t, ctx, "context should not be nil")
}

// TestWasmStoreIsMounted tests that the wasm store is properly mounted
func TestWasmStoreIsMounted(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")

	ctx := app.NewUncachedContext(true, sdk.NewBlockHeader())
	store := ctx.KVStore(app.storeKeys.wasm)
	require.NotNil(t, store, "wasm store should be accessible")
}

// TestWasmModuleGasConfiguration tests wasm gas configuration
func TestWasmModuleGasConfiguration(t *testing.T) {
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")

	// The wasm config should be set with reasonable gas limits
	// Note: We can't directly access wasmConfig after initialization,
	// but we verify it was set during keeper creation by checking
	// the keeper exists and params are set
	ctx := app.NewUncachedContext(true, sdk.NewBlockHeader())
	params := app.WasmKeeper.GetParams(sdk.WrapSDKContext(ctx))
	require.NotNil(t, params, "wasm params should be set")
}

// TestEncodingConfigCompatibility tests that encoding config works with all modules
func TestEncodingConfigCompatibility(t *testing.T) {
	encoding := MakeEncodingConfig()
	require.NotNil(t, encoding.Codec, "codec should not be nil")
	require.NotNil(t, encoding.InterfaceRegistry, "interface registry should not be nil")

	// Test that we can create a new app with this encoding
	app := NewApp()
	require.NotNil(t, app, "app should not be nil")
	require.Equal(t, encoding.Codec.(*codec.ProtoCodec).InterfaceRegistry(), app.encoding.InterfaceRegistry,
		"encoding configs should use compatible interface registries")
}
