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

type PrivacyKeeperTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func (suite *PrivacyKeeperTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // authKeeper
		nil, // bankKeeper
	)
	suite.ctx = input.Ctx
}

func TestPrivacyKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(PrivacyKeeperTestSuite))
}

// Params Tests

func (suite *PrivacyKeeperTestSuite) TestGetParams() {
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().NotNil(params)
}

func (suite *PrivacyKeeperTestSuite) TestSetParams() {
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	retrieved := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(params, retrieved)
}

// Commitment Tests

func TestCreateCommitment(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	sender := keepertest.GenTestAddr().String()
	commitment := []byte("commitment_hash_12345678901234567890")

	commitmentID, err := k.CreateCommitment(input.Ctx, sender, commitment)
	require.NoError(t, err)
	require.NotEmpty(t, commitmentID)
}

func TestGetCommitment(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	sender := keepertest.GenTestAddr().String()
	commitment := []byte("commitment_hash")

	commitmentID, err := k.CreateCommitment(input.Ctx, sender, commitment)
	require.NoError(t, err)

	retrieved, found := k.GetCommitment(input.Ctx, commitmentID)
	require.True(t, found)
	require.Equal(t, commitment, retrieved.Commitment)
}

func TestGetNonExistentCommitment(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	commitment, found := k.GetCommitment(input.Ctx, "nonexistent")
	require.False(t, found)
	require.Nil(t, commitment)
}

func TestVerifyCommitment(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	sender := keepertest.GenTestAddr().String()
	commitment := []byte("commitment_hash")
	secret := []byte("secret_value")

	commitmentID, err := k.CreateCommitment(input.Ctx, sender, commitment)
	require.NoError(t, err)

	valid := k.VerifyCommitment(input.Ctx, commitmentID, secret)
	require.True(t, valid)
}

func TestVerifyCommitmentInvalidSecret(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	sender := keepertest.GenTestAddr().String()
	commitment := []byte("commitment_hash")
	wrongSecret := []byte("wrong_secret")

	commitmentID, err := k.CreateCommitment(input.Ctx, sender, commitment)
	require.NoError(t, err)

	valid := k.VerifyCommitment(input.Ctx, commitmentID, wrongSecret)
	require.False(t, valid)
}

// Nullifier Tests

func TestCreateNullifier(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	nullifier := []byte("nullifier_hash_123")

	err := k.CreateNullifier(input.Ctx, nullifier)
	require.NoError(t, err)
}

func TestNullifierExists(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	nullifier := []byte("nullifier_hash_123")

	err := k.CreateNullifier(input.Ctx, nullifier)
	require.NoError(t, err)

	exists := k.NullifierExists(input.Ctx, nullifier)
	require.True(t, exists)
}

func TestDuplicateNullifier(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	nullifier := []byte("nullifier_hash_123")

	err := k.CreateNullifier(input.Ctx, nullifier)
	require.NoError(t, err)

	// Try to create again
	err = k.CreateNullifier(input.Ctx, nullifier)
	require.Error(t, err, "Should not allow duplicate nullifier")
}

// Merkle Tree Tests

func TestAddLeaf(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	leaf := []byte("leaf_data_123")

	index, err := k.AddLeaf(input.Ctx, leaf)
	require.NoError(t, err)
	require.GreaterOrEqual(t, index, uint64(0))
}

func TestGetLeaf(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	leaf := []byte("leaf_data_123")

	index, err := k.AddLeaf(input.Ctx, leaf)
	require.NoError(t, err)

	retrieved, found := k.GetLeaf(input.Ctx, index)
	require.True(t, found)
	require.Equal(t, leaf, retrieved)
}

func TestGetMerkleRoot(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	// Add several leaves
	for i := 0; i < 10; i++ {
		leaf := []byte("leaf_" + string(rune('A'+i)))
		_, err := k.AddLeaf(input.Ctx, leaf)
		require.NoError(t, err)
	}

	root := k.GetMerkleRoot(input.Ctx)
	require.NotNil(t, root)
	require.NotEmpty(t, root)
}

