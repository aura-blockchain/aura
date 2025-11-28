package types

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrInvalidParam", ErrInvalidParam, "invalid parameter"},
		{"ErrInvalidRequest", ErrInvalidRequest, "invalid request"},
		{"ErrPoolAlreadyExists", ErrPoolAlreadyExists, "pool already exists"},
		{"ErrPoolNotFound", ErrPoolNotFound, "pool not found"},
		{"ErrInsufficientLiquidity", ErrInsufficientLiquidity, "insufficient liquidity"},
		{"ErrInsufficientPoolLiquidity", ErrInsufficientPoolLiquidity, "insufficient pool liquidity"},
		{"ErrInsufficientLPTokens", ErrInsufficientLPTokens, "insufficient LP tokens"},
		{"ErrNotLiquidityProvider", ErrNotLiquidityProvider, "not a liquidity provider"},
		{"ErrSlippageExceeded", ErrSlippageExceeded, "slippage exceeded"},
		{"ErrSlippageTooHigh", ErrSlippageTooHigh, "slippage too high"},
		{"ErrPriceImpactTooHigh", ErrPriceImpactTooHigh, "price impact too high"},
		{"ErrCircuitBreakerActive", ErrCircuitBreakerActive, "circuit breaker is active"},
		{"ErrFrontRunningDetected", ErrFrontRunningDetected, "front-running detected"},
		{"ErrDustAttack", ErrDustAttack, "dust attack detected"},
		{"ErrInsufficientBalance", ErrInsufficientBalance, "insufficient balance"},
		{"ErrOrderNotFound", ErrOrderNotFound, "order not found"},
		{"ErrOrderAlreadyCanceled", ErrOrderAlreadyCanceled, "order already canceled"},
		{"ErrOrderAlreadyExecuted", ErrOrderAlreadyExecuted, "order already executed"},
		{"ErrHTLCNotFound", ErrHTLCNotFound, "HTLC not found"},
		{"ErrHTLCAlreadyClaimed", ErrHTLCAlreadyClaimed, "HTLC already claimed"},
		{"ErrHTLCExpired", ErrHTLCExpired, "HTLC expired"},
		{"ErrInvalidSecret", ErrInvalidSecret, "invalid secret"},
		{"ErrFlashLoanDetected", ErrFlashLoanDetected, "flash loan attack detected"},
		{"ErrMEVDetected", ErrMEVDetected, "MEV attack detected"},
		{"ErrTradeTooLarge", ErrTradeTooLarge, "trade size exceeds maximum"},
		{"ErrLiquidityLocked", ErrLiquidityLocked, "liquidity is locked"},
		{"ErrOrderManipulation", ErrOrderManipulation, "order manipulation detected"},
		{"ErrWashTradingDetected", ErrWashTradingDetected, "wash trading detected"},
		{"ErrPoolCreationCooldown", ErrPoolCreationCooldown, "pool creation cooldown active"},
		{"ErrMaxPoolsExceeded", ErrMaxPoolsExceeded, "maximum pools per creator exceeded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrorsAreErrors(t *testing.T) {
	// Test that all defined error variables implement the error interface
	var _ error = ErrInvalidParam
	var _ error = ErrInvalidRequest
	var _ error = ErrPoolAlreadyExists
	var _ error = ErrPoolNotFound
	var _ error = ErrInsufficientLiquidity
	var _ error = ErrInsufficientPoolLiquidity
	var _ error = ErrInsufficientLPTokens
	var _ error = ErrNotLiquidityProvider
	var _ error = ErrSlippageExceeded
	var _ error = ErrSlippageTooHigh
	var _ error = ErrPriceImpactTooHigh
	var _ error = ErrCircuitBreakerActive
	var _ error = ErrFrontRunningDetected
	var _ error = ErrDustAttack
	var _ error = ErrInsufficientBalance
	var _ error = ErrOrderNotFound
	var _ error = ErrOrderAlreadyCanceled
	var _ error = ErrOrderAlreadyExecuted
	var _ error = ErrHTLCNotFound
	var _ error = ErrHTLCAlreadyClaimed
	var _ error = ErrHTLCExpired
	var _ error = ErrInvalidSecret
	var _ error = ErrFlashLoanDetected
	var _ error = ErrMEVDetected
	var _ error = ErrTradeTooLarge
	var _ error = ErrLiquidityLocked
	var _ error = ErrOrderManipulation
	var _ error = ErrWashTradingDetected
	var _ error = ErrPoolCreationCooldown
	var _ error = ErrMaxPoolsExceeded
}

func TestErrorComparison(t *testing.T) {
	// Test that errors can be compared using errors.Is
	err := ErrPoolNotFound
	if !errors.Is(err, ErrPoolNotFound) {
		t.Error("errors.Is should return true for same error")
	}

	if errors.Is(err, ErrOrderNotFound) {
		t.Error("errors.Is should return false for different error")
	}
}
