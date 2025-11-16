package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

func TestNewKeeper(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	if keeper == nil {
		t.Fatal("NewKeeper returned nil")
	}

	// Verify default monitoring rules were created
	rules := keeper.GetMonitoringRules()
	if len(rules) == 0 {
		t.Error("No default monitoring rules created")
	}

	expectedRules := []string{
		"velocity_24h",
		"large_transaction",
		"structuring",
		"smurfing",
		"round_amounts",
		"rapid_movement",
		"high_risk_jurisdiction",
	}

	for _, ruleID := range expectedRules {
		if _, exists := rules[ruleID]; !exists {
			t.Errorf("Expected rule %s not found", ruleID)
		}
	}
}

func TestSubmitKYC(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	tests := []struct {
		name           string
		address        string
		kycLevel       types.KYCLevel
		provider       string
		verificationID string
		documents      []string
		jurisdiction   string
		expectError    bool
	}{
		{
			name:           "Valid KYC submission",
			address:        "aura1abc123",
			kycLevel:       types.KYCLevelBasic,
			provider:       "TestProvider",
			verificationID: "VER123",
			documents:      []string{"passport", "utility_bill"},
			jurisdiction:   "US",
			expectError:    false,
		},
		{
			name:           "Empty address",
			address:        "",
			kycLevel:       types.KYCLevelBasic,
			provider:       "TestProvider",
			verificationID: "VER123",
			documents:      []string{},
			jurisdiction:   "US",
			expectError:    true,
		},
		{
			name:           "Empty provider",
			address:        "aura1abc123",
			kycLevel:       types.KYCLevelBasic,
			provider:       "",
			verificationID: "VER123",
			documents:      []string{},
			jurisdiction:   "US",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.SubmitKYC(
				tt.address,
				tt.kycLevel,
				tt.provider,
				tt.verificationID,
				tt.documents,
				tt.jurisdiction,
			)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify KYC record was stored
				record, err := keeper.GetKYCRecord(tt.address)
				if err != nil {
					t.Errorf("Failed to retrieve KYC record: %v", err)
				}
				if record.Address != tt.address {
					t.Errorf("Address mismatch: got %s, want %s", record.Address, tt.address)
				}
				if record.KYCLevel != tt.kycLevel {
					t.Errorf("KYC level mismatch: got %d, want %d", record.KYCLevel, tt.kycLevel)
				}
			}
		})
	}
}

func TestValidateKYCLevel(t *testing.T) {
	params := types.DefaultParams()
	params.KYCRequired = true
	params.MinimumKYCLevel = types.KYCLevelBasic
	keeper := NewKeeper(params)

	// Submit valid KYC
	err := keeper.SubmitKYC(
		"aura1test",
		types.KYCLevelAdvanced,
		"TestProvider",
		"VER123",
		[]string{"passport"},
		"US",
	)
	if err != nil {
		t.Fatalf("Failed to submit KYC: %v", err)
	}

	// Validate - should pass
	err = keeper.ValidateKYCLevel("aura1test")
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test address without KYC
	err = keeper.ValidateKYCLevel("aura1nokyc")
	if err == nil {
		t.Error("Expected error for address without KYC")
	}

	// Test insufficient KYC level
	err = keeper.SubmitKYC(
		"aura1low",
		types.KYCLevelNone,
		"TestProvider",
		"VER456",
		[]string{},
		"US",
	)
	if err != nil {
		t.Fatalf("Failed to submit KYC: %v", err)
	}

	err = keeper.ValidateKYCLevel("aura1low")
	if err == nil {
		t.Error("Expected error for insufficient KYC level")
	}
}

func TestReportSuspiciousActivity(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	activityID, err := keeper.ReportSuspiciousActivity(
		"reporter1",
		"aura1suspicious",
		"txhash123",
		"structuring",
		"Multiple transactions just below threshold",
		"9900",
		[]string{"high_frequency", "round_amounts"},
	)

	if err != nil {
		t.Fatalf("Failed to report suspicious activity: %v", err)
	}

	if activityID == "" {
		t.Error("Activity ID should not be empty")
	}

	// Verify AML profile was created/updated
	profile, err := keeper.GetAMLProfile("aura1suspicious")
	if err != nil {
		t.Fatalf("Failed to get AML profile: %v", err)
	}

	if len(profile.SuspiciousActivities) == 0 {
		t.Error("Suspicious activity not added to profile")
	}

	// Risk level should increase from Low (but may not reach Medium with just one SA)
	if profile.RiskLevel == types.AMLRiskUnspecified {
		t.Error("Risk level should be set after suspicious activity")
	}
}

