// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/common/determinism"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	gogotypes "github.com/cosmos/gogoproto/types"
)

func (k Keeper) ConfigureSession(ctx context.Context, walletID string, timeoutDuration *gogotypes.Duration, autoLockEnabled bool, inactivityThresholdSeconds int32) (*wsproto.SessionConfig, error) {
	config := &wsproto.SessionConfig{
		SessionId:                  fmt.Sprintf("session_%s_%d", walletID, determinism.GetBlockTime(ctx).Unix()),
		WalletId:                   walletID,
		TimeoutDuration:            timeoutDuration,
		AutoLockEnabled:            autoLockEnabled,
		InactivityThresholdSeconds: inactivityThresholdSeconds,
		StartedAt:                  blockTimeToGogoTimestamp(ctx),
		Locked:                     false,
	}
	configBytes, err := k.cdc.Marshal(config)
	if err != nil {
		return nil, err
	}
	if err := k.SetSessionConfig(ctx, config.SessionId, configBytes); err != nil {
		return nil, err
	}
	return config, nil
}

func (k Keeper) LockSession(ctx context.Context, sessionID string) error {
	configBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get for SessionId: %w", err)
	}
	var config wsproto.SessionConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	config.Locked = true
	configBytes, err = k.cdc.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	return k.SetSessionConfig(ctx, sessionID, configBytes)
}

func (k Keeper) UnlockSession(ctx context.Context, sessionID string, authProof []byte) error {
	configBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	var config wsproto.SessionConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}
	if !config.Locked {
		return fmt.Errorf("session is not locked")
	}
	config.Locked = false
	config.LastActivity = blockTimeToGogoTimestamp(ctx)
	configBytes, err = k.cdc.Marshal(&config)
	if err != nil {
		return fmt.Errorf("error in UnlockSession: %w", err)
	}
	return k.SetSessionConfig(ctx, sessionID, configBytes)
}

// NOTE: Biometric authentication functions (EnrollBiometric, AuthenticateBiometric) have been
// removed as they are deprecated. See BIOMETRIC_DEPRECATION.md for details.
// True biometric authentication cannot work on blockchain due to determinism requirements.
// Use hardware wallet integration, multi-sig, or social recovery instead.

func (k Keeper) StoreInSecureEnclave(ctx context.Context, walletID string, enclaveType wsproto.EnclaveType, encryptedKeyMaterial []byte, attestationCertificate string) (*wsproto.SecureEnclaveConfig, error) {
	config := &wsproto.SecureEnclaveConfig{
		WalletId:               walletID,
		EnclaveId:              fmt.Sprintf("enclave_%s_%d", walletID, determinism.GetBlockTime(ctx).Unix()),
		EnclaveType:            enclaveType,
		EncryptedKeyMaterial:   encryptedKeyMaterial,
		CreatedAt:              blockTimeToGogoTimestamp(ctx),
		HardwareBacked:         true,
		AttestationCertificate: attestationCertificate,
	}
	configBytes, err := k.cdc.Marshal(config)
	if err != nil {
		return nil, err
	}
	if err := k.SetSecureEnclaveConfig(ctx, walletID, configBytes); err != nil {
		return nil, err
	}
	return config, nil
}

func (k Keeper) CreateEncryptedBackup(ctx context.Context, walletID string, encryptedSeed []byte, encryptionAlgorithm string, keyDerivationFunction string, salt []byte, iterations int32, location wsproto.BackupLocation) (*wsproto.EncryptedBackup, error) {
	backupID := fmt.Sprintf("backup_%s_%d", walletID, determinism.GetBlockTime(ctx).Unix())
	backup := &wsproto.EncryptedBackup{
		BackupId:              backupID,
		WalletId:              walletID,
		EncryptedSeed:         encryptedSeed,
		EncryptionAlgorithm:   encryptionAlgorithm,
		KeyDerivationFunction: keyDerivationFunction,
		Salt:                  salt,
		Iterations:            iterations,
		CreatedAt:             blockTimeToGogoTimestamp(ctx),
		Location:              location,
		Checksum:              fmt.Sprintf("%x", encryptedSeed),
	}
	backupBytes, err := k.cdc.Marshal(backup)
	if err != nil {
		return nil, err
	}
	if err := k.SetEncryptedBackup(ctx, backupID, backupBytes); err != nil {
		return nil, err
	}
	return backup, nil
}

func (k Keeper) ConfigureDustFilter(ctx context.Context, walletID string, enabled bool, minimumAmount string, maxDustTransactionsPerBlock int32, suspiciousPatternThreshold int32) (*wsproto.DustAttackFilter, error) {
	filter := &wsproto.DustAttackFilter{
		WalletId:                    walletID,
		Enabled:                     enabled,
		MinimumAmount:               minimumAmount,
		MaxDustTransactionsPerBlock: maxDustTransactionsPerBlock,
		SuspiciousPatternThreshold:  suspiciousPatternThreshold,
		LastUpdated:                 blockTimeToGogoTimestamp(ctx),
	}
	filterBytes, err := k.cdc.Marshal(filter)
	if err != nil {
		return nil, err
	}
	if err := k.SetDustFilter(ctx, walletID, filterBytes); err != nil {
		return nil, err
	}
	return filter, nil
}

func (k Keeper) ValidateAddressChecksum(ctx context.Context, address string, algorithm wsproto.ChecksumAlgorithm) (bool, string, error) {
	if len(address) == 0 {
		return false, "", fmt.Errorf("address cannot be empty")
	}
	checksum := fmt.Sprintf("%x", []byte(address))
	return true, checksum, nil
}
