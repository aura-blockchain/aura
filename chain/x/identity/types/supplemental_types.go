// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// RecoveryRecord - Supplemental type for account recovery
// ============================================================================

// RecoveryRecord holds information about identity recovery mechanisms
type RecoveryRecord struct {
	DID             string    `json:"did"`
	RecoveryAddress string    `json:"recovery_address"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Attempts        uint32    `json:"attempts"`
	MaxAttempts     uint32    `json:"max_attempts"`
	LastAttemptAt   time.Time `json:"last_attempt_at"`
	Metadata        string    `json:"metadata"`
}

// Reset implements the proto.Message interface
func (m *RecoveryRecord) Reset() {
	*m = RecoveryRecord{}
}

// String implements the proto.Message interface
func (m *RecoveryRecord) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("RecoveryRecord{DID:%s, RecoveryAddress:%s, Status:%s, CreatedAt:%v, UpdatedAt:%v, ExpiresAt:%v, Attempts:%d, MaxAttempts:%d}",
		m.DID, m.RecoveryAddress, m.Status, m.CreatedAt, m.UpdatedAt, m.ExpiresAt, m.Attempts, m.MaxAttempts)
}

// ProtoMessage implements the proto.Message interface
func (*RecoveryRecord) ProtoMessage() {}

// Validate performs basic validation
func (m *RecoveryRecord) Validate() error {
	if m.DID == "" {
		return ErrInvalidDID.Wrap("DID cannot be empty")
	}
	if m.RecoveryAddress == "" {
		return ErrInvalidInput.Wrap("recovery address cannot be empty")
	}
	if m.Status == "" {
		return ErrInvalidInput.Wrap("status cannot be empty")
	}
	if m.MaxAttempts > 0 && m.Attempts > m.MaxAttempts {
		return ErrInvalidInput.Wrap("attempts cannot exceed max attempts")
	}
	return nil
}

// ============================================================================
// VerificationRecord - Supplemental type for identity verification
// ============================================================================

// VerificationRecord tracks identity verification levels and status
type VerificationRecord struct {
	DID          string    `json:"did"`
	Level        int32     `json:"level"`
	VerifiedAt   time.Time `json:"verified_at"`
	VerifiedBy   string    `json:"verified_by"`
	Method       string    `json:"method"`
	ExpiresAt    time.Time `json:"expires_at"`
	Status       string    `json:"status"`
	Documents    []string  `json:"documents"`
	Attestations []string  `json:"attestations"`
	Metadata     string    `json:"metadata"`
}

// Reset implements the proto.Message interface
func (m *VerificationRecord) Reset() {
	*m = VerificationRecord{}
}

// String implements the proto.Message interface
func (m *VerificationRecord) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("VerificationRecord{DID:%s, Level:%d, VerifiedAt:%v, VerifiedBy:%s, Method:%s, Status:%s, Documents:%d, Attestations:%d}",
		m.DID, m.Level, m.VerifiedAt, m.VerifiedBy, m.Method, m.Status, len(m.Documents), len(m.Attestations))
}

// ProtoMessage implements the proto.Message interface
func (*VerificationRecord) ProtoMessage() {}

// Validate performs basic validation
func (m *VerificationRecord) Validate() error {
	if m.DID == "" {
		return ErrInvalidDID.Wrap("DID cannot be empty")
	}
	if m.Level < 0 {
		return ErrInvalidInput.Wrap("verification level cannot be negative")
	}
	if m.VerifiedBy == "" {
		return ErrInvalidInput.Wrap("verified_by cannot be empty")
	}
	if m.Status == "" {
		return ErrInvalidInput.Wrap("status cannot be empty")
	}
	return nil
}

// ============================================================================
// DelegationRecord - Supplemental type for permission delegation
// ============================================================================

// DelegationRecord tracks delegated permissions from one identity to another
type DelegationRecord struct {
	DID         string    `json:"did"`
	DelegatedTo string    `json:"delegated_to"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      string    `json:"status"`
	CanRevoke   bool      `json:"can_revoke"`
	Metadata    string    `json:"metadata"`
}

// Reset implements the proto.Message interface
func (m *DelegationRecord) Reset() {
	*m = DelegationRecord{}
}

// String implements the proto.Message interface
func (m *DelegationRecord) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("DelegationRecord{DID:%s, DelegatedTo:%s, Permissions:[%s], CreatedAt:%v, ExpiresAt:%v, Status:%s, CanRevoke:%v}",
		m.DID, m.DelegatedTo, strings.Join(m.Permissions, ","), m.CreatedAt, m.ExpiresAt, m.Status, m.CanRevoke)
}

// ProtoMessage implements the proto.Message interface
func (*DelegationRecord) ProtoMessage() {}

// Validate performs basic validation
func (m *DelegationRecord) Validate() error {
	if m.DID == "" {
		return ErrInvalidDID.Wrap("DID cannot be empty")
	}
	if m.DelegatedTo == "" {
		return ErrInvalidInput.Wrap("delegated_to cannot be empty")
	}
	if len(m.Permissions) == 0 {
		return ErrInvalidInput.Wrap("permissions cannot be empty")
	}
	if m.Status == "" {
		return ErrInvalidInput.Wrap("status cannot be empty")
	}
	return nil
}

// ============================================================================
// FederationRecord - Supplemental type for federated identity
// ============================================================================

