package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	pb "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// InitGenesis initializes the module state from genesis data
func (k Keeper) InitGenesis(ctx context.Context, data *pb.GenesisState) error {
	if data == nil {
		return fmt.Errorf("genesis state cannot be nil")
	}

	store := k.getStore(ctx)

	// Set params (Params field is non-nullable in proto)
	paramsBytes, err := k.cdc.Marshal(&data.Params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	store.Set(types.ParamsKey, paramsBytes)

	// Import hardware wallet configs
	for _, hw := range data.HardwareWallets {
		if hw == nil || hw.WalletId == "" {
			continue
		}
		hwBytes, err := k.cdc.Marshal(hw)
		if err != nil {
			return fmt.Errorf("failed to marshal hardware wallet: %w", err)
		}
		key := types.GetHardwareWalletKey(hw.WalletId)
		store.Set(key, hwBytes)
	}

	// Import multi-sig wallets
	for _, ms := range data.MultisigWallets {
		if ms == nil || ms.WalletId == "" {
			continue
		}
		msBytes, err := k.cdc.Marshal(ms)
		if err != nil {
			return fmt.Errorf("failed to marshal multisig wallet: %w", err)
		}
		key := types.GetMultiSigWalletKey(ms.WalletId)
		store.Set(key, msBytes)
	}

	// Import pending multi-sig transactions
	for _, tx := range data.PendingTransactions {
		if tx == nil || tx.TxId == "" {
			continue
		}
		txBytes, err := k.cdc.Marshal(tx)
		if err != nil {
			return fmt.Errorf("failed to marshal pending tx: %w", err)
		}
		key := types.GetPendingMultiSigTxKey(tx.TxId)
		store.Set(key, txBytes)
	}

	// Import social recovery configs
	for _, sr := range data.RecoveryConfigs {
		if sr == nil || sr.WalletId == "" {
			continue
		}
		srBytes, err := k.cdc.Marshal(sr)
		if err != nil {
			return fmt.Errorf("failed to marshal social recovery: %w", err)
		}
		key := types.GetSocialRecoveryKey(sr.WalletId)
		store.Set(key, srBytes)
	}

	// Import recovery requests
	for _, req := range data.RecoveryRequests {
		if req == nil || req.RequestId == "" {
			continue
		}
		reqBytes, err := k.cdc.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal recovery request: %w", err)
		}
		key := types.GetRecoveryRequestKey(req.RequestId)
		store.Set(key, reqBytes)
	}

	// Import domain verifications
	for _, dv := range data.DomainVerifications {
		if dv == nil || dv.Domain == "" {
			continue
		}
		dvBytes, err := k.cdc.Marshal(dv)
		if err != nil {
			return fmt.Errorf("failed to marshal domain verification: %w", err)
		}
		key := types.GetDomainVerificationKey(dv.Domain)
		store.Set(key, dvBytes)
	}

	// Import spending limits
	for _, sl := range data.SpendingLimits {
		if sl == nil || sl.WalletId == "" {
			continue
		}
		slBytes, err := k.cdc.Marshal(sl)
		if err != nil {
			return fmt.Errorf("failed to marshal spending limit: %w", err)
		}
		key := types.GetSpendingLimitKey(sl.WalletId, sl.Denom)
		store.Set(key, slBytes)
	}

	// Import session configs
	for _, sc := range data.SessionConfigs {
		if sc == nil || sc.SessionId == "" {
			continue
		}
		scBytes, err := k.cdc.Marshal(sc)
		if err != nil {
			return fmt.Errorf("failed to marshal session config: %w", err)
		}
		key := types.GetSessionKey(sc.SessionId)
		store.Set(key, scBytes)
	}

	// Import biometric auth configs
	for _, ba := range data.BiometricConfigs {
		if ba == nil || ba.WalletId == "" {
			continue
		}
		baBytes, err := k.cdc.Marshal(ba)
		if err != nil {
			return fmt.Errorf("failed to marshal biometric auth: %w", err)
		}
		key := types.GetBiometricAuthKey(ba.WalletId)
		store.Set(key, baBytes)
	}

	// Import secure enclave configs
	for _, ec := range data.EnclaveConfigs {
		if ec == nil || ec.WalletId == "" {
			continue
		}
		ecBytes, err := k.cdc.Marshal(ec)
		if err != nil {
			return fmt.Errorf("failed to marshal enclave config: %w", err)
		}
		key := types.GetSecureEnclaveKey(ec.WalletId)
		store.Set(key, ecBytes)
	}

	// Import encrypted backups
	for _, eb := range data.EncryptedBackups {
		if eb == nil || eb.BackupId == "" {
			continue
		}
		ebBytes, err := k.cdc.Marshal(eb)
		if err != nil {
			return fmt.Errorf("failed to marshal encrypted backup: %w", err)
		}
		key := types.GetEncryptedBackupKey(eb.BackupId)
		store.Set(key, ebBytes)
	}

	// Import dust filters
	for _, df := range data.DustFilters {
		if df == nil || df.WalletId == "" {
			continue
		}
		dfBytes, err := k.cdc.Marshal(df)
		if err != nil {
			return fmt.Errorf("failed to marshal dust filter: %w", err)
		}
		key := types.GetDustFilterKey(df.WalletId)
		store.Set(key, dfBytes)
	}

	// Import dust transactions
	for _, dt := range data.DustTransactions {
		if dt == nil || dt.TxHash == "" {
			continue
		}
		dtBytes, err := k.cdc.Marshal(dt)
		if err != nil {
			return fmt.Errorf("failed to marshal dust transaction: %w", err)
		}
		key := types.GetDustTransactionKey(dt.TxHash)
		store.Set(key, dtBytes)
	}

	// Import security metrics
	for _, sm := range data.SecurityMetrics {
		if sm == nil || sm.WalletId == "" {
			continue
		}
		smBytes, err := k.cdc.Marshal(sm)
		if err != nil {
			return fmt.Errorf("failed to marshal security metrics: %w", err)
		}
		key := types.GetSecurityMetricsKey(sm.WalletId)
		store.Set(key, smBytes)
	}

	k.logger.Info("Wallet security genesis imported",
		"hardware_wallets", len(data.HardwareWallets),
		"multisig_wallets", len(data.MultisigWallets),
		"session_configs", len(data.SessionConfigs))

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k Keeper) ExportGenesis(ctx context.Context) *pb.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)

	// Export params
	var params pb.WalletSecurityParams
	paramsBytes, _ := store.Get(types.ParamsKey)
	if paramsBytes != nil {
		if err := k.cdc.Unmarshal(paramsBytes, &params); err != nil {
			k.logger.Error("failed to unmarshal params", "error", err)
			// Continue with default params on error
		}
	}

	genesis := &pb.GenesisState{
		Params:              params,
		HardwareWallets:     []*pb.HardwareWalletConfig{},
		MultisigWallets:     []*pb.MultiSigWallet{},
		PendingTransactions: []*pb.PendingMultiSigTransaction{},
		RecoveryConfigs:     []*pb.SocialRecoveryConfig{},
		RecoveryRequests:    []*pb.RecoveryRequest{},
		DomainVerifications: []*pb.DomainVerification{},
		PhishingConfigs:     []*pb.PhishingProtectionConfig{},
		SpendingLimits:      []*pb.SpendingLimit{},
		SessionConfigs:      []*pb.SessionConfig{},
		BiometricConfigs:    []*pb.BiometricAuth{},
		EnclaveConfigs:      []*pb.SecureEnclaveConfig{},
		EncryptedBackups:    []*pb.EncryptedBackup{},
		DustFilters:         []*pb.DustAttackFilter{},
		DustTransactions:    []*pb.DustTransaction{},
		SecurityMetrics:     []*pb.WalletSecurityMetrics{},
	}

	// Get the KVStore for iteration
	kvStore := k.storeService.OpenKVStore(ctx)

	// Export all hardware wallets
	iter := storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.HardwareWalletPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var hw pb.HardwareWalletConfig
		if err := k.cdc.Unmarshal(iter.Value(), &hw); err == nil {
			genesis.HardwareWallets = append(genesis.HardwareWallets, &hw)
		}
	}

	// Export all multi-sig wallets
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.MultiSigWalletPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var ms pb.MultiSigWallet
		if err := k.cdc.Unmarshal(iter.Value(), &ms); err == nil {
			genesis.MultisigWallets = append(genesis.MultisigWallets, &ms)
		}
	}

	// Export all pending multi-sig txs
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.PendingMultiSigTxPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var tx pb.PendingMultiSigTransaction
		if err := k.cdc.Unmarshal(iter.Value(), &tx); err == nil {
			genesis.PendingTransactions = append(genesis.PendingTransactions, &tx)
		}
	}

	// Export all social recovery configs
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.SocialRecoveryPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var sr pb.SocialRecoveryConfig
		if err := k.cdc.Unmarshal(iter.Value(), &sr); err == nil {
			genesis.RecoveryConfigs = append(genesis.RecoveryConfigs, &sr)
		}
	}

	// Export all recovery requests
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.RecoveryRequestPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var req pb.RecoveryRequest
		if err := k.cdc.Unmarshal(iter.Value(), &req); err == nil {
			genesis.RecoveryRequests = append(genesis.RecoveryRequests, &req)
		}
	}

	// Export all domain verifications
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.DomainVerificationPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var dv pb.DomainVerification
		if err := k.cdc.Unmarshal(iter.Value(), &dv); err == nil {
			genesis.DomainVerifications = append(genesis.DomainVerifications, &dv)
		}
	}

	// Export all spending limits
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.SpendingLimitPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var sl pb.SpendingLimit
		if err := k.cdc.Unmarshal(iter.Value(), &sl); err == nil {
			genesis.SpendingLimits = append(genesis.SpendingLimits, &sl)
		}
	}

	// Export all session configs
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.SessionPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var sc pb.SessionConfig
		if err := k.cdc.Unmarshal(iter.Value(), &sc); err == nil {
			genesis.SessionConfigs = append(genesis.SessionConfigs, &sc)
		}
	}

	// Export all biometric auth configs
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.BiometricAuthPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var ba pb.BiometricAuth
		if err := k.cdc.Unmarshal(iter.Value(), &ba); err == nil {
			genesis.BiometricConfigs = append(genesis.BiometricConfigs, &ba)
		}
	}

	// Export all secure enclave configs
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.SecureEnclavePrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var ec pb.SecureEnclaveConfig
		if err := k.cdc.Unmarshal(iter.Value(), &ec); err == nil {
			genesis.EnclaveConfigs = append(genesis.EnclaveConfigs, &ec)
		}
	}

	// Export all encrypted backups
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.EncryptedBackupPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var eb pb.EncryptedBackup
		if err := k.cdc.Unmarshal(iter.Value(), &eb); err == nil {
			genesis.EncryptedBackups = append(genesis.EncryptedBackups, &eb)
		}
	}

	// Export all dust filters
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.DustFilterPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var df pb.DustAttackFilter
		if err := k.cdc.Unmarshal(iter.Value(), &df); err == nil {
			genesis.DustFilters = append(genesis.DustFilters, &df)
		}
	}

	// Export all dust transactions
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.DustTransactionPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var dt pb.DustTransaction
		if err := k.cdc.Unmarshal(iter.Value(), &dt); err == nil {
			genesis.DustTransactions = append(genesis.DustTransactions, &dt)
		}
	}

	// Export all security metrics
	iter = storetypes.KVStorePrefixIterator(sdkCtx.KVStore(kvStore.(storetypes.StoreKey)), types.SecurityMetricsPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var sm pb.WalletSecurityMetrics
		if err := k.cdc.Unmarshal(iter.Value(), &sm); err == nil {
			genesis.SecurityMetrics = append(genesis.SecurityMetrics, &sm)
		}
	}

	return genesis
}
