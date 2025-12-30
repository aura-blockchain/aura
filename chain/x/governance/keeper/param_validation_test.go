// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupParamValidationKeeper(t *testing.T) (*Keeper, sdk.Context, *MockStakingKeeper) {
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
	return keeper, ctx, mockStaking
}

func TestValidateGovernanceParams_Valid(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	params := types.DefaultParams()
	err := keeper.ValidateGovernanceParams(params)
	require.NoError(t, err)
}

func TestValidateGovernanceParams_Invalid(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	tests := []struct {
		name      string
		params    *types.GovernanceParams
		wantError bool
	}{
		{
			name:      "nil params",
			params:    nil,
			wantError: true,
		},
		{
			name: "empty min deposit",
			params: &types.GovernanceParams{
				MinDeposit:       "",
				MaxDepositPeriod: gogotypes.DurationProto(172800000000000),
				VotingPeriod:     gogotypes.DurationProto(604800000000000),
				Quorum:           "0.334",
				Threshold:        "0.5",
				VetoThreshold:    "0.334",
			},
			wantError: true,
		},
		{
			name: "empty quorum",
			params: &types.GovernanceParams{
				MinDeposit:       "10000000",
				MaxDepositPeriod: gogotypes.DurationProto(172800000000000),
				VotingPeriod:     gogotypes.DurationProto(604800000000000),
				Quorum:           "",
				Threshold:        "0.5",
				VetoThreshold:    "0.334",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.ValidateGovernanceParams(tt.params)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateGovernanceParams_Success(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Get current params
	oldParams := keeper.GetParams(ctx)
	require.NotNil(t, oldParams)

	// Create new params with different values
	newParams := types.DefaultParams()
	newParams.MinDeposit = "20000000"

	// Update params
	err := keeper.UpdateGovernanceParams(ctx, newParams)
	require.NoError(t, err)

	// Verify params were updated
	updatedParams := keeper.GetParams(ctx)
	require.Equal(t, "20000000", updatedParams.MinDeposit)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	// Find the governance_params_updated event
	foundEvent := false
	for _, event := range events {
		if event.Type == "governance_params_updated" {
			foundEvent = true
			break
		}
	}
	require.True(t, foundEvent, "governance_params_updated event should be emitted")
}

func TestUpdateGovernanceParams_InvalidParams(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Try to update with invalid params
	invalidParams := &types.GovernanceParams{
		MinDeposit:       "", // Invalid - empty
		MaxDepositPeriod: gogotypes.DurationProto(172800000000000),
		VotingPeriod:     gogotypes.DurationProto(604800000000000),
		Quorum:           "0.334",
		Threshold:        "0.5",
		VetoThreshold:    "0.334",
	}

	err := keeper.UpdateGovernanceParams(ctx, invalidParams)
	require.Error(t, err)

	// Verify old params are still set
	params := keeper.GetParams(ctx)
	require.NotEmpty(t, params.MinDeposit)
}

func TestValidateProposalContent_ValidTitle(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	err := keeper.ValidateProposalContent(
		"Valid Proposal Title",
		"Valid proposal description",
		types.CategoryText,
	)
	require.NoError(t, err)
}

func TestValidateProposalContent_EmptyTitle(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	err := keeper.ValidateProposalContent(
		"",
		"Valid proposal description",
		types.CategoryText,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "title cannot be empty")
}

func TestValidateProposalContent_TitleTooLong(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	// Create a title longer than 200 characters
	longTitle := ""
	for i := 0; i < 201; i++ {
		longTitle += "a"
	}

	err := keeper.ValidateProposalContent(
		longTitle,
		"Valid proposal description",
		types.CategoryText,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "title too long")
}

func TestValidateProposalContent_EmptyDescription(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	err := keeper.ValidateProposalContent(
		"Valid Title",
		"",
		types.CategoryText,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "description cannot be empty")
}

func TestValidateProposalContent_DescriptionTooLong(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	// Create a description longer than 10000 characters
	longDesc := ""
	for i := 0; i < 10001; i++ {
		longDesc += "a"
	}

	err := keeper.ValidateProposalContent(
		"Valid Title",
		longDesc,
		types.CategoryText,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "description too long")
}

func TestValidateProposalContent_AllCategories(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	categories := []types.ProposalCategory{
		types.CategoryText,
		types.CategoryParameterChange,
		types.CategorySoftwareUpgrade,
		types.CategorySpending,
	}

	for _, category := range categories {
		t.Run(category.String(), func(t *testing.T) {
			err := keeper.ValidateProposalContent(
				"Test Proposal",
				"Test Description",
				category,
			)
			require.NoError(t, err)
		})
	}
}

func TestValidateVote_Success(t *testing.T) {
	keeper, ctx, mockStaking := setupParamValidationKeeper(t)

	// Create a proposal in voting period
	proposal := validProposal(1, types.StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	voter := testAddr("voter1")

	// Set up voting power for the voter
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	// Validate vote
	err := keeper.ValidateVote(ctx, 1, voter, types.VoteOptionYes)
	require.NoError(t, err)
}

func TestValidateVote_ProposalNotFound(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	voter := testAddr("voter1")

	// Try to validate vote on non-existent proposal
	err := keeper.ValidateVote(ctx, 999, voter, types.VoteOptionYes)
	require.Error(t, err)
}

func TestValidateVote_InvalidProposalStatus(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Create a proposal not in voting period
	proposal := validProposal(1, types.StatusDepositPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	voter := testAddr("voter1")

	// Try to validate vote - should fail because not in voting period
	err := keeper.ValidateVote(ctx, 1, voter, types.VoteOptionYes)
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidProposalStatus, err)
}

func TestValidateVote_InvalidVoteOption(t *testing.T) {
	keeper, ctx, mockStaking := setupParamValidationKeeper(t)

	// Create a proposal in voting period
	proposal := validProposal(1, types.StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	voter := testAddr("voter1")

	// Set up voting power for the voter
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	// Try to validate vote with invalid option
	err := keeper.ValidateVote(ctx, 1, voter, types.VoteOption(999))
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidVoteOption, err)
}

func TestValidateVote_NoVotingPower(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Create a proposal in voting period
	proposal := validProposal(1, types.StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	// Use an address with no voting power (mock staking keeper returns zero for unknown addresses)
	voter := testAddr("novotingpower")

	// Validate vote - should fail due to no voting power
	err := keeper.ValidateVote(ctx, 1, voter, types.VoteOptionYes)
	require.Error(t, err)
	require.Equal(t, types.ErrNoVotingPower, err)
}

func TestIsValidVoteOption_AllValidOptions(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	validOptions := []types.VoteOption{
		types.OptionYes,
		types.OptionNo,
		types.OptionAbstain,
		types.OptionNoWithVeto,
	}

	for _, option := range validOptions {
		t.Run(option.String(), func(t *testing.T) {
			require.True(t, keeper.isValidVoteOption(option))
		})
	}
}

func TestIsValidVoteOption_InvalidOption(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	invalidOption := types.VoteOption(999)
	require.False(t, keeper.isValidVoteOption(invalidOption))
}

func TestValidateDeposit_Success(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Create a proposal in deposit period
	proposal := validProposal(1, types.StatusDepositPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	depositor := testAddr("depositor1")

	// Validate deposit
	err := keeper.ValidateDeposit(ctx, 1, depositor, "1000")
	require.NoError(t, err)
}

func TestValidateDeposit_ProposalNotFound(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	depositor := testAddr("depositor1")

	// Try to validate deposit on non-existent proposal
	err := keeper.ValidateDeposit(ctx, 999, depositor, "1000")
	require.Error(t, err)
}

func TestValidateDeposit_InvalidProposalStatus(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Create a proposal in voting period (not deposit period)
	proposal := validProposal(1, types.StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	depositor := testAddr("depositor1")

	// Try to validate deposit - should fail because not in deposit period
	err := keeper.ValidateDeposit(ctx, 1, depositor, "1000")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidProposalStatus, err)
}

func TestValidateDeposit_EmptyAmount(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Create a proposal in deposit period
	proposal := validProposal(1, types.StatusDepositPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	depositor := testAddr("depositor1")

	// Try to validate deposit with empty amount
	err := keeper.ValidateDeposit(ctx, 1, depositor, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be positive")
}

func TestValidateDeposit_ZeroAmount(t *testing.T) {
	keeper, ctx, _ := setupParamValidationKeeper(t)

	// Create a proposal in deposit period
	proposal := validProposal(1, types.StatusDepositPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	depositor := testAddr("depositor1")

	// Try to validate deposit with zero amount
	err := keeper.ValidateDeposit(ctx, 1, depositor, "0")
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be positive")
}

func TestGetParameterValidationRules(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	rules := keeper.GetParameterValidationRules()
	require.NotNil(t, rules)

	// Verify voting period rules
	require.Equal(t, uint64(3600), rules.MinVotingPeriod)
	require.Equal(t, uint64(2592000), rules.MaxVotingPeriod)

	// Verify deposit period rules
	require.Equal(t, uint64(3600), rules.MinDepositPeriod)
	require.Equal(t, uint64(604800), rules.MaxDepositPeriod)

	// Verify quorum rules
	require.Equal(t, uint64(0), rules.MinQuorum)
	require.Equal(t, uint64(10000), rules.MaxQuorum)

	// Verify threshold rules
	require.Equal(t, uint64(0), rules.MinThreshold)
	require.Equal(t, uint64(10000), rules.MaxThreshold)

	// Verify veto threshold rules
	require.Equal(t, uint64(0), rules.MinVetoThreshold)
	require.Equal(t, uint64(10000), rules.MaxVetoThreshold)

	// Verify execution delay
	require.Equal(t, uint64(604800), rules.MaxExecutionDelay)

	// Verify delegation limits
	require.Equal(t, uint64(100), rules.MaxDelegationsPerUser)
	require.Equal(t, uint64(1), rules.MinVoteCreditsPerToken)

	// Verify content limits
	require.Equal(t, uint64(200), rules.MaxProposalTitleLength)
	require.Equal(t, uint64(10000), rules.MaxProposalDescLength)
}

func TestValidateProposalContent_BoundaryTitleLength(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	// Test exactly 200 characters (should pass)
	title200 := ""
	for i := 0; i < 200; i++ {
		title200 += "a"
	}
	err := keeper.ValidateProposalContent(title200, "Description", types.CategoryText)
	require.NoError(t, err)

	// Test 201 characters (should fail)
	title201 := title200 + "a"
	err = keeper.ValidateProposalContent(title201, "Description", types.CategoryText)
	require.Error(t, err)
	require.Contains(t, err.Error(), "title too long")
}

func TestValidateProposalContent_BoundaryDescriptionLength(t *testing.T) {
	keeper, _, _ := setupParamValidationKeeper(t)

	// Test exactly 10000 characters (should pass)
	desc10000 := ""
	for i := 0; i < 10000; i++ {
		desc10000 += "a"
	}
	err := keeper.ValidateProposalContent("Title", desc10000, types.CategoryText)
	require.NoError(t, err)

	// Test 10001 characters (should fail)
	desc10001 := desc10000 + "a"
	err = keeper.ValidateProposalContent("Title", desc10001, types.CategoryText)
	require.Error(t, err)
	require.Contains(t, err.Error(), "description too long")
}

func TestValidateVote_AllVoteOptions(t *testing.T) {
	keeper, ctx, mockStaking := setupParamValidationKeeper(t)

	// Create a proposal in voting period
	proposal := validProposal(1, types.StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)
	keeper.SetNextProposalID(ctx, 2)

	voter := testAddr("voter1")

	// Set up voting power for the voter
	mockStaking.SetDelegatorBonded(voter, sdkmath.NewInt(1000000))

	options := []types.VoteOption{
		types.OptionYes,
		types.OptionNo,
		types.OptionAbstain,
		types.OptionNoWithVeto,
	}

	for _, option := range options {
		t.Run(option.String(), func(t *testing.T) {
			err := keeper.ValidateVote(ctx, 1, voter, option)
			require.NoError(t, err)
		})
	}
}
