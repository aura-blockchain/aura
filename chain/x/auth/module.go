package auth

import (
	"context"

	"github.com/aequitas/aura/chain/x/auth/keeper"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	"google.golang.org/grpc"
)

// AppModule represents the auth module
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new auth module
func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the auth module
func (am AppModule) RegisterGRPCGatewayRoutes(clientCtx interface{}, mux interface{}) {
	// Register HTTP handlers if needed
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg interface{}) error {
	// In a real implementation, this would register with the Cosmos SDK module manager
	return nil
}

// RegisterGRPCServer registers the gRPC services
func (am AppModule) RegisterGRPCServer(server *grpc.Server) {
	authproto.RegisterMsgServer(server, NewMsgServer(am.keeper))
	authproto.RegisterQueryServer(server, NewQueryServer(am.keeper))
}

// MsgServer implements the auth Msg service
type MsgServer struct {
	keeper *keeper.Keeper
	authproto.UnimplementedMsgServer
}

// NewMsgServer creates a new MsgServer
func NewMsgServer(keeper *keeper.Keeper) authproto.MsgServer {
	return &MsgServer{keeper: keeper}
}

// CreateRole creates a new role
func (s *MsgServer) CreateRole(ctx context.Context, req *authproto.MsgCreateRole) (*authproto.MsgCreateRoleResponse, error) {
	role, err := s.keeper.CreateRole(ctx, req.Creator, req.Name, req.Permissions, req.Description)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateRoleResponse{Role: role}, nil
}

// AssignRole assigns a role to an address
func (s *MsgServer) AssignRole(ctx context.Context, req *authproto.MsgAssignRole) (*authproto.MsgAssignRoleResponse, error) {
	assignment, err := s.keeper.AssignRole(ctx, req.Assigner, req.Address, req.RoleName, req.ExpiresInSeconds)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgAssignRoleResponse{Assignment: assignment}, nil
}

// RevokeRole revokes a role from an address
func (s *MsgServer) RevokeRole(ctx context.Context, req *authproto.MsgRevokeRole) (*authproto.MsgRevokeRoleResponse, error) {
	err := s.keeper.RevokeRole(ctx, req.Revoker, req.Address, req.RoleName)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgRevokeRoleResponse{Success: true}, nil
}

// CreateMultisigWallet creates a new multisig wallet
func (s *MsgServer) CreateMultisigWallet(ctx context.Context, req *authproto.MsgCreateMultisigWallet) (*authproto.MsgCreateMultisigWalletResponse, error) {
	wallet, err := s.keeper.CreateMultisigWallet(ctx, req.Creator, req.Signers, req.Threshold, req.WalletType)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateMultisigWalletResponse{Wallet: wallet}, nil
}

// CreateMultisigProposal creates a new multisig proposal
func (s *MsgServer) CreateMultisigProposal(ctx context.Context, req *authproto.MsgCreateMultisigProposal) (*authproto.MsgCreateMultisigProposalResponse, error) {
	proposal, err := s.keeper.CreateMultisigProposal(ctx, req.Proposer, req.WalletId, req.Title, req.Description, req.Payload, req.ExpiresInSeconds)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateMultisigProposalResponse{Proposal: proposal}, nil
}

