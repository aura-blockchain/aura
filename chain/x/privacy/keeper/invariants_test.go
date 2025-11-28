package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
	"github.com/aequitas/aura/chain/x/privacy/types"
)

type InvariantsTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func (suite *InvariantsTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // authKeeper
		nil, // bankKeeper
	)
	suite.ctx = input.Ctx

	// Set default params
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

// TestAllInvariants tests that all invariants pass on empty store
func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.ctx

	// Test: All invariants on empty store
	inv := keeper.AllInvariants(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

// TestParamsInvariant_Valid tests ParamsInvariant with valid parameters
func (suite *InvariantsTestSuite) TestParamsInvariant_Valid() {
	ctx := suite.ctx

	// Set valid params
	params := types.DefaultParams()
	err := suite.keeper.SetParams(ctx, params)
	suite.Require().NoError(err)

	// Test invariant
	inv := keeper.ParamsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "params invariant should pass with valid params")
	suite.Empty(msg)
}

// TestParamsInvariant_InvalidMinRingSize tests ParamsInvariant with invalid min ring size
func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidMinRingSize() {
	ctx := suite.ctx

	// Set params with invalid min ring size (< 2)
	params := types.DefaultParams()
	params.MinRingSize = 1
	err := suite.keeper.SetParams(ctx, params)
	suite.Require().NoError(err)

	// Test invariant
	inv := keeper.ParamsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.True(broken, "params invariant should fail with min ring size < 2")
	suite.Contains(msg, "min ring size")
}

// TestParamsInvariant_InvalidMaxRingSize tests ParamsInvariant with invalid max ring size
func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidMaxRingSize() {
	ctx := suite.ctx

	// Set params with max < min
	params := types.DefaultParams()
	params.MinRingSize = 5
	params.MaxRingSize = 3 // max < min
	err := suite.keeper.SetParams(ctx, params)
	suite.Require().NoError(err)

	// Test invariant
	inv := keeper.ParamsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.True(broken, "params invariant should fail when max < min")
	suite.Contains(msg, "max ring size")
}

// TestMixingStateConsistencyInvariant_EmptyStore tests MixingStateConsistencyInvariant on empty store
func (suite *InvariantsTestSuite) TestMixingStateConsistencyInvariant_EmptyStore() {
	ctx := suite.ctx

	// Test invariant on empty store
	inv := keeper.MixingStateConsistencyInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "mixing state invariant should pass on empty store")
	suite.Empty(msg)
}

// TestCommitmentValidityInvariant_EmptyStore tests CommitmentValidityInvariant on empty store
func (suite *InvariantsTestSuite) TestCommitmentValidityInvariant_EmptyStore() {
	ctx := suite.ctx

	// Test invariant on empty store
	inv := keeper.CommitmentValidityInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "commitment validity invariant should pass on empty store")
	suite.Empty(msg)
}

// TestAllInvariants_WithValidData tests all invariants with valid data
func (suite *InvariantsTestSuite) TestAllInvariants_WithValidData() {
	ctx := suite.ctx

	// Set valid params
	params := types.DefaultParams()
	params.MinRingSize = 3
	params.MaxRingSize = 10
	params.MinMixingParticipants = 2
	err := suite.keeper.SetParams(ctx, params)
	suite.Require().NoError(err)

	// Test all invariants
	inv := keeper.AllInvariants(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass with valid data")
	suite.Empty(msg)
}

// TestAllInvariants_WithInvalidParams tests that AllInvariants detects invalid params
func (suite *InvariantsTestSuite) TestAllInvariants_WithInvalidParams() {
	ctx := suite.ctx

	// Set invalid params
	params := types.DefaultParams()
	params.MinRingSize = 0 // Invalid - must be >= 2
	err := suite.keeper.SetParams(ctx, params)
	suite.Require().NoError(err)

	// Test all invariants
	inv := keeper.AllInvariants(suite.keeper)
	msg, broken := inv(ctx)
	suite.True(broken, "all invariants should detect invalid params")
	suite.NotEmpty(msg)
}

// TestParamsInvariant_EdgeCase_MinEqualsMax tests params where min == max
func (suite *InvariantsTestSuite) TestParamsInvariant_EdgeCase_MinEqualsMax() {
	ctx := suite.ctx

	// Set params where min == max (should be valid)
	params := types.DefaultParams()
	params.MinRingSize = 5
	params.MaxRingSize = 5
	err := suite.keeper.SetParams(ctx, params)
	suite.Require().NoError(err)

	// Test invariant
	inv := keeper.ParamsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "params invariant should pass when min == max")
	suite.Empty(msg)
}

// TestInvariantFunctions_NotPanic tests that invariant functions don't panic
func (suite *InvariantsTestSuite) TestInvariantFunctions_NotPanic() {
	ctx := suite.ctx

	// Test ParamsInvariant doesn't panic
	suite.NotPanics(func() {
		inv := keeper.ParamsInvariant(suite.keeper)
		inv(ctx)
	})

	// Test MixingStateConsistencyInvariant doesn't panic
	suite.NotPanics(func() {
		inv := keeper.MixingStateConsistencyInvariant(suite.keeper)
		inv(ctx)
	})

	// Test CommitmentValidityInvariant doesn't panic
	suite.NotPanics(func() {
		inv := keeper.CommitmentValidityInvariant(suite.keeper)
		inv(ctx)
	})

	// Test AllInvariants doesn't panic
	suite.NotPanics(func() {
		inv := keeper.AllInvariants(suite.keeper)
		inv(ctx)
	})
}
