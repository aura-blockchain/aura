package keeper

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// Keeper maintains the state for the cryptography module
// All state is persisted in the KV store - NO in-memory caching
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	logger    log.Logger
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
		cdc:       cdc,
		storeKey:  storeKey,
		logger:    logger,
		authority: authority,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// getStore returns the KVStore
func (k Keeper) getStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(k.storeKey)
}

// GetParams gets the module parameters
func (k Keeper) GetParams(ctx context.Context) (*cryptoproto.Params, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		params := types.DefaultParams()
		return &params, nil
	}

	var params cryptoproto.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}
	return &params, nil
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx context.Context, params *cryptoproto.Params) error {
	if err := types.ValidateParams(params); err != nil {
		return fmt.Errorf("error in SetParams for ValidateParams: %w", err)
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(params)
	store.Set(types.ParamsKey, bz)
	return nil
}

// UpdateParams updates the module parameters
func (k Keeper) UpdateParams(ctx context.Context, authority string, params *cryptoproto.Params) error {
	if err := k.ValidateAuthority(authority); err != nil {
		return fmt.Errorf("error in UpdateParams for ValidateAuthority: %w", err)
	}
	return k.SetParams(ctx, params)
}

// ValidateAuthority checks if the signer is the module authority
func (k Keeper) ValidateAuthority(signer string) error {
	if signer != k.authority {
		return types.ErrUnauthorized
	}
	return nil
}

// ============================================================================
// Quantum Resistant Keys - KV Store Operations
// ============================================================================

// GetQuantumResistantKey retrieves a quantum-resistant key from KV store
func (k Keeper) GetQuantumResistantKey(ctx context.Context, keyID string) (*cryptoproto.QuantumResistantKey, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetQuantumResistantKeyKey(keyID))
	if bz == nil {
		return nil, types.ErrQuantumKeyNotFound
	}

	var key cryptoproto.QuantumResistantKey
	if err := k.cdc.Unmarshal(bz, &key); err != nil {
		return nil, types.ErrInvalidState
	}

	// Check expiration
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	if key.ExpiresAt != nil && !key.ExpiresAt.IsZero() && key.ExpiresAt.Before(blockTime) {
		return nil, types.ErrKeyExpired
	}

	return &key, nil
}

// SetQuantumResistantKey stores a quantum-resistant key
func (k Keeper) SetQuantumResistantKey(ctx context.Context, key *cryptoproto.QuantumResistantKey) error {
	if key == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(key)
	store.Set(types.GetQuantumResistantKeyKey(key.KeyId), bz)
	return nil
}

// DeleteQuantumResistantKey deletes a quantum-resistant key
func (k Keeper) DeleteQuantumResistantKey(ctx context.Context, keyID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetQuantumResistantKeyKey(keyID))

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("deleted quantum-resistant key", "key_id", keyID)
	return nil
}

// IterateQuantumKeys iterates over all quantum-resistant keys
func (k Keeper) IterateQuantumKeys(ctx context.Context, fn func(key *cryptoproto.QuantumResistantKey) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.QuantumResistantKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var key cryptoproto.QuantumResistantKey
		if err := k.cdc.Unmarshal(iterator.Value(), &key); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&key) {
			break
		}
	}

	return nil
}

// ============================================================================
// Certificate Pins - KV Store Operations
// ============================================================================

// GetCertificatePin retrieves a certificate pin from KV store
func (k Keeper) GetCertificatePin(ctx context.Context, hostname string) (*cryptoproto.CertificatePin, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetCertificatePinKey(hostname))
	if bz == nil {
		return nil, types.ErrCertificatePinNotFound
	}

	var pin cryptoproto.CertificatePin
	if err := k.cdc.Unmarshal(bz, &pin); err != nil {
		return nil, types.ErrInvalidState
	}
	return &pin, nil
}

// SetCertificatePin stores a certificate pin
func (k Keeper) SetCertificatePin(ctx context.Context, pin *cryptoproto.CertificatePin) error {
	if pin == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(pin)
	store.Set(types.GetCertificatePinKey(pin.Hostname), bz)
	return nil
}

