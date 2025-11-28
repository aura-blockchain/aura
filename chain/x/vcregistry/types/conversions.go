package types

import (
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// VCPolicyFromProto converts a protobuf VCPolicy to types.VCPolicy
func VCPolicyFromProto(pb *vcregistrypb.VCPolicy) *VCPolicy {
	if pb == nil {
		return nil
	}
	return &VCPolicy{
		VcTypeName:            pb.VcTypeName,
		VcTypeEnum:            VCType(pb.VcTypeEnum),
		CsThreshold:           pb.CsThreshold,
		RequiredIrIds:         pb.RequiredIrIds,
		RequiredArena:         pb.RequiredArena,
		RequiredArenaScore:    pb.RequiredArenaScore,
		ExpiryDurationDays:    pb.ExpiryDurationDays,
		Singleton:             pb.Singleton,
		RequiresAnnualRenewal: pb.RequiresAnnualRenewal,
		MetadataUri:           pb.MetadataUri,
		Status:                VCPolicyStatus(pb.Status),
		Version:               pb.Version,
		CreatedAt:             pb.CreatedAt,
		CreatedHeight:         pb.CreatedHeight,
		Creator:               pb.Creator,
	}
}

// VerificationMethodFromProto converts a protobuf VerificationMethod to types.VerificationMethod
func VerificationMethodFromProto(pb *vcregistrypb.VerificationMethod) *VerificationMethod {
	if pb == nil {
		return nil
	}
	return &VerificationMethod{
		Id:         pb.Id,
		Type:       pb.Type,
		Controller: pb.Controller,
		PublicKey:  pb.PublicKey,
	}
}

// VerificationMethodsFromProto converts a slice of protobuf VerificationMethods
func VerificationMethodsFromProto(pbs []*vcregistrypb.VerificationMethod) []*VerificationMethod {
	if pbs == nil {
		return nil
	}
	result := make([]*VerificationMethod, len(pbs))
	for i, pb := range pbs {
		result[i] = VerificationMethodFromProto(pb)
	}
	return result
}

// VCTypeFromProto converts protobuf VCType to types.VCType
func VCTypeFromProto(pb vcregistrypb.VCType) VCType {
	return VCType(pb)
}

// VCStatusFromProto converts protobuf VCStatus to types.VCStatus
func VCStatusFromProto(pb vcregistrypb.VCStatus) VCStatus {
	return VCStatus(pb)
}

// VCPolicyStatusFromProto converts protobuf VCPolicyStatus to types.VCPolicyStatus
func VCPolicyStatusFromProto(pb vcregistrypb.VCPolicyStatus) VCPolicyStatus {
	return VCPolicyStatus(pb)
}

// RevocationReasonFromProto converts protobuf RevocationReason to types.RevocationReason
func RevocationReasonFromProto(pb vcregistrypb.RevocationReason) RevocationReason {
	return RevocationReason(pb)
}

// VCTypeToProto converts types.VCType to protobuf VCType
func VCTypeToProto(t VCType) vcregistrypb.VCType {
	return vcregistrypb.VCType(t)
}

// VCStatusToProto converts types.VCStatus to protobuf VCStatus
func VCStatusToProto(s VCStatus) vcregistrypb.VCStatus {
	return vcregistrypb.VCStatus(s)
}

// VCPolicyStatusToProto converts types.VCPolicyStatus to protobuf VCPolicyStatus
func VCPolicyStatusToProto(s VCPolicyStatus) vcregistrypb.VCPolicyStatus {
	return vcregistrypb.VCPolicyStatus(s)
}

// RevocationReasonToProto converts types.RevocationReason to protobuf RevocationReason
func RevocationReasonToProto(r RevocationReason) vcregistrypb.RevocationReason {
	return vcregistrypb.RevocationReason(r)
}

// VCRecordToProto converts types.VCRecord to protobuf VCRecord
func VCRecordToProto(r *VCRecord) *vcregistrypb.VCRecord {
	if r == nil {
		return nil
	}
	return &vcregistrypb.VCRecord{
		VcId:               r.VcId,
		VcType:             vcregistrypb.VCType(r.VcType),
		VcTypeCustom:       r.VcTypeCustom,
		HolderDid:          r.HolderDid,
		HolderAddress:      r.HolderAddress,
		Status:             vcregistrypb.VCStatus(r.Status),
		IssuedAt:           r.IssuedAt,
		ExpiresAt:          r.ExpiresAt,
		IssuedHeight:       r.IssuedHeight,
		CredentialHash:     r.CredentialHash,
		VerifierPluginHash: r.VerifierPluginHash,
		IssuerAssistant:    r.IssuerAssistant,
		PrerequisiteIrIds:  r.PrerequisiteIrIds,
		Metadata:           r.Metadata,
		CsAtMint:           r.CsAtMint,
		PolicyVersion:      r.PolicyVersion,
	}
}

// VCPolicyToProto converts types.VCPolicy to protobuf VCPolicy
func VCPolicyToProto(p *VCPolicy) *vcregistrypb.VCPolicy {
	if p == nil {
		return nil
	}
	return &vcregistrypb.VCPolicy{
		VcTypeName:            p.VcTypeName,
		VcTypeEnum:            vcregistrypb.VCType(p.VcTypeEnum),
		CsThreshold:           p.CsThreshold,
		RequiredIrIds:         p.RequiredIrIds,
		RequiredArena:         p.RequiredArena,
		RequiredArenaScore:    p.RequiredArenaScore,
		ExpiryDurationDays:    p.ExpiryDurationDays,
		Singleton:             p.Singleton,
		RequiresAnnualRenewal: p.RequiresAnnualRenewal,
		MetadataUri:           p.MetadataUri,
		Status:                vcregistrypb.VCPolicyStatus(p.Status),
		Version:               p.Version,
		CreatedAt:             p.CreatedAt,
		CreatedHeight:         p.CreatedHeight,
		Creator:               p.Creator,
	}
}

// DIDDocumentToProto converts types.DIDDocument to protobuf DIDDocument
func DIDDocumentToProto(d *DIDDocument) *vcregistrypb.DIDDocument {
	if d == nil {
		return nil
	}

	// Convert verification methods
	pbMethods := make([]*vcregistrypb.VerificationMethod, len(d.VerificationMethods))
	for i, m := range d.VerificationMethods {
		if m != nil {
			pbMethods[i] = &vcregistrypb.VerificationMethod{
				Id:         m.Id,
				Type:       m.Type,
				Controller: m.Controller,
				PublicKey:  m.PublicKey,
			}
		}
	}

	return &vcregistrypb.DIDDocument{
		Did:                 d.Did,
		Controller:          d.Controller,
		VerificationMethods: pbMethods,
		CredentialIds:       d.CredentialIds,
		Created:             d.Created,
		Updated:             d.Updated,
		MetadataUri:         d.MetadataUri,
		ServiceEndpoints:    d.ServiceEndpoints,
	}
}

// ParamsToProto converts types.Params to protobuf Params
// Note: types.Params and vcregistrypb.Params are identical, so we just cast the pointer
func ParamsToProto(p *Params) *vcregistrypb.Params {
	if p == nil {
		return nil
	}
	return &vcregistrypb.Params{
		MaxVcsPerUser:                   p.MaxVcsPerUser,
		MaxMintPerDay:                   p.MaxMintPerDay,
		MaxMintPerHour:                  p.MaxMintPerHour,
		DefaultVcExpiryDays:             p.DefaultVcExpiryDays,
		RevocationMerkleUpdateFrequency: p.RevocationMerkleUpdateFrequency,
		DidPrefix:                       p.DidPrefix,
		DidNetwork:                      p.DidNetwork,
		MintFee:                         p.MintFee,
		RevokeFee:                       p.RevokeFee,
		PolicyCreationDeposit:           p.PolicyCreationDeposit,
		RateLimitingEnabled:             p.RateLimitingEnabled,
	}
}

// RevocationRecordToProto converts types.RevocationRecord to protobuf RevocationRecord
func RevocationRecordToProto(r *RevocationRecord) *vcregistrypb.RevocationRecord {
	if r == nil {
		return nil
	}
	return &vcregistrypb.RevocationRecord{
		VcId:          r.VcId,
		RevokedAt:     r.RevokedAt,
		RevokedHeight: r.RevokedHeight,
		Reason:        vcregistrypb.RevocationReason(r.Reason),
		Revoker:       r.Revoker,
		Evidence:      r.Evidence,
		MerkleProof:   r.MerkleProof,
	}
}

// RevocationListToProto converts types.RevocationList to protobuf RevocationList
func RevocationListToProto(l *RevocationList) *vcregistrypb.RevocationList {
	if l == nil {
		return nil
	}
	return &vcregistrypb.RevocationList{
		MerkleRoot:        l.MerkleRoot,
		TotalRevocations:  l.TotalRevocations,
		LastUpdatedHeight: l.LastUpdatedHeight,
		LastUpdated:       l.LastUpdated,
	}
}
