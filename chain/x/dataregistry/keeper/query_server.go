package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/aequitas/aura/chain/x/dataregistry/types"
	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// Ensure queryServer implements QueryServer interface
var _ pb.QueryServer = &queryServer{}

// queryServer implements the gRPC QueryServer interface
type queryServer struct {
	pb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer creates a new QueryServer instance
func NewQueryServer(k *Keeper) pb.QueryServer {
	return &queryServer{keeper: k}
}

// DataItem retrieves a specific data item
func (s *queryServer) DataItem(ctx context.Context, req *pb.QueryDataItemRequest) (*pb.QueryDataItemResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.DataId == "" {
		return nil, fmt.Errorf("data_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	item, exists := s.keeper.GetDataItem(sdkCtx, req.DataId)
	if !exists {
		return &pb.QueryDataItemResponse{
			DataItem:  nil,
			Exists:    false,
			HasAccess: false,
		}, nil
	}

	hasAccess := s.keeper.CheckAccess(sdkCtx, req.DataId, req.Requester)

	// If no access, return limited info
	if !hasAccess {
		return &pb.QueryDataItemResponse{
			DataItem:  nil,
			Exists:    true,
			HasAccess: false,
		}, nil
	}

	return &pb.QueryDataItemResponse{
		DataItem:  &item,
		Exists:    true,
		HasAccess: true,
	}, nil
}

// UserDataItems lists all data items for a user
func (s *queryServer) UserDataItems(ctx context.Context, req *pb.QueryUserDataItemsRequest) (*pb.QueryUserDataItemsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.OwnerAddress == "" {
		return nil, fmt.Errorf("owner_address cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	items := s.keeper.ListUserDataItems(sdkCtx, req.OwnerAddress, req.TypeFilter, req.StatusFilter)

	pagedItems, pageRes, err := paginateDataItems(items, req.Pagination)
	if err != nil {
		return nil, err
	}

	itemPtrs := make([]*pb.DataItem, len(pagedItems))
	for i := range pagedItems {
		itemPtrs[i] = &pagedItems[i]
	}

	return &pb.QueryUserDataItemsResponse{
		DataItems:  itemPtrs,
		Pagination: pageRes,
	}, nil
}

// SearchDataItems searches for data items
func (s *queryServer) SearchDataItems(ctx context.Context, req *pb.QuerySearchDataItemsRequest) (*pb.QuerySearchDataItemsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	items := s.keeper.SearchDataItems(
		sdkCtx,
		req.SearchQuery,
		req.Tags,
		req.TypeFilter,
		req.NearLocation,
		req.RadiusKm,
		req.Requester,
	)

	pagedItems, pageRes, err := paginateDataItems(items, req.Pagination)
	if err != nil {
		return nil, err
	}

	itemPtrs := make([]*pb.DataItem, len(pagedItems))
	for i := range pagedItems {
		itemPtrs[i] = &pagedItems[i]
	}

	return &pb.QuerySearchDataItemsResponse{
		DataItems:  itemPtrs,
		Pagination: pageRes,
	}, nil
}

// DataItemVerifications retrieves verifications for a data item
func (s *queryServer) DataItemVerifications(ctx context.Context, req *pb.QueryDataItemVerificationsRequest) (*pb.QueryDataItemVerificationsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.DataId == "" {
		return nil, fmt.Errorf("data_id cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	verifications, err := s.keeper.GetDataItemVerifications(sdkCtx, req.DataId)
	if err != nil {
		return nil, err
	}

	return &pb.QueryDataItemVerificationsResponse{
		Verifications: verifications,
	}, nil
}

// Stats retrieves registry statistics
func (s *queryServer) Stats(ctx context.Context, req *pb.QueryStatsRequest) (*pb.QueryStatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	stats := s.keeper.GetStats(sdkCtx)

	return &pb.QueryStatsResponse{
		TotalDataItems:     stats.TotalDataItems,
		TotalVerifiedItems: stats.TotalVerifiedItems,
		TotalStorageBytes:  stats.TotalStorageBytes,
		ItemsByType:        stats.ItemsByType,
		TotalVerifications: stats.TotalVerifications,
	}, nil
}

// Params retrieves module parameters
func (s *queryServer) Params(ctx context.Context, req *pb.QueryParamsRequest) (*pb.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	params, _ := s.keeper.GetParams(ctx)

	return &pb.QueryParamsResponse{
		Params: &params,
	}, nil
}

func paginateDataItems(items []types.DataItem, pageReq *query.PageRequest) ([]types.DataItem, *query.PageResponse, error) {
	if pageReq == nil {
		return items, nil, nil
	}

	if len(pageReq.Key) != 0 {
		return nil, nil, fmt.Errorf("key-based pagination is not supported for data registry")
	}

	offset := int(pageReq.Offset)
	if offset > len(items) {
		offset = len(items)
	}

	limit := len(items)
	if pageReq.Limit > 0 {
		limit = int(pageReq.Limit)
	}

	pageRes := &query.PageResponse{}
	if pageReq.CountTotal {
		pageRes.Total = uint64(len(items))
	}

	if offset >= len(items) {
		return []types.DataItem{}, pageRes, nil
	}

	remaining := len(items) - offset
	if limit > remaining {
		limit = remaining
	}

	end := offset + limit
	paged := items[offset:end]

	if end < len(items) {
		pageRes.NextKey = []byte(fmt.Sprintf("%d", end))
	}

	return paged, pageRes, nil
}
