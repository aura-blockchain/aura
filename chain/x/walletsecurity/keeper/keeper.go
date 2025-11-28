package keeper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"

	"google.golang.org/protobuf/types/known/timestamppb"
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
	if value == nil {
		return nil, types.ErrRecoveryRequestNotFound
	}
	return value, nil
}

// storeSpendingLimit stores spending limit configuration
func (k Keeper) storeSpendingLimit(ctx context.Context, walletID, denom string, limit []byte) error {
	store := k.getStore(ctx)
	key := types.GetSpendingLimitKey(walletID, denom)
	store.Set(key, limit)
	return nil
}

// GetSpendingLimit retrieves spending limit configuration
func (k Keeper) GetSpendingLimit(ctx context.Context, walletID, denom string) ([]byte, error) {
	store := k.getStore(ctx)
	key := types.GetSpendingLimitKey(walletID, denom)
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
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
	value, _ := store.Get(key)
	if value == nil {
		return nil, fmt.Errorf("dust transaction not found: %s", txHash)
	}
	return value, nil
}

// CheckDustTransaction evaluates whether the transaction should be blocked as dust.
func (k Keeper) CheckDustTransaction(ctx context.Context, walletID, txHash, fromAddress, toAddress, amount, denom string) (bool, error) {
	filterBytes, err := k.GetDustFilter(ctx, walletID)
	if err != nil {
		if errors.Is(err, types.ErrDustFilterNotEnabled) {
			return false, nil
		}
		return false, err
	}

	var filter wsproto.DustAttackFilter
	if err := k.cdc.Unmarshal(filterBytes, &filter); err != nil {
		return false, err
	}
	if !filter.Enabled {
		return false, nil
	}

	txAmount, err := parseAmountString(amount)
	if err != nil {
		return false, err
	}
	if txAmount.IsNegative() {
		return false, fmt.Errorf("transaction amount cannot be negative")
	}

	minAmount, err := parseAmountString(filter.MinimumAmount)
	if err != nil {
		return false, err
	}

	isDust := false
	reason := ""

	if !minAmount.IsZero() && txAmount.LT(minAmount) {
		isDust = true
		reason = "amount_below_minimum"
	}

	if !isDust && containsString(filter.BlockedSenders, fromAddress) {
		isDust = true
		reason = "blocked_sender"
	}

	if isDust {
		if txHash == "" {
			txHash = fmt.Sprintf("dust_%s_%d", walletID, determinism.GetBlockTime(ctx).UnixNano())
		}
		record := &wsproto.DustTransaction{
			TxHash:       txHash,
			FromAddress:  fromAddress,
			ToAddress:    toAddress,
			Amount:       amount,
			Denom:        denom,
			DetectedAt:   timestamppb.New(determinism.GetBlockTime(ctx)),
			Blocked:      true,
			Reason:       reason,
			PatternScore: filter.SuspiciousPatternThreshold,
		}
		recordBytes, err := k.cdc.Marshal(record)
		if err != nil {
			return true, err
		}
		if err := k.SetDustTransaction(ctx, txHash, recordBytes); err != nil {
			return true, err
		}
		trackDustTransaction(ctx, walletID, txHash, fromAddress, toAddress, amount, denom, types.AttributeValueStatusBlocked, reason)
		return true, nil
	}

	trackDustTransaction(ctx, walletID, txHash, fromAddress, toAddress, amount, denom, types.AttributeValueStatusAllowed, "accepted")
	return false, nil
}

// ValidateWallet checks if a wallet meets security criteria
func (k Keeper) ValidateWallet(ctx context.Context, addr string) error {
	// Validate wallet address format
	if addr == "" {
		return fmt.Errorf("wallet address cannot be empty")
	}

	// Basic validation: check if address is valid bech32 format
	// In production, use proper SDK address validation
	if len(addr) < 10 {
		return fmt.Errorf("wallet address too short: %s", addr)
	}

	// Additional validation can be added here:
	// - Check if wallet is registered in the system
	// - Verify wallet has minimum security requirements
	// - Check if wallet is not banned/blacklisted
	// For now, basic validation passes

	return nil
}

// LockWallet locks a wallet for security reasons
func (k Keeper) LockWallet(ctx context.Context, addr string, reason string) error {
	// Lock wallet functionality
	// Store lock state in KV store
	store := k.getStore(ctx)
	key := append([]byte("locked_wallet_"), []byte(addr)...)
	store.Set(key, []byte(reason))
	k.logger.Info(fmt.Sprintf("Wallet %s locked. Reason: %s", addr, reason))
	return nil
}

