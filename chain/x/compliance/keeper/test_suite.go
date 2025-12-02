package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

// KeeperTestSuite provides a shared harness for compliance keeper tests.
type KeeperTestSuite struct {
	suite.Suite

	Keeper   *Keeper
	SdkCtx   sdk.Context
	Cdc      codec.Codec
	StoreKey storetypes.StoreKey
}

func (suite *KeeperTestSuite) SetupTest() {
	// Configure SDK with Aura-specific prefixes (safe to call multiple times)
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInputWithKeys(suite.T(), "compliance")

	suite.Keeper = NewKeeper(input.Cdc, input.StoreKey)
	suite.SdkCtx = input.Ctx
	suite.Cdc = input.Cdc
	suite.StoreKey = input.StoreKey
}
