package keeper

import (
	"github.com/aequitas/aura/chain/testing/testutil"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/dex/types"
)

// FuzzGetQuote_ConstantProductBounds ensures GetQuote maintains invariants across random inputs.
func FuzzGetQuote_ConstantProductBounds(f *testing.F) {
	f.Add(uint64(1_000_000), uint64(2_000_000), uint64(10_000))
	f.Add(uint64(10_000), uint64(50_000), uint64(1_000))
	f.Add(uint64(5_000_000), uint64(5_000_000), uint64(100))

	f.Fuzz(func(t *testing.T, reserveA, reserveB, in uint64) {
		// Normalize inputs to keep calculations in safe range and avoid zero reserves.
		if reserveA == 0 {
			reserveA = 1
		}
		if reserveB == 0 {
			reserveB = 1
		}
		if in == 0 {
			in = 1
		}

		// Cap values to avoid hitting max swap guardrails or overflow in SafeMul.
		const max = uint64(1_000_000_000_000) // 1e12
		reserveA %= max
		reserveB %= max
		in %= max
		if in == 0 {
			in = 1
		}

		input := keepertest.CreateTestInput(t)
		k := NewKeeper(input.Cdc, input.StoreKey, testutil.NewMockBankKeeper(), testutil.NewMockAccountKeeper(), testutil.NewMockVCRegistryKeeper(), testutil.NewMockSecurityKeeper())
		ctx := input.Ctx

		pool := &types.LiquidityPool{
			PoolId:                "pool-a",
			DenomA:                "tokenA",
			DenomB:                "tokenB",
			ReserveA:              sdkmath.NewIntFromUint64(reserveA),
			ReserveB:              sdkmath.NewIntFromUint64(reserveB),
			FeePercentage:         sdkmath.LegacyMustNewDecFromStr("0.003"),
			ProtocolFeePercentage: sdkmath.LegacyMustNewDecFromStr("0.001"),
			TotalLpTokens:         sdkmath.NewIntFromUint64(reserveA + reserveB),
			Providers:             []types.LiquidityProvider{},
		}
		k.SetPool(ctx, pool)

		amountIn := sdkmath.NewIntFromUint64(in)
		out, price, impact, fee, err := k.GetQuote(ctx, pool.PoolId, pool.DenomA, amountIn)
		if err != nil {
			return // invalid combos are acceptable; focus on invariants for valid quotes
		}

		require.False(t, out.IsNegative())
		require.False(t, fee.IsNegative())
		require.True(t, out.LTE(sdkmath.NewIntFromUint64(reserveB)), "output should not exceed reserve")

		require.True(t, price.GTE(sdkmath.LegacyZeroDec()))
		require.True(t, impact.GTE(sdkmath.LegacyZeroDec()))
		require.True(t, impact.LTE(sdkmath.LegacyNewDec(100)), "price impact should be <=100%%")

		// Slippage: expected output with zero fees should be at least as large as fee-adjusted output.
		// Compare against a zero-fee quote to ensure fees never increase output.
		pool.FeePercentage = sdkmath.LegacyZeroDec()
		pool.ProtocolFeePercentage = sdkmath.LegacyZeroDec()
		k.SetPool(ctx, pool)
		noFeeOut, _, _, _, err := k.GetQuote(ctx, pool.PoolId, pool.DenomA, amountIn)
		if err == nil {
			require.True(t, noFeeOut.GTE(out))
		}
	})
}
