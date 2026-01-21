// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

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

// SimpleTestSuite provides a minimal test harness for non-suite tests.
// Used by tests that don't need the full testify suite functionality.
type SimpleTestSuite struct {
	Keeper   *Keeper
	Ctx      sdk.Context
	Cdc      codec.Codec
	StoreKey storetypes.StoreKey
}

// NewTestSuite creates a simple test suite for use in standard Go tests.
// This is a lightweight alternative to KeeperTestSuite for tests that don't
// use the testify suite pattern.
func NewTestSuite(t *testing.T) *SimpleTestSuite {
	t.Helper()

	// Configure SDK with Aura-specific prefixes (safe to call multiple times)
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInputWithKeys(t, "compliance")

	return &SimpleTestSuite{
		Keeper:   NewKeeper(input.Cdc, input.StoreKey),
		Ctx:      input.Ctx,
		Cdc:      input.Cdc,
		StoreKey: input.StoreKey,
	}
}
