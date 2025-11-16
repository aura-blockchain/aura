package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// GetQueryCmd returns the query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "vcregistry",
		Short:                      "Querying commands for the VC Registry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdVerifyPresentation(),
		CmdShowDisclosurePolicy(),
		CmdParseVoiceCommand(),
		CmdListAttributeVCs(),
	)

	return cmd
}

// CmdVerifyPresentation verifies a QR code presentation
func CmdVerifyPresentation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-presentation [qr-data]",
		Short: "Verify a QR code presentation",
		Long: `Verify a QR code presentation and extract disclosed attributes.

Examples:
  aurad query vcregistry verify-presentation "aura://verify?data=base64encodeddata"
  aurad query vcregistry verify-presentation "aura://verify?data=..." --verifier aura1abc...

The command will:
  1. Parse and validate the QR code data
  2. Verify the signature
  3. Check VC validity and expiration
  4. Extract disclosed attributes based on context
  5. Return verification result
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistrypb.NewQueryClient(clientCtx)

			// Get verifier address from flag or use client address
			verifier, _ := cmd.Flags().GetString("verifier")
			if verifier == "" {
				verifier = clientCtx.GetFromAddress().String()
			}

			req := &vcregistrypb.QueryVerifyPresentationRequest{
				QrCodeData:      args[0],
				VerifierAddress: verifier,
			}

			res, err := queryClient.VerifyPresentation(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("verifier", "", "Verifier address (optional)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdShowDisclosurePolicy shows user's disclosure policy
func CmdShowDisclosurePolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-disclosure-policy [address]",
		Short: "Show user's selective disclosure policy",
		Long: `Show a user's selective disclosure policy.

Examples:
  aurad query vcregistry show-disclosure-policy aura1abc...
  aurad query vcregistry show-disclosure-policy $(aurad keys show alice -a)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistrypb.NewQueryClient(clientCtx)

			req := &vcregistrypb.QueryGetDisclosurePolicyRequest{
				Address: args[0],
			}

			res, err := queryClient.GetDisclosurePolicy(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdParseVoiceCommand parses a voice command to attribute types
func CmdParseVoiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse-voice-command [command-text]",
		Short: "Parse a voice command into attribute types",
		Long: `Parse a voice command into attribute types for selective disclosure.

Examples:
  aurad query vcregistry parse-voice-command "AURA show my age and address"
  aurad query vcregistry parse-voice-command "AURA show everything"
  aurad query vcregistry parse-voice-command "AURA show only that I'm over 21"

Supported voice commands:
  - "AURA show my age" -> [ATTRIBUTE_TYPE_AGE]
  - "AURA show my name and address" -> [ATTRIBUTE_TYPE_FULL_NAME, ATTRIBUTE_TYPE_ADDRESS_FULL]
  - "AURA show everything" -> [ALL_ATTRIBUTES]
  - "AURA show only that I'm over 21" -> [ATTRIBUTE_TYPE_AGE] with ZK proof flag
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistrypb.NewQueryClient(clientCtx)

			req := &vcregistrypb.QueryParseVoiceCommandRequest{
				CommandText: args[0],
			}

			res, err := queryClient.ParseVoiceCommand(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdListAttributeVCs lists user's attribute VCs
func CmdListAttributeVCs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-attribute-vcs [address]",
		Short: "List user's attribute VCs",
		Long: `List all attribute VCs for a user, optionally filtered by attribute type.

Examples:
  aurad query vcregistry list-attribute-vcs aura1abc...
  aurad query vcregistry list-attribute-vcs aura1abc... --attribute-type age
  aurad query vcregistry list-attribute-vcs $(aurad keys show alice -a)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := vcregistrypb.NewQueryClient(clientCtx)

			// Parse optional attribute type filter
			attributeTypeStr, _ := cmd.Flags().GetString("attribute-type")
			var attributeTypes []vcregistrypb.AttributeType
			if attributeTypeStr != "" {
				attrType := parseAttributeType(attributeTypeStr)
				if attrType != vcregistrypb.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED {
					attributeTypes = []vcregistrypb.AttributeType{attrType}
				}
			}

			req := &vcregistrypb.QueryListAttributeVCsRequest{
				Address:        args[0],
				AttributeTypes: attributeTypes,
			}

			res, err := queryClient.ListAttributeVCs(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("attribute-type", "", "Filter by attribute type (optional)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// Helper function to pretty print QR code data (for debugging)
func prettyPrintQRData(qrData string) error {
	// Extract base64 data from URI
	const prefix = "aura://verify?data="
	if len(qrData) < len(prefix) {
		return fmt.Errorf("invalid QR code format")
	}

	encodedData := qrData[len(prefix):]

	// Decode from base64
	jsonData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	// Pretty print JSON
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(jsonData, &prettyJSON); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	prettyBytes, err := json.MarshalIndent(prettyJSON, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(prettyBytes))
	return nil
}
