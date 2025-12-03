package types

import "errors"

var (
	ErrInvalidParam          = errors.New("invalid parameter")
	ErrDuplicateAttestation  = errors.New("duplicate attestation")
	ErrWithdrawalNotFound    = errors.New("withdrawal not found")
	ErrChainNotFound         = errors.New("chain not found")
	ErrTransferNotFound      = errors.New("transfer not found")
	ErrCircuitBreakerTripped = errors.New("circuit breaker tripped - amount exceeds limit")
	ErrTimelockNotElapsed    = errors.New("timelock period has not elapsed")
	ErrChainDisabled         = errors.New("chain disabled")
	ErrInvalidSignature      = errors.New("invalid cryptographic signature")
	ErrCorruptedData         = errors.New("corrupted or invalid data in storage")
)
