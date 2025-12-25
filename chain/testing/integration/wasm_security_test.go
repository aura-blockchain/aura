// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/wasm/types"
	contractpb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

// WASMSecurityTestSuite validates the security guardrails enforced by the WASM
// keeper and contract registry integration.
type WASMSecurityTestSuite struct {
	suite.Suite
	ctx WASMTestContext
}

func TestWASMSecurity(t *testing.T) {
	suite.Run(t, new(WASMSecurityTestSuite))
}

func (s *WASMSecurityTestSuite) SetupTest() {
	s.ctx = SetupTestAppWithWasm(s.T())
}

// TestDoSViaRapidExecution simulates a user spamming executions. The per-user
// rate limit should throttle them while leaving other users unaffected.
func (s *WASMSecurityTestSuite) TestDoSViaRapidExecution() {
	uploader := s.ctx.CreateAuthorizedUploader()
	policy := contractpb.SecurityPolicy{
		AllowPause:       true,
		RateLimitPerUser: 1,
	}
	contractAddr := s.ctx.SetupCompleteContractWithPolicy(
		uploader,
		securityMetadata("dos-guard"),
		policy,
		contractpb.ComplianceRequirements{},
	)
	info, found := s.ctx.RegistryKeeper.GetContractInfo(s.ctx.Ctx, contractAddr.String())
	s.Require().True(found)
	s.Require().Equal(uint64(1), info.SecurityPolicy.RateLimitPerUser)

	attacker := sdk.AccAddress([]byte("dos-attacker"))
	s.Require().NoError(s.ctx.ExecuteAsUser(contractAddr, attacker, true, nil))

	s.ctx.AdvanceBlock(time.Second)
	err := s.ctx.ExecuteAsUser(contractAddr, attacker, true, nil)
	s.Require().Error(err, "attacker should be throttled by rate limit")
	s.Contains(err.Error(), "rate limit")

	legitUser := sdk.AccAddress([]byte("legit-user"))
	s.Require().NoError(s.ctx.ExecuteAsUser(contractAddr, legitUser, true, nil), "other users should remain unaffected")
}

// TestUnauthorizedUploadAttempt ensures restricted upload access blocks
// unauthorized code submission.
func (s *WASMSecurityTestSuite) TestUnauthorizedUploadAttempt() {
	uploader := s.ctx.CreateAuthorizedUploader()
	params := s.ctx.WasmKeeper.GetParams(s.ctx.Ctx)
	params.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    uploader.String(),
	}
	err := s.ctx.WasmKeeper.SetParams(s.ctx.Ctx, params)
	s.Require().NoError(err)

	codeBytes := s.ctx.CreateMockWASMCode()
	unauthorized := sdk.AccAddress([]byte("unauthorized-uploader"))
	err = s.ctx.WasmKeeper.ValidateContractUpload(s.ctx.Ctx, unauthorized.String(), codeBytes)
	s.Require().Error(err, "unauthorized uploader should be rejected")
	s.Contains(err.Error(), "not authorized")

	err = s.ctx.WasmKeeper.ValidateContractUpload(s.ctx.Ctx, uploader.String(), codeBytes)
	s.Require().NoError(err, "authorized uploader should be able to submit code")
}

// TestBlacklistedExecutionBlocked verifies the registry's blacklist is enforced
// prior to contract execution.
func (s *WASMSecurityTestSuite) TestBlacklistedExecutionBlocked() {
	uploader := s.ctx.CreateAuthorizedUploader()
	blacklisted := sdk.AccAddress([]byte("blacklisted-user"))
	policy := contractpb.SecurityPolicy{
		AllowPause:           true,
		BlacklistedAddresses: []string{blacklisted.String()},
	}
	contractAddr := s.ctx.SetupCompleteContractWithPolicy(
		uploader,
		securityMetadata("blacklist-guard"),
		policy,
		contractpb.ComplianceRequirements{},
	)

	err := s.ctx.ExecuteAsUser(contractAddr, blacklisted, true, nil)
	s.Require().Error(err, "blacklisted address must be blocked")
	s.Contains(err.Error(), "blacklisted")

	permitted := sdk.AccAddress([]byte("permitted-user"))
	s.Require().NoError(s.ctx.ExecuteAsUser(contractAddr, permitted, true, nil))
}

// defaultMetadata duplicates the helper locally to avoid circular imports.
func securityMetadata(label string) contractpb.ContractMetadata {
	return contractpb.ContractMetadata{
		Name:        label,
		Description: fmt.Sprintf("%s security contract", label),
		Tags:        []string{"wasm", "security"},
	}
}
