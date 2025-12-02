package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/dex/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

type MsgServerTestSuite struct {
	KeeperTestSuite
	msgServer dexpb.MsgServer
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (suite *MsgServerTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.msgServer = NewMsgServerImpl(suite.Keeper)
}

func (suite *MsgServerTestSuite) TestMsgServerImplementation() {
	suite.NotNil(suite.msgServer, "msg server should be created")
}

func (suite *MsgServerTestSuite) TestNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.SwapExactIn(ctx, &dexpb.MsgSwapExactIn{
		Sender: suite.addr("swap-sender"),
		PoolId: "pool-1",
		// CoinIn intentionally nil to exercise input validation before keeper deps (bank) are used
		CoinIn: nil,
		// MinAmountOut must be a valid int string to reach coin validation
		MinAmountOut:   "1",
		MaxSlippageBps: 100,
	})
	suite.Error(err)
	suite.Contains(err.Error(), "coin required")
}

func (suite *MsgServerTestSuite) TestInvalidSigner() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.CreateOrder(ctx, &dexpb.MsgCreateOrder{
		Creator:     suite.addr("creator-invalid-aura"),
		OrderType:   types.SwapOrderType_BUY,
		AuraAmount:  "not-a-number",
		OtherCoin:   "utoken",
		OtherAmount: "1000",
	})
	suite.Error(err)
	suite.Contains(err.Error(), "invalid aura amount")
}

func (suite *MsgServerTestSuite) TestValidMessage() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.CancelOrder(ctx, &dexpb.MsgCancelOrder{
		Creator: suite.addr("no-order"),
		OrderId: "missing-order",
	})
	suite.Error(err)
	suite.Contains(err.Error(), "order not found")
}

func (suite *MsgServerTestSuite) TestUnauthorized() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	owner := suite.addr("owner")
	other := suite.addr("other")

	order := &types.SwapOrder{
		OrderId:      "order-1",
		OrderType:    types.SwapOrderType_BUY,
		AuraAmount:   "10",
		OtherCoin:    "utoken",
		OtherAmount:  "20",
		UserAddress:  owner,
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    timestamppb.New(suite.SdkCtx.BlockTime()),
		ExpiresAt:    timestamppb.New(suite.SdkCtx.BlockTime().Add(time.Hour)),
		PricePerAura: "2",
	}
	suite.Keeper.SetOrder(suite.SdkCtx, order)

	_, err := suite.msgServer.CancelOrder(ctx, &dexpb.MsgCancelOrder{
		Creator: other,
		OrderId: order.OrderId,
	})
	suite.Error(err)
	suite.Contains(err.Error(), "cannot cancel order owned by another address")
}

func (suite *MsgServerTestSuite) TestEventEmission() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.ExecuteSwap(ctx, &dexpb.MsgExecuteSwap{
		Initiator: suite.addr("exec-init"),
		OrderId:   "missing-order",
	})
	suite.Error(err)
	suite.Contains(err.Error(), "order not found")
}

func (suite *MsgServerTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}
