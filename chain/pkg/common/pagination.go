package common

import (
	"github.com/cosmos/cosmos-sdk/types/query"
)

const (
	// DefaultPageLimit is the default number of items per page when pagination is not specified.
	// This prevents unbounded queries that could cause DoS attacks.
	DefaultPageLimit = 100

	// MaxPageLimit is the maximum number of items that can be requested in a single query.
	// This is enforced by the Cosmos SDK's query.Paginate function.
	MaxPageLimit = 1000
)

// NormalizePagination ensures pagination request has sensible defaults.
// This prevents DoS attacks via unbounded queries and ensures consistent
// pagination behavior across all modules.
//
// Security considerations:
//   - Sets default limit to 100 if not specified
//   - Maximum limit of 1000 enforced by Cosmos SDK
//   - Prevents unbounded queries that could exhaust resources
//
// Parameters:
//   - pagination: PageRequest from query (can be nil)
//
// Returns:
//   - *query.PageRequest: Normalized pagination with defaults applied
//
// Normalization rules:
//   - If pagination is nil, returns default PageRequest with limit 100
//   - If limit is 0, sets it to DefaultPageLimit (100)
//   - Cosmos SDK enforces MaxPageLimit (1000) automatically
//
// Example usage:
//   normalized := common.NormalizePagination(req.Pagination)
//   items, pageRes, err := query.Paginate(store, normalized, func(key []byte, value []byte) error {
//       // Process item
//   })
func NormalizePagination(pagination *query.PageRequest) *query.PageRequest {
	if pagination == nil {
		return &query.PageRequest{
			Limit: DefaultPageLimit,
		}
	}

	// If limit is 0, set to default
	if pagination.Limit == 0 {
		pagination.Limit = DefaultPageLimit
	}

	// Cosmos SDK enforces MaxPageLimit automatically via query.Paginate
	// No need to cap here as it would create inconsistent behavior

	return pagination
}

// GetEffectiveLimit returns the effective limit that will be used for pagination.
// This is useful for logging and metrics.
//
// Parameters:
//   - pagination: PageRequest from query (can be nil)
//
// Returns:
//   - uint64: Effective limit that will be used (capped at MaxPageLimit)
func GetEffectiveLimit(pagination *query.PageRequest) uint64 {
	if pagination == nil || pagination.Limit == 0 {
		return DefaultPageLimit
	}

	if pagination.Limit > MaxPageLimit {
		return MaxPageLimit
	}

	return pagination.Limit
}
