package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkmath "cosmossdk.io/math"

	governancev1beta1 "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

// GetTxCmd returns the transaction commands for the governance module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "governance",
		Aliases:                    []string{"gov", "proposal"},
		Short:                      "Governance transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// Proposal commands
		CmdSubmitProposal(),

		// Voting commands
		CmdDeposit(),
		CmdVote(),
		CmdVoteWeighted(),
		CmdRevealSecretVote(),

		// Vote delegation commands
		CmdDelegateVote(),
		CmdUndelegateVote(),

		// Veto commands
		CmdSubmitVeto(),
		CmdCosignVeto(),

		// Execution commands
		CmdExecuteProposal(),

		// Snapshot voting
		CmdSubmitSnapshotVote(),
	)

	return cmd
}

// ============================================================================
// Proposal Commands
// ============================================================================

// CmdSubmitProposal submits a governance proposal
func CmdSubmitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [title] [description] [category]",
		Short: "Submit a governance proposal",
		Long: `Submit a new governance proposal for voting.

Examples:
  aurad tx governance submit-proposal "Community Pool Funding" "Allocate 100,000 AURA for marketing" text --from alice --initial-deposit 1000uaura
  aurad tx governance submit-proposal "Update Min Deposit" "Change min deposit to 500 AURA" parameter-change --from alice --initial-deposit 5000uaura
  aurad tx governance submit-proposal "Upgrade to v2.0" "Upgrade chain to version 2.0" software-upgrade --from alice --initial-deposit 10000uaura
  aurad tx governance submit-proposal "Treasury Spend" "Fund development team" spending --from alice --initial-deposit 2000uaura
  aurad tx governance submit-proposal "Emergency Fix" "Critical security patch" emergency --from alice --initial-deposit 1000uaura --emergency

Categories:
  text              - Simple text proposals for signaling
  parameter-change  - Changes to chain parameters
  software-upgrade  - Software/protocol upgrades
  spending          - Treasury spending proposals
  emergency         - Emergency proposals (fast-tracked)
  constitution      - Constitution changes (highest threshold)

Flags:
  --initial-deposit: Initial deposit amount (e.g., 1000uaura)
  --emergency: Mark as emergency proposal for fast-track voting
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title := args[0]
			description := args[1]
			categoryStr := args[2]

			// Parse category
			category, err := parseProposalCategory(categoryStr)
			if err != nil {
				return err
			}

			// Get flags
			initialDepositStr, _ := cmd.Flags().GetString("initial-deposit")
			isEmergency, _ := cmd.Flags().GetBool("emergency")

			msg := &governancev1beta1.MsgSubmitProposal{
				Title:          title,
				Description:    description,
				Category:       category,
				Proposer:       clientCtx.GetFromAddress().String(),
				InitialDeposit: initialDepositStr,
				IsEmergency:    isEmergency,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("initial-deposit", "0", "Initial deposit amount (e.g., 1000uaura)")
	cmd.Flags().Bool("emergency", false, "Mark proposal as emergency for fast-track voting")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Voting Commands
// ============================================================================

// CmdDeposit adds a deposit to a proposal
func CmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [amount]",
		Short: "Add a deposit to a proposal",
		Long: `Add a deposit to a proposal to help it reach the minimum deposit threshold.

Examples:
  aurad tx governance deposit 1 1000uaura --from alice
  aurad tx governance deposit 5 5000uaura --from bob

