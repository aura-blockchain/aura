package types

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
