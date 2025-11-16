package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// MonitorTransaction records and analyzes a transaction
func (k *Keeper) MonitorTransaction(tx *types.TransactionMonitorData) error {
	if !k.params.EnableTransactionMonitoring {
		return nil
	}

	if tx == nil {
		return types.ErrInvalidTransaction
	}

	// Validate transaction data
	if err := k.validateTransaction(tx); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Store transaction
	k.transactions[tx.TxHash] = tx

	// Update Prometheus metrics
	k.metrics.TotalTransactions.WithLabelValues(tx.Status, tx.Module).Inc()
	k.metrics.TransactionGasUsed.WithLabelValues(tx.Module).Observe(float64(tx.GasUsed))

	// Check for large transactions
	if tx.Amount >= k.params.LargeTransactionThreshold {
		tx.IsLargeTransfer = true
		k.metrics.LargeTransactions.Inc()

		// Create alert for large transaction
		if k.params.EnableAlerts {
			k.createLargeTransactionAlert(tx)
		}
	}

	// Track failed transactions
	if tx.Status == "failed" {
		k.metrics.FailedTransactions.WithLabelValues("unknown", tx.Module).Inc()
		k.recordFailedTransaction(tx)
	}

	// Run anomaly detection
	if k.params.EnableAnomalyDetection {
		anomalyScore, err := k.detectTransactionAnomaly(tx)
		if err == nil {
			tx.AnomalyScore = anomalyScore
			if anomalyScore >= k.params.AnomalyThreshold {
				k.createAnomalyAlert(tx, anomalyScore)
			}
		}
	}

	return nil
}

// validateTransaction validates transaction data
func (k *Keeper) validateTransaction(tx *types.TransactionMonitorData) error {
	if tx.TxHash == "" {
		return fmt.Errorf("transaction hash cannot be empty")
	}
	if tx.Timestamp.IsZero() {
		return fmt.Errorf("transaction timestamp cannot be zero")
	}
	if tx.BlockHeight < 0 {
		return fmt.Errorf("block height cannot be negative")
	}
	return nil
}

// GetTransaction retrieves a monitored transaction by hash
func (k *Keeper) GetTransaction(txHash string) (*types.TransactionMonitorData, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	tx, exists := k.transactions[txHash]
	if !exists {
		return nil, types.ErrInvalidTransaction
	}

	return tx, nil
}

// GetRecentTransactions returns recent transactions
func (k *Keeper) GetRecentTransactions(limit int) []*types.TransactionMonitorData {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var txs []*types.TransactionMonitorData
	for _, tx := range k.transactions {
		txs = append(txs, tx)
		if len(txs) >= limit {
			break
		}
	}

	return txs
}

// GetTransactionsByModule returns transactions for a specific module
func (k *Keeper) GetTransactionsByModule(module string, limit int) []*types.TransactionMonitorData {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var txs []*types.TransactionMonitorData
	for _, tx := range k.transactions {
		if tx.Module == module {
			txs = append(txs, tx)
			if len(txs) >= limit {
				break
			}
		}
	}

	return txs
}

// GetLargeTransactions returns all large transactions
func (k *Keeper) GetLargeTransactions() []*types.TransactionMonitorData {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var txs []*types.TransactionMonitorData
	for _, tx := range k.transactions {
		if tx.IsLargeTransfer {
			txs = append(txs, tx)
		}
	}

	return txs
}

// createLargeTransactionAlert creates an alert for a large transaction
func (k *Keeper) createLargeTransactionAlert(tx *types.TransactionMonitorData) {
	alert := &types.Alert{
		ID:       generateID("alert-large-tx"),
		Type:     types.AlertTypeLargeTransaction,
		Severity: types.SeverityMedium,
		Message:  fmt.Sprintf("Large transaction detected: %d tokens", tx.Amount),
		Details: map[string]interface{}{
			"tx_hash":      tx.TxHash,
			"amount":       tx.Amount,
			"sender":       tx.Sender,
			"receiver":     tx.Receiver,
			"block_height": tx.BlockHeight,
		},
		Timestamp:        time.Now(),
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	k.alerts[alert.ID] = alert
	k.metrics.TotalAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
	k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
}

// recordFailedTransaction records a failed transaction for pattern analysis
func (k *Keeper) recordFailedTransaction(tx *types.TransactionMonitorData) {
	// This will be processed by the failed transaction analysis worker
	// For now, just track the metrics
	k.metrics.FailedTransactions.WithLabelValues("processing", tx.Module).Inc()
}

// GetTransactionStats returns transaction statistics
func (k *Keeper) GetTransactionStats() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var totalTxs, successTxs, failedTxs, largeTxs int
	var totalGasUsed uint64

	for _, tx := range k.transactions {
		totalTxs++
		if tx.Status == "success" {
			successTxs++
		} else if tx.Status == "failed" {
			failedTxs++
		}
		if tx.IsLargeTransfer {
			largeTxs++
		}
		totalGasUsed += tx.GasUsed
	}

	avgGasUsed := uint64(0)
	if totalTxs > 0 {
		avgGasUsed = totalGasUsed / uint64(totalTxs)
	}

	successRate := 0.0
	if totalTxs > 0 {
		successRate = float64(successTxs) / float64(totalTxs) * 100
	}

	return map[string]interface{}{
		"total_transactions":   totalTxs,
		"successful_transactions": successTxs,
		"failed_transactions":  failedTxs,
		"large_transactions":   largeTxs,
		"total_gas_used":       totalGasUsed,
		"average_gas_used":     avgGasUsed,
		"success_rate":         successRate,
	}
}
