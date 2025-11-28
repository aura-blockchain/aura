package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func seedBridgeTransfer(t *testing.T, input keepertest.TestInput, transferID string, amount string, requiredConfirmations uint64) {
	t.Helper()

	transfer := &types.CrossChainTransfer{
		TransferId:            transferID,
		SourceChain:           "aura",
		TargetChain:           "paw",
		Sender:                keepertest.GenTestAddr().String(),
		Recipient:             keepertest.GenTestAddr().String(),
		Denom:                 "uaura",
		Amount:                amount,
		Status:                types.TransferStatus_PENDING,
		Timestamp:             timestamppb.New(input.Ctx.BlockTime()),
		RequiredConfirmations: requiredConfirmations,
	}

	store := input.Ctx.KVStore(input.StoreKey)
	store.Set(types.TransferKey(transferID), input.Cdc.MustMarshal(transfer))
}

func getBridgeTransfer(t *testing.T, input keepertest.TestInput, transferID string) types.CrossChainTransfer {
	t.Helper()
	store := input.Ctx.KVStore(input.StoreKey)
	bz := store.Get(types.TransferKey(transferID))
	require.NotNil(t, bz)
	var transfer types.CrossChainTransfer
	require.NoError(t, input.Cdc.Unmarshal(bz, &transfer))
	return transfer
}
