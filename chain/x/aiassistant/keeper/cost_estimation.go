// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	storeprefix "cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// Pricing and accounting prefixes.
var (
	pricingStoreKey        = []byte{0x31}
	discountStoreKey       = []byte{0x32}
	historyStoreKey        = []byte{0x33}
	subscriptionStoreKey   = []byte{0x34}
	userSubscriptionKey    = []byte{0x35}
	freeTierStoreKey       = []byte{0x36}
	peakMultiplierStoreKey = []byte{0x37}
)

// ModelPricing captures configurable pricing knobs for a model.
type ModelPricing struct {
	ModelHash      string      `json:"model_hash"`
	BasePrice      sdkmath.Int `json:"base_price"`
	PerTokenPrice  sdkmath.Int `json:"per_token_price"`
	PerComputeUnit sdkmath.Int `json:"per_compute_unit"`
	MinimumCharge  sdkmath.Int `json:"minimum_charge"`
}

// CostDiscount allows operators to configure percentage discounts.
type CostDiscount struct {
	Address         string    `json:"address"`
	DiscountPercent uint64    `json:"discount_percent"`
	MinimumQueries  uint64    `json:"minimum_queries"`
	ValidUntil      time.Time `json:"valid_until"`
}

// CostBreakdown returns a transparent view into pricing components.
type CostBreakdown struct {
	BaseCost    sdkmath.Int `json:"base_cost"`
	ComputeCost sdkmath.Int `json:"compute_cost"`
	StorageCost sdkmath.Int `json:"storage_cost"`
	TotalCost   sdkmath.Int `json:"total_cost"`
}

// QueryCost is the public cost result returned to callers.
type QueryCost struct {
	Amount    sdkmath.Int   `json:"amount"`
	Breakdown CostBreakdown `json:"breakdown"`
}

// SubscriptionPlan models paid plans with included usage.
type SubscriptionPlan struct {
	Name            string      `json:"name"`
	MonthlyPrice    sdkmath.Int `json:"monthly_price"`
	IncludedQueries uint64      `json:"included_queries"`
	OverageRate     sdkmath.Int `json:"overage_rate"`
}

// FreeTier governs the complimentary tier.
type FreeTier struct {
	DailyQueries   uint64 `json:"daily_queries"`
	MonthlyQueries uint64 `json:"monthly_queries"`
	MaxInputLength uint64 `json:"max_input_length"`
}

func (k Keeper) pricingStore(ctx sdk.Context) storeprefix.Store {
	return storeprefix.NewStore(ctx.KVStore(k.storeKey), pricingStoreKey)
}

func (k Keeper) discountStore(ctx sdk.Context) storeprefix.Store {
	return storeprefix.NewStore(ctx.KVStore(k.storeKey), discountStoreKey)
}

func (k Keeper) historyStore(ctx sdk.Context) storeprefix.Store {
	return storeprefix.NewStore(ctx.KVStore(k.storeKey), historyStoreKey)
}

func (k Keeper) subscriptionStore(ctx sdk.Context) storeprefix.Store {
	return storeprefix.NewStore(ctx.KVStore(k.storeKey), subscriptionStoreKey)
}

func (k Keeper) userSubscriptionStore(ctx sdk.Context) storeprefix.Store {
	return storeprefix.NewStore(ctx.KVStore(k.storeKey), userSubscriptionKey)
}

func (k Keeper) freeTierStore(ctx sdk.Context) storeprefix.Store {
	return storeprefix.NewStore(ctx.KVStore(k.storeKey), freeTierStoreKey)
}

func defaultModelPricing(modelHash string) ModelPricing {
	return ModelPricing{
		ModelHash:      strings.ToLower(modelHash),
		BasePrice:      sdkmath.NewInt(1_000),
		PerTokenPrice:  sdkmath.NewInt(2),
		PerComputeUnit: sdkmath.NewInt(5),
		MinimumCharge:  sdkmath.NewInt(100),
	}
}

func (k Keeper) SetModelPricing(ctx sdk.Context, pricing ModelPricing) error {
	if pricing.ModelHash == "" {
		return fmt.Errorf("model hash cannot be empty")
	}
	if pricing.BasePrice.IsNil() {
		pricing.BasePrice = sdkmath.ZeroInt()
	}
	if pricing.PerTokenPrice.IsNil() {
		pricing.PerTokenPrice = sdkmath.ZeroInt()
	}
	if pricing.PerComputeUnit.IsNil() {
		pricing.PerComputeUnit = sdkmath.ZeroInt()
	}
	if pricing.MinimumCharge.IsNil() {
		pricing.MinimumCharge = sdkmath.ZeroInt()
	}
	if pricing.BasePrice.IsNegative() || pricing.PerTokenPrice.IsNegative() || pricing.PerComputeUnit.IsNegative() {
		return fmt.Errorf("pricing values must be non-negative")
	}
	store := k.pricingStore(ctx)
	bz, err := json.Marshal(pricing)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store.Set([]byte(strings.ToLower(pricing.ModelHash)), bz)
	return nil
}

