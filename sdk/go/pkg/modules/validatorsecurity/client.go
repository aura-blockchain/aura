package validatorsecurity

import (
	"context"
	"fmt"

	"github.com/aura-chain/aura/sdk/go/client"
	validatorsecuritypb "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the validatorsecurity module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient validatorsecuritypb.QueryClient
	msgClient   validatorsecuritypb.MsgClient
}

// NewClient creates a new validatorsecurity client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: validatorsecuritypb.NewQueryClient(grpcConn),
		msgClient:   validatorsecuritypb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*validatorsecuritypb.ValidatorSecurityParams, error) {
	req := &validatorsecuritypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return &resp.Params, nil
}
