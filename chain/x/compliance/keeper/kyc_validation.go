package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// IsKYCExpired checks if a KYC record has expired based on the expires_at timestamp.
// Returns true if:
//   - No KYC record is found (treat as expired)
//   - The current block time is after the expires_at timestamp
//
// This method is the primary expiry check for KYC verification status.
//
// Parameters:
//   - ctx: SDK context for accessing block time and state
//   - address: Blockchain address to check
//
// Returns:
//   - bool: true if KYC is expired or not found, false if valid
//
// Security considerations:
//   - Default deny: No KYC found = expired (fail-safe)
//   - Time-based: Uses blockchain time (cannot be manipulated by users)
//   - Immutable timestamps: expires_at is set at KYC submission time
//
// Compliance:
//   - FinCEN: KYC verification must be current and periodically refreshed
//   - FATF Recommendation 10: Ongoing due diligence requires fresh KYC
//   - Prevents stale verification from being used indefinitely
//
// Example usage:
//   if k.IsKYCExpired(ctx, userAddress) {
//       return errorsmod.Wrap(types.ErrKYCExpired, "re-verification required")
//   }
func (k Keeper) IsKYCExpired(ctx sdk.Context, address string) bool {
	record, err := k.GetKYCRecord(ctx, address)
	if err != nil {
		// No KYC record = treat as expired (fail-safe default)
		return true
	}

	// Check if current time is after expiry time
	// ExpiresAt is *time.Time (nullable), dereference if not nil
	if record.ExpiresAt != nil {
		return ctx.BlockTime().After(*record.ExpiresAt)
	}
	// No expiry time set = never expires
	return false
}

// ValidateKYCStatus performs comprehensive KYC validation for compliance operations.
// This is the primary enforcement function that should be called before any operation
// that requires valid KYC verification.
//
// Validation checks (in order):
//   1. KYC record exists
//   2. KYC has not expired (expires_at check)
//   3. KYC level meets minimum requirement from params
//
// This method MUST be called before:
//   - High-value transactions
//   - Bridge operations
//   - DEX trading (if KYC required)
//   - Identity credential issuance
//   - Any operation requiring verified identity
//
// Parameters:
//   - ctx: SDK context for state access and time
//   - address: Blockchain address to validate
//
// Returns:
//   - error: Specific error indicating validation failure, nil if valid
//
// Error types:
//   - ErrKYCNotFound: No KYC record exists for address
//   - ErrKYCExpired: KYC record has expired (re-verification needed)
//   - ErrInsufficientKYCLevel: KYC level below minimum required
//
// Compliance:
//   - FinCEN: Customer identification program (CIP) enforcement
//   - FATF Recommendation 10: Customer due diligence verification
//   - BSA: Know Your Customer requirements
//   - AML: Identity verification for regulated operations
//
// Security considerations:
//   - Default deny: Returns error if any check fails
//   - Immutable audit: All checks logged via state access
//   - Cannot be bypassed: Centralized enforcement point
//   - Time-based: Automatic expiry prevents stale verification
//
// Example usage:
//   if err := k.ValidateKYCStatus(ctx, sender); err != nil {
//       return errorsmod.Wrap(err, "sender KYC validation failed")
//   }
func (k Keeper) ValidateKYCStatus(ctx sdk.Context, address string) error {
	// Check if KYC record exists
	record, err := k.GetKYCRecord(ctx, address)
	if err != nil {
		return errorsmod.Wrapf(types.ErrKYCNotFound,
			"no KYC record found for address %s", address)
	}

	// Check if KYC has expired
	if k.IsKYCExpired(ctx, address) {
		expiryTimeStr := "never"
		if record.ExpiresAt != nil {
			expiryTimeStr = record.ExpiresAt.Format("2006-01-02 15:04:05")
		}
		return errorsmod.Wrapf(types.ErrKYCExpired,
			"KYC expired for address %s (expired at %s, current time %s)",
			address,
			expiryTimeStr,
			ctx.BlockTime().Format("2006-01-02 15:04:05"))
	}

	// Check if KYC level meets minimum requirement
	params := k.GetParams(ctx)
	if record.KycLevel < params.MinimumKycLevel {
		return errorsmod.Wrapf(types.ErrInsufficientKYCLevel,
			"KYC level %s is below minimum required level %s for address %s",
			record.KycLevel.String(),
			params.MinimumKycLevel.String(),
			address)
	}

	return nil
}

