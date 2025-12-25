// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package privacy

import (
	"crypto/elliptic"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Ring Signatures

func TestRingSigner_SignAndVerify(t *testing.T) {
	signer := NewRingSigner()

	// Generate proper elliptic curve public keys (33 bytes compressed)
	curve := signer.curve
	publicKeys := make([][]byte, 3)

	// Generate keys for ring members
	// The signer will be at index 1
	signerIndex := 1
	privateKey := big.NewInt(1100) // This corresponds to publicKeys[1]

	for i := 0; i < 3; i++ {
		privKey := big.NewInt(int64(1000 + i*100))
		x, y := curve.ScalarBaseMult(privKey.Bytes())
		publicKeys[i] = elliptic.MarshalCompressed(curve, x, y)
	}

	message := []byte("test message to sign")

	// Sign message
	signature, err := signer.Sign(signerIndex, privateKey, publicKeys, message)
	require.NoError(t, err)
	assert.NotNil(t, signature)

	// Verify signature
	valid, err := signer.Verify(signature, message)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestRingSigner_InvalidRingSize(t *testing.T) {
	signer := NewRingSigner()

	privateKey := big.NewInt(12345)
	publicKeys := [][]byte{[]byte("single_key")}
	message := []byte("test message")

	// Try to sign with ring size < 2
	_, err := signer.Sign(0, privateKey, publicKeys, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ring size must be at least 2")
}

func TestMLSAGSigner_SignAndVerify(t *testing.T) {
	signer := NewMLSAGSigner()

	// Generate proper elliptic curve public keys (33 bytes compressed P256)
	curve := signer.curve

	// Generate ring of public keys (each row has 2 keys)
	publicKeyMatrix := make([][][]byte, 3) // 3 ring members
	for i := 0; i < 3; i++ {
		publicKeyMatrix[i] = make([][]byte, 2) // 2 keys per member
		for j := 0; j < 2; j++ {
			privKey := big.NewInt(int64(1000 + i*100 + j*10))
			x, y := curve.ScalarBaseMult(privKey.Bytes())
			publicKeyMatrix[i][j] = elliptic.MarshalCompressed(curve, x, y)
		}
	}

	// The signer is at index 1, so use corresponding private keys
	privateKeys := []*big.Int{
		big.NewInt(1100), // Corresponds to publicKeyMatrix[1][0]
		big.NewInt(1110), // Corresponds to publicKeyMatrix[1][1]
	}

	message := []byte("mlsag test message")

	// Sign
	signature, err := signer.SignMLSAG(1, privateKeys, publicKeyMatrix, message)
	require.NoError(t, err)
	assert.NotNil(t, signature)

	// Verify
	valid, err := signer.VerifyMLSAG(signature, message)
	require.NoError(t, err)
	assert.True(t, valid)
}

// Test Confidential Transactions

func TestConfidentialTransaction_CreateAndVerifyCommitment(t *testing.T) {
	ct, err := NewConfidentialTransactionSystem(32)
	require.NoError(t, err)

	value := big.NewInt(5000)

	// Create commitment
	commitment, err := ct.CreateCommitment(value)
	require.NoError(t, err)
	assert.NotNil(t, commitment)

	// Verify commitment
	valid := ct.VerifyCommitment(
		commitment.Commitment,
		commitment.Value,
		commitment.BlindingFactor,
	)
	assert.True(t, valid)
}

func TestConfidentialTransaction_Bulletproof(t *testing.T) {
	ct, err := NewConfidentialTransactionSystem(32)
	require.NoError(t, err)

	value := big.NewInt(1000)
	blindingFactor := big.NewInt(12345)

	// Generate bulletproof
	proof, err := ct.GenerateBulletproof(value, blindingFactor)
	require.NoError(t, err)
	assert.NotNil(t, proof)

	// Verify bulletproof
	valid, err := ct.VerifyBulletproof(proof)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestConfidentialTransaction_RingCT(t *testing.T) {
	ct, err := NewConfidentialTransactionSystem(32)
	require.NoError(t, err)

	inputAmounts := []*big.Int{big.NewInt(1000), big.NewInt(500)}
	outputAmounts := []*big.Int{big.NewInt(1200)}
	fee := big.NewInt(300)

	// Create RingCT
	ringCT, err := ct.CreateRingCT(inputAmounts, outputAmounts, fee)
	require.NoError(t, err)
	assert.NotNil(t, ringCT)

	// Verify RingCT
	valid, err := ct.VerifyRingCT(ringCT)
	require.NoError(t, err)
	assert.True(t, valid)
}

// Test Network Privacy

func TestNetworkPrivacyManager_TorClient(t *testing.T) {
	config := &NetworkPrivacyConfig{
		NetworkType:     NetworkTypeTor,
		TorProxyAddr:    "127.0.0.1:9050",
		CircuitLifetime: 10 * time.Minute,
		StreamIsolation: true,
	}

	npm, err := NewNetworkPrivacyManager(config)
	require.NoError(t, err)
	assert.NotNil(t, npm)

	// Create circuit
	circuit, err := npm.CreateCircuit()
	require.NoError(t, err)
	assert.NotNil(t, circuit)

	// Destroy circuit
	err = npm.DestroyCircuit(circuit.ID)
	require.NoError(t, err)
}

func TestNetworkPrivacyManager_I2PClient(t *testing.T) {
	config := &NetworkPrivacyConfig{
		NetworkType:     NetworkTypeI2P,
		I2PProxyAddr:    "127.0.0.1:4444",
		CircuitLifetime: 10 * time.Minute,
	}

	npm, err := NewNetworkPrivacyManager(config)
	require.NoError(t, err)
	assert.NotNil(t, npm)

	// Create circuit (I2P tunnel)
	circuit, err := npm.CreateCircuit()
	require.NoError(t, err)
	assert.NotNil(t, circuit)
}

// Test Mixing Service

func TestMixingService_CreateAndJoinPool(t *testing.T) {
	ms := NewMixingService(3)

	denomination := big.NewInt(1000)
	fee := big.NewInt(10)
	now := time.Now()

	// Create pool
	pool, err := ms.CreatePool(denomination, 3, 10, 2, 5, 1*time.Hour, fee, now)
	require.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Equal(t, PoolStatusPending, pool.Status)

	// Join pool
	// Generate proper 33-byte elliptic curve commitment
	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult(big.NewInt(12345).Bytes())
	commitment := elliptic.MarshalCompressed(curve, x, y)
	outputAddr := []byte("output_address")
	blindingFactor := big.NewInt(54321)

	err = ms.JoinPool(pool.ID, "participant1", commitment, outputAddr, blindingFactor, now)
	require.NoError(t, err)

	// Check pool status
	poolStatus, err := ms.GetPoolStatus(pool.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(poolStatus.Participants))
}

func TestMixingService_ExecuteMixing(t *testing.T) {
	ms := NewMixingService(2)

	denomination := big.NewInt(1000)
	fee := big.NewInt(10)
	now := time.Now()

	// Create pool
	pool, err := ms.CreatePool(denomination, 2, 5, 1, 3, 1*time.Hour, fee, now)
	require.NoError(t, err)

	// Add participants
	curve := elliptic.P256()
	for i := 0; i < 3; i++ {
		// Generate proper 33-byte elliptic curve commitment
		x, y := curve.ScalarBaseMult(big.NewInt(int64(10000 + i*1000)).Bytes())
		commitment := elliptic.MarshalCompressed(curve, x, y)
		outputAddr := []byte("output_address_" + string(rune(i)))
		blindingFactor := big.NewInt(int64(i * 1000))

		err = ms.JoinPool(
			pool.ID,
			"participant_"+string(rune(i)),
			commitment,
			outputAddr,
			blindingFactor,
			now,
		)
		require.NoError(t, err)
	}

	// Execute mixing
	result, err := ms.ExecuteMixing(pool.ID, now)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result.Outputs))
}

// Test Encryption

func TestMemoEncryptor_AES256GCM(t *testing.T) {
	encryptor := NewMemoEncryptor(AlgorithmAES256GCM)

	memo := []byte("This is a secret transaction memo")

	// Generate a proper recipient key pair
	curve := elliptic.P256()
	recipientPrivKey := big.NewInt(54321)
	recipientPubX, recipientPubY := curve.ScalarBaseMult(recipientPrivKey.Bytes())
	recipientPubKey := elliptic.MarshalCompressed(curve, recipientPubX, recipientPubY)

	// Encrypt
	encrypted, err := encryptor.Encrypt(memo, recipientPubKey)
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.Equal(t, AlgorithmAES256GCM, encrypted.Algorithm)

	// Use the matching private key for decryption
	privateKey := recipientPrivKey.Bytes()

	// Decrypt
	decrypted, err := encryptor.Decrypt(encrypted, privateKey)
	require.NoError(t, err)
	assert.NotNil(t, decrypted)
	assert.Equal(t, memo, decrypted)
}

func TestMemoEncryptor_ChaCha20Poly1305(t *testing.T) {
	encryptor := NewMemoEncryptor(AlgorithmChaCha20Poly1305)

	memo := []byte("Another secret memo")

	// Generate a proper recipient key pair
	curve := elliptic.P256()
	recipientPrivKey := big.NewInt(98765)
	recipientPubX, recipientPubY := curve.ScalarBaseMult(recipientPrivKey.Bytes())
	recipientPubKey := elliptic.MarshalCompressed(curve, recipientPubX, recipientPubY)

	// Encrypt
	encrypted, err := encryptor.Encrypt(memo, recipientPubKey)
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.Equal(t, AlgorithmChaCha20Poly1305, encrypted.Algorithm)
}

// Test View Keys

func TestViewKeyManager_GenerateAndRetrieve(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_incoming", "view_balance"}
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	// Generate view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeIncoming,
		address,
		permissions,
		&expiresAt,
		now,
	)
	require.NoError(t, err)
	assert.NotNil(t, viewKey)

	// Retrieve view key
	retrieved, err := vkm.GetViewKey(viewKey.PublicKey, now)
	require.NoError(t, err)
	assert.Equal(t, viewKey.Type, retrieved.Type)
	assert.Equal(t, viewKey.Permissions, retrieved.Permissions)
}

func TestViewKeyManager_RevokeKey(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_all"}
	now := time.Now()

	// Generate view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeAudit,
		address,
		permissions,
		nil,
		now,
	)
	require.NoError(t, err)

	// Revoke view key
	err = vkm.RevokeViewKey(viewKey.PublicKey, now)
	require.NoError(t, err)

	// Try to retrieve revoked key
	_, err = vkm.GetViewKey(viewKey.PublicKey, now)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestViewKeyManager_VerifyPermission(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_incoming", "view_outgoing"}
	now := time.Now()

	// Generate view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeFull,
		address,
		permissions,
		nil,
		now,
	)
	require.NoError(t, err)

	// Verify existing permission
	hasPermission, err := vkm.VerifyPermission(viewKey.PublicKey, "view_incoming", now)
	require.NoError(t, err)
	assert.True(t, hasPermission)

	// Verify non-existing permission
	hasPermission, err = vkm.VerifyPermission(viewKey.PublicKey, "admin", now)
	require.NoError(t, err)
	assert.False(t, hasPermission)
}

