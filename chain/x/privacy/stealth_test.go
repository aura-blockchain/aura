// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStealthAddressScheme_GenerateKeys(t *testing.T) {
	scheme := NewStealthAddressScheme()

	keys, err := scheme.GenerateStealthKeys()
	require.NoError(t, err)
	assert.NotNil(t, keys)
	assert.NotNil(t, keys.SpendKeyPair)
	assert.NotNil(t, keys.ViewKeyPair)
	assert.NotEmpty(t, keys.SpendKeyPair.PrivateKey)
	assert.NotEmpty(t, keys.SpendKeyPair.PublicKey)
	assert.NotEmpty(t, keys.ViewKeyPair.PrivateKey)
	assert.NotEmpty(t, keys.ViewKeyPair.PublicKey)
}

func TestStealthAddressScheme_GenerateAndScan(t *testing.T) {
	scheme := NewStealthAddressScheme()

	// Generate recipient keys
	recipientKeys, err := scheme.GenerateStealthKeys()
	require.NoError(t, err)

	// Generate stealth address
	stealthAddr, err := scheme.GenerateStealthAddress(
		recipientKeys.SpendKeyPair.PublicKey,
		recipientKeys.ViewKeyPair.PublicKey,
	)
	require.NoError(t, err)
	assert.NotNil(t, stealthAddr)
	assert.NotEmpty(t, stealthAddr.OneTimePublicKey)
	assert.NotEmpty(t, stealthAddr.TxPublicKey)

	// Scan for payment
	isForMe, err := scheme.ScanForStealthPayments(
		stealthAddr.TxPublicKey,
		stealthAddr.OneTimePublicKey,
		recipientKeys.ViewKeyPair.PrivateKey,
		recipientKeys.SpendKeyPair.PublicKey,
	)
	require.NoError(t, err)
	assert.True(t, isForMe)
}

func TestStealthAddressScheme_DerivePrivateKey(t *testing.T) {
	scheme := NewStealthAddressScheme()

	// Generate keys
	keys, err := scheme.GenerateStealthKeys()
	require.NoError(t, err)

	// Generate stealth address
	stealthAddr, err := scheme.GenerateStealthAddress(
		keys.SpendKeyPair.PublicKey,
		keys.ViewKeyPair.PublicKey,
	)
	require.NoError(t, err)

	// Derive private key for spending
	privateKey, err := scheme.DerivePrivateKey(
		stealthAddr.TxPublicKey,
		keys.ViewKeyPair.PrivateKey,
		keys.SpendKeyPair.PrivateKey,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, privateKey)
}

func TestCurve25519StealthAddress_GenerateAndScan(t *testing.T) {
	scheme := NewCurve25519StealthAddress()

	// Generate keys
	keys, err := scheme.GenerateKeys()
	require.NoError(t, err)
	assert.NotNil(t, keys)

	// Convert to fixed-size arrays
	var spendPub, viewPub [32]byte
	copy(spendPub[:], keys.SpendKeyPair.PublicKey)
	copy(viewPub[:], keys.ViewKeyPair.PublicKey)

	// Create stealth address
	stealthAddr, err := scheme.CreateOneTimeAddress(spendPub, viewPub)
	require.NoError(t, err)
	assert.NotNil(t, stealthAddr)

	// Convert keys for scanning
	var txPub, oneTimeAddr, viewPriv [32]byte
	copy(txPub[:], stealthAddr.TxPublicKey)
	copy(oneTimeAddr[:], stealthAddr.OneTimePublicKey)
	copy(viewPriv[:], keys.ViewKeyPair.PrivateKey)

	// Scan transaction
	isForMe := scheme.ScanTransaction(txPub, oneTimeAddr, viewPriv, spendPub)
	assert.True(t, isForMe)
}

func TestEncryptDecryptAmount(t *testing.T) {
	amount := big.NewInt(12345)
	sharedSecret := []byte("shared_secret_32_bytes_long_key!")

	// Encrypt amount
	encrypted, err := EncryptAmount(amount, sharedSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// Decrypt amount
	decrypted, err := DecryptAmount(encrypted, sharedSecret)
	require.NoError(t, err)
	assert.Equal(t, amount.String(), decrypted.String())
}

func TestDualKeyStealthAddress(t *testing.T) {
	// Create key pairs
	viewKey := &KeyPair{
		PrivateKey: []byte("view_private_key"),
		PublicKey:  []byte("view_public_key"),
	}
	spendKey := &KeyPair{
		PrivateKey: []byte("spend_private_key"),
		PublicKey:  []byte("spend_public_key"),
	}

	// Create dual-key address
	dualKey := NewDualKeyStealthAddress(viewKey, spendKey)
	assert.NotNil(t, dualKey)

	// Get address
	address := dualKey.GetAddress()
	assert.NotEmpty(t, address)

	// Test view permission
	canView := dualKey.CanView(viewKey.PrivateKey)
	assert.True(t, canView)

	// Test spend permission
	canSpend := dualKey.CanSpend(spendKey.PrivateKey)
	assert.True(t, canSpend)

	// Test wrong keys
	canView = dualKey.CanView([]byte("wrong_key"))
	assert.False(t, canView)

	canSpend = dualKey.CanSpend([]byte("wrong_key"))
	assert.False(t, canSpend)
}
