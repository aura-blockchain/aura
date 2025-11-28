package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// MsgServer implements the gRPC message service for the monitoring module
type MsgServer struct {
	keeper *Keeper
}

// NewMsgServer creates a new message server
func NewMsgServer(k *Keeper) types.MsgServer {
	return &MsgServer{keeper: k}
}

// Ensure MsgServer implements the MsgServer interface
var _ types.MsgServer = &MsgServer{}

// AcknowledgeAlert acknowledges an alert
func (ms *MsgServer) AcknowledgeAlert(ctx context.Context, msg *types.MsgAcknowledgeAlert) (*types.MsgAcknowledgeAlertResponse, error) {
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

	return &types.MsgAcknowledgeAlertResponse{
		Success: true,
	}, nil
}

// ResolveAlert resolves an alert
func (ms *MsgServer) ResolveAlert(ctx context.Context, msg *types.MsgResolveAlert) (*types.MsgResolveAlertResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidTransaction
	}

	// Validate input
	if msg.AlertId == "" {
		return nil, types.ErrAlertNotFound
	}

	// Resolve the alert
	err := ms.keeper.ResolveAlert(ctx, msg.AlertId)
	if err != nil {
		return nil, err
	}

	// Emit event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"alert_resolved",
		sdk.NewAttribute("alert_id", msg.AlertId),
	))

	return &types.MsgResolveAlertResponse{
		Success: true,
	}, nil
}

// NOTE: AcknowledgeAlert and ResolveAlert are implemented in alerts.go
