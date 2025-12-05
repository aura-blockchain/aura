package types

import (
	"errors"
	"fmt"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// DefaultGenesisState returns the default genesis state for the economics module
func DefaultGenesisState() *economicspb.GenesisState {
	return &economicspb.GenesisState{
		Params:               *DefaultParams(),
		VestingSchedules:     []economicspb.VestingSchedule{},
		Proposals:            []economicspb.Proposal{},
		Votes:                []economicspb.Vote{},
		Deposits:             []economicspb.Deposit{},
		VoteLocks:            []economicspb.VoteLock{},
		VoteDelegations:      []economicspb.VoteDelegation{},
		PendingTreasuryTxs:   []economicspb.PendingTreasuryTx{},
		UserMevBalances:      make(map[string]string),
		LastLargeTxTimes:     make(map[string]int64),
		NextProposalId:       1,
		NextVestingScheduleId: 1,
		NextVoteLockId:       1,
		NextTreasuryTxId:     1,
	}
}

// ValidateGenesisState performs basic validation of genesis data
func ValidateGenesisState(gs *economicspb.GenesisState) error {
	if gs == nil {
		return errors.New("genesis state cannot be nil")
	}

	// Validate params
	if err := ValidateParamsProto(&gs.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate vesting schedules
	scheduleIDs := make(map[string]bool)
	for i, schedule := range gs.VestingSchedules {
		if schedule.Id == "" {
			return fmt.Errorf("vesting schedule %d: ID cannot be empty", i)
		}
		if scheduleIDs[schedule.Id] {
			return fmt.Errorf("vesting schedule %d: duplicate ID %s", i, schedule.Id)
		}
		scheduleIDs[schedule.Id] = true

		if schedule.Address == "" {
			return fmt.Errorf("vesting schedule %d: address cannot be empty", i)
		}
	}

	// Validate vote locks
	lockIDs := make(map[string]bool)
	for i, lock := range gs.VoteLocks {
		if lock.Id == "" {
			return fmt.Errorf("vote lock %d: ID cannot be empty", i)
		}
		if lockIDs[lock.Id] {
			return fmt.Errorf("vote lock %d: duplicate ID %s", i, lock.Id)
		}
		lockIDs[lock.Id] = true

		if lock.Owner == "" {
			return fmt.Errorf("vote lock %d: owner cannot be empty", i)
		}
	}

	// Validate proposals
	proposalIDs := make(map[uint64]bool)
	for i, proposal := range gs.Proposals {
		if proposal.Id == 0 {
			return fmt.Errorf("proposal %d: ID cannot be zero", i)
		}
		if proposalIDs[proposal.Id] {
			return fmt.Errorf("proposal %d: duplicate ID %d", i, proposal.Id)
		}
		proposalIDs[proposal.Id] = true

		if proposal.Proposer == "" {
			return fmt.Errorf("proposal %d: proposer cannot be empty", i)
		}
		if proposal.Title == "" {
			return fmt.Errorf("proposal %d: title cannot be empty", i)
		}
	}

	// Validate votes
	for i, vote := range gs.Votes {
		if vote.ProposalId == 0 {
			return fmt.Errorf("vote %d: proposal ID cannot be zero", i)
		}
		if vote.Voter == "" {
			return fmt.Errorf("vote %d: voter cannot be empty", i)
		}
	}

	// Validate deposits
	for i, deposit := range gs.Deposits {
		if deposit.ProposalId == 0 {
			return fmt.Errorf("deposit %d: proposal ID cannot be zero", i)
		}
		if deposit.Depositor == "" {
			return fmt.Errorf("deposit %d: depositor cannot be empty", i)
		}
	}

	// Validate pending treasury transactions
	txIDs := make(map[string]bool)
	for i, tx := range gs.PendingTreasuryTxs {
		if tx.TxId == "" {
			return fmt.Errorf("pending tx %d: ID cannot be empty", i)
		}
		if txIDs[tx.TxId] {
			return fmt.Errorf("pending tx %d: duplicate ID %s", i, tx.TxId)
		}
		txIDs[tx.TxId] = true

		if tx.Recipient == "" {
			return fmt.Errorf("pending tx %d: recipient cannot be empty", i)
		}
	}

	return nil
}

// ValidateParamsProto validates the proto Params message
func ValidateParamsProto(p *economicspb.Params) error {
	if p == nil {
		return errors.New("params cannot be nil")
	}

	// Validate fee params
	if p.Fees.TargetBlockUtilization > 10000 {
		return errors.New("target block utilization cannot exceed 100%")
	}
	if p.Fees.MaxFeeMultiplier < p.Fees.MinFeeMultiplier {
		return errors.New("max fee multiplier must be >= min fee multiplier")
	}

	// Validate vesting params
	if p.Vesting.MaxVestingDuration < p.Vesting.MinVestingDuration {
		return errors.New("max vesting duration must be >= min vesting duration")
	}

	// Validate governance params
	if p.Governance.Quorum > 10000 {
		return errors.New("governance quorum cannot exceed 100%")
	}
	if p.Governance.Threshold > 10000 {
		return errors.New("governance threshold cannot exceed 100%")
	}
	if p.Governance.VetoThreshold > 10000 {
		return errors.New("governance veto threshold cannot exceed 100%")
	}

	// Validate tokenomics params
	if p.Tokenomics.MaxInflationRate < p.Tokenomics.MinInflationRate {
		return errors.New("max inflation rate must be >= min inflation rate")
	}

	// Validate whale protection params
	if p.WhaleProtection.MaxHoldingPercentage > 10000 {
		return errors.New("max holding percentage cannot exceed 100%")
	}

	return nil
}
