// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"

// Re-export all proto types
type (
	// Enums
	TransferStatus       = pb.TransferStatus
	FeeType              = pb.FeeType
	FraudProofStatus     = pb.FraudProofStatus
	FraudType            = pb.FraudType
	SlashReason          = pb.SlashReason
	CircuitBreakerStatus = pb.CircuitBreakerStatus
	ClaimStatus          = pb.ClaimStatus
	PermissionType       = pb.PermissionType
	RotationStatus       = pb.RotationStatus
	TimeLockStatus       = pb.TimeLockStatus

	// Core types - Bridge
	CrossChainTransfer = pb.CrossChainTransfer
	CrossChainSwap     = pb.CrossChainSwap
	WrappedToken       = pb.WrappedToken
	ChainConfig        = pb.ChainConfig
	MerkleProof        = pb.MerkleProof
	BridgeFee          = pb.BridgeFee
	NonceTracker       = pb.NonceTracker
	SharedIdentity     = pb.SharedIdentity

	// Core types - Validators
	BridgeValidator      = pb.BridgeValidator
	ValidatorSignature   = pb.ValidatorSignature
	TSSSignature         = pb.TSSSignature
	ValidatorSetRotation = pb.ValidatorSetRotation
	RelayerStats         = pb.RelayerStats

	// Core types - Security
	CircuitBreaker    = pb.CircuitBreaker
	TimeLock          = pb.TimeLock
	PendingTransfer   = pb.PendingTransfer
	FraudProof        = pb.FraudProof
	SlashingEvent     = pb.SlashingEvent
	InsuranceFund     = pb.InsuranceFund
	InsuranceClaim    = pb.InsuranceClaim
	WithdrawalLimit   = pb.WithdrawalLimit
	AddressPermission = pb.AddressPermission

	// Params and Genesis
	BridgeParams = pb.BridgeParams
	GenesisState = pb.GenesisState

	// Message types
	MsgLockTokens             = pb.MsgLockTokens
	MsgLockTokensResponse     = pb.MsgLockTokensResponse
	MsgUnlockTokens           = pb.MsgUnlockTokens
	MsgUnlockTokensResponse   = pb.MsgUnlockTokensResponse
	MsgMintTokens             = pb.MsgMintTokens
	MsgMintTokensResponse     = pb.MsgMintTokensResponse
	MsgBurnTokens             = pb.MsgBurnTokens
	MsgBurnTokensResponse     = pb.MsgBurnTokensResponse
	MsgRelayTransfer          = pb.MsgRelayTransfer
	MsgRelayTransferResponse  = pb.MsgRelayTransferResponse
	MsgCrossChainSwap         = pb.MsgCrossChainSwap
	MsgCrossChainSwapResponse = pb.MsgCrossChainSwapResponse
	MsgLinkAddress            = pb.MsgLinkAddress
	MsgLinkAddressResponse    = pb.MsgLinkAddressResponse

	// Query types
	QueryTransferRequest          = pb.QueryTransferRequest
	QueryTransferResponse         = pb.QueryTransferResponse
	QueryUserTransfersRequest     = pb.QueryUserTransfersRequest
	QueryUserTransfersResponse    = pb.QueryUserTransfersResponse
	QueryAllTransfersRequest      = pb.QueryAllTransfersRequest
	QueryAllTransfersResponse     = pb.QueryAllTransfersResponse
	QueryWrappedTokenRequest      = pb.QueryWrappedTokenRequest
	QueryWrappedTokenResponse     = pb.QueryWrappedTokenResponse
	QueryAllWrappedTokensRequest  = pb.QueryAllWrappedTokensRequest
	QueryAllWrappedTokensResponse = pb.QueryAllWrappedTokensResponse
	QueryChainConfigRequest       = pb.QueryChainConfigRequest
	QueryChainConfigResponse      = pb.QueryChainConfigResponse
	QueryAllChainsRequest         = pb.QueryAllChainsRequest
	QueryAllChainsResponse        = pb.QueryAllChainsResponse
	QueryValidatorsRequest        = pb.QueryValidatorsRequest
	QueryValidatorsResponse       = pb.QueryValidatorsResponse
	QueryRelayerStatsRequest      = pb.QueryRelayerStatsRequest
	QueryRelayerStatsResponse     = pb.QueryRelayerStatsResponse
	QueryBridgeStatsRequest       = pb.QueryBridgeStatsRequest
	QueryBridgeStatsResponse      = pb.QueryBridgeStatsResponse
	QueryCrossChainSwapRequest    = pb.QueryCrossChainSwapRequest
	QueryCrossChainSwapResponse   = pb.QueryCrossChainSwapResponse
	QuerySharedIdentityRequest    = pb.QuerySharedIdentityRequest
	QuerySharedIdentityResponse   = pb.QuerySharedIdentityResponse
)

