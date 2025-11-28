package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type InvariantsTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
}

func (suite *InvariantsTestSuite) SetupTest() {
	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create test context
	testCtx := testutil.DefaultContextWithDB(suite.T(), storeKey, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	suite.cdc = codec.NewProtoCodec(registry)

	// Create keeper
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		storeKey,
		nil, // wasmd keeper not needed for invariant tests
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.ctx

	// Test: All invariants on empty store
	inv := keeper.AllInvariants(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Test individual invariants directly without registry
	// Cosmos SDK v0.50 doesn't have NewInvariantRegistry

	ctx := suite.ctx

	// Test ParamsInvariant
	inv := keeper.ParamsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "params invariant should pass")
	suite.Empty(msg)

	// Test SecurityStatsInvariant
	inv = keeper.SecurityStatsInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "security stats invariant should pass")
	suite.Empty(msg)

	// Test PausedContractsInvariant
	inv = keeper.PausedContractsInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "paused contracts invariant should pass")
	suite.Empty(msg)

	// Test AuthorizedUploadersInvariant
	inv = keeper.AuthorizedUploadersInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "authorized uploaders invariant should pass")
	suite.Empty(msg)
}
