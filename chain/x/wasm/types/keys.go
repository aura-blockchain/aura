package types

const (
	// ModuleName defines the module name
	ModuleName = "aura_wasm_security"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for wasm
	RouterKey = ModuleName

	// QuerierRoute is the querier route for wasm
	QuerierRoute = ModuleName
)

// KVStore keys
var (
	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x01}

	// ContractAuthKey is the key for storing authorized contract uploaders
	ContractAuthKey = []byte{0x02}
	AuthorizedUploaderPrefix = []byte{0x02}

	// ContractPauseKey is the key for storing paused contracts
	ContractPauseKey = []byte{0x03}
	PausedContractPrefix = []byte{0x03}

	// SecurityStatsKey is the key for storing security statistics
	SecurityStatsKey = []byte{0x04}

	// ContractAdminPrefix is the key prefix for storing contract admins
	ContractAdminPrefix = []byte{0x06}
)

// GetContractAuthKey returns the key for a specific authorized uploader
func GetContractAuthKey(address string) []byte {
	return append(ContractAuthKey, []byte(address)...)
}

// GetContractAdminKey returns the key for a specific contract's admin
func GetContractAdminKey(contractAddr string) []byte {
	return append(ContractAdminPrefix, []byte(contractAddr)...)
}

// GetAuthorizedUploaderKey returns the key for a specific authorized uploader
func GetAuthorizedUploaderKey(address string) []byte {
	return append(AuthorizedUploaderPrefix, []byte(address)...)
}

// GetContractPauseKey returns the key for a specific paused contract
func GetContractPauseKey(contractAddr string) []byte {
	return append(ContractPauseKey, []byte(contractAddr)...)
}

// GetPausedContractKey returns the key for a specific paused contract
func GetPausedContractKey(contractAddr string) []byte {
	return append(PausedContractPrefix, []byte(contractAddr)...)
}

// GetContractExecutingKey returns the key for tracking contract execution (reentrancy protection)
func GetContractExecutingKey(contractAddr string) []byte {
	return append([]byte{0x05}, []byte(contractAddr)...)
}
