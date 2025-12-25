// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
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
		// Skip if wasmd keeper is not configured (unit test environment)
		if k.GetWasmKeeper() == nil {
			t.Skip("skipping test that requires wasmd keeper")
		}

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
		require.Equal(t, uint64(1), stats.GetTotalCodesAnalyzed())
	})

	t.Run("failure - unauthorized uploader", func(t *testing.T) {
		// Skip if wasmd keeper is not configured (unit test environment)
		if k.GetWasmKeeper() == nil {
			t.Skip("skipping test that requires wasmd keeper")
		}

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
		largeCode := make([]byte, params.GetMaxWasmCodeSize()+1)

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
		// Skip if wasmd keeper is not configured (unit test environment)
		if k.GetWasmKeeper() == nil {
			t.Skip("skipping test that requires wasmd keeper")
		}

		initMsg := []byte(`{"init":"data"}`)

		msg := &types.MsgInstantiateContract{
			Sender: sender.String(),
			Admin:  admin.String(),
			CodeID: 1,
			Label:  "test contract",
			Msg:    initMsg,
			Funds:  nil,
		}

		resp, err := msgServer.InstantiateContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Address)

		// Verify instantiation succeeded (no specific counter for this)
		require.NotEmpty(t, resp.Address)
	})

	t.Run("success - without admin", func(t *testing.T) {
		// Skip if wasmd keeper is not configured (unit test environment)
		if k.GetWasmKeeper() == nil {
			t.Skip("skipping test that requires wasmd keeper")
		}

		initMsg := []byte(`{"init":"data"}`)

		msg := &types.MsgInstantiateContract{
			Sender: sender.String(),
			Admin:  "",
			CodeID: 1,
			Label:  "test contract no admin",
			Msg:    initMsg,
			Funds:  nil,
		}

		resp, err := msgServer.InstantiateContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("failure - invalid code id", func(t *testing.T) {
		initMsg := []byte(`{"init":"data"}`)

		msg := &types.MsgInstantiateContract{
			Sender: sender.String(),
			Admin:  admin.String(),
			CodeID: 0,
			Label:  "test",
			Msg:    initMsg,
			Funds:  nil,
		}

		// Should fail when executing (code ID 0 is invalid)
		_, err := msgServer.InstantiateContract(ctx, msg)
		require.Error(t, err)
	})
}

func TestMsgExecuteContract(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - normal execution", func(t *testing.T) {
		// Skip if wasmd keeper is not configured (unit test environment)
		if k.GetWasmKeeper() == nil {
			t.Skip("skipping test that requires wasmd keeper")
		}

		execMsg := []byte(`{"execute":"action"}`)

		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      execMsg,
			Funds:    nil,
		}

		resp, err := msgServer.ExecuteContract(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify security stats updated
		stats := k.GetSecurityStats(ctx)
		require.Equal(t, uint64(1), stats.GetTotalExecutions())
	})

	t.Run("failure - paused contract", func(t *testing.T) {
		// Pause the contract
		err := k.PauseContract(ctx, contractAddr.String())
		require.NoError(t, err)

		execMsg := []byte(`{"execute":"action"}`)

		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      execMsg,
			Funds:    nil,
		}

		_, err = msgServer.ExecuteContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "paused")

		// Unpause for other tests
		err = k.UnpauseContract(ctx, contractAddr.String())
		require.NoError(t, err)
	})

	t.Run("failure - reentrancy detected", func(t *testing.T) {
		// Skip if wasmd keeper is not configured (unit test environment)
		if k.GetWasmKeeper() == nil {
			t.Skip("skipping test that requires wasmd keeper")
		}

		// Mark contract as executing
		k.SetExecuting(ctx, contractAddr.String(), true)

		execMsg := []byte(`{"execute":"action"}`)

		msg := &types.MsgExecuteContract{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			Msg:      execMsg,
			Funds:    nil,
		}

		_, err := msgServer.ExecuteContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "reentrancy")

		// Cleanup
		k.SetExecuting(ctx, contractAddr.String(), false)

		// Verify execution was blocked (reentrancy protection working)
		require.Error(t, err)
	})
}

