package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/compliance/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all compliance module invariants.
// nolint:staticcheck // Uses deprecated SDK invariant interfaces required by module wiring.
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "kyc-record-consistency", KYCRecordConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sanctions-screening-validity", SanctionsScreeningInvariant(k))
	ir.RegisterRoute(types.ModuleName, "gdpr-data-integrity", GDPRDataIntegrityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "tax-record-consistency", TaxRecordConsistencyInvariant(k))
}

// AllInvariants runs all invariants of the compliance module
// nolint:staticcheck // Uses deprecated SDK invariant interfaces required by module wiring.
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			KYCRecordConsistencyInvariant(k),
			SanctionsScreeningInvariant(k),
			GDPRDataIntegrityInvariant(k),
			TaxRecordConsistencyInvariant(k),
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
		params := k.GetParams(ctx)
		if err := types.ValidateParams(params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid params: %s", err.Error()),
			), true
		}
		return "", false
	}
}

// KYCRecordConsistencyInvariant checks that all KYC records have valid state
func KYCRecordConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, KYCRecordsKeyPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var record types.KYCRecord
			if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"kyc-record-consistency",
					fmt.Sprintf("failed to unmarshal KYC record: %s", err.Error()),
				), true
			}

			// Check address is valid
			if _, err := sdk.AccAddressFromBech32(record.Address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"kyc-record-consistency",
					fmt.Sprintf("invalid address in KYC record: %s", record.Address),
				), true
			}

			// Check PII commitment is exactly 32 bytes (SHA-256)
			if len(record.PiiCommitment) != 32 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"kyc-record-consistency",
					fmt.Sprintf("KYC record for %s has invalid PII commitment length: %d (expected 32)", record.Address, len(record.PiiCommitment)),
				), true
			}

			// Check KYC level is valid (should be between 1-4 for actual KYC levels)
			validLevels := []types.KYCLevel{
				types.KYCLevel_KYC_LEVEL_NONE,
				types.KYCLevel_KYC_LEVEL_BASIC,
				types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
				types.KYCLevel_KYC_LEVEL_ADVANCED,
			}
			levelValid := false
			for _, vl := range validLevels {
				if record.KycLevel == vl {
					levelValid = true
					break
				}
			}
			if !levelValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"kyc-record-consistency",
					fmt.Sprintf("KYC record for %s has invalid KYC level: %d", record.Address, record.KycLevel),
				), true
			}

			// Check verified_at timestamp
			// VerifiedAt is time.Time (non-nullable), check for zero value
			if record.VerifiedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"kyc-record-consistency",
					fmt.Sprintf("KYC record for %s has zero verified_at", record.Address),
				), true
			}
		}

		return "", false
	}
}

// SanctionsScreeningInvariant checks that sanctions screening results are valid
func SanctionsScreeningInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, SanctionsResultsKeyPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var result types.SanctionsScreeningResult
			if err := k.cdc.Unmarshal(iterator.Value(), &result); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sanctions-screening-validity",
					fmt.Sprintf("failed to unmarshal sanctions screening result: %s", err.Error()),
				), true
			}

			// Check address is valid
			if _, err := sdk.AccAddressFromBech32(result.Address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sanctions-screening-validity",
					fmt.Sprintf("invalid address in sanctions screening: %s", result.Address),
				), true
			}

			// Check screening date
			// ScreenedAt is time.Time (non-nullable), check for zero value
			if result.ScreenedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sanctions-screening-validity",
					fmt.Sprintf("sanctions screening for %s has zero screened_at", result.Address),
				), true
			}

			// If status indicates a match, must have matches
			if result.Status == types.SanctionsStatus_SANCTIONS_MATCH && len(result.Matches) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sanctions-screening-validity",
					fmt.Sprintf("sanctions screening for %s has match status but no matches", result.Address),
				), true
			}
		}

		return "", false
	}
}

// GDPRDataIntegrityInvariant checks GDPR data handling integrity
func GDPRDataIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, GDPRRequestsKeyPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var request types.GDPRDataRequest
			if err := k.cdc.Unmarshal(iterator.Value(), &request); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"gdpr-data-integrity",
					fmt.Sprintf("failed to unmarshal GDPR request: %s", err.Error()),
				), true
			}

			// Check requester address is valid
			if _, err := sdk.AccAddressFromBech32(request.Address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"gdpr-data-integrity",
					fmt.Sprintf("invalid address in GDPR request: %s", request.Address),
				), true
			}

			// Check request type is valid
			validTypes := []string{"access", "deletion", "rectification", "portability"}
			typeValid := false
			for _, vt := range validTypes {
				if request.RequestType == vt {
					typeValid = true
					break
				}
			}
			if !typeValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"gdpr-data-integrity",
					fmt.Sprintf("GDPR request has invalid type: %s", request.RequestType),
				), true
			}

			// Check timestamps
			// RequestedAt is time.Time (non-nullable), check for zero value
			if request.RequestedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"gdpr-data-integrity",
					"GDPR request has zero requested_at",
				), true
			}

			// If completed, must have completed_at
			if request.Status == "completed" && request.CompletedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"gdpr-data-integrity",
					"GDPR request marked completed but has nil completed_at",
				), true
			}
		}

		return "", false
	}
}

// TaxRecordConsistencyInvariant checks tax record consistency
func TaxRecordConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, TaxReportsKeyPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			// TaxReports are stored as TaxReportList keyed by address
			var reportList types.TaxReportList
			if err := k.cdc.Unmarshal(iterator.Value(), &reportList); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"tax-record-consistency",
					fmt.Sprintf("failed to unmarshal tax report list: %s", err.Error()),
				), true
			}

			// Check each report in the list
			for _, report := range reportList.Reports {
				// Check address is valid
				if _, err := sdk.AccAddressFromBech32(report.Address); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"tax-record-consistency",
						fmt.Sprintf("invalid address in tax report: %s", report.Address),
					), true
				}

				// Check tax year is valid (string format, should be YYYY)
				if len(report.TaxYear) != 4 {
					return sdk.FormatInvariant(
						types.ModuleName,
						"tax-record-consistency",
						fmt.Sprintf("tax report has invalid year format: %s", report.TaxYear),
					), true
				}

				// Parse and validate tax year is in reasonable range (1970 to current year + 1)
				var year int
				if _, err := fmt.Sscanf(report.TaxYear, "%d", &year); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"tax-record-consistency",
						fmt.Sprintf("tax report has non-numeric year: %s", report.TaxYear),
					), true
				}

				// Tax year must be >= 1970 (Unix epoch) and <= current year + 1 (for future filing)
				currentYear := ctx.BlockTime().Year()
				if year < 1970 || year > currentYear+1 {
					return sdk.FormatInvariant(
						types.ModuleName,
						"tax-record-consistency",
						fmt.Sprintf("tax report for %s has invalid year %d (must be between 1970 and %d)", report.Address, year, currentYear+1),
					), true
				}

				// Check jurisdiction is not empty
				if report.Jurisdiction == "" {
					return sdk.FormatInvariant(
						types.ModuleName,
						"tax-record-consistency",
						fmt.Sprintf("tax report for %s has empty jurisdiction", report.Address),
					), true
				}
			}
		}

		return "", false
	}
}
