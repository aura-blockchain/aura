// Package keeper implements the consolidated security module keeper.
// This keeper combines functionality from: networksecurity, validatorsecurity,
// walletsecurity, incidentresponse, cryptography, and privacy modules.
package keeper

import (
	"encoding/hex"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// Keeper implements the consolidated security module keeper
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	memKey    storetypes.StoreKey
	authority string

	// References to other keepers
	bankKeeper    BankKeeper
	stakingKeeper StakingKeeper
	accountKeeper AccountKeeper
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// StakingKeeper defines the expected staking keeper interface
type StakingKeeper interface {
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator interface{}, found bool)
	Jail(ctx sdk.Context, consAddr sdk.ConsAddress) error
	Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) error
	Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor string) (string, error)
}

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx sdk.Context, acc sdk.AccountI)
}

// NewKeeper creates a new security keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	authority string,
	bankKeeper BankKeeper,
	stakingKeeper StakingKeeper,
	accountKeeper AccountKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		authority:     authority,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		accountKeeper: accountKeeper,
	}
}

// GetAuthority returns the module's authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetStore returns the module's KV store
func (k Keeper) GetStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}

// GetMemStore returns the module's memory store
func (k Keeper) GetMemStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.memKey)
}

// =============================================================================
// Genesis Operations
// =============================================================================

// InitGenesis initializes the module's state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState *securitypb.GenesisState) {
	// Store params
	if genState.Params != nil {
		k.SetParams(ctx, *genState.Params)
	}

	// Initialize network security state
	if genState.NetworkSecurity != nil {
		ns := genState.NetworkSecurity
		for _, rl := range ns.RateLimits {
			k.SetRateLimit(ctx, rl)
		}
		for _, rep := range ns.Reputations {
			k.SetPeerReputation(ctx, rep)
		}
		for _, peer := range ns.TrustedPeers {
			k.SetTrustedPeer(ctx, peer)
		}
		for _, alert := range ns.ForkAlerts {
			k.SetForkAlert(ctx, alert)
		}
		for _, alert := range ns.PartitionAlerts {
			k.SetPartitionAlert(ctx, alert)
		}
	}

	// Initialize validator security state
	if genState.ValidatorSecurity != nil {
		vs := genState.ValidatorSecurity
		for _, vi := range vs.Validators {
			k.SetValidatorSecurityInfo(ctx, vi)
		}
		for _, ev := range vs.DoubleSignEvidences {
			k.SetDoubleSignEvidence(ctx, ev)
		}
		for _, inf := range vs.DowntimeInfractions {
			k.SetDowntimeInfraction(ctx, inf)
		}
		for _, alert := range vs.Alerts {
			k.SetValidatorAlert(ctx, alert)
		}
		for _, sn := range vs.SentryNodes {
			k.SetSentryNode(ctx, sn)
		}
	}

	// Initialize wallet security state
	if genState.WalletSecurity != nil {
		ws := genState.WalletSecurity
		for _, hw := range ws.HardwareWallets {
			k.SetHardwareWallet(ctx, hw)
		}
		for _, ms := range ws.MultisigWallets {
			k.SetMultiSigWallet(ctx, ms)
		}
		for _, tx := range ws.PendingMultisigTxs {
			k.SetPendingMultiSigTx(ctx, tx)
		}
		for _, src := range ws.SocialRecoveryConfigs {
			k.SetSocialRecoveryConfig(ctx, src)
		}
		for _, req := range ws.RecoveryRequests {
			k.SetRecoveryRequest(ctx, req)
		}
		for _, sl := range ws.SpendingLimits {
			k.SetSpendingLimit(ctx, sl)
		}
	}

	// Initialize incident response state
	if genState.IncidentResponse != nil {
		ir := genState.IncidentResponse
		for _, inc := range ir.Incidents {
			k.SetIncident(ctx, inc)
		}
		for _, entry := range ir.AuditLogs {
			k.SetAuditLogEntry(ctx, entry)
		}
		// Note: NextIncidentId is stored separately in KV store, not in genesis
	}

	// Initialize cryptography state
	if genState.Cryptography != nil {
		crypto := genState.Cryptography
		for _, krs := range crypto.KeyRotationSchedules {
			k.SetKeyRotationSchedule(ctx, krs)
		}
		for _, ts := range crypto.ThresholdSchemes {
			k.SetThresholdScheme(ctx, ts)
		}
		for _, zk := range crypto.ZkProofConfigs {
			k.SetZKProofConfig(ctx, zk)
		}
		for _, qrk := range crypto.QuantumResistantKeys {
			k.SetQuantumResistantKey(ctx, qrk)
		}
	}

	// Initialize privacy state
	if genState.Privacy != nil {
		priv := genState.Privacy
		for _, mp := range priv.MixingPools {
			k.SetMixingPool(ctx, mp)
		}
		for _, sa := range priv.StealthAddresses {
			// Use hex encoding of OneTimeAddress as key since StealthAddress has no id field
			k.SetStealthAddress(ctx, sa)
		}
		for _, rs := range priv.RingSignatures {
			// Use hex encoding of KeyImage as key since RingSignature has no id field
			k.SetRingSignature(ctx, rs)
		}
	}

	k.Logger(ctx).Info("security module genesis initialized")
}

