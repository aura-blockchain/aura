package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/security/keeper"
	"github.com/aequitas/aura/chain/x/security/types"
)

type SecurityGuardsTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper *keeper.Keeper
}

func TestSecurityGuardsTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityGuardsTestSuite))
}

func (suite *SecurityGuardsTestSuite) SetupTest() {
	// Create test keeper
	suite.ctx, suite.keeper = testutil.NewSecurityKeeperForTest(suite.T())
}

// =============================================================================
// Reentrancy Protection Tests
// =============================================================================

func (suite *SecurityGuardsTestSuite) TestReentrancyGuard_BasicProtection() {
	ctx := suite.ctx
	k := suite.keeper

	key := "test:operation:user1"

	// First entry should succeed
	err := k.EnterNoReentrant(ctx, key)
	suite.Require().NoError(err, "first entry should succeed")

	// Second entry with same key should fail (reentrancy detected)
	err = k.EnterNoReentrant(ctx, key)
	suite.Require().Error(err, "second entry should fail")
	suite.Require().ErrorIs(err, types.ErrReentrancyDetected)

	// Exit the guard
	k.ExitNoReentrant(ctx, key)

	// After exit, should be able to enter again
	err = k.EnterNoReentrant(ctx, key)
	suite.Require().NoError(err, "entry after exit should succeed")

	k.ExitNoReentrant(ctx, key)
}

func (suite *SecurityGuardsTestSuite) TestReentrancyGuard_ScopedLocks() {
	ctx := suite.ctx
	k := suite.keeper

	key1 := "pool:123"
	key2 := "pool:456"

	// Enter with key1
	err := k.EnterNoReentrant(ctx, key1)
	suite.Require().NoError(err)

	// Enter with key2 should succeed (different scope)
	err = k.EnterNoReentrant(ctx, key2)
	suite.Require().NoError(err)

	// Reentering key1 should fail
	err = k.EnterNoReentrant(ctx, key1)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrReentrancyDetected)

	// Reentering key2 should fail
	err = k.EnterNoReentrant(ctx, key2)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrReentrancyDetected)

	// Exit both
	k.ExitNoReentrant(ctx, key1)
	k.ExitNoReentrant(ctx, key2)
}

func (suite *SecurityGuardsTestSuite) TestReentrancyGuard_WithGuardHelper() {
	ctx := suite.ctx
	k := suite.keeper

	key := "operation:user1"
	executedCount := 0

	// Execute function with reentrancy guard
	err := k.WithReentrancyGuard(ctx, key, func() error {
		executedCount++

		// Try to reenter (should fail)
		err := k.EnterNoReentrant(ctx, key)
		suite.Require().Error(err, "reentrancy should be detected inside guarded function")
		suite.Require().ErrorIs(err, types.ErrReentrancyDetected)

		return nil
	})

	suite.Require().NoError(err)
	suite.Require().Equal(1, executedCount, "function should execute once")

	// After WithReentrancyGuard completes, lock should be released
	err = k.EnterNoReentrant(ctx, key)
	suite.Require().NoError(err, "lock should be released after WithReentrancyGuard")
	k.ExitNoReentrant(ctx, key)
}

func (suite *SecurityGuardsTestSuite) TestReentrancyGuard_ConcurrentOperations() {
	ctx := suite.ctx
	k := suite.keeper

	// Simulate concurrent operations on different resources
	// In real blockchain, different transactions would use different ctx
	keys := []string{
		"swap:pool1:user1",
		"swap:pool1:user2",
		"swap:pool2:user1",
		"liquidity:pool1:user1",
	}

	// All should be able to acquire locks
	for _, key := range keys {
		err := k.EnterNoReentrant(ctx, key)
		suite.Require().NoError(err, "different keys should not block each other")
	}

	// All should be locked now, reentering any should fail
	for _, key := range keys {
		err := k.EnterNoReentrant(ctx, key)
		suite.Require().Error(err, "all keys should be locked")
		suite.Require().ErrorIs(err, types.ErrReentrancyDetected)
	}

	// Release all
	for _, key := range keys {
		k.ExitNoReentrant(ctx, key)
	}
}

// =============================================================================
// Pause Guard Tests
// =============================================================================

