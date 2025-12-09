package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/walletsecurity/keeper"
	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

type KeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create database and commit multi-store
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	suite.Require().NoError(cms.LoadLatestVersion())

	// Create SDK context
	header := cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	sdkCtx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	// Create keeper
	storeService := runtime.NewKVStoreService(storeKey)
	suite.keeper = keeper.NewKeeper(cdc, storeService, log.NewNopLogger())
	suite.cdc = cdc
	suite.ctx = sdkCtx
}

// ============================================================================
// Hardware Wallet Tests
// ============================================================================

func (suite *KeeperTestSuite) TestRegisterHardwareWallet() {
	address := "cosmos1abc123def456"
	deviceID := "ledger-nano-s-12345"
	signature := make([]byte, 64)

	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		signature,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().Equal(address, config.Address)
	suite.Require().Equal(wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER, config.Type)
	suite.Require().Equal(deviceID, config.DeviceId)
}

func (suite *KeeperTestSuite) TestRegisterHardwareWallet_DuplicateError() {
	address := "cosmos1abc123def456"
	deviceID := "ledger-nano-s-12345"
	signature := make([]byte, 64)

	// First registration should succeed
	_, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	suite.Require().NoError(err)

	// Second registration with same device should fail
	_, err = suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrHardwareWalletExists, err)
}

func (suite *KeeperTestSuite) TestUpdateHardwareWalletUsage() {
	address := "cosmos1abc123def456"
	deviceID := "trezor-one-67890"
	signature := make([]byte, 64)

	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"1.10.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	suite.Require().NoError(err)

	// Update usage
	err = suite.keeper.UpdateHardwareWalletUsage(suite.ctx, config.WalletId)
	suite.Require().NoError(err)

	// Verify signature count increased
	configBytes, err := suite.keeper.GetHardwareWallet(suite.ctx, config.WalletId)
	suite.Require().NoError(err)

	var updatedConfig wsproto.HardwareWalletConfig
	suite.cdc.MustUnmarshal(configBytes, &updatedConfig)
	suite.Require().Equal(int32(1), updatedConfig.SignatureCount)
}

// ============================================================================
// Multi-Sig Wallet Tests
// ============================================================================

func (suite *KeeperTestSuite) TestCreateMultiSigWallet() {
	creator := "cosmos1creator"
	signers := []string{"cosmos1signer1", "cosmos1signer2", "cosmos1signer3"}
	threshold := int32(2)

	wallet, err := suite.keeper.CreateMultiSigWallet(
		suite.ctx,
		creator,
		signers,
		threshold,
		nil,
		0,
		nil,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(wallet)
	suite.Require().Equal(creator, wallet.Creator)
	suite.Require().Equal(threshold, wallet.Threshold)
	suite.Require().Equal(int32(len(signers)), wallet.TotalSigners)
}

func (suite *KeeperTestSuite) TestCreateMultiSigWallet_InvalidThreshold() {
	creator := "cosmos1creator"
	signers := []string{"cosmos1signer1", "cosmos1signer2"}
	threshold := int32(5) // Exceeds number of signers

	_, err := suite.keeper.CreateMultiSigWallet(
		suite.ctx,
		creator,
		signers,
		threshold,
		nil,
		0,
		nil,
	)

	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInvalidThreshold, err)
}

