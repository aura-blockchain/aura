package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
)

// GetTxCmd returns the transaction commands for the networksecurity module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "networksecurity",
		Short:                      "Network security transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdAddTrustedPeer(),
		CmdRemoveTrustedPeer(),
		CmdBanPeer(),
		CmdUnbanPeer(),
		CmdUpdatePeerReputation(),
		CmdResolveForkAlert(),
		CmdResolvePartitionAlert(),
	)

	return cmd
}

// CmdAddTrustedPeer adds a trusted peer
func CmdAddTrustedPeer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-trusted-peer [peer-id] [address]",
		Short: "Add a trusted peer to the network (requires authority)",
		Long: `Add a peer to the trusted peer list for enhanced network security.

Trusted peers are prioritized for connections and given more lenient rate limits.

Examples:
  aurad tx networksecurity add-trusted-peer peer123 node1.aura.network:26656 --from authority --description "Core validator node"
  aurad tx networksecurity add-trusted-peer peer456 203.0.113.5:26656 --from authority --description "Relayer node"

Flags:
  --description: Optional description of the peer
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			peerID := args[0]
			address := args[1]
			description, _ := cmd.Flags().GetString("description")

			peer := v1beta1.TrustedPeer{
				PeerId:      peerID,
				Address:     address,
				Description: description,
			}

			msg := &v1beta1.MsgAddTrustedPeer{
				Authority: clientCtx.GetFromAddress().String(),
				Peer:      peer,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("description", "", "Description of the trusted peer")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRemoveTrustedPeer removes a trusted peer
func CmdRemoveTrustedPeer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-trusted-peer [peer-id]",
		Short: "Remove a trusted peer from the network (requires authority)",
		Long: `Remove a peer from the trusted peer list.

Examples:
  aurad tx networksecurity remove-trusted-peer peer123 --from authority
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgRemoveTrustedPeer{
				Authority: clientCtx.GetFromAddress().String(),
				PeerId:    args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdBanPeer manually bans a peer
func CmdBanPeer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ban-peer [peer-id] [duration-seconds] [reason]",
		Short: "Manually ban a peer (requires authority)",
		Long: `Ban a peer from the network for a specified duration.

Examples:
  aurad tx networksecurity ban-peer peer789 86400 "Excessive spam" --from authority
  aurad tx networksecurity ban-peer peer012 3600 "DDoS attempt" --from authority

Duration is specified in seconds.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			peerID := args[0]
			duration, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
			reason := args[2]

			msg := &v1beta1.MsgBanPeer{
				Authority:       clientCtx.GetFromAddress().String(),
				PeerId:          peerID,
				DurationSeconds: duration,
				Reason:          reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUnbanPeer unbans a peer
func CmdUnbanPeer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unban-peer [peer-id]",
		Short: "Unban a peer (requires authority)",
		Long: `Remove a peer from the ban list.

Examples:
  aurad tx networksecurity unban-peer peer789 --from authority
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgUnbanPeer{
				Authority: clientCtx.GetFromAddress().String(),
				PeerId:    args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdatePeerReputation manually updates peer reputation
func CmdUpdatePeerReputation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-reputation [peer-id] [score] [reason]",
		Short: "Manually update peer reputation score (requires authority)",
		Long: `Update the reputation score for a peer.

Reputation scores typically range from 0-100, with higher scores indicating better behavior.

Examples:
  aurad tx networksecurity update-reputation peer123 95 "Consistent uptime and performance" --from authority
  aurad tx networksecurity update-reputation peer456 20 "Frequent timeout issues" --from authority
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			peerID := args[0]
			score, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid score: %w", err)
			}
			reason := args[2]

			msg := &v1beta1.MsgUpdatePeerReputation{
				Authority: clientCtx.GetFromAddress().String(),
				PeerId:    peerID,
				Score:     score,
				Reason:    reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdResolveForkAlert marks a fork alert as resolved
func CmdResolveForkAlert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-fork-alert [alert-id] [resolution-details]",
		Short: "Mark a fork alert as resolved (requires authority)",
		Long: `Resolve a network fork alert after the issue has been addressed.

Examples:
  aurad tx networksecurity resolve-fork-alert alert123 "Fork resolved at height 12345, canonical chain confirmed" --from authority
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgResolveForkAlert{
				Authority:         clientCtx.GetFromAddress().String(),
				AlertId:           args[0],
				ResolutionDetails: args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdResolvePartitionAlert marks a partition alert as resolved
func CmdResolvePartitionAlert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-partition-alert [alert-id]",
		Short: "Mark a network partition alert as resolved (requires authority)",
		Long: `Resolve a network partition alert after the network has healed.

Network partitions occur when the network splits into disconnected groups of nodes.

Examples:
  aurad tx networksecurity resolve-partition-alert alert456 --from authority
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgResolvePartitionAlert{
				Authority: clientCtx.GetFromAddress().String(),
				AlertId:   args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
