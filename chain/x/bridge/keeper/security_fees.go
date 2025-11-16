package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// BRIDGE TRANSFER FEES
// ============================================================================

// CalculateTransferFee calculates the fee for a bridge transfer
func (k Keeper) CalculateTransferFee(
	ctx sdk.Context,
	amount sdk.Int,
	feeType types.FeeType,
) (sdk.Int, error) {
	params := k.GetParams(ctx)

	// Get base fees
	fixedFee := params.FixedTransferFee
	percentageBPS := params.PercentageFeeBPS

	// Calculate percentage fee
	percentageFee := amount.MulRaw(int64(percentageBPS)).QuoRaw(10000)

	// Total fee is fixed + percentage
	totalFee := fixedFee.Add(percentageFee)

	// Ensure fee is reasonable (not more than 5% of amount)
	maxFee := amount.MulRaw(5).QuoRaw(100)
	if totalFee.GT(maxFee) {
		totalFee = maxFee
	}

	// Minimum fee (to prevent dust attacks)
	minFee := sdk.NewInt(100000) // 0.1 token minimum
	if totalFee.LT(minFee) {
		totalFee = minFee
	}

	return totalFee, nil
}

// CollectTransferFee collects fees from a bridge transfer
func (k Keeper) CollectTransferFee(
	ctx sdk.Context,
	sender sdk.AccAddress,
	amount sdk.Int,
	feeType types.FeeType,
) (sdk.Int, error) {
	fee, err := k.CalculateTransferFee(ctx, amount, feeType)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// Collect fee from sender
	feeCoins := sdk.NewCoins(sdk.NewCoin("aura", fee))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		types.ModuleName,
		feeCoins,
	); err != nil {
		return sdk.ZeroInt(), fmt.Errorf("failed to collect fee: %w", err)
	}

	// Distribute fee to insurance fund and other recipients
	k.DistributeFees(ctx, fee)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_fee_collected",
			sdk.NewAttribute("sender", sender.String()),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("fee", fee.String()),
			sdk.NewAttribute("fee_type", feeType.String()),
		),
	)

	return fee, nil
}

// DistributeFees distributes collected fees
func (k Keeper) DistributeFees(ctx sdk.Context, totalFee sdk.Int) {
	params := k.GetParams(ctx)

	// Calculate insurance fund portion
	insurancePortion := totalFee.MulRaw(int64(params.InsuranceFundContributionBPS)).QuoRaw(10000)

	// Remaining goes to validators/stakers
	validatorPortion := totalFee.Sub(insurancePortion)

	// Add to insurance fund
	insuranceFund := k.GetInsuranceFund(ctx)
	if insuranceFund == nil {
		insuranceFund = &types.InsuranceFund{
			TotalBalance:        sdk.ZeroInt(),
			TotalClaimsPaid:     sdk.ZeroInt(),
			PendingClaims:       []*types.InsuranceClaim{},
			ContributionRateBps: params.InsuranceFundContributionBPS,
			LastUpdated:         ctx.BlockTime(),
		}
	}

	insuranceFund.TotalBalance = insuranceFund.TotalBalance.Add(insurancePortion)
	insuranceFund.LastUpdated = ctx.BlockTime()
	k.SetInsuranceFund(ctx, insuranceFund)

	// TODO: Distribute validator portion to stakers
	// For now, keep it in the module account

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fees_distributed",
			sdk.NewAttribute("total_fee", totalFee.String()),
			sdk.NewAttribute("insurance_portion", insurancePortion.String()),
			sdk.NewAttribute("validator_portion", validatorPortion.String()),
		),
	)
}

// SetBridgeFee sets the fee structure for a specific fee type
func (k Keeper) SetBridgeFee(ctx sdk.Context, fee *types.BridgeFee) {
	store := ctx.KVStore(k.storeKey)
	key := types.BridgeFeeKey(fee.FeeType)

	bz := k.cdc.MustMarshal(fee)
	store.Set(key, bz)
}

// GetBridgeFee retrieves the fee structure for a fee type
func (k Keeper) GetBridgeFee(ctx sdk.Context, feeType types.FeeType) *types.BridgeFee {
	store := ctx.KVStore(k.storeKey)
	key := types.BridgeFeeKey(feeType)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var fee types.BridgeFee
	k.cdc.MustUnmarshal(bz, &fee)
	return &fee
}

// ============================================================================
// INSURANCE FUND
// ============================================================================

// GetInsuranceFund retrieves the insurance fund
func (k Keeper) GetInsuranceFund(ctx sdk.Context) *types.InsuranceFund {
	store := ctx.KVStore(k.storeKey)
	key := types.InsuranceFundKey()

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var fund types.InsuranceFund
	k.cdc.MustUnmarshal(bz, &fund)
	return &fund
}

// SetInsuranceFund stores the insurance fund
func (k Keeper) SetInsuranceFund(ctx sdk.Context, fund *types.InsuranceFund) {
	store := ctx.KVStore(k.storeKey)
	key := types.InsuranceFundKey()

	bz := k.cdc.MustMarshal(fund)
	store.Set(key, bz)
}

