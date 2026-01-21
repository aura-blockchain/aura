// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// Vesting Schedule Tests
// =============================================================================

func TestCreateVestingSchedule_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)

	startTime := time.Unix(500, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		86400,  // 1 day cliff
		604800, // 7 days vesting
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)

	require.NoError(t, err)
	require.NotEmpty(t, scheduleID)
	require.Contains(t, scheduleID, "vs:")
}

func TestCreateVestingSchedule_EmptyBeneficiary(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)

	startTime := time.Unix(500, 0)
	_, err := k.CreateVestingSchedule(
		ctx,
		"", // Empty beneficiary
		"10000000",
		startTime,
		86400,
		604800,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)

	require.ErrorIs(t, err, types.ErrInvalidBeneficiary)
}

func TestCreateVestingSchedule_InvalidAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)

	startTime := time.Unix(500, 0)
	_, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"not-a-number",
		startTime,
		86400,
		604800,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)

	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestCreateVestingSchedule_ZeroAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)

	startTime := time.Unix(500, 0)
	_, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"0",
		startTime,
		86400,
		604800,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)

	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestCreateVestingSchedule_NegativeAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)

	startTime := time.Unix(500, 0)
	_, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"-1000",
		startTime,
		86400,
		604800,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)

	require.ErrorIs(t, err, types.ErrInvalidAmount)
}

func TestCreateVestingSchedule_ZeroVestingDuration(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 1000)

	startTime := time.Unix(500, 0)
	_, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		86400,
		0, // Zero vesting duration
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)

	require.ErrorIs(t, err, types.ErrInvalidDuration)
}

func TestReleaseVestedTokens_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		100,  // cliff at 200
		1000, // full vest at 1100
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Move time past cliff and partially through vesting
	_ = k.SetCurrentTime(ctx, startTimeUnix+500) // 50% through vesting

	released, err := k.ReleaseVestedTokens(ctx, "aura1beneficiary1", scheduleID)
	require.NoError(t, err)
	require.NotEmpty(t, released)
	require.NotEqual(t, "0", released)
}

func TestReleaseVestedTokens_WrongBeneficiary(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 100)

	startTime := time.Unix(100, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		100,
		1000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Wrong beneficiary tries to release
	_, err = k.ReleaseVestedTokens(ctx, "aura1wrongperson", scheduleID)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestReleaseVestedTokens_BeforeCliff(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		1000, // cliff at 1100
		2000, // full vest at 2100
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Try to release before cliff
	_ = k.SetCurrentTime(ctx, startTimeUnix+500) // before cliff

	_, err = k.ReleaseVestedTokens(ctx, "aura1beneficiary1", scheduleID)
	require.ErrorIs(t, err, types.ErrCliffNotReached)
}

func TestReleaseVestedTokens_NotFound(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.ReleaseVestedTokens(ctx, "aura1beneficiary1", "nonexistent-schedule")
	require.Error(t, err)
}

func TestRevokeVestingSchedule_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		100,  // cliff at 200
		1000, // full vest at 1100
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Move past cliff for partial vesting
	_ = k.SetCurrentTime(ctx, startTimeUnix+500)

	unvested, err := k.RevokeVestingSchedule(ctx, scheduleID, "team member left")
	require.NoError(t, err)
	require.NotEmpty(t, unvested)

	// Verify schedule is marked as revoked
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	require.NoError(t, err)
	require.True(t, schedule.Revoked)
	require.Equal(t, "team member left", schedule.RevokedReason)
}

func TestRevokeVestingSchedule_AlreadyRevoked(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		100,
		1000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Revoke first time
	_ = k.SetCurrentTime(ctx, startTimeUnix+200)
	_, err = k.RevokeVestingSchedule(ctx, scheduleID, "first revoke")
	require.NoError(t, err)

	// Try to revoke again
	_, err = k.RevokeVestingSchedule(ctx, scheduleID, "second revoke")
	require.ErrorIs(t, err, types.ErrVestingAlreadyRevoked)
}

func TestRevokeVestingSchedule_NotFound(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, err := k.RevokeVestingSchedule(ctx, "nonexistent-schedule", "reason")
	require.Error(t, err)
}

func TestGetUserVestingSchedules_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	schedules, err := k.GetUserVestingSchedules(ctx, "aura1userwithnoschedules")
	require.NoError(t, err)
	require.Empty(t, schedules)
}

