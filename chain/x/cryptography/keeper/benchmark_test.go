// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
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

// setupCryptoBenchmark creates a keeper and context for benchmarking
func setupCryptoBenchmark(b *testing.B) (*Keeper, sdk.Context) {
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
		log.NewNopLogger(),
		"",
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, log.NewNopLogger())

	// Set default params
	params := types.DefaultParams()
	_ = k.SetParams(ctx, &params)

	return k, ctx
}

// ============================================================================
// Hashing Benchmarks
// ============================================================================

// BenchmarkSHA256 benchmarks SHA256 hashing
func BenchmarkSHA256(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"32B", 32},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
		{"64KB", 65536},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := make([]byte, size.size)
			_, _ = rand.Read(data)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = sha256.Sum256(data)
			}
		})
	}
}

// BenchmarkSHA512 benchmarks SHA512 hashing
func BenchmarkSHA512(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"32B", 32},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := make([]byte, size.size)
			_, _ = rand.Read(data)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				h := sha512.New()
				h.Write(data)
				_ = h.Sum(nil)
			}
		})
	}
}

// BenchmarkSHA3_256 benchmarks SHA3-256 hashing
func BenchmarkSHA3_256(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"32B", 32},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := make([]byte, size.size)
			_, _ = rand.Read(data)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				h := sha3.New256()
				h.Write(data)
				_ = h.Sum(nil)
			}
		})
	}
}

// BenchmarkBLAKE2b benchmarks BLAKE2b-512 hashing
func BenchmarkBLAKE2b(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"32B", 32},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := make([]byte, size.size)
			_, _ = rand.Read(data)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				h, _ := blake2b.New512(nil)
				h.Write(data)
				_ = h.Sum(nil)
			}
		})
	}
}

// BenchmarkHashComparison benchmarks constant-time hash comparison
func BenchmarkHashComparison(b *testing.B) {
	k, _ := setupCryptoBenchmark(b)

	hash1 := make([]byte, 32)
	hash2 := make([]byte, 32)
	_, _ = rand.Read(hash1)
	copy(hash2, hash1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.CompareHashes(hash1, hash2)
	}
}

// BenchmarkHashWithCustomSalt benchmarks custom salt hashing
func BenchmarkHashWithCustomSalt(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	data := make([]byte, 64)
	salt := make([]byte, 16)
	_, _ = rand.Read(data)
	_, _ = rand.Read(salt)

	algorithms := []struct {
		name string
		algo cryptoproto.HashAlgorithm
	}{
		{"SHA256", cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256},
		{"SHA512", cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA512},
		{"SHA3_256", cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA3_256},
		{"BLAKE2b", cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE2B},
	}

	for _, algo := range algorithms {
		b.Run(algo.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = k.HashWithCustomSalt(ctx, data, salt, algo.algo, 1)
			}
		})
	}
}

// BenchmarkIteratedHash benchmarks hash with multiple iterations
func BenchmarkIteratedHash(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	data := make([]byte, 64)
	salt := make([]byte, 16)
	_, _ = rand.Read(data)
	_, _ = rand.Read(salt)

	iterations := []struct {
		name string
		iter int32
	}{
		{"1_iter", 1},
		{"10_iter", 10},
		{"100_iter", 100},
		{"1000_iter", 1000},
	}

	for _, iter := range iterations {
		b.Run(iter.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = k.HashWithCustomSalt(ctx, data, salt, cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256, iter.iter)
			}
		})
	}
}

// ============================================================================
// Key Stretching Benchmarks
// ============================================================================

// BenchmarkPBKDF2 benchmarks PBKDF2-SHA512 key stretching
func BenchmarkPBKDF2(b *testing.B) {
	iterations := []struct {
		name string
		iter int
	}{
		{"1000_iter", 1000},
		{"10000_iter", 10000},
		{"100000_iter", 100000},
	}

	password := []byte("test_password_123")
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	for _, iter := range iterations {
		b.Run(iter.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = pbkdf2.Key(password, salt, iter.iter, 64, sha512.New)
			}
		})
	}
}