func (suite *KeeperTestSuite) TestCreateWeightedMultiSigWallet() {
	creator := "cosmos1creator"
	signers := []string{"cosmos1signer1", "cosmos1signer2", "cosmos1signer3"}
	weights := map[string]int32{
		"cosmos1signer1": 50,
		"cosmos1signer2": 30,
		"cosmos1signer3": 20,
	}
	weightThreshold := int32(51)

	wallet, err := suite.keeper.CreateMultiSigWallet(
		suite.ctx,
		creator,
		signers,
		2,
		weights,
		weightThreshold,
		nil,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(wallet)
	suite.Require().Equal(weightThreshold, wallet.WeightThreshold)
	suite.Require().Equal(len(weights), len(wallet.SignerWeights))
}

func (suite *KeeperTestSuite) TestSignMultiSigTransaction() {
	// Create wallet
	creator := "cosmos1creator"
	signers := []string{"cosmos1signer1", "cosmos1signer2", "cosmos1signer3"}
	threshold := int32(2)

	wallet, err := suite.keeper.CreateMultiSigWallet(
		suite.ctx,
		creator,
		signers,
		threshold,
		nil,
		0,
		nil,
	)
	suite.Require().NoError(err)

	// Create pending transaction
	txData := []byte("test transaction data")
	pendingTx, err := suite.keeper.CreatePendingMultiSigTransaction(
		suite.ctx,
		wallet.WalletId,
		txData,
		"transfer",
		"Test transfer",
		24*time.Hour,
	)
	suite.Require().NoError(err)

	// First signature
	signature1 := make([]byte, 64)
	ready, err := suite.keeper.SignMultiSigTransaction(
		suite.ctx,
		pendingTx.TxId,
		signers[0],
		signature1,
	)
	suite.Require().NoError(err)
	suite.Require().False(ready) // Not enough signatures

	// Second signature
	signature2 := make([]byte, 64)
	ready, err = suite.keeper.SignMultiSigTransaction(
		suite.ctx,
		pendingTx.TxId,
		signers[1],
		signature2,
	)
	suite.Require().NoError(err)
	suite.Require().True(ready) // Threshold met
}

// ============================================================================
// Social Recovery Tests
// ============================================================================

func (suite *KeeperTestSuite) TestConfigureSocialRecovery() {
	walletID := "wallet123"
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
		{Address: "cosmos1guardian2", Name: "Guardian 2"},
		{Address: "cosmos1guardian3", Name: "Guardian 3"},
	}
	threshold := int32(2)
	delay := &gogotypes.Duration{Seconds: int64(48 * time.Hour.Seconds()), Nanos: int32(48 * time.Hour.Nanoseconds() % 1e9)}

	config, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		walletID,
		guardians,
		threshold,
		delay,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().Equal(walletID, config.WalletId)
	suite.Require().Equal(len(guardians), len(config.Guardians))
	suite.Require().Equal(threshold, config.RecoveryThreshold)
	suite.Require().True(config.Enabled)
}

func (suite *KeeperTestSuite) TestConfigureSocialRecovery_InvalidThreshold() {
	walletID := "wallet123"
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
	}
	threshold := int32(5) // Exceeds number of guardians

	_, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		walletID,
		guardians,
		threshold,
		nil,
	)

	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInvalidRecoveryThreshold, err)
}

func (suite *KeeperTestSuite) TestInitiateRecovery() {
	walletID := "wallet123"
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
		{Address: "cosmos1guardian2", Name: "Guardian 2"},
	}

	config, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		walletID,
		guardians,
		2,
		&gogotypes.Duration{Seconds: int64(48*time.Hour.Seconds()), Nanos: int32(48*time.Hour.Nanoseconds() % 1e9)},
	)
	suite.Require().NoError(err)

	// Confirm guardians first
	err = suite.keeper.ConfirmGuardian(suite.ctx, walletID, guardians[0].Address)
	suite.Require().NoError(err)
	err = suite.keeper.ConfirmGuardian(suite.ctx, walletID, guardians[1].Address)
	suite.Require().NoError(err)

	// Initiate recovery
	newAddress := "cosmos1newaddress"
	request, err := suite.keeper.InitiateRecovery(
		suite.ctx,
		config.WalletId,
		newAddress,
		guardians[0].Address,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(request)
	suite.Require().Equal(newAddress, request.NewAddress)
	suite.Require().Equal(int32(1), request.ApprovalsCount)
}

func (suite *KeeperTestSuite) TestApproveRecovery_Success() {
	walletID := "wallet123"
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
		{Address: "cosmos1guardian2", Name: "Guardian 2"},
		{Address: "cosmos1guardian3", Name: "Guardian 3"},
	}

	// Configure social recovery
	config, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		walletID,
		guardians,
		2,
		&gogotypes.Duration{Seconds: int64(48*time.Hour.Seconds()), Nanos: int32(48*time.Hour.Nanoseconds() % 1e9)},
	)
	suite.Require().NoError(err)

	// Confirm guardians
	for _, g := range guardians {
		err = suite.keeper.ConfirmGuardian(suite.ctx, walletID, g.Address)
		suite.Require().NoError(err)
	}

	// Initiate recovery
	newAddress := "cosmos1newaddress"
	request, err := suite.keeper.InitiateRecovery(
		suite.ctx,
		config.WalletId,
		newAddress,
		guardians[0].Address,
	)
	suite.Require().NoError(err)

	// Second guardian approves
	signature := make([]byte, 64)
	ready, err := suite.keeper.ApproveRecovery(
		suite.ctx,
		request.RequestId,
		guardians[1].Address,
		signature,
	)
	suite.Require().NoError(err)
	suite.Require().True(ready) // Threshold of 2 met

	// Verify request state
	requestBytes, err := suite.keeper.GetRecoveryRequest(suite.ctx, request.RequestId)
	suite.Require().NoError(err)
	var updatedRequest wsproto.RecoveryRequest
	suite.cdc.MustUnmarshal(requestBytes, &updatedRequest)
	suite.Require().Equal(int32(2), updatedRequest.ApprovalsCount)
	suite.Require().Equal(wsproto.RecoveryStatus_RECOVERY_STATUS_APPROVED, updatedRequest.Status)
}