func TestGetUserVestingSchedules_WithSchedules(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	_ = k.SetCurrentTime(ctx, 100)

	startTime := time.Unix(100, 0)
	beneficiary := "aura1beneficiary1"

	// Create multiple schedules for the same user
	_, err := k.CreateVestingSchedule(ctx, beneficiary, "10000000", startTime, 100, 1000, types.VestingTypeLinear, types.ScheduleType_SCHEDULE_TYPE_TEAM)
	require.NoError(t, err)

	_ = k.SetCurrentTime(ctx, 101) // Slightly different time for unique ID
	_, err = k.CreateVestingSchedule(ctx, beneficiary, "5000000", startTime, 200, 2000, types.VestingTypeLinear, types.ScheduleType_SCHEDULE_TYPE_INVESTOR)
	require.NoError(t, err)

	schedules, err := k.GetUserVestingSchedules(ctx, beneficiary)
	require.NoError(t, err)
	require.Len(t, schedules, 2)
}

func TestCalculateVestedAmount_Linear_BeforeCliff(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)

	schedule := &types.VestingSchedule{
		ScheduleId:      "vs:test",
		TotalAmount:     "10000000",
		VestedAmount:    "0",
		StartTime:       time.Unix(startTimeUnix, 0),
		CliffDuration:   500, // cliff at 600
		VestingDuration: 1000,
		VestingType:     types.VestingTypeLinear,
		Revoked:         false,
	}

	// Set time before cliff
	_ = k.SetCurrentTime(ctx, startTimeUnix+400)

	_, err := k.calculateVestedAmount(ctx, schedule)
	require.ErrorIs(t, err, types.ErrCliffNotReached)
}

func TestCalculateVestedAmount_Linear_AfterCliff(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)

	schedule := &types.VestingSchedule{
		ScheduleId:      "vs:test",
		TotalAmount:     "10000000",
		VestedAmount:    "0",
		StartTime:       time.Unix(startTimeUnix, 0),
		CliffDuration:   100, // cliff at 200
		VestingDuration: 1000,
		VestingType:     types.VestingTypeLinear,
		Revoked:         false,
	}

	// Set time to 50% through vesting
	_ = k.SetCurrentTime(ctx, startTimeUnix+500)

	vested, err := k.calculateVestedAmount(ctx, schedule)
	require.NoError(t, err)
	require.Equal(t, "5000000", vested) // 50% of 10000000
}

func TestCalculateVestedAmount_Linear_FullyVested(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)

	schedule := &types.VestingSchedule{
		ScheduleId:      "vs:test",
		TotalAmount:     "10000000",
		VestedAmount:    "0",
		StartTime:       time.Unix(startTimeUnix, 0),
		CliffDuration:   100,
		VestingDuration: 1000,
		VestingType:     types.VestingTypeLinear,
		Revoked:         false,
	}

	// Set time past full vesting
	_ = k.SetCurrentTime(ctx, startTimeUnix+2000)

	vested, err := k.calculateVestedAmount(ctx, schedule)
	require.NoError(t, err)
	require.Equal(t, "10000000", vested) // Fully vested
}

func TestCalculateVestedAmount_CliffThenLinear(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)

	schedule := &types.VestingSchedule{
		ScheduleId:      "vs:test",
		TotalAmount:     "10000000",
		VestedAmount:    "0",
		StartTime:       time.Unix(startTimeUnix, 0),
		CliffDuration:   200, // cliff at 300
		VestingDuration: 1000,
		VestingType:     types.VestingTypeCliffThenLinear,
		Revoked:         false,
	}

	// Set time to halfway between cliff and end
	// cliffEnd = 300, vestingEnd = 1100, midpoint = 700
	_ = k.SetCurrentTime(ctx, 700)

	vested, err := k.calculateVestedAmount(ctx, schedule)
	require.NoError(t, err)
	require.NotEmpty(t, vested)
}

func TestCalculateVestedAmount_Milestone(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)

	schedule := &types.VestingSchedule{
		ScheduleId:      "vs:test",
		TotalAmount:     "10000000",
		VestedAmount:    "0",
		StartTime:       time.Unix(startTimeUnix, 0),
		CliffDuration:   100,
		VestingDuration: 1000,
		VestingType:     types.VestingTypeMilestone,
		Revoked:         false,
	}

	// Set time to 50% through
	_ = k.SetCurrentTime(ctx, startTimeUnix+500)

	vested, err := k.calculateVestedAmount(ctx, schedule)
	require.NoError(t, err)
	require.NotEmpty(t, vested)
}

func TestCalculateVestedAmount_Revoked(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	schedule := &types.VestingSchedule{
		ScheduleId:   "vs:test",
		TotalAmount:  "10000000",
		VestedAmount: "3000000", // Was revoked with 3M vested
		Revoked:      true,
	}

	vested, err := k.calculateVestedAmount(ctx, schedule)
	require.NoError(t, err)
	require.Equal(t, "3000000", vested) // Returns vested amount at revocation
}

