package keeper

import (
	"crypto/sha256"
	"encoding/hex"
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

// ============================================================================
// Identity Creation Benchmarks
// ============================================================================

// BenchmarkCreateIdentity benchmarks DID creation
func BenchmarkCreateIdentity(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	controller := "aura1controller"
	did := "did:aura:test"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testDID := fmt.Sprintf("%s%d", did, i)
		_ = keeper.CreateIdentity(ctx, controller, testDID, nil, nil)
	}
}

// BenchmarkCreateIdentity_WithVerificationMethods benchmarks DID creation with verification methods
func BenchmarkCreateIdentity_WithVerificationMethods(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	controller := "aura1controller"
	did := "did:aura:test"

	// Create verification methods
	verificationMethods := []*types.VerificationMethod{
		{
			Id:         "key-1",
			Type:       "EcdsaSecp256k1VerificationKey2019",
			Controller: controller,
			PublicKeyMultibase: "zQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme",
		},
		{
			Id:         "key-2",
			Type:       "Ed25519VerificationKey2020",
			Controller: controller,
			PublicKeyMultibase: "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testDID := fmt.Sprintf("%s%d", did, i)
		_ = keeper.CreateIdentity(ctx, controller, testDID, verificationMethods, nil)
	}
}

// ============================================================================
// Credential Verification Benchmarks
// ============================================================================

// BenchmarkVerifyCredential benchmarks credential verification
func BenchmarkVerifyCredential(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create a test identity
	did := "did:aura:test"
	controller := "aura1controller"
	keeper.CreateIdentity(ctx, controller, did, nil, nil)

	// Create a test credential
	credentialID := "cred-1"
	credential := &types.VerifiableCredential{
		CredentialId: credentialID,
		Issuer:       did,
		Subject:      did,
		IssuedAt:     time.Now().Unix(),
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
		Status:       types.CredentialStatusActive,
	}
	keeper.SetCredential(ctx, credential)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.VerifyCredential(ctx, credentialID, did)
	}
}

// BenchmarkVerifyCredential_Batch benchmarks batch credential verification
func BenchmarkVerifyCredential_Batch(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create a test identity
	did := "did:aura:test"
	controller := "aura1controller"
	keeper.CreateIdentity(ctx, controller, did, nil, nil)

	// Create 100 test credentials
	credentials := make([]*types.VerifiableCredential, 100)
	for i := 0; i < 100; i++ {
		credentialID := fmt.Sprintf("cred-%d", i)
		credentials[i] = &types.VerifiableCredential{
			CredentialId: credentialID,
			Issuer:       did,
			Subject:      did,
			IssuedAt:     time.Now().Unix(),
			ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
			Status:       types.CredentialStatusActive,
		}
		keeper.SetCredential(ctx, credentials[i])
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, cred := range credentials {
			_ = keeper.VerifyCredential(ctx, cred.CredentialId, did)
		}
	}
}

// ============================================================================
// ZK Proof Verification Benchmarks
// ============================================================================

// BenchmarkZKProofVerification benchmarks zero-knowledge proof verification
func BenchmarkZKProofVerification(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create a simple test proof (this is a mock proof for benchmarking)
	proofType := ZKProofTypeSimple
	proof := createMockZKProof()
	publicInputs := createMockPublicInputs()

	// Register verification key
	verificationKey := createMockVerificationKey()
	keeper.SetZKVerificationKey(ctx, proofType, verificationKey, "test key")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.VerifyZKProof(ctx, proofType, proof, publicInputs)
	}
}

// BenchmarkZKProofVerification_Groth16 benchmarks Groth16 proof verification
func BenchmarkZKProofVerification_Groth16(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	proofType := ZKProofTypeGroth16
	proof := createMockZKProof()
	publicInputs := createMockPublicInputs()

	// Register verification key
	verificationKey := createMockVerificationKey()
	keeper.SetZKVerificationKey(ctx, proofType, verificationKey, "groth16 test key")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.VerifyZKProof(ctx, proofType, proof, publicInputs)
	}
}

// BenchmarkZKProofVerification_PLONK benchmarks PLONK proof verification
func BenchmarkZKProofVerification_PLONK(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	proofType := ZKProofTypePLONK
	proof := createMockZKProof()
	publicInputs := createMockPublicInputs()

	// Register verification key
	verificationKey := createMockVerificationKey()
	keeper.SetZKVerificationKey(ctx, proofType, verificationKey, "plonk test key")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.VerifyZKProof(ctx, proofType, proof, publicInputs)
	}
}

// ============================================================================
// Session Management Benchmarks
// ============================================================================

// BenchmarkCreateSession benchmarks session creation
func BenchmarkCreateSession(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	userAddress := "aura1user"
	expirySeconds := uint64(3600) // 1 hour

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.CreateSession(ctx, userAddress, expirySeconds)
	}
}

// BenchmarkValidateSession benchmarks session validation
func BenchmarkValidateSession(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	userAddress := "aura1user"
	expirySeconds := uint64(3600) // 1 hour

	// Create a session
	session, _ := keeper.CreateSession(ctx, userAddress, expirySeconds)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.ValidateSession(ctx, session.SessionId, userAddress)
	}
}

// ============================================================================
// DID Key Rotation Benchmarks
// ============================================================================

