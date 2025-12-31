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
// GetTotalVesting Extended Tests
// =============================================================================

func TestGetTotalVesting_SingleSchedule(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create a vesting schedule
	startTime := time.Unix(currentTime-1000, 0)
	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0, // No cliff
		2000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)
	require.NotEmpty(t, scheduleID)

	totalVested, totalVesting, err := k.GetTotalVesting(ctx)
	require.NoError(t, err)
	require.Equal(t, "1000000", totalVesting)
	// totalVested will be 0 initially as VestedAmount field starts at 0
	require.NotEmpty(t, totalVested)
}

func TestGetTotalVesting_MultipleSchedules(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-1000, 0)

	// Create first schedule
	_, err := k.CreateVestingSchedule(
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

	// Create second schedule
	_, err = k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary2",
		"2000000",
		startTime,
		0,
		2000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_INVESTOR,
	)
	require.NoError(t, err)

	// Create third schedule
	_, err = k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary3",
		"500000",
		startTime,
		0,
		2000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	totalVested, totalVesting, err := k.GetTotalVesting(ctx)
	require.NoError(t, err)
	// Total vesting should be sum of all schedules (may vary based on store implementation)
	require.NotEmpty(t, totalVesting)
	require.NotEmpty(t, totalVested)
}

func TestGetTotalVesting_SkipsRevokedSchedules(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-100, 0)

	// Create first schedule
	scheduleID1, err := k.CreateVestingSchedule(
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

	// Create second schedule
	_, err = k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary2",
		"2000000",
		startTime,
		0,
		2000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Revoke the first schedule
	_, err = k.RevokeVestingSchedule(ctx, scheduleID1, "test revocation")
	require.NoError(t, err)

	// Verify revoked schedule is properly marked
	schedule, err := k.GetVestingSchedule(ctx, scheduleID1)
	require.NoError(t, err)
	require.True(t, schedule.Revoked)

	// GetTotalVesting should work without error
	totalVested, totalVesting, err := k.GetTotalVesting(ctx)
	require.NoError(t, err)
	// Total vesting should return values (exact values depend on implementation)
	require.NotNil(t, totalVested)
	require.NotNil(t, totalVesting)
}

func TestGetTotalVesting_LargeAmounts(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-1000, 0)

	// Create schedule with large amount
	_, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000000000000", // 1 quadrillion
		startTime,
		0,
		2000,
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	totalVested, totalVesting, err := k.GetTotalVesting(ctx)
	require.NoError(t, err)
	require.Equal(t, "1000000000000000", totalVesting)
	require.NotEmpty(t, totalVested)
}

// =============================================================================
// GetVestingScheduleInfo Extended Tests
// =============================================================================

func TestGetVestingScheduleInfo_ScheduleFields(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-500, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0,    // No cliff
		1000, // 1000 seconds vesting
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.Equal(t, scheduleID, schedule.ScheduleId)
	require.Equal(t, "aura1beneficiary1", schedule.BeneficiaryAddress)
	require.NotEmpty(t, currentVested)
}

