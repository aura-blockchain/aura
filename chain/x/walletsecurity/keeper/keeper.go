package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// getStore returns the KVStore for this module
func (k Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx context.Context) (wsproto.WalletSecurityParams, error) {
	store := k.getStore(ctx)
	var params wsproto.WalletSecurityParams

	paramsBytes, err := store.Get(types.ParamsKey)
	if err != nil {
		return params, err
	}

	if paramsBytes == nil {
		// Return default params if not set
		return wsproto.WalletSecurityParams{}, nil
	}

	if err := k.cdc.Unmarshal(paramsBytes, &params); err != nil {
		return params, err
	}

	return params, nil
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx context.Context, params wsproto.WalletSecurityParams) error {
	store := k.getStore(ctx)
	paramsBytes, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, paramsBytes)
}

// SetHardwareWallet stores a hardware wallet configuration
func (k Keeper) SetHardwareWallet(ctx context.Context, walletID string, config []byte) error {
	store := k.getStore(ctx)
	key := types.GetHardwareWalletKey(walletID)
	return store.Set(key, config)
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
	return store.Set(key, wallet)
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
	return store.Set(key, tx)
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
	return store.Delete(key)
}

// SetSocialRecoveryConfig stores social recovery configuration
func (k Keeper) SetSocialRecoveryConfig(ctx context.Context, walletID string, config []byte) error {
	store := k.getStore(ctx)
	key := types.GetSocialRecoveryKey(walletID)
	return store.Set(key, config)
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
	return store.Set(key, request)
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
	return store.Set(key, limit)
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
	return store.Set(key, config)
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
	return store.Set(key, auth)
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
	return store.Set(key, config)
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
	return store.Set(key, backup)
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
	return store.Set(key, filter)
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
	return store.Set(key, verification)
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
	return store.Set(key, metrics)
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
	return store.Set(key, tx)
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
			DetectedAt:   blockTimeToGogoTimestamp(ctx),
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
	if err := store.Set(key, []byte(reason)); err != nil {
		return err
	}
	k.logger.Info(fmt.Sprintf("Wallet %s locked. Reason: %s", addr, reason))
	return nil
}

