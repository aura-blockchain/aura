package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/aequitas/aura/chain/x/dataregistry/types"
	dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "dataregistry",
		Short:                      "Data Registry transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdStoreDataItem(),
		CmdUpdateDataItem(),
		CmdDeleteDataItem(),
		CmdUpdateAccessPolicy(),
		CmdVerifyDataItem(),
	)

	return cmd
}

// CmdStoreDataItem stores a new data item
func CmdStoreDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store-data-item [data-type] [title] [content-hash] [storage-location]",
		Short: "Store a new data item on-chain",
		Long: `Store a new data item with metadata and access policy.

Examples:
  aurad tx dataregistry store-data-item photo "Vacation Photo" abc123... ipfs://Qm... --from alice
  aurad tx dataregistry store-data-item document "Report" def456... https://... --encrypted --from alice
  aurad tx dataregistry store-data-item photo "Beach" abc... ipfs://... --description "Beach photo from 2024" --tags "vacation,beach,2024" --from alice

Data types:
  - photo, video, audio
  - document, text, code
  - health_record, biometric
  - location, sensor_data
  - certificate, credential
  - contract, receipt
  - social_post, message
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse data type
			dataType := parseDataType(args[0])
			if dataType == types.DataItemTypeUnspecified {
				return fmt.Errorf("invalid data type: %s", args[0])
			}

			title := args[1]

			// Parse content hash (hex)
			contentHash, err := hex.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("invalid content hash (must be hex): %w", err)
			}

			storageLocation := args[3]

			// Parse optional flags
			description, _ := cmd.Flags().GetString("description")
			isEncrypted, _ := cmd.Flags().GetBool("encrypted")
			tagsStr, _ := cmd.Flags().GetString("tags")
			metadataStr, _ := cmd.Flags().GetString("metadata")
			accessPolicyStr, _ := cmd.Flags().GetString("access-policy")
			geoLocationStr, _ := cmd.Flags().GetString("geo-location")

			// Parse tags
			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i, tag := range tags {
					tags[i] = strings.TrimSpace(tag)
				}
			}

			// Parse metadata JSON
			var metadata map[string]string
			if metadataStr != "" {
				if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
					return fmt.Errorf("invalid metadata JSON: %w", err)
				}
			}

			// Parse access policy JSON
			var accessPolicy *types.AccessPolicy
			if accessPolicyStr != "" {
				accessPolicy = &types.AccessPolicy{}
				if err := json.Unmarshal([]byte(accessPolicyStr), accessPolicy); err != nil {
					return fmt.Errorf("invalid access-policy JSON: %w", err)
				}
			} else {
				// Default to private
				accessPolicy = &types.AccessPolicy{
					Mode: types.AccessModePrivate,
				}
			}

			// Parse geo-location JSON
			var geoLocation *types.GeoLocation
			if geoLocationStr != "" {
				geoLocation = &types.GeoLocation{}
				if err := json.Unmarshal([]byte(geoLocationStr), geoLocation); err != nil {
					return fmt.Errorf("invalid geo-location JSON: %w", err)
				}
			}

			// Create message
			msg := &dataregistrypb.MsgStoreDataItem{
				Creator:         clientCtx.GetFromAddress().String(),
				DataType:        convertDataTypeToProto(dataType),
				Title:           title,
				Description:     description,
				ContentHash:     contentHash,
				StorageLocation: storageLocation,
				IsEncrypted:     isEncrypted,
				GeoLocation:     convertGeoLocationToProto(geoLocation),
				Metadata:        metadata,
				AccessPolicy:    convertAccessPolicyToProto(accessPolicy),
				Tags:            tags,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("description", "", "Data item description")
	cmd.Flags().Bool("encrypted", false, "Whether the data is encrypted")
	cmd.Flags().String("tags", "", "Comma-separated tags")
	cmd.Flags().String("metadata", "", "Metadata as JSON object")
	cmd.Flags().String("access-policy", "", "Access policy as JSON object")
	cmd.Flags().String("geo-location", "", "Geo-location as JSON object {\"latitude\":0.0,\"longitude\":0.0}")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateDataItem updates an existing data item
func CmdUpdateDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-data-item [data-id]",
		Short: "Update an existing data item's metadata",
		Long: `Update an existing data item's title, description, metadata, access policy, or tags.

