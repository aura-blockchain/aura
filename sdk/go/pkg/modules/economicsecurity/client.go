package economicsecurity

import (
	"context"
	"fmt"

	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the economicsecurity module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient economicsecuritypb.QueryClient
	msgClient   economicsecuritypb.MsgClient
}

// NewClient creates a new economicsecurity client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: economicsecuritypb.NewQueryClient(grpcConn),
		msgClient:   economicsecuritypb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*economicsecuritypb.Params, error) {
	req := &economicsecuritypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}
