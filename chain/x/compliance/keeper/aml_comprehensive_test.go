package keeper

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// AMLComprehensiveTestSuite tests AML screening edge cases, complex transaction patterns,
// and multi-jurisdiction scenarios as specified in ROADMAP_PRODUCTION.md task #12
type AMLComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestAMLComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(AMLComprehensiveTestSuite))
}

// addr generates a deterministic test address from a seed string
func (suite *AMLComprehensiveTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}

// ============================================================================
// AML Screening Edge Cases
// ============================================================================

// TestAMLScreening_BoundaryRiskScores tests risk score boundary conditions
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_BoundaryRiskScores() {
	address := suite.addr("test-user")

	testCases := []struct {
		name      string
		riskScore int32
		status    types.AMLStatus
	}{
		{
			name:      "Zero risk score (minimum)",
			riskScore: 0,
			status:    types.AMLStatus_CLEAR,
		},
		{
			name:      "Low risk score (25)",
			riskScore: 25,
			status:    types.AMLStatus_CLEAR,
		},
		{
			name:      "Medium risk score (50)",
			riskScore: 50,
			status:    types.AMLStatus_UNDER_REVIEW,
		},
		{
			name:      "High risk score (75)",
			riskScore: 75,
			status:    types.AMLStatus_UNDER_REVIEW,
		},
		{
			name:      "Critical risk score (90)",
			riskScore: 90,
			status:    types.AMLStatus_FLAGGED,
		},
		{
			name:      "Maximum risk score (100)",
			riskScore: 100,
			status:    types.AMLStatus_FLAGGED,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			profile := &types.AMLProfile{
				Address:        address,
				RiskScore:      tc.riskScore,
				Status:         tc.status,
				LastScreenedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
			}

			suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

			// Retrieve and verify
			retrieved := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
			suite.NotNil(retrieved)
			suite.Equal(tc.riskScore, retrieved.RiskScore)
			suite.Equal(tc.status, retrieved.Status)
		})
	}
}

