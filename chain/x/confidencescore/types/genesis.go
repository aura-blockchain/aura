package types

import (
	"fmt"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// GenesisState defines the initial state of the confidencescore module
type GenesisState struct {
	Params       Params
	UserRecords  []UserConfidenceRecord
	SlashRecords []SlashRecord
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:       DefaultParams(),
		UserRecords:  []UserConfidenceRecord{},
		SlashRecords: []SlashRecord{},
	}
}

// GenesisStateFromProto converts proto genesis state to internal type
func GenesisStateFromProto(pb *confidencescorepb.GenesisState) GenesisState {
	if pb == nil {
		return DefaultGenesisState()
	}

	userRecords := make([]UserConfidenceRecord, len(pb.UserRecords))
	for i, record := range pb.UserRecords {
		userRecords[i] = UserConfidenceRecordFromProto(record)
	}

	slashRecords := make([]SlashRecord, len(pb.SlashRecords))
	for i, record := range pb.SlashRecords {
		slashRecords[i] = SlashRecordFromProto(record)
	}

	return GenesisState{
		Params:       ParamsFromProto(pb.Params),
		UserRecords:  userRecords,
		SlashRecords: slashRecords,
	}
}

// GenesisStateToProto converts internal genesis state to proto type
func GenesisStateToProto(gs GenesisState) *confidencescorepb.GenesisState {
	userRecords := make([]*confidencescorepb.UserConfidenceRecord, len(gs.UserRecords))
	for i, record := range gs.UserRecords {
		userRecords[i] = UserConfidenceRecordToProto(record)
	}

	slashRecords := make([]*confidencescorepb.SlashRecord, len(gs.SlashRecords))
	for i, record := range gs.SlashRecords {
		slashRecords[i] = SlashRecordToProto(record)
	}

	return &confidencescorepb.GenesisState{
		Params:       ParamsToProto(gs.Params),
		UserRecords:  userRecords,
		SlashRecords: slashRecords,
		Completions:  []*confidencescorepb.IRCompletion{}, // Embedded in user records
		History:      []*confidencescorepb.ConfidenceHistory{},
	}
}

// Validate performs validation on the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate user records
	walletsSeen := make(map[string]bool)
	for i, record := range gs.UserRecords {
		if record.WalletAddress == "" {
			return fmt.Errorf("user record[%d]: wallet address cannot be empty", i)
		}

		if walletsSeen[record.WalletAddress] {
			return fmt.Errorf("duplicate wallet address: %s", record.WalletAddress)
		}
		walletsSeen[record.WalletAddress] = true

		// Validate that total score matches sum of completions
		var calculatedScore uint64
		for _, completion := range record.CompletedIRs {
			calculatedScore += completion.FinalScore
		}

		if calculatedScore != record.TotalScore {
			return fmt.Errorf("user %s: total score mismatch (expected %d, got %d)",
				record.WalletAddress, calculatedScore, record.TotalScore)
		}

		// Validate verification status
		if record.TotalScore >= gs.Params.VerificationThreshold {
			if record.Status == VerificationStatusUnverified {
				return fmt.Errorf("user %s: score >= threshold but status is unverified",
					record.WalletAddress)
			}
		}
	}

	return nil
}

// TestGenesisState returns a genesis state with test data for testing
func TestGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
		UserRecords: []UserConfidenceRecord{
			{
				WalletAddress: "aura1testuser1",
				TotalScore:    12000,
				HasAnchor:     true,
				Status:        VerificationStatusVerified,
				ArenaScores: map[string]ArenaScore{
					"Biometric": {
						ArenaType:        "Biometric",
						TotalScore:       6000,
						IRCount:          5,
						FocusBonusActive: true,
					},
					"Knowledge": {
						ArenaType:  "Knowledge",
						TotalScore: 6000,
						IRCount:    4,
					},
				},
				AnchorInfo: &AnchorInfo{
					Completed:   true,
					CompletedAt: 1699900000,
					BlockHeight: 100,
				},
			},
		},
		SlashRecords: []SlashRecord{},
	}
}
