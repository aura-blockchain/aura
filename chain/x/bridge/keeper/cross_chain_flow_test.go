package keeper_test

import (
	"crypto/sha256"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// TestCrossChainTransferFlowWithMerkleProof simulates a complete bridge lock flow
// and proves the recorded transfer via the module's Merkle proof helpers. This
// mirrors the roadmap requirement to exercise a cross-chain transfer plus Merkle verification.
func TestCrossChainTransferFlowWithMerkleProof(t *testing.T) {
	input := keepertest.CreateTestInput(t)

	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge").
		WithKeyTable(types.ParamKeyTable())

	k := keeper.NewKeeper(input.Cdc, input.StoreKey, &ps, nil, nil, nil)
	ctx := input.Ctx

	params := types.DefaultParams()
	params.MinConfirmations = 2
	require.NoError(t, k.SetParams(ctx, params))

	require.NoError(t, k.AddSupportedChain(ctx, types.ChainConfig{
		ChainId:          "paw",
		ChainName:        "PAW Testnet",
		AddressPrefix:    "paw",
		RpcEndpoint:      "http://localhost:8545",
		BridgeContract:   "0xbridge",
		MinConfirmations: 2,
		NativeDenom:      "upaw",
		Enabled:          true,
	}))

	msgServer := keeper.NewMsgServerImpl(k)
	sender := keepertest.GenTestAddr().String()

	lock := func(denom string, amt int64, recipient string) string {
		coin := sdk.NewCoin(denom, sdkmath.NewInt(amt))
		resp, err := msgServer.LockTokens(sdk.WrapSDKContext(ctx), &types.MsgLockTokens{
			Sender:      sender,
			Recipient:   recipient,
			TargetChain: "paw",
			Amount:      &coin,
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp.TransferId)
		return resp.TransferId
	}

	transferID := lock("uaura", 5_000_000, "paw1qpqhemockrecipient")
	_ = lock("uaura", 2_500_000, "paw1additionalrecipient")

	validators := keepertest.GenTestAddrs(2)
	for _, val := range validators {
		require.NoError(t, k.SubmitAttestation(ctx, transferID, val.String(), true))
	}
	require.True(t, k.CheckAttestationThreshold(ctx, transferID))

	transfer := getBridgeTransfer(t, input, transferID)
	require.Equal(t, uint64(2), transfer.Confirmations)
	require.Len(t, transfer.ValidatorSignatures, 2)

	exported := k.ExportGenesis(ctx)
	require.GreaterOrEqual(t, len(exported.Transfers), 2)

	leaves := make([][]byte, len(exported.Transfers))
	targetIdx := -1
	for i, tf := range exported.Transfers {
		leaves[i] = input.Cdc.MustMarshal(tf)
		if tf.TransferId == transferID {
			targetIdx = i
		}
	}
	require.NotEqual(t, -1, targetIdx, "target transfer not found in exported state")

	root := keeper.ComputeMerkleRoot(leaves)
	require.NotEmpty(t, root)

	proof, err := keeper.GenerateMerkleProof(leaves, targetIdx)
	require.NoError(t, err)
	require.Equal(t, root, proof.Root)
	require.True(t, keeper.VerifyMerkleProof(proof))

	expectedLeaf := sha256.Sum256(leaves[targetIdx])
	require.Equal(t, expectedLeaf[:], proof.Leaf, "leaf hash must match proof payload")
}
