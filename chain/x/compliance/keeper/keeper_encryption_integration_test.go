package keeper_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/keeper"
	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

// ============================================================================
// Keeper Encryption Service Integration Tests
// ============================================================================

func TestKeeper_SetEncryptionService(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)
	require.NotNil(t, k)

	// Initially, encryption should be disabled
	_, enabled := k.GetEncryptionService()
	require.False(t, enabled, "encryption should be disabled by default")

	// Generate master key
	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	// Create encryption service
	encService, err := keeper.NewEncryptionService(masterKey)
	require.NoError(t, err)
	require.NotNil(t, encService)

	// Set encryption service on keeper
	k.SetEncryptionService(encService)

	// Verify encryption is now enabled
	retrievedService, enabled := k.GetEncryptionService()
	require.True(t, enabled, "encryption should be enabled after SetEncryptionService")
	require.NotNil(t, retrievedService)
	require.Equal(t, encService, retrievedService)
}

func TestKeeper_GetEncryptionService_Disabled(t *testing.T) {
	// Setup keeper without encryption
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Get encryption service (should be disabled)
	service, enabled := k.GetEncryptionService()
	require.False(t, enabled, "encryption should be disabled by default")
	require.Nil(t, service)
}

func TestKeeper_GetEncryptionService_Enabled(t *testing.T) {
	// Setup keeper with encryption
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Create and set encryption service
	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	encService, err := keeper.NewEncryptionService(masterKey)
	require.NoError(t, err)

	k.SetEncryptionService(encService)

	// Get encryption service (should be enabled)
	retrievedService, enabled := k.GetEncryptionService()
	require.True(t, enabled, "encryption should be enabled")
	require.NotNil(t, retrievedService)
	require.Equal(t, encService, retrievedService)
}

func TestKeeper_EncryptionService_RoundTrip(t *testing.T) {
	// Setup keeper with encryption
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	encService, err := keeper.NewEncryptionService(masterKey)
	require.NoError(t, err)

	k.SetEncryptionService(encService)

	// Test encryption/decryption through keeper
	service, enabled := k.GetEncryptionService()
	require.True(t, enabled)

	plaintext := []byte("sensitive KYC data")
	context := "kyc:test:1"

	// Encrypt
	ciphertext, err := service.Encrypt(plaintext, context)
	require.NoError(t, err)
	require.NotNil(t, ciphertext)
	require.NotEqual(t, plaintext, ciphertext)

	// Decrypt
	decrypted, err := service.Decrypt(ciphertext, context)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestKeeper_EncryptionService_MultipleContexts(t *testing.T) {
	// Setup keeper with encryption
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	encService, err := keeper.NewEncryptionService(masterKey)
	require.NoError(t, err)

	k.SetEncryptionService(encService)

	service, _ := k.GetEncryptionService()

	// Test different contexts produce different ciphertexts
	plaintext := []byte("test data")

	ciphertext1, err := service.Encrypt(plaintext, "context:kyc")
	require.NoError(t, err)

	ciphertext2, err := service.Encrypt(plaintext, "context:aml")
	require.NoError(t, err)

	ciphertext3, err := service.Encrypt(plaintext, "context:tax")
	require.NoError(t, err)

	// All ciphertexts should be different
	require.NotEqual(t, ciphertext1, ciphertext2)
	require.NotEqual(t, ciphertext2, ciphertext3)
	require.NotEqual(t, ciphertext1, ciphertext3)

	// Each should decrypt correctly with its own context
	decrypted1, err := service.Decrypt(ciphertext1, "context:kyc")
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted1)

	decrypted2, err := service.Decrypt(ciphertext2, "context:aml")
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted2)

	decrypted3, err := service.Decrypt(ciphertext3, "context:tax")
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted3)

	// Cross-context decryption should fail
	_, err = service.Decrypt(ciphertext1, "context:aml")
	require.Error(t, err)
}

// ============================================================================
// Data Protection Service Tests
// ============================================================================

func TestKeeper_GetDataProtectionService(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Get data protection service (should always be available)
	service := k.GetDataProtectionService()
	require.NotNil(t, service, "data protection service should always be available")
}