Examples:
  aurad tx dataregistry update-data-item data:abc123... --title "New Title" --from alice
  aurad tx dataregistry update-data-item data:abc123... --description "Updated description" --tags "new,tags" --from alice
  aurad tx dataregistry update-data-item data:abc123... --access-policy '{"mode":"public"}' --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dataID := args[0]

			// Parse optional flags
			title, _ := cmd.Flags().GetString("title")
			description, _ := cmd.Flags().GetString("description")
			tagsStr, _ := cmd.Flags().GetString("tags")
			metadataStr, _ := cmd.Flags().GetString("metadata")
			accessPolicyStr, _ := cmd.Flags().GetString("access-policy")

			// Parse tags
			var tags []string
			if tagsStr != "" {
				tags = strings.Split(tagsStr, ",")
				for i, tag := range tags {
					tags[i] = strings.TrimSpace(tag)
				}
			}

			// Parse metadata JSON
			var metadata map[string]string
			if metadataStr != "" {
				if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
					return fmt.Errorf("invalid metadata JSON: %w", err)
				}
			}

			// Parse access policy JSON
			var accessPolicy *types.AccessPolicy
			if accessPolicyStr != "" {
				accessPolicy = &types.AccessPolicy{}
				if err := json.Unmarshal([]byte(accessPolicyStr), accessPolicy); err != nil {
					return fmt.Errorf("invalid access-policy JSON: %w", err)
				}
			}

			// Create message
			msg := &dataregistrypb.MsgUpdateDataItem{
				Creator:      clientCtx.GetFromAddress().String(),
				DataId:       dataID,
				Title:        title,
				Description:  description,
				Metadata:     metadata,
				AccessPolicy: convertAccessPolicyToProto(accessPolicy),
				Tags:         tags,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("tags", "", "New comma-separated tags")
	cmd.Flags().String("metadata", "", "New metadata as JSON object")
	cmd.Flags().String("access-policy", "", "New access policy as JSON object")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdDeleteDataItem deletes a data item
func CmdDeleteDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-data-item [data-id]",
		Short: "Delete a data item",
		Long: `Delete a data item from the registry. Only the owner can delete their data.

Examples:
  aurad tx dataregistry delete-data-item data:abc123... --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &dataregistrypb.MsgDeleteDataItem{
				Creator: clientCtx.GetFromAddress().String(),
				DataId:  args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateAccessPolicy updates a data item's access policy
func CmdUpdateAccessPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-access-policy [data-id] [policy-json]",
		Short: "Update a data item's access policy",
		Long: `Update a data item's access policy to control who can access it.

Examples:
  aurad tx dataregistry update-access-policy data:abc... '{"mode":"public"}' --from alice
  aurad tx dataregistry update-access-policy data:abc... '{"mode":"whitelist","allowed_addresses":["aura1...","aura2..."]}' --from alice
  aurad tx dataregistry update-access-policy data:abc... '{"mode":"verified_users"}' --from alice

Access modes:
  - private: Only owner can access
  - public: Anyone can access
  - whitelist: Only specified addresses can access
  - verified_users: Only users with AURA VCs can access
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dataID := args[0]

			// Parse access policy JSON
			var accessPolicy types.AccessPolicy
			if err := json.Unmarshal([]byte(args[1]), &accessPolicy); err != nil {
				return fmt.Errorf("invalid access-policy JSON: %w", err)
			}

			// Create message
			msg := &dataregistrypb.MsgUpdateDataItem{
				Creator:      clientCtx.GetFromAddress().String(),
				DataId:       dataID,
				AccessPolicy: convertAccessPolicyToProto(&accessPolicy),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdVerifyDataItem adds a verification to a data item
func CmdVerifyDataItem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-data-item [data-id] [verification-level] [confidence-score]",
		Short: "Verify a data item",
		Long: `Add a verification to a data item.

Examples:
  aurad tx dataregistry verify-data-item data:abc... peer_verified 85 --from bob
  aurad tx dataregistry verify-data-item data:abc... expert_verified 95 --notes "Verified authenticity" --from expert
  aurad tx dataregistry verify-data-item data:abc... authority_verified 100 --method "manual_inspection" --from authority

Verification levels:
  - self_attested (0)
  - peer_verified (1)
  - expert_verified (2)
  - authority_verified (3)
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dataID := args[0]

			// Parse verification level
			level := parseVerificationLevel(args[1])
			if level == types.VerificationLevelUnspecified {
				return fmt.Errorf("invalid verification level: %s", args[1])
			}

			// Parse confidence score
			confidenceScore, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid confidence score: %w", err)
			}
			if confidenceScore > 100 {
				return fmt.Errorf("confidence score must be between 0 and 100")
			}

			// Parse optional flags
			notes, _ := cmd.Flags().GetString("notes")
			method, _ := cmd.Flags().GetString("method")
			proofStr, _ := cmd.Flags().GetString("proof")

			// Parse proof (hex)
			var proof []byte
			if proofStr != "" {
				proof, err = hex.DecodeString(proofStr)
				if err != nil {
					return fmt.Errorf("invalid proof (must be hex): %w", err)
				}
			}

			// Create message
			msg := &dataregistrypb.MsgVerifyDataItem{
				Verifier:           clientCtx.GetFromAddress().String(),
				DataId:             dataID,
				Level:              convertVerificationLevelToProto(level),
				ConfidenceScore:    confidenceScore,
				Notes:              notes,
				VerificationMethod: method,
				Proof:              proof,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("notes", "", "Verification notes")
	cmd.Flags().String("method", "", "Verification method")
	cmd.Flags().String("proof", "", "Verification proof as hex string")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// parseDataType converts a string to DataItemType enum
func parseDataType(s string) types.DataItemType {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")

	typeMap := map[string]types.DataItemType{
		"photo":         types.DataItemTypePhoto,
		"video":         types.DataItemTypeVideo,
		"audio":         types.DataItemTypeAudio,
		"document":      types.DataItemTypeDocument,
		"text":          types.DataItemTypeText,
		"code":          types.DataItemTypeCode,
		"health_record": types.DataItemTypeHealthRecord,
		"biometric":     types.DataItemTypeBiometric,
		"location":      types.DataItemTypeLocation,
		"sensor_data":   types.DataItemTypeSensorData,
		"certificate":   types.DataItemTypeCertificate,
		"credential":    types.DataItemTypeCredential,
		"contract":      types.DataItemTypeContract,
		"receipt":       types.DataItemTypeReceipt,
		"social_post":   types.DataItemTypeSocialPost,
		"message":       types.DataItemTypeMessage,
	}

	if dataType, ok := typeMap[s]; ok {
		return dataType
	}

	return types.DataItemTypeUnspecified
}

// parseVerificationLevel converts a string to VerificationLevel enum
func parseVerificationLevel(s string) types.VerificationLevel {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")

	levelMap := map[string]types.VerificationLevel{
		"self_attested":      types.VerificationLevelSelfAttested,
		"peer_verified":      types.VerificationLevelPeerVerified,
		"expert_verified":    types.VerificationLevelExpertVerified,
		"authority_verified": types.VerificationLevelAuthorityVerified,
	}

	if level, ok := levelMap[s]; ok {
		return level
	}

	return types.VerificationLevelUnspecified
}

// convertDataTypeToProto converts local DataItemType to proto DataItemType
func convertDataTypeToProto(dt types.DataItemType) dataregistrypb.DataItemType {
	return dataregistrypb.DataItemType(dt)
}

// convertVerificationLevelToProto converts local VerificationLevel to proto VerificationLevel
func convertVerificationLevelToProto(vl types.VerificationLevel) dataregistrypb.VerificationLevel {
	return dataregistrypb.VerificationLevel(vl)
}

// convertGeoLocationToProto converts local GeoLocation to proto GeoLocation
func convertGeoLocationToProto(gl *types.GeoLocation) *dataregistrypb.GeoLocation {
	if gl == nil {
		return nil
	}
	return &dataregistrypb.GeoLocation{
		Latitude:  gl.Latitude,
		Longitude: gl.Longitude,
		Country:   gl.Country,
		Region:    gl.Region,
		City:      gl.City,
	}
}

// convertAccessPolicyToProto converts local AccessPolicy to proto AccessPolicy
func convertAccessPolicyToProto(ap *types.AccessPolicy) *dataregistrypb.AccessPolicy {
	if ap == nil {
		return nil
	}
	return &dataregistrypb.AccessPolicy{
		Mode:          dataregistrypb.AccessMode(ap.Mode),
		AllowedUsers:  ap.AllowedUsers,
		RequiredRoles: ap.RequiredRoles,
		ExpiresAt:     ap.ExpiresAt,
	}
}
