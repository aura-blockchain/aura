package keeper

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

// verifySigner checks that the claimed address matches the transaction signer.
// This prevents unauthorized operations where an attacker could specify another
// user's address in the message without being the actual transaction signer.
//
// Security: This is critical for preventing authorization bypass attacks.
func verifySigner(msg sdk.Msg, claimedAddr string) error {
	// Type assert to get access to GetSigners method
	signerMsg, ok := msg.(interface{ GetSigners() []sdk.AccAddress })
	if !ok {
		return status.Error(codes.Internal, "message does not implement GetSigners")
	}

	signers := signerMsg.GetSigners()
	if len(signers) == 0 {
		return status.Error(codes.Unauthenticated, "no signers in transaction")
	}

	claimed, err := sdk.AccAddressFromBech32(claimedAddr)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid address format: %s", err.Error())
	}

	if !claimed.Equals(signers[0]) {
		return status.Errorf(codes.PermissionDenied,
			"creator/sender must be transaction signer (claimed: %s, signer: %s)",
			claimed.String(), signers[0].String())
	}

	return nil
}

func (ms msgServer) CreatePool(goCtx context.Context, msg *dexpb.MsgCreatePool) (*dexpb.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message creator is the transaction signer
	if err := verifySigner(msg, msg.Creator); err != nil {
		return nil, err
	}

	// Validate coins (they are value types, not pointers)
	if err := msg.AmountA.Validate(); err != nil {
		return nil, err
	}
	if err := msg.AmountB.Validate(); err != nil {
		return nil, err
	}

	pool, lpTokens, err := ms.keeper.SecureCreatePool(ctx, msg.Creator, msg.DenomA, msg.DenomB, msg.AmountA, msg.AmountB)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgCreatePoolResponse{
		PoolId:   pool.PoolId,
		LpTokens: lpTokens,
	}, nil
}

func (ms msgServer) AddLiquidity(goCtx context.Context, msg *dexpb.MsgAddLiquidity) (*dexpb.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message provider is the transaction signer
	if err := verifySigner(msg, msg.Provider); err != nil {
		return nil, err
	}

	// Validate coins (they are value types, not pointers)
	if err := msg.AmountA.Validate(); err != nil {
		return nil, err
	}
	if err := msg.AmountB.Validate(); err != nil {
		return nil, err
	}

	lpTokens, poolShare, err := ms.keeper.SecureAddLiquidity(ctx, msg.Provider, msg.PoolId, msg.AmountA, msg.AmountB)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgAddLiquidityResponse{
		LpTokensMinted:   lpTokens,
		PoolSharePercent: poolShare,
	}, nil
}

func (ms msgServer) RemoveLiquidity(goCtx context.Context, msg *dexpb.MsgRemoveLiquidity) (*dexpb.MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message provider is the transaction signer
	if err := verifySigner(msg, msg.Provider); err != nil {
		return nil, err
	}

	// LpTokens is already math.Int type (customtype in proto)
	coinA, coinB, err := ms.keeper.SecureRemoveLiquidity(ctx, msg.Provider, msg.PoolId, msg.LpTokens)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgRemoveLiquidityResponse{
		AmountA: coinA,
		AmountB: coinB,
	}, nil
}

func (ms msgServer) SwapExactIn(goCtx context.Context, msg *dexpb.MsgSwapExactIn) (*dexpb.MsgSwapExactInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message sender is the transaction signer
	if err := verifySigner(msg, msg.Sender); err != nil {
		return nil, err
	}

	// Validate coin (value type, not pointer)
	if err := msg.CoinIn.Validate(); err != nil {
		return nil, err
	}
	// MinAmountOut is already math.Int type (customtype in proto)

	amountOut, effectivePrice, priceImpact, err := ms.keeper.SecureSwapExactIn(ctx, msg.Sender, msg.PoolId, msg.CoinIn, msg.MinAmountOut, msg.MaxSlippageBps)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgSwapExactInResponse{
		AmountOut:          amountOut,
		EffectivePrice:     effectivePrice,
		PriceImpactPercent: priceImpact,
	}, nil
}

func (ms msgServer) CreateOrder(goCtx context.Context, msg *dexpb.MsgCreateOrder) (*dexpb.MsgCreateOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message creator is the transaction signer
	if err := verifySigner(msg, msg.Creator); err != nil {
		return nil, err
	}

	// AuraAmount and OtherAmount are already math.Int types (customtype in proto)
	order, err := ms.keeper.CreateOrder(ctx, msg.Creator, msg.OrderType, msg.AuraAmount, msg.OtherCoin, msg.OtherAmount, 1440)
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

	// CRITICAL SECURITY: Verify that the message creator is the transaction signer
	// This prevents an attacker from cancelling another user's order by simply
	// specifying their address in msg.Creator without being the actual signer.
	if err := verifySigner(msg, msg.Creator); err != nil {
		return nil, err
	}

	order := ms.keeper.GetOrder(ctx, msg.OrderId)
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	// Verify that the authenticated signer owns the order
	if order.UserAddress != msg.Creator {
		return nil, status.Error(codes.PermissionDenied, "cannot cancel order owned by another address")
	}

	if err := ms.keeper.CancelOrder(ctx, msg.OrderId, "user_cancelled"); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %s", err.Error())
	}

	return &dexpb.MsgCancelOrderResponse{Success: true}, nil
}

