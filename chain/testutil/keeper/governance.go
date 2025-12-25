// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/governance/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

// MockStakingKeeper is a mock implementation of the StakingKeeper interface for testing
type MockStakingKeeper struct {
	delegatorBonded map[string]sdkmath.Int
}

// NewMockStakingKeeper creates a new mock staking keeper
func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		delegatorBonded: make(map[string]sdkmath.Int),
	}
}

// GetDelegatorBonded returns the bonded tokens for a delegator
func (m *MockStakingKeeper) GetDelegatorBonded(ctx context.Context, delegator sdk.AccAddress) (sdkmath.Int, error) {
	if amount, ok := m.delegatorBonded[delegator.String()]; ok {
		return amount, nil
	}
	return sdkmath.ZeroInt(), nil
}

// TotalBondedTokens returns the total bonded tokens across all delegators
func (m *MockStakingKeeper) TotalBondedTokens(ctx context.Context) (sdkmath.Int, error) {
	total := sdkmath.ZeroInt()
	for _, amount := range m.delegatorBonded {
		total = total.Add(amount)
	}
	return total, nil
}

// SetDelegatorBonded sets the bonded tokens for a delegator (test helper)
func (m *MockStakingKeeper) SetDelegatorBonded(delegator sdk.AccAddress, amount sdkmath.Int) {
	m.delegatorBonded[delegator.String()] = amount
}

const (
	GovernanceStoreKey = "governance"
)

// GovernanceKeeper creates a governance keeper for testing
func GovernanceKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(GovernanceStoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create mock keepers
	stakingKeeper := NewMockStakingKeeper()
	bankKeeper := NewMockBankKeeper()
	securityKeeper := NewMockSecurityKeeper()

	k := keeper.NewKeeper(cdc, storeKey, stakingKeeper, bankKeeper, securityKeeper)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	k.SetParams(ctx, params)

	return k, ctx
}

// MockBankKeeper is a mock bank keeper for testing
type MockBankKeeper struct{}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{}
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.ZeroInt())
}

// MockSecurityKeeper is a mock security keeper for testing
type MockSecurityKeeper struct{}

func NewMockSecurityKeeper() *MockSecurityKeeper {
	return &MockSecurityKeeper{}
}

func (m *MockSecurityKeeper) EnterNoReentrant(ctx sdk.Context, key string) error {
	return nil
}

func (m *MockSecurityKeeper) ExitNoReentrant(ctx sdk.Context, key string) {}

func (m *MockSecurityKeeper) WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error {
	return fn()
}

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

func (m *MockSecurityKeeper) CheckGuardRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error {
	return nil
}

func (m *MockSecurityKeeper) IncrementGuardRateLimit(ctx sdk.Context, key string, window time.Duration) {}

func (m *MockSecurityKeeper) ValidateAddress(address string) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateAmount(amount sdkmath.Int, min, max sdkmath.Int) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateNonEmpty(value string, fieldName string) error {
	return nil
}

func (m *MockSecurityKeeper) ValidateStringLength(value string, fieldName string, minLen, maxLen int) error {
	return nil
}

func (m *MockSecurityKeeper) CheckAuthorization(ctx sdk.Context, address string, action string) error {
	return nil
}

func (m *MockSecurityKeeper) LogSecurityEvent(ctx sdk.Context, eventType string, severity string, actor string, action string, details string) {
}
