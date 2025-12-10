package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestUpdateInflationCheckTimestampDebug(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	params := keeper.GetParams()
	params.Tokenomics = &types.TokenomicsConfig{
		MaxSupply:         "1000000000",
		CirculatingSupply: "100000000",
		InflationRate:     500,
	}
	require.NoError(t, keeper.SetParams(params))

	// Check the params before update
	beforeParams := keeper.GetParams()
	t.Logf("Before: LastInflationCheck.IsZero() = %v", beforeParams.Tokenomics.LastInflationCheck.IsZero())
	t.Logf("Before: LastInflationCheck = %v", beforeParams.Tokenomics.LastInflationCheck)

	// Get block time from context
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	t.Logf("BlockTime = %v", sdkCtx.BlockTime())
	t.Logf("BlockTime.IsZero() = %v", sdkCtx.BlockTime().IsZero())

	// Update timestamp
	err := keeper.UpdateInflationCheckTimestamp(ctx)
	require.NoError(t, err)

	// Verify timestamp was set
	updatedParams := keeper.GetParams()
	t.Logf("After: LastInflationCheck.IsZero() = %v", updatedParams.Tokenomics.LastInflationCheck.IsZero())
	t.Logf("After: LastInflationCheck = %v", updatedParams.Tokenomics.LastInflationCheck)
	t.Logf("After: Tokenomics pointer = %p", updatedParams.Tokenomics)
	
	require.False(t, updatedParams.Tokenomics.LastInflationCheck.IsZero())

	// Verify it's recent (within last minute)
	checkTime := updatedParams.Tokenomics.LastInflationCheck
	require.True(t, time.Since(checkTime) < time.Minute)
}