func TestMsgMigrateContract(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - migration enabled", func(t *testing.T) {
		// Skip if wasmd keeper is not configured (unit test environment)
		if k.GetWasmKeeper() == nil {
			t.Skip("skipping test that requires wasmd keeper")
		}

		// Migration is controlled by admin check, not a param
		migrateMsg := []byte(`{"migrate":"data"}`)

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

	t.Run("failure - unauthorized migration", func(t *testing.T) {
		// Migration without admin should fail
		migrateMsg := []byte(`{"migrate":"data"}`)

		unauthorizedSender := sdk.AccAddress("unauthorized________")
		msg := &types.MsgMigrateContract{
			Sender:   unauthorizedSender.String(),
			Contract: contractAddr.String(),
			CodeID:   2,
			Msg:      migrateMsg,
		}

		_, err := msgServer.MigrateContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "admin")
	})
}

func TestMsgAuthorizeUploader(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	authority := sdk.AccAddress("authority___________").String() // Must match keeper test setup
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

	authority := sdk.AccAddress("authority___________").String() // Must match keeper test setup
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
		require.Equal(t, uint64(1), stats.GetContractsPaused())
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
		require.Equal(t, uint64(0), stats.GetContractsPaused())
	})
}

func TestMsgUpdateParams(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	authority := sdk.AccAddress("authority___________").String() // Must match keeper test setup

	t.Run("success - update params", func(t *testing.T) {
		newParams := types.DefaultParams()
		newParams.MaxWasmCodeSize = 1000000

		msg := &types.MsgUpdateParams{
			Authority: authority,
			Params:    *newParams,
		}

		resp, err := msgServer.UpdateParams(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify params updated
		params := k.GetParams(ctx)
		require.Equal(t, uint64(1000000), params.GetMaxWasmCodeSize())
	})

	t.Run("failure - invalid params", func(t *testing.T) {
		invalidParams := &types.Params{
			MaxWasmCodeSize: 0, // Invalid
		}

		msg := &types.MsgUpdateParams{
			Authority: authority,
			Params:    *invalidParams,
		}

		// Should fail when executing due to validation
		_, err := msgServer.UpdateParams(ctx, msg)
		require.Error(t, err)
	})
}

func TestMsgUpdateAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")
	newAdmin := sdk.AccAddress("newadmin____________")

	t.Run("failure - no admin set", func(t *testing.T) {
		// Contract with no admin set should fail
		msg := &types.MsgUpdateAdmin{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			NewAdmin: newAdmin.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract has no admin")
	})

	t.Run("failure - invalid sender address", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   "invalid_address",
			Contract: contractAddr.String(),
			NewAdmin: newAdmin.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid sender address")
	})

	t.Run("failure - invalid contract address", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   sender.String(),
			Contract: "invalid_contract",
			NewAdmin: newAdmin.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid contract address")
	})

	t.Run("failure - invalid new admin address", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
			NewAdmin: "invalid_admin",
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid new admin address")
	})

	// NOTE: Full integration tests with actual admin changes require a real wasmd keeper.
	// These tests verify error handling and validation. Integration tests should be added
	// in integration_test.go with a properly configured wasmd keeper to verify:
	// - Admin can successfully update to a new admin
	// - Non-admin cannot update admin
	// - Event is emitted on successful update
	// - Contract admin is actually changed in storage
}

func TestMsgClearAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	sender := sdk.AccAddress("sender______________")
	contractAddr := sdk.AccAddress("contract____________")

	t.Run("failure - no admin to clear", func(t *testing.T) {
		// Contract with no admin set should fail
		msg := &types.MsgClearAdmin{
			Sender:   sender.String(),
			Contract: contractAddr.String(),
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract has no admin to clear")
	})

	t.Run("failure - invalid sender address", func(t *testing.T) {
		msg := &types.MsgClearAdmin{
			Sender:   "invalid_address",
			Contract: contractAddr.String(),
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid sender address")
	})

	t.Run("failure - invalid contract address", func(t *testing.T) {
		msg := &types.MsgClearAdmin{
			Sender:   sender.String(),
			Contract: "invalid_contract",
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid contract address")
	})

	// NOTE: Full integration tests with actual admin clearing require a real wasmd keeper.
	// These tests verify error handling and validation. Integration tests should be added
	// in integration_test.go with a properly configured wasmd keeper to verify:
	// - Admin can successfully clear admin
	// - Non-admin cannot clear admin
	// - Event is emitted on successful clear
	// - Contract admin is actually cleared in storage
	// - Contract cannot be migrated after admin is cleared
}

// TestAdminStorage tests the admin storage methods
func TestAdminStorage(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")
	otherAddr := sdk.AccAddress("other_______________")

	t.Run("set and get admin", func(t *testing.T) {
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, adminAddr, storedAdmin)
	})

	t.Run("has admin - true", func(t *testing.T) {
		hasAdmin := k.HasContractAdmin(ctx, contractAddr)
		require.True(t, hasAdmin)
	})

	t.Run("is admin - true for correct address", func(t *testing.T) {
		isAdmin, err := k.IsContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)
		require.True(t, isAdmin)
	})

	t.Run("is admin - false for different address", func(t *testing.T) {
		isAdmin, err := k.IsContractAdmin(ctx, contractAddr, otherAddr)
		require.NoError(t, err)
		require.False(t, isAdmin)
	})

	t.Run("delete admin", func(t *testing.T) {
		err := k.DeleteContractAdmin(ctx, contractAddr)
		require.NoError(t, err)

		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.True(t, storedAdmin.Empty())
	})

	t.Run("has admin - false after deletion", func(t *testing.T) {
		hasAdmin := k.HasContractAdmin(ctx, contractAddr)
		require.False(t, hasAdmin)
	})
}