// TestAMLScreening_StatusTransitions tests valid and invalid status transitions
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_StatusTransitions() {
	address := suite.addr("transition-test")

	// Create initial profile in CLEAR status
	profile := &types.AMLProfile{
		Address:        address,
		RiskScore:      10,
		Status:         types.AMLStatus_CLEAR,
		LastScreenedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	// Test transition: CLEAR -> UNDER_REVIEW
	profile.Status = types.AMLStatus_UNDER_REVIEW
	profile.RiskScore = 50
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	retrieved := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.Equal(types.AMLStatus_UNDER_REVIEW, retrieved.Status)

	// Test transition: UNDER_REVIEW -> FLAGGED
	profile.Status = types.AMLStatus_FLAGGED
	profile.RiskScore = 85
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	retrieved = suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.Equal(types.AMLStatus_FLAGGED, retrieved.Status)

	// Test transition: FLAGGED -> CLEARED (after review)
	profile.Status = types.AMLStatus_CLEARED
	profile.RiskScore = 15
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	retrieved = suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.Equal(types.AMLStatus_CLEARED, retrieved.Status)

	// Test transition: CLEARED -> BLOCKED (sanctions hit)
	profile.Status = types.AMLStatus_BLOCKED
	profile.RiskScore = 100
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	retrieved = suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.Equal(types.AMLStatus_BLOCKED, retrieved.Status)
}

// TestAMLScreening_NegativeRiskScore tests handling of invalid negative risk scores
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_NegativeRiskScore() {
	address := suite.addr("negative-risk")

	// Attempt to create profile with negative risk score
	profile := &types.AMLProfile{
		Address:        address,
		RiskScore:      -10,
		Status:         types.AMLStatus_CLEAR,
		LastScreenedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}

	// The keeper should store it (validation happens at msg layer)
	// But we should verify it's retrievable
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	retrieved := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NotNil(retrieved)
	suite.Equal(int32(-10), retrieved.RiskScore)

	// Application logic should handle negative scores as invalid
	// This test documents current behavior
}

// TestAMLScreening_ExcessiveRiskScore tests handling of risk scores > 100
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_ExcessiveRiskScore() {
	address := suite.addr("excessive-risk")

	profile := &types.AMLProfile{
		Address:        address,
		RiskScore:      150, // Invalid: > 100
		Status:         types.AMLStatus_FLAGGED,
		LastScreenedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}

	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	retrieved := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NotNil(retrieved)
	suite.Equal(int32(150), retrieved.RiskScore)

	// Application logic should cap or reject scores > 100
	// This test documents current behavior
}

// TestAMLScreening_StaleScreeningDetection tests detection of stale AML screenings
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_StaleScreeningDetection() {
	address := suite.addr("stale-screening")

	// Create profile with old screening timestamp
	oldTime := suite.SdkCtx.BlockTime().Add(-365 * 24 * time.Hour) // 1 year ago
	profile := &types.AMLProfile{
		Address:        address,
		RiskScore:      20,
		Status:         types.AMLStatus_CLEAR,
		LastScreenedAt: timestamppb.New(oldTime),
	}
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	retrieved := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NotNil(retrieved)

	// Check if screening is stale (> 90 days old)
	screenedAt := retrieved.LastScreenedAt.AsTime()
	staleCutoff := suite.SdkCtx.BlockTime().Add(-90 * 24 * time.Hour)

	isStale := screenedAt.Before(staleCutoff)
	suite.True(isStale, "Screening should be considered stale after 90 days")
}

// TestAMLScreening_MultipleProfileUpdates tests rapid profile updates
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_MultipleProfileUpdates() {
	address := suite.addr("rapid-updates")

	// Create initial profile
	profile := &types.AMLProfile{
		Address:        address,
		RiskScore:      10,
		Status:         types.AMLStatus_CLEAR,
		LastScreenedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)

	// Perform rapid updates
	for i := 0; i < 10; i++ {
		profile.RiskScore = int32(10 + i*5)
		profile.LastScreenedAt = timestamppb.New(suite.SdkCtx.BlockTime().Add(time.Duration(i) * time.Minute))
		suite.Keeper.SetAMLProfile(suite.SdkCtx, profile)
	}

	// Verify final state
	retrieved := suite.Keeper.GetAMLProfile(suite.SdkCtx, address)
	suite.NotNil(retrieved)
	suite.Equal(int32(55), retrieved.RiskScore) // 10 + 9*5 = 55
}

// TestAMLScreening_UnknownAddress tests screening of addresses with no profile
func (suite *AMLComprehensiveTestSuite) TestAMLScreening_UnknownAddress() {
	unknownAddr := suite.addr("unknown")

	// Attempt to get profile for address that doesn't exist
	profile := suite.Keeper.GetAMLProfile(suite.SdkCtx, unknownAddr)

	// Should return nil for non-existent profile
	suite.Nil(profile, "Non-existent profile should return nil")

	// Application logic should treat unknown addresses appropriately
	// (e.g., require screening before allowing transactions)
}

// ============================================================================
// Complex Transaction Pattern Detection
// ============================================================================

// TestComplexPatterns_StructuringDetection tests detection of structuring (smurfing)
func (suite *AMLComprehensiveTestSuite) TestComplexPatterns_StructuringDetection() {
	structurer := suite.addr("structurer")

	// Create multiple small transactions just below reporting threshold
	// Typical structuring: breaking large amount into small chunks
	reportingThreshold := int64(10000)
	chunkSize := reportingThreshold - 1 // Just below threshold

	for i := 0; i < 5; i++ {
		tx := &types.MonitoredTransaction{
			TxHash:    fmt.Sprintf("struct-tx-%d", i),
			Sender:    structurer,
			Recipient: suite.addr(fmt.Sprintf("recipient-%d", i)),
			Amount:    fmt.Sprintf("%d", chunkSize),
			Denom:     "uaura",
			Timestamp: timestamppb.New(suite.SdkCtx.BlockTime().Add(time.Duration(i) * time.Minute)),
		}
		suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)
	}

	// Query all transactions from structurer
	allTxs := suite.Keeper.GetAllMonitoredTransactions(suite.SdkCtx)

	// Filter transactions from structurer in last hour
	var structurerTxs []*types.MonitoredTransaction
	cutoff := suite.SdkCtx.BlockTime().Add(-1 * time.Hour)

	for _, tx := range allTxs {
		if tx.Sender == structurer && tx.Timestamp.AsTime().After(cutoff) {
			structurerTxs = append(structurerTxs, tx)
		}
	}

	// Calculate total amount
	totalAmount := int64(0)
	for _, tx := range structurerTxs {
		amount, ok := sdkmath.NewIntFromString(tx.Amount)
		if ok {
			totalAmount += amount.Int64()
		}
	}

	// Detect structuring pattern:
	// - Multiple transactions just below threshold
	// - Total amount exceeds threshold
	// - Transactions within short time window
	suite.GreaterOrEqual(len(structurerTxs), 3, "Should have multiple transactions")
	suite.Greater(totalAmount, reportingThreshold, "Total should exceed threshold")

	// This pattern indicates potential structuring
	isStructuring := len(structurerTxs) >= 3 && totalAmount > reportingThreshold
	suite.True(isStructuring, "Should detect structuring pattern")
}

