// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "auth",
		Short:                      "Querying commands for the auth module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryRole(),
		GetCmdQueryListRoles(),
		GetCmdQueryRoleAssignments(),
		GetCmdQueryHasPermission(),
		GetCmdQueryMultisigWallet(),
		GetCmdQueryListMultisigWallets(),
		GetCmdQueryMultisigProposal(),
		GetCmdQueryListMultisigProposals(),
		GetCmdQueryTimeLockedAction(),
		GetCmdQueryListTimeLockedActions(),
		GetCmdQueryEmergencyAdmin(),
		GetCmdQueryListEmergencyAdmins(),
		GetCmdQueryValidatorKeyRotation(),
		GetCmdQuerySession(),
		GetCmdQueryListSessions(),
		GetCmdQueryRateLimitStatus(),
		GetCmdQueryAuditLogs(),
		GetCmdQueryParams(),
	)

	return cmd
}

// GetCmdQueryRole queries a specific role by name
func GetCmdQueryRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role [name]",
		Short: "Query a role by name",
		Long: `Query details of a specific role including its permissions and metadata.

Example:
  aurad query auth role admin
  aurad query auth role operator`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetRoleRequest{
				Name: args[0],
			}

			res, err := queryClient.GetRole(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryListRoles queries all roles
func GetCmdQueryListRoles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roles",
		Short: "List all roles",
		Long: `List all roles defined in the system.

Example:
  aurad query auth roles`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryListRolesRequest{}

			res, err := queryClient.ListRoles(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryRoleAssignments queries role assignments for an address
func GetCmdQueryRoleAssignments() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "role-assignments [address]",
		Short: "Query role assignments for an address",
		Long: `Query all role assignments for a specific address.

Example:
  aurad query auth role-assignments aura1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetRoleAssignmentsRequest{
				Address: args[0],
			}

			res, err := queryClient.GetRoleAssignments(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryHasPermission checks if an address has a specific permission
func GetCmdQueryHasPermission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "has-permission [address] [permission]",
		Short: "Check if an address has a specific permission",
		Long: `Check if an address has a specific permission through any of its assigned roles.

Example:
  aurad query auth has-permission aura1abc... CREATE_ROLE`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryHasPermissionRequest{
				Address:    args[0],
				Permission: args[1],
			}

			res, err := queryClient.HasPermission(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryMultisigWallet queries a multisig wallet
func GetCmdQueryMultisigWallet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisig-wallet [id]",
		Short: "Query a multisig wallet by ID",
		Long: `Query details of a specific multisig wallet.

Example:
  aurad query auth multisig-wallet wallet123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetMultisigWalletRequest{
				Id: args[0],
			}

			res, err := queryClient.GetMultisigWallet(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryListMultisigWallets lists all multisig wallets
func GetCmdQueryListMultisigWallets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisig-wallets",
		Short: "List all multisig wallets",
		Long: `List all multisig wallets in the system.

Example:
  aurad query auth multisig-wallets`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryListMultisigWalletsRequest{}

			res, err := queryClient.ListMultisigWallets(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryMultisigProposal queries a multisig proposal
func GetCmdQueryMultisigProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisig-proposal [id]",
		Short: "Query a multisig proposal by ID",
		Long: `Query details of a specific multisig proposal.

Example:
  aurad query auth multisig-proposal proposal123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetMultisigProposalRequest{
				Id: args[0],
			}

			res, err := queryClient.GetMultisigProposal(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryListMultisigProposals lists multisig proposals
func GetCmdQueryListMultisigProposals() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisig-proposals [wallet-id]",
		Short: "List multisig proposals for a wallet",
		Long: `List all multisig proposals for a specific wallet, optionally filtered by status.

Example:
  aurad query auth multisig-proposals wallet123
  aurad query auth multisig-proposals wallet123 --status pending`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			statusStr, err := cmd.Flags().GetString("status")
			if err != nil {
				return err
			}

			status := parseProposalStatus(statusStr)

			req := &v1beta1.QueryListMultisigProposalsRequest{
				WalletId: args[0],
				Status:   status,
			}

			res, err := queryClient.ListMultisigProposals(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by status (pending, approved, executed, rejected, expired)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryTimeLockedAction queries a time-locked action
func GetCmdQueryTimeLockedAction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timelocked-action [id]",
		Short: "Query a time-locked action by ID",
		Long: `Query details of a specific time-locked action.

Example:
  aurad query auth timelocked-action action123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetTimeLockedActionRequest{
				Id: args[0],
			}

			res, err := queryClient.GetTimeLockedAction(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryListTimeLockedActions lists time-locked actions
func GetCmdQueryListTimeLockedActions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timelocked-actions",
		Short: "List all time-locked actions",
		Long: `List all time-locked actions, optionally filtered by status.

Example:
  aurad query auth timelocked-actions
  aurad query auth timelocked-actions --status pending`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			statusStr, err := cmd.Flags().GetString("status")
			if err != nil {
				return err
			}

			status := parseActionStatus(statusStr)

			req := &v1beta1.QueryListTimeLockedActionsRequest{
				Status: status,
			}

			res, err := queryClient.ListTimeLockedActions(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by status (pending, ready, executed, cancelled)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryEmergencyAdmin queries emergency admin status
func GetCmdQueryEmergencyAdmin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "emergency-admin [address]",
		Short: "Query emergency admin status",
		Long: `Query the emergency admin status for a specific address.

Example:
  aurad query auth emergency-admin aura1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetEmergencyAdminRequest{
				Address: args[0],
			}

			res, err := queryClient.GetEmergencyAdmin(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryListEmergencyAdmins lists all emergency admins
func GetCmdQueryListEmergencyAdmins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "emergency-admins",
		Short: "List all emergency admins",
		Long: `List all emergency admins in the system.

Example:
  aurad query auth emergency-admins`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryListEmergencyAdminsRequest{}

			res, err := queryClient.ListEmergencyAdmins(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryValidatorKeyRotation queries validator key rotation status
func GetCmdQueryValidatorKeyRotation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-key-rotation [validator-address]",
		Short: "Query validator key rotation status",
		Long: `Query the key rotation status for a specific validator.

Example:
  aurad query auth validator-key-rotation auravaloper1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetValidatorKeyRotationRequest{
				ValidatorAddress: args[0],
			}

			res, err := queryClient.GetValidatorKeyRotation(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQuerySession queries a session
func GetCmdQuerySession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session [session-id]",
		Short: "Query a session by ID",
		Long: `Query details of a specific session.

Example:
  aurad query auth session session123`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetSessionRequest{
				SessionId: args[0],
			}

			res, err := queryClient.GetSession(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryListSessions lists sessions for a user
func GetCmdQueryListSessions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions [user-address]",
		Short: "List sessions for a user",
		Long: `List all sessions for a specific user address.

Example:
  aurad query auth sessions aura1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryListSessionsRequest{
				UserAddress: args[0],
			}

			res, err := queryClient.ListSessions(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryRateLimitStatus queries rate limit status
func GetCmdQueryRateLimitStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit-status [user-address]",
		Short: "Query rate limit status for a user",
		Long: `Query the current rate limit status for a specific user address.

Example:
  aurad query auth rate-limit-status aura1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetRateLimitStatusRequest{
				UserAddress: args[0],
			}

			res, err := queryClient.GetRateLimitStatus(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryAuditLogs queries audit logs
func GetCmdQueryAuditLogs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit-logs",
		Short: "Query audit logs",
		Long: `Query audit logs with optional filters for actor, action, and time range.

Example:
  aurad query auth audit-logs
  aurad query auth audit-logs --actor aura1abc...
  aurad query auth audit-logs --action CREATE_ROLE --start-time 1609459200 --end-time 1609545600 --limit 100`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			actor, err := cmd.Flags().GetString("actor")
			if err != nil {
				return err
			}

			action, err := cmd.Flags().GetString("action")
			if err != nil {
				return err
			}

			startTime, err := cmd.Flags().GetInt64("start-time")
			if err != nil {
				return err
			}

			endTime, err := cmd.Flags().GetInt64("end-time")
			if err != nil {
				return err
			}

			limit, err := cmd.Flags().GetUint64("limit")
			if err != nil {
				return err
			}

			req := &v1beta1.QueryGetAuditLogsRequest{
				Actor:     actor,
				Action:    action,
				StartTime: startTime,
				EndTime:   endTime,
				Limit:     limit,
			}

			res, err := queryClient.GetAuditLogs(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("actor", "", "Filter by actor address")
	cmd.Flags().String("action", "", "Filter by action type")
	cmd.Flags().Int64("start-time", 0, "Start time (Unix timestamp)")
	cmd.Flags().Int64("end-time", 0, "End time (Unix timestamp)")
	cmd.Flags().Uint64("limit", 100, "Maximum number of logs to return")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryParams queries module parameters
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query auth module parameters",
		Long: `Query the current parameters for the auth module.

Example:
  aurad query auth params`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryGetParamsRequest{}

			res, err := queryClient.GetParams(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// parseProposalStatus converts a string to ProposalStatus enum
func parseProposalStatus(s string) v1beta1.ProposalStatus {
	switch s {
	case "pending":
		return v1beta1.ProposalStatus_PROPOSAL_STATUS_PENDING
	case "approved":
		return v1beta1.ProposalStatus_PROPOSAL_STATUS_APPROVED
	case "executed":
		return v1beta1.ProposalStatus_PROPOSAL_STATUS_EXECUTED
	case "rejected":
		return v1beta1.ProposalStatus_PROPOSAL_STATUS_REJECTED
	case "expired":
		return v1beta1.ProposalStatus_PROPOSAL_STATUS_EXPIRED
	default:
		return v1beta1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	}
}

// parseActionStatus converts a string to ActionStatus enum
func parseActionStatus(s string) v1beta1.ActionStatus {
	switch s {
	case "pending":
		return v1beta1.ActionStatus_ACTION_STATUS_PENDING
	case "ready":
		return v1beta1.ActionStatus_ACTION_STATUS_READY
	case "executed":
		return v1beta1.ActionStatus_ACTION_STATUS_EXECUTED
	case "cancelled":
		return v1beta1.ActionStatus_ACTION_STATUS_CANCELLED
	default:
		return v1beta1.ActionStatus_ACTION_STATUS_UNSPECIFIED
	}
}
