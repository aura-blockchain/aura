package keeper

import (
	"errors"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper *KeeperKV
}

// NewMsgServerImpl creates a new message server implementation
func NewMsgServerImpl(keeper *KeeperKV) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// ReportIncident creates a new security incident
func (ms msgServer) ReportIncident(goCtx interface{}, msg *types.MsgReportIncident) (*types.MsgReportIncidentResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	incidentID, err := ms.Keeper.ReportIncident(
		ctx,
		msg.Title,
		msg.Description,
		types.IncidentSeverity(msg.Severity),
		msg.Reporter,
		msg.AffectedSystems,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to report incident: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"incident_reported",
			sdk.NewAttribute("incident_id", incidentID),
			sdk.NewAttribute("reporter", msg.Reporter),
			sdk.NewAttribute("severity", msg.Severity),
		),
	)

	return &types.MsgReportIncidentResponse{
		IncidentId: incidentID,
	}, nil
}

// UpdateIncidentStatus updates the status of an existing incident
func (ms msgServer) UpdateIncidentStatus(goCtx interface{}, msg *types.MsgUpdateIncidentStatus) (*types.MsgUpdateIncidentStatusResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	err := ms.Keeper.UpdateIncidentStatus(
		ctx,
		msg.IncidentId,
		types.IncidentStatus(msg.Status),
		msg.UpdatedBy,
		msg.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update incident status: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"incident_status_updated",
			sdk.NewAttribute("incident_id", msg.IncidentId),
			sdk.NewAttribute("status", msg.Status),
			sdk.NewAttribute("updated_by", msg.UpdatedBy),
		),
	)

	return &types.MsgUpdateIncidentStatusResponse{}, nil
}

// RequestChainPause initiates an emergency chain pause
func (ms msgServer) RequestChainPause(goCtx interface{}, msg *types.MsgRequestChainPause) (*types.MsgRequestChainPauseResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	duration := time.Duration(msg.DurationSeconds) * time.Second

	err := ms.Keeper.RequestChainPause(
		ctx,
		msg.Requester,
		types.PauseLevel(msg.PauseLevel),
		msg.Reason,
		msg.IncidentId,
		duration,
	)
	if err != nil {
		if errors.Is(err, types.ErrUnauthorizedPause) {
			return nil, mapUnauthorizedPause(err)
		}
		return nil, fmt.Errorf("failed to request chain pause: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"chain_pause_requested",
			sdk.NewAttribute("requester", msg.Requester),
			sdk.NewAttribute("pause_level", msg.PauseLevel),
			sdk.NewAttribute("incident_id", msg.IncidentId),
		),
	)

	return &types.MsgRequestChainPauseResponse{}, nil
}

// ResumeChain resumes chain operations after a pause
func (ms msgServer) ResumeChain(goCtx interface{}, msg *types.MsgResumeChain) (*types.MsgResumeChainResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	err := ms.Keeper.ResumeChain(ctx, msg.Resumer, msg.Reason)
	if err != nil {
		if errors.Is(err, types.ErrUnauthorizedPause) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, fmt.Errorf("failed to resume chain: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"chain_resumed",
			sdk.NewAttribute("resumer", msg.Resumer),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &types.MsgResumeChainResponse{}, nil
}

// SetWalletLimits configures security limits for a hot wallet
func (ms msgServer) SetWalletLimits(goCtx interface{}, msg *types.MsgSetWalletLimits) (*types.MsgSetWalletLimitsResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	err := ms.Keeper.SetWalletLimits(
		ctx,
		msg.Address,
		msg.MaxBalance,
		msg.MaxTransactionSize,
		msg.DailyLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set wallet limits: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"wallet_limits_set",
			sdk.NewAttribute("address", msg.Address),
			sdk.NewAttribute("max_balance", msg.MaxBalance),
		),
	)

	return &types.MsgSetWalletLimitsResponse{}, nil
}

// CreatePostMortem creates a post-mortem analysis for an incident
func (ms msgServer) CreatePostMortem(goCtx interface{}, msg *types.MsgCreatePostMortem) (*types.MsgCreatePostMortemResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	err := ms.Keeper.CreatePostMortem(
		ctx,
		msg.IncidentId,
		msg.Creator,
		msg.Summary,
		msg.RootCause,
		msg.Impact,
		msg.Resolution,
		msg.LessonsLearned,
		msg.ActionItems,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create post-mortem: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"post_mortem_created",
			sdk.NewAttribute("incident_id", msg.IncidentId),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgCreatePostMortemResponse{}, nil
}

// CloseIncident closes an incident after post-mortem is complete
func (ms msgServer) CloseIncident(goCtx interface{}, msg *types.MsgCloseIncident) (*types.MsgCloseIncidentResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	err := ms.Keeper.CloseIncident(ctx, msg.IncidentId, msg.Closer)
	if err != nil {
		return nil, fmt.Errorf("failed to close incident: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"incident_closed",
			sdk.NewAttribute("incident_id", msg.IncidentId),
			sdk.NewAttribute("closer", msg.Closer),
		),
	)

	return &types.MsgCloseIncidentResponse{}, nil
}

// TriggerBackup initiates a manual backup operation
func (ms msgServer) TriggerBackup(goCtx interface{}, msg *types.MsgTriggerBackup) (*types.MsgTriggerBackupResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	backupID, err := ms.Keeper.TriggerBackup(ctx, msg.BackupType, msg.Requester)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger backup: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"backup_triggered",
			sdk.NewAttribute("backup_id", backupID),
			sdk.NewAttribute("backup_type", msg.BackupType),
			sdk.NewAttribute("requester", msg.Requester),
		),
	)

	return &types.MsgTriggerBackupResponse{
		BackupId: backupID,
	}, nil
}

// TriggerInsuranceClaim submits an insurance claim for an incident
func (ms msgServer) TriggerInsuranceClaim(goCtx interface{}, msg *types.MsgTriggerInsuranceClaim) (*types.MsgTriggerInsuranceClaimResponse, error) {
	ctx := goCtx.(sdk.Context)

	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	claimID, err := ms.Keeper.TriggerInsuranceClaim(
		ctx,
		msg.IncidentId,
		msg.Amount,
		msg.Signers,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger insurance claim: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"insurance_claim_triggered",
			sdk.NewAttribute("claim_id", claimID),
			sdk.NewAttribute("incident_id", msg.IncidentId),
			sdk.NewAttribute("amount", msg.Amount),
		),
	)

	return &types.MsgTriggerInsuranceClaimResponse{
		ClaimId: claimID,
	}, nil
}
