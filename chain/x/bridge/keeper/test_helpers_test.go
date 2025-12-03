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

// seedBridgeTransferWithPending seeds both a transfer and pending transfer for fraud proof tests
func seedBridgeTransferWithPending(t *testing.T, input keepertest.TestInput, transferID string, amount string, requiredConfirmations uint64) {
	t.Helper()

	// Create the regular transfer
	seedBridgeTransfer(t, input, transferID, amount, requiredConfirmations)

	// Create the pending transfer with unlock time in the future
	// Use the fraud proof window from default params (7 days)
	unlockTime := input.Ctx.BlockTime().Add(types.DefaultFraudProofWindow)

	pending := &types.PendingTransfer{
		TransferId:   transferID,
		Recipient:    keepertest.GenTestAddr().String(),
		Amount:       amount, // Amount is stored as string
		Denom:        "uaura",
		SourceChain:  "paw",
		SourceTxHash: "0xabcd1234",
		CreatedAt:    timestamppb.New(input.Ctx.BlockTime()),
		UnlockTime:   timestamppb.New(unlockTime),
		Challenged:   false,
		FraudProofId: "",
	}

	store := input.Ctx.KVStore(input.StoreKey)
	store.Set(types.PendingTransferKey(transferID), input.Cdc.MustMarshal(pending))
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