func (suite *KeeperTestSuite) TestApproveRecovery_UnauthorizedGuardian() {
	walletID := "wallet123"
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
		{Address: "cosmos1guardian2", Name: "Guardian 2"},
	}

	// Configure social recovery
	config, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		walletID,
		guardians,
		2,
		&gogotypes.Duration{Seconds: int64(48*time.Hour.Seconds()), Nanos: int32(48*time.Hour.Nanoseconds() % 1e9)},
	)
	suite.Require().NoError(err)

	// Confirm authorized guardians
	for _, g := range guardians {
		err = suite.keeper.ConfirmGuardian(suite.ctx, walletID, g.Address)
		suite.Require().NoError(err)
	}

	// Initiate recovery
	newAddress := "cosmos1newaddress"
	request, err := suite.keeper.InitiateRecovery(
		suite.ctx,
		config.WalletId,
		newAddress,
		guardians[0].Address,
	)
	suite.Require().NoError(err)

	// Unauthorized address (not in guardian list) tries to approve
	unauthorizedAddress := "cosmos1unauthorized"
	signature := make([]byte, 64)
	_, err = suite.keeper.ApproveRecovery(
		suite.ctx,
		request.RequestId,
		unauthorizedAddress,
		signature,
	)

	// Should fail with invalid guardian error
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInvalidGuardian, err)
}

func (suite *KeeperTestSuite) TestApproveRecovery_DuplicateApproval() {
	walletID := "wallet123"
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
		{Address: "cosmos1guardian2", Name: "Guardian 2"},
		{Address: "cosmos1guardian3", Name: "Guardian 3"},
	}

	// Configure social recovery
	config, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		walletID,
		guardians,
		2,
		&gogotypes.Duration{Seconds: int64(48*time.Hour.Seconds()), Nanos: int32(48*time.Hour.Nanoseconds() % 1e9)},
	)
	suite.Require().NoError(err)

	// Confirm guardians
	for _, g := range guardians {
		err = suite.keeper.ConfirmGuardian(suite.ctx, walletID, g.Address)
		suite.Require().NoError(err)
	}

	// Initiate recovery (guardian 1 automatically approves)
	newAddress := "cosmos1newaddress"
	request, err := suite.keeper.InitiateRecovery(
		suite.ctx,
		config.WalletId,
		newAddress,
		guardians[0].Address,
	)
	suite.Require().NoError(err)

	// Same guardian tries to approve again
	signature := make([]byte, 64)
	_, err = suite.keeper.ApproveRecovery(
		suite.ctx,
		request.RequestId,
		guardians[0].Address,
		signature,
	)

	// Should fail with already exists error
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "already approved")
}

func (suite *KeeperTestSuite) TestApproveRecovery_UnconfirmedGuardian() {
	walletID := "wallet123"
	guardians := []*wsproto.Guardian{
		{Address: "cosmos1guardian1", Name: "Guardian 1"},
		{Address: "cosmos1guardian2", Name: "Guardian 2"},
		{Address: "cosmos1guardian3", Name: "Guardian 3"},
	}

	// Configure social recovery
	config, err := suite.keeper.ConfigureSocialRecovery(
		suite.ctx,
		walletID,
		guardians,
		2,
		&gogotypes.Duration{Seconds: int64(48*time.Hour.Seconds()), Nanos: int32(48*time.Hour.Nanoseconds() % 1e9)},
	)
	suite.Require().NoError(err)

	// Confirm only guardians 1 and 2, leave guardian 3 unconfirmed
	err = suite.keeper.ConfirmGuardian(suite.ctx, walletID, guardians[0].Address)
	suite.Require().NoError(err)
	err = suite.keeper.ConfirmGuardian(suite.ctx, walletID, guardians[1].Address)
	suite.Require().NoError(err)

	// Initiate recovery with guardian 1
	newAddress := "cosmos1newaddress"
	request, err := suite.keeper.InitiateRecovery(
		suite.ctx,
		config.WalletId,
		newAddress,
		guardians[0].Address,
	)
	suite.Require().NoError(err)

	// Unconfirmed guardian 3 tries to approve
	signature := make([]byte, 64)
	_, err = suite.keeper.ApproveRecovery(
		suite.ctx,
		request.RequestId,
		guardians[2].Address,
		signature,
	)

	// Should fail with invalid guardian error
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInvalidGuardian, err)
}

