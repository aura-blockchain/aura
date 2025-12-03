package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	wspb "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

type MsgServerTestSuite struct {
	KeeperTestSuite
	msgServer wspb.MsgServer
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (suite *MsgServerTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.msgServer = NewMsgServerImpl(&suite.keeper)
}

func (suite *MsgServerTestSuite) TestMsgServerImplementation() {
	suite.NotNil(suite.msgServer, "msg server should be created")
}

// TestSignMultiSigTransaction_UnauthorizedSigner tests that unauthorized signers are rejected
func (suite *MsgServerTestSuite) TestSignMultiSigTransaction_UnauthorizedSigner() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Create test addresses
	creator := sdk.AccAddress("creator_____________")
	authorizedSigner1 := sdk.AccAddress("authorized1_________")
	authorizedSigner2 := sdk.AccAddress("authorized2_________")
	unauthorizedSigner := sdk.AccAddress("unauthorized________")

	// Create a multi-sig wallet
	wallet := &wspb.MultiSigWallet{
		WalletId:     "test_wallet_1",
		Signers:      []string{authorizedSigner1.String(), authorizedSigner2.String()},
		Threshold:    2,
		TotalSigners: 2,
		CreatedAt:    timestamppb.Now(),
		Creator:      creator.String(),
	}

	walletBytes, err := suite.cdc.Marshal(wallet)
	suite.Require().NoError(err)
	err = suite.keeper.SetMultiSigWallet(ctx, "test_wallet_1", walletBytes)
	suite.Require().NoError(err)

	// Create a pending transaction
	tx := &wspb.PendingMultiSigTransaction{
		TxId:          "test_tx_1",
		WalletId:      "test_wallet_1",
		TxData:        []byte("test_data"),
		Signatures:    []string{},
		SignedBy:      []string{},
		CurrentWeight: 0,
		CreatedAt:     timestamppb.Now(),
	}

	txBytes, err := suite.cdc.Marshal(tx)
	suite.Require().NoError(err)
	err = suite.keeper.SetPendingMultiSigTx(ctx, "test_tx_1", txBytes)
	suite.Require().NoError(err)

	// Attempt to sign with unauthorized signer
	msg := &wspb.MsgSignMultiSigTransaction{
		TxId:      "test_tx_1",
		Signer:    unauthorizedSigner.String(),
		Signature: []byte("signature"),
	}

	// Note: In real scenario, GetSigners() would return the transaction signer
	// For this test, we're testing the authorization check, not the transaction signer verification
	// The actual implementation would reject based on GetSigners() first

	_, err = suite.msgServer.SignMultiSigTransaction(suite.ctx, msg)
	suite.Require().Error(err)

	// Check that it's a permission denied error
	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.PermissionDenied, st.Code())
	suite.Contains(st.Message(), "signer not authorized for this wallet")
}

// TestSignMultiSigTransaction_DuplicateSignature tests that duplicate signatures are rejected
func (suite *MsgServerTestSuite) TestSignMultiSigTransaction_DuplicateSignature() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Create test addresses
	creator := sdk.AccAddress("creator_____________")
	signer1 := sdk.AccAddress("signer1_____________")
	signer2 := sdk.AccAddress("signer2_____________")

	// Create a multi-sig wallet
	wallet := &wspb.MultiSigWallet{
		WalletId:     "test_wallet_2",
		Signers:      []string{signer1.String(), signer2.String()},
		Threshold:    2,
		TotalSigners: 2,
		CreatedAt:    timestamppb.Now(),
		Creator:      creator.String(),
	}

	walletBytes, err := suite.cdc.Marshal(wallet)
	suite.Require().NoError(err)
	err = suite.keeper.SetMultiSigWallet(ctx, "test_wallet_2", walletBytes)
	suite.Require().NoError(err)

	// Create a pending transaction with one signature already
	tx := &wspb.PendingMultiSigTransaction{
		TxId:          "test_tx_2",
		WalletId:      "test_wallet_2",
		TxData:        []byte("test_data"),
		Signatures:    []string{"signature1"},
		SignedBy:      []string{signer1.String()},
		CurrentWeight: 1,
		CreatedAt:     timestamppb.Now(),
	}

	txBytes, err := suite.cdc.Marshal(tx)
	suite.Require().NoError(err)
	err = suite.keeper.SetPendingMultiSigTx(ctx, "test_tx_2", txBytes)
	suite.Require().NoError(err)

	// Attempt to sign again with the same signer
	msg := &wspb.MsgSignMultiSigTransaction{
		TxId:      "test_tx_2",
		Signer:    signer1.String(),
		Signature: []byte("signature2"),
	}

	_, err = suite.msgServer.SignMultiSigTransaction(suite.ctx, msg)
	suite.Require().Error(err)

	// Check that it's an already exists error
	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.AlreadyExists, st.Code())
	suite.Contains(st.Message(), "signer already signed this transaction")
}

