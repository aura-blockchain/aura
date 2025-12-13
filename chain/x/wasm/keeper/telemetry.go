package keeper

//lint:file-ignore U1000 -- telemetry hooks are defined for future instrumentation but not yet wired

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ============================================================================
// WASM MODULE PROMETHEUS METRICS
// ============================================================================
// Implements monitoring probes for WASM transaction failures, state load
// errors, and circuit breaker status as specified in ROADMAP_PRODUCTION.md

var (
	// Transaction metrics
	wasmTxTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "tx_total",
			Help:      "Total number of WASM transactions processed",
		},
		[]string{"contract_address", "operation"},
	)

	wasmTxFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "tx_failures_total",
			Help:      "Total number of WASM transaction failures by contract and error type",
		},
		[]string{"contract_address", "error_type", "operation"},
	)

	// Instantiation metrics
	wasmInstantiationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "instantiation_total",
			Help:      "Total number of WASM contract instantiations",
		},
		[]string{"code_id"},
	)

	wasmInstantiationFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "instantiation_failures_total",
			Help:      "Total number of WASM contract instantiation failures",
		},
		[]string{"code_id", "error_type"},
	)

	// Circuit breaker metrics
	wasmCircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "circuit_breaker_state",
			Help:      "WASM circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"state"},
	)

	wasmCircuitBreakerTransitions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "circuit_breaker_transitions_total",
			Help:      "Total number of circuit breaker state transitions",
		},
		[]string{"from_state", "to_state"},
	)

	// Validation cache metrics
	wasmValidationCacheTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "validation_cache_total",
			Help:      "Total number of validation cache lookups",
		},
	)

	wasmValidationCacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "validation_cache_hits_total",
			Help:      "Total number of validation cache hits",
		},
	)

	wasmValidationCacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "validation_cache_misses_total",
			Help:      "Total number of validation cache misses",
		},
	)

	// Registry integration metrics
	wasmRegistryErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "registry_errors_total",
			Help:      "Total number of contract registry integration errors",
		},
		[]string{"operation", "error_type"},
	)

	// Execution metrics
	wasmExecutionDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "execution_duration_seconds",
			Help:      "WASM contract execution duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"contract_address", "success"},
	)

	wasmGasUsedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "gas_used_total",
			Help:      "Total gas used by WASM contract executions",
		},
		[]string{"contract_address"},
	)

	// Hook metrics
	wasmHookDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "aura",
			Subsystem: "wasm",
			Name:      "hook_duration_seconds",
			Help:      "WASM hook execution duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10), // 0.1ms to ~100ms
		},
		[]string{"hook_type", "contract_address"},
	)
)

// ============================================================================
// METRIC RECORDING FUNCTIONS
// ============================================================================

// recordWasmTx records a WASM transaction attempt
func (k Keeper) recordWasmTx(contractAddr string, operation string) {
	wasmTxTotal.WithLabelValues(contractAddr, operation).Inc()
}

// recordWasmTxFailure records a WASM transaction failure
func (k Keeper) recordWasmTxFailure(contractAddr string, errorType string, operation string) {
	wasmTxFailuresTotal.WithLabelValues(contractAddr, errorType, operation).Inc()
}

// recordWasmInstantiation records a contract instantiation
func (k Keeper) recordWasmInstantiation(codeID string) {
	wasmInstantiationTotal.WithLabelValues(codeID).Inc()
}

// recordWasmInstantiationFailure records an instantiation failure
func (k Keeper) recordWasmInstantiationFailure(codeID string, errorType string) {
	wasmInstantiationFailuresTotal.WithLabelValues(codeID, errorType).Inc()
}

// updateCircuitBreakerMetrics updates circuit breaker state metrics
func (k Keeper) updateCircuitBreakerMetrics(state string) {
	// Reset all state gauges
	wasmCircuitBreakerState.Reset()

	// Set the current state
	var stateValue float64
	switch state {
	case "closed":
		stateValue = 0
	case "open":
		stateValue = 1
	case "half-open":
		stateValue = 2
	}
	wasmCircuitBreakerState.WithLabelValues(state).Set(stateValue)
}

// recordCircuitBreakerTransition records a state transition
func (k Keeper) recordCircuitBreakerTransition(fromState, toState string) {
	wasmCircuitBreakerTransitions.WithLabelValues(fromState, toState).Inc()
}

// recordValidationCacheHit records a cache hit
func (k Keeper) recordValidationCacheHit() {
	wasmValidationCacheTotal.Inc()
	wasmValidationCacheHitsTotal.Inc()
}

// recordValidationCacheMiss records a cache miss
func (k Keeper) recordValidationCacheMiss() {
	wasmValidationCacheTotal.Inc()
	wasmValidationCacheMissesTotal.Inc()
}

// recordRegistryError records a registry integration error
func (k Keeper) recordRegistryError(operation string, errorType string) {
	wasmRegistryErrorsTotal.WithLabelValues(operation, errorType).Inc()
}

// recordWasmExecution records execution metrics
func (k Keeper) recordWasmExecution(contractAddr string, duration time.Duration, gasUsed uint64, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}

	wasmExecutionDurationSeconds.WithLabelValues(contractAddr, successStr).Observe(duration.Seconds())
	wasmGasUsedTotal.WithLabelValues(contractAddr).Add(float64(gasUsed))
}

// recordWasmHook records hook execution metrics
func (k Keeper) recordWasmHook(hookType string, contractAddr string, duration time.Duration) {
	wasmHookDurationSeconds.WithLabelValues(hookType, contractAddr).Observe(duration.Seconds())
}

// ============================================================================
// ERROR TYPE CLASSIFICATION
// ============================================================================

// classifyWasmError classifies error into metric-friendly categories
func classifyWasmError(err error) string {
	if err == nil {
		return "unknown"
	}

	errMsg := err.Error()

	// Classify common WASM errors
	switch {
	case contains(errMsg, "out of gas"):
		return "out_of_gas"
	case contains(errMsg, "contract not found"):
		return "contract_not_found"
	case contains(errMsg, "unauthorized"):
		return "unauthorized"
	case contains(errMsg, "invalid"):
		return "invalid_input"
	case contains(errMsg, "exceeded"):
		return "limit_exceeded"
	case contains(errMsg, "paused"):
		return "contract_paused"
	case contains(errMsg, "registry"):
		return "registry_error"
	case contains(errMsg, "rate limit"):
		return "rate_limited"
	case contains(errMsg, "circuit breaker"):
		return "circuit_breaker_open"
	default:
		return "other"
	}
}

// contains checks if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
