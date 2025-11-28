package v1beta1

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/common/validation"
)

const (
	// MaxContentHashSize is the maximum size for content hashes
	MaxContentHashSize = 128
	// MinContentHashSize is the minimum size for content hashes (32 bytes for SHA-256)
	MinContentHashSize = 32
	// MaxStorageLocationLength is the maximum length for storage location (IPFS CID)
	MaxStorageLocationLength = 256
	// MinStorageLocationLength is the minimum length for storage location
	MinStorageLocationLength = 1
	// MaxSpecializedDataSize is the maximum size for specialized data (1MB)
	MaxSpecializedDataSize = 1024 * 1024
	// MaxTagsCount is the maximum number of tags
	MaxTagsCount = 50
	// MaxTagLength is the maximum length for a tag
	MaxTagLength = 64
	// MaxProofSize is the maximum size for verification proof
	MaxProofSize = 4096
	// MaxNotesLength is the maximum length for verification notes
	MaxNotesLength = 1000
	// MaxReasonLength is the maximum length for revocation reason
	MaxReasonLength = 500
	// MaxMethodLength is the maximum length for verification method
	MaxMethodLength = 128
	// MaxConfidenceScore is the maximum confidence score
	MaxConfidenceScore = uint64(100)
)

// ValidateBasic implements the sdk.Msg interface for MsgStoreDataItem
func (m *MsgStoreDataItem) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Data type enum is validated at protobuf level

	// Validate title
	if err := validation.ValidateBoundedString(m.Title, 1, validation.MaxNameLength, "title"); err != nil {
		return err
	}

	// Validate description
	if err := validation.ValidateBoundedString(m.Description, 1, validation.MaxDescriptionLength, "description"); err != nil {
		return err
	}

	// Validate content hash
	if err := validation.ValidateBytes(m.ContentHash, MinContentHashSize, MaxContentHashSize, "content_hash"); err != nil {
		return err
	}

	// Validate storage location (IPFS CID)
	if err := validation.ValidateBoundedString(m.StorageLocation, MinStorageLocationLength, MaxStorageLocationLength, "storage_location"); err != nil {
		return err
	}

	// Validate specialized data (optional)
	if len(m.SpecializedData) > 0 {
		if err := validation.ValidateBytes(m.SpecializedData, 1, MaxSpecializedDataSize, "specialized_data"); err != nil {
			return err
		}
	}

	// Validate tags
	if len(m.Tags) > MaxTagsCount {
		return fmt.Errorf("tags: cannot exceed %d tags, got %d", MaxTagsCount, len(m.Tags))
	}

	for i, tag := range m.Tags {
		if err := validation.ValidateBoundedString(tag, 1, MaxTagLength, fmt.Sprintf("tags[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateDataItem
func (m *MsgUpdateDataItem) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate data ID
	if err := validation.ValidateID(m.DataId, "data_id"); err != nil {
		return err
	}

	// Validate title (optional, but if present must be valid)
	if m.Title != "" {
		if err := validation.ValidateBoundedString(m.Title, 1, validation.MaxNameLength, "title"); err != nil {
			return err
		}
	}

	// Validate description (optional, but if present must be valid)
	if m.Description != "" {
		if err := validation.ValidateBoundedString(m.Description, 1, validation.MaxDescriptionLength, "description"); err != nil {
			return err
		}
	}

	// Validate tags
	if len(m.Tags) > MaxTagsCount {
		return fmt.Errorf("tags: cannot exceed %d tags, got %d", MaxTagsCount, len(m.Tags))
	}

	for i, tag := range m.Tags {
		if err := validation.ValidateBoundedString(tag, 1, MaxTagLength, fmt.Sprintf("tags[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDeleteDataItem
func (m *MsgDeleteDataItem) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate data ID
	if err := validation.ValidateID(m.DataId, "data_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgVerifyDataItem
func (m *MsgVerifyDataItem) ValidateBasic() error {
	// Validate verifier address
	if err := validation.ValidateAccAddress(m.Verifier); err != nil {
		return fmt.Errorf("verifier: %w", err)
	}

	// Validate data ID
	if err := validation.ValidateID(m.DataId, "data_id"); err != nil {
		return err
	}

	// Verification level enum is validated at protobuf level

	// Validate confidence score
	if m.ConfidenceScore > MaxConfidenceScore {
		return fmt.Errorf("confidence_score cannot exceed %d, got %d", MaxConfidenceScore, m.ConfidenceScore)
	}

	// Validate notes (optional)
	if m.Notes != "" {
		if err := validation.ValidateBoundedString(m.Notes, 0, MaxNotesLength, "notes"); err != nil {
			return err
		}
	}

	// Validate verification method
	if err := validation.ValidateBoundedString(m.VerificationMethod, 1, MaxMethodLength, "verification_method"); err != nil {
		return err
	}

	// Validate proof (optional)
	if len(m.Proof) > 0 {
		if err := validation.ValidateBytes(m.Proof, 1, MaxProofSize, "proof"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRevokeDataItem
func (m *MsgRevokeDataItem) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate data ID
	if err := validation.ValidateID(m.DataId, "data_id"); err != nil {
		return err
	}

	// Validate reason
	if err := validation.ValidateBoundedString(m.Reason, 1, MaxReasonLength, "reason"); err != nil {
		return err
	}

	return nil
}
