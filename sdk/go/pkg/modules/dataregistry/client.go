package dataregistry

import (
	"context"
	"fmt"

	dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the dataregistry module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient dataregistrypb.QueryClient
	msgClient   dataregistrypb.MsgClient
}

// NewClient creates a new dataregistry client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: dataregistrypb.NewQueryClient(grpcConn),
		msgClient:   dataregistrypb.NewMsgClient(grpcConn),
	}
}

// GetParams retrieves module parameters
func (c *Client) GetParams(ctx context.Context) (*dataregistrypb.Params, error) {
	req := &dataregistrypb.QueryParamsRequest{}

	resp, err := c.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	return resp.Params, nil
}
