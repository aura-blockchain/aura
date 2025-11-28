package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// GetGasPriceTracking retrieves gas price tracking data from the KV store
func (k Keeper) GetGasPriceTracking(ctx context.Context) (*types.GasPriceTracking, error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(GasPriceTrackingKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		// Return empty gas price tracking data
		return &types.GasPriceTracking{
			PriceHistory:   []types.GasPricePoint{},
			TrendDirection: "stable",
		}, nil
	}

	var tracking types.GasPriceTracking
	if err := json.Unmarshal(bz, &tracking); err != nil {
		return nil, err
	}

	return &tracking, nil
}

// TrackGasPrice records a gas price observation
func (k Keeper) TrackGasPrice(ctx context.Context, price uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if !params.EnableGasPriceTracking {
		return nil
	}

	if price == 0 {
		return types.ErrInvalidGasPrice
	}

	// Get current tracking data
	tracking, err := k.GetGasPriceTracking(ctx)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	// Add to price history
	pricePoint := types.GasPricePoint{
		Timestamp: blockTime,
		Price:     price,
	}
	tracking.PriceHistory = append(tracking.PriceHistory, pricePoint)

	// Keep only recent history
	if len(tracking.PriceHistory) > params.GasPriceHistorySize {
		tracking.PriceHistory = tracking.PriceHistory[len(tracking.PriceHistory)-params.GasPriceHistorySize:]
	}

	// Update current price
	tracking.CurrentPrice = price
	tracking.Timestamp = blockTime

	// Calculate statistics
	k.calculateGasPriceStats(tracking)

	// Persist to KV store
	if err := k.SetGasPriceTracking(ctx, tracking); err != nil {
		return err
	}

	// Update Prometheus metrics (non-consensus, observability only)
	if k.metrics != nil {
		k.metrics.CurrentGasPrice.Set(float64(price))
		k.metrics.AverageGasPrice.Set(float64(tracking.AveragePrice))
		k.metrics.GasPriceVolatility.Set(tracking.VolatilityScore)
	}

	// Check for price spikes
	if tracking.AveragePrice > 0 {
		ratio := float64(price) / float64(tracking.AveragePrice)
		if ratio >= params.GasPriceSpikeThreshold {
			if k.metrics != nil {
				k.metrics.GasPriceSpikes.Inc()
			}
			if params.EnableAlerts {
				if err := k.createGasPriceSpikeAlert(ctx, price, ratio); err != nil {
					// Log error but don't fail the update
					return nil
				}
			}
		}
	}

	return nil
}

// calculateGasPriceStats calculates gas price statistics
func (k Keeper) calculateGasPriceStats(tracking *types.GasPriceTracking) {
	if len(tracking.PriceHistory) == 0 {
		return
	}

	var sum uint64
	var prices []uint64
	minPrice := tracking.PriceHistory[0].Price
	maxPrice := tracking.PriceHistory[0].Price

	for _, point := range tracking.PriceHistory {
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
	avgPrice := sum / uint64(len(tracking.PriceHistory))
	tracking.AveragePrice = avgPrice

	// Median
	sort.Slice(prices, func(i, j int) bool {
		return prices[i] < prices[j]
	})
	medianIdx := len(prices) / 2
	if len(prices)%2 == 0 && medianIdx > 0 {
		tracking.MedianPrice = (prices[medianIdx-1] + prices[medianIdx]) / 2
	} else {
		tracking.MedianPrice = prices[medianIdx]
	}

	tracking.MinPrice = minPrice
	tracking.MaxPrice = maxPrice

	// Volatility (coefficient of variation)
	var variance float64
	for _, point := range tracking.PriceHistory {
		diff := float64(point.Price) - float64(avgPrice)
		variance += diff * diff
	}
	variance /= float64(len(tracking.PriceHistory))
	stdDev := math.Sqrt(variance)

	if avgPrice > 0 {
		tracking.VolatilityScore = stdDev / float64(avgPrice)
	}

	// Trend direction
	tracking.TrendDirection = k.calculateGasPriceTrend(tracking.PriceHistory)
}

// calculateGasPriceTrend calculates the price trend direction
func (k Keeper) calculateGasPriceTrend(history []types.GasPricePoint) string {
	historyLen := len(history)
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
		recentSum += history[i].Price
		recentCount++
	}

	for i := 0; i < recentLen && i < historyLen-recentLen; i++ {
		olderSum += history[i].Price
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

// createGasPriceSpikeAlert creates an alert for gas price spikes
func (k Keeper) createGasPriceSpikeAlert(ctx context.Context, price uint64, ratio float64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	tracking, err := k.GetGasPriceTracking(ctx)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	alert := &types.Alert{
		ID:       k.generateID(ctx, "alert-gas-spike"),
		Type:     types.AlertTypeGasPriceSpike,
		Severity: k.determineGasSpikeSeverity(ratio),
		Message:  fmt.Sprintf("Gas price spike detected: %.2fx above average", ratio),
		Details: map[string]interface{}{
			"current_price": price,
			"average_price": tracking.AveragePrice,
			"ratio":         ratio,
			"threshold":     params.GasPriceSpikeThreshold,
		},
		Timestamp:        blockTime,
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	// Store alert in KV store
	if err := k.SetAlert(ctx, alert); err != nil {
		return err
	}

	// Update metrics
	if k.metrics != nil {
		k.metrics.TotalAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
		k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
	}

	return nil
}

// determineGasSpikeSeverity determines severity based on spike ratio
func (k Keeper) determineGasSpikeSeverity(ratio float64) types.AlertSeverity {
	if ratio >= 5.0 {
		return types.SeverityCritical
	} else if ratio >= 3.0 {
		return types.SeverityHigh
	} else if ratio >= 2.0 {
		return types.SeverityMedium
	}
	return types.SeverityLow
}
