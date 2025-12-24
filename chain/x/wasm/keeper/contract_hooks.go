package keeper

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/wasm/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// CIRCUIT BREAKER PATTERN (KV Store-based for Consensus)
// ============================================================================

// Circuit breaker KV store keys
var (
	circuitBreakerFailureCountKey       = []byte("circuit_breaker_failure_count")
	circuitBreakerLastFailureKey        = []byte("circuit_breaker_last_failure")
	circuitBreakerStateKey              = []byte("circuit_breaker_state")
	circuitBreakerConsecutiveSuccessKey = []byte("circuit_breaker_consecutive_success")
)

func clampUint64ToInt(u uint64) int {
	if u > math.MaxInt {
		return math.MaxInt
	}
	return int(u)
}

func clampIntToUint64(v int) uint64 {
	if v < 0 {
		return 0
	}
	if v > math.MaxInt {
		return math.MaxUint64
	}
	return uint64(v)
}

const (
	circuitBreakerThreshold        = 5  // failures before opening
	circuitBreakerTimeout          = 60 // seconds before attempting half-open
	circuitBreakerSuccessThreshold = 3  // successes before closing from half-open
)

// Circuit breaker states
const (
	circuitBreakerStateClosed   = "closed"
	circuitBreakerStateOpen     = "open"
	circuitBreakerStateHalfOpen = "half-open"
)

// circuitBreakerData represents the circuit breaker state (stored in KV store)
type circuitBreakerData struct {
	FailureCount       int
	LastFailure        time.Time
	State              string
	ConsecutiveSuccess int
}

// getCircuitBreakerState reads circuit breaker state from KV store
func (k Keeper) getCircuitBreakerState(ctx sdk.Context) circuitBreakerData {
	store := ctx.KVStore(k.storeKey)

	// Read failure count
	failureCount := 0
	if bz := store.Get(circuitBreakerFailureCountKey); bz != nil {
		failureCount = clampUint64ToInt(binary.BigEndian.Uint64(bz))
	}

	// Read last failure time
	lastFailure := time.Time{}
	if bz := store.Get(circuitBreakerLastFailureKey); bz != nil {
		if err := lastFailure.UnmarshalBinary(bz); err != nil {
			k.Logger(ctx).Error("failed to unmarshal last failure time", "error", err)
		}
	}

	// Read state
	state := circuitBreakerStateClosed
	if bz := store.Get(circuitBreakerStateKey); bz != nil {
		state = string(bz)
	}

	// Read consecutive success count
	consecutiveSuccess := 0
	if bz := store.Get(circuitBreakerConsecutiveSuccessKey); bz != nil {
		consecutiveSuccess = clampUint64ToInt(binary.BigEndian.Uint64(bz))
	}

	return circuitBreakerData{
		FailureCount:       failureCount,
		LastFailure:        lastFailure,
		State:              state,
		ConsecutiveSuccess: consecutiveSuccess,
	}
}

// setCircuitBreakerState writes circuit breaker state to KV store
func (k Keeper) setCircuitBreakerState(ctx sdk.Context, data circuitBreakerData) {
	store := ctx.KVStore(k.storeKey)

	// Write failure count
	failureCountBz := make([]byte, 8)
	binary.BigEndian.PutUint64(failureCountBz, clampIntToUint64(data.FailureCount))
	store.Set(circuitBreakerFailureCountKey, failureCountBz)

	// Write last failure time
	if !data.LastFailure.IsZero() {
		lastFailureBz, err := data.LastFailure.MarshalBinary()
		if err != nil {
			k.Logger(ctx).Error("failed to marshal last failure time", "error", err)
		} else {
			store.Set(circuitBreakerLastFailureKey, lastFailureBz)
		}
	}

	// Write state
	store.Set(circuitBreakerStateKey, []byte(data.State))

	// Write consecutive success count
	consecutiveSuccessBz := make([]byte, 8)
	binary.BigEndian.PutUint64(consecutiveSuccessBz, clampIntToUint64(data.ConsecutiveSuccess))
	store.Set(circuitBreakerConsecutiveSuccessKey, consecutiveSuccessBz)
}

