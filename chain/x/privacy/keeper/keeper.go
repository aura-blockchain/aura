package keeper

import (
	"context"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// Keeper handles privacy module state
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	authority      string
	zkProofSystem  any // Privacy utilities - avoiding import cycle
	mixingService  any
	viewKeyManager any
	networkPrivacy any
	memoEncryptor  any
}

// NewKeeper creates a new privacy keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		authority:      authority,
		zkProofSystem:  nil,
		mixingService:  nil,
		viewKeyManager: nil,
		networkPrivacy: nil,
		memoEncryptor:  nil,
	}
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/privacy")
}

// SetParams sets the privacy module parameters
func (k Keeper) SetParams(ctx context.Context, params *privacyproto.Params) error {
	store := k.getStore(ctx)
	bz, err := k.cdc.Marshal(params)
	if err != nil {
		return err
	}
	store.Set([]byte("params"), bz)
	return nil
}

// GetParams gets the privacy module parameters
func (k Keeper) GetParams(ctx context.Context) (*privacyproto.Params, error) {
	store := k.getStore(ctx)
	bz := store.Get([]byte("params"))
	if bz == nil {
		// Return default params
		return &privacyproto.Params{
			EnableZkProofs:                 true,
			EnableStealthAddresses:         true,
			EnableRingSignatures:           true,
			EnableConfidentialTransactions: true,
			EnableNetworkPrivacy:           true,
			EnableMixing:                   true,
			MinRingSize:                    3,
			MaxRingSize:                    16,
			MinMixingParticipants:          5,
			MixingFee:                      "1000",
			ZkProofVerificationCost:        10000,
		}, nil
	}

	params := &privacyproto.Params{}
	if err := k.cdc.Unmarshal(bz, params); err != nil {
		return nil, err
	}
	return params, nil
}

// getStore returns the KVStore for the privacy module
func (k Keeper) getStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(k.storeKey)
}
