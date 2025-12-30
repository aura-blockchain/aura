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
)

func setupInvariantKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := NewMockStakingKeeper()
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	mockSecurity := &MockSecurityKeeper{}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank, mockSecurity)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})
	return keeper, ctx
}

// testAddr generates a valid bech32 address for testing
func testAddr(name string) string {
	// Pad name to 20 bytes for valid AccAddress
	padded := name + "________________"
	return sdk.AccAddress(padded[:20]).String()
}

// validProposal creates a complete valid proposal for testing
func validProposal(id uint64, status types.ProposalStatus) *types.Proposal {
	now := time.Now()
	submitTime, _ := gogotypes.TimestampProto(now)
	votingStart, _ := gogotypes.TimestampProto(now)
	votingEnd, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	return &types.Proposal{
		Id:              id,
		Title:           "Test Proposal",
		Description:     "Test Description",
		Proposer:        testAddr("proposer1"),
		Status:          status,
		Category:        types.CategoryText,
		SubmitTime:      submitTime,
		VotingStartTime: votingStart,
		VotingEndTime:   votingEnd,
	}
}

// validVote creates a complete valid vote for testing
func validVote(proposalID uint64, voterName string) *types.Vote {
	ts, _ := gogotypes.TimestampProto(time.Now())
	return &types.Vote{
		ProposalId:  proposalID,
		Voter:       testAddr(voterName),
		Option:      types.VoteOptionYes,
		VotingPower: "1000",
		Timestamp:   ts,
	}
}

// validDeposit creates a complete valid deposit for testing
func validDeposit(proposalID uint64, depositorName string) *types.Deposit {
	ts, _ := gogotypes.TimestampProto(time.Now())
	return &types.Deposit{
		ProposalId: proposalID,
		Depositor:  testAddr(depositorName),
		Amount:     "1000",
		Timestamp:  ts,
	}
}

func TestRegisterInvariants(t *testing.T) {
	keeper, _ := setupInvariantKeeper(t)

	// Create a mock invariant registry
	registry := &mockInvariantRegistry{routes: make(map[string]sdk.Invariant)}

	// Register invariants
	RegisterInvariants(registry, keeper)

	// Verify all invariants are registered
	require.Contains(t, registry.routes, "governance/params-valid")
	require.Contains(t, registry.routes, "governance/proposal-validity")
	require.Contains(t, registry.routes, "governance/vote-consistency")
	require.Contains(t, registry.routes, "governance/deposit-consistency")
	require.Contains(t, registry.routes, "governance/voting-power-consistency")
}

type mockInvariantRegistry struct {
	routes map[string]sdk.Invariant
}

func (m *mockInvariantRegistry) RegisterRoute(moduleName, route string, invar sdk.Invariant) {
	m.routes[moduleName+"/"+route] = invar
}

func TestAllInvariants(t *testing.T) {
	keeper, ctx := setupInvariantKeeper(t)

	// Initialize with valid params
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	// Run all invariants - should pass with valid state
	invariant := AllInvariants(keeper)
	msg, broken := invariant(ctx)

	require.False(t, broken, "invariants should not be broken with valid state: %s", msg)
	require.Empty(t, msg)
}

func TestParamsInvariant(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Keeper, sdk.Context)
		wantBroken bool
		wantMsg    string
	}{
		{
			name: "valid params",
			setup: func(k *Keeper, ctx sdk.Context) {
				params := types.DefaultParams()
				k.SetParams(ctx, params)
			},
			wantBroken: false,
		},
		{
			name: "empty min deposit",
			setup: func(k *Keeper, ctx sdk.Context) {
				params := &types.GovernanceParams{
					MinDeposit:       "",
					MaxDepositPeriod: gogotypes.DurationProto(172800000000000),
					VotingPeriod:     gogotypes.DurationProto(604800000000000),
					Quorum:           "0.334",
					Threshold:        "0.5",
					VetoThreshold:    "0.334",
				}
				k.SetParams(ctx, params)
			},
			wantBroken: true,
			wantMsg:    "min_deposit is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupInvariantKeeper(t)
			tt.setup(keeper, ctx)

			invariant := ParamsInvariant(keeper)
			msg, broken := invariant(ctx)

			require.Equal(t, tt.wantBroken, broken)
			if tt.wantMsg != "" {
				require.Contains(t, msg, tt.wantMsg)
			}
		})
	}
}

func TestProposalValidityInvariant(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Keeper, sdk.Context)
		wantBroken bool
		wantMsg    string
	}{
		{
			name: "no proposals - valid",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
			},
			wantBroken: false,
		},
		{
			name: "valid proposal",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				k.SetProposal(ctx, validProposal(1, types.StatusVotingPeriod))
			},
			wantBroken: false,
		},
		{
			name: "proposal with empty title",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				proposal := validProposal(1, types.StatusVotingPeriod)
				proposal.Title = "" // Make title empty
				k.SetProposal(ctx, proposal)
			},
			wantBroken: true,
			wantMsg:    "empty title",
		},
		{
			name: "proposal with empty description",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				proposal := validProposal(1, types.StatusVotingPeriod)
				proposal.Description = "" // Make description empty
				k.SetProposal(ctx, proposal)
			},
			wantBroken: true,
			wantMsg:    "empty description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupInvariantKeeper(t)
			tt.setup(keeper, ctx)

			invariant := ProposalValidityInvariant(keeper)
			msg, broken := invariant(ctx)

			require.Equal(t, tt.wantBroken, broken, "invariant msg: %s", msg)
			if tt.wantMsg != "" {
				require.Contains(t, msg, tt.wantMsg)
			}
		})
	}
}