// shouldSkipCircuitBreaker returns true if we should skip registry calls
// Now deterministic: reads from KV store and uses block time
func (k Keeper) shouldSkipCircuitBreaker(ctx sdk.Context) bool {
	data := k.getCircuitBreakerState(ctx)

	// If closed, never skip
	if data.State == circuitBreakerStateClosed {
		return false
	}

	// If open, check if timeout has elapsed (use block time for determinism)
	if data.State == circuitBreakerStateOpen {
		if !data.LastFailure.IsZero() {
			elapsed := ctx.BlockTime().Sub(data.LastFailure)
			if elapsed.Seconds() >= circuitBreakerTimeout {
				// Transition to half-open
				data.State = circuitBreakerStateHalfOpen
				data.ConsecutiveSuccess = 0
				k.setCircuitBreakerState(ctx, data)
				return false // Try again in half-open state
			}
		}
		return true // Still open, skip
	}

	// Half-open: don't skip (allow attempts)
	return false
}

// recordCircuitBreakerSuccess records a successful registry call
func (k Keeper) recordCircuitBreakerSuccess(ctx sdk.Context) {
	data := k.getCircuitBreakerState(ctx)

	data.ConsecutiveSuccess++
	data.FailureCount = 0

	if data.State == circuitBreakerStateHalfOpen && data.ConsecutiveSuccess >= circuitBreakerSuccessThreshold {
		data.State = circuitBreakerStateClosed
		data.ConsecutiveSuccess = 0
	}

	k.setCircuitBreakerState(ctx, data)
}

// recordCircuitBreakerFailure records a failed registry call
func (k Keeper) recordCircuitBreakerFailure(ctx sdk.Context) {
	data := k.getCircuitBreakerState(ctx)

	data.FailureCount++
	data.LastFailure = ctx.BlockTime() // Use block time for determinism
	data.ConsecutiveSuccess = 0

	if data.State == circuitBreakerStateHalfOpen {
		data.State = circuitBreakerStateOpen
	} else if data.FailureCount >= circuitBreakerThreshold {
		data.State = circuitBreakerStateOpen
	}

	k.setCircuitBreakerState(ctx, data)
}

// getCircuitBreakerStateString returns the current circuit breaker state as a string
func (k Keeper) getCircuitBreakerStateString(ctx sdk.Context) string {
	data := k.getCircuitBreakerState(ctx)
	return data.State
}

// ============================================================================
// VALIDATION CACHE (Performance Optimization)
// ============================================================================

type validationCacheEntry struct {
	allowed     bool
	error       error
	timestamp   time.Time
	blockHeight int64
}

type validationCache struct {
	mu      sync.RWMutex
	entries map[string]*validationCacheEntry
}

var (
	valCache = &validationCache{
		entries: make(map[string]*validationCacheEntry),
	}
)

// getCacheKey generates a cache key for validation
func getCacheKey(contractAddr, sender string, blockHeight int64) string {
	return fmt.Sprintf("%s:%s:%d", contractAddr, sender, blockHeight)
}

// get retrieves a cached validation result
// FIXED: Use block height only for cache validity (not wall-clock time)
func (vc *validationCache) get(key string, blockHeight int64) (*validationCacheEntry, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	entry, found := vc.entries[key]
	if !found {
		return nil, false
	}

	// Invalidate if from a different block only (removed non-deterministic time.Since check)
	// Cache is valid for same block height only - deterministic across validators
	if entry.blockHeight != blockHeight {
		return nil, false
	}

	return entry, true
}

// set stores a validation result
func (vc *validationCache) set(key string, entry *validationCacheEntry) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.entries[key] = entry
}

