package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/privacy/types"
	pb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

func TestInitGenesis(t *testing.T) {
	t.Run("init with nil genesis", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		err := k.InitGenesisProto(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesisProto(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params := k.GetParams(ctx)
		require.NotNil(t, params)
	})

	t.Run("init with custom params", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		genesis := &pb.GenesisState{
			Params: pb.Params{
				EnableZkProofs:                 true,
				EnableStealthAddresses:         true,
				EnableRingSignatures:           true,
				EnableConfidentialTransactions: true,
				EnableNetworkPrivacy:           false,
				EnableMixing:                   true,
				MinRingSize:                    5,
				MaxRingSize:                    11,
				MinMixingParticipants:          3,
				MixingFee:                      sdkmath.NewInt(100),
				ZkProofVerificationCost:        50,
			},
			MixingPools:        []*pb.MixingPool{},
			RegisteredViewKeys: []*pb.ViewKey{},
		}

		err := k.InitGenesisProto(ctx, genesis)
		require.NoError(t, err)

		// Verify custom params
		params := k.GetParams(ctx)
		require.True(t, params.EnableZkProofs)
		require.True(t, params.EnableStealthAddresses)
		require.True(t, params.EnableRingSignatures)
		require.Equal(t, uint32(5), params.MinRingSize)
		require.Equal(t, uint32(11), params.MaxRingSize)
	})

	t.Run("init with all privacy features enabled", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		genesis := &pb.GenesisState{
			Params: pb.Params{
				EnableZkProofs:                 true,
				EnableStealthAddresses:         true,
				EnableRingSignatures:           true,
				EnableConfidentialTransactions: true,
				EnableNetworkPrivacy:           true,
				EnableMixing:                   true,
				MinRingSize:                    3,
				MaxRingSize:                    15,
				MinMixingParticipants:          2,
				MixingFee:                      sdkmath.NewInt(200),
				ZkProofVerificationCost:        100,
			},
			MixingPools:        []*pb.MixingPool{},
			RegisteredViewKeys: []*pb.ViewKey{},
		}

		err := k.InitGenesisProto(ctx, genesis)
		require.NoError(t, err)

		params := k.GetParams(ctx)
		require.True(t, params.EnableZkProofs)
		require.True(t, params.EnableConfidentialTransactions)
		require.True(t, params.EnableNetworkPrivacy)
		require.True(t, params.EnableMixing)
	})

	t.Run("init with minimal privacy features", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		genesis := &pb.GenesisState{
			Params: pb.Params{
				EnableZkProofs:                 false,
				EnableStealthAddresses:         false,
				EnableRingSignatures:           false,
				EnableConfidentialTransactions: false,
				EnableNetworkPrivacy:           false,
				EnableMixing:                   false,
				MinRingSize:                    3,
				MaxRingSize:                    7,
				MinMixingParticipants:          2,
				MixingFee:                      sdkmath.NewInt(50),
				ZkProofVerificationCost:        25,
			},
			MixingPools:        []*pb.MixingPool{},
			RegisteredViewKeys: []*pb.ViewKey{},
		}

		err := k.InitGenesisProto(ctx, genesis)
		require.NoError(t, err)

		params := k.GetParams(ctx)
		require.False(t, params.EnableZkProofs)
		require.False(t, params.EnableConfidentialTransactions)
		require.False(t, params.EnableNetworkPrivacy)
	})

	t.Run("init with invalid ring size fails", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		genesis := &pb.GenesisState{
			Params: pb.Params{
				EnableZkProofs:                 true,
				EnableStealthAddresses:         true,
				EnableRingSignatures:           true,
				EnableConfidentialTransactions: true,
				EnableNetworkPrivacy:           true,
				EnableMixing:                   true,
				MinRingSize:                    10, // Invalid - greater than max
				MaxRingSize:                    5,
				MinMixingParticipants:          2,
				MixingFee:                      sdkmath.NewInt(100),
				ZkProofVerificationCost:        50,
			},
			MixingPools:        []*pb.MixingPool{},
			RegisteredViewKeys: []*pb.ViewKey{},
		}

		err := k.InitGenesisProto(ctx, genesis)
		require.Error(t, err)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		genesis := k.ExportGenesisProto(ctx)

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.MixingPools)
		require.NotNil(t, genesis.RegisteredViewKeys)
	})

	t.Run("export after init preserves params", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		originalGenesis := &pb.GenesisState{
			Params: pb.Params{
				EnableZkProofs:                 true,
				EnableStealthAddresses:         false,
				EnableRingSignatures:           true,
				EnableConfidentialTransactions: false,
				EnableNetworkPrivacy:           true,
				EnableMixing:                   false,
				MinRingSize:                    7,
				MaxRingSize:                    13,
				MinMixingParticipants:          4,
				MixingFee:                      sdkmath.NewInt(150),
				ZkProofVerificationCost:        75,
			},
			MixingPools:        []*pb.MixingPool{},
			RegisteredViewKeys: []*pb.ViewKey{},
		}

		err := k.InitGenesisProto(ctx, originalGenesis)
		require.NoError(t, err)

		exported := k.ExportGenesisProto(ctx)

		require.Equal(t, originalGenesis.Params.EnableZkProofs, exported.Params.EnableZkProofs)
		require.Equal(t, originalGenesis.Params.EnableStealthAddresses, exported.Params.EnableStealthAddresses)
		require.Equal(t, originalGenesis.Params.MinRingSize, exported.Params.MinRingSize)
		require.Equal(t, originalGenesis.Params.MaxRingSize, exported.Params.MaxRingSize)
		require.Equal(t, originalGenesis.Params.MixingFee, exported.Params.MixingFee)
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		originalGenesis := &pb.GenesisState{
			Params: pb.Params{
				EnableZkProofs:                 true,
				EnableStealthAddresses:         true,
				EnableRingSignatures:           true,
				EnableConfidentialTransactions: true,
				EnableNetworkPrivacy:           false,
				EnableMixing:                   true,
				MinRingSize:                    5,
				MaxRingSize:                    11,
				MinMixingParticipants:          3,
				MixingFee:                      sdkmath.NewInt(100),
				ZkProofVerificationCost:        50,
			},
			MixingPools:        []*pb.MixingPool{},
			RegisteredViewKeys: []*pb.ViewKey{},
		}

		// Import
		err := k.InitGenesisProto(ctx, originalGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesisProto(ctx)

		// Verify all params match
		require.Equal(t, originalGenesis.Params.EnableZkProofs, exported.Params.EnableZkProofs)
		require.Equal(t, originalGenesis.Params.EnableStealthAddresses, exported.Params.EnableStealthAddresses)
		require.Equal(t, originalGenesis.Params.EnableRingSignatures, exported.Params.EnableRingSignatures)
		require.Equal(t, originalGenesis.Params.EnableConfidentialTransactions, exported.Params.EnableConfidentialTransactions)
		require.Equal(t, originalGenesis.Params.EnableNetworkPrivacy, exported.Params.EnableNetworkPrivacy)
		require.Equal(t, originalGenesis.Params.EnableMixing, exported.Params.EnableMixing)
		require.Equal(t, originalGenesis.Params.MinRingSize, exported.Params.MinRingSize)
		require.Equal(t, originalGenesis.Params.MaxRingSize, exported.Params.MaxRingSize)
		require.Equal(t, originalGenesis.Params.MinMixingParticipants, exported.Params.MinMixingParticipants)
		require.Equal(t, originalGenesis.Params.MixingFee, exported.Params.MixingFee)
		require.Equal(t, originalGenesis.Params.ZkProofVerificationCost, exported.Params.ZkProofVerificationCost)
	})

	t.Run("multiple round trips are deterministic", func(t *testing.T) {
		k1, ctx1 := setupTestKeeperWithCtx(t)
		k2, ctx2 := setupTestKeeperWithCtx(t)

		genesis := types.DefaultGenesisState()
		genesis.Params.MinRingSize = 9
		genesis.Params.MaxRingSize = 17

		// First round trip
		err := k1.InitGenesisProto(ctx1, genesis)
		require.NoError(t, err)
		export1 := k1.ExportGenesisProto(ctx1)

		// Second round trip
		err = k2.InitGenesisProto(ctx2, export1)
		require.NoError(t, err)
		export2 := k2.ExportGenesisProto(ctx2)

		// Verify exports match
		require.Equal(t, export1.Params.MinRingSize, export2.Params.MinRingSize)
		require.Equal(t, export1.Params.MaxRingSize, export2.Params.MaxRingSize)
	})
}

