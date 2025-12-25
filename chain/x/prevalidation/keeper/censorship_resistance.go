// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// CensorshipMonitor monitors for censorship attempts
type CensorshipMonitor struct {
	SuspiciousPatterns map[string]int
	BlockedTxs         map[string]time.Time
	LastCheck          time.Time
}

// DetectCensorship detects potential censorship of transactions
func (k *Keeper) DetectCensorship(ctx sdk.Context, tx types.Transaction) (bool, error) {
	// NOTE: Future enhancement - Add EnableCensorResist and MaxTxAge to Params proto if needed
	// For now, censorship resistance is always enabled with default max age
	enableCensorResist := true
	maxTxAge := int64(3600) // 1 hour in seconds

	if !enableCensorResist {
		return false, nil
	}

	// Check if transaction has been in mempool too long
	txHash := k.GetTransactionHash(tx)
	mempoolAge := k.GetMempoolAge(ctx, txHash)

	maxAge := time.Duration(maxTxAge) * time.Second
	if mempoolAge > maxAge {
		// Transaction has been waiting too long - possible censorship
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"censorship_detected",
			sdk.NewAttribute("tx_hash", txHash),
			sdk.NewAttribute("mempool_age", mempoolAge.String()),
			sdk.NewAttribute("sender", tx.Sender),
		))

		return true, nil
	}

	// Check for systematic exclusion patterns
	if k.DetectSystematicExclusion(ctx, tx.Sender) {
		return true, nil
	}

	return false, nil
}

// GetMempoolAge returns how long a transaction has been in the mempool
func (k *Keeper) GetMempoolAge(ctx sdk.Context, txHash string) time.Duration {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("mempool_time_%s", txHash))

	bz := store.Get(key)
	if bz == nil {
		// Transaction just added, record timestamp
		timestamp := sdk.Uint64ToBigEndian(uint64(ctx.BlockTime().Unix()))
		store.Set(key, timestamp)
		return 0
	}

	timestamp := sdk.BigEndianToUint64(bz)
	addedAt := time.Unix(int64(timestamp), 0)

	return ctx.BlockTime().Sub(addedAt)
}

// DetectSystematicExclusion detects if an address is being systematically excluded
func (k *Keeper) DetectSystematicExclusion(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("exclusion_count_%s", address))

	bz := store.Get(key)
	var count uint64
	if bz != nil {
		count = sdk.BigEndianToUint64(bz)
	}

	// Increment exclusion count
	count++
	store.Set(key, sdk.Uint64ToBigEndian(count))

	// If more than 5 transactions have been delayed, flag as systematic
	if count > 5 {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			"systematic_exclusion_detected",
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("exclusion_count", fmt.Sprintf("%d", count)),
		))
		return true
	}

	return false
}

// PromoteTransaction promotes a censored transaction to priority queue
func (k *Keeper) PromoteTransaction(ctx sdk.Context, txHash string) error {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("priority_tx_%s", txHash))

	// Add to priority queue
	store.Set(key, []byte(txHash))

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"transaction_promoted",
		sdk.NewAttribute("tx_hash", txHash),
		sdk.NewAttribute("reason", "censorship_resistance"),
	))

	return nil
}

// GetPriorityTransactions returns transactions in priority queue
func (k *Keeper) GetPriorityTransactions(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte("priority_tx_")

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	txHashes := make([]string, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		txHashes = append(txHashes, string(iterator.Value()))
	}

	return txHashes
}

// ValidateTransactionInclusion ensures transaction inclusion fairness
func (k *Keeper) ValidateTransactionInclusion(ctx sdk.Context) error {
	// Get all mempool transactions
	mempoolTxs := k.GetMempoolTransactions(ctx)

	for _, tx := range mempoolTxs {
		censored, err := k.DetectCensorship(ctx, tx)
		if err != nil {
			continue
		}

		if censored {
			// Promote transaction
			txHash := k.GetTransactionHash(tx)
			if err := k.PromoteTransaction(ctx, txHash); err != nil {
				return fmt.Errorf("failed to promote transaction %s: %w", txHash, err)
			}
		}
	}

	return nil
}

// ResetExclusionCount resets exclusion count for an address
func (k *Keeper) ResetExclusionCount(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	key := []byte(fmt.Sprintf("exclusion_count_%s", address))
	store.Delete(key)
}
