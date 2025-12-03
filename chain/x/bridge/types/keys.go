package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	// Cross-chain transfer keys
	TransferPrefix = []byte{0x01}

	// Wrapped token keys
	WrappedTokenPrefix = []byte{0x02}

	// Shared identity keys
	SharedIdentityPrefix = []byte{0x03}

	// Relayer keys
	RelayerPrefix = []byte{0x04}

	// Chain configuration keys
	ChainConfigPrefix = []byte{0x05}

	// Validator metadata
	ValidatorPrefix = []byte{0x06}

	// Attestations (transfer + validator)
	AttestationPrefix = []byte{0x07}

	// Relayer statistics
	RelayerStatsPrefix = []byte{0x08}

	// Cross-chain swaps
	SwapPrefix = []byte{0x09}

	// Transfer hash index (tx hash -> transfer ID)
	TransferHashIndexPrefix = []byte{0x0a}

	// Auto-increment counter key
	TransferCounterKey = []byte{0x0b}

	// Processed source hashes (replay attack prevention)
	ProcessedSourceHashPrefix = []byte{0x0c}

	// Signature set tracking (prevent reusing same signature sets)
	SignatureSetPrefix = []byte{0x0e}

	// Validator snapshot tracking (for historical validator sets)
	ValidatorSnapshotPrefix = []byte{0x0f}

	// Daily mint tracking (supply cap enforcement)
	DailyMintPrefix = []byte{0x20}

	// Hourly mint tracking (rate limiting)
	HourlyMintPrefix = []byte{0x21}

	// Verified block hashes (for Merkle proof verification)
	VerifiedBlockHashPrefix = []byte{0x22}

	// Pending transfers (awaiting fraud proof window expiry)
	PendingTransferPrefix = []byte{0x23}

	// Unlock nonces (replay protection for unlock operations)
	UnlockNoncePrefix = []byte{0x24}
)

// TransferKey returns the store key for a cross-chain transfer
func TransferKey(transferID string) []byte {
	return append(TransferPrefix, []byte(transferID)...)
}

// WrappedTokenKey returns the store key for a wrapped token
func WrappedTokenKey(wrappedDenom string) []byte {
	return append(WrappedTokenPrefix, []byte(wrappedDenom)...)
}

// SharedIdentityKey returns the store key for a shared identity
func SharedIdentityKey(auraAddress string) []byte {
	return append(SharedIdentityPrefix, []byte(auraAddress)...)
}

// RelayerKey returns the store key for a relayer
func RelayerKey(relayerAddress string) []byte {
	return append(RelayerPrefix, []byte(relayerAddress)...)
}

// ChainConfigKey returns the store key for a chain configuration
func ChainConfigKey(chainID string) []byte {
	return append(ChainConfigPrefix, []byte(chainID)...)
}

// ValidatorKey returns the key for a bridge validator entry
func ValidatorKey(address string) []byte {
	return append(ValidatorPrefix, []byte(address)...)
}

// AttestationKey stores attestations per transfer/validator pair
func AttestationKey(transferID, validator string) []byte {
	key := append(AttestationPrefix, []byte(transferID)...)
	key = append(key, byte(0x00))
	return append(key, []byte(validator)...)
}

// TransferHashIndexKey maintains a mapping from external hashes to transfer IDs
func TransferHashIndexKey(hash string) []byte {
	return append(TransferHashIndexPrefix, []byte(hash)...)
}

// SwapKey returns the store key for a cross-chain swap
func SwapKey(swapID string) []byte {
	return append(SwapPrefix, []byte(swapID)...)
}

// ProcessedSourceHashKey returns the store key for a processed source hash
// Format: ProcessedSourceHashPrefix + sourceChain:sourceHash
func ProcessedSourceHashKey(sourceChain, sourceHash string) []byte {
	compositeKey := sourceChain + ":" + sourceHash
	return append(ProcessedSourceHashPrefix, []byte(compositeKey)...)
}

// SignatureSetKey returns the store key for a signature set hash
// Format: SignatureSetPrefix + transferID + ":" + signatureSetHash
func SignatureSetKey(transferID string, signatureSetHash []byte) []byte {
	key := append(SignatureSetPrefix, []byte(transferID)...)
	key = append(key, byte(0x00))
	return append(key, signatureSetHash...)
}

// ValidatorSnapshotKey returns the store key for a validator set snapshot at a specific height
// Format: ValidatorSnapshotPrefix + blockHeight (8 bytes big-endian)
func ValidatorSnapshotKey(blockHeight int64) []byte {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(blockHeight))
	return append(ValidatorSnapshotPrefix, heightBytes...)
}

// DailyMintKey returns the store key for daily mint tracking
// Format: DailyMintPrefix + date (YYYYMMDD) + ":" + denom
// The date is derived from the current block time
func DailyMintKey(date string, denom string) []byte {
	compositeKey := date + ":" + denom
	return append(DailyMintPrefix, []byte(compositeKey)...)
}

// HourlyMintKey returns the store key for hourly mint tracking
// Format: HourlyMintPrefix + datetime (YYYYMMDDHH) + ":" + denom
// The datetime is derived from the current block time
func HourlyMintKey(datetime string, denom string) []byte {
	compositeKey := datetime + ":" + denom
	return append(HourlyMintPrefix, []byte(compositeKey)...)
}

// VerifiedBlockHashKey returns the store key for a verified block hash
// Format: VerifiedBlockHashPrefix + sourceChain + ":" + blockHeight (8 bytes big-endian)
func VerifiedBlockHashKey(sourceChain string, blockHeight uint64) []byte {
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, blockHeight)
	compositeKey := sourceChain + ":"
	key := append(VerifiedBlockHashPrefix, []byte(compositeKey)...)
	return append(key, heightBytes...)
}

// PendingTransferKey returns the store key for a pending transfer
// Format: PendingTransferPrefix + transferID
func PendingTransferKey(transferID string) []byte {
	return append(PendingTransferPrefix, []byte(transferID)...)
}

// UnlockNonceKey returns the store key for unlock nonce tracking
// Format: UnlockNoncePrefix + transferID
// This tracks how many times unlock has been attempted for a transfer (replay protection)
func UnlockNonceKey(transferID string) []byte {
	return append(UnlockNoncePrefix, []byte(transferID)...)
}
