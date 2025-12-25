// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

//lint:file-ignore SA1019 -- use invariant registry until Cosmos SDK removes the legacy APIs

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all module invariants
//nolint:staticcheck // invariant registry uses deprecated SDK interfaces until crisis removal
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "query-stats-non-negative",
		QueryStatsNonNegativeInvariant(k))
	ir.RegisterRoute(types.ModuleName, "message-stats-non-negative",
		MessageStatsNonNegativeInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rate-limits-valid",
		RateLimitsValidInvariant(k))
	ir.RegisterRoute(types.ModuleName, "state-consistency",
		StateConsistencyInvariant(k))
}

// QueryStatsNonNegativeInvariant checks that all query statistics are non-negative
func QueryStatsNonNegativeInvariant(k *Keeper) sdk.Invariant { //nolint:staticcheck // invariant signature relies on deprecated type
	return func(ctx sdk.Context) (string, bool) {
		k.mu.RLock()
		defer k.mu.RUnlock()

		var (
			msg   string
			count int
		)

		for queryType, stat := range k.queryStats {
			// uint64 is always non-negative, but check for consistency
			if queryType == "" {
				count++
				msg += fmt.Sprintf("\tempty query type with stat %d\n", stat)
			}
		}

		broken := count > 0
		//nolint:staticcheck // FormatInvariant is deprecated but still required until crisis removal
		return sdk.FormatInvariant(
			types.ModuleName, "query-stats-non-negative",
			fmt.Sprintf("%d invalid query stats found\n%s", count, msg),
		), broken
	}
}

// MessageStatsNonNegativeInvariant checks that all message statistics are non-negative
func MessageStatsNonNegativeInvariant(k *Keeper) sdk.Invariant { //nolint:staticcheck // invariant signature relies on deprecated type
	return func(ctx sdk.Context) (string, bool) {
		k.mu.RLock()
		defer k.mu.RUnlock()

		var (
			msg   string
			count int
		)

		for msgType, stat := range k.messageStats {
			// uint64 is always non-negative, but check for consistency
			if msgType == "" {
				count++
				msg += fmt.Sprintf("\tempty message type with stat %d\n", stat)
			}
		}

		broken := count > 0
		//nolint:staticcheck // FormatInvariant is deprecated but still required until crisis removal
		return sdk.FormatInvariant(
			types.ModuleName, "message-stats-non-negative",
			fmt.Sprintf("%d invalid message stats found\n%s", count, msg),
		), broken
	}
}

// RateLimitsValidInvariant checks that rate limits are within valid bounds
func RateLimitsValidInvariant(k *Keeper) sdk.Invariant { //nolint:staticcheck // invariant signature relies on deprecated type
	return func(ctx sdk.Context) (string, bool) {
		k.mu.RLock()
		defer k.mu.RUnlock()

		var (
			msg   string
			count int
		)

		for address, limit := range k.queryRateLimits {
			if address == "" {
				count++
				msg += fmt.Sprintf("\tempty address with rate limit %d\n", limit)
			}

			if limit < 0 {
				count++
				msg += fmt.Sprintf("\taddress %s has negative rate limit %d\n", address, limit)
			}

			if limit > types.MaxQueriesPerBlock {
				count++
				msg += fmt.Sprintf("\taddress %s exceeds max queries per block: %d > %d\n",
					address, limit, types.MaxQueriesPerBlock)
			}
		}

		broken := count > 0
		//nolint:staticcheck // FormatInvariant is deprecated but still required until crisis removal
		return sdk.FormatInvariant(
			types.ModuleName, "rate-limits-valid",
			fmt.Sprintf("%d invalid rate limits found\n%s", count, msg),
		), broken
	}
}

// StateConsistencyInvariant checks overall state consistency
func StateConsistencyInvariant(k *Keeper) sdk.Invariant { //nolint:staticcheck // invariant signature relies on deprecated type
	return func(ctx sdk.Context) (string, bool) {
		k.mu.RLock()
		defer k.mu.RUnlock()

		var (
			msg   string
			count int
		)

		// Check that maps are initialized
		if k.queryStats == nil {
			count++
			msg += "\tquery stats map is nil\n"
		}

		if k.messageStats == nil {
			count++
			msg += "\tmessage stats map is nil\n"
		}

		if k.queryRateLimits == nil {
			count++
			msg += "\tquery rate limits map is nil\n"
		}

		// Check that current block height is valid
		if k.currentBlock < 0 {
			count++
			msg += fmt.Sprintf("\tcurrent block is negative: %d\n", k.currentBlock)
		}

		// Check that the current block matches or is less than context block
		if k.currentBlock > ctx.BlockHeight() {
			count++
			msg += fmt.Sprintf("\tcurrent block %d exceeds context block %d\n",
				k.currentBlock, ctx.BlockHeight())
		}

		broken := count > 0
		//nolint:staticcheck // FormatInvariant is deprecated but still required until crisis removal
		return sdk.FormatInvariant(
			types.ModuleName, "state-consistency",
			fmt.Sprintf("%d state consistency issues found\n%s", count, msg),
		), broken
	}
}

// AllInvariants runs all invariants of the module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := QueryStatsNonNegativeInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = MessageStatsNonNegativeInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = RateLimitsValidInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		return StateConsistencyInvariant(k)(ctx)
	}
}