// DeleteCertificatePin deletes a certificate pin
func (k Keeper) DeleteCertificatePin(ctx context.Context, hostname string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetCertificatePinKey(hostname))
	return nil
}

// IterateCertificatePins iterates over all certificate pins
func (k Keeper) IterateCertificatePins(ctx context.Context, fn func(pin *cryptoproto.CertificatePin) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.CertificatePinPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pin cryptoproto.CertificatePin
		if err := k.cdc.Unmarshal(iterator.Value(), &pin); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&pin) {
			break
		}
	}

	return nil
}

// ListCertificatePins returns all certificate pins
func (k Keeper) ListCertificatePins(ctx context.Context) []*cryptoproto.CertificatePin {
	pins := make([]*cryptoproto.CertificatePin, 0)
	_ = k.IterateCertificatePins(ctx, func(pin *cryptoproto.CertificatePin) bool {
		pins = append(pins, pin)
		return false
	})
	return pins
}

// ============================================================================
// Key Rotation Schedules - KV Store Operations
// ============================================================================

// GetKeyRotationSchedule retrieves a key rotation schedule from KV store
func (k Keeper) GetKeyRotationSchedule(ctx context.Context, scheduleID string) (*cryptoproto.KeyRotationSchedule, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetKeyRotationScheduleKey(scheduleID))
	if bz == nil {
		return nil, types.ErrKeyRotationScheduleNotFound
	}

	var schedule cryptoproto.KeyRotationSchedule
	if err := k.cdc.Unmarshal(bz, &schedule); err != nil {
		return nil, types.ErrInvalidState
	}
	return &schedule, nil
}

// SetKeyRotationSchedule stores a key rotation schedule
func (k Keeper) SetKeyRotationSchedule(ctx context.Context, schedule *cryptoproto.KeyRotationSchedule) error {
	if schedule == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(schedule)
	store.Set(types.GetKeyRotationScheduleKey(schedule.Id), bz)
	return nil
}

// DeleteKeyRotationSchedule deletes a key rotation schedule
func (k Keeper) DeleteKeyRotationSchedule(ctx context.Context, scheduleID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetKeyRotationScheduleKey(scheduleID))
	return nil
}

// IterateKeyRotationSchedules iterates over all key rotation schedules
func (k Keeper) IterateKeyRotationSchedules(ctx context.Context, fn func(schedule *cryptoproto.KeyRotationSchedule) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRotationSchedulePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var schedule cryptoproto.KeyRotationSchedule
		if err := k.cdc.Unmarshal(iterator.Value(), &schedule); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&schedule) {
			break
		}
	}

	return nil
}

// GetAllKeyRotationSchedules retrieves all key rotation schedules
func (k Keeper) GetAllKeyRotationSchedules(ctx context.Context) []*cryptoproto.KeyRotationSchedule {
	schedules := make([]*cryptoproto.KeyRotationSchedule, 0)
	_ = k.IterateKeyRotationSchedules(ctx, func(schedule *cryptoproto.KeyRotationSchedule) bool {
		schedules = append(schedules, schedule)
		return false
	})
	return schedules
}

// GetSchedulesForKey returns all rotation schedules for a given key
func (k Keeper) GetSchedulesForKey(ctx context.Context, keyID string) []*cryptoproto.KeyRotationSchedule {
	schedules := make([]*cryptoproto.KeyRotationSchedule, 0)
	_ = k.IterateKeyRotationSchedules(ctx, func(schedule *cryptoproto.KeyRotationSchedule) bool {
		if schedule.KeyId == keyID {
			schedules = append(schedules, schedule)
		}
		return false
	})
	return schedules
}

// ============================================================================
// Key Stretching Configs - KV Store Operations
// ============================================================================

// GetKeyStretchingConfig retrieves a key stretching config from KV store
func (k Keeper) GetKeyStretchingConfig(ctx context.Context, configID string) (*cryptoproto.KeyStretchingConfig, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetKeyStretchingConfigKey(configID))
	if bz == nil {
		return nil, types.ErrInvalidKeyStretchingConfig
	}

	var config cryptoproto.KeyStretchingConfig
	if err := k.cdc.Unmarshal(bz, &config); err != nil {
		return nil, types.ErrInvalidState
	}
	return &config, nil
}