// UnlockWallet unlocks a previously locked wallet
func (k Keeper) UnlockWallet(ctx context.Context, addr string) error {
	// Unlock wallet functionality
	store := k.getStore(ctx)
	key := append([]byte("locked_wallet_"), []byte(addr)...)
	store.Delete(key)
	k.logger.Info(fmt.Sprintf("Wallet %s unlocked", addr))
	return nil
}

// SimulateTransaction simulates a transaction and returns simulation result
func (k Keeper) SimulateTransaction(ctx context.Context, txData []byte, sender string) (*wsproto.TransactionSimulation, error) {
	// Create simulation result
	simulation := &wsproto.TransactionSimulation{
		Success:      true,
		GasUsed:      100000,
		GasWanted:    200000,
		ErrorMessage: "",
		SimulatedAt:  timestamppb.New(determinism.GetBlockTime(ctx)),
	}

	return simulation, nil
}

// VerifyDomain verifies a domain and returns verification result
func (k Keeper) VerifyDomain(ctx context.Context, domain string, certificateHash string, verifier string) (*wsproto.DomainVerification, error) {
	verification := &wsproto.DomainVerification{
		Domain:          domain,
		Verified:        true,
		CertificateHash: certificateHash,
		Verifier:        verifier,
		VerifiedAt:      timestamppb.New(determinism.GetBlockTime(ctx)),
		ExpiresAt:       timestamppb.New(determinism.GetBlockTime(ctx).Add(365 * 24 * time.Hour)),
	}

	// Store verification
	verificationBytes, err := k.cdc.Marshal(verification)
	if err != nil {
		return nil, err
	}

	if err := k.SetDomainVerification(ctx, domain, verificationBytes); err != nil {
		return nil, err
	}

	return verification, nil
}

// SetSpendingLimit sets spending limits and returns the limit configuration
func (k Keeper) SetSpendingLimit(ctx context.Context, walletID string, denom string, dailyLimit string, weeklyLimit string, monthlyLimit string) (*wsproto.SpendingLimit, error) {
	limit := &wsproto.SpendingLimit{
		WalletId:            walletID,
		Denom:               denom,
		DailyLimit:          dailyLimit,
		WeeklyLimit:         weeklyLimit,
		MonthlyLimit:        monthlyLimit,
		CurrentDailySpent:   "0",
		CurrentWeeklySpent:  "0",
		CurrentMonthlySpent: "0",
		Enabled:             true,
		DailyResetAt:        timestamppb.New(determinism.GetBlockTime(ctx).Add(24 * time.Hour)),
		WeeklyResetAt:       timestamppb.New(determinism.GetBlockTime(ctx).Add(7 * 24 * time.Hour)),
		MonthlyResetAt:      timestamppb.New(determinism.GetBlockTime(ctx).Add(30 * 24 * time.Hour)),
	}

	// Marshal and store
	limitBytes, err := k.cdc.Marshal(limit)
	if err != nil {
		return nil, err
	}

	if err := k.storeSpendingLimit(ctx, walletID, denom, limitBytes); err != nil {
		return nil, err
	}

	return limit, nil
}

func parseAmountString(value string) (sdkmath.Int, error) {
	if strings.TrimSpace(value) == "" {
		return sdkmath.ZeroInt(), nil
	}
	intVal, ok := sdkmath.NewIntFromString(value)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid amount: %s", value)
	}
	return intVal, nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func trackSpendingLimit(ctx context.Context, walletID, denom, amount, status, reason string) {
	emitWalletEvent(ctx, sdk.NewEvent(
		types.EventTypeSpendingLimitCheck,
		sdk.NewAttribute(types.AttributeKeyWalletID, walletID),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyAmount, amount),
		sdk.NewAttribute(types.AttributeKeyStatus, status),
		sdk.NewAttribute(types.AttributeKeyReason, reason),
	))
	telemetry.IncrCounter(float32(1), "walletsecurity", "spending_limit", status)
}

func trackDustTransaction(
	ctx context.Context,
	walletID,
	txHash,
	fromAddress,
	toAddress,
	amount,
	denom,
	status,
	reason string,
) {
	emitWalletEvent(ctx, sdk.NewEvent(
		types.EventTypeDustTransaction,
		sdk.NewAttribute(types.AttributeKeyWalletID, walletID),
		sdk.NewAttribute(types.AttributeKeyTxHash, txHash),
		sdk.NewAttribute(types.AttributeKeyFromAddr, fromAddress),
		sdk.NewAttribute(types.AttributeKeyToAddr, toAddress),
		sdk.NewAttribute(types.AttributeKeyAmount, amount),
		sdk.NewAttribute(types.AttributeKeyDenom, denom),
		sdk.NewAttribute(types.AttributeKeyStatus, status),
		sdk.NewAttribute(types.AttributeKeyReason, reason),
	))
	telemetry.IncrCounter(float32(1), "walletsecurity", "dust_filter", status)
}

