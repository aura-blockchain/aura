// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestCaptureMEV(t *testing.T) {
	tests := []struct {
		name        string
		mevEnabled  bool
		amount      string
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "capture MEV successfully",
			mevEnabled: true,
			amount:     "1000000",
			wantErr:    false,
		},
		{
			name:        "MEV disabled returns error",
			mevEnabled:  false,
			amount:      "1000000",
			wantErr:     true,
			expectedErr: types.ErrMEVRedistributionDisabled,
		},
		{
			name:        "invalid amount returns error",
			mevEnabled:  true,
			amount:      "invalid",
			wantErr:     true,
			expectedErr: types.ErrInvalidAmount,
		},
		{
			name:       "zero amount succeeds",
			mevEnabled: true,
			amount:     "0",
			wantErr:    false,
		},
		{
			name:       "large amount succeeds",
			mevEnabled: true,
			amount:     "999999999999999999",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			params.Mev.Enabled = tt.mevEnabled

			k, ctx := setupKeeperWithCustomParams(t, params)

			err := k.CaptureMEV(ctx, tt.amount)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)

				// Verify pending MEV increased
				pending, err := k.GetTotalMEVPending(ctx)
				require.NoError(t, err)
				require.Equal(t, tt.amount, pending)
			}
		})
	}
}

func TestCaptureMEVAccumulation(t *testing.T) {
	params := types.DefaultParams()
	params.Mev.Enabled = true
	k, ctx := setupKeeperWithCustomParams(t, params)

	// Capture multiple MEV amounts
	_ = k.CaptureMEV(ctx, "1000")
	_ = k.CaptureMEV(ctx, "2000")
	_ = k.CaptureMEV(ctx, "3000")

	pending, err := k.GetTotalMEVPending(ctx)
	require.NoError(t, err)
	require.Equal(t, "6000", pending)
}

func TestDistributeMEV(t *testing.T) {
	tests := []struct {
		name              string
		mevEnabled        bool
		pendingMEV        string
		activeUsers       []string
		userActivity      map[string]uint64
		userIRScores      map[string]uint64
		strategy          types.MEVRedistributionStrategy
		wantErr           bool
		expectedValidator string
		expectedTreasury  string
		expectedBurn      string
	}{
		{
			name:         "MEV disabled returns zero shares",
			mevEnabled:   false,
			pendingMEV:   "1000000",
			activeUsers:  []string{"user1"},
			userActivity: map[string]uint64{"user1": 100},
			userIRScores: map[string]uint64{"user1": 50},
			wantErr:      true,
		},
		{
			name:              "no pending MEV returns zeros",
			mevEnabled:        true,
			pendingMEV:        "0",
			activeUsers:       []string{"user1"},
			userActivity:      map[string]uint64{"user1": 100},
			userIRScores:      map[string]uint64{"user1": 50},
			expectedValidator: "0",
			expectedTreasury:  "0",
			expectedBurn:      "0",
		},
		{
			name:       "equal distribution strategy",
			mevEnabled: true,
			pendingMEV: "10000",
			activeUsers: []string{
				"aura1user1xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"aura1user2xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
			userActivity: map[string]uint64{},
			userIRScores: map[string]uint64{},
			strategy:     types.MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION,
		},
		{
			name:       "proportional to activity strategy",
			mevEnabled: true,
			pendingMEV: "10000",
			activeUsers: []string{
				"aura1user1xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"aura1user2xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
			userActivity: map[string]uint64{
				"aura1user1xxxxxxxxxxxxxxxxxxxxxxxxxxxxx": 75,
				"aura1user2xxxxxxxxxxxxxxxxxxxxxxxxxxxxx": 25,
			},
			userIRScores: map[string]uint64{},
			strategy:     types.MEVRedistributionStrategy_MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY,
		},
		{
			name:       "IR weighted strategy",
			mevEnabled: true,
			pendingMEV: "10000",
			activeUsers: []string{
				"aura1user1xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"aura1user2xxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
			userActivity: map[string]uint64{},
			userIRScores: map[string]uint64{
				"aura1user1xxxxxxxxxxxxxxxxxxxxxxxxxxxxx": 80,
				"aura1user2xxxxxxxxxxxxxxxxxxxxxxxxxxxxx": 20,
			},
			strategy: types.MEVRedistributionStrategy_MEV_STRATEGY_IR_WEIGHTED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			params.Mev.Enabled = tt.mevEnabled
			if tt.strategy != 0 {
				params.Mev.Strategy = tt.strategy
			}

			k, ctx := setupKeeperWithCustomParams(t, params)

			// Set pending MEV
			if tt.pendingMEV != "0" {
				_ = k.SetTotalMEVPending(ctx, tt.pendingMEV)
			}

			validatorShare, treasuryShare, burnShare, err := k.DistributeMEV(
				ctx,
				tt.activeUsers,
				tt.userActivity,
				tt.userIRScores,
			)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expectedValidator != "" {
					require.Equal(t, tt.expectedValidator, validatorShare)
				}
				if tt.expectedTreasury != "" {
					require.Equal(t, tt.expectedTreasury, treasuryShare)
				}
				if tt.expectedBurn != "" {
					require.Equal(t, tt.expectedBurn, burnShare)
				}
			}
		})
	}
}

func TestClaimMEVRewards(t *testing.T) {
	tests := []struct {
		name           string
		address        string
		initialBalance string
		wantErr        bool
		expectedClaim  string
	}{
		{
			name:           "claim with balance",
			address:        "aura1claimer",
			initialBalance: "5000",
			wantErr:        false,
			expectedClaim:  "5000",
		},
		{
			name:           "claim with zero balance fails",
			address:        "aura1nobalance",
			initialBalance: "0",
			wantErr:        true,
		},
		{
			name:           "claim with empty balance fails",
			address:        "aura1emptyuser",
			initialBalance: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)

			// Set initial balance
			if tt.initialBalance != "" && tt.initialBalance != "0" {
				_ = k.SetUserMEVBalance(ctx, tt.address, tt.initialBalance)
			}

			claimed, err := k.ClaimMEVRewards(ctx, tt.address)

			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, types.ErrInsufficientMEVBalance)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedClaim, claimed)

				// Verify balance is now zero
				balance, _ := k.GetUserMEVBalance(ctx, tt.address)
				require.Equal(t, "0", balance)
			}
		})
	}
}