func TestVerifyMerklePath(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	leaf := []byte("leaf_data")
	index, err := k.AddLeaf(input.Ctx, leaf)
	require.NoError(t, err)

	path := k.GetMerklePath(input.Ctx, index)
	require.NotNil(t, path)

	valid := k.VerifyMerklePath(input.Ctx, leaf, path, index)
	require.True(t, valid)
}

// ZK Proof Tests

func TestSubmitZKProof(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	prover := keepertest.GenTestAddr().String()
	proof := []byte("zk_proof_data")
	publicInputs := []byte("public_inputs")

	proofID, err := k.SubmitZKProof(input.Ctx, prover, proof, publicInputs)
	require.NoError(t, err)
	require.NotEmpty(t, proofID)
}

func TestVerifyZKProof(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	prover := keepertest.GenTestAddr().String()
	proof := []byte("valid_zk_proof")
	publicInputs := []byte("public_inputs")

	proofID, err := k.SubmitZKProof(input.Ctx, prover, proof, publicInputs)
	require.NoError(t, err)

	valid := k.VerifyZKProof(input.Ctx, proofID)
	require.True(t, valid)
}

func TestVerifyInvalidZKProof(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	prover := keepertest.GenTestAddr().String()
	invalidProof := []byte("invalid_proof")
	publicInputs := []byte("public_inputs")

	proofID, err := k.SubmitZKProof(input.Ctx, prover, invalidProof, publicInputs)
	require.NoError(t, err)

	valid := k.VerifyZKProof(input.Ctx, proofID)
	require.False(t, valid)
}

// Shielded Transfer Tests

func TestShieldedTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	sender := keepertest.GenTestAddr().String()
	amount := math.NewInt(1000000)
	commitment := []byte("transfer_commitment")
	proof := []byte("transfer_proof")

	transferID, err := k.ShieldedTransfer(input.Ctx, sender, amount, commitment, proof)
	require.NoError(t, err)
	require.NotEmpty(t, transferID)
}

func TestShieldedTransferZeroAmount(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	sender := keepertest.GenTestAddr().String()
	amount := math.NewInt(0)
	commitment := []byte("transfer_commitment")
	proof := []byte("transfer_proof")

	_, err := k.ShieldedTransfer(input.Ctx, sender, amount, commitment, proof)
	require.Error(t, err, "Should not allow zero amount transfer")
}

func TestGetShieldedTransfer(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	sender := keepertest.GenTestAddr().String()
	amount := math.NewInt(1000000)
	commitment := []byte("transfer_commitment")
	proof := []byte("transfer_proof")

	transferID, err := k.ShieldedTransfer(input.Ctx, sender, amount, commitment, proof)
	require.NoError(t, err)

	transfer, found := k.GetShieldedTransfer(input.Ctx, transferID)
	require.True(t, found)
	require.Equal(t, sender, transfer.Sender)
}

// Unshield Tests

func TestUnshield(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	recipient := keepertest.GenTestAddr().String()
	amount := math.NewInt(1000000)
	nullifier := []byte("nullifier_hash")
	proof := []byte("unshield_proof")

	err := k.Unshield(input.Ctx, recipient, amount, nullifier, proof)
	require.NoError(t, err)

	// Nullifier should now exist
	exists := k.NullifierExists(input.Ctx, nullifier)
	require.True(t, exists)
}

func TestUnshieldDuplicateNullifier(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	recipient := keepertest.GenTestAddr().String()
	amount := math.NewInt(1000000)
	nullifier := []byte("nullifier_hash")
	proof := []byte("unshield_proof")

	err := k.Unshield(input.Ctx, recipient, amount, nullifier, proof)
	require.NoError(t, err)

	// Try to unshield again with same nullifier
	err = k.Unshield(input.Ctx, recipient, amount, nullifier, proof)
	require.Error(t, err, "Should not allow reuse of nullifier")
}

// Genesis Tests

func TestInitGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(input.Ctx, *genesisState)
	require.NoError(t, err)
}

func TestExportGenesis(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	genesisState := types.DefaultGenesis()
	err := k.InitGenesis(input.Ctx, *genesisState)
	require.NoError(t, err)

	exported := k.ExportGenesis(input.Ctx)
	require.NotNil(t, exported)
	require.Equal(t, genesisState.Params, exported.Params)
}
