package types

import "errors"

var (
	// ErrInvalidDataID is returned when a data ID is invalid
	ErrInvalidDataID = errors.New("invalid data ID")

	// ErrDataItemNotFound is returned when a data item is not found
	ErrDataItemNotFound = errors.New("data item not found")

	// ErrInvalidOwner is returned when the owner address is invalid
	ErrInvalidOwner = errors.New("invalid owner address")

	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrAccessDenied is returned when access is denied
	ErrAccessDenied = errors.New("access denied")

	// ErrInvalidDataType is returned when data type is invalid
	ErrInvalidDataType = errors.New("invalid data type")

	// ErrInvalidContentHash is returned when content hash is invalid
	ErrInvalidContentHash = errors.New("invalid content hash")

	// ErrInvalidStorageLocation is returned when storage location is invalid
	ErrInvalidStorageLocation = errors.New("invalid storage location")

	// ErrDataItemAlreadyExists is returned when data item already exists
	ErrDataItemAlreadyExists = errors.New("data item already exists")

	// ErrDataItemRevoked is returned when data item is revoked
	ErrDataItemRevoked = errors.New("data item is revoked")

	// ErrDataItemExpired is returned when data item is expired
	ErrDataItemExpired = errors.New("data item is expired")

	// ErrInvalidVerifier is returned when verifier is invalid
	ErrInvalidVerifier = errors.New("invalid verifier")

	// ErrInvalidVerificationLevel is returned when verification level is invalid
	ErrInvalidVerificationLevel = errors.New("invalid verification level")

	// ErrMaxDataItemsExceeded is returned when max data items limit is exceeded
	ErrMaxDataItemsExceeded = errors.New("maximum data items per user exceeded")

	// ErrStorageSizeExceeded is returned when storage size limit is exceeded
	ErrStorageSizeExceeded = errors.New("storage size limit exceeded")

	// ErrInvalidGeoLocation is returned when geo location is invalid
	ErrInvalidGeoLocation = errors.New("invalid geo location")

	// ErrInvalidAccessPolicy is returned when access policy is invalid
	ErrInvalidAccessPolicy = errors.New("invalid access policy")
)
