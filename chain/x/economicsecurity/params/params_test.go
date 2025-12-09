package params

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestStoreSetParamsValidation(t *testing.T) {
	store := NewStore(*types.DefaultParams())

	current := store.GetParams()
	require.NotEmpty(t, current.Tokenomics.MaxSupply)

	invalid := current
	invalid.Tokenomics.MaxSupply = ""

	err := store.SetParams(invalid)
	require.Error(t, err, "invalid params should be rejected")

	// Verify the original params were not mutated when validation failed.
	stored := store.GetParams()
	require.Equal(t, current.Tokenomics.MaxSupply, stored.Tokenomics.MaxSupply)
}

func TestStoreSetParamsSuccess(t *testing.T) {
	store := NewStore(*types.DefaultParams())

	update := store.GetParams()
	update.Tokenomics.InflationRate = 1200

	err := store.SetParams(update)
	require.NoError(t, err)

	applied := store.GetParams()
	require.Equal(t, uint64(1200), applied.Tokenomics.InflationRate)
	require.Equal(t, update.Governance.QuorumPercentage, applied.Governance.QuorumPercentage)
}
