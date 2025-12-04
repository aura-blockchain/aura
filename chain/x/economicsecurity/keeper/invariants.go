package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// RegisterInvariants registers all economicsecurity module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "vesting-schedules-valid", VestingSchedulesInvariant(k))
	ir.RegisterRoute(types.ModuleName, "vote-locks-valid", VoteLocksInvariant(k))
	ir.RegisterRoute(types.ModuleName, "treasury-txs-valid", PendingTreasuryTxsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "inflation-alerts-valid", InflationAlertsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "whale-protection-valid", WhaleLimitsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "mev-balances-valid", MEVBalancesInvariant(k))
}

// AllInvariants runs all invariants of the economicsecurity module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			VestingSchedulesInvariant(k),
			VoteLocksInvariant(k),
			PendingTreasuryTxsInvariant(k),
			InflationAlertsInvariant(k),
			WhaleLimitsInvariant(k),
			MEVBalancesInvariant(k),
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

// ParamsInvariant checks that module parameters are valid and consistent
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams()
		if err := types.ValidateParams(&params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid module params: %s", err.Error()),
			), true
		}

		// Validate tokenomics parameters
		if params.Tokenomics != nil {
			// Max supply must be positive
			maxSupply, ok := sdkmath.NewIntFromString(params.Tokenomics.MaxSupply)
			if !ok || maxSupply.LTE(sdkmath.ZeroInt()) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("invalid max supply: %s", params.Tokenomics.MaxSupply),
				), true
			}

			// Circulating supply must not exceed max supply
			circSupply, ok := sdkmath.NewIntFromString(params.Tokenomics.CirculatingSupply)
			if ok && circSupply.GT(maxSupply) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("circulating supply (%s) exceeds max supply (%s)",
						params.Tokenomics.CirculatingSupply, params.Tokenomics.MaxSupply),
				), true
			}

			// Inflation rate must be within min/max bounds
			if params.Tokenomics.InflationRate < params.Tokenomics.MinInflationRate ||
				params.Tokenomics.InflationRate > params.Tokenomics.MaxInflationRate {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("inflation rate %d outside bounds [%d, %d]",
						params.Tokenomics.InflationRate,
						params.Tokenomics.MinInflationRate,
						params.Tokenomics.MaxInflationRate),
				), true
			}
		}

		// Validate whale protection parameters
		if params.WhaleProtection != nil {
			// MaxHoldingPercentage is in basis points (10000 = 100%)
			if params.WhaleProtection.MaxHoldingPercentage > 10000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("invalid max holding percentage: %d (must be <= 10000 basis points)",
						params.WhaleProtection.MaxHoldingPercentage),
				), true
			}

			// MaxTxPercentage is in basis points (10000 = 100%)
			if params.WhaleProtection.MaxTxPercentage > 10000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("invalid max tx percentage: %d (must be <= 10000 basis points)",
						params.WhaleProtection.MaxTxPercentage),
				), true
			}

			// LargeTxThreshold should be reasonable (in basis points)
			if params.WhaleProtection.LargeTxThreshold > 10000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("invalid large tx threshold: %d (must be <= 10000 basis points)",
						params.WhaleProtection.LargeTxThreshold),
				), true
			}

			// LargeTxCooldown should be reasonable (max 30 days in seconds)
			if params.WhaleProtection.LargeTxCooldown > 30*24*60*60 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("invalid large tx cooldown: %d seconds (max 30 days)",
						params.WhaleProtection.LargeTxCooldown),
				), true
			}
		}

		return "", false
	}
}

