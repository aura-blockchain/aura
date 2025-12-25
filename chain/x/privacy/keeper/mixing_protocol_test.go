// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
	"github.com/aequitas/aura/chain/x/privacy/types"
)

type MixingProtocolTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func TestMixingProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(MixingProtocolTestSuite))
}

func (suite *MixingProtocolTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
	)
	suite.ctx = input.Ctx

	// Enable mixing and set params
	params := types.DefaultParams()
	params.EnableMixing = true
	params.MinMixingParticipants = 3
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
}

// JoinMixingRound Tests

func (suite *MixingProtocolTestSuite) TestJoinMixingRound_NotEnabled() {
	// Disable mixing
	params := suite.keeper.GetParams(suite.ctx)
	params.EnableMixing = false
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	participantID, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		keepertest.GenTestAddr().String(),
		[]byte("input"),
		math.NewInt(1000),
	)

	suite.Require().Error(err)
	suite.Require().Empty(participantID)
	suite.Require().Contains(err.Error(), "mixing not enabled")
}

func (suite *MixingProtocolTestSuite) TestJoinMixingRound_ZeroAmount() {
	participantID, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		keepertest.GenTestAddr().String(),
		[]byte("input"),
		math.NewInt(0),
	)

	suite.Require().Error(err)
	suite.Require().Empty(participantID)
	suite.Require().Contains(err.Error(), "invalid amount")
}

func (suite *MixingProtocolTestSuite) TestJoinMixingRound_NegativeAmount() {
	participantID, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		keepertest.GenTestAddr().String(),
		[]byte("input"),
		math.NewInt(-1000),
	)

	suite.Require().Error(err)
	suite.Require().Empty(participantID)
	suite.Require().Contains(err.Error(), "invalid amount")
}

func (suite *MixingProtocolTestSuite) TestJoinMixingRound_Success() {
	participant := keepertest.GenTestAddr().String()

	participantID, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input"),
		math.NewInt(1000),
	)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(participantID)

	// Verify pool was created and participant added
	pool, err := suite.keeper.GetMixingPool(suite.ctx, "pool1")
	suite.Require().NoError(err)
	suite.Require().Equal("open", pool.Status)
	suite.Require().Len(pool.Participants, 1)
}

func (suite *MixingProtocolTestSuite) TestJoinMixingRound_MultipleInputs() {
	participant := keepertest.GenTestAddr().String()

	// Join first time
	pid1, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input1"),
		math.NewInt(1000),
	)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(pid1)

	// Same participant can join again with different input (different participant ID generated)
	pid2, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input2"),
		math.NewInt(1000),
	)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(pid2)
	suite.Require().NotEqual(pid1, pid2, "Different inputs should generate different participant IDs")
}

func (suite *MixingProtocolTestSuite) TestJoinMixingRound_PoolBecomesReady() {
	params := suite.keeper.GetParams(suite.ctx)

	// Join with min participants
	for i := 0; i < int(params.MinMixingParticipants); i++ {
		participant := keepertest.GenTestAddr().String()
		_, err := suite.keeper.JoinMixingRound(
			suite.ctx,
			"pool1",
			participant,
			[]byte("input"),
			math.NewInt(1000),
		)
		suite.Require().NoError(err)
	}

	// Pool should be ready
	pool, err := suite.keeper.GetMixingPool(suite.ctx, "pool1")
	suite.Require().NoError(err)
	suite.Require().Equal("ready", pool.Status)
}

func (suite *MixingProtocolTestSuite) TestJoinMixingRound_PoolFull() {
	params := suite.keeper.GetParams(suite.ctx)
	maxParticipants := params.MinMixingParticipants * 4

	// Fill the pool
	for i := uint32(0); i < maxParticipants; i++ {
		participant := keepertest.GenTestAddr().String()
		_, err := suite.keeper.JoinMixingRound(
			suite.ctx,
			"pool1",
			participant,
			[]byte("input"),
			math.NewInt(1000),
		)
		suite.Require().NoError(err)
	}

	// Try to join full pool
	participant := keepertest.GenTestAddr().String()
	participantID, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input"),
		math.NewInt(1000),
	)

	suite.Require().Error(err)
	suite.Require().Empty(participantID)
	suite.Require().Contains(err.Error(), "pool is full")
}

