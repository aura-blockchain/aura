package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

// KeeperTestSuite is the base test suite for keeper tests
type KeeperTestSuite struct {
	suite.Suite

	Keeper *Keeper
	SdkCtx sdk.Context
}

// SetupTest initializes the test suite
func (suite *KeeperTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.Keeper = NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)
	suite.SdkCtx = input.Ctx

	// Set default params
	params := types.DefaultParams()
	suite.Require().NoError(suite.Keeper.SetParams(params))
}

// TestKeeperTestSuite runs the test suite
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
