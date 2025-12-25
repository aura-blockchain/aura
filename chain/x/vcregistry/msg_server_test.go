// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package vcregistry

import (
	"fmt"
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

	"github.com/aequitas/aura/chain/x/vcregistry/keeper"
	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

func setupMsgServerKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	k := keeper.NewKeeper(params.NewStore(*types.DefaultParams()), "authority").WithStore(storeKey, cdc)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	return k, ctx
}

type staticCSKeeper struct {
	score uint64
}

func (s staticCSKeeper) GetUserScore(string) (uint64, bool) { return s.score, true }
func (s staticCSKeeper) HasCompletedIR(string, string) bool { return true }
func (s staticCSKeeper) GetArenaScore(string, string) (uint64, error) {
	return 100, nil
}
func (s staticCSKeeper) GetAnchorInfo(string) (interface{}, bool) { return struct{}{}, true }
func (s staticCSKeeper) IsVerified(string) bool                   { return true }

func TestMsgServerMintVCSyncsBlockMetadata(t *testing.T) {
	k, sdkCtx := setupMsgServerKeeper(t)
	blockTime := time.Unix(1_700_000_000, 0)
	sdkCtx = sdkCtx.WithBlockTime(blockTime).WithBlockHeight(42)
	k.SetCurrentTime(blockTime.Unix())
	k.SetCurrentHeight(uint64(sdkCtx.BlockHeight()))
	k.SetConfidenceScoreKeeper(staticCSKeeper{score: 200})

	policy := types.VCPolicy{
		VcTypeName:    fmt.Sprintf("%d", vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION),
		VcTypeEnum:    types.VCType_VC_TYPE_KYC_VERIFICATION,
		Status:        types.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		Version:       "1.0.0",
		CsThreshold:   100,
		RequiredIrIds: []string{},
	}
	require.NoError(t, k.SetVCPolicy(sdkCtx, policy))

	// Generate valid bech32 address
	holderAccAddr := sdk.AccAddress([]byte("holder______________")[:20])
	holderAddr := holderAccAddr.String()
	holderDID := "did:aura:holder"
	require.NoError(t, k.RegisterDID(sdkCtx, holderDID, holderAddr, []*types.VerificationMethod{}, "meta"))

	srv := keeper.NewMsgServer(k)
	resp, err := srv.MintVC(sdk.WrapSDKContext(sdkCtx), &vcregistrypb.MsgMintVC{
		HolderAddress: holderAddr,
		HolderDid:     holderDID,
		VcType:        vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION,
		Metadata:      map[string]string{"source": "test"},
	})
	require.NoError(t, err)
	require.Equal(t, blockTime.Unix(), resp.IssuedAt.Seconds)

	record, ok := k.GetVCRecord(sdkCtx, resp.VcId)
	require.True(t, ok)
	require.Equal(t, uint64(sdkCtx.BlockHeight()), record.IssuedHeight)

	doc, ok := k.GetDIDDocument(sdkCtx, holderDID)
	require.True(t, ok)
	require.Equal(t, blockTime.Unix(), doc.Updated.Seconds)

	// Advance a week and ensure cleanup removes the previous rate-limit counter.
	future := blockTime.Add(8 * 24 * time.Hour)
	k.SetCurrentTime(future.Unix())
	k.CleanupOldMintCounts(sdkCtx)
	require.NoError(t, k.CheckMintRateLimit(sdkCtx, holderAddr))
}

func TestMsgServerDisclosureFlowUpdatesIndices(t *testing.T) {
	k, sdkCtx := setupMsgServerKeeper(t)
	blockTime := time.Unix(1_700_500_000, 0)
	sdkCtx = sdkCtx.WithBlockTime(blockTime).WithBlockHeight(7)
	k.SetCurrentTime(blockTime.Unix())
	srv := keeper.NewMsgServer(k)

	// Generate valid bech32 addresses
	holderAccAddr := sdk.AccAddress([]byte("attr________________")[:20])
	holder := holderAccAddr.String()
	issuerAccAddr := sdk.AccAddress([]byte("issuer______________")[:20])
	issuer := issuerAccAddr.String()
	verifierAccAddr := sdk.AccAddress([]byte("verifier____________")[:20])
	verifier := verifierAccAddr.String()

	// Create an attribute VC
	_, err := srv.CreateAttributeVC(sdk.WrapSDKContext(sdkCtx), &vcregistrypb.MsgCreateAttributeVC{
		Creator:        holder,
		AttributeType:  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		EncryptedValue: []byte("cipher"),
		Issuer:         issuer,
	})
	require.NoError(t, err)

	// Create a disclosure request
	reqResp, err := srv.CreateDisclosureRequest(sdk.WrapSDKContext(sdkCtx), &vcregistrypb.MsgCreateDisclosureRequest{
		HolderAddress: holder,
		Verifier:      verifier,
		VerifierName:  "Verifier One",
		RequestedAttributes: []vcregistrypb.AttributeType{
			vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		},
		ExpiresInSeconds: 900,
	})
	require.NoError(t, err)
	genesis := k.ExportGenesis(sdkCtx)
	require.Contains(t, genesis.PendingDisclosureIndex[holder].Ids, reqResp.RequestId)

	// Respond in a new block to ensure timestamps follow context metadata.
	nextCtx := sdkCtx.WithBlockTime(blockTime.Add(10 * time.Minute)).WithBlockHeight(sdkCtx.BlockHeight() + 1)
	k.SetCurrentTime(nextCtx.BlockTime().Unix())
	resp, err := srv.RespondToDisclosureRequest(sdk.WrapSDKContext(nextCtx), &vcregistrypb.MsgRespondToDisclosureRequest{
		Creator:   holder,
		RequestId: reqResp.RequestId,
		Approved:  true,
	})
	require.NoError(t, err)
	require.Equal(t, nextCtx.BlockTime().Unix(), resp.Response.RespondedAt.Seconds)
	after := k.ExportGenesis(nextCtx)
	if idx := after.PendingDisclosureIndex[holder]; idx != nil {
		require.NotContains(t, idx.Ids, reqResp.RequestId)
	}
}
