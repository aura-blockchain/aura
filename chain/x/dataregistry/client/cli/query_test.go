package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetQueryCmd(t *testing.T) {
	cmd := GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "dataregistry", cmd.Use)
	require.Contains(t, cmd.Short, "Data Registry")

	subcommands := cmd.Commands()
	require.Len(t, subcommands, 5)
}

func TestCmdShowDataItem(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid data ID",
			args:    []string{"data:abc123"},
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
			cmd := CmdShowDataItem()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "show-data-item")

			// Check requester flag exists
			require.NotNil(t, cmd.Flags().Lookup("requester"))

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

func TestCmdListDataItems(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid owner address",
			args:    []string{"aura1owner123"},
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
			cmd := CmdListDataItems()
			require.NotNil(t, cmd)

			// Check flags exist
			require.NotNil(t, cmd.Flags().Lookup("type"))
			require.NotNil(t, cmd.Flags().Lookup("status"))

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

func TestCmdSearchDataItems(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid JSON query",
			args:    []string{`{"tags":["vacation","2024"]}`},
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
			cmd := CmdSearchDataItems()
			require.NotNil(t, cmd)
			require.Contains(t, cmd.Use, "search-data-items")

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

func TestCmdGetStats(t *testing.T) {
	cmd := CmdGetStats()
	require.NotNil(t, cmd)
	require.Equal(t, "stats", cmd.Use)

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestCmdGetParams(t *testing.T) {
	cmd := CmdGetParams()
	require.NotNil(t, cmd)
	require.Equal(t, "params", cmd.Use)

	err := cmd.Args(cmd, []string{})
	require.NoError(t, err)
}

func TestParseDataTypeProto(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"photo", "PHOTO"},
		{"video", "VIDEO"},
		{"document", "DOCUMENT_PDF"},
		{"golf_score", "GOLF_SCORE"},
		{"vaccination_record", "VACCINATION_RECORD"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseDataTypeProto(tt.input)
			require.NotEqual(t, 0, result, "should parse valid type")
		})
	}
}

func TestParseDataItemStatusProto(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"pending_verification", true},
		{"verified", true},
		{"rejected", true},
		{"expired", true},
		{"revoked", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseDataItemStatusProto(tt.input)
			if tt.valid {
				require.NotEqual(t, 0, result)
			}
		})
	}
}

func TestAllQueryCommandsHaveHelp(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"show-data-item", CmdShowDataItem()},
		{"list-data-items", CmdListDataItems()},
		{"search-data-items", CmdSearchDataItems()},
		{"stats", CmdGetStats()},
		{"params", CmdGetParams()},
	}

	for _, tc := range commands {
		t.Run(tc.name+" has help", func(t *testing.T) {
			require.NotEmpty(t, tc.cmd.Short)
			require.NotEmpty(t, tc.cmd.Long)
		})
	}
}
