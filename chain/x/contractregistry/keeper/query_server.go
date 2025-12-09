package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.QueryServer = (*queryServer)(nil)

type queryServer struct {
	pb.UnimplementedQueryServer
	*Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) pb.QueryServer {
	return &queryServer{Keeper: keeper}
}

// ContractInfo returns information about a registered contract
func (qs queryServer) ContractInfo(goCtx context.Context, req *pb.QueryContractInfoRequest) (*pb.QueryContractInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	info, found := qs.GetContractInfo(ctx, req.ContractAddress)
	if !found {
		return nil, status.Error(codes.NotFound, "contract not found")
	}

	return &pb.QueryContractInfoResponse{
		Contract: &info,
	}, nil
}

// ContractsByCreator returns all contracts created by a specific address
func (qs queryServer) ContractsByCreator(goCtx context.Context, req *pb.QueryContractsByCreatorRequest) (*pb.QueryContractsByCreatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var contracts []*pb.ContractInfo

	// Get all contracts - simplified for now
	var creatorContracts []*pb.ContractInfo
	qs.IterateContractInfo(ctx, func(info *pb.ContractInfo) bool {
		if info.Creator == req.CreatorAddress {
			creatorContracts = append(creatorContracts, info)
		}
		return false
	})

	// Apply pagination if requested
	if req.Pagination != nil {
		// Manual pagination since we're using an index
		start := int(req.Pagination.Offset)
		limit := int(req.Pagination.Limit)
		if limit == 0 {
			limit = 100 // default
		}

		end := start + limit
		if end > len(creatorContracts) {
			end = len(creatorContracts)
		}

		if start < len(creatorContracts) {
			for i := start; i < end; i++ {
				contracts = append(contracts, creatorContracts[i])
			}
		}
	} else {
		for i := range creatorContracts {
			contracts = append(contracts, creatorContracts[i])
		}
	}

	return &pb.QueryContractsByCreatorResponse{
		Contracts: contracts,
		Pagination: &query.PageResponse{
			Total: uint64(len(creatorContracts)),
		},
	}, nil
}

// ContractsByTag returns all contracts with a specific tag
func (qs queryServer) ContractsByTag(goCtx context.Context, req *pb.QueryContractsByTagRequest) (*pb.QueryContractsByTagResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var contracts []*pb.ContractInfo

	// Get all contracts with tag - simplified for now
	var tagContracts []*pb.ContractInfo
	qs.IterateContractInfo(ctx, func(info *pb.ContractInfo) bool {
		// Metadata is a value type, check if it has tags
		for _, tag := range info.Metadata.Tags {
			if tag == req.Tag {
				tagContracts = append(tagContracts, info)
				break
			}
		}
		return false
	})

	// Apply pagination if requested
	if req.Pagination != nil {
		start := int(req.Pagination.Offset)
		limit := int(req.Pagination.Limit)
		if limit == 0 {
			limit = 100
		}

		end := start + limit
		if end > len(tagContracts) {
			end = len(tagContracts)
		}

		if start < len(tagContracts) {
			for i := start; i < end; i++ {
				contracts = append(contracts, tagContracts[i])
			}
		}
	} else {
		for i := range tagContracts {
			contracts = append(contracts, tagContracts[i])
		}
	}

	return &pb.QueryContractsByTagResponse{
		Contracts: contracts,
		Pagination: &query.PageResponse{
			Total: uint64(len(tagContracts)),
		},
	}, nil
}

// RegisteredContracts returns all registered contracts with pagination
func (qs queryServer) RegisteredContracts(goCtx context.Context, req *pb.QueryRegisteredContractsRequest) (*pb.QueryRegisteredContractsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var contracts []*pb.ContractInfo

	store := ctx.KVStore(qs.storeKey)

	// Use manual iteration for prefix queries
	iterator := storetypes.KVStorePrefixIterator(store, types.ContractInfoPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info pb.ContractInfo
		if err := qs.cdc.Unmarshal(iterator.Value(), &info); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Apply status filter if specified
		if req.Status != pb.ContractStatus_CONTRACT_STATUS_UNSPECIFIED {
			if info.Status != req.Status {
				continue // Skip this contract, continue to next iteration
			}
		}

		contracts = append(contracts, &info)
	}

	// Simple pagination response
	return &pb.QueryRegisteredContractsResponse{
		Contracts: contracts,
		Pagination: &query.PageResponse{
			Total: uint64(len(contracts)),
		},
	}, nil
}

// ContractMetrics returns usage metrics for a contract
func (qs queryServer) ContractMetrics(goCtx context.Context, req *pb.QueryContractMetricsRequest) (*pb.QueryContractMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	metrics, found := qs.GetContractMetrics(ctx, req.ContractAddress)
	if !found {
		// Return zero metrics if not found
		metrics = &pb.ContractMetrics{
			ContractAddress: req.ContractAddress,
		}
	}

	return &pb.QueryContractMetricsResponse{
		Metrics: metrics,
	}, nil
}
