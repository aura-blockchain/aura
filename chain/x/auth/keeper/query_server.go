package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ authproto.QueryServer = queryServer{}

// queryServer implements the QueryServer interface
type queryServer struct {
	authproto.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) authproto.QueryServer {
	return &queryServer{Keeper: keeper}
}

// GetRole queries a role by name
func (qs queryServer) GetRole(goCtx context.Context, req *authproto.QueryGetRoleRequest) (*authproto.QueryGetRoleResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	role, err := qs.Keeper.GetRole(req.Name)
	if err != nil {
		return nil, err
	}

	return &authproto.QueryGetRoleResponse{Role: role}, nil
}

// ListRoles lists all roles
func (qs queryServer) ListRoles(goCtx context.Context, req *authproto.QueryListRolesRequest) (*authproto.QueryListRolesResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	roles := qs.Keeper.ListRoles()

	return &authproto.QueryListRolesResponse{Roles: roles}, nil
}

// GetRoleAssignments queries role assignments for an address
func (qs queryServer) GetRoleAssignments(goCtx context.Context, req *authproto.QueryGetRoleAssignmentsRequest) (*authproto.QueryGetRoleAssignmentsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	assignments := qs.Keeper.GetRoleAssignments(req.Address)

	return &authproto.QueryGetRoleAssignmentsResponse{Assignments: assignments}, nil
}

// HasPermission checks if an address has a specific permission
func (qs queryServer) HasPermission(goCtx context.Context, req *authproto.QueryHasPermissionRequest) (*authproto.QueryHasPermissionResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	hasPermission := qs.Keeper.HasPermission(req.Address, req.Permission)
	matchingRoles := qs.Keeper.GetMatchingRoles(req.Address, req.Permission)

	return &authproto.QueryHasPermissionResponse{
		HasPermission: hasPermission,
		MatchingRoles: matchingRoles,
	}, nil
}

// GetMultisigWallet queries a multisig wallet
func (qs queryServer) GetMultisigWallet(goCtx context.Context, req *authproto.QueryGetMultisigWalletRequest) (*authproto.QueryGetMultisigWalletResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	wallet, err := qs.Keeper.GetMultisigWallet(req.Id)
	if err != nil {
		return nil, err
	}

	return &authproto.QueryGetMultisigWalletResponse{Wallet: wallet}, nil
}

// ListMultisigWallets lists all multisig wallets
func (qs queryServer) ListMultisigWallets(goCtx context.Context, req *authproto.QueryListMultisigWalletsRequest) (*authproto.QueryListMultisigWalletsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	wallets := qs.Keeper.ListMultisigWallets()

	return &authproto.QueryListMultisigWalletsResponse{Wallets: wallets}, nil
}

// GetMultisigProposal queries a multisig proposal
func (qs queryServer) GetMultisigProposal(goCtx context.Context, req *authproto.QueryGetMultisigProposalRequest) (*authproto.QueryGetMultisigProposalResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	proposal, err := qs.Keeper.GetMultisigProposal(req.Id)
	if err != nil {
		return nil, err
	}

	return &authproto.QueryGetMultisigProposalResponse{Proposal: proposal}, nil
}

// ListMultisigProposals lists multisig proposals for a wallet
func (qs queryServer) ListMultisigProposals(goCtx context.Context, req *authproto.QueryListMultisigProposalsRequest) (*authproto.QueryListMultisigProposalsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	proposals := qs.Keeper.ListMultisigProposals(req.WalletId, req.Status)

	return &authproto.QueryListMultisigProposalsResponse{Proposals: proposals}, nil
}

// GetTimeLockedAction queries a time-locked action
func (qs queryServer) GetTimeLockedAction(goCtx context.Context, req *authproto.QueryGetTimeLockedActionRequest) (*authproto.QueryGetTimeLockedActionResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	action, err := qs.Keeper.GetTimeLockedAction(req.Id)
	if err != nil {
		return nil, err
	}

	return &authproto.QueryGetTimeLockedActionResponse{Action: action}, nil
}

// ListTimeLockedActions lists all time-locked actions
func (qs queryServer) ListTimeLockedActions(goCtx context.Context, req *authproto.QueryListTimeLockedActionsRequest) (*authproto.QueryListTimeLockedActionsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	actions := qs.Keeper.ListTimeLockedActions(req.Status)

	return &authproto.QueryListTimeLockedActionsResponse{Actions: actions}, nil
}

// GetEmergencyAdmin queries emergency admin status
func (qs queryServer) GetEmergencyAdmin(goCtx context.Context, req *authproto.QueryGetEmergencyAdminRequest) (*authproto.QueryGetEmergencyAdminResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	admin, err := qs.Keeper.GetEmergencyAdmin(req.Address)
	if err != nil {
		return nil, err
	}

	return &authproto.QueryGetEmergencyAdminResponse{Admin: admin}, nil
}

// ListEmergencyAdmins lists all emergency admins
func (qs queryServer) ListEmergencyAdmins(goCtx context.Context, req *authproto.QueryListEmergencyAdminsRequest) (*authproto.QueryListEmergencyAdminsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	admins := qs.Keeper.ListEmergencyAdmins()

	return &authproto.QueryListEmergencyAdminsResponse{Admins: admins}, nil
}

// GetValidatorKeyRotation queries validator key rotation status
func (qs queryServer) GetValidatorKeyRotation(goCtx context.Context, req *authproto.QueryGetValidatorKeyRotationRequest) (*authproto.QueryGetValidatorKeyRotationResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	rotation, err := qs.Keeper.GetValidatorKeyRotation(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	return &authproto.QueryGetValidatorKeyRotationResponse{Rotation: rotation}, nil
}

// GetSession queries a session
func (qs queryServer) GetSession(goCtx context.Context, req *authproto.QueryGetSessionRequest) (*authproto.QueryGetSessionResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	session, err := qs.Keeper.GetSession(req.SessionId)
	if err != nil {
		return nil, err
	}

	return &authproto.QueryGetSessionResponse{Session: session}, nil
}

// ListSessions lists sessions for a user
func (qs queryServer) ListSessions(goCtx context.Context, req *authproto.QueryListSessionsRequest) (*authproto.QueryListSessionsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	sessions := qs.Keeper.ListSessions(req.UserAddress)

	return &authproto.QueryListSessionsResponse{Sessions: sessions}, nil
}

// GetRateLimitStatus queries rate limit status for a user
func (qs queryServer) GetRateLimitStatus(goCtx context.Context, req *authproto.QueryGetRateLimitStatusRequest) (*authproto.QueryGetRateLimitStatusResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	config, isLimited := qs.Keeper.GetRateLimitStatus(req.UserAddress)

	return &authproto.QueryGetRateLimitStatusResponse{
		Config:    config,
		IsLimited: isLimited,
	}, nil
}

// GetAuditLogs queries audit logs
func (qs queryServer) GetAuditLogs(goCtx context.Context, req *authproto.QueryGetAuditLogsRequest) (*authproto.QueryGetAuditLogsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	logs := qs.Keeper.GetAuditLogs(req.Actor, req.Action, req.StartTime, req.EndTime, req.Limit)

	return &authproto.QueryGetAuditLogsResponse{Logs: logs}, nil
}

// GetParams queries the auth module parameters
func (qs queryServer) GetParams(goCtx context.Context, req *authproto.QueryGetParamsRequest) (*authproto.QueryGetParamsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx

	params := qs.Keeper.GetParams()

	return &authproto.QueryGetParamsResponse{Params: params}, nil
}
