package types

import "fmt"

// Helper functions for GovernanceParams

// GetVotingPeriodSeconds returns voting period in seconds
func GetVotingPeriodSeconds(p *GovernanceParams) uint64 {
	if p == nil || p.VotingPeriod == nil {
		return 604800 // Default 7 days
	}
	return uint64(p.VotingPeriod.Seconds)
}

// GetDepositPeriodSeconds returns deposit period in seconds
func GetDepositPeriodSeconds(p *GovernanceParams) uint64 {
	if p == nil || p.MaxDepositPeriod == nil {
		return 604800 // Default 7 days
	}
	return uint64(p.MaxDepositPeriod.Seconds)
}

// GetExecutionDelaySeconds returns execution delay in seconds
func GetExecutionDelaySeconds(p *GovernanceParams) uint64 {
	if p == nil || p.ExecutionDelay == nil {
		return 172800 // Default 48 hours
	}
	return uint64(p.ExecutionDelay.Seconds)
}

// GetMaxDelegationsPerUser returns max delegations per user (default if not configured)
func GetMaxDelegationsPerUser(p *GovernanceParams) uint64 {
	// Default max delegations per user
	return 10
}

// GetRefundFailedProposals returns whether failed proposals get refunded (default false)
func GetRefundFailedProposals(p *GovernanceParams) bool {
	// Default: no refund for failed proposals
	return false
}

// GetFailedProposalRefundPercentage returns the refund percentage for failed proposals
func GetFailedProposalRefundPercentage(p *GovernanceParams) uint64 {
	// Default: 50% refund if enabled
	return 5000 // 50% in basis points
}

// ValidateGovernanceParams validates governance parameters
func ValidateGovernanceParams(p *GovernanceParams) error {
	if p == nil {
		return fmt.Errorf("governance params cannot be nil")
	}
	if p.MinDeposit == "" {
		return fmt.Errorf("min_deposit cannot be empty")
	}
	if p.Quorum == "" {
		return fmt.Errorf("quorum cannot be empty")
	}
	if p.Threshold == "" {
		return fmt.Errorf("threshold cannot be empty")
	}
	return nil
}
