// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/privacy/types"
)

// testAddr creates a valid bech32 address from a seed string
func testAddr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	addr, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	if err != nil {
		panic(err)
	}
	return addr
}

// ============================================================================
// Benchmark Setup Helper
// ============================================================================

// setupPrivacyBenchmark creates a keeper and context for benchmarking
func setupPrivacyBenchmark(b *testing.B) (*Keeper, sdk.Context) {
	b.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		b.Fatal(err)
	}

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := NewKeeper(
		cdc,
		storeKey,
		nil, // accountKeeper
		nil, // bankKeeper
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, log.NewNopLogger())

	// Set default params with privacy features enabled
	params := types.DefaultParams()
	params.EnableRingSignatures = true
	params.EnableZkProofs = true
	params.EnableStealthAddresses = true
	params.EnableConfidentialTransactions = true
	params.MinRingSize = 4
	params.MaxRingSize = 64
	_ = k.SetParams(ctx, params)

	return k, ctx
}

// generateTestKeyPair generates a test ECDSA key pair
func generateTestKeyPair(b *testing.B) *ecdsa.PrivateKey {
	b.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		b.Fatalf("Failed to generate key: %v", err)
	}
	return key
}

// generateTestPublicKeys generates n test public keys
func generateTestPublicKeys(b *testing.B, n int) [][]byte {
	b.Helper()
	curve := elliptic.P256()
	keys := make([][]byte, n)
	for i := 0; i < n; i++ {
		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			b.Fatalf("Failed to generate key %d: %v", i, err)
		}
		keys[i] = elliptic.Marshal(curve, key.PublicKey.X, key.PublicKey.Y)
	}
	return keys
}

// ============================================================================
// Ring Signature Benchmarks
// ============================================================================

// BenchmarkGenerateRingSignature benchmarks ring signature generation
func BenchmarkGenerateRingSignature(b *testing.B) {
	ringSizes := []struct {
		name string
		size int
	}{
		{"RingSize_4", 4},
		{"RingSize_8", 8},
		{"RingSize_16", 16},
		{"RingSize_32", 32},
	}

	for _, rs := range ringSizes {
		b.Run(rs.name, func(b *testing.B) {
			k, ctx := setupPrivacyBenchmark(b)

			// Generate public keys for ring
			publicKeys := generateTestPublicKeys(b, rs.size)

			// Generate signer's key pair
			signerKey := generateTestKeyPair(b)
			curve := elliptic.P256()
			publicKeys[0] = elliptic.Marshal(curve, signerKey.PublicKey.X, signerKey.PublicKey.Y)

			message := []byte("test message for ring signature")

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = k.GenerateRingSignature(ctx, message, publicKeys, signerKey, 0)
			}
		})
	}
}

// BenchmarkVerifyRingSignature benchmarks ring signature verification
func BenchmarkVerifyRingSignature(b *testing.B) {
	ringSizes := []struct {
		name string
		size int
	}{
		{"RingSize_4", 4},
		{"RingSize_8", 8},
		{"RingSize_16", 16},
		{"RingSize_32", 32},
	}

	for _, rs := range ringSizes {
		b.Run(rs.name, func(b *testing.B) {
			k, ctx := setupPrivacyBenchmark(b)

			// Generate public keys for ring
			publicKeys := generateTestPublicKeys(b, rs.size)

			// Generate signer's key pair
			signerKey := generateTestKeyPair(b)
			curve := elliptic.P256()
			publicKeys[0] = elliptic.Marshal(curve, signerKey.PublicKey.X, signerKey.PublicKey.Y)

			message := []byte("test message for ring signature")

			// Generate signature once
			sig, err := k.GenerateRingSignature(ctx, message, publicKeys, signerKey, 0)
			if err != nil {
				b.Fatalf("Failed to generate signature: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = k.verifyRingSignatureCrypto(sig)
			}
		})
	}
}

// BenchmarkRingMemberSelection benchmarks random ring member selection
func BenchmarkRingMemberSelection(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	// Add ring members to the pool
	curve := elliptic.P256()
	for i := 0; i < 100; i++ {
		key, _ := ecdsa.GenerateKey(curve, rand.Reader)
		pubKey := elliptic.Marshal(curve, key.PublicKey.X, key.PublicKey.Y)
		_ = k.AddRingMember(ctx, pubKey)
	}

	ringSizes := []int{8, 16, 32}

	for _, size := range ringSizes {
		b.Run("Size_"+string(rune('0'+size/10))+string(rune('0'+size%10)), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = k.GetRingMembers(ctx, size)
			}
		})
	}
}

