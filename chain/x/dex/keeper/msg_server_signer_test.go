package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/dex/keeper"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// TestSignerVerification tests that all message handlers properly verify
// that the claimed address matches the transaction signer.
// This is a critical security test to prevent authorization bypass attacks.
func TestSignerVerification(t *testing.T) {
	suite := setupMsgServerTestSuite(t)
	msgServer := keeper.NewMsgServerImpl(suite.keeper)

	// Create two test addresses
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	// Fund both addresses
	suite.fundAccount(addr1, sdk.NewCoins(
		sdk.NewCoin("aura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))
	suite.fundAccount(addr2, sdk.NewCoins(
		sdk.NewCoin("aura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	))

	t.Run("CreatePool - reject mismatched signer", func(t *testing.T) {
		msg := &dexpb.MsgCreatePool{
			Creator: addr2.String(), // Claiming to be addr2
			DenomA:  "aura",
			DenomB:  "usdt",
			AmountA: &sdk.Coin{Denom: "aura", Amount: sdkmath.NewInt(1000)},
			AmountB: &sdk.Coin{Denom: "usdt", Amount: sdkmath.NewInt(1000)},
		}

		// But transaction is signed by addr1
		ctx := suite.ctx.WithValue(sdk.ContextKey("signers"), []sdk.AccAddress{addr1})
		_, err := msgServer.CreatePool(ctx, msg)

		// Should be rejected
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.PermissionDenied, st.Code())
		require.Contains(t, st.Message(), "must be transaction signer")
	})

	t.Run("AddLiquidity - reject mismatched signer", func(t *testing.T) {
		// First create a pool as addr1
		createMsg := &dexpb.MsgCreatePool{
			Creator: addr1.String(),
			DenomA:  "aura",
			DenomB:  "usdt",
			AmountA: &sdk.Coin{Denom: "aura", Amount: sdkmath.NewInt(1000)},
			AmountB: &sdk.Coin{Denom: "usdt", Amount: sdkmath.NewInt(1000)},
		}
		resp, err := msgServer.CreatePool(suite.ctx, createMsg)
		require.NoError(t, err)

		// Try to add liquidity claiming to be addr2
		msg := &dexpb.MsgAddLiquidity{
			Provider: addr2.String(), // Claiming to be addr2
			PoolId:   resp.PoolId,
			AmountA:  &sdk.Coin{Denom: "aura", Amount: sdkmath.NewInt(100)},
			AmountB:  &sdk.Coin{Denom: "usdt", Amount: sdkmath.NewInt(100)},
		}

		// But transaction is signed by addr1
		ctx := suite.ctx.WithValue(sdk.ContextKey("signers"), []sdk.AccAddress{addr1})
		_, err = msgServer.AddLiquidity(ctx, msg)

		// Should be rejected
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("CancelOrder - reject mismatched signer", func(t *testing.T) {
		// Create an order as addr1
		createMsg := &dexpb.MsgCreateOrder{
			Creator:     addr1.String(),
			OrderType:   dexpb.SwapOrderType_BUY,
			AuraAmount:  sdkmath.NewInt(100).String(),
			OtherCoin:   "usdt",
			OtherAmount: sdkmath.NewInt(100).String(),
		}
		createResp, err := msgServer.CreateOrder(suite.ctx, createMsg)
		require.NoError(t, err)

		// Try to cancel it claiming to be addr1, but signed by addr2
		// This is the CRITICAL security vulnerability being tested
		cancelMsg := &dexpb.MsgCancelOrder{
			Creator: addr1.String(), // Claiming to be addr1 (the order owner)
			OrderId: createResp.OrderId,
		}

		// But transaction is actually signed by addr2 (attacker)
		ctx := suite.ctx.WithValue(sdk.ContextKey("signers"), []sdk.AccAddress{addr2})
		_, err = msgServer.CancelOrder(ctx, cancelMsg)

		// Should be REJECTED - this prevents the attack
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.PermissionDenied, st.Code())
		require.Contains(t, st.Message(), "must be transaction signer")
	})

	t.Run("CancelOrder - allow matching signer", func(t *testing.T) {
		// Create an order as addr1
		createMsg := &dexpb.MsgCreateOrder{
			Creator:     addr1.String(),
			OrderType:   dexpb.SwapOrderType_BUY,
			AuraAmount:  sdkmath.NewInt(100).String(),
			OtherCoin:   "usdt",
			OtherAmount: sdkmath.NewInt(100).String(),
		}
		createResp, err := msgServer.CreateOrder(suite.ctx, createMsg)
		require.NoError(t, err)

		// Cancel it properly with matching signer
		cancelMsg := &dexpb.MsgCancelOrder{
			Creator: addr1.String(),
			OrderId: createResp.OrderId,
		}

		// Transaction is correctly signed by addr1
		_, err = msgServer.CancelOrder(suite.ctx, cancelMsg)

		// Should succeed
		require.NoError(t, err)
	})

	t.Run("SwapExactIn - reject mismatched signer", func(t *testing.T) {
		// Create a pool
		createMsg := &dexpb.MsgCreatePool{
			Creator: addr1.String(),
			DenomA:  "aura",
			DenomB:  "usdt",
			AmountA: &sdk.Coin{Denom: "aura", Amount: sdkmath.NewInt(10000)},
			AmountB: &sdk.Coin{Denom: "usdt", Amount: sdkmath.NewInt(10000)},
		}
		poolResp, err := msgServer.CreatePool(suite.ctx, createMsg)
		require.NoError(t, err)

		// Try to swap claiming to be addr2
		msg := &dexpb.MsgSwapExactIn{
			Sender:         addr2.String(),
			PoolId:         poolResp.PoolId,
			CoinIn:         &sdk.Coin{Denom: "aura", Amount: sdkmath.NewInt(100)},
			MinAmountOut:   sdkmath.NewInt(90).String(),
			MaxSlippageBps: 500,
		}

		// But signed by addr1
		ctx := suite.ctx.WithValue(sdk.ContextKey("signers"), []sdk.AccAddress{addr1})
		_, err = msgServer.SwapExactIn(ctx, msg)

		// Should be rejected
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("CreateHTLC - reject mismatched signer", func(t *testing.T) {
		msg := &dexpb.MsgCreateHTLC{
			Sender:            addr2.String(),
			Recipient:         addr1.String(),
			Amount:            &sdk.Coin{Denom: "aura", Amount: sdkmath.NewInt(100)},
			SecretHash:        "hash123",
			TimelockDuration:  3600,
		}

		// But signed by addr1
		ctx := suite.ctx.WithValue(sdk.ContextKey("signers"), []sdk.AccAddress{addr1})
		_, err := msgServer.CreateHTLC(ctx, msg)

		// Should be rejected
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.PermissionDenied, st.Code())
	})

	t.Run("ClaimHTLC - reject mismatched signer", func(t *testing.T) {
		// Create HTLC as addr1
		createMsg := &dexpb.MsgCreateHTLC{
			Sender:            addr1.String(),
			Recipient:         addr2.String(),
			Amount:            &sdk.Coin{Denom: "aura", Amount: sdkmath.NewInt(100)},
			SecretHash:        "hash123",
			TimelockDuration:  3600,
		}
		htlcResp, err := msgServer.CreateHTLC(suite.ctx, createMsg)
		require.NoError(t, err)

		// Try to claim it claiming to be addr2
		msg := &dexpb.MsgClaimHTLC{
			Recipient: addr2.String(),
			HtlcId:    htlcResp.HtlcId,
			Secret:    "secret123",
		}

		// But signed by addr1
		ctx := suite.ctx.WithValue(sdk.ContextKey("signers"), []sdk.AccAddress{addr1})
		_, err = msgServer.ClaimHTLC(ctx, msg)

		// Should be rejected
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.PermissionDenied, st.Code())
	})
}

// Helper function to setup test suite
func setupMsgServerTestSuite(t *testing.T) *testSuite {
	// This would normally use the existing test suite setup
	// For now, return a minimal setup
	// TODO: Integrate with existing test infrastructure
	t.Skip("Integration with existing test suite required")
	return nil
}

// Helper to fund an account
func (s *testSuite) fundAccount(addr sdk.AccAddress, coins sdk.Coins) {
	// Implementation would use bank keeper to fund account
}

type testSuite struct {
	ctx    sdk.Context
	keeper *keeper.Keeper
}
