package keeper

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// setupIdentityBenchmark creates a keeper for benchmarking
func setupIdentityBenchmark(b *testing.B) (*Keeper, sdk.Context) {
	b.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		b.Fatal(err)
	}

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create store service
	storeService := &mockStoreService{storeKey: storeKey}

	k := NewKeeper(
		storeService,
		storeKey,
		cdc,
		"authority",
		log.NewNopLogger(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	// Initialize params
	params := types.DefaultParams()
	if err := k.SetParams(ctx, params); err != nil {
		b.Fatal(err)
	}

	return k, ctx
}

// Mock store service for testing
type mockStoreService struct {
	storeKey storetypes.StoreKey
}

func (m *mockStoreService) OpenKVStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(m.storeKey)
}

// BenchmarkCreateIdentity benchmarks DID registration
func BenchmarkCreateIdentity(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		did := fmt.Sprintf("did:aura:test%d", i)
		identity := &types.DIDocument{
			Did:        did,
			Controller: "aura1controller",
			CreatedAt:  uint64(ctx.BlockTime().Unix()),
			UpdatedAt:  uint64(ctx.BlockTime().Unix()),
		}
		_ = keeper.SetDIDDocument(ctx, identity)
	}
}

// BenchmarkVerifyCredential benchmarks credential validation
func BenchmarkVerifyCredential(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	// Create a test DID
	did := "did:aura:test"
	identity := &types.DIDocument{
		Did:        did,
		Controller: "aura1controller",
		CreatedAt:  uint64(ctx.BlockTime().Unix()),
		UpdatedAt:  uint64(ctx.BlockTime().Unix()),
	}
	keeper.SetDIDDocument(ctx, identity)

	credID := "cred-1"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.VerifyCredential(ctx, credID, did)
	}
}

// BenchmarkZKProofVerification benchmarks ZK proof verification
func BenchmarkZKProofVerification(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	proofType := ZKProofTypeSimple
	proof := make([]byte, 32)
	publicInputs := make([]byte, 32)

	// Register verification key
	verificationKey := make([]byte, 64)
	keeper.SetZKVerificationKey(ctx, proofType, verificationKey, "test key")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.VerifyZKProof(ctx, proofType, proof, publicInputs)
	}
}

// BenchmarkCreateSession benchmarks session creation
func BenchmarkCreateSession(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	userAddress := "aura1user"
	expirySeconds := uint64(3600)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.CreateSession(ctx, userAddress, expirySeconds)
	}
}

// BenchmarkVerifySignatureWithKey benchmarks signature verification
func BenchmarkVerifySignatureWithKey(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	did := "did:aura:test"
	message := []byte("test message")
	signature := make([]byte, 64)
	hash := sha256.Sum256(message)
	copy(signature, hash[:])

	verificationMethod := "key-1"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.VerifySignatureWithKey(ctx, did, verificationMethod, message, signature)
	}
}

// BenchmarkVerifyPIICommitment benchmarks PII commitment verification
func BenchmarkVerifyPIICommitment(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	did := "did:aura:test"
	piiData := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.VerifyPIICommitment(ctx, did, piiData)
	}
}

// BenchmarkCreateChangeRequest benchmarks change request creation
func BenchmarkCreateChangeRequest(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	requester := "aura1requester"
	targetDID := "did:aura:target"
	irID := "ir-123"
	metadataHash := fmt.Sprintf("%x", sha256.Sum256([]byte("metadata")))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.CreateChangeRequest(ctx, requester, targetDID, irID, metadataHash)
	}
}

// BenchmarkCreateRole benchmarks role creation
func BenchmarkCreateRole(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	creator := "aura1creator"
	permissions := []string{"permission1", "permission2"}
	description := "Test role"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		roleName := fmt.Sprintf("role-%d", i)
		_, _ = keeper.CreateRole(ctx, creator, roleName, permissions, description)
	}
}

// BenchmarkGetDIDDocument benchmarks DID retrieval
func BenchmarkGetDIDDocument(b *testing.B) {
	keeper, ctx := setupIdentityBenchmark(b)

	did := "did:aura:test"
	identity := &types.DIDocument{
		Did:        did,
		Controller: "aura1controller",
		CreatedAt:  uint64(ctx.BlockTime().Unix()),
		UpdatedAt:  uint64(ctx.BlockTime().Unix()),
	}
	keeper.SetDIDDocument(ctx, identity)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.GetDIDDocument(ctx, did)
	}
}
