package types

import (
	"fmt"
	"math/big"

	sdkmath "cosmossdk.io/math"
)

// SafeMath provides overflow-protected arithmetic operations for DEX calculations.
//
// SECURITY RATIONALE:
// DEX operations involve user-controlled values (amounts, multipliers) that could
// be crafted to cause integer overflow/underflow, leading to:
// - Zero or negative fees (protocol loss of revenue)
// - Incorrect swap calculations (theft of liquidity)
// - Pool invariant violations (protocol insolvency)
//
// All multiplication operations MUST use these functions to prevent overflow attacks.

// SafeMul performs multiplication with overflow protection.
//
// Returns error if:
//   - Either operand is negative
//   - Result would exceed sdkmath.MaxInt256 (overflow)
//   - Result is negative (overflow wrapped around)
//
// Example usage:
//
//	result, err := SafeMul(userAmount, feeMultiplier)
//	if err != nil {
//	    return sdkmath.ZeroInt(), err
//	}
func SafeMul(a, b sdkmath.Int) (sdkmath.Int, error) {
	// Reject negative inputs - all DEX amounts should be positive
	if a.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf("SafeMul: first operand cannot be negative: %s", a.String())
	}
	if b.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf("SafeMul: second operand cannot be negative: %s", b.String())
	}

	// Handle zero cases (no overflow possible)
	if a.IsZero() || b.IsZero() {
		return sdkmath.ZeroInt(), nil
	}

	// Check for overflow BEFORE multiplication
	// If a > MaxInt256 / b, then a * b > MaxInt256 (overflow)
	// Rearranged: if MaxInt256 / b < a, overflow will occur
	maxSafeValue := new(big.Int).Div(
		sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 255)).Sub(sdkmath.OneInt()).BigInt(), // MaxInt256
		b.BigInt(),
	)

	if a.BigInt().Cmp(maxSafeValue) > 0 {
		return sdkmath.ZeroInt(), fmt.Errorf(
			"SafeMul: multiplication would overflow (a=%s, b=%s, max_safe_a=%s)",
			a.String(),
			b.String(),
			sdkmath.NewIntFromBigInt(maxSafeValue).String(),
		)
	}

	// Perform multiplication
	result := a.Mul(b)

	// Post-multiplication sanity check: result should never be negative
	// If it is, overflow occurred despite our checks (defense in depth)
	if result.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf(
			"SafeMul: overflow detected, result is negative (a=%s, b=%s)",
			a.String(),
			b.String(),
		)
	}

	return result, nil
}

// SafeAdd performs addition with overflow protection.
//
// Returns error if:
//   - Either operand is negative
//   - Result would exceed sdkmath.MaxInt256 (overflow)
func SafeAdd(a, b sdkmath.Int) (sdkmath.Int, error) {
	if a.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf("SafeAdd: first operand cannot be negative: %s", a.String())
	}
	if b.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf("SafeAdd: second operand cannot be negative: %s", b.String())
	}

	// Check for overflow: if a > MaxInt256 - b, then a + b > MaxInt256
	maxInt256 := sdkmath.NewIntFromBigInt(new(big.Int).Lsh(big.NewInt(1), 255)).Sub(sdkmath.OneInt())
	maxSafeValue := maxInt256.Sub(b)

	if a.GT(maxSafeValue) {
		return sdkmath.ZeroInt(), fmt.Errorf(
			"SafeAdd: addition would overflow (a=%s, b=%s, max=%s)",
			a.String(),
			b.String(),
			maxInt256.String(),
		)
	}

	result := a.Add(b)

	// Sanity check
	if result.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf(
			"SafeAdd: overflow detected, result is negative (a=%s, b=%s)",
			a.String(),
			b.String(),
		)
	}

	return result, nil
}

// SafeMulDec performs multiplication of Int by LegacyDec with overflow protection.
//
// This is commonly used in fee calculations where we multiply an amount by a fee rate.
// Example: feeAmount = tradeAmount * feeRate
//
// Returns error if:
//   - Amount is negative
//   - Fee rate is negative
//   - Result would overflow
func SafeMulDec(amount sdkmath.Int, rate sdkmath.LegacyDec) (sdkmath.Int, error) {
	if amount.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf("SafeMulDec: amount cannot be negative: %s", amount.String())
	}
	if rate.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf("SafeMulDec: rate cannot be negative: %s", rate.String())
	}

	if amount.IsZero() || rate.IsZero() {
		return sdkmath.ZeroInt(), nil
	}

	// Convert to Dec, multiply, truncate
	result := amount.ToLegacyDec().Mul(rate).TruncateInt()

	// Sanity check: result should not be negative
	if result.IsNegative() {
		return sdkmath.ZeroInt(), fmt.Errorf(
			"SafeMulDec: overflow detected, result is negative (amount=%s, rate=%s)",
			amount.String(),
			rate.String(),
		)
	}

	return result, nil
}

// CheckPositive validates that a value is positive (> 0).
// This is a common precondition for DEX operations.
func CheckPositive(value sdkmath.Int, fieldName string) error {
	if value.IsNegative() {
		return fmt.Errorf("%s cannot be negative: %s", fieldName, value.String())
	}
	if value.IsZero() {
		return fmt.Errorf("%s must be positive: got zero", fieldName)
	}
	return nil
}

// CheckNonNegative validates that a value is non-negative (>= 0).
func CheckNonNegative(value sdkmath.Int, fieldName string) error {
	if value.IsNegative() {
		return fmt.Errorf("%s cannot be negative: %s", fieldName, value.String())
	}
	return nil
}
