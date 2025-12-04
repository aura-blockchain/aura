package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// setupKeeperWithKVStore creates a keeper wired to an in-memory KV store for full KV code paths.
func setupKeeperWithKVStore(t *testing.T) (*Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	keeper := NewKeeper(params.NewStore(*types.DefaultParams()), "authority").WithStore(storeKey, cdc)
	header := tmproto.Header{Height: 1, Time: time.Now()}
	ctx := sdk.NewContext(stateStore, header, false, log.NewNopLogger())
	return keeper, ctx
}

func TestKVAttributeVCLifecycle(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	avc := types.AttributeVC{
		AttributeVcId:  keeper.GenerateAttributeVCID(ctx, "addr1", types.AttributeType_ATTRIBUTE_TYPE_AGE),
		HolderAddress:  "addr1",
		AttributeType:  types.AttributeType_ATTRIBUTE_TYPE_AGE,
		EncryptedValue: []byte("ciphertext"),
		ExpiresAt:      timestamppb.New(time.Unix(now+3600, 0)),
	}

	require.NoError(t, keeper.CreateAttributeVC(ctx, avc))

	stored, ok := keeper.GetAttributeVC(ctx, avc.AttributeVcId)
	require.True(t, ok)
	require.Equal(t, avc.AttributeType, stored.AttributeType)

	listed := keeper.ListAttributeVCs(ctx, avc.HolderAddress, nil)
	require.Len(t, listed, 1)
	require.Equal(t, avc.AttributeVcId, listed[0].AttributeVcId)

	require.NoError(t, keeper.RevokeAttributeVC(ctx, avc.AttributeVcId, "reason"))
	revoked, ok := keeper.GetAttributeVC(ctx, avc.AttributeVcId)
	require.True(t, ok)
	require.Equal(t, types.VCStatus_VC_STATUS_REVOKED, revoked.Status)
}

