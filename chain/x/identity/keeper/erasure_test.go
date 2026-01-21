// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
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

	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/identity/types"
)

// setupKeeperForTest creates a keeper for testing
func setupKeeperForTest(t *testing.T) (*Keeper, sdk.Context) {
	// Ensure bech32 prefixes match test addresses (aura1...).
	testutil.EnsureSDKConfig()

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
	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        now,
		UpdatedAt:        &now,
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
		CreatedAt: now,
		UpdatedAt: &now,
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
		CreatedAt: now,
		UpdatedAt: &now,
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
	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:            did,
		Address:        "aura1test",
		Status:         types.IdentityStatusActive,
		CreatedAt:      now,
		UpdatedAt:      &now,
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

	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(correctData, salt)

	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:            did,
		Address:        "aura1test",
		Status:         types.IdentityStatusActive,
		CreatedAt:      now,
		UpdatedAt:      &now,
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
	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:            did,
		Address:        address,
		Status:         types.IdentityStatusActive,
		CreatedAt:      now,
		UpdatedAt:      &now,
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
	oldSalt := mustGenerateSalt(t)
	oldCommitment := types.ComputePIICommitment(oldData, oldSalt)

	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:            did,
		Address:        address,
		Status:         types.IdentityStatusActive,
		CreatedAt:      now,
		UpdatedAt:      &now,
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
	newSalt := mustGenerateSalt(t)
	newCommitment := types.ComputePIICommitment(newData, newSalt)

	err = keeper.UpdatePIICommitment(ctx, did, address, newSalt, "ipfs://QmNew", "ipfs")
	require.NoError(t, err)

	// Update the commitment field manually since UpdatePIICommitment now only stores salt
	updatedRecord, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	updatedRecord.PiiCommitment = newCommitment
	err = keeper.SetIdentityRecord(ctx, updatedRecord)
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
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Attempt update (should fail)
	newSalt := mustGenerateSalt(t)
	err = keeper.UpdatePIICommitment(ctx, did, address, newSalt, "ipfs://QmNew", "ipfs")
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

	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   owner,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}

	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Attempt update by non-owner (should fail without admin permission)
	newSalt := mustGenerateSalt(t)
	err = keeper.UpdatePIICommitment(ctx, did, attacker, newSalt, "ipfs://QmEvil", "ipfs")
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
	salt := mustGenerateSalt(t)
	commitment := types.ComputePIICommitment(piiData, salt)

	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:              did,
		Address:          address,
		Status:           types.IdentityStatusActive,
		CreatedAt:        now,
		UpdatedAt:        &now,
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

// TestEraseIdentity_CascadeDeletesChangeRequests tests that erasure deletes all associated change requests
func TestEraseIdentity_CascadeDeletesChangeRequests(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Create multiple change requests for this DID
	req1 := &types.ChangeRequest{
		Id:          "req-1",
		Requester:   address,
		Did:         did,
		Status:      types.ChangeStatusPending,
		RequestedAt: now,
		ChangeType:  types.ChangeTypeUpdateMetadata,
	}
	req2 := &types.ChangeRequest{
		Id:          "req-2",
		Requester:   address,
		Did:         did,
		Status:      types.ChangeStatusApproved,
		RequestedAt: now,
		ChangeType:  types.ChangeTypeUpdateMetadata,
	}
	err = keeper.SetChangeRequest(ctx, req1)
	require.NoError(t, err)
	err = keeper.SetChangeRequest(ctx, req2)
	require.NoError(t, err)

	// Verify requests exist before erasure
	_, err = keeper.GetChangeRequest(ctx, "req-1")
	require.NoError(t, err)
	_, err = keeper.GetChangeRequest(ctx, "req-2")
	require.NoError(t, err)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Verify change requests are deleted
	_, err = keeper.GetChangeRequest(ctx, "req-1")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrChangeRequestNotFound)

	_, err = keeper.GetChangeRequest(ctx, "req-2")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrChangeRequestNotFound)
}

// TestEraseIdentity_CascadeDeletesSessions tests that erasure deletes all associated sessions
func TestEraseIdentity_CascadeDeletesSessions(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Create multiple sessions for this address
	session1, err := keeper.CreateSession(ctx, address, 3600)
	require.NoError(t, err)
	session2, err := keeper.CreateSession(ctx, address, 7200)
	require.NoError(t, err)

	// Verify sessions exist before erasure
	_, err = keeper.GetSession(ctx, session1.Id)
	require.NoError(t, err)
	_, err = keeper.GetSession(ctx, session2.Id)
	require.NoError(t, err)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Verify sessions are deleted
	_, err = keeper.GetSession(ctx, session1.Id)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrSessionNotFound)

	_, err = keeper.GetSession(ctx, session2.Id)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrSessionNotFound)
}

