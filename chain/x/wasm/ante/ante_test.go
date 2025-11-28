package ante_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/wasm/ante"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
)

func TestWasmGasDecorator(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	decorator := ante.NewWasmGasDecorator(k)

	sender := sdk.AccAddress("sender______________")
	err := k.AuthorizeUploader(ctx, sender.String())
	require.NoError(t, err)

	t.Run("success - store code", func(t *testing.T) {
		msg := &types.MsgStoreCode{
			Sender:       sender.String(),
			WasmByteCode: []byte("dummy code"),
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.NoError(t, err)
	})

	t.Run("failure - contract too large for block", func(t *testing.T) {
		params := k.GetParams(ctx)
		largeCode := make([]byte, params.MaxWasmCodeSize+1)

		msg := &types.MsgStoreCode{
			Sender:       sender.String(),
			WasmByteCode: largeCode,
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.Error(t, err)
		require.Contains(t, err.Error(), "per block exceeded")
	})
}

func TestWasmAuthDecorator(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	decorator := ante.NewWasmAuthDecorator(k)

	authorizedSender := sdk.AccAddress("authorized__________")
	unauthorizedSender := sdk.AccAddress("unauthorized________")

	err := k.AuthorizeUploader(ctx, authorizedSender.String())
	require.NoError(t, err)

	t.Run("success - authorized uploader", func(t *testing.T) {
		msg := &types.MsgStoreCode{
			Sender:       authorizedSender.String(),
			WasmByteCode: []byte("dummy code"),
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.NoError(t, err)
	})

	t.Run("failure - unauthorized uploader", func(t *testing.T) {
		msg := &types.MsgStoreCode{
			Sender:       unauthorizedSender.String(),
			WasmByteCode: []byte("dummy code"),
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not authorized")
	})
}

func TestWasmReentrancyDecorator(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	decorator := ante.NewWasmReentrancyDecorator(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - single execution", func(t *testing.T) {
		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      []byte(`{"execute":"action"}`),
			Funds:    nil,
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.NoError(t, err)
	})

	t.Run("failure - reentrancy in same tx", func(t *testing.T) {
		msg1 := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      []byte(`{"execute":"action1"}`),
			Funds:    nil,
		}

		msg2 := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      []byte(`{"execute":"action2"}`),
			Funds:    nil,
		}

		tx := mockTx{msgs: []sdk.Msg{msg1, msg2}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reentrancy")
	})
}

func TestWasmSecurityDecorator(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	decorator := ante.NewWasmSecurityDecorator(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - valid code", func(t *testing.T) {
		msg := &types.MsgStoreCode{
			Sender:       sender.String(),
			WasmByteCode: []byte("valid wasm code"),
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.NoError(t, err)
	})

	t.Run("failure - empty code", func(t *testing.T) {
		msg := &types.MsgStoreCode{
			Sender:       sender.String(),
			WasmByteCode: []byte{},
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err := decorator.AnteHandle(ctx, tx, false, mockNext)
		require.Error(t, err)
	})

	t.Run("success - migration with admin requirement", func(t *testing.T) {
		// Ensure admin requirement is enabled
		params := k.GetParams(ctx)
		params.RequireAdminForMigrate = true
		err := k.SetParams(ctx, params)
		require.NoError(t, err)

		msg := &types.MsgMigrateContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			CodeId:   2,
			Msg:      []byte(`{"migrate":"data"}`),
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err = decorator.AnteHandle(ctx, tx, false, mockNext)
		require.NoError(t, err)
	})

	t.Run("failure - paused contract", func(t *testing.T) {
		// Pause contract
		err := k.PauseContract(ctx, contractAddr.String())
		require.NoError(t, err)

		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      []byte(`{"execute":"action"}`),
			Funds:    nil,
		}

		tx := mockTx{msgs: []sdk.Msg{msg}}

		_, err = decorator.AnteHandle(ctx, tx, false, mockNext)
		require.Error(t, err)
		require.Contains(t, err.Error(), "paused")

		// Cleanup
		err = k.UnpauseContract(ctx, contractAddr.String())
		require.NoError(t, err)
	})
}

// Mock types for testing

type mockTx struct {
	msgs []sdk.Msg
}

func (tx mockTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

// GetMsgsV2 implements the sdk.Tx interface for SDK v0.50
func (tx mockTx) GetMsgsV2() ([]protov2.Message, error) {
	// Return nil for test purposes - this is sufficient for ante decorator testing
	return nil, nil
}

func mockNext(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	return ctx, nil
}
