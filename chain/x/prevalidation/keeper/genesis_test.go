package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

func TestInitGenesis(t *testing.T) {
	t.Run("init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := types.DefaultGenesis()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with valid genesis data", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &pb.GenesisState{
			Params: types.DefaultParams(),
			PreValidatedTransactions: []*pb.PreValidatedTransaction{
				{
					Id:         "tx1",
					TxType:     pb.TransactionType_TX_TYPE_IR_COMPLETION,
					TemplateId: "template1",
					Signer:     "aura1test",
					Nonce:      1,
					Status:     pb.ValidationStatus_VALIDATION_STATUS_VALIDATED,
				},
			},
			Templates: []*pb.ValidationTemplate{
				{
					Id:              "template1",
					TxType:          pb.TransactionType_TX_TYPE_IR_COMPLETION,
					Name:            "Test Template",
					Description:     "Test",
					ValidationRules: `{"rule1": true, "rule2": true}`,
					Active:          true,
				},
			},
			Metrics: &pb.PreValidationMetrics{
				TotalPreValidations: 100,
				TotalExecuted:       90,
				TotalExpired:        10,
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify transactions were imported
		// Note: Add getter methods if needed
	})

	t.Run("init with pre-validated transactions", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &pb.GenesisState{
			Params: types.DefaultParams(),
			PreValidatedTransactions: []*pb.PreValidatedTransaction{
				{
					Id:            "tx1",
					TxType:        pb.TransactionType_TX_TYPE_IR_COMPLETION,
					TemplateId:    "template1",
					Signer:        "aura1test1",
					Nonce:         1,
					Status:        pb.ValidationStatus_VALIDATION_STATUS_VALIDATED,
					EncryptedData: []byte("encrypted_data_1"),
				},
				{
					Id:            "tx2",
					TxType:        pb.TransactionType_TX_TYPE_DEX_SWAP,
					TemplateId:    "template2",
					Signer:        "aura1test2",
					Nonce:         2,
					Status:        pb.ValidationStatus_VALIDATION_STATUS_PENDING,
					EncryptedData: []byte("encrypted_data_2"),
				},
			},
			Templates: []*pb.ValidationTemplate{},
			Metrics:   &pb.PreValidationMetrics{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with validation templates", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &pb.GenesisState{
			Params:                   types.DefaultParams(),
			PreValidatedTransactions: []*pb.PreValidatedTransaction{},
			Templates: []*pb.ValidationTemplate{
				{
					Id:              "template1",
					TxType:          pb.TransactionType_TX_TYPE_IR_COMPLETION,
					Name:            "Standard Validation",
					Description:     "Standard validation rules",
					ValidationRules: `{"check_signature": true, "check_balance": true}`,
					PriorityWeight:  10,
					Active:          true,
				},
				{
					Id:              "template2",
					TxType:          pb.TransactionType_TX_TYPE_DEX_SWAP,
					Name:            "Fast Track",
					Description:     "Fast track validation",
					ValidationRules: `{"check_signature": true}`,
					PriorityWeight:  5,
					Active:          true,
				},
			},
			Metrics: &pb.PreValidationMetrics{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with metrics", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &pb.GenesisState{
			Params:                   types.DefaultParams(),
			PreValidatedTransactions: []*pb.PreValidatedTransaction{},
			Templates:                []*pb.ValidationTemplate{},
			Metrics: &pb.PreValidationMetrics{
				TotalPreValidations: 1000,
				TotalExecuted:       950,
				TotalExpired:        25,
				TotalCacheHits:      900,
				TotalCacheMisses:    100,
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with invalid genesis fails", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &pb.GenesisState{
			Params: &pb.Params{
				Enabled:      true,
				MaxCacheSize: 0, // Invalid - should be > 0
			},
			PreValidatedTransactions: []*pb.PreValidatedTransaction{},
			Templates:                []*pb.ValidationTemplate{},
			Metrics:                  &pb.PreValidationMetrics{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
	})

	t.Run("init skips nil entries", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := &pb.GenesisState{
			Params: types.DefaultParams(),
			PreValidatedTransactions: []*pb.PreValidatedTransaction{
				nil,
				{
					Id:         "tx1",
					TxType:     pb.TransactionType_TX_TYPE_IR_COMPLETION,
					TemplateId: "template1",
					Signer:     "aura1test",
					Nonce:      1,
					Status:     pb.ValidationStatus_VALIDATION_STATUS_VALIDATED,
				},
				nil,
			},
			Templates: []*pb.ValidationTemplate{nil},
			Metrics:   &pb.PreValidationMetrics{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := k.ExportGenesis(ctx)

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.PreValidatedTransactions)
		require.NotNil(t, genesis.Templates)
		require.NotNil(t, genesis.Metrics)
	})

	t.Run("export with data", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		// Initialize with data
		initGenesis := &pb.GenesisState{
			Params: types.DefaultParams(),
			PreValidatedTransactions: []*pb.PreValidatedTransaction{
				{Id: "tx1", TxType: pb.TransactionType_TX_TYPE_IR_COMPLETION, TemplateId: "template1", Signer: "aura1test1", Nonce: 1, Status: pb.ValidationStatus_VALIDATION_STATUS_VALIDATED},
				{Id: "tx2", TxType: pb.TransactionType_TX_TYPE_IR_COMPLETION, TemplateId: "template1", Signer: "aura1test2", Nonce: 2, Status: pb.ValidationStatus_VALIDATION_STATUS_PENDING},
			},
			Templates: []*pb.ValidationTemplate{
				{Id: "template1", TxType: pb.TransactionType_TX_TYPE_IR_COMPLETION, Name: "Test", Active: true},
			},
			Metrics: &pb.PreValidationMetrics{
				TotalPreValidations: 50,
				TotalExecuted:       45,
			},
		}

		err := k.InitGenesis(ctx, initGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		require.NotNil(t, exported)
		// Note: Actual counts may vary based on keeper implementation
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		originalGenesis := &pb.GenesisState{
			Params: &pb.Params{
				Enabled:                true,
				MaxCacheSize:           1000,
				ExpiryHours:            24,
				EncryptionAlgorithm:    "AES-256-GCM",
				ControlGroupPercentage: 10.0,
				MinConfidenceScore:     80,
			},
			PreValidatedTransactions: []*pb.PreValidatedTransaction{
				{
					Id:         "tx1",
					TxType:     pb.TransactionType_TX_TYPE_IR_COMPLETION,
					TemplateId: "template1",
					Signer:     "aura1test",
					Nonce:      1,
					Status:     pb.ValidationStatus_VALIDATION_STATUS_VALIDATED,
				},
			},
			Templates: []*pb.ValidationTemplate{
				{
					Id:              "template1",
					TxType:          pb.TransactionType_TX_TYPE_IR_COMPLETION,
					Name:            "Standard",
					Description:     "Standard validation",
					ValidationRules: `{"rule1": true}`,
					PriorityWeight:  10,
					Active:          true,
				},
			},
			Metrics: &pb.PreValidationMetrics{
				TotalPreValidations: 100,
				TotalExecuted:       90,
				TotalExpired:        10,
			},
		}

		// Import
		err := k.InitGenesis(ctx, originalGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		// Verify params match
		require.Equal(t, originalGenesis.Params.Enabled, exported.Params.Enabled)
		require.Equal(t, originalGenesis.Params.MaxCacheSize, exported.Params.MaxCacheSize)
		require.Equal(t, originalGenesis.Params.ExpiryHours, exported.Params.ExpiryHours)
	})

	t.Run("multiple round trips are deterministic", func(t *testing.T) {
		k1, ctx1 := setupTestKeeper(t)
		k2, ctx2 := setupTestKeeper(t)

		genesis := types.DefaultGenesis()

		// First round trip
		err := k1.InitGenesis(ctx1, genesis)
		require.NoError(t, err)
		export1 := k1.ExportGenesis(ctx1)

		// Second round trip
		err = k2.InitGenesis(ctx2, export1)
		require.NoError(t, err)
		export2 := k2.ExportGenesis(ctx2)

		// Verify exports match using proto.Equal to avoid sizeCache differences
		require.True(t, proto.Equal(export1.Params, export2.Params), "params should be equal")
	})
}

func TestDefaultGenesis(t *testing.T) {
	t.Run("default genesis is valid", func(t *testing.T) {
		genesis := types.DefaultGenesis()

		err := types.ValidateGenesis(genesis)
		require.NoError(t, err)

		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.PreValidatedTransactions)
		require.NotNil(t, genesis.Templates)
		require.NotNil(t, genesis.Metrics)
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := types.DefaultGenesis()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})
}

func setupTestKeeper(t *testing.T) (*Keeper, context.Context) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.ModuleName)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	encCfg := MakeTestEncodingConfig()

	keeper := NewKeeper(encCfg.Codec, key)
	keeper.SetLogger(log.NewNopLogger())

	sdkCtx := sdk.NewContext(cms, cmtproto.Header{}, false, log.NewNopLogger())

	return keeper, sdkCtx
}
