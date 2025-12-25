package identity

import (
	"context"
	"fmt"

	"github.com/aura-chain/aura/sdk/go/client"
	"github.com/aura-chain/aura/sdk/go/pkg/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the identity module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient identitypb.QueryClient
	msgClient   identitypb.MsgClient
}

// NewClient creates a new identity client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: identitypb.NewQueryClient(grpcConn),
		msgClient:   identitypb.NewMsgClient(grpcConn),
	}
}

// RequestIdentityChangeParams contains parameters for requesting an identity change
type RequestIdentityChangeParams struct {
	Requester    string
	TargetDID    string
	MetadataHash string
	IrID         string
	ProofHash    string
}

// RequestIdentityChange initiates an identity change request
func (c *Client) RequestIdentityChange(ctx context.Context, params *RequestIdentityChangeParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Requester == "" {
		return nil, fmt.Errorf("requester is required")
	}
	if params.TargetDID == "" {
		return nil, fmt.Errorf("target DID is required")
	}

	msg := &identitypb.MsgRequestIdentityChange{
		Requester:    params.Requester,
		TargetDid:    params.TargetDID,
		MetadataHash: params.MetadataHash,
		IrId:         params.IrID,
		ProofHash:    params.ProofHash,
	}

	addr, err := sdk.AccAddressFromBech32(params.Requester)
	if err != nil {
		return nil, fmt.Errorf("invalid requester address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to request identity change: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// CreateRoleParams contains parameters for creating a role
type CreateRoleParams struct {
	Creator     string
	RoleName    string
	Permissions []string
	Description string
}

// CreateRole creates a new role
func (c *Client) CreateRole(ctx context.Context, params *CreateRoleParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Creator == "" {
		return nil, fmt.Errorf("creator is required")
	}
	if params.RoleName == "" {
		return nil, fmt.Errorf("role name is required")
	}

	msg := &identitypb.MsgCreateRole{
		Creator:     params.Creator,
		RoleName:    params.RoleName,
		Permissions: params.Permissions,
		Description: params.Description,
	}

	addr, err := sdk.AccAddressFromBech32(params.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// AssignRoleParams contains parameters for assigning a role
type AssignRoleParams struct {
	Assigner string
	Address  string
	RoleName string
}

// AssignRole assigns a role to an address
func (c *Client) AssignRole(ctx context.Context, params *AssignRoleParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Assigner == "" {
		return nil, fmt.Errorf("assigner is required")
	}
	if params.Address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if params.RoleName == "" {
		return nil, fmt.Errorf("role name is required")
	}

	msg := &identitypb.MsgAssignRole{
		Assigner: params.Assigner,
		Address:  params.Address,
		RoleName: params.RoleName,
	}

	addr, err := sdk.AccAddressFromBech32(params.Assigner)
	if err != nil {
		return nil, fmt.Errorf("invalid assigner address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to assign role: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// CreateMultisigWalletParams contains parameters for creating a multisig wallet
type CreateMultisigWalletParams struct {
	Creator    string
	Signers    []string
	Threshold  uint32
	WalletType identitypb.WalletType
}

// CreateMultisigWallet creates a new multisig wallet
func (c *Client) CreateMultisigWallet(ctx context.Context, params *CreateMultisigWalletParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Creator == "" {
		return nil, fmt.Errorf("creator is required")
	}
	if len(params.Signers) == 0 {
		return nil, fmt.Errorf("at least one signer is required")
	}
	if params.Threshold == 0 {
		return nil, fmt.Errorf("threshold must be greater than 0")
	}

	msg := &identitypb.MsgCreateMultisigWallet{
		Creator:    params.Creator,
		Signers:    params.Signers,
		Threshold:  params.Threshold,
		WalletType: params.WalletType,
	}

	addr, err := sdk.AccAddressFromBech32(params.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create multisig wallet: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// CreateSessionParams contains parameters for creating a session
type CreateSessionParams struct {
	Address           string
	DeviceFingerprint string
	IpAddress         string
	Metadata          map[string]string
}

// CreateSession creates a new authentication session
func (c *Client) CreateSession(ctx context.Context, params *CreateSessionParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Address == "" {
		return nil, fmt.Errorf("address is required")
	}

	msg := &identitypb.MsgCreateSession{
		Address:           params.Address,
		DeviceFingerprint: params.DeviceFingerprint,
		IpAddress:         params.IpAddress,
		Metadata:          params.Metadata,
	}

	addr, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// EndSession ends an authentication session
func (c *Client) EndSession(ctx context.Context, address, sessionID string) (*types.TxResponse, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	msg := &identitypb.MsgEndSession{
		Address:   address,
		SessionId: sessionID,
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to end session: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// GetIdentityRecord retrieves an identity record by DID
func (c *Client) GetIdentityRecord(ctx context.Context, did string) (*identitypb.IdentityRecord, error) {
	if did == "" {
		return nil, fmt.Errorf("DID is required")
	}

	req := &identitypb.QueryIdentityRecordRequest{
		Did: did,
	}

	resp, err := c.queryClient.IdentityRecord(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity record: %w", err)
	}

	return &resp.Record, nil
}

// GetIdentityRecordByAddress retrieves an identity record by address
func (c *Client) GetIdentityRecordByAddress(ctx context.Context, address string) (*identitypb.IdentityRecord, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &identitypb.QueryIdentityRecordByAddressRequest{
		Address: address,
	}

	resp, err := c.queryClient.IdentityRecordByAddress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity record by address: %w", err)
	}

	return &resp.Record, nil
}

// ListIdentityRecords lists all identity records
func (c *Client) ListIdentityRecords(ctx context.Context) ([]identitypb.IdentityRecord, error) {
	req := &identitypb.QueryAllIdentityRecordsRequest{}

	resp, err := c.queryClient.AllIdentityRecords(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list identity records: %w", err)
	}

	return resp.Records, nil
}

// GetChangeRequest retrieves a change request by ID
func (c *Client) GetChangeRequest(ctx context.Context, requestID string) (*identitypb.ChangeRequest, error) {
	if requestID == "" {
		return nil, fmt.Errorf("request ID is required")
	}

	req := &identitypb.QueryChangeRequestRequest{
		RequestId: requestID,
	}

	resp, err := c.queryClient.ChangeRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get change request: %w", err)
	}

	return &resp.Request, nil
}

// GetChangeRequestsByDID retrieves change requests for a DID
func (c *Client) GetChangeRequestsByDID(ctx context.Context, did string) ([]identitypb.ChangeRequest, error) {
	if did == "" {
		return nil, fmt.Errorf("DID is required")
	}

	req := &identitypb.QueryChangeRequestsByDIDRequest{
		Did: did,
	}

	resp, err := c.queryClient.ChangeRequestsByDID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get change requests by DID: %w", err)
	}

	return resp.Requests, nil
}

// GetChangeHistory retrieves change history for a DID
func (c *Client) GetChangeHistory(ctx context.Context, did string) ([]identitypb.ChangeHistory, error) {
	if did == "" {
		return nil, fmt.Errorf("DID is required")
	}

	req := &identitypb.QueryChangeHistoryRequest{
		Did: did,
	}

	resp, err := c.queryClient.ChangeHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get change history: %w", err)
	}

	return resp.Entries, nil
}

// GetRole retrieves a role by name
func (c *Client) GetRole(ctx context.Context, roleName string) (*identitypb.Role, error) {
	if roleName == "" {
		return nil, fmt.Errorf("role name is required")
	}

	req := &identitypb.QueryRoleRequest{
		RoleName: roleName,
	}

	resp, err := c.queryClient.Role(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &resp.Role, nil
}

// ListRoles lists all roles
func (c *Client) ListRoles(ctx context.Context) ([]identitypb.Role, error) {
	req := &identitypb.QueryAllRolesRequest{}

	resp, err := c.queryClient.AllRoles(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return resp.Roles, nil
}

// GetRoleAssignments retrieves role assignments for an address
func (c *Client) GetRoleAssignments(ctx context.Context, address string) ([]identitypb.RoleAssignment, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &identitypb.QueryRoleAssignmentsRequest{
		Address: address,
	}

	resp, err := c.queryClient.RoleAssignments(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get role assignments: %w", err)
	}

	return resp.Assignments, nil
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*identitypb.Params, error) {
	req := &identitypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return &resp.Params, nil
}
