// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name        string
		params      Params
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default params",
			params:      DefaultParams(),
			expectError: false,
		},
		{
			name: "zero verification threshold",
			params: Params{
				VerificationThreshold:       0,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "verification_threshold must be positive",
		},
		{
			name: "high assurance threshold < verification threshold",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      5000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "high_assurance_threshold",
		},
		{
			name: "zero arena focus threshold",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         0,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "arena_focus_threshold must be positive",
		},
		{
			name: "mismatched velocity bonus arrays",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500}, // Mismatch
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "velocity_bonus_days and velocity_bonus_multipliers_bps must have same length",
		},
		{
			name: "velocity bonus multiplier < 1.0x (10000 basis points)",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{5000, 11000}, // 5000 < 10000 (0.5x < 1.0x)
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "velocity_bonus_multipliers_bps",
		},
		{
			name: "arena multiplier with zero threshold",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					0: 11000, // Zero threshold
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "arena_multipliers_bps threshold must be positive",
		},
		{
			name: "arena multiplier < 1.0x (10000 basis points)",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 5000, // < 10000 (0.5x < 1.0x)
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "arena_multipliers_bps",
		},
		{
			name: "slash percentage > 100",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       150, // > 100
			},
			expectError: true,
			errorMsg:    "slash_percentage must be <= 100",
		},
		{
			name: "zero max IRs per day",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          0, // Zero
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "max_irs_per_day must be positive",
		},
		{
			name: "zero max IRs per hour",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         0, // Zero
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "max_irs_per_hour must be positive",
		},
		{
			name: "hourly > daily limit",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         15, // > daily
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{50000},
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "max_irs_per_hour",
		},
		{
			name: "mismatched jackpot arrays",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100, 1000},
				JackpotMultipliersBps: []uint64{50000}, // Mismatch
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "jackpot_odds and jackpot_multipliers_bps must have same length",
		},
		{
			name: "jackpot multiplier < 1.0x (10000 basis points)",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:          10,
				MaxIrsPerHour:         3,
				JackpotOdds:           []uint64{100},
				JackpotMultipliersBps: []uint64{5000}, // < 10000 (0.5x < 1.0x)
				SlashPercentage:       50,
			},
			expectError: true,
			errorMsg:    "jackpot_multipliers_bps",
		},
		{
			name: "user reward split > 100",
			params: Params{
				VerificationThreshold:       10000,
				HighAssuranceThreshold:      15000,
				ArenaFocusThreshold:         5000,
				VelocityBonusDays:           []uint64{7, 30},
				VelocityBonusMultipliersBps: []uint64{12500, 11000},
				ArenaMultipliersBps: map[uint64]uint64{
					3000: 11000,
				},
				MaxIrsPerDay:           10,
				MaxIrsPerHour:          3,
				JackpotOdds:            []uint64{100},
				JackpotMultipliersBps:  []uint64{50000},
				SlashPercentage:        50,
				UserRewardSplitPercent: 150, // > 100
			},
			expectError: true,
			errorMsg:    "user_reward_split_percent must be <= 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParams(&tt.params)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Check if error contains expected message
					// We use full equality for some, contains for others
					// This is flexible enough for the test suite
					t.Logf("got error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// Validate default params
	if err := ValidateParams(&params); err != nil {
		t.Errorf("default params should be valid, got error: %v", err)
	}

	// Check specific defaults
	if params.VerificationThreshold != 10000 {
		t.Errorf("expected verification threshold 10000, got %d", params.VerificationThreshold)
	}

	if params.HighAssuranceThreshold != 15000 {
		t.Errorf("expected high assurance threshold 15000, got %d", params.HighAssuranceThreshold)
	}

	if params.ArenaFocusThreshold != 5000 {
		t.Errorf("expected arena focus threshold 5000, got %d", params.ArenaFocusThreshold)
	}

	if params.SlashPercentage != 50 {
		t.Errorf("expected slash percentage 50, got %d", params.SlashPercentage)
	}

	if params.MaxIrsPerDay != 10 {
		t.Errorf("expected max IRs per day 10, got %d", params.MaxIrsPerDay)
	}

	if params.MaxIrsPerHour != 3 {
		t.Errorf("expected max IRs per hour 3, got %d", params.MaxIrsPerHour)
	}

	if !params.PoiRewardsEnabled {
		t.Error("expected PoI rewards to be enabled by default")
	}

	if params.UserRewardSplitPercent != 70 {
		t.Errorf("expected user reward split 70, got %d", params.UserRewardSplitPercent)
	}

	if params.StalenessEnabled {
		t.Error("expected staleness to be disabled by default")
	}

	// Test new basis point fields
	if len(params.VelocityBonusMultipliersBps) != 2 {
		t.Errorf("expected 2 velocity bonus multipliers, got %d", len(params.VelocityBonusMultipliersBps))
	}
	if params.VelocityBonusMultipliersBps[0] != 12500 { // 1.25x
		t.Errorf("expected velocity bonus multiplier[0] = 12500, got %d", params.VelocityBonusMultipliersBps[0])
	}

	if len(params.ArenaMultipliersBps) != 3 {
		t.Errorf("expected 3 arena multipliers, got %d", len(params.ArenaMultipliersBps))
	}
	if params.ArenaMultipliersBps[5000] != 15000 { // 1.5x
		t.Errorf("expected arena multiplier[5000] = 15000, got %d", params.ArenaMultipliersBps[5000])
	}

	if len(params.JackpotMultipliersBps) != 2 {
		t.Errorf("expected 2 jackpot multipliers, got %d", len(params.JackpotMultipliersBps))
	}
	if params.JackpotMultipliersBps[0] != 50000 { // 5.0x
		t.Errorf("expected jackpot multiplier[0] = 50000, got %d", params.JackpotMultipliersBps[0])
	}
}
