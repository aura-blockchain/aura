package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	sdkmath "cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// setupKeeper creates a test keeper for aiassistant module
func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	t.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		ChainID: "aiassistant-test",
		Time:    time.Now().UTC(),
	}, false, log.NewNopLogger())

	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	bank := newMockBankKeeper()
	k := NewKeeper(cdc, key, "", bank)
	return &k, ctx
}

// mockBankKeeper is a minimal bank keeper for testing
type mockBankKeeper struct{}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{}
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ sdk.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ sdk.Context, _, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) BurnCoins(_ sdk.Context, _ string, _ sdk.Coins) error {
	return nil
}

func TestRecordAnalytics(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name    string
		record  AnalyticsRecord
		wantErr bool
	}{
		{
			name: "valid analytics record",
			record: AnalyticsRecord{
				UserAddress:   "user1",
				ModelHash:     "model-hash-1",
				ComputeUnits:  100,
				Cost:          sdkmath.NewInt(1000),
				OperationType: "query",
				Success:       true,
				ResponseTime:  500,
			},
			wantErr: false,
		},
		{
			name: "high compute units",
			record: AnalyticsRecord{
				UserAddress:   "user2",
				ModelHash:     "model-hash-2",
				ComputeUnits:  1000000,
				Cost:          sdkmath.NewInt(10000000),
				OperationType: "training",
				Success:       true,
				ResponseTime:  5000,
			},
			wantErr: false,
		},
		{
			name: "failed operation",
			record: AnalyticsRecord{
				UserAddress:   "user3",
				ModelHash:     "model-hash-1",
				ComputeUnits:  50,
				Cost:          sdkmath.NewInt(500),
				OperationType: "inference",
				Success:       false,
				ResponseTime:  100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.RecordAnalytics(ctx, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordAnalytics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
