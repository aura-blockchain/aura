package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/common/determinism"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// PurchaseInsurance allows a wallet to purchase insurance coverage
func (k Keeper) PurchaseInsurance(ctx context.Context, walletID string, coverageAmount math.Int, premium math.Int) (*wsproto.InsurancePolicy, error) {
	policyID := fmt.Sprintf("policy_%s_%d", walletID, determinism.GetBlockTime(ctx).UnixNano())

	policy := &wsproto.InsurancePolicy{
		PolicyId:       policyID,
		WalletId:       walletID,
		CoverageAmount: coverageAmount.String(),
		Premium:        premium.String(),
		Active:         true,
		PurchasedAt:    blockTimeToGogoTimestamp(ctx),
		ExpiresAt:      blockTimeWithOffsetToGogoTimestamp(ctx, 365 * 24 * time.Hour), // 1 year
		ClaimsPaid:     "0",
	}

	policyBytes, err := k.cdc.Marshal(policy)
	if err != nil {
		return nil, err
	}

	kvStore := k.getStore(ctx)
	key := []byte(fmt.Sprintf("insurance_%s", policyID))
	if err := kvStore.Set(key, policyBytes); err != nil {
		return nil, err
	}

	return policy, nil
}

// FileClaim files an insurance claim
func (k Keeper) FileClaim(ctx context.Context, policyID, reason string, claimAmount math.Int) (string, error) {
	// Get policy
	kvStore := k.getStore(ctx)
	key := []byte(fmt.Sprintf("insurance_%s", policyID))

	policyBytes, _ := kvStore.Get(key)
	if policyBytes == nil {
		return "", fmt.Errorf("policy not found")
	}

	var policy wsproto.InsurancePolicy
	if err := k.cdc.Unmarshal(policyBytes, &policy); err != nil {
		return "", err
	}

	if !policy.Active {
		return "", fmt.Errorf("policy not active")
	}

	// Create claim
	claimID := fmt.Sprintf("claim_%s_%d", policyID, determinism.GetBlockTime(ctx).UnixNano())

	claim := &wsproto.InsuranceClaim{
		ClaimId:    claimID,
		PolicyId:   policyID,
		WalletId:   policy.WalletId,
		Amount:     claimAmount.String(),
		Reason:     reason,
		Status:     "pending",
		FiledAt:    blockTimeToGogoTimestamp(ctx),
		ApprovedAt: nil,
	}

	claimBytes, err := k.cdc.Marshal(claim)
	if err != nil {
		return "", err
	}

	claimKey := []byte(fmt.Sprintf("claim_%s", claimID))
	if err := kvStore.Set(claimKey, claimBytes); err != nil {
		return "", err
	}

	return claimID, nil
}

// ProcessClaim processes an insurance claim
func (k Keeper) ProcessClaim(ctx context.Context, claimID string, approved bool) error {
	kvStore := k.getStore(ctx)
	key := []byte(fmt.Sprintf("claim_%s", claimID))

	claimBytes, _ := kvStore.Get(key)
	if claimBytes == nil {
		return fmt.Errorf("claim not found")
	}

	var claim wsproto.InsuranceClaim
	if err := k.cdc.Unmarshal(claimBytes, &claim); err != nil {
		return fmt.Errorf("error in ProcessClaim: %w", err)
	}

	if approved {
		claim.Status = "approved"
		claim.ApprovedAt = blockTimeToGogoTimestamp(ctx)

		// Update policy claims paid
		policyKey := []byte(fmt.Sprintf("insurance_%s", claim.PolicyId))
		policyBytes, _ := kvStore.Get(policyKey)

		var policy wsproto.InsurancePolicy
		if err := k.cdc.Unmarshal(policyBytes, &policy); err == nil {
			currentPaid, _ := math.NewIntFromString(policy.ClaimsPaid)
			claimAmount, _ := math.NewIntFromString(claim.Amount)
			newPaid := currentPaid.Add(claimAmount)
			policy.ClaimsPaid = newPaid.String()

			updatedPolicyBytes, _ := k.cdc.Marshal(&policy)
			if err := kvStore.Set(policyKey, updatedPolicyBytes); err != nil {
				return fmt.Errorf("failed to marshal for currentPaid: %w", err)
			}
		}
	} else {
		claim.Status = "denied"
	}

	updatedBytes, err := k.cdc.Marshal(&claim)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return kvStore.Set(key, updatedBytes)
}
