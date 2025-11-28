package security

import (
	"cosmossdk.io/math"
)

// SafeMath provides safe arithmetic operations with overflow protection
type SafeMath struct{}

// NewSafeMath creates a new SafeMath instance
func NewSafeMath() *SafeMath {
	return &SafeMath{}
}

// SafeAdd performs safe addition with overflow check
func (sm *SafeMath) SafeAdd(a, b math.Int) (math.Int, error) {
	if a.IsNil() || b.IsNil() {
		return math.Int{}, ErrInvalidAmount
	}

	result := a.Add(b)

	// Check for overflow: if b is positive, result should be > a
	if b.IsPositive() && !result.GT(a) {
		return math.Int{}, ErrIntegerOverflow
	}

	// Check for underflow: if b is negative, result should be < a
	if b.IsNegative() && !result.LT(a) {
		return math.Int{}, ErrIntegerUnderflow
	}

	return result, nil
}

// SafeSub performs safe subtraction with underflow check
func (sm *SafeMath) SafeSub(a, b math.Int) (math.Int, error) {
	if a.IsNil() || b.IsNil() {
		return math.Int{}, ErrInvalidAmount
	}

	if a.LT(b) {
		return math.Int{}, ErrIntegerUnderflow
	}

	result := a.Sub(b)
	return result, nil
}

// SafeMul performs safe multiplication with overflow check
func (sm *SafeMath) SafeMul(a, b math.Int) (math.Int, error) {
	if a.IsNil() || b.IsNil() {
		return math.Int{}, ErrInvalidAmount
	}

	// Zero multiplication is always safe
	if a.IsZero() || b.IsZero() {
		return math.ZeroInt(), nil
	}

	result := a.Mul(b)

	// Check for overflow by dividing back
	if !result.Quo(b).Equal(a) {
		return math.Int{}, ErrIntegerOverflow
	}

	return result, nil
}

// SafeDiv performs safe division with divide-by-zero check
func (sm *SafeMath) SafeDiv(a, b math.Int) (math.Int, error) {
	if a.IsNil() || b.IsNil() {
		return math.Int{}, ErrInvalidAmount
	}

	if b.IsZero() {
		return math.Int{}, ErrZeroAmount
	}

	result := a.Quo(b)
	return result, nil
}

// SafeAddDec performs safe decimal addition
func (sm *SafeMath) SafeAddDec(a, b math.LegacyDec) (math.LegacyDec, error) {
	result := a.Add(b)

	// LegacyDec has built-in overflow protection
	// The operation itself will panic on overflow, so if we get here it's safe
	return result, nil
}

// SafeSubDec performs safe decimal subtraction
func (sm *SafeMath) SafeSubDec(a, b math.LegacyDec) (math.LegacyDec, error) {
	result := a.Sub(b)

	// LegacyDec has built-in overflow protection
	return result, nil
}

// SafeMulDec performs safe decimal multiplication
func (sm *SafeMath) SafeMulDec(a, b math.LegacyDec) (math.LegacyDec, error) {
	result := a.Mul(b)

	// LegacyDec has built-in overflow protection
	return result, nil
}

// SafeDivDec performs safe decimal division
func (sm *SafeMath) SafeDivDec(a, b math.LegacyDec) (math.LegacyDec, error) {
	if b.IsZero() {
		return math.LegacyDec{}, ErrZeroAmount
	}

	result := a.Quo(b)

	// LegacyDec has built-in overflow protection
	return result, nil
}

// CheckNoOverflow validates that an operation won't overflow
func (sm *SafeMath) CheckNoOverflow(a, b math.Int, op string) error {
	if a.IsNil() || b.IsNil() {
		return ErrInvalidAmount
	}

	switch op {
	case "add":
		_, err := sm.SafeAdd(a, b)
		return err
	case "sub":
		_, err := sm.SafeSub(a, b)
		return err
	case "mul":
		_, err := sm.SafeMul(a, b)
		return err
	case "div":
		_, err := sm.SafeDiv(a, b)
		return err
	default:
		return ErrInvalidInput
	}
}
