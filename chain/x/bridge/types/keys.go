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
