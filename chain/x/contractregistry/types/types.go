// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ContractStatus represents the status of a contract
type ContractStatus int32

const (
	ContractStatus_CONTRACT_STATUS_UNSPECIFIED ContractStatus = 0
	ContractStatus_CONTRACT_STATUS_ACTIVE      ContractStatus = 1
	ContractStatus_CONTRACT_STATUS_PAUSED      ContractStatus = 2
	ContractStatus_CONTRACT_STATUS_DEPRECATED  ContractStatus = 3
	ContractStatus_CONTRACT_STATUS_FROZEN      ContractStatus = 4
)

func (s ContractStatus) String() string {
	switch s {
	case ContractStatus_CONTRACT_STATUS_ACTIVE:
		return "ACTIVE"
	case ContractStatus_CONTRACT_STATUS_PAUSED:
		return "PAUSED"
	case ContractStatus_CONTRACT_STATUS_DEPRECATED:
		return "DEPRECATED"
	case ContractStatus_CONTRACT_STATUS_FROZEN:
		return "FROZEN"
	default:
		return "UNSPECIFIED"
	}
}

// ContractInfo represents contract information
type ContractInfo struct {
	ContractAddress string
	CodeId          uint64
	Admin           string
	Status          ContractStatus
	Metadata        *ContractMetadata
	SecurityPolicy  *SecurityPolicy
	ComplianceReqs  *ComplianceRequirements
	Verified        bool
	VerifiedAt      *timestamppb.Timestamp
	AuditScore      uint32
	LastAudit       *timestamppb.Timestamp
	CreatedAt       *timestamppb.Timestamp
	MigrationTarget string
	MigratedFrom    string
	MigratedAt      *timestamppb.Timestamp
}

// ContractMetadata represents contract metadata
type ContractMetadata struct {
	Name            string
	Description     string
	Version         string
	Tags            []string
	ContractAddress string
	CodeHash        []byte
	Creator         string
	CreatedAt       *timestamppb.Timestamp
	UpdatedAt       *timestamppb.Timestamp
}

// SecurityPolicy represents contract security policy
type SecurityPolicy struct {
	AllowPause           bool
	RequireKyc           bool
	MinKycLevel          uint32
	RequireVc            bool
	AllowedVcTypes       []string
	MinConfidenceScore   uint64
	RateLimitPerUser     uint64
	MaxGasPerExecution   uint64
	EnableAccessControl  bool
	AllowedExecutors     []string
}

// ComplianceRequirements represents compliance requirements
type ComplianceRequirements struct {
	MinKycLevel             uint32
	RestrictedJurisdictions []string
}

// WhitelistEntry represents a whitelist entry
type WhitelistEntry struct {
	ContractAddress string
	AddedAt         int64
	AddedBy         string
	Reason          string
}

// BlacklistEntry represents a blacklist entry
type BlacklistEntry struct {
	ContractAddress string
	AddedAt         int64
	AddedBy         string
	Reason          string
}

// AuditEntry represents an audit trail entry
type AuditEntry struct {
	Id              uint64
	ContractAddress string
	Timestamp       *timestamppb.Timestamp
	Action          string
	EventType       string            // Event type (DEPLOY, UPGRADE, PAUSE, etc.)
	Actor           string
	Details         string
	Metadata        map[string]string // Additional metadata
	Success         bool
}

// AuditStatistics represents audit statistics
type AuditStatistics struct {
	ContractAddress string
	TotalEntries    uint64
	TotalEvents     uint64            // Alias for TotalEntries (for test compatibility)
	SuccessCount    uint64
	FailureCount    uint64
	ActionCounts    map[string]uint64
	EventTypeCounts map[string]uint64 // Count by event type
}

// ContractMetrics represents contract usage metrics
type ContractMetrics struct {
	TotalExecutions      uint64
	SuccessfulExecutions uint64
	FailedExecutions     uint64
	TotalGasUsed         uint64
	RateLimitViolations  uint64
	ComplianceFailures   uint64
	LastExecuted         *timestamppb.Timestamp
}

