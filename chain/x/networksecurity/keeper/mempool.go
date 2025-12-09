package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

// MempoolManager manages mempool security
type MempoolManager struct {
	keeper *Keeper
}

// NewMempoolManager creates a new mempool manager
func NewMempoolManager(k *Keeper) *MempoolManager {
	return &MempoolManager{
		keeper: k,
	}
}

// ValidateTransaction validates a transaction for mempool admission
func (k Keeper) ValidateTransaction(ctx sdk.Context, tx sdk.Tx, txBytes []byte, sender string) error {
	params, _ := k.GetParams(ctx)
	stats := k.GetMempoolStats(ctx)

	// 1. Check mempool size limits
	if stats.TxCount >= params.Mempool.MaxSize {
		// Mempool is full, check if we should evict
		if !params.Mempool.EnablePriorityFees {
			return types.ErrMempoolFull
		}
		// Will check priority fee later to see if we can evict
	}

	// 2. Check mempool byte size
	if stats.SizeBytes+uint64(len(txBytes)) > params.Mempool.MaxBytes {
		return types.ErrMempoolFull
	}

	// 3. Extract and validate priority fee
	var priorityFee math.Int
	if feeTx, ok := tx.(sdk.FeeTx); ok {
		fees := feeTx.GetFee()
		if len(fees) > 0 {
			priorityFee = fees[0].Amount
		} else {
			priorityFee = math.ZeroInt()
		}
	} else {
		priorityFee = math.ZeroInt()
	}

	// 4. Check minimum priority fee
	if params.Mempool.EnablePriorityFees {
		if priorityFee.LT(params.Mempool.MinPriorityFee) {
			return types.ErrPriorityFeeTooLow
		}
	}

	// 5. Check per-account transaction limits
	if sender != "" {
		txCount := k.GetAccountMempoolTxCount(ctx, sender)
		if txCount >= params.Mempool.MaxTxsPerAccount {
			return types.ErrMaxTxsPerAccountReached
		}
	}

	// 6. If mempool is full and priority fees are enabled, check if we can evict
	if stats.TxCount >= params.Mempool.MaxSize && params.Mempool.EnablePriorityFees {
		// Check if this tx has higher priority than lowest in mempool
		if priorityFee.LTE(stats.AvgPriorityFee) {
			return types.ErrPriorityFeeTooLow
		}
		// Evict lowest priority transaction
		k.EvictLowestPriorityTx(ctx)
	}

	return nil
}

// AddToMempool adds a transaction to mempool tracking
func (k Keeper) AddToMempool(ctx sdk.Context, tx sdk.Tx, txBytes []byte, sender string) error {
	stats := k.GetMempoolStats(ctx)

	// Extract fee
	var priorityFee math.Int
	if feeTx, ok := tx.(sdk.FeeTx); ok {
		fees := feeTx.GetFee()
		if len(fees) > 0 {
			priorityFee = fees[0].Amount
		} else {
			priorityFee = math.ZeroInt()
		}
	} else {
		priorityFee = math.ZeroInt()
	}

	// Update stats
	stats.TxCount++
	stats.SizeBytes += uint64(len(txBytes))

	// Update average priority fee (running average)
	if stats.TxCount == 1 {
		stats.AvgPriorityFee = priorityFee
	} else {
		// Calculate new average: avg = (avg * (n-1) + new) / n
		totalFee := stats.AvgPriorityFee.Mul(math.NewInt(int64(stats.TxCount - 1)))
		totalFee = totalFee.Add(priorityFee)
		stats.AvgPriorityFee = totalFee.Quo(math.NewInt(int64(stats.TxCount)))
	}

	k.SetMempoolStats(ctx, stats)

	// Increment account tx count
	if sender != "" {
		count := k.GetAccountMempoolTxCount(ctx, sender)
		k.SetAccountMempoolTxCount(ctx, sender, count+1)
	}

	return nil
}

// RemoveFromMempool removes a transaction from mempool tracking
func (k Keeper) RemoveFromMempool(ctx sdk.Context, txBytes []byte, sender string) error {
	stats := k.GetMempoolStats(ctx)

	if stats.TxCount > 0 {
		stats.TxCount--
	}

	if stats.SizeBytes >= uint64(len(txBytes)) {
		stats.SizeBytes -= uint64(len(txBytes))
	}

	k.SetMempoolStats(ctx, stats)

	// Decrement account tx count
	if sender != "" {
		count := k.GetAccountMempoolTxCount(ctx, sender)
		if count > 0 {
			k.SetAccountMempoolTxCount(ctx, sender, count-1)
		}
	}

	return nil
}

// EvictLowestPriorityTx evicts the lowest priority transaction from mempool
func (k Keeper) EvictLowestPriorityTx(ctx sdk.Context) {
	stats := k.GetMempoolStats(ctx)
	stats.EvictedCount++
	k.SetMempoolStats(ctx, stats)

	// Actual eviction would be handled by the mempool implementation
	// This just updates our statistics
	k.logger.Info("Evicting lowest priority transaction from mempool")
}

// RejectTransaction records a rejected transaction (spam)
func (k Keeper) RejectTransaction(ctx sdk.Context, reason string) {
	stats := k.GetMempoolStats(ctx)
	stats.RejectedCount++
	k.SetMempoolStats(ctx, stats)

	k.logger.Debug(fmt.Sprintf("Rejected transaction: %s", reason))
}

// GetAccountMempoolTxCount gets transaction count for an account in mempool
func (k Keeper) GetAccountMempoolTxCount(ctx sdk.Context, address string) uint32 {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte("mempool_account_"), []byte(address)...)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return 0
	}

	if len(bz) >= 4 {
		return uint32(bz[0]) | uint32(bz[1])<<8 | uint32(bz[2])<<16 | uint32(bz[3])<<24
	}
	return 0
}

