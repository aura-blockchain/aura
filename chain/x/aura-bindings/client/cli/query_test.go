package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

type QueryTestSuite struct {
	suite.Suite
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

// Test command registration
func (s *QueryTestSuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal(types.ModuleName, cmd.Use)
	require.Equal("Querying commands for the aura-bindings module", cmd.Short)

	// Verify all expected subcommands are registered
	subCmds := cmd.Commands()
	require.Equal(3, len(subCmds))

	cmdNames := make(map[string]bool)
	for _, subCmd := range subCmds {
		cmdNames[subCmd.Name()] = true
	}

	expectedCmds := []string{"query-stats", "message-stats", "all-stats"}
	for _, expectedCmd := range expectedCmds {
		require.True(cmdNames[expectedCmd], "command %s should be registered", expectedCmd)
	}
}

// Test GetCmdQueryStats
func (s *QueryTestSuite) TestGetCmdQueryStats() {
	cmd := GetCmdQueryStats()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("query-stats", cmd.Use)
	require.Equal("Query the query usage statistics", cmd.Short)

	// Test that Args validator accepts 0 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{})
	require.NoError(err)
}

func (s *QueryTestSuite) TestGetCmdQueryStats_RejectsArgs() {
	cmd := GetCmdQueryStats()
	require := s.Require()

	// Test that extra args are rejected
	err := cmd.Args(cmd, []string{"extra"})
	require.Error(err)
}

// Test GetCmdMessageStats
func (s *QueryTestSuite) TestGetCmdMessageStats() {
	cmd := GetCmdMessageStats()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("message-stats", cmd.Use)
	require.Equal("Query the message usage statistics", cmd.Short)

	// Test that Args validator accepts 0 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{})
	require.NoError(err)
}

func (s *QueryTestSuite) TestGetCmdMessageStats_RejectsArgs() {
	cmd := GetCmdMessageStats()
	require := s.Require()

	// Test that extra args are rejected
	err := cmd.Args(cmd, []string{"extra"})
	require.Error(err)
}

// Test GetCmdAllStats
func (s *QueryTestSuite) TestGetCmdAllStats() {
	cmd := GetCmdAllStats()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("all-stats", cmd.Use)
	require.Equal("Query all usage statistics (queries and messages)", cmd.Short)

	// Test that Args validator accepts 0 arguments
	require.NotNil(cmd.Args)
	err := cmd.Args(cmd, []string{})
	require.NoError(err)
}

func (s *QueryTestSuite) TestGetCmdAllStats_RejectsArgs() {
	cmd := GetCmdAllStats()
	require := s.Require()

	// Test that extra args are rejected
	err := cmd.Args(cmd, []string{"extra"})
	require.Error(err)
}

// Test query flags are added to commands
func (s *QueryTestSuite) TestQueryFlagsAdded() {
	require := s.Require()

	commands := []*cobra.Command{
		GetCmdQueryStats(),
		GetCmdMessageStats(),
		GetCmdAllStats(),
	}

	for _, cmd := range commands {
		s.Run(cmd.Name(), func() {
			// Standard query flags should be added
			require.NotNil(cmd.Flags().Lookup("output"))
			require.NotNil(cmd.Flags().Lookup("node"))
		})
	}
}

// Test command structure consistency
func (s *QueryTestSuite) TestCommandStructureConsistency() {
	require := s.Require()

	cmd := GetQueryCmd()

	// Parent command should validate subcommands
	require.NotNil(cmd.RunE)

	// All subcommands should have RunE defined
	for _, subCmd := range cmd.Commands() {
		require.NotNil(subCmd.RunE, "subcommand %s should have RunE defined", subCmd.Name())
	}
}

// Benchmark tests
func BenchmarkQueryStatsCommand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := GetCmdQueryStats()
		_ = cmd.Use
	}
}

func BenchmarkGetQueryCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := GetQueryCmd()
		_ = cmd.Commands()
	}
}
