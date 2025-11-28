package vcregistry

import (
	"fmt"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/vcregistry/keeper"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// testCSKeeper is a lightweight confidence score keeper implementation for tests.
type testCSKeeper struct {
	score        uint64
	anchor       bool
	completedIRs map[string]bool
	arenaScores  map[string]uint64
}

func (t testCSKeeper) GetUserScore(string) (uint64, bool) {
	return t.score, true
}

func (t testCSKeeper) HasCompletedIR(_, ir string) bool {
	return t.completedIRs != nil && t.completedIRs[ir]
}

func (t testCSKeeper) GetArenaScore(_, arena string) (uint64, error) {
	if t.arenaScores == nil {
		return 0, fmt.Errorf("missing arena score for %s", arena)
	}
	score, ok := t.arenaScores[arena]
	if !ok {
		return 0, fmt.Errorf("missing arena score for %s", arena)
	}
	return score, nil
}

func (t testCSKeeper) GetAnchorInfo(string) (interface{}, bool) {
	if t.anchor {
		return struct{}{}, true
	}
	return nil, false
}

func (t testCSKeeper) IsVerified(string) bool { return true }

func TestQueryServerCheckVCStatusReflectsContextTime(t *testing.T) {
	k, ctx := setupMsgServerKeeper(t)
	base := time.Now().Unix()
	k.SetCurrentTime(base)
	k.SetCurrentHeight(1)

	vc := types.VCRecord{
		VcId:          "vc-context",
		HolderAddress: "holder1",
		HolderDid:     "did:aura:holder1",
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		VcType:        types.VCType_VC_TYPE_KYC_VERIFICATION,
		IssuedAt:      timestamppb.New(time.Unix(base, 0)),
		ExpiresAt:     timestamppb.New(time.Unix(base+10, 0)),
	}
	require.NoError(t, k.SetVCRecord(ctx, vc))

	query := keeper.NewQueryServer(k)
	laterCtx := ctx.WithBlockTime(time.Unix(base+100, 0)).WithBlockHeight(ctx.BlockHeight() + 1)
	resp, err := query.CheckVCStatus(sdk.WrapSDKContext(laterCtx), &vcregistrypb.QueryCheckVCStatusRequest{VcId: vc.VcId})
	require.NoError(t, err)
	require.Equal(t, vcregistrypb.VCStatus_VC_STATUS_EXPIRED, resp.Status)
	require.False(t, resp.Valid)

	stored, ok := k.GetVCRecord(laterCtx, vc.VcId)
	require.True(t, ok)
	require.Equal(t, types.VCStatus_VC_STATUS_EXPIRED, stored.Status)
}

func TestQueryServerValidateMintEligibilityEdgeCases(t *testing.T) {
	k, ctx := setupMsgServerKeeper(t)
	query := keeper.NewQueryServer(k)
	holder := "aura1holder"
	req := &vcregistrypb.QueryValidateMintEligibilityRequest{
		HolderAddress: holder,
		VcType:        vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION,
	}

	t.Run("policy missing", func(t *testing.T) {
		resp, err := query.ValidateMintEligibility(sdk.WrapSDKContext(ctx), req)
		require.NoError(t, err)
		require.False(t, resp.Eligible)
		require.Contains(t, resp.MissingRequirements, "Policy not found for VC type")
	})

	policy := types.VCPolicy{
		VcTypeName:         vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION.String(),
		VcTypeEnum:         types.VCType_VC_TYPE_KYC_VERIFICATION,
		Status:             types.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		CsThreshold:        100,
		RequiredIrIds:      []string{"ir-1"},
		RequiredArena:      "main",
		RequiredArenaScore: 10,
		Singleton:          true,
		Version:            "1.0.0",
		CreatedAt:          timestamppb.New(ctx.BlockTime()),
		CreatedHeight:      uint64(ctx.BlockHeight()),
		Creator:            "governance",
	}
	require.NoError(t, k.SetVCPolicy(ctx, policy))

	t.Run("insufficient confidence score", func(t *testing.T) {
		k.SetConfidenceScoreKeeper(testCSKeeper{
			score:        50,
			anchor:       true,
			completedIRs: map[string]bool{"ir-1": true},
			arenaScores:  map[string]uint64{"main": 20},
		})

		resp, err := query.ValidateMintEligibility(sdk.WrapSDKContext(ctx), req)
		require.NoError(t, err)
		require.False(t, resp.Eligible)
		require.True(t, strings.Contains(strings.Join(resp.MissingRequirements, ","), "Insufficient confidence score"))
	})

	t.Run("singleton constraint", func(t *testing.T) {
		k.SetConfidenceScoreKeeper(testCSKeeper{
			score:        150,
			anchor:       true,
			completedIRs: map[string]bool{"ir-1": true},
			arenaScores:  map[string]uint64{"main": 20},
		})

		existingVC := types.VCRecord{
			VcId:          "vc-singleton",
			HolderAddress: holder,
			HolderDid:     "did:aura:holder",
			VcType:        types.VCType_VC_TYPE_KYC_VERIFICATION,
			Status:        types.VCStatus_VC_STATUS_ACTIVE,
			IssuedAt:      timestamppb.New(ctx.BlockTime()),
		}
		require.NoError(t, k.SetVCRecord(ctx, existingVC))

		resp, err := query.ValidateMintEligibility(sdk.WrapSDKContext(ctx), req)
		require.NoError(t, err)
		require.False(t, resp.Eligible)
		require.Contains(t, resp.MissingRequirements, "Singleton VC: user already has an active VC of this type")
	})
}

func TestQueryServerAttributeDisclosureIndexing(t *testing.T) {
	k, ctx := setupMsgServerKeeper(t)
	srv := keeper.NewMsgServer(k)
	query := keeper.NewQueryServer(k)
	holder := "aura1attr"

	_, err := srv.CreateAttributeVC(sdk.WrapSDKContext(ctx), &vcregistrypb.MsgCreateAttributeVC{
		Creator:        holder,
		AttributeType:  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		EncryptedValue: []byte("cipher"),
		Issuer:         "issuer",
	})
	require.NoError(t, err)

	attrResp, err := query.ListAttributeVCs(sdk.WrapSDKContext(ctx), &vcregistrypb.QueryAttributeVCsRequest{HolderAddress: holder})
	require.NoError(t, err)
	require.Len(t, attrResp.AttributeVcs, 1)
	require.Equal(t, holder, attrResp.AttributeVcs[0].HolderAddress)

	reqResp, err := srv.CreateDisclosureRequest(sdk.WrapSDKContext(ctx), &vcregistrypb.MsgCreateDisclosureRequest{
		HolderAddress: holder,
		Verifier:      "verifier1",
		RequestedAttributes: []vcregistrypb.AttributeType{
			vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		},
	})
	require.NoError(t, err)

	dr, err := query.GetDisclosureRequest(sdk.WrapSDKContext(ctx), &vcregistrypb.QueryDisclosureRequestRequest{RequestId: reqResp.RequestId})
	require.NoError(t, err)
	require.True(t, dr.Exists)
	require.Equal(t, reqResp.RequestId, dr.Request.RequestId)
}

func TestQueryServerResolveDIDIncludesMintedCredential(t *testing.T) {
	k, ctx := setupMsgServerKeeper(t)
	srv := keeper.NewMsgServer(k)
	query := keeper.NewQueryServer(k)
	holder := "aura1holder"
	holderDID := "did:aura:holder"

	require.NoError(t, k.RegisterDID(ctx, holderDID, holder, []*types.VerificationMethod{}, ""))

	policy := types.VCPolicy{
		VcTypeName:         fmt.Sprintf("%d", vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION),
		VcTypeEnum:         types.VCType_VC_TYPE_KYC_VERIFICATION,
		Status:             types.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		CsThreshold:        10,
		RequiredIrIds:      []string{"ir-1"},
		RequiredArena:      "main",
		RequiredArenaScore: 0,
		Singleton:          false,
		Version:            "1.0.0",
		CreatedAt:          timestamppb.New(ctx.BlockTime()),
		CreatedHeight:      uint64(ctx.BlockHeight()),
		Creator:            "governance",
	}
	require.NoError(t, k.SetVCPolicy(ctx, policy))

	k.SetConfidenceScoreKeeper(testCSKeeper{
		score:        200,
		anchor:       true,
		completedIRs: map[string]bool{"ir-1": true},
		arenaScores:  map[string]uint64{"main": 100},
	})

	resp, err := srv.MintVC(sdk.WrapSDKContext(ctx), &vcregistrypb.MsgMintVC{
		HolderAddress: holder,
		HolderDid:     holderDID,
		VcType:        vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION,
		Metadata:      map[string]string{"source": "test"},
	})
	require.NoError(t, err)

	resolveResp, err := query.ResolveDID(sdk.WrapSDKContext(ctx), &vcregistrypb.QueryResolveDIDRequest{Did: holderDID})
	require.NoError(t, err)
	require.True(t, resolveResp.Exists)
	require.Len(t, resolveResp.Credentials, 1)
	require.Equal(t, resp.VcId, resolveResp.Credentials[0].VcId)
	require.Contains(t, resolveResp.DidDocument.CredentialIds, resp.VcId)
}