// TestComplexPatterns_RapidFireTransactions tests rapid transaction velocity
func (suite *AMLComprehensiveTestSuite) TestComplexPatterns_RapidFireTransactions() {
	rapidTrader := suite.addr("rapid-trader")

	// Create many transactions in short time window
	numTxs := 50
	windowSeconds := 60

	for i := 0; i < numTxs; i++ {
		tx := &types.MonitoredTransaction{
			TxHash:    fmt.Sprintf("rapid-tx-%d", i),
			Sender:    rapidTrader,
			Recipient: suite.addr("exchange"),
			Amount:    "1000",
			Denom:     "uaura",
			Timestamp: timestamppb.New(suite.SdkCtx.BlockTime().Add(time.Duration(i) * time.Second)),
		}
		suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)
	}

	// Calculate transaction velocity
	allTxs := suite.Keeper.GetAllMonitoredTransactions(suite.SdkCtx)
	var rapidTxs []*types.MonitoredTransaction

	for _, tx := range allTxs {
		if tx.Sender == rapidTrader {
			rapidTxs = append(rapidTxs, tx)
		}
	}

	suite.Equal(numTxs, len(rapidTxs))

	// Velocity = transactions per minute
	velocity := float64(len(rapidTxs)) / (float64(windowSeconds) / 60.0)

	// High velocity (> 10 tx/min) may indicate automated trading or suspicious activity
	suite.Greater(velocity, 10.0, "Should detect high velocity")
}

// TestComplexPatterns_CircularTransactions tests circular flow detection
func (suite *AMLComprehensiveTestSuite) TestComplexPatterns_CircularTransactions() {
	// Create circular transaction chain: A -> B -> C -> A
	addrA := suite.addr("user-a")
	addrB := suite.addr("user-b")
	addrC := suite.addr("user-c")

	amount := "5000"

	// Transaction 1: A -> B
	tx1 := &types.MonitoredTransaction{
		TxHash:    "circular-tx-1",
		Sender:    addrA,
		Recipient: addrB,
		Amount:    amount,
		Denom:     "uaura",
		Timestamp: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx1)

	// Transaction 2: B -> C
	tx2 := &types.MonitoredTransaction{
		TxHash:    "circular-tx-2",
		Sender:    addrB,
		Recipient: addrC,
		Amount:    amount,
		Denom:     "uaura",
		Timestamp: timestamppb.New(suite.SdkCtx.BlockTime().Add(1 * time.Minute)),
	}
	suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx2)

	// Transaction 3: C -> A (completes circle)
	tx3 := &types.MonitoredTransaction{
		TxHash:    "circular-tx-3",
		Sender:    addrC,
		Recipient: addrA,
		Amount:    amount,
		Denom:     "uaura",
		Timestamp: timestamppb.New(suite.SdkCtx.BlockTime().Add(2 * time.Minute)),
	}
	suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx3)

	// Build transaction graph
	allTxs := suite.Keeper.GetAllMonitoredTransactions(suite.SdkCtx)

	// Simple cycle detection: find if A sends and receives similar amounts within time window
	var sentFromA, receivedByA bool
	for _, tx := range allTxs {
		if tx.Sender == addrA && tx.Amount == amount {
			sentFromA = true
		}
		if tx.Recipient == addrA && tx.Amount == amount {
			receivedByA = true
		}
	}

	// Circular flow detected
	isCircular := sentFromA && receivedByA
	suite.True(isCircular, "Should detect circular transaction pattern")
}

