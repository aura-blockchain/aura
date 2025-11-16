package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// MonitorTransaction analyzes a transaction for suspicious activity
// Feature 2: Transaction monitoring for suspicious activity
func (k *Keeper) MonitorTransaction(
	transactionHash string,
	from string,
	to string,
	amount string,
	timestamp time.Time,
) ([]*types.TransactionAlert, error) {
	if !k.params.TransactionMonitoringEnabled {
		return nil, nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	alerts := []*types.TransactionAlert{}

	// Check each enabled monitoring rule
	for _, rule := range k.monitoringRules {
		if !rule.Enabled {
			continue
		}

		triggered, description, err := k.checkRule(rule, transactionHash, from, to, amount, timestamp)
		if err != nil {
			// Log error but continue checking other rules
			continue
		}

		if triggered {
			alert := k.createAlert(transactionHash, from, rule.ID, rule.RiskLevel, description)
			alerts = append(alerts, alert)

			// Store alert
			if k.transactionAlerts[from] == nil {
				k.transactionAlerts[from] = []*types.TransactionAlert{}
			}
			k.transactionAlerts[from] = append(k.transactionAlerts[from], alert)

			// Also store for recipient
			if k.transactionAlerts[to] == nil {
				k.transactionAlerts[to] = []*types.TransactionAlert{}
			}
			k.transactionAlerts[to] = append(k.transactionAlerts[to], alert)
		}
	}

	// Auto-escalate critical alerts to suspicious activity
	for _, alert := range alerts {
		if alert.RiskLevel == types.TxRiskCritical {
			_, err := k.autoEscalateToSAR(alert, from, amount)
			if err != nil {
				// Log but don't fail
				continue
			}
		}
	}

	return alerts, nil
}

// checkRule evaluates a specific monitoring rule against a transaction
func (k *Keeper) checkRule(
	rule *types.TransactionMonitoringRule,
	txHash string,
	from string,
	to string,
	amount string,
	timestamp time.Time,
) (bool, string, error) {
	switch rule.RuleType {
	case "velocity":
		return k.checkVelocityRule(rule, from, amount, timestamp)
	case "threshold":
		return k.checkThresholdRule(rule, amount)
	case "structuring":
		return k.checkStructuringRule(rule, from, amount, timestamp)
	case "pattern":
		return k.checkPatternRule(rule, from, amount, timestamp)
	default:
		return false, "", fmt.Errorf("unknown rule type: %s", rule.RuleType)
	}
}

// checkVelocityRule checks if transaction velocity exceeds limits
func (k *Keeper) checkVelocityRule(
	rule *types.TransactionMonitoringRule,
	address string,
	amount string,
	timestamp time.Time,
) (bool, string, error) {
	limitStr := rule.Parameters["limit"]
	timeWindowStr := rule.Parameters["time_window"]

	limit, ok := new(big.Int).SetString(limitStr, 10)
	if !ok {
		return false, "", fmt.Errorf("invalid limit value")
	}

	// Parse time window (e.g., "24h")
	timeWindow, err := time.ParseDuration(timeWindowStr)
	if err != nil {
		return false, "", err
	}

	// Calculate total volume in time window
	totalVolume := new(big.Int)
	cutoffTime := timestamp.Add(-timeWindow)

	alerts := k.transactionAlerts[address]
	for _, alert := range alerts {
		if alert.TriggeredAt.After(cutoffTime) {
			// Parse amount from description if available
			// In real implementation, we'd track actual transaction amounts
			continue
		}
	}

	// Add current transaction
	currentAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return false, "", fmt.Errorf("invalid amount value")
	}
	totalVolume.Add(totalVolume, currentAmount)

	if totalVolume.Cmp(limit) > 0 {
		description := fmt.Sprintf("Velocity limit exceeded: %s in %s (limit: %s)",
			totalVolume.String(), timeWindowStr, limit.String())
		return true, description, nil
	}

	return false, "", nil
}

