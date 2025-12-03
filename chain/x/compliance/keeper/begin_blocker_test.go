package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// ============================================================================
// BeginBlocker Tests
// ============================================================================

func TestBeginBlocker_NoExpiredRecords(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create valid, non-expired KYC records
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

	// Run BeginBlocker
	keeper.BeginBlocker(ctx)

	// Check that no events were emitted for KYC expiry
	events := ctx.EventManager().Events()
	kycExpiredEvents := 0
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			kycExpiredEvents++
		}
	}

	require.Equal(t, 0, kycExpiredEvents, "should not emit any KYC expired events")
}

func TestBeginBlocker_SingleExpiredRecord(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create expired KYC record
	expiredRecord := &types.KYCRecord{
		Address:    "aura1expired",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-400 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(now.Add(-30 * 24 * time.Hour)), // Expired 30 days ago
		Jurisdiction: "US",
	}
	err := keeper.SetKYCRecord(ctx, expiredRecord)
	require.NoError(t, err)

	// Create valid KYC record
	validRecord := &types.KYCRecord{
		Address:    "aura1valid",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now),
		ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
		Jurisdiction: "GB",
	}
	err = keeper.SetKYCRecord(ctx, validRecord)
	require.NoError(t, err)

	// Run BeginBlocker
	keeper.BeginBlocker(ctx)

	// Check that exactly one KYC expired event was emitted
	events := ctx.EventManager().Events()
	kycExpiredEvents := []sdk.Event{}
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			kycExpiredEvents = append(kycExpiredEvents, event)
		}
	}

	require.Len(t, kycExpiredEvents, 1, "should emit exactly one KYC expired event")

	// Verify event attributes
	event := kycExpiredEvents[0]
	attributes := make(map[string]string)
	for _, attr := range event.Attributes {
		attributes[attr.Key] = attr.Value
	}

	require.Equal(t, "aura1expired", attributes[types.AttributeKeyAddress], "event should contain expired address")
	require.Equal(t, "KYC_LEVEL_BASIC", attributes["kyc_level"], "event should contain KYC level")
	require.Equal(t, "test_provider", attributes["provider"], "event should contain provider")
	require.Equal(t, "US", attributes["jurisdiction"], "event should contain jurisdiction")
}

func TestBeginBlocker_MultipleExpiredRecords(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create multiple expired KYC records
	expiredAddresses := []string{"aura1expired1", "aura1expired2", "aura1expired3"}
	for _, addr := range expiredAddresses {
		record := &types.KYCRecord{
			Address:    addr,
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test_provider",
			VerifiedAt: timestamppb.New(now.Add(-400 * 24 * time.Hour)),
			ExpiresAt:  timestamppb.New(now.Add(-30 * 24 * time.Hour)),
			Jurisdiction: "US",
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Create some valid records
	validAddresses := []string{"aura1valid1", "aura1valid2"}
	for _, addr := range validAddresses {
		record := &types.KYCRecord{
			Address:    addr,
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test_provider",
			VerifiedAt: timestamppb.New(now),
			ExpiresAt:  timestamppb.New(now.Add(365 * 24 * time.Hour)),
			Jurisdiction: "GB",
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Run BeginBlocker
	keeper.BeginBlocker(ctx)

	// Check that exactly 3 KYC expired events were emitted
	events := ctx.EventManager().Events()
	kycExpiredEvents := []sdk.Event{}
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			kycExpiredEvents = append(kycExpiredEvents, event)
		}
	}

	require.Len(t, kycExpiredEvents, 3, "should emit exactly 3 KYC expired events")

	// Verify that all expired addresses have events
	emittedAddresses := make(map[string]bool)
	for _, event := range kycExpiredEvents {
		for _, attr := range event.Attributes {
			if attr.Key == types.AttributeKeyAddress {
				emittedAddresses[attr.Value] = true
			}
		}
	}

	for _, addr := range expiredAddresses {
		require.True(t, emittedAddresses[addr], "should emit event for expired address %s", addr)
	}
}

func TestBeginBlocker_JustExpired(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create KYC record that expires exactly 1 second before now
	justExpired := &types.KYCRecord{
		Address:    "aura1justexpired",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-365 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(now.Add(-1 * time.Second)), // Just expired
		Jurisdiction: "US",
	}
	err := keeper.SetKYCRecord(ctx, justExpired)
	require.NoError(t, err)

	// Run BeginBlocker
	keeper.BeginBlocker(ctx)

	// Check that event was emitted
	events := ctx.EventManager().Events()
	kycExpiredEvents := 0
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			kycExpiredEvents++
		}
	}

	require.Equal(t, 1, kycExpiredEvents, "should emit event for record that just expired")
}

func TestBeginBlocker_ExactlyAtExpiry(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create KYC record that expires exactly now (not yet expired)
	atExpiry := &types.KYCRecord{
		Address:    "aura1atexpiry",
		KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:   "test_provider",
		VerifiedAt: timestamppb.New(now.Add(-365 * 24 * time.Hour)),
		ExpiresAt:  timestamppb.New(now), // Expires exactly now
		Jurisdiction: "US",
	}
	err := keeper.SetKYCRecord(ctx, atExpiry)
	require.NoError(t, err)

	// Run BeginBlocker
	keeper.BeginBlocker(ctx)

	// Check that no event was emitted (equal is not after)
	events := ctx.EventManager().Events()
	kycExpiredEvents := 0
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			kycExpiredEvents++
		}
	}

	require.Equal(t, 0, kycExpiredEvents, "should not emit event for record at exact expiry time")
}

