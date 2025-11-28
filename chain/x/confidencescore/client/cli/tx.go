package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// GetTxCmd returns the transaction commands for the confidencescore module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "confidencescore",
		Aliases:                    []string{"cs", "score"},
		Short:                      "Confidence score transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		CmdRecordIRCompletion(),
		CmdRecalculateScore(),
		CmdSlashScore(),
		CmdAppealSlash(),
		CmdResolveAppeal(),
	)

	return txCmd
}

// CmdRecordIRCompletion records an IR completion with assistant attestation
func CmdRecordIRCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record-completion [wallet-address] [ir-id] [proof-hash] [verifier-hash]",
		Short: "Record an Inclusion Routine completion (assistant only)",
		Long: `Record the completion of an Inclusion Routine (IR) by a user. This command can only be executed by
an AI assistant that has verified the user's completion.

Arguments:
  wallet-address: The Aura wallet address of the user completing the IR
  ir-id:          The IR identifier (e.g., IR-102, IR-000)
  proof-hash:     SHA256 hash of the proof data (hex encoded)
  verifier-hash:  SHA256 hash of the verifier plugin used (hex encoded)

The command calculates and applies:
- Base score from IR definition
- Velocity bonus (1.0-1.25x based on time since last IR)
- Arena multiplier (1.0-1.5x based on arena focus)
- Jackpot bonus (rare 5x or 25x multiplier)

Example:
  $ aurad tx confidencescore record-completion \
      aura1user... IR-102 \
      a1b2c3d4e5f6... 9f8e7d6c5b4a... \
      --from assistant-key \
      --chain-id aura-testnet-1

Response includes:
- Score earned (with all multipliers applied)
- New total confidence score
- Whether verification threshold was achieved (>= 10,000)
- Multipliers applied`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Decode proof hash
			proofHash, err := hex.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("invalid proof hash (must be hex): %w", err)
			}

			// Decode verifier hash
			verifierHash, err := hex.DecodeString(args[3])
			if err != nil {
				return fmt.Errorf("invalid verifier hash (must be hex): %w", err)
			}

			msg := &v1beta1.MsgRecordIRCompletion{
				WalletAddress:    args[0],
				IrId:             args[1],
				ProofHash:        proofHash,
				VerifierHash:     verifierHash,
				AssistantAddress: clientCtx.GetFromAddress().String(),
				Timestamp:        timestamppb.New(time.Now()),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdRecalculateScore recalculates a user's total score (admin/governance only)
func CmdRecalculateScore() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recalculate-score [wallet-address]",
		Short: "Recalculate a user's confidence score (governance only)",
		Long: `Recalculate a user's total confidence score from their IR completion history. This is an
administrative command that can only be executed by the governance module.

Use cases:
- Fix score discrepancies after a bug fix
- Audit score calculation accuracy
- Recover from database corruption
- Apply retroactive parameter changes

The command will:
1. Sum all verified IR completions
2. Apply current multiplier rules
3. Detect any discrepancies
4. Update the user's total score

Example:
  $ aurad tx confidencescore recalculate-score aura1abc... \
      --from governance-key \
      --chain-id aura-mainnet-1

Response includes:
- Previous score
- Recalculated score
- List of any discrepancies found`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgRecalculateScore{
				WalletAddress: args[0],
				Authority:     clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdSlashScore slashes a user's score for fraud
func CmdSlashScore() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slash [wallet-address] [ir-id] [slash-amount] [reason]",
		Short: "Slash a user's confidence score for fraud (governance only)",
		Long: `Slash (deduct) confidence score points from a user due to fraud or policy violations.
This is a governance function typically triggered by fraud detection or appeals process.

Arguments:
  wallet-address: The wallet address to slash
  ir-id:          The IR being disputed (if applicable)
  slash-amount:   Number of CS points to deduct
  reason:         Reason for slash (fraud_detected, false_attestation, collusion, duplicate_completion)

Flags:
  --evidence:     IPFS hash or metadata link to supporting evidence

The slash will:
- Deduct the specified CS points
- Record the slash with full metadata
- Set an appeal deadline (configurable, default 30 days)
- Potentially revoke verification status if score drops below 10,000

Example:
  $ aurad tx confidencescore slash \
      aura1user... IR-305 5000 fraud_detected \
      --evidence QmX3d9f2h... \
      --from governance-key \
      --chain-id aura-mainnet-1

Response includes:
- Previous score
- New score after slash
- Whether verification was revoked
- Slash transaction hash (for appeals)`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			slashAmount, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid slash amount: %w", err)
			}

			evidence, _ := cmd.Flags().GetString("evidence")

			msg := &v1beta1.MsgSlashScore{
				WalletAddress: args[0],
				IrId:          args[1],
				SlashAmount:   slashAmount,
				Reason:        args[3],
				Authority:     clientCtx.GetFromAddress().String(),
				Evidence:      evidence,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("evidence", "", "IPFS hash or URL to evidence supporting the slash")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdAppealSlash allows a user to appeal a slash
func CmdAppealSlash() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appeal [slash-tx-hash] [deposit]",
		Short: "Appeal a confidence score slash",
		Long: `File an appeal against a confidence score slash. Users who believe they were slashed
incorrectly can submit an appeal with supporting evidence.

Arguments:
  slash-tx-hash:  Transaction hash of the slash to appeal
  deposit:        Appeal deposit amount (e.g., "1000aeq") - returned if appeal succeeds

Flags:
  --evidence:     IPFS hash or URL to evidence supporting the appeal

The appeal process:
1. User submits appeal with deposit
2. Governance reviews evidence within review period
3. If successful: score restored, deposit returned
4. If unsuccessful: slash stands, deposit forfeit

Requirements:
- Must appeal before the deadline (default 30 days after slash)
- Must provide appeal deposit (default 1000 AURA, configured in params)
- Can only appeal once per slash

Example:
  $ aurad tx confidencescore appeal \
      A1B2C3D4E5F6... 1000aeq \
      --evidence QmY4e8g3k... \
      --from user-key \
      --chain-id aura-mainnet-1

Response includes:
- Whether appeal was accepted
- Review deadline for governance`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			evidence, _ := cmd.Flags().GetString("evidence")

			msg := &v1beta1.MsgAppealSlash{
				WalletAddress: clientCtx.GetFromAddress().String(),
				SlashTxHash:   args[0],
				Deposit:       args[1],
				Evidence:      evidence,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("evidence", "", "IPFS hash or URL to evidence supporting the appeal")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdResolveAppeal resolves an appeal (governance only)
func CmdResolveAppeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-appeal [wallet-address] [slash-tx-hash] [restore]",
		Short: "Resolve a slash appeal (governance only)",
		Long: `Resolve a user's appeal of a confidence score slash. This is a governance function that
determines whether to restore the slashed score.

Arguments:
  wallet-address: The wallet address that filed the appeal
  slash-tx-hash:  Transaction hash of the original slash
  restore:        Whether to restore the score (true/false)

Flags:
  --notes:        Resolution notes explaining the decision

The resolution will:
- If restore=true: Restore slashed score, return deposit
- If restore=false: Keep slash, forfeit deposit
- Record the decision with notes
- Close the appeal

Example (restore score):
  $ aurad tx confidencescore resolve-appeal \
      aura1user... A1B2C3D4E5F6... true \
      --notes "Evidence validates user claim" \
      --from governance-key \
      --chain-id aura-mainnet-1

Example (deny appeal):
  $ aurad tx confidencescore resolve-appeal \
      aura1user... A1B2C3D4E5F6... false \
      --notes "Insufficient evidence provided" \
      --from governance-key \
      --chain-id aura-mainnet-1

Response includes:
- Amount restored (if applicable)
- Whether deposit was returned`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			restore, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("invalid restore value (must be true/false): %w", err)
			}

			notes, _ := cmd.Flags().GetString("notes")

			msg := &v1beta1.MsgResolveAppeal{
				WalletAddress:   args[0],
				SlashTxHash:     args[1],
				RestoreScore:    restore,
				Authority:       clientCtx.GetFromAddress().String(),
				ResolutionNotes: notes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("notes", "", "Resolution notes explaining the governance decision")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
