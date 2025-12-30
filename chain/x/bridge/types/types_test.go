// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestTransferStatusEnum(t *testing.T) {
	tests := []struct {
		name   string
		status types.TransferStatus
		value  int32
	}{
		{"PENDING", types.TransferStatus_PENDING, 0},
		{"CONFIRMED", types.TransferStatus_CONFIRMED, 1},
		{"RELAYED", types.TransferStatus_RELAYED, 2},
		{"COMPLETED", types.TransferStatus_COMPLETED, 3},
		{"FAILED", types.TransferStatus_FAILED, 4},
		{"REFUNDED", types.TransferStatus_REFUNDED, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.value, int32(tt.status))
		})
	}

	// Verify all statuses are distinct
	require.NotEqual(t, types.TransferStatus_PENDING, types.TransferStatus_CONFIRMED)
	require.NotEqual(t, types.TransferStatus_CONFIRMED, types.TransferStatus_RELAYED)
	require.NotEqual(t, types.TransferStatus_RELAYED, types.TransferStatus_COMPLETED)
	require.NotEqual(t, types.TransferStatus_COMPLETED, types.TransferStatus_FAILED)
	require.NotEqual(t, types.TransferStatus_FAILED, types.TransferStatus_REFUNDED)
}

func TestFeeTypeEnum(t *testing.T) {
	tests := []struct {
		name    string
		feeType types.FeeType
		value   int32
	}{
		{"FEE_TRANSFER", types.FeeType_FEE_TRANSFER, 0},
		{"FEE_MINT_WRAPPED", types.FeeType_FEE_MINT_WRAPPED, 1},
		{"FEE_BURN_WRAPPED", types.FeeType_FEE_BURN_WRAPPED, 2},
		{"FEE_FAST_TRANSFER", types.FeeType_FEE_FAST_TRANSFER, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.value, int32(tt.feeType))
		})
	}
}

func TestFraudProofStatusEnum(t *testing.T) {
	tests := []struct {
		name   string
		status types.FraudProofStatus
	}{
		{"PENDING", types.FraudProofStatus_FRAUD_PROOF_PENDING},
		{"INVESTIGATING", types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING},
		{"VALID", types.FraudProofStatus_FRAUD_PROOF_VALID},
		{"INVALID", types.FraudProofStatus_FRAUD_PROOF_INVALID},
		{"EXPIRED", types.FraudProofStatus_FRAUD_PROOF_EXPIRED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.status)
		})
	}

	// Verify all statuses are distinct
	require.NotEqual(t, types.FraudProofStatus_FRAUD_PROOF_PENDING, types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING)
	require.NotEqual(t, types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING, types.FraudProofStatus_FRAUD_PROOF_VALID)
	require.NotEqual(t, types.FraudProofStatus_FRAUD_PROOF_VALID, types.FraudProofStatus_FRAUD_PROOF_INVALID)
	require.NotEqual(t, types.FraudProofStatus_FRAUD_PROOF_INVALID, types.FraudProofStatus_FRAUD_PROOF_EXPIRED)
}

func TestFraudTypeEnum(t *testing.T) {
	tests := []struct {
		name      string
		fraudType types.FraudType
	}{
		{"INVALID_MERKLE_PROOF", types.FraudType_FRAUD_INVALID_MERKLE_PROOF},
		{"DOUBLE_SPEND", types.FraudType_FRAUD_DOUBLE_SPEND},
		{"INVALID_SIGNATURE", types.FraudType_FRAUD_INVALID_SIGNATURE},
		{"AMOUNT_MISMATCH", types.FraudType_FRAUD_AMOUNT_MISMATCH},
		{"UNAUTHORIZED_MINT", types.FraudType_FRAUD_UNAUTHORIZED_MINT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.fraudType)
		})
	}
}

func TestSlashReasonEnum(t *testing.T) {
	tests := []struct {
		name   string
		reason types.SlashReason
	}{
		{"UNSPECIFIED", types.SlashReason_SLASH_REASON_UNSPECIFIED},
		{"INVALID_PROOF", types.SlashReason_SLASH_INVALID_PROOF},
		{"DOUBLE_SIGN", types.SlashReason_SLASH_DOUBLE_SIGN},
		{"UNAUTHORIZED_MINT", types.SlashReason_SLASH_UNAUTHORIZED_MINT},
		{"FRAUD_ATTEMPT", types.SlashReason_SLASH_FRAUD_ATTEMPT},
		{"DOWNTIME", types.SlashReason_SLASH_DOWNTIME},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.reason)
		})
	}
}

