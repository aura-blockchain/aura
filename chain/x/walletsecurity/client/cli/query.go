// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// GetQueryCmd returns the query commands for the walletsecurity module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "walletsecurity",
		Short:                      "Querying commands for the wallet security module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryHardwareWallet(),
		CmdQueryMultiSigWallet(),
		CmdQueryPendingMultiSigTx(),
		CmdQuerySocialRecoveryConfig(),
		CmdQueryRecoveryRequest(),
		CmdQuerySpendingLimit(),
		CmdQuerySessionConfig(),
		CmdQuerySecurityMetrics(),
		CmdQueryDomainVerification(),
		CmdQueryDustFilter(),
		CmdQueryParams(),
	)

	return cmd
}

// CmdQueryHardwareWallet queries hardware wallet configuration
func CmdQueryHardwareWallet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hw-wallet [wallet-id]",
		Short: "Query hardware wallet configuration",
		Long: `Query hardware wallet registration details including:
  - Wallet type (Ledger, Trezor, etc.)
  - Device ID
  - Firmware version
  - Derivation path
  - Registration time

Examples:
  aurad query walletsecurity hw-wallet wallet123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetHardwareWallet(context.Background(), &v1beta1.QueryGetHardwareWalletRequest{
				WalletId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryMultiSigWallet queries multi-sig wallet configuration
func CmdQueryMultiSigWallet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisig [wallet-id]",
		Short: "Query multi-signature wallet configuration",
		Long: `Query multi-sig wallet details including:
  - Signers and their addresses
  - Signature threshold
  - Optional signer weights
  - Time lock configuration
  - Current signatures on pending transactions

Examples:
  aurad query walletsecurity multisig wallet456
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetMultiSigWallet(context.Background(), &v1beta1.QueryGetMultiSigWalletRequest{
				WalletId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPendingMultiSigTx queries a pending multi-sig transaction
func CmdQueryPendingMultiSigTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-multisig-tx [tx-id]",
		Short: "Query a pending multi-signature transaction",
		Long: `Query details of a pending multi-sig transaction including:
  - Transaction data
  - Required signatures
  - Current signatures collected
  - Signers who have signed
  - Ready to execute status

Examples:
  aurad query walletsecurity pending-multisig-tx tx789
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetPendingMultiSigTx(context.Background(), &v1beta1.QueryGetPendingMultiSigTxRequest{
				TxId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySocialRecoveryConfig queries social recovery configuration
func CmdQuerySocialRecoveryConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "social-recovery [wallet-id]",
		Short: "Query social recovery configuration",
		Long: `Query social recovery setup including:
  - Guardian addresses
  - Recovery threshold
  - Recovery delay period
  - Active recovery requests

Examples:
  aurad query walletsecurity social-recovery wallet123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetSocialRecoveryConfig(context.Background(), &v1beta1.QueryGetSocialRecoveryConfigRequest{
				WalletId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryRecoveryRequest queries a recovery request
func CmdQueryRecoveryRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recovery-request [request-id]",
		Short: "Query a social recovery request",
		Long: `Query details of a recovery request including:
  - Wallet being recovered
  - New address
  - Guardian approvals
  - Threshold status
  - Executable time

Examples:
  aurad query walletsecurity recovery-request req123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetRecoveryRequest(context.Background(), &v1beta1.QueryGetRecoveryRequestRequest{
				RequestId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySpendingLimit queries spending limit configuration
func CmdQuerySpendingLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spending-limit [wallet-id] [denom]",
		Short: "Query spending limits for a wallet",
		Long: `Query spending limits including:
  - Daily limit
  - Weekly limit
  - Monthly limit
  - Current spending amounts
  - Reset times

Examples:
  aurad query walletsecurity spending-limit wallet123 uaura
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetSpendingLimit(context.Background(), &v1beta1.QueryGetSpendingLimitRequest{
				WalletId: args[0],
				Denom:    args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySessionConfig queries session configuration
func CmdQuerySessionConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session [session-id]",
		Short: "Query wallet session configuration",
		Long: `Query session settings including:
  - Timeout duration
  - Auto-lock status
  - Inactivity threshold
  - Current lock status
  - Last activity time

Examples:
  aurad query walletsecurity session session456
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetSessionConfig(context.Background(), &v1beta1.QueryGetSessionConfigRequest{
				SessionId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySecurityMetrics queries wallet security metrics
func CmdQuerySecurityMetrics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security-metrics [wallet-id]",
		Short: "Query comprehensive security metrics for a wallet",
		Long: `Query security metrics including:
  - Security score
  - Enabled security features
  - Recent security events
  - Risk assessment
  - Recommendations

Examples:
  aurad query walletsecurity security-metrics wallet123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetSecurityMetrics(context.Background(), &v1beta1.QueryGetSecurityMetricsRequest{
				WalletId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryDomainVerification queries domain verification status
func CmdQueryDomainVerification() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain-verification [domain]",
		Short: "Query domain verification status",
		Long: `Query domain verification for phishing protection including:
  - Verification status
  - Certificate hash
  - Verifier
  - Verification time

Examples:
  aurad query walletsecurity domain-verification app.aura.network
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetDomainVerification(context.Background(), &v1beta1.QueryGetDomainVerificationRequest{
				Domain: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryDustFilter queries dust filter configuration
func CmdQueryDustFilter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dust-filter [wallet-id]",
		Short: "Query dust attack filter configuration",
		Long: `Query dust filter settings including:
  - Filter enabled status
  - Minimum amount threshold
  - Maximum dust transactions per block
  - Suspicious pattern threshold
  - Recent dust attack detections

Examples:
  aurad query walletsecurity dust-filter wallet123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.GetDustFilter(context.Background(), &v1beta1.QueryGetDustFilterRequest{
				WalletId: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryParams queries the walletsecurity module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the wallet security module parameters",
		Long: `Query all parameters for the wallet security module.

Examples:
  aurad query walletsecurity params

Returns:
  - Hardware wallet requirements
  - Multi-signature configuration settings
  - Social recovery parameters
  - Spending limit defaults
  - Session timeout settings
  - Domain verification requirements
  - Dust filter thresholds
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &v1beta1.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
