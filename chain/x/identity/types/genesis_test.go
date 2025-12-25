// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := DefaultGenesisState()

	require.NotNil(t, gs)
	require.NotNil(t, gs.Params.Auth)
	require.NotNil(t, gs.Params.Change)
	require.Equal(t, uint64(1), gs.NextAuditLogId)
	require.False(t, gs.IdentityChangesSuspended)
	require.Empty(t, gs.Roles)
	require.Empty(t, gs.IdentityRecords)
}

func TestValidateGenesisState_NilState(t *testing.T) {
	err := ValidateGenesisState(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be nil")
}

func TestValidateGenesisState_Valid(t *testing.T) {
	gs := DefaultGenesisState()
	err := ValidateGenesisState(gs)
	require.NoError(t, err)
}

func TestValidateGenesisState_InvalidParams(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*identitypb.GenesisState)
		wantErr string
	}{
		{
			name: "zero max roles per account",
			modify: func(gs *identitypb.GenesisState) {
				gs.Params.Auth.MaxRolesPerAccount = 0
			},
			wantErr: "max_roles_per_account",
		},
		{
			name: "zero requests per minute",
			modify: func(gs *identitypb.GenesisState) {
				gs.Params.Auth.DefaultRequestsPerMinute = 0
			},
			wantErr: "default_requests_per_minute",
		},
		{
			name: "zero max requests per wallet per month",
			modify: func(gs *identitypb.GenesisState) {
				gs.Params.Change.MaxRequestsPerWalletPerMonth = 0
			},
			wantErr: "max_requests_per_wallet_per_month",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			tc.modify(gs)
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_Roles(t *testing.T) {
	tests := []struct {
		name    string
		roles   []identitypb.Role
		wantErr string
	}{
		{
			name: "empty role name",
			roles: []identitypb.Role{
				{Name: "", Permissions: []string{"read"}},
			},
			wantErr: "role name cannot be empty",
		},
		{
			name: "duplicate role",
			roles: []identitypb.Role{
				{Name: "admin", Permissions: []string{"read"}},
				{Name: "admin", Permissions: []string{"write"}},
			},
			wantErr: "duplicate role",
		},
		{
			name: "role with no permissions",
			roles: []identitypb.Role{
				{Name: "empty", Permissions: []string{}},
			},
			wantErr: "has no permissions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			gs.Roles = tc.roles
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_RoleAssignments(t *testing.T) {
	tests := []struct {
		name        string
		roles       []identitypb.Role
		assignments []identitypb.RoleAssignment
		wantErr     string
	}{
		{
			name: "empty address",
			assignments: []identitypb.RoleAssignment{
				{Address: "", RoleName: "admin"},
			},
			wantErr: "address cannot be empty",
		},
		{
			name: "empty role name",
			assignments: []identitypb.RoleAssignment{
				{Address: "aura1xxx", RoleName: ""},
			},
			wantErr: "role name cannot be empty",
		},
		{
			name: "role not found",
			assignments: []identitypb.RoleAssignment{
				{Address: "aura1xxx", RoleName: "nonexistent"},
			},
			wantErr: "role not found",
		},
		{
			name: "duplicate assignment",
			roles: []identitypb.Role{
				{Name: "moderator", Permissions: []string{"read"}},
			},
			assignments: []identitypb.RoleAssignment{
				{Address: "aura1xxx", RoleName: "moderator"},
				{Address: "aura1xxx", RoleName: "moderator"},
			},
			wantErr: "duplicate role assignment",
		},
		{
			name: "valid assignment with builtin role",
			assignments: []identitypb.RoleAssignment{
				{Address: "aura1xxx", RoleName: RoleAdmin},
				{Address: "aura1yyy", RoleName: RoleUser},
			},
			wantErr: "", // Should pass
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			gs.Roles = tc.roles
			gs.RoleAssignments = tc.assignments
			err := ValidateGenesisState(gs)
			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestValidateGenesisState_MultisigWallets(t *testing.T) {
	tests := []struct {
		name    string
		wallets []identitypb.MultisigWallet
		wantErr string
	}{
		{
			name: "empty wallet ID",
			wallets: []identitypb.MultisigWallet{
				{Id: "", Threshold: 1, Signers: []string{"a"}},
			},
			wantErr: "wallet ID cannot be empty",
		},
		{
			name: "duplicate wallet ID",
			wallets: []identitypb.MultisigWallet{
				{Id: "wallet1", Threshold: 1, Signers: []string{"a"}},
				{Id: "wallet1", Threshold: 2, Signers: []string{"b", "c"}},
			},
			wantErr: "duplicate wallet ID",
		},
		{
			name: "zero threshold",
			wallets: []identitypb.MultisigWallet{
				{Id: "wallet1", Threshold: 0, Signers: []string{"a"}},
			},
			wantErr: "threshold must be greater than 0",
		},
		{
			name: "threshold exceeds signers",
			wallets: []identitypb.MultisigWallet{
				{Id: "wallet1", Threshold: 3, Signers: []string{"a", "b"}},
			},
			wantErr: "threshold (3) cannot exceed number of signers (2)",
		},
		{
			name: "empty signer address",
			wallets: []identitypb.MultisigWallet{
				{Id: "wallet1", Threshold: 1, Signers: []string{"a", ""}},
			},
			wantErr: "signer address cannot be empty",
		},
		{
			name: "duplicate signer",
			wallets: []identitypb.MultisigWallet{
				{Id: "wallet1", Threshold: 2, Signers: []string{"a", "a"}},
			},
			wantErr: "duplicate signer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			gs.MultisigWallets = tc.wallets
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_IdentityRecords(t *testing.T) {
	tests := []struct {
		name    string
		records []identitypb.IdentityRecord
		wantErr string
	}{
		{
			name: "empty DID",
			records: []identitypb.IdentityRecord{
				{Did: "", Address: "aura1xxx", VerificationMethods: []string{"key1"}},
			},
			wantErr: "DID cannot be empty",
		},
		{
			name: "duplicate DID",
			records: []identitypb.IdentityRecord{
				{Did: "did:aura:1", Address: "aura1xxx", VerificationMethods: []string{"key1"}},
				{Did: "did:aura:1", Address: "aura1yyy", VerificationMethods: []string{"key2"}},
			},
			wantErr: "duplicate DID",
		},
		{
			name: "empty address",
			records: []identitypb.IdentityRecord{
				{Did: "did:aura:1", Address: "", VerificationMethods: []string{"key1"}},
			},
			wantErr: "identity address cannot be empty",
		},
		{
			name: "duplicate address",
			records: []identitypb.IdentityRecord{
				{Did: "did:aura:1", Address: "aura1xxx", VerificationMethods: []string{"key1"}},
				{Did: "did:aura:2", Address: "aura1xxx", VerificationMethods: []string{"key2"}},
			},
			wantErr: "duplicate identity address",
		},
		{
			name: "no verification methods",
			records: []identitypb.IdentityRecord{
				{Did: "did:aura:1", Address: "aura1xxx", VerificationMethods: []string{}},
			},
			wantErr: "must have at least one verification method",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			gs.IdentityRecords = tc.records
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_Sessions(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		sessions []identitypb.Session
		wantErr  string
	}{
		{
			name: "empty session ID",
			sessions: []identitypb.Session{
				{Id: "", Address: "aura1xxx", CreatedAt: now, ExpiresAt: now.Add(time.Hour)},
			},
			wantErr: "session ID cannot be empty",
		},
		{
			name: "duplicate session ID",
			sessions: []identitypb.Session{
				{Id: "sess1", Address: "aura1xxx", CreatedAt: now, ExpiresAt: now.Add(time.Hour)},
				{Id: "sess1", Address: "aura1yyy", CreatedAt: now, ExpiresAt: now.Add(time.Hour)},
			},
			wantErr: "duplicate session ID",
		},
		{
			name: "empty address",
			sessions: []identitypb.Session{
				{Id: "sess1", Address: "", CreatedAt: now, ExpiresAt: now.Add(time.Hour)},
			},
			wantErr: "session address cannot be empty",
		},
		{
			name: "expiry before creation",
			sessions: []identitypb.Session{
				{Id: "sess1", Address: "aura1xxx", CreatedAt: now, ExpiresAt: now.Add(-time.Hour)},
			},
			wantErr: "expiry must be after creation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			gs.Sessions = tc.sessions
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_ChangeRequests(t *testing.T) {
	// Setup with identity record
	gs := DefaultGenesisState()
	gs.IdentityRecords = []identitypb.IdentityRecord{
		{Did: "did:aura:1", Address: "aura1xxx", VerificationMethods: []string{"key1"}},
	}

	tests := []struct {
		name     string
		requests []identitypb.ChangeRequest
		wantErr  string
	}{
		{
			name: "empty request ID",
			requests: []identitypb.ChangeRequest{
				{Id: "", Requester: "aura1xxx", Did: "did:aura:1"},
			},
			wantErr: "request ID cannot be empty",
		},
		{
			name: "duplicate request ID",
			requests: []identitypb.ChangeRequest{
				{Id: "req1", Requester: "aura1xxx", Did: "did:aura:1"},
				{Id: "req1", Requester: "aura1yyy", Did: "did:aura:1"},
			},
			wantErr: "duplicate request ID",
		},
		{
			name: "empty requester",
			requests: []identitypb.ChangeRequest{
				{Id: "req1", Requester: "", Did: "did:aura:1"},
			},
			wantErr: "requester cannot be empty",
		},
		{
			name: "empty DID",
			requests: []identitypb.ChangeRequest{
				{Id: "req1", Requester: "aura1xxx", Did: ""},
			},
			wantErr: "target DID cannot be empty",
		},
		{
			name: "non-existent DID",
			requests: []identitypb.ChangeRequest{
				{Id: "req1", Requester: "aura1xxx", Did: "did:aura:nonexistent"},
			},
			wantErr: "references non-existent DID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testGs := DefaultGenesisState()
			testGs.IdentityRecords = gs.IdentityRecords
			testGs.ChangeRequests = tc.requests
			err := ValidateGenesisState(testGs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_CredentialRevocations(t *testing.T) {
	gs := DefaultGenesisState()
	gs.IdentityRecords = []identitypb.IdentityRecord{
		{Did: "did:aura:1", Address: "aura1xxx", VerificationMethods: []string{"key1"}},
	}

	tests := []struct {
		name        string
		revocations []identitypb.CredentialRevocation
		wantErr     string
	}{
		{
			name: "empty credential ID",
			revocations: []identitypb.CredentialRevocation{
				{CredentialId: "", Did: "did:aura:1"},
			},
			wantErr: "credential ID cannot be empty",
		},
		{
			name: "duplicate credential ID",
			revocations: []identitypb.CredentialRevocation{
				{CredentialId: "cred1", Did: "did:aura:1"},
				{CredentialId: "cred1", Did: "did:aura:1"},
			},
			wantErr: "duplicate credential revocation",
		},
		{
			name: "empty DID",
			revocations: []identitypb.CredentialRevocation{
				{CredentialId: "cred1", Did: ""},
			},
			wantErr: "DID cannot be empty",
		},
		{
			name: "non-existent DID",
			revocations: []identitypb.CredentialRevocation{
				{CredentialId: "cred1", Did: "did:aura:nonexistent"},
			},
			wantErr: "references non-existent DID",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testGs := DefaultGenesisState()
			testGs.IdentityRecords = gs.IdentityRecords
			testGs.CredentialRevocations = tc.revocations
			err := ValidateGenesisState(testGs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_DIDKeyRotations(t *testing.T) {
	gs := DefaultGenesisState()
	gs.IdentityRecords = []identitypb.IdentityRecord{
		{Did: "did:aura:1", Address: "aura1xxx", VerificationMethods: []string{"key1"}},
	}

	tests := []struct {
		name      string
		rotations []identitypb.DIDKeyRotation
		wantErr   string
	}{
		{
			name: "empty DID",
			rotations: []identitypb.DIDKeyRotation{
				{Did: "", OldVerificationMethod: "old", NewVerificationMethod: "new"},
			},
			wantErr: "DID cannot be empty",
		},
		{
			name: "non-existent DID",
			rotations: []identitypb.DIDKeyRotation{
				{Did: "did:aura:nonexistent", OldVerificationMethod: "old", NewVerificationMethod: "new"},
			},
			wantErr: "references non-existent DID",
		},
		{
			name: "duplicate pending rotations",
			rotations: []identitypb.DIDKeyRotation{
				{Did: "did:aura:1", OldVerificationMethod: "old1", NewVerificationMethod: "new1", Status: identitypb.DIDKeyRotationStatus_DID_KEY_ROTATION_STATUS_PENDING},
				{Did: "did:aura:1", OldVerificationMethod: "old2", NewVerificationMethod: "new2", Status: identitypb.DIDKeyRotationStatus_DID_KEY_ROTATION_STATUS_PENDING},
			},
			wantErr: "duplicate pending key rotation",
		},
		{
			name: "empty new verification method",
			rotations: []identitypb.DIDKeyRotation{
				{Did: "did:aura:1", OldVerificationMethod: "old", NewVerificationMethod: ""},
			},
			wantErr: "new verification method cannot be empty",
		},
		{
			name: "empty old verification method",
			rotations: []identitypb.DIDKeyRotation{
				{Did: "did:aura:1", OldVerificationMethod: "", NewVerificationMethod: "new"},
			},
			wantErr: "old verification method cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testGs := DefaultGenesisState()
			testGs.IdentityRecords = gs.IdentityRecords
			testGs.DidKeyRotations = tc.rotations
			err := ValidateGenesisState(testGs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_MultisigProposals(t *testing.T) {
	gs := DefaultGenesisState()
	gs.MultisigWallets = []identitypb.MultisigWallet{
		{Id: "wallet1", Threshold: 2, Signers: []string{"a", "b", "c"}},
	}

	tests := []struct {
		name      string
		proposals []identitypb.MultisigProposal
		wantErr   string
	}{
		{
			name: "empty proposal ID",
			proposals: []identitypb.MultisigProposal{
				{Id: "", WalletId: "wallet1"},
			},
			wantErr: "proposal ID cannot be empty",
		},
		{
			name: "duplicate proposal ID",
			proposals: []identitypb.MultisigProposal{
				{Id: "prop1", WalletId: "wallet1"},
				{Id: "prop1", WalletId: "wallet1"},
			},
			wantErr: "duplicate proposal ID",
		},
		{
			name: "empty wallet ID",
			proposals: []identitypb.MultisigProposal{
				{Id: "prop1", WalletId: ""},
			},
			wantErr: "wallet ID cannot be empty",
		},
		{
			name: "non-existent wallet",
			proposals: []identitypb.MultisigProposal{
				{Id: "prop1", WalletId: "nonexistent"},
			},
			wantErr: "references non-existent wallet",
		},
		{
			name: "duplicate signers",
			proposals: []identitypb.MultisigProposal{
				{Id: "prop1", WalletId: "wallet1", Signatures: []string{"a", "a"}},
			},
			wantErr: "duplicate signers",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testGs := DefaultGenesisState()
			testGs.MultisigWallets = gs.MultisigWallets
			testGs.MultisigProposals = tc.proposals
			err := ValidateGenesisState(testGs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_EmergencyAdmins(t *testing.T) {
	tests := []struct {
		name    string
		admins  []identitypb.EmergencyAdmin
		wantErr string
	}{
		{
			name: "empty address",
			admins: []identitypb.EmergencyAdmin{
				{Address: ""},
			},
			wantErr: "admin address cannot be empty",
		},
		{
			name: "duplicate address",
			admins: []identitypb.EmergencyAdmin{
				{Address: "aura1xxx"},
				{Address: "aura1xxx"},
			},
			wantErr: "duplicate emergency admin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			gs.EmergencyAdmins = tc.admins
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_TimeLockedActions(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		actions []identitypb.TimeLockedAction
		wantErr string
	}{
		{
			name: "empty action ID",
			actions: []identitypb.TimeLockedAction{
				{Id: "", Proposer: "aura1xxx", ProposedAt: now, ExecutableAt: now.Add(time.Hour)},
			},
			wantErr: "action ID cannot be empty",
		},
		{
			name: "duplicate action ID",
			actions: []identitypb.TimeLockedAction{
				{Id: "action1", Proposer: "aura1xxx", ProposedAt: now, ExecutableAt: now.Add(time.Hour)},
				{Id: "action1", Proposer: "aura1yyy", ProposedAt: now, ExecutableAt: now.Add(time.Hour)},
			},
			wantErr: "duplicate time-locked action ID",
		},
		{
			name: "empty proposer",
			actions: []identitypb.TimeLockedAction{
				{Id: "action1", Proposer: "", ProposedAt: now, ExecutableAt: now.Add(time.Hour)},
			},
			wantErr: "proposer cannot be empty",
		},
		{
			name: "executable before proposed",
			actions: []identitypb.TimeLockedAction{
				{Id: "action1", Proposer: "aura1xxx", ProposedAt: now, ExecutableAt: now.Add(-time.Hour)},
			},
			wantErr: "executable time must be after proposal time",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gs := DefaultGenesisState()
			gs.TimeLockedActions = tc.actions
			err := ValidateGenesisState(gs)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestValidateGenesisState_AuditLogCounter(t *testing.T) {
	gs := DefaultGenesisState()
	gs.NextAuditLogId = 0
	err := ValidateGenesisState(gs)
	require.Error(t, err)
	require.Contains(t, err.Error(), "next audit log ID must be at least 1")
}

func TestValidateGenesisState_ComplexValid(t *testing.T) {
	// Test a complex but valid genesis state
	now := time.Now()
	gs := DefaultGenesisState()

	gs.Roles = []identitypb.Role{
		{Name: "moderator", Permissions: []string{"read", "write"}},
		{Name: "auditor", Permissions: []string{"read"}},
	}

	gs.RoleAssignments = []identitypb.RoleAssignment{
		{Address: "aura1admin", RoleName: RoleAdmin},
		{Address: "aura1mod", RoleName: "moderator"},
	}

	gs.IdentityRecords = []identitypb.IdentityRecord{
		{Did: "did:aura:1", Address: "aura1user1", VerificationMethods: []string{"key1"}},
		{Did: "did:aura:2", Address: "aura1user2", VerificationMethods: []string{"key2"}},
	}

	gs.MultisigWallets = []identitypb.MultisigWallet{
		{Id: "wallet1", Threshold: 2, Signers: []string{"aura1a", "aura1b", "aura1c"}},
	}

	gs.Sessions = []identitypb.Session{
		{Id: "sess1", Address: "aura1user1", CreatedAt: now, ExpiresAt: now.Add(24 * time.Hour)},
	}

	err := ValidateGenesisState(gs)
	require.NoError(t, err)
}
