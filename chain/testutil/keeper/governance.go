package keeper

import (
	"testing"

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
func (m *MockStakingKeeper) GetDelegatorBonded(ctx sdk.Context, delegator sdk.AccAddress) sdkmath.Int {
	if amount, ok := m.delegatorBonded[delegator.String()]; ok {
		return amount
	}
	return sdkmath.ZeroInt()
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

	// Create mock staking keeper
	stakingKeeper := NewMockStakingKeeper()

	k := keeper.NewKeeper(cdc, storeKey, stakingKeeper)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	k.SetParams(ctx, params)

	return k, ctx
}
