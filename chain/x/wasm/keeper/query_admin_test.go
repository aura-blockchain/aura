package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestQueryContractAdmin(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")
	adminAddr := sdk.AccAddress("admin_______________")

	t.Run("success - query contract with admin", func(t *testing.T) {
		// Set admin
		err := k.SetContractAdmin(ctx, contractAddr, adminAddr)
		require.NoError(t, err)

		// Query admin
		req := &types.QueryContractAdminRequest{
			Address: contractAddr.String(),
		}

		resp, err := queryServer.ContractAdmin(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, adminAddr.String(), resp.Admin)
	})

	t.Run("success - query contract without admin", func(t *testing.T) {
		// Create a different contract without admin
		contractAddr2 := sdk.AccAddress("contract2___________")

		req := &types.QueryContractAdminRequest{
			Address: contractAddr2.String(),
		}

		resp, err := queryServer.ContractAdmin(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Empty(t, resp.Admin)
	})

	t.Run("failure - empty request", func(t *testing.T) {
		_, err := queryServer.ContractAdmin(ctx, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty request")
	})

	t.Run("failure - empty address", func(t *testing.T) {
		req := &types.QueryContractAdminRequest{
			Address: "",
		}

		_, err := queryServer.ContractAdmin(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("failure - invalid address", func(t *testing.T) {
		req := &types.QueryContractAdminRequest{
			Address: "invalid_address",
		}

		_, err := queryServer.ContractAdmin(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid address")
	})

	t.Run("success - admin persists after update", func(t *testing.T) {
		// Update admin
		newAdmin := sdk.AccAddress("newadmin____________")
		err := k.SetContractAdmin(ctx, contractAddr, newAdmin)
		require.NoError(t, err)

		// Query admin
		req := &types.QueryContractAdminRequest{
			Address: contractAddr.String(),
		}

		resp, err := queryServer.ContractAdmin(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, newAdmin.String(), resp.Admin)
	})

	t.Run("success - admin removed after delete", func(t *testing.T) {
		// Delete admin
		err := k.DeleteContractAdmin(ctx, contractAddr)
		require.NoError(t, err)

		// Query admin
		req := &types.QueryContractAdminRequest{
			Address: contractAddr.String(),
		}

		resp, err := queryServer.ContractAdmin(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Empty(t, resp.Admin)
	})
}
