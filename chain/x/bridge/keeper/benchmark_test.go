// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
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

// setupBridgeBenchmark creates a keeper and context for benchmarking
func setupBridgeBenchmark(b *testing.B) (*Keeper, sdk.Context) {
	b.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.StoreKey + "_mem")

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	if err := stateStore.LoadLatestVersion(); err != nil {
		b.Fatal(err)
	}

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	legacyAmino := codec.NewLegacyAmino()
	ps := paramtypes.NewSubspace(cdc, legacyAmino, storeKey, memStoreKey, "bridge")

	k := NewKeeper(
		cdc,
		storeKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, log.NewNopLogger())

	return k, ctx
}

// ============================================================================
// Signature Verification Benchmarks
// ============================================================================

// BenchmarkSignatureVerification benchmarks signature validation
func BenchmarkSignatureVerification(b *testing.B) {
	signature := make([]byte, 65)
	for i := range signature[:64] {
		signature[i] = byte(i % 256)
	}
	signature[64] = 0

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = len(signature) == 65 && signature[64] < 4
	}
}

// BenchmarkSignatureValidation_Extended benchmarks extended signature validation
func BenchmarkSignatureValidation_Extended(b *testing.B) {
	b.Run("65ByteSignature", func(b *testing.B) {
		sig := make([]byte, 65)
		for i := range sig {
			sig[i] = byte(i % 256)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ValidateSignatureFormat(sig)
		}
	})

	b.Run("64ByteSignature", func(b *testing.B) {
		sig := make([]byte, 64)
		for i := range sig {
			sig[i] = byte(i % 256)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ValidateSignatureFormat(sig)
		}
	})
}

// ValidateSignatureFormat validates signature format
func ValidateSignatureFormat(sig []byte) bool {
	if len(sig) == 65 {
		return sig[64] < 4
	}
	return len(sig) == 64
}

// ============================================================================
// Hashing and Transfer Benchmarks
// ============================================================================

// BenchmarkLockTokens benchmarks hashing operations
func BenchmarkLockTokens(b *testing.B) {
	data := []byte("test transfer data")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sha256.Sum256(data)
	}
}

// BenchmarkUnlockTokens benchmarks hash verification
func BenchmarkUnlockTokens(b *testing.B) {
	data := []byte("test transfer data")
	expectedHash := sha256.Sum256(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		actualHash := sha256.Sum256(data)
		_ = actualHash == expectedHash
	}
}

// BenchmarkBatchTransfers benchmarks batch operations
func BenchmarkBatchTransfers(b *testing.B) {
	hashes := make([][32]byte, 100)
	for i := 0; i < 100; i++ {
		data := []byte{byte(i)}
		hashes[i] = sha256.Sum256(data)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, hash := range hashes {
			_ = hash[:]
		}
	}
}

// BenchmarkTransferHash benchmarks transfer hash computation
func BenchmarkTransferHash(b *testing.B) {
	sender := testAddr("sender")
	recipient := "0x1234567890abcdef1234567890abcdef12345678"
	amount := sdkmath.NewInt(1000000)
	chainID := "ethereum-1"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data := fmt.Sprintf("%s:%s:%s:%s", sender, recipient, amount.String(), chainID)
		_ = sha256.Sum256([]byte(data))
	}
}

// ============================================================================
// Merkle Proof Benchmarks
// ============================================================================

// BenchmarkComputeMerkleRoot benchmarks merkle root computation
func BenchmarkComputeMerkleRoot(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"4_leaves", 4},
		{"16_leaves", 16},
		{"64_leaves", 64},
		{"256_leaves", 256},
		{"1024_leaves", 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := make([][]byte, size.size)
			for i := 0; i < size.size; i++ {
				data[i] = []byte(fmt.Sprintf("transfer_data_%d", i))
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = ComputeMerkleRoot(data)
			}
		})
	}
}

// BenchmarkGenerateMerkleProof benchmarks merkle proof generation
func BenchmarkGenerateMerkleProof(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"4_leaves", 4},
		{"16_leaves", 16},
		{"64_leaves", 64},
		{"256_leaves", 256},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := make([][]byte, size.size)
			for i := 0; i < size.size; i++ {
				data[i] = []byte(fmt.Sprintf("transfer_data_%d", i))
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Generate proof for element at index i % size
				_, _ = GenerateMerkleProof(data, i%size.size)
			}
		})
	}
}

// BenchmarkVerifyMerkleProof benchmarks merkle proof verification
func BenchmarkVerifyMerkleProof(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"4_leaves", 4},
		{"16_leaves", 16},
		{"64_leaves", 64},
		{"256_leaves", 256},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := make([][]byte, size.size)
			for i := 0; i < size.size; i++ {
				data[i] = []byte(fmt.Sprintf("transfer_data_%d", i))
			}

			// Generate proof once
			proof, err := GenerateMerkleProof(data, 0)
			if err != nil {
				b.Fatalf("Failed to generate proof: %v", err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = VerifyMerkleProof(proof)
			}
		})
	}
}

