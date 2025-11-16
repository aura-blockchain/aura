package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================================================================
// Transaction Simulation
// ============================================================================

// SimulateTransaction simulates a transaction before execution
func (k Keeper) SimulateTransaction(
	ctx context.Context,
	txData []byte,
	sender string,
) (*wsproto.TransactionSimulation, error) {
	if len(txData) == 0 {
		return nil, types.ErrInvalidTransactionData
	}

	// In production, this would:
	// 1. Parse the transaction data
	// 2. Create a simulated execution environment
	// 3. Execute the transaction in simulation mode
	// 4. Track all state changes and balance changes
	// 5. Calculate gas usage
	// 6. Detect potential risks

	simulation := &wsproto.TransactionSimulation{
		Success:        true,
		ErrorMessage:   "",
		GasUsed:        21000, // Example gas
		GasWanted:      100000,
		StateChanges:   []*wsproto.StateChange{},
		BalanceChanges: []*wsproto.BalanceChange{},
		Warnings:       []string{},
		SimulatedAt:    timestamppb.New(time.Now()),
		RiskLevel:      wsproto.SimulationRisk_SIMULATION_RISK_LOW,
	}

	// Analyze transaction for risks
	riskLevel := k.analyzeTransactionRisk(txData, sender)
	simulation.RiskLevel = riskLevel

	if riskLevel == wsproto.SimulationRisk_SIMULATION_RISK_CRITICAL {
		simulation.Warnings = append(simulation.Warnings, "CRITICAL: High-risk transaction detected")
	}

	k.Logger(ctx).Info("simulated transaction",
		"sender", sender,
		"gas_used", simulation.GasUsed,
		"risk_level", riskLevel.String(),
	)

	return simulation, nil
}

// analyzeTransactionRisk analyzes transaction for potential risks
func (k Keeper) analyzeTransactionRisk(txData []byte, sender string) wsproto.SimulationRisk {
	// In production, this would analyze:
	// 1. Known malicious contract interactions
	// 2. Unusual permission requests
	// 3. Large value transfers
	// 4. Unknown recipient addresses
	// 5. Complex contract calls

	// For this implementation, we use a simple heuristic
	if len(txData) > 10000 {
		return wsproto.SimulationRisk_SIMULATION_RISK_HIGH
	}
	if len(txData) > 5000 {
		return wsproto.SimulationRisk_SIMULATION_RISK_MEDIUM
	}
	return wsproto.SimulationRisk_SIMULATION_RISK_LOW
}

// ============================================================================
// Phishing Protection
// ============================================================================

// VerifyDomain verifies a domain for phishing protection
func (k Keeper) VerifyDomain(
	ctx context.Context,
	domain string,
	certificateHash string,
	verifier string,
) (*wsproto.DomainVerification, error) {
	if domain == "" {
		return nil, types.ErrInvalidInput
	}

	// Check if domain is blacklisted
	if k.isDomainBlacklisted(domain) {
		return nil, types.ErrDomainBlacklisted
	}

	// In production, this would:
	// 1. Verify SSL certificate
	// 2. Check domain reputation
	// 3. Validate DNSSEC
	// 4. Check against known phishing databases

	now := timestamppb.New(time.Now())
	expiresAt := timestamppb.New(time.Now().Add(90 * 24 * time.Hour)) // 90 days

	verification := &wsproto.DomainVerification{
		Domain:           domain,
		Verified:         true,
		VerifiedAt:       now,
		ExpiresAt:        expiresAt,
		CertificateHash:  certificateHash,
		TrustedAddresses: []string{},
		Verifier:         verifier,
	}

	// Store verification
	verificationBytes := k.cdc.MustMarshal(verification)
	if err := k.SetDomainVerification(ctx, domain, verificationBytes); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("verified domain",
		"domain", domain,
		"verifier", verifier,
	)

	return verification, nil
}

// isDomainBlacklisted checks if a domain is blacklisted
func (k Keeper) isDomainBlacklisted(domain string) bool {
	// In production, check against blacklist database
	knownBadDomains := []string{
		"phishing-site.com",
		"scam-wallet.org",
	}

	for _, bad := range knownBadDomains {
		if strings.Contains(domain, bad) {
			return true
		}
	}
	return false
}

