// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Bridge module error codes (300-399 range)
var (
	ErrInvalidParam          = errorsmod.Register(ModuleName, 300, "invalid parameter")
	ErrDuplicateAttestation  = errorsmod.Register(ModuleName, 301, "duplicate attestation")
	ErrWithdrawalNotFound    = errorsmod.Register(ModuleName, 302, "withdrawal not found")
	ErrChainNotFound         = errorsmod.Register(ModuleName, 303, "chain not found")
	ErrTransferNotFound      = errorsmod.Register(ModuleName, 304, "transfer not found")
	ErrCircuitBreakerTripped = errorsmod.Register(ModuleName, 305, "circuit breaker tripped - amount exceeds limit")
	ErrTimelockNotElapsed    = errorsmod.Register(ModuleName, 306, "timelock period has not elapsed")
	ErrChainDisabled         = errorsmod.Register(ModuleName, 307, "chain disabled")
	ErrInvalidSignature      = errorsmod.Register(ModuleName, 308, "invalid cryptographic signature")
	ErrCorruptedData         = errorsmod.Register(ModuleName, 309, "corrupted or invalid data in storage")
	ErrSignatureReplay       = errorsmod.Register(ModuleName, 310, "signature already used - replay attack prevented")
	ErrSignatureRateLimit    = errorsmod.Register(ModuleName, 311, "signature verification rate limit exceeded")
	ErrInvalidRecoveryID     = errorsmod.Register(ModuleName, 312, "invalid ECDSA recovery ID")
	ErrMarshalFailed         = errorsmod.Register(ModuleName, 313, "failed to marshal data")
	ErrUnmarshalFailed       = errorsmod.Register(ModuleName, 314, "failed to unmarshal data")
	ErrIBCNotEnabled         = errorsmod.Register(ModuleName, 399, "IBC not enabled for bridge module - IBC-based bridging will be available in v2.0")
)
