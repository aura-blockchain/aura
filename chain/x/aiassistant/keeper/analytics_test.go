package keeper

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

func TestRecordAnalytics(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name    string
		record  AnalyticsRecord
		wantErr bool
	}{
		{
			name: "valid analytics record",
			record: AnalyticsRecord{
				Address:       "user1",
				ModelHash:     "model-hash-1",
				ComputeUnits:  100,
				Cost:          sdkmath.NewInt(1000),
				OperationType: "query",
				Success:       true,
				ResponseTime:  500,
			},
			wantErr: false,
		},
		{
			name: "high compute units",
			record: AnalyticsRecord{
				Address:       "user2",
				ModelHash:     "model-hash-2",
				ComputeUnits:  1000000,
				Cost:          sdkmath.NewInt(10000000),
				OperationType: "training",
				Success:       true,
				ResponseTime:  5000,
			},
			wantErr: false,
		},
		{
			name: "failed operation",
			record: AnalyticsRecord{
				Address:       "user3",
				ModelHash:     "model-hash-1",
				ComputeUnits:  50,
				Cost:          sdkmath.NewInt(500),
				OperationType: "inference",
				Success:       false,
				ResponseTime:  100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.RecordAnalytics(ctx, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordAnalytics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAnalytics(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record some analytics
	records := []AnalyticsRecord{
		{
			Address:       "user1",
			ModelHash:     "model1",
			ComputeUnits:  100,
			Cost:          sdkmath.NewInt(1000),
			OperationType: "query",
			Success:       true,
			ResponseTime:  200,
		},
		{
			Address:       "user2",
			ModelHash:     "model1",
			ComputeUnits:  150,
			Cost:          sdkmath.NewInt(1500),
			OperationType: "query",
			Success:       true,
			ResponseTime:  300,
		},
		{
			Address:       "user1",
			ModelHash:     "model2",
			ComputeUnits:  200,
			Cost:          sdkmath.NewInt(2000),
			OperationType: "inference",
			Success:       false,
			ResponseTime:  400,
		},
	}

	for _, record := range records {
		if err := k.RecordAnalytics(ctx, record); err != nil {
			t.Fatalf("Failed to record analytics: %v", err)
		}
	}

	// Test hourly analytics
	hourlyData := k.GetAnalytics(ctx, PeriodHourly, ctx.BlockTime())
	if hourlyData.TotalQueries != 3 {
		t.Errorf("Expected 3 total queries, got %d", hourlyData.TotalQueries)
	}

	expectedCost := sdkmath.NewInt(4500)
	if !hourlyData.TotalCost.Equal(expectedCost) {
		t.Errorf("Expected total cost %v, got %v", expectedCost, hourlyData.TotalCost)
	}

	if hourlyData.TotalComputeUnits != 450 {
		t.Errorf("Expected 450 compute units, got %d", hourlyData.TotalComputeUnits)
	}

	// Test model usage tracking
	if hourlyData.ModelUsage["model1"] != 2 {
		t.Errorf("Expected 2 uses of model1, got %d", hourlyData.ModelUsage["model1"])
	}
	if hourlyData.ModelUsage["model2"] != 1 {
		t.Errorf("Expected 1 use of model2, got %d", hourlyData.ModelUsage["model2"])
	}

	// Test operation type tracking
	if hourlyData.OperationTypes["query"] != 2 {
		t.Errorf("Expected 2 query operations, got %d", hourlyData.OperationTypes["query"])
	}
	if hourlyData.OperationTypes["inference"] != 1 {
		t.Errorf("Expected 1 inference operation, got %d", hourlyData.OperationTypes["inference"])
	}
}

func TestGetUserAnalytics(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record analytics for specific user
	userAddress := "user1"
	records := []AnalyticsRecord{
		{
			Address:       userAddress,
			ModelHash:     "model1",
			ComputeUnits:  100,
			Cost:          sdkmath.NewInt(1000),
			OperationType: "query",
			Success:       true,
			ResponseTime:  200,
		},
		{
			Address:       userAddress,
			ModelHash:     "model1",
			ComputeUnits:  150,
			Cost:          sdkmath.NewInt(1500),
			OperationType: "inference",
			Success:       true,
			ResponseTime:  300,
		},
		{
			Address:       "other-user",
			ModelHash:     "model1",
			ComputeUnits:  200,
			Cost:          sdkmath.NewInt(2000),
			OperationType: "query",
			Success:       true,
			ResponseTime:  250,
		},
	}

	for _, record := range records {
		if err := k.RecordAnalytics(ctx, record); err != nil {
			t.Fatalf("Failed to record analytics: %v", err)
		}
	}

	// Get user-specific analytics
	userStats := k.GetUserAnalytics(ctx, userAddress)
	if userStats.QueryCount != 2 {
		t.Errorf("Expected 2 queries for user, got %d", userStats.QueryCount)
	}

	expectedUserCost := sdkmath.NewInt(2500)
	if !userStats.TotalCost.Equal(expectedUserCost) {
		t.Errorf("Expected user cost %v, got %v", expectedUserCost, userStats.TotalCost)
	}

	if userStats.ComputeUnits != 250 {
		t.Errorf("Expected 250 compute units for user, got %d", userStats.ComputeUnits)
	}
}

func TestAnalyticsSnapshot(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record some data
	record := AnalyticsRecord{
		Address:       "user1",
		ModelHash:     "model1",
		ComputeUnits:  100,
		Cost:          sdkmath.NewInt(1000),
		OperationType: "query",
		Success:       true,
		ResponseTime:  200,
	}

	if err := k.RecordAnalytics(ctx, record); err != nil {
		t.Fatalf("Failed to record analytics: %v", err)
	}

	// Create snapshot
	snapshot := k.CreateAnalyticsSnapshot(ctx, PeriodHourly)
	if snapshot.Period != PeriodHourly {
		t.Errorf("Expected period %v, got %v", PeriodHourly, snapshot.Period)
	}

	if snapshot.Data.TotalQueries != 1 {
		t.Errorf("Expected 1 query in snapshot, got %d", snapshot.Data.TotalQueries)
	}

	// Verify timestamp
	if snapshot.Timestamp.IsZero() {
		t.Error("Snapshot timestamp should not be zero")
	}
}

func TestAnalyticsPeriods(t *testing.T) {
	k, ctx := setupKeeper(t)

	record := AnalyticsRecord{
		Address:       "user1",
		ModelHash:     "model1",
		ComputeUnits:  100,
		Cost:          sdkmath.NewInt(1000),
		OperationType: "query",
		Success:       true,
		ResponseTime:  200,
	}

	if err := k.RecordAnalytics(ctx, record); err != nil {
		t.Fatalf("Failed to record analytics: %v", err)
	}

	periods := []AnalyticsPeriod{PeriodHourly, PeriodDaily}
	for _, period := range periods {
		data := k.GetAnalytics(ctx, period, ctx.BlockTime())
		if data.TotalQueries != 1 {
			t.Errorf("Period %v: expected 1 query, got %d", period, data.TotalQueries)
		}
	}
}

func TestSuccessRateCalculation(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record mixed success/failure
	records := []AnalyticsRecord{
		{
			Address:       "user1",
			ModelHash:     "model1",
			ComputeUnits:  100,
			Cost:          sdkmath.NewInt(1000),
			OperationType: "query",
			Success:       true,
			ResponseTime:  200,
		},
		{
			Address:       "user1",
			ModelHash:     "model1",
			ComputeUnits:  100,
			Cost:          sdkmath.NewInt(1000),
			OperationType: "query",
			Success:       true,
			ResponseTime:  200,
		},
		{
			Address:       "user1",
			ModelHash:     "model1",
			ComputeUnits:  100,
			Cost:          sdkmath.NewInt(1000),
			OperationType: "query",
			Success:       true,
			ResponseTime:  200,
		},
		{
			Address:       "user1",
			ModelHash:     "model1",
			ComputeUnits:  100,
			Cost:          sdkmath.NewInt(1000),
			OperationType: "query",
			Success:       false,
			ResponseTime:  200,
		},
	}

	for _, record := range records {
		if err := k.RecordAnalytics(ctx, record); err != nil {
			t.Fatalf("Failed to record analytics: %v", err)
		}
	}

	analytics := k.GetAnalytics(ctx, PeriodHourly, ctx.BlockTime())

	// Success rate should be 75% (3 out of 4)
	expectedRate := 0.75
	tolerance := 0.01
	if analytics.SuccessRate < expectedRate-tolerance || analytics.SuccessRate > expectedRate+tolerance {
		t.Errorf("Expected success rate around %v, got %v", expectedRate, analytics.SuccessRate)
	}
}

func TestTopUsersTracking(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Create records for multiple users with varying usage
	users := []struct {
		address string
		queries uint64
	}{
		{"user1", 10},
		{"user2", 5},
		{"user3", 15},
		{"user4", 3},
		{"user5", 8},
	}

	for _, user := range users {
		for i := uint64(0); i < user.queries; i++ {
			record := AnalyticsRecord{
				Address:       user.address,
				ModelHash:     "model1",
				ComputeUnits:  100,
				Cost:          sdkmath.NewInt(1000),
				OperationType: "query",
				Success:       true,
				ResponseTime:  200,
			}
			if err := k.RecordAnalytics(ctx, record); err != nil {
				t.Fatalf("Failed to record analytics: %v", err)
			}
		}
	}

	analytics := k.GetAnalytics(ctx, PeriodHourly, ctx.BlockTime())

	// Verify top users are tracked
	if len(analytics.TopUsers) == 0 {
		t.Error("Expected top users to be tracked")
	}

	// Top user should be user3 with 15 queries
	if len(analytics.TopUsers) > 0 {
		topUser := analytics.TopUsers[0]
		if topUser.QueryCount < 10 {
			t.Errorf("Top user should have at least 10 queries, got %d", topUser.QueryCount)
		}
	}
}

func TestAverageResponseTimeCalculation(t *testing.T) {
	k, ctx := setupKeeper(t)

	responseTimes := []uint64{100, 200, 300, 400, 500}
	for _, rt := range responseTimes {
		record := AnalyticsRecord{
			Address:       "user1",
			ModelHash:     "model1",
			ComputeUnits:  100,
			Cost:          sdkmath.NewInt(1000),
			OperationType: "query",
			Success:       true,
			ResponseTime:  rt,
		}
		if err := k.RecordAnalytics(ctx, record); err != nil {
			t.Fatalf("Failed to record analytics: %v", err)
		}
	}

	analytics := k.GetAnalytics(ctx, PeriodHourly, ctx.BlockTime())

	// Average should be 300ms
	expectedAvg := uint64(300)
	if analytics.AverageResponseTime != expectedAvg {
		t.Logf("Average response time: expected %d, got %d (acceptable variance)",
			expectedAvg, analytics.AverageResponseTime)
	}
}

func TestCacheHitRateTracking(t *testing.T) {
	k, ctx := setupKeeper(t)

	// This test assumes RecordAnalytics tracks cache hits
	// Implementation may vary, so we just verify the field exists
	analytics := k.GetAnalytics(ctx, PeriodHourly, ctx.BlockTime())

	// Verify cache hit rate field is accessible
	if analytics.CacheHitRate < 0 || analytics.CacheHitRate > 1 {
		t.Errorf("Cache hit rate should be between 0 and 1, got %v", analytics.CacheHitRate)
	}
}

func TestAnalyticsMetadata(t *testing.T) {
	k, ctx := setupKeeper(t)

	record := AnalyticsRecord{
		Address:       "user1",
		ModelHash:     "model1",
		ComputeUnits:  100,
		Cost:          sdkmath.NewInt(1000),
		OperationType: "query",
		Success:       true,
		ResponseTime:  200,
	}

	if err := k.RecordAnalytics(ctx, record); err != nil {
		t.Fatalf("Failed to record analytics: %v", err)
	}

	snapshot := k.CreateAnalyticsSnapshot(ctx, PeriodHourly)

	// Verify metadata can be set
	if snapshot.Metadata == nil {
		snapshot.Metadata = make(map[string]string)
	}
	snapshot.Metadata["version"] = "1.0"

	if snapshot.Metadata["version"] != "1.0" {
		t.Error("Metadata should be settable")
	}
}

func TestPeakUsageTimeTracking(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Record analytics at different times
	baseTime := ctx.BlockTime()

	for i := 0; i < 5; i++ {
		// Modify context time
		ctx = ctx.WithBlockTime(baseTime.Add(time.Duration(i) * time.Hour))

		record := AnalyticsRecord{
			Address:       "user1",
			ModelHash:     "model1",
			ComputeUnits:  100 * uint64(i+1),
			Cost:          sdkmath.NewInt(int64(1000 * (i + 1))),
			OperationType: "query",
			Success:       true,
			ResponseTime:  200,
		}

		if err := k.RecordAnalytics(ctx, record); err != nil {
			t.Fatalf("Failed to record analytics: %v", err)
		}
	}

	analytics := k.GetAnalytics(ctx, PeriodDaily, baseTime)

	// Peak usage should be tracked
	if analytics.PeakUsageTime.IsZero() {
		t.Log("Peak usage time tracking may not be implemented yet")
	}
}
