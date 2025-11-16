package keeper

import (
	"context"
	"fmt"

	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ dextypes.MsgServer = msgServer{}

// msgServer implements the MsgServer interface
type msgServer struct {
	types.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) dextypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

// CreatePool creates a new AMM liquidity pool
func (ms msgServer) CreatePool(goCtx context.Context, msg *dextypes.MsgCreatePool) (*dextypes.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Creator == "" || msg.DenomA == "" || msg.DenomB == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("creator and denoms are required")
	}

	if !msg.AmountA.IsValid() || !msg.AmountB.IsValid() || msg.AmountA.IsZero() || msg.AmountB.IsZero() {
		return nil, dextypes.ErrInvalidAmount.Wrap("invalid or zero amounts")
	}

	// Create pool
	poolID, lpTokens, err := ms.Keeper.CreateLiquidityPool(ctx, msg.Creator, msg.DenomA, msg.DenomB, msg.AmountA, msg.AmountB)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgCreatePoolResponse{
		PoolId:   poolID,
		LpTokens: lpTokens,
	}, nil
}

// AddLiquidity adds liquidity to existing pool
func (ms msgServer) AddLiquidity(goCtx context.Context, msg *dextypes.MsgAddLiquidity) (*dextypes.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Provider == "" || msg.PoolId == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("provider and pool_id are required")
	}

	if !msg.AmountA.IsValid() || !msg.AmountB.IsValid() || msg.AmountA.IsZero() || msg.AmountB.IsZero() {
		return nil, dextypes.ErrInvalidAmount.Wrap("invalid or zero amounts")
	}

	// Add liquidity
	lpTokensMinted, poolSharePercent, err := ms.Keeper.AddLiquidity(ctx, msg.Provider, msg.PoolId, msg.AmountA, msg.AmountB)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgAddLiquidityResponse{
		LpTokensMinted:   lpTokensMinted,
		PoolSharePercent: poolSharePercent,
	}, nil
}

// RemoveLiquidity removes liquidity from pool
func (ms msgServer) RemoveLiquidity(goCtx context.Context, msg *dextypes.MsgRemoveLiquidity) (*dextypes.MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Provider == "" || msg.PoolId == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("provider and pool_id are required")
	}

	// Remove liquidity
	amountA, amountB, err := ms.Keeper.RemoveLiquidity(ctx, msg.Provider, msg.PoolId, msg.LpTokens)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgRemoveLiquidityResponse{
		AmountA: amountA,
		AmountB: amountB,
	}, nil
}

// SwapExactIn swaps exact input amount for minimum output
func (ms msgServer) SwapExactIn(goCtx context.Context, msg *dextypes.MsgSwapExactIn) (*dextypes.MsgSwapExactInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Sender == "" || msg.PoolId == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("sender and pool_id are required")
	}

	if !msg.CoinIn.IsValid() || msg.CoinIn.IsZero() {
		return nil, dextypes.ErrInvalidAmount.Wrap("invalid or zero coin_in")
	}

	if msg.MaxSlippageBps > 10000 {
		return nil, dextypes.ErrInvalidInput.Wrap("max_slippage_bps cannot exceed 10000 (100%)")
	}

	// Execute swap
	amountOut, effectivePrice, priceImpact, err := ms.Keeper.SwapExactIn(ctx, msg.Sender, msg.PoolId, msg.CoinIn, msg.MinAmountOut, msg.MaxSlippageBps)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgSwapExactInResponse{
		AmountOut:          amountOut,
		EffectivePrice:     effectivePrice,
		PriceImpactPercent: priceImpact,
	}, nil
}

// CreateOrder creates P2P swap order
func (ms msgServer) CreateOrder(goCtx context.Context, msg *dextypes.MsgCreateOrder) (*dextypes.MsgCreateOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Creator == "" || msg.OtherCoin == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("creator and other_coin are required")
	}

	// Create order
	orderID, status, matchedWith, message, err := ms.Keeper.CreateSwapOrder(ctx, msg.Creator, msg.OrderType, msg.AuraAmount, msg.OtherCoin, msg.OtherAmount)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgCreateOrderResponse{
		OrderId:     orderID,
		Status:      status,
		MatchedWith: matchedWith,
		Message:     message,
	}, nil
}

// CancelOrder cancels pending order
func (ms msgServer) CancelOrder(goCtx context.Context, msg *dextypes.MsgCancelOrder) (*dextypes.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Creator == "" || msg.OrderId == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("creator and order_id are required")
	}

	// Cancel order
	err := ms.Keeper.CancelSwapOrder(ctx, msg.Creator, msg.OrderId)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgCancelOrderResponse{Success: true}, nil
}

// ExecuteSwap executes a matched swap order
func (ms msgServer) ExecuteSwap(goCtx context.Context, msg *dextypes.MsgExecuteSwap) (*dextypes.MsgExecuteSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Executor == "" || msg.OrderId == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("executor and order_id are required")
	}

	// Execute swap
	err := ms.Keeper.ExecuteSwapOrder(ctx, msg.Executor, msg.OrderId)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgExecuteSwapResponse{Success: true}, nil
}

// CreateHTLC creates an HTLC atomic swap
func (ms msgServer) CreateHTLC(goCtx context.Context, msg *dextypes.MsgCreateHTLC) (*dextypes.MsgCreateHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Sender == "" || msg.Recipient == "" || msg.HashLock == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("sender, recipient, and hash_lock are required")
	}

	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return nil, dextypes.ErrInvalidAmount.Wrap("invalid or zero amount")
	}

	if msg.TimeLock == 0 {
		return nil, dextypes.ErrInvalidInput.Wrap("time_lock must be greater than 0")
	}

	// Create HTLC
	htlcID, err := ms.Keeper.CreateHTLC(ctx, msg.Sender, msg.Recipient, msg.Amount, msg.HashLock, msg.TimeLock)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgCreateHTLCResponse{HtlcId: htlcID}, nil
}

// ClaimHTLC claims an HTLC with secret
func (ms msgServer) ClaimHTLC(goCtx context.Context, msg *dextypes.MsgClaimHTLC) (*dextypes.MsgClaimHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Recipient == "" || msg.HtlcId == "" || msg.Secret == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("recipient, htlc_id, and secret are required")
	}

	// Claim HTLC
	err := ms.Keeper.ClaimHTLC(ctx, msg.Recipient, msg.HtlcId, msg.Secret)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgClaimHTLCResponse{Success: true}, nil
}

// RefundHTLC refunds an expired HTLC
func (ms msgServer) RefundHTLC(goCtx context.Context, msg *dextypes.MsgRefundHTLC) (*dextypes.MsgRefundHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate inputs
	if msg.Sender == "" || msg.HtlcId == "" {
		return nil, dextypes.ErrInvalidInput.Wrap("sender and htlc_id are required")
	}

	// Refund HTLC
	err := ms.Keeper.RefundHTLC(ctx, msg.Sender, msg.HtlcId)
	if err != nil {
		return nil, err
	}

	return &dextypes.MsgRefundHTLCResponse{Success: true}, nil
}
