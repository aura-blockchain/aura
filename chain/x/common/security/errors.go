package security

import (
	errorsmod "cosmossdk.io/errors"
)

// Common security module name for error registration
const ModuleName = "security-common"

// Common security error codes
var (
	// Reentrancy errors (1-9)
	ErrReentrancyDetected = errorsmod.Register(ModuleName, 1, "reentrancy attack detected")

	// Pause errors (10-19)
	ErrModulePaused  = errorsmod.Register(ModuleName, 10, "module is currently paused")
	ErrAlreadyPaused = errorsmod.Register(ModuleName, 11, "module is already paused")
	ErrNotPaused     = errorsmod.Register(ModuleName, 12, "module is not paused")

	// Access control errors (20-29)
	ErrUnauthorized = errorsmod.Register(ModuleName, 20, "unauthorized: caller does not have required permissions")

	// Input validation errors (30-49)
	ErrInvalidAddress = errorsmod.Register(ModuleName, 30, "invalid address")
	ErrInvalidAmount  = errorsmod.Register(ModuleName, 31, "invalid amount")
	ErrNegativeAmount = errorsmod.Register(ModuleName, 32, "amount cannot be negative")
	ErrZeroAmount     = errorsmod.Register(ModuleName, 33, "amount cannot be zero")
	ErrInvalidInput   = errorsmod.Register(ModuleName, 34, "invalid input")
	ErrEmptyField     = errorsmod.Register(ModuleName, 35, "required field is empty")
	ErrFieldTooLong   = errorsmod.Register(ModuleName, 36, "field exceeds maximum length")
	ErrFieldTooShort  = errorsmod.Register(ModuleName, 37, "field is below minimum length")

	// Gas errors (50-59)
	ErrGasLimitExceeded = errorsmod.Register(ModuleName, 50, "gas limit exceeded")
	ErrZeroGasLimit     = errorsmod.Register(ModuleName, 51, "gas limit cannot be zero")
	ErrInsufficientGas  = errorsmod.Register(ModuleName, 52, "insufficient gas remaining")

	// Overflow errors (60-69)
	ErrIntegerOverflow  = errorsmod.Register(ModuleName, 60, "integer overflow detected")
	ErrIntegerUnderflow = errorsmod.Register(ModuleName, 61, "integer underflow detected")

	// External call errors (70-79)
	ErrExternalCallFailed = errorsmod.Register(ModuleName, 70, "external call failed")
	ErrInvalidCallResult  = errorsmod.Register(ModuleName, 71, "invalid external call result")
)
