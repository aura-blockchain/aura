// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// ============================================================================
// Trezor Signature Validation Tests
// ============================================================================

func TestTrezorSignature_FullFormat97Bytes(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	// Generate a key pair
	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-model-t-001"

	// Register the hardware wallet first
	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.3",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Create transaction and sign it
	txData := []byte("trezor-transaction-data-to-sign")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	// Create 97-byte format: pubkey[33] || sig[64]
	txPayload := append(privKey.PubKey().Bytes(), sig...)
	require.Len(t, txPayload, 97)

	// Validate transaction
	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, txPayload)
	require.NoError(t, err)
}

func TestTrezorSignature_RecoverableFormat65Bytes(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-model-one-002"

	// Register hardware wallet
	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"1.10.5",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Create transaction and sign
	txData := []byte("recoverable-sig-tx-data")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	// Create 65-byte recoverable signature: r[32] || s[32] || v[1]
	// The recovery ID v is typically 0-3, but can also be 27-30 (EIP-155 style)
	recoverableSig := make([]byte, 65)
	copy(recoverableSig[:64], sig)
	recoverableSig[64] = 0 // Recovery ID = 0

	// This will attempt public key recovery - the code path is exercised
	// Recovery may fail or succeed depending on signature, but validates the 65-byte path
	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, recoverableSig)
	// Expect an error - either recovery fails or recovered key doesn't match address
	require.Error(t, err)
}

func TestTrezorSignature_CompactFormat64Bytes_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-compact-test"

	// Register hardware wallet
	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Create 64-byte compact signature (should fail as it requires pubkey context)
	txData := []byte("compact-sig-tx-data")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	compactSig := sig[:64]
	require.Len(t, compactSig, 64)

	// Compact signature should fail
	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, compactSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "compact signature requires pubkey context")
}

func TestTrezorSignature_TooShort_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-short-sig-test"

	// Register hardware wallet
	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Create signature shorter than 64 bytes
	txData := []byte("short-sig-tx-data")
	shortSig := make([]byte, 32) // Too short

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, shortSig)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDeviceSignature)
}

func TestTrezorSignature_WrongAddress_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-wrong-addr-test"

	// Register hardware wallet
	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Sign with a different key (address mismatch)
	otherKey := secp256k1.GenPrivKey()
	txData := []byte("wrong-addr-tx-data")
	txDigest := sha256.Sum256(txData)
	sig, err := otherKey.Sign(txDigest[:])
	require.NoError(t, err)

	// Create 97-byte payload with wrong pubkey
	wrongPayload := append(otherKey.PubKey().Bytes(), sig...)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, wrongPayload)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address mismatch")
}

// ============================================================================
// KeepKey Signature Validation Tests
// ============================================================================

func TestKeepKeySignature_FullFormat97Bytes(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "keepkey-classic-001"

	// Register hardware wallet
	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY,
		deviceID,
		"7.7.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Create and sign transaction
	txData := []byte("keepkey-transaction-data")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	// 97-byte format
	txPayload := append(privKey.PubKey().Bytes(), sig...)
	require.Len(t, txPayload, 97)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, txPayload)
	require.NoError(t, err)
}

func TestKeepKeySignature_RecoverableFormat65Bytes(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "keepkey-recover-001"

	// Register hardware wallet
	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY,
		deviceID,
		"7.7.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Create recoverable signature
	txData := []byte("keepkey-recoverable-tx")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	recoverableSig := make([]byte, 65)
	copy(recoverableSig[:64], sig)
	recoverableSig[64] = 27 // EIP-155 style recovery ID

	// This exercises the 65-byte code path - expect error as recovery typically fails
	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, recoverableSig)
	// Expect an error - either recovery fails or recovered key doesn't match address
	require.Error(t, err)
}

func TestKeepKeySignature_CompactFormat_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "keepkey-compact-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY,
		deviceID,
		"7.7.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	txData := []byte("keepkey-compact-tx")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	compactSig := sig[:64]

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, compactSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "compact signature requires pubkey context")
}

