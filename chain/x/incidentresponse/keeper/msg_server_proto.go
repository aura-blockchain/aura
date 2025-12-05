package keeper

import (
	"context"
	"fmt"
	"time"

	incidentresponsepb "github.com/aequitas/aura/proto/aura/incidentresponse/v1beta1"
	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ incidentresponsepb.MsgServer = protoMsgServer{}

type protoMsgServer struct {
	incidentresponsepb.UnimplementedMsgServer
	keeper *KeeperKV
}

// NewProtoMsgServerImpl creates a new proto-based message server implementation
func NewProtoMsgServerImpl(keeper *KeeperKV) incidentresponsepb.MsgServer {
	return &protoMsgServer{keeper: keeper}
}

// ReportIncident creates a new security incident
func (ms protoMsgServer) ReportIncident(goCtx context.Context, msg *incidentresponsepb.MsgReportIncident) (*incidentresponsepb.MsgReportIncidentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.Title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if msg.Description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if msg.ReportedBy == "" {
		return nil, fmt.Errorf("reported_by cannot be empty")
	}

	incidentID, err := ms.keeper.ReportIncident(
		ctx,
		msg.Title,
		msg.Description,
		types.IncidentSeverity(msg.Severity),
		msg.ReportedBy,
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
			sdk.NewAttribute("reported_by", msg.ReportedBy),
			sdk.NewAttribute("severity", msg.Severity),
		),
	)

	return &incidentresponsepb.MsgReportIncidentResponse{
		IncidentId: incidentID,
	}, nil
}

// UpdateIncidentStatus updates the status of an existing incident
func (ms protoMsgServer) UpdateIncidentStatus(goCtx context.Context, msg *incidentresponsepb.MsgUpdateIncidentStatus) (*incidentresponsepb.MsgUpdateIncidentStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.IncidentId == "" {
		return nil, fmt.Errorf("incident_id cannot be empty")
	}
	if msg.Status == "" {
		return nil, fmt.Errorf("status cannot be empty")
	}
	if msg.UpdatedBy == "" {
		return nil, fmt.Errorf("updated_by cannot be empty")
	}

	err := ms.keeper.UpdateIncidentStatus(
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

	return &incidentresponsepb.MsgUpdateIncidentStatusResponse{
		Success: true,
	}, nil
}

// RequestChainPause initiates an emergency chain pause
func (ms protoMsgServer) RequestChainPause(goCtx context.Context, msg *incidentresponsepb.MsgRequestChainPause) (*incidentresponsepb.MsgRequestChainPauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.Requester == "" {
		return nil, fmt.Errorf("requester cannot be empty")
	}
	if msg.PauseLevel == "" {
		return nil, fmt.Errorf("pause_level cannot be empty")
	}

	var duration time.Duration
	if msg.Duration != nil {
		duration = msg.Duration.AsDuration()
	}

	err := ms.keeper.RequestChainPause(
		ctx,
		msg.Requester,
		types.PauseLevel(msg.PauseLevel),
		msg.Reason,
		msg.IncidentId,
		duration,
	)
	if err != nil {
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

	return &incidentresponsepb.MsgRequestChainPauseResponse{
		Success: true,
	}, nil
}

// ResumeChain resumes chain operations after a pause
func (ms protoMsgServer) ResumeChain(goCtx context.Context, msg *incidentresponsepb.MsgResumeChain) (*incidentresponsepb.MsgResumeChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.ResumedBy == "" {
		return nil, fmt.Errorf("resumed_by cannot be empty")
	}

	err := ms.keeper.ResumeChain(ctx, msg.ResumedBy, msg.Reason)
	if err != nil {
		return nil, fmt.Errorf("failed to resume chain: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"chain_resumed",
			sdk.NewAttribute("resumed_by", msg.ResumedBy),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &incidentresponsepb.MsgResumeChainResponse{
		Success: true,
	}, nil
}

// SetWalletLimits configures security limits for a hot wallet
func (ms protoMsgServer) SetWalletLimits(goCtx context.Context, msg *incidentresponsepb.MsgSetWalletLimits) (*incidentresponsepb.MsgSetWalletLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.Authority == "" {
		return nil, fmt.Errorf("authority cannot be empty")
	}
	if msg.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	err := ms.keeper.SetWalletLimits(
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

	return &incidentresponsepb.MsgSetWalletLimitsResponse{
		Success: true,
	}, nil
}

// CreatePostMortem creates a post-mortem analysis for an incident
func (ms protoMsgServer) CreatePostMortem(goCtx context.Context, msg *incidentresponsepb.MsgCreatePostMortem) (*incidentresponsepb.MsgCreatePostMortemResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.IncidentId == "" {
		return nil, fmt.Errorf("incident_id cannot be empty")
	}
	if msg.CreatedBy == "" {
		return nil, fmt.Errorf("created_by cannot be empty")
	}

	// Convert to action items (proto doesn't have full structure, using empty for now)
	var actionItems []types.ActionItem

	err := ms.keeper.CreatePostMortem(
		ctx,
		msg.IncidentId,
		msg.CreatedBy,
		msg.Summary,
		msg.RootCause,
		msg.Impact,
		msg.Resolution,
		msg.LessonsLearned,
		actionItems,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create post-mortem: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"post_mortem_created",
			sdk.NewAttribute("incident_id", msg.IncidentId),
			sdk.NewAttribute("created_by", msg.CreatedBy),
		),
	)

	return &incidentresponsepb.MsgCreatePostMortemResponse{
		Success: true,
	}, nil
}

// TriggerBackup initiates a manual backup operation
func (ms protoMsgServer) TriggerBackup(goCtx context.Context, msg *incidentresponsepb.MsgTriggerBackup) (*incidentresponsepb.MsgTriggerBackupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.BackupType == "" {
		return nil, fmt.Errorf("backup_type cannot be empty")
	}
	if msg.TriggeredBy == "" {
		return nil, fmt.Errorf("triggered_by cannot be empty")
	}

	backupID, err := ms.keeper.TriggerBackup(ctx, msg.BackupType, msg.TriggeredBy)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger backup: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"backup_triggered",
			sdk.NewAttribute("backup_id", backupID),
			sdk.NewAttribute("backup_type", msg.BackupType),
			sdk.NewAttribute("triggered_by", msg.TriggeredBy),
		),
	)

	return &incidentresponsepb.MsgTriggerBackupResponse{
		BackupId: backupID,
	}, nil
}

// TriggerInsuranceClaim submits an insurance claim for an incident
func (ms protoMsgServer) TriggerInsuranceClaim(goCtx context.Context, msg *incidentresponsepb.MsgTriggerInsuranceClaim) (*incidentresponsepb.MsgTriggerInsuranceClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, fmt.Errorf("empty request")
	}

	// Validate input
	if msg.IncidentId == "" {
		return nil, fmt.Errorf("incident_id cannot be empty")
	}
	if msg.Amount == "" {
		return nil, fmt.Errorf("amount cannot be empty")
	}
	if len(msg.Signers) == 0 {
		return nil, fmt.Errorf("signers cannot be empty")
	}

	claimID, err := ms.keeper.TriggerInsuranceClaim(
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

	return &incidentresponsepb.MsgTriggerInsuranceClaimResponse{
		ClaimId: claimID,
	}, nil
}
