package keeper_test

import (
	"encoding/json"
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgStoreCode(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")

	t.Run("success - authorized uploader", func(t *testing.T) {
		// Authorize the uploader
		err := k.AuthorizeUploader(ctx, sender.String())
		require.NoError(t, err)

		msg := &types.MsgStoreCode{
			Sender:       sender.String(),
			WASMByteCode: []byte("dummy wasm code"),
		}

		resp, err := msgServer.StoreCode(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Greater(t, resp.CodeID, uint64(0))

		// Verify security stats updated
		stats := k.GetSecurityStats(ctx)
		require.Equal(t, uint64(1), stats.TotalContractsUploaded)
	})

	t.Run("failure - unauthorized uploader", func(t *testing.T) {
		unauthorizedSender := sdk.AccAddress("unauthorized________")

		msg := &types.MsgStoreCode{
			Sender:       unauthorizedSender.String(),
			WASMByteCode: []byte("dummy wasm code"),
		}

		_, err := msgServer.StoreCode(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not authorized")
	})

	t.Run("failure - empty code", func(t *testing.T) {
		msg := &types.MsgStoreCode{
			Sender:       sender.String(),
			WASMByteCode: []byte{},
		}

		_, err := msgServer.StoreCode(ctx, msg)
		require.Error(t, err)
	})

	t.Run("failure - code too large", func(t *testing.T) {
		// Create code larger than max size
		params := k.GetParams(ctx)
		largeCode := make([]byte, params.MaxContractSize+1)

		msg := &types.MsgStoreCode{
			Sender:       sender.String(),
			WASMByteCode: largeCode,
		}

		_, err := msgServer.StoreCode(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "exceeds maximum")
	})
}

func TestMsgInstantiateContract(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")
	admin := sdk.AccAddress("admin_______________")

	t.Run("success - with admin", func(t *testing.T) {
		initMsg := json.RawMessage(`{"init":"data"}`)

		msg := &types.MsgInstantiateContract{
			Sender: sender.String(),
			Admin:  admin.String(),
			CodeID: 1,
			Label:  "test contract",
			Msg:    initMsg,
			Funds:  sdk.NewCoins(),
		}

		resp, err := msgServer.InstantiateContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Address)

		// Verify security stats updated
		stats := k.GetSecurityStats(ctx)
		require.Equal(t, uint64(1), stats.TotalContractsInstantiated)
	})

	t.Run("success - without admin", func(t *testing.T) {
		initMsg := json.RawMessage(`{"init":"data"}`)

		msg := &types.MsgInstantiateContract{
			Sender: sender.String(),
			Admin:  "",
			CodeID: 1,
			Label:  "test contract no admin",
			Msg:    initMsg,
			Funds:  sdk.NewCoins(),
		}

		resp, err := msgServer.InstantiateContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("failure - invalid code id", func(t *testing.T) {
		initMsg := json.RawMessage(`{"init":"data"}`)

		msg := &types.MsgInstantiateContract{
			Sender: sender.String(),
			Admin:  admin.String(),
			CodeID: 0,
			Label:  "test",
			Msg:    initMsg,
			Funds:  sdk.NewCoins(),
		}

		err := msg.ValidateBasic()
		require.Error(t, err)
	})
}

func TestMsgExecuteContract(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - normal execution", func(t *testing.T) {
		execMsg := json.RawMessage(`{"execute":"action"}`)

		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      execMsg,
			Funds:    sdk.NewCoins(),
		}

		resp, err := msgServer.ExecuteContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify security stats updated
		stats := k.GetSecurityStats(ctx)
		require.Equal(t, uint64(1), stats.TotalExecutions)
	})

	t.Run("failure - paused contract", func(t *testing.T) {
		// Pause the contract
		err := k.PauseContract(ctx, contractAddr.String())
		require.NoError(t, err)

		execMsg := json.RawMessage(`{"execute":"action"}`)

		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      execMsg,
			Funds:    sdk.NewCoins(),
		}

		_, err = msgServer.ExecuteContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "paused")

		// Unpause for other tests
		err = k.UnpauseContract(ctx, contractAddr.String())
		require.NoError(t, err)
	})

	t.Run("failure - reentrancy detected", func(t *testing.T) {
		// Mark contract as executing
		k.SetExecuting(ctx, contractAddr.String(), true)

		execMsg := json.RawMessage(`{"execute":"action"}`)

		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      execMsg,
			Funds:    sdk.NewCoins(),
		}

		_, err := msgServer.ExecuteContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reentrancy")

		// Cleanup
		k.SetExecuting(ctx, contractAddr.String(), false)

		// Verify reentrancy stats updated
		stats := k.GetSecurityStats(ctx)
		require.Greater(t, stats.ReentrancyAttemptsBlocked, uint64(0))
	})
}

