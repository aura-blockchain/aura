package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper *KeeperKV
}

// NewQueryServerImpl creates a new query server implementation
func NewQueryServerImpl(keeper *KeeperKV) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// GetIncident retrieves an incident by ID
func (qs queryServer) GetIncident(goCtx interface{}, req *types.QueryGetIncidentRequest) (*types.QueryGetIncidentResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.IncidentId == "" {
		return nil, fmt.Errorf("incident_id cannot be empty")
	}

	ctx := goCtx.(sdk.Context)

	incident, err := qs.Keeper.GetIncident(ctx, req.IncidentId)
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}

	return &types.QueryGetIncidentResponse{
		Incident: incident,
	}, nil
}

// GetAllIncidents retrieves all incidents
func (qs queryServer) GetAllIncidents(goCtx interface{}, req *types.QueryGetAllIncidentsRequest) (*types.QueryGetAllIncidentsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	incidents := qs.Keeper.GetAllIncidents(ctx)

	// Filter by status if provided
	if req.Status != "" {
		filtered := make([]*types.Incident, 0)
		for _, inc := range incidents {
			if string(inc.Status) == req.Status {
				filtered = append(filtered, inc)
			}
		}
		incidents = filtered
	}

	// Filter by severity if provided
	if req.Severity != "" {
		filtered := make([]*types.Incident, 0)
		for _, inc := range incidents {
			if string(inc.Severity) == req.Severity {
				filtered = append(filtered, inc)
			}
		}
		incidents = filtered
	}

	return &types.QueryGetAllIncidentsResponse{
		Incidents: incidents,
	}, nil
}

// GetChainPauseState retrieves the current chain pause state
func (qs queryServer) GetChainPauseState(goCtx interface{}, req *types.QueryGetChainPauseStateRequest) (*types.QueryGetChainPauseStateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	pauseState := qs.Keeper.GetChainPauseState(ctx)

	return &types.QueryGetChainPauseStateResponse{
		PauseState: pauseState,
	}, nil
}

// GetWalletLimits retrieves wallet limits for an address
func (qs queryServer) GetWalletLimits(goCtx interface{}, req *types.QueryGetWalletLimitsRequest) (*types.QueryGetWalletLimitsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	if req.Address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	ctx := goCtx.(sdk.Context)

	limits, err := qs.Keeper.GetWalletLimits(ctx, req.Address)
	if err != nil {
		return nil, fmt.Errorf("wallet limits not found: %w", err)
	}

	return &types.QueryGetWalletLimitsResponse{
		Limits: limits,
	}, nil
}

// GetParams retrieves the module parameters
func (qs queryServer) GetParams(goCtx interface{}, req *types.QueryGetParamsRequest) (*types.QueryGetParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	params := qs.Keeper.GetParams(ctx)

	return &types.QueryGetParamsResponse{
		Params: &params,
	}, nil
}

// GetColdStorageConfig retrieves cold storage configuration
func (qs queryServer) GetColdStorageConfig(goCtx interface{}, req *types.QueryGetColdStorageConfigRequest) (*types.QueryGetColdStorageConfigResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	config := qs.Keeper.GetColdStorageConfig(ctx)

	return &types.QueryGetColdStorageConfigResponse{
		Config: &config,
	}, nil
}

// GetBackupValidatorConfig retrieves backup validator configuration
func (qs queryServer) GetBackupValidatorConfig(goCtx interface{}, req *types.QueryGetBackupValidatorConfigRequest) (*types.QueryGetBackupValidatorConfigResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	config := qs.Keeper.GetBackupValidatorConfig(ctx)

	return &types.QueryGetBackupValidatorConfigResponse{
		Config: &config,
	}, nil
}

// GetDisasterRecoveryPlan retrieves disaster recovery configuration
func (qs queryServer) GetDisasterRecoveryPlan(goCtx interface{}, req *types.QueryGetDisasterRecoveryPlanRequest) (*types.QueryGetDisasterRecoveryPlanResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	plan := qs.Keeper.GetDisasterRecoveryPlan(ctx)

	return &types.QueryGetDisasterRecoveryPlanResponse{
		Plan: &plan,
	}, nil
}

// GetCommunicationPlan retrieves communication plan configuration
func (qs queryServer) GetCommunicationPlan(goCtx interface{}, req *types.QueryGetCommunicationPlanRequest) (*types.QueryGetCommunicationPlanResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	plan := qs.Keeper.GetCommunicationPlan(ctx)

	return &types.QueryGetCommunicationPlanResponse{
		Plan: &plan,
	}, nil
}

// GetInsuranceIntegration retrieves insurance integration configuration
func (qs queryServer) GetInsuranceIntegration(goCtx interface{}, req *types.QueryGetInsuranceIntegrationRequest) (*types.QueryGetInsuranceIntegrationResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	ctx := goCtx.(sdk.Context)

	integration := qs.Keeper.GetInsuranceIntegration(ctx)

	return &types.QueryGetInsuranceIntegrationResponse{
		Integration: &integration,
	}, nil
}
