// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
	"github.com/aequitas/aura/chain/x/privacy/types"
)

type ConfidentialTransactionsTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func TestConfidentialTransactionsTestSuite(t *testing.T) {
	suite.Run(t, new(ConfidentialTransactionsTestSuite))
}

func (suite *ConfidentialTransactionsTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
	)
	suite.ctx = input.Ctx

	// Enable confidential transactions
	params := types.DefaultParams()
	params.EnableConfidentialTransactions = true
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)
}

// ValidateConfidentialTransaction Tests

func (suite *ConfidentialTransactionsTestSuite) TestValidateConfidentialTransaction_NotEnabled() {
	// Disable confidential transactions
	params := suite.keeper.GetParams(suite.ctx)
	params.EnableConfidentialTransactions = false
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte("input1")},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}

	valid, err := suite.keeper.ValidateConfidentialTransaction(suite.ctx, ctxTx)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "confidential transactions not enabled")
}

func (suite *ConfidentialTransactionsTestSuite) TestValidateConfidentialTransaction_BalanceVerificationFailed() {
	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{}, // Empty inputs
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}

	valid, err := suite.keeper.ValidateConfidentialTransaction(suite.ctx, ctxTx)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "balance verification failed")
}

func (suite *ConfidentialTransactionsTestSuite) TestValidateConfidentialTransaction_RangeProofFailed() {
	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte("input1")},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("invalid_proof"), // Invalid proof
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}

	// Create input commitments first
	sender := keepertest.GenTestAddr().String()
	_, err := suite.keeper.CreateCommitment(suite.ctx, sender, []byte("input1"))
	suite.Require().NoError(err)

	valid, err := suite.keeper.ValidateConfidentialTransaction(suite.ctx, ctxTx)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "range proof verification failed")
}

func (suite *ConfidentialTransactionsTestSuite) TestValidateConfidentialTransaction_InputNotFound() {
	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte("nonexistent_input")},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}

	valid, err := suite.keeper.ValidateConfidentialTransaction(suite.ctx, ctxTx)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "input commitment not found")
}

func (suite *ConfidentialTransactionsTestSuite) TestValidateConfidentialTransaction_InputAlreadySpent() {
	// Create and spend a commitment
	sender := keepertest.GenTestAddr().String()
	commitment := []byte("input1")
	commitmentID, err := suite.keeper.CreateCommitment(suite.ctx, sender, commitment)
	suite.Require().NoError(err)

	// Mark it as spent
	ctxTx1 := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte(commitmentID)},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}
	err = suite.keeper.StoreConfidentialTx(suite.ctx, "tx1", ctxTx1)
	suite.Require().NoError(err)

	// Try to spend it again
	ctxTx2 := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte(commitmentID)},
		OutputCommitments: [][]byte{[]byte("output2")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}

	valid, err := suite.keeper.ValidateConfidentialTransaction(suite.ctx, ctxTx2)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "already spent")
}

func (suite *ConfidentialTransactionsTestSuite) TestValidateConfidentialTransaction_InvalidSignature() {
	sender := keepertest.GenTestAddr().String()
	commitment := []byte("input1")
	commitmentID, err := suite.keeper.CreateCommitment(suite.ctx, sender, commitment)
	suite.Require().NoError(err)

	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte(commitmentID)},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("invalid_signature"), // Invalid signature
		Fee:               math.NewInt(1000),
	}

	valid, err := suite.keeper.ValidateConfidentialTransaction(suite.ctx, ctxTx)

	suite.Require().Error(err)
	suite.Require().False(valid)
	suite.Require().Contains(err.Error(), "signature verification failed")
}

func (suite *ConfidentialTransactionsTestSuite) TestValidateConfidentialTransaction_Success() {
	sender := keepertest.GenTestAddr().String()
	commitment := []byte("input1")
	commitmentID, err := suite.keeper.CreateCommitment(suite.ctx, sender, commitment)
	suite.Require().NoError(err)

	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte(commitmentID)},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("valid_proof"),
		Signature:         []byte("valid_signature"),
		Fee:               math.NewInt(1000),
	}

	valid, err := suite.keeper.ValidateConfidentialTransaction(suite.ctx, ctxTx)

	suite.Require().NoError(err)
	suite.Require().True(valid)
}

// VerifyBalance Tests

func TestVerifyBalance(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	testCases := []struct {
		name     string
		ctxTx    *keeper.ConfidentialTransaction
		expected bool
	}{
		{
			name: "Valid with inputs and outputs",
			ctxTx: &keeper.ConfidentialTransaction{
				InputCommitments:  [][]byte{[]byte("input1")},
				OutputCommitments: [][]byte{[]byte("output1")},
			},
			expected: true,
		},
		{
			name: "No inputs",
			ctxTx: &keeper.ConfidentialTransaction{
				InputCommitments:  [][]byte{},
				OutputCommitments: [][]byte{[]byte("output1")},
			},
			expected: false,
		},
		{
			name: "No outputs",
			ctxTx: &keeper.ConfidentialTransaction{
				InputCommitments:  [][]byte{[]byte("input1")},
				OutputCommitments: [][]byte{},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := k.VerifyBalance(tc.ctxTx)
			require.Equal(t, tc.expected, result)
		})
	}
}

// VerifyRangeProof Tests