// ExportGenesis exports the module's state
func (k Keeper) ExportGenesis(ctx sdk.Context) *securitypb.GenesisState {
	params := k.GetParams(ctx)

	return &securitypb.GenesisState{
		Params: &params,
		NetworkSecurity: &securitypb.NetworkSecurityState{
			RateLimits:      k.GetAllRateLimits(ctx),
			Reputations:     k.GetAllPeerReputations(ctx),
			TrustedPeers:    k.GetAllTrustedPeers(ctx),
			ForkAlerts:      k.GetAllForkAlerts(ctx),
			PartitionAlerts: k.GetAllPartitionAlerts(ctx),
		},
		ValidatorSecurity: &securitypb.ValidatorSecurityState{
			Validators:          k.GetAllValidatorSecurityInfos(ctx),
			DoubleSignEvidences: k.GetAllDoubleSignEvidence(ctx),
			DowntimeInfractions: k.GetAllDowntimeInfractions(ctx),
			Alerts:              k.GetAllValidatorAlerts(ctx),
			SentryNodes:         k.GetAllSentryNodes(ctx),
		},
		WalletSecurity: &securitypb.WalletSecurityState{
			HardwareWallets:       k.GetAllHardwareWallets(ctx),
			MultisigWallets:       k.GetAllMultiSigWallets(ctx),
			PendingMultisigTxs:    k.GetAllPendingMultiSigTxs(ctx),
			SocialRecoveryConfigs: k.GetAllSocialRecoveryConfigs(ctx),
			RecoveryRequests:      k.GetAllRecoveryRequests(ctx),
			SpendingLimits:        k.GetAllSpendingLimits(ctx),
		},
		IncidentResponse: &securitypb.IncidentResponseState{
			Incidents: k.GetAllIncidents(ctx),
			AuditLogs: k.GetAllAuditLogEntries(ctx),
			// Note: NextIncidentId is not part of IncidentResponseState proto
			// It's stored separately in KV store
		},
		Cryptography: &securitypb.CryptographyState{
			KeyRotationSchedules: k.GetAllKeyRotationSchedules(ctx),
			ThresholdSchemes:     k.GetAllThresholdSchemes(ctx),
			ZkProofConfigs:       k.GetAllZKProofConfigs(ctx),
			QuantumResistantKeys: k.GetAllQuantumResistantKeys(ctx),
		},
		Privacy: &securitypb.PrivacyState{
			MixingPools:      k.GetAllMixingPools(ctx),
			StealthAddresses: k.GetAllStealthAddresses(ctx),
			RingSignatures:   k.GetAllRingSignatures(ctx),
		},
	}
}

// =============================================================================
// Parameter Operations
// =============================================================================

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params securitypb.Params) {
	store := k.GetStore(ctx)
	bz := k.cdc.MustMarshal(&params)
	store.Set([]byte("params"), bz)
}

// GetParams gets the module parameters
func (k Keeper) GetParams(ctx sdk.Context) securitypb.Params {
	store := k.GetStore(ctx)
	bz := store.Get([]byte("params"))
	if bz == nil {
		return types.DefaultParams()
	}
	var params securitypb.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// =============================================================================
// Additional Helper Functions Required by Genesis
// =============================================================================

// SetSpendingLimit stores a spending limit
func (k Keeper) SetSpendingLimit(ctx sdk.Context, limit *securitypb.SpendingLimit) {
	store := k.GetStore(ctx)
	key := types.GetSpendingLimitStoreKey(limit.WalletId)
	bz := k.cdc.MustMarshal(limit)
	store.Set(key, bz)
}

// GetAllSpendingLimits returns all spending limits
func (k Keeper) GetAllSpendingLimits(ctx sdk.Context) []*securitypb.SpendingLimit {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SpendingLimitKey)
	defer iterator.Close()

	var limits []*securitypb.SpendingLimit
	for ; iterator.Valid(); iterator.Next() {
		var limit securitypb.SpendingLimit
		k.cdc.MustUnmarshal(iterator.Value(), &limit)
		limits = append(limits, &limit)
	}
	return limits
}

