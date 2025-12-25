// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/aiassistant/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

const (
	testStakeAmount       int64 = 10_000_000
	testSponsorshipAmount int64 = 1_000_000
)

func TestMsgServerRegisterAssistant(t *testing.T) {
	fx := newMsgServerFixture(t)
	resp := fx.registerDefaultAssistant(t, []string{"EN-us", "es-ES"})

	require.Equal(t, types.AssistantStatus_ACTIVE, resp.Status)
	require.Equal(t, []string{"en-us", "es-es"}, resp.Locales)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(types.DefaultStakeDenom, testStakeAmount+testSponsorshipAmount)), fx.bank.moduleBalance(types.ModuleName))
}

func TestMsgServerUpdateLocales(t *testing.T) {
	fx := newMsgServerFixture(t)
	assistant := fx.registerDefaultAssistant(t, []string{"en-us"})

	resp, err := fx.server.UpdateLocales(fx.wrapCtx(), &types.MsgUpdateLocales{
		AssistantAddress: assistant.AssistantAddress,
		OwnerAddress:     fx.owner.String(),
		Locales:          []string{"fr-FR", "en-US", "ES-es"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"en-us", "es-es", "fr-fr"}, resp.Assistant.Locales)

	_, err = fx.server.UpdateLocales(fx.wrapCtx(), &types.MsgUpdateLocales{
		AssistantAddress: assistant.AssistantAddress,
		OwnerAddress:     randAddr().String(),
		Locales:          []string{"jp-JP"},
	})
	require.ErrorIs(t, err, types.ErrUnauthorizedOperator)
}

func TestMsgServerHeartbeat(t *testing.T) {
	fx := newMsgServerFixture(t)
	assistant := fx.registerDefaultAssistant(t, []string{"en-us"})

	resp, err := fx.server.Heartbeat(fx.wrapCtx(), &types.MsgHeartbeat{
		AssistantAddress: assistant.AssistantAddress,
		OperatorAddress:  fx.owner.String(),
	})
	require.NoError(t, err)

	params, err := fx.keeper.GetParams(fx.ctx)
	require.NoError(t, err)
	expectedNextSlash := fx.ctx.BlockTime().Add(time.Duration(params.HeartbeatWindowSeconds+params.HeartbeatGraceSeconds) * time.Second).Unix()
	require.Equal(t, expectedNextSlash, int64(resp.NextSlashTime))

	stored, ok := fx.keeper.GetAssistant(fx.ctx, assistant.AssistantAddress)
	require.True(t, ok)
	require.True(t, stored.LastHeartbeat.Equal(fx.ctx.BlockTime()))

	_, err = fx.server.Heartbeat(fx.wrapCtx(), &types.MsgHeartbeat{
		AssistantAddress: assistant.AssistantAddress,
		OperatorAddress:  randAddr().String(),
	})
	require.ErrorIs(t, err, types.ErrUnauthorizedOperator)
}

func TestMsgServerReportMisbehavior(t *testing.T) {
	fx := newMsgServerFixture(t)
	assistant := fx.registerDefaultAssistant(t, []string{"en-us"})

	resp, err := fx.server.ReportMisbehavior(fx.wrapCtx(), &types.MsgReportMisbehavior{
		Reporter:         randAddr().String(),
		AssistantAddress: assistant.AssistantAddress,
		Infraction:       "double-sign",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Assistant)
	require.False(t, resp.SlashAmount.Amount.IsZero())
	require.NotEqual(t, types.AssistantStatus_ACTIVE, resp.Assistant.Status)
	require.Equal(t, uint64(1), resp.Assistant.MisbehaviorReports)
}

func TestMsgServerUpdateParamsAuthorization(t *testing.T) {
	authority := randAddr().String()
	k, ctx, _ := setupKeeperWithAuthority(t, authority)
	server := keeper.NewMsgServer(k)

	params := types.DefaultParams()
	params.MaxLocales = 7

	_, err := server.UpdateParams(sdk.WrapSDKContext(ctx), &types.MsgUpdateParams{
		Authority: authority,
		Params:    params,
	})
	require.NoError(t, err)
	updatedParams, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(7), updatedParams.MaxLocales)

	_, err = server.UpdateParams(sdk.WrapSDKContext(ctx), &types.MsgUpdateParams{
		Authority: randAddr().String(),
		Params:    params,
	})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	require.ErrorContains(t, err, types.ErrUnauthorizedOperator.Error())
}

type msgServerFixture struct {
	t             *testing.T
	keeper        *keeper.Keeper
	server        types.MsgServer
	ctx           sdk.Context
	bank          *mockBankKeeper
	owner         sdk.AccAddress
	assistantAddr string
}

func newMsgServerFixture(t *testing.T) *msgServerFixture {
	t.Helper()
	k, ctx, bank := setupKeeper(t)
	ctx = ctx.WithBlockTime(time.Date(2025, 12, 15, 10, 0, 0, 0, time.UTC))

	owner := randAddr()
	bank.fundAccount(owner.String(), sdk.NewCoins(sdk.NewInt64Coin(types.DefaultStakeDenom, 50_000_000)))

	return &msgServerFixture{
		t:             t,
		keeper:        k,
		server:        keeper.NewMsgServer(k),
		ctx:           ctx,
		bank:          bank,
		owner:         owner,
		assistantAddr: randAddr().String(),
	}
}

func (f *msgServerFixture) registerDefaultAssistant(t *testing.T, locales []string) *types.Assistant {
	t.Helper()
	msg := f.buildRegisterMsg(locales)
	resp, err := f.server.RegisterAssistant(f.wrapCtx(), msg)
	require.NoError(t, err)
	return resp.Assistant
}

func (f *msgServerFixture) buildRegisterMsg(locales []string) *types.MsgRegisterAssistant {
	if len(locales) == 0 {
		locales = []string{"en-us"}
	}
	return &types.MsgRegisterAssistant{
		AssistantAddress:  f.assistantAddr,
		OwnerAddress:      f.owner.String(),
		Locales:           locales,
		ModelHash:         "model-v1",
		ApiKeyFingerprint: "fp_test",
		Stake: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: sdkmath.NewInt(testStakeAmount),
		},
		Sponsorship: types.Balance{
			Denom:  types.DefaultStakeDenom,
			Amount: sdkmath.NewInt(testSponsorshipAmount),
		},
	}
}

func (f *msgServerFixture) wrapCtx() context.Context {
	return sdk.WrapSDKContext(f.ctx)
}

func randAddr() sdk.AccAddress {
	return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}
