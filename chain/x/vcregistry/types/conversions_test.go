package types

import (
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
)

func TestVCPolicyFromProto_Nil(t *testing.T) {
	result := VCPolicyFromProto(nil)
	require.Nil(t, result)
}

func TestVCPolicyFromProto_Valid(t *testing.T) {
	now := &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: 0}
	pb := &VCPolicy{
		VcTypeName:            "VerifiedHuman",
		VcTypeEnum:            VCType_VC_TYPE_VERIFIED_HUMAN,
		CsThreshold:           50,
		RequiredIrIds:         []string{"ir1", "ir2"},
		RequiredArena:         "biometric",
		RequiredArenaScore:    200,
		ExpiryDurationDays:    365,
		Singleton:             true,
		RequiresAnnualRenewal: true,
		MetadataUri:           "https://metadata.example.com",
		Status:                VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		Version:               "1.0.0",
		CreatedAt:             now,
		CreatedHeight:         1000,
		Creator:               "aura1creator",
	}

	result := VCPolicyFromProto(pb)
	require.NotNil(t, result)
	require.Equal(t, pb.VcTypeName, result.VcTypeName)
	require.Equal(t, pb.VcTypeEnum, result.VcTypeEnum)
	require.Equal(t, pb.CsThreshold, result.CsThreshold)
	require.Equal(t, pb.RequiredIrIds, result.RequiredIrIds)
	require.Equal(t, pb.RequiredArena, result.RequiredArena)
	require.Equal(t, pb.RequiredArenaScore, result.RequiredArenaScore)
	require.Equal(t, pb.ExpiryDurationDays, result.ExpiryDurationDays)
	require.Equal(t, pb.Singleton, result.Singleton)
	require.Equal(t, pb.RequiresAnnualRenewal, result.RequiresAnnualRenewal)
	require.Equal(t, pb.MetadataUri, result.MetadataUri)
	require.Equal(t, pb.Status, result.Status)
	require.Equal(t, pb.Version, result.Version)
	require.Equal(t, pb.CreatedAt, result.CreatedAt)
	require.Equal(t, pb.CreatedHeight, result.CreatedHeight)
	require.Equal(t, pb.Creator, result.Creator)
}

func TestVerificationMethodFromProto_Nil(t *testing.T) {
	result := VerificationMethodFromProto(nil)
	require.Nil(t, result)
}

func TestVerificationMethodFromProto_Valid(t *testing.T) {
	pb := &VerificationMethod{
		Id:         "did:aura:test#key1",
		Type:       "Ed25519VerificationKey2020",
		Controller: "did:aura:test",
		PublicKey:  []byte("publickey"),
	}

	result := VerificationMethodFromProto(pb)
	require.NotNil(t, result)
	require.Equal(t, pb.Id, result.Id)
	require.Equal(t, pb.Type, result.Type)
	require.Equal(t, pb.Controller, result.Controller)
	require.Equal(t, pb.PublicKey, result.PublicKey)
}

func TestVerificationMethodsFromProto_Nil(t *testing.T) {
	result := VerificationMethodsFromProto(nil)
	require.Nil(t, result)
}

func TestVerificationMethodsFromProto_Empty(t *testing.T) {
	result := VerificationMethodsFromProto([]*VerificationMethod{})
	require.NotNil(t, result)
	require.Empty(t, result)
}

func TestVerificationMethodsFromProto_Multiple(t *testing.T) {
	pbs := []*VerificationMethod{
		{
			Id:         "did:aura:test#key1",
			Type:       "Ed25519VerificationKey2020",
			Controller: "did:aura:test",
			PublicKey:  []byte("publickey1"),
		},
		{
			Id:         "did:aura:test#key2",
			Type:       "EcdsaSecp256k1VerificationKey2019",
			Controller: "did:aura:test",
			PublicKey:  []byte("publickey2"),
		},
	}

	result := VerificationMethodsFromProto(pbs)
	require.NotNil(t, result)
	require.Len(t, result, 2)
	require.Equal(t, pbs[0].Id, result[0].Id)
	require.Equal(t, pbs[1].Id, result[1].Id)
}

func TestVCRecordToProto_Nil(t *testing.T) {
	result := VCRecordToProto(nil)
	require.Nil(t, result)
}

func TestVCRecordToProto_Valid(t *testing.T) {
	now := &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: 0}
	record := &VCRecord{
		VcId:               "vc123",
		VcType:             VCType_VC_TYPE_VERIFIED_HUMAN,
		VcTypeCustom:       "",
		HolderDid:          "did:aura:holder",
		HolderAddress:      "aura1holder",
		Status:             VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:           now,
		ExpiresAt:          now,
		IssuedHeight:       1000,
		CredentialHash:     []byte("hash"),
		VerifierPluginHash: []byte("plugin"),
		IssuerAssistant:    "assistant",
		PrerequisiteIrIds:  []string{"ir1"},
		Metadata:           map[string]string{"key": "value"},
		CsAtMint:           75,
		PolicyVersion:      "1.0.0",
	}

	result := VCRecordToProto(record)
	require.NotNil(t, result)
	require.Equal(t, record.VcId, result.VcId)
	require.Equal(t, record.VcType, result.VcType)
	require.Equal(t, record.HolderDid, result.HolderDid)
	require.Equal(t, record.HolderAddress, result.HolderAddress)
	require.Equal(t, record.Status, result.Status)
	require.Equal(t, record.IssuedAt, result.IssuedAt)
	require.Equal(t, record.ExpiresAt, result.ExpiresAt)
}

func TestVCPolicyToProto_Nil(t *testing.T) {
	result := VCPolicyToProto(nil)
	require.Nil(t, result)
}