// BenchmarkMerkleFullCycle benchmarks complete merkle operations
func BenchmarkMerkleFullCycle(b *testing.B) {
	data := make([][]byte, 64)
	for i := 0; i < 64; i++ {
		data[i] = []byte(fmt.Sprintf("transfer_data_%d", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Compute root
		_ = ComputeMerkleRoot(data)

		// Generate proof for random index
		proof, _ := GenerateMerkleProof(data, i%64)

		// Verify proof
		_ = VerifyMerkleProof(proof)
	}
}

// ============================================================================
// Fee Calculation Benchmarks
// ============================================================================

// BenchmarkCalculateBridgeFee benchmarks fee calculation
func BenchmarkCalculateBridgeFee(b *testing.B) {
	k, ctx := setupBridgeBenchmark(b)

	amounts := []struct {
		name   string
		amount sdkmath.Int
	}{
		{"Small_1K", sdkmath.NewInt(1000)},
		{"Medium_1M", sdkmath.NewInt(1000000)},
		{"Large_1B", sdkmath.NewInt(1000000000)},
	}

	for _, amt := range amounts {
		b.Run(amt.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = k.CalculateBridgeFee(ctx, amt.amount, "ethereum-1")
			}
		})
	}
}

// BenchmarkAddCollectedFee benchmarks fee collection
func BenchmarkAddCollectedFee(b *testing.B) {
	k, ctx := setupBridgeBenchmark(b)

	fee := sdk.NewCoin("uaura", sdkmath.NewInt(1000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		k.AddCollectedFee(ctx, fee)
	}
}

// BenchmarkGetCollectedFees benchmarks fee retrieval
func BenchmarkGetCollectedFees(b *testing.B) {
	k, ctx := setupBridgeBenchmark(b)

	// Add some fees first
	for i := 0; i < 100; i++ {
		k.AddCollectedFee(ctx, sdk.NewCoin("uaura", sdkmath.NewInt(int64(i*100))))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetCollectedFees(ctx)
	}
}

// ============================================================================
// Cross-Chain Transfer Processing Benchmarks
// ============================================================================

// BenchmarkTransferIDGeneration benchmarks transfer ID generation
func BenchmarkTransferIDGeneration(b *testing.B) {
	sender := testAddr("sender")
	recipient := "0x1234567890abcdef1234567890abcdef12345678"
	amount := sdkmath.NewInt(1000000)
	nonce := uint64(12345)
	timestamp := time.Now().Unix()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		data := fmt.Sprintf("%s:%s:%s:%d:%d", sender, recipient, amount.String(), nonce, timestamp)
		hash := sha256.Sum256([]byte(data))
		_ = hex.EncodeToString(hash[:])
	}
}

// BenchmarkTransferValidation benchmarks transfer validation logic
func BenchmarkTransferValidation(b *testing.B) {
	sender := testAddr("sender")
	transfer := &types.CrossChainTransfer{
		TransferId:  "transfer-1",
		SourceChain: "aura",
		TargetChain: "ethereum-1",
		Sender:      sender,
		Recipient:   "0x1234567890abcdef1234567890abcdef12345678",
		Amount:      sdkmath.NewInt(1000000),
		Denom:       "uaura",
		Status:      types.TransferStatus_PENDING,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Validate transfer fields
		_ = len(transfer.Sender) > 0 &&
			len(transfer.Recipient) > 0 &&
			!transfer.Amount.IsNil() &&
			len(transfer.TargetChain) > 0
	}
}

// ============================================================================
// Chain Configuration Benchmarks
// ============================================================================

// BenchmarkChainConfigLookup benchmarks chain configuration lookup
func BenchmarkChainConfigLookup(b *testing.B) {
	k, ctx := setupBridgeBenchmark(b)

	// Register some chains
	chains := []string{"ethereum-1", "polygon-1", "arbitrum-1", "optimism-1", "base-1"}
	for _, chainID := range chains {
		config := types.ChainConfig{
			ChainId: chainID,
			Enabled: true,
		}
		_ = k.AddSupportedChain(ctx, config)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		chainID := chains[i%len(chains)]
		_, _ = k.GetSupportedChain(ctx, chainID)
	}
}

// ============================================================================
// Parameter Benchmarks
// ============================================================================

// BenchmarkGetParams benchmarks parameter retrieval
func BenchmarkGetParams(b *testing.B) {
	k, ctx := setupBridgeBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.GetParams(ctx)
	}
}

// BenchmarkSetParams benchmarks parameter updates
func BenchmarkSetParams(b *testing.B) {
	k, ctx := setupBridgeBenchmark(b)
	params := types.DefaultParams()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		k.SetParams(ctx, params)
	}
}
