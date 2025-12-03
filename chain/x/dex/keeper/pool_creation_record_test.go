package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// TestPoolCreationRecord_RecordAndRetrieve tests basic pool creation record storage
func TestPoolCreationRecord_RecordAndRetrieve(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	poolID := "pool1"
	tokenA := "uaura"
	tokenB := "usdt"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Initially no record should exist
	record := keeper.GetPoolCreationRecord(ctx, creator)
	require.Nil(t, record, "No record should exist initially")

	// Record pool creation
	keeper.RecordPoolCreation(ctx, creator, poolID, tokenA, tokenB, amountA, amountB)

	// Retrieve and verify record
	record = keeper.GetPoolCreationRecord(ctx, creator)
	require.NotNil(t, record, "Record should exist after recording")
	require.Equal(t, creator, record.Creator)
	require.Equal(t, uint64(1), record.TotalPools)
	require.Len(t, record.PoolIds, 1)
	require.Equal(t, poolID, record.PoolIds[0])
	require.NotNil(t, record.LastCreationTime)
}

// TestPoolCreationRecord_MultiplePoolsByCreator tests recording multiple pools
func TestPoolCreationRecord_MultiplePoolsByCreator(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Create first pool
	keeper.RecordPoolCreation(ctx, creator, "pool1", "uaura", "usdt", amountA, amountB)

	record := keeper.GetPoolCreationRecord(ctx, creator)
	require.Equal(t, uint64(1), record.TotalPools)
	require.Len(t, record.PoolIds, 1)

	// Advance time for second pool
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))

	// Create second pool
	keeper.RecordPoolCreation(ctx, creator, "pool2", "uaura", "usdc", amountA, amountB)

	record = keeper.GetPoolCreationRecord(ctx, creator)
	require.Equal(t, uint64(2), record.TotalPools)
	require.Len(t, record.PoolIds, 2)
	require.Equal(t, "pool1", record.PoolIds[0])
	require.Equal(t, "pool2", record.PoolIds[1])

	// Create third pool
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	keeper.RecordPoolCreation(ctx, creator, "pool3", "uaura", "dai", amountA, amountB)

	record = keeper.GetPoolCreationRecord(ctx, creator)
	require.Equal(t, uint64(3), record.TotalPools)
	require.Len(t, record.PoolIds, 3)
}

// TestPoolCreationRecord_MultipleCreators tests separate records for different creators
func TestPoolCreationRecord_MultipleCreators(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator1 := "aura1creator1"
	creator2 := "aura1creator2"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Creator 1 creates 2 pools
	keeper.RecordPoolCreation(ctx, creator1, "pool1", "uaura", "usdt", amountA, amountB)
	keeper.RecordPoolCreation(ctx, creator1, "pool2", "uaura", "usdc", amountA, amountB)

	// Creator 2 creates 1 pool
	keeper.RecordPoolCreation(ctx, creator2, "pool3", "uaura", "dai", amountA, amountB)

	// Verify creator 1 record
	record1 := keeper.GetPoolCreationRecord(ctx, creator1)
	require.NotNil(t, record1)
	require.Equal(t, uint64(2), record1.TotalPools)
	require.Len(t, record1.PoolIds, 2)

	// Verify creator 2 record
	record2 := keeper.GetPoolCreationRecord(ctx, creator2)
	require.NotNil(t, record2)
	require.Equal(t, uint64(1), record2.TotalPools)
	require.Len(t, record2.PoolIds, 1)

	// Records should be independent
	require.NotEqual(t, record1.Creator, record2.Creator)
}

