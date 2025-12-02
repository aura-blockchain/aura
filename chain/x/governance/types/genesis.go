package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	pb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

// GenesisState re-exports the protobuf governance genesis definition.
type GenesisState = pb.GenesisState

// DefaultGenesis returns default governance genesis state.
func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	return &pb.GenesisState{Params: params}
}

// ValidateGenesis validates the governance genesis configuration.
func ValidateGenesis(g *GenesisState) error {
	if g == nil {
		return fmt.Errorf("governance genesis cannot be nil")
	}
	if g.Params == nil {
		return fmt.Errorf("governance params cannot be nil")
	}

	// Validate main thresholds (percentage values between 0.0 and 1.0)
	if err := validateThreshold("quorum", g.Params.Quorum); err != nil {
		return err
	}
	if err := validateThreshold("threshold", g.Params.Threshold); err != nil {
		return err
	}
	if err := validateThreshold("veto_threshold", g.Params.VetoThreshold); err != nil {
		return err
	}

	// Parse thresholds for logical consistency checks
	threshold, _ := sdkmath.LegacyNewDecFromStr(g.Params.Threshold)
	vetoThreshold, _ := sdkmath.LegacyNewDecFromStr(g.Params.VetoThreshold)

	// Logical consistency: veto threshold must be less than pass threshold
	if vetoThreshold.GTE(threshold) {
		return fmt.Errorf("veto_threshold (%s) must be < threshold (%s)", vetoThreshold, threshold)
	}

	// Validate emergency thresholds
	if g.Params.EmergencyQuorum != "" {
		if err := validateThreshold("emergency_quorum", g.Params.EmergencyQuorum); err != nil {
			return err
		}
	}
	if g.Params.EmergencyThreshold != "" {
		if err := validateThreshold("emergency_threshold", g.Params.EmergencyThreshold); err != nil {
			return err
		}
	}

	// Validate deposit amount
	minDeposit, err := sdk.ParseCoinsNormalized(g.Params.MinDeposit)
	if err != nil {
		return fmt.Errorf("invalid min_deposit: %w", err)
	}
	if !minDeposit.IsAllPositive() {
		return fmt.Errorf("min_deposit must be positive, got %s", minDeposit)
	}

	// Validate time periods (must be >= 1 minute and <= 1 year)
	if g.Params.MaxDepositPeriod == nil {
		return fmt.Errorf("max_deposit_period cannot be nil")
	}
	if err := validatePeriod("max_deposit_period", g.Params.MaxDepositPeriod.Seconds); err != nil {
		return err
	}

	if g.Params.VotingPeriod == nil {
		return fmt.Errorf("voting_period cannot be nil")
	}
	if err := validatePeriod("voting_period", g.Params.VotingPeriod.Seconds); err != nil {
		return err
	}

	// Validate emergency voting period if set
	if g.Params.EmergencyVotingPeriod != nil {
		if err := validatePeriod("emergency_voting_period", g.Params.EmergencyVotingPeriod.Seconds); err != nil {
			return err
		}
	}

	// Validate execution delay (can be 0 for emergency proposals)
	if g.Params.ExecutionDelay != nil && g.Params.ExecutionDelay.Seconds < 0 {
		return fmt.Errorf("execution_delay cannot be negative")
	}
	if g.Params.ExecutionDelay != nil && g.Params.ExecutionDelay.Seconds > 365*24*3600 {
		return fmt.Errorf("execution_delay must be <= 1 year")
	}

	// Validate token lock duration if token locks are required
	if g.Params.RequireTokenLock {
		if g.Params.TokenLockDuration == nil {
			return fmt.Errorf("token_lock_duration cannot be nil when require_token_lock is true")
		}
		if g.Params.TokenLockDuration.Seconds < 60 {
			return fmt.Errorf("token_lock_duration must be >= 1 minute")
		}
		if g.Params.TokenLockDuration.Seconds > 365*24*3600 {
			return fmt.Errorf("token_lock_duration must be <= 1 year")
		}
	}

	// Validate reveal period if secret ballot is enabled
	if g.Params.SecretBallotEnabled {
		if g.Params.RevealPeriod == nil {
			return fmt.Errorf("reveal_period cannot be nil when secret_ballot_enabled is true")
		}
		if g.Params.RevealPeriod.Seconds < 60 {
			return fmt.Errorf("reveal_period must be >= 1 minute")
		}
		if g.Params.RevealPeriod.Seconds > 30*24*3600 {
			return fmt.Errorf("reveal_period must be <= 30 days")
		}
	}

	// Validate category-specific parameters
	for category, params := range g.Params.CategoryParams {
		if err := validateCategoryParams(category, params); err != nil {
			return err
		}
	}

	return nil
}

