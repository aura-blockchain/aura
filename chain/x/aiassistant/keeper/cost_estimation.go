package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// CostEstimate represents the estimated cost for an AI operation
type CostEstimate struct {
	TokenCost      sdkmath.Int // Cost in native tokens
	ComputeUnits   uint64      // Estimated compute units
	ModelComplexity string      // Model complexity tier
	EstimatedTime  uint64      // Estimated time in milliseconds
}

// AIOperationType defines types of AI operations
type AIOperationType int

const (
	OpTypeQuery AIOperationType = iota
	OpTypeGeneration
	OpTypeAnalysis
	OpTypeTraining
	OpTypeInference
)

// ModelComplexity defines AI model complexity tiers
const (
	ComplexityLow    = "low"
	ComplexityMedium = "medium"
	ComplexityHigh   = "high"
	ComplexityUltra  = "ultra"
)

// EstimateCost calculates the cost for an AI operation
func (k Keeper) EstimateCost(ctx sdk.Context, opType AIOperationType, inputSize uint64, modelHash string) (CostEstimate, error) {
	params := k.GetParams(ctx)

	// Get base cost from params
	baseCost := sdkmath.NewInt(1000000) // Base cost in smallest denomination

	// Calculate complexity multiplier
	complexity := k.determineModelComplexity(modelHash)
	multiplier := k.getComplexityMultiplier(complexity)

	// Calculate compute units based on input size and operation type
	computeUnits := k.calculateComputeUnits(opType, inputSize)

	// Calculate final token cost
	tokenCost := baseCost.Mul(sdkmath.NewInt(int64(multiplier))).Mul(sdkmath.NewInt(int64(computeUnits / 1000)))

	// Estimate processing time (in milliseconds)
	estimatedTime := k.estimateProcessingTime(computeUnits, complexity)

	estimate := CostEstimate{
		TokenCost:       tokenCost,
		ComputeUnits:    computeUnits,
		ModelComplexity: complexity,
		EstimatedTime:   estimatedTime,
	}

	// Validate against maximum cost limits
	if err := k.validateCostLimit(ctx, estimate, params); err != nil {
		return CostEstimate{}, err
	}

	return estimate, nil
}

// determineModelComplexity analyzes model hash to determine complexity
func (k Keeper) determineModelComplexity(modelHash string) string {
	// In production, this would query model registry
	// For now, simple heuristic based on hash
	if len(modelHash) == 0 {
		return ComplexityLow
	}

	// Simple hash-based complexity (placeholder)
	hashByte := modelHash[0]
	switch {
	case hashByte < '4':
		return ComplexityLow
	case hashByte < '8':
		return ComplexityMedium
	case hashByte < 'c':
		return ComplexityHigh
	default:
		return ComplexityUltra
	}
}

// getComplexityMultiplier returns cost multiplier for complexity tier
func (k Keeper) getComplexityMultiplier(complexity string) uint64 {
	switch complexity {
	case ComplexityLow:
		return 1
	case ComplexityMedium:
		return 3
	case ComplexityHigh:
		return 10
	case ComplexityUltra:
		return 30
	default:
		return 1
	}
}

// calculateComputeUnits estimates compute units for an operation
func (k Keeper) calculateComputeUnits(opType AIOperationType, inputSize uint64) uint64 {
	baseUnits := uint64(1000)

	// Operation type multiplier
	var opMultiplier uint64
	switch opType {
	case OpTypeQuery:
		opMultiplier = 1
	case OpTypeGeneration:
		opMultiplier = 5
	case OpTypeAnalysis:
		opMultiplier = 3
	case OpTypeTraining:
		opMultiplier = 100
	case OpTypeInference:
		opMultiplier = 2
	default:
		opMultiplier = 1
	}

	// Input size factor (logarithmic scaling)
	sizeFactor := uint64(1)
	if inputSize > 1000 {
		sizeFactor = inputSize / 1000
	}

	return baseUnits * opMultiplier * sizeFactor
}

// estimateProcessingTime estimates processing time in milliseconds
func (k Keeper) estimateProcessingTime(computeUnits uint64, complexity string) uint64 {
	baseTime := uint64(100) // 100ms base

	complexityFactor := k.getComplexityMultiplier(complexity)

	return baseTime * (computeUnits / 1000) * complexityFactor
}

// validateCostLimit validates cost against configured limits
func (k Keeper) validateCostLimit(ctx sdk.Context, estimate CostEstimate, params types.Params) error {
	// Maximum cost per operation (from params or hardcoded)
	maxCostPerOp := sdkmath.NewInt(1000000000) // 1000 tokens max

	if estimate.TokenCost.GT(maxCostPerOp) {
		return fmt.Errorf("operation cost %s exceeds maximum %s", estimate.TokenCost, maxCostPerOp)
	}

	// Maximum compute units
	maxComputeUnits := uint64(1000000)
	if estimate.ComputeUnits > maxComputeUnits {
		return fmt.Errorf("compute units %d exceed maximum %d", estimate.ComputeUnits, maxComputeUnits)
	}

	return nil
}

// DeductCost deducts the cost from user's account
func (k Keeper) DeductCost(ctx sdk.Context, userAddr sdk.AccAddress, estimate CostEstimate) error {
	denom := types.DefaultStakeDenom
	cost := sdk.NewCoin(denom, estimate.TokenCost)

	// Transfer cost to module account
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, userAddr, types.ModuleName, sdk.NewCoins(cost)); err != nil {
		return err
	}

	return nil
}

// GetCostHistory returns historical cost data for analytics
func (k Keeper) GetCostHistory(ctx sdk.Context, userAddr string, limit uint64) []CostRecord {
	// Placeholder - would query historical cost records from store
	return []CostRecord{}
}

// CostRecord represents a historical cost record
type CostRecord struct {
	Timestamp   int64
	UserAddress string
	Cost        CostEstimate
	OperationType AIOperationType
}
