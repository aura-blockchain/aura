package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
	monitoringpb "github.com/aequitas/aura/proto/aura/monitoring/v1beta1"
)

// MsgServer implements the gRPC message service for the monitoring module
type MsgServer struct {
	monitoringpb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer creates a new message server
func NewMsgServer(k *Keeper) monitoringpb.MsgServer {
	return &MsgServer{keeper: k}
}

// Ensure MsgServer implements the MsgServer interface
var _ monitoringpb.MsgServer = &MsgServer{}

// AcknowledgeAlert acknowledges an alert
func (ms *MsgServer) AcknowledgeAlert(ctx context.Context, msg *monitoringpb.MsgAcknowledgeAlert) (*monitoringpb.MsgAcknowledgeAlertResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidTransaction
	}

	// Validate input
	if msg.AlertId == "" {
		return nil, types.ErrAlertNotFound
	}

	if msg.AcknowledgedBy == "" {
		return nil, types.ErrInvalidTransaction
	}

	// Acknowledge the alert
	err := ms.keeper.AcknowledgeAlert(ctx, msg.AlertId, msg.AcknowledgedBy)
	if err != nil {
		return nil, err
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"alert_acknowledged",
		sdk.NewAttribute("alert_id", msg.AlertId),
		sdk.NewAttribute("acknowledged_by", msg.AcknowledgedBy),
	))

	return &monitoringpb.MsgAcknowledgeAlertResponse{
		Success: true,
	}, nil
}

// ResolveAlert resolves an alert
func (ms *MsgServer) ResolveAlert(ctx context.Context, msg *monitoringpb.MsgResolveAlert) (*monitoringpb.MsgResolveAlertResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidTransaction
	}

	// Validate input
	if msg.AlertId == "" {
		return nil, types.ErrAlertNotFound
	}

	// Resolve the alert (using Resolver field from proto)
	err := ms.keeper.ResolveAlert(ctx, msg.AlertId)
	if err != nil {
		return nil, err
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"alert_resolved",
		sdk.NewAttribute("alert_id", msg.AlertId),
		sdk.NewAttribute("resolver", msg.Resolver),
	))

	return &monitoringpb.MsgResolveAlertResponse{
		Success: true,
	}, nil
}

// NOTE: AcknowledgeAlert and ResolveAlert are implemented in alerts.go
