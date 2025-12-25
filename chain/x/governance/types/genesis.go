// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"sort"

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

	// Validate proposals
	proposalIDs := make(map[uint64]bool)
	for i, proposal := range g.Proposals {
		if proposal == nil {
			return fmt.Errorf("proposal at index %d is nil", i)
		}
		if proposal.Id == 0 {
			return fmt.Errorf("proposal at index %d has invalid ID 0", i)
		}
		if proposalIDs[proposal.Id] {
			return fmt.Errorf("duplicate proposal ID %d", proposal.Id)
		}
		proposalIDs[proposal.Id] = true
		if proposal.Title == "" {
			return fmt.Errorf("proposal %d has empty title", proposal.Id)
		}
		if proposal.Proposer == "" {
			return fmt.Errorf("proposal %d has empty proposer", proposal.Id)
		}
		// Validate proposer address
		if _, err := sdk.AccAddressFromBech32(proposal.Proposer); err != nil {
			return fmt.Errorf("proposal %d has invalid proposer address %s: %w", proposal.Id, proposal.Proposer, err)
		}
	}

	// Validate deposits
	for i, deposit := range g.Deposits {
		if deposit == nil {
			return fmt.Errorf("deposit at index %d is nil", i)
		}
		if deposit.ProposalId == 0 {
			return fmt.Errorf("deposit at index %d has invalid proposal ID 0", i)
		}
		if deposit.Depositor == "" {
			return fmt.Errorf("deposit at index %d for proposal %d has empty depositor", i, deposit.ProposalId)
		}
		// Validate depositor address
		if _, err := sdk.AccAddressFromBech32(deposit.Depositor); err != nil {
			return fmt.Errorf("deposit at index %d has invalid depositor address %s: %w", i, deposit.Depositor, err)
		}
		// Validate deposit amount
		if deposit.Amount == "" {
			return fmt.Errorf("deposit at index %d for proposal %d has empty amount", i, deposit.ProposalId)
		}
		coins, err := sdk.ParseCoinsNormalized(deposit.Amount)
		if err != nil {
			return fmt.Errorf("deposit at index %d has invalid amount %s: %w", i, deposit.Amount, err)
		}
		if !coins.IsAllPositive() {
			return fmt.Errorf("deposit at index %d has non-positive amount %s", i, deposit.Amount)
		}
	}

	// Validate votes
	for i, vote := range g.Votes {
		if vote == nil {
			return fmt.Errorf("vote at index %d is nil", i)
		}
		if vote.ProposalId == 0 {
			return fmt.Errorf("vote at index %d has invalid proposal ID 0", i)
		}
		if vote.Voter == "" {
			return fmt.Errorf("vote at index %d for proposal %d has empty voter", i, vote.ProposalId)
		}
		// Validate voter address
		if _, err := sdk.AccAddressFromBech32(vote.Voter); err != nil {
			return fmt.Errorf("vote at index %d has invalid voter address %s: %w", i, vote.Voter, err)
		}
		// Validate vote option
		if vote.Option < 1 || vote.Option > 4 {
			return fmt.Errorf("vote at index %d has invalid option %d", i, vote.Option)
		}
	}

	// Validate vote delegations
	for i, delegation := range g.VoteDelegations {
		if delegation == nil {
			return fmt.Errorf("vote delegation at index %d is nil", i)
		}
		if delegation.Delegator == "" {
			return fmt.Errorf("vote delegation at index %d has empty delegator", i)
		}
		if delegation.Delegate == "" {
			return fmt.Errorf("vote delegation at index %d has empty delegate", i)
		}
		// Validate delegator address
		if _, err := sdk.AccAddressFromBech32(delegation.Delegator); err != nil {
			return fmt.Errorf("vote delegation at index %d has invalid delegator address %s: %w", i, delegation.Delegator, err)
		}
		// Validate delegate address
		if _, err := sdk.AccAddressFromBech32(delegation.Delegate); err != nil {
			return fmt.Errorf("vote delegation at index %d has invalid delegate address %s: %w", i, delegation.Delegate, err)
		}
		// Prevent self-delegation
		if delegation.Delegator == delegation.Delegate {
			return fmt.Errorf("vote delegation at index %d: delegator and delegate cannot be the same address", i)
		}
	}

	// Validate token locks
	for i, lock := range g.TokenLocks {
		if lock == nil {
			return fmt.Errorf("token lock at index %d is nil", i)
		}
		if lock.Owner == "" {
			return fmt.Errorf("token lock at index %d has empty owner", i)
		}
		// Validate owner address
		if _, err := sdk.AccAddressFromBech32(lock.Owner); err != nil {
			return fmt.Errorf("token lock at index %d has invalid owner address %s: %w", i, lock.Owner, err)
		}
		if lock.ProposalId == 0 {
			return fmt.Errorf("token lock at index %d has invalid proposal ID 0", i)
		}
		if lock.LockedAmount == "" {
			return fmt.Errorf("token lock at index %d has empty locked amount", i)
		}
		// Validate locked amount
		coins, err := sdk.ParseCoinsNormalized(lock.LockedAmount)
		if err != nil {
			return fmt.Errorf("token lock at index %d has invalid locked amount %s: %w", i, lock.LockedAmount, err)
		}
		if !coins.IsAllPositive() {
			return fmt.Errorf("token lock at index %d has non-positive locked amount %s", i, lock.LockedAmount)
		}
	}

	// Validate veto requests
	for i, veto := range g.VetoRequests {
		if veto == nil {
			return fmt.Errorf("veto request at index %d is nil", i)
		}
		if veto.ProposalId == 0 {
			return fmt.Errorf("veto request at index %d has invalid proposal ID 0", i)
		}
		if veto.Vetoer == "" {
			return fmt.Errorf("veto request at index %d has empty vetoer", i)
		}
		// Validate vetoer address
		if _, err := sdk.AccAddressFromBech32(veto.Vetoer); err != nil {
			return fmt.Errorf("veto request at index %d has invalid vetoer address %s: %w", i, veto.Vetoer, err)
		}
		// Validate cosigners
		for j, cosigner := range veto.Cosigners {
			if cosigner == "" {
				return fmt.Errorf("veto request at index %d has empty cosigner at index %d", i, j)
			}
			if _, err := sdk.AccAddressFromBech32(cosigner); err != nil {
				return fmt.Errorf("veto request at index %d has invalid cosigner address %s at index %d: %w", i, cosigner, j, err)
			}
		}
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
	// CONSENSUS-CRITICAL: Sort category names for deterministic iteration order
	categoryNames := make([]string, 0, len(g.Params.CategoryParams))
	for category := range g.Params.CategoryParams {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)
	for _, category := range categoryNames {
		params := g.Params.CategoryParams[category]
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