// SetKeyStretchingConfig and DeleteKeyStretchingConfig are in key_stretching.go

// IterateKeyStretchingConfigs iterates over all key stretching configs
func (k Keeper) IterateKeyStretchingConfigs(ctx context.Context, fn func(config *cryptoproto.KeyStretchingConfig) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyStretchingConfigPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var config cryptoproto.KeyStretchingConfig
		if err := k.cdc.Unmarshal(iterator.Value(), &config); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&config) {
			break
		}
	}

	return nil
}

// ============================================================================
// Secure Enclaves - KV Store Operations
// ============================================================================

// GetSecureEnclave retrieves a secure enclave from KV store
func (k Keeper) GetSecureEnclave(ctx context.Context, enclaveID string) (*cryptoproto.SecureEnclaveConfig, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetSecureEnclaveKey(enclaveID))
	if bz == nil {
		return nil, types.ErrSecureEnclaveNotFound
	}

	var enclave cryptoproto.SecureEnclaveConfig
	if err := k.cdc.Unmarshal(bz, &enclave); err != nil {
		return nil, types.ErrInvalidState
	}
	return &enclave, nil
}

// SetSecureEnclaveConfig stores a secure enclave config
func (k Keeper) SetSecureEnclaveConfig(ctx context.Context, config *cryptoproto.SecureEnclaveConfig) error {
	if config == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(config)
	store.Set(types.GetSecureEnclaveKey(config.EnclaveId), bz)
	return nil
}

// DeleteSecureEnclave deletes a secure enclave
func (k Keeper) DeleteSecureEnclave(ctx context.Context, enclaveID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetSecureEnclaveKey(enclaveID))
	return nil
}

// IterateSecureEnclaves iterates over all secure enclaves
func (k Keeper) IterateSecureEnclaves(ctx context.Context, fn func(enclave *cryptoproto.SecureEnclaveConfig) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SecureEnclavePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var enclave cryptoproto.SecureEnclaveConfig
		if err := k.cdc.Unmarshal(iterator.Value(), &enclave); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&enclave) {
			break
		}
	}

	return nil
}

// ListSecureEnclaves returns all registered secure enclaves
func (k Keeper) ListSecureEnclaves(ctx context.Context) []*cryptoproto.SecureEnclaveConfig {
	enclaves := make([]*cryptoproto.SecureEnclaveConfig, 0)
	_ = k.IterateSecureEnclaves(ctx, func(enclave *cryptoproto.SecureEnclaveConfig) bool {
		enclaves = append(enclaves, enclave)
		return false
	})
	return enclaves
}

// ============================================================================
// Random Sources - KV Store Operations
// ============================================================================

// GetRandomSource retrieves a random source from KV store
func (k Keeper) GetRandomSource(ctx context.Context, sourceID string) (*cryptoproto.CryptoRandomSource, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetCryptoRandomSourceKey(sourceID))
	if bz == nil {
		return nil, types.ErrRandomSourceFailed
	}

	var source cryptoproto.CryptoRandomSource
	if err := k.cdc.Unmarshal(bz, &source); err != nil {
		return nil, types.ErrInvalidState
	}
	return &source, nil
}

// SetRandomSource stores a random source
func (k Keeper) SetRandomSource(ctx context.Context, source *cryptoproto.CryptoRandomSource) error {
	if source == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(source)
	store.Set(types.GetCryptoRandomSourceKey(source.SourceId), bz)
	return nil
}

// DeleteRandomSource deletes a random source
func (k Keeper) DeleteRandomSource(ctx context.Context, sourceID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetCryptoRandomSourceKey(sourceID))
	return nil
}

// IterateRandomSources iterates over all random sources
func (k Keeper) IterateRandomSources(ctx context.Context, fn func(source *cryptoproto.CryptoRandomSource) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.CryptoRandomSourcePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var source cryptoproto.CryptoRandomSource
		if err := k.cdc.Unmarshal(iterator.Value(), &source); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&source) {
			break
		}
	}

	return nil
}

