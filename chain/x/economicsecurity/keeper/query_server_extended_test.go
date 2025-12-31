// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	pb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// =============================================================================
// Query Server Tests
// =============================================================================

func TestQueryParams_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.Params(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.Params(ctx, &pb.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
}

func TestQueryVestingSchedule_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Test nil request
	resp, err := queryServer.VestingSchedule(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test not found
	resp, err = queryServer.VestingSchedule(ctx, &pb.QueryVestingScheduleRequest{
		ScheduleId: "nonexistent",
	})
	require.Error(t, err)
	require.Nil(t, resp)

	// Create a vesting schedule
	startTime := time.Unix(currentTime-1000, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0,
		2000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Test found
	resp, err = queryServer.VestingSchedule(ctx, &pb.QueryVestingScheduleRequest{
		ScheduleId: scheduleID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Schedule)
}

func TestQueryVestingSchedulesByBeneficiary_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Test nil request
	resp, err := queryServer.VestingSchedulesByBeneficiary(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test empty beneficiary
	resp, err = queryServer.VestingSchedulesByBeneficiary(ctx, &pb.QueryVestingSchedulesByBeneficiaryRequest{
		BeneficiaryAddress: "",
	})
	require.Error(t, err)

	// Create a vesting schedule
	startTime := time.Unix(currentTime-1000, 0)
	_, err = k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0,
		2000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Test found
	resp, err = queryServer.VestingSchedulesByBeneficiary(ctx, &pb.QueryVestingSchedulesByBeneficiaryRequest{
		BeneficiaryAddress: "aura1beneficiary1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryVoteLock_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.VoteLock(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test not found
	resp, err = queryServer.VoteLock(ctx, &pb.QueryVoteLockRequest{
		LockId: "nonexistent",
	})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestQueryVoteLocksByOwner_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.VoteLocksByOwner(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test empty owner
	resp, err = queryServer.VoteLocksByOwner(ctx, &pb.QueryVoteLocksByOwnerRequest{
		Owner: "",
	})
	require.Error(t, err)

	// Test with valid owner (empty result)
	resp, err = queryServer.VoteLocksByOwner(ctx, &pb.QueryVoteLocksByOwnerRequest{
		Owner: "aura1owner",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryVotingPower_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.VotingPower(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test empty address
	resp, err = queryServer.VotingPower(ctx, &pb.QueryVotingPowerRequest{
		Address: "",
	})
	require.Error(t, err)

	// Test valid address
	resp, err = queryServer.VotingPower(ctx, &pb.QueryVotingPowerRequest{
		Address: "aura1voter",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryPendingTreasuryTx_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.PendingTreasuryTx(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test not found
	resp, err = queryServer.PendingTreasuryTx(ctx, &pb.QueryPendingTreasuryTxRequest{
		TxId: "nonexistent",
	})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestQueryPendingTreasuryTxs_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.PendingTreasuryTxs(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request (empty result)
	resp, err = queryServer.PendingTreasuryTxs(ctx, &pb.QueryPendingTreasuryTxsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryInflationMetrics_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.InflationMetrics(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.InflationMetrics(ctx, &pb.QueryInflationMetricsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryInflationAlerts_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.InflationAlerts(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.InflationAlerts(ctx, &pb.QueryInflationAlertsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryLiquidityMiningStats_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.LiquidityMiningStats(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.LiquidityMiningStats(ctx, &pb.QueryLiquidityMiningStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryMEVStats_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.MEVStats(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.MEVStats(ctx, &pb.QueryMEVStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryUserMEVBalance_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.UserMEVBalance(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test empty address
	resp, err = queryServer.UserMEVBalance(ctx, &pb.QueryUserMEVBalanceRequest{
		Address: "",
	})
	require.Error(t, err)

	// Test valid address
	resp, err = queryServer.UserMEVBalance(ctx, &pb.QueryUserMEVBalanceRequest{
		Address: "aura1user",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryTokenomicsStats_Extended(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	queryServer := NewQueryServer(k)

	// Test nil request
	resp, err := queryServer.TokenomicsStats(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.TokenomicsStats(ctx, &pb.QueryTokenomicsStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
