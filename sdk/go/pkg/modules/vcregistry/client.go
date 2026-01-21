package vcregistry

import (
	"context"
	"fmt"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"github.com/aura-chain/aura/sdk/go/pkg/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the vcregistry module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient vcregistrypb.QueryClient
	msgClient   vcregistrypb.MsgClient
}

// NewClient creates a new vcregistry client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: vcregistrypb.NewQueryClient(grpcConn),
		msgClient:   vcregistrypb.NewMsgClient(grpcConn),
	}
}

// MintVCParams contains parameters for minting a verifiable credential
type MintVCParams struct {
	HolderAddress string
	HolderDID     string
	VCType        vcregistrypb.VCType
	VCTypeCustom  string
	Metadata      map[string]string
}

// MintVC mints a new verifiable credential
func (c *Client) MintVC(ctx context.Context, params *MintVCParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.HolderAddress == "" {
		return nil, fmt.Errorf("holder address is required")
	}
	if params.HolderDID == "" {
		return nil, fmt.Errorf("holder DID is required")
	}

	msg := &vcregistrypb.MsgMintVC{
		HolderAddress: params.HolderAddress,
		HolderDid:     params.HolderDID,
		VcType:        params.VCType,
		VcTypeCustom:  params.VCTypeCustom,
		Metadata:      params.Metadata,
	}

	// Sign and broadcast the message
	addr, err := sdk.AccAddressFromBech32(params.HolderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid holder address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to mint VC: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// RevokeVCParams contains parameters for revoking a credential
type RevokeVCParams struct {
	HolderAddress string
	VCID          string
	ReasonText    string
}

// RevokeVC revokes a verifiable credential
func (c *Client) RevokeVC(ctx context.Context, params *RevokeVCParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.HolderAddress == "" {
		return nil, fmt.Errorf("holder address is required")
	}
	if params.VCID == "" {
		return nil, fmt.Errorf("VC ID is required")
	}

	msg := &vcregistrypb.MsgRevokeVC{
		HolderAddress: params.HolderAddress,
		VcId:          params.VCID,
		ReasonText:    params.ReasonText,
	}

	addr, err := sdk.AccAddressFromBech32(params.HolderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid holder address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke VC: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// AdminRevokeVCParams contains parameters for admin revocation
type AdminRevokeVCParams struct {
	Authority string
	VCID      string
	Reason    vcregistrypb.RevocationReason
	Evidence  string
}

// AdminRevokeVC revokes a credential (governance)
func (c *Client) AdminRevokeVC(ctx context.Context, params *AdminRevokeVCParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Authority == "" {
		return nil, fmt.Errorf("authority is required")
	}
	if params.VCID == "" {
		return nil, fmt.Errorf("VC ID is required")
	}

	msg := &vcregistrypb.MsgAdminRevokeVC{
		Authority: params.Authority,
		VcId:      params.VCID,
		Reason:    params.Reason,
		Evidence:  params.Evidence,
	}

	addr, err := sdk.AccAddressFromBech32(params.Authority)
	if err != nil {
		return nil, fmt.Errorf("invalid authority address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to admin revoke VC: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// RegisterDIDParams contains parameters for registering a DID
type RegisterDIDParams struct {
	Controller          string
	DID                 string
	VerificationMethods []*vcregistrypb.VerificationMethod
	MetadataURI         string
}

// RegisterDID registers a new DID document
func (c *Client) RegisterDID(ctx context.Context, params *RegisterDIDParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Controller == "" {
		return nil, fmt.Errorf("controller is required")
	}
	if params.DID == "" {
		return nil, fmt.Errorf("DID is required")
	}

	msg := &vcregistrypb.MsgRegisterDID{
		Controller:          params.Controller,
		Did:                 params.DID,
		VerificationMethods: params.VerificationMethods,
		MetadataUri:         params.MetadataURI,
	}

	addr, err := sdk.AccAddressFromBech32(params.Controller)
	if err != nil {
		return nil, fmt.Errorf("invalid controller address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to register DID: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// UpdateDIDParams contains parameters for updating a DID document
type UpdateDIDParams struct {
	Controller          string
	DID                 string
	VerificationMethods []*vcregistrypb.VerificationMethod
	MetadataURI         string
}

// UpdateDIDDocument updates a DID document
func (c *Client) UpdateDIDDocument(ctx context.Context, params *UpdateDIDParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Controller == "" {
		return nil, fmt.Errorf("controller is required")
	}
	if params.DID == "" {
		return nil, fmt.Errorf("DID is required")
	}

	msg := &vcregistrypb.MsgUpdateDIDDocument{
		Controller:          params.Controller,
		Did:                 params.DID,
		VerificationMethods: params.VerificationMethods,
		MetadataUri:         params.MetadataURI,
	}

	addr, err := sdk.AccAddressFromBech32(params.Controller)
	if err != nil {
		return nil, fmt.Errorf("invalid controller address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to update DID document: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// GetVC retrieves a specific verifiable credential
func (c *Client) GetVC(ctx context.Context, vcID string) (*vcregistrypb.VCRecord, error) {
	if vcID == "" {
		return nil, fmt.Errorf("VC ID is required")
	}

	req := &vcregistrypb.QueryGetVCRequest{
		VcId: vcID,
	}

	resp, err := c.queryClient.GetVC(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get VC: %w", err)
	}

	if !resp.Exists {
		return nil, fmt.Errorf("VC not found: %s", vcID)
	}

	return resp.Vc, nil
}

// ListVCsParams contains parameters for listing VCs
type ListVCsParams struct {
	HolderAddress string
	StatusFilter  vcregistrypb.VCStatus
	TypeFilter    vcregistrypb.VCType
}

// ListVCs lists all VCs for a user
func (c *Client) ListVCs(ctx context.Context, params *ListVCsParams) ([]*vcregistrypb.VCRecord, error) {
	if params == nil || params.HolderAddress == "" {
		return nil, fmt.Errorf("holder address is required")
	}

	req := &vcregistrypb.QueryListUserVCsRequest{
		HolderAddress: params.HolderAddress,
		StatusFilter:  params.StatusFilter,
		TypeFilter:    params.TypeFilter,
	}

	resp, err := c.queryClient.ListUserVCs(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list VCs: %w", err)
	}

	return resp.Vcs, nil
}

// VerifyVC checks if a VC is valid
func (c *Client) VerifyVC(ctx context.Context, vcID string) (bool, *vcregistrypb.QueryCheckVCStatusResponse, error) {
	if vcID == "" {
		return false, nil, fmt.Errorf("VC ID is required")
	}

	req := &vcregistrypb.QueryCheckVCStatusRequest{
		VcId: vcID,
	}

	resp, err := c.queryClient.CheckVCStatus(ctx, req)
	if err != nil {
		return false, nil, fmt.Errorf("failed to verify VC: %w", err)
	}

	return resp.Valid, resp, nil
}

// BatchVCStatus checks multiple VCs
func (c *Client) BatchVCStatus(ctx context.Context, vcIDs []string) (map[string]*vcregistrypb.VCStatusInfo, error) {
	if len(vcIDs) == 0 {
		return nil, fmt.Errorf("VC IDs are required")
	}

	req := &vcregistrypb.QueryBatchVCStatusRequest{
		VcIds: vcIDs,
	}

	resp, err := c.queryClient.BatchVCStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to batch check VC status: %w", err)
	}

	return resp.Statuses, nil
}

// ResolveDID resolves a DID to its document
func (c *Client) ResolveDID(ctx context.Context, did string) (*vcregistrypb.DIDDocument, []*vcregistrypb.VCRecord, error) {
	if did == "" {
		return nil, nil, fmt.Errorf("DID is required")
	}

	req := &vcregistrypb.QueryResolveDIDRequest{
		Did: did,
	}

	resp, err := c.queryClient.ResolveDID(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve DID: %w", err)
	}

	if !resp.Exists {
		return nil, nil, fmt.Errorf("DID not found: %s", did)
	}

	return resp.DidDocument, resp.Credentials, nil
}

// GetDIDByAddress gets DID by controller address
func (c *Client) GetDIDByAddress(ctx context.Context, address string) ([]string, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &vcregistrypb.QueryGetDIDByAddressRequest{
		Controller: address,
	}

	resp, err := c.queryClient.GetDIDByAddress(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get DID by address: %w", err)
	}

	return resp.Dids, nil
}

// ValidateMintEligibility checks if user can mint a VC
func (c *Client) ValidateMintEligibility(ctx context.Context, holderAddress string, vcType vcregistrypb.VCType, vcTypeCustom string) (*vcregistrypb.QueryValidateMintEligibilityResponse, error) {
	if holderAddress == "" {
		return nil, fmt.Errorf("holder address is required")
	}

	req := &vcregistrypb.QueryValidateMintEligibilityRequest{
		HolderAddress: holderAddress,
		VcType:        vcType,
		VcTypeCustom:  vcTypeCustom,
	}

	resp, err := c.queryClient.ValidateMintEligibility(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate mint eligibility: %w", err)
	}

	return resp, nil
}

// GetVCPolicy retrieves a VC policy
func (c *Client) GetVCPolicy(ctx context.Context, vcTypeName string) (*vcregistrypb.VCPolicy, error) {
	if vcTypeName == "" {
		return nil, fmt.Errorf("VC type name is required")
	}

	req := &vcregistrypb.QueryGetVCPolicyRequest{
		VcTypeName: vcTypeName,
	}

	resp, err := c.queryClient.GetVCPolicy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get VC policy: %w", err)
	}

	if !resp.Exists {
		return nil, fmt.Errorf("VC policy not found: %s", vcTypeName)
	}

	return resp.Policy, nil
}

// ListVCPolicies lists all policies
func (c *Client) ListVCPolicies(ctx context.Context, statusFilter vcregistrypb.VCPolicyStatus) ([]*vcregistrypb.VCPolicy, error) {
	req := &vcregistrypb.QueryListVCPoliciesRequest{
		StatusFilter: statusFilter,
	}

	resp, err := c.queryClient.ListVCPolicies(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list VC policies: %w", err)
	}

	return resp.Policies, nil
}

// GetRevocationList retrieves the revocation Merkle root
func (c *Client) GetRevocationList(ctx context.Context) (*vcregistrypb.RevocationList, error) {
	req := &vcregistrypb.QueryGetRevocationListRequest{}

	resp, err := c.queryClient.GetRevocationList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get revocation list: %w", err)
	}

	return resp.RevocationList, nil
}

// CheckRevocation checks if a VC is revoked
func (c *Client) CheckRevocation(ctx context.Context, vcID string) (bool, *vcregistrypb.RevocationRecord, []byte, error) {
	if vcID == "" {
		return false, nil, nil, fmt.Errorf("VC ID is required")
	}

	req := &vcregistrypb.QueryCheckRevocationRequest{
		VcId: vcID,
	}

	resp, err := c.queryClient.CheckRevocation(ctx, req)
	if err != nil {
		return false, nil, nil, fmt.Errorf("failed to check revocation: %w", err)
	}

	return resp.Revoked, resp.Record, resp.MerkleProof, nil
}

// GetStats retrieves registry statistics
func (c *Client) GetStats(ctx context.Context) (*vcregistrypb.QueryStatsResponse, error) {
	req := &vcregistrypb.QueryStatsRequest{}

	resp, err := c.queryClient.Stats(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return resp, nil
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*vcregistrypb.Params, error) {
	req := &vcregistrypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}