// GetRandomSourceStatus returns the status of all random sources
func (k Keeper) GetRandomSourceStatus(ctx context.Context) []*cryptoproto.CryptoRandomSource {
	sources := make([]*cryptoproto.CryptoRandomSource, 0)
	_ = k.IterateRandomSources(ctx, func(source *cryptoproto.CryptoRandomSource) bool {
		sources = append(sources, source)
		return false
	})
	return sources
}

// ============================================================================
// Threshold Signature Schemes - KV Store Operations
// ============================================================================

// GetThresholdScheme retrieves a threshold signature scheme from KV store
func (k Keeper) GetThresholdScheme(ctx context.Context, schemeID string) (*cryptoproto.ThresholdSignatureScheme, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetThresholdSchemeKey(schemeID))
	if bz == nil {
		return nil, types.ErrThresholdSchemeNotFound
	}

	var scheme cryptoproto.ThresholdSignatureScheme
	if err := k.cdc.Unmarshal(bz, &scheme); err != nil {
		return nil, types.ErrInvalidState
	}
	return &scheme, nil
}

// SetThresholdScheme stores a threshold signature scheme
func (k Keeper) SetThresholdScheme(ctx context.Context, scheme *cryptoproto.ThresholdSignatureScheme) error {
	if scheme == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(scheme)
	store.Set(types.GetThresholdSchemeKey(scheme.SchemeId), bz)
	return nil
}

// DeleteThresholdScheme deletes a threshold signature scheme
func (k Keeper) DeleteThresholdScheme(ctx context.Context, schemeID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetThresholdSchemeKey(schemeID))
	return nil
}

// IterateThresholdSchemes iterates over all threshold signature schemes
func (k Keeper) IterateThresholdSchemes(ctx context.Context, fn func(scheme *cryptoproto.ThresholdSignatureScheme) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ThresholdSchemePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var scheme cryptoproto.ThresholdSignatureScheme
		if err := k.cdc.Unmarshal(iterator.Value(), &scheme); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&scheme) {
			break
		}
	}

	return nil
}

// GetAllThresholdSchemes retrieves all threshold signature schemes
func (k Keeper) GetAllThresholdSchemes(ctx context.Context) []*cryptoproto.ThresholdSignatureScheme {
	schemes := make([]*cryptoproto.ThresholdSignatureScheme, 0)
	_ = k.IterateThresholdSchemes(ctx, func(scheme *cryptoproto.ThresholdSignatureScheme) bool {
		schemes = append(schemes, scheme)
		return false
	})
	return schemes
}

// ============================================================================
// Threshold Signature Shares - KV Store Operations
// ============================================================================

// GetThresholdSignatureShare retrieves a threshold signature share from KV store
func (k Keeper) GetThresholdSignatureShare(ctx context.Context, schemeID, participantID string) (*cryptoproto.ThresholdSignatureShare, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetThresholdSignatureShareKey(schemeID, participantID))
	if bz == nil {
		return nil, types.ErrInvalidSignatureShare
	}

	var share cryptoproto.ThresholdSignatureShare
	if err := k.cdc.Unmarshal(bz, &share); err != nil {
		return nil, types.ErrInvalidState
	}
	return &share, nil
}

// SetThresholdSignatureShare stores a threshold signature share
func (k Keeper) SetThresholdSignatureShare(ctx context.Context, share *cryptoproto.ThresholdSignatureShare) error {
	if share == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(share)
	store.Set(types.GetThresholdSignatureShareKey(share.SchemeId, share.ParticipantId), bz)
	return nil
}

// DeleteThresholdSignatureShare deletes a threshold signature share
func (k Keeper) DeleteThresholdSignatureShare(ctx context.Context, schemeID, participantID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetThresholdSignatureShareKey(schemeID, participantID))
	return nil
}