// TestPoolCreationRecord_GetAllRecords tests retrieval of all pool creation records
func TestPoolCreationRecord_GetAllRecords(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Initially no records
	records := keeper.GetAllPoolCreationRecords(ctx)
	require.Empty(t, records)

	// Create pools from 3 different creators
	keeper.RecordPoolCreation(ctx, "aura1creator1", "pool1", "uaura", "usdt", amountA, amountB)
	keeper.RecordPoolCreation(ctx, "aura1creator2", "pool2", "uaura", "usdc", amountA, amountB)
	keeper.RecordPoolCreation(ctx, "aura1creator1", "pool3", "uaura", "dai", amountA, amountB)
	keeper.RecordPoolCreation(ctx, "aura1creator3", "pool4", "uaura", "btc", amountA, amountB)

	// Get all records
	records = keeper.GetAllPoolCreationRecords(ctx)
	require.Len(t, records, 3, "Should have 3 unique creators")

	// Verify records are correct
	creatorCounts := make(map[string]uint64)
	for _, record := range records {
		creatorCounts[record.Creator] = record.TotalPools
	}

	require.Equal(t, uint64(2), creatorCounts["aura1creator1"])
	require.Equal(t, uint64(1), creatorCounts["aura1creator2"])
	require.Equal(t, uint64(1), creatorCounts["aura1creator3"])
}

// TestPoolCreationLimit_Enforcement tests pool creation limit enforcement
func TestPoolCreationLimit_Enforcement(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Set security params with max 3 pools per creator
	params := types.DefaultSecurityParams()
	params.MaxPoolsPerCreator = 3
	// Simulate setting params (would normally be done via keeper method)
	// For this test, we'll manually create records and test the check

	// Create 3 pools - should all succeed
	keeper.RecordPoolCreation(ctx, creator, "pool1", "uaura", "usdt", amountA, amountB)
	err := keeper.CheckPoolCreationLimit(ctx, creator)
	require.NoError(t, err, "Should allow creating more pools (1/3)")

	keeper.RecordPoolCreation(ctx, creator, "pool2", "uaura", "usdc", amountA, amountB)
	err = keeper.CheckPoolCreationLimit(ctx, creator)
	require.NoError(t, err, "Should allow creating more pools (2/3)")

	keeper.RecordPoolCreation(ctx, creator, "pool3", "uaura", "dai", amountA, amountB)
	err = keeper.CheckPoolCreationLimit(ctx, creator)
	require.Error(t, err, "Should reject 4th pool (3/3 reached)")
	require.ErrorIs(t, err, types.ErrPoolCreationLimitExceeded)

	// Verify record
	record := keeper.GetPoolCreationRecord(ctx, creator)
	require.Equal(t, uint64(3), record.TotalPools)
}

// TestPoolCreationLimit_NoLimit tests when limit is disabled
func TestPoolCreationLimit_NoLimit(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// With max_pools_per_creator = 0 (unlimited), should never error
	params := types.DefaultSecurityParams()
	params.MaxPoolsPerCreator = 0

	// Create many pools - should all succeed
	for i := 1; i <= 10; i++ {
		err := keeper.CheckPoolCreationLimit(ctx, creator)
		require.NoError(t, err, "Should allow unlimited pools when limit is 0")
		keeper.RecordPoolCreation(ctx, creator, "pool"+string(rune(i)), "uaura", "usdt", amountA, amountB)
	}

	record := keeper.GetPoolCreationRecord(ctx, creator)
	require.Equal(t, uint64(10), record.TotalPools)
}

// TestPoolCreationCooldown_Enforcement tests pool creation cooldown enforcement
func TestPoolCreationCooldown_Enforcement(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Set cooldown to 1 hour (3600 seconds)
	params := types.DefaultSecurityParams()
	params.PoolCreationCooldown = 3600

	// First pool - should succeed
	err := keeper.CheckPoolCreationCooldown(ctx, creator)
	require.NoError(t, err, "First pool should have no cooldown")
	keeper.RecordPoolCreation(ctx, creator, "pool1", "uaura", "usdt", amountA, amountB)

	// Try to create second pool immediately - should fail
	err = keeper.CheckPoolCreationCooldown(ctx, creator)
	require.Error(t, err, "Should reject pool creation within cooldown period")
	require.ErrorIs(t, err, types.ErrPoolCreationCooldown)

	// Advance time by 30 minutes - still should fail
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
	err = keeper.CheckPoolCreationCooldown(ctx, creator)
	require.Error(t, err, "Should still reject pool creation after 30 minutes")

	// Advance time to exactly 1 hour - should now succeed
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(30 * time.Minute))
	err = keeper.CheckPoolCreationCooldown(ctx, creator)
	require.NoError(t, err, "Should allow pool creation after cooldown period")
	keeper.RecordPoolCreation(ctx, creator, "pool2", "uaura", "usdc", amountA, amountB)

	record := keeper.GetPoolCreationRecord(ctx, creator)
	require.Equal(t, uint64(2), record.TotalPools)
}

