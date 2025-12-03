package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	pb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

func TestInitGenesis(t *testing.T) {
	t.Run("init with default genesis", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, *genesis)
		require.NoError(t, err)
	})

	t.Run("init with VC records", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params: types.DefaultParamsProto(),
			VcRecords: []*pb.VCRecord{
				{
					VcId:       "vc1",
					Holder:     "holder1",
					Issuer:     "issuer1",
					VcType:     "identity",
					IssuedAt:   1000,
					ExpiresAt:  5000,
					Revoked:    false,
					Attributes: map[string]string{"name": "John"},
				},
				{
					VcId:       "vc2",
					Holder:     "holder2",
					Issuer:     "issuer1",
					VcType:     "credential",
					IssuedAt:   2000,
					ExpiresAt:  6000,
					Revoked:    false,
					Attributes: map[string]string{"role": "admin"},
				},
			},
			RevocationRecords:     []*pb.RevocationRecord{},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:          []*pb.DIDDocument{},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify VCs were imported
		vc1, ok := k.GetVCRecord(ctx, "vc1")
		require.True(t, ok)
		require.Equal(t, "holder1", vc1.Holder)

		vc2, ok := k.GetVCRecord(ctx, "vc2")
		require.True(t, ok)
		require.Equal(t, "holder2", vc2.Holder)
	})

	t.Run("init with revocation records", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params:    types.DefaultParamsProto(),
			VcRecords: []*pb.VCRecord{},
			RevocationRecords: []*pb.RevocationRecord{
				{
					VcId:       "vc1",
					RevokedAt:  1000,
					Reason:     "compromised",
					RevokedBy:  "issuer1",
				},
				{
					VcId:       "vc2",
					RevokedAt:  2000,
					Reason:     "expired",
					RevokedBy:  "issuer2",
				},
			},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{"vc1", "vc2"}},
			DidDocuments:          []*pb.DIDDocument{},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify revocations were imported
		list, err := k.GetRevocationList(ctx)
		require.NoError(t, err)
		require.Len(t, list.RevokedVcIds, 2)
	})

	t.Run("init with DID documents", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params:            types.DefaultParamsProto(),
			VcRecords:         []*pb.VCRecord{},
			RevocationRecords: []*pb.RevocationRecord{},
			RevocationList:    &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments: []*pb.DIDDocument{
				{
					Id:         "did:aura:1",
					Controller: "controller1",
					PublicKeys: []string{"key1", "key2"},
					CreatedAt:  1000,
					UpdatedAt:  1000,
					Active:     true,
				},
				{
					Id:         "did:aura:2",
					Controller: "controller2",
					PublicKeys: []string{"key3"},
					CreatedAt:  2000,
					UpdatedAt:  2000,
					Active:     true,
				},
			},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with VC policies", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params:            types.DefaultParamsProto(),
			VcRecords:         []*pb.VCRecord{},
			RevocationRecords: []*pb.RevocationRecord{},
			RevocationList:    &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:      []*pb.DIDDocument{},
			VcPolicies: []*pb.VCPolicy{
				{
					PolicyId:         "policy1",
					Name:             "Standard Policy",
					Description:      "Standard verification policy",
					RequiredVcTypes:  []string{"identity", "credential"},
					MinVerifications: 2,
					Active:           true,
				},
			},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with user mint counts", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params:                types.DefaultParamsProto(),
			VcRecords:             []*pb.VCRecord{},
			RevocationRecords:     []*pb.RevocationRecord{},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:          []*pb.DIDDocument{},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{"user1": 5, "user2": 10},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify mint counts were imported
		count, err := k.GetUserMintCount(ctx, "user1")
		require.NoError(t, err)
		require.Equal(t, uint64(5), count)
	})

	t.Run("init with presentations", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params:            types.DefaultParamsProto(),
			VcRecords:         []*pb.VCRecord{},
			RevocationRecords: []*pb.RevocationRecord{},
			RevocationList:    &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:      []*pb.DIDDocument{},
			VcPolicies:        []*pb.VCPolicy{},
			UserMintCounts:    map[string]uint64{},
			Presentations: []*pb.VCPresentation{
				{
					PresentationId: "pres1",
					Holder:         "holder1",
					VcIds:          []string{"vc1", "vc2"},
					CreatedAt:      1000,
					ExpiresAt:      5000,
					Verified:       true,
				},
			},
			UserPresentationIndex: map[string]*pb.PresentationIds{
				"holder1": {PresentationIds: []string{"pres1"}},
			},
			AttributeVcs:       []*pb.AttributeVC{},
			UserAttributeIndex: map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with attribute VCs", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params:                types.DefaultParamsProto(),
			VcRecords:             []*pb.VCRecord{},
			RevocationRecords:     []*pb.RevocationRecord{},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:          []*pb.DIDDocument{},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs: []*pb.AttributeVC{
				{
					AttributeVcId: "attr1",
					Holder:        "holder1",
					AttributeName: "age",
					AttributeHash: []byte("hash1"),
					IssuedAt:      1000,
					Disclosed:     false,
				},
			},
			UserAttributeIndex: map[string]*pb.AttributeVcIds{
				"holder1": {AttributeVcIds: []string{"attr1"}},
			},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)
	})

	t.Run("init with invalid genesis fails", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params: &pb.Params{
				MaxVcsPerUser: 0, // Invalid
			},
			VcRecords:             []*pb.VCRecord{},
			RevocationRecords:     []*pb.RevocationRecord{},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:          []*pb.DIDDocument{},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
	})

	t.Run("init skips nil entries", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.GenesisState{
			Params: types.DefaultParamsProto(),
			VcRecords: []*pb.VCRecord{
				nil,
				{VcId: "vc1", Holder: "holder1", Issuer: "issuer1"},
				nil,
			},
			RevocationRecords:     []*pb.RevocationRecord{nil},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:          []*pb.DIDDocument{nil},
			VcPolicies:            []*pb.VCPolicy{nil},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{nil},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{nil},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
		}

		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify only valid VC was imported
		vc, ok := k.GetVCRecord(ctx, "vc1")
		require.True(t, ok)
		require.Equal(t, "holder1", vc.Holder)
	})
}

