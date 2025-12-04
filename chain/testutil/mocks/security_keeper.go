package mocks

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockSecurityKeeper is a mock implementation of the SecurityKeeper interface
// for testing purposes. All methods return success/no-op by default.
type MockSecurityKeeper struct{}

// NewMockSecurityKeeper creates a new MockSecurityKeeper instance
func NewMockSecurityKeeper() *MockSecurityKeeper {
	return &MockSecurityKeeper{}
}

// Reentrancy Protection methods

func (m *MockSecurityKeeper) EnterNoReentrant(ctx sdk.Context, key string) error {
	return nil
}

func (m *MockSecurityKeeper) ExitNoReentrant(ctx sdk.Context, key string) {}

func (m *MockSecurityKeeper) WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error {
	return fn()
}

// Emergency Pause (Circuit Breaker) methods

func (m *MockSecurityKeeper) RequireNotPaused(ctx sdk.Context, moduleName string) error {
	return nil
}

func (m *MockSecurityKeeper) PauseModule(ctx sdk.Context, moduleName string, pausedBy string) error {
	return nil
}

func (m *MockSecurityKeeper) UnpauseModule(ctx sdk.Context, moduleName string, unpausedBy string) error {
	return nil
}

func (m *MockSecurityKeeper) IsModulePaused(ctx sdk.Context, moduleName string) bool {
	return false
}

// Rate Limiting methods

func (m *MockSecurityKeeper) CheckGuardRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error {
	return nil
}

func (m *MockSecurityKeeper) IncrementGuardRateLimit(ctx sdk.Context, key string, window time.Duration) {}

// Input Validation methods

func (m *MockSecurityKeeper) ValidateAddress(address string) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateAmount(amount math.Int, min, max math.Int) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateNonEmpty(value string, fieldName string) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateStringLength(value string, fieldName string, minLen, maxLen int) error {
	return nil
}

// Access Control methods

func (m *MockSecurityKeeper) CheckAuthorization(ctx sdk.Context, address string, action string) error {
	return nil
}

// Audit Logging methods

func (m *MockSecurityKeeper) LogSecurityEvent(ctx sdk.Context, eventType string, severity string, actor string, action string, details string) {
	// No-op for mock
}
