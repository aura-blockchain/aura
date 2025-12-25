// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// BeginBlocker processes expired KYC records at the beginning of each block.
// This implements automatic KYC expiry monitoring and enforcement.
//
// Processing workflow:
//   1. Every block: Check expiration index for expired records (O(k) where k = expired count)
//   2. Batch limit: Process max 100 expired records per block to prevent long blocks
//   3. Emit EventTypeKYCExpired for each newly expired record
//   4. Remove processed records from expiration index to prevent re-processing
//
// Performance optimization:
//   - Uses expiration index for O(k) lookup instead of O(n) full scan
//   - Batch limit prevents chain halt: max 100 records per block
//   - With 100K+ records: only processes expired ones (typically 0-10 per block)
//   - Index cleanup prevents re-processing same records
//   - No longer needs 50-block batching - runs every block efficiently
//
// The events emitted allow:
//   - Off-chain monitoring systems to detect expiry
//   - User notifications for re-verification
//   - Compliance audit trails
//   - Indexers to track expiry status
//
// Security considerations:
//   - Read-only: Does not modify KYC records (immutable audit trail)
//   - Time-based: Uses blockchain time (cannot be manipulated)
//   - Event-driven: External systems react to events
//   - Gas-bounded: Max 100 records per block prevents DoS
//   - Index cleanup: Prevents duplicate event emission
//
// Compliance:
//   - FinCEN: Continuous KYC status monitoring
//   - FATF Recommendation 10: Ongoing due diligence enforcement
//   - BSA: Customer verification status tracking
//   - Provides immutable audit trail of all expiry events
//   - Real-time detection (no batching delay)
//
// Events emitted:
//   - EventTypeKYCExpired: For each record that has expired
//     Attributes: address, expired_at, kyc_level, provider
//
// Example event consumer (off-chain):
//   // Listen for kyc_expired events
//   if event.Type == "kyc_expired" {
//       address := event.GetAttribute("address")
//       // Send re-verification notification to user
//       // Flag account in monitoring system
//       // Trigger compliance review
//   }
func (k *Keeper) BeginBlocker(ctx sdk.Context) {
	// Get current block time for expiry checks
	currentTime := ctx.BlockTime()

	// Batch limit: process max 100 expired records per block
	// This prevents long blocks even with 100K+ total records
	const maxExpiredPerBlock = 100

	// Track processed records for event emission and index cleanup
	// Pre-allocate based on max processed per block to avoid reallocations
	expiredRecords := make([]*types.KYCRecord, 0, maxExpiredPerBlock)

	// Use expiration index for efficient O(k) lookup
	// Only iterates over expired records, not all records
	// Note: IterateExpiredRecords automatically cleans up stale index entries
	processedCount := k.IterateExpiredRecords(ctx, currentTime, maxExpiredPerBlock, func(address string) bool {
		// Get full record to emit complete event
		// Note: IterateExpiredRecords already verified record exists
		record, err := k.GetKYCRecord(ctx, address)
		if err != nil {
			// This shouldn't happen since IterateExpiredRecords verifies existence
			k.logger(ctx).Error(
				"unexpected: expired record not found in store",
				"address", address,
				"error", err,
			)
			return false
		}

		// Double-check expiration (defensive programming)
		if record.ExpiresAt == nil || !currentTime.After(*record.ExpiresAt) {
			// Not actually expired - index may be stale
			return false
		}

		// Track for event emission and cleanup
		expiredRecords = append(expiredRecords, record)

		return false // Continue processing up to batch limit
	})

	// Emit events and cleanup index for all expired records
	for _, record := range expiredRecords {
		// Emit event for expired KYC record
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeKYCExpired,
				sdk.NewAttribute(types.AttributeKeyAddress, record.Address),
				sdk.NewAttribute("expired_at", record.ExpiresAt.Format("2006-01-02 15:04:05 MST")),
				sdk.NewAttribute("kyc_level", record.KycLevel.String()),
				sdk.NewAttribute("provider", record.Provider),
				sdk.NewAttribute("jurisdiction", record.Jurisdiction),
				sdk.NewAttribute("verified_at", record.VerifiedAt.Format("2006-01-02 15:04:05 MST")),
				sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
				sdk.NewAttribute(types.AttributeKeyBlockTime, currentTime.Format("2006-01-02 15:04:05 MST")),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", currentTime.Unix())),
			),
		)

		// Log expiry for node operators
		k.logger(ctx).Info(
			"KYC record expired",
			"address", record.Address,
			"expired_at", record.ExpiresAt.Format("2006-01-02 15:04:05"),
			"kyc_level", record.KycLevel.String(),
			"provider", record.Provider,
		)

		// Remove from expiration index to prevent re-processing
		// This ensures we don't emit duplicate events
		k.RemoveFromExpirationIndex(ctx, record)
	}

	// Note: Stale index entries (records deleted but index remaining)
	// are automatically cleaned up by IterateExpiredRecords

	// Log summary if any records processed
	if processedCount > 0 {
		k.logger(ctx).Info(
			"BeginBlocker: processed expired KYC records",
			"count", processedCount,
			"block_height", ctx.BlockHeight(),
			"block_time", currentTime.Format("2006-01-02 15:04:05"),
		)
	}
}
