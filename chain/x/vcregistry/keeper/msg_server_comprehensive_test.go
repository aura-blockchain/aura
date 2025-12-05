package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

type MsgServerComprehensiveTestSuite struct {
	KeeperTestSuite
	msgServer vcregistrypb.MsgServer
}

func TestMsgServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerComprehensiveTestSuite))
}

func (suite *MsgServerComprehensiveTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.msgServer = NewMsgServer(suite.Keeper)
}

// Helper function to generate valid test addresses
func (suite *MsgServerComprehensiveTestSuite) testAddress(name string) string {
	// Create a 20-byte address and convert to bech32
	addr := sdk.AccAddress([]byte(name + "____________")[:20])
	return addr.String()
}

// ============================
// MintVC Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestMintVCNilRequest() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test nil request handling
	_, err := suite.msgServer.MintVC(ctx, nil)
	suite.Require().Error(err, "nil request should fail")
}

func (suite *MsgServerComprehensiveTestSuite) TestMintVCEmptyHolderAddress() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test empty holder address validation
	msg := &vcregistrypb.MsgMintVC{
		HolderAddress: "",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		HolderDid:     "did:aura:holder1",
	}

	_, err := suite.msgServer.MintVC(ctx, msg)
	suite.Require().Error(err, "empty holder address should fail")
	suite.Require().ErrorIs(err, types.ErrInvalidHolderAddress)
}

func (suite *MsgServerComprehensiveTestSuite) TestMintVCInvalidVCType() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test invalid VC type
	msg := &vcregistrypb.MsgMintVC{
		HolderAddress: "aura1holder123",
		VcType:        vcregistrypb.VCType_VC_TYPE_UNSPECIFIED,
		HolderDid:     "did:aura:holder1",
	}

	_, err := suite.msgServer.MintVC(ctx, msg)
	suite.Require().Error(err, "unspecified VC type should fail")
}

func (suite *MsgServerComprehensiveTestSuite) TestMintVCSuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test successful VC minting
	msg := &vcregistrypb.MsgMintVC{
		HolderAddress: "aura1holder123",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		HolderDid:     "did:aura:holder1",
	}

	resp, err := suite.msgServer.MintVC(ctx, msg)
	suite.Require().NoError(err, "valid mint should succeed")
	suite.Require().NotEmpty(resp.VcId, "response should contain VC ID")

	// Verify VC was created
	vc, ok := suite.Keeper.GetVCRecord(suite.SdkCtx, resp.VcId)
	suite.Require().True(ok, "VC should be retrievable")
	suite.Require().Equal(msg.HolderAddress, vc.HolderAddress)
	suite.Require().Equal(msg.VcType, vc.VcType)
	suite.Require().Equal(types.VCStatus_VC_STATUS_ACTIVE, vc.Status)
}

// ============================
// RevokeVC Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestRevokeVCNonExistent() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test revoking non-existent credential
	msg := &vcregistrypb.MsgRevokeVC{
		HolderAddress: "aura1holder123",
		VcId:          "nonexistent-vc-id",
		ReasonText:    "test",
	}

	_, err := suite.msgServer.RevokeVC(ctx, msg)
	suite.Require().Error(err, "revoking non-existent VC should fail")
	suite.Require().ErrorIs(err, types.ErrVCNotFound)
}

func (suite *MsgServerComprehensiveTestSuite) TestRevokeVCSuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// First create a VC
	mintMsg := &vcregistrypb.MsgMintVC{
		HolderAddress: "aura1holder123",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		HolderDid:     "did:aura:holder1",
	}

	mintResp, err := suite.msgServer.MintVC(ctx, mintMsg)
	suite.Require().NoError(err)

	// Now revoke it
	revokeMsg := &vcregistrypb.MsgRevokeVC{
		HolderAddress: "aura1holder123",
		VcId:          mintResp.VcId,
		ReasonText:    "user requested revocation",
	}

	_, err = suite.msgServer.RevokeVC(ctx, revokeMsg)
	suite.Require().NoError(err, "revocation should succeed")

	// Verify VC is revoked
	vc, ok := suite.Keeper.GetVCRecord(suite.SdkCtx, mintResp.VcId)
	suite.Require().True(ok)
	suite.Require().Equal(types.VCStatus_VC_STATUS_REVOKED, vc.Status)

	// Verify revocation record exists
	revRecord, ok := suite.Keeper.GetRevocationRecord(suite.SdkCtx, mintResp.VcId)
	suite.Require().True(ok)
	suite.Require().Equal(mintResp.VcId, revRecord.VcId)
}