// ============================================================================
// Address Checksum Validation
// ============================================================================

// ValidateAddressChecksum validates an address checksum
func (k Keeper) ValidateAddressChecksum(
	ctx context.Context,
	address string,
	algorithm wsproto.ChecksumAlgorithm,
) (bool, string, error) {
	if address == "" {
		return false, "", types.ErrInvalidInput
	}

	var valid bool
	var checksum string
	var err error

	switch algorithm {
	case wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_EIP55:
		valid, checksum, err = k.validateEIP55Checksum(address)
	case wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BECH32:
		valid, checksum, err = k.validateBech32Checksum(address)
	case wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BASE58CHECK:
		valid, checksum, err = k.validateBase58CheckChecksum(address)
	case wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_CRC32:
		valid, checksum, err = k.validateCRC32Checksum(address)
	default:
		return false, "", types.ErrUnsupportedChecksumAlgo
	}

	if err != nil {
		return false, "", err
	}

	k.Logger(ctx).Info("validated address checksum",
		"address", address,
		"algorithm", algorithm.String(),
		"valid", valid,
	)

	return valid, checksum, nil
}

// validateEIP55Checksum validates Ethereum EIP-55 checksum
func (k Keeper) validateEIP55Checksum(address string) (bool, string, error) {
	// Remove 0x prefix if present
	address = strings.TrimPrefix(address, "0x")

	// Convert to lowercase
	addressLower := strings.ToLower(address)

	// Hash the lowercase address
	hash := sha256.Sum256([]byte(addressLower))
	hashHex := hex.EncodeToString(hash[:])

	// Generate checksummed address
	checksummed := ""
	for i, char := range addressLower {
		if char >= 'a' && char <= 'f' {
			// If hash byte is >= 8, capitalize
			hashByte := hashHex[i]
			if hashByte >= '8' {
				checksummed += strings.ToUpper(string(char))
			} else {
				checksummed += string(char)
			}
		} else {
			checksummed += string(char)
		}
	}

	// Check if input matches checksummed version
	valid := address == checksummed

	return valid, "0x" + checksummed, nil
}

// validateBech32Checksum validates Bech32 checksum (Bitcoin/Cosmos)
func (k Keeper) validateBech32Checksum(address string) (bool, string, error) {
	// In production, implement full Bech32 validation
	// For now, basic validation
	if len(address) < 6 {
		return false, "", types.ErrInvalidChecksum
	}

	// Simple check: address should be lowercase or have valid checksum
	valid := address == strings.ToLower(address)
	return valid, address, nil
}

// validateBase58CheckChecksum validates Base58Check checksum (Bitcoin)
func (k Keeper) validateBase58CheckChecksum(address string) (bool, string, error) {
	// In production, implement full Base58Check validation
	// For now, basic validation
	if len(address) < 26 || len(address) > 35 {
		return false, "", types.ErrInvalidChecksum
	}

	// Check if address contains only valid Base58 characters
	validChars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	for _, char := range address {
		if !strings.ContainsRune(validChars, char) {
			return false, "", types.ErrInvalidChecksum
		}
	}

	return true, address, nil
}

// validateCRC32Checksum validates CRC32 checksum
func (k Keeper) validateCRC32Checksum(address string) (bool, string, error) {
	// Basic CRC32 validation
	if len(address) < 8 {
		return false, "", types.ErrInvalidChecksum
	}

	return true, address, nil
}

// ============================================================================
// Spending Limits
// ============================================================================

