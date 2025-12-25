// Package batch provides utilities for batching transactions and queries.
package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Config holds configuration for batch operations
type Config struct {
	MaxBatchSize        int
	ParallelQueries     int
	RetryOnFailure      bool
	RetryAttempts       int
	DelayBetweenRetries time.Duration
}

// DefaultConfig returns the default batch configuration
func DefaultConfig() Config {
	return Config{
		MaxBatchSize:        50,
		ParallelQueries:     10,
		RetryOnFailure:      true,
		RetryAttempts:       3,
		DelayBetweenRetries: time.Second,
	}
}

// TransferItem represents a single transfer in a batch
type TransferItem struct {
	Recipient string
	Amount    sdk.Coins
}

// Message represents a generic message for transactions
type Message struct {
	TypeURL string
	Value   interface{}
}

// BatchResult represents the result of a batch operation
type BatchResult[T any] struct {
	Success bool
	Results []T
	Errors  []BatchError
	TxHash  string
}

// BatchError represents an error at a specific index
type BatchError struct {
	Index int
	Error error
}

// QueryResult represents the result of batch queries
type QueryResult[T any] struct {
	Results []T
	Errors  []BatchError
}

// ============================================================================
// Transaction Builder
// ============================================================================

// TransactionBuilder helps construct multi-message transactions
type TransactionBuilder struct {
	messages []Message
	memo     string
}

// NewTransactionBuilder creates a new batch transaction builder
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		messages: make([]Message, 0),
	}
}

// AddMessage adds a message to the batch
func (b *TransactionBuilder) AddMessage(msg Message) *TransactionBuilder {
	b.messages = append(b.messages, msg)
	return b
}

// AddMessages adds multiple messages to the batch
func (b *TransactionBuilder) AddMessages(msgs []Message) *TransactionBuilder {
	b.messages = append(b.messages, msgs...)
	return b
}

// AddSend adds a bank send message
func (b *TransactionBuilder) AddSend(sender, recipient string, amount sdk.Coins) *TransactionBuilder {
	b.messages = append(b.messages, Message{
		TypeURL: "/cosmos.bank.v1beta1.MsgSend",
		Value: map[string]interface{}{
			"from_address": sender,
			"to_address":   recipient,
			"amount":       amount,
		},
	})
	return b
}

// AddBatchSends adds multiple send messages
func (b *TransactionBuilder) AddBatchSends(sender string, transfers []TransferItem) *TransactionBuilder {
	for _, t := range transfers {
		b.AddSend(sender, t.Recipient, t.Amount)
	}
	return b
}

// AddDelegate adds a delegate message
func (b *TransactionBuilder) AddDelegate(delegator, validator string, amount sdk.Coin) *TransactionBuilder {
	b.messages = append(b.messages, Message{
		TypeURL: "/cosmos.staking.v1beta1.MsgDelegate",
		Value: map[string]interface{}{
			"delegator_address": delegator,
			"validator_address": validator,
			"amount":            amount,
		},
	})
	return b
}

// AddUndelegate adds an undelegate message
func (b *TransactionBuilder) AddUndelegate(delegator, validator string, amount sdk.Coin) *TransactionBuilder {
	b.messages = append(b.messages, Message{
		TypeURL: "/cosmos.staking.v1beta1.MsgUndelegate",
		Value: map[string]interface{}{
			"delegator_address": delegator,
			"validator_address": validator,
			"amount":            amount,
		},
	})
	return b
}

// AddVote adds a governance vote message
func (b *TransactionBuilder) AddVote(voter string, proposalID uint64, option int) *TransactionBuilder {
	b.messages = append(b.messages, Message{
		TypeURL: "/cosmos.gov.v1beta1.MsgVote",
		Value: map[string]interface{}{
			"proposal_id": proposalID,
			"voter":       voter,
			"option":      option,
		},
	})
	return b
}

// WithMemo sets the transaction memo
func (b *TransactionBuilder) WithMemo(memo string) *TransactionBuilder {
	b.memo = memo
	return b
}

// GetMessages returns all messages in the batch
func (b *TransactionBuilder) GetMessages() []Message {
	msgs := make([]Message, len(b.messages))
	copy(msgs, b.messages)
	return msgs
}

// GetMemo returns the memo
func (b *TransactionBuilder) GetMemo() string {
	return b.memo
}

// Size returns the number of messages
func (b *TransactionBuilder) Size() int {
	return len(b.messages)
}

// Clear removes all messages
func (b *TransactionBuilder) Clear() *TransactionBuilder {
	b.messages = make([]Message, 0)
	b.memo = ""
	return b
}

// ============================================================================
// Query Batching
// ============================================================================

// QueryFunc is a function that performs a query
type QueryFunc[T any] func(ctx context.Context) (T, error)

// BatchQueries executes multiple queries in parallel with batching
func BatchQueries[T any](ctx context.Context, queries []QueryFunc[T], config Config) QueryResult[T] {
	results := make([]T, len(queries))
	var errors []BatchError
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create semaphore for parallel execution
	sem := make(chan struct{}, config.ParallelQueries)

	for i, query := range queries {
		wg.Add(1)
		go func(index int, q QueryFunc[T]) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := executeWithRetry(ctx, q, config)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errors = append(errors, BatchError{Index: index, Error: err})
			} else {
				results[index] = result
			}
		}(i, query)
	}

	wg.Wait()

	return QueryResult[T]{
		Results: results,
		Errors:  errors,
	}
}

func executeWithRetry[T any](ctx context.Context, fn QueryFunc[T], config Config) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt < config.RetryAttempts; attempt++ {
		result, lastErr = fn(ctx)
		if lastErr == nil {
			return result, nil
		}

		if !config.RetryOnFailure || attempt == config.RetryAttempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(config.DelayBetweenRetries * time.Duration(1<<attempt)):
			// Continue to next attempt
		}
	}

	return result, lastErr
}

// ============================================================================
// Utility Functions
// ============================================================================

// Chunk splits a slice into chunks of specified size
func Chunk[T any](items []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// ValidateBatchSize validates that batch size doesn't exceed maximum
func ValidateBatchSize(items int, maxSize int) error {
	if items > maxSize {
		return fmt.Errorf("batch size %d exceeds maximum allowed size of %d", items, maxSize)
	}
	return nil
}

// EstimateBatchGas calculates estimated gas for batch operations
func EstimateBatchGas(messageCount int, baseGasPerMessage, overheadGas uint64) uint64 {
	return overheadGas + uint64(messageCount)*baseGasPerMessage
}

// ============================================================================
// Multi-Sig Batch Helpers
// ============================================================================

// MultiSigBatchItem represents an item in a multi-sig batch
type MultiSigBatchItem struct {
	Messages []Message
	Memo     string
}

// MultiSigBatch represents the result of creating a multi-sig batch
type MultiSigBatch struct {
	Transactions  []MultiSigBatchItem
	TotalMessages int
}

// CreateMultiSigBatch creates multiple transactions for multi-sig signing
func CreateMultiSigBatch(items []MultiSigBatchItem) MultiSigBatch {
	totalMessages := 0
	for _, item := range items {
		totalMessages += len(item.Messages)
	}

	return MultiSigBatch{
		Transactions:  items,
		TotalMessages: totalMessages,
	}
}

// NewBatchTransfers creates a new batch transaction builder
func NewBatchTransfers() *TransactionBuilder {
	return NewTransactionBuilder()
}
