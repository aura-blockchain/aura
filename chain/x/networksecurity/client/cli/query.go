// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
)

// GetQueryCmd returns the query commands for the networksecurity module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "networksecurity",
		Short:                      "Querying commands for the network security module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryPeerInfo(),
		CmdQueryAllPeers(),
		CmdQueryTrustedPeers(),
		CmdQueryPeerReputation(),
		CmdQueryRateLimitStatus(),
		CmdQueryMempoolStats(),
		CmdQueryForkAlerts(),
		CmdQueryPartitionAlerts(),
		CmdQueryNetworkHealth(),
	)

	return cmd
}

// CmdQueryParams queries the module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the network security module parameters",
		Long: `Query the current network security module parameters including:
  - Rate limiting settings
  - Peer reputation thresholds
  - Ban durations
  - Fork detection parameters
  - Partition detection settings

Examples:
  aurad query networksecurity params
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

// CmdQueryPeerInfo queries information about a specific peer
func CmdQueryPeerInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peer [peer-id]",
		Short: "Query information about a specific peer",
		Long: `Query detailed information about a peer including:
  - Peer ID and address
  - Connection status
  - Reputation score
  - Rate limit status
  - Ban status
  - Recent behavior metrics

Examples:
  aurad query networksecurity peer peer123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.PeerInfo(context.Background(), &v1beta1.QueryPeerInfoRequest{
				PeerId: args[0],
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

// CmdQueryAllPeers queries information about all connected peers
func CmdQueryAllPeers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peers",
		Short: "Query information about all connected peers",
		Long: `Query information about all currently connected peers.

Returns a list of all peers with their:
  - Peer ID and address
  - Connection status
  - Reputation score
  - Recent activity

Examples:
  aurad query networksecurity peers
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.AllPeers(context.Background(), &v1beta1.QueryAllPeersRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTrustedPeers queries all trusted peers
func CmdQueryTrustedPeers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trusted-peers",
		Short: "Query all trusted peers",
		Long: `Query the list of all trusted peers.

Trusted peers receive preferential treatment for:
  - Connection priority
  - Higher rate limits
  - Enhanced security features

Examples:
  aurad query networksecurity trusted-peers
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.TrustedPeers(context.Background(), &v1beta1.QueryTrustedPeersRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPeerReputation queries reputation for a specific peer
func CmdQueryPeerReputation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reputation [peer-id]",
		Short: "Query reputation score for a specific peer",
		Long: `Query the reputation score and history for a peer.

Reputation is based on:
  - Uptime and availability
  - Message validity
  - Response times
  - Adherence to protocol

Examples:
  aurad query networksecurity reputation peer123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.PeerReputation(context.Background(), &v1beta1.QueryPeerReputationRequest{
				PeerId: args[0],
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

// CmdQueryRateLimitStatus queries rate limit status for a peer
func CmdQueryRateLimitStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit [peer-id]",
		Short: "Query rate limit status for a specific peer",
		Long: `Query the current rate limit status for a peer including:
  - Current rate limit tier
  - Requests in current window
  - Remaining capacity
  - Window reset time

Examples:
  aurad query networksecurity rate-limit peer123
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.RateLimitStatus(context.Background(), &v1beta1.QueryRateLimitStatusRequest{
				PeerId: args[0],
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

// CmdQueryMempoolStats queries current mempool statistics
func CmdQueryMempoolStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mempool-stats",
		Short: "Query current mempool statistics",
		Long: `Query mempool statistics including:
  - Total transactions in mempool
  - Transaction sizes
  - Fee distribution
  - Spam detection metrics
  - Priority queue status

Examples:
  aurad query networksecurity mempool-stats
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.MempoolStats(context.Background(), &v1beta1.QueryMempoolStatsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryForkAlerts queries active fork alerts
func CmdQueryForkAlerts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fork-alerts",
		Short: "Query active fork alerts",
		Long: `Query all fork alerts in the network.

Fork alerts are raised when multiple chains are detected at the same height,
which may indicate:
  - Network consensus issues
  - Validator misbehavior
  - Software bugs

Examples:
  aurad query networksecurity fork-alerts
  aurad query networksecurity fork-alerts --include-resolved
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			includeResolved, _ := cmd.Flags().GetBool("include-resolved")

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.ForkAlerts(context.Background(), &v1beta1.QueryForkAlertsRequest{
				IncludeResolved: includeResolved,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool("include-resolved", false, "Include resolved fork alerts")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPartitionAlerts queries active partition alerts
func CmdQueryPartitionAlerts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "partition-alerts",
		Short: "Query active network partition alerts",
		Long: `Query all network partition alerts.

Partition alerts are raised when the network splits into disconnected groups,
which can be caused by:
  - Network infrastructure issues
  - Geographic connectivity problems
  - DDoS attacks
  - Routing issues

Examples:
  aurad query networksecurity partition-alerts
  aurad query networksecurity partition-alerts --include-resolved
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			includeResolved, _ := cmd.Flags().GetBool("include-resolved")

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.PartitionAlerts(context.Background(), &v1beta1.QueryPartitionAlertsRequest{
				IncludeResolved: includeResolved,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool("include-resolved", false, "Include resolved partition alerts")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryNetworkHealth queries overall network health status
func CmdQueryNetworkHealth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Query overall network health status",
		Long: `Query comprehensive network health metrics including:
  - Total connected peers
  - Healthy peer count
  - Banned peer count
  - Average reputation score
  - Partition status
  - Active fork status
  - Mempool statistics

This provides a high-level overview of the network's security and health.

Examples:
  aurad query networksecurity health
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.NetworkHealth(context.Background(), &v1beta1.QueryNetworkHealthRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
