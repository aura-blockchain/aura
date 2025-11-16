package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
)

// Keeper manages wallet security operations
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger
}

// NewKeeper creates a new wallet security keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		logger:       logger,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger
}

// getStore returns the KVStore for this module
func (k Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

// SetHardwareWallet stores a hardware wallet configuration
func (k Keeper) SetHardwareWallet(ctx context.Context, walletID string, config []byte) error {
	store := k.getStore(ctx)
	key := types.GetHardwareWalletKey(walletID)
	store.Set(key, config)
	return nil
}

// GetHardwareWallet retrieves a hardware wallet configuration
func (k Keeper) GetHardwareWallet(ctx context.Context, walletID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetHardwareWalletKey(walletID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrHardwareWalletNotFound
	}
	return value, nil
}

// SetMultiSigWallet stores a multi-sig wallet configuration
func (k Keeper) SetMultiSigWallet(ctx context.Context, walletID string, wallet []byte) error {
	store := k.getStore(ctx)
	key := types.GetMultiSigWalletKey(walletID)
	store.Set(key, wallet)
	return nil
}

// GetMultiSigWallet retrieves a multi-sig wallet configuration
func (k Keeper) GetMultiSigWallet(ctx context.Context, walletID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetMultiSigWalletKey(walletID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrMultiSigWalletNotFound
	}
	return value, nil
}

// SetPendingMultiSigTx stores a pending multi-sig transaction
func (k Keeper) SetPendingMultiSigTx(ctx context.Context, txID string, tx []byte) error {
	store := k.getStore(ctx)
	key := types.GetPendingMultiSigTxKey(txID)
	store.Set(key, tx)
	return nil
}

// GetPendingMultiSigTx retrieves a pending multi-sig transaction
func (k Keeper) GetPendingMultiSigTx(ctx context.Context, txID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetPendingMultiSigTxKey(txID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrMultiSigTxNotFound
	}
	return value, nil
}

// DeletePendingMultiSigTx removes a pending multi-sig transaction
func (k Keeper) DeletePendingMultiSigTx(ctx context.Context, txID string) error {
	store := k.getStore(ctx)
	key := types.GetPendingMultiSigTxKey(txID)
	store.Delete(key)
	return nil
}

// SetSocialRecoveryConfig stores social recovery configuration
func (k Keeper) SetSocialRecoveryConfig(ctx context.Context, walletID string, config []byte) error {
	store := k.getStore(ctx)
	key := types.GetSocialRecoveryKey(walletID)
	store.Set(key, config)
	return nil
}

// GetSocialRecoveryConfig retrieves social recovery configuration
func (k Keeper) GetSocialRecoveryConfig(ctx context.Context, walletID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetSocialRecoveryKey(walletID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrRecoveryNotEnabled
	}
	return value, nil
}

// SetRecoveryRequest stores a recovery request
func (k Keeper) SetRecoveryRequest(ctx context.Context, requestID string, request []byte) error {
	store := k.getStore(ctx)
	key := types.GetRecoveryRequestKey(requestID)
	store.Set(key, request)
	return nil
}

// GetRecoveryRequest retrieves a recovery request
func (k Keeper) GetRecoveryRequest(ctx context.Context, requestID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetRecoveryRequestKey(requestID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrRecoveryRequestNotFound
	}
	return value, nil
}

// SetSpendingLimit stores spending limit configuration
func (k Keeper) SetSpendingLimit(ctx context.Context, walletID, denom string, limit []byte) error {
	store := k.getStore(ctx)
	key := types.GetSpendingLimitKey(walletID, denom)
	store.Set(key, limit)
	return nil
}

// GetSpendingLimit retrieves spending limit configuration
func (k Keeper) GetSpendingLimit(ctx context.Context, walletID, denom string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetSpendingLimitKey(walletID, denom)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrSpendingLimitNotFound
	}
	return value, nil
}

// SetSessionConfig stores session configuration
func (k Keeper) SetSessionConfig(ctx context.Context, sessionID string, config []byte) error {
	store := k.getStore(ctx)
	key := types.GetSessionConfigKey(sessionID)
	store.Set(key, config)
	return nil
}