func TestDefaultGenesis(t *testing.T) {
	t.Run("default genesis is valid", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.MixingPools)
		require.NotNil(t, genesis.RegisteredViewKeys)

		// Verify sensible defaults
		require.GreaterOrEqual(t, genesis.Params.MaxRingSize, genesis.Params.MinRingSize)
		require.Greater(t, genesis.Params.MinMixingParticipants, uint32(0))
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeperWithCtx(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesisProto(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params := k.GetParams(ctx)
		require.Equal(t, genesis.Params.EnableZkProofs, params.EnableZkProofs)
		require.Equal(t, genesis.Params.MinRingSize, params.MinRingSize)
	})

	t.Run("default genesis has reasonable values", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		// Check that ring sizes are valid
		require.Less(t, genesis.Params.MinRingSize, genesis.Params.MaxRingSize)
		require.GreaterOrEqual(t, genesis.Params.MinRingSize, uint32(3))

		// Check mixing params
		require.GreaterOrEqual(t, genesis.Params.MinMixingParticipants, uint32(2))
	})
}

func setupTestKeeper(t *testing.T) *Keeper {
	k, _ := setupTestKeeperWithCtx(t)
	return k
}

// setupTestKeeperWithCtx creates a test keeper with a proper SDK context
func setupTestKeeperWithCtx(t *testing.T) (*Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey("privacy")
	testCtx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test"))

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := NewKeeper(cdc, storeKey, nil, nil)
	return k, testCtx.Ctx
}

// NewTestKeeper creates a test keeper with a mock store
func NewTestKeeper(t *testing.T) *Keeper {
	k, _ := setupTestKeeperWithCtx(t)
	return k
}