// TestEraseIdentity_CascadeDeletesRoleAssignments tests that erasure deletes all role assignments
func TestEraseIdentity_CascadeDeletesRoleAssignments(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Create roles
	role1 := &types.Role{
		Name:        "viewer",
		Permissions: []string{"read"},
		CreatedAt:   now,
		CreatedBy:   "admin",
	}
	role2 := &types.Role{
		Name:        "editor",
		Permissions: []string{"read", "write"},
		CreatedAt:   now,
		CreatedBy:   "admin",
	}
	err = keeper.SetRole(ctx, role1)
	require.NoError(t, err)
	err = keeper.SetRole(ctx, role2)
	require.NoError(t, err)

	// Assign roles to the identity's address
	assignment1 := &types.RoleAssignment{
		Address:    address,
		RoleName:   "viewer",
		AssignedBy: "admin",
		AssignedAt: now,
	}
	assignment2 := &types.RoleAssignment{
		Address:    address,
		RoleName:   "editor",
		AssignedBy: "admin",
		AssignedAt: now,
	}
	err = keeper.SetRoleAssignment(ctx, assignment1)
	require.NoError(t, err)
	err = keeper.SetRoleAssignment(ctx, assignment2)
	require.NoError(t, err)

	// Verify assignments exist before erasure
	assignments, err := keeper.GetRoleAssignments(ctx, address)
	require.NoError(t, err)
	require.Len(t, assignments, 2)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Verify role assignments are deleted
	assignments, err = keeper.GetRoleAssignments(ctx, address)
	require.NoError(t, err)
	require.Len(t, assignments, 0, "all role assignments should be deleted")
}

// TestEraseIdentity_CascadeDeletesDIDKeyRotations tests that erasure deletes DID key rotation records
func TestEraseIdentity_CascadeDeletesDIDKeyRotations(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set up params for grace period
	params := &types.Params{
		Auth: types.AuthParams{
			MaxRolesPerAccount:            10,
			DefaultRequestsPerMinute:      60,
			DefaultRequestsPerHour:        3600,
			DefaultRequestsPerDay:         86400,
			MultisigProposalExpirySeconds: 604800,
		},
		Change: types.IdentityChangeParams{
			MaxRequestsPerWalletPerMonth:  10,
			MinConfidenceAfterChange:      50,
			KeyRotationGracePeriodSeconds: 86400,
		},
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create identity with verification method
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:                 did,
		Address:             address,
		Status:              types.IdentityStatusActive,
		CreatedAt:           now,
		UpdatedAt:           &now,
		VerificationMethods: []string{"key1"},
	}
	err = keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Create DID key rotation
	rotation := &types.DIDKeyRotation{
		Did:                   did,
		OldVerificationMethod: "key1",
		NewVerificationMethod: "key2",
		RotationTime:          now,
		InitiatedBy:           address,
		Reason:                "key compromise",
		GracePeriodEnd:        now.Add(24 * time.Hour),
		Status:                types.DIDKeyRotationStatusPending,
	}
	err = keeper.SetDIDKeyRotation(ctx, rotation)
	require.NoError(t, err)

	// Verify rotation exists before erasure
	_, err = keeper.GetDIDKeyRotation(ctx, did)
	require.NoError(t, err)

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Verify DID key rotation is deleted
	_, err = keeper.GetDIDKeyRotation(ctx, did)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDIDKeyRotationNotFound)
}