// BenchmarkKeyImageCheck benchmarks key image existence check
func BenchmarkKeyImageCheck(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	// Generate and store some key images
	for i := 0; i < 1000; i++ {
		keyImage := make([]byte, 65)
		_, _ = rand.Read(keyImage)
		_ = k.StoreKeyImage(ctx, keyImage)
	}

	// Generate key image for lookup
	testKeyImage := make([]byte, 65)
	_, _ = rand.Read(testKeyImage)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.KeyImageExists(ctx, testKeyImage)
	}
}

// ============================================================================
// Pedersen Commitment Benchmarks
// ============================================================================

// BenchmarkCreatePedersenCommitment benchmarks Pedersen commitment creation
func BenchmarkCreatePedersenCommitment(b *testing.B) {
	k, _ := setupPrivacyBenchmark(b)

	value := math.NewInt(1000000)
	blindingFactor := make([]byte, 32)
	_, _ = rand.Read(blindingFactor)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.CreatePedersenCommitment(value, blindingFactor)
	}
}

// BenchmarkVerifyPedersenCommitment benchmarks Pedersen commitment verification
func BenchmarkVerifyPedersenCommitment(b *testing.B) {
	k, _ := setupPrivacyBenchmark(b)

	value := math.NewInt(1000000)
	blindingFactor := make([]byte, 32)
	_, _ = rand.Read(blindingFactor)

	commitment := k.CreatePedersenCommitment(value, blindingFactor)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.VerifyPedersenCommitment(commitment, value, blindingFactor)
	}
}

// BenchmarkAggregateCommitments benchmarks commitment aggregation
func BenchmarkAggregateCommitments(b *testing.B) {
	k, _ := setupPrivacyBenchmark(b)

	sizes := []struct {
		name  string
		count int
	}{
		{"2_commitments", 2},
		{"4_commitments", 4},
		{"8_commitments", 8},
		{"16_commitments", 16},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			commitments := make([][]byte, size.count)
			for i := 0; i < size.count; i++ {
				value := math.NewInt(int64(i * 1000))
				blindingFactor := make([]byte, 32)
				_, _ = rand.Read(blindingFactor)
				commitments[i] = k.CreatePedersenCommitment(value, blindingFactor)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = k.AggregateCommitments(commitments)
			}
		})
	}
}

// ============================================================================
// Range Proof Benchmarks
// ============================================================================

// BenchmarkGenerateRangeProof benchmarks range proof generation
func BenchmarkGenerateRangeProof(b *testing.B) {
	k, _ := setupPrivacyBenchmark(b)

	sizes := []struct {
		name  string
		count int
	}{
		{"1_value", 1},
		{"2_values", 2},
		{"4_values", 4},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			values := make([]math.Int, size.count)
			blindingFactors := make([][]byte, size.count)
			for i := 0; i < size.count; i++ {
				values[i] = math.NewInt(int64((i + 1) * 1000))
				blindingFactors[i] = make([]byte, 32)
				_, _ = rand.Read(blindingFactors[i])
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = k.GenerateRangeProof(values, blindingFactors)
			}
		})
	}
}

// BenchmarkVerifyRangeProof benchmarks range proof verification
func BenchmarkVerifyRangeProof(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	values := []math.Int{math.NewInt(1000), math.NewInt(2000)}
	blindingFactors := make([][]byte, 2)
	commitments := make([][]byte, 2)

	for i := 0; i < 2; i++ {
		blindingFactors[i] = make([]byte, 32)
		_, _ = rand.Read(blindingFactors[i])
		commitments[i] = k.CreatePedersenCommitment(values[i], blindingFactors[i])
	}

	proof, _ := k.GenerateRangeProof(values, blindingFactors)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.VerifyRangeProof(ctx, proof, commitments)
	}
}

// ============================================================================
// Confidential Transaction Benchmarks
// ============================================================================

