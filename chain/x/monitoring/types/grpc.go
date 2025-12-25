// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
)

// Query service request/response types

type QueryNetworkHealthRequest struct{}

type QueryNetworkHealthResponse struct {
	Health *NetworkHealth `json:"health"`
}

type QueryValidatorUptimeRequest struct {
	ValidatorAddress string `json:"validator_address"`
}

type QueryValidatorUptimeResponse struct {
	Uptime *ValidatorUptime `json:"uptime"`
}

type QueryAlertsRequest struct {
	Severity string `json:"severity,omitempty"`
	Type     string `json:"type,omitempty"`
}

type QueryAlertsResponse struct {
	Alerts []*Alert `json:"alerts"`
}

type QueryGasPriceRequest struct{}

type QueryGasPriceResponse struct {
	Tracking *GasPriceTracking `json:"tracking"`
}

type QueryTVLRequest struct{}

type QueryTVLResponse struct {
	Tvl *TVLMonitoring `json:"tvl"`
}

// Message service request/response types

type MsgAcknowledgeAlert struct {
	AlertId        string `json:"alert_id"`
	AcknowledgedBy string `json:"acknowledged_by"`
}

type MsgAcknowledgeAlertResponse struct {
	Success bool `json:"success"`
}

type MsgResolveAlert struct {
	AlertId string `json:"alert_id"`
}

type MsgResolveAlertResponse struct {
	Success bool `json:"success"`
}

// QueryServer defines the gRPC query service interface
type QueryServer interface {
	GetNetworkHealth(context.Context, *QueryNetworkHealthRequest) (*QueryNetworkHealthResponse, error)
	GetValidatorUptime(context.Context, *QueryValidatorUptimeRequest) (*QueryValidatorUptimeResponse, error)
	GetAlerts(context.Context, *QueryAlertsRequest) (*QueryAlertsResponse, error)
	GetGasPriceTracking(context.Context, *QueryGasPriceRequest) (*QueryGasPriceResponse, error)
	GetTVLMonitoring(context.Context, *QueryTVLRequest) (*QueryTVLResponse, error)
}

// MsgServer defines the gRPC message service interface
type MsgServer interface {
	AcknowledgeAlert(context.Context, *MsgAcknowledgeAlert) (*MsgAcknowledgeAlertResponse, error)
	ResolveAlert(context.Context, *MsgResolveAlert) (*MsgResolveAlertResponse, error)
}

// RegisterQueryServer registers the query server (stub for compilation)
func RegisterQueryServer(server interface{}, impl QueryServer) {
	// In a real implementation, this would register with the gRPC server
}

// RegisterMsgServer registers the message server (stub for compilation)
func RegisterMsgServer(server interface{}, impl MsgServer) {
	// In a real implementation, this would register with the gRPC server
}