func TestKeepKeySignature_TooShort_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "keepkey-short-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY,
		deviceID,
		"7.7.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	txData := []byte("keepkey-short-tx")
	shortSig := make([]byte, 48) // Too short

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, shortSig)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDeviceSignature)
}

// ============================================================================
// ColdCard Signature Validation Tests
// ============================================================================

func TestColdCardSignature_FullFormat97Bytes(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "coldcard-mk4-001"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Create and sign transaction
	txData := []byte("coldcard-bitcoin-style-tx")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	txPayload := append(privKey.PubKey().Bytes(), sig...)
	require.Len(t, txPayload, 97)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, txPayload)
	require.NoError(t, err)
}

func TestColdCardSignature_RecoverableFormat65Bytes(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "coldcard-mk4-recover"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	txData := []byte("coldcard-recoverable-tx")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	recoverableSig := make([]byte, 65)
	copy(recoverableSig[:64], sig)
	recoverableSig[64] = 1 // Recovery ID = 1

	// This exercises the 65-byte code path - expect error as recovery typically fails
	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, recoverableSig)
	// Expect an error - either recovery fails or recovered key doesn't match address
	require.Error(t, err)
}

func TestColdCardSignature_CompactFormat_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "coldcard-compact-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	txData := []byte("coldcard-compact-tx")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)

	compactSig := sig[:64]

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, compactSig)
	require.Error(t, err)
	require.Contains(t, err.Error(), "compact signature requires pubkey context")
}

func TestColdCardSignature_TooShort_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "coldcard-short-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	txData := []byte("coldcard-short-tx")
	shortSig := make([]byte, 16)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, shortSig)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDeviceSignature)
}

func TestColdCardSignature_InvalidLength_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "coldcard-invalid-len"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	txData := []byte("coldcard-invalid-len-tx")
	// 80 bytes - not 64, 65, or 97
	invalidLengthSig := make([]byte, 80)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, invalidLengthSig)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDeviceSignature)
}

// ============================================================================
// PIN Confirmation Requirements Tests
// ============================================================================

func TestRequiresPinConfirmation_TrezorDefaultsToTrue(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-pin-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPin, err := suite.keeper.RequiresPinConfirmation(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.True(t, requiresPin)
}

func TestRequiresPinConfirmation_LedgerDefaultsToTrue(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "ledger-pin-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPin, err := suite.keeper.RequiresPinConfirmation(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.True(t, requiresPin)
}

func TestRequiresPinConfirmation_KeepKeyDefaultsToTrue(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "keepkey-pin-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY,
		deviceID,
		"7.7.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPin, err := suite.keeper.RequiresPinConfirmation(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.True(t, requiresPin)
}

func TestRequiresPinConfirmation_ColdCardDefaultsToTrue(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "coldcard-pin-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPin, err := suite.keeper.RequiresPinConfirmation(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.True(t, requiresPin)
}

func TestRequiresPinConfirmation_NonExistentWallet_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	_, err := suite.keeper.RequiresPinConfirmation(suite.ctx, "non-existent-wallet-id")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrHardwareWalletNotFound)
}

// ============================================================================
// Passphrase Requirements Tests
// ============================================================================

func TestRequiresPassphrase_TrezorDefaultsToFalse(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-passphrase-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPassphrase, err := suite.keeper.RequiresPassphrase(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.False(t, requiresPassphrase)
}

func TestRequiresPassphrase_LedgerDefaultsToFalse(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "ledger-passphrase-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPassphrase, err := suite.keeper.RequiresPassphrase(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.False(t, requiresPassphrase)
}

func TestRequiresPassphrase_KeepKeyDefaultsToFalse(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "keepkey-passphrase-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY,
		deviceID,
		"7.7.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPassphrase, err := suite.keeper.RequiresPassphrase(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.False(t, requiresPassphrase)
}

func TestRequiresPassphrase_ColdCardDefaultsToFalse(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "coldcard-passphrase-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	requiresPassphrase, err := suite.keeper.RequiresPassphrase(suite.ctx, config.WalletId)
	require.NoError(t, err)
	require.False(t, requiresPassphrase)
}

func TestRequiresPassphrase_NonExistentWallet_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	_, err := suite.keeper.RequiresPassphrase(suite.ctx, "non-existent-wallet-id")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrHardwareWalletNotFound)
}

