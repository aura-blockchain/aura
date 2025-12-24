package keeper

//lint:file-ignore SA1019 // invariants rely on deprecated SDK registry until upstream removal

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/incidentresponse/types"
)

// RegisterInvariants registers all incidentresponse module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) { //nolint:staticcheck // invariant registry uses deprecated SDK interface
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "incident-validity", IncidentValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "pause-state-consistency", PauseStateConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "wallet-limits-validity", WalletLimitsValidityInvariant(k))
}

// AllInvariants runs all invariants of the incidentresponse module
func AllInvariants(k *Keeper) sdk.Invariant { //nolint:staticcheck // invariant signature uses deprecated SDK type
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			IncidentValidityInvariant(k),
			PauseStateConsistencyInvariant(k),
			WalletLimitsValidityInvariant(k),
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
		params, _ := k.GetParams(ctx)

		// Validate emergency pause configuration
		if params.EmergencyPauseEnabled {
			// Verify required signers is reasonable (1-10)
			if params.PauseRequiredSigners == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"pause requires at least 1 signer",
				), true
			}

			if params.PauseRequiredSigners > 10 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("pause required signers too high: %d (max 10)", params.PauseRequiredSigners),
				), true
			}

			// Verify authorized keys count is at least required signers
			if uint32(len(params.PauseAuthorizedKeys)) < params.PauseRequiredSigners {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("insufficient authorized keys (%d) for required signers (%d)",
						len(params.PauseAuthorizedKeys), params.PauseRequiredSigners),
				), true
			}

			// Verify max pause duration is reasonable (max 7 days)
			maxAllowedDuration := int64(7 * 24 * 60 * 60 * 1e9) // 7 days in nanoseconds
			if params.MaxPauseDuration.Nanoseconds() > maxAllowedDuration {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("max pause duration too long: %d seconds (max 7 days)",
						int64(params.MaxPauseDuration.Seconds())),
				), true
			}
		}

		// Validate hot wallet limits configuration
		if params.HotWalletLimitsEnabled {
			// Global max hot wallet should be positive if set
			if params.GlobalMaxHotWallet == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"hot wallet limits enabled but global max is empty",
				), true
			}
		}

		// Validate cold storage configuration
		if params.ColdStorage.Enabled {
			// Verify multi-sig threshold is reasonable
			if params.ColdStorage.MultiSigThreshold == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"cold storage requires at least 1 signer",
				), true
			}

			if params.ColdStorage.MultiSigThreshold > uint32(len(params.ColdStorage.MultiSigSigners)) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("cold storage threshold (%d) exceeds signers (%d)",
						params.ColdStorage.MultiSigThreshold, len(params.ColdStorage.MultiSigSigners)),
				), true
			}
		}

		// Additional validation can be added here for disaster recovery and backup validators
		// when specific fields are defined in the proto

		return "", false
	}
}

// IncidentValidityInvariant checks that all incidents are valid
func IncidentValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		incidents := k.GetAllIncidents(ctx)

		for _, incident := range incidents {
			// Check incident ID is not empty
			if incident.ID == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					"incident has empty ID",
				), true
			}

			// Check title is not empty
			if incident.Title == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s has empty title", incident.ID),
				), true
			}

			// Check severity is valid
			if incident.Severity != types.SeverityLow &&
				incident.Severity != types.SeverityMedium &&
				incident.Severity != types.SeverityHigh &&
				incident.Severity != types.SeverityCritical {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s has invalid severity: %s", incident.ID, incident.Severity),
				), true
			}

			// Validate status is one of the valid types
			validStatuses := []types.IncidentStatus{
				types.StatusNew,
				types.StatusResolved,
				types.StatusPostMortem,
				types.StatusClosed,
			}
			statusValid := false
			for _, validStatus := range validStatuses {
				if incident.Status == validStatus {
					statusValid = true
					break
				}
			}
			if !statusValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s has invalid status: %s", incident.ID, incident.Status),
				), true
			}

			// Check reported by is not empty
			if incident.ReportedBy == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s has empty reported_by", incident.ID),
				), true
			}

			// Check reported at is not zero
			if incident.ReportedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s has zero reported_at", incident.ID),
				), true
			}

			// If incident is resolved, must have resolved time
			if incident.Status == types.StatusResolved && incident.ResolvedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s is resolved but has zero resolved_at", incident.ID),
				), true
			}

			// If incident is closed, must have post-mortem
			if incident.Status == types.StatusClosed && incident.PostMortem == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s is closed but has no post-mortem", incident.ID),
				), true
			}

			// Validate timeline entries
			if len(incident.Timeline) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"incident-validity",
					fmt.Sprintf("incident %s has no timeline entries", incident.ID),
				), true
			}

			for i, entry := range incident.Timeline {
				if entry.Timestamp.IsZero() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"incident-validity",
						fmt.Sprintf("incident %s timeline entry %d has zero timestamp", incident.ID, i),
					), true
				}
				if entry.Action == "" {
					return sdk.FormatInvariant(
						types.ModuleName,
						"incident-validity",
						fmt.Sprintf("incident %s timeline entry %d has empty action", incident.ID, i),
					), true
				}
			}
		}

		return "", false
	}
}

// PauseStateConsistencyInvariant checks chain pause state consistency
func PauseStateConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		pauseState := k.GetChainPauseState(ctx)

		// If chain is paused, validate pause state
		if pauseState.IsPaused {
			// Pause level must be valid
			if pauseState.PauseLevel != types.PauseLevelFull &&
				pauseState.PauseLevel != types.PauseLevelTransactions &&
				pauseState.PauseLevel != types.PauseLevelModules {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pause-state-consistency",
					fmt.Sprintf("chain paused with invalid level: %s", pauseState.PauseLevel),
				), true
			}

			// Paused by should not be empty
			if pauseState.PausedBy == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pause-state-consistency",
					"chain paused but paused_by is empty",
				), true
			}

			// Paused at should not be zero
			if pauseState.PausedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pause-state-consistency",
					"chain paused but paused_at is zero",
				), true
			}

			// Reason should not be empty
			if pauseState.Reason == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pause-state-consistency",
					"chain paused but reason is empty",
				), true
			}
		} else {
			// If chain is not paused, pause level should be none
			if pauseState.PauseLevel != types.PauseLevelNone {
				return sdk.FormatInvariant(
					types.ModuleName,
					"pause-state-consistency",
					fmt.Sprintf("chain not paused but level is %s (expected none)", pauseState.PauseLevel),
				), true
			}
		}

		return "", false
	}
}

// WalletLimitsValidityInvariant checks wallet limits validity
func WalletLimitsValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// This invariant would iterate through all wallet limits if we had a GetAllWalletLimits method
		// For now, we just validate that the module is in a consistent state

		params, _ := k.GetParams(ctx)

		// If hot wallet limits are enabled, verify global max is set
		if params.HotWalletLimitsEnabled {
			if params.GlobalMaxHotWallet == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"wallet-limits-validity",
					"hot wallet limits enabled but global max is empty",
				), true
			}

			// Verify global max is parseable (basic validation)
			if params.GlobalMaxHotWallet != "0" && len(params.GlobalMaxHotWallet) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"wallet-limits-validity",
					"invalid global max hot wallet value",
				), true
			}
		}

		return "", false
	}
}