// ContractVerification represents contract code verification
type ContractVerification struct {
	ContractAddress string
	CodeHash        string
	SourceUrl       string
	CompilerVersion string
	Verified        bool
	VerifiedAt      *timestamppb.Timestamp
	VerifiedBy      string
}

// AuditReport represents a contract audit report
type AuditReport struct {
	Id              uint64
	ContractAddress string
	AuditHash       string
	Auditor         string
	Score           uint32
	Findings        string
	SubmittedAt     *timestamppb.Timestamp
	SubmittedBy     string
}

// MigrationRecord represents a contract migration record
type MigrationRecord struct {
	Id                 uint64
	OldContractAddress string
	NewContractAddress string
	MigratedAt         *timestamppb.Timestamp
	MigratedBy         string
	Reason             string
	OldCodeId          uint64
	NewCodeId          uint64
}

// Params represents module parameters
type Params struct {
	AuditWarningDays       uint64
	MaxContractsPerCreator uint64
}

// Validate validates module parameters
func (p Params) Validate() error {
	return nil
}

// Message types (simplified - in production these would be proto-generated)

type MsgRegisterContract struct {
	Creator         string
	ContractAddress string
	CodeId          uint64
	Admin           string
	Metadata        *ContractMetadata
	SecurityPolicy  *SecurityPolicy
}

func (m *MsgRegisterContract) ValidateBasic() error {
	if m.Creator == "" {
		return ErrUnauthorized
	}
	if m.ContractAddress == "" {
		return ErrContractNotFound
	}
	return nil
}

type MsgUpdateContractMetadata struct {
	Admin           string
	ContractAddress string
	Metadata        *ContractMetadata
}

func (m *MsgUpdateContractMetadata) ValidateBasic() error {
	if m.Admin == "" {
		return ErrUnauthorized
	}
	return nil
}

type MsgUpdateSecurityPolicy struct {
	Admin           string
	ContractAddress string
	SecurityPolicy  *SecurityPolicy
}

func (m *MsgUpdateSecurityPolicy) ValidateBasic() error {
	if m.Admin == "" {
		return ErrUnauthorized
	}
	return nil
}

type MsgPauseContract struct {
	Signer          string
	ContractAddress string
	Reason          string
}

func (m *MsgPauseContract) ValidateBasic() error {
	return nil
}

type MsgUnpauseContract struct {
	Signer          string
	ContractAddress string
}

func (m *MsgUnpauseContract) ValidateBasic() error {
	return nil
}

type MsgDeprecateContract struct {
	Signer          string
	ContractAddress string
	Reason          string
	MigrationTarget string
}

func (m *MsgDeprecateContract) ValidateBasic() error {
	return nil
}

type MsgWhitelistContract struct {
	Authority       string
	ContractAddress string
	Reason          string
}

func (m *MsgWhitelistContract) ValidateBasic() error {
	return nil
}

type MsgBlacklistContract struct {
	Authority       string
	ContractAddress string
	Reason          string
}

func (m *MsgBlacklistContract) ValidateBasic() error {
	return nil
}

type MsgRemoveFromBlacklist struct {
	Authority       string
	ContractAddress string
	Reason          string
}

func (m *MsgRemoveFromBlacklist) ValidateBasic() error {
	return nil
}

type MsgAuditContract struct {
	Submitter       string
	ContractAddress string
	AuditHash       string
	AuditScore      uint32
	Auditor         string
	Findings        string
}

func (m *MsgAuditContract) ValidateBasic() error {
	return nil
}

type MsgVerifyContract struct {
	Submitter       string
	ContractAddress string
	CodeHash        string
	SourceUrl       string
	CompilerVersion string
}

func (m *MsgVerifyContract) ValidateBasic() error {
	return nil
}
