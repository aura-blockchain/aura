package keeper_test

import (
	"crypto/sha256"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

func TestLedgerRegistrationAndTxValidation(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "ledger-nano-x-98765"

	regDigest := sha256.Sum256([]byte("aura-hww:" + address + ":" + deviceID))
	regSig, err := privKey.Sign(regDigest[:])
	require.NoError(t, err)
	regPayload := append(privKey.PubKey().Bytes(), regSig...)

	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		regPayload,
	)
	require.NoError(t, err)
	require.NotNil(t, config)

	txData := []byte("sample-ledger-tx")
	txDigest := sha256.Sum256(txData)
	txSig, err := privKey.Sign(txDigest[:])
	require.NoError(t, err)
	txPayload := append(privKey.PubKey().Bytes(), txSig...)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, txPayload)
	require.NoError(t, err)
}

func TestLedgerValidationFailsOnWrongAddress(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address()).String()
	deviceID := "ledger-nano-x-wrong"

	regDigest := sha256.Sum256([]byte("aura-hww:" + address + ":" + deviceID))
	regSig, err := privKey.Sign(regDigest[:])
	require.NoError(t, err)
	regPayload := append(privKey.PubKey().Bytes(), regSig...)

	config, err := suite.keeper.RegisterHardwareWallet(
		suite.ctx,
		address,
		wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER,
		deviceID,
		"2.1.0",
		"m/44'/118'/0'/0/0",
		regPayload,
	)
	require.NoError(t, err)
	require.NotNil(t, config)

	txData := []byte("tampered-ledger-tx")
	txDigest := sha256.Sum256(txData)

	// Sign with a different key (address mismatch)
	otherKey := secp256k1.GenPrivKey()
	badSig, err := otherKey.Sign(txDigest[:])
	require.NoError(t, err)
	badPayload := append(otherKey.PubKey().Bytes(), badSig...)

	err = suite.keeper.ValidateHardwareWalletTransaction(suite.ctx, config.WalletId, txData, badPayload)
	require.Error(t, err)
}
