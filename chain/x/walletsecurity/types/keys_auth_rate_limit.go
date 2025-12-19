package types

import "fmt"

const (
	// AuthRateLimitPrefix is used to track per-height auth attempts.
	AuthRateLimitPrefix = "auth_rate_limit"
)

// GetAuthRateLimitKey returns the store key for auth attempt tracking scoped by block height.
func GetAuthRateLimitKey(blockHeight int64, address string) []byte {
	return []byte(fmt.Sprintf("%s/%d/%s", AuthRateLimitPrefix, blockHeight, address))
}