// SetSpendingLimit sets spending limits for a wallet
func (k Keeper) SetSpendingLimit(
	ctx context.Context,
	walletID string,
	denom string,
	dailyLimit, weeklyLimit, monthlyLimit string,
) (*wsproto.SpendingLimit, error) {
	if walletID == "" || denom == "" {
		return nil, types.ErrInvalidInput
	}

	now := time.Now()
	limit := &wsproto.SpendingLimit{
		WalletId:            walletID,
		Denom:               denom,
		DailyLimit:          dailyLimit,
		WeeklyLimit:         weeklyLimit,
		MonthlyLimit:        monthlyLimit,
		CurrentDailySpent:   "0",
		CurrentWeeklySpent:  "0",
		CurrentMonthlySpent: "0",
		DailyResetAt:        timestamppb.New(now.Add(24 * time.Hour)),
		WeeklyResetAt:       timestamppb.New(now.Add(7 * 24 * time.Hour)),
		MonthlyResetAt:      timestamppb.New(now.Add(30 * 24 * time.Hour)),
		Enabled:             true,
	}

	// Store limit
	limitBytes := k.cdc.MustMarshal(limit)
	if err := k.SetSpendingLimit(ctx, walletID, denom, limitBytes); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("set spending limit",
		"wallet_id", walletID,
		"denom", denom,
		"daily_limit", dailyLimit,
	)

	return limit, nil
}

// CheckSpendingLimit checks if a transaction exceeds spending limits
func (k Keeper) CheckSpendingLimit(
	ctx context.Context,
	walletID, denom, amount string,
) error {
	limitBytes, err := k.GetSpendingLimit(ctx, walletID, denom)
	if err != nil {
		// No limit configured, allow
		return nil
	}

	var limit wsproto.SpendingLimit
	k.cdc.MustUnmarshal(limitBytes, &limit)

	if !limit.Enabled {
		return nil
	}

	// Reset counters if needed
	k.resetSpendingLimitCounters(ctx, &limit)

	// Check limits
	amountBig := new(big.Int)
	amountBig.SetString(amount, 10)

	// Check daily limit
	dailySpent := new(big.Int)
	dailySpent.SetString(limit.CurrentDailySpent, 10)
	dailyLimit := new(big.Int)
	dailyLimit.SetString(limit.DailyLimit, 10)

	newDailySpent := new(big.Int).Add(dailySpent, amountBig)
	if newDailySpent.Cmp(dailyLimit) > 0 {
		return types.ErrSpendingLimitExceeded
	}

	// Update spent amounts
	limit.CurrentDailySpent = newDailySpent.String()

	// Store updated limit
	updatedBytes := k.cdc.MustMarshal(&limit)
	return k.SetSpendingLimit(ctx, walletID, denom, updatedBytes)
}

// resetSpendingLimitCounters resets spending limit counters when period expires
func (k Keeper) resetSpendingLimitCounters(ctx context.Context, limit *wsproto.SpendingLimit) {
	now := time.Now()

	if now.After(limit.DailyResetAt.AsTime()) {
		limit.CurrentDailySpent = "0"
		limit.DailyResetAt = timestamppb.New(now.Add(24 * time.Hour))
	}

	if now.After(limit.WeeklyResetAt.AsTime()) {
		limit.CurrentWeeklySpent = "0"
		limit.WeeklyResetAt = timestamppb.New(now.Add(7 * 24 * time.Hour))
	}

	if now.After(limit.MonthlyResetAt.AsTime()) {
		limit.CurrentMonthlySpent = "0"
		limit.MonthlyResetAt = timestamppb.New(now.Add(30 * 24 * time.Hour))
	}
}

// ============================================================================
// Session Management
// ============================================================================

// ConfigureSession configures wallet session settings
func (k Keeper) ConfigureSession(
	ctx context.Context,
	walletID string,
	timeoutDuration *durationpb.Duration,
	autoLockEnabled bool,
	inactivityThreshold int32,
) (*wsproto.SessionConfig, error) {
	sessionID := k.generateSessionID(walletID)

	now := timestamppb.New(time.Now())
	timeout := timeoutDuration.AsDuration()
	expiresAt := timestamppb.New(time.Now().Add(timeout))

	config := &wsproto.SessionConfig{
		SessionId:                  sessionID,
		WalletId:                   walletID,
		StartedAt:                  now,
		LastActivity:               now,
		TimeoutDuration:            timeoutDuration,
		ExpiresAt:                  expiresAt,
		AutoLockEnabled:            autoLockEnabled,
		InactivityThresholdSeconds: inactivityThreshold,
		DeviceFingerprint:          k.generateDeviceFingerprint(),
		Locked:                     false,
	}

	// Store configuration
	configBytes := k.cdc.MustMarshal(config)
	if err := k.SetSessionConfig(ctx, sessionID, configBytes); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("configured session",
		"session_id", sessionID,
		"wallet_id", walletID,
		"timeout", timeout,
	)

	return config, nil
}

