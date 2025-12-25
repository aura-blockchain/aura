// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

type queryServer struct {
	Keeper
	types.UnimplementedQueryServer
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params queries the module parameters
func (qs queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := qs.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// PeerInfo queries information about a specific peer
func (qs queryServer) PeerInfo(goCtx context.Context, req *types.QueryPeerInfoRequest) (*types.QueryPeerInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	peer, found := qs.GetPeerInfo(ctx, req.PeerId)
	if !found {
		return nil, types.ErrPeerNotFound
	}

	return &types.QueryPeerInfoResponse{Peer: peer}, nil
}

// AllPeers queries information about all connected peers with pagination
func (qs queryServer) AllPeers(goCtx context.Context, req *types.QueryAllPeersRequest) (*types.QueryAllPeersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for peer info
	peerStore := prefix.NewStore(store, types.PeerInfoPrefix)

	var peers []types.PeerInfo
	pageRes, err := query.Paginate(peerStore, req.Pagination, func(key []byte, value []byte) error {
		var peer types.PeerInfo
		if err := qs.Keeper.cdc.Unmarshal(value, &peer); err != nil {
			return fmt.Errorf("failed to unmarshal peer info: %w", err)
		}
		peers = append(peers, peer)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPeersResponse{
		Peers:      peers,
		Pagination: pageRes,
	}, nil
}

// TrustedPeers queries all trusted peers with pagination
func (qs queryServer) TrustedPeers(goCtx context.Context, req *types.QueryTrustedPeersRequest) (*types.QueryTrustedPeersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for trusted peers
	trustedPeerStore := prefix.NewStore(store, types.TrustedPeerPrefix)

	var peers []types.TrustedPeer
	pageRes, err := query.Paginate(trustedPeerStore, req.Pagination, func(key []byte, value []byte) error {
		var peer types.TrustedPeer
		if err := qs.Keeper.cdc.Unmarshal(value, &peer); err != nil {
			return fmt.Errorf("failed to unmarshal trusted peer: %w", err)
		}
		peers = append(peers, peer)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTrustedPeersResponse{
		Peers:      peers,
		Pagination: pageRes,
	}, nil
}

// PeerReputation queries reputation for a specific peer
func (qs queryServer) PeerReputation(goCtx context.Context, req *types.QueryPeerReputationRequest) (*types.QueryPeerReputationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	reputation, found := qs.GetReputation(ctx, req.PeerId)
	if !found {
		return nil, types.ErrPeerNotFound
	}

	return &types.QueryPeerReputationResponse{Reputation: reputation}, nil
}

// RateLimitStatus queries rate limit status for a peer
func (qs queryServer) RateLimitStatus(goCtx context.Context, req *types.QueryRateLimitStatusRequest) (*types.QueryRateLimitStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	status, found := qs.GetRateLimitEntry(ctx, req.PeerId)
	if !found {
		// Return empty status if not found
		status = types.RateLimitEntry{
			PeerId: req.PeerId,
		}
	}

	return &types.QueryRateLimitStatusResponse{Status: status}, nil
}

// MempoolStats queries current mempool statistics
func (qs queryServer) MempoolStats(goCtx context.Context, req *types.QueryMempoolStatsRequest) (*types.QueryMempoolStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	stats := qs.GetMempoolStats(ctx)

	return &types.QueryMempoolStatsResponse{Stats: stats}, nil
}

// ForkAlerts queries active fork alerts with pagination
func (qs queryServer) ForkAlerts(goCtx context.Context, req *types.QueryForkAlertsRequest) (*types.QueryForkAlertsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for fork alerts
	forkAlertStore := prefix.NewStore(store, types.ForkAlertPrefix)

	var alerts []types.ForkAlert
	pageRes, err := query.Paginate(forkAlertStore, req.Pagination, func(key []byte, value []byte) error {
		var alert types.ForkAlert
		if err := qs.Keeper.cdc.Unmarshal(value, &alert); err != nil {
			return fmt.Errorf("failed to unmarshal fork alert: %w", err)
		}
		// Filter by resolved status if requested
		if !req.IncludeResolved && alert.Resolved {
			return nil // Skip resolved alerts
		}
		alerts = append(alerts, alert)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryForkAlertsResponse{
		Alerts:     alerts,
		Pagination: pageRes,
	}, nil
}

// PartitionAlerts queries active partition alerts with pagination
func (qs queryServer) PartitionAlerts(goCtx context.Context, req *types.QueryPartitionAlertsRequest) (*types.QueryPartitionAlertsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(qs.Keeper.storeKey)

	// Create prefix store for partition alerts
	partitionAlertStore := prefix.NewStore(store, types.PartitionAlertPrefix)

	var alerts []types.PartitionAlert
	pageRes, err := query.Paginate(partitionAlertStore, req.Pagination, func(key []byte, value []byte) error {
		var alert types.PartitionAlert
		if err := qs.Keeper.cdc.Unmarshal(value, &alert); err != nil {
			return fmt.Errorf("failed to unmarshal partition alert: %w", err)
		}
		// Filter by resolved status if requested
		if !req.IncludeResolved && alert.Resolved {
			return nil // Skip resolved alerts
		}
		alerts = append(alerts, alert)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPartitionAlertsResponse{
		Alerts:     alerts,
		Pagination: pageRes,
	}, nil
}

// NetworkHealth queries overall network health status
func (qs queryServer) NetworkHealth(goCtx context.Context, req *types.QueryNetworkHealthRequest) (*types.QueryNetworkHealthResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all peers
	allPeers := qs.GetAllPeers(ctx)
	connectedPeers := uint32(len(allPeers))

	// Get healthy peers
	healthyPeers := qs.GetHealthyPeers(ctx)
	healthyPeerCount := uint32(len(healthyPeers))

	// Get banned peers
	bannedPeers := qs.GetBannedPeers(ctx)
	bannedPeerCount := uint32(len(bannedPeers))

	// Calculate average reputation
	avgReputation := qs.CalculateAverageReputation(ctx)

	// Check for active partition
	isPartitioned := false
	partitionAlerts := qs.GetAllPartitionAlerts(ctx, false)
	if len(partitionAlerts) > 0 {
		isPartitioned = true
	}

	// Check for active forks
	hasActiveForks := false
	forkAlerts := qs.GetAllForkAlerts(ctx, false)
	if len(forkAlerts) > 0 {
		hasActiveForks = true
	}

	// Get mempool stats
	mempoolStats := qs.GetMempoolStats(ctx)

	return &types.QueryNetworkHealthResponse{
		ConnectedPeers:    connectedPeers,
		HealthyPeers:      healthyPeerCount,
		BannedPeers:       bannedPeerCount,
		AverageReputation: avgReputation,
		IsPartitioned:     isPartitioned,
		HasActiveForks:    hasActiveForks,
		MempoolStats:      mempoolStats,
	}, nil
}