func (suite *SecurityGuardsTestSuite) TestPauseGuard_BasicPause() {
	ctx := suite.ctx
	k := suite.keeper

	moduleName := "dex"
	authority := k.GetAuthority()

	// Initially not paused
	err := k.RequireNotPaused(ctx, moduleName)
	suite.Require().NoError(err)
	suite.Require().False(k.IsModulePaused(ctx, moduleName))

	// Pause the module
	err = k.PauseModule(ctx, moduleName, authority)
	suite.Require().NoError(err)

	// Should now be paused
	suite.Require().True(k.IsModulePaused(ctx, moduleName))

	// RequireNotPaused should return error
	err = k.RequireNotPaused(ctx, moduleName)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrSystemPaused)

	// Unpause
	err = k.UnpauseModule(ctx, moduleName, authority)
	suite.Require().NoError(err)

	// Should be operational again
	err = k.RequireNotPaused(ctx, moduleName)
	suite.Require().NoError(err)
	suite.Require().False(k.IsModulePaused(ctx, moduleName))
}

func (suite *SecurityGuardsTestSuite) TestPauseGuard_UnauthorizedPause() {
	ctx := suite.ctx
	k := suite.keeper

	moduleName := "dex"
	unauthorizedUser := "aura1unauthorized"

	// Unauthorized user cannot pause
	err := k.PauseModule(ctx, moduleName, unauthorizedUser)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrUnauthorized)

	// Module should still not be paused
	suite.Require().False(k.IsModulePaused(ctx, moduleName))

	// Pause with authority
	authority := k.GetAuthority()
	err = k.PauseModule(ctx, moduleName, authority)
	suite.Require().NoError(err)

	// Unauthorized user cannot unpause
	err = k.UnpauseModule(ctx, moduleName, unauthorizedUser)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrUnauthorized)

	// Module should still be paused
	suite.Require().True(k.IsModulePaused(ctx, moduleName))
}

func (suite *SecurityGuardsTestSuite) TestPauseGuard_MultipleModules() {
	ctx := suite.ctx
	k := suite.keeper
	authority := k.GetAuthority()

	modules := []string{"dex", "bridge", "governance"}

	// Initially all operational
	for _, mod := range modules {
		err := k.RequireNotPaused(ctx, mod)
		suite.Require().NoError(err)
	}

	// Pause first two modules
	err := k.PauseModule(ctx, "dex", authority)
	suite.Require().NoError(err)
	err = k.PauseModule(ctx, "bridge", authority)
	suite.Require().NoError(err)

	// Check status
	suite.Require().True(k.IsModulePaused(ctx, "dex"))
	suite.Require().True(k.IsModulePaused(ctx, "bridge"))
	suite.Require().False(k.IsModulePaused(ctx, "governance"))

	// Governance should still work
	err = k.RequireNotPaused(ctx, "governance")
	suite.Require().NoError(err)
}

func (suite *SecurityGuardsTestSuite) TestPauseGuard_DoublePause() {
	ctx := suite.ctx
	k := suite.keeper
	authority := k.GetAuthority()

	moduleName := "dex"

	// First pause should succeed
	err := k.PauseModule(ctx, moduleName, authority)
	suite.Require().NoError(err)

	// Second pause should fail (already paused)
	err = k.PauseModule(ctx, moduleName, authority)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInvalidState)
}

func (suite *SecurityGuardsTestSuite) TestPauseGuard_UnpauseNotPaused() {
	ctx := suite.ctx
	k := suite.keeper
	authority := k.GetAuthority()

	moduleName := "dex"

	// Unpause when not paused should fail
	err := k.UnpauseModule(ctx, moduleName, authority)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrInvalidState)
}

// =============================================================================
// Rate Limiting Tests
// =============================================================================

func (suite *SecurityGuardsTestSuite) TestRateLimit_BasicLimiting() {
	ctx := suite.ctx
	k := suite.keeper

	key := "operation:user1"
	limit := uint64(3)
	window := 1 * time.Minute

	// First 3 operations should succeed
	for i := 0; i < 3; i++ {
		err := k.CheckRateLimit(ctx, key, limit, window)
		suite.Require().NoError(err, "operation %d should be within limit", i+1)
		k.IncrementRateLimit(ctx, key, window)
	}

	// 4th operation should fail (limit exceeded)
	err := k.CheckRateLimit(ctx, key, limit, window)
	suite.Require().Error(err, "operation should exceed rate limit")
	suite.Require().ErrorIs(err, types.ErrRateLimitExceeded)
}

