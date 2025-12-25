// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package compliance_test

import (
	stderrors "errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/compliance"
	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
)

func TestIBCModule_DisabledHandlersReturnErrors(t *testing.T) {
	ctx := keeper.CreateTestInput(t).Ctx
	im := compliance.NewIBCModule(nil)

	_, err := im.OnChanOpenInit(ctx, channeltypes.UNORDERED, nil, "compliance", "channel-0", nil, channeltypes.Counterparty{}, "v0")
	require.Error(t, err)
	require.True(t, stderrors.Is(err, compliancetypes.ErrIBCNotEnabled))

	_, err = im.OnChanOpenTry(ctx, channeltypes.UNORDERED, nil, "compliance", "channel-0", nil, channeltypes.Counterparty{}, "v0")
	require.Error(t, err)
	require.True(t, stderrors.Is(err, compliancetypes.ErrIBCNotEnabled))

	require.True(t, stderrors.Is(im.OnChanOpenAck(ctx, "compliance", "channel-0", "counterparty-0", "v0"), compliancetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnChanOpenConfirm(ctx, "compliance", "channel-0"), compliancetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnChanCloseInit(ctx, "compliance", "channel-0"), compliancetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnChanCloseConfirm(ctx, "compliance", "channel-0"), compliancetypes.ErrIBCNotEnabled))

	ack := im.OnRecvPacket(ctx, channeltypes.Packet{}, sdk.AccAddress{})
	require.NotNil(t, ack)
	require.False(t, ack.Success())
	require.NotEmpty(t, ack.Acknowledgement())
	var ackObj channeltypes.Acknowledgement
	require.NoError(t, channeltypes.SubModuleCdc.UnmarshalJSON(ack.Acknowledgement(), &ackObj))
	require.False(t, ackObj.Success())
	require.NoError(t, ackObj.ValidateBasic())

	require.True(t, stderrors.Is(im.OnAcknowledgementPacket(ctx, channeltypes.Packet{}, nil, sdk.AccAddress{}), compliancetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnTimeoutPacket(ctx, channeltypes.Packet{}, sdk.AccAddress{}), compliancetypes.ErrIBCNotEnabled))

	_, err = im.NegotiateAppVersion(ctx, channeltypes.UNORDERED, "connection-0", "compliance", channeltypes.Counterparty{}, "v0")
	require.Error(t, err)
	require.True(t, stderrors.Is(err, compliancetypes.ErrIBCNotEnabled))

	version, ok := im.GetAppVersion(ctx, "compliance", "channel-0")
	require.False(t, ok)
	require.Empty(t, version)

	err = im.SendPacket(ctx, "compliance", "channel-0", 0, 0, []byte("data"))
	require.Error(t, err)
	require.True(t, stderrors.Is(err, compliancetypes.ErrIBCNotEnabled))
}

