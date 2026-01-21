package cryptography

import (
	"context"
	"fmt"

	cryptographypb "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the cryptography module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient cryptographypb.QueryClient
	msgClient   cryptographypb.MsgClient
}

// NewClient creates a new cryptography client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: cryptographypb.NewQueryClient(grpcConn),
		msgClient:   cryptographypb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*cryptographypb.Params, error) {
	req := &cryptographypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}
