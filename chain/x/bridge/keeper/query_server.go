package keeper

import (
	"context"

	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ bridgetypes.QueryServer = queryServer{}

// queryServer implements the QueryServer interface
type queryServer struct {
	types.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) bridgetypes.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Transfer queries a cross-chain transfer by ID
func (qs queryServer) Transfer(goCtx context.Context, req *bridgetypes.QueryTransferRequest) (*bridgetypes.QueryTransferResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	transfer, err := qs.Keeper.GetTransfer(ctx, req.TransferId)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.QueryTransferResponse{Transfer: *transfer}, nil
}

// AllTransfers queries all cross-chain transfers
func (qs queryServer) AllTransfers(goCtx context.Context, req *bridgetypes.QueryAllTransfersRequest) (*bridgetypes.QueryAllTransfersResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	transfers := qs.Keeper.GetAllTransfers(ctx, req.Status)

	return &bridgetypes.QueryAllTransfersResponse{Transfers: transfers}, nil
}

// UserTransfers queries transfers for specific user
func (qs queryServer) UserTransfers(goCtx context.Context, req *bridgetypes.QueryUserTransfersRequest) (*bridgetypes.QueryUserTransfersResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	transfers := qs.Keeper.GetUserTransfers(ctx, req.Address, req.Chain)

	return &bridgetypes.QueryUserTransfersResponse{Transfers: transfers}, nil
}

// ChainConfig queries configuration for a chain
func (qs queryServer) ChainConfig(goCtx context.Context, req *bridgetypes.QueryChainConfigRequest) (*bridgetypes.QueryChainConfigResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config, err := qs.Keeper.GetChainConfig(ctx, req.ChainId)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.QueryChainConfigResponse{Config: *config}, nil
}

// AllChains queries all connected chains
func (qs queryServer) AllChains(goCtx context.Context, req *bridgetypes.QueryAllChainsRequest) (*bridgetypes.QueryAllChainsResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	chains := qs.Keeper.GetAllChainConfigs(ctx)

	return &bridgetypes.QueryAllChainsResponse{Chains: chains}, nil
}

// WrappedToken queries wrapped token info
func (qs queryServer) WrappedToken(goCtx context.Context, req *bridgetypes.QueryWrappedTokenRequest) (*bridgetypes.QueryWrappedTokenResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	token, err := qs.Keeper.GetWrappedToken(ctx, req.WrappedDenom)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.QueryWrappedTokenResponse{Token: *token}, nil
}

// AllWrappedTokens queries all wrapped tokens
func (qs queryServer) AllWrappedTokens(goCtx context.Context, req *bridgetypes.QueryAllWrappedTokensRequest) (*bridgetypes.QueryAllWrappedTokensResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	tokens := qs.Keeper.GetAllWrappedTokens(ctx)

	return &bridgetypes.QueryAllWrappedTokensResponse{Tokens: tokens}, nil
}

// SharedIdentity queries shared identity across chains
func (qs queryServer) SharedIdentity(goCtx context.Context, req *bridgetypes.QuerySharedIdentityRequest) (*bridgetypes.QuerySharedIdentityResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	identity, err := qs.Keeper.GetSharedIdentity(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.QuerySharedIdentityResponse{Identity: *identity}, nil
}

// CrossChainSwap queries cross-chain swap status
func (qs queryServer) CrossChainSwap(goCtx context.Context, req *bridgetypes.QueryCrossChainSwapRequest) (*bridgetypes.QueryCrossChainSwapResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	swap, err := qs.Keeper.GetCrossChainSwap(ctx, req.SwapId)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.QueryCrossChainSwapResponse{Swap: *swap}, nil
}

// BridgeStats queries bridge statistics
func (qs queryServer) BridgeStats(goCtx context.Context, req *bridgetypes.QueryBridgeStatsRequest) (*bridgetypes.QueryBridgeStatsResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stats := qs.Keeper.GetBridgeStats(ctx)

	return &bridgetypes.QueryBridgeStatsResponse{
		TotalTransfers:     stats.TotalTransfers,
		TransfersByStatus:  stats.TransfersByStatus,
		VolumeByChain:      stats.VolumeByChain,
		TotalWrappedTokens: stats.TotalWrappedTokens,
		ActiveValidators:   stats.ActiveValidators,
		ActiveRelayers:     stats.ActiveRelayers,
		AvgCompletionTime:  stats.AvgCompletionTime,
	}, nil
}

// Validators queries active bridge validators
func (qs queryServer) Validators(goCtx context.Context, req *bridgetypes.QueryValidatorsRequest) (*bridgetypes.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	validators := qs.Keeper.GetBridgeValidators(ctx)

	return &bridgetypes.QueryValidatorsResponse{Validators: validators}, nil
}

// RelayerStats queries relayer performance stats
func (qs queryServer) RelayerStats(goCtx context.Context, req *bridgetypes.QueryRelayerStatsRequest) (*bridgetypes.QueryRelayerStatsResponse, error) {
	if req == nil {
		return nil, bridgetypes.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stats, err := qs.Keeper.GetRelayerStats(ctx, req.RelayerAddress)
	if err != nil {
		return nil, err
	}

	return &bridgetypes.QueryRelayerStatsResponse{Stats: *stats}, nil
}
