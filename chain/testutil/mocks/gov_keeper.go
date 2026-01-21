// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// MockGovKeeper implements a mock GovKeeper for testing
type MockGovKeeper struct {
	Proposals      map[uint64]*govtypes.Proposal
	Votes          map[uint64]map[string]govtypes.Vote
	Deposits       map[uint64]map[string]govtypes.Deposit
	NextProposalID uint64
}

// NewMockGovKeeper creates a new mock governance keeper
func NewMockGovKeeper() *MockGovKeeper {
	return &MockGovKeeper{
		Proposals:      make(map[uint64]*govtypes.Proposal),
		Votes:          make(map[uint64]map[string]govtypes.Vote),
		Deposits:       make(map[uint64]map[string]govtypes.Deposit),
		NextProposalID: 1,
	}
}

// GetProposal returns a proposal by ID
func (m *MockGovKeeper) GetProposal(ctx context.Context, proposalID uint64) (govtypes.Proposal, bool) {
	if prop, ok := m.Proposals[proposalID]; ok {
		return *prop, true
	}
	return govtypes.Proposal{}, false
}

// SetProposal sets a proposal
func (m *MockGovKeeper) SetProposal(ctx context.Context, proposal govtypes.Proposal) {
	m.Proposals[proposal.Id] = &proposal
}

// SubmitProposal submits a new proposal
func (m *MockGovKeeper) SubmitProposal(ctx context.Context, messages []sdk.Msg, metadata string, title string, summary string, proposer sdk.AccAddress) (govtypes.Proposal, error) {
	proposal := govtypes.Proposal{
		Id:               m.NextProposalID,
		Messages:         nil, // Would need proper conversion
		Status:           govtypes.StatusDepositPeriod,
		FinalTallyResult: &govtypes.TallyResult{},
		SubmitTime:       nil, // Would set to ctx time
		DepositEndTime:   nil,
		TotalDeposit:     sdk.NewCoins(),
		VotingStartTime:  nil,
		VotingEndTime:    nil,
		Metadata:         metadata,
		Title:            title,
		Summary:          summary,
		Proposer:         proposer.String(),
	}

	m.Proposals[m.NextProposalID] = &proposal
	m.NextProposalID++

	return proposal, nil
}

// AddVote adds a vote to a proposal
func (m *MockGovKeeper) AddVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress, options govtypes.WeightedVoteOptions, metadata string) error {
	if _, ok := m.Votes[proposalID]; !ok {
		m.Votes[proposalID] = make(map[string]govtypes.Vote)
	}

	vote := govtypes.Vote{
		ProposalId: proposalID,
		Voter:      voterAddr.String(),
		Options:    options,
		Metadata:   metadata,
	}

	m.Votes[proposalID][voterAddr.String()] = vote
	return nil
}

// GetVote returns a vote
func (m *MockGovKeeper) GetVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress) (govtypes.Vote, bool) {
	if votes, ok := m.Votes[proposalID]; ok {
		if vote, ok := votes[voterAddr.String()]; ok {
			return vote, true
		}
	}
	return govtypes.Vote{}, false
}

// AddDeposit adds a deposit to a proposal
func (m *MockGovKeeper) AddDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress, depositAmount sdk.Coins) error {
	if _, ok := m.Deposits[proposalID]; !ok {
		m.Deposits[proposalID] = make(map[string]govtypes.Deposit)
	}

	deposit := govtypes.Deposit{
		ProposalId: proposalID,
		Depositor:  depositorAddr.String(),
		Amount:     depositAmount,
	}

	m.Deposits[proposalID][depositorAddr.String()] = deposit

	// Update total deposit on proposal
	if prop, ok := m.Proposals[proposalID]; ok {
		prop.TotalDeposit = sdk.NewCoins(prop.TotalDeposit...).Add(depositAmount...)
		m.Proposals[proposalID] = prop
	}

	return nil
}

// GetDeposit returns a deposit
func (m *MockGovKeeper) GetDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress) (govtypes.Deposit, bool) {
	if deposits, ok := m.Deposits[proposalID]; ok {
		if deposit, ok := deposits[depositorAddr.String()]; ok {
			return deposit, true
		}
	}
	return govtypes.Deposit{}, false
}

// GetParams returns governance params (mock)
func (m *MockGovKeeper) GetParams(ctx context.Context) govtypes.Params {
	return govtypes.DefaultParams()
}

// SetProposalStatus sets a proposal status (test helper)
func (m *MockGovKeeper) SetProposalStatus(proposalID uint64, status govtypes.ProposalStatus) {
	if prop, ok := m.Proposals[proposalID]; ok {
		prop.Status = status
		m.Proposals[proposalID] = prop
	}
}

// DeleteProposal deletes a proposal (test helper)
func (m *MockGovKeeper) DeleteProposal(proposalID uint64) {
	delete(m.Proposals, proposalID)
	delete(m.Votes, proposalID)
	delete(m.Deposits, proposalID)
}
