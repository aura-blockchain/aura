package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DEX module error codes
var (
	// General errors (1-19)
	ErrInvalidParam   = errorsmod.Register(ModuleName, 1, "invalid parameter")
	ErrInvalidRequest = errorsmod.Register(ModuleName, 2, "invalid request")

	// Pool errors (20-39)
	ErrPoolAlreadyExists         = errorsmod.Register(ModuleName, 20, "pool already exists")
	ErrPoolNotFound              = errorsmod.Register(ModuleName, 21, "pool not found")
	ErrInsufficientLiquidity     = errorsmod.Register(ModuleName, 22, "insufficient liquidity")
	ErrInsufficientPoolLiquidity = errorsmod.Register(ModuleName, 23, "insufficient pool liquidity")
	ErrInsufficientLPTokens      = errorsmod.Register(ModuleName, 24, "insufficient LP tokens")
	ErrNotLiquidityProvider      = errorsmod.Register(ModuleName, 25, "not a liquidity provider")
	ErrLiquidityLocked           = errorsmod.Register(ModuleName, 26, "liquidity is locked")
	ErrPoolCreationCooldown      = errorsmod.Register(ModuleName, 27, "pool creation cooldown active")
	ErrMaxPoolsExceeded          = errorsmod.Register(ModuleName, 28, "maximum pools per creator exceeded")
	ErrPoolCreationLimitExceeded = errorsmod.Register(ModuleName, 29, "pool creation limit exceeded")

	// Slippage and price errors (40-49)
	ErrSlippageExceeded   = errorsmod.Register(ModuleName, 40, "slippage exceeded")
	ErrSlippageTooHigh    = errorsmod.Register(ModuleName, 41, "slippage too high")
	ErrPriceImpactTooHigh = errorsmod.Register(ModuleName, 42, "price impact too high")
	ErrPriceManipulation  = errorsmod.Register(ModuleName, 43, "price manipulation detected")

	// Order errors (50-69)
	ErrOrderNotFound        = errorsmod.Register(ModuleName, 50, "order not found")
	ErrOrderAlreadyCanceled = errorsmod.Register(ModuleName, 51, "order already canceled")
	ErrOrderAlreadyExecuted = errorsmod.Register(ModuleName, 52, "order already executed")
	ErrOrderManipulation    = errorsmod.Register(ModuleName, 53, "order manipulation detected")
	ErrTradeTooLarge        = errorsmod.Register(ModuleName, 54, "trade size exceeds maximum")

	// Commitment/reveal errors (70-79)
	ErrCommitmentNotFound      = errorsmod.Register(ModuleName, 70, "order commitment not found")
	ErrRevealDeadlineExpired   = errorsmod.Register(ModuleName, 71, "reveal deadline has expired")
	ErrHashMismatch            = errorsmod.Register(ModuleName, 72, "commitment hash does not match revealed order")
	ErrCommitmentAlreadyExists = errorsmod.Register(ModuleName, 73, "commitment already exists for this sender")

	// HTLC errors (80-89)
	ErrHTLCNotFound       = errorsmod.Register(ModuleName, 80, "HTLC not found")
	ErrHTLCAlreadyClaimed = errorsmod.Register(ModuleName, 81, "HTLC already claimed")
	ErrHTLCExpired        = errorsmod.Register(ModuleName, 82, "HTLC expired")
	ErrInvalidSecret      = errorsmod.Register(ModuleName, 83, "invalid secret")

	// Security and attack prevention errors (90-109)
	ErrCircuitBreakerActive = errorsmod.Register(ModuleName, 90, "circuit breaker is active")
	ErrFrontRunningDetected = errorsmod.Register(ModuleName, 91, "front-running detected")
	ErrDustAttack           = errorsmod.Register(ModuleName, 92, "dust attack detected")
	ErrFlashLoanDetected    = errorsmod.Register(ModuleName, 93, "flash loan attack detected")
	ErrMEVDetected          = errorsmod.Register(ModuleName, 94, "MEV attack detected")
	ErrWashTradingDetected  = errorsmod.Register(ModuleName, 95, "wash trading detected")

	// Balance errors (110-119)
	ErrInsufficientBalance = errorsmod.Register(ModuleName, 110, "insufficient balance")

	// Serialization errors (120-129)
	ErrMarshalFailed   = errorsmod.Register(ModuleName, 120, "failed to marshal data")
	ErrUnmarshalFailed = errorsmod.Register(ModuleName, 121, "failed to unmarshal data")
)
