package types

import (
	"fmt"
	"time"
)

// CategoryParams defines category-specific governance parameters
type CategoryParams struct {
	MinDeposit     string
	VotingPeriod   time.Duration
	Quorum         string
	Threshold      string
	VetoThreshold  string
	ExecutionDelay time.Duration
}

// GovernanceParams defines the parameters for the governance module
type GovernanceParams struct {
	// Deposit parameters
	MinDeposit       string
	MaxDepositPeriod time.Duration

	// Voting parameters
	VotingPeriod  time.Duration
	Quorum        string // Minimum participation percentage (e.g., "0.33" for 33%)
	Threshold     string // Minimum yes percentage (e.g., "0.50" for 50%)
	VetoThreshold string // Minimum veto percentage (e.g., "0.33" for 33%)

	// Execution delay (time-lock)
	ExecutionDelay time.Duration

	// Emergency fast-track
	EmergencyVotingPeriod time.Duration
	EmergencyQuorum       string
	EmergencyThreshold    string

	// Category-specific parameters
	CategoryParams map[string]CategoryParams

	// Veto parameters
	VetoCosignersRequired   uint32
	VetoAuthorizedAddresses []string

	// Token lock parameters
	RequireTokenLock  bool
	TokenLockDuration time.Duration

	// Snapshot voting
	SnapshotVotingEnabled  bool
	SnapshotLookbackBlocks uint64

	// Secret ballot
	SecretBallotEnabled bool
	RevealPeriod        time.Duration
}

// DefaultParams returns default governance parameters
func DefaultParams() GovernanceParams {
	return GovernanceParams{
		// Deposit parameters (10,000 tokens minimum)
		MinDeposit:       "10000000000",
		MaxDepositPeriod: 48 * time.Hour,

		// Voting parameters (3 days voting, 33% quorum, 50% threshold)
		VotingPeriod:  72 * time.Hour,
		Quorum:        "0.334",
		Threshold:     "0.500",
		VetoThreshold: "0.334",

		// Execution delay (24 hour time-lock)
		ExecutionDelay: 24 * time.Hour,

		// Emergency fast-track (24 hours, higher thresholds)
		EmergencyVotingPeriod: 24 * time.Hour,
		EmergencyQuorum:       "0.500",
		EmergencyThreshold:    "0.667",

		// Category-specific parameters
		CategoryParams: map[string]CategoryParams{
			"TEXT": {
				MinDeposit:     "5000000000",
				VotingPeriod:   48 * time.Hour,
				Quorum:         "0.250",
				Threshold:      "0.500",
				VetoThreshold:  "0.334",
				ExecutionDelay: 0,
			},
			"PARAMETER_CHANGE": {
				MinDeposit:     "15000000000",
				VotingPeriod:   72 * time.Hour,
				Quorum:         "0.400",
				Threshold:      "0.500",
				VetoThreshold:  "0.334",
				ExecutionDelay: 48 * time.Hour,
			},
			"SOFTWARE_UPGRADE": {
				MinDeposit:     "20000000000",
				VotingPeriod:   96 * time.Hour,
				Quorum:         "0.500",
				Threshold:      "0.667",
				VetoThreshold:  "0.334",
				ExecutionDelay: 72 * time.Hour,
			},
			"SPENDING": {
				MinDeposit:     "25000000000",
				VotingPeriod:   96 * time.Hour,
				Quorum:         "0.500",
				Threshold:      "0.600",
				VetoThreshold:  "0.334",
				ExecutionDelay: 48 * time.Hour,
			},
			"EMERGENCY": {
				MinDeposit:     "50000000000",
				VotingPeriod:   24 * time.Hour,
				Quorum:         "0.600",
				Threshold:      "0.750",
				VetoThreshold:  "0.250",
				ExecutionDelay: 0,
			},
			"CONSTITUTION": {
				MinDeposit:     "100000000000",
				VotingPeriod:   168 * time.Hour, // 7 days
				Quorum:         "0.667",
				Threshold:      "0.750",
				VetoThreshold:  "0.250",
				ExecutionDelay: 168 * time.Hour, // 7 days
			},
		},

		// Veto parameters (require 3 of 5 authorized addresses)
		VetoCosignersRequired:   3,
		VetoAuthorizedAddresses: []string{},

		// Token lock parameters
		RequireTokenLock:  true,
		TokenLockDuration: 24 * time.Hour,

		// Snapshot voting
		SnapshotVotingEnabled:  true,
		SnapshotLookbackBlocks: 100,

		// Secret ballot
		SecretBallotEnabled: true,
		RevealPeriod:        24 * time.Hour,
	}
}