func TestKeeper_DataProtectionService_GenerateCommitment(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	service := k.GetDataProtectionService()

	// Generate commitment for sensitive data
	sensitiveData := map[string]string{
		"ssn":      "123-45-6789",
		"passport": "AB123456",
		"dob":      "1990-01-01",
	}

	commitment, err := service.GenerateCommitment(sensitiveData)
	require.NoError(t, err)
	require.NotNil(t, commitment)
	require.Len(t, commitment, 32, "SHA-256 commitment should be 32 bytes")
}

func TestKeeper_DataProtectionService_VerifyCommitment(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	service := k.GetDataProtectionService()

	// Generate commitment
	data := "sensitive user data"
	commitment, err := service.GenerateCommitment(data)
	require.NoError(t, err)

	// Verify with correct data
	valid, err := service.VerifyCommitment(data, commitment)
	require.NoError(t, err)
	require.True(t, valid, "valid data should verify")

	// Verify with incorrect data
	invalidValid, err := service.VerifyCommitment("tampered data", commitment)
	require.NoError(t, err)
	require.False(t, invalidValid, "tampered data should not verify")
}

func TestKeeper_DataProtectionService_PIICommitment(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	service := k.GetDataProtectionService()

	// Create PII data
	pii := &keeper.PIIData{
		FullName:       "Alice Smith",
		DateOfBirth:    "1990-01-01",
		SSN:            "123-45-6789",
		PassportNumber: "AB123456",
		Addresses:      []string{"123 Main St"},
		PhoneNumbers:   []string{"+1-555-0123"},
		EmailAddresses: []string{"alice@example.com"},
	}

	// Generate commitment
	commitment, err := service.GeneratePIICommitment(pii)
	require.NoError(t, err)
	require.NotNil(t, commitment)
	require.Len(t, commitment, 32)

	// Verify commitment
	valid, err := service.VerifyPIICommitment(pii, commitment)
	require.NoError(t, err)
	require.True(t, valid)
}

func TestKeeper_DataProtectionService_FieldCommitments(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	service := k.GetDataProtectionService()

	// Generate field commitments
	fields := map[string]interface{}{
		"ssn":      "123-45-6789",
		"passport": "AB123456",
		"address":  "123 Main St",
	}

	commitments, err := service.GenerateFieldCommitments(fields)
	require.NoError(t, err)
	require.Len(t, commitments, 3)

	// Verify field commitments
	valid, err := service.VerifyFieldCommitments(fields, commitments)
	require.NoError(t, err)
	require.True(t, valid)
}

// ============================================================================
// StoreKey Tests
// ============================================================================

func TestKeeper_StoreKey(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Get store key
	storeKey := k.StoreKey()
	require.NotNil(t, storeKey)
	require.Equal(t, input.StoreKey, storeKey)
}

func TestKeeper_StoreKey_ConsistentAcrossCalls(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Get store key multiple times
	storeKey1 := k.StoreKey()
	storeKey2 := k.StoreKey()
	storeKey3 := k.StoreKey()

	// All should be the same reference
	require.Equal(t, storeKey1, storeKey2)
	require.Equal(t, storeKey2, storeKey3)
}

// ============================================================================
// Integration Test: Encryption + Data Protection Together
// ============================================================================

