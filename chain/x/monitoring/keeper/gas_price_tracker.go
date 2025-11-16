package keeper

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// gasPriceWorker periodically tracks gas prices
func (k *Keeper) gasPriceWorker() {
	defer k.wg.Done()
	ticker := time.NewTicker(k.params.GasPriceCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-ticker.C:
			k.updateGasPriceMetrics()
		}
	}
}

// TrackGasPrice records a gas price observation
func (k *Keeper) TrackGasPrice(price uint64) error {
	if !k.params.EnableGasPriceTracking {
		return nil
	}

	if price == 0 {
		return types.ErrInvalidGasPrice
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()

	// Add to price history
	pricePoint := types.GasPricePoint{
		Timestamp: now,
		Price:     price,
	}
	k.gasPriceTracking.PriceHistory = append(k.gasPriceTracking.PriceHistory, pricePoint)

	// Keep only recent history
	if len(k.gasPriceTracking.PriceHistory) > k.params.GasPriceHistorySize {
		k.gasPriceTracking.PriceHistory = k.gasPriceTracking.PriceHistory[1:]
	}

	// Update current price
	k.gasPriceTracking.CurrentPrice = price
	k.gasPriceTracking.Timestamp = now

	// Calculate statistics
	k.calculateGasPriceStats()

	// Update Prometheus metrics
	k.metrics.CurrentGasPrice.Set(float64(price))
	k.metrics.AverageGasPrice.Set(float64(k.gasPriceTracking.AveragePrice))
	k.metrics.GasPriceVolatility.Set(k.gasPriceTracking.VolatilityScore)

	// Check for price spikes
	if k.gasPriceTracking.AveragePrice > 0 {
		ratio := float64(price) / float64(k.gasPriceTracking.AveragePrice)
		if ratio >= k.params.GasPriceSpikeThreshold {
			k.metrics.GasPriceSpikes.Inc()
			if k.params.EnableAlerts {
				k.createGasPriceSpikeAlert(price, ratio)
			}
		}
	}

	return nil
}

// calculateGasPriceStats calculates gas price statistics
func (k *Keeper) calculateGasPriceStats() {
	if len(k.gasPriceTracking.PriceHistory) == 0 {
		return
	}

	var sum uint64
	var prices []uint64
	minPrice := k.gasPriceTracking.PriceHistory[0].Price
	maxPrice := k.gasPriceTracking.PriceHistory[0].Price

	for _, point := range k.gasPriceTracking.PriceHistory {
		sum += point.Price
		prices = append(prices, point.Price)

		if point.Price < minPrice {
			minPrice = point.Price
		}
		if point.Price > maxPrice {
			maxPrice = point.Price
		}
	}

	// Average
	avgPrice := sum / uint64(len(k.gasPriceTracking.PriceHistory))
	k.gasPriceTracking.AveragePrice = avgPrice

	// Median
	sort.Slice(prices, func(i, j int) bool {
		return prices[i] < prices[j]
	})
	medianIdx := len(prices) / 2
	if len(prices)%2 == 0 {
		k.gasPriceTracking.MedianPrice = (prices[medianIdx-1] + prices[medianIdx]) / 2
	} else {
		k.gasPriceTracking.MedianPrice = prices[medianIdx]
	}

	k.gasPriceTracking.MinPrice = minPrice
	k.gasPriceTracking.MaxPrice = maxPrice

	// Volatility (coefficient of variation)
	var variance float64
	for _, point := range k.gasPriceTracking.PriceHistory {
		diff := float64(point.Price) - float64(avgPrice)
		variance += diff * diff
	}
	variance /= float64(len(k.gasPriceTracking.PriceHistory))
	stdDev := math.Sqrt(variance)

	if avgPrice > 0 {
		k.gasPriceTracking.VolatilityScore = stdDev / float64(avgPrice)
	}

	// Trend direction
	k.gasPriceTracking.TrendDirection = k.calculateTrend()
}

// calculateTrend calculates the price trend direction
func (k *Keeper) calculateTrend() string {
	historyLen := len(k.gasPriceTracking.PriceHistory)
	if historyLen < 10 {
		return "stable"
	}

	// Compare recent prices with older prices
	recentLen := historyLen / 4
	if recentLen < 5 {
		recentLen = 5
	}

	var recentSum, olderSum uint64
	recentCount := 0
	olderCount := 0

	for i := historyLen - recentLen; i < historyLen; i++ {
		recentSum += k.gasPriceTracking.PriceHistory[i].Price
		recentCount++
	}

	for i := 0; i < recentLen && i < historyLen-recentLen; i++ {
		olderSum += k.gasPriceTracking.PriceHistory[i].Price
		olderCount++
	}

	if olderCount == 0 || recentCount == 0 {
		return "stable"
	}

	recentAvg := float64(recentSum) / float64(recentCount)
	olderAvg := float64(olderSum) / float64(olderCount)

	changePercent := (recentAvg - olderAvg) / olderAvg

	if changePercent > 0.1 {
		return "increasing"
	} else if changePercent < -0.1 {
		return "decreasing"
	}
	return "stable"
}

// GetGasPriceTracking returns current gas price tracking data
func (k *Keeper) GetGasPriceTracking() *types.GasPriceTracking {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.gasPriceTracking
}

// updateGasPriceMetrics updates gas price metrics
func (k *Keeper) updateGasPriceMetrics() {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.gasPriceTracking != nil {
		k.metrics.CurrentGasPrice.Set(float64(k.gasPriceTracking.CurrentPrice))
		k.metrics.AverageGasPrice.Set(float64(k.gasPriceTracking.AveragePrice))
		k.metrics.GasPriceVolatility.Set(k.gasPriceTracking.VolatilityScore)
	}
}

// createGasPriceSpikeAlert creates an alert for gas price spikes
func (k *Keeper) createGasPriceSpikeAlert(price uint64, ratio float64) {
	alert := &types.Alert{
		ID:       generateID("alert-gas-spike"),
		Type:     types.AlertTypeGasPriceSpike,
		Severity: k.determineGasSpikeSeverity(ratio),
		Message:  fmt.Sprintf("Gas price spike detected: %.2fx above average", ratio),
		Details: map[string]interface{}{
			"current_price": price,
			"average_price": k.gasPriceTracking.AveragePrice,
			"ratio":         ratio,
			"threshold":     k.params.GasPriceSpikeThreshold,
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

// determineGasSpikeSeverity determines severity based on spike ratio
func (k *Keeper) determineGasSpikeSeverity(ratio float64) types.AlertSeverity {
	if ratio >= 5.0 {
		return types.SeverityCritical
	} else if ratio >= 3.0 {
		return types.SeverityHigh
	} else if ratio >= 2.0 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}
