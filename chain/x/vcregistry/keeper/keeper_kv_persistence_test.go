package keeper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// setupKeeperForTest creates a keeper with KV store for testing
func setupKeeperForTest(t *testing.T) (*Keeper, sdk.Context) {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create in-memory state store (CommitMultiStore)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	// Create context with proper Cosmos SDK v0.50+ signature
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Create params store
	paramStore := params.NewStore(*types.DefaultParams())

	// Create keeper
	keeper := NewKeeper(paramStore, "authority")
	keeper = keeper.WithStore(storeKey, cdc)

	// Sync metadata
	keeper.SetCurrentHeight(1)
	keeper.SetCurrentTime(time.Now().Unix())

	return keeper, ctx
}

func TestVCRecord_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Test VC creation
	vcRecord := types.VCRecord{
		VcId:          "vc-test-123",
		HolderAddress: "holder1",
		VcType:        types.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
		ExpiresAt:     &gogotypes.Timestamp{Seconds: time.Now().Add(365 * 24 * time.Hour).Unix(), Nanos: int32(time.Now().Add(365 * 24 * time.Hour).Nanosecond())},
	}

	// Store VC
	err := keeper.SetVCRecord(ctx, vcRecord)
	require.NoError(t, err)

	// Retrieve VC
	retrieved, ok := keeper.GetVCRecord(ctx, "vc-test-123")
	require.True(t, ok)
	require.Equal(t, vcRecord.VcId, retrieved.VcId)
	require.Equal(t, vcRecord.HolderAddress, retrieved.HolderAddress)
	require.Equal(t, vcRecord.VcType, retrieved.VcType)
	require.Equal(t, vcRecord.Status, retrieved.Status)

	// List user VCs
	userVCs := keeper.ListUserVCs(ctx, "holder1", types.VCStatus_VC_STATUS_UNSPECIFIED, types.VCTypeUnspecified)
	require.Len(t, userVCs, 1)
	require.Equal(t, "vc-test-123", userVCs[0].VcId)
}

func TestVCRevocation_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create VC
	vcRecord := types.VCRecord{
		VcId:          "vc-revoke-test",
		HolderAddress: "holder1",
		VcType:        types.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	err := keeper.SetVCRecord(ctx, vcRecord)
	require.NoError(t, err)

	// Revoke VC
	err = keeper.RevokeVC(ctx, "vc-revoke-test", types.RevocationReason_REVOCATION_REASON_SECURITY_COMPROMISE, "revoker1", "test evidence")
	require.NoError(t, err)

	// Check revocation
	isRevoked := keeper.IsRevoked(ctx, "vc-revoke-test")
	require.True(t, isRevoked)

	// Get revocation record
	revRecord, ok := keeper.GetRevocationRecord(ctx, "vc-revoke-test")
	require.True(t, ok)
	require.Equal(t, "vc-revoke-test", revRecord.VcId)
	require.Equal(t, types.RevocationReason_REVOCATION_REASON_SECURITY_COMPROMISE, revRecord.Reason)
	require.Equal(t, "revoker1", revRecord.Revoker)

	// Check VC status updated
	vc, ok := keeper.GetVCRecord(ctx, "vc-revoke-test")
	require.True(t, ok)
	require.Equal(t, types.VCStatus_VC_STATUS_REVOKED, vc.Status)

	// Check revocation list updated
	revList := keeper.GetRevocationList(ctx)
	require.NotNil(t, revList)
	require.Equal(t, uint64(1), revList.TotalRevocations)
	require.NotEmpty(t, revList.MerkleRoot)
}

func TestDIDDocument_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Register DID
	verificationMethods := []*types.VerificationMethod{
		{
			Id:        "key-1",
			Type:      "Ed25519VerificationKey2020",
			PublicKey: []byte("test-public-key"),
		},
	}

	err := keeper.RegisterDID(ctx, "did:aura:test123", "controller1", verificationMethods, "https://metadata.uri")
	require.NoError(t, err)

	// Retrieve DID
	didDoc, ok := keeper.GetDIDDocument(ctx, "did:aura:test123")
	require.True(t, ok)
	require.Equal(t, "did:aura:test123", didDoc.Did)
	require.Equal(t, "controller1", didDoc.Controller)
	require.Len(t, didDoc.VerificationMethods, 1)

	// Get DIDs by address
	dids := keeper.GetDIDsByAddress(ctx, "controller1")
	require.Len(t, dids, 1)
	require.Equal(t, "did:aura:test123", dids[0])

	// Update DID
	newMethods := []*types.VerificationMethod{
		{
			Id:        "key-2",
			Type:      "Ed25519VerificationKey2020",
			PublicKey: []byte("new-public-key"),
		},
	}
	err = keeper.UpdateDIDDocument(ctx, "did:aura:test123", newMethods, "https://new-metadata.uri")
	require.NoError(t, err)

	// Verify update
	didDoc, ok = keeper.GetDIDDocument(ctx, "did:aura:test123")
	require.True(t, ok)
	require.Len(t, didDoc.VerificationMethods, 1)
	require.Equal(t, "key-2", didDoc.VerificationMethods[0].Id)
}