// GetSessionConfig retrieves session configuration
func (k Keeper) GetSessionConfig(ctx context.Context, sessionID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetSessionConfigKey(sessionID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrSessionNotFound
	}
	return value, nil
}

// SetBiometricAuth stores biometric authentication configuration
func (k Keeper) SetBiometricAuth(ctx context.Context, walletID string, auth []byte) error {
	store := k.getStore(ctx)
	key := types.GetBiometricAuthKey(walletID)
	store.Set(key, auth)
	return nil
}

// GetBiometricAuth retrieves biometric authentication configuration
func (k Keeper) GetBiometricAuth(ctx context.Context, walletID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetBiometricAuthKey(walletID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrBiometricNotEnrolled
	}
	return value, nil
}

// SetSecureEnclaveConfig stores secure enclave configuration
func (k Keeper) SetSecureEnclaveConfig(ctx context.Context, walletID string, config []byte) error {
	store := k.getStore(ctx)
	key := types.GetSecureEnclaveKey(walletID)
	store.Set(key, config)
	return nil
}

// GetSecureEnclaveConfig retrieves secure enclave configuration
func (k Keeper) GetSecureEnclaveConfig(ctx context.Context, walletID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetSecureEnclaveKey(walletID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrEnclaveNotAvailable
	}
	return value, nil
}

// SetEncryptedBackup stores encrypted backup
func (k Keeper) SetEncryptedBackup(ctx context.Context, backupID string, backup []byte) error {
	store := k.getStore(ctx)
	key := types.GetEncryptedBackupKey(backupID)
	store.Set(key, backup)
	return nil
}

// GetEncryptedBackup retrieves encrypted backup
func (k Keeper) GetEncryptedBackup(ctx context.Context, backupID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetEncryptedBackupKey(backupID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrBackupNotFound
	}
	return value, nil
}

// SetDustFilter stores dust filter configuration
func (k Keeper) SetDustFilter(ctx context.Context, walletID string, filter []byte) error {
	store := k.getStore(ctx)
	key := types.GetDustFilterKey(walletID)
	store.Set(key, filter)
	return nil
}

// GetDustFilter retrieves dust filter configuration
func (k Keeper) GetDustFilter(ctx context.Context, walletID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetDustFilterKey(walletID)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrDustFilterNotEnabled
	}
	return value, nil
}

// SetDomainVerification stores domain verification
func (k Keeper) SetDomainVerification(ctx context.Context, domain string, verification []byte) error {
	store := k.getStore(ctx)
	key := types.GetDomainVerificationKey(domain)
	store.Set(key, verification)
	return nil
}

// GetDomainVerification retrieves domain verification
func (k Keeper) GetDomainVerification(ctx context.Context, domain string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetDomainVerificationKey(domain)
	value := store.Get(key)
	if value == nil {
		return nil, types.ErrDomainNotVerified
	}
	return value, nil
}

// SetSecurityMetrics stores security metrics
func (k Keeper) SetSecurityMetrics(ctx context.Context, walletID string, metrics []byte) error {
	store := k.getStore(ctx)
	key := types.GetSecurityMetricsKey(walletID)
	store.Set(key, metrics)
	return nil
}

// GetSecurityMetrics retrieves security metrics
func (k Keeper) GetSecurityMetrics(ctx context.Context, walletID string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetSecurityMetricsKey(walletID)
	value := store.Get(key)
	if value == nil {
		// Return empty metrics if not found
		return nil, fmt.Errorf("security metrics not found for wallet %s", walletID)
	}
	return value, nil
}

// SetDustTransaction stores a dust transaction record
func (k Keeper) SetDustTransaction(ctx context.Context, txHash string, tx []byte) error {
	store := k.getStore(ctx)
	key := types.GetDustTransactionKey(txHash)
	store.Set(key, tx)
	return nil
}

// GetDustTransaction retrieves a dust transaction record
func (k Keeper) GetDustTransaction(ctx context.Context, txHash string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetDustTransactionKey(txHash)
	value := store.Get(key)
	if value == nil {
		return nil, fmt.Errorf("dust transaction not found: %s", txHash)
	}
	return value, nil
}
