package keeper

import (
	"context"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
)

// KeeperTestSuite is the base test suite for keeper tests
type KeeperTestSuite struct {
	suite.Suite
	ctx    context.Context
	keeper Keeper
	cdc    codec.BinaryCodec
}

// SetupTest initializes the test suite with a keeper and context
func (suite *KeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create database and commit multi-store
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	suite.Require().NoError(cms.LoadLatestVersion())

	// Create SDK context
	header := cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	sdkCtx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	// Create keeper
	storeService := runtime.NewKVStoreService(storeKey)
	suite.keeper = NewKeeper(cdc, storeService, log.NewNopLogger())
	suite.cdc = cdc
	suite.ctx = sdkCtx
}

// GetContext returns the test context
func (suite *KeeperTestSuite) GetContext() context.Context {
	return suite.ctx
}

// GetKeeper returns the keeper instance
func (suite *KeeperTestSuite) GetKeeper() Keeper {
	return suite.keeper
}

// GetCodec returns the codec instance
func (suite *KeeperTestSuite) GetCodec() codec.BinaryCodec {
	return suite.cdc
}
