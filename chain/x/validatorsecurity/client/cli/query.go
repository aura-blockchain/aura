package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
)

// GetQueryCmd returns the query commands for the validatorsecurity module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "validatorsecurity",
		Short:                      "Querying commands for the validator security module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryValidatorSecurityInfo(),
		CmdQueryAllValidators(),
		CmdQueryJailedValidators(),
		CmdQueryTombstonedValidators(),
		CmdQueryDoubleSignEvidences(),
		CmdQueryValidatorAlerts(),
		CmdQuerySentryNodes(),
	)

	return cmd
}

// CmdQueryParams queries the module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the validator security module parameters",
		Long: `Query the current validator security module parameters including:
  - Downtime jail duration
  - Slash fraction for double signing
  - Slash fraction for downtime
  - Minimum sentry nodes required
  - Geographic decentralization requirements
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

// CmdQueryValidatorSecurityInfo queries security info for a validator
func CmdQueryValidatorSecurityInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [validator-address]",
		Short: "Query security information for a specific validator",
		Long: `Query comprehensive security information for a validator including:
  - Hot and cold keys
  - Geographic location (region, country, coordinates)
  - Backup validators
  - Jailing status
  - Tombstone status
  - Security alerts
  - Sentry nodes

Examples:
  aurad query validatorsecurity validator auravaloper1abc...
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.ValidatorSecurityInfo(context.Background(), &v1beta1.QueryValidatorSecurityInfoRequest{
				ValidatorAddress: args[0],
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

// CmdQueryAllValidators queries all validator security info
func CmdQueryAllValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query security information for all validators",
		Long: `Query security information for all registered validators.

Returns a list of all validators with their:
  - Address
  - Security configuration
  - Geographic distribution
  - Status (active, jailed, tombstoned)

Examples:
  aurad query validatorsecurity validators
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.AllValidators(context.Background(), &v1beta1.QueryAllValidatorsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryJailedValidators queries all jailed validators
func CmdQueryJailedValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jailed",
		Short: "Query all jailed validators",
		Long: `Query all validators that are currently jailed.

Validators can be jailed for:
  - Excessive downtime
  - Missing too many blocks
  - Temporary security issues

Jailed validators can unjail themselves after addressing the issue.

Examples:
  aurad query validatorsecurity jailed
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.JailedValidators(context.Background(), &v1beta1.QueryJailedValidatorsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTombstonedValidators queries all tombstoned validators
func CmdQueryTombstonedValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tombstoned",
		Short: "Query all tombstoned validators",
		Long: `Query all validators that are permanently tombstoned.

Validators are tombstoned for severe violations:
  - Double signing
  - Malicious behavior
  - Critical security breaches

Tombstoned validators cannot be unjailed and are permanently removed.

Examples:
  aurad query validatorsecurity tombstoned
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.TombstonedValidators(context.Background(), &v1beta1.QueryTombstonedValidatorsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryDoubleSignEvidences queries all double sign evidences
func CmdQueryDoubleSignEvidences() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evidences",
		Short: "Query all double signing evidences",
		Long: `Query all recorded double signing evidences.

Each evidence includes:
  - Validator address
  - Height of double signing
  - Both conflicting votes
  - Reporter
  - Slash result

Examples:
  aurad query validatorsecurity evidences
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.DoubleSignEvidences(context.Background(), &v1beta1.QueryDoubleSignEvidencesRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryValidatorAlerts queries alerts for a validator
func CmdQueryValidatorAlerts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alerts [validator-address]",
		Short: "Query security alerts for a specific validator",
		Long: `Query all security alerts for a validator.

Alert types include:
  - Downtime warnings
  - Geographic concentration
  - Version upgrade requirements
  - Performance degradation
  - Security patches

Examples:
  aurad query validatorsecurity alerts auravaloper1abc...
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.ValidatorAlerts(context.Background(), &v1beta1.QueryValidatorAlertsRequest{
				ValidatorAddress: args[0],
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

// CmdQuerySentryNodes queries sentry nodes for a validator
func CmdQuerySentryNodes() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sentry-nodes [validator-address]",
		Short: "Query sentry nodes for a specific validator",
		Long: `Query all registered sentry nodes for a validator.

Sentry nodes provide DDoS protection by acting as a protective layer
between the validator and the public network.

Information includes:
  - Sentry node address
  - IP address
  - Port
  - Registration time
  - Status

Examples:
  aurad query validatorsecurity sentry-nodes auravaloper1abc...
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.SentryNodes(context.Background(), &v1beta1.QuerySentryNodesRequest{
				ValidatorAddress: args[0],
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