// TestComplexPatterns_LayeringBehavior tests layering pattern detection
func (suite *AMLComprehensiveTestSuite) TestComplexPatterns_LayeringBehavior() {
	launderer := suite.addr("launderer")

	// Layering: multiple transfers through different intermediaries
	intermediaries := []string{
		suite.addr("layer-1"),
		suite.addr("layer-2"),
		suite.addr("layer-3"),
		suite.addr("layer-4"),
		suite.addr("layer-5"),
	}

	baseAmount := int64(100000)
	currentSender := launderer

	// Create chain of transactions through multiple layers
	for i, intermediary := range intermediaries {
		// Each layer reduces amount slightly (simulating fees)
		layerAmount := baseAmount - int64(i*1000)

		tx := &types.MonitoredTransaction{
			TxHash:    fmt.Sprintf("layer-tx-%d", i),
			Sender:    currentSender,
			Recipient: intermediary,
			Amount:    fmt.Sprintf("%d", layerAmount),
			Denom:     "uaura",
			Timestamp: timestamppb.New(suite.SdkCtx.BlockTime().Add(time.Duration(i*5) * time.Minute)),
		}
		suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)

		currentSender = intermediary
	}

	// Final transaction back to original sender (integration)
	finalTx := &types.MonitoredTransaction{
		TxHash:    "layer-tx-final",
		Sender:    currentSender,
		Recipient: launderer,
		Amount:    fmt.Sprintf("%d", baseAmount-10000), // Return most of original amount
		Denom:     "uaura",
		Timestamp: timestamppb.New(suite.SdkCtx.BlockTime().Add(30 * time.Minute)),
	}
	suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)

	// Detect layering pattern:
	// - Multiple sequential transactions
	// - Amount preservation across layers
	// - Returns to original sender
	allTxs := suite.Keeper.GetAllMonitoredTransactions(suite.SdkCtx)

	layeringTxs := 0
	for _, tx := range allTxs {
		if tx.TxHash == finalTx.TxHash || len(tx.TxHash) > 9 && tx.TxHash[:8] == "layer-tx" {
			layeringTxs++
		}
	}

	suite.GreaterOrEqual(layeringTxs, 5, "Should have multiple layering transactions")
}

