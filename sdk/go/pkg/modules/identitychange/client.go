package identitychange

import (
	"github.com/aura-chain/aura/sdk/go/client"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the identitychange module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient identitychangepb.QueryClient
	msgClient   identitychangepb.MsgClient
}

// NewClient creates a new identitychange client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: identitychangepb.NewQueryClient(grpcConn),
		msgClient:   identitychangepb.NewMsgClient(grpcConn),
	}
}

// Note: Module has QueryClient and MsgClient but no Params query
// Available queries: IdentityRecord, IdentityChangeRequest, IdentityChangeHistory
// Add specific identity change queries/transactions as needed based on proto definitions
