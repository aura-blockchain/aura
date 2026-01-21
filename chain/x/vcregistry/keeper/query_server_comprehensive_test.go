// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

type QueryServerComprehensiveTestSuite struct {
	KeeperTestSuite
	queryServer vcregistrypb.QueryServer
}

func TestQueryServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerComprehensiveTestSuite))
}

func (suite *QueryServerComprehensiveTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.queryServer = NewQueryServer(suite.Keeper)
}

// ============================
// GetVC Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestGetVCNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test nil request handling
	_, err := suite.queryServer.GetVC(ctx, nil)
	suite.Require().Error(err, "nil request should fail")
}

func (suite *QueryServerComprehensiveTestSuite) TestGetVCEmptyID() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test empty credential ID
	req := &vcregistrypb.QueryGetVCRequest{
		VcId: "",
	}

	_, err := suite.queryServer.GetVC(ctx, req)
	suite.Require().Error(err, "empty VC ID should fail")
	suite.Require().ErrorIs(err, types.ErrInvalidVCID)
}

func (suite *QueryServerComprehensiveTestSuite) TestGetVCNonExistent() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test querying non-existent credential
	req := &vcregistrypb.QueryGetVCRequest{
		VcId: "nonexistent-vc-id",
	}

	_, err := suite.queryServer.GetVC(ctx, req)
	suite.Require().Error(err, "non-existent VC should fail")
	suite.Require().ErrorIs(err, types.ErrVCNotFound)
}

func (suite *QueryServerComprehensiveTestSuite) TestGetVCValid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// First create a VC
	vcRecord := types.VCRecord{
		VcId:          "test-vc-123",
		HolderAddress: "aura1holder",
		HolderDid:     "did:aura:holder1",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	err := suite.Keeper.SetVCRecord(suite.SdkCtx, vcRecord)
	suite.Require().NoError(err)

	// Query the VC
	req := &vcregistrypb.QueryGetVCRequest{
		VcId: "test-vc-123",
	}

	resp, err := suite.queryServer.GetVC(ctx, req)
	suite.Require().NoError(err, "valid VC query should succeed")
	suite.Require().NotNil(resp.Vc, "response should contain VC")
	suite.Require().Equal("test-vc-123", resp.Vc.VcId)
	suite.Require().Equal("aura1holder", resp.Vc.HolderAddress)
}

// ============================
// ListUserVCs Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestListUserVCsNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test nil request
	_, err := suite.queryServer.ListUserVCs(ctx, nil)
	suite.Require().Error(err, "nil request should fail")
}

func (suite *QueryServerComprehensiveTestSuite) TestListUserVCsInvalidAddress() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test invalid address
	req := &vcregistrypb.QueryListUserVCsRequest{
		HolderAddress: "",
	}

	_, err := suite.queryServer.ListUserVCs(ctx, req)
	suite.Require().Error(err, "empty address should fail")
	suite.Require().ErrorIs(err, types.ErrInvalidHolderAddress)
}

func (suite *QueryServerComprehensiveTestSuite) TestListUserVCsEmpty() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test subject with no credentials
	req := &vcregistrypb.QueryListUserVCsRequest{
		HolderAddress: "aura1novcsholder",
	}

	resp, err := suite.queryServer.ListUserVCs(ctx, req)
	suite.Require().NoError(err, "query for user with no VCs should succeed")
	suite.Require().Empty(resp.Vcs, "response should be empty")
}

func (suite *QueryServerComprehensiveTestSuite) TestListUserVCsWithFilters() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	holderAddr := "aura1multivc"

	// Create multiple VCs
	vcActive := types.VCRecord{
		VcId:          "vc-active",
		HolderAddress: holderAddr,
		HolderDid:     "did:aura:multi",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	vcRevoked := types.VCRecord{
		VcId:          "vc-revoked",
		HolderAddress: holderAddr,
		HolderDid:     "did:aura:multi",
		VcType:        vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION,
		Status:        types.VCStatus_VC_STATUS_REVOKED,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}

	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcActive))
	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcRevoked))

	// Query all VCs
	req := &vcregistrypb.QueryListUserVCsRequest{
		HolderAddress: holderAddr,
	}

	resp, err := suite.queryServer.ListUserVCs(ctx, req)
	suite.Require().NoError(err)
	suite.Require().Len(resp.Vcs, 2, "should return all VCs")

	// Query only active VCs
	reqActive := &vcregistrypb.QueryListUserVCsRequest{
		HolderAddress: holderAddr,
		StatusFilter:  vcregistrypb.VCStatus_VC_STATUS_ACTIVE,
	}

	respActive, err := suite.queryServer.ListUserVCs(ctx, reqActive)
	suite.Require().NoError(err)
	suite.Require().Len(respActive.Vcs, 1, "should return only active VCs")
	suite.Require().Equal("vc-active", respActive.Vcs[0].VcId)
}