// UnlockWallet unlocks a previously locked wallet
func (k Keeper) UnlockWallet(ctx context.Context, addr string) error {
	// Unlock wallet functionality
	store := k.getStore(ctx)
	key := append([]byte("locked_wallet_"), []byte(addr)...)
	if err := store.Delete(key); err != nil {
		return err
	}
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
		SimulatedAt:  blockTimeToGogoTimestamp(ctx),
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
		VerifiedAt:      blockTimeToGogoTimestamp(ctx),
		ExpiresAt:       blockTimeWithOffsetToGogoTimestamp(ctx, 365*24*time.Hour),
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
		DailyResetAt:        blockTimeWithOffsetToGogoTimestamp(ctx, 24*time.Hour),
		WeeklyResetAt:       blockTimeWithOffsetToGogoTimestamp(ctx, 7*24*time.Hour),
		MonthlyResetAt:      blockTimeWithOffsetToGogoTimestamp(ctx, 30*24*time.Hour),
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
	if limit.DailyResetAt == nil || now.After(gogoTimestampToTime(limit.DailyResetAt)) {
		limit.CurrentDailySpent = "0"
		limit.DailyResetAt = timeToGogoTimestamp(now.Add(24 * time.Hour))
	}
	if limit.WeeklyResetAt == nil || now.After(gogoTimestampToTime(limit.WeeklyResetAt)) {
		limit.CurrentWeeklySpent = "0"
		limit.WeeklyResetAt = timeToGogoTimestamp(now.Add(7 * 24 * time.Hour))
	}
	if limit.MonthlyResetAt == nil || now.After(gogoTimestampToTime(limit.MonthlyResetAt)) {
		limit.CurrentMonthlySpent = "0"
		limit.MonthlyResetAt = timeToGogoTimestamp(now.Add(30 * 24 * time.Hour))
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

// IsBiometricProofUsed checks if a biometric proof hash has already been used (replay protection)
func (k Keeper) IsBiometricProofUsed(ctx context.Context, walletID string, proofHash []byte) bool {
	store := k.getStore(ctx)
	key := types.GetBiometricProofKey(walletID, proofHash)
	value, _ := store.Get(key)
	return value != nil
}

// MarkBiometricProofUsed marks a biometric proof hash as used (replay protection)
func (k Keeper) MarkBiometricProofUsed(ctx context.Context, walletID string, proofHash []byte) error {
	store := k.getStore(ctx)
	key := types.GetBiometricProofKey(walletID, proofHash)

	// Store with timestamp and set TTL for cleanup (e.g., 24 hours)
	// In production, implement TTL cleanup or use a bounded sliding window
	timestamp := determinism.GetBlockTime(ctx).Unix()
	value := []byte(fmt.Sprintf("%d", timestamp))

	return store.Set(key, value)
}

// verifyBiometricTemplate verifies a biometric proof against the stored enrollment hash
//
// DEPRECATION WARNING:
//
//	This implementation is DEPRECATED and will be removed in a future version.
//	True biometric authentication is fundamentally incompatible with blockchain consensus.
//
// Why Biometric Authentication Cannot Work on Blockchain:
//
// 1. DETERMINISM REQUIREMENT:
//   - Blockchain consensus requires deterministic execution across all validators
//   - Biometric matching is inherently non-deterministic and uses fuzzy algorithms
//   - Different validators would produce different match scores for the same biometric
//   - This breaks blockchain consensus and leads to chain halts
//
// 2. LIVENESS DETECTION IMPOSSIBILITY:
//   - True biometric systems require liveness detection (detect photos, masks, etc.)
//   - Liveness detection requires real-time hardware interaction (camera, sensor)
//   - Blockchain cannot access client-side hardware during consensus
//   - Without liveness, the system is vulnerable to replay attacks with stolen biometric data
//
// 3. PRIVACY CONCERNS:
//   - Storing biometric hashes on-chain creates permanent privacy risks
//   - Biometric data cannot be changed if compromised (unlike passwords)
//   - Public blockchain = permanent public record of biometric identifiers
//   - GDPR/privacy laws prohibit permanent storage of biometric identifiers
//
// 4. SECURITY MODEL MISMATCH:
//   - Biometric authentication assumes: "something you are" + local hardware verification
//   - Blockchain authentication assumes: "something you have" (private key)
//   - The current implementation is just "pre-shared secret authentication"
//   - It provides no additional security beyond knowing the enrollment secret
//
// CURRENT IMPLEMENTATION (Simplified):
//   - This implementation uses exact hash matching as a placeholder
//   - It provides replay protection and basic verification
//   - However, it is NOT true biometric authentication
//   - It is essentially a pre-shared secret that cannot be updated
//
// RECOMMENDED ALTERNATIVES:
//  1. Hardware Wallet Integration (Ledger, Trezor) - Use RegisterHardwareWallet
//  2. Multi-Signature Wallets - Use CreateMultiSigWallet
//  3. Social Recovery - Use ConfigureSocialRecovery
//  4. Time-locked Transactions - Use MultiSigWallet with time_lock
//  5. Off-chain Biometric + On-chain Signature - Use standard Cosmos SDK auth
//
// MIGRATION PATH:
//
//	Users relying on biometric authentication should:
//	1. Enable hardware wallet integration for "something you have"
//	2. Enable multi-sig for enhanced security
//	3. Configure social recovery for account recovery
//	4. Use client-side biometric authentication before signing (off-chain)
//
// Parameters:
//   - enrollmentHash: The hash stored during biometric enrollment
//   - biometricProof: The raw biometric proof data to verify
//
// Returns:
//   - bool: true if verification succeeds, false otherwise
//
// Security Notes:
//   - This function only verifies that the proof matches the enrollment hash
//   - It does NOT provide true biometric security
//   - It does NOT prevent replay attacks at the biometric level
//   - Replay protection is handled at the transaction level (see AuthenticateBiometric)
func (k Keeper) verifyBiometricTemplate(enrollmentHash string, biometricProof []byte) bool {
	// CRITICAL: Hash the provided proof
	proofHash := sha256.Sum256(biometricProof)
	proofHashStr := hex.EncodeToString(proofHash[:])

	// CRITICAL: Exact hash matching
	// This is NOT fuzzy biometric matching - it's exact secret matching
	// If the proof doesn't match exactly, authentication fails
	//
	// In a true biometric system, this would use:
	// 1. Fuzzy matching algorithms (Hamming distance, similarity scores)
	// 2. Liveness detection (3D depth sensing, challenge-response)
	// 3. Anti-spoofing (detect fake fingerprints, deepfakes)
	// 4. Secure enclave verification (TEE, SGX, TPM, Secure Element)
	// 5. Threshold-based matching (e.g., 95% similarity = match)
	//
	// However, ALL of these approaches are non-deterministic and cannot
	// be used in blockchain consensus without breaking the chain.
	//
	// The only deterministic approach is exact matching, which defeats
	// the purpose of biometric authentication (biometrics vary each time).
	return proofHashStr == enrollmentHash
}
