package keeper

import (
	"context"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// GAS PRICE PREDICTION (Feature 5)
// ============================

// PredictGasPrice predicts future gas price based on historical data
// Uses simple linear regression to extrapolate from recent utilization trends
// Returns predicted price, confidence level (0-10000 basis points), and any error
//
// All calculations use deterministic integer arithmetic to ensure consensus safety.
func (k *Keeper) PredictGasPrice(ctx context.Context, blocksAhead uint64) (string, uint64, error) {
	params, _ := k.GetParams(ctx)

	if !params.DynamicFees.Enabled {
		return params.DynamicFees.BaseFee, 0, nil
	}

	// Get historical utilization data
	utilizationData := params.DynamicFees.RecentUtilization
	if len(utilizationData) == 0 {
		return params.DynamicFees.BaseFee, 0, nil
	}

	// Calculate trend using simple linear regression with integer arithmetic
	// All intermediate values are scaled by PRECISION to maintain accuracy
	const PRECISION = int64(1000000) // Scale factor for fixed-point arithmetic

	n := int64(len(utilizationData))
	sumX := int64(0)
	sumY := int64(0)
	sumXY := int64(0)
	sumX2 := int64(0)

	for i, util := range utilizationData {
		x := int64(i)
		y := int64(util)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (trend) and intercept using integer arithmetic
	// slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX^2)
	// intercept = (sumY - slope*sumX) / n
	// Scale numerator by PRECISION before division to maintain accuracy
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		// No variance in data, use current price
		return k.calculateCurrentGasPrice(params), 5000, nil
	}

	// slope = ((n*sumXY - sumX*sumY) * PRECISION) / denominator
	slopeNumerator := (n*sumXY - sumX*sumY) * PRECISION
	slope := slopeNumerator / denominator

	// intercept = ((sumY * PRECISION) / n) - (slope * sumX) / n
	intercept := (sumY*PRECISION)/n - (slope*sumX)/n

	// Predict utilization for blocksAhead
	// predictedUtilization = intercept + slope * (n + blocksAhead)
	futureX := n + int64(blocksAhead)
	predictedUtilizationScaled := intercept + (slope*futureX)/PRECISION

	// Clamp to valid range [0, 10000]
	predictedUtilization := predictedUtilizationScaled
	if predictedUtilization < 0 {
		predictedUtilization = 0
	}
	if predictedUtilization > 10000 {
		predictedUtilization = 10000
	}

	// Calculate predicted multiplier based on target utilization
	targetUtilization := int64(params.DynamicFees.TargetUtilization)
	deviation := predictedUtilization - targetUtilization

	predictedMultiplier := int64(params.DynamicFees.CurrentMultiplier)
	if deviation > 0 {
		// Higher utilization -> higher fees
		// adjustment = (deviation * adjustmentSpeed) / BasisPoints
		adjustment := (deviation * int64(params.DynamicFees.AdjustmentSpeed)) / int64(types.BasisPoints)
		predictedMultiplier += adjustment
	} else {
		// Lower utilization -> lower fees
		// Use absolute value without float math
		absDeviation := -deviation
		if absDeviation < 0 {
			absDeviation = deviation
		}
		adjustment := (absDeviation * int64(params.DynamicFees.AdjustmentSpeed)) / int64(types.BasisPoints)
		predictedMultiplier -= adjustment
	}

	// Clamp to min/max
	if predictedMultiplier < int64(params.DynamicFees.MinMultiplier) {
		predictedMultiplier = int64(params.DynamicFees.MinMultiplier)
	}
	if predictedMultiplier > int64(params.DynamicFees.MaxMultiplier) {
		predictedMultiplier = int64(params.DynamicFees.MaxMultiplier)
	}

	// Calculate predicted gas price
	baseFee := new(big.Int)
	if _, ok := baseFee.SetString(params.DynamicFees.BaseFee, 10); !ok {
		return params.DynamicFees.BaseFee, 0, types.ErrInvalidAmount
	}

	predictedFee := new(big.Int).Mul(baseFee, big.NewInt(predictedMultiplier))
	predictedFee.Div(predictedFee, big.NewInt(types.BasisPoints))

	confidence := k.calculatePredictionConfidence(utilizationData)

	return predictedFee.String(), confidence, nil
}

// calculateCurrentGasPrice calculates the current gas price from params
func (k *Keeper) calculateCurrentGasPrice(params types.Params) string {
	baseFee := new(big.Int)
	if _, ok := baseFee.SetString(params.DynamicFees.BaseFee, 10); !ok {
		return params.DynamicFees.BaseFee
	}

	currentFee := new(big.Int).Mul(baseFee, big.NewInt(int64(params.DynamicFees.CurrentMultiplier)))
	currentFee.Div(currentFee, big.NewInt(types.BasisPoints))

	return currentFee.String()
}

// calculatePredictionConfidence calculates confidence level (0-10000 basis points)
// Lower variance in historical data = higher confidence
//
// Uses deterministic integer arithmetic with Newton-Raphson square root approximation.
func (k *Keeper) calculatePredictionConfidence(data []uint64) uint64 {
	if len(data) < 2 {
		return 0 // No confidence with insufficient data
	}

	// Calculate mean using integer arithmetic
	sum := int64(0)
	for _, val := range data {
		sum += int64(val)
	}
	mean := sum / int64(len(data))

	// Calculate variance using integer arithmetic
	varianceSum := int64(0)
	for _, val := range data {
		diff := int64(val) - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / int64(len(data))

	// Calculate standard deviation using integer square root
	stdDev := intSqrt(variance)

	// Lower variance = higher confidence
	// Normalize to 0-10000 basis points
	// Assume maximum expected standard deviation is 5000 (half the range)
	maxStdDev := int64(5000)

	// Confidence decreases linearly with standard deviation
	// confidence = 10000 - (stdDev * 10000) / maxStdDev
	confidenceReduction := (stdDev * 10000) / maxStdDev
	if confidenceReduction > 10000 {
		confidenceReduction = 10000
	}
	confidence := uint64(10000 - confidenceReduction)

	// Ensure confidence is in valid range
	if confidence > 10000 {
		confidence = 10000
	}

	// Bonus confidence for more data points
	dataBonus := uint64(len(data)) * 10
	if dataBonus > 1000 {
		dataBonus = 1000
	}
	confidence += dataBonus
	if confidence > 10000 {
		confidence = 10000
	}

	return confidence
}

// intSqrt computes integer square root using Newton-Raphson method
// This is deterministic and consensus-safe, unlike math.Sqrt which uses float64
func intSqrt(n int64) int64 {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}

	// Newton-Raphson: x_{n+1} = (x_n + n/x_n) / 2
	x := n
	for {
		x1 := (x + n/x) / 2
		if x1 >= x {
			return x
		}
		x = x1
	}
}

// GetGasPredictionStatistics returns gas prediction statistics for different time horizons
// Returns predictions for 1, 10, and 100 blocks ahead along with average confidence
func (k *Keeper) GetGasPredictionStatistics(ctx context.Context) (price1, price10, price100 string, avgConfidence uint64, err error) {
	// Predict for 1, 10, and 100 blocks ahead
	price1, conf1, err1 := k.PredictGasPrice(ctx, 1)
	if err1 != nil {
		return "", "", "", 0, err1
	}

	price10, conf10, err10 := k.PredictGasPrice(ctx, 10)
	if err10 != nil {
		return "", "", "", 0, err10
	}

	price100, conf100, err100 := k.PredictGasPrice(ctx, 100)
	if err100 != nil {
		return "", "", "", 0, err100
	}

	// Average confidence
	avgConfidence = (conf1 + conf10 + conf100) / 3

	return price1, price10, price100, avgConfidence, nil
}

// GetRecommendedGasPrice returns recommended gas price for different priority levels
// Priority levels: "low", "medium", "high", "urgent"
// Each level adjusts the base price by a different multiplier
func (k *Keeper) GetRecommendedGasPrice(ctx context.Context, priority string) (string, error) {
	params, _ := k.GetParams(ctx)

	baseFee := new(big.Int)
	if _, ok := baseFee.SetString(params.DynamicFees.BaseFee, 10); !ok {
		return "", types.ErrInvalidAmount
	}

	multiplier := params.DynamicFees.CurrentMultiplier

	switch priority {
	case "low":
		// 90% of current price - for non-urgent transactions
		multiplier = (multiplier * 90) / 100
	case "medium":
		// Current price - standard priority
		// multiplier stays the same
	case "high":
		// 110% of current price - for important transactions
		multiplier = (multiplier * 110) / 100
	case "urgent":
		// 150% of current price - for critical transactions
		multiplier = (multiplier * 150) / 100
	default:
		return "", types.ErrInvalidPriority
	}

	// Ensure multiplier doesn't exceed max
	if multiplier > params.DynamicFees.MaxMultiplier {
		multiplier = params.DynamicFees.MaxMultiplier
	}
	// Ensure multiplier doesn't go below min
	if multiplier < params.DynamicFees.MinMultiplier {
		multiplier = params.DynamicFees.MinMultiplier
	}

	fee := new(big.Int).Mul(baseFee, big.NewInt(int64(multiplier)))
	fee.Div(fee, big.NewInt(types.BasisPoints))

	return fee.String(), nil
}

// EstimateTransactionCost estimates the total cost for a transaction
// Returns the estimated gas cost based on current and predicted prices
func (k *Keeper) EstimateTransactionCost(ctx context.Context, estimatedGasUsage uint64, blocksUntilSubmission uint64) (string, error) {
	// Get predicted gas price for when transaction will be submitted
	predictedPrice, confidence, err := k.PredictGasPrice(ctx, blocksUntilSubmission)
	if err != nil {
		return "", err
	}

	// If confidence is low, add safety margin
	safetyMultiplier := uint64(10000) // 100% (in basis points)
	if confidence < 5000 {            // Less than 50% confidence
		safetyMultiplier = 11000 // Add 10% safety margin
	}

	price := new(big.Int)
	if _, ok := price.SetString(predictedPrice, 10); !ok {
		return "", types.ErrInvalidAmount
	}

	// Calculate total cost: price * gasUsage * safetyMultiplier
	totalCost := new(big.Int).Mul(price, big.NewInt(int64(estimatedGasUsage)))
	totalCost.Mul(totalCost, big.NewInt(int64(safetyMultiplier)))
	totalCost.Div(totalCost, big.NewInt(types.BasisPoints))

	return totalCost.String(), nil
}

// GetGasPriceTrend analyzes the trend of gas prices
// Returns: "increasing", "decreasing", "stable" along with the trend strength (0-10000 basis points)
func (k *Keeper) GetGasPriceTrend(ctx context.Context) (direction string, strength uint64) {
	params, _ := k.GetParams(ctx)

	utilizationData := params.DynamicFees.RecentUtilization
	if len(utilizationData) < 3 {
		return "stable", 0
	}

	// Calculate simple moving average of first half vs second half
	halfPoint := len(utilizationData) / 2
	firstHalfSum := uint64(0)
	secondHalfSum := uint64(0)

	for i := 0; i < halfPoint; i++ {
		firstHalfSum += utilizationData[i]
	}
	for i := halfPoint; i < len(utilizationData); i++ {
		secondHalfSum += utilizationData[i]
	}

	firstHalfAvg := firstHalfSum / uint64(halfPoint)
	secondHalfAvg := secondHalfSum / uint64(len(utilizationData)-halfPoint)

	// Calculate difference
	var diff uint64
	if secondHalfAvg > firstHalfAvg {
		diff = secondHalfAvg - firstHalfAvg
		direction = "increasing"
	} else if firstHalfAvg > secondHalfAvg {
		diff = firstHalfAvg - secondHalfAvg
		direction = "decreasing"
	} else {
		direction = "stable"
		return direction, 0
	}

	// Calculate strength as percentage of range
	// Strength is normalized to 0-10000 basis points
	strength = (diff * types.BasisPoints) / 10000
	if strength > types.BasisPoints {
		strength = types.BasisPoints
	}

	return direction, strength
}

// GetOptimalSubmissionTime suggests the optimal number of blocks to wait before submitting
// a transaction to minimize gas costs. Returns recommended blocks to wait.
func (k *Keeper) GetOptimalSubmissionTime(ctx context.Context, maxBlocksToWait uint64) (uint64, string, error) {
	if maxBlocksToWait == 0 {
		maxBlocksToWait = 100 // Default maximum
	}
	if maxBlocksToWait > 1000 {
		maxBlocksToWait = 1000 // Cap at reasonable limit
	}

	lowestPrice := ""
	lowestPriceBig := new(big.Int)
	optimalBlocks := uint64(0)

	// Check price predictions for each block up to max
	for i := uint64(0); i <= maxBlocksToWait; i++ {
		price, _, err := k.PredictGasPrice(ctx, i)
		if err != nil {
			continue
		}

		priceBig := new(big.Int)
		if _, ok := priceBig.SetString(price, 10); !ok {
			continue
		}

		if lowestPrice == "" || priceBig.Cmp(lowestPriceBig) < 0 {
			lowestPrice = price
			lowestPriceBig = priceBig
			optimalBlocks = i
		}
	}

	if lowestPrice == "" {
		// Fallback to current price
		params, _ := k.GetParams(ctx)
		lowestPrice = k.calculateCurrentGasPrice(params)
		optimalBlocks = 0
	}

	return optimalBlocks, lowestPrice, nil
}
