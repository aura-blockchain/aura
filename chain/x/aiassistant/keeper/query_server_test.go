package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

func TestQueryServerParams(t *testing.T) {
	k, ctx, _ := setupKeeper(t)
	server := keeper.NewQueryServer(k)

	resp, err := server.Params(sdk.WrapSDKContext(ctx), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), *resp.Params)
}

func TestQueryServerAssistant(t *testing.T) {
	k, ctx, bank := setupKeeper(t)
	server := keeper.NewQueryServer(k)

	assistant := buildTestAssistant(t, ctx, k, bank, "assistant1")

	found, err := server.Assistant(sdk.WrapSDKContext(ctx), &types.QueryAssistantRequest{
		AssistantAddress: assistant.AssistantAddress,
	})
	require.NoError(t, err)
	require.Equal(t, assistant.AssistantAddress, found.Assistant.AssistantAddress)
}

func TestQueryServerAssistants(t *testing.T) {
	k, ctx, bank := setupKeeper(t)
	server := keeper.NewQueryServer(k)

	buildTestAssistant(t, ctx, k, bank, "assistant1")
	buildTestAssistant(t, ctx, k, bank, "assistant2")

	resp, err := server.Assistants(sdk.WrapSDKContext(ctx), &types.QueryAssistantsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Assistants, 2)
}

func buildTestAssistant(t *testing.T, ctx sdk.Context, k *keeper.Keeper, bank *mockBankKeeper, label string) *types.Assistant {
	t.Helper()
	owner := fixedAddr("owner")
	bank.fundAccount(owner.String(), sdk.NewCoins(sdk.NewInt64Coin(types.DefaultStakeDenom, 20_000_000)))
	k.SetParams(ctx, types.DefaultParams())

	msg := &types.MsgRegisterAssistant{
		AssistantAddress:  fixedAddr(label).String(),
		OwnerAddress:      owner.String(),
		Locales:           []string{"en-us"},
		ModelHash:         "model",
		ApiKeyFingerprint: "fp",
		Stake: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: sdkmath.NewInt(10_000_000),
		},
		Sponsorship: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: sdkmath.ZeroInt(),
		},
	}
	asst, err := k.RegisterAssistant(ctx, msg)
	require.NoError(t, err)
	return asst
}

func fixedAddr(label string) sdk.AccAddress {
	base := label + "____________________"
	return sdk.AccAddress(base[:20])
}
