package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

var _ identitypb.QueryServer = (*queryServer)(nil)

type queryServer struct {
	identitypb.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl creates a new query server implementation
func NewQueryServerImpl(keeper *Keeper) identitypb.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Params queries the module parameters
func (qs queryServer) Params(goCtx context.Context, req *identitypb.QueryParamsRequest) (*identitypb.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params, err := qs.Keeper.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryParamsResponse{Params: params}, nil
}

// IdentityRecord queries an identity record by DID
func (qs queryServer) IdentityRecord(goCtx context.Context, req *identitypb.QueryIdentityRecordRequest) (*identitypb.QueryIdentityRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Did == "" {
		return nil, status.Error(codes.InvalidArgument, "DID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	record, err := qs.Keeper.GetIdentityRecord(ctx, req.Did)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QueryIdentityRecordResponse{Record: record}, nil
}

// IdentityRecordByAddress queries an identity record by address
func (qs queryServer) IdentityRecordByAddress(goCtx context.Context, req *identitypb.QueryIdentityRecordByAddressRequest) (*identitypb.QueryIdentityRecordByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Find identity record by iterating through all records
	allRecords, err := qs.Keeper.GetAllIdentityRecords(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, record := range allRecords {
		if record.Address == req.Address {
			return &identitypb.QueryIdentityRecordByAddressResponse{Record: record}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "identity record not found for address")
}

// AllIdentityRecords queries all identity records
func (qs queryServer) AllIdentityRecords(goCtx context.Context, req *identitypb.QueryAllIdentityRecordsRequest) (*identitypb.QueryAllIdentityRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	records, err := qs.Keeper.GetAllIdentityRecords(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryAllIdentityRecordsResponse{Records: records}, nil
}

// ChangeRequest queries a change request by ID
func (qs queryServer) ChangeRequest(goCtx context.Context, req *identitypb.QueryChangeRequestRequest) (*identitypb.QueryChangeRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.RequestId == "" {
		return nil, status.Error(codes.InvalidArgument, "request ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	request, err := qs.Keeper.GetChangeRequest(ctx, req.RequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QueryChangeRequestResponse{Request: request}, nil
}

// ChangeRequestsByDID queries change requests for a DID
func (qs queryServer) ChangeRequestsByDID(goCtx context.Context, req *identitypb.QueryChangeRequestsByDIDRequest) (*identitypb.QueryChangeRequestsByDIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Did == "" {
		return nil, status.Error(codes.InvalidArgument, "DID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	allRequests, err := qs.Keeper.GetAllChangeRequests(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Filter by DID
	var filtered []*types.ChangeRequest
	for _, request := range allRequests {
		if request.Did == req.Did {
			filtered = append(filtered, request)
		}
	}

	return &identitypb.QueryChangeRequestsByDIDResponse{Requests: filtered}, nil
}

// ChangeHistory queries change history for a DID
func (qs queryServer) ChangeHistory(goCtx context.Context, req *identitypb.QueryChangeHistoryRequest) (*identitypb.QueryChangeHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Did == "" {
		return nil, status.Error(codes.InvalidArgument, "DID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	entries, err := qs.Keeper.GetChangeHistory(ctx, req.Did)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryChangeHistoryResponse{Entries: entries}, nil
}

// Role queries a role by name
func (qs queryServer) Role(goCtx context.Context, req *identitypb.QueryRoleRequest) (*identitypb.QueryRoleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.RoleName == "" {
		return nil, status.Error(codes.InvalidArgument, "role name cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	role, err := qs.Keeper.GetRole(ctx, req.RoleName)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QueryRoleResponse{Role: role}, nil
}

// AllRoles queries all roles
func (qs queryServer) AllRoles(goCtx context.Context, req *identitypb.QueryAllRolesRequest) (*identitypb.QueryAllRolesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	roles, err := qs.Keeper.GetAllRoles(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryAllRolesResponse{Roles: roles}, nil
}

// RoleAssignments queries role assignments for an address
func (qs queryServer) RoleAssignments(goCtx context.Context, req *identitypb.QueryRoleAssignmentsRequest) (*identitypb.QueryRoleAssignmentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	assignments, err := qs.Keeper.GetRoleAssignments(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryRoleAssignmentsResponse{Assignments: assignments}, nil
}

// HasPermission checks if an address has a specific permission
func (qs queryServer) HasPermission(goCtx context.Context, req *identitypb.QueryHasPermissionRequest) (*identitypb.QueryHasPermissionResponse, error) {
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
	hasPermission := qs.Keeper.HasPermission(ctx, req.Address, req.Permission)

	// Get roles that grant this permission
	var grantingRoles []string
	assignments, err := qs.Keeper.GetRoleAssignments(ctx, req.Address)
	if err == nil {
		for _, assignment := range assignments {
			role, err := qs.Keeper.GetRole(ctx, assignment.RoleName)
			if err != nil {
				continue
			}
			for _, perm := range role.Permissions {
				if perm == req.Permission || perm == types.PermissionAdmin {
					grantingRoles = append(grantingRoles, role.Name)
					break
				}
			}
		}
	}

	return &identitypb.QueryHasPermissionResponse{
		HasPermission: hasPermission,
		Roles:         grantingRoles,
	}, nil
}

// MultisigWallet queries a multisig wallet by ID
func (qs queryServer) MultisigWallet(goCtx context.Context, req *identitypb.QueryMultisigWalletRequest) (*identitypb.QueryMultisigWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	wallet, err := qs.Keeper.GetMultisigWallet(ctx, req.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QueryMultisigWalletResponse{Wallet: &wallet}, nil
}

// AllMultisigWallets queries all multisig wallets
func (qs queryServer) AllMultisigWallets(goCtx context.Context, req *identitypb.QueryAllMultisigWalletsRequest) (*identitypb.QueryAllMultisigWalletsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	wallets, err := qs.Keeper.GetAllMultisigWallets(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryAllMultisigWalletsResponse{Wallets: wallets}, nil
}

// MultisigProposal queries a multisig proposal by ID
func (qs queryServer) MultisigProposal(goCtx context.Context, req *identitypb.QueryMultisigProposalRequest) (*identitypb.QueryMultisigProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ProposalId == "" {
		return nil, status.Error(codes.InvalidArgument, "proposal ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, err := qs.Keeper.GetMultisigProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QueryMultisigProposalResponse{Proposal: &proposal}, nil
}

// MultisigProposalsByWallet queries proposals for a wallet
func (qs queryServer) MultisigProposalsByWallet(goCtx context.Context, req *identitypb.QueryMultisigProposalsByWalletRequest) (*identitypb.QueryMultisigProposalsByWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.WalletId == "" {
		return nil, status.Error(codes.InvalidArgument, "wallet ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	allProposals, err := qs.Keeper.GetAllMultisigProposals(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Filter by wallet ID
	var filtered []*types.MultisigProposal
	for _, proposal := range allProposals {
		if proposal.WalletId == req.WalletId {
			filtered = append(filtered, proposal)
		}
	}

	return &identitypb.QueryMultisigProposalsByWalletResponse{Proposals: filtered}, nil
}

// TimeLockedAction queries a time-locked action by ID
func (qs queryServer) TimeLockedAction(goCtx context.Context, req *identitypb.QueryTimeLockedActionRequest) (*identitypb.QueryTimeLockedActionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ActionId == "" {
		return nil, status.Error(codes.InvalidArgument, "action ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	action, err := qs.Keeper.GetTimeLockedAction(ctx, req.ActionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QueryTimeLockedActionResponse{Action: action}, nil
}

// AllTimeLockedActions queries all time-locked actions
func (qs queryServer) AllTimeLockedActions(goCtx context.Context, req *identitypb.QueryAllTimeLockedActionsRequest) (*identitypb.QueryAllTimeLockedActionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	actions, err := qs.Keeper.GetAllTimeLockedActions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryAllTimeLockedActionsResponse{Actions: actions}, nil
}

// EmergencyAdmin queries an emergency admin by address
func (qs queryServer) EmergencyAdmin(goCtx context.Context, req *identitypb.QueryEmergencyAdminRequest) (*identitypb.QueryEmergencyAdminResponse, error) {
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

	return &identitypb.QueryEmergencyAdminResponse{Admin: admin}, nil
}

// AllEmergencyAdmins queries all emergency admins
func (qs queryServer) AllEmergencyAdmins(goCtx context.Context, req *identitypb.QueryAllEmergencyAdminsRequest) (*identitypb.QueryAllEmergencyAdminsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	admins, err := qs.Keeper.GetAllEmergencyAdmins(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryAllEmergencyAdminsResponse{Admins: admins}, nil
}

// ValidatorRotation queries a validator rotation by address
func (qs queryServer) ValidatorRotation(goCtx context.Context, req *identitypb.QueryValidatorRotationRequest) (*identitypb.QueryValidatorRotationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ValidatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	rotation, err := qs.Keeper.GetValidatorRotation(ctx, req.ValidatorAddress)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QueryValidatorRotationResponse{Rotation: rotation}, nil
}

// Session queries a session by ID
func (qs queryServer) Session(goCtx context.Context, req *identitypb.QuerySessionRequest) (*identitypb.QuerySessionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	session, err := qs.Keeper.GetSession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &identitypb.QuerySessionResponse{Session: session}, nil
}

// SessionsByAddress queries sessions for an address
func (qs queryServer) SessionsByAddress(goCtx context.Context, req *identitypb.QuerySessionsByAddressRequest) (*identitypb.QuerySessionsByAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	sessionIDs, err := qs.Keeper.GetUserSessions(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get full session objects
	var sessions []*types.Session
	for _, sessionID := range sessionIDs {
		session, err := qs.Keeper.GetSession(ctx, sessionID)
		if err != nil {
			continue
		}
		sessions = append(sessions, &session)
	}

	return &identitypb.QuerySessionsByAddressResponse{Sessions: sessions}, nil
}

// RateLimit queries rate limit config for an address
func (qs queryServer) RateLimit(goCtx context.Context, req *identitypb.QueryRateLimitRequest) (*identitypb.QueryRateLimitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	config, err := qs.Keeper.GetRateLimitConfig(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryRateLimitResponse{Config: config}, nil
}

// AuditLogs queries audit logs
func (qs queryServer) AuditLogs(goCtx context.Context, req *identitypb.QueryAuditLogsRequest) (*identitypb.QueryAuditLogsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	logs, err := qs.Keeper.GetAllAuditLogs(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &identitypb.QueryAuditLogsResponse{Logs: logs}, nil
}

// AuditLogsByActor queries audit logs by actor
func (qs queryServer) AuditLogsByActor(goCtx context.Context, req *identitypb.QueryAuditLogsByActorRequest) (*identitypb.QueryAuditLogsByActorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Actor == "" {
		return nil, status.Error(codes.InvalidArgument, "actor cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	allLogs, err := qs.Keeper.GetAllAuditLogs(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Filter by actor
	var filtered []*types.AuditLog
	for _, log := range allLogs {
		if log.Actor == req.Actor {
			filtered = append(filtered, log)
		}
	}

	return &identitypb.QueryAuditLogsByActorResponse{Logs: filtered}, nil
}
