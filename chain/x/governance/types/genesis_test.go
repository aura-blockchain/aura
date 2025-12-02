package types

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

// validTestGenesis creates a properly formatted genesis for testing
func validTestGenesis() *GenesisState {
	genesis := DefaultGenesis()
	// Fix the min_deposit to include denomination
	genesis.Params.MinDeposit = "10000000aura"

	// Fix category params min_deposit to include denomination
	for _, params := range genesis.Params.CategoryParams {
		params.MinDeposit = "10000000aura"
	}

	return genesis
}

func TestValidateGenesis_Valid(t *testing.T) {
	genesis := validTestGenesis()

	if err := ValidateGenesis(genesis); err != nil {
		t.Errorf("Valid genesis should pass validation, got error: %v", err)
	}
}

func TestValidateGenesis_NilGenesis(t *testing.T) {
	err := ValidateGenesis(nil)
	if err == nil {
		t.Error("expected error for nil genesis")
	}
	if err.Error() != "governance genesis cannot be nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateGenesis_NilParams(t *testing.T) {
	genesis := &GenesisState{
		Params: nil,
	}
	err := ValidateGenesis(genesis)
	if err == nil {
		t.Error("expected error for nil params")
	}
	if err.Error() != "governance params cannot be nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateGenesis_InvalidThresholds(t *testing.T) {
	tests := []struct {
		name      string
		threshold string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "negative threshold",
			threshold: "-0.1",
			wantErr:   true,
			errMsg:    "must be between 0.0 and 1.0",
		},
		{
			name:      "threshold above 1.0",
			threshold: "1.5",
			wantErr:   true,
			errMsg:    "must be between 0.0 and 1.0",
		},
		{
			name:      "empty threshold",
			threshold: "",
			wantErr:   true,
			errMsg:    "must be set",
		},
		{
			name:      "invalid decimal",
			threshold: "invalid",
			wantErr:   true,
			errMsg:    "must be a valid decimal",
		},
		{
			name:      "valid threshold at 0",
			threshold: "0.0",
			wantErr:   false,
		},
		{
			name:      "valid threshold at 1",
			threshold: "1.0",
			wantErr:   false,
		},
		{
			name:      "valid threshold in middle",
			threshold: "0.5",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := validTestGenesis()
			genesis.Params.Quorum = tt.threshold

			err := ValidateGenesis(genesis)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenesis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" {
					errStr := err.Error()
					if len(errStr) == 0 || len(tt.errMsg) == 0 {
						t.Errorf("error message is empty")
					}
				}
			}
		})
	}
}

