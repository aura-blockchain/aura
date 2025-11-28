package keeper_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/economicsecurity/keeper"
	"github.com/aequitas/aura/chain/x/economicsecurity/params"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"github.com/stretchr/testify/require"
)

// Test Transaction Batching
func TestBatchTransaction(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))
	k.SetCurrentTime(1000000)

	txID, err := k.BatchTransaction("sender1", "recipient1", "100", 5)
	require.NoError(t, err)
	require.NotEmpty(t, txID)

	// Check pending batch
	count, total, status := k.GetPendingBatch()
	require.Equal(t, uint64(1), count)
	require.NotEmpty(t, total)
}

// Test Gas Price Prediction
func TestPredictGasPrice(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))

	// Record some utilization data
	k.RecordBlockUtilization(5000)
	k.RecordBlockUtilization(6000)
	k.RecordBlockUtilization(5500)

	price, confidence, err := k.PredictGasPrice(10)
	require.NoError(t, err)
	require.NotEmpty(t, price)
	require.Greater(t, confidence, uint64(0))
}

// Test Attack Detection
func TestDetectEconomicAttacks(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))
	k.SetCurrentTime(1000000)
	k.SetCurrentHeight(1000)

	alerts := k.DetectEconomicAttacks()
	require.NotNil(t, alerts)
}

// Test Circuit Breakers
func TestCheckCircuitBreakers(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))
	k.SetCurrentTime(1000000)

	events, err := k.CheckCircuitBreakers()
	require.NoError(t, err)
	require.NotNil(t, events)
}

// Test Incentive Analysis
func TestAnalyzeEconomicIncentives(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))

	result, err := k.AnalyzeEconomicIncentives("1000000", "500000", 1000, 100)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.ValidatorRewards)
	require.GreaterOrEqual(t, result.IncentiveEfficiency, float64(0))
}

// Test Tokenomics Simulation
func TestSimulateTokenomics(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))

	params := keeper.SimulationParameters{
		DurationBlocks:       1000,
		InitialSupply:        "1000000",
		InflationRate:        500,
		BurnRate:             100,
		StakingRatio:         5000,
		ActiveUsers:          100,
		TransactionsPerBlock: 50,
		AverageGasPrice:      "1000",
	}

	result, err := k.SimulateTokenomics(params)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.FinalSupply)
	require.Greater(t, len(result.Projections), 0)
}

// Test MEV Auction
func TestMEVAuction(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))
	k.SetCurrentTime(1000000)

	// Create auction
	auctionID, err := k.CreateMEVAuction(1000)
	require.NoError(t, err)
	require.NotEmpty(t, auctionID)

	// Place bids
	bidID1, err := k.PlaceMEVBid(auctionID, "bidder1", "1000", 5)
	require.NoError(t, err)
	require.NotEmpty(t, bidID1)

	bidID2, err := k.PlaceMEVBid(auctionID, "bidder2", "1500", 7)
	require.NoError(t, err)
	require.NotEmpty(t, bidID2)

	// Close auction
	winningBid, amount, err := k.CloseMEVAuction(auctionID)
	require.NoError(t, err)
	require.NotNil(t, winningBid)
	require.NotEmpty(t, amount)
}

// Test Process Batch
func TestProcessBatch(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))
	k.SetCurrentTime(1000000)

	// Add transactions to batch
	for i := 0; i < 5; i++ {
		_, err := k.BatchTransaction("sender", "recipient", "100", 5)
		require.NoError(t, err)
	}

	// Process batch
	count, batchID, err := k.ProcessBatch()
	require.NoError(t, err)
	require.Greater(t, count, uint64(0))
	require.NotEmpty(t, batchID)
}

// Test Recommended Gas Price
func TestGetRecommendedGasPrice(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))

	testCases := []struct {
		priority string
		valid    bool
	}{
		{"low", true},
		{"medium", true},
		{"high", true},
		{"urgent", true},
		{"invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.priority, func(t *testing.T) {
			price, err := k.GetRecommendedGasPrice(tc.priority)
			if tc.valid {
				require.NoError(t, err)
				require.NotEmpty(t, price)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// Test Attack Statistics
func TestGetAttackStatistics(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))

	total, mitigated, critical, warning := k.GetAttackStatistics()
	require.GreaterOrEqual(t, total, uint64(0))
	require.GreaterOrEqual(t, mitigated, uint64(0))
	require.GreaterOrEqual(t, critical, uint64(0))
	require.GreaterOrEqual(t, warning, uint64(0))
}

// Test Circuit Breaker Activation
func TestActivateCircuitBreaker(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))
	k.SetCurrentTime(1000000)

	err := k.ActivateCircuitBreaker(types.CircuitBreakerTypePriceVolatility, "Test activation")
	require.NoError(t, err)

	active := k.GetActiveCircuitBreakers()
	require.Greater(t, len(active), 0)
}

// Test Tokenomics Optimization
func TestOptimizeTokenomicsParameters(t *testing.T) {
	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()))

	recommendations := k.OptimizeTokenomicsParameters(0.05, 1)
	require.NotNil(t, recommendations)
	require.Greater(t, recommendations["optimal_inflation_rate"], uint64(0))
	require.Greater(t, recommendations["optimal_burn_rate"], uint64(0))
}