// GetThresholdSignatureSharesForScheme retrieves all shares for a scheme
func (k Keeper) GetThresholdSignatureSharesForScheme(ctx context.Context, schemeID string, messageHash []byte) []*cryptoproto.ThresholdSignatureShare {
	shares := make([]*cryptoproto.ThresholdSignatureShare, 0)
	store := k.getStore(ctx)

	// Create prefix for this scheme's shares
	prefix := append(types.ThresholdSignatureSharePrefix, []byte(schemeID)...)
	prefix = append(prefix, []byte("/")...)

	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var share cryptoproto.ThresholdSignatureShare
		if err := k.cdc.Unmarshal(iterator.Value(), &share); err != nil {
			// Skip invalid entries
			continue
		}

		// Filter by message hash if provided
		if messageHash == nil || string(share.MessageHash) == string(messageHash) {
			shares = append(shares, &share)
		}
	}

	return shares
}

// ============================================================================
// ZK Proof Configs - KV Store Operations (implemented in zk_proofs.go)
// ============================================================================
// Note: GetZKProofConfig, SetZKProofConfig, and other ZK proof methods
// are implemented in zk_proofs.go

// ============================================================================
// ZK Proofs - KV Store Operations
// ============================================================================

// GetZKProof retrieves a ZK proof from KV store
func (k Keeper) GetZKProof(ctx context.Context, proofID string) (*cryptoproto.ZKProof, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetZKProofKey(proofID))
	if bz == nil {
		return nil, types.ErrInvalidZKProof
	}

	var proof cryptoproto.ZKProof
	if err := k.cdc.Unmarshal(bz, &proof); err != nil {
		return nil, types.ErrInvalidState
	}
	return &proof, nil
}

// SetZKProof stores a ZK proof
func (k Keeper) SetZKProof(ctx context.Context, proof *cryptoproto.ZKProof) error {
	if proof == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(proof)
	store.Set(types.GetZKProofKey(proof.ProofId), bz)
	return nil
}

// DeleteZKProof deletes a ZK proof
func (k Keeper) DeleteZKProof(ctx context.Context, proofID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetZKProofKey(proofID))
	return nil
}

// IterateZKProofs iterates over all ZK proofs
func (k Keeper) IterateZKProofs(ctx context.Context, fn func(proof *cryptoproto.ZKProof) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ZKProofPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proof cryptoproto.ZKProof
		if err := k.cdc.Unmarshal(iterator.Value(), &proof); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&proof) {
			break
		}
	}

	return nil
}

// ============================================================================
// ZK Proof Verifications - KV Store Operations
// ============================================================================

// GetZKProofVerification retrieves a ZK proof verification from KV store
func (k Keeper) GetZKProofVerification(ctx context.Context, verificationID string) (*cryptoproto.ZKProofVerification, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetZKProofVerificationKey(verificationID))
	if bz == nil {
		return nil, types.ErrInvalidZKProof
	}

	var verification cryptoproto.ZKProofVerification
	if err := k.cdc.Unmarshal(bz, &verification); err != nil {
		return nil, types.ErrInvalidState
	}
	return &verification, nil
}

// SetZKProofVerification stores a ZK proof verification
func (k Keeper) SetZKProofVerification(ctx context.Context, verification *cryptoproto.ZKProofVerification) error {
	if verification == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(verification)
	store.Set(types.GetZKProofVerificationKey(verification.VerificationId), bz)
	return nil
}

// DeleteZKProofVerification deletes a ZK proof verification
func (k Keeper) DeleteZKProofVerification(ctx context.Context, verificationID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetZKProofVerificationKey(verificationID))
	return nil
}

// IterateZKProofVerifications iterates over all ZK proof verifications
func (k Keeper) IterateZKProofVerifications(ctx context.Context, fn func(verification *cryptoproto.ZKProofVerification) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ZKProofVerificationPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var verification cryptoproto.ZKProofVerification
		if err := k.cdc.Unmarshal(iterator.Value(), &verification); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&verification) {
			break
		}
	}

	return nil
}

