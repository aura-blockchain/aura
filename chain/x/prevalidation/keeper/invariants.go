package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// RegisterInvariants registers all prevalidation module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "mempool-consistency", MempoolConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "nonce-validity", NonceValidityInvariant(k))
}

// AllInvariants runs all invariants of the prevalidation module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			MempoolConsistencyInvariant(k),
			NonceValidityInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}

		if params == nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"params are nil",
			), true
		}

		// Validate max mempool size
		if params.MaxMempoolSize > 0 && params.MaxMempoolSize < 100 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("max mempool size too small: %d (min 100)", params.MaxMempoolSize),
			), true
		}

		// Validate max transaction age (should be at least 1 hour if set)
		if params.MaxTransactionAge > 0 && params.MaxTransactionAge < 3600 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("max transaction age too short: %d seconds (min 1 hour)", params.MaxTransactionAge),
			), true
		}

		// Validate batch size configuration
		if params.BatchConfig != nil {
			if params.BatchConfig.MaxBatchSize == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"batch config max batch size cannot be zero",
				), true
			}

			if params.BatchConfig.MaxBatchSize > 10000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("batch config max batch size too large: %d (max 10000)", params.BatchConfig.MaxBatchSize),
				), true
			}

			if params.BatchConfig.BatchTimeoutMs == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"batch config timeout cannot be zero",
				), true
			}

			if params.BatchConfig.BatchTimeoutMs > 60000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("batch config timeout too long: %d ms (max 60000)", params.BatchConfig.BatchTimeoutMs),
				), true
			}
		}

		// Validate scheduler configuration
		if params.SchedulerConfig != nil && params.SchedulerConfig.Enabled {
			// Off-peak hours should be valid (0-23)
			if params.SchedulerConfig.OffPeakStartHour > 23 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("off-peak start hour invalid: %d (max 23)", params.SchedulerConfig.OffPeakStartHour),
				), true
			}

			if params.SchedulerConfig.OffPeakEndHour > 23 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("off-peak end hour invalid: %d (max 23)", params.SchedulerConfig.OffPeakEndHour),
				), true
			}

			// Auto-scaling thresholds should be reasonable (0-100%)
			if params.SchedulerConfig.AutoScaleThreshold > 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("auto-scale threshold invalid: %d (max 100)", params.SchedulerConfig.AutoScaleThreshold),
				), true
			}
		}

		return "", false
	}
}

// MempoolConsistencyInvariant checks mempool state consistency
func MempoolConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		transactions := k.GetMempoolTransactions(ctx)

		// Get max mempool size from params
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"mempool-consistency",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}

		// Check mempool size doesn't exceed max
		if params != nil && params.MaxMempoolSize > 0 {
			if uint64(len(transactions)) > params.MaxMempoolSize {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mempool-consistency",
					fmt.Sprintf("mempool size (%d) exceeds max (%d)",
						len(transactions), params.MaxMempoolSize),
				), true
			}
		}

		// Track seen transaction hashes to detect duplicates
		seenHashes := make(map[string]bool)

		for _, tx := range transactions {
			// Basic validation of transaction fields
			if tx.Sender == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mempool-consistency",
					"mempool contains transaction with empty sender",
				), true
			}

			if tx.Recipient == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mempool-consistency",
					"mempool contains transaction with empty recipient",
				), true
			}

			if tx.Amount == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mempool-consistency",
					"mempool contains transaction with empty amount",
				), true
			}

			// Check for duplicate transactions in mempool
			txHash := k.GetTransactionHash(tx)
			if seenHashes[txHash] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mempool-consistency",
					fmt.Sprintf("duplicate transaction in mempool: %s", txHash),
				), true
			}
			seenHashes[txHash] = true

			// Validate nonce is reasonable
			currentNonce := k.GetNonce(ctx, tx.Sender)
			if tx.Nonce < currentNonce {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mempool-consistency",
					fmt.Sprintf("transaction in mempool with stale nonce: sender=%s, tx_nonce=%d, current_nonce=%d",
						tx.Sender, tx.Nonce, currentNonce),
				), true
			}

			// Nonce shouldn't be too far in the future (indicates potential issue)
			if tx.Nonce > currentNonce+1000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mempool-consistency",
					fmt.Sprintf("transaction in mempool with nonce too far in future: sender=%s, tx_nonce=%d, current_nonce=%d",
						tx.Sender, tx.Nonce, currentNonce),
				), true
			}
		}

		return "", false
	}
}

// NonceValidityInvariant checks nonce values are reasonable
func NonceValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Get all transactions to check nonces
		transactions := k.GetMempoolTransactions(ctx)

		// Track nonces per address
		addressNonces := make(map[string]uint64)

		for _, tx := range transactions {
			// Get current nonce for sender
			currentNonce := k.GetNonce(ctx, tx.Sender)

			// Track highest nonce seen for this address
			if highestNonce, exists := addressNonces[tx.Sender]; exists {
				if tx.Nonce > highestNonce {
					addressNonces[tx.Sender] = tx.Nonce
				}
			} else {
				addressNonces[tx.Sender] = tx.Nonce
			}

			// Nonce should not be negative (uint64 can't be, but check for wraparound)
			if tx.Nonce == ^uint64(0) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"nonce-validity",
					fmt.Sprintf("transaction has maximum uint64 nonce (potential wraparound): %s", tx.Sender),
				), true
			}

			// Validate transaction is valid according to current nonce
			valid, err := k.ValidateTransaction(ctx, tx)
			if err != nil && !valid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"nonce-validity",
					fmt.Sprintf("invalid transaction in mempool: sender=%s, error=%s", tx.Sender, err.Error()),
				), true
			}
		}

		// Check for gaps in nonce sequences that are too large
		for address, highestNonce := range addressNonces {
			currentNonce := k.GetNonce(ctx, address)

			// If there's a large gap, it might indicate a problem
			// Allow gaps up to 100 transactions
			if highestNonce > currentNonce+100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"nonce-validity",
					fmt.Sprintf("large nonce gap for address %s: current=%d, highest_pending=%d",
						address, currentNonce, highestNonce),
				), true
			}
		}

		return "", false
	}
}
