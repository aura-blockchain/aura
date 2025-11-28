package types

import (
	"context"
)

// QueryServer is the server API for Query service
type QueryServer interface {
	// QueryStats returns query usage statistics
	QueryStats(context.Context, *QueryStatsRequest) (*QueryStatsResponse, error)

	// MessageStats returns message usage statistics
	MessageStats(context.Context, *MessageStatsRequest) (*MessageStatsResponse, error)

	// AllStats returns both query and message statistics
	AllStats(context.Context, *AllStatsRequest) (*AllStatsResponse, error)
}

// QueryClient is a stub for the query client interface
type QueryClient interface {
	QueryStats(ctx context.Context, req *QueryStatsRequest) (*QueryStatsResponse, error)
	MessageStats(ctx context.Context, req *MessageStatsRequest) (*MessageStatsResponse, error)
	AllStats(ctx context.Context, req *AllStatsRequest) (*AllStatsResponse, error)
}

// NewQueryClient creates a stub query client (to be implemented with gRPC)
func NewQueryClient(clientCtx interface{}) QueryClient {
	return nil // Stub - would be implemented with proper gRPC client
}

// QueryStatsRequest is the request for query statistics
type QueryStatsRequest struct{}

// QueryStatsResponse is the response containing query statistics
type QueryStatsResponse struct {
	QueryStats map[string]uint64 `json:"query_stats"`
}

// MessageStatsRequest is the request for message statistics
type MessageStatsRequest struct{}

// MessageStatsResponse is the response containing message statistics
type MessageStatsResponse struct {
	MessageStats map[string]uint64 `json:"message_stats"`
}

// AllStatsRequest is the request for all statistics
type AllStatsRequest struct{}

// AllStatsResponse is the response containing all statistics
type AllStatsResponse struct {
	QueryStats   map[string]uint64 `json:"query_stats"`
	MessageStats map[string]uint64 `json:"message_stats"`
}

// RegisterQueryServer registers the query server
func RegisterQueryServer(s interface{}, impl QueryServer) {
	// Stub - would be implemented by gRPC server registration
}
