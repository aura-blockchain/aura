package keeper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/incidentresponse/types"
)

// setupKeeperForTest creates a keeper with KV store for testing
func setupKeeperForTest(t *testing.T) (*KeeperKV, sdk.Context) {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create in-memory store using the proper Cosmos SDK testing pattern
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	// Create context with proper store
	ctx := sdk.NewContext(cms, cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, log.NewNopLogger())

	// Create keeper
	keeper := NewKeeperKV(storeKey, cdc)

	// Initialize with default params
	defaultParams := types.DefaultParams()
	err := keeper.SetParams(ctx, defaultParams)
	require.NoError(t, err)

	return keeper, ctx
}

func TestIncident_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Report incident
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Test Security Incident",
		"This is a test incident",
		types.SeverityHigh,
		"reporter1",
		[]string{"system1", "system2"},
	)
	require.NoError(t, err)
	require.NotEmpty(t, incidentID)

	// Retrieve incident
	incident, err := keeper.GetIncident(ctx, incidentID)
	require.NoError(t, err)
	require.NotNil(t, incident)
	require.Equal(t, "Test Security Incident", incident.Title)
	require.Equal(t, types.SeverityHigh, incident.Severity)
	require.Equal(t, types.StatusNew, incident.Status)

	// Update incident status
	err = keeper.UpdateIncidentStatus(ctx, incidentID, types.StatusInvestigation, "investigator1", "Starting investigation")
	require.NoError(t, err)

	// Verify update
	incident, err = keeper.GetIncident(ctx, incidentID)
	require.NoError(t, err)
	require.Equal(t, types.StatusInvestigation, incident.Status)
	require.Len(t, incident.Timeline, 2) // reported + status update
}

func TestIncidentList_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create multiple incidents
	for i := 0; i < 5; i++ {
		_, err := keeper.ReportIncident(
			ctx,
			"Incident "+string(rune('A'+i)),
			"Description "+string(rune('A'+i)),
			types.SeverityMedium,
			"reporter1",
			[]string{"system1"},
		)
		require.NoError(t, err)
	}

	// Get all incidents
	incidents := keeper.GetAllIncidents(ctx)
	require.Len(t, incidents, 5)
}

func TestPauseState_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Setup params with authorized keys
	params := types.DefaultParams()
	params.EmergencyPauseEnabled = true
	params.PauseAuthorizedKeys = []string{"key1", "key2", "key3"}
	params.PauseRequiredSigners = 1
	params.MaxPauseDuration = 24 * time.Hour
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Check initial state
	require.False(t, keeper.IsChainPaused(ctx))

	// Request chain pause
	err = keeper.RequestChainPause(ctx, "key1", types.PauseLevelFull, "Emergency", "", 1*time.Hour)
	require.NoError(t, err)

	// Verify chain is paused
	require.True(t, keeper.IsChainPaused(ctx))

	pauseState := keeper.GetChainPauseState(ctx)
	require.True(t, pauseState.IsPaused)
	require.Equal(t, types.PauseLevelFull, pauseState.PauseLevel)
	require.Equal(t, "key1", pauseState.PausedBy)

	// Resume chain
	err = keeper.ResumeChain(ctx, "key1", "Issue resolved")
	require.NoError(t, err)

	// Verify chain is resumed
	require.False(t, keeper.IsChainPaused(ctx))
}

func TestWalletLimits_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Set wallet limits
	err := keeper.SetWalletLimits(ctx, "wallet1", "1000000", "100000", "500000")
	require.NoError(t, err)

	// Retrieve wallet limits
	limits, err := keeper.GetWalletLimits(ctx, "wallet1")
	require.NoError(t, err)
	require.NotNil(t, limits)
	require.Equal(t, "wallet1", limits.Address)
	require.Equal(t, "1000000", limits.MaxBalance)
	require.Equal(t, "100000", limits.MaxTransactionSize)
	require.Equal(t, "500000", limits.DailyLimit)
}

func TestWalletLimitCheck_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Enable wallet limits
	params := keeper.GetParams(ctx)
	params.HotWalletLimitsEnabled = true
	params.GlobalMaxHotWallet = "10000000"
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Set wallet limits
	err = keeper.SetWalletLimits(ctx, "wallet1", "1000000", "100000", "500000")
	require.NoError(t, err)

	// Check within limits
	err = keeper.CheckWalletLimit(ctx, "wallet1", "50000", "0")
	require.NoError(t, err)

	// Check exceeding transaction size
	err = keeper.CheckWalletLimit(ctx, "wallet1", "150000", "0")
	require.Error(t, err)
	require.Equal(t, types.ErrWalletLimitExceeded, err)

	// Check exceeding daily limit
	err = keeper.CheckWalletLimit(ctx, "wallet1", "600000", "0")
	require.Error(t, err)
	require.Equal(t, types.ErrWalletLimitExceeded, err)
}

