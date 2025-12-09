package params

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dataregistry/types"
)

func TestStoreSetParamsValidation(t *testing.T) {
	store := NewStore(types.DefaultParams())

	current := store.GetParams()
	require.Equal(t, uint64(10485760), current.MaxStorageBytes)

	invalid := current
	invalid.MaxStorageBytes = 0

	err := store.SetParams(invalid)
	require.Error(t, err, "invalid params should be rejected")

	// Ensure params were not mutated on validation failure.
	stored := store.GetParams()
	require.Equal(t, current.MaxStorageBytes, stored.MaxStorageBytes)
}

func TestStoreSetParamsSuccess(t *testing.T) {
	store := NewStore(types.DefaultParams())

	updated := store.GetParams()
	updated.MaxDataItemsPerUser = 2000
	updated.StorageFee = "250"

	require.NoError(t, store.SetParams(updated))

	stored := store.GetParams()
	require.Equal(t, uint64(2000), stored.MaxDataItemsPerUser)
	require.Equal(t, "250", stored.StorageFee)
}
