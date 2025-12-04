package keeper

import (
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/auth/types"
)

// KeeperTestSuite is a test suite for the auth keeper
type KeeperTestSuite struct {
	suite.Suite

	Ctx    sdk.Context
	SdkCtx sdk.Context // Alias for Ctx for compatibility
	Keeper *Keeper
	Cdc    codec.BinaryCodec
}

// SetupTest initializes the test suite before each test
func (suite *KeeperTestSuite) SetupTest() {
	// Create in-memory database
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	suite.Require().NoError(stateStore.LoadLatestVersion())

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	suite.Cdc = codec.NewProtoCodec(registry)

	// Create context
	suite.Ctx = sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())
	suite.SdkCtx = suite.Ctx

	// Create keeper
	suite.Keeper = NewKeeper(suite.Cdc, storeKey)

	// Initialize default params
	err := suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
	suite.Require().NoError(err)
}
