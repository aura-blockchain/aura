// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/governance/types"
)

// ============================
// PROPOSAL DEPOSIT REFUND LOGIC (Feature 7)
// ============================

// ProcessDepositRefunds processes deposit refunds for a finalized proposal
func (k *Keeper) ProcessDepositRefunds(ctx sdk.Context, proposalID uint64) error {
	proposal, err := k.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)

	// Determine refund policy based on proposal outcome
	shouldRefund := k.shouldRefundDeposits(proposal, params)
	refundPercentage := k.getRefundPercentage(proposal, params)

	deposits := k.GetDeposits(ctx, proposalID)

	for _, deposit := range deposits {
		if shouldRefund {
			refundAmount := k.calculateRefundAmount(deposit.Amount, refundPercentage)

			// Process refund (simplified - would interact with bank module)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"deposit_refunded",
					sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
					sdk.NewAttribute("depositor", deposit.Depositor),
					sdk.NewAttribute("amount", refundAmount),
					sdk.NewAttribute("percentage", fmt.Sprintf("%d", refundPercentage)),
				),
			)
		} else {
			// Burn or redirect to community pool
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"deposit_slashed",
					sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
					sdk.NewAttribute("depositor", deposit.Depositor),
					sdk.NewAttribute("amount", deposit.Amount),
				),
			)
		}
	}

	return nil
}

// shouldRefundDeposits determines if deposits should be refunded
func (k *Keeper) shouldRefundDeposits(proposal *types.Proposal, params *types.GovernanceParams) bool {
	switch proposal.Status {
	case types.ProposalStatus_PROPOSAL_STATUS_PASSED, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED:
		return true // Full refund for passed proposals
	case types.ProposalStatus_PROPOSAL_STATUS_FAILED:
		return types.GetRefundFailedProposals(params) // Configurable
	case types.ProposalStatus_PROPOSAL_STATUS_REJECTED:
		return false // No refund for rejected (vetoed) proposals
	default:
		return false
	}
}

// getRefundPercentage returns refund percentage based on proposal outcome
func (k *Keeper) getRefundPercentage(proposal *types.Proposal, params *types.GovernanceParams) uint64 {
	switch proposal.Status {
	case types.ProposalStatus_PROPOSAL_STATUS_PASSED, types.ProposalStatus_PROPOSAL_STATUS_EXECUTED:
		return 10000 // 100% refund
	case types.ProposalStatus_PROPOSAL_STATUS_FAILED:
		return types.GetFailedProposalRefundPercentage(params) // Partial refund
	case types.ProposalStatus_PROPOSAL_STATUS_REJECTED:
		return 0 // No refund
	default:
		return 0
	}
}

// calculateRefundAmount calculates refund amount based on percentage
func (k *Keeper) calculateRefundAmount(depositAmount string, percentage uint64) string {
	amount := new(big.Int)
	amount.SetString(depositAmount, 10)

	refund := new(big.Int).Mul(amount, big.NewInt(int64(percentage)))
	refund.Div(refund, big.NewInt(10000))

	return refund.String()
}