// cleanup removes old entries
func (vc *validationCache) cleanup(currentHeight int64) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	for key, entry := range vc.entries {
		if entry.blockHeight < currentHeight-1 {
			delete(vc.entries, key)
		}
	}
}

// ============================================================================
// METRICS BUFFER (Batch Updates)
// ============================================================================

type metricsUpdate struct {
	contractAddr string
	gasUsed      uint64
	success      bool
	timestamp    time.Time
}

type metricsBuffer struct {
	mu      sync.Mutex
	updates []metricsUpdate
}

var (
	metricsBuf = &metricsBuffer{
		updates: make([]metricsUpdate, 0, 100),
	}
)

const (
	metricsBufferSize     = 50 // flush after this many updates
	metricsBufferDuration = 10 // flush after this many seconds
)

// add adds a metrics update to the buffer
func (mb *metricsBuffer) add(update metricsUpdate) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.updates = append(mb.updates, update)
}

// shouldFlush returns true if buffer should be flushed
func (mb *metricsBuffer) shouldFlush() bool {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	return len(mb.updates) >= metricsBufferSize
}

// flush returns all pending updates and clears the buffer
func (mb *metricsBuffer) flush() []metricsUpdate {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	updates := make([]metricsUpdate, len(mb.updates))
	copy(updates, mb.updates)
	mb.updates = mb.updates[:0] // clear
	return updates
}

// ============================================================================
// HOOK METHODS
// ============================================================================

// BeforeInstantiateHook is called before instantiating a contract
// Auto-registers the contract in the registry with default settings
func (k Keeper) BeforeInstantiateHook(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	admin sdk.AccAddress,
	label string,
) error {
	// Skip if registry not available or circuit breaker is open
	if k.contractRegistry == nil {
		k.Logger(ctx).Debug("contract registry not available, skipping auto-registration")
		return nil
	}

	if k.shouldSkipCircuitBreaker(ctx) {
		k.Logger(ctx).Warn("circuit breaker open, skipping registry auto-registration",
			"state", k.getCircuitBreakerStateString(ctx))
		// Emit alert event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"contract_registry_degraded",
				sdk.NewAttribute("circuit_breaker_state", k.getCircuitBreakerStateString(ctx)),
				sdk.NewAttribute("operation", "auto_register"),
			),
		)
		return nil // Graceful degradation - don't block instantiation
	}

	// Rate limit registration attempts per creator
	if err := k.checkRegistrationRateLimit(ctx, creator.String()); err != nil {
		k.Logger(ctx).Error("registration rate limit exceeded", "creator", creator.String())
		return types.ErrUnauthorized.Wrapf("registration rate limit exceeded for creator %s", creator.String())
	}

	startTime := determinism.GetBlockTime(ctx)

	// Validate that the creator is eligible (check contract limit)
	params, _ := k.contractRegistry.GetParams(ctx)
	if params.MaxContractsPerCreator > 0 {
		// Use GetCreatorContracts to count existing contracts
		contracts := k.contractRegistry.GetCreatorContracts(ctx, creator.String())
		count := uint64(len(contracts))
		if count >= params.MaxContractsPerCreator {
			k.Logger(ctx).Error("creator contract limit exceeded",
				"creator", creator.String(),
				"current", count,
				"max", params.MaxContractsPerCreator)
			k.recordCircuitBreakerFailure(ctx)
			return types.ErrUnauthorized.Wrapf("creator contract limit exceeded: %d >= %d", count, params.MaxContractsPerCreator)
		}
	}

	// Record success
	k.recordCircuitBreakerSuccess(ctx)

	// Log metrics
	elapsed := time.Since(startTime)
	k.Logger(ctx).Debug("before instantiate hook completed",
		"code_id", codeID,
		"creator", creator.String(),
		"elapsed_ms", elapsed.Milliseconds())

	return nil
}

