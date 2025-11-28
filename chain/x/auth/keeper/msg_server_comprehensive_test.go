package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/auth/keeper"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

type AuthMsgServerTestSuite struct {
	suite.Suite
	keeper    *keeper.Keeper
	msgServer authproto.MsgServer
	ctx       context.Context
	fixtures  *testutil.TestFixtures
}

func (s *AuthMsgServerTestSuite) SetupTest() {
	testCtx := testutil.SetupTestContext(s.T())
	s.ctx = testCtx.Ctx
	// Note: keeper initialization would require proper setup
	s.keeper = &keeper.Keeper{}
	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	s.fixtures = testutil.NewTestFixtures()
}

func TestAuthMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthMsgServerTestSuite))
}

// Role management tests
func (s *AuthMsgServerTestSuite) TestCreateRole_Success() {
	msg := &authproto.MsgCreateRole{
		Creator:     s.fixtures.Addresses[0].String(),
		Name:        "admin",
		Permissions: []string{"read", "write", "delete"},
		Description: "Administrator role",
	}

	resp, err := s.msgServer.CreateRole(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Role)
	s.Require().Equal("admin", resp.Role.Name)
}

func (s *AuthMsgServerTestSuite) TestCreateRole_EmptyName() {
	msg := &authproto.MsgCreateRole{
		Creator:     s.fixtures.Addresses[0].String(),
		Name:        "",
		Permissions: []string{"read"},
	}

	_, err := s.msgServer.CreateRole(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "role name cannot be empty")
}

func (s *AuthMsgServerTestSuite) TestCreateRole_EmptyPermissions() {
	msg := &authproto.MsgCreateRole{
		Creator:     s.fixtures.Addresses[0].String(),
		Name:        "testrole",
		Permissions: []string{},
	}

	_, err := s.msgServer.CreateRole(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "permissions cannot be empty")
}

func (s *AuthMsgServerTestSuite) TestAssignRole_Success() {
	msg := &authproto.MsgAssignRole{
		Assigner:         s.fixtures.Addresses[0].String(),
		Address:          s.fixtures.Addresses[1].String(),
		RoleName:         "admin",
		ExpiresInSeconds: 86400,
	}

	resp, err := s.msgServer.AssignRole(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Assignment)
}

func (s *AuthMsgServerTestSuite) TestRevokeRole_Success() {
	msg := &authproto.MsgRevokeRole{
		Revoker:  s.fixtures.Addresses[0].String(),
		Address:  s.fixtures.Addresses[1].String(),
		RoleName: "admin",
	}

	resp, err := s.msgServer.RevokeRole(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)
}

// Multisig wallet tests
func (s *AuthMsgServerTestSuite) TestCreateMultisigWallet_Success() {
	msg := &authproto.MsgCreateMultisigWallet{
		Creator:    s.fixtures.Addresses[0].String(),
		Signers:    []string{s.fixtures.Addresses[1].String(), s.fixtures.Addresses[2].String()},
		Threshold:  2,
		WalletType: authproto.MultisigWalletType_STANDARD,
	}

	resp, err := s.msgServer.CreateMultisigWallet(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Wallet)
}

func (s *AuthMsgServerTestSuite) TestCreateMultisigWallet_InvalidThreshold() {
	msg := &authproto.MsgCreateMultisigWallet{
		Creator:    s.fixtures.Addresses[0].String(),
		Signers:    []string{s.fixtures.Addresses[1].String()},
		Threshold:  0,
		WalletType: authproto.MultisigWalletType_STANDARD,
	}

	_, err := s.msgServer.CreateMultisigWallet(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "threshold must be greater than 0")
}

func (s *AuthMsgServerTestSuite) TestCreateMultisigProposal_Success() {
	msg := &authproto.MsgCreateMultisigProposal{
		Proposer:         s.fixtures.Addresses[0].String(),
		WalletId:         "wallet123",
		Title:            "Test Proposal",
		Description:      "Test Description",
		Payload:          []byte("test payload"),
		ExpiresInSeconds: 3600,
	}

	resp, err := s.msgServer.CreateMultisigProposal(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Proposal)
}

// Time-locked actions tests
func (s *AuthMsgServerTestSuite) TestProposeTimeLockedAction_Success() {
	msg := &authproto.MsgProposeTimeLockedAction{
		Proposer:     s.fixtures.Addresses[0].String(),
		ActionType:   "upgrade",
		Payload:      []byte("upgrade data"),
		DelaySeconds: 86400,
	}

	resp, err := s.msgServer.ProposeTimeLockedAction(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Action)
}

func (s *AuthMsgServerTestSuite) TestProposeTimeLockedAction_ZeroDelay() {
	msg := &authproto.MsgProposeTimeLockedAction{
		Proposer:     s.fixtures.Addresses[0].String(),
		ActionType:   "upgrade",
		Payload:      []byte("upgrade data"),
		DelaySeconds: 0,
	}

	_, err := s.msgServer.ProposeTimeLockedAction(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "delay seconds must be greater than 0")
}

// Emergency admin tests
func (s *AuthMsgServerTestSuite) TestActivateEmergencyAdmin_Success() {
	msg := &authproto.MsgActivateEmergencyAdmin{
		Activator:        s.fixtures.Addresses[0].String(),
		AdminAddress:     s.fixtures.Addresses[1].String(),
		Privileges:       []string{"halt_chain", "emergency_upgrade"},
		ExpiresInSeconds: 3600,
	}

	resp, err := s.msgServer.ActivateEmergencyAdmin(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Admin)
}

func (s *AuthMsgServerTestSuite) TestDeactivateEmergencyAdmin_Success() {
	msg := &authproto.MsgDeactivateEmergencyAdmin{
		Deactivator:  s.fixtures.Addresses[0].String(),
		AdminAddress: s.fixtures.Addresses[1].String(),
	}

	resp, err := s.msgServer.DeactivateEmergencyAdmin(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)
}

// Session management tests
func (s *AuthMsgServerTestSuite) TestCreateSession_Success() {
	msg := &authproto.MsgCreateSession{
		UserAddress: s.fixtures.Addresses[0].String(),
	}

	resp, err := s.msgServer.CreateSession(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Session)
}

func (s *AuthMsgServerTestSuite) TestRevokeSession_Success() {
	msg := &authproto.MsgRevokeSession{
		UserAddress: s.fixtures.Addresses[0].String(),
		SessionId:   "session123",
	}

	resp, err := s.msgServer.RevokeSession(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().True(resp.Success)
}

// Validator key rotation tests
func (s *AuthMsgServerTestSuite) TestInitiateValidatorKeyRotation_Success() {
	msg := &authproto.MsgInitiateValidatorKeyRotation{
		Initiator:          s.fixtures.Addresses[0].String(),
		ValidatorAddress:   s.fixtures.ValidatorAddrs[0].String(),
		NewConsensusPubkey: "new_pubkey",
	}

	resp, err := s.msgServer.InitiateValidatorKeyRotation(s.ctx, msg)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Rotation)
}

// Nil request tests
func (s *AuthMsgServerTestSuite) TestNilRequests() {
	_, err := s.msgServer.CreateRole(s.ctx, nil)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "empty request")

	_, err = s.msgServer.AssignRole(s.ctx, nil)
	s.Require().Error(err)

	_, err = s.msgServer.CreateMultisigWallet(s.ctx, nil)
	s.Require().Error(err)

	_, err = s.msgServer.CreateSession(s.ctx, nil)
	s.Require().Error(err)
}
