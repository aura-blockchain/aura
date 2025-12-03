package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testutil/keeper"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestSetContractAdmin tests setting contract admin
func TestSetContractAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")

	t.Run("success - set admin", func(t *testing.T) {
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Verify admin was set
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, adminAddr, storedAdmin)
	})

	t.Run("failure - empty contract address", func(t *testing.T) {
		err := k.SetContractAdmin(ctx, sdk.AccAddress{}, adminAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract address cannot be empty")
	})

	t.Run("failure - empty admin address", func(t *testing.T) {
		err := k.SetContractAdmin(ctx, contractAddr, sdk.AccAddress{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "admin address cannot be empty")
	})
}

// TestGetContractAdmin tests retrieving contract admin
func TestGetContractAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")

	t.Run("success - get existing admin", func(t *testing.T) {
		// Set admin first
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Retrieve admin
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, adminAddr, storedAdmin)
	})

	t.Run("success - get non-existent admin returns empty", func(t *testing.T) {
		nonExistentContract := sdk.AccAddress("nonexistent_________")
		storedAdmin, err := k.GetContractAdmin(ctx, nonExistentContract)
		require.NoError(t, err)
		require.True(t, storedAdmin.Empty())
	})

	t.Run("failure - empty contract address", func(t *testing.T) {
		_, err := k.GetContractAdmin(ctx, sdk.AccAddress{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract address cannot be empty")
	})
}

// TestDeleteContractAdmin tests deleting contract admin
func TestDeleteContractAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")

	t.Run("success - delete existing admin", func(t *testing.T) {
		// Set admin first
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Verify admin exists
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, adminAddr, storedAdmin)

		// Delete admin
		err = k.DeleteContractAdmin(ctx, contractAddr)
		require.NoError(t, err)

		// Verify admin is gone
		storedAdmin, err = k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.True(t, storedAdmin.Empty())
	})

	t.Run("success - delete non-existent admin (idempotent)", func(t *testing.T) {
		nonExistentContract := sdk.AccAddress("nonexistent_________")
		err := k.DeleteContractAdmin(ctx, nonExistentContract)
		require.NoError(t, err)
	})

	t.Run("failure - empty contract address", func(t *testing.T) {
		err := k.DeleteContractAdmin(ctx, sdk.AccAddress{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract address cannot be empty")
	})
}

// TestHasContractAdmin tests checking if contract has admin
func TestHasContractAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")

	t.Run("has admin - true", func(t *testing.T) {
		// Set admin
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Check has admin
		hasAdmin := k.HasContractAdmin(ctx, contractAddr)
		require.True(t, hasAdmin)
	})

	t.Run("has admin - false for non-existent contract", func(t *testing.T) {
		nonExistentContract := sdk.AccAddress("nonexistent_________")
		hasAdmin := k.HasContractAdmin(ctx, nonExistentContract)
		require.False(t, hasAdmin)
	})

	t.Run("has admin - false after deletion", func(t *testing.T) {
		// Set admin
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Delete admin
		err = k.DeleteContractAdmin(ctx, contractAddr)
		require.NoError(t, err)

		// Check has admin
		hasAdmin := k.HasContractAdmin(ctx, contractAddr)
		require.False(t, hasAdmin)
	})
}

// TestIsContractAdmin tests checking if address is contract admin
func TestIsContractAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")
	otherAddr := sdk.AccAddress("other_______________")

	t.Run("is admin - true", func(t *testing.T) {
		// Set admin
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Check is admin
		isAdmin, err := k.IsContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)
		require.True(t, isAdmin)
	})

	t.Run("is admin - false for different address", func(t *testing.T) {
		// Set admin
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Check is admin with different address
		isAdmin, err := k.IsContractAdmin(ctx, contractAddr, otherAddr)
		require.NoError(t, err)
		require.False(t, isAdmin)
	})

	t.Run("is admin - false for non-existent contract", func(t *testing.T) {
		nonExistentContract := sdk.AccAddress("nonexistent_________")
		isAdmin, err := k.IsContractAdmin(ctx, nonExistentContract, adminAddr)
		require.NoError(t, err)
		require.False(t, isAdmin)
	})
}

// TestMsgUpdateAdmin tests the UpdateAdmin message handler
func TestMsgUpdateAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")
	currentAdmin := sdk.AccAddress("currentadmin________")
	newAdmin := sdk.AccAddress("newadmin____________")
	nonAdmin := sdk.AccAddress("nonadmin____________")

	// Setup: Set initial admin
	err := k.SetContractAdmin(ctx, contractAddr, currentAdmin)
	require.NoError(t, err)

	t.Run("success - admin updates admin", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   currentAdmin.String(),
			Contract: contractAddr.String(),
			NewAdmin: newAdmin.String(),
		}

		resp, err := msgServer.UpdateAdmin(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify admin was updated
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, newAdmin, storedAdmin)
	})

	t.Run("failure - non-admin cannot update", func(t *testing.T) {
		// Reset admin to currentAdmin
		err := k.SetContractAdmin(ctx, contractAddr, currentAdmin)
		require.NoError(t, err)

		msg := &types.MsgUpdateAdmin{
			Sender:   nonAdmin.String(),
			Contract: contractAddr.String(),
			NewAdmin: newAdmin.String(),
		}

		_, err = msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")

		// Verify admin was not changed
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, currentAdmin, storedAdmin)
	})

	t.Run("failure - contract has no admin", func(t *testing.T) {
		noAdminContract := sdk.AccAddress("noadmincontract_____")

		msg := &types.MsgUpdateAdmin{
			Sender:   currentAdmin.String(),
			Contract: noAdminContract.String(),
			NewAdmin: newAdmin.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract has no admin set")
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
			Sender:   currentAdmin.String(),
			Contract: "invalid_contract",
			NewAdmin: newAdmin.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid contract address")
	})

	t.Run("failure - invalid new admin address", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   currentAdmin.String(),
			Contract: contractAddr.String(),
			NewAdmin: "invalid_admin",
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid new admin address")
	})
}

