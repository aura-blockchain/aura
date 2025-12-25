// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"

// Re-export all proto types
type (
	// Enums
	IRStatus    = pb.IRStatus
	PrivacyTier = pb.PrivacyTier
	Arena       = pb.Arena

	// Core types
	IRDefinition   = pb.IRDefinition
	IRPrerequisite = pb.IRPrerequisite
	IRGraphNode    = pb.IRGraphNode
	IRRateLimit    = pb.IRRateLimit
	Params         = pb.Params

	// Message types
	MsgCreateIR                 = pb.MsgCreateIR
	MsgCreateIRResponse         = pb.MsgCreateIRResponse
	MsgUpdateIR                 = pb.MsgUpdateIR
	MsgUpdateIRResponse         = pb.MsgUpdateIRResponse
	MsgDeleteIR                 = pb.MsgDeleteIR
	MsgDeleteIRResponse         = pb.MsgDeleteIRResponse
	MsgActivateIR               = pb.MsgActivateIR
	MsgActivateIRResponse       = pb.MsgActivateIRResponse
	MsgSuspendIR                = pb.MsgSuspendIR
	MsgSuspendIRResponse        = pb.MsgSuspendIRResponse
	MsgSetIRPrerequisites       = pb.MsgSetIRPrerequisites
	MsgSetIRPrerequisitesResponse = pb.MsgSetIRPrerequisitesResponse
	MsgSetIRRateLimit           = pb.MsgSetIRRateLimit
	MsgSetIRRateLimitResponse   = pb.MsgSetIRRateLimitResponse

	// Query types
	QueryIRRequest        = pb.QueryIRRequest
	QueryIRResponse       = pb.QueryIRResponse
	QueryListIRsRequest   = pb.QueryListIRsRequest
	QueryListIRsResponse  = pb.QueryListIRsResponse
	QueryIRGraphRequest   = pb.QueryIRGraphRequest
	QueryIRGraphResponse  = pb.QueryIRGraphResponse
	QueryRateLimitRequest = pb.QueryRateLimitRequest
	QueryRateLimitResponse = pb.QueryRateLimitResponse
	QueryParamsRequest    = pb.QueryParamsRequest
	QueryParamsResponse   = pb.QueryParamsResponse

	// Genesis types
	GenesisState = pb.GenesisState
)

// Re-export enum values for IRStatus
const (
	IRStatus_IR_STATUS_UNSPECIFIED = pb.IRStatus_IR_STATUS_UNSPECIFIED
	IRStatus_IR_STATUS_DRAFT       = pb.IRStatus_IR_STATUS_DRAFT
	IRStatus_IR_STATUS_REVIEWING   = pb.IRStatus_IR_STATUS_REVIEWING
	IRStatus_IR_STATUS_APPROVED    = pb.IRStatus_IR_STATUS_APPROVED
	IRStatus_IR_STATUS_ACTIVE      = pb.IRStatus_IR_STATUS_ACTIVE
	IRStatus_IR_STATUS_SUSPENDED   = pb.IRStatus_IR_STATUS_SUSPENDED
	IRStatus_IR_STATUS_DEPRECATED  = pb.IRStatus_IR_STATUS_DEPRECATED
	IRStatus_IR_STATUS_RETIRED     = pb.IRStatus_IR_STATUS_RETIRED
)

// Re-export enum values for PrivacyTier
const (
	PrivacyTier_PRIVACY_TIER_UNSPECIFIED = pb.PrivacyTier_PRIVACY_TIER_UNSPECIFIED
	PrivacyTier_PRIVACY_TIER_LOW         = pb.PrivacyTier_PRIVACY_TIER_LOW
	PrivacyTier_PRIVACY_TIER_MEDIUM      = pb.PrivacyTier_PRIVACY_TIER_MEDIUM
	PrivacyTier_PRIVACY_TIER_HIGH        = pb.PrivacyTier_PRIVACY_TIER_HIGH
)

// Re-export enum values for Arena
const (
	Arena_ARENA_UNSPECIFIED    = pb.Arena_ARENA_UNSPECIFIED
	Arena_ARENA_ANCHOR         = pb.Arena_ARENA_ANCHOR
	Arena_ARENA_BIOMETRIC      = pb.Arena_ARENA_BIOMETRIC
	Arena_ARENA_POSSESSION     = pb.Arena_ARENA_POSSESSION
	Arena_ARENA_KNOWLEDGE      = pb.Arena_ARENA_KNOWLEDGE
	Arena_ARENA_SOCIAL         = pb.Arena_ARENA_SOCIAL
	Arena_ARENA_GEOLOCATION    = pb.Arena_ARENA_GEOLOCATION
	Arena_ARENA_HIGH_ASSURANCE = pb.Arena_ARENA_HIGH_ASSURANCE
	Arena_ARENA_PERSISTENCE    = pb.Arena_ARENA_PERSISTENCE
	Arena_ARENA_SPECIALIZED    = pb.Arena_ARENA_SPECIALIZED
)