// BenchmarkValidateConfidentialTransaction benchmarks confidential tx validation
func BenchmarkValidateConfidentialTransaction(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	// Create test confidential transaction
	ctxTx := &ConfidentialTransaction{
		InputCommitments:  make([][]byte, 2),
		OutputCommitments: make([][]byte, 2),
		RangeProof:        make([]byte, 64),
		Signature:         []byte("valid_signature"),
		Fee:               math.NewInt(1000),
	}

	// Generate commitments
	for i := 0; i < 2; i++ {
		value := math.NewInt(int64((i + 1) * 1000000))
		blindingFactor := make([]byte, 32)
		_, _ = rand.Read(blindingFactor)
		ctxTx.InputCommitments[i] = k.CreatePedersenCommitment(value, blindingFactor)
		ctxTx.OutputCommitments[i] = k.CreatePedersenCommitment(value, blindingFactor)

		// Store input commitments
		store := k.getStore(ctx)
		key := append(types.CommitmentPrefix, ctxTx.InputCommitments[i]...)
		store.Set(key, ctxTx.InputCommitments[i])
	}

	// Generate range proof
	values := []math.Int{math.NewInt(1000000), math.NewInt(2000000)}
	blindingFactors := make([][]byte, 2)
	for i := range blindingFactors {
		blindingFactors[i] = make([]byte, 32)
		_, _ = rand.Read(blindingFactors[i])
	}
	ctxTx.RangeProof, _ = k.GenerateRangeProof(values, blindingFactors)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.ValidateConfidentialTransaction(ctx, ctxTx)
	}
}

// BenchmarkVerifyBalance benchmarks balance verification
func BenchmarkVerifyBalance(b *testing.B) {
	k, _ := setupPrivacyBenchmark(b)

	ctxTx := &ConfidentialTransaction{
		InputCommitments:  make([][]byte, 2),
		OutputCommitments: make([][]byte, 2),
		Fee:               math.NewInt(1000),
	}

	for i := 0; i < 2; i++ {
		ctxTx.InputCommitments[i] = make([]byte, 32)
		ctxTx.OutputCommitments[i] = make([]byte, 32)
		_, _ = rand.Read(ctxTx.InputCommitments[i])
		_, _ = rand.Read(ctxTx.OutputCommitments[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.VerifyBalance(ctxTx)
	}
}

// ============================================================================
// Stealth Address Benchmarks
// ============================================================================

// BenchmarkHashToPoint benchmarks hash-to-point operations
func BenchmarkHashToPoint(b *testing.B) {
	curve := elliptic.P256()
	key, _ := ecdsa.GenerateKey(curve, rand.Reader)
	pubKey := elliptic.Marshal(curve, key.PublicKey.X, key.PublicKey.Y)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = hashToPoint(curve, pubKey)
	}
}

// BenchmarkGenerateSyntheticRingMembers benchmarks synthetic member generation
func BenchmarkGenerateSyntheticRingMembers(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	sizes := []struct {
		name  string
		count int
	}{
		{"4_members", 4},
		{"8_members", 8},
		{"16_members", 16},
		{"32_members", 32},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = k.generateSyntheticRingMembers(ctx, size.count)
			}
		})
	}
}

// ============================================================================
// ZK Proof Benchmarks
// ============================================================================

// BenchmarkSubmitZKProof benchmarks ZK proof submission
func BenchmarkSubmitZKProof(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	prover := testAddr("prover")
	proof := make([]byte, 256)
	_, _ = rand.Read(proof)
	publicInputs := make([]byte, 64)
	_, _ = rand.Read(publicInputs)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.SubmitZKProof(ctx, prover, proof, publicInputs)
	}
}

// BenchmarkVerifyZKProof benchmarks ZK proof verification
func BenchmarkVerifyZKProof(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	prover := testAddr("prover")
	proof := make([]byte, 256)
	_, _ = rand.Read(proof)
	publicInputs := make([]byte, 64)
	_, _ = rand.Read(publicInputs)

	proofID, _ := k.SubmitZKProof(ctx, prover, proof, publicInputs)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.VerifyZKProof(ctx, proofID)
	}
}

// ============================================================================
// Nullifier Benchmarks
// ============================================================================

// BenchmarkCreateNullifier benchmarks nullifier creation
func BenchmarkCreateNullifier(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		nullifier := make([]byte, 32)
		nullifier[0] = byte(i >> 24)
		nullifier[1] = byte(i >> 16)
		nullifier[2] = byte(i >> 8)
		nullifier[3] = byte(i)
		_ = k.CreateNullifier(ctx, nullifier)
	}
}

