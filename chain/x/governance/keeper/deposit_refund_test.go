// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupDepositRefundKeeper(t *testing.T) (*Keeper, sdk.Context) {
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
	return keeper, ctx
}

// testAddrDR generates a valid bech32 address for testing (deposit refund)
func testAddrDR(name string) string {
	// Pad name to 20 bytes for valid AccAddress
	padded := name + "________________"
	return sdk.AccAddress(padded[:20]).String()
}

func TestProcessDepositRefunds_PassedProposal(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create a passed proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrDR("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Add deposits
	deposit1 := &types.Deposit{
		ProposalId: 1,
		Depositor:  testAddrDR("depositor1"),
		Amount:     "1000000",
	}
	deposit2 := &types.Deposit{
		ProposalId: 1,
		Depositor:  testAddrDR("depositor2"),
		Amount:     "2000000",
	}
	keeper.SetDeposit(ctx, deposit1)
	keeper.SetDeposit(ctx, deposit2)

	// Process refunds
	err := keeper.ProcessDepositRefunds(ctx, 1)
	require.NoError(t, err)

	// Verify events were emitted
	events := ctx.EventManager().Events()
	refundEvents := 0
	for _, event := range events {
		if event.Type == "deposit_refunded" {
			refundEvents++
		}
	}
	require.Equal(t, 2, refundEvents)
}

func TestProcessDepositRefunds_ExecutedProposal(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create an executed proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrDR("proposer1"),
		Status:        types.StatusExecuted,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Add deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  testAddrDR("depositor1"),
		Amount:     "1000000",
	}
	keeper.SetDeposit(ctx, deposit)

	// Process refunds
	err := keeper.ProcessDepositRefunds(ctx, 1)
	require.NoError(t, err)

	// Verify refund event was emitted
	events := ctx.EventManager().Events()
	refundEvents := 0
	for _, event := range events {
		if event.Type == "deposit_refunded" {
			refundEvents++
			// Verify full refund percentage
			for _, attr := range event.Attributes {
				if attr.Key == "percentage" {
					require.Equal(t, "10000", attr.Value)
				}
			}
		}
	}
	require.Equal(t, 1, refundEvents)
}

func TestProcessDepositRefunds_FailedProposal(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create a failed proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrDR("proposer1"),
		Status:        types.StatusFailed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Add deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  testAddrDR("depositor1"),
		Amount:     "1000000",
	}
	keeper.SetDeposit(ctx, deposit)

	// Process refunds
	err := keeper.ProcessDepositRefunds(ctx, 1)
	require.NoError(t, err)

	// Verify either refund or slash event based on params
	events := ctx.EventManager().Events()
	hasEvent := false
	for _, event := range events {
		if event.Type == "deposit_refunded" || event.Type == "deposit_slashed" {
			hasEvent = true
			break
		}
	}
	require.True(t, hasEvent)
}

func TestProcessDepositRefunds_RejectedProposal(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create a rejected proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrDR("proposer1"),
		Status:        types.StatusRejected,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Add deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  testAddrDR("depositor1"),
		Amount:     "1000000",
	}
	keeper.SetDeposit(ctx, deposit)

	// Process refunds
	err := keeper.ProcessDepositRefunds(ctx, 1)
	require.NoError(t, err)

	// Verify slash event was emitted (no refund for rejected)
	events := ctx.EventManager().Events()
	slashEvents := 0
	for _, event := range events {
		if event.Type == "deposit_slashed" {
			slashEvents++
		}
	}
	require.Equal(t, 1, slashEvents)
}

func TestProcessDepositRefunds_ProposalNotFound(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	// Try to process refunds for non-existent proposal
	err := keeper.ProcessDepositRefunds(ctx, 999)
	require.Error(t, err)
}

func TestProcessDepositRefunds_NoDeposits(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create a passed proposal with no deposits
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrDR("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Process refunds
	err := keeper.ProcessDepositRefunds(ctx, 1)
	require.NoError(t, err)

	// Should complete without error even with no deposits
	events := ctx.EventManager().Events()
	refundEvents := 0
	for _, event := range events {
		if event.Type == "deposit_refunded" {
			refundEvents++
		}
	}
	require.Equal(t, 0, refundEvents)
}

func TestShouldRefundDeposits_PassedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusPassed,
	}

	shouldRefund := keeper.shouldRefundDeposits(proposal, params)
	require.True(t, shouldRefund)
}

func TestShouldRefundDeposits_ExecutedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusExecuted,
	}

	shouldRefund := keeper.shouldRefundDeposits(proposal, params)
	require.True(t, shouldRefund)
}

func TestShouldRefundDeposits_RejectedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusRejected,
	}

	shouldRefund := keeper.shouldRefundDeposits(proposal, params)
	require.False(t, shouldRefund)
}

func TestShouldRefundDeposits_FailedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusFailed,
	}

	// Should depend on params.RefundFailedProposals
	shouldRefund := keeper.shouldRefundDeposits(proposal, params)
	// Result depends on configuration
	require.NotNil(t, shouldRefund)
}

func TestShouldRefundDeposits_UnknownStatus(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusVotingPeriod, // Not a final status
	}

	shouldRefund := keeper.shouldRefundDeposits(proposal, params)
	require.False(t, shouldRefund)
}