// GetAllZKProofVerifications retrieves all verifications for a proof
func (k Keeper) GetAllZKProofVerifications(ctx context.Context, proofID string) []*cryptoproto.ZKProofVerification {
	verifications := make([]*cryptoproto.ZKProofVerification, 0)
	_ = k.IterateZKProofVerifications(ctx, func(verification *cryptoproto.ZKProofVerification) bool {
		if proofID == "" || verification.ProofId == proofID {
			verifications = append(verifications, verification)
		}
		return false
	})
	return verifications
}

// ============================================================================
// Salted Hashes - KV Store Operations
// ============================================================================

// GetSaltedHash retrieves a salted hash from KV store
func (k Keeper) GetSaltedHash(ctx context.Context, hashID string) (*cryptoproto.SaltedHash, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetSaltedHashKey(hashID))
	if bz == nil {
		return nil, types.ErrInvalidHashAlgorithm
	}

	var hash cryptoproto.SaltedHash
	if err := k.cdc.Unmarshal(bz, &hash); err != nil {
		return nil, types.ErrInvalidState
	}
	return &hash, nil
}

// SetSaltedHash stores a salted hash
func (k Keeper) SetSaltedHash(ctx context.Context, hash *cryptoproto.SaltedHash) error {
	if hash == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(hash)
	store.Set(types.GetSaltedHashKey(hash.HashId), bz)
	return nil
}

// DeleteSaltedHash deletes a salted hash
func (k Keeper) DeleteSaltedHash(ctx context.Context, hashID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetSaltedHashKey(hashID))
	return nil
}

// IterateSaltedHashes iterates over all salted hashes
func (k Keeper) IterateSaltedHashes(ctx context.Context, fn func(hash *cryptoproto.SaltedHash) bool) error {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SaltedHashPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var hash cryptoproto.SaltedHash
		if err := k.cdc.Unmarshal(iterator.Value(), &hash); err != nil {
			// Skip invalid entries
			continue
		}
		if fn(&hash) {
			break
		}
	}

	return nil
}

// ============================================================================
// HD Key Derivations - KV Store Operations
// ============================================================================

// GetHDKeyDerivation retrieves an HD key derivation from KV store
func (k Keeper) GetHDKeyDerivation(ctx context.Context, masterKeyID string) (*cryptoproto.HDKeyDerivation, error) {
	store := k.getStore(ctx)
	bz := store.Get(types.GetHDKeyDerivationKey(masterKeyID))
	if bz == nil {
		return nil, types.ErrHDKeyDerivationFailed
	}

	var derivation cryptoproto.HDKeyDerivation
	if err := k.cdc.Unmarshal(bz, &derivation); err != nil {
		return nil, types.ErrInvalidState
	}
	return &derivation, nil
}

// SetHDKeyDerivation stores an HD key derivation
func (k Keeper) SetHDKeyDerivation(ctx context.Context, derivation *cryptoproto.HDKeyDerivation) error {
	if derivation == nil {
		return nil
	}

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(derivation)
	store.Set(types.GetHDKeyDerivationKey(derivation.MasterKeyId), bz)
	return nil
}

// DeleteHDKeyDerivation deletes an HD key derivation
func (k Keeper) DeleteHDKeyDerivation(ctx context.Context, masterKeyID string) error {
	store := k.getStore(ctx)
	store.Delete(types.GetHDKeyDerivationKey(masterKeyID))
	return nil
}

// ============================================================================
// Utility Functions
// ============================================================================

// GenerateSecureRandomBytes generates cryptographically secure random bytes.
//
// WARNING: This function uses crypto/rand which is NON-DETERMINISTIC and will break
// consensus if called from a message handler (MsgServer method).
//
// DO NOT call this from message handlers. This is for client-side utilities only.
//
// For on-chain operations that need randomness:
// - Require the client to provide randomness in the message (generated off-chain)
// - Use deterministic sources like block hash, tx hash, or VRF if available
func (k Keeper) GenerateSecureRandomBytes(length int) ([]byte, error) {
	if length < 1 {
		return nil, types.ErrInsufficientEntropy
	}

	// WARNING: Non-deterministic - see function documentation
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

// GenerateRandomUint64 is in random.go
// CompareHashes is in hashing.go