func TestGetMEVStats(t *testing.T) {
	params := types.DefaultParams()
	params.Mev.Enabled = true
	params.Mev.TotalMevCaptured = "1000000"
	params.Mev.TotalMevRedistributed = "500000"
	params.Mev.UserRedistributionPercentage = 5000 // 50%
	params.Mev.Strategy = types.MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetTotalMEVPending(ctx, "100000")

	enabled, captured, redistributed, pending, userPercentage, strategy := k.GetMEVStats(ctx)

	require.True(t, enabled)
	require.Equal(t, "1000000", captured)
	require.Equal(t, "500000", redistributed)
	require.Equal(t, "100000", pending)
	require.Equal(t, uint64(5000), userPercentage)
	require.Equal(t, types.MEVRedistributionStrategy_MEV_STRATEGY_EQUAL_DISTRIBUTION, strategy)
}

func TestGetMEVStatsDisabled(t *testing.T) {
	params := types.DefaultParams()
	params.Mev.Enabled = false

	k, ctx := setupKeeperWithCustomParams(t, params)

	enabled, captured, redistributed, pending, userPercentage, strategy := k.GetMEVStats(ctx)

	require.False(t, enabled)
	require.Equal(t, "0", captured)
	require.Equal(t, "0", redistributed)
	require.Equal(t, "0", pending)
	require.Equal(t, uint64(0), userPercentage)
	require.Equal(t, types.MEVRedistributionStrategy_MEV_STRATEGY_UNSPECIFIED, strategy)
}

func TestAddUserMEVBalance(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	user := "aura1testuser"

	// Initial balance should be 0
	balance, err := k.GetUserMEVBalance(ctx, user)
	require.NoError(t, err)
	require.Equal(t, "0", balance)

	// Add balance
	err = k.SetUserMEVBalance(ctx, user, "1000")
	require.NoError(t, err)

	balance, err = k.GetUserMEVBalance(ctx, user)
	require.NoError(t, err)
	require.Equal(t, "1000", balance)

	// Add more
	err = k.SetUserMEVBalance(ctx, user, "2500")
	require.NoError(t, err)

	balance, err = k.GetUserMEVBalance(ctx, user)
	require.NoError(t, err)
	require.Equal(t, "2500", balance)
}

func TestGetAllUserMEVBalances(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set balances for multiple users
	users := map[string]string{
		"aura1user1xxxxxxxxxxxxxxxxxxxxxxxxxxxxx": "1000",
		"aura1user2xxxxxxxxxxxxxxxxxxxxxxxxxxxxx": "2000",
		"aura1user3xxxxxxxxxxxxxxxxxxxxxxxxxxxxx": "3000",
	}

	for user, balance := range users {
		_ = k.SetUserMEVBalance(ctx, user, balance)
	}

	balances, err := k.GetAllUserMEVBalances(ctx)
	require.NoError(t, err)
	require.Len(t, balances, 3)

	for user, expectedBalance := range users {
		require.Equal(t, expectedBalance, balances[user])
	}
}
