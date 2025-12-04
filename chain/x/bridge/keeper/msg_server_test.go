package keeper_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

func TestMsgServerLockTokens_Success(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	require.NoError(t, k.AddSupportedChain(ctx, types.ChainConfig{ChainId: "paw", Enabled: true}))

	amount := sdk.NewCoin("uaura", sdkmath.NewInt(1_000))
	msg := &types.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		TargetChain: "paw",
		Recipient:   "paw1recipient",
		Amount:      &amount,
	}

	resp, err := ms.LockTokens(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.NotEmpty(t, resp.TransferId)

	exported := k.ExportGenesis(ctx)
	require.Len(t, exported.Transfers, 1)
	require.Equal(t, "paw", exported.Transfers[0].TargetChain)
}

func TestMsgServerLockTokens_MissingChain(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	amount := sdk.NewCoin("uaura", sdkmath.NewInt(1))
	msg := &types.MsgLockTokens{
		Sender:      keepertest.GenTestAddr().String(),
		TargetChain: "unknown",
		Recipient:   "paw1recipient",
		Amount:      &amount,
	}

	_, err := ms.LockTokens(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err)
}

func TestMsgServerMintTokens_CreatesWrappedToken(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	msg := &types.MsgMintTokens{
		Validator:    keepertest.GenTestAddr().String(),
		SourceChain:  "paw",
		SourceTxHash: "0xabc",
		Recipient:    keepertest.GenTestAddr().String(),
		Amount:       "1000",
		Denom:        "paw.token",
	}

	resp, err := ms.MintTokens(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.True(t, resp.Success)

	exported := k.ExportGenesis(ctx)
	require.Len(t, exported.WrappedTokens, 1)
	require.Equal(t, "paw.paw.token", exported.WrappedTokens[0].WrappedDenom)
}

func TestMsgServerUnlockTokens_CompletesTransfer(t *testing.T) {
	// STEP 1: Setup keeper with minimum 3 confirmations requirement
	minConfirmations := uint64(3)
	input := keepertest.CreateTestInput(t)

	// Register crypto interfaces so secp256k1 keys can be marshaled
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	input.Cdc = cdc

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	k := keeper.NewKeeper(cdc, input.StoreKey, &ps, nil, nil, nil, nil)

	// Set security params requiring 3 validators
	params := types.DefaultParams()
	params.MinConfirmations = minConfirmations
	params.ValidatorThresholdPercentage = 67 // 67% threshold
	k.SetParams(input.Ctx, params)

	ctx := input.Ctx
	msgServer := keeper.NewMsgServerImpl(k)

	// STEP 2: Create 5 validators (all active) using secp256k1 keys
	validators := createMultipleValidators(t, input, k, 5, 5)

	// STEP 3: Create a transfer to unlock
	burnTxHash := "0xabcd1234567890"
	transferID := "transfer-unlock-test"
	sender := keepertest.GenTestAddr().String()
	amount := "1000000"
	denom := "uaura"
	sourceChain := "paw"

	transfer := &bridgepb.CrossChainTransfer{
		TransferId:            transferID,
		Sender:                sender,
		Recipient:             sender,
		Amount:                amount,
		Denom:                 denom,
		SourceChain:           sourceChain,
		TargetChain:           "aura",
		Status:                bridgepb.TransferStatus_PENDING,
		SourceTxHash:          burnTxHash,
		RequiredConfirmations: minConfirmations,
		Confirmations:         0,
		Timestamp:             timestamppb.New(ctx.BlockTime()),
	}
	k.SetTransfer(ctx, transfer)
	k.IndexTransferHash(ctx, burnTxHash, transferID)

	// STEP 4: Build message to sign (must match msg_server.go format)
	msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
		sourceChain, burnTxHash, sender, amount, denom)
	msgHash := sha256.Sum256([]byte(msgToSign))

	// STEP 5: Test with 3 signatures (should succeed)
	t.Run("with_3_signatures_succeeds", func(t *testing.T) {
		var signatures [][]byte
		for i := 0; i < 3; i++ {
			sig, err := validators[i].PrivKey.Sign(msgHash[:])
			require.NoError(t, err)
			signatures = append(signatures, sig)
		}

		unlockMsg := &bridgepb.MsgUnlockTokens{
			BurnTxHash:          burnTxHash,
			Sender:              sender,
			Amount:              amount,
			Denom:               denom,
			SourceChain:         sourceChain,
			ValidatorSignatures: signatures,
			MerkleProof:         []byte{},
			MerkleRoot:          []byte{},
			SourceBlockHash:     []byte{},
			SourceBlockHeight:   0,
		}

		resp, err := msgServer.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)
		require.NoError(t, err, "Unlock should succeed with 3 valid signatures")
		require.NotNil(t, resp)
		require.True(t, resp.Success)
	})

	// STEP 6: Test with <3 signatures (should fail - security requirement)
	t.Run("with_2_signatures_fails", func(t *testing.T) {
		// Create a NEW transfer with different hash for this test
		burnTxHash2 := "0xdifferenthash999"
		transferID2 := "transfer-unlock-test-2"

		transfer2 := &bridgepb.CrossChainTransfer{
			TransferId:            transferID2,
			Sender:                sender,
			Recipient:             sender,
			Amount:                amount,
			Denom:                 denom,
			SourceChain:           sourceChain,
			TargetChain:           "aura",
			Status:                bridgepb.TransferStatus_PENDING,
			SourceTxHash:          burnTxHash2,
			RequiredConfirmations: minConfirmations,
			Confirmations:         0,
			Timestamp:             timestamppb.New(ctx.BlockTime()),
		}
		k.SetTransfer(ctx, transfer2)
		k.IndexTransferHash(ctx, burnTxHash2, transferID2)

		// Build message to sign for the second transfer
		msgToSign2 := fmt.Sprintf("%s:%s:%s:%s:%s",
			sourceChain, burnTxHash2, sender, amount, denom)
		msgHash2 := sha256.Sum256([]byte(msgToSign2))

		var signatures [][]byte
		for i := 0; i < 2; i++ { // Only 2 signatures
			sig, err := validators[i].PrivKey.Sign(msgHash2[:])
			require.NoError(t, err)
			signatures = append(signatures, sig)
		}

		unlockMsg := &bridgepb.MsgUnlockTokens{
			BurnTxHash:          burnTxHash2,
			Sender:              sender,
			Amount:              amount,
			Denom:               denom,
			SourceChain:         sourceChain,
			ValidatorSignatures: signatures,
			MerkleProof:         []byte{},
			MerkleRoot:          []byte{},
			SourceBlockHash:     []byte{},
			SourceBlockHeight:   0,
		}

		_, err := msgServer.UnlockTokens(sdk.WrapSDKContext(ctx), unlockMsg)
		require.Error(t, err, "Should fail with only 2 signatures when 3 required")
		require.Contains(t, err.Error(), "insufficient")
	})
}

// Helper function to create multiple test validators
func createMultipleValidators(t *testing.T, input keepertest.TestInput, k *keeper.Keeper, total int, activeCount int) []testValidator {
	validators := make([]testValidator, total)

	for i := 0; i < total; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address()).String()

		isActive := i < activeCount

		pubKeyAny, err := input.Cdc.MarshalInterface(pubKey)
		require.NoError(t, err)

		bridgeValidator := &types.BridgeValidator{
			Address:   addr,
			PublicKey: pubKeyAny,
			Power:     1000,
			Active:    isActive,
			Chains:    []string{"aura", "paw", "xai"},
		}

		k.SetValidator(input.Ctx, bridgeValidator)

		validators[i] = testValidator{
			Address: addr,
			PrivKey: privKey,
			PubKey:  pubKey,
			Active:  isActive,
		}
	}

	return validators
}

type testValidator struct {
	Address string
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Active  bool
}