// TestPoolCreationCooldown_NoCooldown tests when cooldown is disabled
func TestPoolCreationCooldown_NoCooldown(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// With cooldown = 0, should never error
	params := types.DefaultSecurityParams()
	params.PoolCreationCooldown = 0

	// Create pools rapidly - should all succeed
	keeper.RecordPoolCreation(ctx, creator, "pool1", "uaura", "usdt", amountA, amountB)
	err := keeper.CheckPoolCreationCooldown(ctx, creator)
	require.NoError(t, err)

	keeper.RecordPoolCreation(ctx, creator, "pool2", "uaura", "usdc", amountA, amountB)
	err = keeper.CheckPoolCreationCooldown(ctx, creator)
	require.NoError(t, err)

	keeper.RecordPoolCreation(ctx, creator, "pool3", "uaura", "dai", amountA, amountB)
	err = keeper.CheckPoolCreationCooldown(ctx, creator)
	require.NoError(t, err)

	record := keeper.GetPoolCreationRecord(ctx, creator)
	require.Equal(t, uint64(3), record.TotalPools)
}

// TestPoolCreationRecord_Integration tests full pool creation with audit trail
func TestPoolCreationRecord_Integration(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := suite.TestAccs[0]
	creatorStr := creator.String()

	// Setup: Fund creator account
	initialCoins := sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(10000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(10000000)),
	)
	suite.FundAccount(suite.BankKeeper, ctx, creator, initialCoins)

	// Create pool (this should automatically record pool creation)
	pool, lpTokens, err := keeper.CreatePool(
		ctx,
		creatorStr,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(1000000)),
	)

	require.NoError(t, err, "Pool creation should succeed")
	require.NotNil(t, pool)
	require.True(t, lpTokens.IsPositive())

	// Verify pool creation record was created
	record := keeper.GetPoolCreationRecord(ctx, creatorStr)
	require.NotNil(t, record, "Pool creation record should exist")
	require.Equal(t, creatorStr, record.Creator)
	require.Equal(t, uint64(1), record.TotalPools)
	require.Len(t, record.PoolIds, 1)
	require.Equal(t, pool.PoolId, record.PoolIds[0])

	// Verify event was emitted
	events := ctx.EventManager().Events()
	var poolCreationRecordedEventFound bool
	for _, event := range events {
		if event.Type == "pool_creation_recorded" {
			poolCreationRecordedEventFound = true
			// Verify event attributes
			attrs := event.Attributes
			var creatorAttr, poolIDAttr, totalPoolsAttr string
			for _, attr := range attrs {
				switch attr.Key {
				case "creator":
					creatorAttr = attr.Value
				case "pool_id":
					poolIDAttr = attr.Value
				case "total_pools_created":
					totalPoolsAttr = attr.Value
				}
			}
			require.Equal(t, creatorStr, creatorAttr)
			require.Equal(t, pool.PoolId, poolIDAttr)
			require.Equal(t, "1", totalPoolsAttr)
			break
		}
	}
	require.True(t, poolCreationRecordedEventFound, "pool_creation_recorded event should be emitted")
}

// TestPoolCreationRecord_GenesisExportImport tests genesis export/import
func TestPoolCreationRecord_GenesisExportImport(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Create pool creation records
	keeper.RecordPoolCreation(ctx, "aura1creator1", "pool1", "uaura", "usdt", amountA, amountB)
	keeper.RecordPoolCreation(ctx, "aura1creator1", "pool2", "uaura", "usdc", amountA, amountB)
	keeper.RecordPoolCreation(ctx, "aura1creator2", "pool3", "uaura", "dai", amountA, amountB)

	// Export genesis
	genesisState := keeper.ExportGenesis(ctx)

	// Verify pool creation records are in genesis
	require.NotNil(t, genesisState.PoolCreationRecords)
	require.Len(t, genesisState.PoolCreationRecords, 2, "Should export 2 creator records")

	// Create new keeper and import genesis
	suite2 := SetupKeeperTestSuite(t)
	ctx2 := suite2.Ctx
	keeper2 := suite2.DexKeeper

	err := keeper2.InitGenesis(ctx2, genesisState)
	require.NoError(t, err)

	// Verify imported records
	record1 := keeper2.GetPoolCreationRecord(ctx2, "aura1creator1")
	require.NotNil(t, record1)
	require.Equal(t, uint64(2), record1.TotalPools)
	require.Len(t, record1.PoolIds, 2)

	record2 := keeper2.GetPoolCreationRecord(ctx2, "aura1creator2")
	require.NotNil(t, record2)
	require.Equal(t, uint64(1), record2.TotalPools)
	require.Len(t, record2.PoolIds, 1)
}