func TestKeeper_EncryptionAndCommitment_KYCWorkflow(t *testing.T) {
	// Setup keeper with both encryption and data protection
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	// Setup encryption
	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	encService, err := keeper.NewEncryptionService(masterKey)
	require.NoError(t, err)
	k.SetEncryptionService(encService)

	// Get both services
	encryptionService, encEnabled := k.GetEncryptionService()
	require.True(t, encEnabled)

	dataProtectionService := k.GetDataProtectionService()
	require.NotNil(t, dataProtectionService)

	// Simulate KYC workflow
	pii := &keeper.PIIData{
		FullName:       "Bob Johnson",
		DateOfBirth:    "1985-06-15",
		SSN:            "987-65-4321",
		PassportNumber: "XY987654",
		Addresses:      []string{"456 Oak Ave"},
	}

	// Step 1: Generate commitment (for on-chain storage)
	commitment, err := dataProtectionService.GeneratePIICommitment(pii)
	require.NoError(t, err)

	// Step 2: Encrypt sensitive fields (for off-chain storage)
	encryptedSSN, err := encryptionService.EncryptString(pii.SSN, "kyc:ssn")
	require.NoError(t, err)

	encryptedPassport, err := encryptionService.EncryptString(pii.PassportNumber, "kyc:passport")
	require.NoError(t, err)

	// Step 3: Store commitment on-chain (simulated)
	// In real code: keeper.SetKYCRecord(ctx, &types.KYCRecord{PiiCommitment: commitment})
	onChainCommitment := commitment

	// Step 4: Store encrypted data off-chain (simulated)
	// In real code: database.StoreEncrypted(...)
	offChainEncryptedSSN := encryptedSSN
	offChainEncryptedPassport := encryptedPassport

	// Step 5: Later retrieval - decrypt off-chain data
	decryptedSSN, err := encryptionService.DecryptString(offChainEncryptedSSN, "kyc:ssn")
	require.NoError(t, err)
	require.Equal(t, pii.SSN, decryptedSSN)

	decryptedPassport, err := encryptionService.DecryptString(offChainEncryptedPassport, "kyc:passport")
	require.NoError(t, err)
	require.Equal(t, pii.PassportNumber, decryptedPassport)

	// Step 6: Verify commitment matches
	valid, err := dataProtectionService.VerifyPIICommitment(pii, onChainCommitment)
	require.NoError(t, err)
	require.True(t, valid, "PII should verify against commitment")
}

func TestKeeper_EncryptionService_KYCMultipleUsers(t *testing.T) {
	// Setup keeper with encryption
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	encService, err := keeper.NewEncryptionService(masterKey)
	require.NoError(t, err)
	k.SetEncryptionService(encService)

	service, _ := k.GetEncryptionService()

	// Simulate encrypting data for multiple users
	users := []struct {
		address string
		ssn     string
		context string
	}{
		{"aura1user1", "111-11-1111", "kyc:aura1user1:ssn"},
		{"aura1user2", "222-22-2222", "kyc:aura1user2:ssn"},
		{"aura1user3", "333-33-3333", "kyc:aura1user3:ssn"},
	}

	encrypted := make(map[string][]byte)

	// Encrypt each user's SSN
	for _, user := range users {
		ciphertext, err := service.Encrypt([]byte(user.ssn), user.context)
		require.NoError(t, err)
		encrypted[user.address] = ciphertext
	}

	// Verify all encrypted values are different
	require.NotEqual(t, encrypted["aura1user1"], encrypted["aura1user2"])
	require.NotEqual(t, encrypted["aura1user2"], encrypted["aura1user3"])

	// Decrypt and verify each user's data
	for _, user := range users {
		decrypted, err := service.Decrypt(encrypted[user.address], user.context)
		require.NoError(t, err)
		require.Equal(t, user.ssn, string(decrypted))
	}

	// Cross-user decryption should fail
	_, err = service.Decrypt(encrypted["aura1user1"], "kyc:aura1user2:ssn")
	require.Error(t, err, "cross-user decryption should fail")
}

// ============================================================================
// Security Tests
// ============================================================================

func TestKeeper_EncryptionService_KeyIsolation(t *testing.T) {
	// Setup two keepers with different keys
	input1 := keepertest.CreateTestInputWithKeys(t, "compliance1")
	k1 := keeper.NewKeeper(input1.Cdc, input1.StoreKey)

	input2 := keepertest.CreateTestInputWithKeys(t, "compliance2")
	k2 := keeper.NewKeeper(input2.Cdc, input2.StoreKey)

	// Create different encryption keys
	key1 := make([]byte, 32)
	_, err := rand.Read(key1)
	require.NoError(t, err)

	key2 := make([]byte, 32)
	_, err = rand.Read(key2)
	require.NoError(t, err)

	encService1, err := keeper.NewEncryptionService(key1)
	require.NoError(t, err)
	k1.SetEncryptionService(encService1)

	encService2, err := keeper.NewEncryptionService(key2)
	require.NoError(t, err)
	k2.SetEncryptionService(encService2)

	// Encrypt with keeper 1
	service1, _ := k1.GetEncryptionService()
	plaintext := []byte("sensitive data")
	ciphertext1, err := service1.Encrypt(plaintext, "test:context")
	require.NoError(t, err)

	// Try to decrypt with keeper 2 (should fail)
	service2, _ := k2.GetEncryptionService()
	_, err = service2.Decrypt(ciphertext1, "test:context")
	require.Error(t, err, "different keys should not decrypt each other's data")
}

