package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Wallet Security Operations
// =============================================================================

// SetHardwareWallet stores a hardware wallet registration
func (k Keeper) SetHardwareWallet(ctx sdk.Context, hw *securitypb.HardwareWalletConfig) {
	store := k.GetStore(ctx)
	key := append(types.HardwareWalletKey, []byte(hw.WalletId)...)
	bz := k.cdc.MustMarshal(hw)
	store.Set(key, bz)
}

// GetHardwareWallet retrieves a hardware wallet registration
func (k Keeper) GetHardwareWallet(ctx sdk.Context, walletId string) (*securitypb.HardwareWalletConfig, bool) {
	store := k.GetStore(ctx)
	key := append(types.HardwareWalletKey, []byte(walletId)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var hw securitypb.HardwareWalletConfig
	k.cdc.MustUnmarshal(bz, &hw)
	return &hw, true
}

// GetAllHardwareWallets returns all hardware wallets
func (k Keeper) GetAllHardwareWallets(ctx sdk.Context) []*securitypb.HardwareWalletConfig {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.HardwareWalletKey)
	defer iterator.Close()

	var wallets []*securitypb.HardwareWalletConfig
	for ; iterator.Valid(); iterator.Next() {
		var hw securitypb.HardwareWalletConfig
		k.cdc.MustUnmarshal(iterator.Value(), &hw)
		wallets = append(wallets, &hw)
	}
	return wallets
}

// SetMultiSigWallet stores a multisig wallet configuration
func (k Keeper) SetMultiSigWallet(ctx sdk.Context, ms *securitypb.MultiSigWallet) {
	store := k.GetStore(ctx)
	key := append(types.MultiSigWalletKey, []byte(ms.WalletId)...)
	bz := k.cdc.MustMarshal(ms)
	store.Set(key, bz)
}

// GetMultiSigWallet retrieves a multisig wallet configuration
func (k Keeper) GetMultiSigWallet(ctx sdk.Context, walletId string) (*securitypb.MultiSigWallet, bool) {
	store := k.GetStore(ctx)
	key := append(types.MultiSigWalletKey, []byte(walletId)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var ms securitypb.MultiSigWallet
	k.cdc.MustUnmarshal(bz, &ms)
	return &ms, true
}

// GetAllMultiSigWallets returns all multisig wallets
func (k Keeper) GetAllMultiSigWallets(ctx sdk.Context) []*securitypb.MultiSigWallet {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.MultiSigWalletKey)
	defer iterator.Close()

	var wallets []*securitypb.MultiSigWallet
	for ; iterator.Valid(); iterator.Next() {
		var ms securitypb.MultiSigWallet
		k.cdc.MustUnmarshal(iterator.Value(), &ms)
		wallets = append(wallets, &ms)
	}
	return wallets
}

// SetPendingMultiSigTx stores a pending multisig transaction
func (k Keeper) SetPendingMultiSigTx(ctx sdk.Context, tx *securitypb.PendingMultiSigTransaction) {
	store := k.GetStore(ctx)
	key := append(types.PendingMultiSigTxKey, []byte(tx.TxId)...)
	bz := k.cdc.MustMarshal(tx)
	store.Set(key, bz)
}

// GetPendingMultiSigTx retrieves a pending multisig transaction
func (k Keeper) GetPendingMultiSigTx(ctx sdk.Context, id string) (*securitypb.PendingMultiSigTransaction, bool) {
	store := k.GetStore(ctx)
	key := append(types.PendingMultiSigTxKey, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var tx securitypb.PendingMultiSigTransaction
	k.cdc.MustUnmarshal(bz, &tx)
	return &tx, true
}

// GetAllPendingMultiSigTxs returns all pending multisig transactions
func (k Keeper) GetAllPendingMultiSigTxs(ctx sdk.Context) []*securitypb.PendingMultiSigTransaction {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.PendingMultiSigTxKey)
	defer iterator.Close()

	var txs []*securitypb.PendingMultiSigTransaction
	for ; iterator.Valid(); iterator.Next() {
		var tx securitypb.PendingMultiSigTransaction
		k.cdc.MustUnmarshal(iterator.Value(), &tx)
		txs = append(txs, &tx)
	}
	return txs
}

// DeletePendingMultiSigTx removes a pending multisig transaction
func (k Keeper) DeletePendingMultiSigTx(ctx sdk.Context, id string) {
	store := k.GetStore(ctx)
	key := append(types.PendingMultiSigTxKey, []byte(id)...)
	store.Delete(key)
}

// SetSocialRecoveryConfig stores a social recovery configuration
func (k Keeper) SetSocialRecoveryConfig(ctx sdk.Context, config *securitypb.SocialRecoveryConfig) {
	store := k.GetStore(ctx)
	key := append(types.SocialRecoveryKey, []byte(config.WalletId)...)
	bz := k.cdc.MustMarshal(config)
	store.Set(key, bz)
}

// GetSocialRecoveryConfig retrieves a social recovery configuration
func (k Keeper) GetSocialRecoveryConfig(ctx sdk.Context, walletId string) (*securitypb.SocialRecoveryConfig, bool) {
	store := k.GetStore(ctx)
	key := append(types.SocialRecoveryKey, []byte(walletId)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var config securitypb.SocialRecoveryConfig
	k.cdc.MustUnmarshal(bz, &config)
	return &config, true
}

// GetAllSocialRecoveryConfigs returns all social recovery configs
func (k Keeper) GetAllSocialRecoveryConfigs(ctx sdk.Context) []*securitypb.SocialRecoveryConfig {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SocialRecoveryKey)
	defer iterator.Close()

	var configs []*securitypb.SocialRecoveryConfig
	for ; iterator.Valid(); iterator.Next() {
		var config securitypb.SocialRecoveryConfig
		k.cdc.MustUnmarshal(iterator.Value(), &config)
		configs = append(configs, &config)
	}
	return configs
}