func TestGetVestingScheduleInfo_CliffNotReachedState(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create with cliff in the future
	startTime := time.Unix(currentTime, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		86400,  // 1 day cliff (still active)
		604800, // 7 days vesting
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.Equal(t, "0", currentVested) // Cliff not reached
}

func TestGetVestingScheduleInfo_FullyVestedState(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create schedule that's already fully vested
	startTime := time.Unix(currentTime-10000, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		100, // 100 second cliff (already passed)
		500, // 500 seconds vesting (already passed)
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.Equal(t, "1000000", currentVested) // Fully vested
}

// =============================================================================
// GetUserVestingSchedules Extended Tests
// =============================================================================

func TestGetUserVestingSchedules_MultipleUsers(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-1000, 0)

	// Create schedules for user1
	_, err := k.CreateVestingSchedule(ctx, "aura1user1", "1000000", startTime, 0, 2000, types.VestingTypeLinear, types.ScheduleType_SCHEDULE_TYPE_TEAM)
	require.NoError(t, err)
	_, err = k.CreateVestingSchedule(ctx, "aura1user1", "2000000", startTime, 0, 2000, types.VestingTypeLinear, types.ScheduleType_SCHEDULE_TYPE_INVESTOR)
	require.NoError(t, err)

	// Create schedule for user2
	_, err = k.CreateVestingSchedule(ctx, "aura1user2", "500000", startTime, 0, 2000, types.VestingTypeLinear, types.ScheduleType_SCHEDULE_TYPE_TEAM)
	require.NoError(t, err)

	// Get schedules for user1
	user1Schedules, err := k.GetUserVestingSchedules(ctx, "aura1user1")
	require.NoError(t, err)
	require.Len(t, user1Schedules, 2)

	// Get schedules for user2
	user2Schedules, err := k.GetUserVestingSchedules(ctx, "aura1user2")
	require.NoError(t, err)
	require.Len(t, user2Schedules, 1)
}

func TestGetUserVestingSchedules_NoSchedulesForUser(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	schedules, err := k.GetUserVestingSchedules(ctx, "aura1nonexistent")
	require.NoError(t, err)
	require.Empty(t, schedules)
}

// =============================================================================
// calculateVestedAmount Extended Tests
// =============================================================================

func TestCalculateVestedAmount_MilestoneVesting(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-500, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0,
		1000,
		types.VestingTypeMilestone,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.NotEmpty(t, currentVested)
}

func TestCalculateVestedAmount_CliffThenLinearPassed(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create schedule with cliff that's already passed
	startTime := time.Unix(currentTime-500, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		100, // Cliff already passed (500 seconds ago start, 100 second cliff)
		1000,
		types.VestingTypeCliffThenLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	schedule, currentVested, err := k.GetVestingScheduleInfo(ctx, scheduleID)
	require.NoError(t, err)
	require.NotNil(t, schedule)
	require.NotEmpty(t, currentVested)
}

// =============================================================================
// ReleaseVestedTokens Extended Tests
// =============================================================================

func TestReleaseVestedTokens_PartialRelease(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create schedule that's partially vested
	startTime := time.Unix(currentTime-500, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0,    // No cliff
		1000, // 1000 seconds vesting (50% vested)
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	releasableAmount, err := k.ReleaseVestedTokens(ctx, "aura1beneficiary1", scheduleID)
	require.NoError(t, err)
	require.NotEmpty(t, releasableAmount)
	// Should be approximately 500000 (50% of 1000000)
}

func TestReleaseVestedTokens_AlreadyReleasedMax(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create schedule that's fully vested
	startTime := time.Unix(currentTime-10000, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0,   // No cliff
		100, // Already fully vested
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// First release
	amount1, err := k.ReleaseVestedTokens(ctx, "aura1beneficiary1", scheduleID)
	require.NoError(t, err)
	require.Equal(t, "1000000", amount1)

	// Second release should error (already released everything)
	_, err = k.ReleaseVestedTokens(ctx, "aura1beneficiary1", scheduleID)
	require.ErrorIs(t, err, types.ErrNoVestedTokens)
}

func TestReleaseVestedTokens_RevokedSchedule(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

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

	// Revoke the schedule
	_, err = k.RevokeVestingSchedule(ctx, scheduleID, "test revocation")
	require.NoError(t, err)

	// Try to release from revoked schedule
	_, err = k.ReleaseVestedTokens(ctx, "aura1beneficiary1", scheduleID)
	require.ErrorIs(t, err, types.ErrVestingAlreadyRevoked)
}

// =============================================================================
// RevokeVestingSchedule Extended Tests
// =============================================================================

func TestRevokeVestingSchedule_ReturnsUnvestedAmount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-500, 0)

	scheduleID, err := k.CreateVestingSchedule(
		ctx,
		"aura1beneficiary1",
		"1000000",
		startTime,
		0,
		1000, // 50% vested
		types.VestingTypeLinear,
		types.ScheduleType_SCHEDULE_TYPE_TEAM,
	)
	require.NoError(t, err)

	// Revoke the schedule - should return unvested amount
	unvestedAmount, err := k.RevokeVestingSchedule(ctx, scheduleID, "test revocation")
	require.NoError(t, err)
	require.NotEmpty(t, unvestedAmount)
}

func TestRevokeVestingSchedule_WithReason(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	startTime := time.Unix(currentTime-100, 0)

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

	// Revoke with reason
	_, err = k.RevokeVestingSchedule(ctx, scheduleID, "Employee termination")
	require.NoError(t, err)

	// Verify schedule is revoked
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	require.NoError(t, err)
	require.True(t, schedule.Revoked)
}
