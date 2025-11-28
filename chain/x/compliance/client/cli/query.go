package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "compliance",
		Short:                      "Querying commands for the compliance module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryKYCRecord(),
		CmdQueryAMLProfile(),
		CmdQuerySanctionsScreening(),
		CmdQueryTransactionAlerts(),
		CmdQueryTaxReport(),
	)

	return cmd
}

// CmdQueryKYCRecord queries KYC record
func CmdQueryKYCRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyc-record [address]",
		Short: "Query KYC record for an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v1beta1.NewQueryClient(clientCtx)
			res, err := queryClient.KycRecord(cmd.Context(), &v1beta1.QueryKYCRecordRequest{Address: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAMLProfile queries AML profile
func CmdQueryAMLProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aml-profile [address]",
		Short: "Query AML risk profile for an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v1beta1.NewQueryClient(clientCtx)
			res, err := queryClient.AmlProfile(cmd.Context(), &v1beta1.QueryAMLProfileRequest{Address: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySanctionsScreening queries sanctions screening result
func CmdQuerySanctionsScreening() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sanctions [address]",
		Short: "Query sanctions screening results for an address",
		Long: `Query sanctions screening results for an address.

Example:
  aurad query compliance sanctions aura1abc...
  aurad query compliance sanctions aura1abc... --force-refresh
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			forceRefresh, _ := cmd.Flags().GetBool("force-refresh")
			queryClient := v1beta1.NewQueryClient(clientCtx)
			res, err := queryClient.SanctionsScreening(cmd.Context(), &v1beta1.QuerySanctionsScreeningRequest{
				Address:      args[0],
				ForceRefresh: forceRefresh,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool("force-refresh", false, "Force new screening instead of using cache")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTransactionAlerts queries transaction alerts
func CmdQueryTransactionAlerts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alerts [address]",
		Short: "Query transaction monitoring alerts for an address",
		Long: `Query transaction monitoring alerts for an address.

Example:
  aurad query compliance alerts aura1abc...
  aurad query compliance alerts aura1abc... --unreviewed-only
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			unreviewedOnly, _ := cmd.Flags().GetBool("unreviewed-only")
			queryClient := v1beta1.NewQueryClient(clientCtx)
			res, err := queryClient.TransactionAlerts(cmd.Context(), &v1beta1.QueryTransactionAlertsRequest{
				Address:        args[0],
				UnreviewedOnly: unreviewedOnly,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool("unreviewed-only", false, "Show only unreviewed alerts")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTaxReport queries tax report
func CmdQueryTaxReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-report [address] [tax-year] [jurisdiction]",
		Short: "Query tax report for an address",
		Long: `Query tax report for a specific year and jurisdiction.

Example:
  aurad query compliance tax-report aura1abc... "2024" "US"
  aurad query compliance tax-report aura1abc... "2024" "UK"
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := v1beta1.NewQueryClient(clientCtx)
			res, err := queryClient.TaxReport(cmd.Context(), &v1beta1.QueryTaxReportRequest{
				Address:      args[0],
				TaxYear:      args[1],
				Jurisdiction: args[2],
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