// SignMultisigProposal signs a multisig proposal
func (s *MsgServer) SignMultisigProposal(ctx context.Context, req *authproto.MsgSignMultisigProposal) (*authproto.MsgSignMultisigProposalResponse, error) {
	proposal, err := s.keeper.SignMultisigProposal(ctx, req.Signer, req.ProposalId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgSignMultisigProposalResponse{Proposal: proposal}, nil
}

// ExecuteMultisigProposal executes an approved multisig proposal
func (s *MsgServer) ExecuteMultisigProposal(ctx context.Context, req *authproto.MsgExecuteMultisigProposal) (*authproto.MsgExecuteMultisigProposalResponse, error) {
	err := s.keeper.ExecuteMultisigProposal(ctx, req.Executor, req.ProposalId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgExecuteMultisigProposalResponse{Success: true}, nil
}

// ProposeTimeLockedAction proposes a time-locked admin action
func (s *MsgServer) ProposeTimeLockedAction(ctx context.Context, req *authproto.MsgProposeTimeLockedAction) (*authproto.MsgProposeTimeLockedActionResponse, error) {
	action, err := s.keeper.ProposeTimeLockedAction(ctx, req.Proposer, req.ActionType, req.Payload, req.DelaySeconds)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgProposeTimeLockedActionResponse{Action: action}, nil
}

// ExecuteTimeLockedAction executes a ready time-locked action
func (s *MsgServer) ExecuteTimeLockedAction(ctx context.Context, req *authproto.MsgExecuteTimeLockedAction) (*authproto.MsgExecuteTimeLockedActionResponse, error) {
	err := s.keeper.ExecuteTimeLockedAction(ctx, req.Executor, req.ActionId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgExecuteTimeLockedActionResponse{Success: true}, nil
}

// CancelTimeLockedAction cancels a pending time-locked action
func (s *MsgServer) CancelTimeLockedAction(ctx context.Context, req *authproto.MsgCancelTimeLockedAction) (*authproto.MsgCancelTimeLockedActionResponse, error) {
	err := s.keeper.CancelTimeLockedAction(ctx, req.Canceller, req.ActionId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCancelTimeLockedActionResponse{Success: true}, nil
}

// ActivateEmergencyAdmin activates an emergency admin
func (s *MsgServer) ActivateEmergencyAdmin(ctx context.Context, req *authproto.MsgActivateEmergencyAdmin) (*authproto.MsgActivateEmergencyAdminResponse, error) {
	admin, err := s.keeper.ActivateEmergencyAdmin(ctx, req.Activator, req.AdminAddress, req.Privileges, req.ExpiresInSeconds)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgActivateEmergencyAdminResponse{Admin: admin}, nil
}

// DeactivateEmergencyAdmin deactivates an emergency admin
func (s *MsgServer) DeactivateEmergencyAdmin(ctx context.Context, req *authproto.MsgDeactivateEmergencyAdmin) (*authproto.MsgDeactivateEmergencyAdminResponse, error) {
	err := s.keeper.DeactivateEmergencyAdmin(ctx, req.Deactivator, req.AdminAddress)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgDeactivateEmergencyAdminResponse{Success: true}, nil
}

// InitiateValidatorKeyRotation initiates validator key rotation
func (s *MsgServer) InitiateValidatorKeyRotation(ctx context.Context, req *authproto.MsgInitiateValidatorKeyRotation) (*authproto.MsgInitiateValidatorKeyRotationResponse, error) {
	rotation, err := s.keeper.InitiateValidatorKeyRotation(ctx, req.Initiator, req.ValidatorAddress, req.NewConsensusPubkey)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgInitiateValidatorKeyRotationResponse{Rotation: rotation}, nil
}

// CompleteValidatorKeyRotation completes validator key rotation
func (s *MsgServer) CompleteValidatorKeyRotation(ctx context.Context, req *authproto.MsgCompleteValidatorKeyRotation) (*authproto.MsgCompleteValidatorKeyRotationResponse, error) {
	err := s.keeper.CompleteValidatorKeyRotation(ctx, req.Completer, req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCompleteValidatorKeyRotationResponse{Success: true}, nil
}

// CreateSession creates a new API session
func (s *MsgServer) CreateSession(ctx context.Context, req *authproto.MsgCreateSession) (*authproto.MsgCreateSessionResponse, error) {
	session, err := s.keeper.CreateSession(ctx, req.UserAddress, req.IpAddress, req.Metadata)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgCreateSessionResponse{Session: session}, nil
}

// RevokeSession revokes an active session
func (s *MsgServer) RevokeSession(ctx context.Context, req *authproto.MsgRevokeSession) (*authproto.MsgRevokeSessionResponse, error) {
	err := s.keeper.RevokeSession(ctx, req.UserAddress, req.SessionId)
	if err != nil {
		return nil, err
	}
	return &authproto.MsgRevokeSessionResponse{Success: true}, nil
}

// QueryServer implements the auth Query service
type QueryServer struct {
	keeper *keeper.Keeper
	authproto.UnimplementedQueryServer
}

// NewQueryServer creates a new QueryServer
func NewQueryServer(keeper *keeper.Keeper) authproto.QueryServer {
	return &QueryServer{keeper: keeper}
}

// GetRole queries a role by name
func (s *QueryServer) GetRole(ctx context.Context, req *authproto.QueryGetRoleRequest) (*authproto.QueryGetRoleResponse, error) {
	role, err := s.keeper.GetRole(req.Name)
	if err != nil {
		return nil, err
	}
	return &authproto.QueryGetRoleResponse{Role: role}, nil
}

// ListRoles lists all roles
func (s *QueryServer) ListRoles(ctx context.Context, req *authproto.QueryListRolesRequest) (*authproto.QueryListRolesResponse, error) {
	roles := s.keeper.ListRoles()
	return &authproto.QueryListRolesResponse{Roles: roles}, nil
}

// GetRoleAssignments queries role assignments for an address
func (s *QueryServer) GetRoleAssignments(ctx context.Context, req *authproto.QueryGetRoleAssignmentsRequest) (*authproto.QueryGetRoleAssignmentsResponse, error) {
	assignments := s.keeper.GetRoleAssignments(req.Address)
	return &authproto.QueryGetRoleAssignmentsResponse{Assignments: assignments}, nil
}

// HasPermission checks if an address has a specific permission
func (s *QueryServer) HasPermission(ctx context.Context, req *authproto.QueryHasPermissionRequest) (*authproto.QueryHasPermissionResponse, error) {
	hasPermission := s.keeper.HasPermission(req.Address, req.Permission)

	// Get matching roles
	matchingRoles := []string{}
	if hasPermission {
		assignments := s.keeper.GetRoleAssignments(req.Address)
		for _, assignment := range assignments {
			role, err := s.keeper.GetRole(assignment.RoleName)
			if err == nil {
				for _, perm := range role.Permissions {
					if perm == req.Permission {
						matchingRoles = append(matchingRoles, role.Name)
						break
					}
				}
			}
		}
	}

	return &authproto.QueryHasPermissionResponse{
		HasPermission: hasPermission,
		MatchingRoles: matchingRoles,
	}, nil
}

// GetMultisigWallet queries a multisig wallet
func (s *QueryServer) GetMultisigWallet(ctx context.Context, req *authproto.QueryGetMultisigWalletRequest) (*authproto.QueryGetMultisigWalletResponse, error) {
	wallet, err := s.keeper.GetMultisigWallet(req.Id)
	if err != nil {
		return nil, err
	}
	return &authproto.QueryGetMultisigWalletResponse{Wallet: wallet}, nil
}

// ListMultisigWallets lists all multisig wallets
func (s *QueryServer) ListMultisigWallets(ctx context.Context, req *authproto.QueryListMultisigWalletsRequest) (*authproto.QueryListMultisigWalletsResponse, error) {
	wallets := s.keeper.ListMultisigWallets()
	return &authproto.QueryListMultisigWalletsResponse{Wallets: wallets}, nil
}

// GetMultisigProposal queries a multisig proposal
func (s *QueryServer) GetMultisigProposal(ctx context.Context, req *authproto.QueryGetMultisigProposalRequest) (*authproto.QueryGetMultisigProposalResponse, error) {
	proposal, err := s.keeper.GetMultisigProposal(req.Id)
	if err != nil {
		return nil, err
	}
	return &authproto.QueryGetMultisigProposalResponse{Proposal: proposal}, nil
}

// ListMultisigProposals lists multisig proposals for a wallet
func (s *QueryServer) ListMultisigProposals(ctx context.Context, req *authproto.QueryListMultisigProposalsRequest) (*authproto.QueryListMultisigProposalsResponse, error) {
	proposals := s.keeper.ListMultisigProposals(req.WalletId, req.Status)
	return &authproto.QueryListMultisigProposalsResponse{Proposals: proposals}, nil
}

// GetTimeLockedAction queries a time-locked action
func (s *QueryServer) GetTimeLockedAction(ctx context.Context, req *authproto.QueryGetTimeLockedActionRequest) (*authproto.QueryGetTimeLockedActionResponse, error) {
	action, err := s.keeper.GetTimeLockedAction(req.Id)
	if err != nil {
		return nil, err
	}
	return &authproto.QueryGetTimeLockedActionResponse{Action: action}, nil
}

// ListTimeLockedActions lists all time-locked actions
func (s *QueryServer) ListTimeLockedActions(ctx context.Context, req *authproto.QueryListTimeLockedActionsRequest) (*authproto.QueryListTimeLockedActionsResponse, error) {
	actions := s.keeper.ListTimeLockedActions(req.Status)
	return &authproto.QueryListTimeLockedActionsResponse{Actions: actions}, nil
}

// GetEmergencyAdmin queries emergency admin status
func (s *QueryServer) GetEmergencyAdmin(ctx context.Context, req *authproto.QueryGetEmergencyAdminRequest) (*authproto.QueryGetEmergencyAdminResponse, error) {
	admin, err := s.keeper.GetEmergencyAdmin(req.Address)
	if err != nil {
		return nil, err
	}
	return &authproto.QueryGetEmergencyAdminResponse{Admin: admin}, nil
}

// ListEmergencyAdmins lists all emergency admins
func (s *QueryServer) ListEmergencyAdmins(ctx context.Context, req *authproto.QueryListEmergencyAdminsRequest) (*authproto.QueryListEmergencyAdminsResponse, error) {
	admins := s.keeper.ListEmergencyAdmins()
	return &authproto.QueryListEmergencyAdminsResponse{Admins: admins}, nil
}

// GetValidatorKeyRotation queries validator key rotation status
func (s *QueryServer) GetValidatorKeyRotation(ctx context.Context, req *authproto.QueryGetValidatorKeyRotationRequest) (*authproto.QueryGetValidatorKeyRotationResponse, error) {
	rotation, err := s.keeper.GetValidatorKeyRotation(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	return &authproto.QueryGetValidatorKeyRotationResponse{Rotation: rotation}, nil
}

// GetSession queries a session
func (s *QueryServer) GetSession(ctx context.Context, req *authproto.QueryGetSessionRequest) (*authproto.QueryGetSessionResponse, error) {
	session, err := s.keeper.GetSession(req.SessionId)
	if err != nil {
		return nil, err
	}
	return &authproto.QueryGetSessionResponse{Session: session}, nil
}

// ListSessions lists sessions for a user
func (s *QueryServer) ListSessions(ctx context.Context, req *authproto.QueryListSessionsRequest) (*authproto.QueryListSessionsResponse, error) {
	sessions := s.keeper.ListSessions(req.UserAddress)
	return &authproto.QueryListSessionsResponse{Sessions: sessions}, nil
}

// GetRateLimitStatus queries rate limit status for a user
func (s *QueryServer) GetRateLimitStatus(ctx context.Context, req *authproto.QueryGetRateLimitStatusRequest) (*authproto.QueryGetRateLimitStatusResponse, error) {
	config, isLimited := s.keeper.GetRateLimitStatus(req.UserAddress)
	return &authproto.QueryGetRateLimitStatusResponse{
		Config:    config,
		IsLimited: isLimited,
	}, nil
}

// GetAuditLogs queries audit logs
func (s *QueryServer) GetAuditLogs(ctx context.Context, req *authproto.QueryGetAuditLogsRequest) (*authproto.QueryGetAuditLogsResponse, error) {
	logs := s.keeper.GetAuditLogs(req.Actor, req.Action, req.StartTime, req.EndTime, req.Limit)
	return &authproto.QueryGetAuditLogsResponse{Logs: logs}, nil
}

// GetParams queries the auth module parameters
func (s *QueryServer) GetParams(ctx context.Context, req *authproto.QueryGetParamsRequest) (*authproto.QueryGetParamsResponse, error) {
	params := s.keeper.GetParams()
	return &authproto.QueryGetParamsResponse{Params: params}, nil
}