// ============================
// CheckVCStatus Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestCheckVCStatusValid() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create a VC
	vcRecord := types.VCRecord{
		VcId:          "status-check-vc",
		HolderAddress: "aura1holder",
		HolderDid:     "did:aura:holder",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcRecord))

	// Check status
	req := &vcregistrypb.QueryCheckVCStatusRequest{
		VcId: "status-check-vc",
	}

	resp, err := suite.queryServer.CheckVCStatus(ctx, req)
	suite.Require().NoError(err)
	suite.Require().Equal(vcregistrypb.VCStatus_VC_STATUS_ACTIVE, resp.Status)
	suite.Require().True(resp.Valid, "active VC should be valid")
}

// ============================
// GetVCPolicy Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestGetVCPolicyNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test nil request
	_, err := suite.queryServer.GetVCPolicy(ctx, nil)
	suite.Require().Error(err, "nil request should fail")
}

func (suite *QueryServerComprehensiveTestSuite) TestGetVCPolicyNonExistent() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test querying non-existent policy
	req := &vcregistrypb.QueryGetVCPolicyRequest{
		VcTypeName: "NonExistentPolicy",
	}

	_, err := suite.queryServer.GetVCPolicy(ctx, req)
	suite.Require().Error(err, "non-existent policy should fail")
	suite.Require().ErrorIs(err, types.ErrPolicyNotFound)
}

func (suite *QueryServerComprehensiveTestSuite) TestGetVCPolicySuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create a policy
	policy := vcregistrypb.VCPolicy{
		VcTypeName:         "TestPolicy",
		VcTypeEnum:         vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		CsThreshold:        100,
		ExpiryDurationDays: 365,
		Status:             vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		CreatedAt:          &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	suite.Require().NoError(suite.Keeper.SetVCPolicy(suite.SdkCtx, policy))

	// Query the policy
	req := &vcregistrypb.QueryGetVCPolicyRequest{
		VcTypeName: "TestPolicy",
	}

	resp, err := suite.queryServer.GetVCPolicy(ctx, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Policy)
	suite.Require().Equal("TestPolicy", resp.Policy.VcTypeName)
	suite.Require().Equal(uint64(100), resp.Policy.CsThreshold)
}

// ============================
// ListVCPolicies Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestListVCPoliciesWithPagination() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create multiple policies
	for i := 1; i <= 5; i++ {
		policy := vcregistrypb.VCPolicy{
			VcTypeName:         string(rune('A'+i-1)) + "Policy",
			VcTypeEnum:         vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
			CsThreshold:        uint64(i * 100),
			ExpiryDurationDays: 365,
			Status:             vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
			CreatedAt:          &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
		}
		suite.Require().NoError(suite.Keeper.SetVCPolicy(suite.SdkCtx, policy))
	}

	// Query all policies
	req := &vcregistrypb.QueryListVCPoliciesRequest{}

	resp, err := suite.queryServer.ListVCPolicies(ctx, req)
	suite.Require().NoError(err)
	suite.Require().Len(resp.Policies, 5, "should return all policies")
}

// ============================
// ResolveDID Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestResolveDIDSuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Register a DID
	verificationMethods := []*vcregistrypb.VerificationMethod{
		{
			Id:        "key-1",
			Type:      "Ed25519VerificationKey2020",
			PublicKey: []byte("test-public-key"),
		},
	}

	err := suite.Keeper.RegisterDID(suite.SdkCtx, "did:aura:testdid", "aura1controller", verificationMethods, "https://metadata.uri")
	suite.Require().NoError(err)

	// Resolve DID
	req := &vcregistrypb.QueryResolveDIDRequest{
		Did: "did:aura:testdid",
	}

	resp, err := suite.queryServer.ResolveDID(ctx, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.DidDocument)
	suite.Require().Equal("did:aura:testdid", resp.DidDocument.Did)
	suite.Require().Equal("aura1controller", resp.DidDocument.Controller)
}

