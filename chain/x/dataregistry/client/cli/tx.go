package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	dataregistryv1beta1 "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// GetTxCmd returns the transaction commands for the dataregistry module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "dataregistry",
		Aliases:                    []string{"data", "dr"},
		Short:                      "Data registry transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdStoreDataItem(),
		CmdUpdateDataItem(),
		CmdDeleteDataItem(),
		CmdVerifyDataItem(),
		CmdRevokeDataItem(),
	)

	return cmd
}

// CmdStoreDataItem stores a new data item
func CmdStoreDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store [data-type] [title] [description] [storage-location]",
		Short: "Store a new data item in the registry",
		Long: `Store a new data item with metadata and access controls.

Examples:
  aurad tx dataregistry store GOLF_SCORE "Pebble Beach Round" "My best round" ipfs://Qm... --from alice
  aurad tx dataregistry store VEHICLE_REGISTRATION "Toyota Registration" "2024 Registration" ipfs://Qm... --from bob
  aurad tx dataregistry store PHOTO "Sunset Photo" "Beautiful sunset" ipfs://Qm... --from alice --is-encrypted --tags "sunset,beach"

Flags:
  --content-hash: SHA256 hash of content (hex-encoded)
  --is-encrypted: Whether content is encrypted (default: false)
  --geo-lat: Latitude for geotagging
  --geo-lon: Longitude for geotagging
  --geo-name: Location name
  --metadata: Additional metadata as key=value pairs (comma-separated)
  --tags: Search tags (comma-separated)
  --access-mode: Access mode (PRIVATE, WHITELIST, PUBLIC, VERIFIED_USERS) (default: PRIVATE)
  --allowed-addresses: Allowed addresses for WHITELIST mode (comma-separated)
  --min-confidence: Minimum confidence score required for access

Data Types:
  VEHICLE_REGISTRATION, VEHICLE_INSURANCE, PROPERTY_DEED, LEASE_AGREEMENT,
  CONTRACT, RECEIPT, WARRANTY, PHOTO, VIDEO, AUDIO, DOCUMENT_PDF,
  GOLF_SCORE, TEST_SCORE, CERTIFICATION, ACHIEVEMENT, NFT, DIGITAL_ART,
  MUSIC_LICENSE, VACCINATION_RECORD, MEDICAL_RECORD, PRESCRIPTION, CUSTOM
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse data type
			dataType, err := parseDataItemType(args[0])
			if err != nil {
				return err
			}

			// Get flags
			contentHashHex, _ := cmd.Flags().GetString("content-hash")
			var contentHash []byte
			if contentHashHex != "" {
				contentHash, err = hex.DecodeString(contentHashHex)
				if err != nil {
					return fmt.Errorf("invalid content-hash: %w", err)
				}
			}

			isEncrypted, _ := cmd.Flags().GetBool("is-encrypted")

			// Parse geo location
			var geoLocation *dataregistryv1beta1.GeoLocation
			geoLat, _ := cmd.Flags().GetFloat64("geo-lat")
			geoLon, _ := cmd.Flags().GetFloat64("geo-lon")
			geoName, _ := cmd.Flags().GetString("geo-name")
			if geoLat != 0 || geoLon != 0 || geoName != "" {
				geoLocation = &dataregistryv1beta1.GeoLocation{
					Latitude:     geoLat,
					Longitude:    geoLon,
					LocationName: geoName,
				}
			}

			// Parse metadata
			metadata := make(map[string]string)
			metadataStr, _ := cmd.Flags().GetString("metadata")
			if metadataStr != "" {
				pairs := strings.Split(metadataStr, ",")
				for _, pair := range pairs {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
						metadata[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}

			// Parse tags
			var tags []string
			tagsStr, _ := cmd.Flags().GetString("tags")
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}

			// Parse access policy
			accessPolicy := &dataregistryv1beta1.AccessPolicy{}
			accessModeStr, _ := cmd.Flags().GetString("access-mode")
			accessPolicy.Mode = parseAccessMode(accessModeStr)

			allowedAddrs, _ := cmd.Flags().GetString("allowed-addresses")
			if allowedAddrs != "" {
				accessPolicy.AllowedAddresses = strings.Split(allowedAddrs, ",")
			}

			minConfidence, _ := cmd.Flags().GetUint64("min-confidence")
			accessPolicy.MinConfidenceScore = minConfidence

			msg := &dataregistryv1beta1.MsgStoreDataItem{
				Creator:         clientCtx.GetFromAddress().String(),
				DataType:        dataType,
				Title:           args[1],
				Description:     args[2],
				ContentHash:     contentHash,
				StorageLocation: args[3],
				IsEncrypted:     isEncrypted,
				GeoLocation:     geoLocation,
				Metadata:        metadata,
				AccessPolicy:    accessPolicy,
				Tags:            tags,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("content-hash", "", "SHA256 hash of content (hex-encoded)")
	cmd.Flags().Bool("is-encrypted", false, "Whether content is encrypted")
	cmd.Flags().Float64("geo-lat", 0, "Latitude for geotagging")
	cmd.Flags().Float64("geo-lon", 0, "Longitude for geotagging")
	cmd.Flags().String("geo-name", "", "Location name")
	cmd.Flags().String("metadata", "", "Additional metadata as key=value pairs (comma-separated)")
	cmd.Flags().String("tags", "", "Search tags (comma-separated)")
	cmd.Flags().String("access-mode", "PRIVATE", "Access mode (PRIVATE, WHITELIST, PUBLIC, VERIFIED_USERS)")
	cmd.Flags().String("allowed-addresses", "", "Allowed addresses for WHITELIST mode (comma-separated)")
	cmd.Flags().Uint64("min-confidence", 0, "Minimum confidence score required for access")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateDataItem updates an existing data item
func CmdUpdateDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [data-id] [title] [description]",
		Short: "Update an existing data item",
		Long: `Update metadata and access policy for an existing data item.

