package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

type queryServer struct {
	types.UnimplementedQueryServer
	Keeper
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

// AllPeers queries information about all connected peers
func (qs queryServer) AllPeers(goCtx context.Context, req *types.QueryAllPeersRequest) (*types.QueryAllPeersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	peers := qs.GetAllPeers(ctx)

	return &types.QueryAllPeersResponse{Peers: peers}, nil
}

// TrustedPeers queries all trusted peers
func (qs queryServer) TrustedPeers(goCtx context.Context, req *types.QueryTrustedPeersRequest) (*types.QueryTrustedPeersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	peers := qs.GetAllTrustedPeers(ctx)

	return &types.QueryTrustedPeersResponse{Peers: peers}, nil
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

// ForkAlerts queries active fork alerts
func (qs queryServer) ForkAlerts(goCtx context.Context, req *types.QueryForkAlertsRequest) (*types.QueryForkAlertsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	alerts := qs.GetAllForkAlerts(ctx, req.IncludeResolved)

	return &types.QueryForkAlertsResponse{Alerts: alerts}, nil
}

// PartitionAlerts queries active partition alerts
func (qs queryServer) PartitionAlerts(goCtx context.Context, req *types.QueryPartitionAlertsRequest) (*types.QueryPartitionAlertsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	alerts := qs.GetAllPartitionAlerts(ctx, req.IncludeResolved)

	return &types.QueryPartitionAlertsResponse{Alerts: alerts}, nil
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
