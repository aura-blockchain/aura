package keeper

import (
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// QueryUserScore retrieves a user's score and status
func (k *Keeper) QueryUserScore(walletAddr string) (uint64, bool, *types.AnchorInfo, map[string]types.ArenaScore, uint32, int64, types.VerificationStatus, uint64, error) {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		// Return empty record for non-existent users
		return 0, false, nil, make(map[string]types.ArenaScore), 0, 0, types.VerificationStatusUnverified, 0, nil
	}

	params := k.GetParams()
	isVerified := record.TotalScore >= params.VerificationThreshold &&
		record.Status == types.VerificationStatusVerified

	return record.TotalScore,
		isVerified,
		record.AnchorInfo,
		record.ArenaScores,
		uint32(len(record.CompletedIRs)),
		record.LastUpdated,
		record.Status,
		record.VerificationAchievedHeight,
		nil
}

// QueryUserCompletions retrieves a user's IR completions with optional filtering
func (k *Keeper) QueryUserCompletions(walletAddr, arenaFilter string, offset, limit int) ([]types.IRCompletion, int) {
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		return []types.IRCompletion{}, 0
	}

	completions := record.CompletedIRs

	// Apply arena filter if specified
	if arenaFilter != "" {
		filtered := []types.IRCompletion{}
		for _, completion := range completions {
			if completion.Arena == arenaFilter {
				filtered = append(filtered, completion)
			}
		}
		completions = filtered
	}

	total := len(completions)

	// Apply pagination
	if offset >= total {
		return []types.IRCompletion{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return completions[offset:end], total
}

// QueryThresholds returns all verification thresholds
func (k *Keeper) QueryThresholds() (uint64, map[string]uint64, map[string]uint64) {
	params := k.GetParams()

	// VC thresholds
	vcThresholds := map[string]uint64{
		"VerifiedHuman": params.VerificationThreshold,
		"HighAssurance": params.HighAssuranceThreshold,
		"AgeOver18":     params.VerificationThreshold,
		"AgeOver21":     params.VerificationThreshold,
		"Professional":  12000,
	}

	// Arena focus thresholds
	arenaThresholds := map[string]uint64{
		"Biometric":     params.ArenaFocusThreshold,
		"GeoLocation":   params.ArenaFocusThreshold,
		"HighAssurance": params.ArenaFocusThreshold,
		"Knowledge":     params.ArenaFocusThreshold,
		"Social":        params.ArenaFocusThreshold,
		"Possession":    params.ArenaFocusThreshold,
	}

	return params.VerificationThreshold, vcThresholds, arenaThresholds
}