// AfterInstantiateHook is called after successful contract instantiation
// Performs the actual registration with the contract address
func (k Keeper) AfterInstantiateHook(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	codeID uint64,
	creator sdk.AccAddress,
	admin sdk.AccAddress,
	label string,
) error {
	// Skip if registry not available or circuit breaker is open
	if k.contractRegistry == nil {
		return nil
	}

	if k.shouldSkipCircuitBreaker(ctx) {
		k.Logger(ctx).Warn("circuit breaker open, skipping contract registration",
			"contract", contractAddr.String(),
			"state", k.getCircuitBreakerStateString(ctx))
		return nil // Graceful degradation
	}

	startTime := determinism.GetBlockTime(ctx)

	// Get default params (returns simple types.Params, not proto params)
	// For proto params with fields like DefaultMaxGas, we need to use hardcoded defaults
	// since the keeper GetParams doesn't return the full proto params
	wasmParams := k.GetParams(ctx)

	// Create contract info with default settings using proto types
	info := pb.ContractInfo{
		Address: contractAddr.String(),
		CodeId:  codeID,
		Creator: creator.String(),
		Admin:   admin.String(),
		Label:   label,
		Metadata: pb.ContractMetadata{
			Name:        label,
			Description: fmt.Sprintf("Auto-registered WASM contract (code ID: %d)", codeID),
			Tags:        []string{"wasm", "auto-registered"},
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			AllowMigration:   !wasmParams.RequireAdminForMigrate,
			MaxGasPerTx:      1_000_000, // Default 1M gas per tx
			RateLimitPerUser: 100,       // Default 100 calls per hour
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}

	// Attempt registration with retry logic
	var err error
	for retry := 0; retry < 3; retry++ {
		err = k.contractRegistry.RegisterContract(ctx, &info)
		if err == nil {
			break
		}

		// Exponential backoff
		if retry < 2 {
			time.Sleep(time.Duration(1<<retry) * 100 * time.Millisecond)
		}
	}

	if err != nil {
		// Record failure
		k.recordCircuitBreakerFailure(ctx)

		// Log error but don't fail instantiation
		k.Logger(ctx).Error("failed to auto-register contract",
			"contract", contractAddr.String(),
			"error", err,
			"retries", 3)

		// Emit error event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"contract_registration_failed",
				sdk.NewAttribute("contract_address", contractAddr.String()),
				sdk.NewAttribute("creator", creator.String()),
				sdk.NewAttribute("error", err.Error()),
			),
		)

		return nil // Don't fail instantiation if registry fails
	}

	// Record success
	k.recordCircuitBreakerSuccess(ctx)

	// Update stats
	k.incrementSecurityStat(ctx, "instantiated")

	// Emit success event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"contract_auto_registered",
			sdk.NewAttribute("contract_address", contractAddr.String()),
			sdk.NewAttribute("code_id", fmt.Sprintf("%d", codeID)),
			sdk.NewAttribute("creator", creator.String()),
			sdk.NewAttribute("label", label),
		),
	)

	// Log metrics
	elapsed := time.Since(startTime)
	k.Logger(ctx).Info("contract auto-registered",
		"contract", contractAddr.String(),
		"code_id", codeID,
		"elapsed_ms", elapsed.Milliseconds())

	return nil
}

