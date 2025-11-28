package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

// printJSON prints the response as formatted JSON
func printJSON(clientCtx client.Context, v interface{}) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return clientCtx.PrintBytes(out)
}

// GetQueryCmd returns the query commands for the wasm module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the wasm module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryCode(),
		GetCmdListCode(),
		GetCmdQueryContractInfo(),
		GetCmdQueryContractState(),
		GetCmdQueryContractHistory(),
		GetCmdQuerySmartContract(),
		GetCmdQueryRawContract(),
		GetCmdQuerySecurityStats(),
		GetCmdQueryAuthorizedUploaders(),
		GetCmdQueryPausedContracts(),
		GetCmdQueryIsAuthorizedUploader(),
		GetCmdQueryIsContractPaused(),
	)

	return queryCmd
}

// GetCmdQueryParams returns the command to query module parameters
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current wasm module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryCode returns the command to query contract code by ID
func GetCmdQueryCode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code [code-id]",
		Short: "Query contract code by code ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid code id: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Code(context.Background(), &types.QueryCodeRequest{
				CodeId: codeID,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdListCode returns the command to list all stored contract codes
func GetCmdListCode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-code",
		Short: "List all stored contract codes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Codes(context.Background(), &types.QueryCodesRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "list-code")
	return cmd
}

// GetCmdQueryContractInfo returns the command to query contract info
func GetCmdQueryContractInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract [contract-address]",
		Short: "Query contract info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ContractInfo(context.Background(), &types.QueryContractInfoRequest{
				Address: contractAddr,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryContractState returns the command to query all contract state
func GetCmdQueryContractState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract-state-all [contract-address]",
		Short: "Query all state of a contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.AllContractState(context.Background(), &types.QueryAllContractStateRequest{
				Address:    contractAddr,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "contract-state-all")
	return cmd
}

// GetCmdQueryContractHistory returns the command to query contract history
func GetCmdQueryContractHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract-history [contract-address]",
		Short: "Query contract history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.ContractHistory(context.Background(), &types.QueryContractHistoryRequest{
				Address:    contractAddr,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "contract-history")
	return cmd
}

// GetCmdQuerySmartContract returns the command to query smart contract state
func GetCmdQuerySmartContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-smart [contract-address] [query-json]",
		Short: "Query smart contract state using a JSON query message",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			var queryMsg json.RawMessage
			if err := json.Unmarshal([]byte(args[1]), &queryMsg); err != nil {
				return fmt.Errorf("invalid query JSON: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.SmartContractState(context.Background(), &types.QuerySmartContractStateRequest{
				Address:   contractAddr,
				QueryData: queryMsg,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryRawContract returns the command to query raw contract state
func GetCmdQueryRawContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-raw [contract-address] [key-hex]",
		Short: "Query raw contract state by key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			key, err := hex.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("invalid hex key: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RawContractState(context.Background(), &types.QueryRawContractStateRequest{
				Address:   contractAddr,
				QueryData: key,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQuerySecurityStats returns the command to query security statistics
func GetCmdQuerySecurityStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security-stats",
		Short: "Query security statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.SecurityStats(context.Background(), &types.QuerySecurityStatsRequest{})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryAuthorizedUploaders returns the command to query authorized uploaders
func GetCmdQueryAuthorizedUploaders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorized-uploaders",
		Short: "Query all authorized contract uploaders",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.AuthorizedUploaders(context.Background(), &types.QueryAuthorizedUploadersRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "authorized-uploaders")
	return cmd
}

// GetCmdQueryPausedContracts returns the command to query paused contracts
func GetCmdQueryPausedContracts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paused-contracts",
		Short: "Query all paused contracts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.PausedContracts(context.Background(), &types.QueryPausedContractsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "paused-contracts")
	return cmd
}

// GetCmdQueryIsAuthorizedUploader returns the command to check if an address is authorized
func GetCmdQueryIsAuthorizedUploader() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is-authorized [address]",
		Short: "Check if an address is authorized to upload contracts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.IsAuthorizedUploader(context.Background(), &types.QueryIsAuthorizedUploaderRequest{
				Address: address,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryIsContractPaused returns the command to check if a contract is paused
func GetCmdQueryIsContractPaused() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is-paused [contract-address]",
		Short: "Check if a contract is paused",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.IsContractPaused(context.Background(), &types.QueryIsContractPausedRequest{
				Address: contractAddr,
			})
			if err != nil {
				return err
			}

			return printJSON(clientCtx, res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