func (suite *SecurityGuardsTestSuite) TestRateLimit_WindowExpiry() {
	ctx := suite.ctx
	k := suite.keeper

	key := "operation:user1"
	limit := uint64(2)
	window := 1 * time.Second

	// Perform 2 operations (reach limit)
	for i := 0; i < 2; i++ {
		err := k.CheckRateLimit(ctx, key, limit, window)
		suite.Require().NoError(err)
		k.IncrementRateLimit(ctx, key, window)
	}

	// Should be at limit
	err := k.CheckRateLimit(ctx, key, limit, window)
	suite.Require().Error(err)

	// Advance time beyond window
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))

	// Old operations should have expired, new operation should succeed
	err = k.CheckRateLimit(ctx, key, limit, window)
	suite.Require().NoError(err, "rate limit should reset after window expiry")
}

func (suite *SecurityGuardsTestSuite) TestRateLimit_DifferentUsers() {
	ctx := suite.ctx
	k := suite.keeper

	limit := uint64(2)
	window := 1 * time.Minute

	user1Key := "operation:user1"
	user2Key := "operation:user2"

	// User1 performs 2 operations (reaches limit)
	for i := 0; i < 2; i++ {
		err := k.CheckRateLimit(ctx, user1Key, limit, window)
		suite.Require().NoError(err)
		k.IncrementRateLimit(ctx, user1Key, window)
	}

	// User1 should be at limit
	err := k.CheckRateLimit(ctx, user1Key, limit, window)
	suite.Require().Error(err)

	// User2 should still have full limit (different key)
	for i := 0; i < 2; i++ {
		err := k.CheckRateLimit(ctx, user2Key, limit, window)
		suite.Require().NoError(err, "user2 should have independent rate limit")
		k.IncrementRateLimit(ctx, user2Key, window)
	}
}

func (suite *SecurityGuardsTestSuite) TestRateLimit_SlidingWindow() {
	ctx := suite.ctx
	k := suite.keeper

	key := "operation:user1"
	limit := uint64(3)
	window := 10 * time.Second

	// Perform 3 operations at time T
	for i := 0; i < 3; i++ {
		k.IncrementRateLimit(ctx, key, window)
	}

	// At T+5s, should still be at limit (all 3 operations in window)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(5 * time.Second))
	err := k.CheckRateLimit(ctx, key, limit, window)
	suite.Require().Error(err, "operations should still be within window")

	// At T+11s, operations should have expired
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(6 * time.Second))
	err = k.CheckRateLimit(ctx, key, limit, window)
	suite.Require().NoError(err, "operations should have expired")
}

// =============================================================================
// Input Validation Tests
// =============================================================================

func (suite *SecurityGuardsTestSuite) TestValidateAddress() {
	k := suite.keeper

	// Valid address
	validAddr := "aura1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq5dncz2"
	err := k.ValidateAddress(validAddr)
	suite.Require().NoError(err)

	// Empty address
	err = k.ValidateAddress("")
	suite.Require().Error(err)

	// Invalid address format
	err = k.ValidateAddress("invalid_address")
	suite.Require().Error(err)
}

func (suite *SecurityGuardsTestSuite) TestValidateAmount() {
	k := suite.keeper

	min := math.NewInt(100)
	max := math.NewInt(10000)

	// Valid amount
	amount := math.NewInt(500)
	err := k.ValidateAmount(amount, min, max)
	suite.Require().NoError(err)

	// Below minimum
	amount = math.NewInt(50)
	err = k.ValidateAmount(amount, min, max)
	suite.Require().Error(err)

	// Above maximum
	amount = math.NewInt(20000)
	err = k.ValidateAmount(amount, min, max)
	suite.Require().Error(err)

	// Negative
	amount = math.NewInt(-100)
	err = k.ValidateAmount(amount, min, max)
	suite.Require().Error(err)

	// Zero maximum means no upper limit
	amount = math.NewInt(1000000)
	err = k.ValidateAmount(amount, min, math.ZeroInt())
	suite.Require().NoError(err, "zero max should mean unlimited")
}

func (suite *SecurityGuardsTestSuite) TestValidateStringLength() {
	k := suite.keeper

	// Valid length
	err := k.ValidateStringLength("hello", "field", 3, 10)
	suite.Require().NoError(err)

	// Too short
	err = k.ValidateStringLength("hi", "field", 3, 10)
	suite.Require().Error(err)

	// Too long
	err = k.ValidateStringLength("this is a very long string", "field", 3, 10)
	suite.Require().Error(err)

	// Zero maxLen means no upper limit
	err = k.ValidateStringLength("very long string that should be accepted", "field", 3, 0)
	suite.Require().NoError(err)
}