func (suite *MsgServerComprehensiveTestSuite) TestRevokeVCAlreadyRevoked() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Create and revoke a VC
	mintMsg := &vcregistrypb.MsgMintVC{
		HolderAddress: "aura1holder123",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		HolderDid:     "did:aura:holder1",
	}

	mintResp, err := suite.msgServer.MintVC(ctx, mintMsg)
	suite.Require().NoError(err)

	revokeMsg := &vcregistrypb.MsgRevokeVC{
		HolderAddress: "aura1holder123",
		VcId:          mintResp.VcId,
		ReasonText:    "first revocation",
	}

	_, err = suite.msgServer.RevokeVC(ctx, revokeMsg)
	suite.Require().NoError(err)

	// Try to revoke again
	revokeMsg2 := &vcregistrypb.MsgRevokeVC{
		HolderAddress: "aura1holder123",
		VcId:          mintResp.VcId,
		ReasonText:    "second revocation attempt",
	}

	_, err = suite.msgServer.RevokeVC(ctx, revokeMsg2)
	suite.Require().Error(err, "revoking already revoked VC should fail")
	suite.Require().ErrorIs(err, types.ErrVCAlreadyRevoked)
}

// ============================
// CreatePresentation Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestCreatePresentationEmptyVCList() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test creating presentation with no VCs
	msg := &vcregistrypb.MsgCreatePresentation{
		Creator:          "aura1holder123",
		VcIds:            []string{},
		Context:          &vcregistrypb.PresentationContext{ShowFullName: true},
		ExpiresInSeconds: 300,
	}

	_, err := suite.msgServer.CreatePresentation(ctx, msg)
	suite.Require().Error(err, "empty VC list should fail")
	suite.Require().ErrorIs(err, types.ErrEmptyVCList)
}

func (suite *MsgServerComprehensiveTestSuite) TestCreatePresentationSuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// First create a VC
	mintMsg := &vcregistrypb.MsgMintVC{
		HolderAddress: "aura1holder123",
		VcType:        vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		HolderDid:     "did:aura:holder1",
	}

	mintResp, err := suite.msgServer.MintVC(ctx, mintMsg)
	suite.Require().NoError(err)

	// Create presentation
	presentMsg := &vcregistrypb.MsgCreatePresentation{
		Creator:          "aura1holder123",
		VcIds:            []string{mintResp.VcId},
		Context:          &vcregistrypb.PresentationContext{ShowFullName: true, ShowAge: true},
		ExpiresInSeconds: 300,
	}

	resp, err := suite.msgServer.CreatePresentation(ctx, presentMsg)
	suite.Require().NoError(err, "presentation creation should succeed")
	suite.Require().NotEmpty(resp.PresentationId, "response should contain presentation ID")
	suite.Require().NotEmpty(resp.QrCodeData, "response should contain QR code")

	// Verify presentation was stored
	presentation, ok := suite.Keeper.GetPresentation(suite.SdkCtx, resp.PresentationId)
	suite.Require().True(ok, "presentation should be retrievable")
	suite.Require().Equal("aura1holder123", presentation.HolderAddress)
	suite.Require().Contains(presentation.VcIds, mintResp.VcId)
}

// ============================
// RegisterDID Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestRegisterDIDSuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Test successful DID registration
	msg := &vcregistrypb.MsgRegisterDID{
		Controller: "aura1controller",
		Did:        "did:aura:testdid123",
		VerificationMethods: []*vcregistrypb.VerificationMethod{
			{
				Id:        "key-1",
				Type:      "Ed25519VerificationKey2020",
				PublicKey: []byte("test-public-key"),
			},
		},
		MetadataUri: "https://example.com/metadata",
	}

	resp, err := suite.msgServer.RegisterDID(ctx, msg)
	suite.Require().NoError(err, "DID registration should succeed")
	suite.Require().Equal(msg.Did, resp.Did)

	// Verify DID was created
	didDoc, ok := suite.Keeper.GetDIDDocument(suite.SdkCtx, msg.Did)
	suite.Require().True(ok, "DID should be retrievable")
	suite.Require().Equal(msg.Controller, didDoc.Controller)
	suite.Require().Equal(msg.Did, didDoc.Did)
	suite.Require().Len(didDoc.VerificationMethods, 1)
}

