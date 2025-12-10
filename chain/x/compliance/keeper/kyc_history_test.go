package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestKYCVersionTracking verifies that version numbers are properly incremented
// on each KYC record update
func TestKYCVersionTracking(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	address := "aura1test"
	provider := "kyc-provider-1"

	// First submission - should get version 1
	record1 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      provider,
		VerifiedAt: ctx.BlockTime(),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	err := keeper.UpdateKYCRecord(ctx, record1, "initial_submission")
	require.NoError(t, err)
	require.Equal(t, uint64(1), record1.Version, "First submission should have version 1")

	// Verify stored record has correct version
	stored, err := keeper.GetKYCRecord(ctx, address)
	require.NoError(t, err)
	require.Equal(t, uint64(1), stored.Version)

	// Second submission - should get version 2
	record2 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      provider,
		VerifiedAt:    timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(366 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	err = keeper.UpdateKYCRecord(ctx, record2, "level_upgrade")
	require.NoError(t, err)
	require.Equal(t, uint64(2), record2.Version, "Second submission should have version 2")

	// Verify stored record has correct version
	stored, err = keeper.GetKYCRecord(ctx, address)
	require.NoError(t, err)
	require.Equal(t, uint64(2), stored.Version)

	// Third submission - should get version 3
	record3 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
		Provider:      provider,
		VerifiedAt: timestamppb.New(ctx.BlockTime().Add(2 * time.Hour)),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(367 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	err = keeper.UpdateKYCRecord(ctx, record3, "level_upgrade")
	require.NoError(t, err)
	require.Equal(t, uint64(3), record3.Version, "Third submission should have version 3")
}

// TestKYCHistoryPreservation verifies that previous versions are archived
// when records are updated
func TestKYCHistoryPreservation(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	address := "aura1test"
	provider := "kyc-provider-1"

	// First submission
	record1 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      provider,
		VerifiedAt: ctx.BlockTime(),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
		PiiCommitment: []byte("commitment1"),
		Jurisdiction:  "US",
	}

	err := keeper.UpdateKYCRecord(ctx, record1, "initial_submission")
	require.NoError(t, err)

	// Verify no history exists yet (first submission has no previous version)
	history, err := keeper.GetKYCHistory(ctx, address)
	require.NoError(t, err)
	require.Empty(t, history, "No history should exist after first submission")

	// Second submission - should archive version 1
	record2 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      provider,
		VerifiedAt: timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(366 * 24 * time.Hour)),
		PiiCommitment: []byte("commitment2"),
		Jurisdiction:  "US",
	}

	err = keeper.UpdateKYCRecord(ctx, record2, "level_upgrade")
	require.NoError(t, err)

	// Verify version 1 is in history
	history, err = keeper.GetKYCHistory(ctx, address)
	require.NoError(t, err)
	require.Len(t, history, 1, "History should contain 1 entry after second submission")
	require.Equal(t, uint64(1), history[0].Version)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_BASIC, history[0].Snapshot.KycLevel)
	require.Equal(t, "level_upgrade", history[0].UpdateReason)

	// Third submission - should archive version 2
	record3 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
		Provider:      provider,
		VerifiedAt: timestamppb.New(ctx.BlockTime().Add(2 * time.Hour)),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(367 * 24 * time.Hour)),
		PiiCommitment: []byte("commitment3"),
		Jurisdiction:  "US",
	}

	err = keeper.UpdateKYCRecord(ctx, record3, "edd_required")
	require.NoError(t, err)

	// Verify both version 1 and 2 are in history
	history, err = keeper.GetKYCHistory(ctx, address)
	require.NoError(t, err)
	require.Len(t, history, 2, "History should contain 2 entries after third submission")

	// Verify chronological order
	require.Equal(t, uint64(1), history[0].Version)
	require.Equal(t, uint64(2), history[1].Version)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_BASIC, history[0].Snapshot.KycLevel)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_INTERMEDIATE, history[1].Snapshot.KycLevel)
}

// TestKYCDuplicateDetection verifies that duplicate submissions are handled
// correctly with no data loss
func TestKYCDuplicateDetection(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	address := "aura1test"
	provider := "kyc-provider-1"

	// Initial submission
	record1 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      provider,
		VerifiedAt: ctx.BlockTime(),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
		PiiCommitment: []byte("commitment1"),
		Jurisdiction:  "US",
	}

	err := keeper.UpdateKYCRecord(ctx, record1, "initial_submission")
	require.NoError(t, err)

	// Duplicate submission (same address)
	record2 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      provider,
		VerifiedAt: timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(366 * 24 * time.Hour)),
		PiiCommitment: []byte("commitment2"),
		Jurisdiction:  "US",
	}

	err = keeper.UpdateKYCRecord(ctx, record2, "duplicate_submission")
	require.NoError(t, err)

	// Verify current record is the latest
	current, err := keeper.GetKYCRecord(ctx, address)
	require.NoError(t, err)
	require.Equal(t, uint64(2), current.Version)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_INTERMEDIATE, current.KycLevel)

	// Verify original record is preserved in history
	history, err := keeper.GetKYCHistory(ctx, address)
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Equal(t, types.KYCLevel_KYC_LEVEL_BASIC, history[0].Snapshot.KycLevel)
}