// TestSignMultiSigTransaction_WeightedMultiSig tests weighted multi-sig functionality
func (suite *MsgServerTestSuite) TestSignMultiSigTransaction_WeightedMultiSig() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Create test addresses
	creator := sdk.AccAddress("creator_____________")
	heavySigner := sdk.AccAddress("heavy_______________") // Weight 3
	normalSigner1 := sdk.AccAddress("normal1_____________") // Weight 1 (default)
	normalSigner2 := sdk.AccAddress("normal2_____________") // Weight 1 (default)

	// Create a weighted multi-sig wallet
	// Threshold: 4 total weight needed
	// Heavy signer has 3, normal signers have 1 each
	wallet := &wspb.MultiSigWallet{
		WalletId:     "test_wallet_3",
		Signers:      []string{heavySigner.String(), normalSigner1.String(), normalSigner2.String()},
		Threshold:    3, // Number of signers (not used when WeightThreshold is set)
		TotalSigners: 3,
		CreatedAt:    timestamppb.Now(),
		Creator:      creator.String(),
		SignerWeights: map[string]int32{
			heavySigner.String():   3,
			normalSigner1.String(): 1,
			normalSigner2.String(): 1,
		},
		WeightThreshold: 4, // Need 4 total weight
	}

	walletBytes, err := suite.cdc.Marshal(wallet)
	suite.Require().NoError(err)
	err = suite.keeper.SetMultiSigWallet(ctx, "test_wallet_3", walletBytes)
	suite.Require().NoError(err)

	// Verify the wallet was stored correctly with weights
	retrievedBytes, err := suite.keeper.GetMultiSigWallet(ctx, "test_wallet_3")
	suite.Require().NoError(err)

	var retrievedWallet wspb.MultiSigWallet
	err = suite.cdc.Unmarshal(retrievedBytes, &retrievedWallet)
	suite.Require().NoError(err)

	// Verify weights are correctly set
	suite.Equal(int32(3), retrievedWallet.SignerWeights[heavySigner.String()])
	suite.Equal(int32(1), retrievedWallet.SignerWeights[normalSigner1.String()])
	suite.Equal(int32(1), retrievedWallet.SignerWeights[normalSigner2.String()])
	suite.Equal(int32(4), retrievedWallet.WeightThreshold)

	// Note: Full integration test with proper transaction signing context
	// would test the actual signing flow. Unit tests verify the storage
	// and configuration logic.
}

