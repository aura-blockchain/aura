package types

const (
	// ModuleName defines the module name
	ModuleName = "vcregistry"

	// StoreKey is the store key string for vcregistry
	StoreKey = ModuleName

	// RouterKey is the message route for vcregistry
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

// State store key prefixes
var (
	// VCRecordKeyPrefix is the prefix for VC records
	VCRecordKeyPrefix = []byte{0x01}

	// UserVCIndexKeyPrefix is the prefix for user VC index
	UserVCIndexKeyPrefix = []byte{0x02}

	// RevocationRecordKeyPrefix is the prefix for revocation records
	RevocationRecordKeyPrefix = []byte{0x03}

	// RevocationListKey is the single key for the revocation list
	RevocationListKey = []byte{0x04}

	// DIDDocumentKeyPrefix is the prefix for DID documents
	DIDDocumentKeyPrefix = []byte{0x05}

	// AddressToDIDIndexKeyPrefix is the prefix for address to DID index
	AddressToDIDIndexKeyPrefix = []byte{0x06}

	// VCPolicyKeyPrefix is the prefix for VC policies
	VCPolicyKeyPrefix = []byte{0x07}

	// UserMintCountKeyPrefix is the prefix for user mint counts
	UserMintCountKeyPrefix = []byte{0x08}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x09}

	// PresentationKeyPrefix is the prefix for presentations
	PresentationKeyPrefix = []byte{0x0a}

	// NonceKeyPrefix is the prefix for used nonces
	NonceKeyPrefix = []byte{0x0b}

	// AttributeVCKeyPrefix is the prefix for attribute VCs
	AttributeVCKeyPrefix = []byte{0x0c}

	// UserAttributeVCIndexKeyPrefix is the prefix for user attribute VC index
	UserAttributeVCIndexKeyPrefix = []byte{0x0d}

	// DisclosurePolicyKeyPrefix is the prefix for disclosure policies
	DisclosurePolicyKeyPrefix = []byte{0x0e}

	// DisclosureRequestKeyPrefix is the prefix for disclosure requests
	DisclosureRequestKeyPrefix = []byte{0x0f}

	// DisclosureResponseKeyPrefix is the prefix for disclosure responses
	DisclosureResponseKeyPrefix = []byte{0x10}

	// UserPresentationIndexKeyPrefix is the prefix for user presentation index
	UserPresentationIndexKeyPrefix = []byte{0x11}

	// PendingDisclosureRequestKeyPrefix is the prefix for pending disclosure requests by holder
	PendingDisclosureRequestKeyPrefix = []byte{0x12}
)

// VCRecordKey returns the key for a VC record
func VCRecordKey(vcID string) []byte {
	return append(VCRecordKeyPrefix, []byte(vcID)...)
}

// UserVCIndexKey returns the key for user VC index
func UserVCIndexKey(address string) []byte {
	return append(UserVCIndexKeyPrefix, []byte(address)...)
}

// RevocationRecordKey returns the key for a revocation record
func RevocationRecordKey(vcID string) []byte {
	return append(RevocationRecordKeyPrefix, []byte(vcID)...)
}

// DIDDocumentKey returns the key for a DID document
func DIDDocumentKey(did string) []byte {
	return append(DIDDocumentKeyPrefix, []byte(did)...)
}

// AddressToDIDIndexKey returns the key for address to DID index
func AddressToDIDIndexKey(address string) []byte {
	return append(AddressToDIDIndexKeyPrefix, []byte(address)...)
}

// VCPolicyKey returns the key for a VC policy
func VCPolicyKey(vcTypeName string) []byte {
	return append(VCPolicyKeyPrefix, []byte(vcTypeName)...)
}

// UserMintCountKey returns the key for user mint count
func UserMintCountKey(address string, dayTimestamp int64) []byte {
	key := append(UserMintCountKeyPrefix, []byte(address)...)
	// Append day timestamp as bytes
	timestampBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		timestampBytes[i] = byte(dayTimestamp >> (8 * (7 - i)))
	}
	return append(key, timestampBytes...)
}

// PresentationKey returns the key for a presentation
func PresentationKey(presentationID string) []byte {
	return append(PresentationKeyPrefix, []byte(presentationID)...)
}

// NonceKey returns the key for a nonce
func NonceKey(nonce uint64) []byte {
	nonceBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		nonceBytes[i] = byte(nonce >> (8 * (7 - i)))
	}
	return append(NonceKeyPrefix, nonceBytes...)
}

// AttributeVCKey returns the key for an attribute VC
func AttributeVCKey(attributeVCID string) []byte {
	return append(AttributeVCKeyPrefix, []byte(attributeVCID)...)
}

// UserAttributeVCIndexKey returns the key for user attribute VC index
func UserAttributeVCIndexKey(address string) []byte {
	return append(UserAttributeVCIndexKeyPrefix, []byte(address)...)
}

// DisclosurePolicyKey returns the key for a disclosure policy
func DisclosurePolicyKey(holderAddress string) []byte {
	return append(DisclosurePolicyKeyPrefix, []byte(holderAddress)...)
}

// DisclosureRequestKey returns the key for a disclosure request
func DisclosureRequestKey(requestID string) []byte {
	return append(DisclosureRequestKeyPrefix, []byte(requestID)...)
}

// DisclosureResponseKey returns the key for a disclosure response
func DisclosureResponseKey(requestID string) []byte {
	return append(DisclosureResponseKeyPrefix, []byte(requestID)...)
}

// UserPresentationIndexKey returns the key for user presentation index
func UserPresentationIndexKey(address string) []byte {
	return append(UserPresentationIndexKeyPrefix, []byte(address)...)
}

// PendingDisclosureRequestKey returns the key for pending disclosure request
func PendingDisclosureRequestKey(holderAddress, requestID string) []byte {
	key := append(PendingDisclosureRequestKeyPrefix, []byte(holderAddress)...)
	key = append(key, byte(':'))
	return append(key, []byte(requestID)...)
}
