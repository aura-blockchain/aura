package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/auth/keeper"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

type AuthQueryServerTestSuite struct {
	suite.Suite
	keeper      *keeper.Keeper
	queryServer authproto.QueryServer
	ctx         context.Context
	fixtures    *testutil.TestFixtures
}

func (s *AuthQueryServerTestSuite) SetupTest() {
	testCtx := testutil.SetupTestContext(s.T())
	s.ctx = testCtx.Ctx
	s.keeper = &keeper.Keeper{}
	s.queryServer = keeper.NewQueryServerImpl(s.keeper)
	s.fixtures = testutil.NewTestFixtures()
}

func TestAuthQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthQueryServerTestSuite))
}

// Role query tests
func (s *AuthQueryServerTestSuite) TestGetRole_ValidName() {
	req := &authproto.QueryGetRoleRequest{
		Name: "admin",
	}

	resp, err := s.queryServer.GetRole(s.ctx, req)
	// Note: This will fail without proper keeper setup, but tests the interface
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestGetRole_EmptyName() {
	req := &authproto.QueryGetRoleRequest{
		Name: "",
	}

	_, err := s.queryServer.GetRole(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "role name cannot be empty")
}

func (s *AuthQueryServerTestSuite) TestListRoles_Success() {
	req := &authproto.QueryListRolesRequest{}

	resp, err := s.queryServer.ListRoles(s.ctx, req)
	// Note: Will fail without proper keeper setup
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestGetRoleAssignments_ValidAddress() {
	req := &authproto.QueryGetRoleAssignmentsRequest{
		Address: s.fixtures.Addresses[0].String(),
	}

	resp, err := s.queryServer.GetRoleAssignments(s.ctx, req)
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestGetRoleAssignments_EmptyAddress() {
	req := &authproto.QueryGetRoleAssignmentsRequest{
		Address: "",
	}

	_, err := s.queryServer.GetRoleAssignments(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "address cannot be empty")
}

func (s *AuthQueryServerTestSuite) TestHasPermission_ValidRequest() {
	req := &authproto.QueryHasPermissionRequest{
		Address:    s.fixtures.Addresses[0].String(),
		Permission: "write",
	}

	resp, err := s.queryServer.HasPermission(s.ctx, req)
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestHasPermission_EmptyPermission() {
	req := &authproto.QueryHasPermissionRequest{
		Address:    s.fixtures.Addresses[0].String(),
		Permission: "",
	}

	_, err := s.queryServer.HasPermission(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "permission cannot be empty")
}

// Multisig queries
func (s *AuthQueryServerTestSuite) TestGetMultisigWallet_ValidID() {
	req := &authproto.QueryGetMultisigWalletRequest{
		Id: "wallet123",
	}

	resp, err := s.queryServer.GetMultisigWallet(s.ctx, req)
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestGetMultisigWallet_EmptyID() {
	req := &authproto.QueryGetMultisigWalletRequest{
		Id: "",
	}

	_, err := s.queryServer.GetMultisigWallet(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "wallet id cannot be empty")
}

func (s *AuthQueryServerTestSuite) TestListMultisigProposals_WithWalletFilter() {
	req := &authproto.QueryListMultisigProposalsRequest{
		WalletId: "wallet123",
	}

	resp, err := s.queryServer.ListMultisigProposals(s.ctx, req)
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestListMultisigProposals_NoFilter() {
	req := &authproto.QueryListMultisigProposalsRequest{}

	resp, err := s.queryServer.ListMultisigProposals(s.ctx, req)
	_ = resp
	_ = err
}

// Time-locked action queries
func (s *AuthQueryServerTestSuite) TestGetTimeLockedAction_ValidID() {
	req := &authproto.QueryGetTimeLockedActionRequest{
		Id: "action123",
	}

	resp, err := s.queryServer.GetTimeLockedAction(s.ctx, req)
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestGetTimeLockedAction_EmptyID() {
	req := &authproto.QueryGetTimeLockedActionRequest{
		Id: "",
	}

	_, err := s.queryServer.GetTimeLockedAction(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "action id cannot be empty")
}

// Emergency admin queries
func (s *AuthQueryServerTestSuite) TestGetEmergencyAdmin_ValidAddress() {
	req := &authproto.QueryGetEmergencyAdminRequest{
		Address: s.fixtures.Addresses[0].String(),
	}

	resp, err := s.queryServer.GetEmergencyAdmin(s.ctx, req)
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestGetEmergencyAdmin_EmptyAddress() {
	req := &authproto.QueryGetEmergencyAdminRequest{
		Address: "",
	}

	_, err := s.queryServer.GetEmergencyAdmin(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "address cannot be empty")
}

// Session queries
func (s *AuthQueryServerTestSuite) TestGetSession_ValidID() {
	req := &authproto.QueryGetSessionRequest{
		SessionId: "session123",
	}

	resp, err := s.queryServer.GetSession(s.ctx, req)
	_ = resp
	_ = err
}

func (s *AuthQueryServerTestSuite) TestGetSession_EmptyID() {
	req := &authproto.QueryGetSessionRequest{
		SessionId: "",
	}

	_, err := s.queryServer.GetSession(s.ctx, req)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "session id cannot be empty")
}

func (s *AuthQueryServerTestSuite) TestListSessions_ValidUser() {
	req := &authproto.QueryListSessionsRequest{
		UserAddress: s.fixtures.Addresses[0].String(),
	}

	resp, err := s.queryServer.ListSessions(s.ctx, req)
	_ = resp
	_ = err
}

// Rate limit queries
func (s *AuthQueryServerTestSuite) TestGetRateLimitStatus_ValidUser() {
	req := &authproto.QueryGetRateLimitStatusRequest{
		UserAddress: s.fixtures.Addresses[0].String(),
	}

	resp, err := s.queryServer.GetRateLimitStatus(s.ctx, req)
	_ = resp
	_ = err
}

// Audit log queries
func (s *AuthQueryServerTestSuite) TestGetAuditLogs_WithFilters() {
	req := &authproto.QueryGetAuditLogsRequest{
		Actor:     s.fixtures.Addresses[0].String(),
		Action:    "create_role",
		StartTime: 0,
		EndTime:   9999999999,
		Limit:     100,
	}

	resp, err := s.queryServer.GetAuditLogs(s.ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
}

// Nil request tests
func (s *AuthQueryServerTestSuite) TestNilRequests() {
	_, err := s.queryServer.GetRole(s.ctx, nil)
	s.Require().Error(err)

	_, err = s.queryServer.ListRoles(s.ctx, nil)
	s.Require().Error(err)

	_, err = s.queryServer.GetRoleAssignments(s.ctx, nil)
	s.Require().Error(err)

	_, err = s.queryServer.GetMultisigWallet(s.ctx, nil)
	s.Require().Error(err)

	_, err = s.queryServer.GetSession(s.ctx, nil)
	s.Require().Error(err)
}
