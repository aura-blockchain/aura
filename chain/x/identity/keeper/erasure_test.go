package keeper

import (
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
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// setupKeeperForTest creates a keeper for testing
func setupKeeperForTest(t *testing.T) (*Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	storeService := runtime.NewKVStoreService(storeKey)
	keeper := NewKeeper(storeService, storeKey, cdc, "authority", log.NewNopLogger())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	return keeper, ctx
}

// TestEraseIdentity_Success tests successful identity erasure
func TestEraseIdentity_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity record
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()

	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}
	salt := types.GenerateCommitmentSalt()
	commitment := types.ComputePIICommitment(piiData, salt)

	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        timestamppb.New(now),
		UpdatedAt:        timestamppb.New(now),
		PiiCommitment:    commitment,
		CommitmentSalt:   salt,
		OffChainDataRef:  "ipfs://QmTest123",
		OffChainDataType: "ipfs",
		Erased:           false,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Verify erasure
	erasedRecord, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	require.True(t, erasedRecord.Erased, "identity should be marked as erased")
	require.Equal(t, types.IdentityStatusErased, erasedRecord.Status, "status should be ERASED")
	require.NotNil(t, erasedRecord.ErasedAt, "erased_at should be set")
	require.Empty(t, erasedRecord.OffChainDataRef, "off-chain reference should be cleared")
	require.Empty(t, erasedRecord.OffChainDataType, "off-chain type should be cleared")

	// Verify audit trail preserved
	require.Equal(t, did, erasedRecord.Did, "DID should be preserved")
	require.Equal(t, address, erasedRecord.Address, "address should be preserved")
	require.NotEmpty(t, erasedRecord.PiiCommitment, "commitment should be preserved")
	require.NotEmpty(t, erasedRecord.CommitmentSalt, "salt should be preserved")
}

// TestEraseIdentity_NotFound tests erasure of non-existent identity
func TestEraseIdentity_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	err := keeper.EraseIdentity(ctx, "did:aura:nonexistent", "aura1test", "test")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrIdentityNotFound)
}

// TestEraseIdentity_AlreadyErased tests double erasure
func TestEraseIdentity_AlreadyErased(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create and erase identity
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()

	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// First erasure
	err = keeper.EraseIdentity(ctx, did, address, "first erasure")
	require.NoError(t, err)

	// Second erasure should fail
	err = keeper.EraseIdentity(ctx, did, address, "second erasure")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrIdentityAlreadyErased)
}

// TestEraseIdentity_Unauthorized tests erasure by unauthorized user
func TestEraseIdentity_Unauthorized(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	owner := "aura1owner"
	attacker := "aura1attacker"
	now := ctx.BlockTime()

	record := &types.IdentityRecord{
		Did:       did,
		Address:   owner,
		Status:    types.IdentityStatusActive,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Attempt erasure by non-owner (should fail without admin permission)
	err = keeper.EraseIdentity(ctx, did, attacker, "unauthorized")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

// TestVerifyPIICommitment_Success tests successful commitment verification
func TestVerifyPIICommitment_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity with commitment
	did := "did:aura:test123"
	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}
	salt := types.GenerateCommitmentSalt()
	commitment := types.ComputePIICommitment(piiData, salt)

	record := &types.IdentityRecord{
		Did:            did,
		Address:        "aura1test",
		Status:         types.IdentityStatusActive,
		CreatedAt:      timestamppb.New(ctx.BlockTime()),
		PiiCommitment:  commitment,
		CommitmentSalt: salt,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Verify with correct data
	valid, err := keeper.VerifyPIICommitment(ctx, did, piiData)
	require.NoError(t, err)
	require.True(t, valid, "verification should succeed with correct data")
}

// TestVerifyPIICommitment_WrongData tests verification with wrong data
func TestVerifyPIICommitment_WrongData(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity with commitment
	did := "did:aura:test123"
	correctData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}
	wrongData := map[string]string{
		"name":  "Bob Jones",
		"email": "bob@example.com",
	}

	salt := types.GenerateCommitmentSalt()
	commitment := types.ComputePIICommitment(correctData, salt)

	record := &types.IdentityRecord{
		Did:            did,
		Address:        "aura1test",
		Status:         types.IdentityStatusActive,
		CreatedAt:      timestamppb.New(ctx.BlockTime()),
		PiiCommitment:  commitment,
		CommitmentSalt: salt,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Verify with wrong data
	valid, err := keeper.VerifyPIICommitment(ctx, did, wrongData)
	require.NoError(t, err)
	require.False(t, valid, "verification should fail with wrong data")
}

// TestVerifyPIICommitment_ErasedIdentity tests verification after erasure
func TestVerifyPIICommitment_ErasedIdentity(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}
	salt := types.GenerateCommitmentSalt()
	commitment := types.ComputePIICommitment(piiData, salt)

	record := &types.IdentityRecord{
		Did:            did,
		Address:        address,
		Status:         types.IdentityStatusActive,
		CreatedAt:      timestamppb.New(ctx.BlockTime()),
		PiiCommitment:  commitment,
		CommitmentSalt: salt,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Attempt verification after erasure (should fail)
	valid, err := keeper.VerifyPIICommitment(ctx, did, piiData)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrIdentityErased)
	require.False(t, valid)
}

