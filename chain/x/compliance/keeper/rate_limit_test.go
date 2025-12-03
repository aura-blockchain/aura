package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestRateLimitEntry tests basic rate limit entry storage and retrieval
func TestRateLimitEntry(t *testing.T) {
	k, ctx := setupKeeper(t)

	address := "aura1test"
	operation := "sanctions_screening"

	// Should not exist initially
	_, found := k.GetRateLimitEntry(ctx, address, operation)
	require.False(t, found, "entry should not exist initially")

	// Create and store entry
	entry := &types.RateLimitEntry{
		Address:     address,
		Operation:   operation,
		Count:       5,
		WindowStart: timestamppb.New(ctx.BlockTime()),
	}

	err := k.SetRateLimitEntry(ctx, entry)
	require.NoError(t, err, "should store entry without error")

	// Retrieve and verify
	retrieved, found := k.GetRateLimitEntry(ctx, address, operation)
	require.True(t, found, "entry should exist after storage")
	require.Equal(t, entry.Address, retrieved.Address)
	require.Equal(t, entry.Operation, retrieved.Operation)
	require.Equal(t, entry.Count, retrieved.Count)

	// Delete entry
	k.DeleteRateLimitEntry(ctx, address, operation)
	_, found = k.GetRateLimitEntry(ctx, address, operation)
	require.False(t, found, "entry should not exist after deletion")
}

// TestCheckRateLimit_FirstRequest tests rate limit check for first request
func TestCheckRateLimit_FirstRequest(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:   3600, // 1 hour
		SanctionsScreeningLimit:  10,
		KycVerificationLimit:     5,
		AmlProfileQueryLimit:     50,
		TaxReportGenerationLimit: 3,
		DefaultQueryLimit:        100,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	operation := "sanctions_screening"

	// First request should succeed
	err = k.CheckRateLimit(ctx, address, operation)
	require.NoError(t, err, "first request should succeed")

	// Verify entry was created with count = 1
	entry, found := k.GetRateLimitEntry(ctx, address, operation)
	require.True(t, found, "entry should be created")
	require.Equal(t, int64(1), entry.Count, "count should be 1 after first request")
}

// TestCheckRateLimit_MultipleRequests tests multiple requests within limit
func TestCheckRateLimit_MultipleRequests(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600, // 1 hour
		SanctionsScreeningLimit: 10,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	operation := "sanctions_screening"

	// Make 10 requests (at the limit)
	for i := 1; i <= 10; i++ {
		err := k.CheckRateLimit(ctx, address, operation)
		require.NoError(t, err, "request %d should succeed", i)

		// Verify count
		entry, found := k.GetRateLimitEntry(ctx, address, operation)
		require.True(t, found)
		require.Equal(t, int64(i), entry.Count, "count should be %d after request %d", i, i)
	}
}

// TestCheckRateLimit_ExceedsLimit tests that limit is enforced
func TestCheckRateLimit_ExceedsLimit(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params with low limit
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600, // 1 hour
		SanctionsScreeningLimit: 3,    // Only 3 requests allowed
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	operation := "sanctions_screening"

	// Make 3 requests (at the limit)
	for i := 1; i <= 3; i++ {
		err := k.CheckRateLimit(ctx, address, operation)
		require.NoError(t, err, "request %d should succeed", i)
	}

	// 4th request should fail
	err = k.CheckRateLimit(ctx, address, operation)
	require.Error(t, err, "4th request should fail")
	require.Contains(t, err.Error(), "rate limit exceeded", "error should mention rate limit")
	require.Contains(t, err.Error(), "3/3", "error should show count/limit")

	// Verify count is still 3 (not incremented on failure)
	entry, found := k.GetRateLimitEntry(ctx, address, operation)
	require.True(t, found)
	require.Equal(t, int64(3), entry.Count, "count should remain 3 after failed request")
}