// BenchmarkVerifySignatureWithKey benchmarks signature verification with DID keys
func BenchmarkVerifySignatureWithKey(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create identity with verification method
	did := "did:aura:test"
	controller := "aura1controller"
	verificationMethods := []*types.VerificationMethod{
		{
			Id:         "key-1",
			Type:       "EcdsaSecp256k1VerificationKey2019",
			Controller: controller,
			PublicKeyMultibase: "zQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme",
		},
	}
	keeper.CreateIdentity(ctx, controller, did, verificationMethods, nil)

	message := []byte("test message")
	signature := createMockSignature(message)
	verificationMethod := "key-1"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.VerifySignatureWithKey(ctx, did, verificationMethod, message, signature)
	}
}

// ============================================================================
// PII Commitment Verification Benchmarks
// ============================================================================

// BenchmarkVerifyPIICommitment benchmarks PII commitment verification
func BenchmarkVerifyPIICommitment(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create identity
	did := "did:aura:test"
	controller := "aura1controller"
	keeper.CreateIdentity(ctx, controller, did, nil, nil)

	// Create PII data
	piiData := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   "30",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.VerifyPIICommitment(ctx, did, piiData)
	}
}

// ============================================================================
// Change Request Benchmarks
// ============================================================================

// BenchmarkCreateChangeRequest benchmarks change request creation
func BenchmarkCreateChangeRequest(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	requester := "aura1requester"
	targetDID := "did:aura:target"
	irID := "ir-123"
	metadataHash := hex.EncodeToString(sha256.New().Sum([]byte("metadata")))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = keeper.CreateChangeRequest(ctx, requester, targetDID, irID, metadataHash)
	}
}

// ============================================================================
// Role Management Benchmarks
// ============================================================================

// BenchmarkCreateRole benchmarks role creation
func BenchmarkCreateRole(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	creator := "aura1creator"
	permissions := []string{"permission1", "permission2", "permission3"}
	description := "Test role"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		roleName := fmt.Sprintf("role-%d", i)
		_, _ = keeper.CreateRole(ctx, creator, roleName, permissions, description)
	}
}

// BenchmarkCheckPermission benchmarks permission checking
func BenchmarkCheckPermission(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create role and assign to user
	creator := "aura1creator"
	userAddress := "aura1user"
	permissions := []string{"permission1", "permission2", "permission3"}
	role, _ := keeper.CreateRole(ctx, creator, "test-role", permissions, "test")
	keeper.AssignRole(ctx, creator, userAddress, role.Name)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.HasPermission(ctx, userAddress, "permission1")
	}
}

// ============================================================================
// Identity Query Benchmarks
// ============================================================================

// BenchmarkGetIdentity benchmarks identity retrieval
func BenchmarkGetIdentity(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create identity
	did := "did:aura:test"
	controller := "aura1controller"
	keeper.CreateIdentity(ctx, controller, did, nil, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.GetIdentity(ctx, did)
	}
}

// BenchmarkGetAllIdentities benchmarks retrieval of all identities
func BenchmarkGetAllIdentities(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create 100 identities
	controller := "aura1controller"
	for i := 0; i < 100; i++ {
		did := fmt.Sprintf("did:aura:test%d", i)
		keeper.CreateIdentity(ctx, controller, did, nil, nil)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.GetAllIdentities(ctx)
	}
}

// ============================================================================
// Attribute Access Benchmarks
// ============================================================================

// BenchmarkGetAttribute benchmarks attribute retrieval
func BenchmarkGetAttribute(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create identity with attributes
	did := "did:aura:test"
	controller := "aura1controller"
	keeper.CreateIdentity(ctx, controller, did, nil, nil)

	// Set attribute
	keeper.SetAttribute(ctx, did, "name", "John Doe")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = keeper.GetAttribute(ctx, did, "name")
	}
}

// BenchmarkSetAttribute benchmarks attribute setting
func BenchmarkSetAttribute(b *testing.B) {
	keeper, ctx := setupIdentityKeeperForBenchmark(b)

	// Create identity
	did := "did:aura:test"
	controller := "aura1controller"
	keeper.CreateIdentity(ctx, controller, did, nil, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		attrName := fmt.Sprintf("attr-%d", i)
		_ = keeper.SetAttribute(ctx, did, attrName, "test value")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// setupIdentityKeeperForBenchmark creates a keeper instance for benchmarking
func setupIdentityKeeperForBenchmark(b *testing.B) (*Keeper, sdk.Context) {
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

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

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

// createMockZKProof creates a mock zero-knowledge proof for benchmarking
func createMockZKProof() []byte {
	// Create a simple mock proof (32 bytes)
	proof := make([]byte, 32)
	for i := range proof {
		proof[i] = byte(i % 256)
	}
	return proof
}

// createMockPublicInputs creates mock public inputs for ZK proof
func createMockPublicInputs() []byte {
	// Create simple mock public inputs (32 bytes)
	inputs := make([]byte, 32)
	for i := range inputs {
		inputs[i] = byte((i * 2) % 256)
	}
	return inputs
}

// createMockVerificationKey creates a mock verification key
func createMockVerificationKey() []byte {
	// Create a simple mock verification key (64 bytes)
	key := make([]byte, 64)
	for i := range key {
		key[i] = byte((i * 3) % 256)
	}
	return key
}

// createMockSignature creates a mock signature for a message
func createMockSignature(message []byte) []byte {
	// Create a simple mock signature (64 bytes)
	sig := make([]byte, 64)
	hash := sha256.Sum256(message)
	copy(sig, hash[:])
	return sig
}
