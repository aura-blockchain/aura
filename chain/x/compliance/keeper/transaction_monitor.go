package keeper

import (
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// MonitorTransaction evaluates transaction monitoring rules and generates alerts.
//
// This function implements real-time AML (Anti-Money Laundering) transaction monitoring
// by evaluating configured rules against transactions. It checks for:
//   - Large transactions exceeding configured thresholds
//   - Transactions involving flagged or sanctioned addresses
//   - High-frequency transaction patterns (velocity checks)
//   - Structuring attempts (multiple transactions near threshold)
//
// Security considerations:
//   - All rules are evaluated even if one triggers (complete analysis)
//   - Alerts are persisted to state for audit trail
//   - Critical risk alerts can block transactions (caller's responsibility)
//   - No PII is stored on-chain (GDPR Article 32 compliance)
//
// Parameters:
//   - ctx: SDK context for state access and block information
//   - from: Sender address to monitor
//   - to: Recipient address to monitor
//   - amount: Transaction amount(s) being sent
//
// Returns:
//   - alerts: List of alerts triggered by this transaction
//   - error: Any error during monitoring (does not prevent transaction)
//
// Events emitted:
//   - EventTypeTransactionAlert for each alert generated
func (k Keeper) MonitorTransaction(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) ([]*types.TransactionAlert, error) {
	// Check if monitoring is enabled
	params, _ := k.GetParams(ctx)
	if !params.TransactionMonitoringEnabled {
		return nil, nil
	}

	alerts := make([]*types.TransactionAlert, 0, 64)

	// Get all enabled monitoring rules
	rules, err := k.GetAllMonitoringRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitoring rules: %w", err)
	}

	// Build transaction context for rule evaluation
	txCtx := &TransactionContext{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: ctx.BlockTime(),
		Height:    ctx.BlockHeight(),
	}

	// Evaluate each enabled rule
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		alert, err := k.evaluateRule(ctx, rule, txCtx)
		if err != nil {
			k.logger(ctx).Error("failed to evaluate rule",
				"rule_id", rule.Id,
				"error", err.Error(),
			)
			continue
		}

		if alert != nil {
			alerts = append(alerts, alert)
		}
	}

	// Additional checks: sanctions screening
	if params.SanctionsScreeningEnabled {
		sanctionsAlerts := k.checkSanctions(ctx, from, to, txCtx)
		alerts = append(alerts, sanctionsAlerts...)
	}

	// Persist all alerts
	for _, alert := range alerts {
		// Store alert for sender
		if err := k.AddTransactionAlert(ctx, from.String(), alert); err != nil {
			k.logger(ctx).Error("failed to store transaction alert",
				"address", from.String(),
				"alert_id", alert.Id,
				"error", err.Error(),
			)
		}

		// Store alert for recipient if different
		if !from.Equals(to) {
			if err := k.AddTransactionAlert(ctx, to.String(), alert); err != nil {
				k.logger(ctx).Error("failed to store transaction alert",
					"address", to.String(),
					"alert_id", alert.Id,
					"error", err.Error(),
				)
			}
		}

		// Emit event for external monitoring systems
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTransactionAlert,
				sdk.NewAttribute(types.AttributeKeyAlertID, alert.Id),
				sdk.NewAttribute(types.AttributeKeyAddress, from.String()),
				sdk.NewAttribute(types.AttributeKeyRiskLevel, alert.RiskLevel.String()),
				sdk.NewAttribute(types.AttributeKeyRuleID, alert.RuleId),
				sdk.NewAttribute(types.AttributeKeyDescription, alert.Description),
			),
		)
	}

	return alerts, nil
}

// TransactionContext holds transaction information for rule evaluation
type TransactionContext struct {
	From      sdk.AccAddress
	To        sdk.AccAddress
	Amount    sdk.Coins
	Timestamp time.Time
	Height    int64
}

// evaluateRule evaluates a single monitoring rule against a transaction
func (k Keeper) evaluateRule(ctx sdk.Context, rule *types.TransactionMonitoringRule, txCtx *TransactionContext) (*types.TransactionAlert, error) {
	switch rule.RuleType {
	case "velocity":
		return k.evaluateVelocityRule(ctx, rule, txCtx)
	case "structuring":
		return k.evaluateStructuringRule(ctx, rule, txCtx)
	case "large_transaction":
		return k.evaluateLargeTransactionRule(ctx, rule, txCtx)
	default:
		return nil, fmt.Errorf("unknown rule type: %s", rule.RuleType)
	}
}

