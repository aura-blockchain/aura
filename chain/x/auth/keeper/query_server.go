// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
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

// ListRoles lists all roles with pagination
func (qs queryServer) ListRoles(goCtx context.Context, req *authproto.QueryListRolesRequest) (*authproto.QueryListRolesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for roles
	roleStore := prefix.NewStore(store, RolesKeyPrefix)

	var roles []*authproto.Role
	pageRes, err := query.Paginate(roleStore, req.Pagination, func(key, value []byte) error {
		var role authproto.Role
		if err := qs.Keeper.cdc.Unmarshal(value, &role); err != nil {
			return fmt.Errorf("failed to unmarshal role: %w", err)
		}
		roles = append(roles, &role)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListRolesResponse{
		Roles:      roles,
		Pagination: pageRes,
	}, nil
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

// ListMultisigWallets lists all multisig wallets with pagination
func (qs queryServer) ListMultisigWallets(goCtx context.Context, req *authproto.QueryListMultisigWalletsRequest) (*authproto.QueryListMultisigWalletsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for multisig wallets
	walletStore := prefix.NewStore(store, MultisigWalletsKeyPrefix)

	var wallets []*authproto.MultisigWallet
	pageRes, err := query.Paginate(walletStore, req.Pagination, func(key, value []byte) error {
		var wallet authproto.MultisigWallet
		if err := qs.Keeper.cdc.Unmarshal(value, &wallet); err != nil {
			return fmt.Errorf("failed to unmarshal multisig wallet: %w", err)
		}
		wallets = append(wallets, &wallet)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListMultisigWalletsResponse{
		Wallets:    wallets,
		Pagination: pageRes,
	}, nil
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

// ListMultisigProposals lists multisig proposals with optional filtering and pagination
func (qs queryServer) ListMultisigProposals(goCtx context.Context, req *authproto.QueryListMultisigProposalsRequest) (*authproto.QueryListMultisigProposalsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for multisig proposals
	proposalStore := prefix.NewStore(store, MultisigProposalsKeyPrefix)

	var proposals []*authproto.MultisigProposal
	pageRes, err := query.Paginate(proposalStore, req.Pagination, func(key, value []byte) error {
		var proposal authproto.MultisigProposal
		if err := qs.Keeper.cdc.Unmarshal(value, &proposal); err != nil {
			return fmt.Errorf("failed to unmarshal multisig proposal: %w", err)
		}

		// Apply optional filters
		if req.WalletId != "" && proposal.WalletId != req.WalletId {
			return nil // Skip this proposal
		}
		if req.Status != authproto.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED && proposal.Status != req.Status {
			return nil // Skip this proposal
		}

		proposals = append(proposals, &proposal)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListMultisigProposalsResponse{
		Proposals:  proposals,
		Pagination: pageRes,
	}, nil
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

// ListTimeLockedActions lists all time-locked actions with optional filtering and pagination
func (qs queryServer) ListTimeLockedActions(goCtx context.Context, req *authproto.QueryListTimeLockedActionsRequest) (*authproto.QueryListTimeLockedActionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for time-locked actions
	actionStore := prefix.NewStore(store, TimeLockedActionsKeyPrefix)

	var actions []*authproto.TimeLockedAction
	pageRes, err := query.Paginate(actionStore, req.Pagination, func(key, value []byte) error {
		var action authproto.TimeLockedAction
		if err := qs.Keeper.cdc.Unmarshal(value, &action); err != nil {
			return fmt.Errorf("failed to unmarshal time-locked action: %w", err)
		}

		// Apply optional status filter
		if req.Status != authproto.ActionStatus_ACTION_STATUS_UNSPECIFIED && action.Status != req.Status {
			return nil // Skip this action
		}

		actions = append(actions, &action)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListTimeLockedActionsResponse{
		Actions:    actions,
		Pagination: pageRes,
	}, nil
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

// ListEmergencyAdmins lists all emergency admins with pagination
func (qs queryServer) ListEmergencyAdmins(goCtx context.Context, req *authproto.QueryListEmergencyAdminsRequest) (*authproto.QueryListEmergencyAdminsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for emergency admins
	adminStore := prefix.NewStore(store, EmergencyAdminsKeyPrefix)

	var admins []*authproto.EmergencyAdmin
	pageRes, err := query.Paginate(adminStore, req.Pagination, func(key, value []byte) error {
		var admin authproto.EmergencyAdmin
		if err := qs.Keeper.cdc.Unmarshal(value, &admin); err != nil {
			return fmt.Errorf("failed to unmarshal emergency admin: %w", err)
		}
		admins = append(admins, &admin)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListEmergencyAdminsResponse{
		Admins:     admins,
		Pagination: pageRes,
	}, nil
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

// ListSessions lists sessions for a user with pagination
func (qs queryServer) ListSessions(goCtx context.Context, req *authproto.QueryListSessionsRequest) (*authproto.QueryListSessionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.UserAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "user address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for sessions
	sessionStore := prefix.NewStore(store, SessionsKeyPrefix)

	var sessions []*authproto.Session
	pageRes, err := query.Paginate(sessionStore, req.Pagination, func(key, value []byte) error {
		var session authproto.Session
		if err := qs.Keeper.cdc.Unmarshal(value, &session); err != nil {
			return fmt.Errorf("failed to unmarshal session: %w", err)
		}

		// Filter by user address
		if session.UserAddress != req.UserAddress {
			return nil // Skip this session
		}

		sessions = append(sessions, &session)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authproto.QueryListSessionsResponse{
		Sessions:   sessions,
		Pagination: pageRes,
	}, nil
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

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get audit logs from KVStore based on filters
	// The keeper's GetAuditLogs signature is: (ctx sdk.Context, actor, action string, startTime, endTime int64, limit uint64)
	// The proto field is 'actor' not 'address'
	logs := qs.Keeper.GetAuditLogs(ctx, req.Actor, req.Action, req.StartTime, req.EndTime, req.Limit)

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

	return &authproto.QueryGetParamsResponse{Params: &params}, nil
}
