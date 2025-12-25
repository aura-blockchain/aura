// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVCType_Constants(t *testing.T) {
	// Test that VC type constants are properly defined
	vcTypes := []VCType{
		VCType_VC_TYPE_VERIFIED_HUMAN,
		VCType_VC_TYPE_AGE_OVER_18,
		VCType_VC_TYPE_AGE_OVER_21,
		VCType_VC_TYPE_RESIDENT_OF,
		VCType_VC_TYPE_BIOMETRIC_AUTH,
		VCType_VC_TYPE_KYC_VERIFICATION,
		VCType_VC_TYPE_NOTARY_PUBLIC,
		VCType_VC_TYPE_PROFESSIONAL_LICENSE,
	}

	// Ensure all VC types are unique
	for i, vt1 := range vcTypes {
		for j, vt2 := range vcTypes {
			if i != j {
				require.NotEqual(t, vt1, vt2)
			}
		}
	}
}

func TestVCType_FocusTypes(t *testing.T) {
	// Test arena focus VC types
	focusTypes := []VCType{
		VCType_VC_TYPE_BIOMETRIC_FOCUS,
		VCType_VC_TYPE_SOCIAL_FOCUS,
		VCType_VC_TYPE_GEOLOCATION_FOCUS,
		VCType_VC_TYPE_HIGH_ASSURANCE_FOCUS,
		VCType_VC_TYPE_POSSESSION_FOCUS,
		VCType_VC_TYPE_KNOWLEDGE_FOCUS,
		VCType_VC_TYPE_PERSISTENCE_FOCUS,
		VCType_VC_TYPE_SPECIALIZED_FOCUS,
	}

	// Ensure all focus types are unique
	for i, ft1 := range focusTypes {
		for j, ft2 := range focusTypes {
			if i != j {
				require.NotEqual(t, ft1, ft2)
			}
		}
	}
}

func TestVCStatus_Constants(t *testing.T) {
	// Test that VC status constants are properly defined
	statuses := []VCStatus{
		VCStatus_VC_STATUS_PENDING,
		VCStatus_VC_STATUS_ACTIVE,
		VCStatus_VC_STATUS_REVOKED,
		VCStatus_VC_STATUS_EXPIRED,
		VCStatus_VC_STATUS_SUSPENDED,
	}

	// Ensure all statuses are unique
	for i, s1 := range statuses {
		for j, s2 := range statuses {
			if i != j {
				require.NotEqual(t, s1, s2)
			}
		}
	}
}

func TestRevocationReason_Constants(t *testing.T) {
	// Test that revocation reason constants are properly defined
	reasons := []RevocationReason{
		RevocationReason_REVOCATION_REASON_USER_REQUEST,
		RevocationReason_REVOCATION_REASON_FRAUD_DETECTED,
		RevocationReason_REVOCATION_REASON_CS_BELOW_THRESHOLD,
		RevocationReason_REVOCATION_REASON_IR_INVALIDATED,
		RevocationReason_REVOCATION_REASON_EXPIRED,
		RevocationReason_REVOCATION_REASON_GOVERNANCE,
		RevocationReason_REVOCATION_REASON_SECURITY_COMPROMISE,
		RevocationReason_REVOCATION_REASON_POLICY_CHANGE,
	}

	// Ensure all reasons are unique
	for i, r1 := range reasons {
		for j, r2 := range reasons {
			if i != j {
				require.NotEqual(t, r1, r2)
			}
		}
	}
}

func TestVCPolicyStatus_Constants(t *testing.T) {
	// Test that VC policy status constants are properly defined
	statuses := []VCPolicyStatus{
		VCPolicyStatus_VC_POLICY_STATUS_DRAFT,
		VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED,
	}

	// Ensure all statuses are unique
	for i, s1 := range statuses {
		for j, s2 := range statuses {
			if i != j {
				require.NotEqual(t, s1, s2)
			}
		}
	}
}

func TestDisclosurePolicyMode_Constants(t *testing.T) {
	// Test that disclosure policy mode constants are properly defined
	modes := []DisclosurePolicyMode{
		DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY,
		DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ASK,
		DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
		DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_CONDITIONAL,
	}

	// Ensure all modes are unique
	for i, m1 := range modes {
		for j, m2 := range modes {
			if i != j {
				require.NotEqual(t, m1, m2)
			}
		}
	}
}

