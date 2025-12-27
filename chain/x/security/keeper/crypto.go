// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Cryptography Operations
// =============================================================================

// SetKeyRotationSchedule stores a key rotation schedule
func (k Keeper) SetKeyRotationSchedule(ctx sdk.Context, schedule *securitypb.KeyRotationSchedule) {
	store := k.GetStore(ctx)
	key := types.GetKeyRotationScheduleStoreKey(schedule.Id)
	bz := k.cdc.MustMarshal(schedule)
	store.Set(key, bz)
}

// GetKeyRotationSchedule retrieves a key rotation schedule
func (k Keeper) GetKeyRotationSchedule(ctx sdk.Context, id string) (*securitypb.KeyRotationSchedule, bool) {
	store := k.GetStore(ctx)
	key := types.GetKeyRotationScheduleStoreKey(id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var schedule securitypb.KeyRotationSchedule
	if err := k.cdc.Unmarshal(bz, &schedule); err != nil {
		k.Logger(ctx).Error("failed to unmarshal key rotation schedule", "error", err, "id", id)
		return nil, false
	}
	return &schedule, true
}

// GetAllKeyRotationSchedules returns all key rotation schedules
func (k Keeper) GetAllKeyRotationSchedules(ctx sdk.Context) []*securitypb.KeyRotationSchedule {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRotationScheduleKey)
	defer iterator.Close()

	var schedules []*securitypb.KeyRotationSchedule
	for ; iterator.Valid(); iterator.Next() {
		var schedule securitypb.KeyRotationSchedule
		if err := k.cdc.Unmarshal(iterator.Value(), &schedule); err != nil {
			k.Logger(ctx).Error("failed to unmarshal key rotation schedule during iteration", "error", err)
			continue
		}
		schedules = append(schedules, &schedule)
	}
	return schedules
}

// DeleteKeyRotationSchedule removes a key rotation schedule
func (k Keeper) DeleteKeyRotationSchedule(ctx sdk.Context, id string) {
	store := k.GetStore(ctx)
	key := types.GetKeyRotationScheduleStoreKey(id)
	store.Delete(key)
}

// SetThresholdScheme stores a threshold signature scheme
func (k Keeper) SetThresholdScheme(ctx sdk.Context, scheme *securitypb.ThresholdSignatureScheme) {
	store := k.GetStore(ctx)
	key := types.GetThresholdSchemeStoreKey(scheme.SchemeId)
	bz := k.cdc.MustMarshal(scheme)
	store.Set(key, bz)
}

// GetThresholdScheme retrieves a threshold scheme
func (k Keeper) GetThresholdScheme(ctx sdk.Context, id string) (*securitypb.ThresholdSignatureScheme, bool) {
	store := k.GetStore(ctx)
	key := types.GetThresholdSchemeStoreKey(id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var scheme securitypb.ThresholdSignatureScheme
	if err := k.cdc.Unmarshal(bz, &scheme); err != nil {
		k.Logger(ctx).Error("failed to unmarshal threshold scheme", "error", err, "id", id)
		return nil, false
	}
	return &scheme, true
}

// GetAllThresholdSchemes returns all threshold schemes
func (k Keeper) GetAllThresholdSchemes(ctx sdk.Context) []*securitypb.ThresholdSignatureScheme {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ThresholdSchemeKey)
	defer iterator.Close()

	var schemes []*securitypb.ThresholdSignatureScheme
	for ; iterator.Valid(); iterator.Next() {
		var scheme securitypb.ThresholdSignatureScheme
		if err := k.cdc.Unmarshal(iterator.Value(), &scheme); err != nil {
			k.Logger(ctx).Error("failed to unmarshal threshold scheme during iteration", "error", err)
			continue
		}
		schemes = append(schemes, &scheme)
	}
	return schemes
}

// SetZKProofConfig stores a ZK proof configuration
func (k Keeper) SetZKProofConfig(ctx sdk.Context, config *securitypb.ZKProofConfig) {
	store := k.GetStore(ctx)
	key := types.GetZKProofConfigStoreKey(config.ProofId)
	bz := k.cdc.MustMarshal(config)
	store.Set(key, bz)
}

// GetZKProofConfig retrieves a ZK proof configuration
func (k Keeper) GetZKProofConfig(ctx sdk.Context, id string) (*securitypb.ZKProofConfig, bool) {
	store := k.GetStore(ctx)
	key := types.GetZKProofConfigStoreKey(id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var config securitypb.ZKProofConfig
	if err := k.cdc.Unmarshal(bz, &config); err != nil {
		k.Logger(ctx).Error("failed to unmarshal ZK proof config", "error", err, "id", id)
		return nil, false
	}
	return &config, true
}

// GetAllZKProofConfigs returns all ZK proof configs
func (k Keeper) GetAllZKProofConfigs(ctx sdk.Context) []*securitypb.ZKProofConfig {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ZKProofConfigKey)
	defer iterator.Close()

	var configs []*securitypb.ZKProofConfig
	for ; iterator.Valid(); iterator.Next() {
		var config securitypb.ZKProofConfig
		if err := k.cdc.Unmarshal(iterator.Value(), &config); err != nil {
			k.Logger(ctx).Error("failed to unmarshal ZK proof config during iteration", "error", err)
			continue
		}
		configs = append(configs, &config)
	}
	return configs
}