// BenchmarkNullifierExists benchmarks nullifier existence check
func BenchmarkNullifierExists(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	// Create many nullifiers
	for i := 0; i < 10000; i++ {
		nullifier := make([]byte, 32)
		nullifier[0] = byte(i >> 24)
		nullifier[1] = byte(i >> 16)
		nullifier[2] = byte(i >> 8)
		nullifier[3] = byte(i)
		_ = k.CreateNullifier(ctx, nullifier)
	}

	// Test nullifier for lookup (not in set)
	testNullifier := make([]byte, 32)
	testNullifier[0] = 0xFF

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.NullifierExists(ctx, testNullifier)
	}
}

// ============================================================================
// Merkle Tree Benchmarks
// ============================================================================

// BenchmarkAddLeaf benchmarks merkle tree leaf addition
func BenchmarkAddLeaf(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	leaf := make([]byte, 32)
	_, _ = rand.Read(leaf)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.AddLeaf(ctx, leaf)
	}
}

// BenchmarkGetMerkleRoot benchmarks merkle root calculation
func BenchmarkGetMerkleRoot(b *testing.B) {
	sizes := []struct {
		name  string
		count int
	}{
		{"10_leaves", 10},
		{"100_leaves", 100},
		{"1000_leaves", 1000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			k, ctx := setupPrivacyBenchmark(b)

			// Add leaves
			for i := 0; i < size.count; i++ {
				leaf := make([]byte, 32)
				leaf[0] = byte(i >> 24)
				leaf[1] = byte(i >> 16)
				leaf[2] = byte(i >> 8)
				leaf[3] = byte(i)
				_, _ = k.AddLeaf(ctx, leaf)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = k.GetMerkleRoot(ctx)
			}
		})
	}
}

// ============================================================================
// Commitment Benchmarks
// ============================================================================

// BenchmarkCreateCommitment benchmarks commitment creation
func BenchmarkCreateCommitment(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	sender := testAddr("sender")
	commitment := make([]byte, 32)
	_, _ = rand.Read(commitment)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.CreateCommitment(ctx, sender, commitment)
	}
}

// BenchmarkGetCommitment benchmarks commitment retrieval
func BenchmarkGetCommitment(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	sender := testAddr("sender")
	commitment := make([]byte, 32)
	_, _ = rand.Read(commitment)

	commitmentID, _ := k.CreateCommitment(ctx, sender, commitment)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.GetCommitment(ctx, commitmentID)
	}
}

// ============================================================================
// Shielded Transfer Benchmarks
// ============================================================================

// BenchmarkShieldedTransfer benchmarks shielded transfer creation
func BenchmarkShieldedTransfer(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	sender := testAddr("sender")
	amount := math.NewInt(1000000)
	commitment := make([]byte, 32)
	_, _ = rand.Read(commitment)
	proof := make([]byte, 256)
	_, _ = rand.Read(proof)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use different commitments for each iteration
		commitment[0] = byte(i >> 24)
		commitment[1] = byte(i >> 16)
		commitment[2] = byte(i >> 8)
		commitment[3] = byte(i)
		_, _ = k.ShieldedTransfer(ctx, sender, amount, commitment, proof)
	}
}

// ============================================================================
// Hashing Benchmarks
// ============================================================================

// BenchmarkSHA256 benchmarks SHA256 hashing (baseline)
func BenchmarkSHA256(b *testing.B) {
	data := make([]byte, 64)
	_, _ = rand.Read(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sha256.Sum256(data)
	}
}

// BenchmarkComputeChallenge benchmarks challenge computation
func BenchmarkComputeChallenge(b *testing.B) {
	curve := elliptic.P256()
	message := []byte("test message")

	// Generate random curve points
	key1, _ := ecdsa.GenerateKey(curve, rand.Reader)
	key2, _ := ecdsa.GenerateKey(curve, rand.Reader)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = computeChallenge(curve, message,
			key1.PublicKey.X, key1.PublicKey.Y,
			key2.PublicKey.X, key2.PublicKey.Y)
	}
}

// ============================================================================
// Parameter Benchmarks
// ============================================================================

// BenchmarkGetParams benchmarks parameter retrieval
func BenchmarkGetParams(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetParams(ctx)
	}
}

// BenchmarkSetParams benchmarks parameter updates
func BenchmarkSetParams(b *testing.B) {
	k, ctx := setupPrivacyBenchmark(b)
	params := types.DefaultParams()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.SetParams(ctx, params)
	}
}
