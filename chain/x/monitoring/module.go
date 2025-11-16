package monitoring

import (
	"context"

	"github.com/aequitas/aura/chain/x/monitoring/keeper"
	"github.com/aequitas/aura/chain/x/monitoring/types"
	"google.golang.org/grpc"
)

// AppModule represents the monitoring application module
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new monitoring module
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

// RegisterGRPCServices registers the module's gRPC services
func (am AppModule) RegisterGRPCServices(server *grpc.Server) {
	types.RegisterQueryServer(server, NewQueryServer(am.keeper))
	types.RegisterMsgServer(server, NewMsgServer(am.keeper))
}

// GetKeeper returns the module's keeper
func (am AppModule) GetKeeper() *keeper.Keeper {
	return am.keeper
}

// QueryServer implements the gRPC query service
type QueryServer struct {
	keeper *keeper.Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(k *keeper.Keeper) *QueryServer {
	return &QueryServer{keeper: k}
}

// GetNetworkHealth returns network health information
func (qs *QueryServer) GetNetworkHealth(ctx context.Context, req *types.QueryNetworkHealthRequest) (*types.QueryNetworkHealthResponse, error) {
	health := qs.keeper.GetNetworkHealth()
	return &types.QueryNetworkHealthResponse{
		Health: health,
	}, nil
}

// GetValidatorUptime returns validator uptime information
func (qs *QueryServer) GetValidatorUptime(ctx context.Context, req *types.QueryValidatorUptimeRequest) (*types.QueryValidatorUptimeResponse, error) {
	uptime, err := qs.keeper.GetValidatorUptime(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	return &types.QueryValidatorUptimeResponse{
		Uptime: uptime,
	}, nil
}

// GetAlerts returns active alerts
func (qs *QueryServer) GetAlerts(ctx context.Context, req *types.QueryAlertsRequest) (*types.QueryAlertsResponse, error) {
	alerts := qs.keeper.GetActiveAlerts()
	return &types.QueryAlertsResponse{
		Alerts: alerts,
	}, nil
}

// GetGasPriceTracking returns gas price tracking data
func (qs *QueryServer) GetGasPriceTracking(ctx context.Context, req *types.QueryGasPriceRequest) (*types.QueryGasPriceResponse, error) {
	tracking := qs.keeper.GetGasPriceTracking()
	return &types.QueryGasPriceResponse{
		Tracking: tracking,
	}, nil
}

// GetTVLMonitoring returns TVL monitoring data
func (qs *QueryServer) GetTVLMonitoring(ctx context.Context, req *types.QueryTVLRequest) (*types.QueryTVLResponse, error) {
	tvl := qs.keeper.GetTVLMonitoring()
	return &types.QueryTVLResponse{
		Tvl: tvl,
	}, nil
}

// MsgServer implements the gRPC message service
type MsgServer struct {
	keeper *keeper.Keeper
}

// NewMsgServer creates a new message server
func NewMsgServer(k *keeper.Keeper) *MsgServer {
	return &MsgServer{keeper: k}
}

// AcknowledgeAlert acknowledges an alert
func (ms *MsgServer) AcknowledgeAlert(ctx context.Context, msg *types.MsgAcknowledgeAlert) (*types.MsgAcknowledgeAlertResponse, error) {
	err := ms.keeper.AcknowledgeAlert(msg.AlertId, msg.AcknowledgedBy)
	if err != nil {
		return nil, err
	}
	return &types.MsgAcknowledgeAlertResponse{Success: true}, nil
}

// ResolveAlert resolves an alert
func (ms *MsgServer) ResolveAlert(ctx context.Context, msg *types.MsgResolveAlert) (*types.MsgResolveAlertResponse, error) {
	err := ms.keeper.ResolveAlert(msg.AlertId)
	if err != nil {
		return nil, err
	}
	return &types.MsgResolveAlertResponse{Success: true}, nil
}