// BenchmarkArgon2id benchmarks Argon2id key stretching
func BenchmarkArgon2id(b *testing.B) {
	configs := []struct {
		name    string
		memory  uint32
		time    uint32
		threads uint8
	}{
		{"Low_16MB_1iter", 16 * 1024, 1, 2},
		{"Medium_64MB_1iter", 64 * 1024, 1, 4},
		{"High_64MB_3iter", 64 * 1024, 3, 4},
	}

	password := []byte("test_password_123")
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	for _, config := range configs {
		b.Run(config.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = argon2.IDKey(password, salt, config.time, config.memory, config.threads, 64)
			}
		})
	}
}

// BenchmarkPerformKeyStretching benchmarks the keeper's key stretching
func BenchmarkPerformKeyStretching(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	password := []byte("test_password_123")
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)

	b.Run("PBKDF2_10000", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = k.PerformKeyStretching(ctx, password, salt, "pbkdf2-sha512", 10000)
		}
	})

	b.Run("Argon2id_1iter", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = k.PerformKeyStretching(ctx, password, salt, "argon2id", 1)
		}
	})
}

// ============================================================================
// Encryption Benchmarks
// ============================================================================

// BenchmarkAESGCM_Encrypt benchmarks AES-GCM encryption
func BenchmarkAESGCM_Encrypt(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"64B", 64},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
		{"16KB", 16384},
	}

	key := make([]byte, 32)
	_, _ = rand.Read(key)

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			plaintext := make([]byte, size.size)
			_, _ = rand.Read(plaintext)

			block, _ := aes.NewCipher(key)
			gcm, _ := cipher.NewGCM(block)
			nonce := make([]byte, gcm.NonceSize())
			_, _ = rand.Read(nonce)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = gcm.Seal(nil, nonce, plaintext, nil)
			}
		})
	}
}

// BenchmarkAESGCM_Decrypt benchmarks AES-GCM decryption
func BenchmarkAESGCM_Decrypt(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"64B", 64},
		{"256B", 256},
		{"1KB", 1024},
		{"4KB", 4096},
	}

	key := make([]byte, 32)
	_, _ = rand.Read(key)

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			plaintext := make([]byte, size.size)
			_, _ = rand.Read(plaintext)

			block, _ := aes.NewCipher(key)
			gcm, _ := cipher.NewGCM(block)
			nonce := make([]byte, gcm.NonceSize())
			_, _ = rand.Read(nonce)

			ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = gcm.Open(nil, nonce, ciphertext, nil)
			}
		})
	}
}

// ============================================================================
// Key Generation Benchmarks
// ============================================================================

// BenchmarkECDSAKeyGeneration benchmarks ECDSA key generation
func BenchmarkECDSAKeyGeneration(b *testing.B) {
	curves := []struct {
		name  string
		curve elliptic.Curve
	}{
		{"P256", elliptic.P256()},
		{"P384", elliptic.P384()},
		{"P521", elliptic.P521()},
	}

	for _, c := range curves {
		b.Run(c.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = ecdsa.GenerateKey(c.curve, rand.Reader)
			}
		})
	}
}

// BenchmarkSecureRandomBytes benchmarks secure random byte generation
func BenchmarkSecureRandomBytes(b *testing.B) {
	k, _ := setupCryptoBenchmark(b)

	sizes := []struct {
		name string
		size int
	}{
		{"16B", 16},
		{"32B", 32},
		{"64B", 64},
		{"256B", 256},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = k.GenerateSecureRandomBytes(size.size)
			}
		})
	}
}

// ============================================================================
// Signature Verification Benchmarks
// ============================================================================

// BenchmarkECDSASign benchmarks ECDSA signing
func BenchmarkECDSASign(b *testing.B) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	hash := make([]byte, 32)
	_, _ = rand.Read(hash)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = ecdsa.Sign(rand.Reader, key, hash)
	}
}

// BenchmarkECDSAVerify benchmarks ECDSA verification
func BenchmarkECDSAVerify(b *testing.B) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	hash := make([]byte, 32)
	_, _ = rand.Read(hash)
	r, s, _ := ecdsa.Sign(rand.Reader, key, hash)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ecdsa.Verify(&key.PublicKey, hash, r, s)
	}
}

