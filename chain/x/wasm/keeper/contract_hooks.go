package keeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/common/determinism"
	contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// CIRCUIT BREAKER PATTERN
// ============================================================================

// circuitBreakerState tracks the health of the contract registry integration
type circuitBreakerState struct {
	mu                sync.RWMutex
	failureCount      int
	lastFailure       time.Time
	state             string // "closed", "open", "half-open"
	consecutiveSuccess int
}

var (
	circuitBreaker = &circuitBreakerState{
		state: "closed",
	}
)

const (
	circuitBreakerThreshold     = 5     // failures before opening
	circuitBreakerTimeout       = 60    // seconds before attempting half-open
	circuitBreakerSuccessThreshold = 3  // successes before closing from half-open
)

// checkCircuitBreaker returns true if we should skip registry calls
func (cb *circuitBreakerState) shouldSkip(ctx context.Context) bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == "closed" {
		return false
	}

	if cb.state == "open" {
		// Check if timeout elapsed using deterministic block time
		elapsed := determinism.TimeSince(ctx, cb.lastFailure)
		if elapsed > time.Duration(circuitBreakerTimeout)*time.Second {
			// Transition to half-open
			return false
		}
		return true
	}

	// half-open state
	return false
}

// recordSuccess records a successful registry call
func (cb *circuitBreakerState) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveSuccess++
	cb.failureCount = 0

	if cb.state == "half-open" && cb.consecutiveSuccess >= circuitBreakerSuccessThreshold {
		cb.state = "closed"
		cb.consecutiveSuccess = 0
	}
}

// recordFailure records a failed registry call
func (cb *circuitBreakerState) recordFailure(ctx context.Context) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = determinism.GetBlockTime(ctx)
	cb.consecutiveSuccess = 0

	if cb.state == "half-open" {
		cb.state = "open"
	} else if cb.failureCount >= circuitBreakerThreshold {
		cb.state = "open"
	}
}

// getState returns the current circuit breaker state
func (cb *circuitBreakerState) getState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ============================================================================
// VALIDATION CACHE (Performance Optimization)
// ============================================================================

type validationCacheEntry struct {
	allowed    bool
	error      error
	timestamp  time.Time
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

const (
	validationCacheDuration = 5 * time.Second // cache for same block
)

// getCacheKey generates a cache key for validation
func getCacheKey(contractAddr, sender string, blockHeight int64) string {
	return fmt.Sprintf("%s:%s:%d", contractAddr, sender, blockHeight)
}

// get retrieves a cached validation result
func (vc *validationCache) get(key string, blockHeight int64) (*validationCacheEntry, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	entry, found := vc.entries[key]
	if !found {
		return nil, false
	}

	// Invalidate if from a different block or too old
	if entry.blockHeight != blockHeight || time.Since(entry.timestamp) > validationCacheDuration {
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
	metricsBufferSize     = 50  // flush after this many updates
	metricsBufferDuration = 10  // flush after this many seconds
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

// setContractRegistryInternal sets the contract registry keeper (internal only)
func (k *Keeper) setContractRegistryInternal(registry *contractregistrykeeper.Keeper) {
	k.contractRegistry = registry
}

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

	if circuitBreaker.shouldSkip(ctx) {
		k.Logger(ctx).Warn("circuit breaker open, skipping registry auto-registration",
			"state", circuitBreaker.getState())
		// Emit alert event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"contract_registry_degraded",
				sdk.NewAttribute("circuit_breaker_state", circuitBreaker.getState()),
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
	params := k.contractRegistry.GetParams(ctx)
	if params.MaxContractsPerCreator > 0 {
		// Use GetCreatorContracts to count existing contracts
		contracts := k.contractRegistry.GetCreatorContracts(ctx, creator.String())
		count := uint64(len(contracts))
		if count >= params.MaxContractsPerCreator {
			k.Logger(ctx).Error("creator contract limit exceeded",
				"creator", creator.String(),
				"current", count,
				"max", params.MaxContractsPerCreator)
			circuitBreaker.recordFailure(ctx)
			return types.ErrUnauthorized.Wrapf("creator contract limit exceeded: %d >= %d", count, params.MaxContractsPerCreator)
		}
	}

	// Record success
	circuitBreaker.recordSuccess()

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

	if circuitBreaker.shouldSkip(ctx) {
		k.Logger(ctx).Warn("circuit breaker open, skipping contract registration",
			"contract", contractAddr.String(),
			"state", circuitBreaker.getState())
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
		Metadata: &pb.ContractMetadata{
			Name:        label,
			Description: fmt.Sprintf("Auto-registered WASM contract (code ID: %d)", codeID),
			Tags:        []string{"wasm", "auto-registered"},
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:       true,
			AllowMigration:   !wasmParams.RequireAdminForMigrate,
			MaxGasPerTx:      1_000_000, // Default 1M gas per tx
			RateLimitPerUser: 100,       // Default 100 calls per hour
		},
		Compliance: &pb.ComplianceRequirements{},
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
		circuitBreaker.recordFailure(ctx)

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
	circuitBreaker.recordSuccess()

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
	if circuitBreaker.shouldSkip(ctx) {
		k.Logger(ctx).Warn("circuit breaker open, permissive mode enabled",
			"contract", contractAddr.String(),
			"state", circuitBreaker.getState())

		// Emit alert
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"contract_registry_degraded",
				sdk.NewAttribute("circuit_breaker_state", circuitBreaker.getState()),
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

		return err
	}

	// Note: IncrementRateLimit is not implemented in the contract registry keeper
	// Rate limiting is checked via CheckRateLimit in ValidateContractExecution above
	// TODO: Implement IncrementRateLimit if explicit increment tracking is needed
	// k.contractRegistry.IncrementRateLimit(ctx, contractAddr.String(), sender.String())

	// Record success
	circuitBreaker.recordSuccess()

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
	if circuitBreaker.shouldSkip(ctx) {
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

// updateMetricsSafe safely updates metrics with timeout protection
func (k Keeper) updateMetricsSafe(ctx sdk.Context, update metricsUpdate) error {
	// Create a channel for the result
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()

		k.contractRegistry.UpdateMetricsOnExecution(
			ctx,
			update.contractAddr,
			update.gasUsed,
			update.success,
		)
		done <- nil
	}()

	// Wait with timeout
	select {
	case err := <-done:
		return err
	case <-time.After(500 * time.Millisecond):
		return fmt.Errorf("metrics update timeout")
	}
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
func (k Keeper) GetCircuitBreakerStatus() string {
	return circuitBreaker.getState()
}

// ResetCircuitBreaker manually resets the circuit breaker (governance/emergency)
func (k Keeper) ResetCircuitBreaker(ctx context.Context) {
	circuitBreaker.mu.Lock()
	defer circuitBreaker.mu.Unlock()

	circuitBreaker.state = "closed"
	circuitBreaker.failureCount = 0
	circuitBreaker.consecutiveSuccess = 0

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("circuit breaker manually reset")
}