// ============================================================================
// GetHardwareWalletByAddress Tests
// ============================================================================

func TestGetHardwareWalletByAddress_ReturnsError(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()

	// Currently this function returns an error indicating wallet ID is needed
	config, err := suite.keeper.GetHardwareWalletByAddress(suite.ctx, address)
	require.Error(t, err)
	require.Nil(t, config)
	require.Contains(t, err.Error(), "use GetHardwareWallet with wallet ID")
}

func TestGetHardwareWalletByAddress_EmptyAddress(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	config, err := suite.keeper.GetHardwareWalletByAddress(suite.ctx, "")
	require.Error(t, err)
	require.Nil(t, config)
}

// ============================================================================
// Registration Edge Cases
// ============================================================================

func TestRegisterHardwareWallet_EmptyAddress_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	deviceID := "ledger-empty-addr"
	signature := makeRegistrationSig(t, privKey, "", deviceID)

	_, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		"",
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidInput)
}

func TestRegisterHardwareWallet_EmptyDeviceID_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	signature := makeRegistrationSig(t, privKey, address, "")

	_, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		"",
		"2.1.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidInput)
}

func TestRegisterHardwareWallet_UnspecifiedType_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "unspecified-type-test"
	signature := makeRegistrationSig(t, privKey, address, deviceID)

	_, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_UNSPECIFIED,
		deviceID,
		"1.0.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnsupportedHardwareWallet)
}

func TestRegisterHardwareWallet_InvalidSignature_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "invalid-sig-test"

	// Create an invalid signature (wrong format)
	invalidSig := make([]byte, 50) // Wrong length

	_, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		invalidSig,
	)
	require.Error(t, err)
}

func TestRegisterHardwareWallet_MismatchedAddressInSignature_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	otherKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "mismatched-addr-test"

	// Sign with otherKey but use privKey's address
	digest := sha256.Sum256([]byte("aura-hww:" + address + ":" + deviceID))
	sig, err := otherKey.Sign(digest[:])
	require.NoError(t, err)
	signature := append(otherKey.PubKey().Bytes(), sig...)

	_, err = suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		signature,
	)
	require.Error(t, err)
}

// ============================================================================
// Validate Hardware Wallet Transaction Edge Cases
// ============================================================================

func TestValidateHardwareWalletTransaction_NonExistentWallet_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	txData := []byte("test-tx")
	signature := make([]byte, 97)

	err := suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, "non-existent-wallet-id", txData, signature)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get")
}

func TestValidateHardwareWalletTransaction_UnsupportedType_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "unsupported-type-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Manually modify the stored config to have an unsupported type
	var storedConfig wsproto.HardwareWalletConfig
	configBytes, err := suite.keeper.GetHardwareWallet(suite.ctx, config.WalletId)
	require.NoError(t, err)
	err = suite.cdc.Unmarshal(configBytes, &storedConfig)
	require.NoError(t, err)

	storedConfig.Type = wsproto.HardwareWalletType(99) // Invalid type
	modifiedBytes, err := suite.cdc.Marshal(&storedConfig)
	require.NoError(t, err)
	err = suite.keeper.SetHardwareWallet(suite.ctx, config.WalletId, modifiedBytes)
	require.NoError(t, err)

	txData := []byte("test-tx")
	txDigest := sha256.Sum256(txData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)
	txPayload := append(privKey.PubKey().Bytes(), sig...)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, txPayload)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrUnsupportedHardwareWallet)
}

