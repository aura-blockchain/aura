package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// TestIsCredentialRevoked tests the basic revocation check
func TestIsCredentialRevoked(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	tests := []struct {
		name           string
		credentialID   string
		setupRevoke    bool
		expectedResult bool
	}{
		{
			name:           "non-revoked credential",
			credentialID:   "cred-001",
			setupRevoke:    false,
			expectedResult: false,
		},
		{
			name:           "revoked credential",
			credentialID:   "cred-002",
			setupRevoke:    true,
			expectedResult: true,
		},
		{
			name:           "empty credential ID",
			credentialID:   "",
			setupRevoke:    false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create identity
			did := "did:aura:test-" + tt.name
			owner := "aura1owner123456789"
			identity := &types.IdentityRecord{
				Did:             did,
				Address:         owner,
				Status:          types.IdentityStatusActive,
				CreatedAt:       timestamppb.New(time.Now()),
				UpdatedAt:       timestamppb.New(time.Now()),
				ConfidenceScore: 80,
			}
			require.NoError(t, k.SetIdentityRecord(ctx, identity))

			// Setup revocation if needed
			if tt.setupRevoke && tt.credentialID != "" {
				err := k.RevokeCredential(ctx, tt.credentialID, did, owner, "test revocation", nil)
				require.NoError(t, err)
			}

			// Test
			result := k.IsCredentialRevoked(ctx, tt.credentialID)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestRevokeCredential tests credential revocation with various scenarios
func TestRevokeCredential(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create identity
	did := "did:aura:test-revoke"
	owner := "aura1owner123456789"
	identity := &types.IdentityRecord{
		Did:             did,
		Address:         owner,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity))

	tests := []struct {
		name          string
		credentialID  string
		did           string
		revoker       string
		reason        string
		metadata      map[string]string
		setupRevoke   bool
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid revocation by owner",
			credentialID: "cred-valid-001",
			did:          did,
			revoker:      owner,
			reason:       "compromised key",
			metadata: map[string]string{
				"severity": "high",
			},
			setupRevoke: false,
			expectError: false,
		},
		{
			name:          "empty credential ID",
			credentialID:  "",
			did:           did,
			revoker:       owner,
			reason:        "test",
			setupRevoke:   false,
			expectError:   true,
			errorContains: "credential ID cannot be empty",
		},
		{
			name:          "empty DID",
			credentialID:  "cred-002",
			did:           "",
			revoker:       owner,
			reason:        "test",
			setupRevoke:   false,
			expectError:   true,
			errorContains: "DID cannot be empty",
		},
		{
			name:          "empty revoker",
			credentialID:  "cred-003",
			did:           did,
			revoker:       "",
			reason:        "test",
			setupRevoke:   false,
			expectError:   true,
			errorContains: "revoker address cannot be empty",
		},
		{
			name:          "already revoked",
			credentialID:  "cred-004",
			did:           did,
			revoker:       owner,
			reason:        "test",
			setupRevoke:   true,
			expectError:   true,
			errorContains: "already revoked",
		},
		{
			name:          "non-existent identity",
			credentialID:  "cred-005",
			did:           "did:aura:nonexistent",
			revoker:       owner,
			reason:        "test",
			setupRevoke:   false,
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:          "unauthorized revoker",
			credentialID:  "cred-006",
			did:           did,
			revoker:       "aura1unauthorized",
			reason:        "test",
			setupRevoke:   false,
			expectError:   true,
			errorContains: "not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Pre-revoke if needed
			if tt.setupRevoke {
				err := k.RevokeCredential(ctx, tt.credentialID, tt.did, owner, "pre-test revocation", nil)
				require.NoError(t, err)
			}

			// Test
			err := k.RevokeCredential(ctx, tt.credentialID, tt.did, tt.revoker, tt.reason, tt.metadata)

			// Verify
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)

				// Verify revocation was stored
				require.True(t, k.IsCredentialRevoked(ctx, tt.credentialID))

				// Verify revocation details
				revocation, err := k.GetCredentialRevocation(ctx, tt.credentialID)
				require.NoError(t, err)
				require.Equal(t, tt.credentialID, revocation.CredentialId)
				require.Equal(t, tt.did, revocation.Did)
				require.Equal(t, tt.revoker, revocation.RevokedBy)
				require.Equal(t, tt.reason, revocation.Reason)
				if tt.metadata != nil {
					require.Equal(t, tt.metadata, revocation.Metadata)
				}
			}
		})
	}
}