func TestExportGenesis(t *testing.T) {
	t.Run("export empty state", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := k.ExportGenesis(ctx)

		require.NotNil(t, genesis.Params)
		require.Empty(t, genesis.VcRecords)
		require.Empty(t, genesis.RevocationRecords)
		require.NotNil(t, genesis.RevocationList)
		require.Empty(t, genesis.DidDocuments)
		require.Empty(t, genesis.VcPolicies)
	})

	t.Run("export with data", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		// Initialize with data
		initGenesis := types.GenesisState{
			Params: types.DefaultParamsProto(),
			VcRecords: []*pb.VCRecord{
				{VcId: "vc1", Holder: "holder1", Issuer: "issuer1"},
				{VcId: "vc2", Holder: "holder2", Issuer: "issuer1"},
			},
			RevocationRecords: []*pb.RevocationRecord{
				{VcId: "vc1", Reason: "compromised"},
			},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{"vc1"}},
			DidDocuments:          []*pb.DIDDocument{{Id: "did:aura:1"}},
			VcPolicies:            []*pb.VCPolicy{{PolicyId: "policy1"}},
			UserMintCounts:        map[string]uint64{"user1": 5},
			Presentations:         []*pb.VCPresentation{{PresentationId: "pres1"}},
			UserPresentationIndex: map[string]*pb.PresentationIds{"user1": {PresentationIds: []string{"pres1"}}},
			AttributeVcs:          []*pb.AttributeVC{{AttributeVcId: "attr1"}},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{"user1": {AttributeVcIds: []string{"attr1"}}},
		}

		err := k.InitGenesis(ctx, initGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		require.Len(t, exported.VcRecords, 2)
		require.Len(t, exported.RevocationRecords, 1)
		require.Len(t, exported.DidDocuments, 1)
		require.Len(t, exported.VcPolicies, 1)
		require.Len(t, exported.UserMintCounts, 1)
		require.Len(t, exported.Presentations, 1)
		require.Len(t, exported.AttributeVcs, 1)
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	t.Run("init then export produces same state", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		originalGenesis := types.GenesisState{
			Params: &pb.Params{
				MaxVcsPerUser:         100,
				MinCredentialDuration: 86400,
				MaxCredentialDuration: 31536000,
				RevocationFee:         "1000",
				AllowSelfIssued:       true,
			},
			VcRecords: []*pb.VCRecord{
				{
					VcId:       "vc1",
					Holder:     "holder1",
					Issuer:     "issuer1",
					VcType:     "identity",
					IssuedAt:   1000,
					ExpiresAt:  5000,
					Revoked:    false,
					Attributes: map[string]string{"name": "Alice"},
				},
			},
			RevocationRecords:  []*pb.RevocationRecord{},
			RevocationList:     &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:       []*pb.DIDDocument{{Id: "did:aura:1", Controller: "controller1"}},
			VcPolicies:         []*pb.VCPolicy{{PolicyId: "policy1", Name: "Standard"}},
			UserMintCounts:     map[string]uint64{"holder1": 1},
			Presentations:      []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:       []*pb.AttributeVC{},
			UserAttributeIndex: map[string]*pb.AttributeVcIds{},
		}

		// Import
		err := k.InitGenesis(ctx, originalGenesis)
		require.NoError(t, err)

		// Export
		exported := k.ExportGenesis(ctx)

		// Verify params match
		require.Equal(t, originalGenesis.Params.MaxVcsPerUser, exported.Params.MaxVcsPerUser)
		require.Equal(t, originalGenesis.Params.MinCredentialDuration, exported.Params.MinCredentialDuration)
		require.Equal(t, originalGenesis.Params.RevocationFee, exported.Params.RevocationFee)

		// Verify counts match
		require.Len(t, exported.VcRecords, len(originalGenesis.VcRecords))
		require.Len(t, exported.DidDocuments, len(originalGenesis.DidDocuments))
		require.Len(t, exported.VcPolicies, len(originalGenesis.VcPolicies))
		require.Len(t, exported.UserMintCounts, len(originalGenesis.UserMintCounts))

		// Verify VC data integrity
		require.Equal(t, "vc1", exported.VcRecords[0].VcId)
		require.Equal(t, "holder1", exported.VcRecords[0].Holder)
	})

	t.Run("multiple round trips are deterministic", func(t *testing.T) {
		k1 := NewKeeper(nil, "authority")
		k2 := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.DefaultGenesisState()
		genesis.VcRecords = []*pb.VCRecord{
			{VcId: "vc1", Holder: "holder1", Issuer: "issuer1"},
		}

		// First round trip
		err := k1.InitGenesis(ctx, *genesis)
		require.NoError(t, err)
		export1 := k1.ExportGenesis(ctx)

		// Second round trip
		err = k2.InitGenesis(ctx, export1)
		require.NoError(t, err)
		export2 := k2.ExportGenesis(ctx)

		// Verify exports match
		require.Len(t, export2.VcRecords, len(export1.VcRecords))
		require.Equal(t, export1.Params.MaxVcsPerUser, export2.Params.MaxVcsPerUser)
	})
}