// ============================================================================
// Update Hardware Wallet Usage Tests
// ============================================================================

func TestUpdateHardwareWalletUsage_MultipleUpdates(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "multi-update-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Update usage multiple times
	for i := 0; i < 5; i++ {
		err = suite.keeper.UpdateHardwareWalletUsage(suite.ctx, config.WalletId)
		require.NoError(t, err)
	}

	// Verify signature count
	configBytes, err := suite.keeper.GetHardwareWallet(suite.ctx, config.WalletId)
	require.NoError(t, err)

	var updatedConfig wsproto.HardwareWalletConfig
	suite.cdc.MustUnmarshal(configBytes, &updatedConfig)
	require.Equal(t, int32(5), updatedConfig.SignatureCount)
}

func TestUpdateHardwareWalletUsage_NonExistentWallet_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	err := suite.keeper.UpdateHardwareWalletUsage(suite.ctx, "non-existent-wallet")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get")
}

// ============================================================================
// Signature Verification with Wrong Transaction Data
// ============================================================================

func TestTrezorSignature_WrongTxData_Fails(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "trezor-wrong-tx-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
		deviceID,
		"2.5.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	// Sign one transaction but validate with different data
	originalTxData := []byte("original-transaction")
	txDigest := sha256.Sum256(originalTxData)
	sig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)
	txPayload := append(privKey.PubKey().Bytes(), sig...)

	// Try to validate with different transaction data
	differentTxData := []byte("different-transaction")
	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, differentTxData, txPayload)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidDeviceSignature)
}

// ============================================================================
// Metadata Tests
// ============================================================================

func TestHardwareWalletMetadata_ContainsDeviceType(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "metadata-test"

	regSignature := makeRegistrationSig(t, privKey, address, deviceID)
	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
		deviceID,
		"5.2.0",
		"m/44'/118'/0'/0/0",
		regSignature,
	)
	require.NoError(t, err)

	require.NotNil(t, config.Metadata)
	require.Contains(t, config.Metadata, "device_type")
	require.Equal(t, wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD.String(), config.Metadata["device_type"])
	require.Contains(t, config.Metadata, "registered_timestamp")
}

// ============================================================================
// All Hardware Wallet Types Transaction Validation Table Test
// ============================================================================

func TestAllHardwareWalletTypes_TransactionValidation(t *testing.T) {
	testCases := []struct {
		name       string
		walletType wsproto.HardwareWalletType
		firmware   string
	}{
		{
			name:       "Ledger",
			walletType: wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
			firmware:   "2.1.0",
		},
		{
			name:       "Trezor",
			walletType: wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR,
			firmware:   "2.5.3",
		},
		{
			name:       "KeepKey",
			walletType: wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY,
			firmware:   "7.7.0",
		},
		{
			name:       "ColdCard",
			walletType: wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD,
			firmware:   "5.2.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := new(KeeperTestSuite)
			suite.SetT(t)
			suite.SetupTest()

			privKey := secp256k1.GenPrivKey()
			address := sdk.AccAddress(privKey.PubKey().Address()).String()
			deviceID := "device-" + tc.name

			// Register
			regSignature := makeRegistrationSig(t, privKey, address, deviceID)
			config, err := suite.keeper.RegisterHardwareWallet(
				suite.ctx,
				address,
				tc.walletType,
				deviceID,
				tc.firmware,
				"m/44'/118'/0'/0/0",
				regSignature,
			)
			require.NoError(t, err)
			require.NotNil(t, config)

			// Validate transaction
			txData := []byte("transaction-for-" + tc.name)
			txDigest := sha256.Sum256(txData)
			sig, err := privKey.Sign(txDigest[:])
			require.NoError(t, err)
			txPayload := append(privKey.PubKey().Bytes(), sig...)

			err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, txPayload)
			require.NoError(t, err)
		})
	}
}