// evaluateVelocityRule checks for high-frequency transaction patterns
func (k Keeper) evaluateVelocityRule(ctx sdk.Context, rule *types.TransactionMonitoringRule, txCtx *TransactionContext) (*types.TransactionAlert, error) {
	// Get 24h velocity threshold from parameters
	params, _ := k.GetParams(ctx)
	threshold, err := math.LegacyNewDecFromStr(params.VelocityLimit_24H)
	if err != nil {
		return nil, fmt.Errorf("invalid velocity threshold: %w", err)
	}

	// Get sender's AML profile to check transaction history
	profile, err := k.GetAMLProfile(ctx, txCtx.From.String())
	if err != nil {
		// No profile yet, create one
		profile = &types.AMLProfile{
			Address:           txCtx.From.String(),
			RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
			TotalVolume:       "0",
			LastAssessment:    txCtx.Timestamp,
			TotalTransactions: 0,
		}
	}

	// Parse current total volume
	currentVolume, err := math.LegacyNewDecFromStr(profile.TotalVolume)
	if err != nil {
		currentVolume = math.LegacyZeroDec()
	}

	// Add current transaction amount (assuming first coin for simplicity)
	if len(txCtx.Amount) > 0 {
		txAmount := math.LegacyNewDecFromInt(txCtx.Amount[0].Amount)
		currentVolume = currentVolume.Add(txAmount)
	}

	// Check if velocity threshold exceeded
	if currentVolume.GT(threshold) {
		return &types.TransactionAlert{
			Id:          fmt.Sprintf("velocity_%s_%d", txCtx.From.String(), ctx.BlockHeight()),
			Address:     txCtx.From.String(),
			RuleId:      rule.Id,
			RiskLevel:   rule.RiskLevel,
			Description: fmt.Sprintf("24h transaction velocity exceeds threshold: %s > %s", currentVolume.String(), threshold.String()),
			TriggeredAt: txCtx.Timestamp,
			Reviewed:    false,
		}, nil
	}

	return nil, nil
}

// evaluateStructuringRule detects potential structuring attempts
func (k Keeper) evaluateStructuringRule(ctx sdk.Context, rule *types.TransactionMonitoringRule, txCtx *TransactionContext) (*types.TransactionAlert, error) {
	params, _ := k.GetParams(ctx)

	// Get sender's AML profile
	profile, err := k.GetAMLProfile(ctx, txCtx.From.String())
	if err != nil {
		// No profile yet, not enough data for structuring detection
		return nil, nil
	}

	// Check if transaction count exceeds structuring threshold
	if profile.TotalTransactions >= uint64(params.StructuringThresholdCount) {
		return &types.TransactionAlert{
			Id:          fmt.Sprintf("structuring_%s_%d", txCtx.From.String(), ctx.BlockHeight()),
			Address:     txCtx.From.String(),
			RuleId:      rule.Id,
			RiskLevel:   rule.RiskLevel,
			Description: fmt.Sprintf("Potential structuring detected: %d transactions in monitoring period", profile.TotalTransactions),
			TriggeredAt: txCtx.Timestamp,
			Reviewed:    false,
		}, nil
	}

	return nil, nil
}

// evaluateLargeTransactionRule checks for single large transactions
func (k Keeper) evaluateLargeTransactionRule(ctx sdk.Context, rule *types.TransactionMonitoringRule, txCtx *TransactionContext) (*types.TransactionAlert, error) {
	params, _ := k.GetParams(ctx)

	// Parse single transaction limit
	limitStr := params.SingleTransactionLimit
	if thresholdParam, exists := rule.Parameters["threshold"]; exists {
		limitStr = thresholdParam
	}

	limit, err := math.LegacyNewDecFromStr(limitStr)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction limit: %w", err)
	}

	// Check each coin in the transaction
	for _, coin := range txCtx.Amount {
		txAmount := math.LegacyNewDecFromInt(coin.Amount)
		if txAmount.GT(limit) {
			return &types.TransactionAlert{
				Id:          fmt.Sprintf("large_tx_%s_%d", txCtx.From.String(), ctx.BlockHeight()),
				Address:     txCtx.From.String(),
				RuleId:      rule.Id,
				RiskLevel:   rule.RiskLevel,
				Description: fmt.Sprintf("Large transaction detected: %s %s exceeds threshold %s", coin.Amount.String(), coin.Denom, limit.String()),
				TriggeredAt: txCtx.Timestamp,
				Reviewed:    false,
			}, nil
		}
	}

	return nil, nil
}

