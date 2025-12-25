// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ComplianceKeeper defines the methods the prevalidation module expects for sanctions checks.
type ComplianceKeeper interface {
	// IsAddressSanctioned returns true if the given address is sanctioned/blocklisted.
	IsAddressSanctioned(ctx sdk.Context, address string) bool
}
