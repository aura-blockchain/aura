package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx

	// Test: All invariants on empty store
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Cosmos SDK v0.53 doesn't expose sdk.NewInvariantRegistry, so exercise each invariant directly.
	invariants := []func(sdk.Context) (string, bool){
		ParamsInvariant(suite.Keeper),
		PoolReservesConsistencyInvariant(suite.Keeper),
		OrderValidityInvariant(suite.Keeper),
		LiquidityProviderConsistencyInvariant(suite.Keeper),
		SecurityLimitsInvariant(suite.Keeper),
		HTLCValidityInvariant(suite.Keeper),
	}

	for _, inv := range invariants {
		msg, broken := inv(suite.SdkCtx)
		suite.False(broken)
		suite.Empty(msg)
	}
}