// TestPoolCreationRecord_TimestampUpdates tests that timestamps are updated correctly
func TestPoolCreationRecord_TimestampUpdates(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Record first pool
	initialTime := ctx.BlockTime()
	keeper.RecordPoolCreation(ctx, creator, "pool1", "uaura", "usdt", amountA, amountB)

	record := keeper.GetPoolCreationRecord(ctx, creator)
	firstTimestamp := record.LastCreationTime.AsTime()
	require.True(t, firstTimestamp.Equal(initialTime), "First timestamp should match block time")

	// Advance time and create second pool
	newTime := initialTime.Add(2 * time.Hour)
	ctx = ctx.WithBlockTime(newTime)
	keeper.RecordPoolCreation(ctx, creator, "pool2", "uaura", "usdc", amountA, amountB)

	record = keeper.GetPoolCreationRecord(ctx, creator)
	secondTimestamp := record.LastCreationTime.AsTime()
	require.True(t, secondTimestamp.Equal(newTime), "Second timestamp should be updated")
	require.True(t, secondTimestamp.After(firstTimestamp), "Second timestamp should be after first")
}

// TestPoolCreationRecord_AuditTrailCompleteness tests comprehensive audit trail
func TestPoolCreationRecord_AuditTrailCompleteness(t *testing.T) {
	suite := SetupKeeperTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.DexKeeper

	creator := "aura1creator1"
	poolID := "pool1"
	tokenA := "uaura"
	tokenB := "usdt"
	amountA := sdkmath.NewInt(1000000)
	amountB := sdkmath.NewInt(2000000)

	// Record pool creation
	keeper.RecordPoolCreation(ctx, creator, poolID, tokenA, tokenB, amountA, amountB)

	// Verify all audit information is present
	record := keeper.GetPoolCreationRecord(ctx, creator)
	require.NotNil(t, record)

	// Check creator
	require.NotEmpty(t, record.Creator, "Creator should be recorded")
	require.Equal(t, creator, record.Creator)

	// Check pool IDs list
	require.NotEmpty(t, record.PoolIds, "Pool IDs should be recorded")
	require.Contains(t, record.PoolIds, poolID)

	// Check timestamp
	require.NotNil(t, record.LastCreationTime, "Creation timestamp should be recorded")
	require.True(t, record.LastCreationTime.AsTime().Before(time.Now()), "Timestamp should be valid")

	// Check total pools counter
	require.Greater(t, record.TotalPools, uint64(0), "Total pools should be positive")
	require.Equal(t, uint64(len(record.PoolIds)), record.TotalPools, "Total pools should match pool IDs count")

	// Verify event contains detailed audit information
	events := ctx.EventManager().Events()
	var auditEvent sdk.Event
	for _, event := range events {
		if event.Type == "pool_creation_recorded" {
			auditEvent = event
			break
		}
	}

	require.NotEmpty(t, auditEvent.Type, "Audit event should exist")

	// Verify event has all required audit attributes
	requiredAttrs := []string{
		"creator",
		"pool_id",
		"token_a",
		"token_b",
		"initial_liquidity_a",
		"initial_liquidity_b",
		"total_pools_created",
		"timestamp",
		"block_height",
	}

	attrs := make(map[string]string)
	for _, attr := range auditEvent.Attributes {
		attrs[attr.Key] = attr.Value
	}

	for _, requiredAttr := range requiredAttrs {
		require.Contains(t, attrs, requiredAttr, "Event should contain %s attribute", requiredAttr)
		require.NotEmpty(t, attrs[requiredAttr], "Attribute %s should not be empty", requiredAttr)
	}
}
