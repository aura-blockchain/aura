package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	// Test error definitions
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidDataID", ErrInvalidDataID},
		{"ErrDataItemNotFound", ErrDataItemNotFound},
		{"ErrInvalidOwner", ErrInvalidOwner},
		{"ErrUnauthorized", ErrUnauthorized},
		{"ErrAccessDenied", ErrAccessDenied},
		{"ErrInvalidDataType", ErrInvalidDataType},
		{"ErrInvalidContentHash", ErrInvalidContentHash},
		{"ErrInvalidStorageLocation", ErrInvalidStorageLocation},
		{"ErrDataItemAlreadyExists", ErrDataItemAlreadyExists},
		{"ErrDataItemRevoked", ErrDataItemRevoked},
		{"ErrDataItemExpired", ErrDataItemExpired},
		{"ErrInvalidVerifier", ErrInvalidVerifier},
		{"ErrInvalidVerificationLevel", ErrInvalidVerificationLevel},
		{"ErrMaxDataItemsExceeded", ErrMaxDataItemsExceeded},
		{"ErrStorageSizeExceeded", ErrStorageSizeExceeded},
		{"ErrInvalidGeoLocation", ErrInvalidGeoLocation},
		{"ErrInvalidAccessPolicy", ErrInvalidAccessPolicy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.err)
			require.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestErrorMessages(t *testing.T) {
	require.Contains(t, ErrInvalidDataID.Error(), "invalid data ID")
	require.Contains(t, ErrDataItemNotFound.Error(), "data item not found")
	require.Contains(t, ErrInvalidOwner.Error(), "invalid owner")
	require.Contains(t, ErrUnauthorized.Error(), "unauthorized")
	require.Contains(t, ErrAccessDenied.Error(), "access denied")
	require.Contains(t, ErrInvalidDataType.Error(), "invalid data type")
	require.Contains(t, ErrInvalidContentHash.Error(), "invalid content hash")
	require.Contains(t, ErrDataItemAlreadyExists.Error(), "already exists")
	require.Contains(t, ErrMaxDataItemsExceeded.Error(), "maximum data items")
	require.Contains(t, ErrStorageSizeExceeded.Error(), "storage size")
}

func TestErrorUniqueness(t *testing.T) {
	// Ensure all errors are distinct
	errors := []error{
		ErrInvalidDataID,
		ErrDataItemNotFound,
		ErrInvalidOwner,
		ErrUnauthorized,
		ErrAccessDenied,
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j {
				require.NotEqual(t, err1, err2)
			}
		}
	}
}
