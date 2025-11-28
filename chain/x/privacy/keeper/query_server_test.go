package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

type QueryServerTestSuite struct {
	suite.Suite
	keeper      *keeper.Keeper
	ctx         sdk.Context
	queryServer privacypb.QueryServer
}

func (suite *QueryServerTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // authKeeper
		nil, // bankKeeper
	)
	suite.ctx = input.Ctx
	suite.queryServer = keeper.NewQueryServerImpl(suite.keeper)
}

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (suite *QueryServerTestSuite) TestQueryServerImplementation() {
	suite.NotNil(suite.queryServer, "query server should be created")
}

func (suite *QueryServerTestSuite) TestParamsQuery() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test valid params query
	req := &privacypb.QueryParamsRequest{}
	resp, err := suite.queryServer.Params(goCtx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.Params)
}

func (suite *QueryServerTestSuite) TestParamsQueryNilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil request
	resp, err := suite.queryServer.Params(goCtx, nil)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestMixingPoolQuery() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test querying non-existent pool
	req := &privacypb.QueryMixingPoolRequest{
		PoolId: "nonexistent",
	}
	resp, err := suite.queryServer.MixingPool(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestMixingPoolQueryNilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil request
	resp, err := suite.queryServer.MixingPool(goCtx, nil)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestMixingPoolQueryEmptyPoolId() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test empty pool ID
	req := &privacypb.QueryMixingPoolRequest{
		PoolId: "",
	}
	resp, err := suite.queryServer.MixingPool(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestMixingPoolsQuery() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test valid query
	req := &privacypb.QueryMixingPoolsRequest{}
	resp, err := suite.queryServer.MixingPools(goCtx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.MixingPools)
}

func (suite *QueryServerTestSuite) TestMixingPoolsQueryNilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil request
	resp, err := suite.queryServer.MixingPools(goCtx, nil)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestViewKeyQuery() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test querying non-existent view key
	req := &privacypb.QueryViewKeyRequest{
		PublicViewKey: []byte("nonexistent"),
	}
	resp, err := suite.queryServer.ViewKey(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestViewKeyQueryNilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil request
	resp, err := suite.queryServer.ViewKey(goCtx, nil)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestViewKeyQueryEmptyKey() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test empty public view key
	req := &privacypb.QueryViewKeyRequest{
		PublicViewKey: []byte{},
	}
	resp, err := suite.queryServer.ViewKey(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestViewKeysQuery() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Create a test address
	testAddr := keepertest.GenTestAddr().String()

	// Test valid query
	req := &privacypb.QueryViewKeysRequest{
		Address: testAddr,
	}
	resp, err := suite.queryServer.ViewKeys(goCtx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.NotNil(resp.ViewKeys)
}

func (suite *QueryServerTestSuite) TestViewKeysQueryNilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil request
	resp, err := suite.queryServer.ViewKeys(goCtx, nil)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestViewKeysQueryEmptyAddress() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test empty address
	req := &privacypb.QueryViewKeysRequest{
		Address: "",
	}
	resp, err := suite.queryServer.ViewKeys(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestVerifyZKProofQuery() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test with valid ZK proof structure
	req := &privacypb.QueryVerifyZKProofRequest{
		ZkProof: &privacypb.ZKProof{
			ProofType:       "GROTH16",
			ProofData:       []byte("test_proof"),
			PublicInputs:    [][]byte{[]byte("test_input_1"), []byte("test_input_2")},
			VerificationKey: []byte("verification_key"),
			CircuitId:       "test_circuit",
		},
	}
	resp, err := suite.queryServer.VerifyZKProof(goCtx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	// Verification may fail, but response should be valid
}

func (suite *QueryServerTestSuite) TestVerifyZKProofQueryNilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil request
	resp, err := suite.queryServer.VerifyZKProof(goCtx, nil)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestVerifyZKProofQueryNilProof() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil proof
	req := &privacypb.QueryVerifyZKProofRequest{
		ZkProof: nil,
	}
	resp, err := suite.queryServer.VerifyZKProof(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestDecryptWithViewKeyQuery() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test with valid data
	req := &privacypb.QueryDecryptWithViewKeyRequest{
		EncryptedData:  []byte("encrypted_test_data"),
		PrivateViewKey: []byte("private_view_key"),
	}
	resp, err := suite.queryServer.DecryptWithViewKey(goCtx, req)
	suite.NoError(err)
	suite.NotNil(resp)
	// Decryption may fail, but response should be valid
}

func (suite *QueryServerTestSuite) TestDecryptWithViewKeyQueryNilRequest() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test nil request
	resp, err := suite.queryServer.DecryptWithViewKey(goCtx, nil)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestDecryptWithViewKeyQueryEmptyData() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test empty encrypted data
	req := &privacypb.QueryDecryptWithViewKeyRequest{
		EncryptedData:  []byte{},
		PrivateViewKey: []byte("private_view_key"),
	}
	resp, err := suite.queryServer.DecryptWithViewKey(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}

func (suite *QueryServerTestSuite) TestDecryptWithViewKeyQueryEmptyViewKey() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// Test empty private view key
	req := &privacypb.QueryDecryptWithViewKeyRequest{
		EncryptedData:  []byte("encrypted_test_data"),
		PrivateViewKey: []byte{},
	}
	resp, err := suite.queryServer.DecryptWithViewKey(goCtx, req)
	suite.Error(err)
	suite.Nil(resp)
}