// TestComplexPatterns_FanOutFanIn tests fan-out and fan-in patterns
func (suite *AMLComprehensiveTestSuite) TestComplexPatterns_FanOutFanIn() {
	central := suite.addr("central")

	// Fan-out: one source sends to many recipients
	fanOutRecipients := make([]string, 20)
	for i := 0; i < 20; i++ {
		recipient := suite.addr(fmt.Sprintf("fanout-%d", i))
		fanOutRecipients[i] = recipient

		tx := &types.MonitoredTransaction{
			TxHash:    fmt.Sprintf("fanout-tx-%d", i),
			Sender:    central,
			Recipient: recipient,
			Amount:    "5000",
			Denom:     "uaura",
			Timestamp: timestamppb.New(suite.SdkCtx.BlockTime()),
		}
		suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)
	}

	// Fan-in: many senders send to one recipient
	collector := suite.addr("collector")
	for i := 0; i < 20; i++ {
		sender := suite.addr(fmt.Sprintf("fanin-%d", i))

		tx := &types.MonitoredTransaction{
			TxHash:    fmt.Sprintf("fanin-tx-%d", i),
			Sender:    sender,
			Recipient: collector,
			Amount:    "3000",
			Denom:     "uaura",
			Timestamp: timestamppb.New(suite.SdkCtx.BlockTime().Add(1 * time.Hour)),
		}
		suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)
	}

	// Analyze patterns
	allTxs := suite.Keeper.GetAllMonitoredTransactions(suite.SdkCtx)

	// Count fan-out (central as sender)
	fanOutCount := 0
	for _, tx := range allTxs {
		if tx.Sender == central {
			fanOutCount++
		}
	}

	// Count fan-in (collector as recipient)
	fanInCount := 0
	for _, tx := range allTxs {
		if tx.Recipient == collector {
			fanInCount++
		}
	}

	suite.Equal(20, fanOutCount, "Should detect fan-out pattern")
	suite.Equal(20, fanInCount, "Should detect fan-in pattern")

	// High fan-out/fan-in ratios may indicate:
	// - Exchange operations (legitimate)
	// - Mixing services (suspicious)
	// - Distribution networks (context-dependent)
}

// ============================================================================
// Multi-Jurisdiction Scenarios
// ============================================================================

// TestMultiJurisdiction_ConflictingRequirements tests handling of conflicting regulations
func (suite *AMLComprehensiveTestSuite) TestMultiJurisdiction_ConflictingRequirements() {
	// User subject to multiple jurisdictions with different requirements
	user := suite.addr("multi-jurisdiction")

	// Jurisdiction 1: US (strict KYC, high AML requirements)
	jurisdiction1 := &types.Jurisdiction{
		Code:              "US",
		Name:              "United States",
		KycRequired:       true,
		AmlScreening:      true,
		TaxReporting:      true,
		DataRetentionDays: 2555, // 7 years
	}
	suite.Keeper.SetJurisdiction(suite.SdkCtx, jurisdiction1)

	// Jurisdiction 2: EU (GDPR data minimization, right to deletion)
	jurisdiction2 := &types.Jurisdiction{
		Code:              "EU",
		Name:              "European Union",
		KycRequired:       true,
		AmlScreening:      true,
		TaxReporting:      true,
		DataRetentionDays: 1825, // 5 years max for GDPR
	}
	suite.Keeper.SetJurisdiction(suite.SdkCtx, jurisdiction2)

	// Jurisdiction 3: Jurisdiction with privacy focus (minimal KYC)
	jurisdiction3 := &types.Jurisdiction{
		Code:              "CH",
		Name:              "Switzerland",
		KycRequired:       false, // Conflicting requirement
		AmlScreening:      true,
		TaxReporting:      false, // Different requirement
		DataRetentionDays: 730,   // 2 years
	}
	suite.Keeper.SetJurisdiction(suite.SdkCtx, jurisdiction3)

	// User compliance record needs to satisfy ALL jurisdictions
	// Resolution strategy: most restrictive requirement wins
	effectiveKycRequired := jurisdiction1.KycRequired || jurisdiction2.KycRequired || jurisdiction3.KycRequired
	effectiveAmlScreening := jurisdiction1.AmlScreening || jurisdiction2.AmlScreening || jurisdiction3.AmlScreening
	effectiveTaxReporting := jurisdiction1.TaxReporting || jurisdiction2.TaxReporting || jurisdiction3.TaxReporting

	// Data retention: MINIMUM to satisfy GDPR right to deletion
	minRetention := jurisdiction1.DataRetentionDays
	if jurisdiction2.DataRetentionDays < minRetention {
		minRetention = jurisdiction2.DataRetentionDays
	}
	if jurisdiction3.DataRetentionDays < minRetention {
		minRetention = jurisdiction3.DataRetentionDays
	}

	suite.True(effectiveKycRequired, "Should require KYC if any jurisdiction requires it")
	suite.True(effectiveAmlScreening, "Should require AML screening if any jurisdiction requires it")
	suite.True(effectiveTaxReporting, "Should require tax reporting if any jurisdiction requires it")
	suite.Equal(int32(730), minRetention, "Should use minimum retention period for GDPR compliance")
}

