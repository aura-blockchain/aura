package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

func TestMsgServerLockTokens_Success(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	require.NoError(t, k.AddSupportedChain(ctx, types.ChainConfig{ChainId: "paw", Enabled: true}))

	amount := sdk.NewCoin("uaura", sdkmath.NewInt(1_000))
	msg := &types.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		TargetChain: "paw",
		Recipient:   "paw1recipient",
		Amount:      &amount,
	}

	resp, err := ms.LockTokens(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.NotEmpty(t, resp.TransferId)

	exported := k.ExportGenesis(ctx)
	require.Len(t, exported.Transfers, 1)
	require.Equal(t, "paw", exported.Transfers[0].TargetChain)
}

func TestMsgServerLockTokens_MissingChain(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	amount := sdk.NewCoin("uaura", sdkmath.NewInt(1))
	msg := &types.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		TargetChain: "unknown",
		Recipient:   "paw1recipient",
		Amount:      &amount,
	}

	_, err := ms.LockTokens(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
}

func TestMsgServerMintTokens_CreatesWrappedToken(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	msg := &types.MsgMintTokens{
		Validator:    keepertest.GenTestAddr().String(),
		SourceChain:  "paw",
		SourceTxHash: "0xabc",
		Recipient:    keepertest.GenTestAddr().String(),
		Amount:       "1000",
		Denom:        "paw.token",
	}

	resp, err := ms.MintTokens(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.True(t, resp.Success)

	exported := k.ExportGenesis(ctx)
	require.Len(t, exported.WrappedTokens, 1)
	require.Equal(t, "paw.paw.token", exported.WrappedTokens[0].WrappedDenom)
}

func TestMsgServerUnlockTokens_CompletesTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	transferID := "transfer-999"
	seedBridgeTransfer(t, input, transferID, sdkmath.NewInt(500).String(), 1)

	msg := &types.MsgUnlockTokens{
		Sender:              keepertest.GenTestAddr().String(),
		SourceChain:         "paw",
		BurnTxHash:          transferID,
		Amount:              "500",
		Denom:               "uaura",
		ValidatorSignatures: [][]byte{{0x01}},
	}

	_, err := ms.UnlockTokens(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	exported := k.ExportGenesis(ctx)
	require.Equal(t, types.TransferStatus_COMPLETED, exported.Transfers[0].Status)
}
