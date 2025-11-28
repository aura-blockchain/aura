package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

type msgServer struct {
	keeper *Keeper
	dexpb.UnimplementedMsgServer
}

// NewMsgServerImpl wires the keeper into the gRPC Msg service.
func NewMsgServerImpl(k *Keeper) dexpb.MsgServer {
	return &msgServer{keeper: k}
}

var _ dexpb.MsgServer = (*msgServer)(nil)

func (ms msgServer) CreatePool(goCtx context.Context, msg *dexpb.MsgCreatePool) (*dexpb.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amountA, err := protoCoinToSDK(msg.AmountA)
	if err != nil {
		return nil, err
	}
	amountB, err := protoCoinToSDK(msg.AmountB)
	if err != nil {
		return nil, err
	}

	pool, lpTokens, err := ms.keeper.SecureCreatePool(ctx, msg.Creator, msg.DenomA, msg.DenomB, amountA, amountB)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgCreatePoolResponse{
		PoolId:   pool.PoolId,
		LpTokens: lpTokens.String(),
	}, nil
}

func (ms msgServer) AddLiquidity(goCtx context.Context, msg *dexpb.MsgAddLiquidity) (*dexpb.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amountA, err := protoCoinToSDK(msg.AmountA)
	if err != nil {
		return nil, err
	}
	amountB, err := protoCoinToSDK(msg.AmountB)
	if err != nil {
		return nil, err
	}

	lpTokens, poolShare, err := ms.keeper.SecureAddLiquidity(ctx, msg.Provider, msg.PoolId, amountA, amountB)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgAddLiquidityResponse{
		LpTokensMinted:   lpTokens.String(),
		PoolSharePercent: poolShare.String(),
	}, nil
}

func (ms msgServer) RemoveLiquidity(goCtx context.Context, msg *dexpb.MsgRemoveLiquidity) (*dexpb.MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lpTokens, ok := sdkmath.NewIntFromString(msg.LpTokens)
	if !ok {
		return nil, fmt.Errorf("invalid lp token amount")
	}

	coinA, coinB, err := ms.keeper.SecureRemoveLiquidity(ctx, msg.Provider, msg.PoolId, lpTokens)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgRemoveLiquidityResponse{
		AmountA: convertCoinToProto(coinA),
		AmountB: convertCoinToProto(coinB),
	}, nil
}

func (ms msgServer) SwapExactIn(goCtx context.Context, msg *dexpb.MsgSwapExactIn) (*dexpb.MsgSwapExactInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coinIn, err := protoCoinToSDK(msg.CoinIn)
	if err != nil {
		return nil, err
	}
	minAmountOut, ok := sdkmath.NewIntFromString(msg.MinAmountOut)
	if !ok {
		return nil, fmt.Errorf("invalid min_amount_out")
	}

	amountOut, effectivePrice, priceImpact, err := ms.keeper.SecureSwapExactIn(ctx, msg.Sender, msg.PoolId, coinIn, minAmountOut, msg.MaxSlippageBps)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgSwapExactInResponse{
		AmountOut:          amountOut.String(),
		EffectivePrice:     effectivePrice.String(),
		PriceImpactPercent: priceImpact.String(),
	}, nil
}

func (ms msgServer) CreateOrder(goCtx context.Context, msg *dexpb.MsgCreateOrder) (*dexpb.MsgCreateOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	auraAmount, ok := sdkmath.NewIntFromString(msg.AuraAmount)
	if !ok {
		return nil, fmt.Errorf("invalid aura amount")
	}
	otherAmount, ok := sdkmath.NewIntFromString(msg.OtherAmount)
	if !ok {
		return nil, fmt.Errorf("invalid other coin amount")
	}

	order, err := ms.keeper.CreateOrder(ctx, msg.Creator, msg.OrderType, auraAmount, msg.OtherCoin, otherAmount, 1440)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgCreateOrderResponse{
		OrderId: order.OrderId,
		Status:  order.Status,
		Message: "order created",
	}, nil
}

func (ms msgServer) CancelOrder(goCtx context.Context, msg *dexpb.MsgCancelOrder) (*dexpb.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	order := ms.keeper.GetOrder(ctx, msg.OrderId)
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.UserAddress != msg.Creator {
		return nil, fmt.Errorf("cannot cancel order owned by another address")
	}

	if err := ms.keeper.CancelOrder(ctx, msg.OrderId, "user_cancelled"); err != nil {
		return nil, err
	}

	return &dexpb.MsgCancelOrderResponse{Success: true}, nil
}

func (ms msgServer) ExecuteSwap(goCtx context.Context, msg *dexpb.MsgExecuteSwap) (*dexpb.MsgExecuteSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.keeper.MatchOrder(ctx, msg.Initiator, msg.OrderId); err != nil {
		return nil, err
	}

	return &dexpb.MsgExecuteSwapResponse{
		Success: true,
		SwapId:  msg.OrderId,
	}, nil
}

func (ms msgServer) CreateHTLC(goCtx context.Context, msg *dexpb.MsgCreateHTLC) (*dexpb.MsgCreateHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amount, err := protoCoinToSDK(msg.Amount)
	if err != nil {
		return nil, err
	}

	htlcID, err := ms.keeper.CreateHTLC(ctx, msg.Sender, msg.Recipient, amount, msg.SecretHash, msg.TimelockDuration)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgCreateHTLCResponse{HtlcId: htlcID}, nil
}

func (ms msgServer) ClaimHTLC(goCtx context.Context, msg *dexpb.MsgClaimHTLC) (*dexpb.MsgClaimHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.keeper.ClaimHTLC(ctx, msg.Recipient, msg.HtlcId, msg.Secret); err != nil {
		return nil, err
	}

	return &dexpb.MsgClaimHTLCResponse{Success: true}, nil
}

func (ms msgServer) RefundHTLC(goCtx context.Context, msg *dexpb.MsgRefundHTLC) (*dexpb.MsgRefundHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.keeper.RefundHTLC(ctx, msg.Sender, msg.HtlcId); err != nil {
		return nil, err
	}

	return &dexpb.MsgRefundHTLCResponse{Success: true}, nil
}

func protoCoinToSDK(coin *sdk.Coin) (sdk.Coin, error) {
	if coin == nil {
		return sdk.Coin{}, fmt.Errorf("coin required")
	}
	if err := coin.Validate(); err != nil {
		return sdk.Coin{}, err
	}
	return *coin, nil
}

func convertCoinToProto(coin sdk.Coin) *sdk.Coin {
	return &sdk.Coin{
		Denom:  coin.Denom,
		Amount: coin.Amount,
	}
}