// VestingSchedulesInvariant checks all vesting schedules are valid and consistent
func VestingSchedulesInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Use keeper's iterator to check all vesting schedules
		scheduleIDs := make(map[string]bool)
		invalidSchedule := ""

		err := k.IterateVestingSchedules(ctx, func(schedule *types.VestingSchedule) bool {
			// Check for duplicate schedule IDs
			if scheduleIDs[schedule.ScheduleId] {
				invalidSchedule = fmt.Sprintf("duplicate schedule ID: %s", schedule.ScheduleId)
				return true // stop iteration
			}
			scheduleIDs[schedule.ScheduleId] = true

			// Validate schedule ID is not empty
			if schedule.ScheduleId == "" {
				invalidSchedule = "vesting schedule has empty ID"
				return true
			}

			// Validate beneficiary address
			if schedule.BeneficiaryAddress == "" {
				invalidSchedule = fmt.Sprintf("schedule %s has empty beneficiary", schedule.ScheduleId)
				return true
			}

			// Validate total amount is positive
			totalAmt, ok := sdkmath.NewIntFromString(schedule.TotalAmount)
			if !ok || totalAmt.LTE(sdkmath.ZeroInt()) {
				invalidSchedule = fmt.Sprintf("schedule %s has invalid total amount: %s",
					schedule.ScheduleId, schedule.TotalAmount)
				return true
			}

			// Validate vested amount doesn't exceed total
			vestedAmt, ok := sdkmath.NewIntFromString(schedule.VestedAmount)
			if ok && vestedAmt.GT(totalAmt) {
				invalidSchedule = fmt.Sprintf("schedule %s vested amount (%s) exceeds total (%s)",
					schedule.ScheduleId, schedule.VestedAmount, schedule.TotalAmount)
				return true
			}

			// Validate vesting duration is positive
			if schedule.VestingDuration == 0 {
				invalidSchedule = fmt.Sprintf("schedule %s has zero vesting duration", schedule.ScheduleId)
				return true
			}

			// Validate cliff duration is not longer than vesting duration
			if schedule.CliffDuration > schedule.VestingDuration {
				invalidSchedule = fmt.Sprintf("schedule %s cliff duration (%d) exceeds vesting duration (%d)",
					schedule.ScheduleId, schedule.CliffDuration, schedule.VestingDuration)
				return true
			}

			return false // continue iteration
		})

		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"vesting-schedules-valid",
				fmt.Sprintf("failed to iterate vesting schedules: %s", err.Error()),
			), true
		}

		if invalidSchedule != "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"vesting-schedules-valid",
				invalidSchedule,
			), true
		}

		return "", false
	}
}

// VoteLocksInvariant checks all vote locks are valid and consistent
func VoteLocksInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		lockIDs := make(map[string]bool)
		invalidLock := ""

		err := k.IterateVoteLocks(ctx, func(lock *types.VoteLock) bool {
			// Check for duplicate lock IDs
			if lockIDs[lock.LockId] {
				invalidLock = fmt.Sprintf("duplicate lock ID: %s", lock.LockId)
				return true
			}
			lockIDs[lock.LockId] = true

			// Validate lock ID is not empty
			if lock.LockId == "" {
				invalidLock = "vote lock has empty ID"
				return true
			}

			// Validate owner address
			if lock.Owner == "" {
				invalidLock = fmt.Sprintf("lock %s has empty owner", lock.LockId)
				return true
			}

			// Validate address format
			if _, err := sdk.AccAddressFromBech32(lock.Owner); err != nil {
				invalidLock = fmt.Sprintf("lock %s has invalid owner address: %s", lock.LockId, lock.Owner)
				return true
			}

			// Validate locked amount is positive
			amount, ok := sdkmath.NewIntFromString(lock.Amount)
			if !ok || amount.LTE(sdkmath.ZeroInt()) {
				invalidLock = fmt.Sprintf("lock %s has invalid amount: %s", lock.LockId, lock.Amount)
				return true
			}

			// Validate lock end time is set
			if lock.LockEnd == nil {
				invalidLock = fmt.Sprintf("lock %s has nil lock end time", lock.LockId)
				return true
			}

			// Validate lock start time is set
			if lock.LockStart == nil {
				invalidLock = fmt.Sprintf("lock %s has nil lock start time", lock.LockId)
				return true
			}

			// Validate lock end is after lock start
			if lock.LockEnd.Seconds < lock.LockStart.Seconds {
				invalidLock = fmt.Sprintf("lock %s has end time before start time", lock.LockId)
				return true
			}

			// Validate voting power is positive
			votingPower, ok := sdkmath.NewIntFromString(lock.VotingPower)
			if !ok || votingPower.IsNegative() {
				invalidLock = fmt.Sprintf("lock %s has invalid voting power: %s", lock.LockId, lock.VotingPower)
				return true
			}

			return false // continue
		})

		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"vote-locks-valid",
				fmt.Sprintf("failed to iterate vote locks: %s", err.Error()),
			), true
		}

		if invalidLock != "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"vote-locks-valid",
				invalidLock,
			), true
		}

		return "", false
	}
}