// =============================================================================
// Audit Logging Tests
// =============================================================================

func (suite *SecurityGuardsTestSuite) TestAuditLogging() {
	ctx := suite.ctx
	k := suite.keeper

	eventType := "transfer"
	severity := "high"
	actor := "aura1actor"
	action := "large_transfer"
	details := "transferred 1000000 tokens"

	// Log security event
	k.LogSecurityEvent(ctx, eventType, severity, actor, action, details)

	// Verify event was emitted
	events := ctx.EventManager().Events()
	suite.Require().Greater(len(events), 0, "event should be emitted")

	// Find the security audit event
	found := false
	for _, event := range events {
		if event.Type == "security_audit" {
			found = true
			// Verify attributes
			for _, attr := range event.Attributes {
				if attr.Key == "event_type" {
					suite.Require().Equal(eventType, attr.Value)
				}
				if attr.Key == "severity" {
					suite.Require().Equal(severity, attr.Value)
				}
			}
			break
		}
	}
	suite.Require().True(found, "security_audit event should be emitted")
}

// =============================================================================
// Integration Tests - Real-World Scenarios
// =============================================================================

func (suite *SecurityGuardsTestSuite) TestIntegration_SwapWithAllGuards() {
	ctx := suite.ctx
	k := suite.keeper

	poolID := "pool1"
	sender := "aura1sender"
	moduleName := "dex"

	// Simulate swap operation with all security guards

	// 1. Check pause status
	err := k.RequireNotPaused(ctx, moduleName)
	suite.Require().NoError(err, "module should not be paused")

	// 2. Reentrancy guard
	reentrancyKey := fmt.Sprintf("swap:%s:%s", poolID, sender)
	err = k.EnterNoReentrant(ctx, reentrancyKey)
	suite.Require().NoError(err, "reentrancy guard should allow first entry")
	defer k.ExitNoReentrant(ctx, reentrancyKey)

	// 3. Rate limiting
	rateLimitKey := fmt.Sprintf("swap:%s", sender)
	err = k.CheckRateLimit(ctx, rateLimitKey, 100, time.Minute)
	suite.Require().NoError(err, "rate limit should allow operation")
	k.IncrementRateLimit(ctx, rateLimitKey, time.Minute)

	// 4. Input validation
	amount := math.NewInt(1000)
	err = k.ValidateAmount(amount, math.NewInt(1), math.NewInt(1000000))
	suite.Require().NoError(err, "amount should be valid")

	// 5. Audit logging
	k.LogSecurityEvent(ctx, "swap", "info", sender, "swap_executed", fmt.Sprintf("pool=%s,amount=%s", poolID, amount))

	// Verify all guards worked correctly
	events := ctx.EventManager().Events()
	suite.Require().Greater(len(events), 0, "events should be emitted")
}

func (suite *SecurityGuardsTestSuite) TestIntegration_EmergencyPauseScenario() {
	ctx := suite.ctx
	k := suite.keeper
	authority := k.GetAuthority()

	// Simulate emergency pause scenario

	// Normal operation
	err := k.RequireNotPaused(ctx, "dex")
	suite.Require().NoError(err)

	user := "aura1user"
	reentrancyKey := "swap:pool1:user"

	// User can perform operations
	err = k.EnterNoReentrant(ctx, reentrancyKey)
	suite.Require().NoError(err)
	k.ExitNoReentrant(ctx, reentrancyKey)

	// EMERGENCY: Exploit detected, pause the module
	err = k.PauseModule(ctx, "dex", authority)
	suite.Require().NoError(err)

	// Now all operations should fail
	err = k.RequireNotPaused(ctx, "dex")
	suite.Require().Error(err, "operations should fail when paused")

	// Even reentrancy checks won't be reached due to pause
	// (in real code, RequireNotPaused is checked first)

	// After fix is deployed, unpause
	err = k.UnpauseModule(ctx, "dex", authority)
	suite.Require().NoError(err)

	// Operations resume
	err = k.RequireNotPaused(ctx, "dex")
	suite.Require().NoError(err, "operations should resume after unpause")
}

func TestValidationFunctions(t *testing.T) {
	// Additional standalone tests
	t.Run("ValidateNonEmpty", func(t *testing.T) {
		ctx, keeper := testutil.NewSecurityKeeperForTest(t)
		_ = ctx

		err := keeper.ValidateNonEmpty("value", "field")
		require.NoError(t, err)

		err = keeper.ValidateNonEmpty("", "field")
		require.Error(t, err)
	})
}