// TestBatchRevokeCredentials tests batch revocation functionality
func TestBatchRevokeCredentials(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create identity
	did := "did:aura:test-batch"
	owner := "aura1owner123456789"
	identity := &types.IdentityRecord{
		Did:             did,
		Address:         owner,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity))

	tests := []struct {
		name          string
		credentialIDs []string
		did           string
		revoker       string
		reason        string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid batch revocation",
			credentialIDs: []string{
				"batch-cred-001",
				"batch-cred-002",
				"batch-cred-003",
			},
			did:         did,
			revoker:     owner,
			reason:      "batch security update",
			expectError: false,
		},
		{
			name:          "empty credential list",
			credentialIDs: []string{},
			did:           did,
			revoker:       owner,
			reason:        "test",
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name: "batch with some already revoked",
			credentialIDs: []string{
				"batch-cred-004",
				"batch-cred-005", // will be pre-revoked
				"batch-cred-006",
			},
			did:         did,
			revoker:     owner,
			reason:      "test",
			expectError: false, // Should succeed for non-revoked ones
		},
		{
			name: "unauthorized batch revocation",
			credentialIDs: []string{
				"batch-cred-007",
			},
			did:           did,
			revoker:       "aura1unauthorized",
			reason:        "test",
			expectError:   true,
			errorContains: "not authorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Pre-revoke one credential for "some already revoked" test
			if tt.name == "batch with some already revoked" && len(tt.credentialIDs) > 1 {
				err := k.RevokeCredential(ctx, tt.credentialIDs[1], tt.did, owner, "pre-test", nil)
				require.NoError(t, err)
			}

			// Test
			err := k.BatchRevokeCredentials(ctx, tt.credentialIDs, tt.did, tt.revoker, tt.reason)

			// Verify
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				// Verify all credentials are revoked
				for _, credID := range tt.credentialIDs {
					if credID != "" {
						require.True(t, k.IsCredentialRevoked(ctx, credID),
							"credential %s should be revoked", credID)
					}
				}
			}
		})
	}
}

