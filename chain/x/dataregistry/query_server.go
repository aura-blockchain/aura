package dataregistry

import (
	"context"

	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// QueryServer defines the query server interface
type QueryServer interface {
	DataItem(ctx context.Context, req *QueryDataItemRequest) (*QueryDataItemResponse, error)
	UserDataItems(ctx context.Context, req *QueryUserDataItemsRequest) (*QueryUserDataItemsResponse, error)
	SearchDataItems(ctx context.Context, req *QuerySearchDataItemsRequest) (*QuerySearchDataItemsResponse, error)
	DataItemVerifications(ctx context.Context, req *QueryDataItemVerificationsRequest) (*QueryDataItemVerificationsResponse, error)
	Stats(ctx context.Context, req *QueryStatsRequest) (*QueryStatsResponse, error)
	Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
}

// queryServer implements QueryServer
type queryServer struct {
	dataregistrypb.UnimplementedQueryServer
	keeper *keeper.Keeper
}

// NewQueryServer creates a new QueryServer
func NewQueryServer(k *keeper.Keeper) QueryServer {
	return &queryServer{keeper: k}
}

// Query request/response types
type QueryDataItemRequest struct {
	DataID    string
	Requester string
}

type QueryDataItemResponse struct {
	DataItem  *types.DataItem
	Exists    bool
	HasAccess bool
}

type QueryUserDataItemsRequest struct {
	OwnerAddress string
	TypeFilter   types.DataItemType
	StatusFilter types.DataItemStatus
}

type QueryUserDataItemsResponse struct {
	DataItems []types.DataItem
}

type QuerySearchDataItemsRequest struct {
	SearchQuery  string
	Tags         []string
	TypeFilter   types.DataItemType
	NearLocation *types.GeoLocation
	RadiusKM     float64
	Requester    string
}

type QuerySearchDataItemsResponse struct {
	DataItems []types.DataItem
}

type QueryDataItemVerificationsRequest struct {
	DataID string
}

type QueryDataItemVerificationsResponse struct {
	Verifications []types.Verification
}

type QueryStatsRequest struct{}

type QueryStatsResponse struct {
	Stats types.RegistryStats
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params types.Params
}

// DataItem retrieves a specific data item
func (s *queryServer) DataItem(ctx context.Context, req *QueryDataItemRequest) (*QueryDataItemResponse, error) {
	item, exists := s.keeper.GetDataItem(req.DataID)
	if !exists {
		return &QueryDataItemResponse{
			DataItem:  nil,
			Exists:    false,
			HasAccess: false,
		}, nil
	}

	hasAccess := s.keeper.CheckAccess(req.DataID, req.Requester)

	// If no access, return limited info
	if !hasAccess {
		return &QueryDataItemResponse{
			DataItem:  nil,
			Exists:    true,
			HasAccess: false,
		}, nil
	}

	return &QueryDataItemResponse{
		DataItem:  &item,
		Exists:    true,
		HasAccess: true,
	}, nil
}

// UserDataItems lists all data items for a user
func (s *queryServer) UserDataItems(ctx context.Context, req *QueryUserDataItemsRequest) (*QueryUserDataItemsResponse, error) {
	items := s.keeper.ListUserDataItems(req.OwnerAddress, req.TypeFilter, req.StatusFilter)

	return &QueryUserDataItemsResponse{
		DataItems: items,
	}, nil
}

// SearchDataItems searches for data items
func (s *queryServer) SearchDataItems(ctx context.Context, req *QuerySearchDataItemsRequest) (*QuerySearchDataItemsResponse, error) {
	items := s.keeper.SearchDataItems(
		req.SearchQuery,
		req.Tags,
		req.TypeFilter,
		req.NearLocation,
		req.RadiusKM,
		req.Requester,
	)

	return &QuerySearchDataItemsResponse{
		DataItems: items,
	}, nil
}

// DataItemVerifications retrieves verifications for a data item
func (s *queryServer) DataItemVerifications(ctx context.Context, req *QueryDataItemVerificationsRequest) (*QueryDataItemVerificationsResponse, error) {
	verifications, err := s.keeper.GetDataItemVerifications(req.DataID)
	if err != nil {
		return nil, err
	}

	return &QueryDataItemVerificationsResponse{
		Verifications: verifications,
	}, nil
}

// Stats retrieves registry statistics
func (s *queryServer) Stats(ctx context.Context, req *QueryStatsRequest) (*QueryStatsResponse, error) {
	stats := s.keeper.GetStats()

	return &QueryStatsResponse{
		Stats: stats,
	}, nil
}

// Params retrieves module parameters
func (s *queryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	params := s.keeper.GetParams()

	return &QueryParamsResponse{
		Params: params,
	}, nil
}
