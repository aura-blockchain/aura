package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

type EncryptedKVStoreTestSuite struct {
	suite.Suite
	keeper *keeper.Keeper
	ctx    sdk.Context
}

func TestEncryptedKVStoreTestSuite(t *testing.T) {
	suite.Run(t, new(EncryptedKVStoreTestSuite))
}

func (suite *EncryptedKVStoreTestSuite) SetupTest() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	suite.keeper = ts.Keeper
	suite.ctx = ts.Ctx

	// Initialize encryption service for tests
	encService := keeper.NewEncryptionService(make([]byte, 32))
	suite.keeper.SetEncryptionService(encService)
}

// ============================================================================
// KYC Record Encrypted Tests
// ============================================================================

func (suite *EncryptedKVStoreTestSuite) TestSetKYCRecordEncrypted_Success() {
	addr := "cosmos1test1"
	record := &types.KYCRecord{
		Address:      addr,
		Status:       types.KYCStatus_APPROVED,
		KycLevel:     1,
		Jurisdiction: "US",
		Provider:     "test-provider",
		SubmittedAt:  time.Now(),
		ApprovedAt:   time.Now(),
	}

	err := suite.keeper.SetKYCRecordEncrypted(suite.ctx, record)
	suite.Require().NoError(err)

	// Verify record was stored
	storedRecord, err := suite.keeper.GetKYCRecord(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().NotNil(storedRecord)
	suite.Require().NotEmpty(storedRecord.PiiCommitment)
}

func (suite *EncryptedKVStoreTestSuite) TestSetKYCRecordEncrypted_NoEncryptionService() {
	// Create keeper without encryption service
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	addr := "cosmos1test2"
	record := &types.KYCRecord{
		Address:      addr,
		Status:       types.KYCStatus_APPROVED,
		KycLevel:     1,
		Jurisdiction: "US",
		Provider:     "test-provider",
		SubmittedAt:  time.Now(),
	}

	err := keeperNoEnc.SetKYCRecordEncrypted(ts.Ctx, record)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "encryption service not configured")
}

func (suite *EncryptedKVStoreTestSuite) TestGetKYCRecordEncrypted_Success() {
	addr := "cosmos1test3"
	record := &types.KYCRecord{
		Address:      addr,
		Status:       types.KYCStatus_APPROVED,
		KycLevel:     1,
		Jurisdiction: "US",
		Provider:     "test-provider",
		SubmittedAt:  time.Now(),
	}

	err := suite.keeper.SetKYCRecordEncrypted(suite.ctx, record)
	suite.Require().NoError(err)

	retrieved, err := suite.keeper.GetKYCRecordEncrypted(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrieved)
	suite.Require().Equal(addr, retrieved.Address)
}

func (suite *EncryptedKVStoreTestSuite) TestGetKYCRecordEncrypted_NoEncryption() {
	// Store unencrypted record first
	addr := "cosmos1test4"
	record := &types.KYCRecord{
		Address:      addr,
		Status:       types.KYCStatus_APPROVED,
		KycLevel:     1,
		Jurisdiction: "US",
		Provider:     "test-provider",
		SubmittedAt:  time.Now(),
	}

	err := suite.keeper.SetKYCRecord(suite.ctx, record)
	suite.Require().NoError(err)

	// Remove encryption service
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	// Copy store data
	err = keeperNoEnc.SetKYCRecord(ts.Ctx, record)
	suite.Require().NoError(err)

	retrieved, err := keeperNoEnc.GetKYCRecordEncrypted(ts.Ctx, addr)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrieved)
	suite.Require().Equal(addr, retrieved.Address)
}

func (suite *EncryptedKVStoreTestSuite) TestGetKYCRecordEncrypted_NotFound() {
	_, err := suite.keeper.GetKYCRecordEncrypted(suite.ctx, "cosmos1nonexistent")
	suite.Require().Error(err)
}

// ============================================================================
// AML Profile Encrypted Tests
// ============================================================================