// TestKYCHistoryMetadata verifies that history entries contain correct metadata
func TestKYCHistoryMetadata(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	address := "aura1test"
	provider1 := "kyc-provider-1"
	provider2 := "kyc-provider-2"

	// Initial submission by provider 1
	record1 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      provider1,
		VerifiedAt: ctx.BlockTime(),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	err := keeper.UpdateKYCRecord(ctx, record1, "initial_submission")
	require.NoError(t, err)

	// Update by provider 2
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(24 * time.Hour))
	record2 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      provider2,
		VerifiedAt: ctx.BlockTime(),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	err = keeper.UpdateKYCRecord(ctx, record2, "provider_change")
	require.NoError(t, err)

	// Verify history metadata
	history, err := keeper.GetKYCHistory(ctx, address)
	require.NoError(t, err)
	require.Len(t, history, 1)

	entry := history[0]
	require.Equal(t, address, entry.Address)
	require.Equal(t, uint64(1), entry.Version)
	require.Equal(t, provider2, entry.UpdatedBy, "UpdatedBy should be the provider performing the update")
	require.Equal(t, "provider_change", entry.UpdateReason)
	require.NotNil(t, entry.UpdatedAt)
	require.NotNil(t, entry.Snapshot)
	require.Equal(t, provider1, entry.Snapshot.Provider, "Snapshot should preserve original provider")
}

// TestKYCVersionEvents verifies that version change events are emitted
func TestKYCVersionEvents(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	address := "aura1test"
	provider := "kyc-provider-1"

	// Create event manager to capture events
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	// Initial submission
	record1 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		Provider:      provider,
		VerifiedAt: ctx.BlockTime(),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	err := keeper.UpdateKYCRecord(ctx, record1, "initial_submission")
	require.NoError(t, err)

	// Verify event was emitted with version
	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)

	foundVersionEvent := false
	for _, event := range events {
		if event.Type == types.EventTypeKYCSubmitted {
			foundVersionEvent = true
			for _, attr := range event.Attributes {
				if attr.Key == "version" {
					require.Equal(t, "1", attr.Value)
				}
				if attr.Key == "update_reason" {
					require.Equal(t, "initial_submission", attr.Value)
				}
			}
		}
	}
	require.True(t, foundVersionEvent, "Version event should be emitted")

	// Update and verify version 2 event
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	record2 := &types.KYCRecord{
		Address:       address,
		KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		Provider:      provider,
		VerifiedAt: timestamppb.New(ctx.BlockTime().Add(time.Hour)),
		ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(366 * 24 * time.Hour)),
		PiiCommitment: make([]byte, 32),
		Jurisdiction:  "US",
	}

	err = keeper.UpdateKYCRecord(ctx, record2, "level_upgrade")
	require.NoError(t, err)

	events = ctx.EventManager().Events()
	foundVersionEvent = false
	for _, event := range events {
		if event.Type == types.EventTypeKYCSubmitted {
			foundVersionEvent = true
			for _, attr := range event.Attributes {
				if attr.Key == "version" {
					require.Equal(t, "2", attr.Value)
				}
			}
		}
	}
	require.True(t, foundVersionEvent, "Version 2 event should be emitted")
}

// TestGetKYCHistoryEmpty verifies that GetKYCHistory returns empty list
// for addresses with no history
func TestGetKYCHistoryEmpty(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	address := "aura1nonexistent"

	history, err := keeper.GetKYCHistory(ctx, address)
	require.NoError(t, err, "GetKYCHistory should not error for non-existent address")
	require.Empty(t, history, "History should be empty for address with no records")
}

// TestGetAllKYCHistory verifies that all KYC history can be retrieved
// for genesis export
func TestGetAllKYCHistory(t *testing.T) {
	suite := NewTestSuite(t)
	ctx := suite.Ctx
	keeper := suite.Keeper

	address1 := "aura1test1"
	address2 := "aura1test2"
	provider := "kyc-provider-1"

	// Create history for address 1
	for i := 0; i < 3; i++ {
		record := &types.KYCRecord{
			Address:       address1,
			KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
			Provider:      provider,
			VerifiedAt: timestamppb.New(ctx.BlockTime().Add(time.Duration(i) * time.Hour)),
			ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
			PiiCommitment: make([]byte, 32),
			Jurisdiction:  "US",
		}
		err := keeper.UpdateKYCRecord(ctx, record, "update")
		require.NoError(t, err)
	}

	// Create history for address 2
	for i := 0; i < 2; i++ {
		record := &types.KYCRecord{
			Address:       address2,
			KycLevel:      types.KYCLevel_KYC_LEVEL_INTERMEDIATE,
			Provider:      provider,
			VerifiedAt: timestamppb.New(ctx.BlockTime().Add(time.Duration(i) * time.Hour)),
			ExpiresAt:     timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
			PiiCommitment: make([]byte, 32),
			Jurisdiction:  "GB",
		}
		err := keeper.UpdateKYCRecord(ctx, record, "update")
		require.NoError(t, err)
	}

	// Get all history
	allHistory, err := keeper.GetAllKYCHistory(ctx)
	require.NoError(t, err)
	require.Len(t, allHistory, 2, "Should have history for 2 addresses")

	// Verify address 1 history (3 updates = 2 history entries)
	require.Len(t, allHistory[address1], 2)
	require.Equal(t, uint64(1), allHistory[address1][0].Version)
	require.Equal(t, uint64(2), allHistory[address1][1].Version)

	// Verify address 2 history (2 updates = 1 history entry)
	require.Len(t, allHistory[address2], 1)
	require.Equal(t, uint64(1), allHistory[address2][0].Version)
}