Examples:
  aurad tx dataregistry update data-123 "Updated Title" "New description" --from alice
  aurad tx dataregistry update data-456 "New Title" "Updated info" --tags "new,tags" --from bob

Note: Only the owner can update a data item.
Content and content hash cannot be updated; create a new version instead.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse metadata
			metadata := make(map[string]string)
			metadataStr, _ := cmd.Flags().GetString("metadata")
			if metadataStr != "" {
				pairs := strings.Split(metadataStr, ",")
				for _, pair := range pairs {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
						metadata[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}

			// Parse tags
			var tags []string
			tagsStr, _ := cmd.Flags().GetString("tags")
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
			}

			// Parse access policy
			accessPolicy := &dataregistryv1beta1.AccessPolicy{}
			accessModeStr, _ := cmd.Flags().GetString("access-mode")
			if accessModeStr != "" {
				accessPolicy.Mode = parseAccessMode(accessModeStr)
			}

			allowedAddrs, _ := cmd.Flags().GetString("allowed-addresses")
			if allowedAddrs != "" {
				accessPolicy.AllowedAddresses = strings.Split(allowedAddrs, ",")
			}

			msg := &dataregistryv1beta1.MsgUpdateDataItem{
				Creator:      clientCtx.GetFromAddress().String(),
				DataId:       args[0],
				Title:        args[1],
				Description:  args[2],
				Metadata:     metadata,
				AccessPolicy: accessPolicy,
				Tags:         tags,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("metadata", "", "Additional metadata as key=value pairs (comma-separated)")
	cmd.Flags().String("tags", "", "Search tags (comma-separated)")
	cmd.Flags().String("access-mode", "", "Access mode (PRIVATE, WHITELIST, PUBLIC, VERIFIED_USERS)")
	cmd.Flags().String("allowed-addresses", "", "Allowed addresses for WHITELIST mode (comma-separated)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdDeleteDataItem deletes a data item
func CmdDeleteDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [data-id]",
		Short: "Delete a data item",
		Long: `Delete a data item from the registry.

Examples:
  aurad tx dataregistry delete data-123 --from alice
  aurad tx dataregistry delete data-456 --from bob

Note: Only the owner can delete a data item.
This marks the item as deleted but retains metadata for audit purposes.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &dataregistryv1beta1.MsgDeleteDataItem{
				Creator: clientCtx.GetFromAddress().String(),
				DataId:  args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdVerifyDataItem verifies a data item
func CmdVerifyDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [data-id] [verification-level] [confidence-score]",
		Short: "Verify a data item",
		Long: `Add verification to a data item with confidence score.

Examples:
  aurad tx dataregistry verify data-123 PEER_VERIFIED 85 --from verifier --notes "Verified via inspection"
  aurad tx dataregistry verify data-456 AI_VERIFIED 92 --from ai-agent --method "OCR + ML"

Verification Levels:
  SELF_ATTESTED - User claims, not verified
  PEER_VERIFIED - Verified by another user
  AI_VERIFIED - Verified by AI agent
  AUTHORITY_VERIFIED - Verified by official authority
  BLOCKCHAIN_ANCHORED - Anchored with external proof

Flags:
  --notes: Verifier notes
  --method: Verification method used
  --proof: Cryptographic proof (hex-encoded)
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse verification level
			verificationLevel, err := parseVerificationLevel(args[1])
			if err != nil {
				return err
			}

			// Parse confidence score
			confidenceScore, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid confidence-score: %w", err)
			}

			if confidenceScore > 100 {
				return fmt.Errorf("confidence score must be between 0 and 100")
			}

			// Get optional flags
			notes, _ := cmd.Flags().GetString("notes")
			method, _ := cmd.Flags().GetString("method")
			proofHex, _ := cmd.Flags().GetString("proof")

			var proof []byte
			if proofHex != "" {
				proof, err = hex.DecodeString(proofHex)
				if err != nil {
					return fmt.Errorf("invalid proof: %w", err)
				}
			}

			msg := &dataregistryv1beta1.MsgVerifyDataItem{
				Verifier:           clientCtx.GetFromAddress().String(),
				DataId:             args[0],
				Level:              verificationLevel,
				ConfidenceScore:    confidenceScore,
				Notes:              notes,
				VerificationMethod: method,
				Proof:              proof,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("notes", "", "Verifier notes")
	cmd.Flags().String("method", "", "Verification method used")
	cmd.Flags().String("proof", "", "Cryptographic proof (hex-encoded)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRevokeDataItem revokes a data item (governance/authority)
func CmdRevokeDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [data-id] [reason]",
		Short: "Revoke a data item (governance/authority only)",
		Long: `Revoke a data item for policy violations or other reasons.

Examples:
  aurad tx dataregistry revoke data-123 "Policy violation" --from authority
  aurad tx dataregistry revoke data-456 "Fraudulent content" --from governance

Note: This command is restricted to authorized authorities.
Revoked items are marked but not deleted for audit purposes.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &dataregistryv1beta1.MsgRevokeDataItem{
				Authority: clientCtx.GetFromAddress().String(),
				DataId:    args[0],
				Reason:    args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// Helper functions

func parseDataItemType(s string) (dataregistryv1beta1.DataItemType, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "VEHICLE_REGISTRATION":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_VEHICLE_REGISTRATION, nil
	case "VEHICLE_INSURANCE":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_VEHICLE_INSURANCE, nil
	case "PROPERTY_DEED":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_PROPERTY_DEED, nil
	case "LEASE_AGREEMENT":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_LEASE_AGREEMENT, nil
	case "CONTRACT":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_CONTRACT, nil
	case "RECEIPT":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_RECEIPT, nil
	case "WARRANTY":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_WARRANTY, nil
	case "PHOTO":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_PHOTO, nil
	case "VIDEO":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_VIDEO, nil
	case "AUDIO":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_AUDIO, nil
	case "DOCUMENT_PDF":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_DOCUMENT_PDF, nil
	case "GOLF_SCORE":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_GOLF_SCORE, nil
	case "TEST_SCORE":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_TEST_SCORE, nil
	case "CERTIFICATION":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_CERTIFICATION, nil
	case "ACHIEVEMENT":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_ACHIEVEMENT, nil
	case "NFT":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_NFT, nil
	case "DIGITAL_ART":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_DIGITAL_ART, nil
	case "MUSIC_LICENSE":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_MUSIC_LICENSE, nil
	case "VACCINATION_RECORD":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_VACCINATION_RECORD, nil
	case "MEDICAL_RECORD":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_MEDICAL_RECORD, nil
	case "PRESCRIPTION":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_PRESCRIPTION, nil
	case "CUSTOM":
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_CUSTOM, nil
	default:
		return dataregistryv1beta1.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED, fmt.Errorf("invalid data item type: %s", s)
	}
}

func parseAccessMode(s string) dataregistryv1beta1.AccessMode {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "WHITELIST":
		return dataregistryv1beta1.AccessMode_ACCESS_MODE_WHITELIST
	case "PUBLIC":
		return dataregistryv1beta1.AccessMode_ACCESS_MODE_PUBLIC
	case "VERIFIED_USERS":
		return dataregistryv1beta1.AccessMode_ACCESS_MODE_VERIFIED_USERS
	default:
		return dataregistryv1beta1.AccessMode_ACCESS_MODE_PRIVATE
	}
}

func parseVerificationLevel(s string) (dataregistryv1beta1.VerificationLevel, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	switch s {
	case "SELF_ATTESTED":
		return dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_SELF_ATTESTED, nil
	case "PEER_VERIFIED":
		return dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_PEER_VERIFIED, nil
	case "AI_VERIFIED":
		return dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_AI_VERIFIED, nil
	case "AUTHORITY_VERIFIED":
		return dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_AUTHORITY_VERIFIED, nil
	case "BLOCKCHAIN_ANCHORED":
		return dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_BLOCKCHAIN_ANCHORED, nil
	default:
		return dataregistryv1beta1.VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED, fmt.Errorf("invalid verification level: %s", s)
	}
}