// SetRecoveryRequest stores a recovery request
func (k Keeper) SetRecoveryRequest(ctx sdk.Context, request *securitypb.RecoveryRequest) {
	store := k.GetStore(ctx)
	key := append(types.RecoveryRequestKey, []byte(request.RequestId)...)
	bz := k.cdc.MustMarshal(request)
	store.Set(key, bz)
}

// GetRecoveryRequest retrieves a recovery request
func (k Keeper) GetRecoveryRequest(ctx sdk.Context, id string) (*securitypb.RecoveryRequest, bool) {
	store := k.GetStore(ctx)
	key := append(types.RecoveryRequestKey, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var request securitypb.RecoveryRequest
	k.cdc.MustUnmarshal(bz, &request)
	return &request, true
}

// GetAllRecoveryRequests returns all recovery requests
func (k Keeper) GetAllRecoveryRequests(ctx sdk.Context) []*securitypb.RecoveryRequest {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.RecoveryRequestKey)
	defer iterator.Close()

	var requests []*securitypb.RecoveryRequest
	for ; iterator.Valid(); iterator.Next() {
		var request securitypb.RecoveryRequest
		k.cdc.MustUnmarshal(iterator.Value(), &request)
		requests = append(requests, &request)
	}
	return requests
}

// SetDeviceFingerprint stores a device fingerprint
func (k Keeper) SetDeviceFingerprint(ctx sdk.Context, fp *types.DeviceFingerprint) {
	store := k.GetStore(ctx)
	key := append(types.DeviceFingerprintKey, []byte(fp.Id)...)
	bz := k.cdc.MustMarshal(fp)
	store.Set(key, bz)
}

// GetAllDeviceFingerprints returns all device fingerprints
func (k Keeper) GetAllDeviceFingerprints(ctx sdk.Context) []*types.DeviceFingerprint {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.DeviceFingerprintKey)
	defer iterator.Close()

	var fps []*types.DeviceFingerprint
	for ; iterator.Valid(); iterator.Next() {
		var fp types.DeviceFingerprint
		k.cdc.MustUnmarshal(iterator.Value(), &fp)
		fps = append(fps, &fp)
	}
	return fps
}

// SetSession stores a session
func (k Keeper) SetSession(ctx sdk.Context, session *types.WalletSession) {
	store := k.GetStore(ctx)
	key := append(types.SessionKey, []byte(session.Id)...)
	bz := k.cdc.MustMarshal(session)
	store.Set(key, bz)
}

// GetSession retrieves a session
func (k Keeper) GetSession(ctx sdk.Context, id string) (*types.WalletSession, bool) {
	store := k.GetStore(ctx)
	key := append(types.SessionKey, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var session types.WalletSession
	k.cdc.MustUnmarshal(bz, &session)
	return &session, true
}

// GetAllSessions returns all sessions
func (k Keeper) GetAllSessions(ctx sdk.Context) []*types.WalletSession {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SessionKey)
	defer iterator.Close()

	var sessions []*types.WalletSession
	for ; iterator.Valid(); iterator.Next() {
		var session types.WalletSession
		k.cdc.MustUnmarshal(iterator.Value(), &session)
		sessions = append(sessions, &session)
	}
	return sessions
}

// DeleteSession removes a session
func (k Keeper) DeleteSession(ctx sdk.Context, id string) {
	store := k.GetStore(ctx)
	key := append(types.SessionKey, []byte(id)...)
	store.Delete(key)
}

// ValidateSession validates a session
func (k Keeper) ValidateSession(ctx sdk.Context, sessionID string) error {
	session, found := k.GetSession(ctx, sessionID)
	if !found {
		return types.ErrInvalidSession
	}
	if session.ExpiresAt != nil && ctx.BlockTime().After(*session.ExpiresAt) {
		k.DeleteSession(ctx, sessionID)
		return types.ErrInvalidSession
	}
	return nil
}

// SetAnomalyDetection stores an anomaly detection record
func (k Keeper) SetAnomalyDetection(ctx sdk.Context, anomaly *types.AnomalyDetection) {
	store := k.GetStore(ctx)
	key := append(types.AnomalyDetectionKey, []byte(anomaly.Id)...)
	bz := k.cdc.MustMarshal(anomaly)
	store.Set(key, bz)
}

// GetAllAnomalyDetections returns all anomaly detections
func (k Keeper) GetAllAnomalyDetections(ctx sdk.Context) []*types.AnomalyDetection {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.AnomalyDetectionKey)
	defer iterator.Close()

	var anomalies []*types.AnomalyDetection
	for ; iterator.Valid(); iterator.Next() {
		var anomaly types.AnomalyDetection
		k.cdc.MustUnmarshal(iterator.Value(), &anomaly)
		anomalies = append(anomalies, &anomaly)
	}
	return anomalies
}

// IsMultiSigWallet checks if an address is a multisig wallet
func (k Keeper) IsMultiSigWallet(ctx sdk.Context, addr string) bool {
	_, found := k.GetMultiSigWallet(ctx, addr)
	return found
}

// ValidateMultiSigTx validates a multisig transaction has enough signatures
func (k Keeper) ValidateMultiSigTx(ctx sdk.Context, txID string) error {
	tx, found := k.GetPendingMultiSigTx(ctx, txID)
	if !found {
		return types.ErrNotFound
	}
	if int32(len(tx.SignedBy)) < tx.CurrentWeight {
		return types.ErrInsufficientSigs
	}
	if tx.ExpiresAt != nil && ctx.BlockTime().After(tx.ExpiresAt.AsTime()) {
		return types.ErrInvalidRequest
	}
	return nil
}
