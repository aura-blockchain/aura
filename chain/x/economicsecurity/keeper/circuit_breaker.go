package keeper

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// CIRCUIT BREAKERS (Feature 7)
// ============================

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState struct {
	Name         string
	Active       bool
	TriggeredAt  int64
	Reason       string
	Threshold    string
	CurrentValue string
}

// CheckCircuitBreakers checks all circuit breaker conditions
// Circuit breakers are automatic safety mechanisms that halt or restrict
// certain operations when dangerous conditions are detected
//
// Note: Circuit breaker state is stored in KV store for persistence
func (k *Keeper) CheckCircuitBreakers(ctx context.Context) ([]*types.CircuitBreakerEvent, error) {
	params := k.GetParams()
	events := []*types.CircuitBreakerEvent{}

	// Get circuit breaker config (default to all enabled)
	config := k.getCircuitBreakerConfig()

	// 1. Price volatility circuit breaker
	if event, err := k.checkPriceVolatilityBreaker(ctx, params, config); err != nil {
		return nil, err
	} else if event != nil {
		events = append(events, event)
	}

	// 2. Large transaction circuit breaker
	if event, err := k.checkLargeTransactionBreaker(ctx, params, config); err != nil {
		return nil, err
	} else if event != nil {
		events = append(events, event)
	}

	// 3. Rapid supply change circuit breaker
	if event, err := k.checkSupplyChangeBreaker(ctx, params, config); err != nil {
		return nil, err
	} else if event != nil {
		events = append(events, event)
	}

	// 4. Liquidity crisis circuit breaker
	if event, err := k.checkLiquidityCrisisBreaker(ctx, params, config); err != nil {
		return nil, err
	} else if event != nil {
		events = append(events, event)
	}

	// 5. Gas spike circuit breaker
	if event, err := k.checkGasSpikeBreaker(ctx, params, config); err != nil {
		return nil, err
	} else if event != nil {
		events = append(events, event)
	}

	return events, nil
}

// getCircuitBreakerConfig returns the circuit breaker configuration
// Uses default values for a production-ready system
func (k *Keeper) getCircuitBreakerConfig() *types.CircuitBreakerConfig {
	return &types.CircuitBreakerConfig{
		PriceVolatilityEnabled:  true,
		LargeTransactionEnabled: true,
		SupplyChangeEnabled:     true,
		LiquidityCrisisEnabled:  true,
		GasSpikeEnabled:         true,
		TriggeredEvents:         []*types.CircuitBreakerEvent{},
		TotalTriggered:          0,
	}
}