func TestVoteConsistencyInvariant(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Keeper, sdk.Context)
		wantBroken bool
		wantMsg    string
	}{
		{
			name: "no votes - valid",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
			},
			wantBroken: false,
		},
		{
			name: "valid vote",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				k.SetProposal(ctx, validProposal(1, types.StatusVotingPeriod))
				k.SetVote(ctx, validVote(1, "voter1"))
			},
			wantBroken: false,
		},
		{
			name: "vote with nil timestamp",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				k.SetProposal(ctx, validProposal(1, types.StatusVotingPeriod))
				vote := &types.Vote{
					ProposalId:  1,
					Voter:       testAddr("voter1"),
					Option:      types.VoteOptionYes,
					VotingPower: "1000",
					Timestamp:   nil, // Invalid
				}
				k.SetVote(ctx, vote)
			},
			wantBroken: true,
			wantMsg:    "vote has nil timestamp",
		},
		{
			name: "vote with invalid voting power",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				k.SetProposal(ctx, validProposal(1, types.StatusVotingPeriod))
				ts, _ := gogotypes.TimestampProto(time.Now())
				vote := &types.Vote{
					ProposalId:  1,
					Voter:       testAddr("voter1"),
					Option:      types.VoteOptionYes,
					VotingPower: "-100", // Invalid - negative
					Timestamp:   ts,
				}
				k.SetVote(ctx, vote)
			},
			wantBroken: true,
			wantMsg:    "vote has invalid voting power",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupInvariantKeeper(t)
			tt.setup(keeper, ctx)

			invariant := VoteConsistencyInvariant(keeper)
			msg, broken := invariant(ctx)

			require.Equal(t, tt.wantBroken, broken)
			if tt.wantMsg != "" {
				require.Contains(t, msg, tt.wantMsg)
			}
		})
	}
}

func TestDepositConsistencyInvariant(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Keeper, sdk.Context)
		wantBroken bool
		wantMsg    string
	}{
		{
			name: "no deposits - valid",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
			},
			wantBroken: false,
		},
		{
			name: "valid deposit",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				k.SetProposal(ctx, validProposal(1, types.StatusDepositPeriod))
				k.SetDeposit(ctx, validDeposit(1, "depositor1"))
			},
			wantBroken: false,
		},
		{
			name: "deposit with nil timestamp",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				k.SetProposal(ctx, validProposal(1, types.StatusDepositPeriod))
				deposit := &types.Deposit{
					ProposalId: 1,
					Depositor:  testAddr("depositor1"),
					Amount:     "1000",
					Timestamp:  nil, // Invalid
				}
				k.SetDeposit(ctx, deposit)
			},
			wantBroken: true,
			wantMsg:    "deposit has nil timestamp",
		},
		{
			name: "deposit with invalid amount",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				k.SetProposal(ctx, validProposal(1, types.StatusDepositPeriod))
				ts, _ := gogotypes.TimestampProto(time.Now())
				deposit := &types.Deposit{
					ProposalId: 1,
					Depositor:  testAddr("depositor1"),
					Amount:     "-500", // Invalid - negative
					Timestamp:  ts,
				}
				k.SetDeposit(ctx, deposit)
			},
			wantBroken: true,
			wantMsg:    "deposit has invalid amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupInvariantKeeper(t)
			tt.setup(keeper, ctx)

			invariant := DepositConsistencyInvariant(keeper)
			msg, broken := invariant(ctx)

			require.Equal(t, tt.wantBroken, broken)
			if tt.wantMsg != "" {
				require.Contains(t, msg, tt.wantMsg)
			}
		})
	}
}

func TestVotingPowerConsistencyInvariant(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Keeper, sdk.Context)
		wantBroken bool
		wantMsg    string
	}{
		{
			name: "no snapshots - valid",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
			},
			wantBroken: false,
		},
		{
			name: "valid voting power snapshot",
			setup: func(k *Keeper, ctx sdk.Context) {
				k.SetParams(ctx, types.DefaultParams())
				proposal := &types.Proposal{
					Id:          1,
					Title:       "Test Proposal",
					Description: "Test Description",
					Proposer:    testAddr("proposer1"),
					Status:      types.StatusVotingPeriod,
					Category:    types.CategoryText,
				}
				k.SetProposal(ctx, proposal)

				// SetVotingPowerSnapshot takes proposalID, voter, power
				_ = k.SetVotingPowerSnapshot(ctx, 1, testAddr("proposer1"), sdkmath.NewInt(1000))
			},
			wantBroken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupInvariantKeeper(t)
			tt.setup(keeper, ctx)

			invariant := VotingPowerConsistencyInvariant(keeper)
			msg, broken := invariant(ctx)

			require.Equal(t, tt.wantBroken, broken)
			if tt.wantMsg != "" {
				require.Contains(t, msg, tt.wantMsg)
			}
		})
	}
}

func TestAllInvariantsWithBrokenState(t *testing.T) {
	keeper, ctx := setupInvariantKeeper(t)

	// Set params with empty min deposit to break params invariant
	params := &types.GovernanceParams{
		MinDeposit:       "",
		MaxDepositPeriod: gogotypes.DurationProto(172800000000000),
		VotingPeriod:     gogotypes.DurationProto(604800000000000),
		Quorum:           "0.334",
		Threshold:        "0.5",
		VetoThreshold:    "0.334",
	}
	keeper.SetParams(ctx, params)

	// Run all invariants - should fail on params
	invariant := AllInvariants(keeper)
	msg, broken := invariant(ctx)

	require.True(t, broken, "invariants should be broken with invalid params")
	require.Contains(t, msg, "min_deposit is empty")
}