// ExecuteMixing Tests

func (suite *MixingProtocolTestSuite) TestExecuteMixing_PoolNotFound() {
	err := suite.keeper.ExecuteMixing(suite.ctx, "nonexistent_pool")

	suite.Require().Error(err)
}

func (suite *MixingProtocolTestSuite) TestExecuteMixing_PoolNotReady() {
	// Create a pool but don't fill it to ready state
	participant := keepertest.GenTestAddr().String()
	_, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input"),
		math.NewInt(1000),
	)
	suite.Require().NoError(err)

	// Try to execute mixing
	err = suite.keeper.ExecuteMixing(suite.ctx, "pool1")

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "pool not ready")
}

func (suite *MixingProtocolTestSuite) TestExecuteMixing_Success() {
	params := suite.keeper.GetParams(suite.ctx)

	// Create a ready pool
	for i := 0; i < int(params.MinMixingParticipants); i++ {
		participant := keepertest.GenTestAddr().String()
		_, err := suite.keeper.JoinMixingRound(
			suite.ctx,
			"pool1",
			participant,
			[]byte("input"),
			math.NewInt(1000),
		)
		suite.Require().NoError(err)
	}

	// Execute mixing
	err := suite.keeper.ExecuteMixing(suite.ctx, "pool1")

	suite.Require().NoError(err)

	// Verify pool status is completed
	pool, err := suite.keeper.GetMixingPool(suite.ctx, "pool1")
	suite.Require().NoError(err)
	suite.Require().Equal("completed", pool.Status)
	suite.Require().Len(pool.Participants, int(params.MinMixingParticipants))
}

// WithdrawFromMixing Tests

func (suite *MixingProtocolTestSuite) TestWithdrawFromMixing_PoolNotFound() {
	err := suite.keeper.WithdrawFromMixing(suite.ctx, "nonexistent_pool", "participant1", []byte("output"))

	suite.Require().Error(err)
}

func (suite *MixingProtocolTestSuite) TestWithdrawFromMixing_MixingNotCompleted() {
	// Create a pool
	participant := keepertest.GenTestAddr().String()
	participantID, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input"),
		math.NewInt(1000),
	)
	suite.Require().NoError(err)

	// Try to withdraw before completion
	err = suite.keeper.WithdrawFromMixing(suite.ctx, "pool1", participantID, []byte("output"))

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "mixing not yet completed")
}

func (suite *MixingProtocolTestSuite) TestWithdrawFromMixing_ParticipantNotFound() {
	params := suite.keeper.GetParams(suite.ctx)

	// Create and complete a mixing pool
	for i := 0; i < int(params.MinMixingParticipants); i++ {
		participant := keepertest.GenTestAddr().String()
		_, err := suite.keeper.JoinMixingRound(
			suite.ctx,
			"pool1",
			participant,
			[]byte("input"),
			math.NewInt(1000),
		)
		suite.Require().NoError(err)
	}

	err := suite.keeper.ExecuteMixing(suite.ctx, "pool1")
	suite.Require().NoError(err)

	// Try to withdraw with non-participant
	err = suite.keeper.WithdrawFromMixing(suite.ctx, "pool1", "nonexistent_participant", []byte("output"))

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "participant not in pool")
}

