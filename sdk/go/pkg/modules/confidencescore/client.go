package confidencescore

import (
	"context"
	"fmt"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the confidencescore module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient confidencescorepb.QueryClient
	msgClient   confidencescorepb.MsgClient
}

// NewClient creates a new confidencescore client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: confidencescorepb.NewQueryClient(grpcConn),
		msgClient:   confidencescorepb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*confidencescorepb.Params, error) {
	req := &confidencescorepb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}