Once the minimum deposit is reached, the proposal enters the voting period.
Deposits are refunded if the proposal passes or is rejected.
Deposits are burned if the proposal is vetoed or fails to meet quorum.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			msg := &governancev1beta1.MsgDeposit{
				ProposalId: proposalID,
				Depositor:  clientCtx.GetFromAddress().String(),
				Amount:     args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdVote casts a vote on a proposal
func CmdVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [option]",
		Short: "Cast a vote on a proposal",
		Long: `Cast a vote on an active governance proposal.

Examples:
  aurad tx governance vote 1 yes --from alice
  aurad tx governance vote 2 no --from bob
  aurad tx governance vote 3 abstain --from charlie
  aurad tx governance vote 4 no-with-veto --from dave
  aurad tx governance vote 5 yes --secret --vote-commitment abc123... --from alice

Vote options:
  yes           - Vote in favor of the proposal
  no            - Vote against the proposal
  abstain       - Abstain from voting (counts toward quorum)
  no-with-veto  - Vote against with veto (can block malicious proposals)

Secret ballot:
  Use --secret flag to cast a secret ballot vote
  Provide --vote-commitment with a hash of (option + secret)
  Later reveal with reveal-secret-vote command

Note: Voting power is determined by token holdings at snapshot height.
Delegated voting power is automatically included.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			// Parse vote option
			option, err := parseVoteOption(args[1])
			if err != nil {
				return err
			}

			// Get secret ballot flags
			isSecret, _ := cmd.Flags().GetBool("secret")
			voteCommitment, _ := cmd.Flags().GetString("vote-commitment")

			if isSecret && voteCommitment == "" {
				return fmt.Errorf("--vote-commitment required for secret ballot")
			}

			msg := &governancev1beta1.MsgVote{
				ProposalId:     proposalID,
				Voter:          clientCtx.GetFromAddress().String(),
				Option:         option,
				IsSecret:       isSecret,
				VoteCommitment: voteCommitment,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool("secret", false, "Cast a secret ballot vote")
	cmd.Flags().String("vote-commitment", "", "Vote commitment hash for secret ballot (hex-encoded)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdVoteWeighted casts a weighted vote on a proposal
func CmdVoteWeighted() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-weighted [proposal-id] [options]",
		Short: "Cast a weighted vote on a proposal",
		Long: `Cast a weighted vote with split voting power across multiple options.

Examples:
  aurad tx governance vote-weighted 1 "yes=0.6,no=0.4" --from alice
  aurad tx governance vote-weighted 2 "yes=0.5,abstain=0.5" --from bob
  aurad tx governance vote-weighted 3 "yes=0.3,no=0.3,abstain=0.4" --from charlie

Options format: "option1=weight1,option2=weight2,..."
Weights must sum to 1.0 and be between 0 and 1.

Vote options: yes, no, abstain, no-with-veto

This allows voters to express nuanced positions on proposals.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			// Parse weighted options
			weightedOptions, err := parseWeightedVoteOptions(args[1])
			if err != nil {
				return err
			}

			msg := &governancev1beta1.MsgVoteWeighted{
				ProposalId: proposalID,
				Voter:      clientCtx.GetFromAddress().String(),
				Options:    weightedOptions,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRevealSecretVote reveals a secret ballot vote
func CmdRevealSecretVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reveal-secret-vote [proposal-id] [option] [reveal-key]",
		Short: "Reveal a secret ballot vote",
		Long: `Reveal a previously cast secret ballot vote after the reveal period begins.

Examples:
  aurad tx governance reveal-secret-vote 1 yes mysecretkey123 --from alice
  aurad tx governance reveal-secret-vote 2 no anothersecret456 --from bob

Arguments:
  proposal-id: The proposal ID
  option: The actual vote option (yes, no, abstain, no-with-veto)
  reveal-key: The secret used to create the vote commitment

The reveal-key must match the commitment hash from the original vote.
Votes must be revealed during the reveal period, or they won't be counted.

Secret ballot process:
  1. Vote with commitment: hash(option + reveal-key)
  2. Wait for voting period to end
  3. Reveal during reveal period
  4. Vote is counted if reveal matches commitment
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			// Parse vote option
			option, err := parseVoteOption(args[1])
			if err != nil {
				return err
			}

			revealKey := args[2]

			msg := &governancev1beta1.MsgRevealSecretVote{
				ProposalId: proposalID,
				Voter:      clientCtx.GetFromAddress().String(),
				Option:     option,
				RevealKey:  revealKey,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Vote Delegation Commands
// ============================================================================

// CmdDelegateVote delegates voting power to another address
func CmdDelegateVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate-vote [delegate-address]",
		Short: "Delegate voting power to another address",
		Long: `Delegate your voting power to a trusted address for governance votes.

Examples:
  aurad tx governance delegate-vote aura1delegate... --from alice
  aurad tx governance delegate-vote aura1delegate... --categories text,parameter-change --from alice

Options:
  --categories: Comma-separated list of proposal categories to delegate
                If not specified, delegates for all categories

Categories: text, parameter-change, software-upgrade, spending, emergency, constitution

The delegate can vote on your behalf for the specified categories.
You can override their vote by voting directly on a proposal.
You can undelegate at any time.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delegateAddress := args[0]

			// Parse categories if provided
			categoriesStr, _ := cmd.Flags().GetString("categories")
			var categories []governancev1beta1.ProposalCategory
			if categoriesStr != "" {
				cats := strings.Split(categoriesStr, ",")
				for _, cat := range cats {
					category, err := parseProposalCategory(strings.TrimSpace(cat))
					if err != nil {
						return err
					}
					categories = append(categories, category)
				}
			}

			msg := &governancev1beta1.MsgDelegateVote{
				Delegator:  clientCtx.GetFromAddress().String(),
				Delegate:   delegateAddress,
				Categories: categories,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("categories", "", "Comma-separated list of proposal categories to delegate")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUndelegateVote removes vote delegation
func CmdUndelegateVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "undelegate-vote [delegate-address]",
		Short: "Remove vote delegation from an address",
		Long: `Remove your voting power delegation from a delegate.

Examples:
  aurad tx governance undelegate-vote aura1delegate... --from alice
  aurad tx governance undelegate-vote aura1delegate... --categories text,parameter-change --from alice

Options:
  --categories: Comma-separated list of proposal categories to undelegate
                If not specified, undelegates all categories

After undelegation, the delegate can no longer vote on your behalf.
You regain full control of your voting power.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delegateAddress := args[0]

			// Parse categories if provided
			categoriesStr, _ := cmd.Flags().GetString("categories")
			var categories []governancev1beta1.ProposalCategory
			if categoriesStr != "" {
				cats := strings.Split(categoriesStr, ",")
				for _, cat := range cats {
					category, err := parseProposalCategory(strings.TrimSpace(cat))
					if err != nil {
						return err
					}
					categories = append(categories, category)
				}
			}

			msg := &governancev1beta1.MsgUndelegateVote{
				Delegator:  clientCtx.GetFromAddress().String(),
				Delegate:   delegateAddress,
				Categories: categories,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("categories", "", "Comma-separated list of proposal categories to undelegate")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Veto Commands
// ============================================================================

// CmdSubmitVeto submits a veto request
func CmdSubmitVeto() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-veto [proposal-id] [reason]",
		Short: "Submit a veto request for a proposal",
		Long: `Submit a veto request to block a malicious or dangerous proposal.

Examples:
  aurad tx governance submit-veto 1 "This proposal contains malicious code" --from validator1
  aurad tx governance submit-veto 2 "Security vulnerability detected" --from validator2

Veto mechanism:
  - Only authorized addresses can submit veto requests
  - Multiple cosigners required to execute veto
  - Veto immediately halts proposal execution
  - Used for emergency situations

If enough cosigners approve, the proposal is immediately vetoed.
Deposits from the proposal are burned as penalty.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			reason := args[1]

			msg := &governancev1beta1.MsgSubmitVeto{
				ProposalId: proposalID,
				Vetoer:     clientCtx.GetFromAddress().String(),
				Reason:     reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCosignVeto cosigns an existing veto request
func CmdCosignVeto() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cosign-veto [proposal-id]",
		Short: "Cosign an existing veto request",
		Long: `Add your signature to an existing veto request.

Examples:
  aurad tx governance cosign-veto 1 --from validator2
  aurad tx governance cosign-veto 2 --from validator3

When enough authorized cosigners approve the veto:
  - The proposal is immediately vetoed
  - All deposits are burned
  - Proposal status changes to VETOED
  - Execution is permanently blocked

Only authorized veto addresses can cosign.
Check governance params for required number of cosigners.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			msg := &governancev1beta1.MsgCosignVeto{
				ProposalId: proposalID,
				Cosigner:   clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Execution Commands
// ============================================================================

// CmdExecuteProposal executes a passed proposal after time-lock
func CmdExecuteProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-proposal [proposal-id]",
		Short: "Execute a passed proposal after time-lock delay",
		Long: `Execute a proposal that has passed voting and completed its time-lock delay.

Examples:
  aurad tx governance execute-proposal 1 --from alice
  aurad tx governance execute-proposal 5 --from bob

Requirements:
  - Proposal must have passed voting
  - Time-lock delay must have elapsed
  - Proposal must be in READY_FOR_EXECUTION status

Time-lock provides a safety period for:
  - Security review of parameter changes
  - Community to prepare for upgrades
  - Detection of malicious proposals
  - Emergency veto if needed

Anyone can execute a ready proposal (permissionless).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			msg := &governancev1beta1.MsgExecuteProposal{
				ProposalId: proposalID,
				Executor:   clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Snapshot Voting Commands
// ============================================================================

// CmdSubmitSnapshotVote submits an off-chain snapshot vote
func CmdSubmitSnapshotVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-snapshot-vote [proposal-id] [option] [signature]",
		Short: "Submit an off-chain snapshot vote",
		Long: `Submit a vote using voting power from a historical snapshot block.

Examples:
  aurad tx governance submit-snapshot-vote 1 yes abc123signature... --from alice
  aurad tx governance submit-snapshot-vote 2 no def456signature... --from bob

Snapshot voting features:
  - Voting power calculated at proposal creation height
  - Prevents vote manipulation via token transfers
  - Requires signature over (proposal_id, option, snapshot_height)
  - Can be submitted off-chain and relayed

This enables gasless voting and vote aggregation.
Signature must be from the voter's address using their private key.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			// Parse vote option
			option, err := parseVoteOption(args[1])
			if err != nil {
				return err
			}

			signature := args[2]

			msg := &governancev1beta1.MsgSubmitSnapshotVote{
				ProposalId: proposalID,
				Voter:      clientCtx.GetFromAddress().String(),
				Option:     option,
				Signature:  signature,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseProposalCategory parses a proposal category string
func parseProposalCategory(categoryStr string) (governancev1beta1.ProposalCategory, error) {
	categoryStr = strings.ToLower(strings.TrimSpace(categoryStr))
	switch categoryStr {
	case "text":
		return governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_TEXT, nil
	case "parameter-change", "param-change", "parameter_change":
		return governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE, nil
	case "software-upgrade", "upgrade", "software_upgrade":
		return governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE, nil
	case "spending", "spend":
		return governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_SPENDING, nil
	case "emergency":
		return governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY, nil
	case "constitution":
		return governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION, nil
	default:
		return governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_UNSPECIFIED,
			fmt.Errorf("invalid category: %s (valid: text, parameter-change, software-upgrade, spending, emergency, constitution)", categoryStr)
	}
}

// parseVoteOption parses a vote option string
func parseVoteOption(optionStr string) (governancev1beta1.VoteOption, error) {
	optionStr = strings.ToLower(strings.TrimSpace(optionStr))
	switch optionStr {
	case "yes", "y":
		return governancev1beta1.VoteOption_VOTE_OPTION_YES, nil
	case "no", "n":
		return governancev1beta1.VoteOption_VOTE_OPTION_NO, nil
	case "abstain", "a":
		return governancev1beta1.VoteOption_VOTE_OPTION_ABSTAIN, nil
	case "no-with-veto", "veto", "nwv", "no_with_veto":
		return governancev1beta1.VoteOption_VOTE_OPTION_NO_WITH_VETO, nil
	default:
		return governancev1beta1.VoteOption_VOTE_OPTION_UNSPECIFIED,
			fmt.Errorf("invalid vote option: %s (valid: yes, no, abstain, no-with-veto)", optionStr)
	}
}

// parseWeightedVoteOptions parses weighted vote options from a string
func parseWeightedVoteOptions(optionsStr string) ([]*governancev1beta1.WeightedVoteOption, error) {
	// Format: "yes=0.6,no=0.4"
	parts := strings.Split(optionsStr, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("no vote options provided")
	}

	var options []*governancev1beta1.WeightedVoteOption
	totalWeight := sdkmath.LegacyZeroDec()

	for _, part := range parts {
		kv := strings.Split(strings.TrimSpace(part), "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid option format: %s (expected option=weight)", part)
		}

		// Parse option
		option, err := parseVoteOption(kv[0])
		if err != nil {
			return nil, err
		}

		// Parse weight
		weight, err := sdkmath.LegacyNewDecFromStr(kv[1])
		if err != nil {
			return nil, fmt.Errorf("invalid weight: %w", err)
		}

		if weight.IsNegative() || weight.GT(sdkmath.LegacyOneDec()) {
			return nil, fmt.Errorf("weight must be between 0 and 1: %s", kv[1])
		}

		totalWeight = totalWeight.Add(weight)

		options = append(options, &governancev1beta1.WeightedVoteOption{
			Option: option,
			Weight: weight.String(),
		})
	}

	// Validate total weight is 1.0
	if !totalWeight.Equal(sdkmath.LegacyOneDec()) {
		return nil, fmt.Errorf("total weight must equal 1.0, got %s", totalWeight.String())
	}

	return options, nil
}
