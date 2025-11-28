package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "economicsecurity",
		Aliases:                    []string{"econ", "es"},
		Short:                      "Economic security transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdCreateVestingSchedule(),
		CmdReleaseVestedTokens(),
		CmdRevokeVestingSchedule(),
		CmdLockVotingTokens(),
		CmdUnlockVotingTokens(),
		CmdProposeTreasurySpend(),
		CmdSignTreasurySpend(),
		CmdExecuteTreasurySpend(),
	)

	return cmd
}

// CmdCreateVestingSchedule creates a vesting schedule
func CmdCreateVestingSchedule() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-vesting [beneficiary] [amount] [cliff-duration] [vesting-duration] [vesting-type] [schedule-type]",
		Short: "Create a new vesting schedule",
		Long: `Create a new token vesting schedule for a beneficiary.

Vesting Types:
  0 - LINEAR: Tokens vest linearly over time
  1 - CLIFF: All tokens vest after cliff period
  2 - EXPONENTIAL: Exponential vesting curve

Schedule Types:
  0 - TEAM: Team member vesting
  1 - INVESTOR: Investor vesting
  2 - ADVISOR: Advisor vesting
  3 - ECOSYSTEM: Ecosystem fund vesting

Example:
  aurad tx economicsecurity create-vesting aura1abc... 1000000uaura 15552000 31104000 0 0 --from alice
  (Creates linear vesting for team member: 180 days cliff, 360 days vesting)
`,
		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			beneficiary := args[0]
			amount := args[1]
			cliffDuration, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid cliff duration: %w", err)
			}
			vestingDuration, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid vesting duration: %w", err)
			}
			vestingTypeInt, err := strconv.ParseInt(args[4], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid vesting type: %w", err)
			}
			scheduleTypeInt, err := strconv.ParseInt(args[5], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid schedule type: %w", err)
			}

			msg := &v1beta1.MsgCreateVestingSchedule{
				Creator:           clientCtx.GetFromAddress().String(),
				BeneficiaryAddress: beneficiary,
				TotalAmount:       amount,
				CliffDuration:     cliffDuration,
				VestingDuration:   vestingDuration,
				VestingType:       v1beta1.VestingType(vestingTypeInt),
				ScheduleType:      v1beta1.ScheduleType(scheduleTypeInt),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdReleaseVestedTokens releases vested tokens
func CmdReleaseVestedTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release-vested [schedule-id]",
		Short: "Release vested tokens from a schedule",
		Long: `Release currently vested tokens from a vesting schedule.

Example:
  aurad tx economicsecurity release-vested "schedule-123" --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scheduleID := args[0]

			msg := &v1beta1.MsgReleaseVestedTokens{
				Beneficiary: clientCtx.GetFromAddress().String(),
				ScheduleId:  scheduleID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRevokeVestingSchedule revokes a vesting schedule
func CmdRevokeVestingSchedule() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-vesting [schedule-id] [reason]",
		Short: "Revoke a vesting schedule",
		Long: `Revoke a vesting schedule and return unvested tokens.

Example:
  aurad tx economicsecurity revoke-vesting "schedule-123" "Terminated employment" --from alice
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			scheduleID := args[0]
			reason := args[1]

			msg := &v1beta1.MsgRevokeVestingSchedule{
				Revoker:    clientCtx.GetFromAddress().String(),
				ScheduleId: scheduleID,
				Reason:     reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdLockVotingTokens locks tokens for voting power
func CmdLockVotingTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-voting [amount] [lock-duration]",
		Short: "Lock tokens to boost voting power",
		Long: `Lock tokens for a specified duration to receive boosted voting power.

Longer lock periods = higher voting power multiplier.

Example:
  aurad tx economicsecurity lock-voting 10000uaura 31536000 --from alice
  (Lock 10000 tokens for ~1 year = 365 days)
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount := args[0]
			lockDuration, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid lock duration: %w", err)
			}

			msg := &v1beta1.MsgLockVotingTokens{
				Owner:        clientCtx.GetFromAddress().String(),
				Amount:       amount,
				LockDuration: lockDuration,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUnlockVotingTokens unlocks voting tokens
func CmdUnlockVotingTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock-voting [lock-id]",
		Short: "Unlock voting tokens after lock period",
		Long: `Unlock voting tokens after the lock period has ended.

Example:
  aurad tx economicsecurity unlock-voting "lock-123" --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			lockID := args[0]

			msg := &v1beta1.MsgUnlockVotingTokens{
				Owner:  clientCtx.GetFromAddress().String(),
				LockId: lockID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdProposeTreasurySpend proposes a treasury spend
func CmdProposeTreasurySpend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-treasury-spend [recipient] [amount] [description]",
		Short: "Propose a treasury spend",
		Long: `Propose a multi-sig treasury spend that requires multiple approvals.

Example:
  aurad tx economicsecurity propose-treasury-spend aura1abc... 100000uaura "Development grant Q1 2024" --from alice
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			recipient := args[0]
			amount := args[1]
			description := args[2]

			msg := &v1beta1.MsgProposeTreasurySpend{
				Proposer:    clientCtx.GetFromAddress().String(),
				Recipient:   recipient,
				Amount:      amount,
				Description: description,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSignTreasurySpend signs a treasury spend proposal
func CmdSignTreasurySpend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-treasury-spend [tx-id]",
		Short: "Sign a pending treasury spend proposal",
		Long: `Sign a pending treasury spend proposal as an authorized signer.

Example:
  aurad tx economicsecurity sign-treasury-spend "tx-123" --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txID := args[0]

			msg := &v1beta1.MsgSignTreasurySpend{
				Signer: clientCtx.GetFromAddress().String(),
				TxId:   txID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdExecuteTreasurySpend executes an approved treasury spend
func CmdExecuteTreasurySpend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-treasury-spend [tx-id]",
		Short: "Execute an approved treasury spend",
		Long: `Execute a treasury spend that has received sufficient signatures.

Example:
  aurad tx economicsecurity execute-treasury-spend "tx-123" --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txID := args[0]

			msg := &v1beta1.MsgExecuteTreasurySpend{
				Executor: clientCtx.GetFromAddress().String(),
				TxId:     txID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
