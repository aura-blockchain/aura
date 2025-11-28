package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/common/determinism"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (k Keeper) ConfigureSession(ctx context.Context, walletID string, timeoutDuration *durationpb.Duration, autoLockEnabled bool, inactivityThresholdSeconds int32) (*wsproto.SessionConfig, error) {
	config := &wsproto.SessionConfig{
		SessionId:                  fmt.Sprintf("session_%s_%d", walletID, determinism.GetBlockTime(ctx).Unix()),
		WalletId:                   walletID,
		TimeoutDuration:            timeoutDuration,
		AutoLockEnabled:            autoLockEnabled,
		InactivityThresholdSeconds: inactivityThresholdSeconds,
		StartedAt:                  timestamppb.New(determinism.GetBlockTime(ctx)),
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
		return err
	}
	var config wsproto.SessionConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return err
	}
	config.Locked = true
	configBytes, err = k.cdc.Marshal(&config)
	if err != nil {
		return err
	}
	return k.SetSessionConfig(ctx, sessionID, configBytes)
}

func (k Keeper) UnlockSession(ctx context.Context, sessionID string, authProof []byte) error {
	configBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return err
	}
	var config wsproto.SessionConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return err
	}
	if !config.Locked {
		return fmt.Errorf("session is not locked")
	}
	config.Locked = false
	config.LastActivity = timestamppb.New(determinism.GetBlockTime(ctx))
	configBytes, err = k.cdc.Marshal(&config)
	if err != nil {
		return err
	}
	return k.SetSessionConfig(ctx, sessionID, configBytes)
}

func (k Keeper) EnrollBiometric(ctx context.Context, walletID string, bioType wsproto.BiometricType, enrollmentData []byte) (*wsproto.BiometricAuth, error) {
	auth := &wsproto.BiometricAuth{
		WalletId:       walletID,
		Type:           bioType,
		EnrolledAt:     timestamppb.New(determinism.GetBlockTime(ctx)),
		Enabled:        true,
		EnrollmentHash: fmt.Sprintf("%x", enrollmentData),
	}
	authBytes, err := k.cdc.Marshal(auth)
	if err != nil {
		return nil, err
	}
	if err := k.SetBiometricAuth(ctx, walletID, authBytes); err != nil {
		return nil, err
	}
	return auth, nil
}

func (k Keeper) AuthenticateBiometric(ctx context.Context, walletID string, biometricProof []byte) (bool, error) {
	authBytes, err := k.GetBiometricAuth(ctx, walletID)
	if err != nil {
		return false, err
	}
	var auth wsproto.BiometricAuth
	if err := k.cdc.Unmarshal(authBytes, &auth); err != nil {
		return false, err
	}
	if !auth.Enabled {
		return false, fmt.Errorf("biometric authentication is disabled")
	}
	if auth.LockedOut {
		return false, fmt.Errorf("biometric authentication is locked out")
	}

	// Compare the provided biometric proof with enrolled hash
	proofHash := fmt.Sprintf("%x", biometricProof)
	authenticated := proofHash == auth.EnrollmentHash

	auth.LastAttempt = timestamppb.New(determinism.GetBlockTime(ctx))
	if !authenticated {
		auth.FailedAttempts++
		// Lock out after 5 failed attempts
		if auth.FailedAttempts >= 5 {
			auth.LockedOut = true
			// Set lockout for 30 minutes
			auth.LockoutUntil = timestamppb.New(determinism.GetBlockTime(ctx).Add(30 * time.Minute))
		}
	} else {
		auth.FailedAttempts = 0
		auth.LockedOut = false
		auth.LockoutUntil = nil
	}
	authBytes, err = k.cdc.Marshal(&auth)
	if err != nil {
		return false, err
	}
	k.SetBiometricAuth(ctx, walletID, authBytes)
	return authenticated, nil
}

func (k Keeper) StoreInSecureEnclave(ctx context.Context, walletID string, enclaveType wsproto.EnclaveType, encryptedKeyMaterial []byte, attestationCertificate string) (*wsproto.SecureEnclaveConfig, error) {
	config := &wsproto.SecureEnclaveConfig{
		WalletId:               walletID,
		EnclaveId:              fmt.Sprintf("enclave_%s_%d", walletID, determinism.GetBlockTime(ctx).Unix()),
		EnclaveType:            enclaveType,
		EncryptedKeyMaterial:   encryptedKeyMaterial,
		CreatedAt:              timestamppb.New(determinism.GetBlockTime(ctx)),
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
		CreatedAt:             timestamppb.New(determinism.GetBlockTime(ctx)),
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
		LastUpdated:                 timestamppb.New(determinism.GetBlockTime(ctx)),
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