func TestBeginBlocker_MixedExpiryTimes(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()

	// Create records with various expiry states
	records := []struct {
		address  string
		expiresAt time.Time
		shouldExpire bool
	}{
		{"aura1expired1", now.Add(-100 * 24 * time.Hour), true},  // Expired 100 days ago
		{"aura1expired2", now.Add(-1 * time.Second), true},       // Just expired
		{"aura1valid1", now.Add(1 * time.Second), false},         // Expires in 1 second
		{"aura1valid2", now.Add(30 * 24 * time.Hour), false},     // Expires in 30 days
		{"aura1valid3", now.Add(365 * 24 * time.Hour), false},    // Expires in 1 year
		{"aura1expired3", now.Add(-365 * 24 * time.Hour), true},  // Expired 1 year ago
	}

	for _, rec := range records {
		record := &types.KYCRecord{
			Address:    rec.address,
			KycLevel:   types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:   "test_provider",
			VerifiedAt: timestamppb.New(now.Add(-400 * 24 * time.Hour)),
			ExpiresAt:  timestamppb.New(rec.expiresAt),
			Jurisdiction: "US",
		}
		err := keeper.SetKYCRecord(ctx, record)
		require.NoError(t, err)
	}

	// Run BeginBlocker
	keeper.BeginBlocker(ctx)

	// Count expired events
	events := ctx.EventManager().Events()
	kycExpiredEvents := []sdk.Event{}
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			kycExpiredEvents = append(kycExpiredEvents, event)
		}
	}

	// Count expected expired records
	expectedExpired := 0
	for _, rec := range records {
		if rec.shouldExpire {
			expectedExpired++
		}
	}

	require.Len(t, kycExpiredEvents, expectedExpired, "should emit events for all expired records")

	// Verify that only expired addresses have events
	emittedAddresses := make(map[string]bool)
	for _, event := range kycExpiredEvents {
		for _, attr := range event.Attributes {
			if attr.Key == types.AttributeKeyAddress {
				emittedAddresses[attr.Value] = true
			}
		}
	}

	for _, rec := range records {
		if rec.shouldExpire {
			require.True(t, emittedAddresses[rec.address], "should emit event for expired address %s", rec.address)
		} else {
			require.False(t, emittedAddresses[rec.address], "should not emit event for valid address %s", rec.address)
		}
	}
}

func TestBeginBlocker_EmptyStore(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Run BeginBlocker on empty store - should not panic
	require.NotPanics(t, func() {
		keeper.BeginBlocker(ctx)
	}, "BeginBlocker should handle empty store gracefully")

	// Verify no events emitted
	events := ctx.EventManager().Events()
	kycExpiredEvents := 0
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			kycExpiredEvents++
		}
	}

	require.Equal(t, 0, kycExpiredEvents, "should not emit any events on empty store")
}

func TestBeginBlocker_EventAttributes(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	now := ctx.BlockTime()
	verifiedAt := now.Add(-400 * 24 * time.Hour)
	expiresAt := now.Add(-30 * 24 * time.Hour)

	// Create expired KYC record with specific attributes
	record := &types.KYCRecord{
		Address:    "aura1test",
		KycLevel:   types.KYCLevel_KYC_LEVEL_ADVANCED,
		Provider:   "kyc_provider_1",
		VerifiedAt: timestamppb.New(verifiedAt),
		ExpiresAt:  timestamppb.New(expiresAt),
		Jurisdiction: "US",
	}
	err := keeper.SetKYCRecord(ctx, record)
	require.NoError(t, err)

	// Run BeginBlocker
	keeper.BeginBlocker(ctx)

	// Find the KYC expired event
	events := ctx.EventManager().Events()
	var kycEvent *sdk.Event
	for _, event := range events {
		if event.Type == types.EventTypeKYCExpired {
			e := event
			kycEvent = &e
			break
		}
	}

	require.NotNil(t, kycEvent, "should emit KYC expired event")

	// Extract attributes
	attributes := make(map[string]string)
	for _, attr := range kycEvent.Attributes {
		attributes[attr.Key] = attr.Value
	}

	// Verify all required attributes are present
	require.Equal(t, "aura1test", attributes[types.AttributeKeyAddress])
	require.Equal(t, "KYC_LEVEL_ADVANCED", attributes["kyc_level"])
	require.Equal(t, "kyc_provider_1", attributes["provider"])
	require.Equal(t, "US", attributes["jurisdiction"])
	require.NotEmpty(t, attributes["expired_at"])
	require.NotEmpty(t, attributes["verified_at"])
	require.NotEmpty(t, attributes[types.AttributeKeyBlockHeight])
	require.NotEmpty(t, attributes[types.AttributeKeyBlockTime])
	require.NotEmpty(t, attributes[types.AttributeKeyTimestamp])
}