func (ms msgServer) ExecuteSwap(goCtx context.Context, msg *dexpb.MsgExecuteSwap) (*dexpb.MsgExecuteSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message initiator is the transaction signer
	if err := verifySigner(msg, msg.Initiator); err != nil {
		return nil, err
	}

	if err := ms.keeper.MatchOrder(ctx, msg.Initiator, msg.OrderId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute swap: %s", err.Error())
	}

	return &dexpb.MsgExecuteSwapResponse{
		Success: true,
		SwapId:  msg.OrderId,
	}, nil
}

func (ms msgServer) CreateHTLC(goCtx context.Context, msg *dexpb.MsgCreateHTLC) (*dexpb.MsgCreateHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message sender is the transaction signer
	if err := verifySigner(msg, msg.Sender); err != nil {
		return nil, err
	}

	// Validate coin (value type, not pointer)
	if err := msg.Amount.Validate(); err != nil {
		return nil, err
	}

	htlcID, err := ms.keeper.CreateHTLC(ctx, msg.Sender, msg.Recipient, msg.Amount, msg.SecretHash, msg.TimelockDuration)
	if err != nil {
		return nil, err
	}

	return &dexpb.MsgCreateHTLCResponse{HtlcId: htlcID}, nil
}

func (ms msgServer) ClaimHTLC(goCtx context.Context, msg *dexpb.MsgClaimHTLC) (*dexpb.MsgClaimHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message recipient is the transaction signer
	if err := verifySigner(msg, msg.Recipient); err != nil {
		return nil, err
	}

	if err := ms.keeper.ClaimHTLC(ctx, msg.Recipient, msg.HtlcId, msg.Secret); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to claim HTLC: %s", err.Error())
	}

	return &dexpb.MsgClaimHTLCResponse{Success: true}, nil
}

func (ms msgServer) RefundHTLC(goCtx context.Context, msg *dexpb.MsgRefundHTLC) (*dexpb.MsgRefundHTLCResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message sender is the transaction signer
	if err := verifySigner(msg, msg.Sender); err != nil {
		return nil, err
	}

	if err := ms.keeper.RefundHTLC(ctx, msg.Sender, msg.HtlcId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refund HTLC: %s", err.Error())
	}

	return &dexpb.MsgRefundHTLCResponse{Success: true}, nil
}

func (ms msgServer) CommitOrder(goCtx context.Context, msg *dexpb.MsgCommitOrder) (*dexpb.MsgCommitOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message sender is the transaction signer
	if err := verifySigner(msg, msg.Sender); err != nil {
		return nil, err
	}

	commitID, err := ms.keeper.CommitOrder(ctx, msg.Sender, msg.CommitHash)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit order: %s", err.Error())
	}

	// Get the commitment to return reveal deadline
	commitment, found := ms.keeper.GetOrderCommitment(ctx, commitID)
	if !found {
		return nil, status.Error(codes.Internal, "commitment not found after creation")
	}

	return &dexpb.MsgCommitOrderResponse{
		CommitId:       commitID,
		RevealDeadline: commitment.RevealDeadline.Format(time.RFC3339),
	}, nil
}

func (ms msgServer) RevealOrder(goCtx context.Context, msg *dexpb.MsgRevealOrder) (*dexpb.MsgRevealOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify that the message sender is the transaction signer
	if err := verifySigner(msg, msg.Sender); err != nil {
		return nil, err
	}

	// AuraAmount and OtherAmount are already math.Int types (customtype in proto)
	orderID, err := ms.keeper.RevealOrder(
		ctx,
		msg.CommitId,
		msg.Sender,
		msg.OrderType,
		msg.AuraAmount,
		msg.OtherCoin,
		msg.OtherAmount,
		msg.Salt,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reveal order: %s", err.Error())
	}

	// Determine message based on batch execution
	params := ms.keeper.GetParams(ctx)
	message := "executed immediately"
	if params.BatchExecutionEnabled {
		message = "queued for batch execution"
	}

	return &dexpb.MsgRevealOrderResponse{
		Success: true,
		OrderId: orderID,
		Message: message,
	}, nil
}
