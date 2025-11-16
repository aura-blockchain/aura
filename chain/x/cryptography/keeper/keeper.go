package keeper

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// Keeper maintains the state for the cryptography module
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	logger   log.Logger

	// Cached data for performance
	mu                sync.RWMutex
	rotationSchedules map[string]*cryptoproto.KeyRotationSchedule
	thresholdSchemes  map[string]*cryptoproto.ThresholdSignatureScheme
	thresholdShares   map[string]map[string]*cryptoproto.ThresholdSignatureShare
	zkProofConfigs    map[string]*cryptoproto.ZKProofConfig
	secureEnclaves    map[string]*cryptoproto.SecureEnclaveConfig
	quantumKeys       map[string]*cryptoproto.QuantumResistantKey
	randomSources     map[string]*cryptoproto.CryptoRandomSource
	certificatePins   map[string]*cryptoproto.CertificatePin

	// Authority for governance actions
	authority string
}

// NewKeeper creates a new cryptography keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	logger log.Logger,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		logger:            logger,
		authority:         authority,
		rotationSchedules: make(map[string]*cryptoproto.KeyRotationSchedule),
		thresholdSchemes:  make(map[string]*cryptoproto.ThresholdSignatureScheme),
		thresholdShares:   make(map[string]map[string]*cryptoproto.ThresholdSignatureShare),
		zkProofConfigs:    make(map[string]*cryptoproto.ZKProofConfig),
		secureEnclaves:    make(map[string]*cryptoproto.SecureEnclaveConfig),
		quantumKeys:       make(map[string]*cryptoproto.QuantumResistantKey),
		randomSources:     make(map[string]*cryptoproto.CryptoRandomSource),
		certificatePins:   make(map[string]*cryptoproto.CertificatePin),
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.logger.With("module", "x/"+types.ModuleName, "height", sdkCtx.BlockHeight())
}

// GetParams gets the module parameters
func (k Keeper) GetParams(ctx context.Context) (*cryptoproto.Params, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams(), nil
	}

	var params cryptoproto.Params
	k.cdc.MustUnmarshal(bz, &params)
	return &params, nil
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx context.Context, params *cryptoproto.Params) error {
	if err := types.ValidateParams(params); err != nil {
		return err
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(params)
	store.Set(types.ParamsKey, bz)
	return nil
}

// getStore returns the KVStore
func (k Keeper) getStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(k.storeKey)
}

// GenerateSecureRandomBytes generates cryptographically secure random bytes
func (k Keeper) GenerateSecureRandomBytes(length int) ([]byte, error) {
	if length < 1 {
		return nil, types.ErrInsufficientEntropy
	}

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, types.ErrRandomSourceFailed
	}

	return randomBytes, nil
}

// GenerateSecureRandomInt generates a cryptographically secure random integer
func (k Keeper) GenerateSecureRandomInt(max int64) (int64, error) {
	if max < 1 {
		return 0, types.ErrInsufficientEntropy
	}

	// Generate 8 random bytes
	randomBytes, err := k.GenerateSecureRandomBytes(8)
	if err != nil {
		return 0, err
	}

	// Convert to uint64
	randomInt := binary.BigEndian.Uint64(randomBytes)

	// Return value in range [0, max)
	return int64(randomInt % uint64(max)), nil
}

// ValidateAuthority checks if the signer is the module authority
func (k Keeper) ValidateAuthority(signer string) error {
	if signer != k.authority {
		return types.ErrUnauthorized
	}
	return nil
}

// GetAllZKProofConfigs retrieves all ZK proof configurations
func (k Keeper) GetAllZKProofConfigs(ctx context.Context) []*cryptoproto.ZKProofConfig {
	k.mu.RLock()
	defer k.mu.RUnlock()

	configs := make([]*cryptoproto.ZKProofConfig, 0, len(k.zkProofConfigs))
	for _, config := range k.zkProofConfigs {
		configs = append(configs, config)
	}
	return configs
}

// GetAllThresholdSchemes retrieves all threshold signature schemes
func (k Keeper) GetAllThresholdSchemes(ctx context.Context) []*cryptoproto.ThresholdSignatureScheme {
	k.mu.RLock()
	defer k.mu.RUnlock()

	schemes := make([]*cryptoproto.ThresholdSignatureScheme, 0, len(k.thresholdSchemes))
	for _, scheme := range k.thresholdSchemes {
		schemes = append(schemes, scheme)
	}
	return schemes
}

// GetAllKeyRotationSchedules retrieves all key rotation schedules

// DeleteQuantumResistantKey deletes a quantum-resistant key
func (k Keeper) DeleteQuantumResistantKey(ctx context.Context, keyID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	store := k.getStore(ctx)
	store.Delete(types.GetQuantumResistantKeyKey(keyID))

	delete(k.quantumKeys, keyID)

	k.Logger(ctx).Info("deleted quantum-resistant key", "key_id", keyID)
	return nil
}

// DeleteZKProofConfig deletes a ZK proof configuration
func (k Keeper) DeleteZKProofConfig(ctx context.Context, proofID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	store := k.getStore(ctx)
	store.Delete(types.GetZKProofConfigKey(proofID))

	delete(k.zkProofConfigs, proofID)

	k.Logger(ctx).Info("deleted ZK proof config", "proof_id", proofID)
	return nil
}