// SubmitInsuranceClaim submits a claim to the insurance fund
func (k Keeper) SubmitInsuranceClaim(
	ctx sdk.Context,
	claimant string,
	transferId string,
	claimAmount sdk.Int,
	reason string,
	evidence []byte,
) (*types.InsuranceClaim, error) {
	insuranceFund := k.GetInsuranceFund(ctx)
	if insuranceFund == nil {
		return nil, fmt.Errorf("insurance fund not initialized")
	}

	// Check if fund has sufficient balance
	if insuranceFund.TotalBalance.LT(claimAmount) {
		return nil, fmt.Errorf("insufficient insurance fund balance")
	}

	claimId := fmt.Sprintf("claim-%s-%d", transferId, ctx.BlockHeight())

	claim := &types.InsuranceClaim{
		ClaimId:        claimId,
		Claimant:       claimant,
		TransferId:     transferId,
		ClaimAmount:    claimAmount,
		Reason:         reason,
		Evidence:       evidence,
		Status:         types.ClaimStatus_CLAIM_PENDING,
		SubmittedAt:    ctx.BlockTime(),
		ApprovedAmount: sdk.ZeroInt(),
	}

	// Add to pending claims
	insuranceFund.PendingClaims = append(insuranceFund.PendingClaims, claim)
	k.SetInsuranceFund(ctx, insuranceFund)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"insurance_claim_submitted",
			sdk.NewAttribute("claim_id", claimId),
			sdk.NewAttribute("claimant", claimant),
			sdk.NewAttribute("amount", claimAmount.String()),
			sdk.NewAttribute("reason", reason),
		),
	)

	return claim, nil
}

// ApproveInsuranceClaim approves and pays out an insurance claim
func (k Keeper) ApproveInsuranceClaim(
	ctx sdk.Context,
	claimId string,
	approvedAmount sdk.Int,
) error {
	insuranceFund := k.GetInsuranceFund(ctx)
	if insuranceFund == nil {
		return fmt.Errorf("insurance fund not initialized")
	}

	// Find the claim
	var claim *types.InsuranceClaim
	claimIndex := -1
	for i, c := range insuranceFund.PendingClaims {
		if c.ClaimId == claimId {
			claim = c
			claimIndex = i
			break
		}
	}

	if claim == nil {
		return fmt.Errorf("claim not found: %s", claimId)
	}

	if claim.Status != types.ClaimStatus_CLAIM_PENDING &&
		claim.Status != types.ClaimStatus_CLAIM_INVESTIGATING {
		return fmt.Errorf("claim is not pending: %s", claim.Status.String())
	}

	// Check if fund has sufficient balance
	if insuranceFund.TotalBalance.LT(approvedAmount) {
		return fmt.Errorf("insufficient insurance fund balance")
	}

	// Pay out the claim
	claimantAddr, err := sdk.AccAddressFromBech32(claim.Claimant)
	if err != nil {
		return fmt.Errorf("invalid claimant address: %w", err)
	}

	coins := sdk.NewCoins(sdk.NewCoin("aura", approvedAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		claimantAddr,
		coins,
	); err != nil {
		return fmt.Errorf("failed to pay claim: %w", err)
	}

	// Update claim status
	claim.Status = types.ClaimStatus_CLAIM_PAID
	claim.ApprovedAmount = approvedAmount
	claim.ResolvedAt = ctx.BlockTime()

	// Update insurance fund
	insuranceFund.TotalBalance = insuranceFund.TotalBalance.Sub(approvedAmount)
	insuranceFund.TotalClaimsPaid = insuranceFund.TotalClaimsPaid.Add(approvedAmount)

	// Remove from pending claims
	insuranceFund.PendingClaims = append(
		insuranceFund.PendingClaims[:claimIndex],
		insuranceFund.PendingClaims[claimIndex+1:]...,
	)

	k.SetInsuranceFund(ctx, insuranceFund)

	// Store resolved claim separately for history
	k.SetInsuranceClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"insurance_claim_approved",
			sdk.NewAttribute("claim_id", claimId),
			sdk.NewAttribute("claimant", claim.Claimant),
			sdk.NewAttribute("approved_amount", approvedAmount.String()),
		),
	)

	return nil
}

// RejectInsuranceClaim rejects an insurance claim
func (k Keeper) RejectInsuranceClaim(ctx sdk.Context, claimId string, reason string) error {
	insuranceFund := k.GetInsuranceFund(ctx)
	if insuranceFund == nil {
		return fmt.Errorf("insurance fund not initialized")
	}

	// Find and remove the claim
	var claim *types.InsuranceClaim
	claimIndex := -1
	for i, c := range insuranceFund.PendingClaims {
		if c.ClaimId == claimId {
			claim = c
			claimIndex = i
			break
		}
	}

	if claim == nil {
		return fmt.Errorf("claim not found: %s", claimId)
	}

	// Update claim status
	claim.Status = types.ClaimStatus_CLAIM_REJECTED
	claim.ResolvedAt = ctx.BlockTime()

	// Remove from pending claims
	insuranceFund.PendingClaims = append(
		insuranceFund.PendingClaims[:claimIndex],
		insuranceFund.PendingClaims[claimIndex+1:]...,
	)

	k.SetInsuranceFund(ctx, insuranceFund)
	k.SetInsuranceClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"insurance_claim_rejected",
			sdk.NewAttribute("claim_id", claimId),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// SetInsuranceClaim stores a resolved insurance claim
func (k Keeper) SetInsuranceClaim(ctx sdk.Context, claim *types.InsuranceClaim) {
	store := ctx.KVStore(k.storeKey)
	key := types.InsuranceClaimKey(claim.ClaimId)

	bz := k.cdc.MustMarshal(claim)
	store.Set(key, bz)
}

// GetInsuranceClaim retrieves an insurance claim
func (k Keeper) GetInsuranceClaim(ctx sdk.Context, claimId string) *types.InsuranceClaim {
	store := ctx.KVStore(k.storeKey)
	key := types.InsuranceClaimKey(claimId)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var claim types.InsuranceClaim
	k.cdc.MustUnmarshal(bz, &claim)
	return &claim
}
