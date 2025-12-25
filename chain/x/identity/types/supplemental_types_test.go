// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRecoveryRecordReset(t *testing.T) {
	now := time.Now()
	record := &RecoveryRecord{
		DID:             "did:aura:test",
		RecoveryAddress: "aura1recovery",
		Status:          RecoveryStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
		Attempts:        1,
		MaxAttempts:     3,
		LastAttemptAt:   now,
		Metadata:        "test",
	}

	record.Reset()

	require.Equal(t, "", record.DID)
	require.Equal(t, "", record.RecoveryAddress)
	require.Equal(t, "", record.Status)
	require.Equal(t, uint32(0), record.Attempts)
	require.Equal(t, uint32(0), record.MaxAttempts)
}

func TestRecoveryRecordString(t *testing.T) {
	tests := []struct {
		name   string
		record *RecoveryRecord
		want   string
	}{
		{
			name:   "nil record",
			record: nil,
			want:   "nil",
		},
		{
			name: "valid record",
			record: &RecoveryRecord{
				DID:             "did:aura:test",
				RecoveryAddress: "aura1recovery",
				Status:          RecoveryStatusActive,
				Attempts:        1,
				MaxAttempts:     3,
			},
			want: "RecoveryRecord{DID:did:aura:test, RecoveryAddress:aura1recovery, Status:active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.record.String()
			require.Contains(t, result, tt.want)
		})
	}
}

func TestRecoveryRecordProtoMessage(t *testing.T) {
	record := &RecoveryRecord{}
	record.ProtoMessage() // Should not panic
}

func TestRecoveryRecordValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *RecoveryRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: &RecoveryRecord{
				DID:             "did:aura:test",
				RecoveryAddress: "aura1recovery",
				Status:          RecoveryStatusActive,
				Attempts:        1,
				MaxAttempts:     3,
			},
			wantErr: false,
		},
		{
			name: "empty DID",
			record: &RecoveryRecord{
				DID:             "",
				RecoveryAddress: "aura1recovery",
				Status:          RecoveryStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty recovery address",
			record: &RecoveryRecord{
				DID:             "did:aura:test",
				RecoveryAddress: "",
				Status:          RecoveryStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty status",
			record: &RecoveryRecord{
				DID:             "did:aura:test",
				RecoveryAddress: "aura1recovery",
				Status:          "",
			},
			wantErr: true,
		},
		{
			name: "attempts exceed max",
			record: &RecoveryRecord{
				DID:             "did:aura:test",
				RecoveryAddress: "aura1recovery",
				Status:          RecoveryStatusActive,
				Attempts:        5,
				MaxAttempts:     3,
			},
			wantErr: true,
		},
		{
			name: "attempts equal max",
			record: &RecoveryRecord{
				DID:             "did:aura:test",
				RecoveryAddress: "aura1recovery",
				Status:          RecoveryStatusActive,
				Attempts:        3,
				MaxAttempts:     3,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.record.CreatedAt.IsZero() {
				tt.record.CreatedAt = now
			}
			if tt.record.UpdatedAt.IsZero() {
				tt.record.UpdatedAt = now
			}

			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVerificationRecordReset(t *testing.T) {
	now := time.Now()
	record := &VerificationRecord{
		DID:          "did:aura:test",
		Level:        2,
		VerifiedAt:   now,
		VerifiedBy:   "aura1verifier",
		Method:       "kyc",
		ExpiresAt:    now.Add(365 * 24 * time.Hour),
		Status:       VerificationStatusVerified,
		Documents:    []string{"doc1", "doc2"},
		Attestations: []string{"att1", "att2"},
		Metadata:     "test",
	}

	record.Reset()

	require.Equal(t, "", record.DID)
	require.Equal(t, int32(0), record.Level)
	require.Equal(t, "", record.VerifiedBy)
	require.Nil(t, record.Documents)
	require.Nil(t, record.Attestations)
}

func TestVerificationRecordString(t *testing.T) {
	tests := []struct {
		name   string
		record *VerificationRecord
		want   string
	}{
		{
			name:   "nil record",
			record: nil,
			want:   "nil",
		},
		{
			name: "valid record",
			record: &VerificationRecord{
				DID:          "did:aura:test",
				Level:        2,
				VerifiedBy:   "aura1verifier",
				Method:       "kyc",
				Status:       VerificationStatusVerified,
				Documents:    []string{"doc1"},
				Attestations: []string{"att1"},
			},
			want: "VerificationRecord{DID:did:aura:test, Level:2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.record.String()
			require.Contains(t, result, tt.want)
		})
	}
}

func TestVerificationRecordProtoMessage(t *testing.T) {
	record := &VerificationRecord{}
	record.ProtoMessage() // Should not panic
}

func TestVerificationRecordValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *VerificationRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: &VerificationRecord{
				DID:        "did:aura:test",
				Level:      2,
				VerifiedBy: "aura1verifier",
				Status:     VerificationStatusVerified,
			},
			wantErr: false,
		},
		{
			name: "empty DID",
			record: &VerificationRecord{
				DID:        "",
				Level:      2,
				VerifiedBy: "aura1verifier",
				Status:     VerificationStatusVerified,
			},
			wantErr: true,
		},
		{
			name: "negative level",
			record: &VerificationRecord{
				DID:        "did:aura:test",
				Level:      -1,
				VerifiedBy: "aura1verifier",
				Status:     VerificationStatusVerified,
			},
			wantErr: true,
		},
		{
			name: "empty verified_by",
			record: &VerificationRecord{
				DID:        "did:aura:test",
				Level:      2,
				VerifiedBy: "",
				Status:     VerificationStatusVerified,
			},
			wantErr: true,
		},
		{
			name: "empty status",
			record: &VerificationRecord{
				DID:        "did:aura:test",
				Level:      2,
				VerifiedBy: "aura1verifier",
				Status:     "",
			},
			wantErr: true,
		},
		{
			name: "zero level is valid",
			record: &VerificationRecord{
				DID:        "did:aura:test",
				Level:      0,
				VerifiedBy: "aura1verifier",
				Status:     VerificationStatusPending,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.record.VerifiedAt.IsZero() {
				tt.record.VerifiedAt = now
			}

			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDelegationRecordReset(t *testing.T) {
	now := time.Now()
	record := &DelegationRecord{
		DID:         "did:aura:test",
		DelegatedTo: "aura1delegate",
		Permissions: []string{"read", "write"},
		CreatedAt:   now,
		ExpiresAt:   now.Add(24 * time.Hour),
		Status:      DelegationStatusActive,
		CanRevoke:   true,
		Metadata:    "test",
	}

	record.Reset()

	require.Equal(t, "", record.DID)
	require.Equal(t, "", record.DelegatedTo)
	require.Nil(t, record.Permissions)
	require.False(t, record.CanRevoke)
}

func TestDelegationRecordString(t *testing.T) {
	tests := []struct {
		name   string
		record *DelegationRecord
		want   string
	}{
		{
			name:   "nil record",
			record: nil,
			want:   "nil",
		},
		{
			name: "valid record",
			record: &DelegationRecord{
				DID:         "did:aura:test",
				DelegatedTo: "aura1delegate",
				Permissions: []string{"read", "write"},
				Status:      DelegationStatusActive,
				CanRevoke:   true,
			},
			want: "DelegationRecord{DID:did:aura:test, DelegatedTo:aura1delegate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.record.String()
			require.Contains(t, result, tt.want)
		})
	}
}

func TestDelegationRecordProtoMessage(t *testing.T) {
	record := &DelegationRecord{}
	record.ProtoMessage() // Should not panic
}

func TestDelegationRecordValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *DelegationRecord
		wantErr bool
	}{
		{
			name: "valid record",
			record: &DelegationRecord{
				DID:         "did:aura:test",
				DelegatedTo: "aura1delegate",
				Permissions: []string{"read", "write"},
				Status:      DelegationStatusActive,
			},
			wantErr: false,
		},
		{
			name: "empty DID",
			record: &DelegationRecord{
				DID:         "",
				DelegatedTo: "aura1delegate",
				Permissions: []string{"read"},
				Status:      DelegationStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty delegated_to",
			record: &DelegationRecord{
				DID:         "did:aura:test",
				DelegatedTo: "",
				Permissions: []string{"read"},
				Status:      DelegationStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty permissions",
			record: &DelegationRecord{
				DID:         "did:aura:test",
				DelegatedTo: "aura1delegate",
				Permissions: []string{},
				Status:      DelegationStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty status",
			record: &DelegationRecord{
				DID:         "did:aura:test",
				DelegatedTo: "aura1delegate",
				Permissions: []string{"read"},
				Status:      "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.record.CreatedAt.IsZero() {
				tt.record.CreatedAt = now
			}

			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFederationRecordReset(t *testing.T) {
	now := time.Now()
	record := &FederationRecord{
		DID:            "did:aura:test",
		FederatedChain: "cosmos-hub",
		ExternalDID:    "did:cosmos:external",
		Status:         FederationStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		Verified:       true,
		VerifiedBy:     "aura1verifier",
		VerifiedAt:     now,
		ProofHash:      "hash123",
		Metadata:       "test",
	}

	record.Reset()

	require.Equal(t, "", record.DID)
	require.Equal(t, "", record.FederatedChain)
	require.Equal(t, "", record.ExternalDID)
	require.False(t, record.Verified)
}

func TestFederationRecordString(t *testing.T) {
	tests := []struct {
		name   string
		record *FederationRecord
		want   string
	}{
		{
			name:   "nil record",
			record: nil,
			want:   "nil",
		},
		{
			name: "valid record",
			record: &FederationRecord{
				DID:            "did:aura:test",
				FederatedChain: "cosmos-hub",
				ExternalDID:    "did:cosmos:external",
				Status:         FederationStatusActive,
				Verified:       true,
			},
			want: "FederationRecord{DID:did:aura:test, FederatedChain:cosmos-hub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.record.String()
			require.Contains(t, result, tt.want)
		})
	}
}

func TestFederationRecordProtoMessage(t *testing.T) {
	record := &FederationRecord{}
	record.ProtoMessage() // Should not panic
}

func TestFederationRecordValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *FederationRecord
		wantErr bool
	}{
		{
			name: "valid record - not verified",
			record: &FederationRecord{
				DID:            "did:aura:test",
				FederatedChain: "cosmos-hub",
				ExternalDID:    "did:cosmos:external",
				Status:         FederationStatusPending,
				Verified:       false,
			},
			wantErr: false,
		},
		{
			name: "valid record - verified",
			record: &FederationRecord{
				DID:            "did:aura:test",
				FederatedChain: "cosmos-hub",
				ExternalDID:    "did:cosmos:external",
				Status:         FederationStatusActive,
				Verified:       true,
				VerifiedBy:     "aura1verifier",
			},
			wantErr: false,
		},
		{
			name: "empty DID",
			record: &FederationRecord{
				DID:            "",
				FederatedChain: "cosmos-hub",
				ExternalDID:    "did:cosmos:external",
				Status:         FederationStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty federated_chain",
			record: &FederationRecord{
				DID:            "did:aura:test",
				FederatedChain: "",
				ExternalDID:    "did:cosmos:external",
				Status:         FederationStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty external_did",
			record: &FederationRecord{
				DID:            "did:aura:test",
				FederatedChain: "cosmos-hub",
				ExternalDID:    "",
				Status:         FederationStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty status",
			record: &FederationRecord{
				DID:            "did:aura:test",
				FederatedChain: "cosmos-hub",
				ExternalDID:    "did:cosmos:external",
				Status:         "",
			},
			wantErr: true,
		},
		{
			name: "verified but no verified_by",
			record: &FederationRecord{
				DID:            "did:aura:test",
				FederatedChain: "cosmos-hub",
				ExternalDID:    "did:cosmos:external",
				Status:         FederationStatusActive,
				Verified:       true,
				VerifiedBy:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.record.CreatedAt.IsZero() {
				tt.record.CreatedAt = now
			}
			if tt.record.UpdatedAt.IsZero() {
				tt.record.UpdatedAt = now
			}

			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCrossChainLinkReset(t *testing.T) {
	now := time.Now()
	link := &CrossChainLink{
		DID:           "did:aura:test",
		TargetChain:   "cosmos-hub",
		TargetDID:     "did:cosmos:target",
		LinkType:      "identity",
		Status:        CrossChainLinkStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
		ConfirmedAt:   now,
		Confirmed:     true,
		ProofData:     "proof",
		ProofHash:     "hash",
		Bidirectional: true,
		RelayAddress:  "aura1relay",
		Metadata:      "test",
	}

	link.Reset()

	require.Equal(t, "", link.DID)
	require.Equal(t, "", link.TargetChain)
	require.Equal(t, "", link.TargetDID)
	require.False(t, link.Confirmed)
	require.False(t, link.Bidirectional)
}

func TestCrossChainLinkString(t *testing.T) {
	tests := []struct {
		name string
		link *CrossChainLink
		want string
	}{
		{
			name: "nil link",
			link: nil,
			want: "nil",
		},
		{
			name: "valid link",
			link: &CrossChainLink{
				DID:           "did:aura:test",
				TargetChain:   "cosmos-hub",
				TargetDID:     "did:cosmos:target",
				LinkType:      "identity",
				Status:        CrossChainLinkStatusActive,
				Confirmed:     true,
				Bidirectional: false,
			},
			want: "CrossChainLink{DID:did:aura:test, TargetChain:cosmos-hub",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.link.String()
			require.Contains(t, result, tt.want)
		})
	}
}

func TestCrossChainLinkProtoMessage(t *testing.T) {
	link := &CrossChainLink{}
	link.ProtoMessage() // Should not panic
}

func TestCrossChainLinkValidate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		link    *CrossChainLink
		wantErr bool
	}{
		{
			name: "valid unidirectional link",
			link: &CrossChainLink{
				DID:           "did:aura:test",
				TargetChain:   "cosmos-hub",
				TargetDID:     "did:cosmos:target",
				LinkType:      "identity",
				Status:        CrossChainLinkStatusActive,
				Bidirectional: false,
			},
			wantErr: false,
		},
		{
			name: "valid bidirectional link",
			link: &CrossChainLink{
				DID:           "did:aura:test",
				TargetChain:   "cosmos-hub",
				TargetDID:     "did:cosmos:target",
				LinkType:      "identity",
				Status:        CrossChainLinkStatusActive,
				Bidirectional: true,
				RelayAddress:  "aura1relay",
			},
			wantErr: false,
		},
		{
			name: "empty DID",
			link: &CrossChainLink{
				DID:         "",
				TargetChain: "cosmos-hub",
				TargetDID:   "did:cosmos:target",
				LinkType:    "identity",
				Status:      CrossChainLinkStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty target_chain",
			link: &CrossChainLink{
				DID:         "did:aura:test",
				TargetChain: "",
				TargetDID:   "did:cosmos:target",
				LinkType:    "identity",
				Status:      CrossChainLinkStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty target_did",
			link: &CrossChainLink{
				DID:         "did:aura:test",
				TargetChain: "cosmos-hub",
				TargetDID:   "",
				LinkType:    "identity",
				Status:      CrossChainLinkStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty link_type",
			link: &CrossChainLink{
				DID:         "did:aura:test",
				TargetChain: "cosmos-hub",
				TargetDID:   "did:cosmos:target",
				LinkType:    "",
				Status:      CrossChainLinkStatusActive,
			},
			wantErr: true,
		},
		{
			name: "empty status",
			link: &CrossChainLink{
				DID:         "did:aura:test",
				TargetChain: "cosmos-hub",
				TargetDID:   "did:cosmos:target",
				LinkType:    "identity",
				Status:      "",
			},
			wantErr: true,
		},
		{
			name: "bidirectional but no relay_address",
			link: &CrossChainLink{
				DID:           "did:aura:test",
				TargetChain:   "cosmos-hub",
				TargetDID:     "did:cosmos:target",
				LinkType:      "identity",
				Status:        CrossChainLinkStatusActive,
				Bidirectional: true,
				RelayAddress:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.link.CreatedAt.IsZero() {
				tt.link.CreatedAt = now
			}

			err := tt.link.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
