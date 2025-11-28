package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
)

// KeeperTestSuite is a base test suite for keeper tests in the keeper_test package
type KeeperTestSuite struct {
	suite.Suite

	keeper   *keeper.Keeper
	SdkCtx   sdk.Context
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// SetupTest initializes the test suite before each test
func (suite *KeeperTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
	)
	suite.SdkCtx = input.Ctx
	suite.cdc = input.Cdc
	suite.storeKey = input.StoreKey
}