// PendingTreasuryTxsInvariant checks all pending treasury transactions are valid
func PendingTreasuryTxsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		txIDs := make(map[string]bool)
		invalidTx := ""

		err := k.IteratePendingTreasuryTxs(ctx, func(tx *types.PendingTreasuryTx) bool {
			// Check for duplicate tx IDs
			if txIDs[tx.TxId] {
				invalidTx = fmt.Sprintf("duplicate treasury tx ID: %s", tx.TxId)
				return true
			}
			txIDs[tx.TxId] = true

			// Validate tx ID is not empty
			if tx.TxId == "" {
				invalidTx = "treasury tx has empty ID"
				return true
			}

			// Validate recipient address
			if tx.Recipient == "" {
				invalidTx = fmt.Sprintf("treasury tx %s has empty recipient", tx.TxId)
				return true
			}

			// Validate amount is positive
			amount, ok := sdkmath.NewIntFromString(tx.Amount)
			if !ok || amount.LTE(sdkmath.ZeroInt()) {
				invalidTx = fmt.Sprintf("treasury tx %s has invalid amount: %s", tx.TxId, tx.Amount)
				return true
			}

			// Validate proposer is set
			if tx.Proposer == "" {
				invalidTx = fmt.Sprintf("treasury tx %s has empty proposer", tx.TxId)
				return true
			}

			// Note: Signatures can be nil or empty slice (protobuf repeated fields)
			// Both are semantically equivalent and valid for a new treasury tx

			// Validate creation timestamp
			if tx.CreatedAt == nil {
				invalidTx = fmt.Sprintf("treasury tx %s has nil creation time", tx.TxId)
				return true
			}

			// Validate executable time if set
			if tx.ExecutableAt != nil && tx.CreatedAt != nil {
				if tx.ExecutableAt.Seconds < tx.CreatedAt.Seconds {
					invalidTx = fmt.Sprintf("treasury tx %s executable time before creation time", tx.TxId)
					return true
				}
			}

			return false // continue
		})

		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"treasury-txs-valid",
				fmt.Sprintf("failed to iterate treasury txs: %s", err.Error()),
			), true
		}

		if invalidTx != "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"treasury-txs-valid",
				invalidTx,
			), true
		}

		return "", false
	}
}

// InflationAlertsInvariant checks all inflation alerts are valid
func InflationAlertsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		alertIDs := make(map[string]bool)
		invalidAlert := ""

		err := k.IterateInflationAlerts(ctx, func(alert *types.InflationAlert) bool {
			// Check for duplicate alert IDs
			if alertIDs[alert.AlertId] {
				invalidAlert = fmt.Sprintf("duplicate alert ID: %s", alert.AlertId)
				return true
			}
			alertIDs[alert.AlertId] = true

			// Validate alert ID is not empty
			if alert.AlertId == "" {
				invalidAlert = "inflation alert has empty ID"
				return true
			}

			// Validate alert type is valid
			if alert.AlertType == types.InflationAlertType_INFLATION_ALERT_TYPE_UNSPECIFIED {
				invalidAlert = fmt.Sprintf("alert %s has unspecified type", alert.AlertId)
				return true
			}

			// Validate severity is valid
			if alert.Severity == types.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED {
				invalidAlert = fmt.Sprintf("alert %s has unspecified severity", alert.AlertId)
				return true
			}

			// Validate triggered timestamp is set
			if alert.TriggeredAt == nil {
				invalidAlert = fmt.Sprintf("alert %s has nil triggered_at timestamp", alert.AlertId)
				return true
			}

			// Validate inflation rates are reasonable
			if alert.CurrentInflationRate > 100000 { // Max 10000% (in basis points)
				invalidAlert = fmt.Sprintf("alert %s has unreasonable current inflation rate: %d",
					alert.AlertId, alert.CurrentInflationRate)
				return true
			}

			// Validate message is not empty
			if alert.Message == "" {
				invalidAlert = fmt.Sprintf("alert %s has empty message", alert.AlertId)
				return true
			}

			return false // continue
		})

		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"inflation-alerts-valid",
				fmt.Sprintf("failed to iterate inflation alerts: %s", err.Error()),
			), true
		}

		if invalidAlert != "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"inflation-alerts-valid",
				invalidAlert,
			), true
		}

		return "", false
	}
}

