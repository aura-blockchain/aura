package cli

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// NewTxCmd builds the tx sub-tree for x/aiassistant.
func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "AI assistant transaction commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		newRegisterCmd(),
		newUpdateLocalesCmd(),
		newHeartbeatCmd(),
		newReportMisbehaviorCmd(),
	)
	return cmd
}

func newRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [assistant-address]",
		Short: "Register a new AI assistant",
		Long: `Registers an AI assistant with stake, locales, and model metadata.
The --locales flag accepts a comma-separated list (e.g., "en-us,es-es").`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			localesStr, _ := cmd.Flags().GetString("locales")
			locales := splitAndNormalize(localesStr)
			if len(locales) == 0 {
				return fmt.Errorf("at least one locale is required")
			}

			stakeStr, _ := cmd.Flags().GetString("stake")
			if stakeStr == "" {
				return fmt.Errorf("--stake is required (e.g., 5000000uaura)")
			}
			stakeCoin, err := sdk.ParseCoinNormalized(stakeStr)
			if err != nil {
				return err
			}

			sponsorshipStr, _ := cmd.Flags().GetString("sponsorship")
			var sponsorshipCoin sdk.Coin
			if sponsorshipStr == "" {
				sponsorshipCoin = sdk.NewCoin(stakeCoin.Denom, sdkmath.ZeroInt())
			} else {
				sponsorshipCoin, err = sdk.ParseCoinNormalized(sponsorshipStr)
				if err != nil {
					return err
				}
			}

			modelHash, _ := cmd.Flags().GetString("model-hash")
			fingerprint, _ := cmd.Flags().GetString("fingerprint")

			msg := &types.MsgRegisterAssistant{
				AssistantAddress:  args[0],
				OwnerAddress:      clientCtx.GetFromAddress().String(),
				Locales:           locales,
				ModelHash:         modelHash,
				ApiKeyFingerprint: fingerprint,
				Stake: types.Balance{
					Denom:  stakeCoin.Denom,
					Amount: stakeCoin.Amount,
				},
				Sponsorship: types.Balance{
					Denom:  sponsorshipCoin.Denom,
					Amount: sponsorshipCoin.Amount,
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("locales", "en-us", "Comma-separated locale codes served by the assistant")
	cmd.Flags().String("stake", "", "Stake to bond (e.g., 5000000uaura)")
	cmd.Flags().String("sponsorship", "", "Optional sponsorship balance (defaults to 0)")
	cmd.Flags().String("model-hash", "", "Model hash/version metadata")
	cmd.Flags().String("fingerprint", "", "API key fingerprint (never the raw key)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newUpdateLocalesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-locales [assistant-address]",
		Short: "Update the locales for an assistant (owner only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			localesStr, _ := cmd.Flags().GetString("locales")
			locales := splitAndNormalize(localesStr)
			if len(locales) == 0 {
				return fmt.Errorf("locales cannot be empty")
			}
			msg := &types.MsgUpdateLocales{
				AssistantAddress: args[0],
				OwnerAddress:     clientCtx.GetFromAddress().String(),
				Locales:          locales,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String("locales", "en-us", "Comma-separated locale codes")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newHeartbeatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "heartbeat [assistant-address]",
		Short: "Send a heartbeat proving the assistant is online",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			attHash, _ := cmd.Flags().GetString("attestation-hash")
			msg := &types.MsgHeartbeat{
				AssistantAddress: args[0],
				OperatorAddress:  clientCtx.GetFromAddress().String(),
				AttestationHash:  attHash,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String("attestation-hash", "", "Optional attestation hash committed by the assistant")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newReportMisbehaviorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report-misbehavior [assistant-address]",
		Short: "Report a malicious or offline assistant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			infraction, _ := cmd.Flags().GetString("infraction")
			if infraction == "" {
				return fmt.Errorf("--infraction is required")
			}
			evidence, _ := cmd.Flags().GetString("evidence")
			msg := &types.MsgReportMisbehavior{
				Reporter:         clientCtx.GetFromAddress().String(),
				AssistantAddress: args[0],
				Infraction:       infraction,
				EvidenceHash:     evidence,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String("infraction", "", "Reason (double-sign, false attestation, offline, etc.)")
	cmd.Flags().String("evidence", "", "Optional hash linking to off-chain evidence")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func splitAndNormalize(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	locales := make([]string, 0, len(parts))
	seen := make(map[string]struct{})
	for _, p := range parts {
		loc := strings.TrimSpace(strings.ToLower(p))
		if loc == "" {
			continue
		}
		if _, ok := seen[loc]; ok {
			continue
		}
		seen[loc] = struct{}{}
		locales = append(locales, loc)
	}
	return locales
}
