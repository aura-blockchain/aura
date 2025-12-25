// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package params

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

func TestStoreSetParamsValidation(t *testing.T) {
	store := NewStore(*types.DefaultParams())

	current := store.GetParams()
	require.NotZero(t, current.MaxVcsPerUser)

	invalid := current
	invalid.MaxMintPerHour = 0 // invalid: must be positive

	err := store.SetParams(invalid)
	require.Error(t, err)

	// Ensure previous params remain unchanged after failed set.
	stored := store.GetParams()
	require.Equal(t, current.MaxMintPerHour, stored.MaxMintPerHour)
}

func TestStoreSetParamsSuccess(t *testing.T) {
	store := NewStore(*types.DefaultParams())

	update := store.GetParams()
	update.MaxMintPerDay = 10
	update.MaxMintPerHour = 5
	update.DidNetwork = "testnet"

	require.NoError(t, store.SetParams(update))

	stored := store.GetParams()
	require.Equal(t, uint64(10), stored.MaxMintPerDay)
	require.Equal(t, uint64(5), stored.MaxMintPerHour)
	require.Equal(t, "testnet", stored.DidNetwork)
}