func TestVerifyRangeProof(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	testCases := []struct {
		name        string
		proof       []byte
		commitments [][]byte
		expected    bool
	}{
		{
			name:        "Valid proof",
			proof:       []byte("valid_proof"),
			commitments: [][]byte{[]byte("commitment1")},
			expected:    true,
		},
		{
			name:        "Empty proof",
			proof:       []byte{},
			commitments: [][]byte{[]byte("commitment1")},
			expected:    false,
		},
		{
			name:        "Invalid proof marker",
			proof:       []byte("invalid_proof"),
			commitments: [][]byte{[]byte("commitment1")},
			expected:    false,
		},
		{
			name:        "No commitments",
			proof:       []byte("proof"),
			commitments: [][]byte{},
			expected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := k.VerifyRangeProof(input.Ctx, tc.proof, tc.commitments)
			require.Equal(t, tc.expected, result)
		})
	}
}

// VerifyConfidentialSignature Tests

func TestVerifyConfidentialSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	testCases := []struct {
		name      string
		signature []byte
		expected  bool
	}{
		{
			name:      "Valid signature",
			signature: []byte("valid_signature"),
			expected:  true,
		},
		{
			name:      "Empty signature",
			signature: []byte{},
			expected:  false,
		},
		{
			name:      "Invalid signature marker",
			signature: []byte("invalid_signature"),
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxTx := &keeper.ConfidentialTransaction{
				Signature: tc.signature,
			}
			result := k.VerifyConfidentialSignature(ctxTx)
			require.Equal(t, tc.expected, result)
		})
	}
}

// Pedersen Commitment Tests

func TestCreatePedersenCommitment(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	value := math.NewInt(1000)
	blindingFactor := []byte("random_blinding_factor")

	commitment := k.CreatePedersenCommitment(value, blindingFactor)

	require.NotNil(t, commitment)
	require.NotEmpty(t, commitment)
	require.Equal(t, 32, len(commitment)) // SHA256 hash
}

func TestVerifyPedersenCommitment(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	value := math.NewInt(1000)
	blindingFactor := []byte("random_blinding_factor")

	commitment := k.CreatePedersenCommitment(value, blindingFactor)

	// Verify with correct value and blinding factor
	valid := k.VerifyPedersenCommitment(commitment, value, blindingFactor)
	require.True(t, valid)

	// Verify with wrong value
	wrongValue := math.NewInt(2000)
	valid = k.VerifyPedersenCommitment(commitment, wrongValue, blindingFactor)
	require.False(t, valid)

	// Verify with wrong blinding factor
	wrongBlindingFactor := []byte("wrong_blinding_factor")
	valid = k.VerifyPedersenCommitment(commitment, value, wrongBlindingFactor)
	require.False(t, valid)
}

// GenerateRangeProof Tests

func TestGenerateRangeProof(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	values := []math.Int{math.NewInt(100), math.NewInt(200)}
	blindingFactors := [][]byte{[]byte("blinding1"), []byte("blinding2")}

	proof, err := k.GenerateRangeProof(values, blindingFactors)

	require.NoError(t, err)
	require.NotNil(t, proof)
	require.NotEmpty(t, proof)
}

func TestGenerateRangeProof_MismatchedArrays(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	values := []math.Int{math.NewInt(100), math.NewInt(200)}
	blindingFactors := [][]byte{[]byte("blinding1")} // Mismatched length

	proof, err := k.GenerateRangeProof(values, blindingFactors)

	require.Error(t, err)
	require.Nil(t, proof)
	require.Contains(t, err.Error(), "mismatched")
}

// AggregateCommitments Tests

func TestAggregateCommitments(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	commitment1 := []byte{0x01, 0x02, 0x03, 0x04}
	commitment2 := []byte{0x05, 0x06, 0x07, 0x08}

	aggregated := k.AggregateCommitments([][]byte{commitment1, commitment2})

	require.NotNil(t, aggregated)
	require.NotEmpty(t, aggregated)
	require.Equal(t, len(commitment1), len(aggregated))
}

func TestAggregateCommitments_Empty(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	aggregated := k.AggregateCommitments([][]byte{})

	require.Nil(t, aggregated)
}

func TestAggregateCommitments_SingleCommitment(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	commitment := []byte{0x01, 0x02, 0x03, 0x04}

	aggregated := k.AggregateCommitments([][]byte{commitment})

	require.NotNil(t, aggregated)
	require.Equal(t, commitment, aggregated)
}

// StoreConfidentialTx Tests

func TestStoreConfidentialTx(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{[]byte("input1")},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}

	err := k.StoreConfidentialTx(input.Ctx, "tx_123", ctxTx)
	require.NoError(t, err)

	// Verify tx was stored
	retrieved, err := k.GetConfidentialTx(input.Ctx, "tx_123")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	// Verify input commitments are marked as spent
	isSpent := k.IsCommitmentSpent(input.Ctx, []byte("input1"))
	require.True(t, isSpent)
}

// GetConfidentialTx Tests

func TestGetConfidentialTx_NotFound(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	retrieved, err := k.GetConfidentialTx(input.Ctx, "nonexistent_tx")

	require.Error(t, err)
	require.Nil(t, retrieved)
	require.Contains(t, err.Error(), "not found")
}

// IsCommitmentSpent Tests

func TestIsCommitmentSpent(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	commitment := []byte("commitment1")

	// Should not be spent initially
	isSpent := k.IsCommitmentSpent(input.Ctx, commitment)
	require.False(t, isSpent)

	// Store a transaction that spends it
	ctxTx := &keeper.ConfidentialTransaction{
		InputCommitments:  [][]byte{commitment},
		OutputCommitments: [][]byte{[]byte("output1")},
		RangeProof:        []byte("proof"),
		Signature:         []byte("signature"),
		Fee:               math.NewInt(1000),
	}

	err := k.StoreConfidentialTx(input.Ctx, "tx_123", ctxTx)
	require.NoError(t, err)

	// Should be spent now
	isSpent = k.IsCommitmentSpent(input.Ctx, commitment)
	require.True(t, isSpent)
}