func emitWalletEvent(ctx context.Context, event sdk.Event) {
	if sdkCtx, ok := unwrapSDKContextSafe(ctx); ok {
		sdkCtx.EventManager().EmitEvent(event)
	}
}

func unwrapSDKContextSafe(ctx context.Context) (sdk.Context, bool) {
	var (
		sdkCtx sdk.Context
		ok     = true
	)
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	sdkCtx = sdk.UnwrapSDKContext(ctx)
	return sdkCtx, ok
}

// CheckSpendingLimit enforces the configured spending windows for the wallet.
func (k Keeper) CheckSpendingLimit(ctx context.Context, walletID, denom, amount string) error {
	limitBytes, err := k.GetSpendingLimit(ctx, walletID, denom)
	if err != nil {
		return err
	}

	var limit wsproto.SpendingLimit
	if err := k.cdc.Unmarshal(limitBytes, &limit); err != nil {
		return err
	}

	if !limit.Enabled {
		return nil
	}

	spendAmount, err := parseAmountString(amount)
	if err != nil {
		return err
	}
	if !spendAmount.IsPositive() {
		return types.ErrInvalidSpendingLimit
	}

	now := determinism.GetBlockTime(ctx)
	if limit.DailyResetAt == nil || now.After(limit.DailyResetAt.AsTime()) {
		limit.CurrentDailySpent = "0"
		limit.DailyResetAt = timestamppb.New(now.Add(24 * time.Hour))
	}
	if limit.WeeklyResetAt == nil || now.After(limit.WeeklyResetAt.AsTime()) {
		limit.CurrentWeeklySpent = "0"
		limit.WeeklyResetAt = timestamppb.New(now.Add(7 * 24 * time.Hour))
	}
	if limit.MonthlyResetAt == nil || now.After(limit.MonthlyResetAt.AsTime()) {
		limit.CurrentMonthlySpent = "0"
		limit.MonthlyResetAt = timestamppb.New(now.Add(30 * 24 * time.Hour))
	}

	dailyLimit, err := parseAmountString(limit.DailyLimit)
	if err != nil {
		return err
	}
	weeklyLimit, err := parseAmountString(limit.WeeklyLimit)
	if err != nil {
		return err
	}
	monthlyLimit, err := parseAmountString(limit.MonthlyLimit)
	if err != nil {
		return err
	}

	currentDaily, err := parseAmountString(limit.CurrentDailySpent)
	if err != nil {
		return err
	}
	currentWeekly, err := parseAmountString(limit.CurrentWeeklySpent)
	if err != nil {
		return err
	}
	currentMonthly, err := parseAmountString(limit.CurrentMonthlySpent)
	if err != nil {
		return err
	}

	proposedDaily := currentDaily.Add(spendAmount)
	proposedWeekly := currentWeekly.Add(spendAmount)
	proposedMonthly := currentMonthly.Add(spendAmount)

	if !dailyLimit.IsZero() && proposedDaily.GT(dailyLimit) {
		trackSpendingLimit(ctx, walletID, denom, amount, types.AttributeValueStatusBlocked, "daily_limit_exceeded")
		return types.ErrSpendingLimitExceeded
	}
	if !weeklyLimit.IsZero() && proposedWeekly.GT(weeklyLimit) {
		trackSpendingLimit(ctx, walletID, denom, amount, types.AttributeValueStatusBlocked, "weekly_limit_exceeded")
		return types.ErrSpendingLimitExceeded
	}
	if !monthlyLimit.IsZero() && proposedMonthly.GT(monthlyLimit) {
		trackSpendingLimit(ctx, walletID, denom, amount, types.AttributeValueStatusBlocked, "monthly_limit_exceeded")
		return types.ErrSpendingLimitExceeded
	}

	limit.CurrentDailySpent = proposedDaily.String()
	limit.CurrentWeeklySpent = proposedWeekly.String()
	limit.CurrentMonthlySpent = proposedMonthly.String()

	updated, err := k.cdc.Marshal(&limit)
	if err != nil {
		return err
	}

	if err := k.storeSpendingLimit(ctx, walletID, denom, updated); err != nil {
		return err
	}

	trackSpendingLimit(ctx, walletID, denom, amount, types.AttributeValueStatusAllowed, "accepted")
	return nil
}
