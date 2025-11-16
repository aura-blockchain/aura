package types

const ModuleName = "inclusionroutines"

// Store key prefixes
const (
	// IRStoreKeyPrefix is the prefix for storing IR definitions
	IRStoreKeyPrefix = "ir/"

	// PrerequisiteStoreKeyPrefix is the prefix for storing IR prerequisites
	PrerequisiteStoreKeyPrefix = "prereq/"

	// RateLimitStoreKeyPrefix is the prefix for storing rate limit configurations
	RateLimitStoreKeyPrefix = "ratelimit/"

	// RateLimitUsageKeyPrefix is the prefix for storing rate limit usage counters
	RateLimitUsageKeyPrefix = "ratelimitusage/"
)

// IRStoreKey returns the store key for a specific IR definition
func IRStoreKey(irID string) string {
	return IRStoreKeyPrefix + irID
}

// PrerequisiteStoreKey returns the store key for IR prerequisites
func PrerequisiteStoreKey(irID string) string {
	return PrerequisiteStoreKeyPrefix + irID
}

// RateLimitStoreKey returns the store key for IR rate limits
func RateLimitStoreKey(irID string) string {
	return RateLimitStoreKeyPrefix + irID
}

// RateLimitUsageKey returns the store key for rate limit usage tracking
// Format: ratelimitusage/{ir_id}/{wallet}/{time_window}
func RateLimitUsageKey(irID, wallet, timeWindow string) string {
	return RateLimitUsageKeyPrefix + irID + "/" + wallet + "/" + timeWindow
}
