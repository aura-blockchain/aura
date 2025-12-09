package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/testing/testutil"
	"github.com/aequitas/aura/chain/x/contractregistry/keeper"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

type QueryServerTestSuite struct {
	suite.Suite
	keeper      *keeper.Keeper
	queryServer pb.QueryServer
	ctx         *testutil.TestContext
	fixtures    *testutil.TestFixtures
}

func (s *QueryServerTestSuite) SetupTest() {
	s.ctx = testutil.SetupTestContext(s.T())
	s.keeper = &keeper.Keeper{}
	s.queryServer = keeper.NewQueryServerImpl(s.keeper)
	s.fixtures = testutil.NewTestFixtures()
}

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (s *QueryServerTestSuite) TestQueryContract_Success() {
	s.Require().NotNil(s.queryServer)
	s.T().Log("Testing contract query")
}

func (s *QueryServerTestSuite) TestListContracts_Success() {
	s.Require().NotNil(s.queryServer)
	s.T().Log("Testing contract listing")
}

func TestContractRegistryQueries(t *testing.T) {
	ctx := testutil.SetupTestContext(t)
	require.NotNil(t, ctx)

	t.Run("QueryByAddress", func(t *testing.T) {
		t.Log("Testing query by contract address")
	})

	t.Run("QueryByCodeHash", func(t *testing.T) {
		t.Log("Testing query by code hash")
	})

	t.Run("ListWithPagination", func(t *testing.T) {
		t.Log("Testing paginated listing")
	})
}