// TestVerifyCredential tests comprehensive credential verification
func TestVerifyCredential(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create active identity
	did := "did:aura:test-verify"
	owner := "aura1owner123456789"
	identity := &types.IdentityRecord{
		Did:             did,
		Address:         owner,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity))

	// Setup: Create erased identity
	erasedDID := "did:aura:test-erased"
	erasedIdentity := &types.IdentityRecord{
		Did:             erasedDID,
		Address:         "aura1erased",
		Status:          types.IdentityStatusErased,
		Erased:          true,
		ErasedAt:        timestamppb.New(time.Now()),
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 0,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, erasedIdentity))

	// Setup: Create suspended identity
	suspendedDID := "did:aura:test-suspended"
	suspendedIdentity := &types.IdentityRecord{
		Did:             suspendedDID,
		Address:         "aura1suspended",
		Status:          types.IdentityStatusSuspended,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 50,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, suspendedIdentity))

	tests := []struct {
		name          string
		credentialID  string
		did           string
		setupRevoke   bool
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid credential",
			credentialID: "verify-cred-001",
			did:          did,
			setupRevoke:  false,
			expectError:  false,
		},
		{
			name:          "revoked credential",
			credentialID:  "verify-cred-002",
			did:           did,
			setupRevoke:   true,
			expectError:   true,
			errorContains: "has been revoked",
		},
		{
			name:          "empty credential ID",
			credentialID:  "",
			did:           did,
			setupRevoke:   false,
			expectError:   true,
			errorContains: "credential ID cannot be empty",
		},
		{
			name:          "empty DID",
			credentialID:  "verify-cred-003",
			did:           "",
			setupRevoke:   false,
			expectError:   true,
			errorContains: "DID cannot be empty",
		},
		{
			name:          "non-existent identity",
			credentialID:  "verify-cred-004",
			did:           "did:aura:nonexistent",
			setupRevoke:   false,
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:          "erased identity",
			credentialID:  "verify-cred-005",
			did:           erasedDID,
			setupRevoke:   false,
			expectError:   true,
			errorContains: "has been erased",
		},
		{
			name:          "suspended identity",
			credentialID:  "verify-cred-006",
			did:           suspendedDID,
			setupRevoke:   false,
			expectError:   true,
			errorContains: "not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Revoke if needed
			if tt.setupRevoke && tt.credentialID != "" && tt.did != "" {
				err := k.RevokeCredential(ctx, tt.credentialID, tt.did, owner, "test revocation", nil)
				require.NoError(t, err)
			}

			// Test
			err := k.VerifyCredential(ctx, tt.credentialID, tt.did)

			// Verify
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestGetCredentialRevocation tests retrieving revocation details
func TestGetCredentialRevocation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create identity and revoke a credential
	did := "did:aura:test-get"
	owner := "aura1owner123456789"
	identity := &types.IdentityRecord{
		Did:             did,
		Address:         owner,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity))

	credID := "get-cred-001"
	reason := "security breach"
	metadata := map[string]string{
		"severity": "critical",
		"incident": "INC-2024-001",
	}

	err := k.RevokeCredential(ctx, credID, did, owner, reason, metadata)
	require.NoError(t, err)

	// Test: Get revocation
	revocation, err := k.GetCredentialRevocation(ctx, credID)
	require.NoError(t, err)
	require.NotNil(t, revocation)
	require.Equal(t, credID, revocation.CredentialId)
	require.Equal(t, did, revocation.Did)
	require.Equal(t, owner, revocation.RevokedBy)
	require.Equal(t, reason, revocation.Reason)
	require.Equal(t, metadata, revocation.Metadata)
	require.NotNil(t, revocation.RevokedAt)

	// Test: Get non-existent revocation
	_, err = k.GetCredentialRevocation(ctx, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")

	// Test: Empty credential ID
	_, err = k.GetCredentialRevocation(ctx, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be empty")
}

// TestGetAllCredentialRevocations tests retrieving all revocations
func TestGetAllCredentialRevocations(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create identity
	did := "did:aura:test-getall"
	owner := "aura1owner123456789"
	identity := &types.IdentityRecord{
		Did:             did,
		Address:         owner,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity))

	// Revoke multiple credentials
	credIDs := []string{
		"getall-cred-001",
		"getall-cred-002",
		"getall-cred-003",
	}

	for _, credID := range credIDs {
		err := k.RevokeCredential(ctx, credID, did, owner, "test", nil)
		require.NoError(t, err)
	}

	// Test
	revocations, err := k.GetAllCredentialRevocations(ctx)
	require.NoError(t, err)

	// Verify we have at least the ones we created
	require.GreaterOrEqual(t, len(revocations), len(credIDs))

	// Verify all our credentials are in the list
	foundCreds := make(map[string]bool)
	for _, rev := range revocations {
		foundCreds[rev.CredentialId] = true
	}
	for _, credID := range credIDs {
		require.True(t, foundCreds[credID], "credential %s not found in revocations", credID)
	}
}

// TestGetCredentialRevocationsByDID tests filtering revocations by DID
func TestGetCredentialRevocationsByDID(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create multiple identities
	did1 := "did:aura:test-filter-1"
	owner1 := "aura1owner1"
	identity1 := &types.IdentityRecord{
		Did:             did1,
		Address:         owner1,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity1))

	did2 := "did:aura:test-filter-2"
	owner2 := "aura1owner2"
	identity2 := &types.IdentityRecord{
		Did:             did2,
		Address:         owner2,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity2))

	// Revoke credentials for both DIDs
	did1Creds := []string{"filter-cred-1-1", "filter-cred-1-2"}
	did2Creds := []string{"filter-cred-2-1"}

	for _, credID := range did1Creds {
		err := k.RevokeCredential(ctx, credID, did1, owner1, "test", nil)
		require.NoError(t, err)
	}

	for _, credID := range did2Creds {
		err := k.RevokeCredential(ctx, credID, did2, owner2, "test", nil)
		require.NoError(t, err)
	}

	// Test: Get revocations for DID1
	revocations1, err := k.GetCredentialRevocationsByDID(ctx, did1)
	require.NoError(t, err)
	require.Len(t, revocations1, len(did1Creds))

	for _, rev := range revocations1 {
		require.Equal(t, did1, rev.Did)
	}

	// Test: Get revocations for DID2
	revocations2, err := k.GetCredentialRevocationsByDID(ctx, did2)
	require.NoError(t, err)
	require.Len(t, revocations2, len(did2Creds))

	for _, rev := range revocations2 {
		require.Equal(t, did2, rev.Did)
	}
}