// TestMsgClearAdmin tests the ClearAdmin message handler
func TestMsgClearAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")
	nonAdmin := sdk.AccAddress("nonadmin____________")

	// Setup: Set initial admin
	err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
	require.NoError(t, err)

	t.Run("success - admin clears admin", func(t *testing.T) {
		msg := &types.MsgClearAdmin{
			Sender:   adminAddr.String(),
			Contract: contractAddr.String(),
		}

		resp, err := msgServer.ClearAdmin(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify admin was cleared
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.True(t, storedAdmin.Empty())
	})

	t.Run("failure - non-admin cannot clear", func(t *testing.T) {
		// Reset admin
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		msg := &types.MsgClearAdmin{
			Sender:   nonAdmin.String(),
			Contract: contractAddr.String(),
		}

		_, err = msgServer.ClearAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")

		// Verify admin was not cleared
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, adminAddr, storedAdmin)
	})

	t.Run("failure - contract has no admin to clear", func(t *testing.T) {
		noAdminContract := sdk.AccAddress("noadmincontract_____")

		msg := &types.MsgClearAdmin{
			Sender:   adminAddr.String(),
			Contract: noAdminContract.String(),
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "contract has no admin to clear")
	})

	t.Run("failure - cannot clear already cleared admin", func(t *testing.T) {
		// Clear admin first
		err := k.DeleteContractAdmin(ctx, contractAddr)
		require.NoError(t, err)

		msg := &types.MsgClearAdmin{
			Sender:   adminAddr.String(),
			Contract: contractAddr.String(),
		}

		_, err = msgServer.ClearAdmin(ctx, msg)
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
			Sender:   adminAddr.String(),
			Contract: "invalid_contract",
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid contract address")
	})
}

// TestAdminOnlyMigration tests that only admins can migrate contracts
func TestAdminOnlyMigration(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")
	nonAdmin := sdk.AccAddress("nonadmin____________")

	// Setup: Set admin
	err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
	require.NoError(t, err)

	t.Run("failure - non-admin cannot migrate", func(t *testing.T) {
		msg := &types.MsgMigrateContract{
			Sender:   nonAdmin.String(),
			Contract: contractAddr.String(),
			CodeId:   2,
			Msg:      []byte("{}"),
		}

		_, err := msgServer.MigrateContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")
	})

	t.Run("failure - cleared admin cannot migrate", func(t *testing.T) {
		// Clear admin
		err := k.DeleteContractAdmin(ctx, contractAddr)
		require.NoError(t, err)

		msg := &types.MsgMigrateContract{
			Sender:   adminAddr.String(),
			Contract: contractAddr.String(),
			CodeId:   2,
			Msg:      []byte("{}"),
		}

		_, err := msgServer.MigrateContract(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")
	})
}

// TestAdminChangeChain tests a chain of admin changes
func TestAdminChangeChain(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")
	admin1 := sdk.AccAddress("admin1______________")
	admin2 := sdk.AccAddress("admin2______________")
	admin3 := sdk.AccAddress("admin3______________")

	// Setup: Set initial admin
	err := k.SetContractAdmin(ctx, contractAddr, admin1)
	require.NoError(t, err)

	t.Run("admin1 transfers to admin2", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   admin1.String(),
			Contract: contractAddr.String(),
			NewAdmin: admin2.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.NoError(t, err)

		// Verify admin2 is now admin
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, admin2, storedAdmin)
	})

	t.Run("admin1 can no longer update admin", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   admin1.String(),
			Contract: contractAddr.String(),
			NewAdmin: admin3.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not the contract admin")
	})

	t.Run("admin2 transfers to admin3", func(t *testing.T) {
		msg := &types.MsgUpdateAdmin{
			Sender:   admin2.String(),
			Contract: contractAddr.String(),
			NewAdmin: admin3.String(),
		}

		_, err := msgServer.UpdateAdmin(ctx, msg)
		require.NoError(t, err)

		// Verify admin3 is now admin
		storedAdmin, err := k.GetContractAdmin(ctx, contractAddr)
		require.NoError(t, err)
		require.Equal(t, admin3, storedAdmin)
	})

	t.Run("admin3 clears admin", func(t *testing.T) {
		msg := &types.MsgClearAdmin{
			Sender:   admin3.String(),
			Contract: contractAddr.String(),
		}

		_, err := msgServer.ClearAdmin(ctx, msg)
		require.NoError(t, err)

		// Verify no admin
		hasAdmin := k.HasContractAdmin(ctx, contractAddr)
		require.False(t, hasAdmin)
	})

	t.Run("no one can update admin after cleared", func(t *testing.T) {
		for _, admin := range []sdk.AccAddress{admin1, admin2, admin3} {
			msg := &types.MsgUpdateAdmin{
				Sender:   admin.String(),
				Contract: contractAddr.String(),
				NewAdmin: admin1.String(),
			}

			_, err := msgServer.UpdateAdmin(ctx, msg)
			require.Error(t, err)
			require.Contains(t, err.Error(), "contract has no admin set")
		}
	})
}