func (suite *QueryServerComprehensiveTestSuite) TestResolveDIDNotFound() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Try to resolve non-existent DID
	req := &vcregistrypb.QueryResolveDIDRequest{
		Did: "did:aura:nonexistent",
	}

	_, err := suite.queryServer.ResolveDID(ctx, req)
	suite.Require().Error(err, "non-existent DID should fail")
	suite.Require().ErrorIs(err, types.ErrDIDNotFound)
}

// ============================
// CheckRevocation Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestCheckRevocationNotRevoked() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create an active VC
	vcRecord := types.VCRecord{
		VcId:          "active-vc",
		HolderAddress: "aura1holder",
		HolderDid:     "did:aura:holder",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcRecord))

	// Check revocation
	req := &vcregistrypb.QueryCheckRevocationRequest{
		VcId: "active-vc",
	}

	resp, err := suite.queryServer.CheckRevocation(ctx, req)
	suite.Require().NoError(err)
	suite.Require().False(resp.Revoked, "active VC should not be revoked")
}

func (suite *QueryServerComprehensiveTestSuite) TestCheckRevocationRevoked() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create a VC
	vcRecord := types.VCRecord{
		VcId:          "revoked-vc",
		HolderAddress: "aura1holder",
		HolderDid:     "did:aura:holder",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcRecord))

	// Revoke it
	err := suite.Keeper.RevokeVC(suite.SdkCtx, "revoked-vc", vcregistrypb.RevocationReason_REVOCATION_REASON_USER_REQUEST, "aura1revoker", "test")
	suite.Require().NoError(err)

	// Check revocation
	req := &vcregistrypb.QueryCheckRevocationRequest{
		VcId: "revoked-vc",
	}

	resp, err := suite.queryServer.CheckRevocation(ctx, req)
	suite.Require().NoError(err)
	suite.Require().True(resp.Revoked, "revoked VC should be marked as revoked")
	suite.Require().NotNil(resp.Record)
	suite.Require().Equal("revoked-vc", resp.Record.VcId)
}

func (suite *QueryServerComprehensiveTestSuite) TestCheckRevocationNonExistent() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Check revocation for non-existent VC
	req := &vcregistrypb.QueryCheckRevocationRequest{
		VcId: "nonexistent-vc",
	}

	resp, err := suite.queryServer.CheckRevocation(ctx, req)
	suite.Require().NoError(err)
	suite.Require().False(resp.Revoked, "non-existent VC should return not revoked")
	suite.Require().Nil(resp.Record)
}

// ============================
// Stats Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestStatsEmpty() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Query stats on empty state
	req := &vcregistrypb.QueryStatsRequest{}

	resp, err := suite.queryServer.Stats(ctx, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Equal(uint64(0), resp.TotalVcsMinted)
	suite.Require().Equal(uint64(0), resp.TotalActiveVcs)
}

func (suite *QueryServerComprehensiveTestSuite) TestStatsWithData() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create some VCs
	vcActive1 := types.VCRecord{
		VcId:          "vc-active-1",
		HolderAddress: "aura1holder1",
		HolderDid:     "did:aura:holder1",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	vcActive2 := types.VCRecord{
		VcId:          "vc-active-2",
		HolderAddress: "aura1holder2",
		HolderDid:     "did:aura:holder2",
		VcType:        vcregistrypb.VCType_VC_TYPE_KYC_VERIFICATION,
		Status:        types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}
	vcRevoked := types.VCRecord{
		VcId:          "vc-revoked",
		HolderAddress: "aura1holder3",
		HolderDid:     "did:aura:holder3",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		Status:        types.VCStatus_VC_STATUS_REVOKED,
		IssuedAt:      &gogotypes.Timestamp{Seconds: time.Now().Unix(), Nanos: int32(time.Now().Nanosecond())},
	}

	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcActive1))
	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcActive2))
	suite.Require().NoError(suite.Keeper.SetVCRecord(suite.SdkCtx, vcRevoked))

	// Query stats
	req := &vcregistrypb.QueryStatsRequest{}

	resp, err := suite.queryServer.Stats(ctx, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Equal(uint64(3), resp.TotalVcsMinted)
	suite.Require().Equal(uint64(2), resp.TotalActiveVcs)
	suite.Require().Equal(uint64(1), resp.TotalRevokedVcs)
}

// ============================
// Params Tests
// ============================

func (suite *QueryServerComprehensiveTestSuite) TestParamsQuery() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Query params
	req := &vcregistrypb.QueryParamsRequest{}

	resp, err := suite.queryServer.Params(ctx, req)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp.Params)
	suite.Require().Greater(resp.Params.MaxVcsPerUser, uint64(0), "max VCs per user should be set")
}
