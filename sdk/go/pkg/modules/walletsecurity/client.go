package walletsecurity

import (
	walletsecuritypb "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the walletsecurity module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient walletsecuritypb.QueryClient
	msgClient   walletsecuritypb.MsgClient
}

// NewClient creates a new walletsecurity client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: walletsecuritypb.NewQueryClient(grpcConn),
		msgClient:   walletsecuritypb.NewMsgClient(grpcConn),
	}
}

// Note: Module has QueryClient and MsgClient but no Params query
// Add specific wallet security queries/transactions as needed based on proto definitions
