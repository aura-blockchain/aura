package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates the module parameters
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Validate and set params
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// AddTrustedPeer adds a trusted peer
func (ms msgServer) AddTrustedPeer(goCtx context.Context, msg *types.MsgAddTrustedPeer) (*types.MsgAddTrustedPeerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Check if peer already exists
	if ms.IsTrustedPeer(ctx, msg.Peer.PeerId) {
		return nil, types.ErrTrustedPeerExists
	}

	// Validate peer data
	if msg.Peer.PeerId == "" {
		return nil, types.ErrInvalidPeerID
	}
	if msg.Peer.Address == "" {
		return nil, types.ErrInvalidIPAddress
	}

	// Set added timestamp
	msg.Peer.AddedAt = ctx.BlockTime()

	// Add trusted peer
	if err := ms.SetTrustedPeer(ctx, msg.Peer); err != nil {
		return nil, err
	}

	return &types.MsgAddTrustedPeerResponse{}, nil
}

// RemoveTrustedPeer removes a trusted peer
func (ms msgServer) RemoveTrustedPeer(goCtx context.Context, msg *types.MsgRemoveTrustedPeer) (*types.MsgRemoveTrustedPeerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Check if peer exists
	if !ms.IsTrustedPeer(ctx, msg.PeerId) {
		return nil, types.ErrPeerNotFound
	}

	// Remove trusted peer
	if err := ms.RemoveTrustedPeer(ctx, msg.PeerId); err != nil {
		return nil, err
	}

	return &types.MsgRemoveTrustedPeerResponse{}, nil
}

// BanPeer manually bans a peer
func (ms msgServer) BanPeer(goCtx context.Context, msg *types.MsgBanPeer) (*types.MsgBanPeerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Ban the peer
	if err := ms.Keeper.BanPeer(ctx, msg.PeerId, msg.DurationSeconds, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgBanPeerResponse{}, nil
}

// UnbanPeer unbans a peer
func (ms msgServer) UnbanPeer(goCtx context.Context, msg *types.MsgUnbanPeer) (*types.MsgUnbanPeerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Unban the peer
	if err := ms.Keeper.UnbanPeer(ctx, msg.PeerId); err != nil {
		return nil, err
	}

	return &types.MsgUnbanPeerResponse{}, nil
}

// UpdatePeerReputation manually updates peer reputation
func (ms msgServer) UpdatePeerReputation(goCtx context.Context, msg *types.MsgUpdatePeerReputation) (*types.MsgUpdatePeerReputationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Update reputation
	if err := ms.Keeper.UpdateReputation(ctx, msg.PeerId, msg.Score, msg.Reason); err != nil {
		return nil, err
	}

	return &types.MsgUpdatePeerReputationResponse{}, nil
}

// ResolveForkAlert marks a fork alert as resolved
func (ms msgServer) ResolveForkAlert(goCtx context.Context, msg *types.MsgResolveForkAlert) (*types.MsgResolveForkAlertResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Get alert
	alert, found := ms.GetForkAlert(ctx, msg.AlertId)
	if !found {
		return nil, types.ErrAlertNotFound
	}

	if alert.Resolved {
		return nil, types.ErrAlreadyResolved
	}

	// Mark as resolved
	alert.Resolved = true
	alert.ResolutionDetails = msg.ResolutionDetails

	if err := ms.SetForkAlert(ctx, alert); err != nil {
		return nil, err
	}

	return &types.MsgResolveForkAlertResponse{}, nil
}

// ResolvePartitionAlert marks a partition alert as resolved
func (ms msgServer) ResolvePartitionAlert(goCtx context.Context, msg *types.MsgResolvePartitionAlert) (*types.MsgResolvePartitionAlertResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority
	if ms.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidPeerID
	}

	// Resolve alert
	if err := ms.Keeper.ResolvePartitionAlert(ctx, msg.AlertId); err != nil {
		return nil, err
	}

	return &types.MsgResolvePartitionAlertResponse{}, nil
}