// TestCheckRateLimit_WindowReset tests that window resets after expiration
func TestCheckRateLimit_WindowReset(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  60, // 1 minute window for testing
		SanctionsScreeningLimit: 2,  // Only 2 requests per window
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	operation := "sanctions_screening"

	// Make 2 requests (at the limit)
	for i := 1; i <= 2; i++ {
		err := k.CheckRateLimit(ctx, address, operation)
		require.NoError(t, err, "request %d should succeed", i)
	}

	// 3rd request should fail
	err = k.CheckRateLimit(ctx, address, operation)
	require.Error(t, err, "3rd request should fail within window")

	// Advance time by 61 seconds (beyond window)
	newTime := ctx.BlockTime().Add(61 * time.Second)
	ctx = ctx.WithBlockTime(newTime)

	// Request should succeed after window reset
	err = k.CheckRateLimit(ctx, address, operation)
	require.NoError(t, err, "request should succeed after window reset")

	// Verify count was reset to 1
	entry, found := k.GetRateLimitEntry(ctx, address, operation)
	require.True(t, found)
	require.Equal(t, int64(1), entry.Count, "count should be 1 after window reset")
}

// TestCheckRateLimit_DifferentOperations tests that operations are tracked separately
func TestCheckRateLimit_DifferentOperations(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:   3600,
		SanctionsScreeningLimit:  2,
		KycVerificationLimit:     2,
		AmlProfileQueryLimit:     2,
		TaxReportGenerationLimit: 2,
		DefaultQueryLimit:        2,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"

	operations := []string{
		"sanctions_screening",
		"kyc_verification",
		"aml_profile_query",
		"tax_report_generation",
		"transaction_alerts",
	}

	// Each operation should have independent rate limit
	for _, operation := range operations {
		// Make 2 requests (at the limit) for each operation
		for i := 1; i <= 2; i++ {
			err := k.CheckRateLimit(ctx, address, operation)
			require.NoError(t, err, "request %d for %s should succeed", i, operation)
		}

		// 3rd request should fail for this operation
		err := k.CheckRateLimit(ctx, address, operation)
		require.Error(t, err, "3rd request for %s should fail", operation)
	}
}

// TestCheckRateLimit_DifferentAddresses tests that addresses are tracked separately
func TestCheckRateLimit_DifferentAddresses(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600,
		SanctionsScreeningLimit: 2,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	addresses := []string{"aura1alice", "aura1bob", "aura1charlie"}
	operation := "sanctions_screening"

	// Each address should have independent rate limit
	for _, address := range addresses {
		// Make 2 requests (at the limit) for each address
		for i := 1; i <= 2; i++ {
			err := k.CheckRateLimit(ctx, address, operation)
			require.NoError(t, err, "request %d for %s should succeed", i, address)
		}

		// 3rd request should fail for this address
		err := k.CheckRateLimit(ctx, address, operation)
		require.Error(t, err, "3rd request for %s should fail", address)
	}
}