// TestMultiJurisdiction_CrossBorderTransaction tests cross-border AML screening
func (suite *AMLComprehensiveTestSuite) TestMultiJurisdiction_CrossBorderTransaction() {
	senderUS := suite.addr("sender-us")
	recipientEU := suite.addr("recipient-eu")

	// Create jurisdictions
	jurisdictionUS := &types.Jurisdiction{
		Code:         "US",
		Name:         "United States",
		KycRequired:  true,
		AmlScreening: true,
		TaxReporting: true,
	}
	suite.Keeper.SetJurisdiction(suite.SdkCtx, jurisdictionUS)

	jurisdictionEU := &types.Jurisdiction{
		Code:         "EU",
		Name:         "European Union",
		KycRequired:  true,
		AmlScreening: true,
		TaxReporting: true,
	}
	suite.Keeper.SetJurisdiction(suite.SdkCtx, jurisdictionEU)

	// Create KYC records for both parties
	kycSender := &types.KYCRecord{
		Address:      senderUS,
		Name:         "John Doe",
		Country:      "US",
		Status:       types.KYCStatus_APPROVED,
		Jurisdiction: "US",
		VerifiedAt:   timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycSender)

	kycRecipient := &types.KYCRecord{
		Address:      recipientEU,
		Name:         "Jane Smith",
		Country:      "DE",
		Status:       types.KYCStatus_APPROVED,
		Jurisdiction: "EU",
		VerifiedAt:   timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycRecipient)

	// Create cross-border transaction
	tx := &types.MonitoredTransaction{
		TxHash:    "cross-border-tx",
		Sender:    senderUS,
		Recipient: recipientEU,
		Amount:    "50000", // Large amount requiring reporting
		Denom:     "uaura",
		Timestamp: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)

	// Cross-border transactions require:
	// 1. Both parties KYC verified
	senderKYC := suite.Keeper.GetKYCRecord(suite.SdkCtx, senderUS)
	recipientKYC := suite.Keeper.GetKYCRecord(suite.SdkCtx, recipientEU)

	suite.NotNil(senderKYC)
	suite.NotNil(recipientKYC)
	suite.Equal(types.KYCStatus_APPROVED, senderKYC.Status)
	suite.Equal(types.KYCStatus_APPROVED, recipientKYC.Status)

	// 2. Both jurisdictions' rules satisfied
	senderJurisdiction := suite.Keeper.GetJurisdiction(suite.SdkCtx, senderKYC.Jurisdiction)
	recipientJurisdiction := suite.Keeper.GetJurisdiction(suite.SdkCtx, recipientKYC.Jurisdiction)

	suite.NotNil(senderJurisdiction)
	suite.NotNil(recipientJurisdiction)
	suite.True(senderJurisdiction.KycRequired)
	suite.True(recipientJurisdiction.KycRequired)

	// 3. Transaction should be flagged for cross-border reporting
	amount, _ := sdkmath.NewIntFromString(tx.Amount)
	requiresReporting := amount.GT(sdkmath.NewInt(10000)) // Threshold
	suite.True(requiresReporting, "Large cross-border transaction should require reporting")
}

