package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	pb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	t.Run("init with default genesis", func(t *testing.T) {
		genesis := types.DefaultGenesisState()
		err := keeper.InitGenesis(*genesis)
		require.NoError(t, err)

		// Verify params are set
		params := keeper.paramsStore.GetParams()
		require.Equal(t, genesis.Params.MaxIrPerLocale, params.MaxIrPerLocale)
		require.Equal(t, genesis.Params.DefaultRateLimitHour, params.DefaultRateLimitHour)
	})

	t.Run("init with valid IR definitions", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		genesis := &pb.GenesisState{
			Params: &pb.Params{
				MaxIrPerLocale:       100,
				DefaultRateLimitHour: 10,
				SuspensionFee:        "1000",
				MinGovernanceDeposit: "5000",
			},
			Irs: []*pb.IRDefinition{
				{
					Id:          "IR-001",
					Name:        "Test IR",
					Arena:       pb.Arena_ARENA_BIOMETRIC,
					Description: "Test description",
					Score:       100,
					PoiReward:   50,
					LocaleTags:  []string{"global"},
					PrivacyTier: pb.PrivacyTier_PRIVACY_TIER_HIGH,
					Version:     "1.0",
					Status:      pb.IRStatus_IR_STATUS_ACTIVE,
				},
			},
			Prerequisites: []*pb.IRPrerequisite{},
			RateLimits:    []*pb.IRRateLimit{},
		}

		genesisState := types.GenesisState{
			Params:        genesis.Params,
			Irs:           genesis.Irs,
			Prerequisites: genesis.Prerequisites,
			RateLimits:    genesis.RateLimits,
		}

		err := keeper.InitGenesis(genesisState)
		require.NoError(t, err)

		// Verify IR was imported
		ir, ok := keeper.GetIR("IR-001")
		require.True(t, ok)
		require.Equal(t, "Test IR", ir.Name)
		require.Equal(t, int64(100), ir.Score)
	})

	t.Run("init with prerequisites", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		genesis := types.GenesisState{
			Params: &pb.Params{
				MaxIrPerLocale:       100,
				DefaultRateLimitHour: 10,
				SuspensionFee:        "1000",
				MinGovernanceDeposit: "5000",
			},
			Irs: []*pb.IRDefinition{
				{
					Id:          "IR-BASE",
					Name:        "Base IR",
					Arena:       pb.Arena_ARENA_ANCHOR,
					Description: "Base IR",
					Score:       0,
					LocaleTags:  []string{"global"},
					PrivacyTier: pb.PrivacyTier_PRIVACY_TIER_LOW,
					Version:     "1.0",
					Status:      pb.IRStatus_IR_STATUS_ACTIVE,
				},
				{
					Id:          "IR-DERIVED",
					Name:        "Derived IR",
					Arena:       pb.Arena_ARENA_BIOMETRIC,
					Description: "Derived IR",
					Score:       50,
					LocaleTags:  []string{"global"},
					PrivacyTier: pb.PrivacyTier_PRIVACY_TIER_MEDIUM,
					Version:     "1.0",
					Status:      pb.IRStatus_IR_STATUS_ACTIVE,
				},
			},
			Prerequisites: []*pb.IRPrerequisite{
				{
					IrId:          "IR-DERIVED",
					RequiredIrIds: []string{"IR-BASE"},
				},
			},
			RateLimits: []*pb.IRRateLimit{},
		}

		err := keeper.InitGenesis(genesis)
		require.NoError(t, err)

		// Verify prerequisite was imported
		prereq, exists := keeper.GetPrerequisites("IR-DERIVED")
		require.True(t, exists)
		require.Len(t, prereq.RequiredIrIds, 1)
		require.Equal(t, "IR-BASE", prereq.RequiredIrIds[0])
	})

	t.Run("init with rate limits", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		genesis := types.GenesisState{
			Params: &pb.Params{
				MaxIrPerLocale:       100,
				DefaultRateLimitHour: 10,
				SuspensionFee:        "1000",
				MinGovernanceDeposit: "5000",
			},
			Irs: []*pb.IRDefinition{
				{
					Id:          "IR-LIMITED",
					Name:        "Rate Limited IR",
					Arena:       pb.Arena_ARENA_BIOMETRIC,
					Description: "Limited IR",
					Score:       100,
					LocaleTags:  []string{"global"},
					PrivacyTier: pb.PrivacyTier_PRIVACY_TIER_HIGH,
					Version:     "1.0",
					Status:      pb.IRStatus_IR_STATUS_ACTIVE,
				},
			},
			Prerequisites: []*pb.IRPrerequisite{},
			RateLimits: []*pb.IRRateLimit{
				{
					IrId:             "IR-LIMITED",
					PerWalletPerHour: 5,
					PerWalletPerDay:  20,
					PerBlockGlobal:   100,
				},
			},
		}

		err := keeper.InitGenesis(genesis)
		require.NoError(t, err)

		// Verify rate limit was imported
		limit, exists := keeper.GetRateLimit("IR-LIMITED")
		require.True(t, exists)
		require.Equal(t, int64(5), limit.PerWalletPerHour)
		require.Equal(t, int64(20), limit.PerWalletPerDay)
		require.Equal(t, int64(100), limit.PerBlockGlobal)
	})

	t.Run("init with invalid genesis fails", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		genesis := types.GenesisState{
			Params: &pb.Params{
				MaxIrPerLocale:       -1, // Invalid
				DefaultRateLimitHour: 10,
				SuspensionFee:        "1000",
				MinGovernanceDeposit: "5000",
			},
			Irs:           []*pb.IRDefinition{},
			Prerequisites: []*pb.IRPrerequisite{},
			RateLimits:    []*pb.IRRateLimit{},
		}

		err := keeper.InitGenesis(genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid genesis state")
	})

	t.Run("init skips nil entries", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		genesis := types.GenesisState{
			Params: &pb.Params{
				MaxIrPerLocale:       100,
				DefaultRateLimitHour: 10,
				SuspensionFee:        "1000",
				MinGovernanceDeposit: "5000",
			},
			Irs: []*pb.IRDefinition{
				nil,
				{
					Id:          "IR-VALID",
					Name:        "Valid IR",
					Arena:       pb.Arena_ARENA_BIOMETRIC,
					Description: "Valid",
					Score:       100,
					LocaleTags:  []string{"global"},
					PrivacyTier: pb.PrivacyTier_PRIVACY_TIER_HIGH,
					Version:     "1.0",
					Status:      pb.IRStatus_IR_STATUS_ACTIVE,
				},
				nil,
			},
			Prerequisites: []*pb.IRPrerequisite{nil},
			RateLimits:    []*pb.IRRateLimit{nil},
		}

		err := keeper.InitGenesis(genesis)
		require.NoError(t, err)

		// Verify only valid IR was imported
		ir, ok := keeper.GetIR("IR-VALID")
		require.True(t, ok)
		require.Equal(t, "Valid IR", ir.Name)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")

		genesis := keeper.ExportGenesis()

		require.NotNil(t, genesis.Params)
		require.Empty(t, genesis.Irs)
		require.Empty(t, genesis.Prerequisites)
		require.Empty(t, genesis.RateLimits)
	})

	t.Run("export with data", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")

		// Create some IRs
		ir1 := types.IRDefinition{
			Id:          "IR-EXPORT-1",
			Name:        "Export Test 1",
			Arena:       types.Arena(pb.Arena_ARENA_BIOMETRIC),
			Description: "Test",
			Score:       100,
			PoiReward:   50,
			LocaleTags:  []string{"global"},
			PrivacyTier: types.PrivacyTier(pb.PrivacyTier_PRIVACY_TIER_HIGH),
			Version:     "1.0",
			Status:      types.IRStatus(pb.IRStatus_IR_STATUS_ACTIVE),
		}
		ir2 := types.IRDefinition{
			Id:          "IR-EXPORT-2",
			Name:        "Export Test 2",
			Arena:       types.Arena(pb.Arena_ARENA_KNOWLEDGE),
			Description: "Test",
			Score:       75,
			PoiReward:   25,
			LocaleTags:  []string{"global"},
			PrivacyTier: types.PrivacyTier(pb.PrivacyTier_PRIVACY_TIER_MEDIUM),
			Version:     "1.0",
			Status:      types.IRStatus(pb.IRStatus_IR_STATUS_ACTIVE),
		}

		err := keeper.CreateIR(ir1)
		require.NoError(t, err)
		err = keeper.CreateIR(ir2)
		require.NoError(t, err)

		// Set prerequisite
		prereq := types.IRPrerequisite{
			IrId:          "IR-EXPORT-2",
			RequiredIrIds: []string{"IR-EXPORT-1"},
		}
		keeper.SetPrerequisites("IR-EXPORT-2", prereq)

		// Set rate limit
		limit := types.IRRateLimit{
			IrId:             "IR-EXPORT-1",
			PerWalletPerHour: 10,
			PerWalletPerDay:  50,
			PerBlockGlobal:   200,
		}
		keeper.SetRateLimit("IR-EXPORT-1", limit)

		// Export genesis
		genesis := keeper.ExportGenesis()

		// Verify exported data
		require.NotNil(t, genesis.Params)
		require.Len(t, genesis.Irs, 2)
		require.Len(t, genesis.Prerequisites, 1)
		require.Len(t, genesis.RateLimits, 1)

		// Verify IR data
		var found1, found2 bool
		for _, ir := range genesis.Irs {
			if ir.Id == "IR-EXPORT-1" {
				found1 = true
				require.Equal(t, "Export Test 1", ir.Name)
				require.Equal(t, int64(100), ir.Score)
			}
			if ir.Id == "IR-EXPORT-2" {
				found2 = true
				require.Equal(t, "Export Test 2", ir.Name)
				require.Equal(t, int64(75), ir.Score)
			}
		}
		require.True(t, found1)
		require.True(t, found2)

		// Verify prerequisite
		require.Equal(t, "IR-EXPORT-2", genesis.Prerequisites[0].IrId)
		require.Equal(t, []string{"IR-EXPORT-1"}, genesis.Prerequisites[0].RequiredIrIds)

		// Verify rate limit
		require.Equal(t, "IR-EXPORT-1", genesis.RateLimits[0].IrId)
		require.Equal(t, int64(10), genesis.RateLimits[0].PerWalletPerHour)
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")

		// Create original genesis
		originalGenesis := types.GenesisState{
			Params: &pb.Params{
				MaxIrPerLocale:       150,
				DefaultRateLimitHour: 15,
				SuspensionFee:        "2000",
				MinGovernanceDeposit: "10000",
			},
			Irs: []*pb.IRDefinition{
				{
					Id:               "IR-RT-1",
					Name:             "RoundTrip Test 1",
					Arena:            pb.Arena_ARENA_BIOMETRIC,
					Description:      "Test roundtrip",
					Score:            200,
					PoiReward:        100,
					LocaleTags:       []string{"global", "us"},
					PrivacyTier:      pb.PrivacyTier_PRIVACY_TIER_HIGH,
					Version:          "2.0",
					MetadataHash:     "hash123",
					Status:           pb.IRStatus_IR_STATUS_ACTIVE,
					ActivationHeight: 100,
					SunsetHeight:     1000,
				},
				{
					Id:               "IR-RT-2",
					Name:             "RoundTrip Test 2",
					Arena:            pb.Arena_ARENA_KNOWLEDGE,
					Description:      "Another test",
					Score:            150,
					PoiReward:        75,
					LocaleTags:       []string{"global"},
					PrivacyTier:      pb.PrivacyTier_PRIVACY_TIER_MEDIUM,
					Version:          "1.5",
					MetadataHash:     "hash456",
					Status:           pb.IRStatus_IR_STATUS_ACTIVE,
					ActivationHeight: 50,
					SunsetHeight:     500,
				},
			},
			Prerequisites: []*pb.IRPrerequisite{
				{
					IrId:          "IR-RT-2",
					RequiredIrIds: []string{"IR-RT-1"},
				},
			},
			RateLimits: []*pb.IRRateLimit{
				{
					IrId:             "IR-RT-1",
					PerWalletPerHour: 8,
					PerWalletPerDay:  40,
					PerBlockGlobal:   150,
				},
				{
					IrId:             "IR-RT-2",
					PerWalletPerHour: 12,
					PerWalletPerDay:  60,
					PerBlockGlobal:   200,
				},
			},
		}

		// Import genesis
		err := keeper.InitGenesis(originalGenesis)
		require.NoError(t, err)

		// Export genesis
		exportedGenesis := keeper.ExportGenesis()

		// Verify params match
		require.Equal(t, originalGenesis.Params.MaxIrPerLocale, exportedGenesis.Params.MaxIrPerLocale)
		require.Equal(t, originalGenesis.Params.DefaultRateLimitHour, exportedGenesis.Params.DefaultRateLimitHour)
		require.Equal(t, originalGenesis.Params.SuspensionFee, exportedGenesis.Params.SuspensionFee)
		require.Equal(t, originalGenesis.Params.MinGovernanceDeposit, exportedGenesis.Params.MinGovernanceDeposit)

		// Verify same number of items
		require.Len(t, exportedGenesis.Irs, len(originalGenesis.Irs))
		require.Len(t, exportedGenesis.Prerequisites, len(originalGenesis.Prerequisites))
		require.Len(t, exportedGenesis.RateLimits, len(originalGenesis.RateLimits))

		// Verify IR data integrity
		for _, originalIR := range originalGenesis.Irs {
			found := false
			for _, exportedIR := range exportedGenesis.Irs {
				if exportedIR.Id == originalIR.Id {
					found = true
					require.Equal(t, originalIR.Name, exportedIR.Name)
					require.Equal(t, originalIR.Arena, exportedIR.Arena)
					require.Equal(t, originalIR.Description, exportedIR.Description)
					require.Equal(t, originalIR.Score, exportedIR.Score)
					require.Equal(t, originalIR.PoiReward, exportedIR.PoiReward)
					require.Equal(t, originalIR.LocaleTags, exportedIR.LocaleTags)
					require.Equal(t, originalIR.PrivacyTier, exportedIR.PrivacyTier)
					require.Equal(t, originalIR.Version, exportedIR.Version)
					require.Equal(t, originalIR.MetadataHash, exportedIR.MetadataHash)
					require.Equal(t, originalIR.Status, exportedIR.Status)
					require.Equal(t, originalIR.ActivationHeight, exportedIR.ActivationHeight)
					require.Equal(t, originalIR.SunsetHeight, exportedIR.SunsetHeight)
					break
				}
			}
			require.True(t, found, "IR %s not found in export", originalIR.Id)
		}
	})

	t.Run("multiple round trips are deterministic", func(t *testing.T) {
		keeper1 := NewKeeper(nil, "authority")
		keeper2 := NewKeeper(nil, "authority")

		genesis := types.DefaultGenesisState()
		genesis.Irs = []*pb.IRDefinition{
			{
				Id:          "IR-DET-1",
				Name:        "Deterministic Test",
				Arena:       pb.Arena_ARENA_BIOMETRIC,
				Description: "Test",
				Score:       50,
				LocaleTags:  []string{"global"},
				PrivacyTier: pb.PrivacyTier_PRIVACY_TIER_HIGH,
				Version:     "1.0",
				Status:      pb.IRStatus_IR_STATUS_ACTIVE,
			},
		}

		// First round trip
		err := keeper1.InitGenesis(*genesis)
		require.NoError(t, err)
		export1 := keeper1.ExportGenesis()

		// Second round trip
		err = keeper2.InitGenesis(export1)
		require.NoError(t, err)
		export2 := keeper2.ExportGenesis()

		// Verify exports match
		require.Equal(t, len(export1.Irs), len(export2.Irs))
		require.Equal(t, export1.Params.MaxIrPerLocale, export2.Params.MaxIrPerLocale)
	})
}

func TestDefaultGenesis(t *testing.T) {
	t.Run("default genesis is valid", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		// Validate default genesis
		err := types.ValidateGenesisState(genesis)
		require.NoError(t, err)

		// Verify structure
		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.Irs)
		require.NotNil(t, genesis.Prerequisites)
		require.NotNil(t, genesis.RateLimits)

		// Verify default params are reasonable
		require.Greater(t, genesis.Params.MaxIrPerLocale, int64(0))
		require.Greater(t, genesis.Params.DefaultRateLimitHour, int64(0))
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		keeper := NewKeeper(nil, "authority")
		genesis := types.DefaultGenesisState()

		err := keeper.InitGenesis(*genesis)
		require.NoError(t, err)

		// Verify params are set
		params := keeper.paramsStore.GetParams()
		require.Equal(t, genesis.Params.MaxIrPerLocale, params.MaxIrPerLocale)
	})
}