func TestVCRecord_Fields(t *testing.T) {
	record := &VCRecord{
		VcId:          "vc123",
		VcType:        VCType_VC_TYPE_VERIFIED_HUMAN,
		HolderDid:     "did:aura:holder1",
		HolderAddress: "aura1test",
		Status:        VCStatus_VC_STATUS_ACTIVE,
		CsAtMint:      75,
	}

	require.Equal(t, "vc123", record.VcId)
	require.Equal(t, VCType_VC_TYPE_VERIFIED_HUMAN, record.VcType)
	require.Equal(t, "did:aura:holder1", record.HolderDid)
	require.Equal(t, "aura1test", record.HolderAddress)
	require.Equal(t, VCStatus_VC_STATUS_ACTIVE, record.Status)
	require.Equal(t, uint64(75), record.CsAtMint)
}

func TestVCPolicy_Fields(t *testing.T) {
	policy := &VCPolicy{
		VcTypeName:            "VerifiedHuman",
		VcTypeEnum:            VCType_VC_TYPE_VERIFIED_HUMAN,
		CsThreshold:           50,
		ExpiryDurationDays:    365,
		Singleton:             true,
		RequiresAnnualRenewal: true,
		Status:                VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		Version:               "1.0.0",
	}

	require.Equal(t, "VerifiedHuman", policy.VcTypeName)
	require.Equal(t, VCType_VC_TYPE_VERIFIED_HUMAN, policy.VcTypeEnum)
	require.Equal(t, uint64(50), policy.CsThreshold)
	require.Equal(t, uint64(365), policy.ExpiryDurationDays)
	require.True(t, policy.Singleton)
	require.True(t, policy.RequiresAnnualRenewal)
	require.Equal(t, VCPolicyStatus_VC_POLICY_STATUS_ACTIVE, policy.Status)
	require.Equal(t, "1.0.0", policy.Version)
}

func TestRevocationRecord_Fields(t *testing.T) {
	record := &RevocationRecord{
		VcId:          "vc123",
		RevokedHeight: 1000,
		Reason:        RevocationReason_REVOCATION_REASON_USER_REQUEST,
		Revoker:       "aura1revoker",
	}

	require.Equal(t, "vc123", record.VcId)
	require.Equal(t, uint64(1000), record.RevokedHeight)
	require.Equal(t, RevocationReason_REVOCATION_REASON_USER_REQUEST, record.Reason)
	require.Equal(t, "aura1revoker", record.Revoker)
}

func TestDIDDocument_Fields(t *testing.T) {
	doc := &DIDDocument{
		Did:        "did:aura:test",
		Controller: "aura1controller",
		VerificationMethods: []*VerificationMethod{
			{
				Id:         "did:aura:test#key1",
				Type:       "Ed25519VerificationKey2020",
				Controller: "did:aura:test",
				PublicKey:  []byte("publickey"),
			},
		},
	}

	require.Equal(t, "did:aura:test", doc.Did)
	require.Equal(t, "aura1controller", doc.Controller)
	require.Len(t, doc.VerificationMethods, 1)
	require.Equal(t, "did:aura:test#key1", doc.VerificationMethods[0].Id)
}

func TestVerificationMethod_Fields(t *testing.T) {
	vm := &VerificationMethod{
		Id:         "did:aura:test#key1",
		Type:       "Ed25519VerificationKey2020",
		Controller: "did:aura:test",
		PublicKey:  []byte("publickey"),
	}

	require.Equal(t, "did:aura:test#key1", vm.Id)
	require.Equal(t, "Ed25519VerificationKey2020", vm.Type)
	require.Equal(t, "did:aura:test", vm.Controller)
	require.Equal(t, []byte("publickey"), vm.PublicKey)
}

func TestRevocationList_Fields(t *testing.T) {
	list := &RevocationList{
		MerkleRoot:        []byte("root"),
		TotalRevocations:  100,
		LastUpdatedHeight: 5000,
	}

	require.Equal(t, []byte("root"), list.MerkleRoot)
	require.Equal(t, uint64(100), list.TotalRevocations)
	require.Equal(t, uint64(5000), list.LastUpdatedHeight)
}

func TestAttributeVC_Fields(t *testing.T) {
	attrVC := &AttributeVC{}

	require.NotNil(t, attrVC)
}

func TestDisclosurePolicy_Fields(t *testing.T) {
	policy := &DisclosurePolicy{
		DefaultMode: DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ASK,
	}

	require.Equal(t, DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ASK, policy.DefaultMode)
}