// ValidateKYCForOperation validates KYC status with context about the operation.
// This is a convenience wrapper around ValidateKYCStatus that includes operation
// context in error messages for better debugging and audit trails.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address to validate
//   - operation: Description of operation requiring KYC (e.g., "bridge transfer", "DEX swap")
//
// Returns:
//   - error: Specific error with operation context, nil if valid
//
// Example usage:
//   if err := k.ValidateKYCForOperation(ctx, sender, "bridge transfer to Ethereum"); err != nil {
//       return err
//   }
func (k Keeper) ValidateKYCForOperation(ctx sdk.Context, address string, operation string) error {
	if err := k.ValidateKYCStatus(ctx, address); err != nil {
		return errorsmod.Wrapf(err, "KYC validation failed for operation: %s", operation)
	}
	return nil
}

// IterateKYCRecords iterates over all KYC records in the store.
// The callback function receives each KYC record and should return:
//   - false to continue iteration
//   - true to stop iteration
//
// This method is used for:
//   - BeginBlocker processing (marking expired records)
//   - Batch queries
//   - Compliance audits
//   - Statistics collection
//
// Parameters:
//   - ctx: SDK context for state access
//   - callback: Function called for each record
//
// Security considerations:
//   - Read-only iteration (callback should not modify records directly)
//   - Deterministic ordering (iterates in key order)
//   - Gas-bounded: Caller responsible for pagination in production
//
// Example usage:
//   k.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
//       if k.IsKYCExpired(ctx, record.Address) {
//           // Process expired record
//       }
//       return false // Continue iteration
//   })
func (k Keeper) IterateKYCRecords(ctx sdk.Context, callback func(record types.KYCRecord) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, KYCRecordsKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.KYCRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			// Log error but continue iteration
			k.logger(ctx).Error(
				"failed to unmarshal KYC record during iteration",
				"error", err,
				"key", fmt.Sprintf("%x", iterator.Key()),
			)
			continue
		}

		if callback(record) {
			break
		}
	}
}

// GetExpiringKYCRecords returns all KYC records that will expire within the specified duration.
// This is useful for:
//   - Sending expiry warnings to users
//   - Proactive re-verification reminders
//   - Compliance reporting
//
// Parameters:
//   - ctx: SDK context for state access
//   - withinDuration: Time window to check (e.g., 30 days)
//
// Returns:
//   - []*types.KYCRecord: List of records expiring within the duration
//
// Example usage:
//   // Get records expiring in next 30 days
//   expiring := k.GetExpiringKYCRecords(ctx, 30*24*time.Hour)
//   for _, record := range expiring {
//       // Send notification to user
//   }
func (k Keeper) GetExpiringKYCRecords(ctx sdk.Context, withinDuration time.Duration) []*types.KYCRecord {
	expiringRecords := make([]*types.KYCRecord, 0, 64)
	expiryThreshold := ctx.BlockTime().Add(withinDuration)

	k.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
		// ExpiresAt is *time.Time (nullable), check for nil
		if record.ExpiresAt != nil {
			expiresAt := *record.ExpiresAt
			// Record expires after now but before threshold
			if expiresAt.After(ctx.BlockTime()) && expiresAt.Before(expiryThreshold) {
				expiringRecords = append(expiringRecords, &record)
			}
		}
		return false // Continue iteration
	})

	return expiringRecords
}

// GetExpiredKYCRecords returns all KYC records that have already expired.
// This is useful for:
//   - BeginBlocker processing
//   - Compliance reporting
//   - Identifying accounts requiring re-verification
//
// Parameters:
//   - ctx: SDK context for state access
//
// Returns:
//   - []*types.KYCRecord: List of expired records
//
// Security considerations:
//   - Read-only: Does not modify state
//   - Time-based: Uses blockchain time (trustworthy)
//   - Gas cost: O(n) where n is total KYC records (paginate in production)
//
// Example usage:
//   expired := k.GetExpiredKYCRecords(ctx)
//   for _, record := range expired {
//       // Emit event or take action
//   }
func (k Keeper) GetExpiredKYCRecords(ctx sdk.Context) []*types.KYCRecord {
	expiredRecords := make([]*types.KYCRecord, 0, 64)

	k.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
		// ExpiresAt is *time.Time (nullable), check for nil and dereference
		if record.ExpiresAt != nil && ctx.BlockTime().After(*record.ExpiresAt) {
			expiredRecords = append(expiredRecords, &record)
		}
		return false // Continue iteration
	})

	return expiredRecords
}
