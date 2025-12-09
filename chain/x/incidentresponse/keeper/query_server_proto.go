package keeper

import (
	"context"
	"fmt"

	incidentresponsepb "github.com/aequitas/aura/proto/aura/incidentresponse/v1beta1"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/types"
)

var _ incidentresponsepb.QueryServer = protoQueryServer{}

type protoQueryServer struct {
	incidentresponsepb.UnimplementedQueryServer
	keeper *KeeperKV
}

// NewProtoQueryServerImpl creates a new proto-based query server implementation
func NewProtoQueryServerImpl(keeper *KeeperKV) incidentresponsepb.QueryServer {
	return &protoQueryServer{keeper: keeper}
}

// GetIncident retrieves an incident by ID
func (qs protoQueryServer) GetIncident(goCtx context.Context, req *incidentresponsepb.QueryGetIncidentRequest) (*incidentresponsepb.QueryGetIncidentResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.IncidentId == "" {
		return nil, fmt.Errorf("incident_id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	incident, err := qs.keeper.GetIncident(ctx, req.IncidentId)
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}

	// Convert to proto incident
	// Note: Proto fields use stdtime=true
	// ReportedAt is non-nullable (time.Time), UpdatedAt/ResolvedAt are nullable (*time.Time)
	protoIncident := &incidentresponsepb.Incident{
		Id:              incident.ID,
		Title:           incident.Title,
		Description:     incident.Description,
		Severity:        string(incident.Severity),
		Status:          string(incident.Status),
		ReportedBy:      incident.ReportedBy,
		ReportedAt:      incident.ReportedAt,
		UpdatedAt:       &incident.UpdatedAt,
		AffectedSystems: incident.AffectedSystems,
		ResponseTeam:    incident.ResponseTeam,
	}

	if !incident.ResolvedAt.IsZero() {
		protoIncident.ResolvedAt = &incident.ResolvedAt
	}

	return &incidentresponsepb.QueryGetIncidentResponse{
		Incident: protoIncident,
	}, nil
}

// GetChainPauseState retrieves the current chain pause state
func (qs protoQueryServer) GetChainPauseState(goCtx context.Context, req *incidentresponsepb.QueryGetChainPauseStateRequest) (*incidentresponsepb.QueryGetChainPauseStateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	pauseState := qs.keeper.GetChainPauseState(ctx)

	// Convert to proto pause state
	// Note: Proto fields use stdtime=true with nullable (*time.Time) fields
	protoPauseState := &incidentresponsepb.ChainPauseState{
		IsPaused:      pauseState.IsPaused,
		PauseLevel:    string(pauseState.PauseLevel),
		PausedBy:      pauseState.PausedBy,
		Reason:        pauseState.Reason,
		IncidentId:    pauseState.IncidentID,
		PausedModules: pauseState.PausedModules,
	}

	if !pauseState.PausedAt.IsZero() {
		protoPauseState.PausedAt = &pauseState.PausedAt
	}

	if !pauseState.EstimatedResume.IsZero() {
		protoPauseState.EstimatedResume = &pauseState.EstimatedResume
	}

	return &incidentresponsepb.QueryGetChainPauseStateResponse{
		PauseState: protoPauseState,
	}, nil
}

// CheckWalletLimit checks if an amount is within wallet limits
func (qs protoQueryServer) CheckWalletLimit(goCtx context.Context, req *incidentresponsepb.QueryCheckWalletLimitRequest) (*incidentresponsepb.QueryCheckWalletLimitResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	if req.Amount == "" {
		return nil, fmt.Errorf("amount cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if wallet limits exist
	limits, err := qs.keeper.GetWalletLimits(ctx, req.Address)
	if err != nil {
		// No limits set, allow by default
		return &incidentresponsepb.QueryCheckWalletLimitResponse{
			Allowed: true,
			Reason:  "no limits configured",
		}, nil
	}

	// Parse amounts
	amount, ok := math.NewIntFromString(req.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", req.Amount)
	}

	maxBalance, ok := math.NewIntFromString(limits.MaxBalance)
	if !ok {
		return nil, fmt.Errorf("invalid max_balance in limits")
	}

	maxTxSize, ok := math.NewIntFromString(limits.MaxTransactionSize)
	if !ok {
		return nil, fmt.Errorf("invalid max_transaction_size in limits")
	}

	dailyLimit, ok := math.NewIntFromString(limits.DailyLimit)
	if !ok {
		return nil, fmt.Errorf("invalid daily_limit in limits")
	}

	currentBalance := math.ZeroInt()
	if req.CurrentBalance != "" {
		currentBalance, ok = math.NewIntFromString(req.CurrentBalance)
		if !ok {
			return nil, fmt.Errorf("invalid current_balance: %s", req.CurrentBalance)
		}
	}

	todayTransferred, ok := math.NewIntFromString(limits.TodayTransferred)
	if !ok {
		todayTransferred = math.ZeroInt()
	}

	// Check limits
	if amount.GT(maxTxSize) {
		return &incidentresponsepb.QueryCheckWalletLimitResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("amount %s exceeds max transaction size %s", amount, maxTxSize),
		}, nil
	}

	newBalance := currentBalance.Add(amount)
	if newBalance.GT(maxBalance) {
		return &incidentresponsepb.QueryCheckWalletLimitResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("new balance %s would exceed max balance %s", newBalance, maxBalance),
		}, nil
	}

	newDailyTotal := todayTransferred.Add(amount)
	if newDailyTotal.GT(dailyLimit) {
		return &incidentresponsepb.QueryCheckWalletLimitResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("daily total %s would exceed daily limit %s", newDailyTotal, dailyLimit),
		}, nil
	}

	return &incidentresponsepb.QueryCheckWalletLimitResponse{
		Allowed: true,
		Reason:  "within all limits",
	}, nil
}

// GetParams retrieves the module parameters
func (qs protoQueryServer) GetParams(goCtx context.Context, req *incidentresponsepb.QueryGetParamsRequest) (*incidentresponsepb.QueryGetParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := qs.keeper.GetParams(ctx)

	// Convert to proto params
	protoParams := &incidentresponsepb.IncidentResponseParams{
		EmergencyPauseEnabled:  params.EmergencyPauseEnabled,
		PauseAuthorizedKeys:    params.PauseAuthorizedKeys,
		PauseRequiredSigners:   params.PauseRequiredSigners,
		MaxPauseDuration:       types.DurationProto(params.MaxPauseDuration),
		HotWalletLimitsEnabled: params.HotWalletLimitsEnabled,
		GlobalMaxHotWallet:     params.GlobalMaxHotWallet,
		GlobalDailyLimit:       params.GlobalDailyLimit,
	}

	return &incidentresponsepb.QueryGetParamsResponse{
		Params: protoParams,
	}, nil
}