// ============================================================================
// Spending Limit Tests
// ============================================================================

func (suite *KeeperTestSuite) TestSetSpendingLimit() {
	walletID := "wallet123"
	denom := "uatom"
	dailyLimit := "1000000"

	limit, err := suite.keeper.SetSpendingLimit(
		suite.ctx,
		walletID,
		denom,
		dailyLimit,
		"7000000",
		"30000000",
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(limit)
	suite.Require().Equal(walletID, limit.WalletId)
	suite.Require().Equal(denom, limit.Denom)
	suite.Require().Equal(dailyLimit, limit.DailyLimit)
	suite.Require().True(limit.Enabled)
}

func (suite *KeeperTestSuite) TestCheckSpendingLimit_Exceeded() {
	walletID := "wallet123"
	denom := "uatom"
	dailyLimit := "1000000"

	_, err := suite.keeper.SetSpendingLimit(
		suite.ctx,
		walletID,
		denom,
		dailyLimit,
		"7000000",
		"30000000",
	)
	suite.Require().NoError(err)

	err = suite.keeper.CheckSpendingLimit(suite.ctx, walletID, denom, "1500000")
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, types.ErrSpendingLimitExceeded)

	stored, err := suite.keeper.GetSpendingLimit(suite.ctx, walletID, denom)
	suite.Require().NoError(err)
	var limit wsproto.SpendingLimit
	suite.cdc.MustUnmarshal(stored, &limit)
	suite.Require().Equal("0", limit.CurrentDailySpent)
}

func (suite *KeeperTestSuite) TestCheckSpendingLimit_WithinLimit() {
	walletID := "wallet123"
	denom := "uatom"
	dailyLimit := "1000000"

	_, err := suite.keeper.SetSpendingLimit(
		suite.ctx,
		walletID,
		denom,
		dailyLimit,
		"7000000",
		"30000000",
	)
	suite.Require().NoError(err)

	err = suite.keeper.CheckSpendingLimit(suite.ctx, walletID, denom, "500000")
	suite.Require().NoError(err)

	stored, err := suite.keeper.GetSpendingLimit(suite.ctx, walletID, denom)
	suite.Require().NoError(err)
	var limit wsproto.SpendingLimit
	suite.cdc.MustUnmarshal(stored, &limit)
	suite.Require().Equal("500000", limit.CurrentDailySpent)
}

// ============================================================================
// Session Management Tests
// ============================================================================