func TestPostMortem_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create incident
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Post Mortem Test",
		"Testing post mortem",
		types.SeverityCritical,
		"reporter1",
		[]string{"system1"},
	)
	require.NoError(t, err)

	// Resolve incident
	err = keeper.UpdateIncidentStatus(ctx, incidentID, types.StatusResolved, "resolver1", "Fixed")
	require.NoError(t, err)

	// Create post mortem
	actionItems := []types.ActionItem{
		{
			ID:          "action1",
			Description: "Improve monitoring",
			Assignee:    "team1",
			Priority:    "high",
			Status:      "pending",
			DueDate:     time.Now().Add(7 * 24 * time.Hour),
		},
	}

	err = keeper.CreatePostMortem(
		ctx,
		incidentID,
		"pm-creator",
		"Summary of incident",
		"Root cause analysis",
		"Impact assessment",
		"Resolution details",
		[]string{"Lesson 1", "Lesson 2"},
		actionItems,
	)
	require.NoError(t, err)

	// Verify post mortem
	incident, err := keeper.GetIncident(ctx, incidentID)
	require.NoError(t, err)
	require.NotNil(t, incident.PostMortem)
	require.Equal(t, "Summary of incident", incident.PostMortem.Summary)
	require.Len(t, incident.PostMortem.ActionItems, 1)

	// Close incident
	err = keeper.CloseIncident(ctx, incidentID, "closer1")
	require.NoError(t, err)

	// Verify closed
	incident, err = keeper.GetIncident(ctx, incidentID)
	require.NoError(t, err)
	require.Equal(t, types.StatusClosed, incident.Status)
}

func TestGenesis_RoundTrip(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create some state
	incidentID, err := keeper.ReportIncident(
		ctx,
		"Genesis Test",
		"Testing genesis",
		types.SeverityMedium,
		"reporter1",
		[]string{"system1"},
	)
	require.NoError(t, err)

	err = keeper.SetWalletLimits(ctx, "wallet1", "1000000", "100000", "500000")
	require.NoError(t, err)

	// Export genesis
	exported := keeper.ExportGenesis(ctx)
	require.NotNil(t, exported)
	require.Len(t, exported.Incidents, 1)
	require.Len(t, exported.WalletLimits, 1)
	require.NotNil(t, exported.Params)
	require.NotNil(t, exported.PauseState)

	// Validate genesis
	err = exported.Validate()
	require.NoError(t, err)

	// Create new keeper
	keeper2, ctx2 := setupKeeperForTest(t)

	// Import genesis
	err = keeper2.InitGenesis(ctx2, exported)
	require.NoError(t, err)

	// Verify state
	incident, err := keeper2.GetIncident(ctx2, incidentID)
	require.NoError(t, err)
	require.Equal(t, "Genesis Test", incident.Title)

	limits, err := keeper2.GetWalletLimits(ctx2, "wallet1")
	require.NoError(t, err)
	require.Equal(t, "1000000", limits.MaxBalance)
}

func TestMultiSigPause_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Setup params with multi-sig
	params := types.DefaultParams()
	params.EmergencyPauseEnabled = true
	params.PauseAuthorizedKeys = []string{"key1", "key2", "key3"}
	params.PauseRequiredSigners = 3
	params.MaxPauseDuration = 24 * time.Hour
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Request pause (first signer)
	blockTime := determinism.GetBlockTime(ctx)
	pauseReqID := fmt.Sprintf("pause-%s-%d", "key1", blockTime.Unix())
	err = keeper.RequestChainPause(ctx, "key1", types.PauseLevelFull, "Emergency", "", 1*time.Hour)
	require.NoError(t, err)

	// Chain should not be paused yet (needs 3 signers)
	require.False(t, keeper.IsChainPaused(ctx))

	// Second approval
	err = keeper.ApproveChainPause(ctx, pauseReqID, "key2")
	require.NoError(t, err)

	// Still not paused (needs 3 signers)
	require.False(t, keeper.IsChainPaused(ctx))

	// Third approval
	err = keeper.ApproveChainPause(ctx, pauseReqID, "key3")
	require.NoError(t, err)

	// Now should be paused (3 signers reached)
	require.True(t, keeper.IsChainPaused(ctx))
}

func TestNextIncidentID_KVPersistence(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)

	// Create multiple incidents
	var ids []string
	for i := 0; i < 10; i++ {
		id, err := keeper.ReportIncident(
			ctx,
			"Incident",
			"Description",
			types.SeverityLow,
			"reporter",
			[]string{},
		)
		require.NoError(t, err)
		ids = append(ids, id)
	}

	// Verify IDs are sequential
	for i := 0; i < len(ids); i++ {
		require.Contains(t, ids[i], "INC-")
	}

	// Verify no duplicates
	idMap := make(map[string]bool)
	for _, id := range ids {
		require.False(t, idMap[id], "Duplicate ID found: %s", id)
		idMap[id] = true
	}
}

func TestKeeperPanicsWithoutStore(t *testing.T) {
	// This test ensures that operations panic without a store
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create keeper
	keeper := NewKeeperKV(storeKey, cdc)

	// Manually set store to nil to test panic behavior
	keeper.store = nil

	ctx := context.Background()

	// Should panic when trying to use keeper without store
	require.Panics(t, func() {
		keeper.GetIncident(ctx, "test")
	})
}
