package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// TRANSACTION BATCHING (Feature 4)
// ============================

// Transaction represents a pending transaction for batching
type Transaction struct {
	ID        string `json:"id"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    string `json:"amount"`
	Priority  uint64 `json:"priority"`
	Timestamp int64  `json:"timestamp"`
	GasPrice  string `json:"gas_price"`
}

// TransactionBatch represents a batch of transactions
type TransactionBatch struct {
	BatchID      string        `json:"batch_id"`
	Transactions []Transaction `json:"transactions"`
	TotalAmount  string        `json:"total_amount"`
	CreatedAt    int64         `json:"created_at"`
	Status       string        `json:"status"`
}

// BatchRecord represents historical batch processing data
type BatchRecord struct {
	BatchID          string  `json:"batch_id"`
	TransactionCount uint64  `json:"transaction_count"`
	TotalAmount      string  `json:"total_amount"`
	ProcessedAt      int64   `json:"processed_at"`
	GasSaved         uint64  `json:"gas_saved"`
	AverageGasPrice  uint64  `json:"average_gas_price"`
	CompressionRatio float32 `json:"compression_ratio"`
}

// BatchStatistics represents aggregate batching statistics
type BatchStatistics struct {
	TotalBatchesProcessed      uint64 `json:"total_batches_processed"`
	TotalTransactionsBatched   uint64 `json:"total_transactions_batched"`
	TotalGasSaved              string `json:"total_gas_saved"`
	AverageCompressionRatio    float32 `json:"average_compression_ratio"`
}

// KV Store keys for batch processing
var (
	PendingBatchKey       = []byte{0x20}
	BatchHistoryPrefix    = []byte{0x21}
	BatchStatisticsKey    = []byte{0x22}
	BatchEnabledKey       = []byte{0x23}
	BatchConfigPrefix     = []byte{0x24}
)

// BatchTransaction adds a transaction to the pending batch
func (k *Keeper) BatchTransaction(ctx context.Context, sender, recipient, amount string, priority uint64) (string, error) {
	// Check if batching is enabled
	enabled, err := k.IsBatchingEnabled(ctx)
	if err != nil {
		return "", err
	}
	if !enabled {
		return "", types.ErrBatchingDisabled
	}

	// Get current time and height from context
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "", err
	}

	// Create transaction
	txID := k.generateTxID(sender, recipient, amount, currentTime)
	tx := Transaction{
		ID:        txID,
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		Priority:  priority,
		Timestamp: currentTime,
		GasPrice:  k.CalculateDynamicFee(),
	}

	// Get or create pending batch
	batch, err := k.GetPendingBatchData(ctx)
	if err != nil {
		return "", err
	}

	if batch == nil {
		currentHeight, err := k.GetCurrentHeight(ctx)
		if err != nil {
			return "", err
		}

		batch = &TransactionBatch{
			BatchID:      k.generateBatchID(currentHeight, currentTime),
			Transactions: []Transaction{},
			TotalAmount:  "0",
			CreatedAt:    currentTime,
			Status:       "pending",
		}
	}

	// Add transaction to batch
	batch.Transactions = append(batch.Transactions, tx)

	// Update total amount
	amt := new(big.Int)
	if _, ok := amt.SetString(amount, 10); !ok {
		return "", types.ErrInvalidAmount
	}
	total := new(big.Int)
	if _, ok := total.SetString(batch.TotalAmount, 10); !ok {
		total = big.NewInt(0)
	}
	total.Add(total, amt)
	batch.TotalAmount = total.String()

	// Check if batch should be marked ready
	maxBatchSize, err := k.GetMaxBatchSize(ctx)
	if err != nil {
		return "", err
	}

	if uint64(len(batch.Transactions)) >= maxBatchSize {
		batch.Status = "ready"
	}

	// Save pending batch
	if err := k.SetPendingBatch(ctx, batch); err != nil {
		return "", err
	}

	return txID, nil
}

// ProcessBatch processes all pending transactions in a batch
func (k *Keeper) ProcessBatch(ctx context.Context) (uint64, string, error) {
	// Check if batching is enabled
	enabled, err := k.IsBatchingEnabled(ctx)
	if err != nil {
		return 0, "", err
	}
	if !enabled {
		return 0, "", types.ErrBatchingDisabled
	}

	// Get pending batch
	batch, err := k.GetPendingBatchData(ctx)
	if err != nil {
		return 0, "", err
	}
	if batch == nil {
		return 0, "", nil
	}

	count := uint64(len(batch.Transactions))

	// Check minimum batch size
	minBatchSize, err := k.GetMinBatchSize(ctx)
	if err != nil {
		return 0, "", err
	}

	if count < minBatchSize {
		return 0, "", types.ErrBatchTooSmall
	}

	// Get current time for timestamp
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return 0, "", err
	}

	// Calculate gas savings
	individualGas := count * 21000 // Base gas per transaction
	batchGas := 21000 + (count-1)*15000 // Base + reduced gas per additional tx
	gasSavings := individualGas - batchGas

	// Calculate average gas price from batch
	avgGasPrice := k.GetAverageUtilization()

	// Create batch record
	record := BatchRecord{
		BatchID:          batch.BatchID,
		TransactionCount: count,
		TotalAmount:      batch.TotalAmount,
		ProcessedAt:      currentTime,
		GasSaved:         gasSavings,
		AverageGasPrice:  avgGasPrice,
		CompressionRatio: float32(batchGas) / float32(individualGas),
	}

	// Save batch record to history
	if err := k.AddBatchRecord(ctx, &record); err != nil {
		return 0, "", err
	}

	// Update statistics
	stats, err := k.GetBatchStatisticsData(ctx)
	if err != nil {
		return 0, "", err
	}

	stats.TotalBatchesProcessed++
	stats.TotalTransactionsBatched += count

	totalGasSaved := new(big.Int)
	if _, ok := totalGasSaved.SetString(stats.TotalGasSaved, 10); !ok {
		totalGasSaved = big.NewInt(0)
	}
	totalGasSaved.Add(totalGasSaved, big.NewInt(int64(gasSavings)))
	stats.TotalGasSaved = totalGasSaved.String()

	if err := k.SetBatchStatistics(ctx, stats); err != nil {
		return 0, "", err
	}

	// Clear pending batch
	if err := k.ClearPendingBatch(ctx); err != nil {
		return 0, "", err
	}

	return count, batch.BatchID, nil
}

// GetPendingBatch returns the current pending batch information
func (k *Keeper) GetPendingBatch() (uint64, string, string) {
	// This is a simplified version for queries without context
	// For full functionality, use GetPendingBatchData with context
	return 0, "0", ""
}

// GetBatchStatistics returns batching statistics
func (k *Keeper) GetBatchStatistics() (uint64, uint64, string, float32) {
	// This is a simplified version for queries without context
	// For full functionality, use GetBatchStatisticsData with context
	return 0, 0, "0", 0.0
}

// ShouldProcessBatch determines if the pending batch should be processed
func (k *Keeper) ShouldProcessBatch(ctx context.Context) (bool, error) {
	enabled, err := k.IsBatchingEnabled(ctx)
	if err != nil || !enabled {
		return false, err
	}

	batch, err := k.GetPendingBatchData(ctx)
	if err != nil {
		return false, err
	}
	if batch == nil {
		return false, nil
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return false, err
	}

	pendingCount := uint64(len(batch.Transactions))

	// Get max batch size
	maxBatchSize, err := k.GetMaxBatchSize(ctx)
	if err != nil {
		return false, err
	}

	// Process if we've reached max batch size
	if pendingCount >= maxBatchSize {
		return true, nil
	}

	// Get min batch size and timeout
	minBatchSize, err := k.GetMinBatchSize(ctx)
	if err != nil {
		return false, err
	}

	timeout, err := k.GetBatchTimeout(ctx)
	if err != nil {
		return false, err
	}

	// Process if we've reached minimum and batch has been pending for timeout duration
	if pendingCount >= minBatchSize {
		batchAge := currentTime - batch.CreatedAt
		if batchAge >= int64(timeout) {
			return true, nil
		}
	}

	return false, nil
}

// ============================
// KV STORE OPERATIONS
// ============================

// GetPendingBatchData retrieves the pending batch from KV store
func (k *Keeper) GetPendingBatchData(ctx context.Context) (*TransactionBatch, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(PendingBatchKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, nil
	}

	var batch TransactionBatch
	if err := json.Unmarshal(bz, &batch); err != nil {
		return nil, err
	}

	return &batch, nil
}

// SetPendingBatch stores the pending batch to KV store
func (k *Keeper) SetPendingBatch(ctx context.Context, batch *TransactionBatch) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(batch)
	if err != nil {
		return err
	}
	return store.Set(PendingBatchKey, bz)
}

// ClearPendingBatch removes the pending batch from KV store
func (k *Keeper) ClearPendingBatch(ctx context.Context) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(PendingBatchKey)
}

// AddBatchRecord adds a batch record to history
func (k *Keeper) AddBatchRecord(ctx context.Context, record *BatchRecord) error {
	store := k.storeService.OpenKVStore(ctx)

	// Use timestamp as part of key to ensure uniqueness
	key := append(BatchHistoryPrefix, []byte(fmt.Sprintf("%d-%s", record.ProcessedAt, record.BatchID))...)

	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return store.Set(key, bz)
}

// GetBatchStatisticsData retrieves batch statistics from KV store
func (k *Keeper) GetBatchStatisticsData(ctx context.Context) (*BatchStatistics, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(BatchStatisticsKey)
	if err != nil {
		return nil, err
	}

	if bz == nil {
		// Return default statistics
		return &BatchStatistics{
			TotalBatchesProcessed:    0,
			TotalTransactionsBatched: 0,
			TotalGasSaved:            "0",
			AverageCompressionRatio:  0.0,
		}, nil
	}

	var stats BatchStatistics
	if err := json.Unmarshal(bz, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// SetBatchStatistics stores batch statistics to KV store
func (k *Keeper) SetBatchStatistics(ctx context.Context, stats *BatchStatistics) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return store.Set(BatchStatisticsKey, bz)
}

// IsBatchingEnabled checks if batching is enabled
func (k *Keeper) IsBatchingEnabled(ctx context.Context) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(BatchEnabledKey)
	if err != nil {
		return false, err
	}
	if bz == nil {
		return true, nil // Default to enabled
	}
	return bz[0] == 1, nil
}

// SetBatchingEnabled sets whether batching is enabled
func (k *Keeper) SetBatchingEnabled(ctx context.Context, enabled bool) error {
	store := k.storeService.OpenKVStore(ctx)
	var bz []byte
	if enabled {
		bz = []byte{1}
	} else {
		bz = []byte{0}
	}
	return store.Set(BatchEnabledKey, bz)
}

// GetMaxBatchSize retrieves the maximum batch size
func (k *Keeper) GetMaxBatchSize(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(BatchConfigPrefix, []byte("max_size")...)
	bz, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 100, nil // Default max size
	}
	return binary.BigEndian.Uint64(bz), nil
}

// SetMaxBatchSize sets the maximum batch size
func (k *Keeper) SetMaxBatchSize(ctx context.Context, size uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(BatchConfigPrefix, []byte("max_size")...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, size)
	return store.Set(key, bz)
}

// GetMinBatchSize retrieves the minimum batch size
func (k *Keeper) GetMinBatchSize(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(BatchConfigPrefix, []byte("min_size")...)
	bz, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 10, nil // Default min size
	}
	return binary.BigEndian.Uint64(bz), nil
}

// SetMinBatchSize sets the minimum batch size
func (k *Keeper) SetMinBatchSize(ctx context.Context, size uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(BatchConfigPrefix, []byte("min_size")...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, size)
	return store.Set(key, bz)
}

// GetBatchTimeout retrieves the batch timeout in seconds
func (k *Keeper) GetBatchTimeout(ctx context.Context) (uint64, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(BatchConfigPrefix, []byte("timeout")...)
	bz, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 60, nil // Default 60 seconds
	}
	return binary.BigEndian.Uint64(bz), nil
}

// SetBatchTimeout sets the batch timeout in seconds
func (k *Keeper) SetBatchTimeout(ctx context.Context, timeout uint64) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(BatchConfigPrefix, []byte("timeout")...)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, timeout)
	return store.Set(key, bz)
}

// generateTxID generates a unique transaction ID
func (k *Keeper) generateTxID(sender, recipient, amount string, currentTime int64) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s-%d", sender, recipient, amount, currentTime)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// generateBatchID generates a unique batch ID
func (k *Keeper) generateBatchID(currentHeight uint64, currentTime int64) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("batch-%d-%d", currentHeight, currentTime)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