func TestClaimStatusEnum(t *testing.T) {
	tests := []struct {
		name   string
		status types.ClaimStatus
	}{
		{"PENDING", types.ClaimStatus_CLAIM_PENDING},
		{"INVESTIGATING", types.ClaimStatus_CLAIM_INVESTIGATING},
		{"APPROVED", types.ClaimStatus_CLAIM_APPROVED},
		{"REJECTED", types.ClaimStatus_CLAIM_REJECTED},
		{"PAID", types.ClaimStatus_CLAIM_PAID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.status)
		})
	}
}

func TestPermissionTypeEnum(t *testing.T) {
	tests := []struct {
		name       string
		permission types.PermissionType
	}{
		{"NONE", types.PermissionType_PERMISSION_NONE},
		{"WHITELISTED", types.PermissionType_PERMISSION_WHITELISTED},
		{"BLACKLISTED", types.PermissionType_PERMISSION_BLACKLISTED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.permission)
		})
	}
}

func TestRotationStatusEnum(t *testing.T) {
	tests := []struct {
		name   string
		status types.RotationStatus
	}{
		{"PENDING", types.RotationStatus_ROTATION_PENDING},
		{"APPROVED", types.RotationStatus_ROTATION_APPROVED},
		{"ACTIVE", types.RotationStatus_ROTATION_ACTIVE},
		{"EXPIRED", types.RotationStatus_ROTATION_EXPIRED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.status)
		})
	}
}

func TestTimeLockStatusEnum(t *testing.T) {
	tests := []struct {
		name   string
		status types.TimeLockStatus
	}{
		{"LOCKED", types.TimeLockStatus_TIMELOCK_LOCKED},
		{"UNLOCKED", types.TimeLockStatus_TIMELOCK_UNLOCKED},
		{"CHALLENGED", types.TimeLockStatus_TIMELOCK_CHALLENGED},
		{"CANCELLED", types.TimeLockStatus_TIMELOCK_CANCELLED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.status)
		})
	}
}

func TestCachedBridgeStats(t *testing.T) {
	stats := types.CachedBridgeStats{
		TotalTransfers: 100,
		TransfersByStatus: map[string]uint64{
			"completed": 90,
			"pending":   5,
			"failed":    5,
		},
		VolumeByChain: map[string]string{
			"paw": "1000000000",
			"xai": "500000000",
		},
		TotalWrappedTokens: 10,
		ActiveValidators:   5,
		ActiveRelayers:     3,
		LastUpdatedHeight:  12345,
	}

	require.Equal(t, uint64(100), stats.TotalTransfers)
	require.Equal(t, uint64(90), stats.TransfersByStatus["completed"])
	require.Equal(t, "1000000000", stats.VolumeByChain["paw"])
	require.Equal(t, uint64(10), stats.TotalWrappedTokens)
	require.Equal(t, uint64(5), stats.ActiveValidators)
	require.Equal(t, uint64(3), stats.ActiveRelayers)
	require.Equal(t, int64(12345), stats.LastUpdatedHeight)
}

func TestCrossChainTransferFields(t *testing.T) {
	transfer := &types.CrossChainTransfer{
		TransferId:            "transfer-1",
		SourceChain:           "aura",
		TargetChain:           "paw",
		Sender:                "aura1sender",
		Recipient:             "paw1recipient",
		Amount:                sdkmath.NewInt(1000000),
		Denom:                 "uaura",
		Status:                types.TransferStatus_PENDING,
		Confirmations:         2,
		RequiredConfirmations: 3,
	}

	require.Equal(t, "transfer-1", transfer.TransferId)
	require.Equal(t, "aura", transfer.SourceChain)
	require.Equal(t, "paw", transfer.TargetChain)
	require.Equal(t, "aura1sender", transfer.Sender)
	require.Equal(t, "paw1recipient", transfer.Recipient)
	require.True(t, transfer.Amount.Equal(sdkmath.NewInt(1000000)))
	require.Equal(t, "uaura", transfer.Denom)
	require.Equal(t, types.TransferStatus_PENDING, transfer.Status)
	require.Equal(t, uint64(2), transfer.Confirmations)
	require.Equal(t, uint64(3), transfer.RequiredConfirmations)
}

func TestBridgeValidatorFields(t *testing.T) {
	validator := &types.BridgeValidator{
		Address: "aura1validator",
		Power:   100,
	}

	require.Equal(t, "aura1validator", validator.Address)
	require.Equal(t, uint64(100), validator.Power)
}

func TestValidatorSignatureFields(t *testing.T) {
	sig := &types.ValidatorSignature{
		ValidatorAddress: "aura1validator",
		Signature:        []byte("signature-bytes"),
	}

	require.Equal(t, "aura1validator", sig.ValidatorAddress)
	require.Equal(t, []byte("signature-bytes"), sig.Signature)
}