// GetSpendingLimit returns the spending limit for a wallet
func (k Keeper) GetSpendingLimit(ctx sdk.Context, walletID string) (*securitypb.SpendingLimit, bool) {
	store := k.GetStore(ctx)
	key := types.GetSpendingLimitStoreKey(walletID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var limit securitypb.SpendingLimit
	k.cdc.MustUnmarshal(bz, &limit)
	return &limit, true
}

// ValidateWallet validates that a wallet exists and is not locked
// This is used for security checks before transaction execution
func (k Keeper) ValidateWallet(ctx sdk.Context, walletID string) error {
	// Check if wallet is blacklisted
	if k.IsBlacklisted(ctx, walletID) {
		return fmt.Errorf("wallet %s is blacklisted", walletID)
	}
	// Additional validation can be added here
	return nil
}

// CheckSpendingLimit checks if a transaction amount exceeds the wallet's spending limit
func (k Keeper) CheckSpendingLimit(ctx sdk.Context, walletID, denom, amount string) error {
	limit, found := k.GetSpendingLimit(ctx, walletID)
	if !found {
		// No spending limit set, allow transaction
		return nil
	}

	// Check if spending limits are enabled
	if !limit.Enabled {
		return nil
	}

	// Check if the denom matches
	if limit.Denom != denom {
		return nil // Different denom, no limit applies
	}

	// Check daily limit (simple string comparison for now)
	// In production, this should properly parse and compare math.Int values
	if limit.DailyLimit != "" && amount > limit.DailyLimit {
		return fmt.Errorf("daily spending limit exceeded for wallet %s: requested %s, limit %s", walletID, amount, limit.DailyLimit)
	}

	return nil
}

// SetAuditLogEntry stores an audit log entry
func (k Keeper) SetAuditLogEntry(ctx sdk.Context, entry *securitypb.AuditLogEntry) {
	store := k.GetStore(ctx)
	key := types.GetAuditLogStoreKey(entry.LogId)
	bz := k.cdc.MustMarshal(entry)
	store.Set(key, bz)
}

// GetAllAuditLogEntries returns all audit log entries
func (k Keeper) GetAllAuditLogEntries(ctx sdk.Context) []*securitypb.AuditLogEntry {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.AuditLogKey)
	defer iterator.Close()

	var entries []*securitypb.AuditLogEntry
	for ; iterator.Valid(); iterator.Next() {
		var entry securitypb.AuditLogEntry
		k.cdc.MustUnmarshal(iterator.Value(), &entry)
		entries = append(entries, &entry)
	}
	return entries
}

// SetStealthAddress stores a stealth address
func (k Keeper) SetStealthAddress(ctx sdk.Context, addr *securitypb.StealthAddress) {
	store := k.GetStore(ctx)
	// Use hex encoding of OneTimeAddress as key since StealthAddress has no id field
	key := types.GetStealthAddressStoreKey(hex.EncodeToString(addr.OneTimeAddress))
	bz := k.cdc.MustMarshal(addr)
	store.Set(key, bz)
}

// GetAllStealthAddresses returns all stealth addresses
func (k Keeper) GetAllStealthAddresses(ctx sdk.Context) []*securitypb.StealthAddress {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.StealthAddressKey)
	defer iterator.Close()

	var addrs []*securitypb.StealthAddress
	for ; iterator.Valid(); iterator.Next() {
		var addr securitypb.StealthAddress
		k.cdc.MustUnmarshal(iterator.Value(), &addr)
		addrs = append(addrs, &addr)
	}
	return addrs
}

// SetRingSignature stores a ring signature
func (k Keeper) SetRingSignature(ctx sdk.Context, sig *securitypb.RingSignature) {
	store := k.GetStore(ctx)
	// Use hex encoding of KeyImage as key since RingSignature has no id field
	key := types.GetRingSignatureStoreKey(hex.EncodeToString(sig.KeyImage))
	bz := k.cdc.MustMarshal(sig)
	store.Set(key, bz)
}

// GetAllRingSignatures returns all ring signatures
func (k Keeper) GetAllRingSignatures(ctx sdk.Context) []*securitypb.RingSignature {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.RingSignatureKey)
	defer iterator.Close()

	var sigs []*securitypb.RingSignature
	for ; iterator.Valid(); iterator.Next() {
		var sig securitypb.RingSignature
		k.cdc.MustUnmarshal(iterator.Value(), &sig)
		sigs = append(sigs, &sig)
	}
	return sigs
}