// TestUpdatePIICommitment_Success tests successful commitment update
func TestUpdatePIICommitment_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	oldData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}
	oldSalt := types.GenerateCommitmentSalt()
	oldCommitment := types.ComputePIICommitment(oldData, oldSalt)

	record := &types.IdentityRecord{
		Did:            did,
		Address:        address,
		Status:         types.IdentityStatusActive,
		CreatedAt:      timestamppb.New(ctx.BlockTime()),
		PiiCommitment:  oldCommitment,
		CommitmentSalt: oldSalt,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Update PII commitment
	newData := map[string]string{
		"name":  "Alice Johnson",
		"email": "alice.johnson@example.com",
	}

	err = keeper.UpdatePIICommitment(ctx, did, address, newData, "ipfs://QmNew", "ipfs")
	require.NoError(t, err)

	// Verify update
	updatedRecord, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)

	// Old commitment should not match
	require.NotEqual(t, oldCommitment, updatedRecord.PiiCommitment)

	// New data should verify
	valid, err := keeper.VerifyPIICommitment(ctx, did, newData)
	require.NoError(t, err)
	require.True(t, valid)

	// Off-chain reference should be updated
	require.Equal(t, "ipfs://QmNew", updatedRecord.OffChainDataRef)
	require.Equal(t, "ipfs", updatedRecord.OffChainDataType)
}

// TestUpdatePIICommitment_ErasedIdentity tests update after erasure
func TestUpdatePIICommitment_ErasedIdentity(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create and erase identity
	did := "did:aura:test123"
	address := "aura1test"
	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: timestamppb.New(ctx.BlockTime()),
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Attempt update (should fail)
	newData := map[string]string{"name": "Alice"}
	err = keeper.UpdatePIICommitment(ctx, did, address, newData, "ipfs://QmNew", "ipfs")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrIdentityErased)
}

// TestUpdatePIICommitment_Unauthorized tests unauthorized update
func TestUpdatePIICommitment_Unauthorized(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	owner := "aura1owner"
	attacker := "aura1attacker"

	record := &types.IdentityRecord{
		Did:       did,
		Address:   owner,
		Status:    types.IdentityStatusActive,
		CreatedAt: timestamppb.New(ctx.BlockTime()),
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Attempt update by non-owner (should fail without admin permission)
	newData := map[string]string{"name": "Attacker"}
	err = keeper.UpdatePIICommitment(ctx, did, attacker, newData, "ipfs://QmEvil", "ipfs")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

// TestEraseIdentity_PreservesCommitment tests that erasure preserves commitment
func TestEraseIdentity_PreservesCommitment(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	piiData := map[string]string{
		"name":  "Alice Smith",
		"email": "alice@example.com",
	}
	salt := types.GenerateCommitmentSalt()
	commitment := types.ComputePIICommitment(piiData, salt)

	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        timestamppb.New(ctx.BlockTime()),
		PiiCommitment:    commitment,
		CommitmentSalt:   salt,
		OffChainDataRef:  "ipfs://QmTest",
		OffChainDataType: "ipfs",
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Store original commitment for comparison
	originalCommitment := make([]byte, len(commitment))
	copy(originalCommitment, commitment)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Verify commitment preserved
	erasedRecord, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	require.Equal(t, originalCommitment, erasedRecord.PiiCommitment, "commitment should be preserved after erasure")
	require.Equal(t, salt, erasedRecord.CommitmentSalt, "salt should be preserved after erasure")
}