// TestSignMultiSigTransaction_WeightThresholdCalculation tests weight threshold vs regular threshold
func (suite *MsgServerTestSuite) TestSignMultiSigTransaction_WeightThresholdCalculation() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Create test addresses
	creator := sdk.AccAddress("creator_____________")
	signer1 := sdk.AccAddress("signer1_____________")
	signer2 := sdk.AccAddress("signer2_____________")

	// Create wallet with both threshold types
	// When WeightThreshold is set, it should be used instead of Threshold
	wallet := &wspb.MultiSigWallet{
		WalletId:     "test_wallet_4",
		Signers:      []string{signer1.String(), signer2.String()},
		Threshold:    2, // Regular threshold (ignored when WeightThreshold is set)
		TotalSigners: 2,
		CreatedAt:    timestamppb.Now(),
		Creator:      creator.String(),
		SignerWeights: map[string]int32{
			signer1.String(): 2,
			signer2.String(): 3,
		},
		WeightThreshold: 3, // Weight threshold (takes precedence)
	}

	walletBytes, err := suite.cdc.Marshal(wallet)
	suite.Require().NoError(err)
	err = suite.keeper.SetMultiSigWallet(ctx, "test_wallet_4", walletBytes)
	suite.Require().NoError(err)

	// Verify the wallet was stored correctly
	retrievedBytes, err := suite.keeper.GetMultiSigWallet(ctx, "test_wallet_4")
	suite.Require().NoError(err)

	var retrievedWallet wspb.MultiSigWallet
	err = suite.cdc.Unmarshal(retrievedBytes, &retrievedWallet)
	suite.Require().NoError(err)

	suite.Equal(int32(3), retrievedWallet.WeightThreshold)
	suite.Equal(int32(2), retrievedWallet.Threshold)
}

// TestSignMultiSigTransaction_DefaultWeightForSignerWithoutExplicitWeight tests default weight assignment
func (suite *MsgServerTestSuite) TestSignMultiSigTransaction_DefaultWeightForSignerWithoutExplicitWeight() {
	ctx := sdk.UnwrapSDKContext(suite.ctx)

	// Create test addresses
	creator := sdk.AccAddress("creator_____________")
	weightedSigner := sdk.AccAddress("weighted____________")
	defaultSigner := sdk.AccAddress("default_____________")

	// Create wallet where one signer has explicit weight, another doesn't
	wallet := &wspb.MultiSigWallet{
		WalletId:     "test_wallet_5",
		Signers:      []string{weightedSigner.String(), defaultSigner.String()},
		Threshold:    2,
		TotalSigners: 2,
		CreatedAt:    timestamppb.Now(),
		Creator:      creator.String(),
		SignerWeights: map[string]int32{
			weightedSigner.String(): 5,
			// defaultSigner not in map, should get default weight of 1
		},
		WeightThreshold: 6,
	}

	walletBytes, err := suite.cdc.Marshal(wallet)
	suite.Require().NoError(err)
	err = suite.keeper.SetMultiSigWallet(ctx, "test_wallet_5", walletBytes)
	suite.Require().NoError(err)

	// Verify the wallet configuration
	retrievedBytes, err := suite.keeper.GetMultiSigWallet(ctx, "test_wallet_5")
	suite.Require().NoError(err)

	var retrievedWallet wspb.MultiSigWallet
	err = suite.cdc.Unmarshal(retrievedBytes, &retrievedWallet)
	suite.Require().NoError(err)

	// Verify weighted signer has explicit weight
	weight, exists := retrievedWallet.SignerWeights[weightedSigner.String()]
	suite.True(exists)
	suite.Equal(int32(5), weight)

	// Verify default signer is not in the weights map (will get default 1)
	_, exists = retrievedWallet.SignerWeights[defaultSigner.String()]
	suite.False(exists)
}

// TestApproveRecovery_NilRequest tests ApproveRecovery with nil request
func (suite *MsgServerTestSuite) TestApproveRecovery_NilRequest() {
	resp, err := suite.msgServer.ApproveRecovery(suite.ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(resp)

	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.InvalidArgument, st.Code())
	suite.Contains(st.Message(), "empty request")
}

