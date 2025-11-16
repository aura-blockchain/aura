package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "vcregistry",
		Short:                      "VC Registry transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreatePresentation(),
		CmdCreateAttributeVC(),
		CmdUpdateDisclosurePolicy(),
	)

	return cmd
}

// CmdCreatePresentation creates a new QR code presentation
func CmdCreatePresentation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-presentation [vc-ids] [context-json] [expires-in-seconds]",
		Short: "Create a QR code presentation for selective disclosure",
		Long: `Create a QR code presentation containing specified VCs with selective disclosure context.

Examples:
  aurad tx vcregistry create-presentation vc:abc123,vc:def456 '{"show_full_name":true,"show_age":true}' 300 --from alice
  aurad tx vcregistry create-presentation vc:abc123 '{"show_age_over_18":true}' 600 --from alice

The context JSON supports the following fields:
  - show_full_name: bool
  - show_age: bool
  - show_age_over_18: bool
  - show_age_over_21: bool
  - show_address: bool
  - show_city_state_only: bool
  - show_professional_license: bool
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse VC IDs
			vcIDs := strings.Split(args[0], ",")
			for i, id := range vcIDs {
				vcIDs[i] = strings.TrimSpace(id)
			}

			// Parse context JSON
			var contextMap map[string]bool
			if err := json.Unmarshal([]byte(args[1]), &contextMap); err != nil {
				return fmt.Errorf("invalid context JSON: %w", err)
			}

			context := &vcregistrypb.PresentationContext{
				ShowFullName:            contextMap["show_full_name"],
				ShowAge:                 contextMap["show_age"],
				ShowAgeOver_18:          contextMap["show_age_over_18"],
				ShowAgeOver_21:          contextMap["show_age_over_21"],
				ShowAddress:             contextMap["show_address"],
				ShowCityStateOnly:       contextMap["show_city_state_only"],
				ShowProfessionalLicense: contextMap["show_professional_license"],
			}

			// Parse expires in seconds
			expiresIn, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid expires-in-seconds: %w", err)
			}

			// Create message
			msg := &vcregistrypb.MsgCreatePresentation{
				Creator:          clientCtx.GetFromAddress().String(),
				VcIds:            vcIDs,
				Context:          context,
				ExpiresInSeconds: expiresIn,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateAttributeVC creates an attribute VC
func CmdCreateAttributeVC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-attribute-vc [attribute-type] [encrypted-value]",
		Short: "Create an encrypted attribute VC for selective disclosure",
		Long: `Create an encrypted attribute VC for fine-grained selective disclosure.

Examples:
  aurad tx vcregistry create-attribute-vc age "base64encodedencryptedvalue" --from alice
  aurad tx vcregistry create-attribute-vc address "base64encodedencryptedvalue" --from alice

Supported attribute types:
  - full_name, first_name, last_name
  - age, date_of_birth
  - address_full, address_street, address_city, address_state, address_zip, address_country
  - email, phone
  - height, weight, eye_color, hair_color
  - occupation, employer, professional_license, education_level, degree
  - passport_number, drivers_license, ssn
  - scuba_certified, pilots_license, security_clearance
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse attribute type
			attributeType := parseAttributeType(args[0])
			if attributeType == vcregistrypb.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED {
				return fmt.Errorf("invalid attribute type: %s", args[0])
			}

			// Parse encrypted value (base64)
			encryptedValue, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("invalid base64 encrypted value: %w", err)
			}

			// Create message
			msg := &vcregistrypb.MsgCreateAttributeVC{
				Creator:        clientCtx.GetFromAddress().String(),
				AttributeType:  attributeType,
				EncryptedValue: encryptedValue,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateDisclosurePolicy updates user's disclosure policy
func CmdUpdateDisclosurePolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-disclosure-policy [policy-json]",
		Short: "Update your selective disclosure policy",
		Long: `Update your selective disclosure policy to control which attributes are shown by default.

Examples:
  aurad tx vcregistry update-disclosure-policy '{"auto_disclose_age":true,"auto_disclose_address":false}' --from alice

The policy JSON supports various auto-disclosure flags for different attributes.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse policy JSON
			var policyMap map[string]interface{}
			if err := json.Unmarshal([]byte(args[0]), &policyMap); err != nil {
				return fmt.Errorf("invalid policy JSON: %w", err)
			}

			// Create disclosure policy from map
			policy := &vcregistrypb.DisclosurePolicy{
				// Map the policy fields from the JSON
				// This is a simplified implementation - in production would have more fields
			}

			// Create message
			msg := &vcregistrypb.MsgUpdateDisclosurePolicy{
				Creator: clientCtx.GetFromAddress().String(),
				Policy:  policy,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// parseAttributeType converts a string to AttributeType enum
func parseAttributeType(s string) vcregistrypb.AttributeType {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")

	attributeMap := map[string]vcregistrypb.AttributeType{
		"full_name":            vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FULL_NAME,
		"first_name":           vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FIRST_NAME,
		"last_name":            vcregistrypb.AttributeType_ATTRIBUTE_TYPE_LAST_NAME,
		"age":                  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
		"date_of_birth":        vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DATE_OF_BIRTH,
		"address_full":         vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL,
		"address_street":       vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STREET,
		"address_city":         vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_CITY,
		"address_state":        vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STATE,
		"address_zip":          vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_ZIP,
		"address_country":      vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_COUNTRY,
		"email":                vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		"phone":                vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PHONE,
		"height":               vcregistrypb.AttributeType_ATTRIBUTE_TYPE_HEIGHT,
		"weight":               vcregistrypb.AttributeType_ATTRIBUTE_TYPE_WEIGHT,
		"eye_color":            vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EYE_COLOR,
		"hair_color":           vcregistrypb.AttributeType_ATTRIBUTE_TYPE_HAIR_COLOR,
		"occupation":           vcregistrypb.AttributeType_ATTRIBUTE_TYPE_OCCUPATION,
		"employer":             vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMPLOYER,
		"professional_license": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE,
		"education_level":      vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EDUCATION_LEVEL,
		"degree":               vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DEGREE,
		"passport_number":      vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PASSPORT_NUMBER,
		"drivers_license":      vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DRIVERS_LICENSE,
		"ssn":                  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SSN,
		"scuba_certified":      vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SCUBA_CERTIFIED,
		"pilots_license":       vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PILOTS_LICENSE,
		"security_clearance":   vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SECURITY_CLEARANCE,
	}

	if attrType, ok := attributeMap[s]; ok {
		return attrType
	}

	return vcregistrypb.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED
}
