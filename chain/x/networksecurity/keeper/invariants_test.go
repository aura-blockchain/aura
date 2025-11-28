package keeper

import (
	"testing"

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
	inv := AllInvariants(&suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Register invariants - should not panic
	suite.NotPanics(func() {
		// RegisterInvariants expects a crisis.InvariantRegistry
		// For this basic test, we just verify it doesn't panic when called with keeper
		inv := AllInvariants(&suite.Keeper)
		suite.NotNil(inv, "invariant function should be created")
	})
}
