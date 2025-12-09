package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Bridge module error codes
var (
	ErrInvalidParam          = errorsmod.Register(ModuleName, 1, "invalid parameter")
	ErrDuplicateAttestation  = errorsmod.Register(ModuleName, 2, "duplicate attestation")
	ErrWithdrawalNotFound    = errorsmod.Register(ModuleName, 3, "withdrawal not found")
	ErrChainNotFound         = errorsmod.Register(ModuleName, 4, "chain not found")
	ErrTransferNotFound      = errorsmod.Register(ModuleName, 5, "transfer not found")
	ErrCircuitBreakerTripped = errorsmod.Register(ModuleName, 6, "circuit breaker tripped - amount exceeds limit")
	ErrTimelockNotElapsed    = errorsmod.Register(ModuleName, 7, "timelock period has not elapsed")
	ErrChainDisabled         = errorsmod.Register(ModuleName, 8, "chain disabled")
	ErrInvalidSignature      = errorsmod.Register(ModuleName, 9, "invalid cryptographic signature")
	ErrCorruptedData         = errorsmod.Register(ModuleName, 10, "corrupted or invalid data in storage")
)