func (suite *EncryptedKVStoreTestSuite) TestSetAMLProfileEncrypted_Success() {
	addr := "cosmos1aml1"
	profile := &types.AMLProfile{
		Address:           addr,
		RiskLevel:         types.RiskLevel_MEDIUM,
		TotalTransactions: 100,
		TotalVolume:       sdkmath.NewInt(1000000),
		LastAssessment:    time.Now(),
		PepStatus:         false,
		RiskFactors:       []string{"factor1", "factor2"},
		SourceOfFunds:     []string{"salary", "investment"},
		Occupation:        "engineer",
	}

	err := suite.keeper.SetAMLProfileEncrypted(suite.ctx, profile)
	suite.Require().NoError(err)

	// Verify profile was stored
	storedProfile, err := suite.keeper.GetAMLProfile(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().NotNil(storedProfile)
	suite.Require().Contains(storedProfile.Occupation, "encrypted:")
}

func (suite *EncryptedKVStoreTestSuite) TestSetAMLProfileEncrypted_NoEncryptionService() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	addr := "cosmos1aml2"
	profile := &types.AMLProfile{
		Address:           addr,
		RiskLevel:         types.RiskLevel_MEDIUM,
		TotalTransactions: 100,
		TotalVolume:       sdkmath.NewInt(1000000),
		LastAssessment:    time.Now(),
		RiskFactors:       []string{"factor1"},
	}

	// Without encryption service, should store plaintext
	err := keeperNoEnc.SetAMLProfileEncrypted(ts.Ctx, profile)
	suite.Require().NoError(err)

	storedProfile, err := keeperNoEnc.GetAMLProfile(ts.Ctx, addr)
	suite.Require().NoError(err)
	suite.Require().NotContains(storedProfile.Occupation, "encrypted:")
}

func (suite *EncryptedKVStoreTestSuite) TestGetAMLProfileEncrypted_Success() {
	addr := "cosmos1aml3"
	profile := &types.AMLProfile{
		Address:           addr,
		RiskLevel:         types.RiskLevel_HIGH,
		TotalTransactions: 200,
		TotalVolume:       sdkmath.NewInt(5000000),
		LastAssessment:    time.Now(),
		RiskFactors:       []string{"high-value", "frequent"},
		SourceOfFunds:     []string{"business"},
		Occupation:        "trader",
	}

	err := suite.keeper.SetAMLProfileEncrypted(suite.ctx, profile)
	suite.Require().NoError(err)

	retrieved, err := suite.keeper.GetAMLProfileEncrypted(suite.ctx, addr)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrieved)
	suite.Require().Equal(addr, retrieved.Address)
	suite.Require().Equal("trader", retrieved.Occupation)
	suite.Require().Equal([]string{"high-value", "frequent"}, retrieved.RiskFactors)
}

func (suite *EncryptedKVStoreTestSuite) TestGetAMLProfileEncrypted_NoEncryption() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	addr := "cosmos1aml4"
	profile := &types.AMLProfile{
		Address:           addr,
		RiskLevel:         types.RiskLevel_LOW,
		TotalTransactions: 10,
		TotalVolume:       sdkmath.NewInt(100000),
		LastAssessment:    time.Now(),
		Occupation:        "employee",
	}

	err := keeperNoEnc.SetAMLProfile(ts.Ctx, profile)
	suite.Require().NoError(err)

	retrieved, err := keeperNoEnc.GetAMLProfileEncrypted(ts.Ctx, addr)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrieved)
	suite.Require().Equal("employee", retrieved.Occupation)
}

func (suite *EncryptedKVStoreTestSuite) TestGetAMLProfileEncrypted_NotFound() {
	_, err := suite.keeper.GetAMLProfileEncrypted(suite.ctx, "cosmos1nonexistent")
	suite.Require().Error(err)
}

// ============================================================================
// Suspicious Activity Encrypted Tests
// ============================================================================

func (suite *EncryptedKVStoreTestSuite) TestSetSuspiciousActivityEncrypted_Success() {
	activity := &types.SuspiciousActivity{
		Id:              "activity1",
		Address:         "cosmos1sus1",
		TransactionHash: "hash123",
		ActivityType:    "structuring",
		Amount:          sdkmath.NewInt(9000),
		DetectedAt:      time.Now(),
		Description:     "Multiple transactions just below threshold",
		Indicators:      []string{"velocity", "amount"},
		FiledSar:        false,
	}

	err := suite.keeper.SetSuspiciousActivityEncrypted(suite.ctx, activity)
	suite.Require().NoError(err)

	// Verify activity was stored
	stored, err := suite.keeper.GetSuspiciousActivity(suite.ctx, "activity1")
	suite.Require().NoError(err)
	suite.Require().NotNil(stored)
	suite.Require().Contains(stored.Description, "encrypted:")
}

