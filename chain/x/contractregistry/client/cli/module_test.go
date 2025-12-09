package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContractRegistryCLISuite struct {
	suite.Suite
}

func TestContractRegistryCLISuite(t *testing.T) {
	suite.Run(t, new(ContractRegistryCLISuite))
}

func (s *ContractRegistryCLISuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("contractregistry", cmd.Use)
	require.True(cmd.DisableFlagParsing)
}

func (s *ContractRegistryCLISuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("contractregistry", cmd.Use)
	require.True(cmd.DisableFlagParsing)
}
