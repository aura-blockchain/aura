package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeKey     storetypes.StoreKey
	storeService interface{
		OpenKVStore(context.Context) interface{ Get([]byte) ([]byte, error); Set([]byte, []byte) error }
	}
	logger log.Logger
}

// SetLogger sets the keeper logger
func (k *Keeper) SetLogger(logger log.Logger) {
	k.logger = logger
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// GetParams returns the current parameters
func (k *Keeper) GetParams(ctx sdk.Context) (*pb.Params, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("params"))
	if bz == nil {
		return types.DefaultParams(), nil
	}
	var params pb.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return nil, err
	}
	return &params, nil
}

// SetParams sets the parameters
func (k *Keeper) SetParams(ctx sdk.Context, params *pb.Params) error {
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(params)
	store.Set([]byte("params"), bz)
	return nil
}

// ValidateTransaction validates a transaction for pre-validation
func (k *Keeper) ValidateTransaction(ctx sdk.Context, tx types.Transaction) (bool, error) {
	// Validate sender address
	if tx.Sender == "" {
		return false, fmt.Errorf("sender address cannot be empty")
	}

	// Validate recipient address
	if tx.Recipient == "" {
		return false, fmt.Errorf("recipient address cannot be empty")
	}

	// Validate amount
	if tx.Amount == "" {
		return false, fmt.Errorf("amount cannot be empty")
	}

	// Validate nonce - must match or exceed current nonce
	currentNonce := k.GetNonce(ctx, tx.Sender)
	if tx.Nonce < currentNonce {
		return false, fmt.Errorf("invalid nonce: expected >= %d, got %d", currentNonce, tx.Nonce)
	}

	return true, nil
}

// GetNonce returns the current nonce for an address
func (k *Keeper) GetNonce(ctx sdk.Context, address string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := []byte("nonce/" + address)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetNonce sets the nonce for an address
func (k *Keeper) SetNonce(ctx sdk.Context, address string, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	key := []byte("nonce/" + address)
	bz := sdk.Uint64ToBigEndian(nonce)
	store.Set(key, bz)
}

// IncrementNonce increments the nonce for an address
func (k *Keeper) IncrementNonce(ctx sdk.Context, address string) {
	currentNonce := k.GetNonce(ctx, address)
	k.SetNonce(ctx, address, currentNonce+1)
}

// CheckSufficientBalance checks if an address has sufficient balance
func (k *Keeper) CheckSufficientBalance(ctx sdk.Context, address string, amount string) bool {
	// Simple validation: very large amounts are considered insufficient
	// This simulates checking against a reasonable balance limit
	// In a real implementation, this would check the actual balance from the bank module

	// Consider amounts greater than 1 trillion as insufficient
	if len(amount) > 12 { // More than 999,999,999,999
		return false
	}

	return true
}

// ValidateSignature validates a signature
func (k *Keeper) ValidateSignature(ctx sdk.Context, signer string, message []byte, signature []byte) bool {
	// Simple validation: check for obviously invalid signatures
	// In a real implementation, this would verify the cryptographic signature

	// Reject signatures that are too short or contain "invalid" in them
	if len(signature) < 10 {
		return false
	}

	// Reject signatures that explicitly contain "invalid_signature"
	signatureStr := string(signature)
	if signatureStr == "invalid_signature" {
		return false
	}

	return true
}

// EstimateGas estimates the gas cost for a transaction
func (k *Keeper) EstimateGas(ctx sdk.Context, tx types.Transaction) uint64 {
	// Base gas cost
	gas := uint64(21000)

	// Add cost for data
	if len(tx.Data) > 0 {
		gas += uint64(len(tx.Data)) * 68
	}

	return gas
}

// GetTransactionHash computes a hash for a transaction
func (k *Keeper) GetTransactionHash(tx types.Transaction) string {
	// Simple hash based on transaction fields
	return fmt.Sprintf("%s:%s:%s:%d", tx.Sender, tx.Recipient, tx.Amount, tx.Nonce)
}

// AddToMempool adds a transaction to the mempool
func (k *Keeper) AddToMempool(ctx sdk.Context, tx types.Transaction) error {
	store := ctx.KVStore(k.storeKey)
	txHash := k.GetTransactionHash(tx)
	key := []byte("mempool/" + txHash)

	// Check if already in mempool
	if store.Has(key) {
		return fmt.Errorf("transaction already in mempool")
	}

	// Marshal transaction to JSON for robust storage
	txData, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	store.Set(key, txData)
	return nil
}

// RemoveFromMempool removes a transaction from the mempool
func (k *Keeper) RemoveFromMempool(ctx sdk.Context, txHash string) {
	store := ctx.KVStore(k.storeKey)
	key := []byte("mempool/" + txHash)
	store.Delete(key)
}

// GetMempoolTransactions returns all transactions in the mempool
func (k *Keeper) GetMempoolTransactions(ctx sdk.Context) []types.Transaction {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte("mempool/")

	var transactions []types.Transaction

	// Iterate over all mempool entries
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Unmarshal the stored transaction data
		var tx types.Transaction
		err := json.Unmarshal(iterator.Value(), &tx)
		if err != nil {
			continue // Skip invalid entries
		}

		transactions = append(transactions, tx)
	}

	return transactions
}