// TestAuthenticateBiometric_Success tests successful biometric authentication
func (suite *MsgServerTestSuite) TestAuthenticateBiometric_Success() {
	// Create test wallet address
	walletAddr := sdk.AccAddress("wallet_addr_________")

	// Enroll biometric first
	enrollmentData := []byte("test_biometric_data_for_enrollment_123456789012345678901234567890")
	enrollMsg := &wspb.MsgEnrollBiometric{
		WalletId:       walletAddr.String(),
		Type:           wspb.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		EnrollmentData: enrollmentData,
	}

	enrollResp, err := suite.msgServer.EnrollBiometric(suite.ctx, enrollMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(enrollResp)
	suite.Equal(walletAddr.String(), enrollResp.Auth.WalletId)
	suite.True(enrollResp.Auth.Enabled)

	// Authenticate with the same biometric data
	authMsg := &wspb.MsgAuthenticateBiometric{
		WalletId:       walletAddr.String(),
		BiometricProof: enrollmentData, // Same data should match
	}

	authResp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, authMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(authResp)
	suite.True(authResp.Authenticated)
	suite.Equal(int32(0), authResp.FailedAttempts)
	suite.False(authResp.LockedOut)
}

// TestAuthenticateBiometric_InvalidProof tests authentication with wrong biometric data
func (suite *MsgServerTestSuite) TestAuthenticateBiometric_InvalidProof() {
	walletAddr := sdk.AccAddress("wallet_addr2________")

	// Enroll biometric
	enrollmentData := []byte("correct_biometric_data_for_enrollment_12345678901234567890123")
	enrollMsg := &wspb.MsgEnrollBiometric{
		WalletId:       walletAddr.String(),
		Type:           wspb.BiometricType_BIOMETRIC_TYPE_FACE,
		EnrollmentData: enrollmentData,
	}

	_, err := suite.msgServer.EnrollBiometric(suite.ctx, enrollMsg)
	suite.Require().NoError(err)

	// Authenticate with DIFFERENT biometric data (should fail)
	wrongData := []byte("wrong_biometric_data_this_should_not_match_12345678901234567890123")
	authMsg := &wspb.MsgAuthenticateBiometric{
		WalletId:       walletAddr.String(),
		BiometricProof: wrongData,
	}

	authResp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, authMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(authResp)
	suite.False(authResp.Authenticated)
	suite.Equal(int32(1), authResp.FailedAttempts)
	suite.False(authResp.LockedOut)
}

// TestAuthenticateBiometric_ReplayAttack tests that the same proof cannot be used twice
func (suite *MsgServerTestSuite) TestAuthenticateBiometric_ReplayAttack() {
	walletAddr := sdk.AccAddress("wallet_addr3________")

	// Enroll biometric
	enrollmentData := []byte("replay_test_biometric_data_for_enrollment_1234567890123456789012")
	enrollMsg := &wspb.MsgEnrollBiometric{
		WalletId:       walletAddr.String(),
		Type:           wspb.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		EnrollmentData: enrollmentData,
	}

	_, err := suite.msgServer.EnrollBiometric(suite.ctx, enrollMsg)
	suite.Require().NoError(err)

	// First authentication - should succeed
	authMsg := &wspb.MsgAuthenticateBiometric{
		WalletId:       walletAddr.String(),
		BiometricProof: enrollmentData,
	}

	authResp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, authMsg)
	suite.Require().NoError(err)
	suite.True(authResp.Authenticated)

	// Second authentication with SAME proof - should fail (replay attack)
	authResp2, err := suite.msgServer.AuthenticateBiometric(suite.ctx, authMsg)
	suite.Require().Error(err)
	suite.Require().Nil(authResp2)

	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.AlreadyExists, st.Code())
	suite.Contains(st.Message(), "replay attack detected")
}

// TestAuthenticateBiometric_ProofTooShort tests that short proofs are rejected
func (suite *MsgServerTestSuite) TestAuthenticateBiometric_ProofTooShort() {
	walletAddr := sdk.AccAddress("wallet_addr4________")

	// Enroll biometric
	enrollmentData := []byte("valid_biometric_data_for_enrollment_this_is_long_enough_12345")
	enrollMsg := &wspb.MsgEnrollBiometric{
		WalletId:       walletAddr.String(),
		Type:           wspb.BiometricType_BIOMETRIC_TYPE_IRIS,
		EnrollmentData: enrollmentData,
	}

	_, err := suite.msgServer.EnrollBiometric(suite.ctx, enrollMsg)
	suite.Require().NoError(err)

	// Authenticate with too short proof (less than 64 bytes)
	shortProof := []byte("short")
	authMsg := &wspb.MsgAuthenticateBiometric{
		WalletId:       walletAddr.String(),
		BiometricProof: shortProof,
	}

	authResp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, authMsg)
	suite.Require().Error(err)
	suite.Require().Nil(authResp)

	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.InvalidArgument, st.Code())
	suite.Contains(st.Message(), "biometric proof too short")
}

