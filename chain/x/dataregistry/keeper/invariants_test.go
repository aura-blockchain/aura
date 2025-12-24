package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/params"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

func TestAllInvariants(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create params store
	paramsStore := params.NewStore(types.DefaultParams())

	// Create keeper
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	// Add test data
	item1 := types.DataItem{
		DataId:          "test-data-1",
		OwnerAddress:    "aura1owner",
		DataType:        types.DataItemType_DATA_ITEM_TYPE_PHOTO,
		ContentHash:     []byte("hash1"),
		StorageLocation: "ipfs://test1",
		Status:          types.DataItemStatus_DATA_ITEM_STATUS_PENDING_VERIFICATION,
	}

	require.NoError(t, k.SetDataItem(input.Ctx, item1))

	// Test all invariants pass
	msg, broken := keeper.AllInvariants(k)()
	require.False(t, broken, "invariants should not be broken")
	require.Empty(t, msg, "no invariant violations expected")
}

func TestRegisterInvariants(t *testing.T) {
	// Test that registering invariants doesn't panic
	require.NotPanics(t, func() {
		input := keepertest.CreateTestInput(t)
		paramsStore := params.NewStore(types.DefaultParams())
		k := keeper.NewKeeper(
			keepertest.WrapStoreService(input.StoreKey),
			input.Cdc,
			paramsStore,
			"aura1authority",
			keepertest.Logger(),
		)

		// Register invariants
		keeper.RegisterInvariants(k)
	})
}

func TestParamsInvariant(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	// Create params store with valid params
	paramsStore := params.NewStore(types.DefaultParams())

	// Create keeper
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		paramsStore,
		"aura1authority",
		keepertest.Logger(),
	)

	t.Run("ValidParams", func(t *testing.T) {
		// Test params invariant with valid params
		msg, broken := keeper.ParamsInvariant(k)()
		require.False(t, broken, "params invariant should not be broken with valid params")
		require.Empty(t, msg, "no invariant violations expected")
	})

	t.Run("CustomParams", func(t *testing.T) {
		// Set custom params
		customParams := types.DefaultParams()
		customParams.MaxStorageBytes = 10000000
		require.NoError(t, k.SetParams(customParams))

		// Test params invariant still passes
		msg, broken := keeper.ParamsInvariant(k)()
		require.False(t, broken, "params invariant should not be broken with custom params")
		require.Empty(t, msg, "no invariant violations expected")

		// Verify params were actually set
		retrieved, _ := k.GetParams(input.Ctx)
		require.Equal(t, uint64(10000000), retrieved.MaxStorageBytes)
	})
}
