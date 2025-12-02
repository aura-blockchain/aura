package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Validator Security Operations
// =============================================================================

// SetValidatorSecurityInfo stores validator security information
func (k Keeper) SetValidatorSecurityInfo(ctx sdk.Context, info *securitypb.ValidatorSecurityInfo) {
	store := k.GetStore(ctx)
	key := append(types.ValidatorInfoKey, []byte(info.ValidatorAddress)...)
	bz := k.cdc.MustMarshal(info)
	store.Set(key, bz)
}

// GetValidatorSecurityInfo retrieves validator security information
func (k Keeper) GetValidatorSecurityInfo(ctx sdk.Context, valAddr string) (*securitypb.ValidatorSecurityInfo, bool) {
	store := k.GetStore(ctx)
	key := append(types.ValidatorInfoKey, []byte(valAddr)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var info securitypb.ValidatorSecurityInfo
	k.cdc.MustUnmarshal(bz, &info)
	return &info, true
}

// GetAllValidatorSecurityInfos returns all validator security infos
func (k Keeper) GetAllValidatorSecurityInfos(ctx sdk.Context) []*securitypb.ValidatorSecurityInfo {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorInfoKey)
	defer iterator.Close()

	var infos []*securitypb.ValidatorSecurityInfo
	for ; iterator.Valid(); iterator.Next() {
		var info securitypb.ValidatorSecurityInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)
		infos = append(infos, &info)
	}
	return infos
}

// SetDoubleSignEvidence stores double sign evidence
func (k Keeper) SetDoubleSignEvidence(ctx sdk.Context, evidence *securitypb.DoubleSignEvidence) {
	store := k.GetStore(ctx)
	// Use ValidatorAddress as key since DoubleSignEvidence has no id field
	key := append(types.DoubleSignEvidenceKey, []byte(evidence.ValidatorAddress)...)
	bz := k.cdc.MustMarshal(evidence)
	store.Set(key, bz)
}

// GetAllDoubleSignEvidence returns all double sign evidence
func (k Keeper) GetAllDoubleSignEvidence(ctx sdk.Context) []*securitypb.DoubleSignEvidence {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.DoubleSignEvidenceKey)
	defer iterator.Close()

	var evidences []*securitypb.DoubleSignEvidence
	for ; iterator.Valid(); iterator.Next() {
		var evidence securitypb.DoubleSignEvidence
		k.cdc.MustUnmarshal(iterator.Value(), &evidence)
		evidences = append(evidences, &evidence)
	}
	return evidences
}

// SetDowntimeInfraction stores a downtime infraction
func (k Keeper) SetDowntimeInfraction(ctx sdk.Context, infraction *securitypb.DowntimeInfraction) {
	store := k.GetStore(ctx)
	// Use ValidatorAddress as key since DowntimeInfraction has no id field
	key := append(types.DowntimeInfractionKey, []byte(infraction.ValidatorAddress)...)
	bz := k.cdc.MustMarshal(infraction)
	store.Set(key, bz)
}

// GetAllDowntimeInfractions returns all downtime infractions
func (k Keeper) GetAllDowntimeInfractions(ctx sdk.Context) []*securitypb.DowntimeInfraction {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.DowntimeInfractionKey)
	defer iterator.Close()

	var infractions []*securitypb.DowntimeInfraction
	for ; iterator.Valid(); iterator.Next() {
		var infraction securitypb.DowntimeInfraction
		k.cdc.MustUnmarshal(iterator.Value(), &infraction)
		infractions = append(infractions, &infraction)
	}
	return infractions
}

// SetValidatorAlert stores a validator alert
func (k Keeper) SetValidatorAlert(ctx sdk.Context, alert *securitypb.ValidatorAlert) {
	store := k.GetStore(ctx)
	key := append(types.ValidatorAlertKey, []byte(alert.Id)...)
	bz := k.cdc.MustMarshal(alert)
	store.Set(key, bz)
}

// GetAllValidatorAlerts returns all validator alerts
func (k Keeper) GetAllValidatorAlerts(ctx sdk.Context) []*securitypb.ValidatorAlert {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorAlertKey)
	defer iterator.Close()

	var alerts []*securitypb.ValidatorAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert securitypb.ValidatorAlert
		k.cdc.MustUnmarshal(iterator.Value(), &alert)
		alerts = append(alerts, &alert)
	}
	return alerts
}

// SetSentryNode stores a sentry node configuration
func (k Keeper) SetSentryNode(ctx sdk.Context, sentry *securitypb.SentryNodeInfo) {
	store := k.GetStore(ctx)
	key := append(types.SentryNodeKey, []byte(sentry.ValidatorAddress)...)
	key = append(key, []byte(sentry.Address)...)
	bz := k.cdc.MustMarshal(sentry)
	store.Set(key, bz)
}

// GetAllSentryNodes returns all sentry nodes
func (k Keeper) GetAllSentryNodes(ctx sdk.Context) []*securitypb.SentryNodeInfo {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SentryNodeKey)
	defer iterator.Close()

	var sentries []*securitypb.SentryNodeInfo
	for ; iterator.Valid(); iterator.Next() {
		var sentry securitypb.SentryNodeInfo
		k.cdc.MustUnmarshal(iterator.Value(), &sentry)
		sentries = append(sentries, &sentry)
	}
	return sentries
}

// JailValidator jails a validator
func (k Keeper) JailValidator(ctx sdk.Context, valAddr sdk.ValAddress) error {
	consAddr, err := sdk.ConsAddressFromBech32(valAddr.String())
	if err != nil {
		return err
	}
	if err := k.stakingKeeper.Jail(ctx, consAddr); err != nil {
		return err
	}
	return nil
}

// UnjailValidator unjails a validator
func (k Keeper) UnjailValidator(ctx sdk.Context, valAddr sdk.ValAddress) error {
	consAddr, err := sdk.ConsAddressFromBech32(valAddr.String())
	if err != nil {
		return err
	}
	if err := k.stakingKeeper.Unjail(ctx, consAddr); err != nil {
		return err
	}
	return nil
}

// SlashValidator slashes a validator
func (k Keeper) SlashValidator(ctx sdk.Context, valAddr sdk.ValAddress, height int64, power int64, factor string) error {
	consAddr, err := sdk.ConsAddressFromBech32(valAddr.String())
	if err != nil {
		return err
	}
	if _, err := k.stakingKeeper.Slash(ctx, consAddr, height, power, factor); err != nil {
		return err
	}
	return nil
}

// TrackMissedBlock records a missed block for a validator
func (k Keeper) TrackMissedBlock(ctx sdk.Context, valAddr string) {
	info, found := k.GetValidatorSecurityInfo(ctx, valAddr)
	if !found {
		info = &securitypb.ValidatorSecurityInfo{
			ValidatorAddress:    valAddr,
			MissedBlocksCounter: 0,
		}
	}
	info.MissedBlocksCounter++
	k.SetValidatorSecurityInfo(ctx, info)
}

// TrackSignedBlock records a signed block for a validator
func (k Keeper) TrackSignedBlock(ctx sdk.Context, valAddr string) {
	info, found := k.GetValidatorSecurityInfo(ctx, valAddr)
	if !found {
		info = &securitypb.ValidatorSecurityInfo{
			ValidatorAddress:    valAddr,
			MissedBlocksCounter: 0,
		}
	}
	// Reset missed blocks counter on successful sign
	if info.MissedBlocksCounter > 0 {
		info.MissedBlocksCounter = 0
	}
	k.SetValidatorSecurityInfo(ctx, info)
}
