// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	governancev1beta1 "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

// GetQueryCmd returns the cli query commands for the governance module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "governance",
		Short:                      "Querying commands for the governance module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// Proposal queries
		CmdQueryProposal(),
		CmdQueryProposals(),

		// Vote queries
		CmdQueryVote(),
		CmdQueryVotes(),

		// Deposit queries
		CmdQueryDeposit(),
		CmdQueryDeposits(),

		// Tally queries
		CmdQueryTallyResult(),

		// Params query
		CmdQueryParams(),

		// Vote delegation queries
		CmdQueryVoteDelegations(),
		CmdQueryVotingPower(),

		// Token lock queries
		CmdQueryTokenLocks(),

		// Veto queries
		CmdQueryVetoRequests(),

		// Snapshot voting queries
		CmdQuerySnapshotVotes(),
	)

	return cmd
}

// ============================================================================
// Proposal Queries
// ============================================================================

// CmdQueryProposal queries a proposal by ID
func CmdQueryProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [proposal-id]",
		Short: "Query a proposal by ID",
		Long: `Query details of a specific governance proposal.

Examples:
  aurad query governance proposal 1
  aurad query governance proposal 5

Returns:
  - Proposal ID and title
  - Description and category
  - Proposer address
  - Current status (deposit period, voting, passed, etc.)
  - Voting statistics and tally
  - Deposit information
  - Submission and voting period timestamps
  - Execution delay and time-lock
  - Emergency status
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryProposalRequest{
				ProposalId: proposalID,
			}

			res, err := queryClient.Proposal(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryProposals queries all proposals with optional filters
func CmdQueryProposals() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals",
		Short: "Query all governance proposals",
		Long: `Query all governance proposals with optional status filtering.

Examples:
  aurad query governance proposals
  aurad query governance proposals --status deposit-period
  aurad query governance proposals --status voting-period
  aurad query governance proposals --status passed
  aurad query governance proposals --voter aura1abc...
  aurad query governance proposals --depositor aura1def...

Status options:
  deposit-period       - Proposals in deposit period
  voting-period        - Proposals in voting period
  passed               - Passed proposals
  rejected             - Rejected proposals
  failed               - Failed proposals
  vetoed               - Vetoed proposals
  execution-delay      - Proposals in time-lock delay
  ready-for-execution  - Proposals ready to execute
  executed             - Executed proposals

Filters:
  --status: Filter by proposal status
  --voter: Filter proposals where address has voted
  --depositor: Filter proposals where address has deposited
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			// Parse filters
			statusStr, _ := cmd.Flags().GetString("status")
			voter, _ := cmd.Flags().GetString("voter")
			depositor, _ := cmd.Flags().GetString("depositor")

			var status governancev1beta1.ProposalStatus
			if statusStr != "" {
				status, err = parseProposalStatus(statusStr)
				if err != nil {
					return err
				}
			}

			req := &governancev1beta1.QueryProposalsRequest{
				Status:    status,
				Voter:     voter,
				Depositor: depositor,
			}

			res, err := queryClient.Proposals(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by proposal status")
	cmd.Flags().String("voter", "", "Filter proposals where address has voted")
	cmd.Flags().String("depositor", "", "Filter proposals where address has deposited")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Vote Queries
// ============================================================================

// CmdQueryVote queries a vote by proposal ID and voter
func CmdQueryVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter]",
		Short: "Query a vote by proposal ID and voter address",
		Long: `Query a specific vote on a proposal.

Examples:
  aurad query governance vote 1 aura1abc...
  aurad query governance vote 5 aura1def...

Returns:
  - Proposal ID
  - Voter address
  - Vote option (yes, no, abstain, no-with-veto)
  - Vote timestamp
  - Secret ballot status
  - Voting power at snapshot height
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			voterAddr := args[1]

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryVoteRequest{
				ProposalId: proposalID,
				Voter:      voterAddr,
			}

			res, err := queryClient.Vote(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVotes queries all votes for a proposal
func CmdQueryVotes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes [proposal-id]",
		Short: "Query all votes for a proposal",
		Long: `Query all votes cast on a specific proposal.

Examples:
  aurad query governance votes 1
  aurad query governance votes 5

Returns a list of all votes with:
  - Voter addresses
  - Vote options
  - Voting power
  - Vote timestamps
  - Secret ballot status

