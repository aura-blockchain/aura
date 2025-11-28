package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	compliancecli "github.com/aequitas/aura/chain/x/compliance/client/cli"
	confidencescorecli "github.com/aequitas/aura/chain/x/confidencescore/client/cli"
	wasmcli "github.com/aequitas/aura/chain/x/wasm/client/cli"
)

// QueryCmd returns the query command for the Aura daemon
func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Query blockchain state",
		Long: `Query blockchain state and module-specific data.

Available subcommands allow querying different modules and their state.`,
	}

	cmd.AddCommand(
		queryBlockCmd(),
		queryTxCmd(),
		queryAccountCmd(),
		queryIdentityChangeCmd(),
		queryInclusionRoutinesCmd(),
		queryConfidenceScoreCmd(),
		queryComplianceCmd(),
		queryVCRegistryCmd(),
		queryDataRegistryCmd(),
		queryGovernanceCmd(),
		queryWasmCmd(),
	)

	return cmd
}

// queryBlockCmd returns the query block command
func queryBlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "block [height]",
		Short: "Query block at a given height",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			height := "latest"
			if len(args) > 0 {
				height = args[0]
			}
			fmt.Printf("Querying block: %s\n", height)
			fmt.Printf("This feature requires full Cosmos SDK integration.\n")
			return nil
		},
	}
}

// queryTxCmd returns the query transaction command
func queryTxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tx [hash]",
		Short: "Query transaction by hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txHash := args[0]
			fmt.Printf("Querying transaction: %s\n", txHash)
			fmt.Printf("This feature requires full Cosmos SDK integration.\n")
			return nil
		},
	}
}

// queryAccountCmd returns the query account command
func queryAccountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "account [address]",
		Short: "Query account information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]
			fmt.Printf("Querying account: %s\n", address)
			fmt.Printf("This feature requires full Cosmos SDK integration.\n")
			return nil
		},
	}
}

// queryIdentityChangeCmd returns the query identity change command
func queryIdentityChangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identitychange",
		Short: "Query identity change module state",
		Long:  "Query identity change records and module parameters.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "params",
			Short: "Query identity change module parameters",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying identity change parameters...\n")
				fmt.Printf("This will use gRPC to query the identitychange module.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "record [address]",
			Short: "Query identity change record for an address",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying identity change record for: %s\n", args[0])
				fmt.Printf("This will use gRPC to query the identitychange module.\n")
				return nil
			},
		},
	)

	return cmd
}

// queryInclusionRoutinesCmd returns the query inclusion routines command
func queryInclusionRoutinesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inclusionroutines",
		Short: "Query inclusion routines module state",
		Long:  "Query inclusion routines, proposals, and module parameters.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "params",
			Short: "Query inclusion routines module parameters",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying inclusion routines parameters...\n")
				fmt.Printf("This will use gRPC to query the inclusionroutines module.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "routine [id]",
			Short: "Query inclusion routine by ID",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying inclusion routine: %s\n", args[0])
				fmt.Printf("This will use gRPC to query the inclusionroutines module.\n")
				return nil
			},
		},
	)

	return cmd
}

// queryConfidenceScoreCmd returns the query confidence score command
func queryConfidenceScoreCmd() *cobra.Command {
	// Use the comprehensive CLI commands from the confidencescore module
	return confidencescorecli.GetQueryCmd()
}

func queryComplianceCmd() *cobra.Command {
	return compliancecli.GetQueryCmd()
}

// queryVCRegistryCmd returns the query VC registry command
func queryVCRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vcregistry",
		Short: "Query VC registry module state",
		Long:  "Query verifiable credentials and module parameters.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "params",
			Short: "Query VC registry module parameters",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying VC registry parameters...\n")
				fmt.Printf("This will use gRPC to query the vcregistry module.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "credential [id]",
			Short: "Query verifiable credential by ID",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying credential: %s\n", args[0])
				fmt.Printf("This will use gRPC to query the vcregistry module.\n")
				return nil
			},
		},
	)

	return cmd
}

// queryDataRegistryCmd returns the query data registry command
func queryDataRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataregistry",
		Short: "Query data registry module state",
		Long:  "Query registered data and module parameters.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "params",
			Short: "Query data registry module parameters",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying data registry parameters...\n")
				fmt.Printf("This will use gRPC to query the dataregistry module.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "data [id]",
			Short: "Query registered data by ID",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying data: %s\n", args[0])
				fmt.Printf("This will use gRPC to query the dataregistry module.\n")
				return nil
			},
		},
	)

	return cmd
}

// queryGovernanceCmd returns the query governance command
func queryGovernanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "governance",
		Short: "Query governance module state",
		Long:  "Query proposals, votes, and module parameters.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "params",
			Short: "Query governance module parameters",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying governance parameters...\n")
				fmt.Printf("This will use gRPC to query the governance module.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "proposal [id]",
			Short: "Query proposal by ID",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying proposal: %s\n", args[0])
				fmt.Printf("This will use gRPC to query the governance module.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "proposals",
			Short: "Query all proposals",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("Querying all proposals...\n")
				fmt.Printf("This will use gRPC to query the governance module.\n")
				return nil
			},
		},
	)

	return cmd
}

// queryWasmCmd returns the wasm query command
func queryWasmCmd() *cobra.Command {
	return wasmcli.GetQueryCmd()
}
