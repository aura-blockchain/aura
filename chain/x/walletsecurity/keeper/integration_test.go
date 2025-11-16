package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// TestCompleteWalletSecurityWorkflow demonstrates a complete security workflow
func TestCompleteWalletSecurityWorkflow(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetupTest()

	// 1. Register hardware wallet
	address := "cosmos1secure123"
	deviceID := "ledger-nano-x-12345"
	signature := make([]byte, 64)

	hwConfig, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	require.NoError(t, err)
	require.NotNil(t, hwConfig)
	t.Logf("Hardware wallet registered: %s", hwConfig.WalletId)

	// 2. Create multi-sig wallet with the hardware wallet
	creator := address
	signers := []string{address, "cosmos1guardian1", "cosmos1guardian2"}

	multiSigWallet, err := suite.keeper.CreateMultiSigWallet(
		suite.ctx,
		creator,
		signers,
		2, // 2-of-3 threshold
		nil,
		0,
		nil,
	)
	require.NoError(t, err)
	require.NotNil(t, multiSigWallet)
	t.Logf("Multi-sig wallet created: %s", multiSigWallet.WalletId)

	// 3. Configure social recovery
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian One"},
		{Address: "cosmos1guardian2", Name: "Guardian Two"},
		{Address: "cosmos1guardian3", Name: "Guardian Three"},
	}

	recoveryConfig, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		multiSigWallet.WalletId,
		guardians,
		2, // 2-of-3 threshold
		durationpb.New(48*time.Hour),
	)
	require.NoError(t, err)
	require.NotNil(t, recoveryConfig)
	t.Logf("Social recovery configured with %d guardians", len(guardians))

	// 4. Set spending limits
	spendingLimit, err := suite.keeper.SetSpendingLimit(
		suite.ctx,
		multiSigWallet.WalletId,
		"uatom",
		"1000000",  // 1 ATOM daily
		"7000000",  // 7 ATOM weekly
		"30000000", // 30 ATOM monthly
	)
	require.NoError(t, err)
	require.NotNil(t, spendingLimit)
	t.Logf("Spending limits set for %s", spendingLimit.Denom)

	// 5. Configure session with auto-lock
	sessionConfig, err := suite.keeper.ConfigureSession(
		suite.ctx,
		multiSigWallet.WalletId,
		durationpb.New(30*time.Minute),
		true, // auto-lock enabled
		300,  // 5 minutes inactivity
	)
	require.NoError(t, err)
	require.NotNil(t, sessionConfig)
	t.Logf("Session configured: %s", sessionConfig.SessionId)

	// 6. Enroll biometric authentication
	biometricData := []byte("fingerprint_template_hash")
	biometricAuth, err := suite.keeper.EnrollBiometric(
		suite.ctx,
		multiSigWallet.WalletId,
		wsproto.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		biometricData,
	)
	require.NoError(t, err)
	require.NotNil(t, biometricAuth)
	t.Logf("Biometric enrolled: %s", biometricAuth.Type.String())

	// 7. Store in secure enclave
	encryptedKey := make([]byte, 256)
	enclaveConfig, err := suite.keeper.StoreInSecureEnclave(
		suite.ctx,
		multiSigWallet.WalletId,
		wsproto.EnclaveType_ENCLAVE_TYPE_TEE,
		encryptedKey,
		"attestation_cert_hash",
	)
	require.NoError(t, err)
	require.NotNil(t, enclaveConfig)
	t.Logf("Key stored in secure enclave: %s", enclaveConfig.EnclaveId)

	// 8. Create encrypted backup
	encryptedSeed := make([]byte, 512)
	salt := make([]byte, 32)
	backup, err := suite.keeper.CreateEncryptedBackup(
		suite.ctx,
		multiSigWallet.WalletId,
		encryptedSeed,
		"AES-256-GCM",
		"PBKDF2-SHA256",
		salt,
		100000,
		wsproto.BackupLocation_BACKUP_LOCATION_CLOUD,
	)
	require.NoError(t, err)
	require.NotNil(t, backup)
	t.Logf("Encrypted backup created: %s", backup.BackupId)

	// 9. Configure dust filter
	dustFilter, err := suite.keeper.ConfigureDustFilter(
		suite.ctx,
		multiSigWallet.WalletId,
		true,
		"1000", // minimum 0.001 ATOM
		10,
		5,
	)
	require.NoError(t, err)
	require.NotNil(t, dustFilter)
	t.Logf("Dust filter configured")

	// 10. Simulate a transaction
	txData := []byte("transfer 100 ATOM to cosmos1receiver")
	simulation, err := suite.keeper.SimulateTransaction(
		suite.ctx,
		txData,
		address,
	)
	require.NoError(t, err)
	require.NotNil(t, simulation)
	require.True(t, simulation.Success)
	t.Logf("Transaction simulated - Risk level: %s", simulation.RiskLevel.String())

	// 11. Check spending limit before transaction
	err = suite.keeper.CheckSpendingLimit(
		suite.ctx,
		multiSigWallet.WalletId,
		"uatom",
		"500000", // 0.5 ATOM - within limit
	)
	require.NoError(t, err)
	t.Logf("Spending limit check passed")

	// 12. Validate address checksum
	testAddress := "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"
	valid, checksum, err := suite.keeper.ValidateAddressChecksum(
		suite.ctx,
		testAddress,
		wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_EIP55,
	)
	require.NoError(t, err)
	require.NotEmpty(t, checksum)
	t.Logf("Address checksum validated: %v", valid)

	t.Log("Complete wallet security workflow successful!")
}