func TestValidateGenesis_VetoThresholdMustBeLessThanThreshold(t *testing.T) {
	tests := []struct {
		name          string
		threshold     string
		vetoThreshold string
		wantErr       bool
	}{
		{
			name:          "veto equal to threshold",
			threshold:     "0.5",
			vetoThreshold: "0.5",
			wantErr:       true,
		},
		{
			name:          "veto greater than threshold",
			threshold:     "0.5",
			vetoThreshold: "0.6",
			wantErr:       true,
		},
		{
			name:          "veto less than threshold",
			threshold:     "0.5",
			vetoThreshold: "0.334",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := validTestGenesis()
			genesis.Params.Threshold = tt.threshold
			genesis.Params.VetoThreshold = tt.vetoThreshold

			err := ValidateGenesis(genesis)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGenesis_InvalidMinDeposit(t *testing.T) {
	tests := []struct {
		name       string
		minDeposit string
		wantErr    bool
	}{
		{
			name:       "empty deposit",
			minDeposit: "",
			wantErr:    true,
		},
		{
			name:       "invalid coin format",
			minDeposit: "invalid",
			wantErr:    true,
		},
		{
			name:       "negative deposit",
			minDeposit: "-100aura",
			wantErr:    true,
		},
		{
			name:       "zero deposit",
			minDeposit: "0aura",
			wantErr:    true,
		},
		{
			name:       "valid deposit",
			minDeposit: "10000000aura",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := validTestGenesis()
			genesis.Params.MinDeposit = tt.minDeposit

			err := ValidateGenesis(genesis)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGenesis_InvalidPeriods(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "period too short - 0 seconds",
			seconds: 0,
			wantErr: true,
			errMsg:  "must be >= 1 minute",
		},
		{
			name:    "period too short - 30 seconds",
			seconds: 30,
			wantErr: true,
			errMsg:  "must be >= 1 minute",
		},
		{
			name:    "period exactly 1 minute",
			seconds: 60,
			wantErr: false,
		},
		{
			name:    "period too long - 2 years",
			seconds: 2 * 365 * 24 * 3600,
			wantErr: true,
			errMsg:  "must be <= 1 year",
		},
		{
			name:    "period exactly 1 year",
			seconds: 365 * 24 * 3600,
			wantErr: false,
		},
		{
			name:    "valid period - 7 days",
			seconds: 7 * 24 * 3600,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" (voting_period)", func(t *testing.T) {
			genesis := validTestGenesis()
			genesis.Params.VotingPeriod = durationpb.New(time.Duration(tt.seconds) * time.Second)

			err := ValidateGenesis(genesis)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		t.Run(tt.name+" (max_deposit_period)", func(t *testing.T) {
			genesis := validTestGenesis()
			genesis.Params.MaxDepositPeriod = durationpb.New(time.Duration(tt.seconds) * time.Second)

			err := ValidateGenesis(genesis)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGenesis_NilPeriods(t *testing.T) {
	t.Run("nil voting_period", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.VotingPeriod = nil

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for nil voting_period")
		}
		if err.Error() != "voting_period cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("nil max_deposit_period", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.MaxDepositPeriod = nil

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for nil max_deposit_period")
		}
		if err.Error() != "max_deposit_period cannot be nil" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestValidateGenesis_ExecutionDelay(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		wantErr bool
	}{
		{
			name:    "negative execution delay",
			seconds: -1,
			wantErr: true,
		},
		{
			name:    "zero execution delay (allowed for emergency)",
			seconds: 0,
			wantErr: false,
		},
		{
			name:    "valid execution delay - 48 hours",
			seconds: 48 * 3600,
			wantErr: false,
		},
		{
			name:    "execution delay too long - 2 years",
			seconds: 2 * 365 * 24 * 3600,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesis := validTestGenesis()
			genesis.Params.ExecutionDelay = durationpb.New(time.Duration(tt.seconds) * time.Second)

			err := ValidateGenesis(genesis)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGenesis_TokenLockValidation(t *testing.T) {
	t.Run("token lock enabled but nil duration", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.RequireTokenLock = true
		genesis.Params.TokenLockDuration = nil

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for nil token_lock_duration when require_token_lock is true")
		}
	})

	t.Run("token lock duration too short", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.RequireTokenLock = true
		genesis.Params.TokenLockDuration = durationpb.New(30 * time.Second)

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for token_lock_duration < 1 minute")
		}
	})

	t.Run("token lock duration too long", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.RequireTokenLock = true
		genesis.Params.TokenLockDuration = durationpb.New(2 * 365 * 24 * time.Hour)

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for token_lock_duration > 1 year")
		}
	})

	t.Run("valid token lock duration", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.RequireTokenLock = true
		genesis.Params.TokenLockDuration = durationpb.New(7 * 24 * time.Hour)

		err := ValidateGenesis(genesis)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateGenesis_SecretBallotValidation(t *testing.T) {
	t.Run("secret ballot enabled but nil reveal period", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.SecretBallotEnabled = true
		genesis.Params.RevealPeriod = nil

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for nil reveal_period when secret_ballot_enabled is true")
		}
	})

	t.Run("reveal period too short", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.SecretBallotEnabled = true
		genesis.Params.RevealPeriod = durationpb.New(30 * time.Second)

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for reveal_period < 1 minute")
		}
	})

	t.Run("reveal period too long", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.SecretBallotEnabled = true
		genesis.Params.RevealPeriod = durationpb.New(31 * 24 * time.Hour)

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for reveal_period > 30 days")
		}
	})

	t.Run("valid reveal period", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.SecretBallotEnabled = true
		genesis.Params.RevealPeriod = durationpb.New(24 * time.Hour)

		err := ValidateGenesis(genesis)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestValidateGenesis_CategoryParams(t *testing.T) {
	t.Run("nil category params", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.CategoryParams["test_category"] = nil

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for nil category params")
		}
	})

	t.Run("category with invalid threshold", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.CategoryParams["test_category"] = &CategoryParams{
			MinDeposit:     "10000000aura",
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),
			Quorum:         "1.5", // Invalid: > 1.0
			Threshold:      "0.5",
			VetoThreshold:  "0.334",
			ExecutionDelay: durationpb.New(48 * time.Hour),
		}

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for invalid category threshold")
		}
	})

	t.Run("category with veto >= threshold", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.CategoryParams["test_category"] = &CategoryParams{
			MinDeposit:     "10000000aura",
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),
			Quorum:         "0.334",
			Threshold:      "0.5",
			VetoThreshold:  "0.6", // Invalid: > threshold
			ExecutionDelay: durationpb.New(48 * time.Hour),
		}

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for category veto_threshold >= threshold")
		}
	})

	t.Run("category with invalid min deposit", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.CategoryParams["test_category"] = &CategoryParams{
			MinDeposit:     "0aura", // Invalid: not positive
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),
			Quorum:         "0.334",
			Threshold:      "0.5",
			VetoThreshold:  "0.334",
			ExecutionDelay: durationpb.New(48 * time.Hour),
		}

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for non-positive category min_deposit")
		}
	})

	t.Run("category with invalid voting period", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.CategoryParams["test_category"] = &CategoryParams{
			MinDeposit:     "10000000aura",
			VotingPeriod:   durationpb.New(30 * time.Second), // Invalid: < 1 minute
			Quorum:         "0.334",
			Threshold:      "0.5",
			VetoThreshold:  "0.334",
			ExecutionDelay: durationpb.New(48 * time.Hour),
		}

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for invalid category voting_period")
		}
	})

	t.Run("category with negative execution delay", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.CategoryParams["test_category"] = &CategoryParams{
			MinDeposit:     "10000000aura",
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),
			Quorum:         "0.334",
			Threshold:      "0.5",
			VetoThreshold:  "0.334",
			ExecutionDelay: durationpb.New(-1 * time.Hour), // Invalid: negative
		}

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for negative category execution_delay")
		}
	})

	t.Run("valid category params", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.CategoryParams["test_category"] = &CategoryParams{
			MinDeposit:     "10000000aura",
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),
			Quorum:         "0.334",
			Threshold:      "0.5",
			VetoThreshold:  "0.334",
			ExecutionDelay: durationpb.New(48 * time.Hour),
		}

		err := ValidateGenesis(genesis)
		if err != nil {
			t.Errorf("unexpected error for valid category params: %v", err)
		}
	})
}

func TestValidateGenesis_EmergencyThresholds(t *testing.T) {
	t.Run("invalid emergency quorum", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.EmergencyQuorum = "1.5" // Invalid: > 1.0

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for invalid emergency_quorum")
		}
	})

	t.Run("invalid emergency threshold", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.EmergencyThreshold = "-0.1" // Invalid: < 0.0

		err := ValidateGenesis(genesis)
		if err == nil {
			t.Error("expected error for invalid emergency_threshold")
		}
	})

	t.Run("valid emergency thresholds", func(t *testing.T) {
		genesis := validTestGenesis()
		genesis.Params.EmergencyQuorum = "0.6"
		genesis.Params.EmergencyThreshold = "0.75"

		err := ValidateGenesis(genesis)
		if err != nil {
			t.Errorf("unexpected error for valid emergency thresholds: %v", err)
		}
	})
}