// SetSecureEnclave stores a secure enclave configuration
func (k Keeper) SetSecureEnclave(ctx sdk.Context, enclave *types.SecureEnclave) {
	store := k.GetStore(ctx)
	key := types.GetSecureEnclaveStoreKey(enclave.Id)
	bz := k.cdc.MustMarshal(enclave)
	store.Set(key, bz)
}

// GetSecureEnclave retrieves a secure enclave
func (k Keeper) GetSecureEnclave(ctx sdk.Context, id string) (*types.SecureEnclave, bool) {
	store := k.GetStore(ctx)
	key := types.GetSecureEnclaveStoreKey(id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var enclave types.SecureEnclave
	if err := k.cdc.Unmarshal(bz, &enclave); err != nil {
		k.Logger(ctx).Error("failed to unmarshal secure enclave", "error", err, "id", id)
		return nil, false
	}
	return &enclave, true
}

// GetAllSecureEnclaves returns all secure enclaves
func (k Keeper) GetAllSecureEnclaves(ctx sdk.Context) []*types.SecureEnclave {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SecureEnclaveKey)
	defer iterator.Close()

	enclaves := make([]*types.SecureEnclave, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var enclave types.SecureEnclave
		if err := k.cdc.Unmarshal(iterator.Value(), &enclave); err != nil {
			k.Logger(ctx).Error("failed to unmarshal secure enclave during iteration", "error", err)
			continue
		}
		enclaves = append(enclaves, &enclave)
	}
	return enclaves
}

// SetQuantumResistantKey stores a quantum-resistant key
func (k Keeper) SetQuantumResistantKey(ctx sdk.Context, qrk *securitypb.QuantumResistantKey) {
	store := k.GetStore(ctx)
	key := types.GetQuantumResistantKeyStoreKey(qrk.KeyId)
	bz := k.cdc.MustMarshal(qrk)
	store.Set(key, bz)
}

// GetQuantumResistantKey retrieves a quantum-resistant key
func (k Keeper) GetQuantumResistantKey(ctx sdk.Context, id string) (*securitypb.QuantumResistantKey, bool) {
	store := k.GetStore(ctx)
	key := types.GetQuantumResistantKeyStoreKey(id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var qrk securitypb.QuantumResistantKey
	if err := k.cdc.Unmarshal(bz, &qrk); err != nil {
		k.Logger(ctx).Error("failed to unmarshal quantum resistant key", "error", err, "id", id)
		return nil, false
	}
	return &qrk, true
}

// GetAllQuantumResistantKeys returns all quantum-resistant keys
func (k Keeper) GetAllQuantumResistantKeys(ctx sdk.Context) []*securitypb.QuantumResistantKey {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.QuantumResistantKeyPrefix)
	defer iterator.Close()

	var keys []*securitypb.QuantumResistantKey
	for ; iterator.Valid(); iterator.Next() {
		var qrk securitypb.QuantumResistantKey
		if err := k.cdc.Unmarshal(iterator.Value(), &qrk); err != nil {
			k.Logger(ctx).Error("failed to unmarshal quantum resistant key during iteration", "error", err)
			continue
		}
		keys = append(keys, &qrk)
	}
	return keys
}

// SetRandomSource stores a random source
func (k Keeper) SetRandomSource(ctx sdk.Context, source *types.RandomSource) {
	store := k.GetStore(ctx)
	key := types.GetRandomSourceStoreKey(source.Id)
	bz := k.cdc.MustMarshal(source)
	store.Set(key, bz)
}

// GetRandomSource retrieves a random source
func (k Keeper) GetRandomSource(ctx sdk.Context, id string) (*types.RandomSource, bool) {
	store := k.GetStore(ctx)
	key := types.GetRandomSourceStoreKey(id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var source types.RandomSource
	if err := k.cdc.Unmarshal(bz, &source); err != nil {
		k.Logger(ctx).Error("failed to unmarshal random source", "error", err, "id", id)
		return nil, false
	}
	return &source, true
}

// GetAllRandomSources returns all random sources
func (k Keeper) GetAllRandomSources(ctx sdk.Context) []*types.RandomSource {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.RandomSourceKey)
	defer iterator.Close()

	sources := make([]*types.RandomSource, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var source types.RandomSource
		if err := k.cdc.Unmarshal(iterator.Value(), &source); err != nil {
			k.Logger(ctx).Error("failed to unmarshal random source during iteration", "error", err)
			continue
		}
		sources = append(sources, &source)
	}
	return sources
}

