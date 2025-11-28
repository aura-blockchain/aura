package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// QueryServer implements the gRPC query service for the monitoring module
type QueryServer struct {
	keeper *Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(k *Keeper) types.QueryServer {
	return &QueryServer{keeper: k}
}

// Ensure QueryServer implements the QueryServer interface
var _ types.QueryServer = &QueryServer{}

// GetNetworkHealth returns network health information
func (qs *QueryServer) GetNetworkHealth(ctx context.Context, req *types.QueryNetworkHealthRequest) (*types.QueryNetworkHealthResponse, error) {
	if req == nil {
		req = &types.QueryNetworkHealthRequest{}
	}

	health, err := qs.keeper.GetNetworkHealth(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryNetworkHealthResponse{
		Health: health,
	}, nil
}

// GetValidatorUptime returns validator uptime information
func (qs *QueryServer) GetValidatorUptime(ctx context.Context, req *types.QueryValidatorUptimeRequest) (*types.QueryValidatorUptimeResponse, error) {
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

	return &types.QueryValidatorUptimeResponse{
		Uptime: uptime,
	}, nil
}

// GetAlerts returns active alerts, optionally filtered by severity and/or type
func (qs *QueryServer) GetAlerts(ctx context.Context, req *types.QueryAlertsRequest) (*types.QueryAlertsResponse, error) {
	if req == nil {
		req = &types.QueryAlertsRequest{}
	}

	var alerts []*types.Alert
	var err error

	// If both filters are empty, return all active alerts
	if req.Severity == "" && req.Type == "" {
		alerts, err = qs.keeper.GetActiveAlerts(ctx)
		if err != nil {
			return nil, err
		}
	} else if req.Severity != "" && req.Type != "" {
		// Filter by both severity and type
		allAlerts, err := qs.keeper.GetActiveAlerts(ctx)
		if err != nil {
			return nil, err
		}
		for _, alert := range allAlerts {
			if string(alert.Severity) == req.Severity && string(alert.Type) == req.Type {
				alerts = append(alerts, alert)
			}
		}
	} else if req.Severity != "" {
		// Filter by severity only
		alerts, err = qs.keeper.GetAlertsBySeverity(ctx, types.AlertSeverity(req.Severity))
		if err != nil {
			return nil, err
		}
	} else if req.Type != "" {
		// Filter by type only
		alerts, err = qs.keeper.GetAlertsByType(ctx, types.AlertType(req.Type))
		if err != nil {
			return nil, err
		}
	}

	// Ensure alerts is never nil, return empty slice instead
	if alerts == nil {
		alerts = []*types.Alert{}
	}

	return &types.QueryAlertsResponse{
		Alerts: alerts,
	}, nil
}

// GetGasPriceTracking returns gas price tracking data
func (qs *QueryServer) GetGasPriceTracking(ctx context.Context, req *types.QueryGasPriceRequest) (*types.QueryGasPriceResponse, error) {
	if req == nil {
		req = &types.QueryGasPriceRequest{}
	}

	tracking, err := qs.keeper.GetGasPriceTracking(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryGasPriceResponse{
		Tracking: tracking,
	}, nil
}

// GetTVLMonitoring returns TVL monitoring data
func (qs *QueryServer) GetTVLMonitoring(ctx context.Context, req *types.QueryTVLRequest) (*types.QueryTVLResponse, error) {
	if req == nil {
		req = &types.QueryTVLRequest{}
	}

	tvl, err := qs.keeper.GetTVLMonitoring(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryTVLResponse{
		Tvl: tvl,
	}, nil
}
