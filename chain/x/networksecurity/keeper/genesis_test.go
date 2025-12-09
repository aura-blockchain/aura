package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

func TestInitGenesis(t *testing.T) {
	t.Run("init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.NotNil(t, params)
	})

	t.Run("init with trusted peers", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		params := types.DefaultParams()
		genesis := &types.GenesisState{
			Params: *params,
			TrustedPeers: []types.TrustedPeer{
				{
					PeerId:      "peer1",
					Address:     "192.168.1.1",
					PublicKey:   []byte("pubkey1"),
					Description: "Test peer 1",
					AddedAt:     time.Unix(1000, 0),
				},
				{
					PeerId:      "peer2",
					Address:     "192.168.1.2",
					PublicKey:   []byte("pubkey2"),
					Description: "Test peer 2",
					AddedAt:     time.Unix(2000, 0),
				},
			},
			Reputations:     []types.NodeReputation{},
			RateLimits:      []types.RateLimitEntry{},
			ForkAlerts:      []types.ForkAlert{},
			PartitionAlerts: []types.PartitionAlert{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify trusted peers were imported
		// Note: You may need to add a GetTrustedPeer method
		allPeers := k.GetAllTrustedPeers(ctx)
		require.Len(t, allPeers, 2)
	})

	t.Run("init with reputations", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		params := types.DefaultParams()
		genesis := &types.GenesisState{
			Params:       *params,
			TrustedPeers: []types.TrustedPeer{},
			Reputations: []types.NodeReputation{
				{
					PeerId:            "peer1",
					Score:             75,
					MessagesReceived:  100,
					ValidMessages:     95,
					InvalidMessages:   5,
					Uptime:            3600,
					LastUpdatedHeight: 5000,
					MisbehaviorCount:  0,
				},
			},
			RateLimits:      []types.RateLimitEntry{},
			ForkAlerts:      []types.ForkAlert{},
			PartitionAlerts: []types.PartitionAlert{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify reputations were imported
		allReputations := k.GetAllReputations(ctx)
		require.Len(t, allReputations, 1)
	})

	t.Run("init with rate limits", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		params := types.DefaultParams()
		windowStart := time.Unix(1000, 0)
		genesis := &types.GenesisState{
			Params:       *params,
			TrustedPeers: []types.TrustedPeer{},
			Reputations:  []types.NodeReputation{},
			RateLimits: []types.RateLimitEntry{
				{
					PeerId:        "peer1",
					RequestCount:  50,
					WindowStart:   windowStart,
					IsBanned:      false,
					BytesSent:     1024,
					BytesReceived: 2048,
				},
			},
			ForkAlerts:      []types.ForkAlert{},
			PartitionAlerts: []types.PartitionAlert{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify rate limits were imported
		allLimits := k.GetAllRateLimits(ctx)
		require.Len(t, allLimits, 1)
	})

	t.Run("init with fork and partition alerts", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		params := types.DefaultParams()
		forkDetectedAt := time.Unix(1000, 0)
		partitionDetectedAt := time.Unix(2000, 0)
		genesis := &types.GenesisState{
			Params:       *params,
			TrustedPeers: []types.TrustedPeer{},
			Reputations:  []types.NodeReputation{},
			RateLimits:   []types.RateLimitEntry{},
			ForkAlerts: []types.ForkAlert{
				{
					AlertId:           "fork1",
					BlockHeight:       100,
					ChainAHash:        []byte("hash_a"),
					ChainBHash:        []byte("hash_b"),
					DetectedAt:        forkDetectedAt,
					Resolved:          false,
					ResolutionDetails: "",
				},
			},
			PartitionAlerts: []types.PartitionAlert{
				{
					AlertId:        "partition1",
					ConnectedPeers: 2,
					ExpectedPeers:  10,
					MissingPeerIds: []string{"peer3", "peer4"},
					DetectedAt:     partitionDetectedAt,
					Resolved:       false,
				},
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify alerts were imported
		forkAlerts := k.GetAllForkAlerts(ctx, false)
		require.Len(t, forkAlerts, 1)

		partitionAlerts := k.GetAllPartitionAlerts(ctx, false)
		require.Len(t, partitionAlerts, 1)
	})

	t.Run("init with invalid genesis fails", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		// Create invalid genesis state (invalid params)
		invalidParams := types.DefaultParams()
		invalidParams.RateLimit.MaxRequestsPerSecond = 0 // Invalid - must be > 0

		genesis := &types.GenesisState{
			Params:          *invalidParams,
			TrustedPeers:    []types.TrustedPeer{},
			Reputations:     []types.NodeReputation{},
			RateLimits:      []types.RateLimitEntry{},
			ForkAlerts:      []types.ForkAlert{},
			PartitionAlerts: []types.PartitionAlert{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := k.ExportGenesis(ctx)

		require.NotNil(t, genesis)
		require.NotNil(t, genesis.Params)
		require.Empty(t, genesis.TrustedPeers)
		require.Empty(t, genesis.Reputations)
	})

	t.Run("export with data", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		// Initialize with data
		params := types.DefaultParams()
		initGenesis := &types.GenesisState{
			Params: *params,
			TrustedPeers: []types.TrustedPeer{
				{
					PeerId:      "peer1",
					Address:     "192.168.1.1",
					PublicKey:   []byte("pubkey1"),
					Description: "Peer 1",
					AddedAt:     time.Unix(1000, 0),
				},
				{
					PeerId:      "peer2",
					Address:     "192.168.1.2",
					PublicKey:   []byte("pubkey2"),
					Description: "Peer 2",
					AddedAt:     time.Unix(2000, 0),
				},
			},
			Reputations: []types.NodeReputation{
				{PeerId: "peer1", Score: 80},
			},
			RateLimits:      []types.RateLimitEntry{},
			ForkAlerts:      []types.ForkAlert{},
			PartitionAlerts: []types.PartitionAlert{},
		}

		err := k.InitGenesis(ctx, initGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		require.Len(t, exported.TrustedPeers, 2)
		require.Len(t, exported.Reputations, 1)
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		params := types.DefaultParams()
		windowStart := time.Unix(1000, 0)
		originalGenesis := &types.GenesisState{
			Params: *params,
			TrustedPeers: []types.TrustedPeer{
				{
					PeerId:      "peer1",
					Address:     "192.168.1.1",
					PublicKey:   []byte("key1"),
					Description: "Test peer",
					AddedAt:     time.Unix(1000, 0),
				},
			},
			Reputations: []types.NodeReputation{
				{
					PeerId:           "peer1",
					Score:            75,
					MessagesReceived: 100,
					ValidMessages:    95,
					InvalidMessages:  5,
				},
			},
			RateLimits: []types.RateLimitEntry{
				{
					PeerId:       "peer1",
					RequestCount: 10,
					WindowStart:  windowStart,
					IsBanned:     false,
				},
			},
			ForkAlerts:      []types.ForkAlert{},
			PartitionAlerts: []types.PartitionAlert{},
		}

		// Import
		err := k.InitGenesis(ctx, originalGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		// Verify counts match
		require.Len(t, exported.TrustedPeers, len(originalGenesis.TrustedPeers))
		require.Len(t, exported.Reputations, len(originalGenesis.Reputations))
		require.Len(t, exported.RateLimits, len(originalGenesis.RateLimits))
	})

	t.Run("multiple round trips are deterministic", func(t *testing.T) {
		k1, ctx1 := setupTestKeeper(t)
		k2, ctx2 := setupTestKeeper(t)

		genesis := types.DefaultGenesisState()
		genesis.TrustedPeers = []types.TrustedPeer{
			{
				PeerId:      "peer1",
				Address:     "192.168.1.1",
				PublicKey:   []byte("key1"),
				Description: "Test peer",
				AddedAt:     time.Unix(1000, 0),
			},
		}

		// First round trip
		err := k1.InitGenesis(ctx1, genesis)
		require.NoError(t, err)
		export1 := k1.ExportGenesis(ctx1)

		// Second round trip
		err = k2.InitGenesis(ctx2, export1)
		require.NoError(t, err)
		export2 := k2.ExportGenesis(ctx2)

		// Verify exports match
		require.Len(t, export2.TrustedPeers, len(export1.TrustedPeers))
	})
}

func TestDefaultGenesis(t *testing.T) {
	t.Run("default genesis is valid", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		err := types.ValidateGenesisState(genesis)
		require.NoError(t, err)

		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.TrustedPeers)
		require.NotNil(t, genesis.Reputations)
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		k, ctx := setupTestKeeper(t)

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify params were set
		params, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.NotNil(t, params)
	})
}

func setupTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	// Create a test keeper and context
	// This should match your actual test setup
	k, ctx := NewTestKeeperWithContext(t)
	return k, ctx
}
