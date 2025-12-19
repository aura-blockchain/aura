package bridge_test

import (
	stderrors "errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
)

func TestIBCModule_DisabledHandlersReturnErrors(t *testing.T) {
	ctx := keeper.CreateTestInput(t).Ctx
	im := bridge.NewIBCModule(nil)

	_, err := im.OnChanOpenInit(ctx, channeltypes.UNORDERED, nil, "bridge", "channel-0", nil, channeltypes.Counterparty{}, "v0")
	require.Error(t, err)
	require.True(t, stderrors.Is(err, bridgetypes.ErrIBCNotEnabled))

	_, err = im.OnChanOpenTry(ctx, channeltypes.UNORDERED, nil, "bridge", "channel-0", nil, channeltypes.Counterparty{}, "v0")
	require.Error(t, err)
	require.True(t, stderrors.Is(err, bridgetypes.ErrIBCNotEnabled))

	require.True(t, stderrors.Is(im.OnChanOpenAck(ctx, "bridge", "channel-0", "counterparty-0", "v0"), bridgetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnChanOpenConfirm(ctx, "bridge", "channel-0"), bridgetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnChanCloseInit(ctx, "bridge", "channel-0"), bridgetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnChanCloseConfirm(ctx, "bridge", "channel-0"), bridgetypes.ErrIBCNotEnabled))

	ack := im.OnRecvPacket(ctx, channeltypes.Packet{}, sdk.AccAddress{})
	require.NotNil(t, ack)
	require.False(t, ack.Success())
	require.NotEmpty(t, ack.Acknowledgement())
	var ackObj channeltypes.Acknowledgement
	require.NoError(t, channeltypes.SubModuleCdc.UnmarshalJSON(ack.Acknowledgement(), &ackObj))
	require.False(t, ackObj.Success())
	require.NoError(t, ackObj.ValidateBasic())

	require.True(t, stderrors.Is(im.OnAcknowledgementPacket(ctx, channeltypes.Packet{}, nil, sdk.AccAddress{}), bridgetypes.ErrIBCNotEnabled))
	require.True(t, stderrors.Is(im.OnTimeoutPacket(ctx, channeltypes.Packet{}, sdk.AccAddress{}), bridgetypes.ErrIBCNotEnabled))

	_, err = im.NegotiateAppVersion(ctx, channeltypes.UNORDERED, "connection-0", "bridge", channeltypes.Counterparty{}, "v0")
	require.Error(t, err)
	require.True(t, stderrors.Is(err, bridgetypes.ErrIBCNotEnabled))

	version, ok := im.GetAppVersion(ctx, "bridge", "channel-0")
	require.False(t, ok)
	require.Empty(t, version)

	err = im.SendPacket(ctx, "bridge", "channel-0", 0, 0, []byte("data"))
	require.Error(t, err)
	require.True(t, stderrors.Is(err, bridgetypes.ErrIBCNotEnabled))
}

