package incidentresponse

import (
	"context"

	"github.com/aequitas/aura/chain/x/incidentresponse/keeper"
	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	"google.golang.org/grpc"
)

// AppModule represents the incident response module
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new incident response module
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		keeper: k,
	}
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (am AppModule) RegisterGRPCGatewayRoutes(server *grpc.Server) {
	types.RegisterIncidentResponseServiceServer(server, &grpcServer{keeper: am.keeper})
}

// grpcServer implements the gRPC service
type grpcServer struct {
	types.UnimplementedIncidentResponseServiceServer
	keeper *keeper.Keeper
}

// ReportIncident handles incident reporting
func (s *grpcServer) ReportIncident(ctx context.Context, req *types.ReportIncidentRequest) (*types.ReportIncidentResponse, error) {
	incidentID, err := s.keeper.ReportIncident(
		req.Title,
		req.Description,
		req.Severity,
		req.ReportedBy,
		req.AffectedSystems,
	)
	if err != nil {
		return nil, err
	}

	return &types.ReportIncidentResponse{
		IncidentId: incidentID,
	}, nil
}

// UpdateIncidentStatus updates incident status
func (s *grpcServer) UpdateIncidentStatus(ctx context.Context, req *types.UpdateIncidentStatusRequest) (*types.UpdateIncidentStatusResponse, error) {
	err := s.keeper.UpdateIncidentStatus(
		req.IncidentId,
		req.Status,
		req.UpdatedBy,
		req.Notes,
	)
	if err != nil {
		return nil, err
	}

	return &types.UpdateIncidentStatusResponse{
		Success: true,
	}, nil
}

// GetIncident retrieves incident details
func (s *grpcServer) GetIncident(ctx context.Context, req *types.GetIncidentRequest) (*types.GetIncidentResponse, error) {
	incident, err := s.keeper.GetIncident(req.IncidentId)
	if err != nil {
		return nil, err
	}

	return &types.GetIncidentResponse{
		Incident: incident,
	}, nil
}

// RequestChainPause handles emergency chain pause requests
func (s *grpcServer) RequestChainPause(ctx context.Context, req *types.RequestChainPauseRequest) (*types.RequestChainPauseResponse, error) {
	err := s.keeper.RequestChainPause(
		req.Requester,
		req.PauseLevel,
		req.Reason,
		req.IncidentId,
		req.Duration,
	)
	if err != nil {
		return nil, err
	}

	return &types.RequestChainPauseResponse{
		Success: true,
	}, nil
}

// ResumeChain handles chain resume requests
func (s *grpcServer) ResumeChain(ctx context.Context, req *types.ResumeChainRequest) (*types.ResumeChainResponse, error) {
	err := s.keeper.ResumeChain(req.ResumedBy, req.Reason)
	if err != nil {
		return nil, err
	}

	return &types.ResumeChainResponse{
		Success: true,
	}, nil
}

// GetChainPauseState retrieves current pause state
func (s *grpcServer) GetChainPauseState(ctx context.Context, req *types.GetChainPauseStateRequest) (*types.GetChainPauseStateResponse, error) {
	state := s.keeper.GetChainPauseState()
	return &types.GetChainPauseStateResponse{
		PauseState: state,
	}, nil
}

// SetWalletLimits configures hot wallet limits
func (s *grpcServer) SetWalletLimits(ctx context.Context, req *types.SetWalletLimitsRequest) (*types.SetWalletLimitsResponse, error) {
	err := s.keeper.SetWalletLimits(
		req.Address,
		req.MaxBalance,
		req.MaxTransactionSize,
		req.DailyLimit,
	)
	if err != nil {
		return nil, err
	}

	return &types.SetWalletLimitsResponse{
		Success: true,
	}, nil
}

// CheckWalletLimit validates a transaction
func (s *grpcServer) CheckWalletLimit(ctx context.Context, req *types.CheckWalletLimitRequest) (*types.CheckWalletLimitResponse, error) {
	err := s.keeper.CheckWalletLimit(req.Address, req.Amount, req.CurrentBalance)

	return &types.CheckWalletLimitResponse{
		Allowed: err == nil,
		Reason:  getErrorMessage(err),
	}, nil
}

// CreatePostMortem creates a post-mortem analysis
func (s *grpcServer) CreatePostMortem(ctx context.Context, req *types.CreatePostMortemRequest) (*types.CreatePostMortemResponse, error) {
	err := s.keeper.CreatePostMortem(
		req.IncidentId,
		req.CreatedBy,
		req.Summary,
		req.RootCause,
		req.Impact,
		req.Resolution,
		req.LessonsLearned,
		req.ActionItems,
	)
	if err != nil {
		return nil, err
	}

	return &types.CreatePostMortemResponse{
		Success: true,
	}, nil
}

// TriggerBackup initiates a backup operation
func (s *grpcServer) TriggerBackup(ctx context.Context, req *types.TriggerBackupRequest) (*types.TriggerBackupResponse, error) {
	backupID, err := s.keeper.TriggerBackup(req.BackupType, req.TriggeredBy)
	if err != nil {
		return nil, err
	}

	return &types.TriggerBackupResponse{
		BackupId: backupID,
	}, nil
}

// TriggerInsuranceClaim submits an insurance claim
func (s *grpcServer) TriggerInsuranceClaim(ctx context.Context, req *types.TriggerInsuranceClaimRequest) (*types.TriggerInsuranceClaimResponse, error) {
	claimID, err := s.keeper.TriggerInsuranceClaim(
		req.IncidentId,
		req.Amount,
		req.Signers,
	)
	if err != nil {
		return nil, err
	}

	return &types.TriggerInsuranceClaimResponse{
		ClaimId: claimID,
	}, nil
}

func getErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
