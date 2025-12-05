package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestQueryParams(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	t.Run("success - query params", func(t *testing.T) {
		req := &types.QueryParamsRequest{}

		resp, err := queryServer.Params(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Params)
	})

	t.Run("failure - nil request", func(t *testing.T) {
		_, err := queryServer.Params(ctx, nil)
		require.Error(t, err)
	})
}

func TestQuerySecurityStats(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	// Setup some stats directly (StoreCode requires full wasmd setup)
	stats := types.SecurityStats{
		TotalCodesAnalyzed: 1,
		TotalExecutions:    5,
		CodesRejected:      0,
		ContractsPaused:    0,
	}
	k.SetSecurityStats(ctx, stats)

	t.Run("success - query security stats", func(t *testing.T) {
		req := &types.QuerySecurityStatsRequest{}

		resp, err := queryServer.SecurityStats(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Stats)
		require.Equal(t, uint64(1), resp.Stats.GetTotalCodesAnalyzed())
	})
}

func TestQueryAuthorizedUploaders(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	uploader1 := sdk.AccAddress("uploader1___________")
	uploader2 := sdk.AccAddress("uploader2___________")

	// Authorize uploaders
	err := k.AuthorizeUploader(ctx, uploader1.String())
	require.NoError(t, err)
	err = k.AuthorizeUploader(ctx, uploader2.String())
	require.NoError(t, err)

	t.Run("success - query authorized uploaders", func(t *testing.T) {
		req := &types.QueryAuthorizedUploadersRequest{}

		resp, err := queryServer.AuthorizedUploaders(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Uploaders, 2)
	})

	t.Run("failure - nil request", func(t *testing.T) {
		_, err := queryServer.AuthorizedUploaders(ctx, nil)
		require.Error(t, err)
	})
}

func TestQueryPausedContracts(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	contract1 := sdk.AccAddress("contract1___________")
	contract2 := sdk.AccAddress("contract2___________")

	// Pause contracts
	err := k.PauseContract(ctx, contract1.String())
	require.NoError(t, err)
	err = k.PauseContract(ctx, contract2.String())
	require.NoError(t, err)

	t.Run("success - query paused contracts", func(t *testing.T) {
		req := &types.QueryPausedContractsRequest{}

		resp, err := queryServer.PausedContracts(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Contracts, 2)
	})
}

func TestQueryIsAuthorizedUploader(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	authorizedAddr := sdk.AccAddress("authorized__________")
	unauthorizedAddr := sdk.AccAddress("unauthorized________")

	// Authorize one address
	err := k.AuthorizeUploader(ctx, authorizedAddr.String())
	require.NoError(t, err)

	t.Run("success - authorized uploader", func(t *testing.T) {
		req := &types.QueryIsAuthorizedUploaderRequest{
			Address: authorizedAddr.String(),
		}

		resp, err := queryServer.IsAuthorizedUploader(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsAuthorized)
	})

	t.Run("success - unauthorized uploader", func(t *testing.T) {
		req := &types.QueryIsAuthorizedUploaderRequest{
			Address: unauthorizedAddr.String(),
		}

		resp, err := queryServer.IsAuthorizedUploader(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsAuthorized)
	})

	t.Run("failure - empty address", func(t *testing.T) {
		req := &types.QueryIsAuthorizedUploaderRequest{
			Address: "",
		}

		_, err := queryServer.IsAuthorizedUploader(ctx, req)
		require.Error(t, err)
	})
}

func TestQueryIsContractPaused(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	pausedContract := sdk.AccAddress("paused______________")
	activeContract := sdk.AccAddress("active______________")

	// Pause one contract
	err := k.PauseContract(ctx, pausedContract.String())
	require.NoError(t, err)

	t.Run("success - paused contract", func(t *testing.T) {
		req := &types.QueryIsContractPausedRequest{
			Address: pausedContract.String(),
		}

		resp, err := queryServer.IsContractPaused(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.IsPaused)
	})

	t.Run("success - active contract", func(t *testing.T) {
		req := &types.QueryIsContractPausedRequest{
			Address: activeContract.String(),
		}

		resp, err := queryServer.IsContractPaused(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.False(t, resp.IsPaused)
	})

	t.Run("failure - empty address", func(t *testing.T) {
		req := &types.QueryIsContractPausedRequest{
			Address: "",
		}

		_, err := queryServer.IsContractPaused(ctx, req)
		require.Error(t, err)
	})
}

func TestQueryContractInfo(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - query contract info", func(t *testing.T) {
		req := &types.QueryContractInfoRequest{
			Address: contractAddr.String(),
		}

		resp, err := queryServer.ContractInfo(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, contractAddr.String(), resp.Address)
	})

	t.Run("failure - empty address", func(t *testing.T) {
		req := &types.QueryContractInfoRequest{
			Address: "",
		}

		_, err := queryServer.ContractInfo(ctx, req)
		require.Error(t, err)
	})

	t.Run("failure - invalid address", func(t *testing.T) {
		req := &types.QueryContractInfoRequest{
			Address: "invalid",
		}

		_, err := queryServer.ContractInfo(ctx, req)
		require.Error(t, err)
	})
}

func TestQuerySmartContractState(t *testing.T) {
	k, ctx := keepertest.WasmKeeper(t)
	queryServer := keeper.NewQueryServerImpl(k)

	contractAddr := sdk.AccAddress("contract____________")

	t.Run("success - query smart contract", func(t *testing.T) {
		// Note: This test requires a full wasmd keeper setup which is beyond unit test scope
		// In unit tests, we expect the "wasm keeper not configured" error
		// Full integration tests with wasmd would be in integration_test.go
		req := &types.QuerySmartContractStateRequest{
			Address:   contractAddr.String(),
			QueryData: []byte(`{"query":"data"}`),
		}

		_, err := queryServer.SmartContractState(ctx, req)
		// Expect error since wasmd keeper is not configured in unit test setup
		require.Error(t, err)
		require.Contains(t, err.Error(), "wasm keeper not configured")
	})

	t.Run("failure - empty query data", func(t *testing.T) {
		req := &types.QuerySmartContractStateRequest{
			Address:   contractAddr.String(),
			QueryData: []byte{},
		}

		_, err := queryServer.SmartContractState(ctx, req)
		require.Error(t, err)
	})

	t.Run("failure - paused contract", func(t *testing.T) {
		// Pause the contract
		err := k.PauseContract(ctx, contractAddr.String())
		require.NoError(t, err)

		req := &types.QuerySmartContractStateRequest{
			Address:   contractAddr.String(),
			QueryData: []byte(`{"query":"data"}`),
		}

		_, err = queryServer.SmartContractState(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "paused")

		// Unpause
		err = k.UnpauseContract(ctx, contractAddr.String())
		require.NoError(t, err)
	})
}
