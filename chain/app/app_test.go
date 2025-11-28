package app

import (
	"testing"
	"time"

	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNewAppRegistersServices(t *testing.T) {
	app := NewApp()
	info := app.GRPCServer().GetServiceInfo()

	// Verify identitychange module services are registered
	require.Contains(t, info, "aura.identitychange.v1beta1.Msg")
	require.Contains(t, info, "aura.identitychange.v1beta1.Query")

	// Verify inclusionroutines module services are registered
	require.Contains(t, info, "aura.inclusionroutines.v1beta1.Msg")
	require.Contains(t, info, "aura.inclusionroutines.v1beta1.Query")

	// Verify confidencescore module services are registered
	require.Contains(t, info, "aura.confidencescore.v1beta1.Msg")
	require.Contains(t, info, "aura.confidencescore.v1beta1.Query")

	// Verify compliance module services are registered
	require.Contains(t, info, "aura.compliance.v1beta1.Msg")
	require.Contains(t, info, "aura.compliance.v1beta1.Query")
	// Verify wallet security module services are registered
	require.Contains(t, info, "aura.walletsecurity.v1beta1.Msg")
	require.Contains(t, info, "aura.walletsecurity.v1beta1.Query")
	// Verify validator security module services are registered
	require.Contains(t, info, "aura.validatorsecurity.v1beta1.Msg")
	require.Contains(t, info, "aura.validatorsecurity.v1beta1.Query")
	// Verify cryptography module services are registered
	require.Contains(t, info, "aura.cryptography.v1beta1.Msg")
	require.Contains(t, info, "aura.cryptography.v1beta1.Query")
}

func TestNewCosmosAppExposesBaseApp(t *testing.T) {
	cApp := NewCosmosApp(nil)
	require.NotNil(t, cApp.BaseApp)
	require.NotNil(t, cApp.Encoding().InterfaceRegistry)
	info := cApp.GRPCServer().GetServiceInfo()

	// Verify identitychange module services are registered
	require.Contains(t, info, "aura.identitychange.v1beta1.Msg")
	require.Contains(t, info, "aura.identitychange.v1beta1.Query")

	// Verify inclusionroutines module services are registered
	require.Contains(t, info, "aura.inclusionroutines.v1beta1.Msg")
	require.Contains(t, info, "aura.inclusionroutines.v1beta1.Query")

	// Verify confidencescore module services are registered
	require.Contains(t, info, "aura.confidencescore.v1beta1.Msg")
	require.Contains(t, info, "aura.confidencescore.v1beta1.Query")

	// Verify compliance module services are registered
	require.Contains(t, info, "aura.compliance.v1beta1.Msg")
	require.Contains(t, info, "aura.compliance.v1beta1.Query")
	// Verify wallet security module services are registered
	require.Contains(t, info, "aura.walletsecurity.v1beta1.Msg")
	require.Contains(t, info, "aura.walletsecurity.v1beta1.Query")
	// Verify validator security module services are registered
	require.Contains(t, info, "aura.validatorsecurity.v1beta1.Msg")
	require.Contains(t, info, "aura.validatorsecurity.v1beta1.Query")
	// Verify cryptography module services are registered
	require.Contains(t, info, "aura.cryptography.v1beta1.Msg")
	require.Contains(t, info, "aura.cryptography.v1beta1.Query")
}

func TestCosmosAppBeginBlockCleansOldMintCounts(t *testing.T) {
	cApp := NewCosmosApp(nil)
	require.NotNil(t, cApp.vcKeeper)

	addr := "aura1kvbeginblock"
	now := time.Now().Unix()
	old := now - (10 * 86400)

	// Seed KV-backed mint counts in the past
	sdkCtx := cApp.BaseApp.NewUncachedContext(false, tmproto.Header{
		Height: 1,
		Time:   time.Unix(old, 0),
	})
	ctx := sdk.WrapSDKContext(sdkCtx)
	cApp.vcKeeper.SetCurrentTime(old)
	cApp.vcKeeper.IncrementMintCount(ctx, addr)

	oldKey := vctypes.UserMintCountKey(addr, old/86400)
	store := sdkCtx.KVStore(cApp.storeKeys.vc)
	require.NotNil(t, store.Get(oldKey), "expected mint counter persisted before BeginBlock")

	// Run module manager BeginBlock with a real SDK context; cleanup should prune the stale counter
	beginCtx := cApp.BaseApp.NewUncachedContext(false, tmproto.Header{
		Height: 2,
		Time:   time.Unix(now, 0),
	})
	cApp.moduleManager.BeginBlock(sdk.WrapSDKContext(beginCtx))

	newCtx := cApp.BaseApp.NewUncachedContext(false, tmproto.Header{})
	require.Nil(t, newCtx.KVStore(cApp.storeKeys.vc).Get(oldKey), "stale mint count key should be deleted")
}

func TestAppBridgeGenesisRoundTrip(t *testing.T) {
	app := NewApp()
	ctx := app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	genesis := *bridgetypes.DefaultGenesis()
	require.NoError(t, app.InitBridgeGenesis(ctx, genesis))

	exported := app.ExportBridgeGenesis(ctx)
	require.Len(t, exported, 1)
	require.NotNil(t, exported[0].Params)
	require.Equal(t, genesis.Params, exported[0].Params)
}
