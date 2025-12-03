package testutil

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/security/keeper"
	"github.com/aequitas/aura/chain/x/security/types"
)

// NewSecurityKeeperForTest creates a new security keeper for testing
func NewSecurityKeeperForTest(t testing.TB) (sdk.Context, *keeper.Keeper) {
	// Create test codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store keys
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	// Create multi-store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	// Create context
	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Create mock keepers
	bankKeeper := &MockBankKeeper{}
	stakingKeeper := &MockStakingKeeper{}
	accountKeeper := &MockAccountKeeper{}

	// Create keeper
	authority := "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" // Test authority
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		authority,
		bankKeeper,
		stakingKeeper,
		accountKeeper,
	)

	// Initialize params
	params := types.DefaultParams()
	k.SetParams(ctx, params)

	return ctx, k
}

// MockBankKeeper is a mock bank keeper for testing
type MockBankKeeper struct{}

func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdk.NewInt(1000000))
}

func (m *MockBankKeeper) SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

// MockStakingKeeper is a mock staking keeper for testing
type MockStakingKeeper struct{}

func (m *MockStakingKeeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator interface{}, found bool) {
	return nil, false
}

func (m *MockStakingKeeper) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) error {
	return nil
}

func (m *MockStakingKeeper) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) error {
	return nil
}

func (m *MockStakingKeeper) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor string) (string, error) {
	return "0", nil
}

// MockAccountKeeper is a mock account keeper for testing
type MockAccountKeeper struct{}

func (m *MockAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m *MockAccountKeeper) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
}
