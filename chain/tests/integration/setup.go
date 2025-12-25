// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil/keeper"
	"github.com/aequitas/aura/chain/testutil/testdata"
)

// IntegrationTestSuite provides setup for integration tests
type IntegrationTestSuite struct {
	T   *testing.T
	App *keeper.TestApp
	Ctx sdk.Context

	// Test accounts
	Accounts []sdk.AccAddress

	// Test validators
	Validators []sdk.ValAddress
}

// SetupIntegrationTest creates a new integration test suite
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	t.Helper()

	app, ctx := keeper.SetupTestApp(t)

	// Create test accounts
	accounts := keeper.CreateTestAccounts(t, 10)

	suite := &IntegrationTestSuite{
		T:          t,
		App:        app,
		Ctx:        ctx,
		Accounts:   accounts,
		Validators: make([]sdk.ValAddress, 0),
	}

	return suite
}

// SetupIntegrationTestWithValidators creates a suite with validators
func SetupIntegrationTestWithValidators(t *testing.T, numValidators int) *IntegrationTestSuite {
	t.Helper()

	app, ctx, validators := keeper.SetupTestAppWithValidators(t, numValidators)

	// Create test accounts
	accounts := keeper.CreateTestAccounts(t, 10)

	// Extract validator addresses
	valAddrs := make([]sdk.ValAddress, len(validators))
	for i, val := range validators {
		valAddr, err := sdk.ValAddressFromBech32(val.OperatorAddress)
		require.NoError(t, err)
		valAddrs[i] = valAddr
	}

	suite := &IntegrationTestSuite{
		T:          t,
		App:        app,
		Ctx:        ctx,
		Accounts:   accounts,
		Validators: valAddrs,
	}

	return suite
}

// Helper methods for cross-module testing

// CreatePoolAndSwap simulates DEX pool creation and swap
func (s *IntegrationTestSuite) CreatePoolAndSwap(
	creator sdk.AccAddress,
	tokenA, tokenB string,
	amountA, amountB int64,
	swapAmount int64,
) error {
	// This would interact with DEX keeper + Bank keeper
	// Placeholder for now
	return nil
}

// SubmitProposalAndVote simulates governance proposal flow
func (s *IntegrationTestSuite) SubmitProposalAndVote(
	proposer sdk.AccAddress,
	voters []sdk.AccAddress,
	title, description string,
) error {
	// This would interact with Gov keeper + multiple modules
	// Placeholder for now
	return nil
}

// BridgeTransfer simulates cross-chain bridge transfer
func (s *IntegrationTestSuite) BridgeTransfer(
	from sdk.AccAddress,
	to string,
	amount sdk.Coins,
	targetChain string,
) error {
	// This would interact with Bridge keeper + Bank keeper
	// Placeholder for now
	return nil
}

// RegisterAndVerifyIdentity simulates identity registration and verification
func (s *IntegrationTestSuite) RegisterAndVerifyIdentity(
	address sdk.AccAddress,
	credentialData interface{},
) error {
	// This would interact with VCRegistry + IdentityChange + ConfidenceScore
	// Placeholder for now
	return nil
}

// AdvanceBlockHeight advances the block height
func (s *IntegrationTestSuite) AdvanceBlockHeight(n int64) {
	header := s.Ctx.BlockHeader()
	header.Height += n
	s.Ctx = s.Ctx.WithBlockHeader(header)
}

// AdvanceTime advances the block time
func (s *IntegrationTestSuite) AdvanceTime(duration time.Duration) {
	header := s.Ctx.BlockHeader()
	header.Time = header.Time.Add(duration)
	s.Ctx = s.Ctx.WithBlockHeader(header)
}

// FundAccount funds a test account (placeholder)
func (s *IntegrationTestSuite) FundAccount(addr sdk.AccAddress, coins sdk.Coins) error {
	// Would use bank keeper to fund account
	return nil
}

// GetAccount returns a test account by index
func (s *IntegrationTestSuite) GetAccount(index int) sdk.AccAddress {
	require.Less(s.T, index, len(s.Accounts), "account index out of range")
	return s.Accounts[index]
}

// GetValidator returns a validator by index
func (s *IntegrationTestSuite) GetValidator(index int) sdk.ValAddress {
	require.Less(s.T, index, len(s.Validators), "validator index out of range")
	return s.Validators[index]
}

// UseStandardTestAccounts returns predefined test addresses
func (s *IntegrationTestSuite) UseStandardTestAccounts() []sdk.AccAddress {
	return []sdk.AccAddress{
		testdata.TestAddr1,
		testdata.TestAddr2,
		testdata.TestAddr3,
		testdata.TestAddr4,
		testdata.TestAddr5,
	}
}

// Cleanup performs cleanup after tests
func (s *IntegrationTestSuite) Cleanup() {
	// Cleanup resources if needed
}
