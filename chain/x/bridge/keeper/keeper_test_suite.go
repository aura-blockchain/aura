package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

// KeeperTestSuite is a base test suite for keeper tests in the keeper package
type KeeperTestSuite struct {
	suite.Suite

	Keeper   *Keeper
	SdkCtx   sdk.Context
	Cdc      codec.BinaryCodec
	StoreKey storetypes.StoreKey
}

// SetupTest initializes the test suite before each test
func (suite *KeeperTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	suite.Keeper = NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)
	suite.SdkCtx = input.Ctx
	suite.Cdc = input.Cdc
	suite.StoreKey = input.StoreKey
}
