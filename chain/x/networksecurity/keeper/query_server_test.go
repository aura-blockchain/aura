package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueryServerTestSuite struct {
	KeeperTestSuite
	queryServer interface{}
}

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (suite *QueryServerTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.queryServer = NewQueryServerImpl(suite.Keeper)
}

func (suite *QueryServerTestSuite) TestQueryServerImplementation() {
	suite.NotNil(suite.queryServer, "query server should be created")
}