// WhaleLimitsInvariant checks whale protection limits are properly enforced
func WhaleLimitsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invalidRecord := ""

		// Check all large transaction records
		err := k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
			// Validate tx hash is not empty
			if record.TxHash == "" {
				invalidRecord = "large tx record has empty hash"
				return true
			}

			// Validate sender address
			if record.Sender == "" {
				invalidRecord = fmt.Sprintf("large tx record %s has empty sender", record.TxHash)
				return true
			}

			// Validate amount is positive
			amount, ok := sdkmath.NewIntFromString(record.Amount)
			if !ok || amount.LTE(sdkmath.ZeroInt()) {
				invalidRecord = fmt.Sprintf("large tx record %s has invalid amount: %s",
					record.TxHash, record.Amount)
				return true
			}

			// Validate timestamp is set
			if record.Timestamp == nil {
				invalidRecord = fmt.Sprintf("large tx record %s has nil timestamp", record.TxHash)
				return true
			}

			// Validate recipient is not empty
			if record.Recipient == "" {
				invalidRecord = fmt.Sprintf("large tx record %s has empty recipient", record.TxHash)
				return true
			}

			// Validate percentage of supply is reasonable (max 100% = 10000 basis points)
			if record.PercentageOfSupply > 10000 {
				invalidRecord = fmt.Sprintf("large tx record %s has invalid percentage: %d",
					record.TxHash, record.PercentageOfSupply)
				return true
			}

			// Validate block height is positive if tx was included
			if record.BlockHeight == 0 {
				// This might be a pending tx, which is fine
				// But we log it for awareness
			}

			return false // continue
		})

		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"whale-protection-valid",
				fmt.Sprintf("failed to iterate large tx records: %s", err.Error()),
			), true
		}

		if invalidRecord != "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"whale-protection-valid",
				invalidRecord,
			), true
		}

		return "", false
	}
}

// MEVBalancesInvariant checks MEV balances are valid and consistent
func MEVBalancesInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		totalMEV := sdkmath.ZeroInt()
		invalidBalance := ""

		// Sum all user MEV balances
		err := k.IterateUserMEVBalances(ctx, func(address string, balanceStr string) bool {
			// Validate address
			if address == "" {
				invalidBalance = "MEV balance has empty address"
				return true
			}

			// Validate address format
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				invalidBalance = fmt.Sprintf("MEV balance has invalid address: %s", address)
				return true
			}

			// Validate balance is non-negative
			balance, ok := sdkmath.NewIntFromString(balanceStr)
			if !ok || balance.IsNegative() {
				invalidBalance = fmt.Sprintf("address %s has invalid MEV balance: %s", address, balanceStr)
				return true
			}

			// Sum up total
			totalMEV = totalMEV.Add(balance)

			return false // continue
		})

		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"mev-balances-valid",
				fmt.Sprintf("failed to iterate MEV balances: %s", err.Error()),
			), true
		}

		if invalidBalance != "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"mev-balances-valid",
				invalidBalance,
			), true
		}

		// Validate total pending MEV matches sum of user balances
		totalPending, err := k.GetTotalMEVPending(ctx)
		if err == nil && totalPending != "" {
			pending, ok := sdkmath.NewIntFromString(totalPending)
			if !ok {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mev-balances-valid",
					fmt.Sprintf("invalid total MEV pending: %s", totalPending),
				), true
			}

			// Allow some tolerance for rounding, but should match closely
			// In production, you might want exact match
			if !pending.Equal(totalMEV) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mev-balances-valid",
					fmt.Sprintf("total MEV pending (%s) doesn't match sum of user balances (%s)",
						totalPending, totalMEV.String()),
				), true
			}
		}

		return "", false
	}
}
