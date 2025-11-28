package types

import (
	"encoding/json"
	"fmt"

	pb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultGenesisState returns the default genesis state for the vcregistry module
func DefaultGenesisState() *pb.GenesisState {
	params := DefaultParams()
	return &pb.GenesisState{
		Params:                 params,
		VcRecords:              []*pb.VCRecord{},
		RevocationRecords:      []*pb.RevocationRecord{},
		RevocationList:         &pb.RevocationList{},
		DidDocuments:           []*pb.DIDDocument{},
		VcPolicies:             []*pb.VCPolicy{},
		UserMintCounts:         make(map[string]uint64),
		Presentations:          []*pb.VCPresentation{},
		UserPresentationIndex:  make(map[string]*pb.PresentationIds),
		AttributeVcs:           []*pb.AttributeVC{},
		UserAttributeIndex:     make(map[string]*pb.AttributeVcIds),
	}
}

// ValidateGenesisState validates the genesis state
func ValidateGenesisState(state *pb.GenesisState) error {
	if state == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	// Validate params - params must not be nil
	if state.Params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	if err := ValidateParams(state.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Track VC IDs to prevent duplicates
	vcIDs := make(map[string]struct{})

	// Validate VC records
	for i, vc := range state.VcRecords {
		if vc == nil {
			return fmt.Errorf("vc record at index %d is nil", i)
		}
		if vc.VcId == "" {
			return fmt.Errorf("vc record at index %d has empty vc_id", i)
		}
		if _, exists := vcIDs[vc.VcId]; exists {
			return fmt.Errorf("duplicate vc_id: %s", vc.VcId)
		}
		vcIDs[vc.VcId] = struct{}{}

		if vc.HolderDid == "" && vc.HolderAddress == "" {
			return fmt.Errorf("vc %s has empty holder_did and holder_address", vc.VcId)
		}
		if vc.IssuerAssistant == "" {
			return fmt.Errorf("vc %s has empty issuer_assistant", vc.VcId)
		}
		if vc.VcType == pb.VCType_VC_TYPE_UNSPECIFIED {
			return fmt.Errorf("vc %s has unspecified type", vc.VcId)
		}
		if vc.IssuedAt == nil {
			return fmt.Errorf("vc %s has nil issued_at", vc.VcId)
		}
	}

	// Validate revocation records
	revocationVCs := make(map[string]struct{})
	for i, revocation := range state.RevocationRecords {
		if revocation == nil {
			return fmt.Errorf("revocation record at index %d is nil", i)
		}
		if revocation.VcId == "" {
			return fmt.Errorf("revocation record at index %d has empty vc_id", i)
		}
		if _, exists := revocationVCs[revocation.VcId]; exists {
			return fmt.Errorf("duplicate revocation for vc_id: %s", revocation.VcId)
		}
		revocationVCs[revocation.VcId] = struct{}{}

		if revocation.Reason == pb.RevocationReason_REVOCATION_REASON_UNSPECIFIED {
			return fmt.Errorf("revocation for vc %s has unspecified reason", revocation.VcId)
		}
		if revocation.RevokedAt == nil {
			return fmt.Errorf("revocation for vc %s has nil revoked_at", revocation.VcId)
		}
	}

	// Validate revocation list exists
	if state.RevocationList == nil {
		return fmt.Errorf("revocation_list cannot be nil")
	}

	// Validate DID documents
	didIDs := make(map[string]struct{})
	for i, did := range state.DidDocuments {
		if did == nil {
			return fmt.Errorf("did document at index %d is nil", i)
		}
		if did.Did == "" {
			return fmt.Errorf("did document at index %d has empty did", i)
		}
		if _, exists := didIDs[did.Did]; exists {
			return fmt.Errorf("duplicate did: %s", did.Did)
		}
		didIDs[did.Did] = struct{}{}

		if did.Controller == "" {
			return fmt.Errorf("did %s has empty controller", did.Did)
		}
	}

	// Validate VC policies
	policyNames := make(map[string]struct{})
	for i, policy := range state.VcPolicies {
		if policy == nil {
			return fmt.Errorf("vc policy at index %d is nil", i)
		}
		if policy.VcTypeName == "" {
			return fmt.Errorf("vc policy at index %d has empty vc_type_name", i)
		}
		if _, exists := policyNames[policy.VcTypeName]; exists {
			return fmt.Errorf("duplicate vc_type_name: %s", policy.VcTypeName)
		}
		policyNames[policy.VcTypeName] = struct{}{}

		if policy.VcTypeEnum == pb.VCType_VC_TYPE_UNSPECIFIED {
			return fmt.Errorf("policy %s has unspecified vc_type_enum", policy.VcTypeName)
		}
	}

	// Validate user mint counts
	for addr, count := range state.UserMintCounts {
		if addr == "" {
			return fmt.Errorf("user mint count has empty address")
		}
		// count >= 0 is always true for uint64
		_ = count
	}

	// Validate presentations
	presentationIDs := make(map[string]struct{})
	for i, presentation := range state.Presentations {
		if presentation == nil {
			return fmt.Errorf("presentation at index %d is nil", i)
		}
		if presentation.PresentationId == "" {
			return fmt.Errorf("presentation at index %d has empty presentation_id", i)
		}
		if _, exists := presentationIDs[presentation.PresentationId]; exists {
			return fmt.Errorf("duplicate presentation_id: %s", presentation.PresentationId)
		}
		presentationIDs[presentation.PresentationId] = struct{}{}

		if presentation.HolderDid == "" && presentation.HolderAddress == "" {
			return fmt.Errorf("presentation %s has empty holder_did and holder_address", presentation.PresentationId)
		}
		if len(presentation.VcIds) == 0 {
			return fmt.Errorf("presentation %s has no vc_ids", presentation.PresentationId)
		}
	}

	// Validate user presentation index
	for addr, presIDs := range state.UserPresentationIndex {
		if addr == "" {
			return fmt.Errorf("user presentation index has empty address")
		}
		if presIDs == nil {
			return fmt.Errorf("user presentation index for %s is nil", addr)
		}
		for _, presID := range presIDs.Ids {
			if _, exists := presentationIDs[presID]; !exists {
				return fmt.Errorf("user presentation index references non-existent presentation %s", presID)
			}
		}
	}

	// Validate attribute VCs
	attrVCIDs := make(map[string]struct{})
	for i, attrVC := range state.AttributeVcs {
		if attrVC == nil {
			return fmt.Errorf("attribute vc at index %d is nil", i)
		}
		if attrVC.AttributeVcId == "" {
			return fmt.Errorf("attribute vc at index %d has empty attribute_vc_id", i)
		}
		if _, exists := attrVCIDs[attrVC.AttributeVcId]; exists {
			return fmt.Errorf("duplicate attribute_vc_id: %s", attrVC.AttributeVcId)
		}
		attrVCIDs[attrVC.AttributeVcId] = struct{}{}

		if attrVC.HolderAddress == "" {
			return fmt.Errorf("attribute vc %s has empty holder_address", attrVC.AttributeVcId)
		}
	}

	// Validate user attribute index
	for addr, attrIDs := range state.UserAttributeIndex {
		if addr == "" {
			return fmt.Errorf("user attribute index has empty address")
		}
		if attrIDs == nil {
			return fmt.Errorf("user attribute index for %s is nil", addr)
		}
		for _, attrID := range attrIDs.Ids {
			if _, exists := attrVCIDs[attrID]; !exists {
				return fmt.Errorf("user attribute index references non-existent attribute vc %s", attrID)
			}
		}
	}

	return nil
}

// DefaultGenesis returns the default genesis as raw JSON
func DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis is an alias for ValidateGenesisState for consistency
func ValidateGenesis(state *pb.GenesisState) error {
	return ValidateGenesisState(state)
}
