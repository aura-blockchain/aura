package batch

import (
	"context"
	"errors"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionBuilderMessages(t *testing.T) {
	builder := NewTransactionBuilder().
		AddSend("aura1from", "aura1to", sdk.NewCoins(sdk.NewInt64Coin("uaura", 10))).
		AddDelegate("aura1from", "auravaloper1xyz", sdk.NewInt64Coin("uaura", 20)).
		AddUndelegate("aura1from", "auravaloper1xyz", sdk.NewInt64Coin("uaura", 5)).
		AddVote("aura1from", 42, 1).
		WithMemo("testing")

	require.Equal(t, 4, builder.Size())
	assert.Equal(t, "testing", builder.GetMemo())

	msgs := builder.GetMessages()
	assert.Len(t, msgs, 4)

	builder.Clear()
	assert.Equal(t, 0, builder.Size())
	assert.Empty(t, builder.GetMemo())
}

func TestTransactionBuilderBatchSends(t *testing.T) {
	transfers := []TransferItem{
		{Recipient: "aura1a", Amount: sdk.NewCoins(sdk.NewInt64Coin("uaura", 1))},
		{Recipient: "aura1b", Amount: sdk.NewCoins(sdk.NewInt64Coin("uaura", 2))},
	}

	builder := NewBatchTransfers().AddBatchSends("aura1sender", transfers)
	assert.Equal(t, 2, builder.Size())
}

func TestChunk(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	chunks := Chunk(items, 2)
	assert.Equal(t, [][]int{{1, 2}, {3, 4}, {5}}, chunks)
}

func TestValidateBatchSize(t *testing.T) {
	assert.NoError(t, ValidateBatchSize(5, 10))
	err := ValidateBatchSize(11, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds")
}

func TestEstimateBatchGas(t *testing.T) {
	gas := EstimateBatchGas(3, 1000, 500)
	assert.Equal(t, uint64(3500), gas)
}

func TestBatchQueriesWithRetry(t *testing.T) {
	attempts := 0
	q := func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 2 {
			return "", errors.New("fail once")
		}
		return "ok", nil
	}

	config := Config{
		ParallelQueries:     1,
		RetryOnFailure:      true,
		RetryAttempts:       3,
		DelayBetweenRetries: time.Millisecond,
	}

	res := BatchQueries(context.Background(), []QueryFunc[string]{q}, config)
	require.Len(t, res.Errors, 0)
	assert.Equal(t, "ok", res.Results[0])
	assert.Equal(t, 2, attempts)
}

func TestBatchQueriesMaxRetryFailure(t *testing.T) {
	q := func(ctx context.Context) (string, error) {
		return "", errors.New("always fail")
	}

	config := Config{
		ParallelQueries:     2,
		RetryOnFailure:      true,
		RetryAttempts:       2,
		DelayBetweenRetries: time.Millisecond,
	}

	res := BatchQueries(context.Background(), []QueryFunc[string]{q, q}, config)
	require.Len(t, res.Errors, 2)
	assert.Zero(t, res.Results[0])
	assert.Zero(t, res.Results[1])
}

func TestCreateMultiSigBatch(t *testing.T) {
	items := []MultiSigBatchItem{
		{Messages: []Message{{TypeURL: "msg1"}}, Memo: "a"},
		{Messages: []Message{{TypeURL: "msg2"}, {TypeURL: "msg3"}}, Memo: "b"},
	}

	batch := CreateMultiSigBatch(items)
	assert.Equal(t, 3, batch.TotalMessages)
	assert.Len(t, batch.Transactions, 2)
}