func (suite *MsgServerComprehensiveTestSuite) TestRegisterDIDDuplicate() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	// Register a DID
	msg := &vcregistrypb.MsgRegisterDID{
		Controller: "aura1controller",
		Did:        "did:aura:duplicate",
		VerificationMethods: []*vcregistrypb.VerificationMethod{
			{
				Id:        "key-1",
				Type:      "Ed25519VerificationKey2020",
				PublicKey: []byte("test-key"),
			},
		},
	}

	_, err := suite.msgServer.RegisterDID(ctx, msg)
	suite.Require().NoError(err)

	// Try to register again
	_, err = suite.msgServer.RegisterDID(ctx, msg)
	suite.Require().Error(err, "duplicate DID registration should fail")
	suite.Require().ErrorIs(err, types.ErrDIDAlreadyExists)
}

// ============================
// VCPolicy Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestCreateVCPolicySuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &vcregistrypb.MsgCreateVCPolicy{
		Authority:          suite.Keeper.GetAuthority(),
		VcTypeName:         "TestPolicy",
		VcTypeEnum:         vcregistrypb.VCType_VC_TYPE_VERIFIED_HUMAN,
		CsThreshold:        100,
		ExpiryDurationDays: 365,
	}

	resp, err := suite.msgServer.CreateVCPolicy(ctx, msg)
	suite.Require().NoError(err, "policy creation should succeed")
	suite.Require().NotEmpty(resp.PolicyId, "response should contain policy ID")

	// Verify policy was created
	policy, ok := suite.Keeper.GetVCPolicy(suite.SdkCtx, msg.VcTypeName)
	suite.Require().True(ok, "policy should be retrievable")
	suite.Require().Equal(msg.CsThreshold, policy.CsThreshold)
	suite.Require().Equal(vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE, policy.Status)
}

// ============================
// AttributeVC Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestCreateAttributeVCSuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &vcregistrypb.MsgCreateAttributeVC{
		Creator:          "aura1issuer",
		AttributeType:    vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
		EncryptedValue:   []byte("encrypted-age-data"),
		Issuer:           "aura1issuer",
		ExpiresInSeconds: 31536000, // 365 days in seconds
	}

	resp, err := suite.msgServer.CreateAttributeVC(ctx, msg)
	suite.Require().NoError(err, "attribute VC creation should succeed")
	suite.Require().NotEmpty(resp.AttributeVcId, "response should contain attribute VC ID")

	// Verify attribute VC was created
	attrVC, ok := suite.Keeper.GetAttributeVC(suite.SdkCtx, resp.AttributeVcId)
	suite.Require().True(ok, "attribute VC should be retrievable")
	suite.Require().Equal(msg.AttributeType, attrVC.AttributeType)
	suite.Require().Equal(types.VCStatus_VC_STATUS_ACTIVE, attrVC.Status)
}

// ============================
// DisclosurePolicy Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestUpdateDisclosurePolicySuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &vcregistrypb.MsgUpdateDisclosurePolicy{
		Creator:     "aura1holder",
		DefaultMode: vcregistrypb.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY,
		Rules: []*vcregistrypb.AttributeDisclosureRule{
			{
				AttributeType: vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
				Mode:          vcregistrypb.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW,
			},
		},
	}

	resp, err := suite.msgServer.UpdateDisclosurePolicy(ctx, msg)
	suite.Require().NoError(err, "policy update should succeed")
	suite.Require().NotNil(resp.UpdatedAt, "response should contain updated timestamp")

	// Verify policy was set
	policy, ok := suite.Keeper.GetDisclosurePolicy(suite.SdkCtx, msg.Creator)
	suite.Require().True(ok, "policy should be retrievable")
	suite.Require().Equal(msg.DefaultMode, policy.DefaultMode)
	suite.Require().Len(policy.Rules, 1)
}

// ============================
// DisclosureRequest Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestCreateDisclosureRequestSuccess() {
	ctx := sdk.WrapSDKContext(suite.SdkCtx)

	msg := &vcregistrypb.MsgCreateDisclosureRequest{
		Verifier:            "aura1verifier",
		HolderAddress:       "aura1holder",
		VerifierName:        "Test Verifier",
		RequestedAttributes: []vcregistrypb.AttributeType{vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE},
		Purpose:             "age verification",
		ExpiresInSeconds:    600,
	}

	resp, err := suite.msgServer.CreateDisclosureRequest(ctx, msg)
	suite.Require().NoError(err, "disclosure request creation should succeed")
	suite.Require().NotEmpty(resp.RequestId, "response should contain request ID")

	// Verify request was created
	request, ok := suite.Keeper.GetDisclosureRequest(suite.SdkCtx, resp.RequestId)
	suite.Require().True(ok, "request should be retrievable")
	suite.Require().Equal(msg.Verifier, request.VerifierAddress)
	suite.Require().Equal(msg.Purpose, request.Purpose)
}