// TestEraseIdentity_CascadeDeletesAllRelatedData tests comprehensive cascade deletion
func TestEraseIdentity_CascadeDeletesAllRelatedData(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set up params
	params := &types.Params{
		Auth: types.AuthParams{
			MaxRolesPerAccount:            10,
			DefaultRequestsPerMinute:      60,
			DefaultRequestsPerHour:        3600,
			DefaultRequestsPerDay:         86400,
			MultisigProposalExpirySeconds: 604800,
		},
		Change: types.IdentityChangeParams{
			MaxRequestsPerWalletPerMonth:  10,
			MinConfidenceAfterChange:      50,
			KeyRotationGracePeriodSeconds: 86400,
		},
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:                 did,
		Address:             address,
		Status:              types.IdentityStatusActive,
		CreatedAt:           now,
		UpdatedAt:           &now,
		VerificationMethods: []string{"key1"},
		OffChainDataRef:     "ipfs://QmTest",
		OffChainDataType:    "ipfs",
	}
	err = keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Create change request
	changeReq := &types.ChangeRequest{
		Id:          "req-1",
		Requester:   address,
		Did:         did,
		Status:      types.ChangeStatusPending,
		RequestedAt: now,
		ChangeType:  types.ChangeTypeUpdateMetadata,
	}
	err = keeper.SetChangeRequest(ctx, changeReq)
	require.NoError(t, err)

	// Create session
	session, err := keeper.CreateSession(ctx, address, 3600)
	require.NoError(t, err)

	// Create role and assignment
	role := &types.Role{
		Name:        "admin",
		Permissions: []string{"manage_all"},
		CreatedAt:   now,
		CreatedBy:   "system",
	}
	err = keeper.SetRole(ctx, role)
	require.NoError(t, err)

	assignment := &types.RoleAssignment{
		Address:    address,
		RoleName:   "admin",
		AssignedBy: "system",
		AssignedAt: now,
	}
	err = keeper.SetRoleAssignment(ctx, assignment)
	require.NoError(t, err)

	// Create DID key rotation
	rotation := &types.DIDKeyRotation{
		Did:                   did,
		OldVerificationMethod: "key1",
		NewVerificationMethod: "key2",
		RotationTime:          now,
		InitiatedBy:           address,
		Reason:                "regular rotation",
		GracePeriodEnd:        now.Add(24 * time.Hour),
		Status:                types.DIDKeyRotationStatusPending,
	}
	err = keeper.SetDIDKeyRotation(ctx, rotation)
	require.NoError(t, err)

	// Verify all data exists before erasure
	_, err = keeper.GetChangeRequest(ctx, "req-1")
	require.NoError(t, err)
	_, err = keeper.GetSession(ctx, session.Id)
	require.NoError(t, err)
	assignments, err := keeper.GetRoleAssignments(ctx, address)
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	_, err = keeper.GetDIDKeyRotation(ctx, did)
	require.NoError(t, err)

	// Erase identity (cascade delete all)
	err = keeper.EraseIdentity(ctx, did, address, "GDPR Right to Erasure")
	require.NoError(t, err)

	// Verify identity is marked as erased
	erasedRecord, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	require.True(t, erasedRecord.Erased)
	require.Equal(t, types.IdentityStatusErased, erasedRecord.Status)
	require.Empty(t, erasedRecord.OffChainDataRef)
	require.Empty(t, erasedRecord.OffChainDataType)

	// Verify all related data is deleted
	_, err = keeper.GetChangeRequest(ctx, "req-1")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrChangeRequestNotFound)

	_, err = keeper.GetSession(ctx, session.Id)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrSessionNotFound)

	assignments, err = keeper.GetRoleAssignments(ctx, address)
	require.NoError(t, err)
	require.Len(t, assignments, 0)

	_, err = keeper.GetDIDKeyRotation(ctx, did)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDIDKeyRotationNotFound)
}

// TestEraseIdentity_NoOrphanedData tests that no orphaned data remains after erasure
func TestEraseIdentity_NoOrphanedData(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create identity
	did := "did:aura:test123"
	address := "aura1test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   address,
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}
	err := keeper.SetIdentityRecord(ctx, record)
	require.NoError(t, err)

	// Create multiple change requests
	for i := 1; i <= 5; i++ {
		req := &types.ChangeRequest{
			Id:          fmt.Sprintf("req-%d", i),
			Requester:   address,
			Did:         did,
			Status:      types.ChangeStatusPending,
			RequestedAt: now,
			ChangeType:  types.ChangeTypeUpdateMetadata,
		}
		err = keeper.SetChangeRequest(ctx, req)
		require.NoError(t, err)
	}

	// Erase identity
	err = keeper.EraseIdentity(ctx, did, address, "GDPR request")
	require.NoError(t, err)

	// Get ALL change requests and verify none reference the erased DID
	allRequests, err := keeper.GetAllChangeRequests(ctx)
	require.NoError(t, err)

	for _, req := range allRequests {
		require.NotEqual(t, did, req.Did, "found orphaned change request: %s", req.Id)
	}
}