Useful for analyzing voting patterns and participation.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryVotesRequest{
				ProposalId: proposalID,
			}

			res, err := queryClient.Votes(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Deposit Queries
// ============================================================================

// CmdQueryDeposit queries a deposit by proposal ID and depositor
func CmdQueryDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [proposal-id] [depositor]",
		Short: "Query a deposit by proposal ID and depositor address",
		Long: `Query a specific deposit on a proposal.

Examples:
  aurad query governance deposit 1 aura1abc...
  aurad query governance deposit 5 aura1def...

Returns:
  - Proposal ID
  - Depositor address
  - Deposit amount
  - Deposit timestamp
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			depositorAddr := args[1]

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryDepositRequest{
				ProposalId: proposalID,
				Depositor:  depositorAddr,
			}

			res, err := queryClient.Deposit(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryDeposits queries all deposits for a proposal
func CmdQueryDeposits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits [proposal-id]",
		Short: "Query all deposits for a proposal",
		Long: `Query all deposits made on a specific proposal.

Examples:
  aurad query governance deposits 1
  aurad query governance deposits 5

Returns a list of all deposits with:
  - Depositor addresses
  - Deposit amounts
  - Deposit timestamps

Shows progress toward minimum deposit requirement.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryDepositsRequest{
				ProposalId: proposalID,
			}

			res, err := queryClient.Deposits(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Tally Queries
// ============================================================================

// CmdQueryTallyResult queries the tally of a proposal vote
func CmdQueryTallyResult() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tally [proposal-id]",
		Short: "Query the tally of a proposal vote",
		Long: `Query the current vote tally for a proposal.

Examples:
  aurad query governance tally 1
  aurad query governance tally 5

Returns:
  - Yes votes (count and percentage)
  - No votes (count and percentage)
  - Abstain votes (count and percentage)
  - No with veto votes (count and percentage)
  - Total voting power
  - Turnout percentage
  - Current outcome (passing/failing)

Tally is updated in real-time as votes are cast.
For active proposals, shows current status.
For completed proposals, shows final result.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryTallyResultRequest{
				ProposalId: proposalID,
			}

			res, err := queryClient.TallyResult(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Params Query
// ============================================================================

// CmdQueryParams queries the governance parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the governance parameters",
		Long: `Query the current governance module parameters.

Example:
  aurad query governance params

Returns:
  Deposit parameters:
    - Minimum deposit amount
    - Maximum deposit period

  Voting parameters:
    - Voting period duration
    - Quorum threshold
    - Pass threshold
    - Veto threshold

  Execution parameters:
    - Execution delay (time-lock)

  Emergency parameters:
    - Emergency voting period
    - Emergency quorum
    - Emergency threshold

  Category-specific parameters:
    - Per-category thresholds and durations

  Veto parameters:
    - Required cosigners
    - Authorized veto addresses

  Token lock parameters:
    - Lock requirement
    - Lock duration

  Snapshot voting parameters:
    - Enabled status
    - Lookback blocks

  Secret ballot parameters:
    - Enabled status
    - Reveal period
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryParamsRequest{}

			res, err := queryClient.Params(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Vote Delegation Queries
// ============================================================================

// CmdQueryVoteDelegations queries vote delegations for a delegator
func CmdQueryVoteDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-delegations [delegator]",
		Short: "Query all vote delegations for a delegator",
		Long: `Query all active vote delegations from a specific delegator.

Examples:
  aurad query governance vote-delegations aura1abc...

Returns a list of delegations with:
  - Delegator address
  - Delegate address
  - Delegated voting power
  - Delegation timestamp
  - Categories delegated (if category-specific)

Shows who can vote on behalf of the delegator.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			delegatorAddr := args[0]

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryVoteDelegationsRequest{
				Delegator: delegatorAddr,
			}

			res, err := queryClient.VoteDelegations(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVotingPower queries the voting power of an address
func CmdQueryVotingPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voting-power [address]",
		Short: "Query the voting power of an address",
		Long: `Query the current voting power of an address for governance.

Examples:
  aurad query governance voting-power aura1abc...
  aurad query governance voting-power aura1abc... --proposal-id 5

Options:
  --proposal-id: Query voting power at specific proposal's snapshot height

Returns:
  - Own voting power (from token holdings)
  - Delegated voting power (from others)
  - Total voting power
  - Snapshot height (if proposal-id specified)

Voting power is based on:
  - Token holdings at snapshot height
  - Delegated voting power from other addresses
  - Staking participation
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			proposalID, _ := cmd.Flags().GetUint64("proposal-id")

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryVotingPowerRequest{
				Address:    address,
				ProposalId: proposalID,
			}

			res, err := queryClient.VotingPower(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint64("proposal-id", 0, "Query voting power at specific proposal's snapshot height")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Token Lock Queries
// ============================================================================

// CmdQueryTokenLocks queries token locks for an address
func CmdQueryTokenLocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-locks [address]",
		Short: "Query all token locks for an address",
		Long: `Query all active token locks from governance voting.

Examples:
  aurad query governance token-locks aura1abc...

Returns a list of token locks with:
  - Owner address
  - Proposal ID
  - Locked amount
  - Lock timestamp
  - Unlock timestamp

Token locks are created when voting if require_token_lock is enabled.
Prevents vote manipulation via token transfers during voting.
Tokens are automatically unlocked after the lock period expires.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryTokenLocksRequest{
				Address: address,
			}

			res, err := queryClient.TokenLocks(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Veto Queries
// ============================================================================

// CmdQueryVetoRequests queries veto requests for a proposal
func CmdQueryVetoRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "veto-requests [proposal-id]",
		Short: "Query all veto requests for a proposal",
		Long: `Query all veto requests submitted for a specific proposal.

Examples:
  aurad query governance veto-requests 1
  aurad query governance veto-requests 5

Returns a list of veto requests with:
  - Proposal ID
  - Vetoer address (initial submitter)
  - Reason for veto
  - Submission timestamp
  - Cosigners (addresses that have approved)
  - Execution status

Shows progress toward veto execution threshold.
Multiple cosigners needed to execute veto.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QueryVetoRequestsRequest{
				ProposalId: proposalID,
			}

			res, err := queryClient.VetoRequests(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Snapshot Voting Queries
// ============================================================================

// CmdQuerySnapshotVotes queries snapshot votes for a proposal
func CmdQuerySnapshotVotes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot-votes [proposal-id]",
		Short: "Query all snapshot votes for a proposal",
		Long: `Query all off-chain snapshot votes submitted for a proposal.

Examples:
  aurad query governance snapshot-votes 1
  aurad query governance snapshot-votes 5

Returns a list of snapshot votes with:
  - Proposal ID
  - Snapshot height
  - Voter address
  - Vote option
  - Voting power at snapshot
  - Signature
  - Submission timestamp

Snapshot votes use historical voting power from proposal creation.
Enables gasless voting via signature verification.
Can be aggregated and submitted in batches.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid proposal-id: %w", err)
			}

			queryClient := governancev1beta1.NewQueryClient(clientCtx)

			req := &governancev1beta1.QuerySnapshotVotesRequest{
				ProposalId: proposalID,
			}

			res, err := queryClient.SnapshotVotes(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseProposalStatus parses a proposal status string
func parseProposalStatus(statusStr string) (governancev1beta1.ProposalStatus, error) {
	statusStr = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(statusStr), "_", "-"))
	switch statusStr {
	case "deposit-period", "deposit":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD, nil
	case "voting-period", "voting":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, nil
	case "passed", "pass":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_PASSED, nil
	case "rejected", "reject":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_REJECTED, nil
	case "failed", "fail":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_FAILED, nil
	case "vetoed", "veto":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VETOED, nil
	case "execution-delay", "timelock":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY, nil
	case "ready-for-execution", "ready":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION, nil
	case "executed", "exec":
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTED, nil
	default:
		return governancev1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED,
			fmt.Errorf("invalid status: %s", statusStr)
	}
}

// FormatProposalStatus formats proposal status for display
func FormatProposalStatus(status governancev1beta1.ProposalStatus) string {
	switch status {
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD:
		return "DEPOSIT_PERIOD"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD:
		return "VOTING_PERIOD"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_PASSED:
		return "PASSED"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_REJECTED:
		return "REJECTED"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_FAILED:
		return "FAILED"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_VETOED:
		return "VETOED"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY:
		return "EXECUTION_DELAY"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION:
		return "READY_FOR_EXECUTION"
	case governancev1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTED:
		return "EXECUTED"
	default:
		return "UNSPECIFIED"
	}
}

// FormatVoteOption formats vote option for display
func FormatVoteOption(option governancev1beta1.VoteOption) string {
	switch option {
	case governancev1beta1.VoteOption_VOTE_OPTION_YES:
		return "YES"
	case governancev1beta1.VoteOption_VOTE_OPTION_NO:
		return "NO"
	case governancev1beta1.VoteOption_VOTE_OPTION_ABSTAIN:
		return "ABSTAIN"
	case governancev1beta1.VoteOption_VOTE_OPTION_NO_WITH_VETO:
		return "NO_WITH_VETO"
	default:
		return "UNSPECIFIED"
	}
}

// FormatProposalCategory formats proposal category for display
func FormatProposalCategory(category governancev1beta1.ProposalCategory) string {
	switch category {
	case governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_TEXT:
		return "TEXT"
	case governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE:
		return "PARAMETER_CHANGE"
	case governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE:
		return "SOFTWARE_UPGRADE"
	case governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_SPENDING:
		return "SPENDING"
	case governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY:
		return "EMERGENCY"
	case governancev1beta1.ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION:
		return "CONSTITUTION"
	default:
		return "UNSPECIFIED"
	}
}