// Re-export enum values for TransferStatus
const (
	TransferStatus_PENDING   = pb.TransferStatus_PENDING
	TransferStatus_CONFIRMED = pb.TransferStatus_CONFIRMED
	TransferStatus_RELAYED   = pb.TransferStatus_RELAYED
	TransferStatus_COMPLETED = pb.TransferStatus_COMPLETED
	TransferStatus_FAILED    = pb.TransferStatus_FAILED
	TransferStatus_REFUNDED  = pb.TransferStatus_REFUNDED
)

// Re-export enum values for FeeType
const (
	FeeType_FEE_TRANSFER      = pb.FeeType_FEE_TRANSFER
	FeeType_FEE_MINT_WRAPPED  = pb.FeeType_FEE_MINT_WRAPPED
	FeeType_FEE_BURN_WRAPPED  = pb.FeeType_FEE_BURN_WRAPPED
	FeeType_FEE_FAST_TRANSFER = pb.FeeType_FEE_FAST_TRANSFER
)

// Re-export enum values for FraudProofStatus
const (
	FraudProofStatus_FRAUD_PROOF_PENDING       = pb.FraudProofStatus_FRAUD_PROOF_PENDING
	FraudProofStatus_FRAUD_PROOF_INVESTIGATING = pb.FraudProofStatus_FRAUD_PROOF_INVESTIGATING
	FraudProofStatus_FRAUD_PROOF_VALID         = pb.FraudProofStatus_FRAUD_PROOF_VALID
	FraudProofStatus_FRAUD_PROOF_INVALID       = pb.FraudProofStatus_FRAUD_PROOF_INVALID
	FraudProofStatus_FRAUD_PROOF_EXPIRED       = pb.FraudProofStatus_FRAUD_PROOF_EXPIRED
)

// Re-export enum values for FraudType
const (
	FraudType_FRAUD_INVALID_MERKLE_PROOF = pb.FraudType_FRAUD_INVALID_MERKLE_PROOF
	FraudType_FRAUD_DOUBLE_SPEND         = pb.FraudType_FRAUD_DOUBLE_SPEND
	FraudType_FRAUD_INVALID_SIGNATURE    = pb.FraudType_FRAUD_INVALID_SIGNATURE
	FraudType_FRAUD_AMOUNT_MISMATCH      = pb.FraudType_FRAUD_AMOUNT_MISMATCH
	FraudType_FRAUD_UNAUTHORIZED_MINT    = pb.FraudType_FRAUD_UNAUTHORIZED_MINT
)

// Re-export enum values for SlashReason
const (
	SlashReason_SLASH_REASON_UNSPECIFIED = pb.SlashReason_SLASH_REASON_UNSPECIFIED
	SlashReason_SLASH_INVALID_PROOF      = pb.SlashReason_SLASH_INVALID_PROOF
	SlashReason_SLASH_DOUBLE_SIGN        = pb.SlashReason_SLASH_DOUBLE_SIGN
	SlashReason_SLASH_UNAUTHORIZED_MINT  = pb.SlashReason_SLASH_UNAUTHORIZED_MINT
	SlashReason_SLASH_FRAUD_ATTEMPT      = pb.SlashReason_SLASH_FRAUD_ATTEMPT
	SlashReason_SLASH_DOWNTIME           = pb.SlashReason_SLASH_DOWNTIME
)