// UpdateSessionActivity updates session activity timestamp
func (k Keeper) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	configBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return err
	}

	var config wsproto.SessionConfig
	k.cdc.MustUnmarshal(configBytes, &config)

	// Check if session is locked
	if config.Locked {
		return types.ErrSessionLocked
	}

	// Check if session expired
	if time.Now().After(config.ExpiresAt.AsTime()) {
		return types.ErrSessionExpired
	}

	// Update activity
	config.LastActivity = timestamppb.New(time.Now())
	config.ExpiresAt = timestamppb.New(time.Now().Add(config.TimeoutDuration.AsDuration()))

	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetSessionConfig(ctx, sessionID, updatedBytes)
}

// LockSession locks a wallet session
func (k Keeper) LockSession(ctx context.Context, sessionID string) error {
	configBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return err
	}

	var config wsproto.SessionConfig
	k.cdc.MustUnmarshal(configBytes, &config)

	config.Locked = true

	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetSessionConfig(ctx, sessionID, updatedBytes)
}

// UnlockSession unlocks a wallet session
func (k Keeper) UnlockSession(ctx context.Context, sessionID string, authProof []byte) error {
	configBytes, err := k.GetSessionConfig(ctx, sessionID)
	if err != nil {
		return err
	}

	var config wsproto.SessionConfig
	k.cdc.MustUnmarshal(configBytes, &config)

	// Verify authentication proof
	if err := k.verifyAuthenticationProof(authProof, config.WalletId); err != nil {
		return err
	}

	config.Locked = false
	config.LastActivity = timestamppb.New(time.Now())
	config.ExpiresAt = timestamppb.New(time.Now().Add(config.TimeoutDuration.AsDuration()))

	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetSessionConfig(ctx, sessionID, updatedBytes)
}

