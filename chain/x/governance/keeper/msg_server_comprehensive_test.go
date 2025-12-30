// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
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

// setupMsgServerTest sets up a test keeper and msg server
func setupMsgServerTest(t *testing.T) (*Keeper, govpb.MsgServer, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})

	// Initialize params
	keeper.SetParams(ctx, types.DefaultParams())

	msgServer := NewMsgServerImpl(keeper)
	return keeper, msgServer, ctx
}

// testGoCtx wraps sdk.Context to implement context.Context for msg server calls
func testGoCtx(ctx sdk.Context) context.Context {
	return sdk.WrapSDKContext(ctx)
}

func TestMsgServer_SubmitProposal(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	proposer := testAddr("proposer1")

	tests := []struct {
		name    string
		msg     *govpb.MsgSubmitProposal
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty title",
			msg: &govpb.MsgSubmitProposal{
				Proposer:    proposer,
				Title:       "",
				Description: "Test Description",
				Category:    types.CategoryText,
			},
			wantErr: true,
			errMsg:  "title",
		},
		{
			name: "empty description",
			msg: &govpb.MsgSubmitProposal{
				Proposer:    proposer,
				Title:       "Test Proposal",
				Description: "",
				Category:    types.CategoryText,
			},
			wantErr: true,
			errMsg:  "description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.SubmitProposal(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}

	// Verify proposal count remains unchanged after failed submissions
	proposals := keeper.GetAllProposals(ctx)
	require.Empty(t, proposals)
}

func TestMsgServer_Deposit(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	depositor := testAddr("depositor1")

	// Create a proposal in deposit period
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(48 * time.Hour))
	proposal := &types.Proposal{
		Id:             1,
		Title:          "Test Proposal",
		Description:    "Test Description",
		Proposer:       testAddr("proposer1"),
		Status:         types.StatusDepositPeriod,
		Category:       types.CategoryText,
		SubmitTime:     ts,
		DepositEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	tests := []struct {
		name    string
		msg     *govpb.MsgDeposit
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "non-existent proposal",
			msg: &govpb.MsgDeposit{
				ProposalId: 999,
				Depositor:  depositor,
				Amount:     "1000uaura",
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "invalid amount",
			msg: &govpb.MsgDeposit{
				ProposalId: 1,
				Depositor:  depositor,
				Amount:     "invalid",
			},
			wantErr: true,
			errMsg:  "amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.Deposit(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_Vote(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	voter := testAddr("voter1")

	// Create a proposal in voting period
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	tests := []struct {
		name    string
		msg     *govpb.MsgVote
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "non-existent proposal",
			msg: &govpb.MsgVote{
				ProposalId: 999,
				Voter:      voter,
				Option:     types.VoteOptionYes,
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "valid yes vote",
			msg: &govpb.MsgVote{
				ProposalId: 1,
				Voter:      voter,
				Option:     types.VoteOptionYes,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.Vote(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_VoteWeighted(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	voter := testAddr("voter1")

	// Create a proposal in voting period
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	tests := []struct {
		name    string
		msg     *govpb.MsgVoteWeighted
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "empty options",
			msg: &govpb.MsgVoteWeighted{
				ProposalId: 1,
				Voter:      voter,
				Options:    nil,
			},
			wantErr: true,
			errMsg:  "options",
		},
		{
			name: "non-existent proposal",
			msg: &govpb.MsgVoteWeighted{
				ProposalId: 999,
				Voter:      voter,
				Options: []*types.WeightedVoteOption{
					{Option: types.VoteOptionYes, Weight: "1.0"},
				},
			},
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.VoteWeighted(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_DelegateVote(t *testing.T) {
	_, msgServer, ctx := setupMsgServerTest(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	tests := []struct {
		name    string
		msg     *govpb.MsgDelegateVote
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "self delegation",
			msg: &govpb.MsgDelegateVote{
				Delegator: delegator,
				Delegate:  delegator,
			},
			wantErr: true,
			errMsg:  "self",
		},
		{
			name: "valid delegation",
			msg: &govpb.MsgDelegateVote{
				Delegator: delegator,
				Delegate:  delegate,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.DelegateVote(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_UndelegateVote(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	delegator := testAddr("delegator1")
	delegate := testAddr("delegate1")

	// Create an existing delegation
	ts, _ := gogotypes.TimestampProto(time.Now())
	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       delegate,
		DelegationTime: ts,
		DelegatedPower: "1000",
	}
	keeper.SetVoteDelegation(ctx, delegation)

	tests := []struct {
		name    string
		msg     *govpb.MsgUndelegateVote
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "successful undelegation",
			msg: &govpb.MsgUndelegateVote{
				Delegator: delegator,
				Delegate:  delegate,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.UndelegateVote(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_ExecuteProposal(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	executor := testAddr("executor1")

	// Create a proposal ready for execution
	ts, _ := gogotypes.TimestampProto(time.Now())
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusReadyForExecution,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   ts,
		FinalTallyResult: &types.TallyResult{
			Yes:        "1000",
			No:         "0",
			Abstain:    "0",
			NoWithVeto: "0",
		},
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	tests := []struct {
		name    string
		msg     *govpb.MsgExecuteProposal
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "non-existent proposal",
			msg: &govpb.MsgExecuteProposal{
				ProposalId: 999,
				Executor:   executor,
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "execute ready proposal",
			msg: &govpb.MsgExecuteProposal{
				ProposalId: 1,
				Executor:   executor,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.ExecuteProposal(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_SubmitVeto(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	vetoer := testAddr("vetoer1")

	// Create a proposal in voting period
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	tests := []struct {
		name    string
		msg     *govpb.MsgSubmitVeto
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "non-existent proposal",
			msg: &govpb.MsgSubmitVeto{
				ProposalId: 999,
				Vetoer:     vetoer,
				Reason:     "Test reason",
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "valid veto submission",
			msg: &govpb.MsgSubmitVeto{
				ProposalId: 1,
				Vetoer:     vetoer,
				Reason:     "Test reason for veto",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.SubmitVeto(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgServer_SubmitSnapshotVote(t *testing.T) {
	keeper, msgServer, ctx := setupMsgServerTest(t)

	voter := testAddr("voter1")

	// Create a proposal in voting period with snapshot voting enabled
	ts, _ := gogotypes.TimestampProto(time.Now())
	endTs, _ := gogotypes.TimestampProto(time.Now().Add(7 * 24 * time.Hour))
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          types.StatusVotingPeriod,
		Category:        types.CategoryText,
		SubmitTime:      ts,
		VotingStartTime: ts,
		VotingEndTime:   endTs,
		SnapshotHeight:  100,
	}
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Set voting power snapshot
	keeper.SetVotingPowerSnapshot(ctx, 1, voter, sdkmath.NewInt(1000))

	tests := []struct {
		name    string
		msg     *govpb.MsgSubmitSnapshotVote
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
			errMsg:  "empty request",
		},
		{
			name: "non-existent proposal",
			msg: &govpb.MsgSubmitSnapshotVote{
				ProposalId: 999,
				Voter:      voter,
				Option:     types.VoteOptionYes,
				Signature:  "sig123",
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "valid snapshot vote",
			msg: &govpb.MsgSubmitSnapshotVote{
				ProposalId: 1,
				Voter:      voter,
				Option:     types.VoteOptionYes,
				Signature:  "valid_signature",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := msgServer.SubmitSnapshotVote(testGoCtx(ctx), tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateProposal(t *testing.T) {
	tests := []struct {
		name    string
		msg     *govpb.MsgSubmitProposal
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid proposal",
			msg: &govpb.MsgSubmitProposal{
				Proposer:    testAddr("proposer1"),
				Title:       "Test Proposal",
				Description: "Test Description",
				Category:    types.CategoryText,
			},
			wantErr: false,
		},
		{
			name: "empty title",
			msg: &govpb.MsgSubmitProposal{
				Proposer:    testAddr("proposer1"),
				Title:       "",
				Description: "Test Description",
				Category:    types.CategoryText,
			},
			wantErr: true,
			errMsg:  "title",
		},
		{
			name: "empty description",
			msg: &govpb.MsgSubmitProposal{
				Proposer:    testAddr("proposer1"),
				Title:       "Test",
				Description: "",
				Category:    types.CategoryText,
			},
			wantErr: true,
			errMsg:  "description",
		},
		{
			name: "empty proposer",
			msg: &govpb.MsgSubmitProposal{
				Proposer:    "",
				Title:       "Test",
				Description: "Test Description",
				Category:    types.CategoryText,
			},
			wantErr: true,
			errMsg:  "proposer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProposal(tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
