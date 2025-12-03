package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// ============================================================================
// IsKYCExpired Tests
// ============================================================================

func TestIsKYCExpired_NoRecord(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// No KYC record exists - should be treated as expired (fail-safe)
	expired := keeper.IsKYCExpired(ctx, "aura1nonexistent")

	require.True(t, expired, "non-existent KYC should be treated as expired")
}

func TestIsKYCExpired_NotExpired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create KYC record that expires in the future
	now := ctx.BlockTime()
	expiresAt := now.Add(365 * 24 * time.Hour) // 1 year from now

	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(expiresAt),
	}

	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Check expiry - should not be expired
	expired := keeper.IsKYCExpired(ctx, "aura1test")

	require.False(t, expired, "KYC with future expiry should not be expired")
}

func TestIsKYCExpired_Expired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create KYC record that expired in the past
	now := ctx.BlockTime()
	expiresAt := now.Add(-24 * time.Hour) // 1 day ago

	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-365 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(expiresAt),
	}

	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Check expiry - should be expired
	expired := keeper.IsKYCExpired(ctx, "aura1test")

	require.True(t, expired, "KYC with past expiry should be expired")
}

func TestIsKYCExpired_ExactlyAtExpiry(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create KYC record that expires exactly now
	now := ctx.BlockTime()

	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-365 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(now),
	}

	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Check expiry - should not be expired (equal is not after)
	expired := keeper.IsKYCExpired(ctx, "aura1test")

	require.False(t, expired, "KYC at exact expiry time should not be expired yet")
}

// ============================================================================
// ValidateKYCStatus Tests
// ============================================================================

func TestValidateKYCStatus_Success(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set minimum KYC level in params
	params := types.ComplianceParams{
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:   365,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create valid, non-expired KYC record
	now := ctx.BlockTime()
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
	}

	err = keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Validate - should succeed
	err = keeper.ValidateKYCStatus(ctx, "aura1test")

	require.NoError(t, err, "valid KYC should pass validation")
}

func TestValidateKYCStatus_NotFound(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Validate non-existent KYC record
	err := keeper.ValidateKYCStatus(ctx, "aura1nonexistent")

	require.Error(t, err, "validation should fail for non-existent KYC")
	require.ErrorIs(t, err, types.ErrKYCNotFound, "should return ErrKYCNotFound")
}

func TestValidateKYCStatus_Expired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set minimum KYC level in params
	params := types.ComplianceParams{
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:   365,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create expired KYC record
	now := ctx.BlockTime()
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-400 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(now.Add(-30 * 24 * time.Hour)), // Expired 30 days ago
	}

	err = keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Validate - should fail due to expiry
	err = keeper.ValidateKYCStatus(ctx, "aura1test")

	require.Error(t, err, "validation should fail for expired KYC")
	require.ErrorIs(t, err, types.ErrKYCExpired, "should return ErrKYCExpired")
	require.Contains(t, err.Error(), "KYC expired", "error message should mention expiry")
}

func TestValidateKYCStatus_InsufficientLevel(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set high minimum KYC level in params
	params := types.ComplianceParams{
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_ADVANCED, // Require ADVANCED
		KycExpiryDays:   365,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create KYC record with BASIC level (insufficient)
	now := ctx.BlockTime()
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC, // Only BASIC
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
	}

	err = keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Validate - should fail due to insufficient level
	err = keeper.ValidateKYCStatus(ctx, "aura1test")

	require.Error(t, err, "validation should fail for insufficient KYC level")
	require.ErrorIs(t, err, types.ErrInsufficientKYCLevel, "should return ErrInsufficientKYCLevel")
	require.Contains(t, err.Error(), "below minimum", "error message should mention level requirement")
}

func TestValidateKYCStatus_AllChecksPass(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set minimum KYC level
	params := types.ComplianceParams{
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		KycExpiryDays:   365,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create valid KYC record with sufficient level
	now := ctx.BlockTime()
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_ADVANCED, // Higher than minimum
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
	}

	err = keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Validate - should succeed (all checks pass)
	err = keeper.ValidateKYCStatus(ctx, "aura1test")

	require.NoError(t, err, "all checks should pass")
}

// ============================================================================
// ValidateKYCForOperation Tests
// ============================================================================

func TestValidateKYCForOperation_Success(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set minimum KYC level
	params := types.ComplianceParams{
		MinimumKycLevel: types.KYCLevel_KYC_LEVEL_BASIC,
		KycExpiryDays:   365,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Create valid KYC record
	now := ctx.BlockTime()
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
	}

	err = keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Validate for specific operation
	err = keeper.ValidateKYCForOperation(ctx, "aura1test", "bridge transfer to Ethereum")

	require.NoError(t, err, "validation should succeed for valid KYC")
}

func TestValidateKYCForOperation_FailureWithContext(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Validate without KYC record - should fail with operation context
	err := keeper.ValidateKYCForOperation(ctx, "aura1test", "DEX swap AURA/USDC")

	require.Error(t, err, "validation should fail")
	require.Contains(t, err.Error(), "DEX swap AURA/USDC", "error should include operation name")
}

// ============================================================================
// IterateKYCRecords Tests
// ============================================================================

