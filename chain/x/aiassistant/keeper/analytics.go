package keeper

import (
	"encoding/json"
	"time"

	"cosmossdk.io/store/prefix"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// UsageAnalytics represents usage analytics data
type UsageAnalytics struct {
	Period              AnalyticsPeriod
	TotalQueries        uint64
	UniqueUsers         uint64
	TotalCost           sdkmath.Int
	AverageCost         sdkmath.Int
	TotalComputeUnits   uint64
	AverageComputeUnits uint64
	ModelUsage          map[string]uint64
	OperationTypes      map[string]uint64
	PeakUsageTime       time.Time
	CacheHitRate        float64
	SuccessRate         float64
	AverageResponseTime uint64
	TopUsers            []UserUsageStats
}

// AnalyticsPeriod defines time period for analytics
type AnalyticsPeriod string

const (
	PeriodHourly  AnalyticsPeriod = "hourly"
	PeriodDaily   AnalyticsPeriod = "daily"
	PeriodWeekly  AnalyticsPeriod = "weekly"
	PeriodMonthly AnalyticsPeriod = "monthly"
)

// UserUsageStats represents usage statistics for a user
type UserUsageStats struct {
	Address          string
	QueryCount       uint64
	TotalCost        sdkmath.Int
	ComputeUnits     uint64
	AverageQueryTime uint64
	LastActive       time.Time
}

// AnalyticsSnapshot represents a snapshot of analytics data
type AnalyticsSnapshot struct {
	Timestamp  time.Time
	Period     AnalyticsPeriod
	Data       UsageAnalytics
	Metadata   map[string]string
}

// RecordAnalytics records analytics data for a query
func (k Keeper) RecordAnalytics(ctx sdk.Context, record AnalyticsRecord) error {
	store := ctx.KVStore(k.storeKey)

	// Update hourly analytics
	hourKey := types.AnalyticsKey(string(PeriodHourly), ctx.BlockTime())
	hourlyData := k.getOrCreateAnalytics(ctx, PeriodHourly, ctx.BlockTime())
	k.updateAnalytics(&hourlyData, record)
	bz, err := json.Marshal(&hourlyData)
	if err != nil {
		return err
	}
	store.Set(hourKey, bz)

	// Update daily analytics
	dayKey := types.AnalyticsKey(string(PeriodDaily), ctx.BlockTime())
	dailyData := k.getOrCreateAnalytics(ctx, PeriodDaily, ctx.BlockTime())
	k.updateAnalytics(&dailyData, record)
	bz, err = json.Marshal(&dailyData)
	if err != nil {
		return err
	}
	store.Set(dayKey, bz)

	return nil
}

// updateAnalytics updates analytics data with new record
func (k Keeper) updateAnalytics(data *UsageAnalytics, record AnalyticsRecord) {
	data.TotalQueries++
	data.TotalCost = data.TotalCost.Add(record.Cost)
	data.TotalComputeUnits += record.ComputeUnits

	if data.ModelUsage == nil {
		data.ModelUsage = make(map[string]uint64)
	}
	data.ModelUsage[record.ModelHash]++

	if data.OperationTypes == nil {
		data.OperationTypes = make(map[string]uint64)
	}
	data.OperationTypes[record.OperationType]++

	// Update averages
	if data.TotalQueries > 0 {
		data.AverageCost = data.TotalCost.Quo(sdkmath.NewInt(int64(data.TotalQueries)))
		data.AverageComputeUnits = data.TotalComputeUnits / data.TotalQueries
	}
}

// GetAnalytics retrieves analytics for a specific period
func (k Keeper) GetAnalytics(ctx sdk.Context, period AnalyticsPeriod, timestamp time.Time) (UsageAnalytics, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.AnalyticsKey(string(period), timestamp)

	bz := store.Get(key)
	if len(bz) == 0 {
		return UsageAnalytics{}, false
	}

	var analytics UsageAnalytics
	if err := json.Unmarshal(bz, &analytics); err != nil {
		return UsageAnalytics{}, false
	}

	return analytics, true
}

// getOrCreateAnalytics gets or creates analytics for period
func (k Keeper) getOrCreateAnalytics(ctx sdk.Context, period AnalyticsPeriod, timestamp time.Time) UsageAnalytics {
	analytics, exists := k.GetAnalytics(ctx, period, timestamp)
	if !exists {
		analytics = UsageAnalytics{
			Period:         period,
			TotalCost:      sdkmath.ZeroInt(),
			AverageCost:    sdkmath.ZeroInt(),
			ModelUsage:     make(map[string]uint64),
			OperationTypes: make(map[string]uint64),
		}
	}
	return analytics
}

// GetAnalyticsSummary returns summary analytics
func (k Keeper) GetAnalyticsSummary(ctx sdk.Context, period AnalyticsPeriod, duration time.Duration) AnalyticsSummary {
	summary := AnalyticsSummary{
		Period:         period,
		StartTime:      ctx.BlockTime().Add(-duration),
		EndTime:        ctx.BlockTime(),
		TotalQueries:   0,
		TotalCost:      sdkmath.ZeroInt(),
		UniqueUsers:    0,
		ModelBreakdown: make(map[string]uint64),
	}

	// Iterate through analytics in range
	currentTime := summary.StartTime
	for currentTime.Before(summary.EndTime) {
		analytics, exists := k.GetAnalytics(ctx, period, currentTime)
		if exists {
			summary.TotalQueries += analytics.TotalQueries
			summary.TotalCost = summary.TotalCost.Add(analytics.TotalCost)

			// Merge model usage
			for model, count := range analytics.ModelUsage {
				summary.ModelBreakdown[model] += count
			}
		}

		// Move to next period
		currentTime = k.getNextPeriod(currentTime, period)
	}

	return summary
}

// getNextPeriod calculates next period timestamp
func (k Keeper) getNextPeriod(t time.Time, period AnalyticsPeriod) time.Time {
	switch period {
	case PeriodHourly:
		return t.Add(time.Hour)
	case PeriodDaily:
		return t.AddDate(0, 0, 1)
	case PeriodWeekly:
		return t.AddDate(0, 0, 7)
	case PeriodMonthly:
		return t.AddDate(0, 1, 0)
	default:
		return t.Add(time.Hour)
	}
}

// GetTopUsers returns top users by query count
func (k Keeper) GetTopUsers(ctx sdk.Context, limit uint64) []UserUsageStats {
	usageRecords := k.ListQueryUsage(ctx)

	// Simple sorting - in production use more efficient approach
	for i := 0; i < len(usageRecords)-1; i++ {
		for j := i + 1; j < len(usageRecords); j++ {
			if usageRecords[i].TotalQueries < usageRecords[j].TotalQueries {
				usageRecords[i], usageRecords[j] = usageRecords[j], usageRecords[i]
			}
		}
	}

	// Convert to UserUsageStats
	var topUsers []UserUsageStats
	for i := 0; i < len(usageRecords) && uint64(i) < limit; i++ {
		topUsers = append(topUsers, UserUsageStats{
			Address:    usageRecords[i].Address,
			QueryCount: usageRecords[i].TotalQueries,
			LastActive: usageRecords[i].LastReset,
		})
	}

	return topUsers
}

// GetRevenueAnalytics returns revenue analytics
func (k Keeper) GetRevenueAnalytics(ctx sdk.Context, period AnalyticsPeriod) RevenueAnalytics {
	analytics := RevenueAnalytics{
		Period:            period,
		TotalRevenue:      sdkmath.ZeroInt(),
		RevenueByModel:    make(map[string]sdkmath.Int),
		RevenueByOperation: make(map[string]sdkmath.Int),
	}

	// Would aggregate from audit logs and cost records
	// Placeholder implementation

	return analytics
}

// CreateAnalyticsSnapshot creates a snapshot of current analytics
func (k Keeper) CreateAnalyticsSnapshot(ctx sdk.Context, period AnalyticsPeriod) error {
	analytics := k.getOrCreateAnalytics(ctx, period, ctx.BlockTime())

	snapshot := AnalyticsSnapshot{
		Timestamp: ctx.BlockTime(),
		Period:    period,
		Data:      analytics,
		Metadata:  make(map[string]string),
	}

	store := ctx.KVStore(k.storeKey)
	key := types.AnalyticsSnapshotKey(string(period), ctx.BlockTime())

	bz, err := json.Marshal(&snapshot)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

// ListAnalyticsSnapshots returns all snapshots for a period
func (k Keeper) ListAnalyticsSnapshots(ctx sdk.Context, period AnalyticsPeriod) []AnalyticsSnapshot {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AnalyticsSnapshotKeyPrefix(string(period)))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var snapshots []AnalyticsSnapshot
	for ; iterator.Valid(); iterator.Next() {
		var snapshot AnalyticsSnapshot
		if err := json.Unmarshal(iterator.Value(), &snapshot); err != nil {
			continue
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

// AnalyticsRecord represents a single analytics record
type AnalyticsRecord struct {
	UserAddress   string
	ModelHash     string
	OperationType string
	Cost          sdkmath.Int
	ComputeUnits  uint64
	ResponseTime  uint64
	Success       bool
	CacheHit      bool
}

// AnalyticsSummary represents summarized analytics
type AnalyticsSummary struct {
	Period         AnalyticsPeriod
	StartTime      time.Time
	EndTime        time.Time
	TotalQueries   uint64
	TotalCost      sdkmath.Int
	UniqueUsers    uint64
	ModelBreakdown map[string]uint64
}

// RevenueAnalytics represents revenue analytics
type RevenueAnalytics struct {
	Period             AnalyticsPeriod
	TotalRevenue       sdkmath.Int
	RevenueByModel     map[string]sdkmath.Int
	RevenueByOperation map[string]sdkmath.Int
	AverageRevenuePerUser sdkmath.Int
}