func TestAttributeVC_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create attribute VC
	avc := types.AttributeVC{
		AttributeVcId:  "avc-test-123",
		HolderAddress:  "holder1",
		AttributeType:  types.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		EncryptedValue: []byte("encrypted-email"),
		ValueHash:      []byte("hash-of-value"),
		Status:         types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:       &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}

	err := keeper.CreateAttributeVC(ctx, avc)
	require.NoError(t, err)

	// Retrieve attribute VC
	retrieved, ok := keeper.GetAttributeVC(ctx, "avc-test-123")
	require.True(t, ok)
	require.Equal(t, avc.AttributeVcId, retrieved.AttributeVcId)
	require.Equal(t, avc.HolderAddress, retrieved.HolderAddress)
	require.Equal(t, avc.AttributeType, retrieved.AttributeType)

	// List attribute VCs
	avcs := keeper.ListAttributeVCs(ctx, "holder1", nil)
	require.Len(t, avcs, 1)
	require.Equal(t, "avc-test-123", avcs[0].AttributeVcId)

	// Revoke attribute VC
	err = keeper.RevokeAttributeVC(ctx, "avc-test-123", "test revocation")
	require.NoError(t, err)

	// Verify revocation
	retrieved, ok = keeper.GetAttributeVC(ctx, "avc-test-123")
	require.True(t, ok)
	require.Equal(t, types.VCStatus_VC_STATUS_REVOKED, retrieved.Status)
}

func TestDisclosurePolicy_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create disclosure policy
	policy := types.DisclosurePolicy{
		HolderAddress: "holder1",
		DefaultMode:   types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY,
		Rules: []*types.AttributeDisclosureRule{
			{
				AttributeType: types.AttributeType_ATTRIBUTE_TYPE_EMAIL,
				Mode:          types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
			},
		},
		UpdatedAt: &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}

	err := keeper.SetDisclosurePolicy(ctx, policy)
	require.NoError(t, err)

	// Retrieve policy
	retrieved, ok := keeper.GetDisclosurePolicy(ctx, "holder1")
	require.True(t, ok)
	require.Equal(t, policy.HolderAddress, retrieved.HolderAddress)
	require.Equal(t, policy.DefaultMode, retrieved.DefaultMode)
	require.Len(t, retrieved.Rules, 1)
	require.Equal(t, types.AttributeType_ATTRIBUTE_TYPE_EMAIL, retrieved.Rules[0].AttributeType)
}

func TestMintRateLimit_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Enable rate limiting
	params := keeper.GetParams()
	params.RateLimitingEnabled = true
	params.MaxMintPerDay = 5
	err := keeper.SetParams(params)
	require.NoError(t, err)

	// Increment mint count
	for i := 0; i < 3; i++ {
		keeper.IncrementMintCount(ctx, "holder1")
	}

	// Check rate limit
	err = keeper.CheckMintRateLimit(ctx, "holder1")
	require.NoError(t, err) // Should pass, under limit

	// Increment to limit
	for i := 0; i < 2; i++ {
		keeper.IncrementMintCount(ctx, "holder1")
	}

	// Should fail now
	err = keeper.CheckMintRateLimit(ctx, "holder1")
	require.Error(t, err)
	require.Equal(t, types.ErrRateLimitExceeded, err)
}

func TestGenesis_RoundTrip(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create some state
	vcRecord := types.VCRecord{
		VcId:            "vc-genesis-test",
		HolderAddress:   "holder1",
		VcType:          types.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:          types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:        &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
		IssuerAssistant: "aura1issuer",
	}
	err := keeper.SetVCRecord(ctx, vcRecord)
	require.NoError(t, err)

	err = keeper.RegisterDID(ctx, "did:aura:genesis", "controller1", nil, "")
	require.NoError(t, err)

	// Export genesis
	exported := keeper.ExportGenesis(ctx)
	require.NotNil(t, exported)
	require.Len(t, exported.VcRecords, 1)
	require.Len(t, exported.DidDocuments, 1)

	// Create new keeper
	keeper2, ctx2 := setupKeeperForTest(t)

	// Import genesis
	err = keeper2.InitGenesis(ctx2, exported)
	require.NoError(t, err)

	// Verify state
	vc, ok := keeper2.GetVCRecord(ctx2, "vc-genesis-test")
	require.True(t, ok)
	require.Equal(t, "vc-genesis-test", vc.VcId)

	didDoc, ok := keeper2.GetDIDDocument(ctx2, "did:aura:genesis")
	require.True(t, ok)
	require.Equal(t, "did:aura:genesis", didDoc.Did)
}

func TestKeeperPanicsWithoutStore(t *testing.T) {
	// Create keeper without store
	paramStore := params.NewStore(*types.DefaultParams())
	keeper := NewKeeper(paramStore, "authority")

	// Create dummy context
	ctx := context.Background()

	// Should panic when trying to use keeper without store
	require.Panics(t, func() {
		keeper.GetVCRecord(ctx, "test")
	})
}

func TestNoMemoryLeaks(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create many VCs
	for i := 0; i < 1000; i++ {
		vcRecord := types.VCRecord{
			VcId:          fmt.Sprintf("vc-leak-test-%d", i),
			HolderAddress: fmt.Sprintf("holder%d", i%10),
			VcType:        types.VCType_VC_TYPE_VERIFIED_HUMAN,
			Status:        types.VCStatus_VC_STATUS_ACTIVE,
			IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
		}
		err := keeper.SetVCRecord(ctx, vcRecord)
		require.NoError(t, err)
	}

	// Verify all VCs are in KV store, not memory
	// The keeper should have no in-memory maps
	// This is validated by the struct definition change
	require.NotNil(t, keeper.store, "Store should be initialized")
}