// validateThreshold validates that a threshold value is a valid percentage (0.0-1.0)
func validateThreshold(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s must be set", name)
	}
	threshold, err := sdkmath.LegacyNewDecFromStr(value)
	if err != nil {
		return fmt.Errorf("invalid %s: must be a valid decimal, got %s", name, value)
	}
	if threshold.IsNegative() || threshold.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("invalid %s: must be between 0.0 and 1.0, got %s", name, threshold)
	}
	return nil
}

// validatePeriod validates that a time period is reasonable (>= 1 minute and <= 1 year)
func validatePeriod(name string, seconds int64) error {
	const minSeconds = 60                // 1 minute
	const maxSeconds = 365 * 24 * 3600   // 1 year

	if seconds < minSeconds {
		return fmt.Errorf("%s must be >= 1 minute, got %d seconds", name, seconds)
	}
	if seconds > maxSeconds {
		return fmt.Errorf("%s must be <= 1 year, got %d seconds", name, seconds)
	}
	return nil
}

// validateCategoryParams validates category-specific governance parameters
func validateCategoryParams(category string, params *pb.CategoryParams) error {
	if params == nil {
		return fmt.Errorf("category params for %s cannot be nil", category)
	}

	// Validate thresholds
	if err := validateThreshold(fmt.Sprintf("category[%s].quorum", category), params.Quorum); err != nil {
		return err
	}
	if err := validateThreshold(fmt.Sprintf("category[%s].threshold", category), params.Threshold); err != nil {
		return err
	}
	if err := validateThreshold(fmt.Sprintf("category[%s].veto_threshold", category), params.VetoThreshold); err != nil {
		return err
	}

	// Parse thresholds for logical consistency
	threshold, _ := sdkmath.LegacyNewDecFromStr(params.Threshold)
	vetoThreshold, _ := sdkmath.LegacyNewDecFromStr(params.VetoThreshold)

	if vetoThreshold.GTE(threshold) {
		return fmt.Errorf("category[%s]: veto_threshold (%s) must be < threshold (%s)",
			category, vetoThreshold, threshold)
	}

	// Validate deposit amount
	if params.MinDeposit != "" {
		minDeposit, err := sdk.ParseCoinsNormalized(params.MinDeposit)
		if err != nil {
			return fmt.Errorf("invalid category[%s].min_deposit: %w", category, err)
		}
		if !minDeposit.IsAllPositive() {
			return fmt.Errorf("category[%s].min_deposit must be positive, got %s", category, minDeposit)
		}
	}

	// Validate voting period
	if params.VotingPeriod != nil {
		if err := validatePeriod(fmt.Sprintf("category[%s].voting_period", category), params.VotingPeriod.Seconds); err != nil {
			return err
		}
	}

	// Validate execution delay (can be 0)
	if params.ExecutionDelay != nil && params.ExecutionDelay.Seconds < 0 {
		return fmt.Errorf("category[%s].execution_delay cannot be negative", category)
	}
	if params.ExecutionDelay != nil && params.ExecutionDelay.Seconds > 365*24*3600 {
		return fmt.Errorf("category[%s].execution_delay must be <= 1 year", category)
	}

	return nil
}
