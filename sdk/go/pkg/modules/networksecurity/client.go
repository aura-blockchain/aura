package networksecurity

import (
	"context"
	"fmt"

	networksecuritypb "github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the networksecurity module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient networksecuritypb.QueryClient
	msgClient   networksecuritypb.MsgClient
}

// NewClient creates a new networksecurity client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: networksecuritypb.NewQueryClient(grpcConn),
		msgClient:   networksecuritypb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*networksecuritypb.Params, error) {
	req := &networksecuritypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return &resp.Params, nil
}