// TestAuthenticateBiometric_NotEnrolled tests authentication without enrollment
func (suite *MsgServerTestSuite) TestAuthenticateBiometric_NotEnrolled() {
	walletAddr := sdk.AccAddress("wallet_addr5________")

	// Try to authenticate without enrolling first
	authMsg := &wspb.MsgAuthenticateBiometric{
		WalletId:       walletAddr.String(),
		BiometricProof: []byte("some_biometric_data_that_is_long_enough_12345678901234567890123"),
	}

	authResp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, authMsg)
	suite.Require().Error(err)
	suite.Require().Nil(authResp)

	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.NotFound, st.Code())
	suite.Contains(st.Message(), "biometric not configured")
}

// TestAuthenticateBiometric_Lockout tests that account locks after max failed attempts
func (suite *MsgServerTestSuite) TestAuthenticateBiometric_Lockout() {
	walletAddr := sdk.AccAddress("wallet_addr6________")

	// Enroll biometric
	enrollmentData := []byte("lockout_test_biometric_data_for_enrollment_123456789012345678901")
	enrollMsg := &wspb.MsgEnrollBiometric{
		WalletId:       walletAddr.String(),
		Type:           wspb.BiometricType_BIOMETRIC_TYPE_VOICE,
		EnrollmentData: enrollmentData,
	}

	_, err := suite.msgServer.EnrollBiometric(suite.ctx, enrollMsg)
	suite.Require().NoError(err)

	// Make 5 failed authentication attempts
	wrongData := []byte("wrong_biometric_data_for_lockout_test_1234567890123456789012345678")
	for i := 0; i < 5; i++ {
		authMsg := &wspb.MsgAuthenticateBiometric{
			WalletId:       walletAddr.String(),
			BiometricProof: wrongData,
		}

		authResp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, authMsg)
		suite.Require().NoError(err)
		suite.False(authResp.Authenticated)

		if i < 4 {
			// First 4 attempts should not lock out
			suite.False(authResp.LockedOut)
			suite.Equal(int32(i+1), authResp.FailedAttempts)
		} else {
			// 5th attempt should trigger lockout
			suite.True(authResp.LockedOut)
			suite.Equal(int32(5), authResp.FailedAttempts)
		}
	}

	// Try with correct data - should still be locked out
	correctAuthMsg := &wspb.MsgAuthenticateBiometric{
		WalletId:       walletAddr.String(),
		BiometricProof: enrollmentData,
	}

	authResp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, correctAuthMsg)
	suite.Require().NoError(err)
	suite.False(authResp.Authenticated)
	suite.True(authResp.LockedOut)
}

// TestEnrollBiometric_EmptyData tests enrollment with no data
func (suite *MsgServerTestSuite) TestEnrollBiometric_EmptyData() {
	walletAddr := sdk.AccAddress("wallet_addr7________")

	enrollMsg := &wspb.MsgEnrollBiometric{
		WalletId:       walletAddr.String(),
		Type:           wspb.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		EnrollmentData: []byte{}, // Empty data
	}

	enrollResp, err := suite.msgServer.EnrollBiometric(suite.ctx, enrollMsg)
	suite.Require().Error(err)
	suite.Require().Nil(enrollResp)

	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.InvalidArgument, st.Code())
	suite.Contains(st.Message(), "enrollment data is required")
}

// TestAuthenticateBiometric_NilRequest tests nil request handling
func (suite *MsgServerTestSuite) TestAuthenticateBiometric_NilRequest() {
	resp, err := suite.msgServer.AuthenticateBiometric(suite.ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(resp)

	st, ok := status.FromError(err)
	suite.Require().True(ok)
	suite.Equal(codes.InvalidArgument, st.Code())
	suite.Contains(st.Message(), "empty request")
}