func TestWrappedTokenFields(t *testing.T) {
	token := &types.WrappedToken{
		WrappedDenom:  "wpaw",
		SourceChain:   "paw",
		OriginalDenom: "upaw",
		TotalSupply:   sdkmath.NewInt(1000000000),
	}

	require.Equal(t, "wpaw", token.WrappedDenom)
	require.Equal(t, "paw", token.SourceChain)
	require.Equal(t, "upaw", token.OriginalDenom)
	require.True(t, token.TotalSupply.Equal(sdkmath.NewInt(1000000000)))
}

func TestChainConfigFields(t *testing.T) {
	config := &types.ChainConfig{
		ChainId:          "paw",
		ChainName:        "PAW Chain",
		AddressPrefix:    "paw",
		MinConfirmations: 12,
		Enabled:          true,
	}

	require.Equal(t, "paw", config.ChainId)
	require.Equal(t, "PAW Chain", config.ChainName)
	require.Equal(t, "paw", config.AddressPrefix)
	require.Equal(t, uint64(12), config.MinConfirmations)
	require.True(t, config.Enabled)
}

func TestMerkleProofFields(t *testing.T) {
	proof := &types.MerkleProof{
		Root:  []byte("merkle-root"),
		Proof: [][]byte{[]byte("hash1"), []byte("hash2")},
		Leaf:  []byte("leaf-data"),
	}

	require.Equal(t, []byte("merkle-root"), proof.Root)
	require.Len(t, proof.Proof, 2)
	require.Equal(t, []byte("leaf-data"), proof.Leaf)
}

func TestBridgeFeeFields(t *testing.T) {
	fee := &types.BridgeFee{
		FeeType:          types.FeeType_FEE_TRANSFER,
		FixedFee:         sdkmath.NewInt(1000),
		PercentageFeeBps: 30,
		MinFee:           sdkmath.NewInt(100),
		MaxFee:           sdkmath.NewInt(10000),
		Recipient:        "aura1recipient",
	}

	require.Equal(t, types.FeeType_FEE_TRANSFER, fee.FeeType)
	require.True(t, fee.FixedFee.Equal(sdkmath.NewInt(1000)))
	require.Equal(t, uint64(30), fee.PercentageFeeBps)
	require.True(t, fee.MinFee.Equal(sdkmath.NewInt(100)))
	require.True(t, fee.MaxFee.Equal(sdkmath.NewInt(10000)))
	require.Equal(t, "aura1recipient", fee.Recipient)
}

func TestSharedIdentityFields(t *testing.T) {
	identity := &types.SharedIdentity{
		Address:         "aura1user",
		ReputationScore: 750,
		AuraIrScore:     85,
	}

	require.Equal(t, "aura1user", identity.Address)
	require.Equal(t, uint64(750), identity.ReputationScore)
	require.Equal(t, uint64(85), identity.AuraIrScore)
}

func TestCrossChainSwapFields(t *testing.T) {
	swap := &types.CrossChainSwap{
		SwapId:          "swap-1",
		Sender:          "aura1sender",
		SourceChain:     "aura",
		TargetChain:     "paw",
		TargetDenom:     "upaw",
		MinTargetAmount: sdkmath.NewInt(900),
	}

	require.Equal(t, "swap-1", swap.SwapId)
	require.Equal(t, "aura1sender", swap.Sender)
	require.Equal(t, "aura", swap.SourceChain)
	require.Equal(t, "paw", swap.TargetChain)
	require.Equal(t, "upaw", swap.TargetDenom)
	require.True(t, swap.MinTargetAmount.Equal(sdkmath.NewInt(900)))
}

func TestRelayerStatsFields(t *testing.T) {
	stats := &types.RelayerStats{
		RelayerAddress:        "aura1relayer",
		TotalTransfersRelayed: 100,
		SuccessfulTransfers:   95,
		FailedTransfers:       5,
		TotalVolume:           sdkmath.NewInt(1000000000),
	}

	require.Equal(t, "aura1relayer", stats.RelayerAddress)
	require.Equal(t, uint64(100), stats.TotalTransfersRelayed)
	require.Equal(t, uint64(95), stats.SuccessfulTransfers)
	require.Equal(t, uint64(5), stats.FailedTransfers)
	require.True(t, stats.TotalVolume.Equal(sdkmath.NewInt(1000000000)))
}

func TestFraudProofFields(t *testing.T) {
	proof := &types.FraudProof{
		ProofId:              "fraud-1",
		ChallengedTransferId: "transfer-1",
		Challenger:           "aura1challenger",
		FraudType:            types.FraudType_FRAUD_DOUBLE_SPEND,
		Evidence:             []byte("evidence"),
		Status:               types.FraudProofStatus_FRAUD_PROOF_PENDING,
	}

	require.Equal(t, "fraud-1", proof.ProofId)
	require.Equal(t, "transfer-1", proof.ChallengedTransferId)
	require.Equal(t, "aura1challenger", proof.Challenger)
	require.Equal(t, types.FraudType_FRAUD_DOUBLE_SPEND, proof.FraudType)
	require.Equal(t, []byte("evidence"), proof.Evidence)
	require.Equal(t, types.FraudProofStatus_FRAUD_PROOF_PENDING, proof.Status)
}

