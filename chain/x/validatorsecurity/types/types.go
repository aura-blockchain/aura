// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	pb "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Re-export all proto types
type (
	// Enums
	ValidatorAlert_Severity  = pb.ValidatorAlert_Severity
	ValidatorAlert_AlertType = pb.ValidatorAlert_AlertType

	// Core types
	ValidatorSecurityInfo  = pb.ValidatorSecurityInfo
	ValidatorAlert         = pb.ValidatorAlert
	SentryNodeInfo         = pb.SentryNodeInfo
	DoubleSignEvidence     = pb.DoubleSignEvidence
	DowntimeInfraction     = pb.DowntimeInfraction
	ValidatorSecurityParams = pb.ValidatorSecurityParams

	// Message types
	MsgRegisterValidator           = pb.MsgRegisterValidator
	MsgRegisterValidatorResponse   = pb.MsgRegisterValidatorResponse
	MsgUpdateSecurityInfo          = pb.MsgUpdateSecurityInfo
	MsgUpdateSecurityInfoResponse  = pb.MsgUpdateSecurityInfoResponse
	MsgRegisterSentryNode          = pb.MsgRegisterSentryNode
	MsgRegisterSentryNodeResponse  = pb.MsgRegisterSentryNodeResponse
	MsgReportDoubleSign            = pb.MsgReportDoubleSign
	MsgReportDoubleSignResponse    = pb.MsgReportDoubleSignResponse
	MsgUnjail                      = pb.MsgUnjail
	MsgUnjailResponse              = pb.MsgUnjailResponse
	MsgAcknowledgeAlert            = pb.MsgAcknowledgeAlert
	MsgAcknowledgeAlertResponse    = pb.MsgAcknowledgeAlertResponse
	MsgUpdateParams                = pb.MsgUpdateParams
	MsgUpdateParamsResponse        = pb.MsgUpdateParamsResponse

	// Query types
	QueryValidatorSecurityInfoRequest = pb.QueryValidatorSecurityInfoRequest
	QueryValidatorSecurityInfoResponse = pb.QueryValidatorSecurityInfoResponse
	QueryValidatorAlertsRequest       = pb.QueryValidatorAlertsRequest
	QueryValidatorAlertsResponse      = pb.QueryValidatorAlertsResponse
	QuerySentryNodesRequest           = pb.QuerySentryNodesRequest
	QuerySentryNodesResponse          = pb.QuerySentryNodesResponse
	QueryDoubleSignEvidencesRequest   = pb.QueryDoubleSignEvidencesRequest
	QueryDoubleSignEvidencesResponse  = pb.QueryDoubleSignEvidencesResponse
	QueryJailedValidatorsRequest      = pb.QueryJailedValidatorsRequest
	QueryJailedValidatorsResponse     = pb.QueryJailedValidatorsResponse
	QueryTombstonedValidatorsRequest  = pb.QueryTombstonedValidatorsRequest
	QueryTombstonedValidatorsResponse = pb.QueryTombstonedValidatorsResponse
	QueryAllValidatorsRequest         = pb.QueryAllValidatorsRequest
	QueryAllValidatorsResponse        = pb.QueryAllValidatorsResponse
	QueryParamsRequest                = pb.QueryParamsRequest
	QueryParamsResponse               = pb.QueryParamsResponse

	// Genesis types
	GenesisState = pb.GenesisState
)

// Re-export enum values for ValidatorAlert_Severity
const (
	ValidatorAlert_INFO     = pb.ValidatorAlert_INFO
	ValidatorAlert_WARNING  = pb.ValidatorAlert_WARNING
	ValidatorAlert_CRITICAL = pb.ValidatorAlert_CRITICAL
)

// Re-export enum values for ValidatorAlert_AlertType
const (
	ValidatorAlert_DOWNTIME             = pb.ValidatorAlert_DOWNTIME
	ValidatorAlert_DOUBLE_SIGN          = pb.ValidatorAlert_DOUBLE_SIGN
	ValidatorAlert_LOW_STAKE            = pb.ValidatorAlert_LOW_STAKE
	ValidatorAlert_SENTRY_NODE_OFFLINE  = pb.ValidatorAlert_SENTRY_NODE_OFFLINE
	ValidatorAlert_GEOGRAPHIC_VIOLATION = pb.ValidatorAlert_GEOGRAPHIC_VIOLATION
	ValidatorAlert_KEY_COMPROMISE       = pb.ValidatorAlert_KEY_COMPROMISE
	ValidatorAlert_FAILOVER_TRIGGERED   = pb.ValidatorAlert_FAILOVER_TRIGGERED
)

// Additional types for invariants (not in proto yet, using local definitions)
type (
	// ValidatorMonitoring represents validator monitoring data
	ValidatorMonitoring struct {
		ValidatorAddress  string
		UptimeBasisPoints uint64 // 10000 = 100%, deterministic integer arithmetic
		MissedBlocks      int64
		TotalBlocks       int64
		LastUpdated       *timestamppb.Timestamp
	}

	// JailingRecord represents a jailing record
	JailingRecord struct {
		ValidatorAddress  string
		Reason            string
		JailedAt          *timestamppb.Timestamp
		ReleaseTime       *timestamppb.Timestamp
		Permanent         bool
		Released          bool
		ActualReleaseTime *timestamppb.Timestamp
	}

	// SlashingRecord represents a slashing record
	SlashingRecord struct {
		ValidatorAddress string
		SlashAmount      string
		SlashFraction    string
		Reason           string
		SlashedAt        *timestamppb.Timestamp
		InfractionHeight int64
	}

	// SentryNode represents a sentry node configuration
	SentryNode struct {
		NodeId           string
		ValidatorAddress string
		IpAddress        string
		Port             int32
		Active           bool
		RegisteredAt     *timestamppb.Timestamp
		LastHeartbeat    *timestamppb.Timestamp
	}
)