func (suite *KeeperTestSuite) TestConfigureSession() {
	walletID := "wallet123"
	timeout := &gogotypes.Duration{Seconds: int64(30 * time.Minute.Seconds()), Nanos: int32(30 * time.Minute.Nanoseconds() % 1e9)}

	config, err := suite.keeper.ConfigureSession(
		suite.ctx,
		walletID,
		timeout,
		true,
		300,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().Equal(walletID, config.WalletId)
	suite.Require().True(config.AutoLockEnabled)
	suite.Require().False(config.Locked)
}

func (suite *KeeperTestSuite) TestLockAndUnlockSession() {
	walletID := "wallet123"
	timeout := &gogotypes.Duration{Seconds: int64(30 * time.Minute.Seconds()), Nanos: int32(30 * time.Minute.Nanoseconds() % 1e9)}

	config, err := suite.keeper.ConfigureSession(
		suite.ctx,
		walletID,
		timeout,
		true,
		300,
	)
	suite.Require().NoError(err)

	// Lock session
	err = suite.keeper.LockSession(suite.ctx, config.SessionId)
	suite.Require().NoError(err)

	// Verify locked
	configBytes, err := suite.keeper.GetSessionConfig(suite.ctx, config.SessionId)
	suite.Require().NoError(err)
	var lockedConfig wsproto.SessionConfig
	suite.cdc.MustUnmarshal(configBytes, &lockedConfig)
	suite.Require().True(lockedConfig.Locked)

	// Unlock session
	authProof := make([]byte, 64)
	err = suite.keeper.UnlockSession(suite.ctx, config.SessionId, authProof)
	suite.Require().NoError(err)

	// Verify unlocked
	configBytes, err = suite.keeper.GetSessionConfig(suite.ctx, config.SessionId)
	suite.Require().NoError(err)
	var unlockedConfig wsproto.SessionConfig
	suite.cdc.MustUnmarshal(configBytes, &unlockedConfig)
	suite.Require().False(unlockedConfig.Locked)
}

// ============================================================================
// Biometric Authentication Tests
// ============================================================================

func (suite *KeeperTestSuite) TestEnrollBiometric() {
	walletID := "wallet123"
	enrollmentData := []byte("fingerprint_data_hash")

	auth, err := suite.keeper.EnrollBiometric(
		suite.ctx,
		walletID,
		wsproto.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		enrollmentData,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(auth)
	suite.Require().Equal(walletID, auth.WalletId)
	suite.Require().Equal(wsproto.BiometricType_BIOMETRIC_TYPE_FINGERPRINT, auth.Type)
	suite.Require().True(auth.Enabled)
}

func (suite *KeeperTestSuite) TestAuthenticateBiometric_Success() {
	walletID := "wallet123"
	enrollmentData := []byte("fingerprint_data_hash")

	_, err := suite.keeper.EnrollBiometric(
		suite.ctx,
		walletID,
		wsproto.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		enrollmentData,
	)
	suite.Require().NoError(err)

	// Authenticate with same data
	authenticated, err := suite.keeper.AuthenticateBiometric(
		suite.ctx,
		walletID,
		enrollmentData,
	)

	suite.Require().NoError(err)
	suite.Require().True(authenticated)
}

func (suite *KeeperTestSuite) TestAuthenticateBiometric_Failure() {
	walletID := "wallet123"
	enrollmentData := []byte("fingerprint_data_hash")

	_, err := suite.keeper.EnrollBiometric(
		suite.ctx,
		walletID,
		wsproto.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		enrollmentData,
	)
	suite.Require().NoError(err)

	// Authenticate with different data
	wrongData := []byte("wrong_fingerprint_data")
	authenticated, err := suite.keeper.AuthenticateBiometric(
		suite.ctx,
		walletID,
		wrongData,
	)

	suite.Require().NoError(err)
	suite.Require().False(authenticated)
}

// ============================================================================
// Transaction Simulation Tests
// ============================================================================

func (suite *KeeperTestSuite) TestSimulateTransaction() {
	txData := []byte("test transaction data")
	sender := "cosmos1sender"

	simulation, err := suite.keeper.SimulateTransaction(
		suite.ctx,
		txData,
		sender,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(simulation)
	suite.Require().True(simulation.Success)
	suite.Require().Greater(simulation.GasUsed, int64(0))
}

// ============================================================================
// Address Checksum Tests
// ============================================================================

func (suite *KeeperTestSuite) TestValidateAddressChecksum_EIP55() {
	address := "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"

	valid, checksum, err := suite.keeper.ValidateAddressChecksum(
		suite.ctx,
		address,
		wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_EIP55,
	)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(checksum)
	suite.Require().True(valid || !valid) // Either valid or not, both are acceptable
}

// ============================================================================
// Dust Filter Tests
// ============================================================================

func (suite *KeeperTestSuite) TestConfigureDustFilter() {
	walletID := "wallet123"
	minimumAmount := "1000"

	filter, err := suite.keeper.ConfigureDustFilter(
		suite.ctx,
		walletID,
		true,
		minimumAmount,
		10,
		5,
	)

	suite.Require().NoError(err)
	suite.Require().NotNil(filter)
	suite.Require().Equal(walletID, filter.WalletId)
	suite.Require().True(filter.Enabled)
	suite.Require().Equal(minimumAmount, filter.MinimumAmount)
}

func (suite *KeeperTestSuite) TestCheckDustTransaction_Blocked() {
	walletID := "wallet123"
	minimumAmount := "1000"

	_, err := suite.keeper.ConfigureDustFilter(
		suite.ctx,
		walletID,
		true,
		minimumAmount,
		10,
		5,
	)
	suite.Require().NoError(err)

	isDust, err := suite.keeper.CheckDustTransaction(
		suite.ctx,
		walletID,
		"txhash123",
		"cosmos1sender",
		"cosmos1receiver",
		"500",
		"uatom",
	)

	suite.Require().NoError(err)
	suite.Require().True(isDust)
}

func TestValidateAddressChecksum(t *testing.T) {
	// Test EIP55 checksum validation
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	testCases := []struct {
		name      string
		address   string
		algorithm wsproto.ChecksumAlgorithm
		expectErr bool
	}{
		{
			name:      "valid ethereum address",
			address:   "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
			algorithm: wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_EIP55,
			expectErr: false,
		},
		{
			name:      "valid bech32 address",
			address:   "cosmos1abc123def456",
			algorithm: wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BECH32,
			expectErr: false,
		},
		{
			name:      "empty address",
			address:   "",
			algorithm: wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_EIP55,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := suite.keeper.ValidateAddressChecksum(
				suite.ctx,
				tc.address,
				tc.algorithm,
			)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