func TestViewKeyManager_ExpiredKey(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_balance"}
	now := time.Now()
	expiresAt := now.Add(-1 * time.Hour) // Already expired

	// Generate expired view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeIncoming,
		address,
		permissions,
		&expiresAt,
		now,
	)
	require.NoError(t, err)

	// Try to retrieve expired key
	_, err = vkm.GetViewKey(viewKey.PublicKey, now)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestViewKeyManager_ListActiveKeys(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_all"}
	now := time.Now()

	// Generate multiple view keys
	for i := 0; i < 3; i++ {
		_, err := vkm.GenerateViewKey(
			ViewKeyTypeAudit,
			address,
			permissions,
			nil,
			now,
		)
		require.NoError(t, err)
	}

	// List active keys
	activeKeys := vkm.ListActiveViewKeys(address, now)
	assert.Equal(t, 3, len(activeKeys))
}

// Test Tumbler Service

func TestTumblerService_ScheduleTumbling(t *testing.T) {
	ts := NewTumblerService(2)

	inputAddr := "input_address"
	outputAddrs := []string{"output1", "output2", "output3"}
	totalAmount := big.NewInt(3000)
	splits := []*big.Int{big.NewInt(1000), big.NewInt(1000), big.NewInt(1000)}
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second}
	now := time.Now()

	// Schedule tumbling
	schedule, err := ts.ScheduleTumbling(
		inputAddr,
		outputAddrs,
		totalAmount,
		splits,
		delays,
		now,
	)
	require.NoError(t, err)
	assert.NotNil(t, schedule)
	assert.Equal(t, "SCHEDULED", schedule.Status)
}

func TestTumblerService_InvalidSplits(t *testing.T) {
	ts := NewTumblerService(2)

	inputAddr := "input_address"
	outputAddrs := []string{"output1", "output2"}
	totalAmount := big.NewInt(2000)
	splits := []*big.Int{big.NewInt(1000), big.NewInt(500)} // Don't sum to total
	delays := []time.Duration{1 * time.Second, 2 * time.Second}

	// Try to schedule with invalid splits
	_, err := ts.ScheduleTumbling(
		inputAddr,
		outputAddrs,
		totalAmount,
		splits,
		delays,
		time.Now(),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do not sum to total")
}
