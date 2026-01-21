// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"

// Re-export all proto types
type (
	// Enums
	IdentityChangeStatus = pb.IdentityChangeStatus

	// Core types
	IdentityRecord        = pb.IdentityRecord
	IdentityChangeRequest = pb.IdentityChangeRequest
	IdentityChangeHistory = pb.IdentityChangeHistory
	Params                = pb.Params

	// Message types
	MsgRequestIdentityChange          = pb.MsgRequestIdentityChange
	MsgRequestIdentityChangeResponse  = pb.MsgRequestIdentityChangeResponse
	MsgSubmitAssistantProof           = pb.MsgSubmitAssistantProof
	MsgSubmitAssistantProofResponse   = pb.MsgSubmitAssistantProofResponse
	MsgApplyIdentityChange            = pb.MsgApplyIdentityChange
	MsgApplyIdentityChangeResponse    = pb.MsgApplyIdentityChangeResponse
	MsgRejectIdentityChange           = pb.MsgRejectIdentityChange
	MsgRejectIdentityChangeResponse   = pb.MsgRejectIdentityChangeResponse
	MsgSuspendIdentityChanges         = pb.MsgSuspendIdentityChanges
	MsgSuspendIdentityChangesResponse = pb.MsgSuspendIdentityChangesResponse

	// Query types
	QueryIdentityRecordRequest         = pb.QueryIdentityRecordRequest
	QueryIdentityRecordResponse        = pb.QueryIdentityRecordResponse
	QueryIdentityChangeRequestRequest  = pb.QueryIdentityChangeRequestRequest
	QueryIdentityChangeRequestResponse = pb.QueryIdentityChangeRequestResponse
	QueryIdentityChangeHistoryRequest  = pb.QueryIdentityChangeHistoryRequest
	QueryIdentityChangeHistoryResponse = pb.QueryIdentityChangeHistoryResponse

	// Genesis types
	GenesisState = pb.GenesisState
)

// Re-export enum values for IdentityChangeStatus
const (
	IdentityChangeStatus_IDENTITY_CHANGE_STATUS_UNSPECIFIED          = pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_UNSPECIFIED
	IdentityChangeStatus_IDENTITY_CHANGE_STATUS_IDLE                 = pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_IDLE
	IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION = pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION
	IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY       = pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY
	IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED             = pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED
	IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED              = pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED
	IdentityChangeStatus_IDENTITY_CHANGE_STATUS_SUSPENDED            = pb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_SUSPENDED
)
