package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/testing/testutil"
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
	// Configure SDK with Aura-specific prefixes (safe to call multiple times)
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInput(suite.T())

	suite.Keeper = NewKeeper(
		input.Cdc,
		input.StoreKey,
		testutil.NewMockBankKeeper(),
		testutil.NewMockAccountKeeper(),
		testutil.NewMockVCRegistryKeeper(),
		testutil.NewMockSecurityKeeper(),
	)
	suite.SdkCtx = input.Ctx
	suite.Cdc = input.Cdc
	suite.StoreKey = input.StoreKey
}
