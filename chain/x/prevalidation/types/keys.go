package types

const (
	// ModuleName defines the module name
	ModuleName = "prevalidation"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for prevalidation
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore keys
var (
	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x01}

	// PreValidatedTransactionPrefix is the prefix for pre-validated transactions
	PreValidatedTransactionPrefix = []byte{0x02}

	// PreValidatedTxPrefix is an alias for PreValidatedTransactionPrefix
	PreValidatedTxPrefix = []byte{0x02}

	// ValidationTemplatePrefix is the prefix for validation templates
	ValidationTemplatePrefix = []byte{0x03}

	// MetricsPrefix is the prefix for metrics
	MetricsPrefix = []byte{0x04}

	// MetricsKey is the key for storing metrics
	MetricsKey = []byte{0x04, 0x01}

	// SchedulerStatePrefix is the prefix for scheduler state
	SchedulerStatePrefix = []byte{0x05}

	// CachePrefix is the prefix for cached validations
	CachePrefix = []byte{0x06}
)

// GetPreValidatedTxKey returns the store key for a pre-validated transaction (alias)
func GetPreValidatedTxKey(txID string) []byte {
	return append(PreValidatedTxPrefix, []byte(txID)...)
}

// GetPreValidatedTransactionKey returns the store key for a pre-validated transaction
func GetPreValidatedTransactionKey(txID string) []byte {
	return append(PreValidatedTransactionPrefix, []byte(txID)...)
}

// GetValidationTemplateKey returns the store key for a validation template
func GetValidationTemplateKey(templateID string) []byte {
	return append(ValidationTemplatePrefix, []byte(templateID)...)
}

// GetCacheKey returns the store key for a cached validation
func GetCacheKey(cacheKey string) []byte {
	return append(CachePrefix, []byte(cacheKey)...)
}
