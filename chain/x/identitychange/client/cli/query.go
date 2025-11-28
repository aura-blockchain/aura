package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	identitychangev1beta1 "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

// GetQueryCmd returns the query commands for the identitychange module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "identitychange",
		Short:                      "Querying commands for the identitychange module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryIdentityRecord(),
		CmdQueryIdentityChangeRequest(),
		CmdQueryIdentityChangeHistory(),
	)

	return cmd
}

// CmdQueryIdentityRecord queries an identity record by DID
func CmdQueryIdentityRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record [did]",
		Short: "Query an identity record by DID",
		Long: `Query the current identity record for a specific DID.

Examples:
  aurad query identitychange record did:aura:abc123
  aurad query identitychange record did:aura:def456

Returns:
  - DID and owner address
  - Current confidence score
  - Metadata hash
  - Latest IR version
  - Last changed block height
  - Current status
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := identitychangev1beta1.NewQueryClient(clientCtx)

			req := &identitychangev1beta1.QueryIdentityRecordRequest{
				Did: args[0],
			}

			res, err := queryClient.IdentityRecord(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryIdentityChangeRequest queries a specific identity change request
func CmdQueryIdentityChangeRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request [request-id]",
		Short: "Query an identity change request by ID",
		Long: `Query detailed information about a specific identity change request.

Examples:
  aurad query identitychange request req-123
  aurad query identitychange request req-456

Returns:
  - Request ID and target DID
  - Requester and assistant addresses
  - Associated IR ID
  - Proof and metadata hashes
  - Current status
  - Reason (if rejected)
  - Creation and verdict block heights
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := identitychangev1beta1.NewQueryClient(clientCtx)

			req := &identitychangev1beta1.QueryIdentityChangeRequestRequest{
				RequestId: args[0],
			}

			res, err := queryClient.IdentityChangeRequest(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryIdentityChangeHistory queries the change history for a DID
func CmdQueryIdentityChangeHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [did]",
		Short: "Query the change history for a DID",
		Long: `Query all historical identity changes for a specific DID.

Examples:
  aurad query identitychange history did:aura:abc123
  aurad query identitychange history did:aura:def456

Returns:
  List of all historical changes with:
  - Request IDs
  - Previous and new confidence scores
  - Transition reasons
  - Block heights when changes occurred

Supports pagination flags.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := identitychangev1beta1.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &identitychangev1beta1.QueryIdentityChangeHistoryRequest{
				Did:        args[0],
				Pagination: pageReq,
			}

			res, err := queryClient.IdentityChangeHistory(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "identity change history")
	return cmd
}