// SetCertificatePin stores a certificate pin
func (k Keeper) SetCertificatePin(ctx sdk.Context, pin *types.CertificatePin) {
	store := k.GetStore(ctx)
	key := types.GetCertificatePinStoreKey(pin.Id)
	bz := k.cdc.MustMarshal(pin)
	store.Set(key, bz)
}

// GetCertificatePin retrieves a certificate pin
func (k Keeper) GetCertificatePin(ctx sdk.Context, id string) (*types.CertificatePin, bool) {
	store := k.GetStore(ctx)
	key := types.GetCertificatePinStoreKey(id)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var pin types.CertificatePin
	if err := k.cdc.Unmarshal(bz, &pin); err != nil {
		k.Logger(ctx).Error("failed to unmarshal certificate pin", "error", err, "id", id)
		return nil, false
	}
	return &pin, true
}

// GetAllCertificatePins returns all certificate pins
func (k Keeper) GetAllCertificatePins(ctx sdk.Context) []*types.CertificatePin {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.CertificatePinKey)
	defer iterator.Close()

	pins := make([]*types.CertificatePin, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var pin types.CertificatePin
		if err := k.cdc.Unmarshal(iterator.Value(), &pin); err != nil {
			k.Logger(ctx).Error("failed to unmarshal certificate pin during iteration", "error", err)
			continue
		}
		pins = append(pins, &pin)
	}
	return pins
}

// DeleteCertificatePin removes a certificate pin
func (k Keeper) DeleteCertificatePin(ctx sdk.Context, id string) {
	store := k.GetStore(ctx)
	key := types.GetCertificatePinStoreKey(id)
	store.Delete(key)
}

// MaxKeyRotationsPerBlock limits the number of key rotations processed per block
// to prevent chain halt under high load (P1 performance fix).
const MaxKeyRotationsPerBlock = 50

// CheckKeyRotationDue checks if any keys need rotation.
// Returns at most MaxKeyRotationsPerBlock schedules to prevent chain halt.
func (k Keeper) CheckKeyRotationDue(ctx sdk.Context) []*securitypb.KeyRotationSchedule {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.KeyRotationScheduleKey)
	defer iterator.Close()

	var dueForRotation []*securitypb.KeyRotationSchedule
	blockTime := ctx.BlockTime()

	// Iterate with early exit when batch limit is reached
	for ; iterator.Valid() && len(dueForRotation) < MaxKeyRotationsPerBlock; iterator.Next() {
		var schedule securitypb.KeyRotationSchedule
		if err := k.cdc.Unmarshal(iterator.Value(), &schedule); err != nil {
			continue
		}
		if schedule.Enabled && !schedule.NextRotationTime.IsZero() {
			if blockTime.After(schedule.NextRotationTime) {
				dueForRotation = append(dueForRotation, &schedule)
			}
		}
	}

	return dueForRotation
}

// ValidateCertificatePin validates a certificate against pinned value
func (k Keeper) ValidateCertificatePin(ctx sdk.Context, domain, certHash string) error {
	pins := k.GetAllCertificatePins(ctx)
	blockTime := ctx.BlockTime()

	for _, pin := range pins {
		if pin.Domain == domain && pin.IsActive {
			// Check validity period
			if pin.ValidFrom != nil && blockTime.Before(*pin.ValidFrom) {
				return types.ErrCertificateInvalid
			}
			if pin.ValidUntil != nil && blockTime.After(*pin.ValidUntil) {
				return types.ErrCertificateInvalid
			}
			// Check hash match
			if pin.PinHash != certHash {
				return types.ErrCertificateInvalid
			}
			return nil
		}
	}
	// No pin found for domain - allow by default
	return nil
}

// RotateKey performs key rotation for a schedule
func (k Keeper) RotateKey(ctx sdk.Context, scheduleId string) error {
	schedule, found := k.GetKeyRotationSchedule(ctx, scheduleId)
	if !found {
		return types.ErrKeyRotationNotFound
	}

	if !schedule.Enabled {
		return types.ErrKeyRotationDisabled
	}

	// Update last rotation time
	now := ctx.BlockTime()
	schedule.LastRotation = &now

	// Calculate next rotation time
	nextRotation := now.Add(time.Duration(schedule.RotationIntervalSeconds) * time.Second)
	schedule.NextRotationTime = nextRotation

	// Save updated schedule
	k.SetKeyRotationSchedule(ctx, schedule)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKeyRotated,
			sdk.NewAttribute(types.AttributeKeyScheduleId, scheduleId),
			sdk.NewAttribute(types.AttributeKeyRotationTime, now.String()),
		),
	)

	return nil
}