// ValidateBasic performs basic validation on governance parameters
func (p GovernanceParams) ValidateBasic() error {
	if p.MinDeposit == "" || p.MinDeposit == "0" {
		return fmt.Errorf("min deposit must be positive")
	}
	if p.MaxDepositPeriod <= 0 {
		return fmt.Errorf("max deposit period must be positive")
	}
	if p.VotingPeriod <= 0 {
		return fmt.Errorf("voting period must be positive")
	}
	if p.Quorum == "" {
		return fmt.Errorf("quorum must be specified")
	}
	if p.Threshold == "" {
		return fmt.Errorf("threshold must be specified")
	}
	if p.VetoThreshold == "" {
		return fmt.Errorf("veto threshold must be specified")
	}
	if p.EmergencyVotingPeriod <= 0 {
		return fmt.Errorf("emergency voting period must be positive")
	}
	if p.RequireTokenLock && p.TokenLockDuration <= 0 {
		return fmt.Errorf("token lock duration must be positive when token lock is required")
	}
	if p.SecretBallotEnabled && p.RevealPeriod <= 0 {
		return fmt.Errorf("reveal period must be positive when secret ballot is enabled")
	}
	return nil
}

// GetCategoryParams returns the parameters for a specific category
func (p GovernanceParams) GetCategoryParams(category ProposalCategory) CategoryParams {
	categoryKey := category.String()
	if params, ok := p.CategoryParams[categoryKey]; ok {
		return params
	}
	// Return default params if category-specific params not found
	return CategoryParams{
		MinDeposit:     p.MinDeposit,
		VotingPeriod:   p.VotingPeriod,
		Quorum:         p.Quorum,
		Threshold:      p.Threshold,
		VetoThreshold:  p.VetoThreshold,
		ExecutionDelay: p.ExecutionDelay,
	}
}

// DefaultGovernanceParams returns default governance parameters
func DefaultGovernanceParams() GovernanceParams {
	return DefaultParams()
}

// Proto message interface methods for GovernanceParams
func (p *GovernanceParams) Reset() {
	*p = GovernanceParams{}
}

func (p *GovernanceParams) String() string {
	return fmt.Sprintf(`GovernanceParams{
  MinDeposit: %s,
  MaxDepositPeriod: %s,
  VotingPeriod: %s,
  Quorum: %s,
  Threshold: %s,
  VetoThreshold: %s,
  ExecutionDelay: %s,
  EmergencyVotingPeriod: %s,
  EmergencyQuorum: %s,
  EmergencyThreshold: %s,
  VetoCosignersRequired: %d,
  RequireTokenLock: %v,
  TokenLockDuration: %s,
  SnapshotVotingEnabled: %v,
  SnapshotLookbackBlocks: %d,
  SecretBallotEnabled: %v,
  RevealPeriod: %s,
}`,
		p.MinDeposit,
		p.MaxDepositPeriod,
		p.VotingPeriod,
		p.Quorum,
		p.Threshold,
		p.VetoThreshold,
		p.ExecutionDelay,
		p.EmergencyVotingPeriod,
		p.EmergencyQuorum,
		p.EmergencyThreshold,
		p.VetoCosignersRequired,
		p.RequireTokenLock,
		p.TokenLockDuration,
		p.SnapshotVotingEnabled,
		p.SnapshotLookbackBlocks,
		p.SecretBallotEnabled,
		p.RevealPeriod,
	)
}

func (p *GovernanceParams) ProtoMessage() {}

// Proto message interface methods for CategoryParams
func (cp *CategoryParams) Reset() {
	*cp = CategoryParams{}
}

func (cp *CategoryParams) String() string {
	return fmt.Sprintf(`CategoryParams{
  MinDeposit: %s,
  VotingPeriod: %s,
  Quorum: %s,
  Threshold: %s,
  VetoThreshold: %s,
  ExecutionDelay: %s,
}`,
		cp.MinDeposit,
		cp.VotingPeriod,
		cp.Quorum,
		cp.Threshold,
		cp.VetoThreshold,
		cp.ExecutionDelay,
	)
}

func (cp *CategoryParams) ProtoMessage() {}