func TestDefaultGenesis(t *testing.T) {
	t.Run("default genesis is valid", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		err := types.ValidateGenesisState(genesis)
		require.NoError(t, err)

		require.NotNil(t, genesis.Params)
		require.NotNil(t, genesis.VcRecords)
		require.NotNil(t, genesis.RevocationList)
	})

	t.Run("can init with default genesis", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		genesis := types.DefaultGenesisState()
		err := k.InitGenesis(ctx, *genesis)
		require.NoError(t, err)
	})

	t.Run("default params are reasonable", func(t *testing.T) {
		genesis := types.DefaultGenesisState()

		require.Greater(t, genesis.Params.MaxVcsPerUser, uint64(0))
		require.Greater(t, genesis.Params.MinCredentialDuration, uint64(0))
		require.Greater(t, genesis.Params.MaxCredentialDuration, genesis.Params.MinCredentialDuration)
	})
}

func TestGenesisIndexNoDuplicates(t *testing.T) {
	t.Run("presentation index not built twice", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		// Genesis with presentations AND their index
		genesis := types.GenesisState{
			Params:            types.DefaultParamsProto(),
			VcRecords:         []*pb.VCRecord{},
			RevocationRecords: []*pb.RevocationRecord{},
			RevocationList:    &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:      []*pb.DIDDocument{},
			VcPolicies:        []*pb.VCPolicy{},
			UserMintCounts:    map[string]uint64{},
			Presentations: []*pb.VCPresentation{
				{
					PresentationId: "pres1",
					HolderAddress:  "holder1",
					VcIds:          []string{"vc1"},
					CreatedAt:      1000,
				},
				{
					PresentationId: "pres2",
					HolderAddress:  "holder1",
					VcIds:          []string{"vc2"},
					CreatedAt:      2000,
				},
			},
			UserPresentationIndex: map[string]*pb.PresentationIds{
				"holder1": {Ids: []string{"pres1", "pres2"}},
			},
			AttributeVcs:       []*pb.AttributeVC{},
			UserAttributeIndex: map[string]*pb.AttributeVcIds{},
		}

		// Import should succeed and validate indexes
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify no duplicates in index
		presentations := k.store.listUserPresentations(ctx, "holder1")
		require.Len(t, presentations, 2)
		require.Contains(t, presentations, "pres1")
		require.Contains(t, presentations, "pres2")

		// Verify each ID appears exactly once
		idCount := make(map[string]int)
		for _, id := range presentations {
			idCount[id]++
		}
		for id, count := range idCount {
			require.Equal(t, 1, count, "ID %s appears %d times, expected 1", id, count)
		}
	})

	t.Run("attribute VC index not built twice", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		// Genesis with attribute VCs AND their index
		genesis := types.GenesisState{
			Params:                types.DefaultParamsProto(),
			VcRecords:             []*pb.VCRecord{},
			RevocationRecords:     []*pb.RevocationRecord{},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:          []*pb.DIDDocument{},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs: []*pb.AttributeVC{
				{
					AttributeVcId: "attr1",
					HolderAddress: "holder1",
					AttributeName: "age",
					IssuedAt:      1000,
				},
				{
					AttributeVcId: "attr2",
					HolderAddress: "holder1",
					AttributeName: "name",
					IssuedAt:      2000,
				},
			},
			UserAttributeIndex: map[string]*pb.AttributeVcIds{
				"holder1": {Ids: []string{"attr1", "attr2"}},
			},
		}

		// Import should succeed and validate indexes
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify no duplicates in index
		attributes := k.store.listUserAttributeVCs(ctx, "holder1")
		require.Len(t, attributes, 2)
		require.Contains(t, attributes, "attr1")
		require.Contains(t, attributes, "attr2")

		// Verify each ID appears exactly once
		idCount := make(map[string]int)
		for _, id := range attributes {
			idCount[id]++
		}
		for id, count := range idCount {
			require.Equal(t, 1, count, "ID %s appears %d times, expected 1", id, count)
		}
	})

	t.Run("index validation detects mismatch", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		// Genesis with presentations but MISMATCHED index
		genesis := types.GenesisState{
			Params:            types.DefaultParamsProto(),
			VcRecords:         []*pb.VCRecord{},
			RevocationRecords: []*pb.RevocationRecord{},
			RevocationList:    &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:      []*pb.DIDDocument{},
			VcPolicies:        []*pb.VCPolicy{},
			UserMintCounts:    map[string]uint64{},
			Presentations: []*pb.VCPresentation{
				{
					PresentationId: "pres1",
					HolderAddress:  "holder1",
					VcIds:          []string{"vc1"},
					CreatedAt:      1000,
				},
			},
			UserPresentationIndex: map[string]*pb.PresentationIds{
				"holder1": {Ids: []string{"pres1", "pres_wrong"}}, // Extra wrong ID
			},
			AttributeVcs:       []*pb.AttributeVC{},
			UserAttributeIndex: map[string]*pb.AttributeVcIds{},
		}

		// Import should fail due to index mismatch
		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "index mismatch")
	})

	t.Run("index validation detects count mismatch", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		// Genesis with presentations but WRONG COUNT in index
		genesis := types.GenesisState{
			Params:            types.DefaultParamsProto(),
			VcRecords:         []*pb.VCRecord{},
			RevocationRecords: []*pb.RevocationRecord{},
			RevocationList:    &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:      []*pb.DIDDocument{},
			VcPolicies:        []*pb.VCPolicy{},
			UserMintCounts:    map[string]uint64{},
			Presentations: []*pb.VCPresentation{
				{
					PresentationId: "pres1",
					HolderAddress:  "holder1",
					VcIds:          []string{"vc1"},
					CreatedAt:      1000,
				},
			},
			UserPresentationIndex: map[string]*pb.PresentationIds{
				"holder1": {Ids: []string{"pres1", "pres2"}}, // Extra entry
			},
			AttributeVcs:       []*pb.AttributeVC{},
			UserAttributeIndex: map[string]*pb.AttributeVcIds{},
		}

		// Import should fail due to count mismatch
		err := k.InitGenesis(ctx, genesis)
		require.Error(t, err)
		require.Contains(t, err.Error(), "built 1 entries, exported 2 entries")
	})

	t.Run("pending disclosure index validated", func(t *testing.T) {
		k := NewKeeper(nil, "authority")
		ctx := context.Background()

		// Genesis with pending disclosure index
		genesis := types.GenesisState{
			Params:                types.DefaultParamsProto(),
			VcRecords:             []*pb.VCRecord{},
			RevocationRecords:     []*pb.RevocationRecord{},
			RevocationList:        &pb.RevocationList{RevokedVcIds: []string{}},
			DidDocuments:          []*pb.DIDDocument{},
			VcPolicies:            []*pb.VCPolicy{},
			UserMintCounts:        map[string]uint64{},
			Presentations:         []*pb.VCPresentation{},
			UserPresentationIndex: map[string]*pb.PresentationIds{},
			AttributeVcs:          []*pb.AttributeVC{},
			UserAttributeIndex:    map[string]*pb.AttributeVcIds{},
			DisclosurePolicies:    []*pb.DisclosurePolicy{},
			DisclosureRequests:    []*pb.DisclosureRequest{},
			DisclosureResponses:   []*pb.DisclosureResponse{},
			PendingDisclosureIndex: map[string]*pb.RequestIds{
				"holder1": {Ids: []string{"req1", "req2"}},
			},
		}

		// Import should succeed
		err := k.InitGenesis(ctx, genesis)
		require.NoError(t, err)

		// Verify pending disclosures imported correctly
		pending := k.store.listPendingDisclosures(ctx, "holder1")
		require.Len(t, pending, 2)
		require.Contains(t, pending, "req1")
		require.Contains(t, pending, "req2")
	})
}