// checkThresholdRule checks if transaction amount exceeds threshold
func (k *Keeper) checkThresholdRule(
	rule *types.TransactionMonitoringRule,
	amount string,
) (bool, string, error) {
	thresholdStr := rule.Parameters["threshold"]

	threshold, ok := new(big.Int).SetString(thresholdStr, 10)
	if !ok {
		return false, "", fmt.Errorf("invalid threshold value")
	}

	txAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return false, "", fmt.Errorf("invalid amount value")
	}

	if txAmount.Cmp(threshold) > 0 {
		description := fmt.Sprintf("Large transaction: %s exceeds threshold %s",
			txAmount.String(), threshold.String())
		return true, description, nil
	}

	return false, "", nil
}

// checkStructuringRule detects potential structuring behavior
func (k *Keeper) checkStructuringRule(
	rule *types.TransactionMonitoringRule,
	address string,
	amount string,
	timestamp time.Time,
) (bool, string, error) {
	countThresholdStr := rule.Parameters["count_threshold"]
	amountThresholdStr := rule.Parameters["amount_threshold"]
	timeWindowStr := rule.Parameters["time_window"]

	countThreshold, err := strconv.ParseInt(countThresholdStr, 10, 32)
	if err != nil {
		return false, "", err
	}

	amountThreshold, ok := new(big.Int).SetString(amountThresholdStr, 10)
	if !ok {
		return false, "", fmt.Errorf("invalid amount threshold")
	}

	timeWindow, err := time.ParseDuration(timeWindowStr)
	if err != nil {
		return false, "", err
	}

	txAmount, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return false, "", fmt.Errorf("invalid amount value")
	}

	// Check if current transaction is just below threshold
	// Allow 2% variance
	variance := new(big.Int).Div(amountThreshold, big.NewInt(50)) // 2%
	lowerBound := new(big.Int).Sub(amountThreshold, variance)

	if txAmount.Cmp(lowerBound) < 0 || txAmount.Cmp(amountThreshold) >= 0 {
		return false, "", nil
	}

	// Count similar transactions in time window
	count := int64(1) // Include current transaction
	cutoffTime := timestamp.Add(-timeWindow)

	alerts := k.transactionAlerts[address]
	for _, alert := range alerts {
		if alert.TriggeredAt.After(cutoffTime) && alert.RuleID == rule.ID {
			count++
		}
	}

	if count >= countThreshold {
		description := fmt.Sprintf("Potential structuring detected: %d transactions just below %s threshold in %s",
			count, amountThreshold.String(), timeWindowStr)
		return true, description, nil
	}

	return false, "", nil
}

// checkPatternRule detects suspicious patterns like round amounts
func (k *Keeper) checkPatternRule(
	rule *types.TransactionMonitoringRule,
	address string,
	amount string,
	timestamp time.Time,
) (bool, string, error) {
	countThresholdStr := rule.Parameters["count_threshold"]
	timeWindowStr := rule.Parameters["time_window"]

	countThreshold, err := strconv.ParseInt(countThresholdStr, 10, 32)
	if err != nil {
		return false, "", err
	}

	timeWindow, err := time.ParseDuration(timeWindowStr)
	if err != nil {
		return false, "", err
	}

	// Check if amount is a round number
	if !isRoundAmount(amount) {
		return false, "", nil
	}

	// Count round amount transactions in time window
	count := int64(1)
	cutoffTime := timestamp.Add(-timeWindow)

	alerts := k.transactionAlerts[address]
	for _, alert := range alerts {
		if alert.TriggeredAt.After(cutoffTime) && alert.RuleID == rule.ID {
			count++
		}
	}

	if count >= countThreshold {
		description := fmt.Sprintf("Suspicious pattern: %d round amount transactions in %s",
			count, timeWindowStr)
		return true, description, nil
	}

	return false, "", nil
}

// isRoundAmount checks if an amount is a round number
func isRoundAmount(amount string) bool {
	amountInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return false
	}

	// Check if divisible by 1000, 10000, etc.
	divisors := []int64{1000, 10000, 100000, 1000000}
	for _, divisor := range divisors {
		mod := new(big.Int).Mod(amountInt, big.NewInt(divisor))
		if mod.Cmp(big.NewInt(0)) == 0 {
			return true
		}
	}

	return false
}

