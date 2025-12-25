package aiassistant

import (
	"context"
	"fmt"

	"github.com/aura-chain/aura/sdk/go/client"
	"github.com/aura-chain/aura/sdk/go/pkg/types"
	aiassistantpb "github.com/aequitas/aura/proto/aura/aiassistant/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the aiassistant module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient aiassistantpb.QueryClient
	msgClient   aiassistantpb.MsgClient
}

// NewClient creates a new aiassistant client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: aiassistantpb.NewQueryClient(grpcConn),
		msgClient:   aiassistantpb.NewMsgClient(grpcConn),
	}
}

// RegisterAssistantParams contains parameters for registering an AI assistant
type RegisterAssistantParams struct {
	AssistantAddress  string
	OwnerAddress      string
	Locales           []string
	ModelHash         string
	ApiKeyFingerprint string
	Stake             aiassistantpb.Balance
	Sponsorship       aiassistantpb.Balance
}

// RegisterAssistant registers a new AI assistant on the chain
func (c *Client) RegisterAssistant(ctx context.Context, params *RegisterAssistantParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.AssistantAddress == "" {
		return nil, fmt.Errorf("assistant address is required")
	}
	if params.OwnerAddress == "" {
		return nil, fmt.Errorf("owner address is required")
	}
	if len(params.Locales) == 0 {
		return nil, fmt.Errorf("at least one locale is required")
	}

	msg := &aiassistantpb.MsgRegisterAssistant{
		AssistantAddress:  params.AssistantAddress,
		OwnerAddress:      params.OwnerAddress,
		Locales:           params.Locales,
		ModelHash:         params.ModelHash,
		ApiKeyFingerprint: params.ApiKeyFingerprint,
		Stake:             params.Stake,
		Sponsorship:       params.Sponsorship,
	}

	addr, err := sdk.AccAddressFromBech32(params.OwnerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to register assistant: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// UpdateLocalesParams contains parameters for updating assistant locales
type UpdateLocalesParams struct {
	AssistantAddress string
	Locales          []string
}

// UpdateLocales updates the locales an assistant supports
func (c *Client) UpdateLocales(ctx context.Context, params *UpdateLocalesParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.AssistantAddress == "" {
		return nil, fmt.Errorf("assistant address is required")
	}
	if len(params.Locales) == 0 {
		return nil, fmt.Errorf("at least one locale is required")
	}

	msg := &aiassistantpb.MsgUpdateLocales{
		AssistantAddress: params.AssistantAddress,
		Locales:          params.Locales,
	}

	addr, err := sdk.AccAddressFromBech32(params.AssistantAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid assistant address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to update locales: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// Heartbeat sends a heartbeat to indicate the assistant is alive
func (c *Client) Heartbeat(ctx context.Context, assistantAddress string) (*types.TxResponse, error) {
	if assistantAddress == "" {
		return nil, fmt.Errorf("assistant address is required")
	}

	msg := &aiassistantpb.MsgHeartbeat{
		AssistantAddress: assistantAddress,
	}

	addr, err := sdk.AccAddressFromBech32(assistantAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid assistant address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to send heartbeat: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// ReportMisbehaviorParams contains parameters for reporting assistant misbehavior
type ReportMisbehaviorParams struct {
	Reporter         string
	AssistantAddress string
	Infraction       string
	EvidenceHash     string
}

// ReportMisbehavior reports misbehavior by an AI assistant
func (c *Client) ReportMisbehavior(ctx context.Context, params *ReportMisbehaviorParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Reporter == "" {
		return nil, fmt.Errorf("reporter is required")
	}
	if params.AssistantAddress == "" {
		return nil, fmt.Errorf("assistant address is required")
	}

	msg := &aiassistantpb.MsgReportMisbehavior{
		Reporter:         params.Reporter,
		AssistantAddress: params.AssistantAddress,
		Infraction:       params.Infraction,
		EvidenceHash:     params.EvidenceHash,
	}

	addr, err := sdk.AccAddressFromBech32(params.Reporter)
	if err != nil {
		return nil, fmt.Errorf("invalid reporter address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to report misbehavior: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// GetAssistant retrieves information about an AI assistant
func (c *Client) GetAssistant(ctx context.Context, address string) (*aiassistantpb.Assistant, error) {
	if address == "" {
		return nil, fmt.Errorf("assistant address is required")
	}

	req := &aiassistantpb.QueryAssistantRequest{
		AssistantAddress: address,
	}

	resp, err := c.queryClient.Assistant(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get assistant: %w", err)
	}

	return resp.Assistant, nil
}

// ListAssistants lists all registered AI assistants
func (c *Client) ListAssistants(ctx context.Context) ([]*aiassistantpb.Assistant, error) {
	req := &aiassistantpb.QueryAssistantsRequest{}

	resp, err := c.queryClient.Assistants(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list assistants: %w", err)
	}

	return resp.Assistants, nil
}

// GetAssistantsByLocale retrieves assistants that support a specific locale
func (c *Client) GetAssistantsByLocale(ctx context.Context, locale string) ([]*aiassistantpb.Assistant, error) {
	if locale == "" {
		return nil, fmt.Errorf("locale is required")
	}

	req := &aiassistantpb.QueryAssistantsByLocaleRequest{
		Locale: locale,
	}

	resp, err := c.queryClient.AssistantsByLocale(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get assistants by locale: %w", err)
	}

	return resp.Assistants, nil
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*aiassistantpb.Params, error) {
	req := &aiassistantpb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}