// FederationRecord tracks cross-chain or cross-system identity federation
type FederationRecord struct {
	DID            string    `json:"did"`
	FederatedChain string    `json:"federated_chain"`
	ExternalDID    string    `json:"external_did"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Verified       bool      `json:"verified"`
	VerifiedBy     string    `json:"verified_by"`
	VerifiedAt     time.Time `json:"verified_at"`
	ProofHash      string    `json:"proof_hash"`
	Metadata       string    `json:"metadata"`
}

// Reset implements the proto.Message interface
func (m *FederationRecord) Reset() {
	*m = FederationRecord{}
}

// String implements the proto.Message interface
func (m *FederationRecord) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("FederationRecord{DID:%s, FederatedChain:%s, ExternalDID:%s, Status:%s, Verified:%v, CreatedAt:%v, UpdatedAt:%v}",
		m.DID, m.FederatedChain, m.ExternalDID, m.Status, m.Verified, m.CreatedAt, m.UpdatedAt)
}

// ProtoMessage implements the proto.Message interface
func (*FederationRecord) ProtoMessage() {}

// Validate performs basic validation
func (m *FederationRecord) Validate() error {
	if m.DID == "" {
		return ErrInvalidDID.Wrap("DID cannot be empty")
	}
	if m.FederatedChain == "" {
		return ErrInvalidInput.Wrap("federated_chain cannot be empty")
	}
	if m.ExternalDID == "" {
		return ErrInvalidInput.Wrap("external_did cannot be empty")
	}
	if m.Status == "" {
		return ErrInvalidInput.Wrap("status cannot be empty")
	}
	if m.Verified && m.VerifiedBy == "" {
		return ErrInvalidInput.Wrap("verified_by cannot be empty when verified is true")
	}
	return nil
}

// ============================================================================
// CrossChainLink - Supplemental type for cross-chain identity linking
// ============================================================================

// CrossChainLink represents a link between identities on different chains
type CrossChainLink struct {
	DID           string    `json:"did"`
	TargetChain   string    `json:"target_chain"`
	TargetDID     string    `json:"target_did"`
	LinkType      string    `json:"link_type"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ConfirmedAt   time.Time `json:"confirmed_at"`
	Confirmed     bool      `json:"confirmed"`
	ProofData     string    `json:"proof_data"`
	ProofHash     string    `json:"proof_hash"`
	Bidirectional bool      `json:"bidirectional"`
	RelayAddress  string    `json:"relay_address"`
	Metadata      string    `json:"metadata"`
}

// Reset implements the proto.Message interface
func (m *CrossChainLink) Reset() {
	*m = CrossChainLink{}
}

// String implements the proto.Message interface
func (m *CrossChainLink) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("CrossChainLink{DID:%s, TargetChain:%s, TargetDID:%s, LinkType:%s, Status:%s, Confirmed:%v, Bidirectional:%v, CreatedAt:%v}",
		m.DID, m.TargetChain, m.TargetDID, m.LinkType, m.Status, m.Confirmed, m.Bidirectional, m.CreatedAt)
}

// ProtoMessage implements the proto.Message interface
func (*CrossChainLink) ProtoMessage() {}

// Validate performs basic validation
func (m *CrossChainLink) Validate() error {
	if m.DID == "" {
		return ErrInvalidDID.Wrap("DID cannot be empty")
	}
	if m.TargetChain == "" {
		return ErrInvalidInput.Wrap("target_chain cannot be empty")
	}
	if m.TargetDID == "" {
		return ErrInvalidInput.Wrap("target_did cannot be empty")
	}
	if m.LinkType == "" {
		return ErrInvalidInput.Wrap("link_type cannot be empty")
	}
	if m.Status == "" {
		return ErrInvalidInput.Wrap("status cannot be empty")
	}
	if m.Bidirectional && m.RelayAddress == "" {
		return ErrInvalidInput.Wrap("relay_address cannot be empty for bidirectional links")
	}
	return nil
}

// ============================================================================
// Status Constants
// ============================================================================

const (
	// Recovery statuses
	RecoveryStatusPending   = "pending"
	RecoveryStatusActive    = "active"
	RecoveryStatusCompleted = "completed"
	RecoveryStatusExpired   = "expired"
	RecoveryStatusRevoked   = "revoked"

	// Verification statuses
	VerificationStatusPending  = "pending"
	VerificationStatusVerified = "verified"
	VerificationStatusExpired  = "expired"
	VerificationStatusRevoked  = "revoked"
	VerificationStatusRejected = "rejected"

	// Delegation statuses
	DelegationStatusActive    = "active"
	DelegationStatusExpired   = "expired"
	DelegationStatusRevoked   = "revoked"
	DelegationStatusSuspended = "suspended"

	// Federation statuses
	FederationStatusPending   = "pending"
	FederationStatusActive    = "active"
	FederationStatusVerified  = "verified"
	FederationStatusRevoked   = "revoked"
	FederationStatusSuspended = "suspended"

	// CrossChainLink statuses
	CrossChainLinkStatusPending   = "pending"
	CrossChainLinkStatusActive    = "active"
	CrossChainLinkStatusConfirmed = "confirmed"
	CrossChainLinkStatusBroken    = "broken"
	CrossChainLinkStatusRevoked   = "revoked"

	// Link types
	LinkTypeOwnership    = "ownership"
	LinkTypeDelegation   = "delegation"
	LinkTypeFederation   = "federation"
	LinkTypeVerification = "verification"
)
