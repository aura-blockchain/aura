package types

import (
	"cosmossdk.io/errors"
)

// WASM module sentinel errors
var (
	ErrUnauthorized = errors.Register(ModuleName, 100, "unauthorized")

	ErrContractPaused = errors.Register(ModuleName, 101, "contract paused")

	ErrContractTooLarge = errors.Register(ModuleName, 102, "contract code exceeds maximum size")

	ErrGasLimitExceeded = errors.Register(ModuleName, 103, "gas limit exceeded")

	ErrInvalidContractCode = errors.Register(ModuleName, 104, "invalid contract code")

	ErrInvalidContractAddress = errors.Register(ModuleName, 105, "invalid contract address")

	ErrMigrationNotAllowed = errors.Register(ModuleName, 106, "migration not allowed for this contract")

	ErrSecurityViolation = errors.Register(ModuleName, 107, "security policy violation")

	ErrReentrancyDetected = errors.Register(ModuleName, 108, "reentrancy detected")

	ErrInvalidAdmin = errors.Register(ModuleName, 109, "invalid admin address")
)
