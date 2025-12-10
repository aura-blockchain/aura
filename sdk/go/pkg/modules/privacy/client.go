package privacy

import (
	"context"
	"fmt"

	"github.com/aura-chain/aura/sdk/go/client"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the privacy module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient privacypb.QueryClient
	msgClient   privacypb.MsgClient
}

// NewClient creates a new privacy client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: privacypb.NewQueryClient(grpcConn),
		msgClient:   privacypb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*privacypb.Params, error) {
	req := &privacypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return &resp.Params, nil
}