// SetAccountMempoolTxCount sets transaction count for an account in mempool
func (k Keeper) SetAccountMempoolTxCount(ctx sdk.Context, address string, count uint32) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append([]byte("mempool_account_"), []byte(address)...)
	bz := make([]byte, 4)
	bz[0] = byte(count)
	bz[1] = byte(count >> 8)
	bz[2] = byte(count >> 16)
	bz[3] = byte(count >> 24)
	return store.Set(key, bz)
}

// CheckMempoolHealth checks mempool health and triggers cleanup if needed
func (k Keeper) CheckMempoolHealth(ctx sdk.Context) {
	params, _ := k.GetParams(ctx)
	stats := k.GetMempoolStats(ctx)

	// Check if mempool is near capacity
	usagePercent := float64(stats.TxCount) / float64(params.Mempool.MaxSize) * 100

	if usagePercent > 80.0 {
		k.logger.Warn(fmt.Sprintf("Mempool is %.1f%% full (%d/%d txs)",
			usagePercent, stats.TxCount, params.Mempool.MaxSize))
	}

	// Check byte usage
	byteUsagePercent := float64(stats.SizeBytes) / float64(params.Mempool.MaxBytes) * 100

	if byteUsagePercent > 80.0 {
		k.logger.Warn(fmt.Sprintf("Mempool byte usage is %.1f%% (%d/%d bytes)",
			byteUsagePercent, stats.SizeBytes, params.Mempool.MaxBytes))
	}
}

// CalculatePriorityScore calculates priority score for transaction ordering
func (k Keeper) CalculatePriorityScore(ctx sdk.Context, tx sdk.Tx, txSize int) int64 {
	var fee math.Int
	if feeTx, ok := tx.(sdk.FeeTx); ok {
		fees := feeTx.GetFee()
		if len(fees) > 0 {
			fee = fees[0].Amount
		} else {
			fee = math.ZeroInt()
		}
	} else {
		fee = math.ZeroInt()
	}

	// Priority = fee / size (fee per byte)
	if txSize == 0 {
		return 0
	}

	priority := fee.Quo(math.NewInt(int64(txSize)))
	return priority.Int64()
}

// AntiSpamCheck performs anti-spam checks on transaction
func (k Keeper) AntiSpamCheck(ctx sdk.Context, tx sdk.Tx, sender string) error {
	params, _ := k.GetParams(ctx)

	// 1. Check if sender is banned
	if k.IsBanned(ctx, sender) {
		k.RejectTransaction(ctx, "sender is banned")
		return types.ErrPeerBanned
	}

	// 2. Check sender reputation
	if reputation, found := k.GetReputation(ctx, sender); found {
		if reputation.Score < params.Reputation.MinScoreToConnect {
			k.RejectTransaction(ctx, "sender has low reputation")
			return types.ErrInvalidPeerReputation
		}
	}

	// 3. Check rate of transactions from this sender
	txCount := k.GetAccountMempoolTxCount(ctx, sender)
	if txCount > params.Mempool.MaxTxsPerAccount {
		k.RejectTransaction(ctx, "too many transactions from sender")

		// Penalize reputation for spam
		k.PenalizeReputation(ctx, sender, params.Reputation.MisbehaviorPenalty/2)

		return types.ErrMaxTxsPerAccountReached
	}

	return nil
}

// MempoolSpamProtection implements comprehensive mempool spam protection
func (k Keeper) MempoolSpamProtection(ctx sdk.Context, tx sdk.Tx, txBytes []byte, sender string) error {
	// 1. Anti-spam checks
	if err := k.AntiSpamCheck(ctx, tx, sender); err != nil {
		return err
	}

	// 2. Validate transaction for mempool admission
	if err := k.ValidateTransaction(ctx, tx, txBytes, sender); err != nil {
		k.RejectTransaction(ctx, err.Error())
		return err
	}

	// 3. Add to mempool tracking
	if err := k.AddToMempool(ctx, tx, txBytes, sender); err != nil {
		return err
	}

	return nil
}

// CleanupMempool performs mempool cleanup and maintenance
func (k Keeper) CleanupMempool(ctx sdk.Context) {
	params, _ := k.GetParams(ctx)
	stats := k.GetMempoolStats(ctx)

	// If mempool is too full, evict based on policy
	for stats.TxCount > params.Mempool.MaxSize {
		switch params.Mempool.EvictionPolicy {
		case "lowest_fee":
			k.EvictLowestPriorityTx(ctx)
		case "oldest":
			// Evict oldest transaction
			k.EvictOldestTx(ctx)
		case "random":
			// Evict random transaction
			k.EvictRandomTx(ctx)
		}

		stats.TxCount--
		k.SetMempoolStats(ctx, stats)
	}

	// Reset statistics periodically
	if ctx.BlockHeight()%1000 == 0 {
		stats.EvictedCount = 0
		stats.RejectedCount = 0
		k.SetMempoolStats(ctx, stats)
	}
}

// EvictOldestTx evicts the oldest transaction
func (k Keeper) EvictOldestTx(ctx sdk.Context) {
	stats := k.GetMempoolStats(ctx)
	stats.EvictedCount++
	k.SetMempoolStats(ctx, stats)
	k.logger.Debug("Evicting oldest transaction from mempool")
}

// EvictRandomTx evicts a random transaction
func (k Keeper) EvictRandomTx(ctx sdk.Context) {
	stats := k.GetMempoolStats(ctx)
	stats.EvictedCount++
	k.SetMempoolStats(ctx, stats)
	k.logger.Debug("Evicting random transaction from mempool")
}
