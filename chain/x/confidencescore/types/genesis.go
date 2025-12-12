package types

import (
	"encoding/json"
	"fmt"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultGenesisState returns the default genesis state for the confidencescore module
func DefaultGenesisState() *confidencescorepb.GenesisState {
	return &confidencescorepb.GenesisState{
		Params:       ParamsToProto(DefaultParams()),
		UserRecords:  []*confidencescorepb.UserConfidenceRecord{},
		Completions:  []*confidencescorepb.IRCompletion{},
		History:      []*confidencescorepb.ConfidenceHistory{},
		SlashRecords: []*confidencescorepb.SlashRecord{},
	}
}

// ValidateGenesisState validates the genesis state
func ValidateGenesisState(state *confidencescorepb.GenesisState) error {
	if state == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	// Validate params
	if state.Params != nil {
		// Convert proto params to local Params type for validation
		params := Params{
			VerificationThreshold:       state.Params.VerificationThreshold,
			HighAssuranceThreshold:      state.Params.HighAssuranceThreshold,
			ArenaFocusThreshold:         state.Params.ArenaFocusThreshold,
			VelocityBonusDays:           state.Params.VelocityBonusDays,
			VelocityBonusMultipliersBps: state.Params.VelocityBonusMultipliersBps,
			ArenaMultipliersBps:         state.Params.ArenaMultipliersBps,
			SlashPercentage:             state.Params.SlashPercentage,
			AppealDeposit:               state.Params.AppealDeposit,
			MaxIrsPerDay:                state.Params.MaxIrsPerDay,
			MaxIrsPerHour:               state.Params.MaxIrsPerHour,
			JackpotOdds:                 state.Params.JackpotOdds,
			JackpotMultipliersBps:       state.Params.JackpotMultipliersBps,
			StalenessEnabled:            state.Params.StalenessEnabled,
			DegradationRatePerYear:      state.Params.DegradationRatePerYear,
			PoiRewardsEnabled:           state.Params.PoiRewardsEnabled,
			UserRewardSplitPercent:      state.Params.UserRewardSplitPercent,
			VelocityBonusEnabled:        state.Params.VelocityBonusEnabled,
		}
		if err := ValidateParams(&params); err != nil {
			return fmt.Errorf("invalid params: %w", err)
		}
	}

	// Track wallet addresses for validation
	walletAddresses := make(map[string]struct{})

	// Validate user records
	for i, record := range state.UserRecords {
		if record == nil {
			return fmt.Errorf("user record at index %d is nil", i)
		}
		if record.WalletAddress == "" {
			return fmt.Errorf("user record at index %d has empty wallet address", i)
		}
		if _, exists := walletAddresses[record.WalletAddress]; exists {
			return fmt.Errorf("duplicate wallet address in user records: %s", record.WalletAddress)
		}
		walletAddresses[record.WalletAddress] = struct{}{}

		// Validate that total score is non-negative
		// Note: uint64 is always non-negative, but we check consistency

		// Validate completed IRs
		irIDs := make(map[string]struct{})
		for j, completion := range record.CompletedIrs {
			if completion == nil {
				return fmt.Errorf("completion at index %d for wallet %s is nil", j, record.WalletAddress)
			}
			if completion.IrId == "" {
				return fmt.Errorf("completion at index %d for wallet %s has empty IR ID", j, record.WalletAddress)
			}
			if _, exists := irIDs[completion.IrId]; exists {
				return fmt.Errorf("duplicate IR completion %s for wallet %s", completion.IrId, record.WalletAddress)
			}
			irIDs[completion.IrId] = struct{}{}

			// Validate completion scores
			if completion.FinalScore < completion.BaseScore {
				return fmt.Errorf("final score cannot be less than base score for IR %s", completion.IrId)
			}
		}

		// Validate arena scores
		var arenaTotal uint64
		for arenaType, arenaScore := range record.ArenaScores {
			if arenaScore == nil {
				return fmt.Errorf("arena score for %s is nil for wallet %s", arenaType, record.WalletAddress)
			}
			if arenaScore.ArenaType != arenaType {
				return fmt.Errorf("arena score type mismatch: key=%s, type=%s", arenaType, arenaScore.ArenaType)
			}
			arenaTotal += arenaScore.TotalScore
		}

		// Validate verification status consistency
		if record.Status == VerificationStatus_VERIFICATION_STATUS_VERIFIED {
			if state.Params != nil && record.TotalScore < state.Params.VerificationThreshold {
				return fmt.Errorf("wallet %s marked as verified but score %d below threshold %d",
					record.WalletAddress, record.TotalScore, state.Params.VerificationThreshold)
			}
		}
	}

	// Validate standalone completions (if any)
	for i, completion := range state.Completions {
		if completion == nil {
			return fmt.Errorf("completion at index %d is nil", i)
		}
		if completion.IrId == "" {
			return fmt.Errorf("completion at index %d has empty IR ID", i)
		}
		if completion.FinalScore < completion.BaseScore {
			return fmt.Errorf("completion %s: final score cannot be less than base score", completion.IrId)
		}
	}

	// Validate history entries
	for i, history := range state.History {
		if history == nil {
			return fmt.Errorf("history entry at index %d is nil", i)
		}
		if history.WalletAddress == "" {
			return fmt.Errorf("history entry at index %d has empty wallet address", i)
		}

		// Validate all score changes
		for j, change := range history.Changes {
			if change == nil {
				return fmt.Errorf("score change at index %d for wallet %s is nil", j, history.WalletAddress)
			}
			if change.Reason == ChangeReason_CHANGE_REASON_UNSPECIFIED {
				return fmt.Errorf("score change at index %d for wallet %s has unspecified reason", j, history.WalletAddress)
			}
		}
	}

	// Validate slash records
	slashIDs := make(map[string]struct{})
	for i, slash := range state.SlashRecords {
		if slash == nil {
			return fmt.Errorf("slash record at index %d is nil", i)
		}
		if slash.WalletAddress == "" {
			return fmt.Errorf("slash record at index %d has empty wallet address", i)
		}
		if slash.Reason == SlashReason_SLASH_REASON_UNSPECIFIED {
			return fmt.Errorf("slash record at index %d for wallet %s has unspecified reason", i, slash.WalletAddress)
		}
		if slash.SlashTxHash == "" {
			return fmt.Errorf("slash record at index %d for wallet %s has empty tx hash", i, slash.WalletAddress)
		}

		slashID := fmt.Sprintf("%s:%s", slash.WalletAddress, slash.SlashTxHash)
		if _, exists := slashIDs[slashID]; exists {
			return fmt.Errorf("duplicate slash record for wallet %s with tx hash %s", slash.WalletAddress, slash.SlashTxHash)
		}
		slashIDs[slashID] = struct{}{}
	}

	return nil
}

// DefaultGenesis returns the default genesis as raw JSON
func DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis is an alias for ValidateGenesisState for consistency
func ValidateGenesis(state *confidencescorepb.GenesisState) error {
	return ValidateGenesisState(state)
}