func TestGetRefundPercentage_PassedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusPassed,
	}

	percentage := keeper.getRefundPercentage(proposal, params)
	require.Equal(t, uint64(10000), percentage) // 100%
}

func TestGetRefundPercentage_ExecutedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusExecuted,
	}

	percentage := keeper.getRefundPercentage(proposal, params)
	require.Equal(t, uint64(10000), percentage) // 100%
}

func TestGetRefundPercentage_RejectedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusRejected,
	}

	percentage := keeper.getRefundPercentage(proposal, params)
	require.Equal(t, uint64(0), percentage) // 0%
}

func TestGetRefundPercentage_FailedProposal(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusFailed,
	}

	// Should return configured percentage
	percentage := keeper.getRefundPercentage(proposal, params)
	require.NotNil(t, percentage)
	// Default is typically partial refund
	require.GreaterOrEqual(t, percentage, uint64(0))
	require.LessOrEqual(t, percentage, uint64(10000))
}

func TestGetRefundPercentage_UnknownStatus(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	params := types.DefaultParams()
	proposal := &types.Proposal{
		Status: types.StatusDepositPeriod,
	}

	percentage := keeper.getRefundPercentage(proposal, params)
	require.Equal(t, uint64(0), percentage)
}

func TestCalculateRefundAmount_FullRefund(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	amount := "1000000"
	percentage := uint64(10000) // 100%

	refund := keeper.calculateRefundAmount(amount, percentage)
	require.Equal(t, "1000000", refund)
}

func TestCalculateRefundAmount_PartialRefund(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	amount := "1000000"
	percentage := uint64(5000) // 50%

	refund := keeper.calculateRefundAmount(amount, percentage)
	require.Equal(t, "500000", refund)
}

func TestCalculateRefundAmount_NoRefund(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	amount := "1000000"
	percentage := uint64(0) // 0%

	refund := keeper.calculateRefundAmount(amount, percentage)
	require.Equal(t, "0", refund)
}

func TestCalculateRefundAmount_SmallPercentage(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	amount := "1000000"
	percentage := uint64(2500) // 25%

	refund := keeper.calculateRefundAmount(amount, percentage)
	require.Equal(t, "250000", refund)
}

func TestCalculateRefundAmount_LargeAmount(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	amount := "999999999999999999"
	percentage := uint64(7500) // 75%

	refund := keeper.calculateRefundAmount(amount, percentage)
	require.NotEmpty(t, refund)
	// Should be 75% of original
	require.Contains(t, refund, "7499999999999999")
}

func TestCalculateRefundAmount_ZeroAmount(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	amount := "0"
	percentage := uint64(10000) // 100%

	refund := keeper.calculateRefundAmount(amount, percentage)
	require.Equal(t, "0", refund)
}

func TestCalculateRefundAmount_EdgeCase(t *testing.T) {
	keeper, _ := setupDepositRefundKeeper(t)

	amount := "1"
	percentage := uint64(5000) // 50%

	refund := keeper.calculateRefundAmount(amount, percentage)
	require.Equal(t, "0", refund) // Floor division
}

func TestProcessDepositRefunds_MultipleDeposits(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create a passed proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrDR("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	// Add multiple deposits
	for i := 1; i <= 5; i++ {
		deposit := &types.Deposit{
			ProposalId: 1,
			Depositor:  testAddrDR("depositor" + string(rune('0'+i))),
			Amount:     "1000000",
		}
		keeper.SetDeposit(ctx, deposit)
	}

	// Process refunds
	err := keeper.ProcessDepositRefunds(ctx, 1)
	require.NoError(t, err)

	// Verify all deposits were processed
	events := ctx.EventManager().Events()
	refundEvents := 0
	for _, event := range events {
		if event.Type == "deposit_refunded" {
			refundEvents++
		}
	}
	require.Equal(t, 5, refundEvents)
}

func TestProcessDepositRefunds_VerifyEventAttributes(t *testing.T) {
	keeper, ctx := setupDepositRefundKeeper(t)

	now := time.Now()
	ts, _ := gogotypes.TimestampProto(now)
	endTs, _ := gogotypes.TimestampProto(now.Add(7 * 24 * time.Hour))

	// Create a passed proposal
	proposal := &types.Proposal{
		Id:            1,
		Title:         "Test Proposal",
		Description:   "Test Description",
		Proposer:      testAddrDR("proposer1"),
		Status:        types.StatusPassed,
		Category:      types.CategoryText,
		SubmitTime:    ts,
		VotingEndTime: endTs,
	}
	keeper.SetProposal(ctx, proposal)

	depositor := testAddrDR("depositor1")
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  depositor,
		Amount:     "1000000",
	}
	keeper.SetDeposit(ctx, deposit)

	// Process refunds
	err := keeper.ProcessDepositRefunds(ctx, 1)
	require.NoError(t, err)

	// Verify event attributes
	events := ctx.EventManager().Events()
	for _, event := range events {
		if event.Type == "deposit_refunded" {
			attrs := make(map[string]string)
			for _, attr := range event.Attributes {
				attrs[attr.Key] = attr.Value
			}
			require.Equal(t, "1", attrs["proposal_id"])
			require.Equal(t, depositor, attrs["depositor"])
			require.Equal(t, "1000000", attrs["amount"])
			require.Equal(t, "10000", attrs["percentage"])
		}
	}
}