// TestMultiSigWithRecoveryWorkflow demonstrates multi-sig with social recovery
func TestMultiSigWithRecoveryWorkflow(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetupTest()

	// Create multi-sig wallet
	creator := "cosmos1creator"
	signers := []string{"cosmos1signer1", "cosmos1signer2", "cosmos1signer3"}

	wallet, err := suite.keeper.CreateMultiSigWallet(
		suite.ctx,
		creator,
		signers,
		2,
		nil,
		0,
		nil,
	)
	require.NoError(t, err)

	// Configure social recovery
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
		{Address: "cosmos1guardian2", Name: "Guardian 2"},
	}

	_, err = suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		wallet.WalletId,
		guardians,
		2,
		durationpb.New(1*time.Second), // Short delay for testing
	)
	require.NoError(t, err)

	// Confirm guardians
	err = suite.keeper.ConfirmGuardian(suite.ctx, wallet.WalletId, guardians[0].Address)
	require.NoError(t, err)
	err = suite.keeper.ConfirmGuardian(suite.ctx, wallet.WalletId, guardians[1].Address)
	require.NoError(t, err)

	// Simulate lost access - initiate recovery
	newAddress := "cosmos1recovered"
	request, err := suite.keeper.InitiateRecovery(
		suite.ctx,
		wallet.WalletId,
		newAddress,
		guardians[0].Address,
	)
	require.NoError(t, err)
	t.Logf("Recovery initiated: %s", request.RequestId)

	// Second guardian approves
	signature := make([]byte, 64)
	ready, err := suite.keeper.ApproveRecovery(
		suite.ctx,
		request.RequestId,
		guardians[1].Address,
		signature,
	)
	require.NoError(t, err)
	require.True(t, ready)
	t.Logf("Recovery approved and ready for execution")

	// Wait for delay (in production, this would be 48 hours)
	time.Sleep(2 * time.Second)

	// Execute recovery
	err = suite.keeper.ExecuteRecovery(suite.ctx, request.RequestId)
	require.NoError(t, err)
	t.Log("Recovery executed successfully")
}

// TestSecurityLayeredDefense demonstrates layered security approach
func TestSecurityLayeredDefense(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetupTest()

	walletID := "secure_wallet_123"

	// Layer 1: Hardware wallet
	hwConfig, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		"cosmos1secure",
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		"device123",
		"2.0.0",
		"m/44'/118'/0'/0/0",
		make([]byte, 64),
	)
	require.NoError(t, err)
	t.Log("Layer 1: Hardware wallet - ACTIVE")

	// Layer 2: Spending limits
	_, err = suite.keeper.SetSpendingLimit(
		suite.ctx,
		walletID,
		"uatom",
		"1000000",
		"7000000",
		"30000000",
	)
	require.NoError(t, err)
	t.Log("Layer 2: Spending limits - ACTIVE")

	// Layer 3: Session timeout
	_, err = suite.keeper.ConfigureSession(
		suite.ctx,
		walletID,
		durationpb.New(15*time.Minute),
		true,
		180,
	)
	require.NoError(t, err)
	t.Log("Layer 3: Session timeout - ACTIVE")

	// Layer 4: Biometric authentication
	_, err = suite.keeper.EnrollBiometric(
		suite.ctx,
		walletID,
		wsproto.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		[]byte("biometric_hash"),
	)
	require.NoError(t, err)
	t.Log("Layer 4: Biometric auth - ACTIVE")

	// Layer 5: Dust filter
	_, err = suite.keeper.ConfigureDustFilter(
		suite.ctx,
		walletID,
		true,
		"1000",
		10,
		5,
	)
	require.NoError(t, err)
	t.Log("Layer 5: Dust filter - ACTIVE")

	// Layer 6: Transaction simulation (risk analysis)
	simulation, err := suite.keeper.SimulateTransaction(
		suite.ctx,
		[]byte("suspicious_transaction"),
		"cosmos1secure",
	)
	require.NoError(t, err)
	t.Logf("Layer 6: Transaction simulation - Risk: %s", simulation.RiskLevel.String())

	t.Log("All security layers successfully activated!")
	t.Logf("Wallet %s is now maximally protected", hwConfig.WalletId)
}