// ============================================================================
// Threshold Signature Benchmarks
// ============================================================================

// BenchmarkCreateThresholdScheme benchmarks threshold scheme creation
func BenchmarkCreateThresholdScheme(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	participants := []string{"p1", "p2", "p3", "p4", "p5"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = k.CreateThresholdScheme(ctx, "creator", 3, 5, participants,
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA)
	}
}

// BenchmarkPerformDKG benchmarks Distributed Key Generation
func BenchmarkPerformDKG(b *testing.B) {
	k, _ := setupCryptoBenchmark(b)

	configs := []struct {
		name      string
		threshold uint32
		total     uint32
	}{
		{"2_of_3", 2, 3},
		{"3_of_5", 3, 5},
		{"5_of_10", 5, 10},
		{"7_of_15", 7, 15},
	}

	for _, cfg := range configs {
		b.Run(cfg.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = k.PerformDKG(cfg.threshold, cfg.total, "test-scheme")
			}
		})
	}
}

// BenchmarkGenerateThresholdSignatureShare benchmarks signature share generation
func BenchmarkGenerateThresholdSignatureShare(b *testing.B) {
	k, _ := setupCryptoBenchmark(b)

	result, _ := k.PerformDKG(3, 5, "test-scheme")
	message := []byte("test message for signing")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Generate share for participant 0
		_ = GenerateThresholdSignatureShare(result.SecretShares[0], message)
	}
}

// BenchmarkLagrangeCoefficient benchmarks Lagrange coefficient computation
func BenchmarkLagrangeCoefficient(b *testing.B) {
	sizes := []struct {
		name    string
		indices []int
	}{
		{"3_of_5", []int{1, 2, 3}},
		{"5_of_10", []int{1, 3, 5, 7, 9}},
		{"7_of_15", []int{1, 3, 5, 7, 9, 11, 13}},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = LagrangeCoefficient(0, size.indices)
			}
		})
	}
}

// BenchmarkHashToG1 benchmarks hash-to-curve operation
func BenchmarkHashToG1(b *testing.B) {
	message := make([]byte, 64)
	_, _ = rand.Read(message)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = HashToG1(message)
	}
}

// ============================================================================
// ZK Proof Benchmarks
// ============================================================================

// BenchmarkZKProofStorage benchmarks ZK proof storage operations
func BenchmarkZKProofStorage(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	proof := &cryptoproto.ZKProof{
		ProofId:      "test-proof-id",
		ProofData:    make([]byte, 256),
		PublicInputs: make([]byte, 64),
		GeneratedAt:  time.Now(),
		Verified:     false,
	}
	_, _ = rand.Read(proof.ProofData)
	_, _ = rand.Read(proof.PublicInputs)

	b.Run("SetZKProof", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			proof.ProofId = string(rune('0' + i%10))
			_ = k.SetZKProof(ctx, proof)
		}
	})

	// Store a proof first
	_ = k.SetZKProof(ctx, proof)

	b.Run("GetZKProof", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = k.GetZKProof(ctx, "test-proof-id")
		}
	})
}

// ============================================================================
// Certificate Pinning Benchmarks
// ============================================================================

// BenchmarkCertificatePinning benchmarks certificate pinning operations
func BenchmarkCertificatePinning(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	domain := "example.com"
	certificate := make([]byte, 1024)
	_, _ = rand.Read(certificate)
	fingerprint := "SHA256:abc123"

	b.Run("PinCertificate", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = k.PinCertificate(ctx, domain, certificate, fingerprint)
		}
	})

	// Pin certificate first
	_ = k.PinCertificate(ctx, domain, certificate, fingerprint)

	b.Run("GetCertificatePin", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = k.GetCertificatePin(ctx, domain)
		}
	})
}

// ============================================================================
// Key Management Benchmarks
// ============================================================================

