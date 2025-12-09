package types

import (
	errorsmod "cosmossdk.io/errors"
)

// DEX module error codes (400-499 range)
var (
	// General errors (400-409)
	ErrInvalidParam   = errorsmod.Register(ModuleName, 400, "invalid parameter")
	ErrInvalidRequest = errorsmod.Register(ModuleName, 401, "invalid request")

	// Pool errors (410-429)
	ErrPoolAlreadyExists         = errorsmod.Register(ModuleName, 410, "pool already exists")
	ErrPoolNotFound              = errorsmod.Register(ModuleName, 411, "pool not found")
	ErrInsufficientLiquidity     = errorsmod.Register(ModuleName, 412, "insufficient liquidity")
	ErrInsufficientPoolLiquidity = errorsmod.Register(ModuleName, 413, "insufficient pool liquidity")
	ErrInsufficientLPTokens      = errorsmod.Register(ModuleName, 414, "insufficient LP tokens")
	ErrNotLiquidityProvider      = errorsmod.Register(ModuleName, 415, "not a liquidity provider")
	ErrLiquidityLocked           = errorsmod.Register(ModuleName, 416, "liquidity is locked")
	ErrPoolCreationCooldown      = errorsmod.Register(ModuleName, 417, "pool creation cooldown active")
	ErrMaxPoolsExceeded          = errorsmod.Register(ModuleName, 418, "maximum pools per creator exceeded")
	ErrPoolCreationLimitExceeded = errorsmod.Register(ModuleName, 419, "pool creation limit exceeded")

	// Slippage and price errors (430-439)
	ErrSlippageExceeded   = errorsmod.Register(ModuleName, 430, "slippage exceeded")
	ErrSlippageTooHigh    = errorsmod.Register(ModuleName, 431, "slippage too high")
	ErrPriceImpactTooHigh = errorsmod.Register(ModuleName, 432, "price impact too high")
	ErrPriceManipulation  = errorsmod.Register(ModuleName, 433, "price manipulation detected")

	// Order errors (440-459)
	ErrOrderNotFound        = errorsmod.Register(ModuleName, 440, "order not found")
	ErrOrderAlreadyCanceled = errorsmod.Register(ModuleName, 441, "order already canceled")
	ErrOrderAlreadyExecuted = errorsmod.Register(ModuleName, 442, "order already executed")
	ErrOrderManipulation    = errorsmod.Register(ModuleName, 443, "order manipulation detected")
	ErrTradeTooLarge        = errorsmod.Register(ModuleName, 444, "trade size exceeds maximum")

	// Commitment/reveal errors (460-469)
	ErrCommitmentNotFound      = errorsmod.Register(ModuleName, 460, "order commitment not found")
	ErrRevealDeadlineExpired   = errorsmod.Register(ModuleName, 461, "reveal deadline has expired")
	ErrHashMismatch            = errorsmod.Register(ModuleName, 462, "commitment hash does not match revealed order")
	ErrCommitmentAlreadyExists = errorsmod.Register(ModuleName, 463, "commitment already exists for this sender")

	// HTLC errors (470-479)
	ErrHTLCNotFound       = errorsmod.Register(ModuleName, 470, "HTLC not found")
	ErrHTLCAlreadyClaimed = errorsmod.Register(ModuleName, 471, "HTLC already claimed")
	ErrHTLCExpired        = errorsmod.Register(ModuleName, 472, "HTLC expired")
	ErrInvalidSecret      = errorsmod.Register(ModuleName, 473, "invalid secret")

	// Security and attack prevention errors (480-489)
	ErrCircuitBreakerActive = errorsmod.Register(ModuleName, 480, "circuit breaker is active")
	ErrFrontRunningDetected = errorsmod.Register(ModuleName, 481, "front-running detected")
	ErrDustAttack           = errorsmod.Register(ModuleName, 482, "dust attack detected")
	ErrFlashLoanDetected    = errorsmod.Register(ModuleName, 483, "flash loan attack detected")
	ErrMEVDetected          = errorsmod.Register(ModuleName, 484, "MEV attack detected")
	ErrWashTradingDetected  = errorsmod.Register(ModuleName, 485, "wash trading detected")

	// Balance errors (490-494)
	ErrInsufficientBalance = errorsmod.Register(ModuleName, 490, "insufficient balance")

	// Serialization errors (495-499)
	ErrMarshalFailed   = errorsmod.Register(ModuleName, 495, "failed to marshal data")
	ErrUnmarshalFailed = errorsmod.Register(ModuleName, 496, "failed to unmarshal data")
)
