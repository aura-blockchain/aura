// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestGetAllKYCRecordsPaginated tests pagination for KYC records
func TestGetAllKYCRecordsPaginated(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test records
	numRecords := 150
	for i := 0; i < numRecords; i++ {
		record := &types.KYCRecord{
			Address:    fmt.Sprintf("address%d", i),
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test-provider",
			VerifiedAt: time.Now(),
			ExpiresAt:  ptrTime(time.Now().Add(365 * 24 * time.Hour)),
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Test 1: First page with limit
	pageReq := &query.PageRequest{
		Limit: 50,
	}
	records, pageRes, err := keeper.GetAllKYCRecordsPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Len(t, records, 50, "First page should return 50 records")
	require.NotNil(t, pageRes)
	require.NotNil(t, pageRes.NextKey, "NextKey should be set for more pages")

	// Test 2: Second page using NextKey
	pageReq2 := &query.PageRequest{
		Key:   pageRes.NextKey,
		Limit: 50,
	}
	records2, pageRes2, err := keeper.GetAllKYCRecordsPaginated(ctx, pageReq2)
	require.NoError(t, err)
	require.Len(t, records2, 50, "Second page should return 50 records")
	require.NotNil(t, pageRes2.NextKey, "NextKey should be set for more pages")

	// Test 3: Last page
	pageReq3 := &query.PageRequest{
		Key:   pageRes2.NextKey,
		Limit: 100,
	}
	records3, pageRes3, err := keeper.GetAllKYCRecordsPaginated(ctx, pageReq3)
	require.NoError(t, err)
	require.Len(t, records3, 50, "Last page should return remaining 50 records")
	require.Nil(t, pageRes3.NextKey, "NextKey should be nil on last page")

	// Test 4: Count total
	pageReqCount := &query.PageRequest{
		Limit:      50,
		CountTotal: true,
	}
	_, pageResCount, err := keeper.GetAllKYCRecordsPaginated(ctx, pageReqCount)
	require.NoError(t, err)
	require.Equal(t, uint64(numRecords), pageResCount.Total, "Total count should match")

	// Test 5: Default pagination (nil request)
	recordsDefault, pageResDefault, err := keeper.GetAllKYCRecordsPaginated(ctx, nil)
	require.NoError(t, err)
	require.NotEmpty(t, recordsDefault, "Should return records with default pagination")
	require.NotNil(t, pageResDefault)
}

// TestGetAllAMLProfilesPaginated tests pagination for AML profiles
func TestGetAllAMLProfilesPaginated(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test profiles
	numProfiles := 75
	for i := 0; i < numProfiles; i++ {
		profile := &types.AMLProfile{
			Address:           fmt.Sprintf("address%d", i),
			RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
			TotalTransactions: uint64(i),
			TotalVolume:       math.NewInt(int64(i * 1000)).String(),
			LastAssessment:    time.Now(),
		}
		err := keeper.SetAMLProfile(ctx, profile)
		require.NoError(t, err)
	}

	// Test pagination
	pageReq := &query.PageRequest{
		Limit: 25,
	}
	profiles, pageRes, err := keeper.GetAllAMLProfilesPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Len(t, profiles, 25, "Should return 25 profiles")
	require.NotNil(t, pageRes.NextKey, "NextKey should be set")

	// Get second page
	pageReq2 := &query.PageRequest{
		Key:   pageRes.NextKey,
		Limit: 25,
	}
	profiles2, pageRes2, err := keeper.GetAllAMLProfilesPaginated(ctx, pageReq2)
	require.NoError(t, err)
	require.Len(t, profiles2, 25)
	require.NotNil(t, pageRes2.NextKey)

	// Verify no duplicate profiles between pages
	addressMap := make(map[string]bool)
	for _, p := range profiles {
		addressMap[p.Address] = true
	}
	for _, p := range profiles2 {
		require.False(t, addressMap[p.Address], "Should not have duplicate addresses across pages")
	}
}

// TestGetAllSanctionsResultsPaginated tests pagination for sanctions results
func TestGetAllSanctionsResultsPaginated(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test sanctions results
	numResults := 120
	for i := 0; i < numResults; i++ {
		result := &types.SanctionsScreeningResult{
			Address:           fmt.Sprintf("address%d", i),
			Status:            types.SanctionsStatus_SANCTIONS_CLEAR,
			ScreenedAt:        time.Now(),
			ScreeningProvider: "test-provider",
		}
		err := keeper.SetSanctionsResult(ctx, result)
		require.NoError(t, err)
	}

	// Test with count total
	pageReq := &query.PageRequest{
		Limit:      40,
		CountTotal: true,
	}
	results, pageRes, err := keeper.GetAllSanctionsResultsPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Len(t, results, 40)
	require.Equal(t, uint64(numResults), pageRes.Total)
	require.NotNil(t, pageRes.NextKey)
}

// TestGetAllGDPRConsentsPaginated tests pagination for GDPR consents
func TestGetAllGDPRConsentsPaginated(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test GDPR consents for different addresses
	numAddresses := 50
	for i := 0; i < numAddresses; i++ {
		consent := &types.GDPRConsent{
			Address:        fmt.Sprintf("address%d", i),
			ConsentType:    "data_processing",
			Consented:      true,
			ConsentGivenAt: time.Now(),
			ConsentVersion: "v1.0",
		}
		err := keeper.SetGDPRConsent(ctx, consent)
		require.NoError(t, err)
	}

	// Test pagination
	pageReq := &query.PageRequest{
		Limit: 20,
	}
	consents, pageRes, err := keeper.GetAllGDPRConsentsPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Len(t, consents, 20, "Should return 20 address consent lists")
	require.NotNil(t, pageRes.NextKey)

	// Verify each consent list has at least one consent
	for _, consentList := range consents {
		require.NotEmpty(t, consentList, "Each address should have consents")
	}
}

// TestGetAllTransactionAlertsPaginated tests pagination for transaction alerts
func TestGetAllTransactionAlertsPaginated(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test transaction alerts
	numAddresses := 30
	for i := 0; i < numAddresses; i++ {
		address := fmt.Sprintf("address%d", i)
		alert := &types.TransactionAlert{
			Id:              fmt.Sprintf("alert-%d", i),
			TransactionHash: fmt.Sprintf("txhash%d", i),
			Address:         address,
			RuleId:          "test-rule",
			RiskLevel:       types.TransactionRiskLevel_TX_RISK_MEDIUM,
			TriggeredAt:     time.Now(),
		}
		err := keeper.AddTransactionAlert(ctx, address, alert)
		require.NoError(t, err)
	}

	// Test pagination
	pageReq := &query.PageRequest{
		Limit: 10,
	}
	alerts, pageRes, err := keeper.GetAllTransactionAlertsPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Len(t, alerts, 10, "Should return 10 address alert lists")
	require.NotNil(t, pageRes.NextKey)
}

// TestGetAllTaxReportsPaginated tests pagination for tax reports
func TestGetAllTaxReportsPaginated(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test tax reports
	numAddresses := 40
	for i := 0; i < numAddresses; i++ {
		address := fmt.Sprintf("address%d", i)
		report := &types.TaxReport{
			Id:           fmt.Sprintf("report-%d", i),
			Address:      address,
			TaxYear:      "2024",
			Jurisdiction: "US",
			ReportType:   "1099-K",
			GeneratedAt:  time.Now(),
		}
		err := keeper.SetTaxReport(ctx, report)
		require.NoError(t, err)
	}

	// Test pagination
	pageReq := &query.PageRequest{
		Limit:      15,
		CountTotal: true,
	}
	reports, pageRes, err := keeper.GetAllTaxReportsPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Len(t, reports, 15, "Should return 15 address report lists")
	require.Equal(t, uint64(numAddresses), pageRes.Total)
}

// TestGetAllGDPRRequestsPaginated tests pagination for GDPR requests
func TestGetAllGDPRRequestsPaginated(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create test GDPR requests
	numRequests := 60
	for i := 0; i < numRequests; i++ {
		request := &types.GDPRDataRequest{
			Id:          fmt.Sprintf("request-%d", i),
			Address:     fmt.Sprintf("address%d", i),
			RequestType: "access",
			RequestedAt: time.Now(),
			Status:      "pending",
		}
		err := keeper.SetGDPRRequest(ctx, request)
		require.NoError(t, err)
	}

	// Test pagination
	pageReq := &query.PageRequest{
		Limit: 20,
	}
	requests, pageRes, err := keeper.GetAllGDPRRequestsPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Len(t, requests, 20)
	require.NotNil(t, pageRes.NextKey)

	// Get all pages and verify total count
	totalCount := len(requests)
	nextKey := pageRes.NextKey

	for nextKey != nil {
		pageReq := &query.PageRequest{
			Key:   nextKey,
			Limit: 20,
		}
		moreRequests, morePageRes, err := keeper.GetAllGDPRRequestsPaginated(ctx, pageReq)
		require.NoError(t, err)
		totalCount += len(moreRequests)
		nextKey = morePageRes.NextKey
	}

	require.Equal(t, numRequests, totalCount, "Total count across all pages should match")
}

// TestPaginationLargeDataset tests pagination with very large datasets
func TestPaginationLargeDataset(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create a large dataset
	numRecords := 500
	for i := 0; i < numRecords; i++ {
		record := &types.KYCRecord{
			Address:    fmt.Sprintf("address%03d", i), // Zero-padded for consistent ordering
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test-provider",
			VerifiedAt: time.Now(),
			ExpiresAt:  ptrTime(time.Now().Add(365 * 24 * time.Hour)),
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Test with small page size
	pageReq := &query.PageRequest{
		Limit:      10,
		CountTotal: true,
	}

	totalRetrieved := 0
	nextKey := []byte(nil)
	iterations := 0
	maxIterations := 60 // Should be enough for 500 records with limit 10

	for {
		pageReq.Key = nextKey
		records, pageRes, err := keeper.GetAllKYCRecordsPaginated(ctx, pageReq)
		require.NoError(t, err)
		require.NotEmpty(t, records, "Should return records")

		totalRetrieved += len(records)
		nextKey = pageRes.NextKey

		iterations++
		if iterations >= maxIterations {
			t.Fatal("Exceeded maximum iterations - possible infinite loop")
		}

		if nextKey == nil {
			break
		}
	}

	require.Equal(t, numRecords, totalRetrieved, "Should retrieve all records across pages")
	require.Less(t, iterations, maxIterations, "Should complete within reasonable iterations")
}

// TestPaginationEmptyStore tests pagination with empty store
func TestPaginationEmptyStore(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Test with empty store
	pageReq := &query.PageRequest{
		Limit: 50,
	}

	records, pageRes, err := keeper.GetAllKYCRecordsPaginated(ctx, pageReq)
	require.NoError(t, err)
	require.Empty(t, records, "Should return empty list for empty store")
	require.NotNil(t, pageRes)
	require.Nil(t, pageRes.NextKey, "NextKey should be nil for empty store")
}

// TestPaginationDefaultLimit tests default pagination behavior
func TestPaginationDefaultLimit(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create records
	numRecords := 200
	for i := 0; i < numRecords; i++ {
		record := &types.KYCRecord{
			Address:    fmt.Sprintf("address%d", i),
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test-provider",
			VerifiedAt: time.Now(),
			ExpiresAt:  ptrTime(time.Now().Add(365 * 24 * time.Hour)),
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Test with nil pagination (uses default)
	records, pageRes, err := keeper.GetAllKYCRecordsPaginated(ctx, nil)
	require.NoError(t, err)
	require.NotEmpty(t, records, "Should return records with default pagination")
	require.NotNil(t, pageRes)
	// Default limit is typically 100 in Cosmos SDK
	require.LessOrEqual(t, len(records), 100, "Should respect default limit")
}

// setupKeeperForTest creates a test keeper and context
// This uses the same setup as keeper_test.go for consistency
func setupKeeperForTest(t *testing.T) (*Keeper, sdk.Context) {
	return setupKeeper(t)
}
