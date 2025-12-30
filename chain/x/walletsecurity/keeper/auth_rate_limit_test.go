// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
)

// =============================================================================
// Auth Rate Limit Tests
// =============================================================================

func TestCheckAuthRateLimit_FirstAttempt(t *testing.T) {
	suite := &KeeperTestSuite{}
	suite.SetT(t)
	suite.SetupTest()

	ctx := suite.GetContext()
	k := suite.GetKeeper()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(cmtproto.Header{Height: 100})

	signer := sdk.AccAddress("test_signer_address1")

	err := k.CheckAuthRateLimit(sdkCtx, signer)
	require.NoError(t, err)
}

func TestCheckAuthRateLimit_UnderLimit(t *testing.T) {
	suite := &KeeperTestSuite{}
	suite.SetT(t)
	suite.SetupTest()

	ctx := suite.GetContext()
	k := suite.GetKeeper()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(cmtproto.Header{Height: 101})

	signer := sdk.AccAddress("test_signer_address2")

	// Make 4 attempts (under the limit of 5)
	for i := 0; i < 4; i++ {
		err := k.CheckAuthRateLimit(sdkCtx, signer)
		require.NoError(t, err, "Attempt %d should succeed", i+1)
	}
}

func TestCheckAuthRateLimit_AtLimit(t *testing.T) {
	suite := &KeeperTestSuite{}
	suite.SetT(t)
	suite.SetupTest()

	ctx := suite.GetContext()
	k := suite.GetKeeper()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(cmtproto.Header{Height: 102})

	signer := sdk.AccAddress("test_signer_address3")

	// Make exactly 5 attempts (at the limit)
	for i := 0; i < 5; i++ {
		err := k.CheckAuthRateLimit(sdkCtx, signer)
		require.NoError(t, err, "Attempt %d should succeed", i+1)
	}
}

func TestCheckAuthRateLimit_ExceedsLimit(t *testing.T) {
	suite := &KeeperTestSuite{}
	suite.SetT(t)
	suite.SetupTest()

	ctx := suite.GetContext()
	k := suite.GetKeeper()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(cmtproto.Header{Height: 103})

	signer := sdk.AccAddress("test_signer_address4")

	// Make 5 successful attempts
	for i := 0; i < 5; i++ {
		err := k.CheckAuthRateLimit(sdkCtx, signer)
		require.NoError(t, err)
	}

	// The 6th attempt should fail
	err := k.CheckAuthRateLimit(sdkCtx, signer)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrAuthRateLimited)
}

func TestCheckAuthRateLimit_EmptySigner(t *testing.T) {
	suite := &KeeperTestSuite{}
	suite.SetT(t)
	suite.SetupTest()

	ctx := suite.GetContext()
	k := suite.GetKeeper()

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	emptySigner := sdk.AccAddress{}

	err := k.CheckAuthRateLimit(sdkCtx, emptySigner)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestCheckAuthRateLimit_DifferentSigners(t *testing.T) {
	suite := &KeeperTestSuite{}
	suite.SetT(t)
	suite.SetupTest()

	ctx := suite.GetContext()
	k := suite.GetKeeper()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(cmtproto.Header{Height: 104})

	signer1 := sdk.AccAddress("test_signer_1_addr_")
	signer2 := sdk.AccAddress("test_signer_2_addr_")

	// signer1 exhausts limit
	for i := 0; i < 5; i++ {
		err := k.CheckAuthRateLimit(sdkCtx, signer1)
		require.NoError(t, err)
	}

	// signer1 should be rate limited
	err := k.CheckAuthRateLimit(sdkCtx, signer1)
	require.ErrorIs(t, err, types.ErrAuthRateLimited)

	// signer2 should still be able to make attempts
	err = k.CheckAuthRateLimit(sdkCtx, signer2)
	require.NoError(t, err)
}

func TestCheckAuthRateLimit_DifferentBlocks(t *testing.T) {
	suite := &KeeperTestSuite{}
	suite.SetT(t)
	suite.SetupTest()

	ctx := suite.GetContext()
	k := suite.GetKeeper()

	signer := sdk.AccAddress("test_signer_blocks_")

	// Block 105: exhaust limit
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeader(cmtproto.Header{Height: 105})

	for i := 0; i < 5; i++ {
		err := k.CheckAuthRateLimit(sdkCtx, signer)
		require.NoError(t, err)
	}

	err := k.CheckAuthRateLimit(sdkCtx, signer)
	require.ErrorIs(t, err, types.ErrAuthRateLimited)

	// Block 106: limit resets
	sdkCtx = sdkCtx.WithBlockHeader(cmtproto.Header{Height: 106})

	err = k.CheckAuthRateLimit(sdkCtx, signer)
	require.NoError(t, err, "Rate limit should reset in new block")
}

func TestMaxAuthAttemptsPerBlock(t *testing.T) {
	require.Equal(t, 5, MaxAuthAttemptsPerBlock)
}
