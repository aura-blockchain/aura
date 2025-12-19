package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ComplianceKeeper defines the methods the prevalidation module expects for sanctions checks.
type ComplianceKeeper interface {
	// IsAddressSanctioned returns true if the given address is sanctioned/blocklisted.
	IsAddressSanctioned(ctx sdk.Context, address string) bool
}
