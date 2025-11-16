package types

import (
	"fmt"
	"time"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// GenesisState defines the initial state of the vcregistry module
type GenesisState struct {
	Params            Params
	VCRecords         []VCRecord
	RevocationRecords []RevocationRecord
	RevocationList    *RevocationList
	DIDDocuments      []DIDDocument
	VCPolicies        []VCPolicy
	UserMintCounts    map[string]map[int64]uint64
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:            DefaultParams(),
		VCRecords:         []VCRecord{},
		RevocationRecords: []RevocationRecord{},
		RevocationList: &RevocationList{
			MerkleRoot:        []byte{},
			TotalRevocations:  0,
			LastUpdatedHeight: 0,
			LastUpdated:       nil,
		},
		DIDDocuments:   []DIDDocument{},
		VCPolicies:     DefaultVCPolicies(),
		UserMintCounts: make(map[string]map[int64]uint64),
	}
}

// DefaultVCPolicies returns the 16 default VC policies for genesis
func DefaultVCPolicies() []VCPolicy {
	genesisTime := timestamppb.Now()

	return []VCPolicy{
		// Core credentials (1-8)
		{
			VcTypeName:            "VerifiedHuman",
			VcTypeEnum:            VCTypeVerifiedHuman,
			CsThreshold:           50,
			RequiredIrIds:         []string{},
			RequiredArena:         "",
			RequiredArenaScore:    0,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmVerifiedHumanPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "AgeOver18",
			VcTypeEnum:            VCTypeAgeOver18,
			CsThreshold:           60,
			RequiredIrIds:         []string{},
			RequiredArena:         "",
			RequiredArenaScore:    0,
			ExpiryDurationDays:    1825, // 5 years
			Singleton:             true,
			RequiresAnnualRenewal: false,
			MetadataUri:           "ipfs://QmAgeOver18Policy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "AgeOver21",
			VcTypeEnum:            VCTypeAgeOver21,
			CsThreshold:           60,
			RequiredIrIds:         []string{},
			RequiredArena:         "",
			RequiredArenaScore:    0,
			ExpiryDurationDays:    1825, // 5 years
			Singleton:             true,
			RequiresAnnualRenewal: false,
			MetadataUri:           "ipfs://QmAgeOver21Policy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "ResidentOf",
			VcTypeEnum:            VCTypeResidentOf,
			CsThreshold:           70,
			RequiredIrIds:         []string{},
			RequiredArena:         "geolocation",
			RequiredArenaScore:    100,
			ExpiryDurationDays:    365,
			Singleton:             false, // Can be resident of multiple places
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmResidentOfPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "BiometricAuth",
			VcTypeEnum:            VCTypeBiometricAuth,
			CsThreshold:           80,
			RequiredIrIds:         []string{},
			RequiredArena:         "biometric",
			RequiredArenaScore:    150,
			ExpiryDurationDays:    730, // 2 years
			Singleton:             true,
			RequiresAnnualRenewal: false,
			MetadataUri:           "ipfs://QmBiometricAuthPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "KYCVerification",
			VcTypeEnum:            VCTypeKYCVerification,
			CsThreshold:           90,
			RequiredIrIds:         []string{},
			RequiredArena:         "high_assurance",
			RequiredArenaScore:    200,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmKYCVerificationPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "NotaryPublic",
			VcTypeEnum:            VCTypeNotaryPublic,
			CsThreshold:           95,
			RequiredIrIds:         []string{},
			RequiredArena:         "specialized",
			RequiredArenaScore:    300,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmNotaryPublicPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "ProfessionalLicense",
			VcTypeEnum:            VCTypeProfessionalLicense,
			CsThreshold:           90,
			RequiredIrIds:         []string{},
			RequiredArena:         "specialized",
			RequiredArenaScore:    250,
			ExpiryDurationDays:    365,
			Singleton:             false, // Can have multiple licenses
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmProfessionalLicensePolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},

		// Arena focus credentials (20-27)
		{
			VcTypeName:            "BiometricFocus",
			VcTypeEnum:            VCTypeBiometricFocus,
			CsThreshold:           75,
			RequiredIrIds:         []string{},
			RequiredArena:         "biometric",
			RequiredArenaScore:    200,
			ExpiryDurationDays:    730, // 2 years
			Singleton:             true,
			RequiresAnnualRenewal: false,
			MetadataUri:           "ipfs://QmBiometricFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "SocialFocus",
			VcTypeEnum:            VCTypeSocialFocus,
			CsThreshold:           70,
			RequiredIrIds:         []string{},
			RequiredArena:         "social",
			RequiredArenaScore:    150,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmSocialFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "GeolocationFocus",
			VcTypeEnum:            VCTypeGeolocationFocus,
			CsThreshold:           70,
			RequiredIrIds:         []string{},
			RequiredArena:         "geolocation",
			RequiredArenaScore:    150,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmGeolocationFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "HighAssuranceFocus",
			VcTypeEnum:            VCTypeHighAssuranceFocus,
			CsThreshold:           95,
			RequiredIrIds:         []string{},
			RequiredArena:         "high_assurance",
			RequiredArenaScore:    300,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmHighAssuranceFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "PossessionFocus",
			VcTypeEnum:            VCTypePossessionFocus,
			CsThreshold:           65,
			RequiredIrIds:         []string{},
			RequiredArena:         "possession",
			RequiredArenaScore:    100,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: false,
			MetadataUri:           "ipfs://QmPossessionFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "KnowledgeFocus",
			VcTypeEnum:            VCTypeKnowledgeFocus,
			CsThreshold:           70,
			RequiredIrIds:         []string{},
			RequiredArena:         "knowledge",
			RequiredArenaScore:    150,
			ExpiryDurationDays:    730, // 2 years
			Singleton:             true,
			RequiresAnnualRenewal: false,
			MetadataUri:           "ipfs://QmKnowledgeFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "PersistenceFocus",
			VcTypeEnum:            VCTypePersistenceFocus,
			CsThreshold:           60,
			RequiredIrIds:         []string{},
			RequiredArena:         "persistence",
			RequiredArenaScore:    100,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: false,
			MetadataUri:           "ipfs://QmPersistenceFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
		{
			VcTypeName:            "SpecializedFocus",
			VcTypeEnum:            VCTypeSpecializedFocus,
			CsThreshold:           85,
			RequiredIrIds:         []string{},
			RequiredArena:         "specialized",
			RequiredArenaScore:    250,
			ExpiryDurationDays:    365,
			Singleton:             true,
			RequiresAnnualRenewal: true,
			MetadataUri:           "ipfs://QmSpecializedFocusPolicy",
			Status:                VCPolicyStatusActive,
			Version:               "1.0.0",
			CreatedAt:             genesisTime,
			CreatedHeight:         0,
			Creator:               "genesis",
		},
	}
}

// GenesisStateFromProto converts proto genesis state to internal type
func GenesisStateFromProto(pb *vcregistrypb.GenesisState) GenesisState {
	if pb == nil {
		return DefaultGenesisState()
	}

	vcRecords := make([]VCRecord, len(pb.VcRecords))
	for i, record := range pb.VcRecords {
		vcRecords[i] = *record
	}

	revocationRecords := make([]RevocationRecord, len(pb.RevocationRecords))
	for i, record := range pb.RevocationRecords {
		revocationRecords[i] = *record
	}

	didDocuments := make([]DIDDocument, len(pb.DidDocuments))
	for i, doc := range pb.DidDocuments {
		didDocuments[i] = *doc
	}

	vcPolicies := make([]VCPolicy, len(pb.VcPolicies))
	for i, policy := range pb.VcPolicies {
		vcPolicies[i] = *policy
	}

	return GenesisState{
		Params:            ParamsFromProto(pb.Params),
		VCRecords:         vcRecords,
		RevocationRecords: revocationRecords,
		RevocationList:    RevocationListFromProto(pb.RevocationList),
		DIDDocuments:      didDocuments,
		VCPolicies:        vcPolicies,
		UserMintCounts:    convertUserMintCountsFromProto(pb.UserMintCounts),
	}
}

// GenesisStateToProto converts internal genesis state to proto type
func GenesisStateToProto(gs GenesisState) *vcregistrypb.GenesisState {
	vcRecords := make([]*vcregistrypb.VCRecord, len(gs.VCRecords))
	for i, record := range gs.VCRecords {
		vcRecords[i] = VCRecordToProto(record)
	}

	revocationRecords := make([]*vcregistrypb.RevocationRecord, len(gs.RevocationRecords))
	for i, record := range gs.RevocationRecords {
		revocationRecords[i] = RevocationRecordToProto(record)
	}

	didDocuments := make([]*vcregistrypb.DIDDocument, len(gs.DIDDocuments))
	for i, doc := range gs.DIDDocuments {
		didDocuments[i] = DIDDocumentToProto(doc)
	}

	vcPolicies := make([]*vcregistrypb.VCPolicy, len(gs.VCPolicies))
	for i, policy := range gs.VCPolicies {
		vcPolicies[i] = VCPolicyToProto(policy)
	}

	return &vcregistrypb.GenesisState{
		Params:            ParamsToProto(gs.Params),
		VcRecords:         vcRecords,
		RevocationRecords: revocationRecords,
		RevocationList:    RevocationListToProto(gs.RevocationList),
		DidDocuments:      didDocuments,
		VcPolicies:        vcPolicies,
		UserMintCounts:    convertUserMintCountsToProto(gs.UserMintCounts),
	}
}

// convertUserMintCountsFromProto converts proto user mint counts to internal format
func convertUserMintCountsFromProto(protoMintCounts map[string]uint64) map[string]map[int64]uint64 {
	// Proto format: map[string]uint64 (address -> today's count)
	// Internal format: map[string]map[int64]uint64 (address -> day_timestamp -> count)
	// For genesis import, we'll create entries for today
	result := make(map[string]map[int64]uint64)
	today := time.Now().Unix() / 86400 * 86400 // Normalize to start of day

	for addr, count := range protoMintCounts {
		if result[addr] == nil {
			result[addr] = make(map[int64]uint64)
		}
		result[addr][today] = count
	}
	return result
}

// convertUserMintCountsToProto converts internal user mint counts to proto format
func convertUserMintCountsToProto(internalMintCounts map[string]map[int64]uint64) map[string]uint64 {
	// Internal format: map[string]map[int64]uint64 (address -> day_timestamp -> count)
	// Proto format: map[string]uint64 (address -> today's count)
	// For genesis export, we'll export only today's counts
	result := make(map[string]uint64)
	today := time.Now().Unix() / 86400 * 86400 // Normalize to start of day

	for addr, dayCounts := range internalMintCounts {
		if count, ok := dayCounts[today]; ok {
			result[addr] = count
		}
	}
	return result
}

// Validate performs validation on the genesis state
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Track seen VC IDs to prevent duplicates
	vcIDsSeen := make(map[string]bool)
	for i, record := range gs.VCRecords {
		if record.VcId == "" {
			return fmt.Errorf("vc_record[%d]: vc_id cannot be empty", i)
		}

		if vcIDsSeen[record.VcId] {
			return fmt.Errorf("duplicate vc_id: %s", record.VcId)
		}
		vcIDsSeen[record.VcId] = true

		if record.HolderAddress == "" {
			return fmt.Errorf("vc_record[%d] (%s): holder_address cannot be empty", i, record.VcId)
		}

		if record.HolderDid == "" {
			return fmt.Errorf("vc_record[%d] (%s): holder_did cannot be empty", i, record.VcId)
		}

		// Validate timestamps
		if record.IssuedAt == nil {
			return fmt.Errorf("vc_record[%d] (%s): issued_at cannot be nil", i, record.VcId)
		}

		if record.ExpiresAt != nil && record.ExpiresAt.Nanos < 0 || (record.ExpiresAt != nil && record.IssuedAt != nil && record.ExpiresAt.Seconds <= record.IssuedAt.Seconds && record.ExpiresAt.Seconds > 0) {
			return fmt.Errorf("vc_record[%d] (%s): expires_at must be after issued_at", i, record.VcId)
		}
	}

	// Track seen revocation record IDs
	revocationIDsSeen := make(map[string]bool)
	for i, record := range gs.RevocationRecords {
		if record.VcId == "" {
			return fmt.Errorf("revocation_record[%d]: vc_id cannot be empty", i)
		}

		if revocationIDsSeen[record.VcId] {
			return fmt.Errorf("duplicate revocation record for vc_id: %s", record.VcId)
		}
		revocationIDsSeen[record.VcId] = true

		// Verify that the revoked VC exists if there are VCRecords
		if len(gs.VCRecords) > 0 && !vcIDsSeen[record.VcId] {
			return fmt.Errorf("revocation_record[%d]: revoked vc_id not found in vc_records: %s", i, record.VcId)
		}

		if record.RevokedAt == nil {
			return fmt.Errorf("revocation_record[%d]: revoked_at cannot be nil", i)
		}
	}

	// Validate revocation list
	if gs.RevocationList != nil {
		if gs.RevocationList.TotalRevocations != uint64(len(gs.RevocationRecords)) {
			return fmt.Errorf("revocation_list: total_revocations mismatch (expected %d, got %d)",
				len(gs.RevocationRecords), gs.RevocationList.TotalRevocations)
		}
	}

	// Track seen DIDs and addresses
	didsSeen := make(map[string]bool)
	addressToDIDs := make(map[string]int)
	for i, doc := range gs.DIDDocuments {
		if doc.Did == "" {
			return fmt.Errorf("did_document[%d]: did cannot be empty", i)
		}

		if didsSeen[doc.Did] {
			return fmt.Errorf("duplicate did: %s", doc.Did)
		}
		didsSeen[doc.Did] = true

		if doc.Controller == "" {
			return fmt.Errorf("did_document[%d]: controller cannot be empty", i)
		}

		addressToDIDs[doc.Controller]++

		if doc.Created == nil {
			return fmt.Errorf("did_document[%d]: created timestamp cannot be nil", i)
		}

		if doc.Updated != nil && doc.Created != nil && doc.Updated.Seconds < doc.Created.Seconds {
			return fmt.Errorf("did_document[%d]: updated cannot be before created", i)
		}
	}

	// Validate VC Policies
	policyTypesSeen := make(map[string]bool)
	for i, policy := range gs.VCPolicies {
		if policy.VcTypeName == "" {
			return fmt.Errorf("vc_policy[%d]: vc_type_name cannot be empty", i)
		}

		if policyTypesSeen[policy.VcTypeName] {
			return fmt.Errorf("duplicate vc policy for type: %s", policy.VcTypeName)
		}
		policyTypesSeen[policy.VcTypeName] = true

		if policy.CreatedAt == nil {
			return fmt.Errorf("vc_policy[%d]: created_at cannot be nil", i)
		}
	}

	// Validate user mint counts
	for address, dayCounts := range gs.UserMintCounts {
		if address == "" {
			return fmt.Errorf("user_mint_counts: address cannot be empty")
		}

		// Validate each day's counts
		for _, count := range dayCounts {
			if count > gs.Params.MaxMintPerDay {
				return fmt.Errorf("user_mint_counts: address %s count (%d) exceeds max_mint_per_day (%d)",
					address, count, gs.Params.MaxMintPerDay)
			}
		}
	}

	return nil
}

// TestGenesisState returns a genesis state with test data for testing
func TestGenesisState() GenesisState {
	return GenesisState{
		Params:            DefaultParams(),
		VCRecords:         []VCRecord{},
		RevocationRecords: []RevocationRecord{},
		RevocationList: &RevocationList{
			MerkleRoot:        []byte{},
			TotalRevocations:  0,
			LastUpdatedHeight: 0,
			LastUpdated:       nil,
		},
		DIDDocuments:   []DIDDocument{},
		VCPolicies:     []VCPolicy{},
		UserMintCounts: make(map[string]map[int64]uint64),
	}
}