// TestCheckRateLimit_DefaultLimits tests fallback to default limits
func TestCheckRateLimit_DefaultLimits(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set params with no rate limits configured (should use defaults)
	params := types.ComplianceParams{
		RateLimitWindowSeconds: 3600,
		// All limits are 0 (will use hardcoded defaults)
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"

	// Test defaults for each operation type
	testCases := []struct {
		operation    string
		defaultLimit int64
	}{
		{"sanctions_screening", 10},
		{"kyc_verification", 5},
		{"aml_profile_query", 50},
		{"tax_report_generation", 3},
		{"unknown_operation", 100},
	}

	for _, tc := range testCases {
		// Make requests up to the default limit
		for i := int64(1); i <= tc.defaultLimit; i++ {
			err := k.CheckRateLimit(ctx, address+tc.operation, tc.operation)
			require.NoError(t, err, "request %d for %s should succeed (default limit %d)", i, tc.operation, tc.defaultLimit)
		}

		// Next request should fail
		err := k.CheckRateLimit(ctx, address+tc.operation, tc.operation)
		require.Error(t, err, "request beyond default limit for %s should fail", tc.operation)
	}
}

// TestCheckRateLimit_ZeroWindowSeconds tests default window duration
func TestCheckRateLimit_ZeroWindowSeconds(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set params with zero window (should use 1 hour default)
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  0, // Will default to 1 hour
		SanctionsScreeningLimit: 2,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	operation := "sanctions_screening"

	// Make 2 requests
	for i := 1; i <= 2; i++ {
		err := k.CheckRateLimit(ctx, address, operation)
		require.NoError(t, err)
	}

	// 3rd request should fail
	err = k.CheckRateLimit(ctx, address, operation)
	require.Error(t, err)

	// Advance time by 59 minutes (within default 1 hour window)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(59 * time.Minute))
	err = k.CheckRateLimit(ctx, address, operation)
	require.Error(t, err, "should still be rate limited within 1 hour")

	// Advance time by 2 more minutes (beyond 1 hour)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Minute))
	err = k.CheckRateLimit(ctx, address, operation)
	require.NoError(t, err, "should succeed after 1 hour default window")
}

// TestQueryServer_SanctionsScreening_RateLimit tests rate limit enforcement in query handler
func TestQueryServer_SanctionsScreening_RateLimit(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := NewQueryServer(k)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:       3600,
		SanctionsScreeningLimit:      2,
		SanctionsScreeningEnabled:    true,
		SanctionsLists:               []string{"OFAC_SDN", "EU_SANCTIONS"},
		ScreeningCacheHours:          24,
		BlockedJurisdictions:         []string{},
		ApprovedKycProviders:         []string{},
		TransactionMonitoringEnabled: false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	req := &types.QuerySanctionsScreeningRequest{
		Address:      address,
		ForceRefresh: false,
	}

	// First 2 requests should succeed (or return not found, but not rate limited)
	for i := 1; i <= 2; i++ {
		_, err := qs.SanctionsScreening(sdk.WrapSDKContext(ctx), req)
		// May fail with NotFound, but should not be ResourceExhausted
		if err != nil {
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.NotEqual(t, codes.ResourceExhausted, st.Code(), "request %d should not be rate limited", i)
		}
	}

	// 3rd request should be rate limited
	_, err = qs.SanctionsScreening(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, st.Code(), "3rd request should be rate limited")
	require.Contains(t, st.Message(), "rate limit exceeded")
}

// TestQueryServer_KycRecord_RateLimit tests rate limit enforcement for KYC queries
func TestQueryServer_KycRecord_RateLimit(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := NewQueryServer(k)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600,
		KycVerificationLimit:    2,
		BlockedJurisdictions:    []string{},
		ApprovedKycProviders:    []string{},
		SanctionsScreeningEnabled: false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	req := &types.QueryKYCRecordRequest{
		Address: address,
	}

	// First 2 requests
	for i := 1; i <= 2; i++ {
		_, err := qs.KycRecord(sdk.WrapSDKContext(ctx), req)
		// May fail with NotFound, but should not be ResourceExhausted
		if err != nil {
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.NotEqual(t, codes.ResourceExhausted, st.Code(), "request %d should not be rate limited", i)
		}
	}

	// 3rd request should be rate limited
	_, err = qs.KycRecord(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, st.Code(), "3rd request should be rate limited")
}

// TestQueryServer_AmlProfile_RateLimit tests rate limit enforcement for AML queries
func TestQueryServer_AmlProfile_RateLimit(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := NewQueryServer(k)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600,
		AmlProfileQueryLimit:    2,
		BlockedJurisdictions:    []string{},
		ApprovedKycProviders:    []string{},
		SanctionsScreeningEnabled: false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	req := &types.QueryAMLProfileRequest{
		Address: address,
	}

	// First 2 requests
	for i := 1; i <= 2; i++ {
		_, err := qs.AmlProfile(sdk.WrapSDKContext(ctx), req)
		// May fail with NotFound, but should not be ResourceExhausted
		if err != nil {
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.NotEqual(t, codes.ResourceExhausted, st.Code(), "request %d should not be rate limited", i)
		}
	}

	// 3rd request should be rate limited
	_, err = qs.AmlProfile(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, st.Code(), "3rd request should be rate limited")
}

// TestQueryServer_TransactionAlerts_RateLimit tests rate limit enforcement for alerts
func TestQueryServer_TransactionAlerts_RateLimit(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := NewQueryServer(k)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600,
		DefaultQueryLimit:       2,
		BlockedJurisdictions:    []string{},
		ApprovedKycProviders:    []string{},
		SanctionsScreeningEnabled: false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	req := &types.QueryTransactionAlertsRequest{
		Address:        address,
		UnreviewedOnly: false,
	}

	// First 2 requests should succeed
	for i := 1; i <= 2; i++ {
		_, err := qs.TransactionAlerts(sdk.WrapSDKContext(ctx), req)
		require.NoError(t, err, "request %d should succeed", i)
	}

	// 3rd request should be rate limited
	_, err = qs.TransactionAlerts(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, st.Code(), "3rd request should be rate limited")
}

// TestQueryServer_TaxReport_RateLimit tests rate limit enforcement for tax reports
func TestQueryServer_TaxReport_RateLimit(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := NewQueryServer(k)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:   3600,
		TaxReportGenerationLimit: 2,
		BlockedJurisdictions:     []string{},
		ApprovedKycProviders:     []string{},
		SanctionsScreeningEnabled: false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	req := &types.QueryTaxReportRequest{
		Address:      address,
		TaxYear:      "2024",
		Jurisdiction: "US",
	}

	// First 2 requests
	for i := 1; i <= 2; i++ {
		_, err := qs.TaxReport(sdk.WrapSDKContext(ctx), req)
		// May fail with NotFound, but should not be ResourceExhausted
		if err != nil {
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.NotEqual(t, codes.ResourceExhausted, st.Code(), "request %d should not be rate limited", i)
		}
	}

	// 3rd request should be rate limited
	_, err = qs.TaxReport(sdk.WrapSDKContext(ctx), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.ResourceExhausted, st.Code(), "3rd request should be rate limited")
}

// TestRateLimit_EventEmission tests that rate limit exceeded events are emitted
func TestRateLimit_EventEmission(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600,
		SanctionsScreeningLimit: 2,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	operation := "sanctions_screening"

	// Make 2 requests (at the limit)
	for i := 1; i <= 2; i++ {
		err := k.CheckRateLimit(ctx, address, operation)
		require.NoError(t, err)
	}

	// 3rd request should fail and emit event
	err = k.CheckRateLimit(ctx, address, operation)
	require.Error(t, err)

	// Check for rate limit exceeded event
	events := ctx.EventManager().Events()
	found := false
	for _, event := range events {
		if event.Type == types.EventTypeRateLimitExceeded {
			found = true
			// Verify event attributes
			attrs := event.Attributes
			addressFound := false
			operationFound := false
			countFound := false
			limitFound := false

			for _, attr := range attrs {
				switch string(attr.Key) {
				case types.AttributeKeyAddress:
					require.Equal(t, address, string(attr.Value))
					addressFound = true
				case types.AttributeKeyOperation:
					require.Equal(t, operation, string(attr.Value))
					operationFound = true
				case types.AttributeKeyCount:
					require.Equal(t, "2", string(attr.Value))
					countFound = true
				case types.AttributeKeyLimit:
					require.Equal(t, "2", string(attr.Value))
					limitFound = true
				}
			}

			require.True(t, addressFound, "event should have address attribute")
			require.True(t, operationFound, "event should have operation attribute")
			require.True(t, countFound, "event should have count attribute")
			require.True(t, limitFound, "event should have limit attribute")
			break
		}
	}

	require.True(t, found, "rate_limit_exceeded event should be emitted")
}

// TestRateLimit_ErrorMessage tests that error messages are informative
func TestRateLimit_ErrorMessage(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600,
		SanctionsScreeningLimit: 5,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	address := "aura1test"
	operation := "sanctions_screening"

	// Exhaust limit
	for i := 1; i <= 5; i++ {
		err := k.CheckRateLimit(ctx, address, operation)
		require.NoError(t, err)
	}

	// Get rate limit error
	err = k.CheckRateLimit(ctx, address, operation)
	require.Error(t, err)

	// Verify error message contains useful information
	errMsg := err.Error()
	require.Contains(t, errMsg, "rate limit exceeded", "should mention rate limit")
	require.Contains(t, errMsg, operation, "should mention operation")
	require.Contains(t, errMsg, "5/5", "should show current count and limit")
	require.Contains(t, errMsg, "resets at", "should show when limit resets")

	// Error message should contain RFC3339 timestamp
	require.Regexp(t, `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, errMsg, "should contain RFC3339 timestamp")
}

// BenchmarkCheckRateLimit benchmarks the rate limit check performance
func BenchmarkCheckRateLimit(b *testing.B) {
	k, ctx := setupKeeperForBench(b)

	// Set rate limit params
	params := types.ComplianceParams{
		RateLimitWindowSeconds:  3600,
		SanctionsScreeningLimit: int64(b.N + 1000), // High enough to not hit limit
	}
	_ = k.SetParams(ctx, params)

	address := "aura1test"
	operation := "sanctions_screening"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = k.CheckRateLimit(ctx, address, operation)
	}
}

// setupKeeperForBench is a simplified setup for benchmarks
func setupKeeperForBench(b *testing.B) (*Keeper, sdk.Context) {
	k, ctx := setupKeeper(&testing.T{})
	return k, ctx
}
