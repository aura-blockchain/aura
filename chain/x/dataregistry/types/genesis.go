package types

import (
	"encoding/json"
	"fmt"

	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultGenesisState returns the default genesis state for the dataregistry module
func DefaultGenesisState() *pb.GenesisState {
	params := DefaultParams()
	return &pb.GenesisState{
		Params:     &params,
		DataItems:  []*pb.DataItem{},
		NextDataId: 1,
	}
}

// ValidateGenesisState validates the genesis state
func ValidateGenesisState(state *pb.GenesisState) error {
	if state == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	// Validate params
	if state.Params != nil {
		if err := ValidateParams(state.Params); err != nil {
			return fmt.Errorf("invalid params: %w", err)
		}
	}

	// Track data item IDs to prevent duplicates
	dataItemIDs := make(map[string]struct{})

	// Validate data items
	for i, item := range state.DataItems {
		if item == nil {
			return fmt.Errorf("data item at index %d is nil", i)
		}

		if item.DataId == "" {
			return fmt.Errorf("data item at index %d has empty data_id", i)
		}

		if _, exists := dataItemIDs[item.DataId]; exists {
			return fmt.Errorf("duplicate data item ID: %s", item.DataId)
		}
		dataItemIDs[item.DataId] = struct{}{}

		if item.OwnerAddress == "" {
			return fmt.Errorf("data item %s has empty owner", item.DataId)
		}

		if item.DataType == DataItemType_DATA_ITEM_TYPE_UNSPECIFIED {
			return fmt.Errorf("data item %s has unspecified data type", item.DataId)
		}

		// Validate content hash
		if len(item.ContentHash) == 0 {
			return fmt.Errorf("data item %s has empty content hash", item.DataId)
		}

		// Validate timestamps
		if item.CreatedAt == nil {
			return fmt.Errorf("data item %s has nil created_at timestamp", item.DataId)
		}

		// Validate expiration if set
		if item.ExpiresAt != nil {
			if item.ExpiresAt.Seconds < item.CreatedAt.Seconds {
				return fmt.Errorf("data item %s expires_at is before created_at", item.DataId)
			}
		}

		// Validate access policy
		if item.AccessPolicy == nil {
			return fmt.Errorf("data item %s has nil access policy", item.DataId)
		}

		// Validate verification list
		verifierAddrs := make(map[string]struct{})
		for j, verification := range item.Verifications {
			if verification == nil {
				return fmt.Errorf("verification at index %d for data item %s is nil", j, item.DataId)
			}

			if verification.VerifierAddress == "" {
				return fmt.Errorf("verification at index %d for data item %s has empty verifier", j, item.DataId)
			}

			if _, exists := verifierAddrs[verification.VerifierAddress]; exists {
				return fmt.Errorf("duplicate verifier %s for data item %s", verification.VerifierAddress, item.DataId)
			}
			verifierAddrs[verification.VerifierAddress] = struct{}{}

			if verification.Level == VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED {
				return fmt.Errorf("verification for data item %s has unspecified level", item.DataId)
			}

			if verification.VerifiedAt == nil {
				return fmt.Errorf("verification for data item %s has nil verified_at timestamp", item.DataId)
			}
		}

		// Note: Type-specific data (VehicleRegistrationData, PhotoData, GolfScoreData)
		// are stored separately and referenced via metadata, not directly on DataItem.
		// They can be stored in the storage_location (IPFS/Arweave) or in metadata map.

		// Validate metadata
		if item.Metadata != nil {
			for key := range item.Metadata {
				if key == "" {
					return fmt.Errorf("data item %s has metadata with empty key", item.DataId)
				}
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