// BeforeExecuteHook is called before contract execution
// Validates execution against all registry policies
func (k Keeper) BeforeExecuteHook(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	sender sdk.AccAddress,
) error {
	// Skip if registry not available
	if k.contractRegistry == nil {
		k.Logger(ctx).Debug("contract registry not available, skipping validation")
		return nil
	}

	// Skip if circuit breaker is open (permissive mode)
	if k.shouldSkipCircuitBreaker(ctx) {
		k.Logger(ctx).Warn("circuit breaker open, permissive mode enabled",
			"contract", contractAddr.String(),
			"state", k.getCircuitBreakerStateString(ctx))

		// Emit alert
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"contract_registry_degraded",
				sdk.NewAttribute("circuit_breaker_state", k.getCircuitBreakerStateString(ctx)),
				sdk.NewAttribute("operation", "validate_execution"),
				sdk.NewAttribute("mode", "permissive"),
			),
		)
		return nil // Permissive mode when registry is down
	}

	startTime := determinism.GetBlockTime(ctx)

	// Check validation cache (performance optimization)
	cacheKey := getCacheKey(contractAddr.String(), sender.String(), ctx.BlockHeight())
	if cached, found := valCache.get(cacheKey, ctx.BlockHeight()); found {
		k.Logger(ctx).Debug("validation cache hit",
			"contract", contractAddr.String(),
			"sender", sender.String())
		return cached.error
	}

	// Cleanup old cache entries periodically
	if ctx.BlockHeight()%100 == 0 {
		valCache.cleanup(ctx.BlockHeight())
	}

	// Get gas limit from context
	gasLimit := ctx.GasMeter().Limit()

	// Validate execution
	err := k.contractRegistry.ValidateContractExecution(
		ctx,
		contractAddr.String(),
		sender.String(),
		gasLimit,
	)

	// Cache the result
	valCache.set(cacheKey, &validationCacheEntry{
		allowed:     err == nil,
		error:       err,
		timestamp:   determinism.GetBlockTime(ctx),
		blockHeight: ctx.BlockHeight(),
	})

	if err != nil {
		// Record failure metric (but not circuit breaker - validation failures are expected)
		k.Logger(ctx).Debug("contract execution validation failed",
			"contract", contractAddr.String(),
			"sender", sender.String(),
			"error", err)

		// Emit validation failure event with details
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"contract_execution_denied",
				sdk.NewAttribute("contract_address", contractAddr.String()),
				sdk.NewAttribute("sender", sender.String()),
				sdk.NewAttribute("reason", err.Error()),
			),
		)

		return fmt.Errorf("operation failed: %w", err)
	}

	// Increment rate limit counter after successful validation
	k.contractRegistry.IncrementRateLimit(ctx, contractAddr.String(), sender.String())

	// Record success
	k.recordCircuitBreakerSuccess(ctx)

	// Log metrics
	elapsed := time.Since(startTime)
	if elapsed > 2*time.Millisecond {
		k.Logger(ctx).Warn("validation exceeded performance target",
			"contract", contractAddr.String(),
			"elapsed_ms", elapsed.Milliseconds(),
			"target_ms", 2)
	}

	k.Logger(ctx).Debug("before execute hook completed",
		"contract", contractAddr.String(),
		"sender", sender.String(),
		"elapsed_ms", elapsed.Milliseconds())

	return nil
}

// AfterExecuteHook is called after contract execution
// Updates contract metrics and analytics
func (k Keeper) AfterExecuteHook(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	gasUsedBefore uint64,
	success bool,
	executeErr error,
) {
	// Skip if registry not available
	if k.contractRegistry == nil {
		return
	}

	// Skip if circuit breaker is open
	if k.shouldSkipCircuitBreaker(ctx) {
		return
	}

	startTime := determinism.GetBlockTime(ctx)

	// Calculate gas used
	gasUsed := ctx.GasMeter().GasConsumed() - gasUsedBefore

	// Add to metrics buffer (batch processing)
	metricsBuf.add(metricsUpdate{
		contractAddr: contractAddr.String(),
		gasUsed:      gasUsed,
		success:      success,
		timestamp:    ctx.BlockTime(),
	})

	// Flush buffer if needed
	if metricsBuf.shouldFlush() {
		k.flushMetricsBuffer(ctx)
	}

	// Update security stats
	k.incrementSecurityStat(ctx, "executions")

	// Emit execution event
	eventAttrs := []sdk.Attribute{
		sdk.NewAttribute("contract_address", contractAddr.String()),
		sdk.NewAttribute("gas_used", fmt.Sprintf("%d", gasUsed)),
		sdk.NewAttribute("success", fmt.Sprintf("%t", success)),
	}

	if !success && executeErr != nil {
		eventAttrs = append(eventAttrs, sdk.NewAttribute("error", executeErr.Error()))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("contract_execution_completed", eventAttrs...),
	)

	// Log metrics
	elapsed := time.Since(startTime)
	k.Logger(ctx).Debug("after execute hook completed",
		"contract", contractAddr.String(),
		"gas_used", gasUsed,
		"success", success,
		"elapsed_ms", elapsed.Milliseconds())
}