func TestIterateKYCRecords_EmptyStore(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Iterate empty store
	count := 0
	keeper.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
		count++
		return false
	})

	require.Equal(t, 0, count, "should not iterate when store is empty")
}

func TestIterateKYCRecords_MultipleRecords(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create multiple KYC records
	addresses := []string{"aura1test1", "aura1test2", "aura1test3"}
	for _, addr := range addresses {
		record := &types.KYCRecord{
			Address:    addr,
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test_provider",
			VerifiedAt: timestamppb.New(now),
			ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Iterate and count records
	count := 0
	keeper.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
		count++
		return false
	})

	require.Equal(t, 3, count, "should iterate all 3 records")
}

func TestIterateKYCRecords_StopEarly(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create multiple KYC records
	for i := 1; i <= 5; i++ {
		record := &types.KYCRecord{
			Address:    fmt.Sprintf("aura1test%d", i),
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test_provider",
			VerifiedAt: timestamppb.New(now),
			ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Iterate and stop after 2 records
	count := 0
	keeper.IterateKYCRecords(ctx, func(record types.KYCRecord) bool {
		count++
		return count >= 2 // Stop after 2
	})

	require.Equal(t, 2, count, "should stop iteration early")
}

// ============================================================================
// GetExpiringKYCRecords Tests
// ============================================================================

func TestGetExpiringKYCRecords_None(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create KYC record expiring far in the future
	now := ctx.BlockTime()
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)), // 1 year away
	}
	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Get records expiring in next 30 days
	expiring := keeper.GetExpiringKYCRecords(ctx, 30*24*time.Hour)

	require.Empty(t, expiring, "should not find records expiring in next 30 days")
}

func TestGetExpiringKYCRecords_SomeExpiring(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create record expiring in 15 days (within window)
	record1 := &types.KYCRecord{
		Address:    "aura1test1",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(15 * 24 * time.Hour)),
	}
	err := keeper.SetKYCRecord(ctx, record1)
	require.NoError(t, err)

	// Create record expiring in 100 days (outside window)
	record2 := &types.KYCRecord{
		Address:    "aura1test2",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(100 * 24 * time.Hour)),
	}
	err = keeper.SetKYCRecord(ctx, record2)
	require.NoError(t, err)

	// Create already expired record (should not be included)
	record3 := &types.KYCRecord{
		Address:    "aura1test3",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-400 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(now.Add(-10 * 24 * time.Hour)),
	}
	err = keeper.SetKYCRecord(ctx, record3)
	require.NoError(t, err)

	// Get records expiring in next 30 days
	expiring := keeper.GetExpiringKYCRecords(ctx, 30*24*time.Hour)

	require.Len(t, expiring, 1, "should find exactly 1 record expiring in window")
	require.Equal(t, "aura1test1", expiring[0].Address, "should return the correct record")
}

// ============================================================================
// GetExpiredKYCRecords Tests
// ============================================================================

func TestGetExpiredKYCRecords_None(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create valid, non-expired KYC record
	now := ctx.BlockTime()
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
	}
	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Get expired records
	expired := keeper.GetExpiredKYCRecords(ctx)

	require.Empty(t, expired, "should not find any expired records")
}

func TestGetExpiredKYCRecords_SomeExpired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create expired record
	record1 := &types.KYCRecord{
		Address:    "aura1test1",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-400 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(now.Add(-30 * 24 * time.Hour)),
	}
	err := keeper.SetKYCRecord(ctx, record1)
	require.NoError(t, err)

	// Create valid record
	record2 := &types.KYCRecord{
		Address:    "aura1test2",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
	}
	err = keeper.SetKYCRecord(ctx, record2)
	require.NoError(t, err)

	// Get expired records
	expired := keeper.GetExpiredKYCRecords(ctx)

	require.Len(t, expired, 1, "should find exactly 1 expired record")
	require.Equal(t, "aura1test1", expired[0].Address, "should return the expired record")
}

func TestGetExpiredKYCRecords_MultipleExpired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create multiple expired records
	expiredAddresses := []string{"aura1test1", "aura1test2", "aura1test3"}
	for _, addr := range expiredAddresses {
		record := &types.KYCRecord{
			Address:    addr,
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test_provider",
			VerifiedAt: timestamppb.New(now.Add(-400 * 24 * time.Hour)),
			ExpiresAt:  timestamppb.New(now.Add(-30 * 24 * time.Hour)),
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Create one valid record
	validRecord := &types.KYCRecord{
		Address:    "aura1valid",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
	}
	err := keeper.SetKYCRecord(ctx, validRecord)
	require.NoError(t, err)

	// Get expired records
	expired := keeper.GetExpiredKYCRecords(ctx)

	require.Len(t, expired, 3, "should find all 3 expired records")

	// Verify all expired addresses are in the result
	foundAddresses := make(map[string]bool)
	for _, rec := range expired {
		foundAddresses[rec.Address] = true
	}

	for _, addr := range expiredAddresses {
		require.True(t, foundAddresses[addr], "should include expired address %s", addr)
	}
	require.False(t, foundAddresses["aura1valid"], "should not include valid address")
}
