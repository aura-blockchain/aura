package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// InitGenesis initializes the module state from genesis data
func (k *Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) error {
	// Validate genesis before importing
	if err := types.ValidateGenesisState(&data); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params
	if data.Params != nil {
		params := types.Params{
			VerificationThreshold:       data.Params.VerificationThreshold,
			HighAssuranceThreshold:      data.Params.HighAssuranceThreshold,
			ArenaFocusThreshold:         data.Params.ArenaFocusThreshold,
			VelocityBonusDays:           data.Params.VelocityBonusDays,
			VelocityBonusMultipliersBps: data.Params.VelocityBonusMultipliersBps,
			ArenaMultipliersBps:         data.Params.ArenaMultipliersBps,
			SlashPercentage:             data.Params.SlashPercentage,
			AppealDeposit:               data.Params.AppealDeposit,
			MaxIrsPerDay:                data.Params.MaxIrsPerDay,
			MaxIrsPerHour:               data.Params.MaxIrsPerHour,
			JackpotOdds:                 data.Params.JackpotOdds,
			JackpotMultipliersBps:       data.Params.JackpotMultipliersBps,
			StalenessEnabled:            data.Params.StalenessEnabled,
			DegradationRatePerYear:      data.Params.DegradationRatePerYear,
			PoiRewardsEnabled:           data.Params.PoiRewardsEnabled,
			UserRewardSplitPercent:      data.Params.UserRewardSplitPercent,
			VelocityBonusEnabled:        data.Params.VelocityBonusEnabled,
		}
		if err := k.paramsStore.SetParams(params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Import user records using KV store
	for _, record := range data.UserRecords {
		if record == nil {
			continue
		}
		if err := k.SetUserRecord(ctx, *record); err != nil {
			return fmt.Errorf("failed to set user record: %w", err)
		}

		// Import completions for this user
		for _, completion := range record.CompletedIrs {
			if completion != nil {
				if err := k.SetIRCompletion(ctx, record.WalletAddress, *completion); err != nil {
					return fmt.Errorf("failed to set IR completion: %w", err)
				}
			}
		}
	}

	// Import slash records using KV store
	for _, slash := range data.SlashRecords {
		if slash == nil {
			continue
		}
		if err := k.AddSlashRecord(ctx, *slash); err != nil {
			return fmt.Errorf("failed to add slash record: %w", err)
		}
	}

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k *Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	store := k.storeService.OpenKVStore(ctx)

	// Export params
	params := k.paramsStore.GetParams()
	protoParams := &confidencescorepb.Params{
		VerificationThreshold:       params.VerificationThreshold,
		HighAssuranceThreshold:      params.HighAssuranceThreshold,
		ArenaFocusThreshold:         params.ArenaFocusThreshold,
		VelocityBonusDays:           params.VelocityBonusDays,
		VelocityBonusMultipliersBps: params.VelocityBonusMultipliersBps,
		ArenaMultipliersBps:         params.ArenaMultipliersBps,
		SlashPercentage:             params.SlashPercentage,
		AppealDeposit:               params.AppealDeposit,
		MaxIrsPerDay:                params.MaxIrsPerDay,
		MaxIrsPerHour:               params.MaxIrsPerHour,
		JackpotOdds:                 params.JackpotOdds,
		JackpotMultipliersBps:       params.JackpotMultipliersBps,
		StalenessEnabled:            params.StalenessEnabled,
		DegradationRatePerYear:      params.DegradationRatePerYear,
		PoiRewardsEnabled:           params.PoiRewardsEnabled,
		UserRewardSplitPercent:      params.UserRewardSplitPercent,
		VelocityBonusEnabled:        params.VelocityBonusEnabled,
	}

	// Export user records from KV store
	userRecords := []*confidencescorepb.UserConfidenceRecord{}
	prefix := []byte(types.UserRecordStoreKeyPrefix)
	// Calculate end bytes for prefix iteration
	endBytes := append([]byte(nil), prefix...)
	endBytes[len(endBytes)-1]++
	iterator, err := store.Iterator(prefix, endBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create user records iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record confidencescorepb.UserConfidenceRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user record at key %x: %w", iterator.Key(), err)
		}
		recordCopy := record
		userRecords = append(userRecords, &recordCopy)
	}

	// Export standalone completions (none for now as they're part of user records)
	completions := []*confidencescorepb.IRCompletion{}

	// Export history (not stored separately in current implementation)
	historyRecords := []*confidencescorepb.ConfidenceHistory{}

	// Export slash records from KV store
	slashRecords := []*confidencescorepb.SlashRecord{}
	slashPrefix := []byte(types.SlashRecordStoreKeyPrefix)
	// Calculate end bytes for prefix iteration
	slashEndBytes := append([]byte(nil), slashPrefix...)
	slashEndBytes[len(slashEndBytes)-1]++
	slashIterator, err := store.Iterator(slashPrefix, slashEndBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create slash records iterator: %w", err)
	}
	defer slashIterator.Close()

	for ; slashIterator.Valid(); slashIterator.Next() {
		var record confidencescorepb.SlashRecord
		if err := k.cdc.Unmarshal(slashIterator.Value(), &record); err != nil {
			return nil, fmt.Errorf("failed to unmarshal slash record at key %x: %w", slashIterator.Key(), err)
		}
		recordCopy := record
		slashRecords = append(slashRecords, &recordCopy)
	}

	return &types.GenesisState{
		Params:       protoParams,
		UserRecords:  userRecords,
		Completions:  completions,
		History:      historyRecords,
		SlashRecords: slashRecords,
	}, nil
}