// flushMetricsBuffer flushes pending metrics updates to the registry
func (k Keeper) flushMetricsBuffer(ctx sdk.Context) {
	updates := metricsBuf.flush()
	if len(updates) == 0 {
		return
	}

	k.Logger(ctx).Debug("flushing metrics buffer", "count", len(updates))

	// Batch update metrics
	for _, update := range updates {
		// Wrap in recover to prevent panic from breaking execution
		func() {
			defer func() {
				if r := recover(); r != nil {
					k.Logger(ctx).Error("panic during metrics update",
						"contract", update.contractAddr,
						"panic", r)
				}
			}()

			// Update metrics (non-blocking)
			err := k.updateMetricsSafe(ctx, update)
			if err != nil {
				k.Logger(ctx).Error("failed to update metrics",
					"contract", update.contractAddr,
					"error", err)
			}
		}()
	}
}

// updateMetricsSafe safely updates metrics synchronously
// CONSENSUS-CRITICAL: This function must be synchronous and deterministic.
// Previous implementation used goroutines with wall-clock timeouts which caused
// different validators to produce different state (some timing out, others not).
func (k Keeper) updateMetricsSafe(ctx sdk.Context, update metricsUpdate) error {
	// Wrap in recover to prevent panic from breaking consensus
	defer func() {
		if r := recover(); r != nil {
			k.Logger(ctx).Error("panic during metrics update",
				"contract", update.contractAddr,
				"panic", r)
		}
	}()

	// Synchronous update - no goroutines, no timeouts
	// All validators will either succeed or fail together
	k.contractRegistry.UpdateMetricsOnExecution(
		ctx,
		update.contractAddr,
		update.gasUsed,
		update.success,
	)
	return nil
}

// checkRegistrationRateLimit checks if creator is rate-limited for registration
func (k Keeper) checkRegistrationRateLimit(ctx sdk.Context, creator string) error {
	// Simple in-memory rate limiting: max 10 contracts per block per creator
	store := ctx.KVStore(k.storeKey)
	key := append([]byte("reg_rate_limit_"), []byte(creator)...)
	key = append(key, []byte(fmt.Sprintf("_%d", ctx.BlockHeight()))...)

	bz := store.Get(key)
	count := uint64(0)
	if bz != nil {
		count = sdk.BigEndianToUint64(bz)
	}

	if count >= 10 {
		return types.ErrUnauthorized.Wrap("registration rate limit exceeded")
	}

	// Increment counter
	count++
	store.Set(key, sdk.Uint64ToBigEndian(count))

	return nil
}

// GetCircuitBreakerStatus returns the current circuit breaker status (for monitoring)
func (k Keeper) GetCircuitBreakerStatus(ctx sdk.Context) string {
	return k.getCircuitBreakerStateString(ctx)
}

// ResetCircuitBreaker manually resets the circuit breaker (governance/emergency)
func (k Keeper) ResetCircuitBreaker(ctx sdk.Context) {
	data := circuitBreakerData{
		FailureCount:       0,
		LastFailure:        time.Time{},
		State:              circuitBreakerStateClosed,
		ConsecutiveSuccess: 0,
	}
	k.setCircuitBreakerState(ctx, data)
	k.Logger(ctx).Info("circuit breaker manually reset")
}
