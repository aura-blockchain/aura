package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	compliancecli "github.com/aequitas/aura/chain/x/compliance/client/cli"
	confidencescorecli "github.com/aequitas/aura/chain/x/confidencescore/client/cli"
	wasmcli "github.com/aequitas/aura/chain/x/wasm/client/cli"
)

// QueryCmd returns the query command tree for all modules plus standard RPC helpers.
func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Query blockchain state",
		Long:    "Query blockchain state, transactions, accounts, and module data via gRPC/RPC.",
		RunE:    client.ValidateCmd,
	}

	cmd.AddCommand(
		authcli.QueryTxCmd(),
		authcli.QueryTxsByEventsCmd(),
		rpc.ValidatorCommand(),
		rpc.WaitTxCmd(),
		rpc.QueryEventForTxCmd(),
		queryAccountCmd(),
	)

	cmd.AddCommand(
		confidencescorecli.GetQueryCmd(),
		compliancecli.GetQueryCmd(),
		wasmcli.GetQueryCmd(),
	)

	return cmd
}

// queryAccountCmd fetches account metadata (including account-number/sequence) via gRPC.
func queryAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query account information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			queryClient := authtypes.NewQueryClient(clientCtx)
			res, err := queryClient.Account(cmd.Context(), &authtypes.QueryAccountRequest{Address: addr.String()})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}
