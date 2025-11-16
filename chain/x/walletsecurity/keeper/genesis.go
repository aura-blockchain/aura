package keeper

import (
	"context"

	wsecproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// InitGenesis initializes the walletsecurity module state from genesis
func (k Keeper) InitGenesis(ctx context.Context, data *wsecproto.GenesisState) error {
	if data == nil {
		return nil
	}

	// Set parameters
	if data.Params != nil {
		if err := k.SetParams(ctx, data.Params); err != nil {
			k.Logger(ctx).Error("failed to set params", "error", err)
			return err
		}
	}

	// Initialize hardware wallets
	for _, wallet := range data.HardwareWallets {
		if err := k.SetHardwareWalletConfig(ctx, wallet); err != nil {
			k.Logger(ctx).Error("failed to initialize hardware wallet", "wallet_id", wallet.WalletId, "error", err)
		}
	}

	// Initialize multisig wallets
	for _, wallet := range data.MultisigWallets {
		if err := k.SetMultiSigWalletConfig(ctx, wallet); err != nil {
			k.Logger(ctx).Error("failed to initialize multisig wallet", "wallet_id", wallet.WalletId, "error", err)
		}
	}

	// Initialize pending multisig transactions
	for _, tx := range data.PendingTransactions {
		if err := k.SetPendingMultiSigTransaction(ctx, tx); err != nil {
			k.Logger(ctx).Error("failed to initialize pending transaction", "tx_id", tx.TxId, "error", err)
		}
	}

	// Initialize social recovery configs
	for _, config := range data.RecoveryConfigs {
		if err := k.SetSocialRecoveryConfigProto(ctx, config); err != nil {
			k.Logger(ctx).Error("failed to initialize recovery config", "wallet_id", config.WalletId, "error", err)
		}
	}

	// Initialize recovery requests
	for _, request := range data.RecoveryRequests {
		if err := k.SetRecoveryRequestProto(ctx, request); err != nil {
			k.Logger(ctx).Error("failed to initialize recovery request", "request_id", request.RequestId, "error", err)
		}
	}

	// Initialize domain verifications
	for _, verification := range data.DomainVerifications {
		if err := k.SetDomainVerification(ctx, verification); err != nil {
			k.Logger(ctx).Error("failed to initialize domain verification", "domain", verification.Domain, "error", err)
		}
	}

	// Initialize phishing protection configs
	for _, config := range data.PhishingConfigs {
		if err := k.SetPhishingProtectionConfig(ctx, config); err != nil {
			k.Logger(ctx).Error("failed to initialize phishing config", "wallet_id", config.WalletId, "error", err)
		}
	}

	// Initialize spending limits
	for _, limit := range data.SpendingLimits {
		if err := k.SetSpendingLimitProto(ctx, limit); err != nil {
			k.Logger(ctx).Error("failed to initialize spending limit", "wallet_id", limit.WalletId, "error", err)
		}
	}

	// Initialize session configs
	for _, config := range data.SessionConfigs {
		if err := k.SetSessionConfig(ctx, config); err != nil {
			k.Logger(ctx).Error("failed to initialize session config", "wallet_id", config.WalletId, "error", err)
		}
	}

	// Initialize biometric auth configs
	for _, config := range data.BiometricConfigs {
		if err := k.SetBiometricAuth(ctx, config); err != nil {
			k.Logger(ctx).Error("failed to initialize biometric config", "wallet_id", config.WalletId, "error", err)
		}
	}

	// Initialize secure enclave configs
	for _, config := range data.EnclaveConfigs {
		if err := k.SetSecureEnclaveConfigProto(ctx, config); err != nil {
			k.Logger(ctx).Error("failed to initialize enclave config", "wallet_id", config.WalletId, "error", err)
		}
	}

	// Initialize encrypted backups
	for _, backup := range data.EncryptedBackups {
		if err := k.SetEncryptedBackup(ctx, backup); err != nil {
			k.Logger(ctx).Error("failed to initialize encrypted backup", "backup_id", backup.BackupId, "error", err)
		}
	}

	// Initialize dust filters
	for _, filter := range data.DustFilters {
		if err := k.SetDustAttackFilter(ctx, filter); err != nil {
			k.Logger(ctx).Error("failed to initialize dust filter", "wallet_id", filter.WalletId, "error", err)
		}
	}

	// Initialize dust transactions
	for _, tx := range data.DustTransactions {
		if err := k.SetDustTransaction(ctx, tx); err != nil {
			k.Logger(ctx).Error("failed to initialize dust transaction", "tx_hash", tx.TxHash, "error", err)
		}
	}

	// Initialize security metrics
	for _, metrics := range data.SecurityMetrics {
		if err := k.SetWalletSecurityMetrics(ctx, metrics); err != nil {
			k.Logger(ctx).Error("failed to initialize security metrics", "wallet_id", metrics.WalletId, "error", err)
		}
	}

	k.Logger(ctx).Info("walletsecurity module initialized from genesis")
	return nil
}

// ExportGenesis exports the walletsecurity module state to genesis
func (k Keeper) ExportGenesis(ctx context.Context) *wsecproto.GenesisState {
	// Get parameters
	params, err := k.GetParams(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to get params during export", "error", err)
		params = &wsecproto.WalletSecurityParams{}
	}

	// Export hardware wallets
	hardwareWallets := k.GetAllHardwareWallets(ctx)

	// Export multisig wallets
	multisigWallets := k.GetAllMultiSigWallets(ctx)

	// Export pending transactions
	pendingTransactions := k.GetAllPendingMultiSigTransactions(ctx)

	// Export recovery configs
	recoveryConfigs := k.GetAllSocialRecoveryConfigs(ctx)

	// Export recovery requests
	recoveryRequests := k.GetAllRecoveryRequests(ctx)

	// Export domain verifications
	domainVerifications := k.GetAllDomainVerifications(ctx)

	// Export phishing configs
	phishingConfigs := k.GetAllPhishingProtectionConfigs(ctx)

	// Export spending limits
	spendingLimits := k.GetAllSpendingLimits(ctx)

	// Export session configs
	sessionConfigs := k.GetAllSessionConfigs(ctx)

	// Export biometric configs
	biometricConfigs := k.GetAllBiometricAuths(ctx)

	// Export enclave configs
	enclaveConfigs := k.GetAllSecureEnclaveConfigs(ctx)

	// Export encrypted backups
	encryptedBackups := k.GetAllEncryptedBackups(ctx)

	// Export dust filters
	dustFilters := k.GetAllDustAttackFilters(ctx)

	// Export dust transactions
	dustTransactions := k.GetAllDustTransactions(ctx)

	// Export security metrics
	securityMetrics := k.GetAllWalletSecurityMetrics(ctx)

	return &wsecproto.GenesisState{
		Params:              params,
		HardwareWallets:     hardwareWallets,
		MultisigWallets:     multisigWallets,
		PendingTransactions: pendingTransactions,
		RecoveryConfigs:     recoveryConfigs,
		RecoveryRequests:    recoveryRequests,
		DomainVerifications: domainVerifications,
		PhishingConfigs:     phishingConfigs,
		SpendingLimits:      spendingLimits,
		SessionConfigs:      sessionConfigs,
		BiometricConfigs:    biometricConfigs,
		EnclaveConfigs:      enclaveConfigs,
		EncryptedBackups:    encryptedBackups,
		DustFilters:         dustFilters,
		DustTransactions:    dustTransactions,
		SecurityMetrics:     securityMetrics,
	}
}