// BenchmarkQuantumResistantKey benchmarks quantum-resistant key operations
func BenchmarkQuantumResistantKey(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	blockTime := time.Now()
	expiryTime := blockTime.Add(24 * time.Hour)

	key := &cryptoproto.QuantumResistantKey{
		KeyId:       "test-key",
		Algorithm:   cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
		PublicKey:   make([]byte, 1952), // Dilithium-3 public key size
		KeyMetadata: []byte("test metadata"),
		CreatedAt:   blockTime,
		ExpiresAt:   &expiryTime,
	}
	_, _ = rand.Read(key.PublicKey)

	b.Run("SetQuantumResistantKey", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = k.SetQuantumResistantKey(ctx, key)
		}
	})

	// Store key first
	_ = k.SetQuantumResistantKey(ctx, key)

	b.Run("GetQuantumResistantKey", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = k.GetQuantumResistantKey(ctx, "test-key")
		}
	})
}

// ============================================================================
// Salted Hash Benchmarks
// ============================================================================

// BenchmarkSaltedHashOperations benchmarks salted hash CRUD
func BenchmarkSaltedHashOperations(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	hash := &cryptoproto.SaltedHash{
		HashId:     "test-hash",
		Salt:       make([]byte, 16),
		Hash:       make([]byte, 32),
		Algorithm:  cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256,
		Iterations: 1000,
		CreatedAt:  time.Now(),
	}
	_, _ = rand.Read(hash.Salt)
	_, _ = rand.Read(hash.Hash)

	b.Run("SetSaltedHash", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = k.SetSaltedHash(ctx, hash)
		}
	})

	// Store hash first
	_ = k.SetSaltedHash(ctx, hash)

	b.Run("GetSaltedHash", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = k.GetSaltedHash(ctx, "test-hash")
		}
	})
}

// ============================================================================
// Parameter Benchmarks
// ============================================================================

// BenchmarkGetParams benchmarks parameter retrieval
func BenchmarkGetParams(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = k.GetParams(ctx)
	}
}

// BenchmarkSetParams benchmarks parameter updates
func BenchmarkSetParams(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)
	params := types.DefaultParams()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = k.SetParams(ctx, &params)
	}
}

// ============================================================================
// Polynomial Benchmarks (Threshold Signatures)
// ============================================================================

// BenchmarkNewShamirPolynomial benchmarks Shamir polynomial creation
func BenchmarkNewShamirPolynomial(b *testing.B) {
	thresholds := []int{3, 5, 10, 20}

	for _, threshold := range thresholds {
		b.Run(string(rune('0'+threshold/10))+string(rune('0'+threshold%10))+"_threshold", func(b *testing.B) {
			var secret fr.Element
			_, _ = secret.SetRandom()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = NewShamirPolynomial(&secret, threshold)
			}
		})
	}
}

// BenchmarkPolynomialEvaluate benchmarks polynomial evaluation
func BenchmarkPolynomialEvaluate(b *testing.B) {
	var secret fr.Element
	_, _ = secret.SetRandom()
	poly, _ := NewShamirPolynomial(&secret, 5)

	var x fr.Element
	x.SetUint64(3)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = poly.Evaluate(&x)
	}
}

// ============================================================================
// HD Key Derivation Benchmarks
// ============================================================================

// BenchmarkHDKeyDerivation benchmarks HD key derivation operations
func BenchmarkHDKeyDerivation(b *testing.B) {
	k, ctx := setupCryptoBenchmark(b)

	derivation := &cryptoproto.HDKeyDerivation{
		MasterKeyId:    "master-key-1",
		DerivationPath: "m/44'/60'/0'/0/0",
		SeedHash:       make([]byte, 32),
		ChainCode:      make([]byte, 32),
		Depth:          5,
		CreatedAt:      time.Now(),
	}
	_, _ = rand.Read(derivation.SeedHash)
	_, _ = rand.Read(derivation.ChainCode)

	b.Run("SetHDKeyDerivation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = k.SetHDKeyDerivation(ctx, derivation)
		}
	})

	// Store derivation first
	_ = k.SetHDKeyDerivation(ctx, derivation)

	b.Run("GetHDKeyDerivation", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = k.GetHDKeyDerivation(ctx, "master-key-1")
		}
	})
}
