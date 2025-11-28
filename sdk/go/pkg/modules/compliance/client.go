package compliance

import (
	"github.com/aura-chain/aura/sdk/go/client"
	_ "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the compliance module
type Client struct {
	auraClient *client.Client
	grpcConn   *grpc.ClientConn
	// Note: Compliance module doesn't currently export gRPC query/msg services
	// Implementation pending proto service definition
}

// NewClient creates a new compliance client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient: auraClient,
		grpcConn:   grpcConn,
	}
}

// TODO: Add methods once gRPC services are defined in proto files
