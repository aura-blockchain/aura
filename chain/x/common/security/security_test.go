package security

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestReentrancyGuard(t *testing.T) {
	guard := NewReentrancyGuard()

	// Test normal operation
	err := guard.WithReentrancyGuard(func() error {
		return nil
	})
	require.NoError(t, err)

	// Test reentrancy detection
	err = guard.WithReentrancyGuard(func() error {
		// Attempt nested call (should fail)
		return guard.WithReentrancyGuard(func() error {
			return nil
		})
	})
	require.Error(t, err)
	require.Equal(t, ErrReentrancyDetected, err)
}

func TestPauseGuard(t *testing.T) {
	admin := "admin123"
	guard := NewPauseGuard(admin)

	// Initial state should be unpaused
	require.False(t, guard.IsPaused())

	// Create mock context
	ctx := sdk.Context{}

	// Test pause by admin
	err := guard.Pause(ctx, admin)
	require.NoError(t, err)
	require.True(t, guard.IsPaused())

	// Test pause when already paused
	err = guard.Pause(ctx, admin)
	require.Error(t, err)
	require.Equal(t, ErrAlreadyPaused, err)

	// Test unpause by admin
	err := guard.Unpause(ctx, admin)
	require.NoError(t, err)
	require.False(t, guard.IsPaused())

	// Test unpause when not paused
	err = guard.Unpause(ctx, admin)
	require.Error(t, err)
	require.Equal(t, ErrNotPaused, err)

	// Test pause by non-admin
	nonAdmin := "user123"
	err = guard.Pause(ctx, nonAdmin)
	require.Error(t, err)
	require.Equal(t, ErrUnauthorized, err)
}

func TestInputValidator(t *testing.T) {
	validator := NewInputValidator()

	// Test address validation
	err := validator.ValidateAddress("aura1validaddress123456")
	require.NoError(t, err)

	err = validator.ValidateAddress("")
	require.Error(t, err)
	require.Equal(t, ErrInvalidAddress, err)

	err = validator.ValidateAddress("short")
	require.Error(t, err)

	// Test amount validation
	validAmount := sdk.NewInt(100)
	err = validator.ValidateAmount(validAmount)
	require.NoError(t, err)

	zeroAmount := sdk.ZeroInt()
	err = validator.ValidateAmount(zeroAmount)
	require.Error(t, err)
	require.Equal(t, ErrZeroAmount, err)

	// Test string validation
	err = validator.ValidateString("valid", "test")
	require.NoError(t, err)

	err = validator.ValidateString("", "test")
	require.Error(t, err)
}

func TestSafeMath(t *testing.T) {
	sm := NewSafeMath()

	// Test safe addition
	a := sdk.NewInt(100)
	b := sdk.NewInt(50)
	result, err := sm.SafeAdd(a, b)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(150), result)

	// Test safe subtraction
	result, err = sm.SafeSub(a, b)
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(50), result)

	// Test safe subtraction underflow
	_, err = sm.SafeSub(b, a)
	require.Error(t, err)
	require.Equal(t, ErrIntegerUnderflow, err)

	// Test safe multiplication
	result, err = sm.SafeMul(sdk.NewInt(10), sdk.NewInt(20))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(200), result)

	// Test safe division
	result, err = sm.SafeDiv(sdk.NewInt(100), sdk.NewInt(5))
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt(20), result)

	// Test division by zero
	_, err = sm.SafeDiv(sdk.NewInt(100), sdk.ZeroInt())
	require.Error(t, err)
	require.Equal(t, ErrZeroAmount, err)
}

func TestGasLimitGuard(t *testing.T) {
	guard := NewGasLimitGuard(1000000)

	// Test valid gas limit
	err := guard.ValidateGasLimit(500000)
	require.NoError(t, err)

	// Test excessive gas limit
	err = guard.ValidateGasLimit(2000000)
	require.Error(t, err)
	require.Equal(t, ErrGasLimitExceeded, err)

	// Test zero gas limit
	err = guard.ValidateGasLimit(0)
	require.Error(t, err)
	require.Equal(t, ErrZeroGasLimit, err)
}

func TestAtomicityGuard(t *testing.T) {
	guard := NewAtomicityGuard()

	rollbackExecuted := false
	guard.AddRollback(func() error {
		rollbackExecuted = true
		return nil
	})

	// Test commit
	guard.Commit()
	require.False(t, rollbackExecuted)

	// Add new rollback
	guard.AddRollback(func() error {
		rollbackExecuted = true
		return nil
	})

	// Test rollback
	err := guard.Rollback()
	require.NoError(t, err)
	require.True(t, rollbackExecuted)
}

func TestAccessControl(t *testing.T) {
	admin1 := "admin1"
	admin2 := "admin2"
	user := "user1"

	ac := NewAccessControl([]string{admin1})

	// Test initial admin
	require.True(t, ac.IsAdmin(admin1))
	require.False(t, ac.IsAdmin(admin2))
	require.False(t, ac.IsAdmin(user))

	// Test add admin by admin
	err := ac.AddAdmin(admin2, admin1)
	require.NoError(t, err)
	require.True(t, ac.IsAdmin(admin2))

	// Test add admin by non-admin
	err = ac.AddAdmin("newadmin", user)
	require.Error(t, err)
	require.Equal(t, ErrUnauthorized, err)

	// Test role management
	err = ac.GrantRole(user, "trader", admin1)
	require.NoError(t, err)
	require.True(t, ac.HasRole(user, "trader"))

	// Test role grant by non-admin
	err = ac.GrantRole(user, "miner", user)
	require.Error(t, err)
	require.Equal(t, ErrUnauthorized, err)

	// Test revoke role
	err = ac.RevokeRole(user, "trader", admin1)
	require.NoError(t, err)
	require.False(t, ac.HasRole(user, "trader"))
}

func TestDecimalSafeMath(t *testing.T) {
	sm := NewSafeMath()

	// Test decimal addition
	a := sdk.NewDec(100)
	b := sdk.NewDec(50)
	result, err := sm.SafeAddDec(a, b)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(150), result)

	// Test decimal subtraction
	result, err = sm.SafeSubDec(a, b)
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(50), result)

	// Test decimal multiplication
	result, err = sm.SafeMulDec(sdk.NewDec(10), sdk.NewDec(5))
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(50), result)

	// Test decimal division
	result, err = sm.SafeDivDec(sdk.NewDec(100), sdk.NewDec(5))
	require.NoError(t, err)
	require.Equal(t, sdk.NewDec(20), result)

	// Test decimal division by zero
	_, err = sm.SafeDivDec(sdk.NewDec(100), sdk.ZeroDec())
	require.Error(t, err)
	require.Equal(t, ErrZeroAmount, err)
}
