package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/keeper"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

func TestRunSchedulerDisabled(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	params := types.DefaultParams()
	params.Enabled = false
	require.NoError(t, k.SetParams(input.Ctx, params))

	err := k.RunScheduler(input.Ctx)
	require.ErrorIs(t, err, types.ErrSchedulerDisabled)
}

func TestRunSchedulerProcessesValidTx(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Enable scheduler execution.
	params := types.DefaultParams()
	params.SchedulerConfig.Enabled = true
	require.NoError(t, k.SetParams(input.Ctx, params))

	tx := types.Transaction{Sender: "aura1sender", Recipient: "aura1rcpt", Amount: "10", Nonce: 0}
	require.NoError(t, k.AddToMempool(input.Ctx, tx))

	err := k.RunScheduler(input.Ctx)
	require.NoError(t, err)

	// Nonce should increment for the sender after successful validation
	require.Equal(t, uint64(1), k.GetNonce(input.Ctx, tx.Sender))
}

func TestCleanupExpiredTransactionsRemovesStale(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	sender := "aura1stale"
	// Set a high current nonce to trigger stale detection
	k.SetNonce(input.Ctx, sender, 200)

	staleTx := types.Transaction{Sender: sender, Recipient: "aura1rcpt", Amount: "1", Nonce: 1}
	require.NoError(t, k.AddToMempool(input.Ctx, staleTx))

	err := k.CleanupExpiredTransactions(input.Ctx)
	require.NoError(t, err)

	// Mempool should be empty after cleanup
	require.Empty(t, k.GetMempoolTransactions(input.Ctx))
}

func TestUpdateMetricsNoop(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Ensure metrics-enabled default path does not error
	err := k.UpdateMetrics(input.Ctx)
	require.NoError(t, err)

	// Disable metrics and ensure it short-circuits
	params := types.DefaultParams()
	params.MetricsEnabled = false
	require.NoError(t, k.SetParams(input.Ctx, params))
	err = k.UpdateMetrics(input.Ctx)
	require.NoError(t, err)
}
