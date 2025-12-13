package keeper

//lint:file-ignore SA1019 -- invariants use deprecated registry interfaces until Cosmos SDK removal

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// RegisterInvariants registers all monitoring module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "params", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "alerts", AlertsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "validator-uptime", ValidatorUptimeInvariant(k))
}

// AllInvariants runs all monitoring module invariants
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Check params invariant
		msg, broken := ParamsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		// Check alerts invariant
		msg, broken = AlertsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		// Check validator uptime invariant
		msg, broken = ValidatorUptimeInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		return "", false
	}
}

// ParamsInvariant checks that params are valid
func ParamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "params",
				fmt.Sprintf("failed to get params: %v", err),
			), true
		}

		if err := types.ValidateParams(params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "params",
				fmt.Sprintf("invalid params: %v", err),
			), true
		}

		return "", false
	}
}

// AlertsInvariant checks that all alerts have valid data
func AlertsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		alerts, err := k.GetAllAlerts(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "alerts",
				fmt.Sprintf("failed to get alerts: %v", err),
			), true
		}

		for _, alert := range alerts {
			if alert.ID == "" {
				return sdk.FormatInvariant(
					types.ModuleName, "alerts",
					"alert with empty ID found",
				), true
			}

			if alert.Timestamp.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName, "alerts",
					fmt.Sprintf("alert %s has zero timestamp", alert.ID),
				), true
			}

			if alert.Resolved && alert.ResolvedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName, "alerts",
					fmt.Sprintf("alert %s is resolved but has zero resolved timestamp", alert.ID),
				), true
			}
		}

		return "", false
	}
}

// ValidatorUptimeInvariant checks that validator uptime data is valid
func ValidatorUptimeInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		err := k.IterateValidatorUptimes(ctx, func(uptime *types.ValidatorUptime) bool {
			if uptime.ValidatorAddress == "" {
				return true
			}

			if uptime.SignedBlocks > uptime.TotalBlocks {
				return true
			}

			if uptime.MissedBlocks > uptime.TotalBlocks {
				return true
			}

			if uptime.TotalBlocks < 0 || uptime.SignedBlocks < 0 || uptime.MissedBlocks < 0 {
				return true
			}

			return false
		})

		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "validator-uptime",
				fmt.Sprintf("invalid validator uptime data: %v", err),
			), true
		}

		return "", false
	}
}

// TransactionMonitorInvariant checks that transaction monitor data is valid
func TransactionMonitorInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		transactions, err := k.GetAllTransactions(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "transaction-monitor",
				fmt.Sprintf("failed to get transactions: %v", err),
			), true
		}

		for _, tx := range transactions {
			if tx.TxHash == "" {
				return sdk.FormatInvariant(
					types.ModuleName, "transaction-monitor",
					"transaction with empty hash found",
				), true
			}

			if tx.Timestamp.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName, "transaction-monitor",
					fmt.Sprintf("transaction %s has zero timestamp", tx.TxHash),
				), true
			}

			if tx.BlockHeight < 0 {
				return sdk.FormatInvariant(
					types.ModuleName, "transaction-monitor",
					fmt.Sprintf("transaction %s has negative block height: %d", tx.TxHash, tx.BlockHeight),
				), true
			}
		}

		return "", false
	}
}

// NetworkHealthInvariant checks that network health data is valid
func NetworkHealthInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		health, err := k.GetNetworkHealth(ctx)
		if err != nil {
			// Network health is optional, not an error if not found
			return "", false
		}

		if health.ActiveValidators > health.TotalValidators {
			return sdk.FormatInvariant(
				types.ModuleName, "network-health",
				fmt.Sprintf("active validators (%d) > total validators (%d)",
					health.ActiveValidators, health.TotalValidators),
			), true
		}

		if health.NetworkCongestion < 0 || health.NetworkCongestion > 1 {
			return sdk.FormatInvariant(
				types.ModuleName, "network-health",
				fmt.Sprintf("network congestion out of range [0,1]: %.2f", health.NetworkCongestion),
			), true
		}

		if health.ConsensusHealth < 0 || health.ConsensusHealth > 1 {
			return sdk.FormatInvariant(
				types.ModuleName, "network-health",
				fmt.Sprintf("consensus health out of range [0,1]: %.2f", health.ConsensusHealth),
			), true
		}

		return "", false
	}
}