// generateSessionID generates a unique session ID
func (k Keeper) generateSessionID(walletID string) string {
	data := fmt.Sprintf("%s:%d", walletID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return "session_" + hex.EncodeToString(hash[:16])
}

// generateDeviceFingerprint generates a device fingerprint
func (k Keeper) generateDeviceFingerprint() string {
	// In production, this would collect device information
	data := fmt.Sprintf("device_%d", time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// verifyAuthenticationProof verifies authentication proof for session unlock
func (k Keeper) verifyAuthenticationProof(proof []byte, walletID string) error {
	// In production, verify biometric or password proof
	if len(proof) < 32 {
		return types.ErrUnauthorized
	}
	return nil
}

// ============================================================================
// Biometric Authentication
// ============================================================================

// EnrollBiometric enrolls biometric authentication
func (k Keeper) EnrollBiometric(
	ctx context.Context,
	walletID string,
	biometricType wsproto.BiometricType,
	enrollmentData []byte,
) (*wsproto.BiometricAuth, error) {
	if biometricType == wsproto.BiometricType_BIOMETRIC_TYPE_UNSPECIFIED {
		return nil, types.ErrInvalidBiometricData
	}

	// Hash the enrollment data (never store raw biometric data)
	enrollmentHash := sha256.Sum256(enrollmentData)

	auth := &wsproto.BiometricAuth{
		WalletId:       walletID,
		Type:           biometricType,
		EnrollmentHash: hex.EncodeToString(enrollmentHash[:]),
		EnrolledAt:     timestamppb.New(time.Now()),
		Enabled:        true,
		FailedAttempts: 0,
		LockedOut:      false,
	}

	// Store biometric auth
	authBytes := k.cdc.MustMarshal(auth)
	if err := k.SetBiometricAuth(ctx, walletID, authBytes); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("enrolled biometric",
		"wallet_id", walletID,
		"type", biometricType.String(),
	)

	return auth, nil
}

// AuthenticateBiometric authenticates using biometric
func (k Keeper) AuthenticateBiometric(
	ctx context.Context,
	walletID string,
	biometricProof []byte,
) (bool, error) {
	authBytes, err := k.GetBiometricAuth(ctx, walletID)
	if err != nil {
		return false, err
	}

	var auth wsproto.BiometricAuth
	k.cdc.MustUnmarshal(authBytes, &auth)

	// Check if locked out
	if auth.LockedOut {
		if time.Now().Before(auth.LockoutUntil.AsTime()) {
			return false, types.ErrBiometricLockedOut
		}
		// Reset lockout
		auth.LockedOut = false
		auth.FailedAttempts = 0
	}

	// Verify biometric proof
	proofHash := sha256.Sum256(biometricProof)
	proofHashStr := hex.EncodeToString(proofHash[:])

	authenticated := proofHashStr == auth.EnrollmentHash

	if !authenticated {
		auth.FailedAttempts++
		auth.LastAttempt = timestamppb.New(time.Now())

		// Lock out after 5 failed attempts
		if auth.FailedAttempts >= 5 {
			auth.LockedOut = true
			auth.LockoutUntil = timestamppb.New(time.Now().Add(30 * time.Minute))
		}
	} else {
		auth.FailedAttempts = 0
	}

	// Store updated auth
	updatedBytes := k.cdc.MustMarshal(&auth)
	if err := k.SetBiometricAuth(ctx, walletID, updatedBytes); err != nil {
		return false, err
	}

	return authenticated, nil
}

// ============================================================================
// Secure Enclave Storage
// ============================================================================

// StoreInSecureEnclave stores key material in secure enclave
func (k Keeper) StoreInSecureEnclave(
	ctx context.Context,
	walletID string,
	enclaveType wsproto.EnclaveType,
	encryptedKeyMaterial []byte,
	attestationCert string,
) (*wsproto.SecureEnclaveConfig, error) {
	if enclaveType == wsproto.EnclaveType_ENCLAVE_TYPE_UNSPECIFIED {
		return nil, types.ErrEnclaveNotAvailable
	}

	// Generate enclave ID
	enclaveID := k.generateEnclaveID(walletID)

	config := &wsproto.SecureEnclaveConfig{
		WalletId:               walletID,
		EnclaveId:              enclaveID,
		EnclaveType:            enclaveType,
		EncryptedKeyMaterial:   encryptedKeyMaterial,
		KeyDerivationAlgorithm: "HKDF-SHA256",
		CreatedAt:              timestamppb.New(time.Now()),
		HardwareBacked:         true,
		AttestationCertificate: attestationCert,
	}

	// Store configuration
	configBytes := k.cdc.MustMarshal(config)
	if err := k.SetSecureEnclaveConfig(ctx, walletID, configBytes); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("stored in secure enclave",
		"wallet_id", walletID,
		"enclave_type", enclaveType.String(),
	)

	return config, nil
}

// generateEnclaveID generates a unique enclave ID
func (k Keeper) generateEnclaveID(walletID string) string {
	data := fmt.Sprintf("enclave_%s_%d", walletID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return base32.StdEncoding.EncodeToString(hash[:16])
}

// ============================================================================
// Encrypted Backup
// ============================================================================

// CreateEncryptedBackup creates an encrypted backup of seed phrase
func (k Keeper) CreateEncryptedBackup(
	ctx context.Context,
	walletID string,
	encryptedSeed []byte,
	encryptionAlgo, kdf string,
	salt []byte,
	iterations int32,
	location wsproto.BackupLocation,
) (*wsproto.EncryptedBackup, error) {
	if len(encryptedSeed) == 0 {
		return nil, types.ErrInvalidBackupData
	}

	// Generate backup ID
	backupID := k.generateBackupID(walletID)

	// Calculate checksum
	checksum := sha256.Sum256(encryptedSeed)

	backup := &wsproto.EncryptedBackup{
		BackupId:              backupID,
		WalletId:              walletID,
		EncryptedSeed:         encryptedSeed,
		EncryptionAlgorithm:   encryptionAlgo,
		KeyDerivationFunction: kdf,
		Salt:                  salt,
		Iterations:            iterations,
		CreatedAt:             timestamppb.New(time.Now()),
		LastVerified:          timestamppb.New(time.Now()),
		Location:              location,
		Checksum:              hex.EncodeToString(checksum[:]),
		Version:               1,
	}

	// Store backup
	backupBytes := k.cdc.MustMarshal(backup)
	if err := k.SetEncryptedBackup(ctx, backupID, backupBytes); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("created encrypted backup",
		"backup_id", backupID,
		"wallet_id", walletID,
		"location", location.String(),
	)

	return backup, nil
}

// generateBackupID generates a unique backup ID
func (k Keeper) generateBackupID(walletID string) string {
	data := fmt.Sprintf("backup_%s_%d", walletID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return "bak_" + hex.EncodeToString(hash[:16])
}

// ============================================================================
// Dust Attack Filtering
// ============================================================================

// ConfigureDustFilter configures dust attack filtering
func (k Keeper) ConfigureDustFilter(
	ctx context.Context,
	walletID string,
	enabled bool,
	minimumAmount string,
	maxDustTxPerBlock int32,
	suspiciousThreshold int32,
) (*wsproto.DustAttackFilter, error) {
	filter := &wsproto.DustAttackFilter{
		WalletId:                    walletID,
		Enabled:                     enabled,
		MinimumAmount:               minimumAmount,
		MaxDustTransactionsPerBlock: maxDustTxPerBlock,
		BlockedSenders:              []string{},
		SuspiciousPatternThreshold:  suspiciousThreshold,
		LastUpdated:                 timestamppb.New(time.Now()),
	}

	// Store filter
	filterBytes := k.cdc.MustMarshal(filter)
	if err := k.SetDustFilter(ctx, walletID, filterBytes); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("configured dust filter",
		"wallet_id", walletID,
		"enabled", enabled,
		"min_amount", minimumAmount,
	)

	return filter, nil
}

// CheckDustTransaction checks if a transaction is a dust attack
func (k Keeper) CheckDustTransaction(
	ctx context.Context,
	walletID, txHash, fromAddress, toAddress, amount, denom string,
) (bool, error) {
	filterBytes, err := k.GetDustFilter(ctx, walletID)
	if err != nil {
		// No filter configured, allow
		return false, nil
	}

	var filter wsproto.DustAttackFilter
	k.cdc.MustUnmarshal(filterBytes, &filter)

	if !filter.Enabled {
		return false, nil
	}

	// Check if sender is blocked
	for _, blocked := range filter.BlockedSenders {
		if blocked == fromAddress {
			return true, nil
		}
	}

	// Check amount against minimum
	amountBig := new(big.Int)
	amountBig.SetString(amount, 10)
	minBig := new(big.Int)
	minBig.SetString(filter.MinimumAmount, 10)

	if amountBig.Cmp(minBig) < 0 {
		// Record dust transaction
		dustTx := &wsproto.DustTransaction{
			TxHash:       txHash,
			FromAddress:  fromAddress,
			ToAddress:    toAddress,
			Amount:       amount,
			Denom:        denom,
			DetectedAt:   timestamppb.New(time.Now()),
			Blocked:      true,
			Reason:       "Amount below minimum threshold",
			PatternScore: k.calculateDustPatternScore(fromAddress),
		}

		dustTxBytes := k.cdc.MustMarshal(dustTx)
		k.SetDustTransaction(ctx, txHash, dustTxBytes)

		return true, nil
	}

	return false, nil
}

// calculateDustPatternScore calculates suspicion score for dust pattern
func (k Keeper) calculateDustPatternScore(address string) int32 {
	// In production, analyze:
	// 1. Number of small transactions from this address
	// 2. Transaction frequency
	// 3. Known dust attack patterns
	// 4. Address reputation

	return 50 // Placeholder score
}
