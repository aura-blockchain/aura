package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	dataregistryv1beta1 "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

func TestGetTxCmd(t *testing.T) {
	cmd := GetTxCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "dataregistry", cmd.Use)
	require.Contains(t, cmd.Short, "Data registry transaction")

	// Check all subcommands are added
	subcommands := cmd.Commands()
	require.Len(t, subcommands, 5)

	cmdNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		cmdNames[subcmd.Name()] = true
	}

	require.True(t, cmdNames["store"])
	require.True(t, cmdNames["update"])
	require.True(t, cmdNames["delete"])
	require.True(t, cmdNames["verify"])
	require.True(t, cmdNames["revoke"])
}

func TestCmdStoreDataItem(t *testing.T) {
	cmd := CmdStoreDataItem()
	require.NotNil(t, cmd)
	require.Equal(t, "store", cmd.Name())
	require.Contains(t, cmd.Short, "Store a new data item")

	// Check flags
	require.NotNil(t, cmd.Flags().Lookup("content-hash"))
	require.NotNil(t, cmd.Flags().Lookup("is-encrypted"))
	require.NotNil(t, cmd.Flags().Lookup("geo-lat"))
	require.NotNil(t, cmd.Flags().Lookup("geo-lon"))
	require.NotNil(t, cmd.Flags().Lookup("geo-name"))
	require.NotNil(t, cmd.Flags().Lookup("metadata"))
	require.NotNil(t, cmd.Flags().Lookup("tags"))
	require.NotNil(t, cmd.Flags().Lookup("access-mode"))
	require.NotNil(t, cmd.Flags().Lookup("allowed-addresses"))
	require.NotNil(t, cmd.Flags().Lookup("min-confidence"))

	// Validate usage
	require.Contains(t, cmd.Use, "[data-type]")
	require.Contains(t, cmd.Use, "[title]")
	require.Contains(t, cmd.Use, "[description]")
	require.Contains(t, cmd.Use, "[storage-location]")
}

func TestCmdUpdateDataItem(t *testing.T) {
	cmd := CmdUpdateDataItem()
	require.NotNil(t, cmd)
	require.Equal(t, "update", cmd.Name())
	require.Contains(t, cmd.Short, "Update an existing data item")

	// Check flags
	require.NotNil(t, cmd.Flags().Lookup("metadata"))
	require.NotNil(t, cmd.Flags().Lookup("tags"))
	require.NotNil(t, cmd.Flags().Lookup("access-mode"))
	require.NotNil(t, cmd.Flags().Lookup("allowed-addresses"))

	// Validate usage
	require.Contains(t, cmd.Use, "[data-id]")
	require.Contains(t, cmd.Use, "[title]")
	require.Contains(t, cmd.Use, "[description]")
}

func TestCmdDeleteDataItem(t *testing.T) {
	cmd := CmdDeleteDataItem()
	require.NotNil(t, cmd)
	require.Equal(t, "delete", cmd.Name())
	require.Contains(t, cmd.Short, "Delete a data item")

	// Validate usage
	require.Contains(t, cmd.Use, "[data-id]")
}

func TestCmdVerifyDataItem(t *testing.T) {
	cmd := CmdVerifyDataItem()
	require.NotNil(t, cmd)
	require.Equal(t, "verify", cmd.Name())
	require.Contains(t, cmd.Short, "Verify a data item")

	// Check flags
	require.NotNil(t, cmd.Flags().Lookup("notes"))
	require.NotNil(t, cmd.Flags().Lookup("method"))
	require.NotNil(t, cmd.Flags().Lookup("proof"))

	// Validate usage
	require.Contains(t, cmd.Use, "[data-id]")
	require.Contains(t, cmd.Use, "[verification-level]")
	require.Contains(t, cmd.Use, "[confidence-score]")
}

func TestCmdRevokeDataItem(t *testing.T) {
	cmd := CmdRevokeDataItem()
	require.NotNil(t, cmd)
	require.Equal(t, "revoke", cmd.Name())
	require.Contains(t, cmd.Short, "Revoke a data item")

	// Validate usage
	require.Contains(t, cmd.Use, "[data-id]")
	require.Contains(t, cmd.Use, "[reason]")
}

