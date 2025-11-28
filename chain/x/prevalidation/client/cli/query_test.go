package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetQueryCmd(t *testing.T) {
	cmd := GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "prevalidation", cmd.Use)
	require.Contains(t, cmd.Short, "prevalidation")

	subcommands := cmd.Commands()
	require.Len(t, subcommands, 6)
}

func TestCmdQueryPreValidatedTransaction(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid tx ID",
			args:    []string{"tx-123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryPreValidatedTransaction()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "transaction")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryPreValidatedTransactions(t *testing.T) {
	cmd := CmdQueryPreValidatedTransactions()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "transactions")

	// Check flags exist
	require.NotNil(t, cmd.Flags().Lookup("type"))
	require.NotNil(t, cmd.Flags().Lookup("status"))
	require.NotNil(t, cmd.Flags().Lookup("signer"))

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryTemplate(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid template ID",
			args:    []string{"template-123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryTemplate()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "template")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryTemplates(t *testing.T) {
	cmd := CmdQueryTemplates()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "templates")

	// Check flags exist
	require.NotNil(t, cmd.Flags().Lookup("type"))
	require.NotNil(t, cmd.Flags().Lookup("active-only"))

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryMetrics(t *testing.T) {
	cmd := CmdQueryMetrics()
	require.NotNil(t, cmd)
	require.Contains(t, cmd.Use, "metrics")

	// Check detailed flag exists
	require.NotNil(t, cmd.Flags().Lookup("detailed"))

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdQueryParams(t *testing.T) {
	cmd := CmdQueryParams()
	require.NotNil(t, cmd)
	require.Equal(t, "params", cmd.Use)

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestParseTransactionType(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"IR_COMPLETION", false},
		{"DEX_SWAP", false},
		{"LP_DEPOSIT", false},
		{"LP_WITHDRAWAL", false},
		{"VC_MINT", false},
		{"BRIDGE_TRANSFER", false},
		{"CONFIDENCE_SCORE_UPDATE", false},
		{"IDENTITY_CHANGE", false},
		{"INVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseTransactionType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseValidationStatus(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"PENDING", false},
		{"VALIDATED", false},
		{"EXPIRED", false},
		{"EXECUTED", false},
		{"FAILED", false},
		{"INVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseValidationStatus(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseCacheStrategy(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"LRU", false},
		{"LFU", false},
		{"FIFO", false},
		{"ADAPTIVE", false},
		{"INVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseCacheStrategy(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseUint32(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"0", false},
		{"100", false},
		{"4294967295", false},
		{"-1", true},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseUint32(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAllQueryCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"transaction", CmdQueryPreValidatedTransaction()},
		{"transactions", CmdQueryPreValidatedTransactions()},
		{"template", CmdQueryTemplate()},
		{"templates", CmdQueryTemplates()},
		{"metrics", CmdQueryMetrics()},
		{"params", CmdQueryParams()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short)
			require.NotEmpty(t, tc.cmd.Long)
		})
	}
}