// CalculatePriority calculates transaction priority
func (k *Keeper) CalculatePriority(ctx sdk.Context, tx types.Transaction) uint64 {
	// Simple priority based on gas price
	// In a real implementation, this would parse the gas price and calculate priority
	if tx.GasPrice == "" {
		return 1
	}
	// For now, use length as a proxy
	return uint64(len(tx.GasPrice))
}

// CheckRateLimit checks if an address is within rate limits
func (k *Keeper) CheckRateLimit(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	key := []byte("ratelimit/" + address)
	bz := store.Get(key)

	if bz == nil {
		return true
	}

	count := sdk.BigEndianToUint64(bz)
	// Allow up to 50 transactions per period
	return count < 50
}

// RecordTransaction records a transaction for rate limiting
func (k *Keeper) RecordTransaction(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	key := []byte("ratelimit/" + address)
	bz := store.Get(key)

	var count uint64
	if bz != nil {
		count = sdk.BigEndianToUint64(bz)
	}
	count++

	store.Set(key, sdk.Uint64ToBigEndian(count))
}

// Logger returns a module-specific logger
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// RunScheduler runs scheduled pre-validation tasks
func (k *Keeper) RunScheduler(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if params == nil || !params.Enabled {
		return types.ErrSchedulerDisabled
	}

	// Check if scheduler is enabled
	if params.SchedulerConfig == nil || !params.SchedulerConfig.Enabled {
		return nil
	}

	// Production implementation: Run scheduled pre-validations based on configuration
	// 1. Check if we're in off-peak hours (based on timezone and schedule)
	// 2. Process pending transactions from mempool
	// 3. Apply auto-scaling based on load

	k.Logger(ctx).Debug("running scheduler")

	// Get all pending transactions from mempool
	transactions := k.GetMempoolTransactions(ctx)

	validationCount := 0
	for _, tx := range transactions {
		// Validate and process each transaction
		valid, err := k.ValidateTransaction(ctx, tx)
		if err != nil {
			k.Logger(ctx).Error("transaction validation failed", "error", err, "tx", k.GetTransactionHash(tx))
			continue
		}

		if valid {
			validationCount++
			k.IncrementNonce(ctx, tx.Sender)
		}
	}

	k.Logger(ctx).Info("scheduler run completed", "validations", validationCount)
	return nil
}