func (suite *EncryptedKVStoreTestSuite) TestSetSuspiciousActivityEncrypted_NoEncryptionService() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	activity := &types.SuspiciousActivity{
		Id:          "activity2",
		Address:     "cosmos1sus2",
		ActivityType: "unusual-pattern",
		Amount:      sdkmath.NewInt(50000),
		DetectedAt:  time.Now(),
		Description: "Unusual transaction pattern",
	}

	err := keeperNoEnc.SetSuspiciousActivityEncrypted(ts.Ctx, activity)
	suite.Require().NoError(err)

	stored, err := keeperNoEnc.GetSuspiciousActivity(ts.Ctx, "activity2")
	suite.Require().NoError(err)
	suite.Require().NotContains(stored.Description, "encrypted:")
}

func (suite *EncryptedKVStoreTestSuite) TestGetSuspiciousActivityEncrypted_Success() {
	activity := &types.SuspiciousActivity{
		Id:          "activity3",
		Address:     "cosmos1sus3",
		ActivityType: "high-risk",
		Amount:      sdkmath.NewInt(100000),
		DetectedAt:  time.Now(),
		Description: "High-risk transaction",
		Indicators:  []string{"sanctioned-country", "high-amount"},
	}

	err := suite.keeper.SetSuspiciousActivityEncrypted(suite.ctx, activity)
	suite.Require().NoError(err)

	retrieved, err := suite.keeper.GetSuspiciousActivityEncrypted(suite.ctx, "activity3")
	suite.Require().NoError(err)
	suite.Require().NotNil(retrieved)
	suite.Require().Equal("High-risk transaction", retrieved.Description)
	suite.Require().Equal([]string{"sanctioned-country", "high-amount"}, retrieved.Indicators)
}

func (suite *EncryptedKVStoreTestSuite) TestGetSuspiciousActivityEncrypted_NoEncryption() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	activity := &types.SuspiciousActivity{
		Id:          "activity4",
		Address:     "cosmos1sus4",
		ActivityType: "normal",
		Amount:      sdkmath.NewInt(1000),
		DetectedAt:  time.Now(),
		Description: "Normal transaction",
	}

	err := keeperNoEnc.SetSuspiciousActivity(ts.Ctx, activity)
	suite.Require().NoError(err)

	retrieved, err := keeperNoEnc.GetSuspiciousActivityEncrypted(ts.Ctx, "activity4")
	suite.Require().NoError(err)
	suite.Require().Equal("Normal transaction", retrieved.Description)
}

// ============================================================================
// GDPR Consent Encrypted Tests
// ============================================================================

func (suite *EncryptedKVStoreTestSuite) TestSetGDPRConsentEncrypted_Success() {
	consent := &types.GDPRConsent{
		Address:        "cosmos1gdpr1",
		ConsentType:    "data_processing",
		Consented:      true,
		ConsentGivenAt: time.Now(),
		ConsentVersion: "v1.0",
	}

	err := suite.keeper.SetGDPRConsentEncrypted(suite.ctx, consent)
	suite.Require().NoError(err)

	// Verify consent was stored
	consents, err := suite.keeper.GetGDPRConsents(suite.ctx, "cosmos1gdpr1")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(consents)
	suite.Require().NotEmpty(consents[0].AuditCommitment)
}

func (suite *EncryptedKVStoreTestSuite) TestSetGDPRConsentEncrypted_NoEncryptionService() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	consent := &types.GDPRConsent{
		Address:        "cosmos1gdpr2",
		ConsentType:    "marketing",
		Consented:      false,
		ConsentGivenAt: time.Now(),
		ConsentVersion: "v1.0",
	}

	err := keeperNoEnc.SetGDPRConsentEncrypted(ts.Ctx, consent)
	suite.Require().NoError(err)
}

