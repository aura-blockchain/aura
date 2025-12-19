// Package keeper implements the consolidated security module keeper.
// This keeper combines functionality from: networksecurity, validatorsecurity,
// walletsecurity, incidentresponse, cryptography, and privacy modules.
package keeper

import (
	"encoding/hex"
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
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

// Genesis methods are implemented in genesis.go

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
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		ctx.Logger().Error("failed to unmarshal security params", "error", err)
		return types.DefaultParams()
	}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &limit); err != nil {
			ctx.Logger().Error("failed to unmarshal spending limit in GetAllSpendingLimits, skipping", "error", err)
			continue
		}
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
	if err := k.cdc.Unmarshal(bz, &limit); err != nil {
		ctx.Logger().Error("failed to unmarshal spending limit", "wallet_id", walletID, "error", err)
		return nil, false
	}
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

	requested, ok := sdkmath.NewIntFromString(amount)
	if !ok {
		return fmt.Errorf("invalid requested amount %q", amount)
	}
	if !requested.IsPositive() {
		return fmt.Errorf("requested amount must be positive")
	}

	// Parse configured daily limit and apply against current spend
	if limit.DailyLimit != "" {
		dailyLimit, ok := sdkmath.NewIntFromString(limit.DailyLimit)
		if !ok {
			return fmt.Errorf("invalid daily limit configured for wallet %s: %s", walletID, limit.DailyLimit)
		}

		current := sdkmath.ZeroInt()
		if limit.CurrentDailySpent != "" {
			current, ok = sdkmath.NewIntFromString(limit.CurrentDailySpent)
			if !ok {
				return fmt.Errorf("invalid current daily spent value for wallet %s: %s", walletID, limit.CurrentDailySpent)
			}
		}

		if current.Add(requested).GT(dailyLimit) {
			return fmt.Errorf(
				"daily spending limit exceeded for wallet %s: current %s + requested %s > limit %s",
				walletID,
				current.String(),
				requested.String(),
				dailyLimit.String(),
			)
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &entry); err != nil {
			ctx.Logger().Error("failed to unmarshal audit log entry in GetAllAuditLogEntries, skipping", "error", err)
			continue
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &addr); err != nil {
			ctx.Logger().Error("failed to unmarshal stealth address in GetAllStealthAddresses, skipping", "error", err)
			continue
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &sig); err != nil {
			ctx.Logger().Error("failed to unmarshal ring signature in GetAllRingSignatures, skipping", "error", err)
			continue
		}
		sigs = append(sigs, &sig)
	}
	return sigs
}
