package keeper

import (
	"context"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	// the address capable of executing a MsgUpdateParams message
	authority string

	// Hooks for staking integration
	stakingKeeper  StakingKeeper
	slashingKeeper SlashingKeeper
	bankKeeper     BankKeeper
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	authority string,
	stakingKeeper StakingKeeper,
	slashingKeeper SlashingKeeper,
	bankKeeper BankKeeper,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		authority:      authority,
		stakingKeeper:  stakingKeeper,
		slashingKeeper: slashingKeeper,
		bankKeeper:     bankKeeper,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx context.Context, params *types.ValidatorSecurityParams) error {
	if err := types.ValidateParams(params); err != nil {
		return err
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(params)
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetParams gets the module parameters
func (k Keeper) GetParams(ctx context.Context) *types.ValidatorSecurityParams {
	store := k.getStore(ctx)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.ValidatorSecurityParams
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		k.logger.Error("failed to unmarshal validator security params", "error", err)
		return types.DefaultParams()
	}
	return &params
}

func (k Keeper) getStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(k.storeKey)
}

// RegisterValidator registers a new validator with security information
func (k Keeper) RegisterValidator(
	ctx context.Context,
	validatorAddr string,
	hotKey, coldKey string,
	region, countryCode string,
	latitude, longitude float64,
	backupValidators []string,
) error {
	// Check if validator already exists
	if k.HasValidatorSecurityInfo(ctx, validatorAddr) {
		return types.ErrValidatorAlreadyRegistered
	}

	params := k.GetParams(ctx)

	// Validate key separation if required
	keysSeparated := hotKey != "" && coldKey != ""
	if keysSeparated && hotKey == coldKey {
		return types.ErrInvalidKeys
	}

	// Validate geographic location
	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return types.ErrInvalidGeographicLocation
	}

	// Check region capacity if geo distribution is enabled
	if params.EnableGeoDistribution {
		if err := k.checkRegionCapacity(ctx, region); err != nil {
			return err
		}
	}

	// Create validator security info
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	info := types.ValidatorSecurityInfo{
		ValidatorAddress:         validatorAddr,
		HotKey:                   hotKey,
		ColdKey:                  coldKey,
		KeysSeparated:            keysSeparated,
		SentryNodeAddresses:      []string{},
		Region:                   region,
		CountryCode:              countryCode,
		Latitude:                 latitude,
		Longitude:                longitude,
		IsJailed:                 false,
		IsTombstoned:             false,
		MissedBlocksCounter:      0,
		IndexOffset:              0,
		LastSeen:                 timestamppb.New(sdkCtx.BlockTime()),
		BackupValidatorAddresses: backupValidators,
		FailoverActive:           false,
	}

	k.SetValidatorSecurityInfo(ctx, info)

	// Increment region validator count
	if params.EnableGeoDistribution {
		k.incrementRegionCount(ctx, region)
	}

	return nil
}

// UpdateValidatorSecurityInfo updates existing validator security metadata.
func (k Keeper) UpdateValidatorSecurityInfo(
	ctx context.Context,
	validatorAddr string,
	hotKey, coldKey string,
	region, countryCode string,
	latitude, longitude float64,
	backupValidators []string,
) error {
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)

	keysSeparated := hotKey != "" && coldKey != ""
	if keysSeparated && hotKey == coldKey {
		return types.ErrInvalidKeys
	}

	if latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180 {
		return types.ErrInvalidGeographicLocation
	}

	if params.EnableGeoDistribution && region != info.Region {
		if err := k.checkRegionCapacity(ctx, region); err != nil {
			return err
		}
	}

	previousRegion := info.Region

	info.HotKey = hotKey
	info.ColdKey = coldKey
	info.KeysSeparated = keysSeparated
	info.Region = region
	info.CountryCode = countryCode
	info.Latitude = latitude
	info.Longitude = longitude
	info.BackupValidatorAddresses = backupValidators

	k.SetValidatorSecurityInfo(ctx, info)

	if params.EnableGeoDistribution && previousRegion != region {
		if previousRegion != "" {
			k.decrementRegionCount(ctx, previousRegion)
		}
		k.incrementRegionCount(ctx, region)
	}

	return nil
}

