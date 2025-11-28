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

	vcregistryv1beta1 "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// GetTxCmd returns the transaction commands for the vcregistry module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "vcregistry",
		Aliases:                    []string{"vc", "vcr"},
		Short:                      "VC Registry transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// Core VC lifecycle
		CmdMintVC(),
		CmdRevokeVC(),
		CmdAdminRevokeVC(),
		CmdSuspendVC(),
		CmdReactivateVC(),

		// VC Policy management (governance)
		CmdCreateVCPolicy(),
		CmdUpdateVCPolicy(),
		CmdDeprecateVCPolicy(),

		// DID management
		CmdRegisterDID(),
		CmdUpdateDIDDocument(),

		// Presentation (will be added when presentation Msg service is defined in proto)
		// CmdCreatePresentation(),

		// Selective disclosure / Attributes (will be added when attributes Msg service is defined in proto)
		// CmdCreateAttributeVC(),
		// CmdRevokeAttributeVC(),
		// CmdUpdateDisclosurePolicy(),
		// CmdRespondToDisclosureRequest(),
		// CmdCreateDisclosureRequest(),
	)

	return cmd
}

// CmdMintVC mints a new verifiable credential
func CmdMintVC() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mint-vc [holder-did] [vc-type]",
		Aliases: []string{"mint", "create-vc"},
		Short:   "Mint a new verifiable credential",
		Long: `Mint a new verifiable credential for the holder.

Examples:
  aurad tx vcregistry mint-vc did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from alice
  aurad tx vcregistry mint-vc did:aura:mainnet:user456 VC_TYPE_AGE_OVER_18 --from bob
  aurad tx vcregistry mint-vc did:aura:mainnet:user789 VC_TYPE_CUSTOM --custom-type "MyCustomVC" --from alice
  aurad tx vcregistry mint-vc did:aura:mainnet:user789 VC_TYPE_KYC_VERIFICATION --metadata "country=US,tier=gold" --from alice

VC Types:
  VC_TYPE_VERIFIED_HUMAN, VC_TYPE_AGE_OVER_18, VC_TYPE_AGE_OVER_21, VC_TYPE_RESIDENT_OF,
  VC_TYPE_BIOMETRIC_AUTH, VC_TYPE_KYC_VERIFICATION, VC_TYPE_NOTARY_PUBLIC,
  VC_TYPE_PROFESSIONAL_LICENSE, VC_TYPE_BIOMETRIC_FOCUS, VC_TYPE_SOCIAL_FOCUS,
  VC_TYPE_GEOLOCATION_FOCUS, VC_TYPE_HIGH_ASSURANCE_FOCUS, VC_TYPE_POSSESSION_FOCUS,
  VC_TYPE_KNOWLEDGE_FOCUS, VC_TYPE_PERSISTENCE_FOCUS, VC_TYPE_SPECIALIZED_FOCUS,
  VC_TYPE_CUSTOM

Requirements:
  - Must meet minimum Confidence Score threshold
  - Must have completed required Inclusion Routines
  - Must not exceed rate limits
  - Policy must be active for the VC type

Flags:
  --custom-type: Custom type name (required for VC_TYPE_CUSTOM)
  --metadata: Key-value pairs (format: "key1=val1,key2=val2")
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			holderDID := args[0]
			vcTypeStr := args[1]

			// Parse VC type
			vcType, err := parseVCType(vcTypeStr)
			if err != nil {
				return err
			}

			// Get optional custom type
			customType, _ := cmd.Flags().GetString("custom-type")
			if vcType == vcregistryv1beta1.VCType_VC_TYPE_CUSTOM && customType == "" {
				return fmt.Errorf("--custom-type is required for VC_TYPE_CUSTOM")
			}

			// Parse metadata
			metadataStr, _ := cmd.Flags().GetString("metadata")
			metadata := make(map[string]string)
			if metadataStr != "" {
				pairs := strings.Split(metadataStr, ",")
				for _, pair := range pairs {
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) == 2 {
						metadata[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}

			msg := &vcregistryv1beta1.MsgMintVC{
				HolderAddress: clientCtx.GetFromAddress().String(),
				HolderDid:     holderDID,
				VcType:        vcType,
				VcTypeCustom:  customType,
				Metadata:      metadata,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("custom-type", "", "Custom VC type name (required for VC_TYPE_CUSTOM)")
	cmd.Flags().String("metadata", "", "Metadata key-value pairs (format: key1=val1,key2=val2)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRevokeVC revokes a credential (user-initiated)
func CmdRevokeVC() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke-vc [vc-id]",
		Aliases: []string{"revoke"},
		Short:   "Revoke your own verifiable credential",
		Long: `Revoke a verifiable credential that you own.

