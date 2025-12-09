package keeper

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestAnalyzeStakingIncentivesHandlesHugeValues(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000000000000000000000000000000" // 1e33
	params.Tokenomics.InflationRate = 9000

	result := k.analyzeStakingIncentives(*params, "500000000000000000000000000000000000") // 5e32

	rewards, ok := new(big.Int).SetString(result.rewards, 10)
	require.True(t, ok, "rewards should parse as big.Int")
	require.False(t, rewards.Sign() < 0, "rewards must be non-negative")
}

func FuzzAnalyzeStakingIncentives_NoOverflow(f *testing.F) {
	f.Add(uint64(1_000_000_000), uint64(500_000_000))
	f.Add(uint64(10), uint64(9))
	f.Add(^uint64(0)-1, ^uint64(0)/2)

	f.Fuzz(func(t *testing.T, supplyRaw uint64, stakedRaw uint64) {
		k, _ := setupKeeperForTest(t)

		params := types.DefaultParams()
		params.Tokenomics.InflationRate = 5000 // keep within bounds for validation

		supply := new(big.Int).SetUint64(supplyRaw)
		if supply.Sign() == 0 {
			supply.SetUint64(1)
		}

		staked := new(big.Int).SetUint64(stakedRaw)
		staked.Mod(staked, new(big.Int).Add(supply, big.NewInt(1))) // ensure staked <= supply

		params.Tokenomics.CirculatingSupply = supply.String()

		result := k.analyzeStakingIncentives(*params, staked.String())

		ratio := new(big.Int).Mul(staked, big.NewInt(10000))
		ratio.Div(ratio, supply)
		require.LessOrEqual(t, ratio.Uint64(), uint64(10000), "staking ratio should stay within 0-10000 bps")

		rewards, ok := new(big.Int).SetString(result.rewards, 10)
		require.True(t, ok, "rewards should be parseable")
		require.False(t, rewards.Sign() < 0, "rewards must not be negative")
	})
}