func (k Keeper) GetModelPricing(ctx sdk.Context, modelHash string) (ModelPricing, bool) {
	store := k.pricingStore(ctx)
	bz := store.Get([]byte(strings.ToLower(modelHash)))
	if bz == nil {
		return ModelPricing{}, false
	}
	var pricing ModelPricing
	if err := json.Unmarshal(bz, &pricing); err != nil {
		return ModelPricing{}, false
	}
	return pricing, true
}

// EstimateComputeUnits returns a coarse compute estimation based on operation type and input length.
func (k Keeper) EstimateComputeUnits(_ sdk.Context, _ string, inputLength uint64, operationType string) uint64 {
	var base uint64
	switch strings.ToLower(operationType) {
	case "training":
		base = 500
	case "fine_tuning":
		base = 250
	case "embedding":
		base = 50
	case "analysis":
		base = 75
	case "inference":
		base = 25
	default:
		base = 20
	}
	// Deterministic integer division: sizeFactor = max(1, inputLength / 100)
	sizeFactor := inputLength / 100
	if sizeFactor < 1 {
		sizeFactor = 1
	}
	return base * sizeFactor
}

// EstimateQueryCost calculates the estimated cost for a model invocation.
func (k Keeper) EstimateQueryCost(ctx sdk.Context, modelHash string, inputLength uint64, operationType string) (QueryCost, error) {
	pricing, found := k.GetModelPricing(ctx, modelHash)
	if !found {
		pricing = defaultModelPricing(modelHash)
	}

	computeUnits := k.EstimateComputeUnits(ctx, modelHash, inputLength, operationType)

	baseCost := pricing.BasePrice
	tokenCost := pricing.PerTokenPrice.Mul(sdkmath.NewIntFromUint64(inputLength))
	computeCost := pricing.PerComputeUnit.Mul(sdkmath.NewIntFromUint64(computeUnits))

	total := baseCost.Add(tokenCost).Add(computeCost)

	if pricing.MinimumCharge.GT(sdkmath.ZeroInt()) && total.LT(pricing.MinimumCharge) {
		total = pricing.MinimumCharge
	}

	mult := k.getPeakHourMultiplier(ctx)
	if !mult.Equal(sdkmath.LegacyOneDec()) {
		total = mult.MulInt(total).RoundInt()
	}

	k.appendHistory(ctx, modelHash, total)

	breakdown := CostBreakdown{
		BaseCost:    baseCost,
		ComputeCost: computeCost,
		StorageCost: tokenCost,
		TotalCost:   total,
	}

	return QueryCost{
		Amount:    total,
		Breakdown: breakdown,
	}, nil
}

// CostEstimate represents the estimated cost for an AI operation
type CostEstimate struct {
	TokenCost       sdkmath.Int // Cost in native tokens
	ComputeUnits    uint64      // Estimated compute units
	ModelComplexity string      // Model complexity tier
	EstimatedTime   uint64      // Estimated time in milliseconds
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
	params, err := k.GetParams(ctx)
	if err != nil {
		return CostEstimate{}, err
	}

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
		return fmt.Errorf("error in DeductCost: %w", err)
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
	Timestamp     int64
	UserAddress   string
	Cost          CostEstimate
	OperationType AIOperationType
}

func (k Keeper) appendHistory(ctx sdk.Context, modelHash string, amount sdkmath.Int) {
	store := k.historyStore(ctx)
	key := []byte(strings.ToLower(modelHash))
	history := k.GetCostEstimationHistory(ctx, modelHash, 0)
	history = append(history, QueryCost{Amount: amount})
	if len(history) > 100 {
		history = history[len(history)-100:]
	}
	bz, err := json.Marshal(history)
	if err != nil {
		return
	}
	store.Set(key, bz)
}

func (k Keeper) GetCostEstimationHistory(ctx sdk.Context, modelHash string, limit uint64) []QueryCost {
	store := k.historyStore(ctx)
	bz := store.Get([]byte(strings.ToLower(modelHash)))
	if bz == nil {
		return []QueryCost{}
	}
	var history []QueryCost
	if err := json.Unmarshal(bz, &history); err != nil {
		return []QueryCost{}
	}
	if limit == 0 || uint64(len(history)) <= limit {
		return history
	}
	return history[len(history)-int(limit):]
}