// TestBiometricLockout tests biometric authentication lockout
func TestBiometricLockout(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetupTest()

	walletID := "biometric_test_wallet"
	correctData := []byte("correct_biometric_data")
	wrongData := []byte("wrong_biometric_data")

	// Enroll biometric
	_, err := suite.keeper.EnrollBiometric(
		suite.ctx,
		walletID,
		wsproto.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		correctData,
	)
	require.NoError(t, err)

	// Attempt with correct data - should succeed
	authenticated, err := suite.keeper.AuthenticateBiometric(suite.ctx, walletID, correctData)
	require.NoError(t, err)
	require.True(t, authenticated)
	t.Log("Correct biometric authentication succeeded")

	// Attempt 5 times with wrong data - should trigger lockout
	for i := 0; i < 5; i++ {
		authenticated, err = suite.keeper.AuthenticateBiometric(suite.ctx, walletID, wrongData)
		require.NoError(t, err)
		require.False(t, authenticated)
		t.Logf("Failed attempt %d/5", i+1)
	}

	// Next attempt should return lockout error
	authBytes, err := suite.keeper.GetBiometricAuth(suite.ctx, walletID)
	require.NoError(t, err)

	var auth wsproto.BiometricAuth
	suite.cdc.MustUnmarshal(authBytes, &auth)
	require.True(t, auth.LockedOut)
	t.Log("Biometric authentication locked out after 5 failed attempts")
}

// TestSpendingLimitEnforcement tests spending limit enforcement
func TestSpendingLimitEnforcement(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetupTest()

	walletID := "spending_test_wallet"
	denom := "uatom"

	// Set daily limit of 1 ATOM (1,000,000 uatom)
	_, err := suite.keeper.SetSpendingLimit(
		suite.ctx,
		walletID,
		denom,
		"1000000",
		"7000000",
		"30000000",
	)
	require.NoError(t, err)

	// Spend 0.5 ATOM - should succeed
	err = suite.keeper.CheckSpendingLimit(suite.ctx, walletID, denom, "500000")
	require.NoError(t, err)
	t.Log("First spend (0.5 ATOM) - PASSED")

	// Spend another 0.3 ATOM - should succeed (total 0.8)
	err = suite.keeper.CheckSpendingLimit(suite.ctx, walletID, denom, "300000")
	require.NoError(t, err)
	t.Log("Second spend (0.3 ATOM) - PASSED")

	// Try to spend 0.5 ATOM - should fail (would exceed 1 ATOM limit)
	err = suite.keeper.CheckSpendingLimit(suite.ctx, walletID, denom, "500000")
	require.Error(t, err)
	t.Log("Third spend (0.5 ATOM) - BLOCKED (would exceed daily limit)")
}

// TestDustAttackProtection tests dust attack filtering
func TestDustAttackProtection(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetupTest()

	walletID := "dust_protection_wallet"

	// Configure dust filter with 1000 uatom minimum
	_, err := suite.keeper.ConfigureDustFilter(
		suite.ctx,
		walletID,
		true,
		"1000",
		5,
		3,
	)
	require.NoError(t, err)

	// Normal transaction (above threshold) - should pass
	isDust, err := suite.keeper.CheckDustTransaction(
		suite.ctx,
		walletID,
		"tx_normal_123",
		"cosmos1sender",
		"cosmos1receiver",
		"50000", // 0.05 ATOM
		"uatom",
	)
	require.NoError(t, err)
	require.False(t, isDust)
	t.Log("Normal transaction (50000 uatom) - ALLOWED")

	// Dust transaction (below threshold) - should be blocked
	isDust, err = suite.keeper.CheckDustTransaction(
		suite.ctx,
		walletID,
		"tx_dust_456",
		"cosmos1attacker",
		"cosmos1receiver",
		"500", // 0.0005 ATOM (below minimum)
		"uatom",
	)
	require.NoError(t, err)
	require.True(t, isDust)
	t.Log("Dust transaction (500 uatom) - BLOCKED")
}