// SetValidatorSecurityInfo sets the security info for a validator
func (k Keeper) SetValidatorSecurityInfo(ctx context.Context, info types.ValidatorSecurityInfo) {
	store := k.getStore(ctx)
	key := types.GetValidatorSecurityInfoKey(info.ValidatorAddress)
	bz := k.cdc.MustMarshal(&info)
	store.Set(key, bz)

	// Update jailed/tombstoned indexes
	if info.IsJailed {
		store.Set(append(types.JailedValidatorsKey, []byte(info.ValidatorAddress)...), []byte{1})
	}
	if info.IsTombstoned {
		store.Set(append(types.TombstonedValidatorsKey, []byte(info.ValidatorAddress)...), []byte{1})
	}
}

// GetValidatorSecurityInfo gets the security info for a validator
func (k Keeper) GetValidatorSecurityInfo(ctx context.Context, validatorAddr string) (types.ValidatorSecurityInfo, error) {
	store := k.getStore(ctx)
	key := types.GetValidatorSecurityInfoKey(validatorAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorSecurityInfo{}, types.ErrValidatorNotFound
	}

	var info types.ValidatorSecurityInfo
	if err := k.cdc.Unmarshal(bz, &info); err != nil {
		k.logger.Error("failed to unmarshal validator security info", "validator", validatorAddr, "error", err)
		return types.ValidatorSecurityInfo{}, types.ErrCorruptedState
	}
	return info, nil
}

// HasValidatorSecurityInfo checks if a validator has security info
func (k Keeper) HasValidatorSecurityInfo(ctx context.Context, validatorAddr string) bool {
	store := k.getStore(ctx)
	key := types.GetValidatorSecurityInfoKey(validatorAddr)
	return store.Has(key)
}

// GetAllValidators returns all validator security info
func (k Keeper) GetAllValidators(ctx context.Context) []types.ValidatorSecurityInfo {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorSecurityInfoKey)
	defer iterator.Close()

	var validators []types.ValidatorSecurityInfo
	for ; iterator.Valid(); iterator.Next() {
		var info types.ValidatorSecurityInfo
		if err := k.cdc.Unmarshal(iterator.Value(), &info); err != nil {
			k.logger.Error("failed to unmarshal validator info in GetAllValidators, skipping", "error", err)
			continue
		}
		validators = append(validators, info)
	}

	return validators
}

func (k Keeper) checkRegionCapacity(ctx context.Context, region string) error {
	params := k.GetParams(ctx)
	currentCount := k.getRegionValidatorCount(ctx, region)

	if currentCount >= int64(params.MaxValidatorsPerRegion) {
		return types.ErrRegionCapacityExceeded
	}

	return nil
}

func (k Keeper) getRegionValidatorCount(ctx context.Context, region string) int64 {
	store := k.getStore(ctx)
	key := types.GetRegionValidatorCountKey(region)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	return int64(sdk.BigEndianToUint64(bz))
}

func (k Keeper) incrementRegionCount(ctx context.Context, region string) {
	store := k.getStore(ctx)
	key := types.GetRegionValidatorCountKey(region)
	count := k.getRegionValidatorCount(ctx, region)
	store.Set(key, sdk.Uint64ToBigEndian(uint64(count+1)))
}

func (k Keeper) decrementRegionCount(ctx context.Context, region string) {
	store := k.getStore(ctx)
	key := types.GetRegionValidatorCountKey(region)
	count := k.getRegionValidatorCount(ctx, region)
	if count > 0 {
		store.Set(key, sdk.Uint64ToBigEndian(uint64(count-1)))
	}
}
