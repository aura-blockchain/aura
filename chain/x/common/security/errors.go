package security

import (
	"errors"
)

var (
	// Reentrancy errors
	ErrReentrancyDetected = errors.New("reentrancy attack detected")

	// Pause errors
	ErrModulePaused  = errors.New("module is currently paused")
	ErrAlreadyPaused = errors.New("module is already paused")
	ErrNotPaused     = errors.New("module is not paused")

	// Access control errors
	ErrUnauthorized = errors.New("unauthorized: caller does not have required permissions")

	// Input validation errors
	ErrInvalidAddress = errors.New("invalid address")
	ErrInvalidAmount  = errors.New("invalid amount")
	ErrNegativeAmount = errors.New("amount cannot be negative")
	ErrZeroAmount     = errors.New("amount cannot be zero")
	ErrInvalidInput   = errors.New("invalid input")
	ErrEmptyField     = errors.New("required field is empty")
	ErrFieldTooLong   = errors.New("field exceeds maximum length")
	ErrFieldTooShort  = errors.New("field is below minimum length")

	// Gas errors
	ErrGasLimitExceeded = errors.New("gas limit exceeded")
	ErrZeroGasLimit     = errors.New("gas limit cannot be zero")
	ErrInsufficientGas  = errors.New("insufficient gas remaining")

	// Overflow errors
	ErrIntegerOverflow  = errors.New("integer overflow detected")
	ErrIntegerUnderflow = errors.New("integer underflow detected")

	// External call errors
	ErrExternalCallFailed = errors.New("external call failed")
	ErrInvalidCallResult  = errors.New("invalid external call result")
)
