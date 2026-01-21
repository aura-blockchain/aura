// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/aiassistant/v1beta1"

type (
	Assistant       = pb.Assistant
	AssistantStatus = pb.AssistantStatus
	Params          = pb.Params
	GenesisState    = pb.GenesisState
	Balance         = pb.Balance

	MsgRegisterAssistant         = pb.MsgRegisterAssistant
	MsgRegisterAssistantResponse = pb.MsgRegisterAssistantResponse
	MsgUpdateLocales             = pb.MsgUpdateLocales
	MsgUpdateLocalesResponse     = pb.MsgUpdateLocalesResponse
	MsgHeartbeat                 = pb.MsgHeartbeat
	MsgHeartbeatResponse         = pb.MsgHeartbeatResponse
	MsgReportMisbehavior         = pb.MsgReportMisbehavior
	MsgReportMisbehaviorResponse = pb.MsgReportMisbehaviorResponse
	MsgUpdateParams              = pb.MsgUpdateParams
	MsgUpdateParamsResponse      = pb.MsgUpdateParamsResponse

	QueryAssistantRequest           = pb.QueryAssistantRequest
	QueryAssistantResponse          = pb.QueryAssistantResponse
	QueryAssistantsRequest          = pb.QueryAssistantsRequest
	QueryAssistantsResponse         = pb.QueryAssistantsResponse
	QueryAssistantsByLocaleRequest  = pb.QueryAssistantsByLocaleRequest
	QueryAssistantsByLocaleResponse = pb.QueryAssistantsByLocaleResponse
	QueryParamsRequest              = pb.QueryParamsRequest
	QueryParamsResponse             = pb.QueryParamsResponse

	MsgServer                = pb.MsgServer
	QueryServer              = pb.QueryServer
	UnimplementedMsgServer   = pb.UnimplementedMsgServer
	UnimplementedQueryServer = pb.UnimplementedQueryServer
)

const (
	AssistantStatus_UNSPECIFIED = pb.AssistantStatus_ASSISTANT_STATUS_UNSPECIFIED
	AssistantStatus_ACTIVE      = pb.AssistantStatus_ASSISTANT_STATUS_ACTIVE
	AssistantStatus_JAILED      = pb.AssistantStatus_ASSISTANT_STATUS_JAILED
	AssistantStatus_TOMBSTONED  = pb.AssistantStatus_ASSISTANT_STATUS_TOMBSTONED
)