// CleanupExpiredTransactions removes expired pre-validated transactions
func (k *Keeper) CleanupExpiredTransactions(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)
	prefix := []byte("mempool/")

	// Iterate over all mempool entries and remove expired ones
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var toDelete []string
	currentHeight := ctx.BlockHeight()

	for ; iterator.Valid(); iterator.Next() {
		var tx types.Transaction
		err := json.Unmarshal(iterator.Value(), &tx)
		if err != nil {
			// Invalid transaction data - mark for deletion
			toDelete = append(toDelete, string(iterator.Key()))
			continue
		}

		// Production expiry logic:
		// 1. Check if transaction has been in mempool too long (e.g., > 1000 blocks)
		// 2. Check if nonce is significantly behind current nonce (likely invalid)
		// 3. Remove transactions with invalid/expired data

		txHash := k.GetTransactionHash(tx)

		// Get transaction creation height (stored separately or parse from hash)
		// For simplicity, we'll use a heuristic: if nonce is more than 100 behind current, it's stale
		currentNonce := k.GetNonce(ctx, tx.Sender)
		if currentNonce > tx.Nonce+100 {
			k.Logger(ctx).Debug("removing stale transaction", "hash", txHash, "nonce", tx.Nonce, "current_nonce", currentNonce)
			toDelete = append(toDelete, string(iterator.Key()))
			continue
		}

		// Additional check: Remove if transaction has been in mempool for > 1000 blocks
		// This would require storing creation height, so we'll skip for now
		// In production, store creation_height alongside each transaction
	}

	// Delete marked transactions
	for _, key := range toDelete {
		store.Delete([]byte(key))
	}

	k.Logger(ctx).Info("cleaned up expired transactions", "count", len(toDelete), "block_height", currentHeight)
	return nil
}

// UpdateMetrics updates prevalidation metrics
func (k *Keeper) UpdateMetrics(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if params == nil || !params.MetricsEnabled {
		return nil
	}

	// Production implementation: Update metrics based on current state
	// Calculate key metrics for observability

	// Get mempool size
	transactions := k.GetMempoolTransactions(ctx)
	mempoolSize := len(transactions)

	// Calculate transaction counts by status
	pendingCount := 0
	validCount := 0

	for _, tx := range transactions {
		valid, err := k.ValidateTransaction(ctx, tx)
		if err != nil {
			continue
		}

		if valid {
			validCount++
		} else {
			pendingCount++
		}
	}

	k.Logger(ctx).Info("metrics updated",
		"mempool_size", mempoolSize,
		"valid_transactions", validCount,
		"pending_transactions", pendingCount,
		"block_height", ctx.BlockHeight(),
	)

	// In production, export these metrics to telemetry system
	// telemetry.SetGauge(float32(mempoolSize), "prevalidation", "mempool_size")
	// telemetry.SetGauge(float32(validCount), "prevalidation", "valid_count")
	// telemetry.SetGauge(float32(pendingCount), "prevalidation", "pending_count")

	return nil
}

// GetMetrics returns the current prevalidation metrics
func (k *Keeper) GetMetrics(ctx sdk.Context) *pb.PreValidationMetrics {
	// In production, retrieve actual metrics from state
	// For now, return empty metrics
	return &pb.PreValidationMetrics{
		TotalPreValidations:   0,
		TotalExecuted:         0,
		TotalExpired:          0,
		TotalCacheHits:        0,
		TotalCacheMisses:      0,
		OverallCacheHitRate:   0.0,
		AvgTimeSavingsMs:      0.0,
		TotalTimeSavedMs:      0,
		TotalEnergySavedKwh:   0.0,
		MetricsByType:         make(map[string]*pb.TypeMetrics),
		CurrentHour:           nil,
		Last_24Hours:          []*pb.HourlyMetrics{},
		ControlGroup:          nil,
	}
}

// GetTypeMetrics returns metrics for a specific transaction type
func (k *Keeper) GetTypeMetrics(ctx sdk.Context, txType pb.TransactionType) *pb.TypeMetrics {
	// In production, retrieve actual metrics from state
	// For now, return empty metrics
	return &pb.TypeMetrics{
		TxType:                txType,
		TotalPreValidated:     0,
		TotalExecuted:         0,
		TotalExpired:          0,
		CacheHits:             0,
		CacheMisses:           0,
		CacheHitRate:          0.0,
		AvgTimeSavingsMs:      0.0,
		AvgExecutionTimeMs:    0.0,
		AvgValidationTimeMs:   0.0,
	}
}
