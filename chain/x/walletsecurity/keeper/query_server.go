package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	walletsecproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ walletsecproto.QueryServer = queryServer{}

type queryServer struct {
	twalletsecproto.UnimplementedQueryServer
	Keeper *Keeper
}

func NewQueryServerImpl(keeper *Keeper) walletsecproto.QueryServer {
	return &queryServer{Keeper: keeper}
}

func (qs queryServer) GetHardwareWallet(goCtx context.Context, req *walletsecproto.QueryGetHardwareWalletRequest) (*walletsecproto.QueryGetHardwareWalletResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config, err := qs.Keeper.GetHardwareWalletConfig(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetHardwareWalletResponse{Config: config}, nil
}

func (qs queryServer) GetMultiSigWallet(goCtx context.Context, req *walletsecproto.QueryGetMultiSigWalletRequest) (*walletsecproto.QueryGetMultiSigWalletResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	wallet, err := qs.Keeper.GetMultiSigWallet(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetMultiSigWalletResponse{Wallet: wallet}, nil
}

func (qs queryServer) GetPendingMultiSigTx(goCtx context.Context, req *walletsecproto.QueryGetPendingMultiSigTxRequest) (*walletsecproto.QueryGetPendingMultiSigTxResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	tx, err := qs.Keeper.GetPendingMultiSigTx(ctx, req.TxId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetPendingMultiSigTxResponse{Tx: tx}, nil
}

func (qs queryServer) GetSocialRecoveryConfig(goCtx context.Context, req *walletsecproto.QueryGetSocialRecoveryConfigRequest) (*walletsecproto.QueryGetSocialRecoveryConfigResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config, err := qs.Keeper.GetSocialRecoveryConfig(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetSocialRecoveryConfigResponse{Config: config}, nil
}

func (qs queryServer) GetRecoveryRequest(goCtx context.Context, req *walletsecproto.QueryGetRecoveryRequestRequest) (*walletsecproto.QueryGetRecoveryRequestResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	request, err := qs.Keeper.GetRecoveryRequest(ctx, req.RequestId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetRecoveryRequestResponse{Request: request}, nil
}

func (qs queryServer) GetSpendingLimit(goCtx context.Context, req *walletsecproto.QueryGetSpendingLimitRequest) (*walletsecproto.QueryGetSpendingLimitResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	limit, err := qs.Keeper.GetSpendingLimit(ctx, req.WalletId, req.Denom)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetSpendingLimitResponse{Limit: limit}, nil
}

func (qs queryServer) GetSessionConfig(goCtx context.Context, req *walletsecproto.QueryGetSessionConfigRequest) (*walletsecproto.QueryGetSessionConfigResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config, err := qs.Keeper.GetSessionConfig(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetSessionConfigResponse{Config: config}, nil
}

func (qs queryServer) GetSecurityMetrics(goCtx context.Context, req *walletsecproto.QueryGetSecurityMetricsRequest) (*walletsecproto.QueryGetSecurityMetricsResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	metrics, err := qs.Keeper.GetSecurityMetrics(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetSecurityMetricsResponse{Metrics: metrics}, nil
}

func (qs queryServer) GetDomainVerification(goCtx context.Context, req *walletsecproto.QueryGetDomainVerificationRequest) (*walletsecproto.QueryGetDomainVerificationResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	verification, err := qs.Keeper.GetDomainVerification(ctx, req.Domain)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetDomainVerificationResponse{Verification: verification}, nil
}

func (qs queryServer) GetDustFilter(goCtx context.Context, req *walletsecproto.QueryGetDustFilterRequest) (*walletsecproto.QueryGetDustFilterResponse, error) {
	if req == nil {
		return nil, types.ErrInvalidInput.Wrap("empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	filter, err := qs.Keeper.GetDustFilter(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.QueryGetDustFilterResponse{Filter: filter}, nil
}
