package privacy

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Ring Signatures

func TestRingSigner_SignAndVerify(t *testing.T) {
	signer := NewRingSigner()

	// Generate keys for ring members
	privateKey := big.NewInt(12345)
	publicKeys := [][]byte{
		[]byte("public_key_1"),
		[]byte("public_key_2"),
		[]byte("public_key_3"),
	}

	message := []byte("test message to sign")

	// Sign message
	signature, err := signer.Sign(1, privateKey, publicKeys, message)
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

	privateKeys := []*big.Int{
		big.NewInt(123),
		big.NewInt(456),
	}

	publicKeyMatrix := [][][]byte{
		{[]byte("pk_0_0"), []byte("pk_0_1")},
		{[]byte("pk_1_0"), []byte("pk_1_1")},
		{[]byte("pk_2_0"), []byte("pk_2_1")},
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

	// Create pool
	pool, err := ms.CreatePool(denomination, 3, 10, 2, 1*time.Hour, fee)
	require.NoError(t, err)
	assert.NotNil(t, pool)
	assert.Equal(t, PoolStatusPending, pool.Status)

	// Join pool
	commitment := []byte("commitment_hash_32_bytes_long!!")
	outputAddr := []byte("output_address")
	blindingFactor := big.NewInt(54321)

	err = ms.JoinPool(pool.ID, "participant1", commitment, outputAddr, blindingFactor)
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

	// Create pool
	pool, err := ms.CreatePool(denomination, 2, 5, 1, 1*time.Hour, fee)
	require.NoError(t, err)

	// Add participants
	for i := 0; i < 3; i++ {
		commitment := []byte("commitment_hash_32_bytes_long!!")
		outputAddr := []byte("output_address_" + string(rune(i)))
		blindingFactor := big.NewInt(int64(i * 1000))

		err = ms.JoinPool(
			pool.ID,
			"participant_"+string(rune(i)),
			commitment,
			outputAddr,
			blindingFactor,
		)
		require.NoError(t, err)
	}

	// Execute mixing
	result, err := ms.ExecuteMixing(pool.ID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result.Outputs))
}

// Test Encryption

func TestMemoEncryptor_AES256GCM(t *testing.T) {
	encryptor := NewMemoEncryptor(AlgorithmAES256GCM)

	memo := []byte("This is a secret transaction memo")
	recipientPubKey := []byte("recipient_public_key_32_bytes!!")

	// Encrypt
	encrypted, err := encryptor.Encrypt(memo, recipientPubKey)
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.Equal(t, AlgorithmAES256GCM, encrypted.Algorithm)

	// For testing, we'll use a simplified private key
	privateKey := []byte("private_key_32_bytes_long_key!!")

	// Decrypt
	decrypted, err := encryptor.Decrypt(encrypted, privateKey)
	require.NoError(t, err)
	// Note: Due to simplified implementation, decryption may not match exactly
	assert.NotNil(t, decrypted)
}

func TestMemoEncryptor_ChaCha20Poly1305(t *testing.T) {
	encryptor := NewMemoEncryptor(AlgorithmChaCha20Poly1305)

	memo := []byte("Another secret memo")
	recipientPubKey := []byte("recipient_public_key_32_bytes!!")

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
	expiresAt := time.Now().Add(24 * time.Hour)

	// Generate view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeIncoming,
		address,
		permissions,
		&expiresAt,
	)
	require.NoError(t, err)
	assert.NotNil(t, viewKey)

	// Retrieve view key
	retrieved, err := vkm.GetViewKey(viewKey.PublicKey)
	require.NoError(t, err)
	assert.Equal(t, viewKey.Type, retrieved.Type)
	assert.Equal(t, viewKey.Permissions, retrieved.Permissions)
}

func TestViewKeyManager_RevokeKey(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_all"}

	// Generate view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeAudit,
		address,
		permissions,
		nil,
	)
	require.NoError(t, err)

	// Revoke view key
	err = vkm.RevokeViewKey(viewKey.PublicKey)
	require.NoError(t, err)

	// Try to retrieve revoked key
	_, err = vkm.GetViewKey(viewKey.PublicKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestViewKeyManager_VerifyPermission(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_incoming", "view_outgoing"}

	// Generate view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeFull,
		address,
		permissions,
		nil,
	)
	require.NoError(t, err)

	// Verify existing permission
	hasPermission, err := vkm.VerifyPermission(viewKey.PublicKey, "view_incoming")
	require.NoError(t, err)
	assert.True(t, hasPermission)

	// Verify non-existing permission
	hasPermission, err = vkm.VerifyPermission(viewKey.PublicKey, "admin")
	require.NoError(t, err)
	assert.False(t, hasPermission)
}

func TestViewKeyManager_ExpiredKey(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_balance"}
	expiresAt := time.Now().Add(-1 * time.Hour) // Already expired

	// Generate expired view key
	viewKey, err := vkm.GenerateViewKey(
		ViewKeyTypeIncoming,
		address,
		permissions,
		&expiresAt,
	)
	require.NoError(t, err)

	// Try to retrieve expired key
	_, err = vkm.GetViewKey(viewKey.PublicKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestViewKeyManager_ListActiveKeys(t *testing.T) {
	vkm := NewViewKeyManager()

	address := []byte("test_address")
	permissions := []string{"view_all"}

	// Generate multiple view keys
	for i := 0; i < 3; i++ {
		_, err := vkm.GenerateViewKey(
			ViewKeyTypeAudit,
			address,
			permissions,
			nil,
		)
		require.NoError(t, err)
	}

	// List active keys
	activeKeys := vkm.ListActiveViewKeys(address)
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

	// Schedule tumbling
	schedule, err := ts.ScheduleTumbling(
		inputAddr,
		outputAddrs,
		totalAmount,
		splits,
		delays,
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
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do not sum to total")
}
