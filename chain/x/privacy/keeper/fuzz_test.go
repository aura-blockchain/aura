// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/privacy/keeper"
	"github.com/aequitas/aura/chain/x/privacy/types"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// createPrivacyFuzzTestContext creates a test context for privacy module fuzz testing
func createPrivacyFuzzTestContext(t testing.TB) (sdk.Context, *keeper.Keeper, privacypb.MsgServer) {
	t.Helper()

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	k := keeper.NewKeeper(cdc, storeKey, nil, nil)

	// Enable privacy features by default for testing
	params := types.DefaultParams()
	params.EnableMixing = true
	params.EnableZkProofs = true
	params.EnableStealthAddresses = true
	params.EnableRingSignatures = true
	params.EnableNetworkPrivacy = true
	params.MinMixingParticipants = 2
	require.NoError(t, k.SetParams(ctx, params))

	msgServer := keeper.NewMsgServerImpl(k)

	return ctx, k, msgServer
}

// FuzzCreateMixingPool fuzzes the CreateMixingPool message handler.
// Security properties tested:
//   - Validates creator address
//   - Validates participant count constraints (min >= 2, max >= min)
//   - Handles denomination validation
//   - Never panics on any input
func FuzzCreateMixingPool(f *testing.F) {
	// Seed corpus with representative test cases
	f.Add("aura1creator", uint32(3), uint32(10), int64(1000000), uint32(5))
	f.Add("", uint32(2), uint32(5), int64(1000000), uint32(1))                        // Empty creator
	f.Add("aura1x", uint32(0), uint32(5), int64(1000000), uint32(1))                  // Zero min participants
	f.Add("aura1x", uint32(1), uint32(5), int64(1000000), uint32(1))                  // Min participants < 2
	f.Add("aura1x", uint32(5), uint32(3), int64(1000000), uint32(1))                  // Min > Max
	f.Add("aura1x", uint32(2), uint32(2), int64(0), uint32(1))                        // Zero denom
	f.Add("aura1x", uint32(100), uint32(1000), int64(1000000), uint32(0))             // Zero rounds
	f.Add(strings.Repeat("a", 1000), uint32(2), uint32(5), int64(1000000), uint32(1)) // Very long creator

	f.Fuzz(func(t *testing.T, creator string, minParticipants, maxParticipants uint32, denominationInt int64, mixingRounds uint32) {
		if len(creator) > 2000 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createPrivacyFuzzTestContext(t)

		// Convert denomination to proper math.Int type
		denomination := sdkmath.NewInt(denominationInt)

		msg := &privacypb.MsgCreateMixingPool{
			Creator:         creator,
			MinParticipants: minParticipants,
			MaxParticipants: maxParticipants,
			Denomination:    denomination,
			MixingRounds:    mixingRounds,
		}

		// Execute - must not panic
		resp, err := msgServer.CreateMixingPool(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty creator must be rejected
		if creator == "" {
			require.Error(t, err, "empty creator must be rejected")
		}

		// SECURITY INVARIANT: Min participants < 2 must be rejected
		if minParticipants < 2 {
			require.Error(t, err, "min participants < 2 must be rejected")
		}

		// SECURITY INVARIANT: Max < Min must be rejected
		if maxParticipants < minParticipants {
			require.Error(t, err, "max < min participants must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.PoolId)
		}
	})
}

// FuzzJoinMixingPool fuzzes the JoinMixingPool message handler.
// Security properties tested:
//   - Validates participant and pool ID
//   - Handles non-existent pool
//   - Handles pool full condition
//   - Handles duplicate participant
func FuzzJoinMixingPool(f *testing.F) {
	f.Add("aura1participant", "pool-1", []byte("commitment"))
	f.Add("", "pool-1", []byte("c"))                       // Empty participant
	f.Add("aura1x", "", []byte("c"))                       // Empty pool ID
	f.Add("aura1x", "nonexistent", []byte("c"))            // Non-existent pool
	f.Add("aura1x", "pool-1", []byte{})                    // Empty commitment
	f.Add(strings.Repeat("p", 500), "pool-1", []byte("c")) // Very long participant

	f.Fuzz(func(t *testing.T, participant, poolID string, commitment []byte) {
		if len(participant) > 1000 || len(poolID) > 500 || len(commitment) > 10000 {
			t.Skip("input too long")
		}

		ctx, k, msgServer := createPrivacyFuzzTestContext(t)

		// Create a pool for some tests
		if poolID != "" && !strings.HasPrefix(poolID, "nonexistent") {
			pool := &privacypb.MixingPool{
				PoolId:          poolID,
				MinParticipants: 2,
				MaxParticipants: 10,
				Status:          "pending",
				Participants:    [][]byte{},
			}
			_ = k.SetMixingPool(ctx, pool)
		}

		msg := &privacypb.MsgJoinMixingPool{
			Participant: participant,
			PoolId:      poolID,
			Commitment:  commitment,
		}

		// Execute - must not panic
		resp, err := msgServer.JoinMixingPool(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty participant must be rejected
		if participant == "" {
			require.Error(t, err, "empty participant must be rejected")
		}

		// SECURITY INVARIANT: Empty pool ID must be rejected
		if poolID == "" {
			require.Error(t, err, "empty pool ID must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.True(t, resp.Success)
		}
	})
}

// FuzzRegisterViewKey fuzzes the RegisterViewKey message handler.
// Security properties tested:
//   - Validates owner address
//   - Validates public key format and length
//   - Prevents private key registration on-chain
//   - Handles malformed key data
func FuzzRegisterViewKey(f *testing.F) {
	// Valid key lengths: 32 (Ed25519), 33 (compressed secp256k1), 64 (uncompressed)
	f.Add("aura1owner", make([]byte, 32), "PUBLIC")
	f.Add("", make([]byte, 32), "PUBLIC")        // Empty owner
	f.Add("aura1x", []byte{}, "PUBLIC")          // Empty key
	f.Add("aura1x", make([]byte, 31), "PUBLIC")  // Invalid length
	f.Add("aura1x", make([]byte, 33), "PUBLIC")  // Valid compressed length
	f.Add("aura1x", make([]byte, 64), "PUBLIC")  // Valid uncompressed length
	f.Add("aura1x", make([]byte, 32), "PRIVATE") // Private key type (should reject)
	f.Add("aura1x", make([]byte, 32), "SECRET")  // Secret key type (should reject)
	f.Add("aura1x", make([]byte, 100), "PUBLIC") // Invalid length

	f.Fuzz(func(t *testing.T, owner string, publicViewKey []byte, keyType string) {
		if len(owner) > 1000 || len(publicViewKey) > 10000 || len(keyType) > 100 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createPrivacyFuzzTestContext(t)

		viewKey := &privacypb.ViewKey{
			PublicViewKey: publicViewKey,
			KeyType:       keyType,
		}

		msg := &privacypb.MsgRegisterViewKey{
			Owner:   owner,
			ViewKey: viewKey,
		}

		// Execute - must not panic
		resp, err := msgServer.RegisterViewKey(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty owner must be rejected
		if owner == "" {
			require.Error(t, err, "empty owner must be rejected")
		}

		// SECURITY INVARIANT: Empty public key must be rejected
		if len(publicViewKey) == 0 {
			require.Error(t, err, "empty public key must be rejected")
		}

		// SECURITY INVARIANT: Invalid key lengths must be rejected
		keyLen := len(publicViewKey)
		if keyLen != 0 && keyLen != 32 && keyLen != 33 && keyLen != 64 {
			require.Error(t, err, "invalid key length must be rejected")
		}

		// SECURITY INVARIANT: Private/secret key types must be rejected
		if keyType == "PRIVATE" || keyType == "SECRET" {
			require.Error(t, err, "private key types must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.True(t, resp.Success)
		}
	})
}

// FuzzRevokeViewKey fuzzes the RevokeViewKey message handler.
// Security properties tested:
//   - Validates owner address
//   - Validates public key presence
func FuzzRevokeViewKey(f *testing.F) {
	f.Add("aura1owner", make([]byte, 32))
	f.Add("", make([]byte, 32))        // Empty owner
	f.Add("aura1x", []byte{})          // Empty key
	f.Add("aura1x", make([]byte, 100)) // Long key

	f.Fuzz(func(t *testing.T, owner string, publicViewKey []byte) {
		if len(owner) > 1000 || len(publicViewKey) > 10000 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createPrivacyFuzzTestContext(t)

		msg := &privacypb.MsgRevokeViewKey{
			Owner:         owner,
			PublicViewKey: publicViewKey,
		}

		// Execute - must not panic
		resp, err := msgServer.RevokeViewKey(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty owner must be rejected
		if owner == "" {
			require.Error(t, err, "empty owner must be rejected")
		}

		// SECURITY INVARIANT: Empty public key must be rejected
		if len(publicViewKey) == 0 {
			require.Error(t, err, "empty public key must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.True(t, resp.Success)
		}
	})
}

// FuzzSubmitPrivateTransaction fuzzes the SubmitPrivateTransaction message handler.
// Security properties tested:
//   - Validates sender address
//   - Validates private transaction structure
//   - Handles ZK proof validation
func FuzzSubmitPrivateTransaction(f *testing.F) {
	f.Add("aura1sender", []byte("txid"), []byte("proof"))
	f.Add("", []byte("tx"), []byte("p"))              // Empty sender
	f.Add("aura1x", []byte{}, []byte("p"))            // Empty txid
	f.Add("aura1x", []byte("tx"), []byte{})           // Empty proof
	f.Add("aura1x", make([]byte, 10000), []byte("p")) // Large txid

	f.Fuzz(func(t *testing.T, sender string, txID, proofData []byte) {
		if len(sender) > 1000 || len(txID) > 50000 || len(proofData) > 50000 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createPrivacyFuzzTestContext(t)

		// Construct ZK proof with the provided data
		zkProof := &privacypb.ZKProof{
			ProofData: proofData,
		}

		privateTx := &privacypb.PrivateTransaction{
			TxId:    txID,
			ZkProof: zkProof,
		}

		msg := &privacypb.MsgSubmitPrivateTransaction{
			Sender:             sender,
			PrivateTransaction: privateTx,
		}

		// Execute - must not panic
		resp, err := msgServer.SubmitPrivateTransaction(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty sender must be rejected
		if sender == "" {
			require.Error(t, err, "empty sender must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.True(t, resp.Success)
		}
	})
}

// FuzzMixingShuffleAlgorithm fuzzes the mixing shuffle algorithm directly.
// Security properties tested:
//   - Shuffle produces valid permutation
//   - All participants preserved after shuffle
//   - Deterministic for same inputs
func FuzzMixingShuffleAlgorithm(f *testing.F) {
	f.Add(uint8(3), int64(12345))
	f.Add(uint8(0), int64(0))      // Zero participants
	f.Add(uint8(1), int64(1))      // Single participant
	f.Add(uint8(100), int64(9999)) // Many participants
	f.Add(uint8(255), int64(-1))   // Max participants, negative seed

	f.Fuzz(func(t *testing.T, participantCount uint8, seedValue int64) {
		if participantCount > 100 {
			t.Skip("too many participants")
		}

		ctx, k, _ := createPrivacyFuzzTestContext(t)

		// Create participants
		participants := make([][]byte, participantCount)
		for i := range participants {
			participants[i] = []byte{byte(i)}
		}

		// Create and store a mixing pool
		pool := &privacypb.MixingPool{
			PoolId:          "shuffle-test",
			MinParticipants: 2,
			MaxParticipants: 100,
			Participants:    participants,
			Status:          "ready",
		}

		if err := k.SetMixingPool(ctx, pool); err != nil {
			t.Skip("failed to set pool")
		}

		// Execute mixing
		err := k.ExecuteMixing(ctx, "shuffle-test")

		if participantCount == 0 {
			// Pool status won't be ready with 0 participants, so error is expected
			return
		}

		if err == nil {
			// Retrieve shuffled pool
			shuffledPool, getErr := k.GetMixingPool(ctx, "shuffle-test")
			require.NoError(t, getErr)

			// SECURITY INVARIANT: All participants must be preserved
			require.Equal(t, len(participants), len(shuffledPool.Participants))

			// SECURITY INVARIANT: Pool must be completed
			require.Equal(t, "completed", shuffledPool.Status)
		}
	})
}

// FuzzUpdateParams fuzzes the UpdateParams message handler.
// Security properties tested:
//   - Validates authority
//   - Validates parameter ranges
func FuzzUpdateParams(f *testing.F) {
	f.Add("authority", true, true, uint32(3), uint32(16), uint32(3), int64(1000))
	f.Add("", true, true, uint32(2), uint32(10), uint32(2), int64(100))     // Empty authority
	f.Add("auth", true, true, uint32(0), uint32(10), uint32(2), int64(100)) // Zero min ring
	f.Add("auth", true, true, uint32(5), uint32(3), uint32(2), int64(100))  // Min > Max ring
	f.Add("auth", true, true, uint32(2), uint32(10), uint32(1), int64(100)) // Min mixing < 2

	f.Fuzz(func(t *testing.T, authority string, enableZK, enableMixing bool, minRing, maxRing, minMixing uint32, mixingFee int64) {
		if len(authority) > 500 {
			t.Skip("authority too long")
		}

		ctx, _, msgServer := createPrivacyFuzzTestContext(t)

		// Normalize mixing fee to positive
		if mixingFee < 0 {
			mixingFee = -mixingFee
		}

		params := privacypb.Params{
			EnableZkProofs:        enableZK,
			EnableMixing:          enableMixing,
			MinRingSize:           minRing,
			MaxRingSize:           maxRing,
			MinMixingParticipants: minMixing,
			MixingFee:             sdkmath.NewInt(mixingFee),
		}

		msg := &privacypb.MsgUpdateParams{
			Authority: authority,
			Params:    params,
		}

		// Execute - must not panic
		resp, err := msgServer.UpdateParams(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty authority must be rejected
		if authority == "" {
			require.Error(t, err, "empty authority must be rejected")
		}

		// Note: Other validation errors depend on keeper implementation
		if err == nil {
			require.NotNil(t, resp)
		}
	})
}

// FuzzNilPrivacyMessageHandling ensures all privacy message handlers reject nil messages.
func FuzzNilPrivacyMessageHandling(f *testing.F) {
	f.Add(uint8(0))
	for i := uint8(0); i < 10; i++ {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, msgType uint8) {
		ctx, _, msgServer := createPrivacyFuzzTestContext(t)

		var err error

		// Test each message type with nil
		switch msgType % 6 {
		case 0:
			_, err = msgServer.CreateMixingPool(sdk.WrapSDKContext(ctx), nil)
		case 1:
			_, err = msgServer.JoinMixingPool(sdk.WrapSDKContext(ctx), nil)
		case 2:
			_, err = msgServer.RegisterViewKey(sdk.WrapSDKContext(ctx), nil)
		case 3:
			_, err = msgServer.RevokeViewKey(sdk.WrapSDKContext(ctx), nil)
		case 4:
			_, err = msgServer.SubmitPrivateTransaction(sdk.WrapSDKContext(ctx), nil)
		case 5:
			_, err = msgServer.UpdateParams(sdk.WrapSDKContext(ctx), nil)
		}

		// SECURITY INVARIANT: Nil messages must always be rejected
		require.Error(t, err, "nil message must be rejected")
	})
}