func TestMonitorTransaction(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	// Test large transaction detection
	alerts, err := keeper.MonitorTransaction(
		"txhash123",
		"aura1sender",
		"aura1receiver",
		"150000", // Exceeds single transaction limit of 100000
		time.Now(),
	)

	if err != nil {
		t.Fatalf("Transaction monitoring failed: %v", err)
	}

	if len(alerts) == 0 {
		t.Error("Expected alert for large transaction")
	}

	// Verify alert was stored
	storedAlerts, err := keeper.GetTransactionAlerts("aura1sender", false)
	if err != nil {
		t.Fatalf("Failed to get alerts: %v", err)
	}

	if len(storedAlerts) == 0 {
		t.Error("Alert not stored")
	}
}

func TestScreenSanctions(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	result, err := keeper.ScreenSanctions("aura1test", false)
	if err != nil {
		t.Fatalf("Sanctions screening failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Address != "aura1test" {
		t.Errorf("Address mismatch: got %s, want aura1test", result.Address)
	}

	// Result should be cached
	result2, err := keeper.ScreenSanctions("aura1test", false)
	if err != nil {
		t.Fatalf("Second screening failed: %v", err)
	}

	// Verify cache was used (timestamps should match)
	if !result.ScreenedAt.Equal(result2.ScreenedAt) {
		t.Error("Cache was not used")
	}

	// Add a small delay to ensure timestamps differ
	time.Sleep(10 * time.Millisecond)

	// Force refresh should create new result
	result3, err := keeper.ScreenSanctions("aura1test", true)
	if err != nil {
		t.Fatalf("Force refresh failed: %v", err)
	}

	if result.ScreenedAt.Equal(result3.ScreenedAt) || result3.ScreenedAt.Before(result.ScreenedAt) {
		t.Error("Force refresh did not create new result")
	}
}

func TestRecordGDPRConsent(t *testing.T) {
	params := types.DefaultParams()
	params.GDPREnabled = true
	keeper := NewKeeper(params)

	err := keeper.RecordGDPRConsent(
		"aura1test",
		"data_processing",
		true,
		"1.0",
		"192.168.1.1",
		"Mozilla/5.0",
	)

	if err != nil {
		t.Fatalf("Failed to record consent: %v", err)
	}

	// Retrieve consent
	consent, err := keeper.GetGDPRConsent("aura1test", "data_processing")
	if err != nil {
		t.Fatalf("Failed to get consent: %v", err)
	}

	if !consent.Consented {
		t.Error("Consent should be true")
	}

	if consent.ConsentVersion != "1.0" {
		t.Errorf("Version mismatch: got %s, want 1.0", consent.ConsentVersion)
	}
}

func TestRequestGDPRData(t *testing.T) {
	params := types.DefaultParams()
	params.GDPREnabled = true
	keeper := NewKeeper(params)

	// Add some data first
	keeper.SubmitKYC("aura1test", types.KYCLevelBasic, "Provider", "VER123", []string{}, "US")

	requestID, err := keeper.RequestGDPRData("aura1test", "access")
	if err != nil {
		t.Fatalf("Failed to request data: %v", err)
	}

	if requestID == "" {
		t.Error("Request ID should not be empty")
	}

	// Give time for async processing
	time.Sleep(2 * time.Second)

	// Retrieve request
	request, err := keeper.GetGDPRDataRequest(requestID)
	if err != nil {
		t.Fatalf("Failed to get request: %v", err)
	}

	if request.Status != "completed" {
		t.Errorf("Request should be completed, got: %s", request.Status)
	}

	if request.FulfillmentData == "" {
		t.Error("Fulfillment data should not be empty")
	}
}

func TestGenerateTaxReport(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	transactions := []*types.TaxTransaction{
		{
			TransactionHash: "tx1",
			Timestamp:       time.Now(),
			TransactionType: "trade",
			Asset:           "AURA",
			Amount:          "1000",
			CostBasis:       "500",
			FairMarketValue: "1000",
			GainLoss:        "500",
			IsIncome:        false,
		},
		{
			TransactionHash: "tx2",
			Timestamp:       time.Now(),
			TransactionType: "stake_reward",
			Asset:           "AURA",
			Amount:          "100",
			FairMarketValue: "100",
			IsIncome:        true,
		},
	}

	reportID, err := keeper.GenerateTaxReport(
		"aura1test",
		"2024",
		"US",
		"1099-MISC",
		transactions,
	)

	if err != nil {
		t.Fatalf("Failed to generate tax report: %v", err)
	}

	if reportID == "" {
		t.Error("Report ID should not be empty")
	}

	// Retrieve report
	report, err := keeper.GetTaxReport("aura1test", "2024")
	if err != nil {
		t.Fatalf("Failed to get tax report: %v", err)
	}

	if report.TotalIncome != "100" {
		t.Errorf("Total income mismatch: got %s, want 100", report.TotalIncome)
	}

	if report.TotalCapitalGains != "500" {
		t.Errorf("Total capital gains mismatch: got %s, want 500", report.TotalCapitalGains)
	}
}

func TestCalculateCapitalGains(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	acquiredDate := time.Now().AddDate(0, 0, -400) // 400 days ago
	soldDate := time.Now()

	gainLoss, isLongTerm, err := keeper.CalculateCapitalGains(
		"AURA",
		acquiredDate,
		soldDate,
		"1000",
		"1500",
	)

	if err != nil {
		t.Fatalf("Failed to calculate capital gains: %v", err)
	}

	if gainLoss != "500" {
		t.Errorf("Gain/loss mismatch: got %s, want 500", gainLoss)
	}

	if !isLongTerm {
		t.Error("Should be long-term (held > 1 year)")
	}

	// Test short-term
	acquiredDate2 := time.Now().AddDate(0, 0, -100) // 100 days ago
	_, isLongTerm2, err := keeper.CalculateCapitalGains(
		"AURA",
		acquiredDate2,
		soldDate,
		"1000",
		"1500",
	)

	if err != nil {
		t.Fatalf("Failed to calculate capital gains: %v", err)
	}

	if isLongTerm2 {
		t.Error("Should be short-term (held < 1 year)")
	}
}

func TestGetAlertStatistics(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	// Create some alerts
	keeper.MonitorTransaction("tx1", "aura1test", "aura1receiver", "150000", time.Now())
	keeper.MonitorTransaction("tx2", "aura1test", "aura1receiver", "160000", time.Now())

	stats := keeper.GetAlertStatistics()

	totalAlerts, ok := stats["total_alerts"].(int)
	if !ok || totalAlerts == 0 {
		t.Error("Expected alerts in statistics")
	}

	pendingAlerts, ok := stats["pending_alerts"].(int)
	if !ok || pendingAlerts == 0 {
		t.Error("Expected pending alerts")
	}
}

func TestGetHighRiskAddresses(t *testing.T) {
	params := types.DefaultParams()
	keeper := NewKeeper(params)

	// Create high-risk profile
	keeper.ReportSuspiciousActivity(
		"system",
		"aura1highrisk",
		"tx1",
		"money_laundering",
		"Suspicious pattern detected",
		"100000",
		[]string{"layering", "structuring", "smurfing"},
	)

	// Add multiple suspicious activities to increase risk
	keeper.ReportSuspiciousActivity(
		"system",
		"aura1highrisk",
		"tx2",
		"terrorist_financing",
		"Link to sanctioned entity",
		"50000",
		[]string{"sanctions_match"},
	)

	highRisk := keeper.GetHighRiskAddresses()
	if len(highRisk) == 0 {
		t.Error("Expected high-risk addresses")
	}

	found := false
	for _, addr := range highRisk {
		if addr == "aura1highrisk" {
			found = true
			break
		}
	}

	if !found {
		t.Error("High-risk address not found in results")
	}
}

func TestCleanupExpiredData(t *testing.T) {
	params := types.DefaultParams()
	params.DataRetentionDays = 1 // 1 day retention for testing
	keeper := NewKeeper(params)

	// Record old consent (simulate old data)
	oldConsent := &types.GDPRConsent{
		Address:        "aura1test",
		ConsentType:    "old_consent",
		Consented:      true,
		ConsentGivenAt: time.Now().AddDate(0, 0, -2), // 2 days ago
		ConsentVersion: "1.0",
	}

	keeper.mu.Lock()
	if keeper.gdprConsents["aura1test"] == nil {
		keeper.gdprConsents["aura1test"] = make(map[string]*types.GDPRConsent)
	}
	keeper.gdprConsents["aura1test"]["old_consent"] = oldConsent
	keeper.mu.Unlock()

	// Run cleanup
	cleaned := keeper.CleanupExpiredData()

	if cleaned == 0 {
		t.Error("Expected some data to be cleaned")
	}

	// Verify old consent was removed
	keeper.mu.RLock()
	_, exists := keeper.gdprConsents["aura1test"]["old_consent"]
	keeper.mu.RUnlock()

	if exists {
		t.Error("Old consent should have been removed")
	}
}