func TestGetTotalVesting_Empty(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	totalVested, totalVesting, err := k.GetTotalVesting(ctx)
	require.NoError(t, err)
	require.Equal(t, "0", totalVested)
	require.Equal(t, "0", totalVesting)
}

func TestGetVestingScheduleInfo_Success(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		100,
		1000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Move past cliff
	_ = k.SetCurrentTime(ctx, startTimeUnix+500)

	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.NotEmpty(t, currentVested)
}

func TestGetVestingScheduleInfo_BeforeCliff(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		500, // cliff at 600
		1000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Stay before cliff
	_ = k.SetCurrentTime(ctx, startTimeUnix+200)

	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.Equal(t, "0", currentVested) // Before cliff, returns 0 without error
}

func TestGetVestingScheduleInfo_NotFound(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	_, _, err := k.GetVestingScheduleInfo(ctx, "nonexistent-schedule")
	require.Error(t, err)
}

func TestVestingSchedule_FullWorkflow(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	beneficiary := "aura1beneficiary1"

	// Create schedule
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		beneficiary,
		"10000000",
		startTime,
		100,  // cliff at 200
		1000, // full vest at 1100
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Verify schedule was created
	schedules, err := k.GetUserVestingSchedules(ctx, beneficiary)
	require.NoError(t, err)
	require.Len(t, schedules, 1)

	// Move past cliff (to 50% vesting)
	_ = k.SetCurrentTime(ctx, startTimeUnix+500)

	// Release vested tokens
	released, err := k.ReleaseVestedTokens(ctx, beneficiary, scheduleID)
	require.NoError(t, err)
	require.Equal(t, "5000000", released) // 50%

	// Check schedule info
	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.Equal(t, "5000000", schedule.VestedAmount)
	require.Equal(t, "5000000", currentVested)

	// Try to release again immediately - should get 0
	released, err = k.ReleaseVestedTokens(ctx, beneficiary, scheduleID)
	require.ErrorIs(t, err, types.ErrNoVestedTokens)
	require.Equal(t, "0", released)

	// Move to 75% vesting
	_ = k.SetCurrentTime(ctx, startTimeUnix+750)

	// Release more
	released, err = k.ReleaseVestedTokens(ctx, beneficiary, scheduleID)
	require.NoError(t, err)
	require.Equal(t, "2500000", released) // 75% - 50% = 25%

	// Verify updated vested amount
	schedule, _, _ = k.GetVestingScheduleInfo(ctx, scheduleID)
	require.Equal(t, "7500000", schedule.VestedAmount)
}

func TestVestingSchedule_RevokeWithPartialVesting(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	startTimeUnix := int64(100)
	_ = k.SetCurrentTime(ctx, startTimeUnix)

	startTime := time.Unix(startTimeUnix, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"10000000",
		startTime,
		100,  // cliff at 200
		1000, // full vest at 1100
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Move to 30% vesting
	_ = k.SetCurrentTime(ctx, startTimeUnix+300)

	// Revoke schedule
	unvested, err := k.RevokeVestingSchedule(ctx, scheduleID, "employee termination")
	require.NoError(t, err)
	require.Equal(t, "7000000", unvested) // 70% unvested

	// Verify schedule is revoked
	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.True(t, schedule.Revoked)
	require.Equal(t, "3000000", currentVested) // 30% was vested at revocation

	// Try to release after revocation - should fail
	_, err = k.ReleaseVestedTokens(ctx, "aura1beneficiary1", scheduleID)
	require.ErrorIs(t, err, types.ErrVestingAlreadyRevoked)
}

func TestCreateVestingSchedule_DifferentScheduleTypes(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	scheduleTypes := []types.ScheduleType{
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
		types.ScheduleType_SCHEDULE_TYPE_INVESTOR,
		types.ScheduleType_SCHEDULE_TYPE_ADVISOR,
		types.ScheduleType_SCHEDULE_TYPE_ECOSYSTEM,
		types.ScheduleType_SCHEDULE_TYPE_COMMUNITY,
	}

	startTime := time.Unix(100, 0)

	for i, st := range scheduleTypes {
		_ = k.SetCurrentTime(ctx, int64(100+i))

		scheduleID, err := k.CreateVestingSchedule(
			ctx,
			"aura1beneficiary1",
			"10000000",
			startTime,
			100,
			1000,
			types.VestingTypeLinear,
			st,
		)
		require.NoError(t, err)
		require.NotEmpty(t, scheduleID)

		schedule, err := k.GetVestingSchedule(ctx, scheduleID)
		require.NoError(t, err)
		require.Equal(t, st, schedule.ScheduleType)
	}
}
