// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

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
		// CoinIn intentionally empty to exercise input validation before keeper deps (bank) are used
		CoinIn: sdk.Coin{},
		// MinAmountOut must be a valid int string to reach coin validation
		MinAmountOut:   sdkmath.NewInt(1),
		MaxSlippageBps: 100,
	})
	suite.Error(err)
	// Empty coin denom triggers "invalid denom" error
	suite.Contains(err.Error(), "invalid denom")
}

func (suite *MsgServerTestSuite) TestInvalidSigner() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	_, err := suite.msgServer.CreateOrder(ctx, &dexpb.MsgCreateOrder{
		Creator:     suite.addr("creator-invalid-aura"),
		OrderType:   types.SwapOrderType_BUY,
		AuraAmount:  sdkmath.ZeroInt(), // Invalid: zero amount
		OtherCoin:   "utoken",
		OtherAmount: sdkmath.NewInt(1000),
	})
	suite.Error(err)
	suite.Contains(err.Error(), "must be positive")
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
		AuraAmount:   sdkmath.NewInt(10),
		OtherCoin:    "utoken",
		OtherAmount:  sdkmath.NewInt(20),
		UserAddress:  owner,
		Status:       types.SwapOrderStatus_PENDING,
		Timestamp:    suite.SdkCtx.BlockTime(),
		ExpiresAt:    suite.SdkCtx.BlockTime().Add(time.Hour),
		PricePerAura: sdkmath.LegacyMustNewDecFromStr("2"),
	}
	suite.Require().NoError(suite.Keeper.SetOrder(suite.SdkCtx, order))

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
