package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// failedTxAnalysisWorker analyzes failed transaction patterns
func (k *Keeper) failedTxAnalysisWorker() {
	defer k.wg.Done()
	ticker := time.NewTicker(k.params.FailedTxAnalysisWindow)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			k.analyzeFailedTransactionPatterns()
		}
	}
}

// RecordFailedTransaction records a failed transaction for pattern analysis
func (k *Keeper) RecordFailedTransaction(tx *types.TransactionMonitorData, failureReason string) error {
	if !k.params.EnableFailedTxAnalysis {
		return nil
	}

	if tx == nil || tx.Status != "failed" {
		return types.ErrInvalidTransaction
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Create or update pattern
	patternKey := fmt.Sprintf("%s:%s", tx.Module, failureReason)
	pattern, exists := k.failedTxPatterns[patternKey]

	if !exists {
		pattern = &types.FailedTransactionPattern{
			ID:                generateID("failed-pattern"),
			Pattern:           patternKey,
			FailureReason:     failureReason,
			Occurrences:       0,
			AffectedAddresses: make([]string, 0),
			FirstSeen:         time.Now(),
			Severity:          types.SeverityLow,
			Metadata: map[string]interface{}{
				"module": tx.Module,
			},
		}
		k.failedTxPatterns[patternKey] = pattern
	}

	pattern.Occurrences++
	pattern.LastSeen = time.Now()

	// Track affected addresses
	if !containsString(pattern.AffectedAddresses, tx.Sender) {
		pattern.AffectedAddresses = append(pattern.AffectedAddresses, tx.Sender)
	}

	// Update severity based on occurrences
	pattern.Severity = k.determinePatternSeverity(pattern.Occurrences)

	// Create alert if threshold exceeded
	if pattern.Occurrences >= k.params.FailedTxPatternThreshold {
		if k.params.EnableAlerts {
			k.createFailedTxPatternAlert(pattern)
		}
	}

	return nil
}

// analyzeFailedTransactionPatterns performs periodic analysis of failed transactions
func (k *Keeper) analyzeFailedTransactionPatterns() {
	k.mu.RLock()
	defer k.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-k.params.FailedTxAnalysisWindow)

	for _, pattern := range k.failedTxPatterns {
		// Check if pattern is still active
		if pattern.LastSeen.After(cutoff) {
			// Pattern is still active, update metrics
			// In a real implementation, we would perform more sophisticated analysis
			continue
		}
	}
}

// GetFailedTransactionPatterns returns all failed transaction patterns
func (k *Keeper) GetFailedTransactionPatterns() []*types.FailedTransactionPattern {
	k.mu.RLock()
	defer k.mu.RUnlock()

	patterns := make([]*types.FailedTransactionPattern, 0, len(k.failedTxPatterns))
	for _, pattern := range k.failedTxPatterns {
		patterns = append(patterns, pattern)
	}

	return patterns
}

// GetFailedTransactionPatternsByReason returns patterns filtered by failure reason
func (k *Keeper) GetFailedTransactionPatternsByReason(reason string) []*types.FailedTransactionPattern {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var filtered []*types.FailedTransactionPattern
	for _, pattern := range k.failedTxPatterns {
		if pattern.FailureReason == reason {
			filtered = append(filtered, pattern)
		}
	}

	return filtered
}

// createFailedTxPatternAlert creates an alert for a failed transaction pattern
func (k *Keeper) createFailedTxPatternAlert(pattern *types.FailedTransactionPattern) {
	alert := &types.Alert{
		ID:       generateID("alert-failed-pattern"),
		Type:     types.AlertTypeFailedTxPattern,
		Severity: pattern.Severity,
		Message:  fmt.Sprintf("Failed transaction pattern detected: %s (%d occurrences)", pattern.FailureReason, pattern.Occurrences),
		Details: map[string]interface{}{
			"pattern_id":         pattern.ID,
			"failure_reason":     pattern.FailureReason,
			"occurrences":        pattern.Occurrences,
			"affected_addresses": len(pattern.AffectedAddresses),
			"first_seen":         pattern.FirstSeen,
			"last_seen":          pattern.LastSeen,
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

// determinePatternSeverity determines severity based on occurrence count
func (k *Keeper) determinePatternSeverity(occurrences int64) types.AlertSeverity {
	if occurrences >= 100 {
		return types.SeverityCritical
	} else if occurrences >= 50 {
		return types.SeverityHigh
	} else if occurrences >= 20 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}

// containsString checks if a string slice contains a value
func containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
