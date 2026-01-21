package inclusionroutines

import (
	"context"
	"fmt"

	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the inclusionroutines module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient inclusionroutinespb.QueryClient
	msgClient   inclusionroutinespb.MsgClient
}

// NewClient creates a new inclusionroutines client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: inclusionroutinespb.NewQueryClient(grpcConn),
		msgClient:   inclusionroutinespb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*inclusionroutinespb.Params, error) {
	req := &inclusionroutinespb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}