// Re-export enum values for CircuitBreakerStatus
const (
	CircuitBreakerStatus_CIRCUIT_CLOSED    = pb.CircuitBreakerStatus_CIRCUIT_CLOSED
	CircuitBreakerStatus_CIRCUIT_OPEN      = pb.CircuitBreakerStatus_CIRCUIT_OPEN
	CircuitBreakerStatus_CIRCUIT_HALF_OPEN = pb.CircuitBreakerStatus_CIRCUIT_HALF_OPEN
)

// Re-export enum values for ClaimStatus
const (
	ClaimStatus_CLAIM_PENDING       = pb.ClaimStatus_CLAIM_PENDING
	ClaimStatus_CLAIM_INVESTIGATING = pb.ClaimStatus_CLAIM_INVESTIGATING
	ClaimStatus_CLAIM_APPROVED      = pb.ClaimStatus_CLAIM_APPROVED
	ClaimStatus_CLAIM_REJECTED      = pb.ClaimStatus_CLAIM_REJECTED
	ClaimStatus_CLAIM_PAID          = pb.ClaimStatus_CLAIM_PAID
)

// Re-export enum values for PermissionType
const (
	PermissionType_PERMISSION_NONE        = pb.PermissionType_PERMISSION_NONE
	PermissionType_PERMISSION_WHITELISTED = pb.PermissionType_PERMISSION_WHITELISTED
	PermissionType_PERMISSION_BLACKLISTED = pb.PermissionType_PERMISSION_BLACKLISTED
)

// Re-export enum values for RotationStatus
const (
	RotationStatus_ROTATION_PENDING  = pb.RotationStatus_ROTATION_PENDING
	RotationStatus_ROTATION_APPROVED = pb.RotationStatus_ROTATION_APPROVED
	RotationStatus_ROTATION_ACTIVE   = pb.RotationStatus_ROTATION_ACTIVE
	RotationStatus_ROTATION_EXPIRED  = pb.RotationStatus_ROTATION_EXPIRED
)

// Re-export enum values for TimeLockStatus
const (
	TimeLockStatus_TIMELOCK_LOCKED     = pb.TimeLockStatus_TIMELOCK_LOCKED
	TimeLockStatus_TIMELOCK_UNLOCKED   = pb.TimeLockStatus_TIMELOCK_UNLOCKED
	TimeLockStatus_TIMELOCK_CHALLENGED = pb.TimeLockStatus_TIMELOCK_CHALLENGED
	TimeLockStatus_TIMELOCK_CANCELLED  = pb.TimeLockStatus_TIMELOCK_CANCELLED
)

// CachedBridgeStats stores pre-computed bridge statistics for O(1) query performance.
// Updated incrementally when transfers are created/modified instead of scanning all records.
type CachedBridgeStats struct {
	// TotalTransfers is the total count of all transfers
	TotalTransfers uint64 `json:"total_transfers"`
	// TransfersByStatus maps status string to count
	TransfersByStatus map[string]uint64 `json:"transfers_by_status"`
	// VolumeByChain maps chain ID to total volume (as string for big int)
	VolumeByChain map[string]string `json:"volume_by_chain"`
	// TotalWrappedTokens is count of wrapped token types
	TotalWrappedTokens uint64 `json:"total_wrapped_tokens"`
	// ActiveValidators is count of active bridge validators
	ActiveValidators uint64 `json:"active_validators"`
	// ActiveRelayers is count of active relayers
	ActiveRelayers uint64 `json:"active_relayers"`
	// LastUpdatedHeight tracks when stats were last computed
	LastUpdatedHeight int64 `json:"last_updated_height"`
}