func (k Keeper) SetCostDiscount(ctx sdk.Context, discount CostDiscount) error {
	if discount.Address == "" {
		return fmt.Errorf("discount address cannot be empty")
	}
	if discount.DiscountPercent > 100 {
		return fmt.Errorf("discount percent must be <= 100")
	}
	store := k.discountStore(ctx)
	bz, err := json.Marshal(discount)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store.Set([]byte(discount.Address), bz)
	return nil
}

func (k Keeper) ApplyDiscount(ctx sdk.Context, address string, amount sdkmath.Int) sdkmath.Int {
	store := k.discountStore(ctx)
	bz := store.Get([]byte(address))
	if bz == nil {
		return amount
	}
	var discount CostDiscount
	if err := json.Unmarshal(bz, &discount); err != nil {
		return amount
	}
	if !discount.ValidUntil.IsZero() && ctx.BlockTime().After(discount.ValidUntil) {
		return amount
	}
	if discount.DiscountPercent == 0 {
		return amount
	}
	discountAmt := amount.Mul(sdkmath.NewIntFromUint64(discount.DiscountPercent)).Quo(sdkmath.NewInt(100))
	final := amount.Sub(discountAmt)
	if final.IsNegative() {
		return sdkmath.ZeroInt()
	}
	return final
}

func (k Keeper) SetPeakHourMultiplier(ctx sdk.Context, multiplier float64) error {
	if multiplier <= 0 {
		return fmt.Errorf("multiplier must be positive")
	}
	decStr := fmt.Sprintf("%.4f", multiplier)
	dec, err := sdkmath.LegacyNewDecFromStr(decStr)
	if err != nil {
		return fmt.Errorf("failed to LegacyNewDecFromStr: %w", err)
	}
	store := ctx.KVStore(k.storeKey)
	bz, _ := dec.Marshal()
	store.Set(peakMultiplierStoreKey, bz)
	return nil
}

func (k Keeper) getPeakHourMultiplier(ctx sdk.Context) sdkmath.LegacyDec {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(peakMultiplierStoreKey)
	if bz == nil {
		return sdkmath.LegacyOneDec()
	}
	var dec sdkmath.LegacyDec
	if err := dec.Unmarshal(bz); err != nil {
		return sdkmath.LegacyOneDec()
	}
	return dec
}

func (k Keeper) CalculateRefund(_ sdk.Context, estimated, actual sdkmath.Int) sdkmath.Int {
	if estimated.GT(actual) {
		return estimated.Sub(actual)
	}
	return sdkmath.ZeroInt()
}

func (k Keeper) GetCostBreakdown(ctx sdk.Context, modelHash string, inputLength uint64, operationType string) CostBreakdown {
	pricing, found := k.GetModelPricing(ctx, modelHash)
	if !found {
		pricing = defaultModelPricing(modelHash)
	}
	computeUnits := k.EstimateComputeUnits(ctx, modelHash, inputLength, operationType)
	baseCost := pricing.BasePrice
	computeCost := pricing.PerComputeUnit.Mul(sdkmath.NewIntFromUint64(computeUnits))
	storageCost := pricing.PerTokenPrice.Mul(sdkmath.NewIntFromUint64(inputLength))
	total := baseCost.Add(computeCost).Add(storageCost)
	if pricing.MinimumCharge.GT(sdkmath.ZeroInt()) && total.LT(pricing.MinimumCharge) {
		total = pricing.MinimumCharge
	}
	return CostBreakdown{
		BaseCost:    baseCost,
		ComputeCost: computeCost,
		StorageCost: storageCost,
		TotalCost:   total,
	}
}

func (k Keeper) SetSubscriptionPlan(ctx sdk.Context, plan SubscriptionPlan) error {
	if plan.Name == "" {
		return fmt.Errorf("subscription name required")
	}
	store := k.subscriptionStore(ctx)
	bz, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store.Set([]byte(plan.Name), bz)
	return nil
}

func (k Keeper) AssignSubscription(ctx sdk.Context, userAddr, planName string) {
	store := k.userSubscriptionStore(ctx)
	store.Set([]byte(userAddr), []byte(planName))
}

func (k Keeper) SetFreeTier(ctx sdk.Context, tier FreeTier) error {
	store := k.freeTierStore(ctx)
	bz, err := json.Marshal(tier)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store.Set([]byte("config"), bz)
	return nil
}

func (k Keeper) QualifiesForFreeTier(ctx sdk.Context, _ string, inputLength uint64) bool {
	store := k.freeTierStore(ctx)
	bz := store.Get([]byte("config"))
	if bz == nil {
		return false
	}
	var tier FreeTier
	if err := json.Unmarshal(bz, &tier); err != nil {
		return false
	}
	if tier.MaxInputLength > 0 && inputLength > tier.MaxInputLength {
		return false
	}
	return tier.DailyQueries > 0 || tier.MonthlyQueries > 0
}
