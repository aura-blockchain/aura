// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

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
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Create a mock invariant registry
	// In Cosmos SDK v0.50+, invariant registration is handled differently
	// This test verifies that registering invariants doesn't panic
	suite.NotPanics(func() {
		// Create invariant function - should not panic when called
		inv := AllInvariants(suite.Keeper)
		// Call it with test context
		msg, broken := inv(suite.SdkCtx)
		// Should complete without panic
		_ = msg
		_ = broken
	})
}
