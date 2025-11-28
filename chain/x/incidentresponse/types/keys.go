package types

const (
	// ModuleName defines the module name
	ModuleName = "incidentresponse"

	// StoreKey is the store key string for incidentresponse
	StoreKey = ModuleName

	// RouterKey is the message route for incidentresponse
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

// State store key prefixes
var (
	// IncidentKeyPrefix is the prefix for incident records
	IncidentKeyPrefix = []byte{0x01}

	// PauseStateKey is the single key for chain pause state
	PauseStateKey = []byte{0x02}

	// PauseVoteKeyPrefix is the prefix for pause vote records
	PauseVoteKeyPrefix = []byte{0x03}

	// WalletLimitKeyPrefix is the prefix for wallet limits
	WalletLimitKeyPrefix = []byte{0x04}

	// NextIncidentIDKey is the key for the next incident ID counter
	NextIncidentIDKey = []byte{0x05}

	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x06}
)

// IncidentKey returns the key for an incident record
func IncidentKey(incidentID string) []byte {
	return append(IncidentKeyPrefix, []byte(incidentID)...)
}

// PauseVoteKey returns the key for a pause vote record
func PauseVoteKey(pauseRequestID string, signer string) []byte {
	key := append(PauseVoteKeyPrefix, []byte(pauseRequestID)...)
	key = append(key, byte(':'))
	return append(key, []byte(signer)...)
}

// PauseVotePrefixKey returns the prefix key for all votes on a pause request
func PauseVotePrefixKey(pauseRequestID string) []byte {
	return append(PauseVoteKeyPrefix, []byte(pauseRequestID)...)
}

// WalletLimitKey returns the key for a wallet limit
func WalletLimitKey(address string) []byte {
	return append(WalletLimitKeyPrefix, []byte(address)...)
}