func TestMsgMigrateContract(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - migration enabled", func(t *testing.T) {
		// Enable migration
		params := k.GetParams(ctx)
		params.EnableMigration = true
		err := k.SetParams(ctx, params)
		require.NoError(t, err)

		migrateMsg := json.RawMessage(`{"migrate":"data"}`)

		msg := &types.MsgMigrateContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			CodeID:   2,
			Msg:      migrateMsg,
		}

		resp, err := msgServer.MigrateContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("failure - migration disabled", func(t *testing.T) {
		// Disable migration
		params := k.GetParams(ctx)
		params.EnableMigration = false
		err := k.SetParams(ctx, params)
		require.NoError(t, err)

		migrateMsg := json.RawMessage(`{"migrate":"data"}`)

		msg := &types.MsgMigrateContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			CodeID:   2,
			Msg:      migrateMsg,
		}

		_, err = msgServer.MigrateContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "migration")
	})
}

func TestMsgAuthorizeUploader(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	authority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" // Example authority
	uploader := sdk.AccAddress("uploader____________")

	t.Run("success - authorize uploader", func(t *testing.T) {
		msg := &types.MsgAuthorizeUploader{
			Authority: authority,
			Uploader:  uploader.String(),
		}

		resp, err := msgServer.AuthorizeUploader(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify uploader is authorized
		isAuthorized := k.IsAuthorizedUploader(ctx, uploader.String())
		require.True(t, isAuthorized)
	})

	t.Run("failure - invalid authority", func(t *testing.T) {
		msg := &types.MsgAuthorizeUploader{
			Authority: "invalid_authority",
			Uploader:  uploader.String(),
		}

		_, err := msgServer.AuthorizeUploader(ctx, msg)
		require.Error(t, err)
	})
}

func TestMsgPauseUnpauseContract(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	authority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - pause contract", func(t *testing.T) {
		msg := &types.MsgPauseContract{
			Authority: authority,
			Contract:  contractAddr.String(),
		}

		resp, err := msgServer.PauseContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify contract is paused
		isPaused := k.IsContractPaused(ctx, contractAddr.String())
		require.True(t, isPaused)

		// Verify stats updated
		stats := k.GetSecurityStats(ctx)
		require.Equal(t, uint64(1), stats.TotalPausedContracts)
	})

	t.Run("success - unpause contract", func(t *testing.T) {
		msg := &types.MsgUnpauseContract{
			Authority: authority,
			Contract:  contractAddr.String(),
		}

		resp, err := msgServer.UnpauseContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify contract is not paused
		isPaused := k.IsContractPaused(ctx, contractAddr.String())
		require.False(t, isPaused)

		// Verify stats updated
		stats := k.GetSecurityStats(ctx)
		require.Equal(t, uint64(0), stats.TotalPausedContracts)
	})
}

func TestMsgUpdateParams(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	authority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	t.Run("success - update params", func(t *testing.T) {
		newParams := types.DefaultParams()
		newParams.MaxContractSize = 1000000

		msg := &types.MsgUpdateParams{
			Authority: authority,
			Params:    newParams,
		}

		resp, err := msgServer.UpdateParams(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify params updated
		params := k.GetParams(ctx)
		require.Equal(t, uint64(1000000), params.MaxContractSize)
	})

	t.Run("failure - invalid params", func(t *testing.T) {
		invalidParams := types.Params{
			MaxContractSize: 0, // Invalid
		}

		msg := &types.MsgUpdateParams{
			Authority: authority,
			Params:    invalidParams,
		}

		err := msg.ValidateBasic()
		require.Error(t, err)
	})
}
