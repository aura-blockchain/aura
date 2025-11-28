package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	aipb "github.com/aequitas/aura/proto/aura/aiassistant/v1beta1"
)

// NewQueryCmd builds the query command tree for x/aiassistant.
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "aiassistant",
		Short:                      "Query commands for the AI assistant registry",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		newQueryAssistantCmd(),
		newQueryAssistantsCmd(),
		newQueryLocaleCmd(),
		newQueryParamsCmd(),
	)
	return cmd
}

func newQueryAssistantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assistant [address]",
		Short: "Fetch a specific assistant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := aipb.NewQueryClient(clientCtx)
			resp, err := queryClient.Assistant(cmd.Context(), &aipb.QueryAssistantRequest{
				AssistantAddress: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(resp)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryAssistantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assistants",
		Short: "List all registered assistants",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := aipb.NewQueryClient(clientCtx)
			resp, err := queryClient.Assistants(cmd.Context(), &aipb.QueryAssistantsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(resp)
		},
	}
	flags.AddPaginationFlagsToCmd(cmd, "assistants")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryLocaleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locale [locale-code]",
		Short: "List assistants serving a specific locale",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			if args[0] == "" {
				return fmt.Errorf("locale required")
			}
			queryClient := aipb.NewQueryClient(clientCtx)
			resp, err := queryClient.AssistantsByLocale(cmd.Context(), &aipb.QueryAssistantsByLocaleRequest{
				Locale: args[0],
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(resp)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show the AI assistant module params",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := aipb.NewQueryClient(clientCtx)
			resp, err := queryClient.Params(cmd.Context(), &aipb.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(resp)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