func TestParseDataItemType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected dataregistryv1beta1.DataItemType
		wantErr  bool
	}{
		{
			name:     "vehicle registration",
			input:    "VEHICLE_REGISTRATION",
			expected: dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION,
			wantErr:  false,
		},
		{
			name:     "vehicle registration lowercase",
			input:    "vehicle_registration",
			expected: dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION,
			wantErr:  false,
		},
		{
			name:     "photo",
			input:    "PHOTO",
			expected: dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_PHOTO,
			wantErr:  false,
		},
		{
			name:     "golf score",
			input:    "GOLF_SCORE",
			expected: dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE,
			wantErr:  false,
		},
		{
			name:     "vaccination record",
			input:    "VACCINATION_RECORD",
			expected: dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_VACCINATION_RECORD,
			wantErr:  false,
		},
		{
			name:     "custom",
			input:    "CUSTOM",
			expected: dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_CUSTOM,
			wantErr:  false,
		},
		{
			name:     "invalid type",
			input:    "INVALID_TYPE",
			expected: dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDataItemType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseAccessMode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected dataregistryv1beta1.AccessMode
	}{
		{
			name:     "private",
			input:    "PRIVATE",
			expected: dataregistryv1beta1.AccessMode_ACCESS_MODE_PRIVATE,
		},
		{
			name:     "whitelist",
			input:    "WHITELIST",
			expected: dataregistryv1beta1.AccessMode_ACCESS_MODE_WHITELIST,
		},
		{
			name:     "public",
			input:    "PUBLIC",
			expected: dataregistryv1beta1.AccessMode_ACCESS_MODE_PUBLIC,
		},
		{
			name:     "verified users",
			input:    "VERIFIED_USERS",
			expected: dataregistryv1beta1.AccessMode_ACCESS_MODE_VERIFIED_USERS,
		},
		{
			name:     "default to private",
			input:    "UNKNOWN",
			expected: dataregistryv1beta1.AccessMode_ACCESS_MODE_PRIVATE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAccessMode(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestParseVerificationLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected dataregistryv1beta1.VerificationLevel
		wantErr  bool
	}{
		{
			name:     "self attested",
			input:    "SELF_ATTESTED",
			expected: dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED,
			wantErr:  false,
		},
		{
			name:     "peer verified",
			input:    "PEER_VERIFIED",
			expected: dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED,
			wantErr:  false,
		},
		{
			name:     "ai verified",
			input:    "AI_VERIFIED",
			expected: dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED,
			wantErr:  false,
		},
		{
			name:     "authority verified",
			input:    "AUTHORITY_VERIFIED",
			expected: dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED,
			wantErr:  false,
		},
		{
			name:     "blockchain anchored",
			input:    "BLOCKCHAIN_ANCHORED",
			expected: dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_BLOCKCHAIN_ANCHORED,
			wantErr:  false,
		},
		{
			name:     "invalid level",
			input:    "INVALID",
			expected: dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseVerificationLevel(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseDataItemTypeCoverage(t *testing.T) {
	// Test all possible data types to ensure comprehensive coverage
	allTypes := []string{
		"VEHICLE_REGISTRATION",
		"VEHICLE_INSURANCE",
		"PROPERTY_DEED",
		"LEASE_AGREEMENT",
		"CONTRACT",
		"RECEIPT",
		"WARRANTY",
		"PHOTO",
		"VIDEO",
		"AUDIO",
		"DOCUMENT_PDF",
		"GOLF_SCORE",
		"TEST_SCORE",
		"CERTIFICATION",
		"ACHIEVEMENT",
		"NFT",
		"DIGITAL_ART",
		"MUSIC_LICENSE",
		"VACCINATION_RECORD",
		"MEDICAL_RECORD",
		"PRESCRIPTION",
		"CUSTOM",
	}

	for _, typeStr := range allTypes {
		t.Run(fmt.Sprintf("parse_%s", typeStr), func(t *testing.T) {
			result, err := parseDataItemType(typeStr)
			require.NoError(t, err)
			require.NotEqual(t, dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED, result)
		})
	}
}
