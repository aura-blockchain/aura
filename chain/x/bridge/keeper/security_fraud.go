package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// FRAUD PROOF SYSTEM
// ============================================================================

// SubmitFraudProof allows anyone to challenge an invalid cross-chain transfer
// Successful challenges earn rewards from the insurance fund
func (k Keeper) SubmitFraudProof(
	ctx sdk.Context,
	challenger string,
	transferId string,
	fraudType types.FraudType,
	evidence []byte,
) (*types.FraudProof, error) {
	params := k.GetParams(ctx)

	// Get the transfer being challenged
	transfer := k.GetTransfer(ctx, transferId)
	if transfer == nil {
		return nil, fmt.Errorf("transfer not found: %s", transferId)
	}

	// Check if transfer is still within challenge window
	timeSinceTransfer := ctx.BlockTime().Sub(transfer.InitiatedAt)
	if timeSinceTransfer > params.FraudProofWindowDuration {
		return nil, fmt.Errorf("transfer is outside fraud proof window")
	}

	// Generate proof ID
	proofId := fmt.Sprintf("fraud-%s-%d", transferId, ctx.BlockHeight())

	// Create fraud proof
	fraudProof := &types.FraudProof{
		ProofId:              proofId,
		ChallengedTransferId: transferId,
		Challenger:           challenger,
		FraudType:            fraudType,
		Evidence:             evidence,
		CounterProof:         nil,
		Status:               types.FraudProofStatus_FRAUD_PROOF_PENDING,
		SubmittedAt:          ctx.BlockTime(),
		ResolvedAt:           time.Time{},
		RewardAmount:         params.FraudProofReward,
	}

	// Store fraud proof
	k.SetFraudProof(ctx, fraudProof)

	// Pause the transfer until investigation completes
	transfer.Status = types.TransferStatus_FAILED
	k.SetTransfer(ctx, transfer)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fraud_proof_submitted",
			sdk.NewAttribute("proof_id", proofId),
			sdk.NewAttribute("transfer_id", transferId),
			sdk.NewAttribute("challenger", challenger),
			sdk.NewAttribute("fraud_type", fraudType.String()),
		),
	)

	return fraudProof, nil
}

// InvestigateFraudProof investigates a fraud proof claim
// This is typically called by governance or authorized validators
func (k Keeper) InvestigateFraudProof(
	ctx sdk.Context,
	proofId string,
	investigator string,
) error {
	fraudProof := k.GetFraudProof(ctx, proofId)
	if fraudProof == nil {
		return fmt.Errorf("fraud proof not found: %s", proofId)
	}

	if fraudProof.Status != types.FraudProofStatus_FRAUD_PROOF_PENDING {
		return fmt.Errorf("fraud proof is not pending: %s", fraudProof.Status.String())
	}

	// Update status to investigating
	fraudProof.Status = types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING
	k.SetFraudProof(ctx, fraudProof)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fraud_proof_investigating",
			sdk.NewAttribute("proof_id", proofId),
			sdk.NewAttribute("investigator", investigator),
		),
	)

	return nil
}

// ResolveFraudProof resolves a fraud proof claim
func (k Keeper) ResolveFraudProof(
	ctx sdk.Context,
	proofId string,
	isValid bool,
	resolution string,
) error {
	params := k.GetParams(ctx)

	fraudProof := k.GetFraudProof(ctx, proofId)
	if fraudProof == nil {
		return fmt.Errorf("fraud proof not found: %s", proofId)
	}

	if fraudProof.Status != types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING {
		return fmt.Errorf("fraud proof is not being investigated: %s", fraudProof.Status.String())
	}

	transfer := k.GetTransfer(ctx, fraudProof.ChallengedTransferId)
	if transfer == nil {
		return fmt.Errorf("transfer not found: %s", fraudProof.ChallengedTransferId)
	}

	if isValid {
		// Fraud proof is valid - challenger wins
		fraudProof.Status = types.FraudProofStatus_FRAUD_PROOF_VALID
		fraudProof.ResolvedAt = ctx.BlockTime()

		// Pay reward to challenger from insurance fund
		challengerAddr, err := sdk.AccAddressFromBech32(fraudProof.Challenger)
		if err == nil {
			k.PayFraudProofReward(ctx, challengerAddr, fraudProof.RewardAmount)
		}

		// Slash the malicious validator who relayed the transfer
		if transfer.RelayerAddress != "" {
			k.SlashValidatorForFraud(ctx, transfer.RelayerAddress, fraudProof.ProofId)
		}

		// Mark transfer as failed and refund if possible
		transfer.Status = types.TransferStatus_REFUNDED

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"fraud_proof_valid",
				sdk.NewAttribute("proof_id", proofId),
				sdk.NewAttribute("challenger", fraudProof.Challenger),
				sdk.NewAttribute("reward", fraudProof.RewardAmount.String()),
				sdk.NewAttribute("slashed_validator", transfer.RelayerAddress),
			),
		)
	} else {
		// Fraud proof is invalid - challenger loses
		fraudProof.Status = types.FraudProofStatus_FRAUD_PROOF_INVALID
		fraudProof.ResolvedAt = ctx.BlockTime()

		// Restore transfer to original status
		transfer.Status = types.TransferStatus_COMPLETED

		// Optional: Penalize frivolous challenges
		// Could slash a small deposit from the challenger

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"fraud_proof_invalid",
				sdk.NewAttribute("proof_id", proofId),
				sdk.NewAttribute("challenger", fraudProof.Challenger),
			),
		)
	}

	k.SetFraudProof(ctx, fraudProof)
	k.SetTransfer(ctx, transfer)

	return nil
}

