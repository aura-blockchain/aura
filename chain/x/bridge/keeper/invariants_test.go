package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx

	// To pass all invariants, we need at least one active validator
	// Add a valid active validator first
	validator := &bridgepb.BridgeValidator{
		Address: sdk.AccAddress("validatoraddr_____").String(),
		Power:   100,
		Active:  true,
	}
	suite.storeValidator(ctx, validator)

	// Test: All invariants should pass with proper setup
	inv := AllInvariants(*suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass with valid validator")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestAllInvariantsEmptyStore() {
	ctx := suite.SdkCtx

	// Empty store should fail validator set invariant (needs at least 1 active validator)
	inv := AllInvariants(*suite.Keeper)
	msg, broken := inv(ctx)
	suite.True(broken, "empty store should fail due to no active validators")
	suite.Contains(msg, "active validators")
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Create a mock invariant registry
	registry := mockInvariantRegistry{}

	// Test registration - should not panic
	suite.NotPanics(func() {
		RegisterInvariants(&registry, *suite.Keeper)
	})

	// Verify all expected invariants are registered
	suite.Equal(7, registry.count, "should register 7 invariants")
}

// mockInvariantRegistry implements sdk.InvariantRegistry for testing
type mockInvariantRegistry struct {
	count int
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	m.count++
}

// storeValidator stores a validator in the keeper
func (suite *InvariantsTestSuite) storeValidator(ctx sdk.Context, validator *bridgepb.BridgeValidator) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(validator)
	store.Set(append(types.ValidatorPrefix, []byte(validator.Address)...), bz)
}
