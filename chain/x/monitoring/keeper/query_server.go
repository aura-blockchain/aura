// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	monitoringpb "github.com/aequitas/aura/proto/aura/monitoring/v1beta1"
)

// QueryServer implements the gRPC query service for the monitoring module
type QueryServer struct {
	monitoringpb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(k *Keeper) monitoringpb.QueryServer {
	return &QueryServer{keeper: k}
}

// Ensure QueryServer implements the QueryServer interface
var _ monitoringpb.QueryServer = &QueryServer{}

// GetNetworkHealth returns network health information
func (qs *QueryServer) GetNetworkHealth(ctx context.Context, _ *monitoringpb.QueryNetworkHealthRequest) (*monitoringpb.QueryNetworkHealthResponse, error) {
	health, err := qs.keeper.GetNetworkHealth(ctx)
	if err != nil {
		return nil, err
	}

	return &monitoringpb.QueryNetworkHealthResponse{
		Health: convertNetworkHealthToProto(health),
	}, nil
}

// GetValidatorUptime returns validator uptime information
func (qs *QueryServer) GetValidatorUptime(ctx context.Context, req *monitoringpb.QueryValidatorUptimeRequest) (*monitoringpb.QueryValidatorUptimeResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidTransaction
	}

	if req.ValidatorAddress == "" {
		return nil, types.ErrValidatorNotFound
	}

	uptime, err := qs.keeper.GetValidatorUptime(ctx, req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &monitoringpb.QueryValidatorUptimeResponse{
		Uptime: convertValidatorUptimeToProto(uptime),
	}, nil
}

// GetAlerts returns active alerts, optionally filtered by severity and/or type
func (qs *QueryServer) GetAlerts(ctx context.Context, req *monitoringpb.QueryAlertsRequest) (*monitoringpb.QueryAlertsResponse, error) {
	if req == nil {
		req = &monitoringpb.QueryAlertsRequest{}
	}

	var alerts []*monitoringpb.Alert

	// If both filters are empty, return all active alerts
	if req.Severity == "" && req.Type == "" {
		typeAlerts, err := qs.keeper.GetActiveAlerts(ctx)
		if err != nil {
			return nil, err
		}
		// Convert types.Alert to monitoringpb.Alert
		alerts = convertAlertsToProto(typeAlerts)
	} else if req.Severity != "" && req.Type != "" {
		// Filter by both severity and type
		allAlerts, err := qs.keeper.GetActiveAlerts(ctx)
		if err != nil {
			return nil, err
		}
		for _, alert := range allAlerts {
			if string(alert.Severity) == req.Severity && string(alert.Type) == req.Type {
				alerts = append(alerts, convertAlertToProto(alert))
			}
		}
	} else if req.Severity != "" {
		// Filter by severity only
		typeAlerts, err := qs.keeper.GetAlertsBySeverity(ctx, types.AlertSeverity(req.Severity))
		if err != nil {
			return nil, err
		}
		alerts = convertAlertsToProto(typeAlerts)
	} else if req.Type != "" {
		// Filter by type only
		typeAlerts, err := qs.keeper.GetAlertsByType(ctx, types.AlertType(req.Type))
		if err != nil {
			return nil, err
		}
		alerts = convertAlertsToProto(typeAlerts)
	}

	// Ensure alerts is never nil, return empty slice instead
	if alerts == nil {
		alerts = []*monitoringpb.Alert{}
	}

	return &monitoringpb.QueryAlertsResponse{
		Alerts: alerts,
	}, nil
}

// GetGasPriceTracking returns gas price tracking data
func (qs *QueryServer) GetGasPriceTracking(ctx context.Context, _ *monitoringpb.QueryGasPriceRequest) (*monitoringpb.QueryGasPriceResponse, error) {
	tracking, err := qs.keeper.GetGasPriceTracking(ctx)
	if err != nil {
		return nil, err
	}

	return &monitoringpb.QueryGasPriceResponse{
		Tracking: convertGasPriceTrackingToProto(tracking),
	}, nil
}

// GetTVLMonitoring returns TVL monitoring data
func (qs *QueryServer) GetTVLMonitoring(ctx context.Context, _ *monitoringpb.QueryTVLRequest) (*monitoringpb.QueryTVLResponse, error) {
	tvl, err := qs.keeper.GetTVLMonitoring(ctx)
	if err != nil {
		return nil, err
	}

	return &monitoringpb.QueryTVLResponse{
		Tvl: convertTVLMonitoringToProto(tvl),
	}, nil
}

// Helper functions to convert types to protobuf types
func convertAlertToProto(alert *types.Alert) *monitoringpb.Alert {
	// Map string-based alert types to proto enums
	var alertType monitoringpb.AlertType
	switch alert.Type {
	case types.AlertTypeSecurityThreat:
		alertType = monitoringpb.AlertType_ALERT_TYPE_SECURITY_THREAT
	case types.AlertTypeGasPriceSpike:
		alertType = monitoringpb.AlertType_ALERT_TYPE_HIGH_GAS_PRICE
	case types.AlertTypeTVLChange:
		alertType = monitoringpb.AlertType_ALERT_TYPE_LOW_TVL
	case types.AlertTypeNetworkCongestion:
		alertType = monitoringpb.AlertType_ALERT_TYPE_NETWORK_CONGESTION
	case types.AlertTypeValidatorDown:
		alertType = monitoringpb.AlertType_ALERT_TYPE_VALIDATOR_DOWN
	default:
		alertType = monitoringpb.AlertType_ALERT_TYPE_UNSPECIFIED
	}

	// Map string-based severity to proto enums
	var severity monitoringpb.AlertSeverity
	switch alert.Severity {
	case types.SeverityInfo:
		severity = monitoringpb.AlertSeverity_ALERT_SEVERITY_INFO
	case types.SeverityLow:
		severity = monitoringpb.AlertSeverity_ALERT_SEVERITY_WARNING
	case types.SeverityMedium, types.SeverityHigh:
		severity = monitoringpb.AlertSeverity_ALERT_SEVERITY_ERROR
	case types.SeverityCritical:
		severity = monitoringpb.AlertSeverity_ALERT_SEVERITY_CRITICAL
	default:
		severity = monitoringpb.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED
	}

	return &monitoringpb.Alert{
		Id:             alert.ID,
		Type:           alertType,
		Severity:       severity,
		Message:        alert.Message,
		Timestamp:      alert.Timestamp.Unix(),
		Acknowledged:   alert.Acknowledged,
		AcknowledgedBy: alert.AcknowledgedBy,
		Resolved:       alert.Resolved,
	}
}

func convertAlertsToProto(alerts []*types.Alert) []*monitoringpb.Alert {
	result := make([]*monitoringpb.Alert, len(alerts))
	for i, alert := range alerts {
		result[i] = convertAlertToProto(alert)
	}
	return result
}

func convertGasPriceTrackingToProto(tracking *types.GasPriceTracking) *monitoringpb.GasPriceTracking {
	if tracking == nil {
		return nil
	}
	return &monitoringpb.GasPriceTracking{
		CurrentPrice: fmt.Sprintf("%d", tracking.CurrentPrice),
		AveragePrice: fmt.Sprintf("%d", tracking.AveragePrice),
		MinPrice:     fmt.Sprintf("%d", tracking.MinPrice),
		MaxPrice:     fmt.Sprintf("%d", tracking.MaxPrice),
		LastUpdated:  tracking.Timestamp.Unix(),
	}
}

func convertTVLMonitoringToProto(tvl *types.TVLMonitoring) *monitoringpb.TVLMonitoring {
	if tvl == nil {
		return nil
	}
	poolValues := make(map[string]string)
	for k, v := range tvl.TVLByModule {
		poolValues[k] = fmt.Sprintf("%d", v)
	}
	return &monitoringpb.TVLMonitoring{
		TotalValueLocked: fmt.Sprintf("%d", tvl.TotalTVL),
		PoolValues:       poolValues,
		LastUpdated:      tvl.Timestamp.Unix(),
	}
}

func convertNetworkHealthToProto(health *types.NetworkHealth) *monitoringpb.NetworkHealth {
	if health == nil {
		return nil
	}
	return &monitoringpb.NetworkHealth{
		IsHealthy:         health.ConsensusHealth > 0.5, // Consider healthy if > 50%
		BlockHeight:       health.BlockHeight,
		BlockTimeSeconds:  health.BlockTime,
		ActiveValidators:  int32(health.ActiveValidators),
		NetworkHashRate:   health.NetworkHashRate,
		TotalTransactions: health.BlockHeight, // Approximate - actual TPS calculation needed
	}
}

func convertValidatorUptimeToProto(uptime *types.ValidatorUptime) *monitoringpb.ValidatorUptime {
	if uptime == nil {
		return nil
	}
	return &monitoringpb.ValidatorUptime{
		ValidatorAddress:  uptime.ValidatorAddress,
		UptimePercentage:  uptime.UptimePercentage,
		BlocksSigned:      uptime.SignedBlocks,
		BlocksMissed:      uptime.MissedBlocks,
		LastActiveHeight:  uptime.LastSeen.Unix(),
	}
}
