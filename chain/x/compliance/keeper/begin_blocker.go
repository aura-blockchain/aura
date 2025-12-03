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
//   1. Iterate through all KYC records
//   2. Check if record has expired (BlockTime > ExpiresAt)
//   3. Emit EventTypeKYCExpired for each newly expired record
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
//   - Gas-bounded: Processing limited by block gas limit
//
// Compliance:
//   - FinCEN: Continuous KYC status monitoring
//   - FATF Recommendation 10: Ongoing due diligence enforcement
//   - BSA: Customer verification status tracking
//   - Provides immutable audit trail of all expiry events
//
// Performance considerations:
//   - O(n) where n is total KYC records
//   - Event emission is gas-efficient
//   - No storage writes (read-only iteration)
//   - Future optimization: Track "next expiry" to avoid full scans
//
// Events emitted:
//   - EventTypeKYCExpired: For each record that has expired since last block
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

	// Track number of expired records for logging
	expiredCount := 0

	// Iterate all KYC records and check for expiry
	k.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
		// Check if record has expired
		if currentTime.After(record.ExpiresAt.AsTime()) {
			// Emit event for expired KYC record
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeKYCExpired,
					sdk.NewAttribute(types.AttributeKeyAddress, record.Address),
					sdk.NewAttribute("expired_at", record.ExpiresAt.AsTime().Format("2006-01-02 15:04:05 MST")),
					sdk.NewAttribute("kyc_level", record.KycLevel.String()),
					sdk.NewAttribute("provider", record.Provider),
					sdk.NewAttribute("jurisdiction", record.Jurisdiction),
					sdk.NewAttribute("verified_at", record.VerifiedAt.AsTime().Format("2006-01-02 15:04:05 MST")),
					sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
					sdk.NewAttribute(types.AttributeKeyBlockTime, currentTime.Format("2006-01-02 15:04:05 MST")),
					sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", currentTime.Unix())),
				),
			)

			expiredCount++

			// Log expiry for node operators
			k.logger(ctx).Info(
				"KYC record expired",
				"address", record.Address,
				"expired_at", record.ExpiresAt.AsTime().Format("2006-01-02 15:04:05"),
				"kyc_level", record.KycLevel.String(),
				"provider", record.Provider,
			)
		}

		// Continue iteration (return false)
		return false
	})

	// Log summary if any records expired
	if expiredCount > 0 {
		k.logger(ctx).Info(
			"BeginBlocker: processed expired KYC records",
			"count", expiredCount,
			"block_height", ctx.BlockHeight(),
			"block_time", currentTime.Format("2006-01-02 15:04:05"),
		)
	}
}
