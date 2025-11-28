package keeper

import (
	"context"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ authproto.QueryServer = queryServer{}

type queryServer struct {
	authproto.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl creates a new query server implementation
func NewQueryServerImpl(keeper *Keeper) authproto.QueryServer {
	return &queryServer{Keeper: keeper}
}

// GetRole queries a role by name
func (qs queryServer) GetRole(goCtx context.Context, req *authproto.QueryGetRoleRequest) (*authproto.QueryGetRoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "role name cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	role, err := qs.Keeper.GetRoleFromStore(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &authproto.QueryGetRoleResponse{Role: role}, nil
}

// ListRoles lists all roles
func (qs queryServer) ListRoles(goCtx context.Context, req *authproto.QueryListRolesRequest) (*authproto.QueryListRolesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	roles, err := qs.Keeper.GetAllRoles(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListRolesResponse{Roles: roles}, nil
}

// GetRoleAssignments queries role assignments for an address
func (qs queryServer) GetRoleAssignments(goCtx context.Context, req *authproto.QueryGetRoleAssignmentsRequest) (*authproto.QueryGetRoleAssignmentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	assignments, err := qs.Keeper.GetRoleAssignmentsForAddress(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryGetRoleAssignmentsResponse{Assignments: assignments}, nil
}

// HasPermission checks if an address has a specific permission
func (qs queryServer) HasPermission(goCtx context.Context, req *authproto.QueryHasPermissionRequest) (*authproto.QueryHasPermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	if req.Permission == "" {
		return nil, status.Error(codes.InvalidArgument, "permission cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all role assignments for the address
	assignments, err := qs.Keeper.GetRoleAssignmentsForAddress(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check each assigned role for the permission
	hasPermission := false
	matchingRoles := []string{}

	for _, assignment := range assignments {
		role, err := qs.Keeper.GetRoleFromStore(ctx, assignment.RoleName)
		if err != nil {
			continue
		}

		// Check if role has the permission
		for _, perm := range role.Permissions {
			if perm == req.Permission {
				hasPermission = true
				matchingRoles = append(matchingRoles, role.Name)
				break
			}
		}
	}

	return &authproto.QueryHasPermissionResponse{
		HasPermission: hasPermission,
		MatchingRoles: matchingRoles,
	}, nil
}

// GetMultisigWallet queries a multisig wallet
func (qs queryServer) GetMultisigWallet(goCtx context.Context, req *authproto.QueryGetMultisigWalletRequest) (*authproto.QueryGetMultisigWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	wallet, err := qs.Keeper.GetMultisigWallet(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &authproto.QueryGetMultisigWalletResponse{Wallet: wallet}, nil
}

// ListMultisigWallets lists all multisig wallets
func (qs queryServer) ListMultisigWallets(goCtx context.Context, req *authproto.QueryListMultisigWalletsRequest) (*authproto.QueryListMultisigWalletsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	wallets, err := qs.Keeper.GetAllMultisigWallets(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListMultisigWalletsResponse{Wallets: wallets}, nil
}

// GetMultisigProposal queries a multisig proposal
func (qs queryServer) GetMultisigProposal(goCtx context.Context, req *authproto.QueryGetMultisigProposalRequest) (*authproto.QueryGetMultisigProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, err := qs.Keeper.GetMultisigProposal(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &authproto.QueryGetMultisigProposalResponse{Proposal: proposal}, nil
}

// ListMultisigProposals lists multisig proposals for a wallet
func (qs queryServer) ListMultisigProposals(goCtx context.Context, req *authproto.QueryListMultisigProposalsRequest) (*authproto.QueryListMultisigProposalsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all proposals
	allProposals, err := qs.Keeper.GetAllMultisigProposals(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Filter by wallet_id if specified
	proposals := allProposals
	if req.WalletId != "" {
		filteredProposals := []*authproto.MultisigProposal{}
		for _, p := range allProposals {
			if p.WalletId == req.WalletId {
				filteredProposals = append(filteredProposals, p)
			}
		}
		proposals = filteredProposals
	}

	return &authproto.QueryListMultisigProposalsResponse{Proposals: proposals}, nil
}

// GetTimeLockedAction queries a time-locked action
func (qs queryServer) GetTimeLockedAction(goCtx context.Context, req *authproto.QueryGetTimeLockedActionRequest) (*authproto.QueryGetTimeLockedActionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "action id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	action, err := qs.Keeper.GetTimeLockedAction(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &authproto.QueryGetTimeLockedActionResponse{Action: action}, nil
}

// ListTimeLockedActions lists all time-locked actions
func (qs queryServer) ListTimeLockedActions(goCtx context.Context, req *authproto.QueryListTimeLockedActionsRequest) (*authproto.QueryListTimeLockedActionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	actions, err := qs.Keeper.GetAllTimeLockedActions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListTimeLockedActionsResponse{Actions: actions}, nil
}

// GetEmergencyAdmin queries emergency admin status
func (qs queryServer) GetEmergencyAdmin(goCtx context.Context, req *authproto.QueryGetEmergencyAdminRequest) (*authproto.QueryGetEmergencyAdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	admin, err := qs.Keeper.GetEmergencyAdmin(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &authproto.QueryGetEmergencyAdminResponse{Admin: admin}, nil
}

// ListEmergencyAdmins lists all emergency admins
func (qs queryServer) ListEmergencyAdmins(goCtx context.Context, req *authproto.QueryListEmergencyAdminsRequest) (*authproto.QueryListEmergencyAdminsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	admins, err := qs.Keeper.GetAllEmergencyAdmins(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListEmergencyAdminsResponse{Admins: admins}, nil
}

// GetValidatorKeyRotation queries validator key rotation status
func (qs queryServer) GetValidatorKeyRotation(goCtx context.Context, req *authproto.QueryGetValidatorKeyRotationRequest) (*authproto.QueryGetValidatorKeyRotationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	rotation, err := qs.Keeper.GetValidatorKeyRotation(ctx, req.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &authproto.QueryGetValidatorKeyRotationResponse{Rotation: rotation}, nil
}

// GetSession queries a session
func (qs queryServer) GetSession(goCtx context.Context, req *authproto.QueryGetSessionRequest) (*authproto.QueryGetSessionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	session, err := qs.Keeper.GetSession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &authproto.QueryGetSessionResponse{Session: session}, nil
}

// ListSessions lists sessions for a user
func (qs queryServer) ListSessions(goCtx context.Context, req *authproto.QueryListSessionsRequest) (*authproto.QueryListSessionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.UserAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "user address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all session IDs for the user
	sessionIDs, err := qs.Keeper.GetUserSessions(ctx, req.UserAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Retrieve full session objects
	sessions := make([]*authproto.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		session, err := qs.Keeper.GetSession(ctx, sessionID)
		if err != nil {
			// Skip sessions that can't be found (may have been deleted)
			continue
		}
		sessions = append(sessions, session)
	}

	return &authproto.QueryListSessionsResponse{Sessions: sessions}, nil
}

// GetRateLimitStatus queries rate limit status for a user
func (qs queryServer) GetRateLimitStatus(goCtx context.Context, req *authproto.QueryGetRateLimitStatusRequest) (*authproto.QueryGetRateLimitStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.UserAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "user address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	config, err := qs.Keeper.GetRateLimitConfig(ctx, req.UserAddress)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Determine if user is currently rate limited
	isLimited := false
	if config.RequestsPerMinute > 0 && config.CurrentMinuteCount >= config.RequestsPerMinute {
		isLimited = true
	}
	if config.RequestsPerHour > 0 && config.CurrentHourCount >= config.RequestsPerHour {
		isLimited = true
	}
	if config.RequestsPerDay > 0 && config.CurrentDayCount >= config.RequestsPerDay {
		isLimited = true
	}

	return &authproto.QueryGetRateLimitStatusResponse{
		Config:    config,
		IsLimited: isLimited,
	}, nil
}

// GetAuditLogs queries audit logs
func (qs queryServer) GetAuditLogs(goCtx context.Context, req *authproto.QueryGetAuditLogsRequest) (*authproto.QueryGetAuditLogsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Note: sdk.Context is not needed for GetAuditLogs since it uses in-memory storage
	// Get audit logs based on filters
	// The keeper's GetAuditLogs signature is: (actor, action string, startTime, endTime int64, limit uint64)
	// The proto field is 'actor' not 'address'
	logs := qs.Keeper.GetAuditLogs(req.Actor, req.Action, req.StartTime, req.EndTime, req.Limit)

	return &authproto.QueryGetAuditLogsResponse{Logs: logs}, nil
}

// GetParams queries the auth module parameters
func (qs queryServer) GetParams(goCtx context.Context, req *authproto.QueryGetParamsRequest) (*authproto.QueryGetParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := qs.Keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryGetParamsResponse{Params: params}, nil
}