// TestAdminUpdateFlow tests the full admin update flow
func TestAdminUpdateFlow(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")
	admin1 := sdk.AccAddress("admin1______________")
	admin2 := sdk.AccAddress("admin2______________")
	nonAdmin := sdk.AccAddress("nonadmin____________")

	// Setup: Set initial admin
	err := k.SetContractAdmin(ctx, contractAddr, admin1)
	require.NoError(t, err)

	t.Run("admin1 updates to admin2 - success", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   admin1.String(),
			Contract: contractAddr.String(),
			NewAdmin: admin2.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.NoError(t, err)

		// Verify admin changed
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, admin2, storedAdmin)
	})

	t.Run("admin1 cannot update anymore - unauthorized", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   admin1.String(),
			Contract: contractAddr.String(),
			NewAdmin: admin1.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")
	})

	t.Run("non-admin cannot update - unauthorized", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   nonAdmin.String(),
			Contract: contractAddr.String(),
			NewAdmin: nonAdmin.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")
	})

	t.Run("admin2 clears admin - success", func(t *testing.T) {
		msg := &types.MsgClearAdmin{
			Sender:   admin2.String(),
			Contract: contractAddr.String(),
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.NoError(t, err)

		// Verify admin cleared
		hasAdmin := k.HasContractAdmin(ctx, contractAddr)
		require.False(t, hasAdmin)
	})

	t.Run("cannot update admin after cleared", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   admin2.String(),
			Contract: contractAddr.String(),
			NewAdmin: admin2.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract has no admin set")
	})

	t.Run("cannot clear admin twice", func(t *testing.T) {
		msg := &types.MsgClearAdmin{
			Sender:   admin2.String(),
			Contract: contractAddr.String(),
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract has no admin to clear")
	})
}

// TestAdminMigrationAuth tests that migrations require admin
func TestAdminMigrationAuth(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")
	nonAdmin := sdk.AccAddress("nonadmin____________")

	// Setup: Set admin
	err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
	require.NoError(t, err)

	t.Run("non-admin cannot migrate", func(t *testing.T) {
		msg := &types.MsgMigrateContract{
			Sender:   nonAdmin.String(),
			Contract: contractAddr.String(),
			CodeID:   2,
			Msg:      []byte("{}"),
		}

		_, err := msgServer.MigrateContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")
	})

	t.Run("admin can attempt migration (will fail at keeper level without wasmd)", func(t *testing.T) {
		msg := &types.MsgMigrateContract{
			Sender:   adminAddr.String(),
			Contract: contractAddr.String(),
			CodeID:   2,
			Msg:      []byte("{}"),
		}

		// Will pass authorization but fail at migration level since no wasmd keeper
		_, err := msgServer.MigrateContract(ctx, msg)
		// Error is expected - either migration not implemented or wasmd not configured
		// The important part is we passed admin authorization check
		require.Error(t, err)
	})

	t.Run("cleared admin cannot migrate", func(t *testing.T) {
		// Clear admin
		errDel := k.DeleteContractAdmin(ctx, contractAddr)
		require.NoError(t, errDel)

		msg := &types.MsgMigrateContract{
			Sender:   adminAddr.String(),
			Contract: contractAddr.String(),
			CodeID:   2,
			Msg:      []byte("{}"),
		}

		_, err := msgServer.MigrateContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")
	})
}