func TestSlashingEventFields(t *testing.T) {
	event := &types.SlashingEvent{
		EventId:          "slash-1",
		ValidatorAddress: "aura1validator",
		Reason:           types.SlashReason_SLASH_DOUBLE_SIGN,
		SlashAmount:      sdkmath.NewInt(1000000),
		EvidenceHash:     []byte("evidence"),
		InfractionHeight: 12345,
		Jailed:           true,
	}

	require.Equal(t, "slash-1", event.EventId)
	require.Equal(t, "aura1validator", event.ValidatorAddress)
	require.Equal(t, types.SlashReason_SLASH_DOUBLE_SIGN, event.Reason)
	require.True(t, event.SlashAmount.Equal(sdkmath.NewInt(1000000)))
	require.Equal(t, []byte("evidence"), event.EvidenceHash)
	require.Equal(t, uint64(12345), event.InfractionHeight)
	require.True(t, event.Jailed)
}

func TestInsuranceFundFields(t *testing.T) {
	fund := &types.InsuranceFund{
		TotalBalance:        sdkmath.NewInt(1000000000),
		TotalClaimsPaid:     sdkmath.NewInt(100000000),
		ContributionRateBps: 2000,
	}

	require.True(t, fund.TotalBalance.Equal(sdkmath.NewInt(1000000000)))
	require.True(t, fund.TotalClaimsPaid.Equal(sdkmath.NewInt(100000000)))
	require.Equal(t, uint64(2000), fund.ContributionRateBps)
}

func TestInsuranceClaimFields(t *testing.T) {
	claim := &types.InsuranceClaim{
		ClaimId:     "claim-1",
		Claimant:    "aura1claimant",
		TransferId:  "transfer-1",
		ClaimAmount: sdkmath.NewInt(100000),
		Reason:      "transfer lost",
		Evidence:    []byte("evidence"),
		Status:      types.ClaimStatus_CLAIM_PENDING,
	}

	require.Equal(t, "claim-1", claim.ClaimId)
	require.Equal(t, "aura1claimant", claim.Claimant)
	require.Equal(t, "transfer-1", claim.TransferId)
	require.True(t, claim.ClaimAmount.Equal(sdkmath.NewInt(100000)))
	require.Equal(t, "transfer lost", claim.Reason)
	require.Equal(t, []byte("evidence"), claim.Evidence)
	require.Equal(t, types.ClaimStatus_CLAIM_PENDING, claim.Status)
}

func TestCircuitBreakerFields(t *testing.T) {
	cb := &types.CircuitBreaker{
		Status: types.CircuitBreakerStatus_CIRCUIT_OPEN,
	}

	require.Equal(t, types.CircuitBreakerStatus_CIRCUIT_OPEN, cb.Status)
}

func TestTimeLockFields(t *testing.T) {
	tl := &types.TimeLock{
		LockId:     "lock-1",
		TransferId: "transfer-1",
		Status:     types.TimeLockStatus_TIMELOCK_LOCKED,
	}

	require.Equal(t, "lock-1", tl.LockId)
	require.Equal(t, "transfer-1", tl.TransferId)
	require.Equal(t, types.TimeLockStatus_TIMELOCK_LOCKED, tl.Status)
}

func TestWithdrawalLimitFields(t *testing.T) {
	limit := &types.WithdrawalLimit{
		Address:        "aura1user",
		DailyLimit:     sdkmath.NewInt(1000000000),
		WithdrawnToday: sdkmath.NewInt(100000000),
		Tier:           2,
	}

	require.Equal(t, "aura1user", limit.Address)
	require.True(t, limit.DailyLimit.Equal(sdkmath.NewInt(1000000000)))
	require.True(t, limit.WithdrawnToday.Equal(sdkmath.NewInt(100000000)))
	require.Equal(t, uint64(2), limit.Tier)
}

func TestAddressPermissionFields(t *testing.T) {
	perm := &types.AddressPermission{
		Address:        "aura1user",
		PermissionType: types.PermissionType_PERMISSION_WHITELISTED,
		Reason:         "trusted partner",
		AddedBy:        "governance",
	}

	require.Equal(t, "aura1user", perm.Address)
	require.Equal(t, types.PermissionType_PERMISSION_WHITELISTED, perm.PermissionType)
	require.Equal(t, "trusted partner", perm.Reason)
	require.Equal(t, "governance", perm.AddedBy)
}