func (suite *EncryptedKVStoreTestSuite) TestGetGDPRConsentsEncrypted_Success() {
	consent := &types.GDPRConsent{
		Address:        "cosmos1gdpr3",
		ConsentType:    "analytics",
		Consented:      true,
		ConsentGivenAt: time.Now(),
		ConsentVersion: "v1.5",
	}

	err := suite.keeper.SetGDPRConsentEncrypted(suite.ctx, consent)
	suite.Require().NoError(err)

	retrieved, err := suite.keeper.GetGDPRConsentsEncrypted(suite.ctx, "cosmos1gdpr3")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(retrieved)
	suite.Require().Equal("cosmos1gdpr3", retrieved[0].Address)
}

func (suite *EncryptedKVStoreTestSuite) TestGetGDPRConsentsEncrypted_NoEncryption() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	keeperNoEnc := ts.Keeper

	consent := &types.GDPRConsent{
		Address:        "cosmos1gdpr4",
		ConsentType:    "essential",
		Consented:      true,
		ConsentGivenAt: time.Now(),
		ConsentVersion: "v1.0",
	}

	err := keeperNoEnc.SetGDPRConsent(ts.Ctx, consent)
	suite.Require().NoError(err)

	retrieved, err := keeperNoEnc.GetGDPRConsentsEncrypted(ts.Ctx, "cosmos1gdpr4")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(retrieved)
}

// ============================================================================
// Utility Methods Tests
// ============================================================================

func (suite *EncryptedKVStoreTestSuite) TestIsEncryptionEnabled() {
	enabled := suite.keeper.IsEncryptionEnabled()
	suite.Require().True(enabled)

	// Test without encryption service
	ts := keeper.NewTestSuite()
	ts.SetupTest()
	enabled = ts.Keeper.IsEncryptionEnabled()
	suite.Require().False(enabled)
}

func (suite *EncryptedKVStoreTestSuite) TestEncryptDecryptField() {
	plaintext := []byte("sensitive data")
	context := "test-context"

	encrypted, err := suite.keeper.EncryptField(plaintext, context)
	suite.Require().NoError(err)
	suite.Require().NotEqual(plaintext, encrypted)

	decrypted, err := suite.keeper.DecryptField(encrypted, context)
	suite.Require().NoError(err)
	suite.Require().Equal(plaintext, decrypted)
}

func (suite *EncryptedKVStoreTestSuite) TestEncryptField_NoService() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()

	_, err := ts.Keeper.EncryptField([]byte("test"), "context")
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "encryption service not configured")
}

func (suite *EncryptedKVStoreTestSuite) TestDecryptField_NoService() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()

	_, err := ts.Keeper.DecryptField([]byte("test"), "context")
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "encryption service not configured")
}

func (suite *EncryptedKVStoreTestSuite) TestEncryptDecryptJSON() {
	data := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
		"field3": true,
	}
	context := "json-context"

	encrypted, err := suite.keeper.EncryptJSON(data, context)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(encrypted)

	var decrypted map[string]interface{}
	err = suite.keeper.DecryptJSON(encrypted, context, &decrypted)
	suite.Require().NoError(err)
	suite.Require().Equal("value1", decrypted["field1"])
	suite.Require().Equal(float64(123), decrypted["field2"]) // JSON numbers are float64
	suite.Require().Equal(true, decrypted["field3"])
}

func (suite *EncryptedKVStoreTestSuite) TestEncryptJSON_NoService() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()

	data := map[string]string{"key": "value"}
	_, err := ts.Keeper.EncryptJSON(data, "context")
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "encryption service not configured")
}

func (suite *EncryptedKVStoreTestSuite) TestDecryptJSON_NoService() {
	ts := keeper.NewTestSuite()
	ts.SetupTest()

	var target map[string]string
	err := ts.Keeper.DecryptJSON([]byte("test"), "context", &target)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "encryption service not configured")
}

func (suite *EncryptedKVStoreTestSuite) TestEncryptDecryptField_WrongContext() {
	plaintext := []byte("sensitive data")
	context1 := "context1"
	context2 := "context2"

	encrypted, err := suite.keeper.EncryptField(plaintext, context1)
	suite.Require().NoError(err)

	// Decrypting with wrong context should fail
	_, err = suite.keeper.DecryptField(encrypted, context2)
	suite.Require().Error(err)
}

// Test coverage for helper function
func TestEncryptedKVStore(t *testing.T) {
	suite.Run(t, new(EncryptedKVStoreTestSuite))
}