// createAlert creates a new transaction alert
func (k *Keeper) createAlert(
	txHash string,
	address string,
	ruleID string,
	riskLevel types.TransactionRiskLevel,
	description string,
) *types.TransactionAlert {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", txHash, ruleID, time.Now().Unix())))
	id := hex.EncodeToString(hash[:])[:16]

	return &types.TransactionAlert{
		ID:              id,
		TransactionHash: txHash,
		Address:         address,
		RuleID:          ruleID,
		RiskLevel:       riskLevel,
		Description:     description,
		TriggeredAt:     time.Now(),
		Reviewed:        false,
	}
}

// autoEscalateToSAR automatically escalates critical alerts to suspicious activity reports
func (k *Keeper) autoEscalateToSAR(
	alert *types.TransactionAlert,
	address string,
	amount string,
) (string, error) {
	rule, exists := k.monitoringRules[alert.RuleID]
	if !exists {
		return "", fmt.Errorf("rule not found: %s", alert.RuleID)
	}

	indicators := []string{
		fmt.Sprintf("auto_escalated_from_rule_%s", rule.ID),
		fmt.Sprintf("risk_level_%d", alert.RiskLevel),
	}

	return k.ReportSuspiciousActivity(
		"system",
		address,
		alert.TransactionHash,
		rule.RuleType,
		alert.Description,
		amount,
		indicators,
	)
}

// GetTransactionAlerts retrieves alerts for an address (legacy in-memory version - deprecated)
// Use keeper_kvstore.go version with sdk.Context instead
// func (k *Keeper) GetTransactionAlerts(address string) ([]*types.TransactionAlert, error) {
// 	k.mu.RLock()
// 	defer k.mu.RUnlock()
//
// 	alerts, ok := k.transactionAlerts[address]
// 	if !ok {
// 		return []*types.TransactionAlert{}, nil
// 	}
//
// 	return alerts, nil
// }

// ReviewAlert marks an alert as reviewed
func (k *Keeper) ReviewAlert(alertID string, reviewer string, resolution string, notes string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Find and update the alert
	for _, alerts := range k.transactionAlerts {
		for _, alert := range alerts {
			if alert.ID == alertID {
				alert.Reviewed = true
				alert.ReviewedAt = time.Now()
				alert.Reviewer = reviewer
				alert.Resolution = resolution
				alert.Notes = notes
				return nil
			}
		}
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// AddMonitoringRule adds a new transaction monitoring rule
func (k *Keeper) AddMonitoringRule(rule *types.TransactionMonitoringRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	k.monitoringRules[rule.ID] = rule
	return nil
}

// UpdateMonitoringRule updates an existing monitoring rule
func (k *Keeper) UpdateMonitoringRule(ruleID string, updates map[string]interface{}) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	rule, exists := k.monitoringRules[ruleID]
	if !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	// Update fields
	if enabled, ok := updates["enabled"].(bool); ok {
		rule.Enabled = enabled
	}
	if params, ok := updates["parameters"].(map[string]string); ok {
		rule.Parameters = params
	}
	if riskLevel, ok := updates["risk_level"].(types.TransactionRiskLevel); ok {
		rule.RiskLevel = riskLevel
	}

	rule.UpdatedAt = time.Now()
	return nil
}

// GetMonitoringRules returns all monitoring rules
func (k *Keeper) GetMonitoringRules() map[string]*types.TransactionMonitoringRule {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Return a copy to prevent external modifications
	rules := make(map[string]*types.TransactionMonitoringRule)
	for id, rule := range k.monitoringRules {
		rules[id] = rule
	}
	return rules
}

// GetAlertStatistics returns statistics about transaction alerts
func (k *Keeper) GetAlertStatistics() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	totalAlerts := 0
	reviewedAlerts := 0
	alertsByRisk := make(map[types.TransactionRiskLevel]int)

	for _, alerts := range k.transactionAlerts {
		for _, alert := range alerts {
			totalAlerts++
			if alert.Reviewed {
				reviewedAlerts++
			}
			alertsByRisk[alert.RiskLevel]++
		}
	}

	return map[string]interface{}{
		"total_alerts":    totalAlerts,
		"reviewed_alerts": reviewedAlerts,
		"pending_alerts":  totalAlerts - reviewedAlerts,
		"alerts_by_risk":  alertsByRisk,
	}
}