func TestVCPolicyToProto_Valid(t *testing.T) {
	now := &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: 0}
	policy := &VCPolicy{
		VcTypeName:            "VerifiedHuman",
		VcTypeEnum:            VCType_VC_TYPE_VERIFIED_HUMAN,
		CsThreshold:           50,
		RequiredIrIds:         []string{"ir1", "ir2"},
		RequiredArena:         "biometric",
		RequiredArenaScore:    200,
		ExpiryDurationDays:    365,
		Singleton:             true,
		RequiresAnnualRenewal: true,
		MetadataUri:           "https://metadata.example.com",
		Status:                VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		Version:               "1.0.0",
		CreatedAt:             now,
		CreatedHeight:         1000,
		Creator:               "aura1creator",
	}

	result := VCPolicyToProto(policy)
	require.NotNil(t, result)
	require.Equal(t, policy.VcTypeName, result.VcTypeName)
	require.Equal(t, policy.VcTypeEnum, result.VcTypeEnum)
	require.Equal(t, policy.CsThreshold, result.CsThreshold)
}

func TestDIDDocumentToProto_Nil(t *testing.T) {
	result := DIDDocumentToProto(nil)
	require.Nil(t, result)
}

func TestDIDDocumentToProto_Valid(t *testing.T) {
	now := &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: 0}
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
		CredentialIds:    []string{"vc1", "vc2"},
		Created:          now,
		Updated:          now,
		MetadataUri:      "https://metadata.example.com",
		ServiceEndpoints: map[string]string{"service1": "https://service.example.com"},
	}

	result := DIDDocumentToProto(doc)
	require.NotNil(t, result)
	require.Equal(t, doc.Did, result.Did)
	require.Equal(t, doc.Controller, result.Controller)
	require.Len(t, result.VerificationMethods, 1)
	require.Equal(t, doc.VerificationMethods[0].Id, result.VerificationMethods[0].Id)
}

func TestParamsToProto_Nil(t *testing.T) {
	result := ParamsToProto(nil)
	require.Nil(t, result)
}

func TestParamsToProto_Valid(t *testing.T) {
	params := &Params{
		MaxVcsPerUser:                   50,
		MaxMintPerDay:                   5,
		MaxMintPerHour:                  2,
		DefaultVcExpiryDays:             365,
		RevocationMerkleUpdateFrequency: 100,
		DidPrefix:                       "did:aura",
		DidNetwork:                      "mainnet",
		MintFee:                         "1000000uaura",
		RevokeFee:                       "0uaura",
		PolicyCreationDeposit:           "10000000uaura",
		RateLimitingEnabled:             true,
	}

	result := ParamsToProto(params)
	require.NotNil(t, result)
	require.Equal(t, params.MaxVcsPerUser, result.MaxVcsPerUser)
	require.Equal(t, params.MaxMintPerDay, result.MaxMintPerDay)
	require.Equal(t, params.MaxMintPerHour, result.MaxMintPerHour)
}

func TestRevocationRecordToProto_Nil(t *testing.T) {
	result := RevocationRecordToProto(nil)
	require.Nil(t, result)
}

func TestRevocationRecordToProto_Valid(t *testing.T) {
	now := &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: 0}
	record := &RevocationRecord{
		VcId:          "vc123",
		RevokedAt:     now,
		RevokedHeight: 1000,
		Reason:        RevocationReason_REVOCATION_REASON_USER_REQUEST,
		Revoker:       "aura1revoker",
		Evidence:      "user requested",
		MerkleProof:   []byte("proof"),
	}

	result := RevocationRecordToProto(record)
	require.NotNil(t, result)
	require.Equal(t, record.VcId, result.VcId)
	require.Equal(t, record.RevokedAt, result.RevokedAt)
	require.Equal(t, record.RevokedHeight, result.RevokedHeight)
	require.Equal(t, record.Reason, result.Reason)
}

func TestRevocationListToProto_Nil(t *testing.T) {
	result := RevocationListToProto(nil)
	require.Nil(t, result)
}

func TestRevocationListToProto_Valid(t *testing.T) {
	now := &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: 0}
	list := &RevocationList{
		MerkleRoot:        []byte("root"),
		TotalRevocations:  100,
		LastUpdatedHeight: 5000,
		LastUpdated:       now,
	}

	result := RevocationListToProto(list)
	require.NotNil(t, result)
	require.Equal(t, list.MerkleRoot, result.MerkleRoot)
	require.Equal(t, list.TotalRevocations, result.TotalRevocations)
	require.Equal(t, list.LastUpdatedHeight, result.LastUpdatedHeight)
	require.Equal(t, list.LastUpdated, result.LastUpdated)
}

func TestEnumConversions(t *testing.T) {
	// Test VCType conversions
	vcType := VCType_VC_TYPE_VERIFIED_HUMAN
	pbType := VCTypeToProto(vcType)
	backType := VCTypeFromProto(pbType)
	require.Equal(t, vcType, backType)

	// Test VCStatus conversions
	status := VCStatus_VC_STATUS_ACTIVE
	pbStatus := VCStatusToProto(status)
	backStatus := VCStatusFromProto(pbStatus)
	require.Equal(t, status, backStatus)

	// Test VCPolicyStatus conversions
	policyStatus := VCPolicyStatus_VC_POLICY_STATUS_ACTIVE
	pbPolicyStatus := VCPolicyStatusToProto(policyStatus)
	backPolicyStatus := VCPolicyStatusFromProto(pbPolicyStatus)
	require.Equal(t, policyStatus, backPolicyStatus)

	// Test RevocationReason conversions
	reason := RevocationReason_REVOCATION_REASON_USER_REQUEST
	pbReason := RevocationReasonToProto(reason)
	backReason := RevocationReasonFromProto(pbReason)
	require.Equal(t, reason, backReason)
}
