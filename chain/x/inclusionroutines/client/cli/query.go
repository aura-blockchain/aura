package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "inclusionroutines",
		Short:                      "Querying commands for the inclusionroutines module",
		Aliases:                    []string{"ir"},
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryIR(),
		GetCmdQueryListIRs(),
		GetCmdQueryIRGraph(),
		GetCmdQueryRateLimit(),
		GetCmdQueryParams(),
	)

	return cmd
}

// GetCmdQueryIR queries a specific IR by ID
func GetCmdQueryIR() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [ir-id]",
		Short: "Query an Inclusion Routine by ID",
		Long: `Query detailed information about a specific Inclusion Routine including:
  - Name and description
  - Arena (verification category)
  - Score and POI reward values
  - Locale tags
  - Privacy tier
  - Version and metadata hash
  - Status (draft, reviewing, approved, active, suspended, deprecated, retired)
  - Activation and sunset heights

Examples:
  aurad query ir show "gov-id-verify"
  aurad query ir show "biometric-face"
  aurad query ir show "social-graph-verify"
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryIRRequest{
				Id: args[0],
			}

			res, err := queryClient.IR(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryListIRs queries all IRs with optional filters
func GetCmdQueryListIRs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all Inclusion Routines with optional filters",
		Long: `List all Inclusion Routines in the system with optional filtering by:
  - Status (draft, reviewing, approved, active, suspended, deprecated, retired)
  - Arena (verification category)
  - Locale

Status values:
  0: UNSPECIFIED (all statuses)
  1: DRAFT - Under development
  2: REVIEWING - Under review
  3: APPROVED - Approved but not yet active
  4: ACTIVE - Currently active and accepting completions
  5: SUSPENDED - Temporarily disabled
  6: DEPRECATED - No longer recommended but still functional
  7: RETIRED - No longer functional

Arena values:
  0: UNSPECIFIED (all arenas)
  1: ANCHOR - Core identity verification
  2: BIOMETRIC - Biometric verification
  3: POSSESSION - Device/asset possession
  4: KNOWLEDGE - Knowledge-based verification
  5: SOCIAL - Social graph verification
  6: GEOLOCATION - Location-based verification
  7: HIGH_ASSURANCE - High-security verification
  8: PERSISTENCE - Long-term verification
  9: SPECIALIZED - Custom verification

Examples:
  aurad query ir list
  aurad query ir list --status 4 (list only ACTIVE IRs)
  aurad query ir list --arena 1 (list only ANCHOR arena IRs)
  aurad query ir list --locale "US" (list IRs available in US)
  aurad query ir list --status 4 --arena 2 --locale "GLOBAL"
  aurad query ir list --page 2 --limit 50
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			statusInt, err := cmd.Flags().GetInt32("status")
			if err != nil {
				return err
			}
			status := v1beta1.IRStatus(statusInt)

			arenaInt, err := cmd.Flags().GetInt32("arena")
			if err != nil {
				return err
			}
			arena := v1beta1.Arena(arenaInt)

			locale, err := cmd.Flags().GetString("locale")
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &v1beta1.QueryListIRsRequest{
				StatusFilter: status,
				ArenaFilter:  arena,
				LocaleFilter: locale,
				Pagination:   pageReq,
			}

			res, err := queryClient.ListIRs(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Int32("status", 0, "Filter by status (0=all, 1=draft, 2=reviewing, 3=approved, 4=active, 5=suspended, 6=deprecated, 7=retired)")
	cmd.Flags().Int32("arena", 0, "Filter by arena (0=all, 1=anchor, 2=biometric, 3=possession, 4=knowledge, 5=social, 6=geolocation, 7=high_assurance, 8=persistence, 9=specialized)")
	cmd.Flags().String("locale", "", "Filter by locale tag (e.g., US, UK, GLOBAL)")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "IRs")
	return cmd
}

// GetCmdQueryIRGraph queries the prerequisite dependency graph for an IR
func GetCmdQueryIRGraph() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph [ir-id]",
		Short: "Query the prerequisite dependency graph for an IR",
		Long: `Query the complete prerequisite dependency graph showing:
  - Which IRs must be completed before this IR (depends_on)
  - Which IRs require this IR as a prerequisite (required_by)
  - Full dependency chain

This helps understand:
  - What users must complete before attempting an IR
  - Which advanced IRs become available after completing an IR
  - The complete verification pathway

Examples:
  aurad query ir graph "advanced-biometric"
  aurad query ir graph "gov-id-verify"
  aurad query ir graph "high-assurance-verify"

The graph shows the entire dependency tree, not just direct prerequisites.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryIRGraphRequest{
				IrId: args[0],
			}

			res, err := queryClient.IRGraph(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryRateLimit queries rate limit settings for an IR
func GetCmdQueryRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit [ir-id]",
		Short: "Query rate limit settings for an IR",
		Long: `Query the current rate limit configuration for an Inclusion Routine:
  - per_wallet_per_hour: Maximum attempts per wallet per hour
  - per_wallet_per_day: Maximum attempts per wallet per day
  - per_block_global: Maximum total attempts per block across all users

Rate limits help:
  - Prevent spam and abuse
  - Ensure fair access to verification tasks
  - Protect verification service resources
  - Maintain system stability

Examples:
  aurad query ir rate-limit "gov-id-verify"
  aurad query ir rate-limit "biometric-face"
  aurad query ir rate-limit "simple-captcha"

A value of 0 means unlimited for that category.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryRateLimitRequest{
				IrId: args[0],
			}

			res, err := queryClient.RateLimit(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryParams queries module parameters
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query inclusionroutines module parameters",
		Long: `Query the current parameters for the inclusionroutines module:
  - max_ir_per_locale: Maximum number of IRs allowed per locale
  - default_rate_limit_hour: Default hourly rate limit for new IRs
  - suspension_fee: Fee required to suspend an IR (spam prevention)
  - min_governance_deposit: Minimum deposit to propose IR changes via governance

Examples:
  aurad query ir params
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryParamsRequest{}

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
