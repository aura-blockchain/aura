package keeper

import (
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AutoScalingDecision represents a scaling decision
type AutoScalingDecision struct {
	TxType        types.TransactionType
	CurrentAmount uint64
	NewAmount     uint64
	CacheHitRate  float64
	Decision      string // "scale_up", "scale_down", "no_change"
	Reason        string
}

// RunAutoScaling executes the auto-scaling algorithm
func (k *Keeper) RunAutoScaling() ([]*types.EventAutoScaling, error) {
	params := k.GetParams()

	if !params.Enabled || params.AutoScalingConfig == nil || !params.AutoScalingConfig.Enabled {
		return nil, nil
	}

	currentTime := time.Unix(k.currentTime, 0)

	// Check cooldown period
	if !k.lastAutoScaleRun.IsZero() {
		cooldownDuration := time.Duration(params.AutoScalingConfig.CooldownMinutes) * time.Minute
		if currentTime.Sub(k.lastAutoScaleRun) < cooldownDuration {
			return nil, nil // Still in cooldown
		}
	}

	k.lastAutoScaleRun = currentTime

	events := []*types.EventAutoScaling{}

	// Evaluate each transaction type
	for txType, currentAmount := range k.typeAmounts {
		if txType == types.TxTypeUnspecified {
			continue
		}

		decision := k.evaluateScaling(txType, currentAmount)

		if decision.Decision != "no_change" {
			// Check per-type cooldown
			if lastAdjustment, ok := k.lastScalingAdjustment[txType]; ok {
				cooldownDuration := time.Duration(params.AutoScalingConfig.CooldownMinutes) * time.Minute
				if currentTime.Sub(lastAdjustment) < cooldownDuration {
					continue // Skip this type, still in cooldown
				}
			}

			// Apply the scaling decision
			k.SetTypeAmount(txType, decision.NewAmount)
			k.lastScalingAdjustment[txType] = currentTime

			// Create event
			event := &types.EventAutoScaling{
				TxType:         txType,
				PreviousAmount: decision.CurrentAmount,
				NewAmount:      decision.NewAmount,
				CacheHitRate:   decision.CacheHitRate,
				Reason:         decision.Reason,
				Timestamp:      timestamppb.New(currentTime),
			}

			events = append(events, event)
		}
	}

	return events, nil
}

// evaluateScaling evaluates whether to scale up or down for a transaction type
func (k *Keeper) evaluateScaling(txType types.TransactionType, currentAmount uint64) AutoScalingDecision {
	params := k.GetParams()
	config := params.AutoScalingConfig

	decision := AutoScalingDecision{
		TxType:        txType,
		CurrentAmount: currentAmount,
		NewAmount:     currentAmount,
		Decision:      "no_change",
		Reason:        "metrics within acceptable range",
	}

	// Get metrics for this type
	metrics := k.metrics.GetTypeMetrics(txType)
	if metrics == nil {
		decision.Reason = "no metrics available yet"
		return decision
	}

	decision.CacheHitRate = metrics.CacheHitRate

	// Evaluate scaling decision
	if config.ShouldScaleUp(metrics) {
		decision.NewAmount = config.CalculateNewAmount(currentAmount, true, txType)
		decision.Decision = "scale_up"
		decision.Reason = "cache hit rate above target - increasing pre-validation amount"
	} else if config.ShouldScaleDown(metrics) {
		decision.NewAmount = config.CalculateNewAmount(currentAmount, false, txType)
		decision.Decision = "scale_down"
		decision.Reason = "cache hit rate below minimum - decreasing pre-validation amount"
	}

	// Additional heuristics

	// 1. Check execution rate
	if metrics.TotalExecuted > 0 && metrics.TotalPreValidated > 0 {
		executionRate := float64(metrics.TotalExecuted) / float64(metrics.TotalPreValidated)

		// If very few pre-validated transactions are being used, scale down
		if executionRate < 0.3 && decision.Decision != "scale_down" {
			decision.NewAmount = config.CalculateNewAmount(currentAmount, false, txType)
			decision.Decision = "scale_down"
			decision.Reason = "low execution rate - too many unused pre-validations"
		}

		// If most pre-validated transactions are being used, scale up
		if executionRate > 0.9 && decision.Decision != "scale_up" {
			decision.NewAmount = config.CalculateNewAmount(currentAmount, true, txType)
			decision.Decision = "scale_up"
			decision.Reason = "high execution rate - demand exceeds supply"
		}
	}

	// 2. Check expiration rate
	if metrics.TotalPreValidated > 0 {
		expirationRate := float64(metrics.TotalExpired) / float64(metrics.TotalPreValidated)

		// If too many are expiring, scale down
		if expirationRate > 0.5 && decision.Decision != "scale_down" {
			decision.NewAmount = config.CalculateNewAmount(currentAmount, false, txType)
			decision.Decision = "scale_down"
			decision.Reason = "high expiration rate - wasting resources"
		}
	}

	// 3. Time savings evaluation
	if metrics.AvgTimeSavingsMs > 0 {
		// If we're saving significant time, maintain or increase
		if metrics.AvgTimeSavingsMs > 100 && decision.Decision == "scale_down" {
			// Don't scale down if we're saving significant time
			decision.Decision = "no_change"
			decision.NewAmount = currentAmount
			decision.Reason = "significant time savings - maintaining current level"
		}
	}

	return decision
}

// GetAutoScalingRecommendations returns recommendations without applying them
func (k *Keeper) GetAutoScalingRecommendations() []AutoScalingDecision {
	recommendations := []AutoScalingDecision{}

	for txType, currentAmount := range k.typeAmounts {
		if txType == types.TxTypeUnspecified {
			continue
		}

		decision := k.evaluateScaling(txType, currentAmount)
		recommendations = append(recommendations, decision)
	}

	return recommendations
}

// AdjustTypeAmount manually adjusts the amount for a transaction type (admin function)
func (k *Keeper) AdjustTypeAmount(txType types.TransactionType, newAmount uint64) error {
	params := k.GetParams()

	if params.AutoScalingConfig == nil {
		return types.ErrInvalidParameters
	}

	// Validate against max
	maxAmount := params.AutoScalingConfig.GetMaxAmount(txType)
	if newAmount > maxAmount {
		newAmount = maxAmount
	}

	// Validate against min
	initialAmount := params.AutoScalingConfig.GetInitialAmount(txType)
	if newAmount < initialAmount {
		newAmount = initialAmount
	}

	k.SetTypeAmount(txType, newAmount)
	k.lastScalingAdjustment[txType] = time.Unix(k.currentTime, 0)

	return nil
}

// ResetAutoScaling resets auto-scaling to initial amounts
func (k *Keeper) ResetAutoScaling() {
	params := k.GetParams()

	if params.AutoScalingConfig == nil {
		return
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	for txTypeStr, amount := range params.AutoScalingConfig.InitialAmounts {
		txType := k.parseTransactionType(txTypeStr)
		k.typeAmounts[txType] = amount
	}

	k.lastScalingAdjustment = make(map[types.TransactionType]time.Time)
	k.lastAutoScaleRun = time.Time{}
}

// GetScalingHistory returns recent scaling adjustments
func (k *Keeper) GetScalingHistory(txType types.TransactionType) time.Time {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if lastAdjustment, ok := k.lastScalingAdjustment[txType]; ok {
		return lastAdjustment
	}

	return time.Time{}
}

// CalculateOptimalAmounts calculates optimal amounts based on historical data
func (k *Keeper) CalculateOptimalAmounts(evaluationPeriodHours uint32) map[types.TransactionType]uint64 {
	optimal := make(map[types.TransactionType]uint64)

	params := k.GetParams()
	if params.AutoScalingConfig == nil {
		return optimal
	}

	for txType := range k.typeAmounts {
		if txType == types.TxTypeUnspecified {
			continue
		}

		metrics := k.metrics.GetTypeMetrics(txType)
		if metrics == nil {
			// Use initial amount if no metrics
			optimal[txType] = params.AutoScalingConfig.GetInitialAmount(txType)
			continue
		}

		// Calculate based on execution rate and cache hit rate
		currentAmount := k.typeAmounts[txType]

		if metrics.TotalExecuted == 0 {
			// No executions yet, use initial amount
			optimal[txType] = params.AutoScalingConfig.GetInitialAmount(txType)
			continue
		}

		// Target: maintain cache hit rate at target level
		targetHitRate := params.AutoScalingConfig.TargetCacheHitRate
		currentHitRate := metrics.CacheHitRate

		if currentHitRate == 0 {
			optimal[txType] = currentAmount
			continue
		}

		// Calculate multiplier needed to reach target
		multiplier := targetHitRate / currentHitRate

		// Apply conservative scaling (don't change too drastically)
		if multiplier > 2.0 {
			multiplier = 2.0
		} else if multiplier < 0.5 {
			multiplier = 0.5
		}

		optimalAmount := uint64(float64(currentAmount) * multiplier)

		// Apply bounds
		maxAmount := params.AutoScalingConfig.GetMaxAmount(txType)
		initialAmount := params.AutoScalingConfig.GetInitialAmount(txType)

		if optimalAmount > maxAmount {
			optimalAmount = maxAmount
		} else if optimalAmount < initialAmount {
			optimalAmount = initialAmount
		}

		optimal[txType] = optimalAmount
	}

	return optimal
}

// ApplyOptimalAmounts applies the calculated optimal amounts
func (k *Keeper) ApplyOptimalAmounts(evaluationPeriodHours uint32) map[types.TransactionType]uint64 {
	optimal := k.CalculateOptimalAmounts(evaluationPeriodHours)

	for txType, amount := range optimal {
		k.SetTypeAmount(txType, amount)
		k.lastScalingAdjustment[txType] = time.Unix(k.currentTime, 0)
	}

	return optimal
}