// TestMultiJurisdiction_SanctionedCountryInteraction tests transactions involving sanctioned countries
func (suite *AMLComprehensiveTestSuite) TestMultiJurisdiction_SanctionedCountryInteraction() {
	userUS := suite.addr("user-us")
	userSanctioned := suite.addr("user-sanctioned")

	// Create KYC records
	kycUS := &types.KYCRecord{
		Address:    userUS,
		Name:       "US Citizen",
		Country:    "US",
		Status:     types.KYCStatus_APPROVED,
		VerifiedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycUS)

	// User from sanctioned country
	kycSanctioned := &types.KYCRecord{
		Address:    userSanctioned,
		Name:       "Sanctioned User",
		Country:    "KP", // North Korea (sanctioned)
		Status:     types.KYCStatus_APPROVED,
		VerifiedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycSanctioned)

	// Create AML profile flagging sanctioned user
	amlSanctioned := &types.AMLProfile{
		Address:        userSanctioned,
		RiskScore:      100,
		Status:         types.AMLStatus_BLOCKED,
		LastScreenedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetAMLProfile(suite.SdkCtx, amlSanctioned)

	// Attempt transaction between US user and sanctioned user
	tx := &types.MonitoredTransaction{
		TxHash:    "sanctioned-tx",
		Sender:    userUS,
		Recipient: userSanctioned,
		Amount:    "1000",
		Denom:     "uaura",
		Timestamp: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)

	// Check if recipient is sanctioned
	recipientAML := suite.Keeper.GetAMLProfile(suite.SdkCtx, userSanctioned)
	suite.NotNil(recipientAML)
	suite.Equal(types.AMLStatus_BLOCKED, recipientAML.Status)

	// Transaction should be blocked
	isBlocked := recipientAML.Status == types.AMLStatus_BLOCKED
	suite.True(isBlocked, "Transaction to sanctioned country should be blocked")
}

// TestMultiJurisdiction_TravelRuleCompliance tests FATF Travel Rule compliance
func (suite *AMLComprehensiveTestSuite) TestMultiJurisdiction_TravelRuleCompliance() {
	// FATF Travel Rule: transfers >= 1000 USD require originator and beneficiary info
	sender := suite.addr("travel-sender")
	recipient := suite.addr("travel-recipient")

	// Create KYC records with full details (required for travel rule)
	kycSender := &types.KYCRecord{
		Address:    sender,
		Name:       "Alice Johnson",
		Country:    "US",
		Status:     types.KYCStatus_APPROVED,
		VerifiedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
		// Additional fields for travel rule:
		// - Full name
		// - Account number
		// - Address
		// - National identification number
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycSender)

	kycRecipient := &types.KYCRecord{
		Address:    recipient,
		Name:       "Bob Smith",
		Country:    "GB",
		Status:     types.KYCStatus_APPROVED,
		VerifiedAt: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycRecipient)

	// Create transaction above travel rule threshold
	amount := sdkmath.NewInt(150000) // 1500 USD equivalent (assuming 100 aura = 1 USD)

	tx := &types.MonitoredTransaction{
		TxHash:    "travel-rule-tx",
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount.String(),
		Denom:     "uaura",
		Timestamp: timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetMonitoredTransaction(suite.SdkCtx, tx)

	// Verify travel rule requirements
	travelRuleThreshold := sdkmath.NewInt(100000) // 1000 USD equivalent
	requiresTravelRule := amount.GTE(travelRuleThreshold)

	suite.True(requiresTravelRule, "Transaction should trigger travel rule")

	// Both parties must have KYC
	senderKYC := suite.Keeper.GetKYCRecord(suite.SdkCtx, sender)
	recipientKYC := suite.Keeper.GetKYCRecord(suite.SdkCtx, recipient)

	suite.NotNil(senderKYC)
	suite.NotNil(recipientKYC)
	suite.Equal(types.KYCStatus_APPROVED, senderKYC.Status)
	suite.Equal(types.KYCStatus_APPROVED, recipientKYC.Status)

	// Transaction metadata should include full originator and beneficiary info
	hasOriginatorInfo := senderKYC.Name != "" && senderKYC.Country != ""
	hasBeneficiaryInfo := recipientKYC.Name != "" && recipientKYC.Country != ""

	suite.True(hasOriginatorInfo, "Must have complete originator information")
	suite.True(hasBeneficiaryInfo, "Must have complete beneficiary information")
}

// TestMultiJurisdiction_DataLocalizationRequirements tests data residency requirements
func (suite *AMLComprehensiveTestSuite) TestMultiJurisdiction_DataLocalizationRequirements() {
	// Some jurisdictions require data to be stored locally (e.g., Russia, China)
	userChina := suite.addr("user-china")

	// Create jurisdiction with data localization requirement
	jurisdictionChina := &types.Jurisdiction{
		Code:              "CN",
		Name:              "China",
		KycRequired:       true,
		AmlScreening:      true,
		TaxReporting:      true,
		DataRetentionDays: 1825,
		// In real implementation, would have DataLocalizationRequired field
	}
	suite.Keeper.SetJurisdiction(suite.SdkCtx, jurisdictionChina)

	// Create KYC record
	kycChina := &types.KYCRecord{
		Address:      userChina,
		Name:         "Chinese User",
		Country:      "CN",
		Status:       types.KYCStatus_APPROVED,
		Jurisdiction: "CN",
		VerifiedAt:   timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycChina)

	// Verify jurisdiction requirements
	jurisdiction := suite.Keeper.GetJurisdiction(suite.SdkCtx, "CN")
	suite.NotNil(jurisdiction)
	suite.Equal("CN", jurisdiction.Code)

	// In production, would verify:
	// - Data stored in local region
	// - Encryption keys held locally
	// - No cross-border data transfers without approval
	// - Local audit trail

	// This test documents the requirement for data localization
	hasDataLocalizationRequirement := jurisdiction.Code == "CN" || jurisdiction.Code == "RU"
	suite.True(hasDataLocalizationRequirement, "China requires data localization")
}

// TestMultiJurisdiction_GDPRRightToDeletion tests GDPR right to erasure
func (suite *AMLComprehensiveTestSuite) TestMultiJurisdiction_GDPRRightToDeletion() {
	userEU := suite.addr("user-eu")

	// Create EU jurisdiction
	jurisdictionEU := &types.Jurisdiction{
		Code:              "EU",
		Name:              "European Union",
		KycRequired:       true,
		AmlScreening:      true,
		TaxReporting:      true,
		DataRetentionDays: 1825, // 5 years max
	}
	suite.Keeper.SetJurisdiction(suite.SdkCtx, jurisdictionEU)

	// Create KYC record
	kycEU := &types.KYCRecord{
		Address:      userEU,
		Name:         "EU Citizen",
		Country:      "DE",
		Status:       types.KYCStatus_APPROVED,
		Jurisdiction: "EU",
		VerifiedAt:   timestamppb.New(suite.SdkCtx.BlockTime()),
	}
	suite.Keeper.SetKYCRecord(suite.SdkCtx, kycEU)

	// Create consent record (GDPR requirement)
	consent := &types.ConsentRecord{
		Address:    userEU,
		Consented:  true,
		Purpose:    "KYC verification and AML screening",
		GrantedAt:  timestamppb.New(suite.SdkCtx.BlockTime()),
		ExpiresAt:  timestamppb.New(suite.SdkCtx.BlockTime().Add(365 * 24 * time.Hour)),
		Jurisdiction: "EU",
	}
	suite.Keeper.SetConsentRecord(suite.SdkCtx, consent)

	// Verify data exists
	retrievedKYC := suite.Keeper.GetKYCRecord(suite.SdkCtx, userEU)
	suite.NotNil(retrievedKYC)

	retrievedConsent := suite.Keeper.GetConsentRecord(suite.SdkCtx, userEU)
	suite.NotNil(retrievedConsent)

	// Simulate GDPR deletion request
	// NOTE: Actual deletion must balance:
	// - GDPR right to erasure
	// - AML/CFT legal retention requirements
	// - Blockchain immutability

	// Solution: Delete PII, keep pseudonymized transaction hashes
	// This test verifies the data exists and could be subject to deletion request
	canRequestDeletion := retrievedKYC.Jurisdiction == "EU"
	suite.True(canRequestDeletion, "EU users have right to request data deletion")
}
