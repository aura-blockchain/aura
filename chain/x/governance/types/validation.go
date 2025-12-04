package types

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

// Enum constant aliases for backward compatibility and convenience
const (
	// ProposalCategory aliases
	CategoryText            = ProposalCategory_PROPOSAL_CATEGORY_TEXT
	CategoryParameterChange = ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE
	CategorySoftwareUpgrade = ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE
	CategorySpending        = ProposalCategory_PROPOSAL_CATEGORY_SPENDING
	CategoryEmergency       = ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY
	CategoryConstitution    = ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION

	// ProposalStatus aliases
	StatusDepositPeriod      = ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD
	StatusVotingPeriod       = ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	StatusPassed             = ProposalStatus_PROPOSAL_STATUS_PASSED
	StatusRejected           = ProposalStatus_PROPOSAL_STATUS_REJECTED
	StatusFailed             = ProposalStatus_PROPOSAL_STATUS_FAILED
	StatusVetoed             = ProposalStatus_PROPOSAL_STATUS_VETOED
	StatusExecutionDelay     = ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY
	StatusReadyForExecution  = ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION
	StatusExecuted           = ProposalStatus_PROPOSAL_STATUS_EXECUTED

	// VoteOption aliases
	OptionUnspecified = VoteOption_VOTE_OPTION_UNSPECIFIED
	OptionYes         = VoteOption_VOTE_OPTION_YES
	OptionAbstain     = VoteOption_VOTE_OPTION_ABSTAIN
	OptionNo          = VoteOption_VOTE_OPTION_NO
	OptionNoWithVeto  = VoteOption_VOTE_OPTION_NO_WITH_VETO
)

// DefaultParams is an alias for DefaultGovernanceParams
func DefaultParams() *GovernanceParams {
	params := DefaultGovernanceParams()
	return &params
}

// DefaultGovernanceParams returns default governance parameters
func DefaultGovernanceParams() GovernanceParams {
	// Initialize category-specific parameters
	categoryParams := map[string]*CategoryParams{
		CategoryText.String(): {
			MinDeposit:     "10000000stake",                      // 10 AURA
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),   // 7 days
			Quorum:         "0.334",                              // 33.4%
			Threshold:      "0.50",                               // 50%
			VetoThreshold:  "0.334",                              // 33.4%
			ExecutionDelay: durationpb.New(48 * time.Hour),       // 48 hours
		},
		CategoryParameterChange.String(): {
			MinDeposit:     "10000000stake",                      // 10 AURA
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),   // 7 days
			Quorum:         "0.334",                              // 33.4%
			Threshold:      "0.50",                               // 50%
			VetoThreshold:  "0.334",                              // 33.4%
			ExecutionDelay: durationpb.New(48 * time.Hour),       // 48 hours
		},
		CategorySoftwareUpgrade.String(): {
			MinDeposit:     "10000000stake",                      // 10 AURA
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),   // 7 days
			Quorum:         "0.334",                              // 33.4%
			Threshold:      "0.50",                               // 50%
			VetoThreshold:  "0.334",                              // 33.4%
			ExecutionDelay: durationpb.New(48 * time.Hour),       // 48 hours
		},
		CategorySpending.String(): {
			MinDeposit:     "10000000stake",                      // 10 AURA
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),   // 7 days
			Quorum:         "0.334",                              // 33.4%
			Threshold:      "0.50",                               // 50%
			VetoThreshold:  "0.334",                              // 33.4%
			ExecutionDelay: durationpb.New(48 * time.Hour),       // 48 hours
		},
		CategoryEmergency.String(): {
			MinDeposit:     "10000000stake",                      // 10 AURA
			VotingPeriod:   durationpb.New(24 * time.Hour),       // 24 hours
			Quorum:         "0.600",                              // 60%
			Threshold:      "0.750",                              // 75%
			VetoThreshold:  "0.334",                              // 33.4%
			ExecutionDelay: durationpb.New(0),                    // No delay for emergency
		},
		CategoryConstitution.String(): {
			MinDeposit:     "10000000stake",                      // 10 AURA
			VotingPeriod:   durationpb.New(7 * 24 * time.Hour),   // 7 days
			Quorum:         "0.667",                              // 66.7%
			Threshold:      "0.750",                              // 75%
			VetoThreshold:  "0.334",                              // 33.4%
			ExecutionDelay: durationpb.New(48 * time.Hour),       // 48 hours
		},
	}

	return GovernanceParams{
		// Deposit parameters
		MinDeposit:       "10000000stake", // 10 AURA (using stake denom for genesis compatibility)
		MaxDepositPeriod: durationpb.New(7 * 24 * time.Hour), // 7 days

		// Voting parameters
		VotingPeriod:   durationpb.New(7 * 24 * time.Hour), // 7 days
		Quorum:         "0.334",                            // 33.4%
		Threshold:      "0.50",                             // 50%
		VetoThreshold:  "0.334",                            // 33.4%

		// Execution delay (time-lock)
		ExecutionDelay: durationpb.New(48 * time.Hour), // 48 hours

		// Emergency fast-track
		EmergencyVotingPeriod: durationpb.New(24 * time.Hour), // 24 hours
		EmergencyQuorum:       "0.50",                         // 50%
		EmergencyThreshold:    "0.667",                        // 66.7%

		// Category-specific thresholds
		CategoryParams: categoryParams,

		// Veto parameters
		VetoCosignersRequired:    3,
		VetoAuthorizedAddresses:  []string{},

		// Token lock parameters
		RequireTokenLock:   false,
		TokenLockDuration:  durationpb.New(7 * 24 * time.Hour), // 7 days

		// Snapshot voting
		SnapshotVotingEnabled: false,
		SnapshotLookbackBlocks: 100,

		// Secret ballot
		SecretBallotEnabled: false,
		RevealPeriod:        durationpb.New(24 * time.Hour), // 24 hours
	}
}