// TestRestoreCredential tests credential restoration (admin function)
func TestRestoreCredential(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create identity
	did := "did:aura:test-restore"
	owner := "aura1owner123456789"
	identity := &types.IdentityRecord{
		Did:             did,
		Address:         owner,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity))

	// Setup: Create admin role and assign it
	adminRole := &types.Role{
		Name:         types.RoleAdmin,
		Permissions:  []string{types.PermissionAdmin},
		Description:  "Administrator",
		CreatedAt:    timestamppb.New(time.Now()),
		UpdatedAt:    timestamppb.New(time.Now()),
		IsSystemRole: true,
	}
	require.NoError(t, k.SetRole(ctx, adminRole))

	admin := "aura1admin123456789"
	assignment := &types.RoleAssignment{
		Address:    admin,
		RoleName:   types.RoleAdmin,
		AssignedAt: timestamppb.New(time.Now()),
		AssignedBy: admin,
	}
	require.NoError(t, k.SetRoleAssignment(ctx, assignment))

	// Setup: Revoke a credential
	credID := "restore-cred-001"
	err := k.RevokeCredential(ctx, credID, did, owner, "test revocation", nil)
	require.NoError(t, err)
	require.True(t, k.IsCredentialRevoked(ctx, credID))

	// Test: Restore credential
	err = k.RestoreCredential(ctx, credID, admin, "administrative correction")
	require.NoError(t, err)
	require.False(t, k.IsCredentialRevoked(ctx, credID))

	// Test: Restore non-revoked credential
	err = k.RestoreCredential(ctx, credID, admin, "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not revoked")

	// Test: Restore with unauthorized user
	err = k.RestoreCredential(ctx, "restore-cred-002", owner, "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not authorized")

	// Test: Restore with empty credential ID
	err = k.RestoreCredential(ctx, "", admin, "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be empty")
}

// TestCredentialRevocationIntegration tests end-to-end revocation workflow
func TestCredentialRevocationIntegration(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	
	

	// Setup: Create identity
	did := "did:aura:test-integration"
	owner := "aura1owner123456789"
	identity := &types.IdentityRecord{
		Did:             did,
		Address:         owner,
		Status:          types.IdentityStatusActive,
		CreatedAt:       timestamppb.New(time.Now()),
		UpdatedAt:       timestamppb.New(time.Now()),
		ConfidenceScore: 80,
		VerificationMethods: []string{
			"pubkey1",
			"pubkey2",
		},
	}
	require.NoError(t, k.SetIdentityRecord(ctx, identity))

	credID := "integration-cred-001"

	// Step 1: Verify credential is valid initially
	err := k.VerifyCredential(ctx, credID, did)
	require.NoError(t, err, "credential should be valid initially")
	require.False(t, k.IsCredentialRevoked(ctx, credID))

	// Step 2: Revoke the credential
	reason := "key compromise detected"
	metadata := map[string]string{
		"incident_id": "INC-2024-001",
		"severity":    "high",
	}
	err = k.RevokeCredential(ctx, credID, did, owner, reason, metadata)
	require.NoError(t, err, "revocation should succeed")

	// Step 3: Verify credential is now rejected
	err = k.VerifyCredential(ctx, credID, did)
	require.Error(t, err, "credential should be rejected after revocation")
	require.Contains(t, err.Error(), "has been revoked")
	require.True(t, k.IsCredentialRevoked(ctx, credID))

	// Step 4: Verify revocation details are correct
	revocation, err := k.GetCredentialRevocation(ctx, credID)
	require.NoError(t, err)
	require.Equal(t, credID, revocation.CredentialId)
	require.Equal(t, did, revocation.Did)
	require.Equal(t, owner, revocation.RevokedBy)
	require.Equal(t, reason, revocation.Reason)
	require.Equal(t, metadata, revocation.Metadata)

	// Step 5: Attempt to revoke again (should fail)
	err = k.RevokeCredential(ctx, credID, did, owner, "duplicate", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already revoked")

	// Step 6: Verify other credentials for same DID still work
	otherCredID := "integration-cred-002"
	err = k.VerifyCredential(ctx, otherCredID, did)
	require.NoError(t, err, "other credentials should still be valid")
}