Examples:
  aurad tx vcregistry revoke-vc vc-123456 --from alice
  aurad tx vcregistry revoke-vc vc-789012 --reason "No longer needed" --from bob

Note:
  - Only the holder can revoke their own VCs
  - Revocation is permanent and cannot be undone
  - The VC will be added to the revocation Merkle tree
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vcID := args[0]
			reason, _ := cmd.Flags().GetString("reason")

			msg := &vcregistryv1beta1.MsgRevokeVC{
				HolderAddress: clientCtx.GetFromAddress().String(),
				VcId:          vcID,
				ReasonText:    reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("reason", "", "Optional reason for revocation")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAdminRevokeVC revokes a credential (governance)
func CmdAdminRevokeVC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin-revoke-vc [vc-id] [reason] [evidence]",
		Short: "Revoke a VC via governance (admin only)",
		Long: `Administratively revoke a verifiable credential through governance.

Examples:
  aurad tx vcregistry admin-revoke-vc vc-123456 REVOCATION_REASON_FRAUD_DETECTED ipfs://Qm... --from governance
  aurad tx vcregistry admin-revoke-vc vc-789012 REVOCATION_REASON_SECURITY_COMPROMISE "Security incident report" --from governance

Revocation Reasons:
  REVOCATION_REASON_FRAUD_DETECTED
  REVOCATION_REASON_CS_BELOW_THRESHOLD
  REVOCATION_REASON_IR_INVALIDATED
  REVOCATION_REASON_GOVERNANCE
  REVOCATION_REASON_SECURITY_COMPROMISE
  REVOCATION_REASON_POLICY_CHANGE

Note: This command requires governance authority
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vcID := args[0]
			reasonStr := args[1]
			evidence := args[2]

			reason, err := parseRevocationReason(reasonStr)
			if err != nil {
				return err
			}

			msg := &vcregistryv1beta1.MsgAdminRevokeVC{
				Authority: clientCtx.GetFromAddress().String(),
				VcId:      vcID,
				Reason:    reason,
				Evidence:  evidence,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSuspendVC temporarily suspends a credential
func CmdSuspendVC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend-vc [vc-id] [reason] [duration-days]",
		Short: "Suspend a VC temporarily (governance only)",
		Long: `Temporarily suspend a verifiable credential through governance.

Examples:
  aurad tx vcregistry suspend-vc vc-123456 "Under investigation" 30 --from governance
  aurad tx vcregistry suspend-vc vc-789012 "Pending review" 0 --from governance (indefinite)

Note:
  - Duration of 0 means indefinite suspension
  - Suspended VCs can be reactivated
  - This command requires governance authority
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vcID := args[0]
			reason := args[1]
			durationDays, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}

			msg := &vcregistryv1beta1.MsgSuspendVC{
				Authority:             clientCtx.GetFromAddress().String(),
				VcId:                  vcID,
				Reason:                reason,
				SuspensionDurationDays: durationDays,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdReactivateVC reactivates a suspended credential
func CmdReactivateVC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reactivate-vc [vc-id]",
		Short: "Reactivate a suspended VC (governance only)",
		Long: `Reactivate a suspended verifiable credential through governance.

Examples:
  aurad tx vcregistry reactivate-vc vc-123456 --from governance

Note: This command requires governance authority
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vcID := args[0]

			msg := &vcregistryv1beta1.MsgReactivateVC{
				Authority: clientCtx.GetFromAddress().String(),
				VcId:      vcID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateVCPolicy creates a new VC policy (governance)
func CmdCreateVCPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-policy [vc-type-name] [vc-type-enum] [cs-threshold]",
		Short: "Create a new VC policy (governance only)",
		Long: `Create a new verifiable credential policy through governance.

Examples:
  aurad tx vcregistry create-policy "Verified Developer" VC_TYPE_CUSTOM 7500 \
    --required-ir-ids "ir-basic-verification,ir-developer-test" \
    --expiry-days 365 \
    --singleton \
    --metadata-uri "ipfs://Qm..." \
    --from governance

  aurad tx vcregistry create-policy "High Assurance Focus" VC_TYPE_HIGH_ASSURANCE_FOCUS 9000 \
    --required-arena "high-assurance" \
    --required-arena-score 8500 \
    --expiry-days 180 \
    --annual-renewal \
    --from governance

Flags:
  --required-ir-ids: Comma-separated list of required IR IDs
  --required-arena: Required arena name
  --required-arena-score: Minimum arena score (uint64)
  --expiry-days: Expiration duration in days (0 = no expiry)
  --singleton: Only one active VC per user
  --annual-renewal: Requires annual renewal
  --metadata-uri: URI with policy details

Note: This command requires governance authority
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vcTypeName := args[0]
			vcTypeEnumStr := args[1]
			csThreshold, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid CS threshold: %w", err)
			}

			vcTypeEnum, err := parseVCType(vcTypeEnumStr)
			if err != nil {
				return err
			}

			// Parse required IR IDs
			requiredIRIDsStr, _ := cmd.Flags().GetString("required-ir-ids")
			var requiredIRIDs []string
			if requiredIRIDsStr != "" {
				requiredIRIDs = strings.Split(requiredIRIDsStr, ",")
				for i := range requiredIRIDs {
					requiredIRIDs[i] = strings.TrimSpace(requiredIRIDs[i])
				}
			}

			requiredArena, _ := cmd.Flags().GetString("required-arena")
			requiredArenaScore, _ := cmd.Flags().GetUint64("required-arena-score")
			expiryDays, _ := cmd.Flags().GetUint64("expiry-days")
			singleton, _ := cmd.Flags().GetBool("singleton")
			annualRenewal, _ := cmd.Flags().GetBool("annual-renewal")
			metadataURI, _ := cmd.Flags().GetString("metadata-uri")

			msg := &vcregistryv1beta1.MsgCreateVCPolicy{
				Authority:              clientCtx.GetFromAddress().String(),
				VcTypeName:             vcTypeName,
				VcTypeEnum:             vcTypeEnum,
				CsThreshold:            csThreshold,
				RequiredIrIds:          requiredIRIDs,
				RequiredArena:          requiredArena,
				RequiredArenaScore:     requiredArenaScore,
				ExpiryDurationDays:     expiryDays,
				Singleton:              singleton,
				RequiresAnnualRenewal:  annualRenewal,
				MetadataUri:            metadataURI,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("required-ir-ids", "", "Comma-separated list of required IR IDs")
	cmd.Flags().String("required-arena", "", "Required arena name")
	cmd.Flags().Uint64("required-arena-score", 0, "Minimum arena score")
	cmd.Flags().Uint64("expiry-days", 365, "Expiration duration in days (0 = no expiry)")
	cmd.Flags().Bool("singleton", false, "Only one active VC per user")
	cmd.Flags().Bool("annual-renewal", false, "Requires annual renewal")
	cmd.Flags().String("metadata-uri", "", "URI with policy details")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateVCPolicy updates an existing VC policy
func CmdUpdateVCPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-policy [vc-type-name] [cs-threshold]",
		Short: "Update a VC policy (governance only)",
		Long: `Update an existing VC policy through governance. Creates a new version.

Examples:
  aurad tx vcregistry update-policy "Verified Developer" 8000 \
    --required-ir-ids "ir-basic-verification,ir-developer-test,ir-advanced" \
    --expiry-days 730 \
    --from governance

Note: This creates a new policy version. This command requires governance authority.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vcTypeName := args[0]
			csThreshold, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid CS threshold: %w", err)
			}

			// Parse optional flags
			requiredIRIDsStr, _ := cmd.Flags().GetString("required-ir-ids")
			var requiredIRIDs []string
			if requiredIRIDsStr != "" {
				requiredIRIDs = strings.Split(requiredIRIDsStr, ",")
				for i := range requiredIRIDs {
					requiredIRIDs[i] = strings.TrimSpace(requiredIRIDs[i])
				}
			}

			requiredArena, _ := cmd.Flags().GetString("required-arena")
			requiredArenaScore, _ := cmd.Flags().GetUint64("required-arena-score")
			expiryDays, _ := cmd.Flags().GetUint64("expiry-days")
			singleton, _ := cmd.Flags().GetBool("singleton")
			annualRenewal, _ := cmd.Flags().GetBool("annual-renewal")
			metadataURI, _ := cmd.Flags().GetString("metadata-uri")

			msg := &vcregistryv1beta1.MsgUpdateVCPolicy{
				Authority:              clientCtx.GetFromAddress().String(),
				VcTypeName:             vcTypeName,
				CsThreshold:            csThreshold,
				RequiredIrIds:          requiredIRIDs,
				RequiredArena:          requiredArena,
				RequiredArenaScore:     requiredArenaScore,
				ExpiryDurationDays:     expiryDays,
				Singleton:              singleton,
				RequiresAnnualRenewal:  annualRenewal,
				MetadataUri:            metadataURI,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("required-ir-ids", "", "Comma-separated list of required IR IDs")
	cmd.Flags().String("required-arena", "", "Required arena name")
	cmd.Flags().Uint64("required-arena-score", 0, "Minimum arena score")
	cmd.Flags().Uint64("expiry-days", 365, "Expiration duration in days (0 = no expiry)")
	cmd.Flags().Bool("singleton", false, "Only one active VC per user")
	cmd.Flags().Bool("annual-renewal", false, "Requires annual renewal")
	cmd.Flags().String("metadata-uri", "", "URI with policy details")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdDeprecateVCPolicy deprecates a VC policy
func CmdDeprecateVCPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deprecate-policy [vc-type-name] [reason]",
		Short: "Deprecate a VC policy (governance only)",
		Long: `Deprecate a VC policy through governance. No new VCs can be minted.

Examples:
  aurad tx vcregistry deprecate-policy "Old Developer VC" "Replaced by new version" --from governance

Note: Existing VCs remain valid. This command requires governance authority.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vcTypeName := args[0]
			reason := args[1]

			msg := &vcregistryv1beta1.MsgDeprecateVCPolicy{
				Authority:  clientCtx.GetFromAddress().String(),
				VcTypeName: vcTypeName,
				Reason:     reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRegisterDID registers a new DID document
func CmdRegisterDID() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-did [did]",
		Aliases: []string{"reg-did", "create-did"},
		Short:   "Register a new DID document",
		Long: `Register a new Decentralized Identifier (DID) document.

Examples:
  aurad tx vcregistry register-did did:aura:mainnet:user123 \
    --metadata-uri "ipfs://Qm..." \
    --verification-method "key1:Ed25519VerificationKey2020:pubkey_hex" \
    --from alice

  aurad tx vcregistry register-did did:aura:testnet:user456 \
    --metadata-uri "https://example.com/did.json" \
    --verification-method "key1:Ed25519VerificationKey2020:pubkey1,key2:EcdsaSecp256k1VerificationKey2019:pubkey2" \
    --from bob

Verification Method Format:
  id:type:public_key_hex

Note: The DID must be unique and follow the did:aura:network:identifier format
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			metadataURI, _ := cmd.Flags().GetString("metadata-uri")
			vmMethodsStr, _ := cmd.Flags().GetString("verification-method")

			// Parse verification methods
			var verificationMethods []*vcregistryv1beta1.VerificationMethod
			if vmMethodsStr != "" {
				methods := strings.Split(vmMethodsStr, ",")
				for _, method := range methods {
					parts := strings.SplitN(strings.TrimSpace(method), ":", 3)
					if len(parts) != 3 {
						return fmt.Errorf("invalid verification method format: %s", method)
					}
					pubKeyBytes, err := hex.DecodeString(parts[2])
					if err != nil {
						return fmt.Errorf("invalid public key hex: %w", err)
					}
					verificationMethods = append(verificationMethods, &vcregistryv1beta1.VerificationMethod{
						Id:        parts[0],
						Type:      parts[1],
						Controller: did,
						PublicKey: pubKeyBytes,
					})
				}
			}

			msg := &vcregistryv1beta1.MsgRegisterDID{
				Controller:            clientCtx.GetFromAddress().String(),
				Did:                   did,
				VerificationMethods:   verificationMethods,
				MetadataUri:           metadataURI,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("metadata-uri", "", "URI to full DID document (IPFS or HTTPS)")
	cmd.Flags().String("verification-method", "", "Verification methods (format: id:type:pubkey_hex, comma-separated)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateDIDDocument updates a DID document
func CmdUpdateDIDDocument() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-did [did]",
		Short: "Update a DID document",
		Long: `Update an existing DID document.

Examples:
  aurad tx vcregistry update-did did:aura:mainnet:user123 \
    --metadata-uri "ipfs://QmNew..." \
    --verification-method "key1:Ed25519VerificationKey2020:new_pubkey_hex" \
    --from alice

Note: Only the controller can update the DID document
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			did := args[0]
			metadataURI, _ := cmd.Flags().GetString("metadata-uri")
			vmMethodsStr, _ := cmd.Flags().GetString("verification-method")

			// Parse verification methods
			var verificationMethods []*vcregistryv1beta1.VerificationMethod
			if vmMethodsStr != "" {
				methods := strings.Split(vmMethodsStr, ",")
				for _, method := range methods {
					parts := strings.SplitN(strings.TrimSpace(method), ":", 3)
					if len(parts) != 3 {
						return fmt.Errorf("invalid verification method format: %s", method)
					}
					pubKeyBytes, err := hex.DecodeString(parts[2])
					if err != nil {
						return fmt.Errorf("invalid public key hex: %w", err)
					}
					verificationMethods = append(verificationMethods, &vcregistryv1beta1.VerificationMethod{
						Id:        parts[0],
						Type:      parts[1],
						Controller: did,
						PublicKey: pubKeyBytes,
					})
				}
			}

			msg := &vcregistryv1beta1.MsgUpdateDIDDocument{
				Controller:            clientCtx.GetFromAddress().String(),
				Did:                   did,
				VerificationMethods:   verificationMethods,
				MetadataUri:           metadataURI,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("metadata-uri", "", "URI to full DID document (IPFS or HTTPS)")
	cmd.Flags().String("verification-method", "", "Verification methods (format: id:type:pubkey_hex, comma-separated)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// Helper functions

func parseVCType(s string) (vcregistryv1beta1.VCType, error) {
	switch strings.ToUpper(s) {
	case "VC_TYPE_VERIFIED_HUMAN":
		return vcregistryv1beta1.VCType_VC_TYPE_VERIFIED_HUMAN, nil
	case "VC_TYPE_AGE_OVER_18":
		return vcregistryv1beta1.VCType_VC_TYPE_AGE_OVER_18, nil
	case "VC_TYPE_AGE_OVER_21":
		return vcregistryv1beta1.VCType_VC_TYPE_AGE_OVER_21, nil
	case "VC_TYPE_RESIDENT_OF":
		return vcregistryv1beta1.VCType_VC_TYPE_RESIDENT_OF, nil
	case "VC_TYPE_BIOMETRIC_AUTH":
		return vcregistryv1beta1.VCType_VC_TYPE_BIOMETRIC_AUTH, nil
	case "VC_TYPE_KYC_VERIFICATION":
		return vcregistryv1beta1.VCType_VC_TYPE_KYC_VERIFICATION, nil
	case "VC_TYPE_NOTARY_PUBLIC":
		return vcregistryv1beta1.VCType_VC_TYPE_NOTARY_PUBLIC, nil
	case "VC_TYPE_PROFESSIONAL_LICENSE":
		return vcregistryv1beta1.VCType_VC_TYPE_PROFESSIONAL_LICENSE, nil
	case "VC_TYPE_BIOMETRIC_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_BIOMETRIC_FOCUS, nil
	case "VC_TYPE_SOCIAL_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_SOCIAL_FOCUS, nil
	case "VC_TYPE_GEOLOCATION_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_GEOLOCATION_FOCUS, nil
	case "VC_TYPE_HIGH_ASSURANCE_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_HIGH_ASSURANCE_FOCUS, nil
	case "VC_TYPE_POSSESSION_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_POSSESSION_FOCUS, nil
	case "VC_TYPE_KNOWLEDGE_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_KNOWLEDGE_FOCUS, nil
	case "VC_TYPE_PERSISTENCE_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_PERSISTENCE_FOCUS, nil
	case "VC_TYPE_SPECIALIZED_FOCUS":
		return vcregistryv1beta1.VCType_VC_TYPE_SPECIALIZED_FOCUS, nil
	case "VC_TYPE_CUSTOM":
		return vcregistryv1beta1.VCType_VC_TYPE_CUSTOM, nil
	default:
		return vcregistryv1beta1.VCType_VC_TYPE_UNSPECIFIED, fmt.Errorf("invalid VC type: %s", s)
	}
}

func parseRevocationReason(s string) (vcregistryv1beta1.RevocationReason, error) {
	switch strings.ToUpper(s) {
	case "REVOCATION_REASON_USER_REQUEST":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_USER_REQUEST, nil
	case "REVOCATION_REASON_FRAUD_DETECTED":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_FRAUD_DETECTED, nil
	case "REVOCATION_REASON_CS_BELOW_THRESHOLD":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_CS_BELOW_THRESHOLD, nil
	case "REVOCATION_REASON_IR_INVALIDATED":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_IR_INVALIDATED, nil
	case "REVOCATION_REASON_EXPIRED":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_EXPIRED, nil
	case "REVOCATION_REASON_GOVERNANCE":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_GOVERNANCE, nil
	case "REVOCATION_REASON_SECURITY_COMPROMISE":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_SECURITY_COMPROMISE, nil
	case "REVOCATION_REASON_POLICY_CHANGE":
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_POLICY_CHANGE, nil
	default:
		return vcregistryv1beta1.RevocationReason_REVOCATION_REASON_UNSPECIFIED, fmt.Errorf("invalid revocation reason: %s", s)
	}
}

func parseAttributeType(s string) (vcregistryv1beta1.AttributeType, error) {
	switch strings.ToUpper(s) {
	case "ATTRIBUTE_TYPE_FULL_NAME":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_FULL_NAME, nil
	case "ATTRIBUTE_TYPE_FIRST_NAME":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_FIRST_NAME, nil
	case "ATTRIBUTE_TYPE_LAST_NAME":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_LAST_NAME, nil
	case "ATTRIBUTE_TYPE_DATE_OF_BIRTH":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_DATE_OF_BIRTH, nil
	case "ATTRIBUTE_TYPE_AGE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_AGE, nil
	case "ATTRIBUTE_TYPE_GENDER":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_GENDER, nil
	case "ATTRIBUTE_TYPE_EMAIL":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_EMAIL, nil
	case "ATTRIBUTE_TYPE_PHONE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_PHONE, nil
	case "ATTRIBUTE_TYPE_ADDRESS_FULL":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL, nil
	case "ATTRIBUTE_TYPE_ADDRESS_STREET":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STREET, nil
	case "ATTRIBUTE_TYPE_ADDRESS_CITY":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_ADDRESS_CITY, nil
	case "ATTRIBUTE_TYPE_ADDRESS_STATE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STATE, nil
	case "ATTRIBUTE_TYPE_ADDRESS_ZIP":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_ADDRESS_ZIP, nil
	case "ATTRIBUTE_TYPE_ADDRESS_COUNTRY":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_ADDRESS_COUNTRY, nil
	case "ATTRIBUTE_TYPE_PASSPORT_NUMBER":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_PASSPORT_NUMBER, nil
	case "ATTRIBUTE_TYPE_DRIVERS_LICENSE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_DRIVERS_LICENSE, nil
	case "ATTRIBUTE_TYPE_SSN":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_SSN, nil
	case "ATTRIBUTE_TYPE_TAX_ID":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_TAX_ID, nil
	case "ATTRIBUTE_TYPE_HEIGHT":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_HEIGHT, nil
	case "ATTRIBUTE_TYPE_WEIGHT":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_WEIGHT, nil
	case "ATTRIBUTE_TYPE_EYE_COLOR":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_EYE_COLOR, nil
	case "ATTRIBUTE_TYPE_HAIR_COLOR":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_HAIR_COLOR, nil
	case "ATTRIBUTE_TYPE_OCCUPATION":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_OCCUPATION, nil
	case "ATTRIBUTE_TYPE_EMPLOYER":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_EMPLOYER, nil
	case "ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE, nil
	case "ATTRIBUTE_TYPE_EDUCATION_LEVEL":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_EDUCATION_LEVEL, nil
	case "ATTRIBUTE_TYPE_DEGREE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_DEGREE, nil
	case "ATTRIBUTE_TYPE_SCUBA_CERTIFIED":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_SCUBA_CERTIFIED, nil
	case "ATTRIBUTE_TYPE_PILOTS_LICENSE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_PILOTS_LICENSE, nil
	case "ATTRIBUTE_TYPE_SECURITY_CLEARANCE":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_SECURITY_CLEARANCE, nil
	case "ATTRIBUTE_TYPE_CUSTOM":
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_CUSTOM, nil
	default:
		return vcregistryv1beta1.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED, fmt.Errorf("invalid attribute type: %s", s)
	}
}

func parseDisclosurePolicyMode(s string) (vcregistryv1beta1.DisclosurePolicyMode, error) {
	switch strings.ToUpper(s) {
	case "DISCLOSURE_POLICY_MODE_DENY":
		return vcregistryv1beta1.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY, nil
	case "DISCLOSURE_POLICY_MODE_ASK":
		return vcregistryv1beta1.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ASK, nil
	case "DISCLOSURE_POLICY_MODE_ALLOW":
		return vcregistryv1beta1.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_ALLOW, nil
	case "DISCLOSURE_POLICY_MODE_CONDITIONAL":
		return vcregistryv1beta1.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_CONDITIONAL, nil
	default:
		return vcregistryv1beta1.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY, fmt.Errorf("invalid disclosure policy mode: %s", s)
	}
}
