// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// IdentityMetrics holds all Prometheus metrics for the Identity module
type IdentityMetrics struct {
	// DID operations
	DIDRegistrations       *prometheus.CounterVec
	DIDUpdates             *prometheus.CounterVec
	DIDResolutions         *prometheus.CounterVec
	ActiveDIDs             prometheus.Gauge
	DIDRegistrationTime    prometheus.Histogram
	DIDKeyRotations        *prometheus.CounterVec
	DIDKeyRotationFails    *prometheus.CounterVec
	DIDVerificationMethods *prometheus.GaugeVec

	// Credential operations
	CredentialsIssued     *prometheus.CounterVec
	CredentialsRevoked    *prometheus.CounterVec
	CredentialsVerified   *prometheus.CounterVec
	ActiveCredentials     *prometheus.GaugeVec
	ExpiredCredentials    prometheus.Gauge
	CredRevocationTime    prometheus.Histogram
	CredVerificationFails *prometheus.CounterVec
	MerkleRootUpdates     prometheus.Counter
	RevocationListSize    prometheus.Gauge
	CredentialAge         *prometheus.GaugeVec

	// Multisig operations
	MultisigProposalsCreated  *prometheus.CounterVec
	MultisigProposalsExecuted *prometheus.CounterVec
	MultisigSignatureCount    *prometheus.GaugeVec
	TimeLockedActionsPending  prometheus.Gauge
	TimeLockedActionsExecuted prometheus.Counter
	EmergencyAdminActions     *prometheus.CounterVec

	// Session & access control
	SessionsActive      prometheus.Gauge
	SessionsCreated     prometheus.Counter
	SessionsTerminated  *prometheus.CounterVec
	RoleAssignments     *prometheus.CounterVec
	PermissionChecks    *prometheus.CounterVec
	RateLimitViolations *prometheus.CounterVec
}

var (
	identityMetricsOnce sync.Once
	identityMetrics     *IdentityMetrics
)

// NewIdentityMetrics creates and registers Identity metrics (singleton pattern)
func NewIdentityMetrics() *IdentityMetrics {
	identityMetricsOnce.Do(func() {
		identityMetrics = &IdentityMetrics{
			// DID operations
			DIDRegistrations: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "did_registrations_total",
					Help:      "Total DID registrations",
				},
				[]string{"did_method"},
			),
			DIDUpdates: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "did_updates_total",
					Help:      "Total DID document updates",
				},
				[]string{"did"},
			),
			DIDResolutions: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "did_resolutions_total",
					Help:      "Total DID resolution requests",
				},
				[]string{"status"},
			),
			ActiveDIDs: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "active_dids",
					Help:      "Currently active DIDs",
				},
			),
			DIDRegistrationTime: promauto.NewHistogram(
				prometheus.HistogramOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "did_registration_latency_seconds",
					Help:      "DID registration latency",
					Buckets:   prometheus.DefBuckets,
				},
			),
			DIDKeyRotations: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "did_key_rotations_total",
					Help:      "DID key rotation events",
				},
				[]string{"did", "key_type"},
			),
			DIDKeyRotationFails: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "did_key_rotation_failures_total",
					Help:      "Failed DID key rotations",
				},
				[]string{"reason"},
			),
			DIDVerificationMethods: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "did_verification_methods_count",
					Help:      "Verification methods per DID",
				},
				[]string{"did"},
			),

			// Credential operations
			CredentialsIssued: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "credentials_issued_total",
					Help:      "Total verifiable credentials issued",
				},
				[]string{"type"},
			),
			CredentialsRevoked: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "credentials_revoked_total",
					Help:      "Total credentials revoked",
				},
				[]string{"reason"},
			),
			CredentialsVerified: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "credentials_verified_total",
					Help:      "Total credential verifications",
				},
				[]string{"status"},
			),
			ActiveCredentials: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "active_credentials",
					Help:      "Currently active credentials",
				},
				[]string{"type"},
			),
			ExpiredCredentials: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "expired_credentials",
					Help:      "Expired but not yet removed credentials",
				},
			),
			CredRevocationTime: promauto.NewHistogram(
				prometheus.HistogramOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "credential_revocation_latency_seconds",
					Help:      "Credential revocation processing time",
					Buckets:   prometheus.DefBuckets,
				},
			),
			CredVerificationFails: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "credential_verification_failures_total",
					Help:      "Failed credential verifications",
				},
				[]string{"reason"},
			),
			MerkleRootUpdates: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "merkle_root_updates_total",
					Help:      "Merkle root update operations",
				},
			),
			RevocationListSize: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "revocation_list_size",
					Help:      "Size of credential revocation list",
				},
			),
			CredentialAge: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "credential_age_seconds",
					Help:      "Age of active credentials",
				},
				[]string{"credential_id"},
			),

			// Multisig operations
			MultisigProposalsCreated: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "multisig_proposals_created_total",
					Help:      "Multisig proposals created",
				},
				[]string{"wallet_id"},
			),
			MultisigProposalsExecuted: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "multisig_proposals_executed_total",
					Help:      "Multisig proposals executed",
				},
				[]string{"status"},
			),
			MultisigSignatureCount: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "multisig_signature_count",
					Help:      "Current signatures on multisig wallet",
				},
				[]string{"wallet_id"},
			),
			TimeLockedActionsPending: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "time_locked_actions_pending",
					Help:      "Pending time-locked actions",
				},
			),
			TimeLockedActionsExecuted: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "time_locked_actions_executed_total",
					Help:      "Time-locked actions executed",
				},
			),
			EmergencyAdminActions: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "emergency_admin_actions_total",
					Help:      "Emergency admin actions performed",
				},
				[]string{"action_type"},
			),

			// Session & access control
			SessionsActive: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "sessions_active",
					Help:      "Currently active sessions",
				},
			),
			SessionsCreated: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "sessions_created_total",
					Help:      "Total sessions created",
				},
			),
			SessionsTerminated: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "sessions_terminated_total",
					Help:      "Sessions terminated",
				},
				[]string{"reason"},
			),
			RoleAssignments: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "role_assignments_total",
					Help:      "Role assignments performed",
				},
				[]string{"role"},
			),
			PermissionChecks: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "permission_checks_total",
					Help:      "Permission check operations",
				},
				[]string{"status"},
			),
			RateLimitViolations: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "identity",
					Name:      "rate_limit_violations_total",
					Help:      "Rate limit violations detected",
				},
				[]string{"operation", "user"},
			),
		}
	})
	return identityMetrics
}

// GetIdentityMetrics returns the singleton Identity metrics instance
func GetIdentityMetrics() *IdentityMetrics {
	if identityMetrics == nil {
		return NewIdentityMetrics()
	}
	return identityMetrics
}