// PayFraudProofReward pays a reward to a successful fraud proof challenger
func (k Keeper) PayFraudProofReward(
	ctx sdk.Context,
	recipient sdk.AccAddress,
	amount sdk.Int,
) error {
	// Pay from insurance fund
	insuranceFund := k.GetInsuranceFund(ctx)
	if insuranceFund.TotalBalance.LT(amount) {
		return fmt.Errorf("insufficient insurance fund balance")
	}

	coins := sdk.NewCoins(sdk.NewCoin("aura", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		recipient,
		coins,
	); err != nil {
		return fmt.Errorf("failed to pay reward: %w", err)
	}

	// Update insurance fund
	insuranceFund.TotalBalance = insuranceFund.TotalBalance.Sub(amount)
	k.SetInsuranceFund(ctx, insuranceFund)

	return nil
}

// SlashValidatorForFraud slashes a validator for submitting fraudulent transfers
func (k Keeper) SlashValidatorForFraud(
	ctx sdk.Context,
	validatorAddr string,
	proofId string,
) error {
	validator := k.GetBridgeValidator(ctx, validatorAddr)
	if validator == nil {
		return fmt.Errorf("validator not found: %s", validatorAddr)
	}

	slashAmount := sdk.NewInt(10000000000) // Severe: 10,000 tokens

	event := &types.SlashingEvent{
		EventId:          fmt.Sprintf("slash-fraud-%d", ctx.BlockHeight()),
		ValidatorAddress: validatorAddr,
		Reason:           types.SlashReason_SLASH_FRAUD_ATTEMPT,
		SlashAmount:      slashAmount,
		EvidenceHash:     []byte(proofId),
		InfractionHeight: uint64(ctx.BlockHeight()),
		Timestamp:        ctx.BlockTime(),
		Jailed:           true,
	}

	k.SetSlashingEvent(ctx, event)

	// Permanently deactivate validator
	validator.Active = false
	k.SetBridgeValidator(ctx, validator)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_slashed_for_fraud",
			sdk.NewAttribute("validator", validatorAddr),
			sdk.NewAttribute("proof_id", proofId),
			sdk.NewAttribute("amount", slashAmount.String()),
		),
	)

	return nil
}

// CheckExpiredFraudProofs checks and expires old fraud proofs
func (k Keeper) CheckExpiredFraudProofs(ctx sdk.Context) {
	params := k.GetParams(ctx)

	fraudProofs := k.GetAllFraudProofs(ctx)
	for _, proof := range fraudProofs {
		if proof.Status == types.FraudProofStatus_FRAUD_PROOF_PENDING ||
			proof.Status == types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING {

			timeSinceSubmission := ctx.BlockTime().Sub(proof.SubmittedAt)
			if timeSinceSubmission > params.FraudProofWindowDuration*2 {
				// Expire old pending proofs
				proof.Status = types.FraudProofStatus_FRAUD_PROOF_EXPIRED
				proof.ResolvedAt = ctx.BlockTime()
				k.SetFraudProof(ctx, proof)

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"fraud_proof_expired",
						sdk.NewAttribute("proof_id", proof.ProofId),
					),
				)
			}
		}
	}
}

// ============================================================================
// STORAGE
// ============================================================================

// GetFraudProof retrieves a fraud proof
func (k Keeper) GetFraudProof(ctx sdk.Context, proofId string) *types.FraudProof {
	store := ctx.KVStore(k.storeKey)
	key := types.FraudProofKey(proofId)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var proof types.FraudProof
	k.cdc.MustUnmarshal(bz, &proof)
	return &proof
}

// SetFraudProof stores a fraud proof
func (k Keeper) SetFraudProof(ctx sdk.Context, proof *types.FraudProof) {
	store := ctx.KVStore(k.storeKey)
	key := types.FraudProofKey(proof.ProofId)

	bz := k.cdc.MustMarshal(proof)
	store.Set(key, bz)
}

// GetAllFraudProofs retrieves all fraud proofs
func (k Keeper) GetAllFraudProofs(ctx sdk.Context) []*types.FraudProof {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.FraudProofPrefix)
	defer iterator.Close()

	proofs := []*types.FraudProof{}
	for ; iterator.Valid(); iterator.Next() {
		var proof types.FraudProof
		k.cdc.MustUnmarshal(iterator.Value(), &proof)
		proofs = append(proofs, &proof)
	}

	return proofs
}
