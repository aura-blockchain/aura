package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "inclusionroutines",
		Short:                      "Inclusion Routines transaction subcommands",
		Aliases:                    []string{"ir"},
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreateIR(),
		CmdUpdateIR(),
		CmdDeleteIR(),
		CmdSetIRPrerequisites(),
		CmdSetIRRateLimit(),
		CmdSuspendIR(),
		CmdActivateIR(),
	)

	return cmd
}

// CmdCreateIR creates a new Inclusion Routine definition
func CmdCreateIR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-ir [id] [name] [arena] [description] [score] [poi-reward]",
		Short: "Create a new Inclusion Routine (IR) definition",
		Long: `Create a new Inclusion Routine for identity verification tasks.

Arenas:
  0: UNSPECIFIED
  1: ANCHOR - Core identity verification (government ID, biometric basics)
  2: BIOMETRIC - Advanced biometric verification (fingerprint, face, voice)
  3: POSSESSION - Device/asset possession verification
  4: KNOWLEDGE - Knowledge-based verification (secrets, answers)
  5: SOCIAL - Social graph verification
  6: GEOLOCATION - Location-based verification
  7: HIGH_ASSURANCE - High-security verification tasks
  8: PERSISTENCE - Long-term verification maintenance
  9: SPECIALIZED - Custom/specialized verification tasks

Privacy Tiers:
  0: UNSPECIFIED
  1: LOW - Minimal privacy requirements
  2: MEDIUM - Moderate privacy protection
  3: HIGH - Maximum privacy protection

Examples:
  aurad tx ir create-ir "gov-id-verify" "Government ID Verification" 1 "Verify government-issued ID" 100 50 \
    --locale-tags "US,UK,EU" \
    --privacy-tier 3 \
    --version "1.0.0" \
    --metadata-hash "0xabc123..." \
    --activation-height 1000 \
    --sunset-height 100000 \
    --from alice

  aurad tx ir create-ir "biometric-face" "Facial Recognition" 2 "Advanced facial biometric verification" 200 100 \
    --locale-tags "GLOBAL" \
    --privacy-tier 3 \
    --version "2.1.0" \
    --from governance
`,
		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id := args[0]
			name := args[1]
			arenaInt, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid arena: %w", err)
			}
			arena := v1beta1.Arena(arenaInt)
			description := args[3]
			score, err := strconv.ParseInt(args[4], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid score: %w", err)
			}
			poiReward, err := strconv.ParseInt(args[5], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid poi-reward: %w", err)
			}

			// Parse optional flags
			localeTags, err := cmd.Flags().GetString("locale-tags")
			if err != nil {
				return err
			}
			var localeTagsList []string
			if localeTags != "" {
				localeTagsList = strings.Split(localeTags, ",")
			}

			privacyTierInt, err := cmd.Flags().GetInt32("privacy-tier")
			if err != nil {
				return err
			}
			privacyTier := v1beta1.PrivacyTier(privacyTierInt)

			version, err := cmd.Flags().GetString("version")
			if err != nil {
				return err
			}

			metadataHash, err := cmd.Flags().GetString("metadata-hash")
			if err != nil {
				return err
			}

			activationHeight, err := cmd.Flags().GetInt64("activation-height")
			if err != nil {
				return err
			}

			sunsetHeight, err := cmd.Flags().GetInt64("sunset-height")
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgCreateIR{
				Authority:        clientCtx.GetFromAddress().String(),
				Id:               id,
				Name:             name,
				Arena:            arena,
				Description:      description,
				Score:            score,
				PoiReward:        poiReward,
				LocaleTags:       localeTagsList,
				PrivacyTier:      privacyTier,
				Version:          version,
				MetadataHash:     metadataHash,
				ActivationHeight: activationHeight,
				SunsetHeight:     sunsetHeight,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("locale-tags", "", "Comma-separated locale tags (e.g., US,UK,EU,GLOBAL)")
	cmd.Flags().Int32("privacy-tier", 0, "Privacy tier (0=UNSPECIFIED, 1=LOW, 2=MEDIUM, 3=HIGH)")
	cmd.Flags().String("version", "1.0.0", "IR version")
	cmd.Flags().String("metadata-hash", "", "Metadata hash for verification")
	cmd.Flags().Int64("activation-height", 0, "Block height when IR becomes active")
	cmd.Flags().Int64("sunset-height", 0, "Block height when IR becomes deprecated")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateIR updates an existing IR definition
func CmdUpdateIR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-ir [id]",
		Short: "Update an existing Inclusion Routine definition",
		Long: `Update an existing Inclusion Routine's metadata and parameters.

Examples:
  aurad tx ir update-ir "gov-id-verify" \
    --name "Updated Government ID Verification" \
    --description "Enhanced government ID verification with liveness detection" \
    --score 150 \
    --poi-reward 75 \
    --locale-tags "US,UK,EU,CA" \
    --privacy-tier 3 \
    --version "1.1.0" \
    --metadata-hash "0xdef456..." \
    --sunset-height 200000 \
    --from governance

  aurad tx ir update-ir "biometric-face" \
    --score 250 \
    --poi-reward 125 \
    --from governance
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id := args[0]

			// Parse optional flags - all fields are optional for updates
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}

			score, err := cmd.Flags().GetInt64("score")
			if err != nil {
				return err
			}

			poiReward, err := cmd.Flags().GetInt64("poi-reward")
			if err != nil {
				return err
			}

			localeTags, err := cmd.Flags().GetString("locale-tags")
			if err != nil {
				return err
			}
			var localeTagsList []string
			if localeTags != "" {
				localeTagsList = strings.Split(localeTags, ",")
			}

			privacyTierInt, err := cmd.Flags().GetInt32("privacy-tier")
			if err != nil {
				return err
			}
			privacyTier := v1beta1.PrivacyTier(privacyTierInt)

			version, err := cmd.Flags().GetString("version")
			if err != nil {
				return err
			}

			metadataHash, err := cmd.Flags().GetString("metadata-hash")
			if err != nil {
				return err
			}

			sunsetHeight, err := cmd.Flags().GetInt64("sunset-height")
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgUpdateIR{
				Authority:    clientCtx.GetFromAddress().String(),
				Id:           id,
				Name:         name,
				Description:  description,
				Score:        score,
				PoiReward:    poiReward,
				LocaleTags:   localeTagsList,
				PrivacyTier:  privacyTier,
				Version:      version,
				MetadataHash: metadataHash,
				SunsetHeight: sunsetHeight,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("name", "", "Updated IR name")
	cmd.Flags().String("description", "", "Updated IR description")
	cmd.Flags().Int64("score", 0, "Updated confidence score")
	cmd.Flags().Int64("poi-reward", 0, "Updated POI reward")
	cmd.Flags().String("locale-tags", "", "Updated comma-separated locale tags")
	cmd.Flags().Int32("privacy-tier", 0, "Updated privacy tier")
	cmd.Flags().String("version", "", "Updated version")
	cmd.Flags().String("metadata-hash", "", "Updated metadata hash")
	cmd.Flags().Int64("sunset-height", 0, "Updated sunset height")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdDeleteIR deletes an IR definition
func CmdDeleteIR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-ir [id]",
		Short: "Delete an Inclusion Routine definition",
		Long: `Delete an existing Inclusion Routine definition.

WARNING: This is a permanent operation and should only be used for IRs that are no longer needed.
Consider using suspend-ir for temporary disabling.

Examples:
  aurad tx ir delete-ir "deprecated-ir-v1" --from governance
  aurad tx ir delete-ir "test-ir-001" --from governance
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgDeleteIR{
				Authority: clientCtx.GetFromAddress().String(),
				Id:        args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSetIRPrerequisites sets prerequisite IRs for an IR
func CmdSetIRPrerequisites() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-prerequisites [ir-id] [required-ir-ids]",
		Short: "Set prerequisite IRs that must be completed before this IR",
		Long: `Set the prerequisite Inclusion Routines that users must complete before
attempting this IR. This creates a dependency graph for verification tasks.

Examples:
  aurad tx ir set-prerequisites "advanced-biometric" "gov-id-verify,basic-biometric" --from governance
  aurad tx ir set-prerequisites "high-assurance-verify" "gov-id-verify,biometric-face,device-possession" --from governance
  aurad tx ir set-prerequisites "basic-verify" "" --from governance (clear prerequisites)

Use cases:
  - Require basic verification before advanced tasks
  - Build progressive trust levels
  - Ensure proper verification order
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			irID := args[0]
			var requiredIRIds []string
			if args[1] != "" {
				requiredIRIds = strings.Split(args[1], ",")
			}

			msg := &v1beta1.MsgSetIRPrerequisites{
				Authority:     clientCtx.GetFromAddress().String(),
				IrId:          irID,
				RequiredIrIds: requiredIRIds,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSetIRRateLimit sets rate limits for an IR
func CmdSetIRRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-rate-limit [ir-id] [per-wallet-hour] [per-wallet-day] [per-block-global]",
		Short: "Set rate limits for an Inclusion Routine",
		Long: `Set rate limits to prevent abuse and ensure fair access to verification tasks.

Rate limit types:
  - per-wallet-hour: Maximum attempts per wallet per hour
  - per-wallet-day: Maximum attempts per wallet per day
  - per-block-global: Maximum total attempts per block across all users

Examples:
  aurad tx ir set-rate-limit "gov-id-verify" 3 10 100 --from governance
    (3/hour per wallet, 10/day per wallet, 100/block globally)

  aurad tx ir set-rate-limit "simple-captcha" 100 1000 10000 --from governance
    (Higher limits for low-value tasks)

  aurad tx ir set-rate-limit "high-assurance" 1 5 50 --from governance
    (Strict limits for high-value tasks)

Use 0 for unlimited in any category.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			irID := args[0]
			perWalletHour, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid per-wallet-hour: %w", err)
			}
			perWalletDay, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid per-wallet-day: %w", err)
			}
			perBlockGlobal, err := strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid per-block-global: %w", err)
			}

			msg := &v1beta1.MsgSetIRRateLimit{
				Authority:        clientCtx.GetFromAddress().String(),
				IrId:             irID,
				PerWalletPerHour: int32(perWalletHour),
				PerWalletPerDay:  int32(perWalletDay),
				PerBlockGlobal:   int32(perBlockGlobal),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSuspendIR suspends an IR temporarily
func CmdSuspendIR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend-ir [ir-id] [reason]",
		Short: "Suspend an Inclusion Routine temporarily",
		Long: `Temporarily suspend an IR to prevent new completions while investigating issues
or performing maintenance. Existing completions remain valid.

Examples:
  aurad tx ir suspend-ir "gov-id-verify" "Security vulnerability detected - investigating" --from governance
  aurad tx ir suspend-ir "biometric-face" "Upgrading verification service" --from governance
  aurad tx ir suspend-ir "social-graph" "Spam detected - reviewing policies" --from governance

To re-enable, use the activate-ir command.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgSuspendIR{
				Authority: clientCtx.GetFromAddress().String(),
				IrId:      args[0],
				Reason:    args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdActivateIR activates a suspended IR
func CmdActivateIR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate-ir [ir-id]",
		Short: "Activate a suspended Inclusion Routine",
		Long: `Activate a previously suspended IR to allow new completions.

Examples:
  aurad tx ir activate-ir "gov-id-verify" --from governance
  aurad tx ir activate-ir "biometric-face" --from governance

The IR must have been previously suspended. For new IRs, ensure the
activation-height has been reached.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgActivateIR{
				Authority: clientCtx.GetFromAddress().String(),
				IrId:      args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