func (suite *MixingProtocolTestSuite) TestWithdrawFromMixing_Success() {
	params := suite.keeper.GetParams(suite.ctx)

	// Create and complete a mixing pool
	var participantID string
	for i := 0; i < int(params.MinMixingParticipants); i++ {
		participant := keepertest.GenTestAddr().String()
		pid, err := suite.keeper.JoinMixingRound(
			suite.ctx,
			"pool1",
			participant,
			[]byte("input"),
			math.NewInt(1000),
		)
		suite.Require().NoError(err)
		if i == 0 {
			participantID = pid
		}
	}

	err := suite.keeper.ExecuteMixing(suite.ctx, "pool1")
	suite.Require().NoError(err)

	// Withdraw
	err = suite.keeper.WithdrawFromMixing(suite.ctx, "pool1", participantID, []byte("output"))

	suite.Require().NoError(err)
}

// GetMixingPoolStatus Tests

func (suite *MixingProtocolTestSuite) TestGetMixingPoolStatus_PoolNotFound() {
	status, err := suite.keeper.GetMixingPoolStatus(suite.ctx, "nonexistent_pool")

	suite.Require().Error(err)
	suite.Require().Empty(status)
}

func (suite *MixingProtocolTestSuite) TestGetMixingPoolStatus_Success() {
	// Create a pool
	participant := keepertest.GenTestAddr().String()
	_, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input"),
		math.NewInt(1000),
	)
	suite.Require().NoError(err)

	// Get status
	status, err := suite.keeper.GetMixingPoolStatus(suite.ctx, "pool1")

	suite.Require().NoError(err)
	suite.Require().Equal("open", status)
}

// CancelMixingPool Tests

func (suite *MixingProtocolTestSuite) TestCancelMixingPool_PoolNotFound() {
	err := suite.keeper.CancelMixingPool(suite.ctx, "nonexistent_pool")

	suite.Require().Error(err)
}

func (suite *MixingProtocolTestSuite) TestCancelMixingPool_AlreadyCompleted() {
	params := suite.keeper.GetParams(suite.ctx)

	// Create and complete a mixing pool
	for i := 0; i < int(params.MinMixingParticipants); i++ {
		participant := keepertest.GenTestAddr().String()
		_, err := suite.keeper.JoinMixingRound(
			suite.ctx,
			"pool1",
			participant,
			[]byte("input"),
			math.NewInt(1000),
		)
		suite.Require().NoError(err)
	}

	err := suite.keeper.ExecuteMixing(suite.ctx, "pool1")
	suite.Require().NoError(err)

	// Try to cancel
	err = suite.keeper.CancelMixingPool(suite.ctx, "pool1")

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot cancel pool that has started mixing")
}

func (suite *MixingProtocolTestSuite) TestCancelMixingPool_Success() {
	// Create a pool
	participant := keepertest.GenTestAddr().String()
	_, err := suite.keeper.JoinMixingRound(
		suite.ctx,
		"pool1",
		participant,
		[]byte("input"),
		math.NewInt(1000),
	)
	suite.Require().NoError(err)

	// Cancel it
	err = suite.keeper.CancelMixingPool(suite.ctx, "pool1")

	suite.Require().NoError(err)

	// Verify status
	status, err := suite.keeper.GetMixingPoolStatus(suite.ctx, "pool1")
	suite.Require().NoError(err)
	suite.Require().Equal("cancelled", status)
}

// Standalone tests

func TestJoinMixingRound_UniqueParticipantIDs(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	params := types.DefaultParams()
	params.EnableMixing = true
	err := k.SetParams(input.Ctx, params)
	require.NoError(t, err)

	// Join with different participants
	participant1 := keepertest.GenTestAddr().String()
	participant2 := keepertest.GenTestAddr().String()

	pid1, err := k.JoinMixingRound(input.Ctx, "pool1", participant1, []byte("input1"), math.NewInt(1000))
	require.NoError(t, err)

	pid2, err := k.JoinMixingRound(input.Ctx, "pool1", participant2, []byte("input2"), math.NewInt(1000))
	require.NoError(t, err)

	require.NotEqual(t, pid1, pid2, "Participant IDs should be unique")
}
