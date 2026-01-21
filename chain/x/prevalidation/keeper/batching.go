// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"sort"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BatchConfig defines batching configuration
type BatchConfig struct {
	MaxBatchSize     int
	MaxWaitTime      time.Duration
	EnableBatching   bool
	OptimizeOrdering bool
}

// TransactionBatch represents a batch of transactions
type TransactionBatch struct {
	ID           string
	Transactions []types.Transaction
	CreatedAt    time.Time
	Status       string
	TotalGas     uint64
}

// BatchValidateTransactions validates multiple transactions in a batch
func (k *Keeper) BatchValidateTransactions(ctx sdk.Context, transactions []types.Transaction) ([]types.ValidationResult, error) {
	// NOTE: Future enhancement - Add EnableBatching and BatchSize to Params proto if needed
	// For now, batching is always enabled with default batch size
	enableBatching := true
	batchSize := 100

	if !enableBatching {
		return k.validateSequentially(ctx, transactions)
	}

	// Process in batches
	if batchSize == 0 {
		batchSize = 100
	}

	results := make([]types.ValidationResult, len(transactions))

	for i := 0; i < len(transactions); i += batchSize {
		end := i + batchSize
		if end > len(transactions) {
			end = len(transactions)
		}

		batch := transactions[i:end]
		batchResults := k.validateBatch(ctx, batch)

		copy(results[i:end], batchResults)
	}

	return results, nil
}

// validateBatch validates a single batch of transactions
func (k *Keeper) validateBatch(ctx sdk.Context, batch []types.Transaction) []types.ValidationResult {
	results := make([]types.ValidationResult, len(batch))

	for i, tx := range batch {
		valid, err := k.ValidateTransaction(ctx, tx)

		result := types.ValidationResult{
			TxHash:      k.GetTransactionHash(tx),
			Valid:       valid,
			GasEstimate: k.EstimateGas(ctx, tx),
		}

		if err != nil {
			result.Error = err.Error()
		}

		results[i] = result
	}

	return results
}

// validateSequentially validates transactions one by one
func (k *Keeper) validateSequentially(ctx sdk.Context, transactions []types.Transaction) ([]types.ValidationResult, error) {
	results := make([]types.ValidationResult, len(transactions))

	for i, tx := range transactions {
		valid, err := k.ValidateTransaction(ctx, tx)

		result := types.ValidationResult{
			TxHash:      k.GetTransactionHash(tx),
			Valid:       valid,
			GasEstimate: k.EstimateGas(ctx, tx),
		}

		if err != nil {
			result.Error = err.Error()
		}

		results[i] = result
	}

	return results, nil
}

// OptimizeTransactionOrder optimizes the order of transactions for better execution.
// CONSENSUS-CRITICAL: This function uses deterministic iteration order to ensure
// all validators produce the same transaction ordering.
func (k *Keeper) OptimizeTransactionOrder(ctx sdk.Context, transactions []types.Transaction) []types.Transaction {
	// Sort by nonce for same sender
	optimized := make([]types.Transaction, len(transactions))
	copy(optimized, transactions)

	// Group by sender
	senderGroups := make(map[string][]types.Transaction)
	for _, tx := range optimized {
		senderGroups[tx.Sender] = append(senderGroups[tx.Sender], tx)
	}

	// CONSENSUS-CRITICAL: Sort sender addresses for deterministic iteration order
	// Without sorting, map iteration order is random and can cause different nodes
	// to produce different transaction orderings, breaking consensus.
	sortedSenders := make([]string, 0, len(senderGroups))
	for sender := range senderGroups {
		sortedSenders = append(sortedSenders, sender)
	}
	sort.Strings(sortedSenders)

	// Sort each sender's transactions by nonce, in deterministic sender order
	result := []types.Transaction{}
	for _, sender := range sortedSenders {
		txs := senderGroups[sender]
		// Simple bubble sort by nonce
		for i := 0; i < len(txs); i++ {
			for j := i + 1; j < len(txs); j++ {
				if txs[i].Nonce > txs[j].Nonce {
					txs[i], txs[j] = txs[j], txs[i]
				}
			}
		}
		result = append(result, txs...)
	}

	return result
}

// CreateBatch creates a new transaction batch
func (k *Keeper) CreateBatch(ctx sdk.Context, transactions []types.Transaction) (string, error) {
	batchID := fmt.Sprintf("batch_%d", ctx.BlockHeight())

	batch := TransactionBatch{
		ID:           batchID,
		Transactions: transactions,
		CreatedAt:    ctx.BlockTime(),
		Status:       "pending",
	}

	// Calculate total gas
	for _, tx := range transactions {
		batch.TotalGas += k.EstimateGas(ctx, tx)
	}

	// Store batch
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("batch_%s", batchID))

	// Simple serialization
	store.Set(key, []byte(batchID))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"batch_created",
		sdk.NewAttribute("batch_id", batchID),
		sdk.NewAttribute("tx_count", fmt.Sprintf("%d", len(transactions))),
		sdk.NewAttribute("total_gas", fmt.Sprintf("%d", batch.TotalGas)),
	))

	return batchID, nil
}

// ProcessBatch processes a batch of transactions
func (k *Keeper) ProcessBatch(ctx sdk.Context, batchID string) error {
	// Retrieve batch
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("batch_%s", batchID))

	if !store.Has(key) {
		return fmt.Errorf("batch not found: %s", batchID)
	}

	// Mark as processed
	store.Delete(key)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"batch_processed",
		sdk.NewAttribute("batch_id", batchID),
	))

	return nil
}
