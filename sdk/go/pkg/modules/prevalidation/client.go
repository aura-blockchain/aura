package prevalidation

import (
	"context"
	"fmt"

	prevalidationpb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the prevalidation module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient prevalidationpb.QueryClient
}

// NewClient creates a new prevalidation client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: prevalidationpb.NewQueryClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*prevalidationpb.Params, error) {
	req := &prevalidationpb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}

// GetPreValidatedTransaction retrieves a pre-validated transaction by ID
func (c *Client) GetPreValidatedTransaction(ctx context.Context, id string) (*prevalidationpb.PreValidatedTransaction, error) {
	if id == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	req := &prevalidationpb.QueryPreValidatedTransactionRequest{
		Id: id,
	}

	resp, err := c.queryClient.PreValidatedTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get pre-validated transaction: %w", err)
	}

	return resp.Transaction, nil
}

// GetPreValidatedTransactionsByStatus retrieves pre-validated transactions by status
func (c *Client) GetPreValidatedTransactionsByStatus(ctx context.Context, status prevalidationpb.ValidationStatus) ([]*prevalidationpb.PreValidatedTransaction, error) {
	req := &prevalidationpb.QueryPreValidatedTransactionsByStatusRequest{
		Status: status,
	}

	resp, err := c.queryClient.PreValidatedTransactionsByStatus(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get pre-validated transactions by status: %w", err)
	}

	return resp.Transactions, nil
}

// GetPreValidatedTransactionsBySigner retrieves pre-validated transactions by signer
func (c *Client) GetPreValidatedTransactionsBySigner(ctx context.Context, signer string) ([]*prevalidationpb.PreValidatedTransaction, error) {
	if signer == "" {
		return nil, fmt.Errorf("signer address is required")
	}

	req := &prevalidationpb.QueryPreValidatedTransactionsBySignerRequest{
		Signer: signer,
	}

	resp, err := c.queryClient.PreValidatedTransactionsBySigner(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get pre-validated transactions by signer: %w", err)
	}

	return resp.Transactions, nil
}

// GetTemplate retrieves a validation template by ID
func (c *Client) GetTemplate(ctx context.Context, templateID string) (*prevalidationpb.ValidationTemplate, error) {
	if templateID == "" {
		return nil, fmt.Errorf("template ID is required")
	}

	req := &prevalidationpb.QueryTemplateRequest{
		TemplateId: templateID,
	}

	resp, err := c.queryClient.Template(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return resp.Template, nil
}

// GetAllTemplates retrieves all validation templates
func (c *Client) GetAllTemplates(ctx context.Context) ([]*prevalidationpb.ValidationTemplate, error) {
	req := &prevalidationpb.QueryAllTemplatesRequest{}

	resp, err := c.queryClient.AllTemplates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get all templates: %w", err)
	}

	return resp.Templates, nil
}

// GetTemplatesByType retrieves templates by transaction type
func (c *Client) GetTemplatesByType(ctx context.Context, txType prevalidationpb.TransactionType) ([]*prevalidationpb.ValidationTemplate, error) {
	req := &prevalidationpb.QueryTemplatesByTypeRequest{
		TxType: txType,
	}

	resp, err := c.queryClient.TemplatesByType(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates by type: %w", err)
	}

	return resp.Templates, nil
}

// GetMetrics retrieves pre-validation metrics
func (c *Client) GetMetrics(ctx context.Context) (*prevalidationpb.PreValidationMetrics, error) {
	req := &prevalidationpb.QueryMetricsRequest{}

	resp, err := c.queryClient.Metrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return resp.Metrics, nil
}

// GetMetricsByType retrieves metrics for a specific transaction type
func (c *Client) GetMetricsByType(ctx context.Context, txType prevalidationpb.TransactionType) (*prevalidationpb.TypeMetrics, error) {
	req := &prevalidationpb.QueryMetricsByTypeRequest{
		TxType: txType,
	}

	resp, err := c.queryClient.MetricsByType(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics by type: %w", err)
	}

	return resp.Metrics, nil
}

// EstimateGasParams contains parameters for gas estimation
type EstimateGasParams struct {
	Sender    string
	Recipient string
	Amount    string
	Data      []byte
	TxType    prevalidationpb.TransactionType
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(ctx context.Context, params *EstimateGasParams) (uint64, uint64, error) {
	if params == nil {
		return 0, 0, fmt.Errorf("params cannot be nil")
	}
	if params.Sender == "" {
		return 0, 0, fmt.Errorf("sender is required")
	}

	req := &prevalidationpb.QueryEstimateGasRequest{
		Sender:    params.Sender,
		Recipient: params.Recipient,
		Amount:    params.Amount,
		Data:      params.Data,
		TxType:    params.TxType,
	}

	resp, err := c.queryClient.EstimateGas(ctx, req)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	return resp.GasEstimate, resp.GasLimit, nil
}

// ValidateTransactionParams contains parameters for transaction validation
type ValidateTransactionParams struct {
	Sender    string
	Recipient string
	Amount    string
	Data      []byte
	Nonce     uint64
	TxType    prevalidationpb.TransactionType
}

// ValidateTransactionResponse contains the validation result
type ValidateTransactionResponse struct {
	Valid             bool
	GasEstimate       uint64
	Error             string
	SufficientBalance bool
}

// ValidateTransaction validates a transaction without pre-validating it
func (c *Client) ValidateTransaction(ctx context.Context, params *ValidateTransactionParams) (*ValidateTransactionResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Sender == "" {
		return nil, fmt.Errorf("sender is required")
	}

	req := &prevalidationpb.QueryValidateTransactionRequest{
		Sender:    params.Sender,
		Recipient: params.Recipient,
		Amount:    params.Amount,
		Data:      params.Data,
		Nonce:     params.Nonce,
		TxType:    params.TxType,
	}

	resp, err := c.queryClient.ValidateTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate transaction: %w", err)
	}

	return &ValidateTransactionResponse{
		Valid:             resp.Valid,
		GasEstimate:       resp.GasEstimate,
		Error:             resp.Error,
		SufficientBalance: resp.SufficientBalance,
	}, nil
}

// GetNonce retrieves the current nonce for an address
func (c *Client) GetNonce(ctx context.Context, address string) (uint64, error) {
	if address == "" {
		return 0, fmt.Errorf("address is required")
	}

	req := &prevalidationpb.QueryGetNonceRequest{
		Address: address,
	}

	resp, err := c.queryClient.GetNonce(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	return resp.Nonce, nil
}