func TestKeeper_DataProtectionService_TamperingDetection(t *testing.T) {
	// Setup keeper
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	service := k.GetDataProtectionService()

	// Original data
	originalData := map[string]interface{}{
		"name":    "Alice",
		"balance": 1000000,
		"tier":    "premium",
	}

	// Generate commitment
	commitment, err := service.GenerateCommitment(originalData)
	require.NoError(t, err)

	// Attacker tampers with data
	tamperedData := map[string]interface{}{
		"name":    "Alice",
		"balance": 9999999999, // Tampered!
		"tier":    "premium",
	}

	// Verification should fail
	valid, err := service.VerifyCommitment(tamperedData, commitment)
	require.NoError(t, err)
	require.False(t, valid, "tampered data should not verify")
}

func TestKeeper_EncryptionService_LargePIIData(t *testing.T) {
	// Setup keeper with encryption
	input := keepertest.CreateTestInputWithKeys(t, "compliance")
	k := keeper.NewKeeper(input.Cdc, input.StoreKey)

	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	require.NoError(t, err)

	encService, err := keeper.NewEncryptionService(masterKey)
	require.NoError(t, err)
	k.SetEncryptionService(encService)

	service, _ := k.GetEncryptionService()

	// Large PII structure (using actual PIIData fields)
	largePII := &keeper.PIIData{
		FullName:       "Alice Marie Elizabeth Smith-Johnson",
		DateOfBirth:    "1990-01-01",
		SSN:            "123-45-6789",
		PassportNumber: "AB123456",
		IDNumber:       "DL987654321",
		Nationality:    "United States",
		Residence:      "United States",
		TaxID:          "987-65-4321",
		Addresses: []string{
			"123 Main Street, Apartment 4B, Springfield, IL 62701, USA",
			"456 Oak Avenue, Unit 12C, Chicago, IL 60601, USA",
			"789 Elm Boulevard, Suite 300, New York, NY 10001, USA",
		},
		PhoneNumbers: []string{
			"+1-555-0123",
			"+1-555-0456",
			"+1-555-0789",
		},
		EmailAddresses: []string{
			"alice.smith@example.com",
			"alice.johnson@company.com",
			"a.smith.personal@email.com",
		},
		SourceOfFunds: []string{
			"Employment - Software Engineering at Tech Corp",
			"Investment Income - Stock Portfolio",
			"Real Estate - Rental Properties",
		},
		Occupation:   "Senior Software Engineer",
		Employer:     "Tech Corporation International LLC",
		AnnualIncome: "$150,000 - $200,000",
		NetWorth:     "$1,000,000 - $5,000,000",
		BankAccounts: []string{
			"1234567890",
			"0987654321",
		},
	}

	// Encrypt as JSON
	encrypted, err := service.EncryptJSON(largePII, "kyc:large:pii")
	require.NoError(t, err)
	require.NotNil(t, encrypted)

	// Decrypt
	var decrypted keeper.PIIData
	err = service.DecryptJSON(encrypted, "kyc:large:pii", &decrypted)
	require.NoError(t, err)
	require.Equal(t, largePII.FullName, decrypted.FullName)
	require.Equal(t, largePII.SSN, decrypted.SSN)
	require.Equal(t, largePII.Addresses, decrypted.Addresses)
	require.Equal(t, largePII.PhoneNumbers, decrypted.PhoneNumbers)
}
