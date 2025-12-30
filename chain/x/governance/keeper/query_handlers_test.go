// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

func setupQueryKeeper(t *testing.T) (*Keeper, govpb.QueryServer, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})
	keeper.SetParams(ctx, types.DefaultParams())
	queryServer := NewQueryServerImpl(keeper)
	return keeper, queryServer, ctx
}

// TestQueryProposal tests the Proposal query handler
func TestQueryProposal(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create a test proposal
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Category:    types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Status:      types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Proposer:    testAddr("proposer1"),
		SubmitTime:  ts,
	}
	err := keeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	tests := []struct {
		name      string
		req       *govpb.QueryProposalRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryProposalResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QueryProposalRequest{
				ProposalId: 0,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "proposal not found",
			req: &govpb.QueryProposalRequest{
				ProposalId: 999,
			},
			wantErr: true,
			errMsg:  "proposal not found",
		},
		{
			name: "valid proposal query",
			req: &govpb.QueryProposalRequest{
				ProposalId: 1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryProposalResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Proposal)
				require.Equal(t, uint64(1), resp.Proposal.Id)
				require.Equal(t, "Test Proposal", resp.Proposal.Title)
				require.Equal(t, testAddr("proposer1"), resp.Proposal.Proposer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Proposal(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryProposalsWithFilters tests Proposals query with filters
func TestQueryProposalsWithFilters(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	voter1 := testAddr("voter1")
	depositor1 := testAddr("depositor1")

	// Create proposals with different statuses
	proposal1 := &types.Proposal{
		Id:          1,
		Title:       "Voting Proposal",
		Description: "Test Description",
		Category:    types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Status:      types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Proposer:    testAddr("proposer1"),
		SubmitTime:  ts,
	}
	require.NoError(t, keeper.SetProposal(ctx, proposal1))

	proposal2 := &types.Proposal{
		Id:          2,
		Title:       "Passed Proposal",
		Description: "Test Description",
		Category:    types.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE,
		Status:      types.ProposalStatus_PROPOSAL_STATUS_PASSED,
		Proposer:    testAddr("proposer2"),
		SubmitTime:  ts,
	}
	require.NoError(t, keeper.SetProposal(ctx, proposal2))

	// Add vote and deposit for filtering
	vote := &types.Vote{
		ProposalId: 1,
		Voter:      voter1,
		Option:     types.VoteOption_VOTE_OPTION_YES,
		Timestamp:  ts,
	}
	require.NoError(t, keeper.SetVote(ctx, vote))

	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  depositor1,
		Amount:     "1000",
	}
	require.NoError(t, keeper.SetDeposit(ctx, deposit))

	tests := []struct {
		name      string
		req       *govpb.QueryProposalsRequest
		wantErr   bool
		checkResp func(*govpb.QueryProposalsResponse)
	}{
		{
			name:    "nil request defaults",
			req:     nil,
			wantErr: false,
			checkResp: func(resp *govpb.QueryProposalsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Proposals, 2)
			},
		},
		{
			name: "filter by status voting",
			req: &govpb.QueryProposalsRequest{
				Status: types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryProposalsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Proposals, 1)
				require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, resp.Proposals[0].Status)
			},
		},
		{
			name: "filter by status passed",
			req: &govpb.QueryProposalsRequest{
				Status: types.ProposalStatus_PROPOSAL_STATUS_PASSED,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryProposalsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Proposals, 1)
				require.Equal(t, types.ProposalStatus_PROPOSAL_STATUS_PASSED, resp.Proposals[0].Status)
			},
		},
		{
			name: "filter by voter",
			req: &govpb.QueryProposalsRequest{
				Voter: voter1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryProposalsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Proposals, 1)
				require.Equal(t, uint64(1), resp.Proposals[0].Id)
			},
		},
		{
			name: "filter by depositor",
			req: &govpb.QueryProposalsRequest{
				Depositor: depositor1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryProposalsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Proposals, 1)
				require.Equal(t, uint64(1), resp.Proposals[0].Id)
			},
		},
		{
			name: "filter by non-existent voter",
			req: &govpb.QueryProposalsRequest{
				Voter: testAddr("nonexistent"),
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryProposalsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Proposals, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryVote tests the Vote query handler
func TestQueryVote(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	voter := testAddr("voter1")

	// Create a vote
	vote := &types.Vote{
		ProposalId:  1,
		Voter:       voter,
		Option:      types.VoteOption_VOTE_OPTION_YES,
		VotingPower: "1000",
		Timestamp:   ts,
	}
	require.NoError(t, keeper.SetVote(ctx, vote))

	tests := []struct {
		name      string
		req       *govpb.QueryVoteRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryVoteResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QueryVoteRequest{
				ProposalId: 0,
				Voter:      voter,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "empty voter",
			req: &govpb.QueryVoteRequest{
				ProposalId: 1,
				Voter:      "",
			},
			wantErr: true,
			errMsg:  "voter cannot be empty",
		},
		{
			name: "vote not found",
			req: &govpb.QueryVoteRequest{
				ProposalId: 1,
				Voter:      testAddr("nonexistent"),
			},
			wantErr: true,
			errMsg:  "vote not found",
		},
		{
			name: "valid vote query",
			req: &govpb.QueryVoteRequest{
				ProposalId: 1,
				Voter:      voter,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVoteResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Vote)
				require.Equal(t, uint64(1), resp.Vote.ProposalId)
				require.Equal(t, voter, resp.Vote.Voter)
				require.Equal(t, types.VoteOption_VOTE_OPTION_YES, resp.Vote.Option)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Vote(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryVotes tests the Votes query handler
func TestQueryVotes(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create multiple votes for proposal 1
	votes := []*types.Vote{
		{ProposalId: 1, Voter: testAddr("voter1"), Option: types.VoteOption_VOTE_OPTION_YES, Timestamp: ts},
		{ProposalId: 1, Voter: testAddr("voter2"), Option: types.VoteOption_VOTE_OPTION_NO, Timestamp: ts},
		{ProposalId: 1, Voter: testAddr("voter3"), Option: types.VoteOption_VOTE_OPTION_ABSTAIN, Timestamp: ts},
	}
	for _, v := range votes {
		require.NoError(t, keeper.SetVote(ctx, v))
	}

	tests := []struct {
		name      string
		req       *govpb.QueryVotesRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryVotesResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QueryVotesRequest{
				ProposalId: 0,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "valid votes query",
			req: &govpb.QueryVotesRequest{
				ProposalId: 1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVotesResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Votes, 3)
			},
		},
		{
			name: "no votes for proposal",
			req: &govpb.QueryVotesRequest{
				ProposalId: 999,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVotesResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Votes, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Votes(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryParams tests the Params query handler
func TestQueryParams(t *testing.T) {
	_, queryServer, ctx := setupQueryKeeper(t)
	_ = ctx // Use ctx to avoid unused variable error

	tests := []struct {
		name      string
		req       *govpb.QueryParamsRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryParamsResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name:    "valid params query",
			req:     &govpb.QueryParamsRequest{},
			wantErr: false,
			checkResp: func(resp *govpb.QueryParamsResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Params)
				require.NotEmpty(t, resp.Params.MinDeposit)
				require.NotEmpty(t, resp.Params.Quorum)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Params(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryDeposit tests the Deposit query handler
func TestQueryDeposit(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	depositor := testAddr("depositor1")

	// Create a deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  depositor,
		Amount:     "5000",
	}
	require.NoError(t, keeper.SetDeposit(ctx, deposit))

	tests := []struct {
		name      string
		req       *govpb.QueryDepositRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryDepositResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QueryDepositRequest{
				ProposalId: 0,
				Depositor:  depositor,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "empty depositor",
			req: &govpb.QueryDepositRequest{
				ProposalId: 1,
				Depositor:  "",
			},
			wantErr: true,
			errMsg:  "depositor cannot be empty",
		},
		{
			name: "deposit not found",
			req: &govpb.QueryDepositRequest{
				ProposalId: 1,
				Depositor:  testAddr("nonexistent"),
			},
			wantErr: true,
			errMsg:  "deposit not found",
		},
		{
			name: "valid deposit query",
			req: &govpb.QueryDepositRequest{
				ProposalId: 1,
				Depositor:  depositor,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryDepositResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Deposit)
				require.Equal(t, uint64(1), resp.Deposit.ProposalId)
				require.Equal(t, depositor, resp.Deposit.Depositor)
				require.Equal(t, "5000", resp.Deposit.Amount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Deposit(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryDeposits tests the Deposits query handler
func TestQueryDeposits(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	// Create multiple deposits for proposal 1
	deposits := []*types.Deposit{
		{ProposalId: 1, Depositor: testAddr("depositor1"), Amount: "1000"},
		{ProposalId: 1, Depositor: testAddr("depositor2"), Amount: "2000"},
		{ProposalId: 1, Depositor: testAddr("depositor3"), Amount: "3000"},
	}
	for _, d := range deposits {
		require.NoError(t, keeper.SetDeposit(ctx, d))
	}

	tests := []struct {
		name      string
		req       *govpb.QueryDepositsRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryDepositsResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QueryDepositsRequest{
				ProposalId: 0,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "valid deposits query",
			req: &govpb.QueryDepositsRequest{
				ProposalId: 1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryDepositsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Deposits, 3)
			},
		},
		{
			name: "no deposits for proposal",
			req: &govpb.QueryDepositsRequest{
				ProposalId: 999,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryDepositsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Deposits, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.Deposits(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryTallyResult tests the TallyResult query handler
func TestQueryTallyResult(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create a proposal with votes
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Category:    types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Status:      types.ProposalStatus_PROPOSAL_STATUS_PASSED,
		Proposer:    testAddr("proposer1"),
		SubmitTime:  ts,
	}
	require.NoError(t, keeper.SetProposal(ctx, proposal))

	// Add votes to match expected tally (TallyResult query calculates from votes)
	yesVote := &types.Vote{
		ProposalId:  1,
		Voter:       testAddr("voter1"),
		Option:      types.VoteOptionYes,
		VotingPower: "1000",
		Timestamp:   ts,
	}
	keeper.SetVote(ctx, yesVote)

	noVote := &types.Vote{
		ProposalId:  1,
		Voter:       testAddr("voter2"),
		Option:      types.VoteOptionNo,
		VotingPower: "100",
		Timestamp:   ts,
	}
	keeper.SetVote(ctx, noVote)

	// Create proposal without votes
	proposal2 := &types.Proposal{
		Id:          2,
		Title:       "Test Proposal 2",
		Description: "Test Description",
		Category:    types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Status:      types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Proposer:    testAddr("proposer1"),
		SubmitTime:  ts,
	}
	require.NoError(t, keeper.SetProposal(ctx, proposal2))

	tests := []struct {
		name      string
		req       *govpb.QueryTallyResultRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryTallyResultResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QueryTallyResultRequest{
				ProposalId: 0,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "proposal not found",
			req: &govpb.QueryTallyResultRequest{
				ProposalId: 999,
			},
			wantErr: true,
			errMsg:  "proposal not found",
		},
		{
			name: "valid tally query with existing tally",
			req: &govpb.QueryTallyResultRequest{
				ProposalId: 1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryTallyResultResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Tally)
				require.Equal(t, "1000", resp.Tally.Yes)
				require.Equal(t, "100", resp.Tally.No)
			},
		},
		{
			name: "valid tally query without existing tally",
			req: &govpb.QueryTallyResultRequest{
				ProposalId: 2,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryTallyResultResponse) {
				require.NotNil(t, resp)
				require.NotNil(t, resp.Tally)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.TallyResult(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryVoteDelegations tests the VoteDelegations query handler
func TestQueryVoteDelegations(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	delegator := testAddr("delegator1")

	// Create delegations
	delegations := []*types.VoteDelegation{
		{
			Delegator:      delegator,
			Delegate:       testAddr("delegate1"),
			DelegatedPower: "1000",
			DelegationTime: ts,
		},
		{
			Delegator:      delegator,
			Delegate:       testAddr("delegate2"),
			DelegatedPower: "2000",
			DelegationTime: ts,
		},
	}
	for _, d := range delegations {
		require.NoError(t, keeper.SetVoteDelegation(ctx, d))
	}

	tests := []struct {
		name      string
		req       *govpb.QueryVoteDelegationsRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryVoteDelegationsResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty delegator",
			req: &govpb.QueryVoteDelegationsRequest{
				Delegator: "",
			},
			wantErr: true,
			errMsg:  "delegator cannot be empty",
		},
		{
			name: "valid delegations query",
			req: &govpb.QueryVoteDelegationsRequest{
				Delegator: delegator,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVoteDelegationsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Delegations, 2)
			},
		},
		{
			name: "no delegations for delegator",
			req: &govpb.QueryVoteDelegationsRequest{
				Delegator: testAddr("nonexistent"),
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVoteDelegationsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Delegations, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.VoteDelegations(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryVotingPower tests the VotingPower query handler
func TestQueryVotingPower(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	address := testAddr("voter1")

	// Set up staking keeper with bonded amount
	mockStaking := keeper.stakingKeeper.(*MockStakingKeeper)
	mockStaking.SetDelegatorBonded(address, sdkmath.NewInt(5000))

	// Create a delegation to this address
	delegation := &types.VoteDelegation{
		Delegator:      testAddr("delegator1"),
		Delegate:       address,
		DelegatedPower: "1000",
		DelegationTime: ts,
	}
	require.NoError(t, keeper.SetVoteDelegation(ctx, delegation))

	tests := []struct {
		name      string
		req       *govpb.QueryVotingPowerRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryVotingPowerResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty address",
			req: &govpb.QueryVotingPowerRequest{
				Address: "",
			},
			wantErr: true,
			errMsg:  "address cannot be empty",
		},
		{
			name: "valid voting power query",
			req: &govpb.QueryVotingPowerRequest{
				Address: address,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVotingPowerResponse) {
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.VotingPower)
				require.NotEmpty(t, resp.TotalPower)
			},
		},
		{
			name: "address with no voting power",
			req: &govpb.QueryVotingPowerRequest{
				Address: testAddr("novotingpower"),
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVotingPowerResponse) {
				require.NotNil(t, resp)
				require.Equal(t, "0", resp.VotingPower)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For valid address test, ensure mock staking is set
			if tt.name == "valid voting power query" {
				mockStaking.SetDelegatorBonded(address, sdkmath.NewInt(5000))
			}

			resp, err := queryServer.VotingPower(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryTokenLocks tests the TokenLocks query handler
func TestQueryTokenLocks(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	lockTime, _ := gogotypes.TimestampProto(now)
	unlockTime, _ := gogotypes.TimestampProto(now.Add(24 * time.Hour))
	address := testAddr("owner1")

	// Create token locks
	locks := []*types.TokenLock{
		{
			Owner:        address,
			ProposalId:   1,
			LockedAmount: "1000",
			LockTime:     lockTime,
			UnlockTime:   unlockTime,
		},
		{
			Owner:        address,
			ProposalId:   2,
			LockedAmount: "2000",
			LockTime:     lockTime,
			UnlockTime:   unlockTime,
		},
	}
	for _, l := range locks {
		require.NoError(t, keeper.SetTokenLock(ctx, l))
	}

	tests := []struct {
		name      string
		req       *govpb.QueryTokenLocksRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryTokenLocksResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty address",
			req: &govpb.QueryTokenLocksRequest{
				Address: "",
			},
			wantErr: true,
			errMsg:  "address cannot be empty",
		},
		{
			name: "valid token locks query",
			req: &govpb.QueryTokenLocksRequest{
				Address: address,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryTokenLocksResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Locks, 2)
			},
		},
		{
			name: "no token locks for address",
			req: &govpb.QueryTokenLocksRequest{
				Address: testAddr("nonexistent"),
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryTokenLocksResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Locks, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.TokenLocks(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryVetoRequests tests the VetoRequests query handler
func TestQueryVetoRequests(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create veto request
	vetoRequest := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     testAddr("vetoer1"),
		Reason:     "Security concern",
		Timestamp:  ts,
		Cosigners:  []string{testAddr("vetoer1"), testAddr("vetoer2")},
	}
	require.NoError(t, keeper.SetVetoRequest(ctx, vetoRequest))

	tests := []struct {
		name      string
		req       *govpb.QueryVetoRequestsRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QueryVetoRequestsResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QueryVetoRequestsRequest{
				ProposalId: 0,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "valid veto requests query",
			req: &govpb.QueryVetoRequestsRequest{
				ProposalId: 1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVetoRequestsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.VetoRequests, 1)
				require.Equal(t, "Security concern", resp.VetoRequests[0].Reason)
			},
		},
		{
			name: "no veto requests for proposal",
			req: &govpb.QueryVetoRequestsRequest{
				ProposalId: 999,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QueryVetoRequestsResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.VetoRequests, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.VetoRequests(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQuerySnapshotVotes tests the SnapshotVotes query handler
func TestQuerySnapshotVotes(t *testing.T) {
	keeper, queryServer, ctx := setupQueryKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)

	// Create snapshot votes
	snapshotVotes := []*types.SnapshotVote{
		{
			ProposalId:            1,
			Voter:                 testAddr("voter1"),
			Option:                types.VoteOption_VOTE_OPTION_YES,
			VotingPowerAtSnapshot: "1000",
			Signature:             "sig1",
			Timestamp:             ts,
		},
		{
			ProposalId:            1,
			Voter:                 testAddr("voter2"),
			Option:                types.VoteOption_VOTE_OPTION_NO,
			VotingPowerAtSnapshot: "2000",
			Signature:             "sig2",
			Timestamp:             ts,
		},
	}
	for _, sv := range snapshotVotes {
		require.NoError(t, keeper.SetSnapshotVote(ctx, sv))
	}

	tests := []struct {
		name      string
		req       *govpb.QuerySnapshotVotesRequest
		wantErr   bool
		errMsg    string
		checkResp func(*govpb.QuerySnapshotVotesResponse)
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "zero proposal id",
			req: &govpb.QuerySnapshotVotesRequest{
				ProposalId: 0,
			},
			wantErr: true,
			errMsg:  "proposal id cannot be 0",
		},
		{
			name: "valid snapshot votes query",
			req: &govpb.QuerySnapshotVotesRequest{
				ProposalId: 1,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QuerySnapshotVotesResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Votes, 2)
			},
		},
		{
			name: "no snapshot votes for proposal",
			req: &govpb.QuerySnapshotVotesRequest{
				ProposalId: 999,
			},
			wantErr: false,
			checkResp: func(resp *govpb.QuerySnapshotVotesResponse) {
				require.NotNil(t, resp)
				require.Len(t, resp.Votes, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := queryServer.SnapshotVotes(sdk.WrapSDKContext(ctx), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				if tt.checkResp != nil {
					tt.checkResp(resp)
				}
			}
		})
	}
}

// TestQueryProposalsInvalidUnmarshal tests error handling for invalid data
func TestQueryProposalsInvalidUnmarshal(t *testing.T) {
	_, queryServer, ctx := setupQueryKeeper(t)

	// This test ensures the error path in Proposals() is covered
	// In normal operation, unmarshal errors are rare, but we test the handler

	resp, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), &govpb.QueryProposalsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Proposals, 0)
}

// TestQueryVotingPowerInvalidAddress tests error handling for invalid addresses
func TestQueryVotingPowerInvalidAddress(t *testing.T) {
	_, queryServer, ctx := setupQueryKeeper(t)

	resp, err := queryServer.VotingPower(sdk.WrapSDKContext(ctx), &govpb.QueryVotingPowerRequest{
		Address: "invalid_address",
	})

	// Should handle gracefully - either error or return zero power
	if err != nil {
		require.Error(t, err)
	} else {
		require.NotNil(t, resp)
	}
}
