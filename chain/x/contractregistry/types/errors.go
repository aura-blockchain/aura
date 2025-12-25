// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Module error codes
const (
	CodeUnauthorized = iota + 2
	CodeContractNotFound
	CodeNotContractAdmin
	CodeContractNotBlacklisted
	CodeInvalidParams
	CodeInvalidMigration
	CodeCircularMigration
	CodeContractNotActive
	CodeContractBlacklisted
	CodeKYCRequired
	CodeSanctioned
	CodeVCRequired
	CodeLowConfidenceScore
	CodeGasLimitExceeded
	CodeRateLimitExceeded
	CodeInvalidContractAddress
	CodeContractAlreadyExists
	CodeInvalidCodeID
	CodeInvalidRequest
	CodeBlacklisted
	CodeNotWhitelisted
	CodeInvalidMetadata
	CodeInvalidSecurityPolicy
	CodeContractPaused
	CodeContractFrozen
	CodeMissingVC
	CodeInsufficientCS
	CodeSanctionsCheckFailed
	CodeContractDeprecated
	CodeNotContractCreator
	CodeInvalidSigner
	CodeInvalidCompliance
	CodeAuditRequired
	CodeVerificationFailed
	CodeInvalidAuditEvent
	CodeRetentionPolicyNotFound
	CodeTooManyContracts
)

var (
	// Authorization errors
	ErrUnauthorized      = errorsmod.Register(ModuleName, CodeUnauthorized, "unauthorized")
	ErrNotContractAdmin  = errorsmod.Register(ModuleName, CodeNotContractAdmin, "not contract admin")
	ErrNotContractCreator = errorsmod.Register(ModuleName, CodeNotContractCreator, "not contract creator")
	ErrInvalidSigner     = errorsmod.Register(ModuleName, CodeInvalidSigner, "invalid signer")

	// Contract state errors
	ErrContractNotFound       = errorsmod.Register(ModuleName, CodeContractNotFound, "contract not found")
	ErrContractNotBlacklisted = errorsmod.Register(ModuleName, CodeContractNotBlacklisted, "contract not blacklisted")
	ErrContractNotActive      = errorsmod.Register(ModuleName, CodeContractNotActive, "contract not active")
	ErrContractBlacklisted    = errorsmod.Register(ModuleName, CodeContractBlacklisted, "contract is blacklisted")
	ErrContractAlreadyExists  = errorsmod.Register(ModuleName, CodeContractAlreadyExists, "contract already exists")
	ErrContractPaused         = errorsmod.Register(ModuleName, CodeContractPaused, "contract is paused")
	ErrContractFrozen         = errorsmod.Register(ModuleName, CodeContractFrozen, "contract is frozen")
	ErrContractDeprecated     = errorsmod.Register(ModuleName, CodeContractDeprecated, "contract is deprecated")

	// Validation errors
	ErrInvalidParams          = errorsmod.Register(ModuleName, CodeInvalidParams, "invalid parameters")
	ErrInvalidMigration       = errorsmod.Register(ModuleName, CodeInvalidMigration, "invalid migration")
	ErrCircularMigration      = errorsmod.Register(ModuleName, CodeCircularMigration, "circular migration detected")
	ErrInvalidContractAddress = errorsmod.Register(ModuleName, CodeInvalidContractAddress, "invalid contract address")
	ErrInvalidCodeID          = errorsmod.Register(ModuleName, CodeInvalidCodeID, "invalid code ID")
	ErrInvalidRequest         = errorsmod.Register(ModuleName, CodeInvalidRequest, "invalid request")
	ErrInvalidMetadata        = errorsmod.Register(ModuleName, CodeInvalidMetadata, "invalid metadata")
	ErrInvalidSecurityPolicy  = errorsmod.Register(ModuleName, CodeInvalidSecurityPolicy, "invalid security policy")
	ErrInvalidCompliance      = errorsmod.Register(ModuleName, CodeInvalidCompliance, "invalid compliance configuration")

	// Compliance errors
	ErrKYCRequired          = errorsmod.Register(ModuleName, CodeKYCRequired, "KYC required")
	ErrSanctioned           = errorsmod.Register(ModuleName, CodeSanctioned, "address is sanctioned")
	ErrVCRequired           = errorsmod.Register(ModuleName, CodeVCRequired, "verifiable credential required")
	ErrLowConfidenceScore   = errorsmod.Register(ModuleName, CodeLowConfidenceScore, "confidence score too low")
	ErrMissingVC            = errorsmod.Register(ModuleName, CodeMissingVC, "missing verifiable credential")
	ErrInsufficientCS       = errorsmod.Register(ModuleName, CodeInsufficientCS, "insufficient confidence score")
	ErrSanctionsCheckFailed = errorsmod.Register(ModuleName, CodeSanctionsCheckFailed, "sanctions check failed")

	// Rate limiting errors
	ErrGasLimitExceeded  = errorsmod.Register(ModuleName, CodeGasLimitExceeded, "gas limit exceeded")
	ErrRateLimitExceeded = errorsmod.Register(ModuleName, CodeRateLimitExceeded, "rate limit exceeded")

	// Access control errors
	ErrBlacklisted    = errorsmod.Register(ModuleName, CodeBlacklisted, "address is blacklisted")
	ErrNotWhitelisted = errorsmod.Register(ModuleName, CodeNotWhitelisted, "address is not whitelisted")

	// Audit and verification errors
	ErrAuditRequired           = errorsmod.Register(ModuleName, CodeAuditRequired, "audit required")
	ErrVerificationFailed      = errorsmod.Register(ModuleName, CodeVerificationFailed, "verification failed")
	ErrInvalidAuditEvent       = errorsmod.Register(ModuleName, CodeInvalidAuditEvent, "invalid audit event")
	ErrRetentionPolicyNotFound = errorsmod.Register(ModuleName, CodeRetentionPolicyNotFound, "retention policy not found")

	// Resource limit errors
	ErrTooManyContracts = errorsmod.Register(ModuleName, CodeTooManyContracts, "too many contracts per creator")
)
