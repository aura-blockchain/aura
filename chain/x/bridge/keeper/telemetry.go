// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

//lint:file-ignore U1000 // telemetry metrics reserved for observability wiring; may be unused in tests

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ============================================================================
// BRIDGE MODULE PROMETHEUS METRICS
// ============================================================================
// Implements monitoring probes for signature verification mismatches and
// state load errors as specified in ROADMAP_PRODUCTION.md

var (
	// Signature verification metrics
	bridgeSignatureVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "signature_verifications_total",
			Help:      "Total number of signature verifications performed",
		},
		[]string{"chain", "signature_type"},
	)

	bridgeSignatureMismatchesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "signature_mismatches_total",
			Help:      "Total number of signature verification mismatches",
		},
		[]string{"chain", "signature_type", "error_reason"},
	)

	bridgeInvalidRecoveryIDTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "invalid_recovery_id_total",
			Help:      "Total number of invalid recovery IDs in signatures",
		},
		[]string{"chain"},
	)

	bridgePubKeyRecoveryFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "pubkey_recovery_failures_total",
			Help:      "Total number of public key recovery failures",
		},
		[]string{"chain", "recovery_id"},
	)

	bridgeSignatureVerificationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "signature_verification_duration_seconds",
			Help:      "Signature verification duration in seconds",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 12), // 0.1ms to ~400ms
		},
		[]string{"chain", "success"},
	)

	// State load error metrics
	stateLoadErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "state",
			Name:      "load_errors_total",
			Help:      "Total number of state load errors by module and store",
		},
		[]string{"module", "store", "error_type"},
	)

	unmarshalErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "unmarshal",
			Name:      "errors_total",
			Help:      "Total number of protobuf unmarshal errors",
		},
		[]string{"module", "proto_type"},
	)

	stateCorruptionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "state",
			Name:      "corruption_total",
			Help:      "Total number of detected state corruption events",
		},
		[]string{"module", "store"},
	)

	kvStoreIterationErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "kvstore",
			Name:      "iteration_errors_total",
			Help:      "Total number of KVStore iteration errors",
		},
		[]string{"store"},
	)

	// Bridge transfer metrics
	bridgeTransfersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "transfers_total",
			Help:      "Total number of bridge transfers initiated",
		},
		[]string{"source_chain", "dest_chain", "status"},
	)

	bridgeTransferDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "transfer_duration_seconds",
			Help:      "Bridge transfer processing duration",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
		},
		[]string{"source_chain", "dest_chain"},
	)

	// Shared identity metrics
	sharedIdentityLinksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "identity_links_total",
			Help:      "Total number of shared identity address links",
		},
		[]string{"chain"},
	)

	sharedIdentityLinkFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "aura",
			Subsystem: "bridge",
			Name:      "identity_link_failures_total",
			Help:      "Total number of failed identity link attempts",
		},
		[]string{"chain", "error_reason"},
	)
)

// ============================================================================
// SIGNATURE VERIFICATION METRICS
// ============================================================================

// recordSignatureVerification records a signature verification attempt
func (k Keeper) recordSignatureVerification(chain string, signatureType string, success bool, duration time.Duration) {
	bridgeSignatureVerificationsTotal.WithLabelValues(chain, signatureType).Inc()

	successStr := "false"
	if success {
		successStr = "true"
	}
	bridgeSignatureVerificationDuration.WithLabelValues(chain, successStr).Observe(duration.Seconds())
}

// recordSignatureMismatch records a signature mismatch
func (k Keeper) recordSignatureMismatch(chain string, signatureType string, errorReason string) {
	bridgeSignatureMismatchesTotal.WithLabelValues(chain, signatureType, errorReason).Inc()
}

// recordInvalidRecoveryID records an invalid recovery ID
func (k Keeper) recordInvalidRecoveryID(chain string) {
	bridgeInvalidRecoveryIDTotal.WithLabelValues(chain).Inc()
}

// recordPubKeyRecoveryFailure records a public key recovery failure
func (k Keeper) recordPubKeyRecoveryFailure(chain string, recoveryID string) {
	bridgePubKeyRecoveryFailuresTotal.WithLabelValues(chain, recoveryID).Inc()
}

// ============================================================================
// STATE LOAD ERROR METRICS
// ============================================================================

// recordStateLoadError records a state load error
func (k Keeper) recordStateLoadError(module string, store string, errorType string) {
	stateLoadErrorsTotal.WithLabelValues(module, store, errorType).Inc()
}

// recordUnmarshalError records a protobuf unmarshal error
func (k Keeper) recordUnmarshalError(module string, protoType string) {
	unmarshalErrorsTotal.WithLabelValues(module, protoType).Inc()
}

// recordStateCorruption records a state corruption event
func (k Keeper) recordStateCorruption(module string, store string) {
	stateCorruptionTotal.WithLabelValues(module, store).Inc()
}

// recordKVStoreIterationError records a KVStore iteration error
func (k Keeper) recordKVStoreIterationError(store string) {
	kvStoreIterationErrorsTotal.WithLabelValues(store).Inc()
}

// ============================================================================
// BRIDGE TRANSFER METRICS
// ============================================================================

// recordBridgeTransfer records a bridge transfer
func (k Keeper) recordBridgeTransfer(sourceChain string, destChain string, status string, duration time.Duration) {
	bridgeTransfersTotal.WithLabelValues(sourceChain, destChain, status).Inc()
	bridgeTransferDuration.WithLabelValues(sourceChain, destChain).Observe(duration.Seconds())
}

// ============================================================================
// SHARED IDENTITY METRICS
// ============================================================================

// recordIdentityLink records a successful identity link
func (k Keeper) recordIdentityLink(chain string) {
	sharedIdentityLinksTotal.WithLabelValues(chain).Inc()
}

// recordIdentityLinkFailure records a failed identity link attempt
func (k Keeper) recordIdentityLinkFailure(chain string, errorReason string) {
	sharedIdentityLinkFailuresTotal.WithLabelValues(chain, errorReason).Inc()
}

// ============================================================================
// ERROR CLASSIFICATION
// ============================================================================

// classifySignatureError classifies signature errors for metrics
func classifySignatureError(err error, recoveredPubKey []byte) string {
	if err == nil && recoveredPubKey == nil {
		return "pubkey_recovery_failed"
	}
	if err == nil {
		return "address_mismatch"
	}

	errMsg := err.Error()
	switch {
	case containsStr(errMsg, "invalid signature length"):
		return "invalid_signature_length"
	case containsStr(errMsg, "invalid recovery id"):
		return "invalid_recovery_id"
	case containsStr(errMsg, "recovery failed"):
		return "recovery_failed"
	case containsStr(errMsg, "verification failed"):
		return "ecdsa_verification_failed"
	case containsStr(errMsg, "malformed"):
		return "malformed_signature"
	default:
		return "other"
	}
}

// classifyStateError classifies state errors for metrics
func classifyStateError(err error) string {
	if err == nil {
		return "unknown"
	}

	errMsg := err.Error()
	switch {
	case containsStr(errMsg, "unmarshal"):
		return "unmarshal_error"
	case containsStr(errMsg, "not found"):
		return "key_not_found"
	case containsStr(errMsg, "corrupted"):
		return "data_corrupted"
	case containsStr(errMsg, "invalid"):
		return "invalid_data"
	case containsStr(errMsg, "decode"):
		return "decode_error"
	default:
		return "other"
	}
}

// containsStr checks if string contains substring
func containsStr(s, substr string) bool {
	// Simple substring check (case-sensitive for error messages)
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