func TestKVDisclosureRequestResponseFlow(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	holder := "addr1"
	req := types.DisclosureRequest{
		RequestId:           "req-kv",
		VerifierAddress:     "verifier1",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		Purpose:             "verification",
		RequestedAt:         timestamppb.New(time.Unix(now, 0)),
		ExpiresInSeconds:    600,
	}

	require.NoError(t, keeper.CreateDisclosureRequest(ctx, holder, req))
	pending := keeper.store.listPendingDisclosures(ctx, holder)
	require.Contains(t, pending, req.RequestId)

	resp := types.DisclosureResponse{
		RequestId:     req.RequestId,
		HolderAddress: holder,
		Approved:      true,
		DisclosedAttributes: []*types.AttributeDisclosure{
			{AttributeType: types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		},
	}

	require.NoError(t, keeper.RespondToDisclosureRequest(ctx, resp))
	_, ok := keeper.store.getDisclosureResponse(ctx, req.RequestId)
	require.True(t, ok)
	require.NotContains(t, keeper.store.listPendingDisclosures(ctx, holder), req.RequestId)
}

func TestKVPersistencePresentation(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	holder := "addr1"
	vc := types.VCRecord{
		VcId:          "vc1",
		HolderAddress: holder,
		HolderDid:     "did:aura:holder",
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      timestamppb.New(time.Unix(now, 0)),
		ExpiresAt:     timestamppb.New(time.Unix(now+3600, 0)),
		VcType:        vcregistrypb.VCType_VC_TYPE_CUSTOM,
	}
	require.NoError(t, keeper.SetVCRecord(ctx, vc))

	presentation, _, err := keeper.CreatePresentation(ctx, holder, []string{vc.VcId}, nil, 300)
	require.NoError(t, err)
	require.NotNil(t, presentation)

	stored, ok := keeper.store.getPresentation(ctx, presentation.PresentationId)
	require.True(t, ok)
	require.Equal(t, presentation.PresentationId, stored.PresentationId)

	idx := keeper.store.listUserPresentations(ctx, holder)
	require.Contains(t, idx, presentation.PresentationId)
}

func TestGenesisRoundTripKVSelectiveDisclosure(t *testing.T) {
	keeper1, ctx1 := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper1.SetCurrentTime(now)

	holder := "addr1"
	avc := types.AttributeVC{
		AttributeVcId:  keeper1.GenerateAttributeVCID(ctx1, holder, types.AttributeType_ATTRIBUTE_TYPE_AGE),
		HolderAddress:  holder,
		AttributeType:  types.AttributeType_ATTRIBUTE_TYPE_AGE,
		EncryptedValue: []byte("cipher"),
		ExpiresAt:      timestamppb.New(time.Unix(now+7200, 0)),
	}
	require.NoError(t, keeper1.CreateAttributeVC(ctx1, avc))

	pol := types.DisclosurePolicy{
		HolderAddress: holder,
		DefaultMode:   types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY,
		Rules: []*types.AttributeDisclosureRule{
			{
				AttributeType: types.AttributeType_ATTRIBUTE_TYPE_AGE,
				Mode:          types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
			},
		},
	}
	require.NoError(t, keeper1.SetDisclosurePolicy(ctx1, pol))

	reqPending := types.DisclosureRequest{
		RequestId:           "req-pending",
		VerifierAddress:     "verifier1",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_AGE},
		RequestedAt:         timestamppb.New(time.Unix(now, 0)),
		ExpiresInSeconds:    600,
	}
	require.NoError(t, keeper1.CreateDisclosureRequest(ctx1, holder, reqPending))

	reqAnswered := types.DisclosureRequest{
		RequestId:           "req-answered",
		VerifierAddress:     "verifier2",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		RequestedAt:         timestamppb.New(time.Unix(now, 0)),
		ExpiresInSeconds:    600,
	}
	require.NoError(t, keeper1.CreateDisclosureRequest(ctx1, holder, reqAnswered))
	resp := types.DisclosureResponse{
		RequestId:     reqAnswered.RequestId,
		HolderAddress: holder,
		Approved:      true,
		DisclosedAttributes: []*types.AttributeDisclosure{
			{AttributeType: types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		},
	}
	require.NoError(t, keeper1.RespondToDisclosureRequest(ctx1, resp))

	// Seed a VC to satisfy presentation creation
	vc := types.VCRecord{
		VcId:              "vc-genesis",
		HolderAddress:     holder,
		HolderDid:         "did:aura:holder",
		Status:            types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:          timestamppb.New(time.Unix(now, 0)),
		ExpiresAt:         timestamppb.New(time.Unix(now+7200, 0)),
		VcType:            vcregistrypb.VCType_VC_TYPE_CUSTOM,
		IssuerAssistant:   "aura1issuer",
	}
	require.NoError(t, keeper1.SetVCRecord(ctx1, vc))

	pres, _, err := keeper1.CreatePresentation(ctx1, holder, []string{vc.VcId}, nil, 300)
	require.NoError(t, err)
	require.NotNil(t, pres)

	gs := keeper1.ExportGenesis(ctx1)

	keeper2, ctx2 := setupKeeperWithKVStore(t)
	keeper2.SetCurrentTime(now)
	require.NoError(t, keeper2.InitGenesis(ctx2, gs))

	restoredAvc, ok := keeper2.GetAttributeVC(ctx2, avc.AttributeVcId)
	require.True(t, ok)
	require.Equal(t, avc.AttributeType, restoredAvc.AttributeType)

	restoredPol, ok := keeper2.GetDisclosurePolicy(ctx2, holder)
	require.True(t, ok)
	require.Equal(t, pol.DefaultMode, restoredPol.DefaultMode)

	restoredReq, ok := keeper2.GetDisclosureRequest(ctx2, reqPending.RequestId)
	require.True(t, ok)
	require.Equal(t, reqPending.VerifierAddress, restoredReq.VerifierAddress)
	require.Contains(t, keeper2.store.listPendingDisclosures(ctx2, holder), reqPending.RequestId)

	_, ok = keeper2.store.getDisclosureResponse(ctx2, reqAnswered.RequestId)
	require.True(t, ok)

	presIdx := keeper2.store.listUserPresentations(ctx2, holder)
	require.Contains(t, presIdx, pres.PresentationId)
}

func TestKVAttributeVCValidationAndIndexing(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	holder := "addr1"
	base := types.AttributeVC{
		AttributeVcId:  keeper.GenerateAttributeVCID(ctx, holder, types.AttributeType_ATTRIBUTE_TYPE_AGE),
		HolderAddress:  holder,
		AttributeType:  types.AttributeType_ATTRIBUTE_TYPE_AGE,
		EncryptedValue: []byte("ciphertext"),
		ExpiresAt:      timestamppb.New(time.Unix(now+600, 0)),
	}

	require.NoError(t, keeper.CreateAttributeVC(ctx, base))

	expired := base
	expired.AttributeVcId = keeper.GenerateAttributeVCID(ctx, holder, types.AttributeType_ATTRIBUTE_TYPE_EMAIL)
	expired.AttributeType = types.AttributeType_ATTRIBUTE_TYPE_EMAIL
	expired.ExpiresAt = timestamppb.New(time.Unix(now-10, 0))
	require.ErrorContains(t, keeper.CreateAttributeVC(ctx, expired), "expired")

	duplicate := base
	duplicate.AttributeVcId = keeper.GenerateAttributeVCID(ctx, holder, base.AttributeType)
	require.Error(t, keeper.CreateAttributeVC(ctx, duplicate))

	otherHolder := base
	otherHolder.HolderAddress = "addr2"
	otherHolder.AttributeVcId = keeper.GenerateAttributeVCID(ctx, otherHolder.HolderAddress, otherHolder.AttributeType)
	require.NoError(t, keeper.CreateAttributeVC(ctx, otherHolder))

	list := keeper.ListAttributeVCs(ctx, holder, nil)
	require.Len(t, list, 1)
	require.Equal(t, holder, list[0].HolderAddress)
}

func TestKVDisclosurePolicyValidationAndUpdate(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	dupRule := types.DisclosurePolicy{
		HolderAddress: "addr1",
		Rules: []*types.AttributeDisclosureRule{
			{
				AttributeType: types.AttributeType_ATTRIBUTE_TYPE_EMAIL,
				Mode:          types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
			},
			{
				AttributeType: types.AttributeType_ATTRIBUTE_TYPE_EMAIL,
				Mode:          types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY,
			},
		},
	}
	require.ErrorContains(t, keeper.SetDisclosurePolicy(ctx, dupRule), "duplicate")

	policy := types.DisclosurePolicy{HolderAddress: dupRule.HolderAddress}
	require.NoError(t, keeper.SetDisclosurePolicy(ctx, policy))

	stored, ok := keeper.GetDisclosurePolicy(ctx, policy.HolderAddress)
	require.True(t, ok)
	require.Equal(t, types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY, stored.DefaultMode)
	require.NotNil(t, stored.UpdatedAt)
	require.Equal(t, now, stored.UpdatedAt.Seconds)

	keeper.SetCurrentTime(now + 15)
	updated := types.DisclosurePolicy{
		HolderAddress: policy.HolderAddress,
		DefaultMode:   types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
		Rules: []*types.AttributeDisclosureRule{
			{
				AttributeType: types.AttributeType_ATTRIBUTE_TYPE_AGE,
				Mode:          types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
			},
		},
	}
	require.NoError(t, keeper.SetDisclosurePolicy(ctx, updated))

	restored, ok := keeper.GetDisclosurePolicy(ctx, policy.HolderAddress)
	require.True(t, ok)
	require.Equal(t, updated.DefaultMode, restored.DefaultMode)
	require.Len(t, restored.Rules, 1)
	require.Equal(t, int64(now+15), restored.UpdatedAt.Seconds)
}

func TestKVDisclosureRequestExpiryAndDefaults(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	holder := "holder1"

	tooLong := types.DisclosureRequest{
		RequestId:           "req-too-long",
		VerifierAddress:     "verifier",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		ExpiresInSeconds:    90000,
		RequestedAt:         timestamppb.New(time.Unix(now, 0)),
	}
	require.ErrorContains(t, keeper.CreateDisclosureRequest(ctx, holder, tooLong), "too long")

	expired := types.DisclosureRequest{
		RequestId:           "req-expired",
		VerifierAddress:     "verifier",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		ExpiresInSeconds:    10,
		RequestedAt:         timestamppb.New(time.Unix(now-20, 0)),
	}
	require.ErrorContains(t, keeper.CreateDisclosureRequest(ctx, holder, expired), "expired")

	defaults := types.DisclosureRequest{
		RequestId:           "req-default",
		VerifierAddress:     "verifier",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
	}
	require.NoError(t, keeper.CreateDisclosureRequest(ctx, holder, defaults))

	stored, ok := keeper.GetDisclosureRequest(ctx, defaults.RequestId)
	require.True(t, ok)
	require.Equal(t, uint64(300), stored.ExpiresInSeconds)
	require.NotNil(t, stored.RequestedAt)
	require.Contains(t, keeper.store.listPendingDisclosures(ctx, holder), defaults.RequestId)
}

func TestKVDisclosureResponseGuards(t *testing.T) {
	keeper, ctx := setupKeeperWithKVStore(t)
	now := time.Now().Unix()
	keeper.SetCurrentTime(now)

	holder := "holder1"
	req := types.DisclosureRequest{
		RequestId:           "req1",
		VerifierAddress:     "verifier1",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_AGE},
		RequestedAt:         timestamppb.New(time.Unix(now, 0)),
		ExpiresInSeconds:    120,
	}
	require.NoError(t, keeper.CreateDisclosureRequest(ctx, holder, req))

	wrongHolderResp := types.DisclosureResponse{
		RequestId:     req.RequestId,
		HolderAddress: "other-holder",
		Approved:      true,
		DisclosedAttributes: []*types.AttributeDisclosure{
			{AttributeType: types.AttributeType_ATTRIBUTE_TYPE_AGE},
		},
	}
	require.ErrorContains(t, keeper.RespondToDisclosureRequest(ctx, wrongHolderResp), "pending for holder")

	keeper.SetCurrentTime(now + 1000)
	expiredResp := types.DisclosureResponse{
		RequestId:     req.RequestId,
		HolderAddress: holder,
		Approved:      true,
		DisclosedAttributes: []*types.AttributeDisclosure{
			{AttributeType: types.AttributeType_ATTRIBUTE_TYPE_AGE},
		},
	}
	require.ErrorContains(t, keeper.RespondToDisclosureRequest(ctx, expiredResp), "expired")

	keeper.SetCurrentTime(now + 10)
	freshReq := types.DisclosureRequest{
		RequestId:           "req2",
		VerifierAddress:     "verifier2",
		RequestedAttributes: []types.AttributeType{types.AttributeType_ATTRIBUTE_TYPE_EMAIL},
		RequestedAt:         timestamppb.New(time.Unix(now+10, 0)),
		ExpiresInSeconds:    200,
	}
	require.NoError(t, keeper.CreateDisclosureRequest(ctx, holder, freshReq))

	invalidResp := types.DisclosureResponse{
		RequestId:     freshReq.RequestId,
		HolderAddress: holder,
		Approved:      true,
		DisclosedAttributes: []*types.AttributeDisclosure{
			{AttributeType: types.AttributeType_ATTRIBUTE_TYPE_AGE},
		},
	}
	require.ErrorContains(t, keeper.RespondToDisclosureRequest(ctx, invalidResp), "was not requested")
}