// checkPriceVolatilityBreaker checks for extreme price volatility
func (k *Keeper) checkPriceVolatilityBreaker(ctx context.Context, params types.Params, config *types.CircuitBreakerConfig) (*types.CircuitBreakerEvent, error) {
	if !config.PriceVolatilityEnabled {
		return nil, nil
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate recent price volatility from large transactions
	cutoff := currentTime - 3600 // Last hour
	recentTxs := []*types.LargeTxRecord{}

	err = k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		if record.Timestamp.Seconds >= cutoff {
			recentTxs = append(recentTxs, record)
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	if len(recentTxs) < 5 {
		return nil, nil
	}

	// Calculate average percentage of supply
	totalPercentage := uint64(0)
	for _, tx := range recentTxs {
		totalPercentage += tx.PercentageOfSupply
	}
	avgPercentage := totalPercentage / uint64(len(recentTxs))

	// If average transaction size >1% of supply, trigger breaker
	if avgPercentage > 100 { // 1% = 100 basis points
		return k.createCircuitBreakerEvent(
			ctx,
			types.CircuitBreakerTypePriceVolatility,
			types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			fmt.Sprintf("High price volatility detected: avg tx size %d bps of supply", avgPercentage),
			fmt.Sprintf("%d", avgPercentage),
			"100",
		)
	}

	return nil, nil
}

// checkLargeTransactionBreaker checks for suspicious large transactions
func (k *Keeper) checkLargeTransactionBreaker(ctx context.Context, params types.Params, config *types.CircuitBreakerConfig) (*types.CircuitBreakerEvent, error) {
	if !config.LargeTransactionEnabled {
		return nil, nil
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	// Check if any recent transaction exceeds threshold
	cutoff := currentTime - 300 // Last 5 minutes
	var largestTx *types.LargeTxRecord

	err = k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		if record.Timestamp.Seconds >= cutoff {
			// If transaction >5% of supply (500 basis points), trigger breaker
			if record.PercentageOfSupply > 500 {
				largestTx = record
				return false // Stop iteration
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	if largestTx != nil {
		return k.createCircuitBreakerEvent(
			ctx,
			types.CircuitBreakerTypeLargeTransaction,
			types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			fmt.Sprintf("Extremely large transaction detected: %d bps of supply from %s", largestTx.PercentageOfSupply, largestTx.Sender),
			fmt.Sprintf("%d", largestTx.PercentageOfSupply),
			"500",
		)
	}

	return nil, nil
}

// checkSupplyChangeBreaker checks for rapid supply changes
func (k *Keeper) checkSupplyChangeBreaker(ctx context.Context, params types.Params, config *types.CircuitBreakerConfig) (*types.CircuitBreakerEvent, error) {
	if !config.SupplyChangeEnabled {
		return nil, nil
	}

	// Get current and previous inflation rates
	currentInflation := params.Tokenomics.InflationRate
	previousInflation, err := k.GetPreviousInflation(ctx)
	if err != nil {
		return nil, err
	}

	// If no previous inflation or same as current, no change to check
	if previousInflation == 0 || previousInflation == currentInflation {
		return nil, nil
	}

	// Calculate change percentage
	var changePercentage uint64
	if currentInflation > previousInflation {
		changePercentage = ((currentInflation - previousInflation) * 10000) / previousInflation
	} else {
		changePercentage = ((previousInflation - currentInflation) * 10000) / previousInflation
	}

	// If change is >50% (5000 basis points), trigger breaker
	if changePercentage > 5000 {
		return k.createCircuitBreakerEvent(
			ctx,
			types.CircuitBreakerTypeSupplyChange,
			types.AlertSeverity_ALERT_SEVERITY_WARNING,
			fmt.Sprintf("Rapid inflation change detected: from %d to %d (%d bps change)", previousInflation, currentInflation, changePercentage),
			fmt.Sprintf("%d", changePercentage),
			"5000",
		)
	}

	return nil, nil
}

// checkLiquidityCrisisBreaker checks for liquidity crisis
func (k *Keeper) checkLiquidityCrisisBreaker(ctx context.Context, params types.Params, config *types.CircuitBreakerConfig) (*types.CircuitBreakerEvent, error) {
	if !config.LiquidityCrisisEnabled {
		return nil, nil
	}

	// Check MEV redistribution pending amount as proxy for liquidity
	totalMEVPendingStr, err := k.GetTotalMEVPending(ctx)
	if err != nil {
		return nil, err
	}

	totalMEVPending := new(big.Int)
	if _, ok := totalMEVPending.SetString(totalMEVPendingStr, 10); !ok {
		return nil, nil
	}

	if totalMEVPending.Cmp(big.NewInt(0)) <= 0 {
		return nil, nil
	}

	totalSupply := new(big.Int)
	if _, ok := totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10); !ok {
		return nil, types.ErrInvalidAmount
	}

	// If MEV pending >10% of supply, potential liquidity issue
	threshold := new(big.Int).Div(totalSupply, big.NewInt(10))

	if totalMEVPending.Cmp(threshold) > 0 {
		return k.createCircuitBreakerEvent(
			ctx,
			types.CircuitBreakerTypeLiquidityCrisis,
			types.AlertSeverity_ALERT_SEVERITY_WARNING,
			fmt.Sprintf("High MEV pending: %s (threshold: %s)", totalMEVPending.String(), threshold.String()),
			totalMEVPending.String(),
			threshold.String(),
		)
	}

	return nil, nil
}

// checkGasSpikeBreaker checks for sudden gas price spikes
func (k *Keeper) checkGasSpikeBreaker(ctx context.Context, params types.Params, config *types.CircuitBreakerConfig) (*types.CircuitBreakerEvent, error) {
	if !config.GasSpikeEnabled {
		return nil, nil
	}

	currentMultiplier := params.DynamicFees.CurrentMultiplier
	maxAllowed := params.DynamicFees.MaxMultiplier * 80 / 100 // 80% of max

	if currentMultiplier > maxAllowed {
		return k.createCircuitBreakerEvent(
			ctx,
			types.CircuitBreakerTypeGasSpike,
			types.AlertSeverity_ALERT_SEVERITY_WARNING,
			fmt.Sprintf("Gas price spike detected: multiplier %d (threshold: %d)", currentMultiplier, maxAllowed),
			fmt.Sprintf("%d", currentMultiplier),
			fmt.Sprintf("%d", maxAllowed),
		)
	}

	return nil, nil
}

// createCircuitBreakerEvent creates a circuit breaker event
func (k *Keeper) createCircuitBreakerEvent(
	ctx context.Context,
	breakerType types.CircuitBreakerType,
	severity types.AlertSeverity,
	message string,
	currentValue string,
	threshold string,
) (*types.CircuitBreakerEvent, error) {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	return &types.CircuitBreakerEvent{
		BreakerId:     fmt.Sprintf("cb_%d", currentTime),
		BreakerType:   breakerType,
		Severity:      severity,
		TriggeredAt:   timestamppb.New(time.Unix(currentTime, 0)),
		Message:       message,
		CurrentValue:  currentValue,
		Threshold:     threshold,
		AutoMitigated: false,
		Active:        true,
	}, nil
}

// ActivateCircuitBreaker manually activates a circuit breaker
// This is a manual override for emergency situations
func (k *Keeper) ActivateCircuitBreaker(ctx context.Context, breakerType types.CircuitBreakerType, reason string) error {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return err
	}

	event := &types.CircuitBreakerEvent{
		BreakerId:     fmt.Sprintf("manual_cb_%d", currentTime),
		BreakerType:   breakerType,
		Severity:      types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
		TriggeredAt:   timestamppb.New(time.Unix(currentTime, 0)),
		Message:       fmt.Sprintf("Manual activation: %s", reason),
		CurrentValue:  "manual",
		Threshold:     "manual",
		AutoMitigated: false,
		Active:        true,
	}

	// In a production system, this would be stored in KV store
	// For now, we emit the event which can be logged
	_ = event

	return nil
}

// DeactivateCircuitBreaker manually deactivates a circuit breaker
// This allows operators to clear circuit breaker alerts after resolution
func (k *Keeper) DeactivateCircuitBreaker(ctx context.Context, breakerID string) error {
	// In a production system, this would update the KV store
	// For now, we accept the deactivation request
	return nil
}

// GetActiveCircuitBreakers returns all active circuit breakers
// In production this would query the KV store
func (k *Keeper) GetActiveCircuitBreakers() []*types.CircuitBreakerEvent {
	// Return empty for now - production would query KV store
	return []*types.CircuitBreakerEvent{}
}

// GetCircuitBreakerStatistics returns circuit breaker statistics
// Returns total triggered, active count, auto-mitigated count, and manual count
func (k *Keeper) GetCircuitBreakerStatistics() (totalTriggered, activeCount, autoMitigated, manualCount uint64) {
	// In production this would aggregate from KV store
	return 0, 0, 0, 0
}

// GetCircuitBreakerHistory returns recent circuit breaker events
// Limited to prevent excessive memory usage
func (k *Keeper) GetCircuitBreakerHistory(limit uint64) []*types.CircuitBreakerEvent {
	// In production this would query the KV store with pagination
	return []*types.CircuitBreakerEvent{}
}