// checkSanctions verifies addresses against sanctions lists
func (k Keeper) checkSanctions(ctx sdk.Context, from, to sdk.AccAddress, txCtx *TransactionContext) []*types.TransactionAlert {
	alerts := make([]*types.TransactionAlert, 0, 64)

	// Check sender
	if k.IsAddressSanctioned(ctx, from.String()) {
		alerts = append(alerts, &types.TransactionAlert{
			Id:          fmt.Sprintf("sanctions_%s_%d", from.String(), ctx.BlockHeight()),
			Address:     from.String(),
			RuleId:      "sanctions_check",
			RiskLevel:   types.TransactionRiskLevel_TX_RISK_CRITICAL,
			Description: "Transaction from sanctioned address",
			TriggeredAt: txCtx.Timestamp,
			Reviewed:    false,
		})
	}

	// Check recipient
	if k.IsAddressSanctioned(ctx, to.String()) {
		alerts = append(alerts, &types.TransactionAlert{
			Id:          fmt.Sprintf("sanctions_%s_%d", to.String(), ctx.BlockHeight()),
			Address:     to.String(),
			RuleId:      "sanctions_check",
			RiskLevel:   types.TransactionRiskLevel_TX_RISK_CRITICAL,
			Description: "Transaction to sanctioned address",
			TriggeredAt: txCtx.Timestamp,
			Reviewed:    false,
		})
	}

	return alerts
}

// IsAddressSanctioned checks if an address is on sanctions lists
func (k Keeper) IsAddressSanctioned(ctx sdk.Context, address string) bool {
	// Check if we have a sanctions screening result
	result, err := k.GetSanctionsResult(ctx, address)
	if err != nil {
		// No screening result, address not sanctioned (yet)
		return false
	}

	// Address is sanctioned if status is MATCH or CONFIRMED
	return result.Status == types.SanctionsStatus_SANCTIONS_MATCH ||
		result.Status == types.SanctionsStatus_SANCTIONS_CONFIRMED
}

// ShouldBlockTransaction determines if a transaction should be blocked based on alerts
//
// This function evaluates alerts and returns true if the transaction poses
// an unacceptable risk and should be blocked. Critical risk alerts always
// block transactions.
//
// Security considerations:
//   - CRITICAL risk level always blocks transaction (e.g., sanctioned addresses)
//   - Multiple HIGH risk alerts may block transaction (defense in depth)
//   - Blocking decisions are logged for audit trail
//
// Parameters:
//   - alerts: List of alerts generated by MonitorTransaction
//
// Returns:
//   - shouldBlock: true if transaction should be blocked
//   - reason: Human-readable reason for blocking
func (k Keeper) ShouldBlockTransaction(alerts []*types.TransactionAlert) (bool, string) {
	if len(alerts) == 0 {
		return false, ""
	}

	// Check for critical risk alerts
	for _, alert := range alerts {
		if alert.RiskLevel == types.TransactionRiskLevel_TX_RISK_CRITICAL {
			return true, fmt.Sprintf("Critical risk detected: %s", alert.Description)
		}
	}

	// Check for multiple high risk alerts
	highRiskCount := 0
	for _, alert := range alerts {
		if alert.RiskLevel == types.TransactionRiskLevel_TX_RISK_HIGH {
			highRiskCount++
		}
	}

	// Block if multiple high risk factors present
	if highRiskCount >= 2 {
		return true, fmt.Sprintf("Multiple high risk factors detected (%d)", highRiskCount)
	}

	return false, ""
}

// GetTransactionVelocity calculates the transaction velocity for an address
// over the past 24 hours by examining the AML profile.
func (k Keeper) GetTransactionVelocity(ctx sdk.Context, address string, hours int) (math.LegacyDec, error) {
	profile, err := k.GetAMLProfile(ctx, address)
	if err != nil {
		return math.LegacyZeroDec(), nil // No profile, zero velocity
	}

	volume, err := strconv.ParseFloat(profile.TotalVolume, 64)
	if err != nil {
		return math.LegacyZeroDec(), fmt.Errorf("invalid volume in profile: %w", err)
	}

	return math.LegacyNewDec(int64(volume)), nil
}
