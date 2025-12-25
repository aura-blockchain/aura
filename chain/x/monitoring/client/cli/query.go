// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "monitoring",
		Short:                      "Querying commands for the monitoring module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryAlerts(),
		CmdQueryAlert(),
		CmdQueryNetworkHealth(),
		CmdQueryValidatorUptime(),
		CmdQueryGasPriceTracking(),
		CmdQueryTVLMonitoring(),
		CmdQueryTransactionMonitor(),
		CmdQueryAnomalies(),
		CmdQuerySecurityEvents(),
	)

	return cmd
}

// CmdQueryParams queries module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query monitoring module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Note: This would use the actual query client once proto messages are defined
			fmt.Println("Monitoring module parameters query - proto messages needed")
			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAlerts queries all alerts
func CmdQueryAlerts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Query all monitoring alerts",
		Long: "Query all monitoring alerts with optional filters.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			severity, _ := cmd.Flags().GetString("severity")
			unresolvedOnly, _ := cmd.Flags().GetBool("unresolved")

			fmt.Printf("Querying alerts - severity: %s, unresolved only: %v\n", severity, unresolvedOnly)
			_ = clientCtx
			return nil
		},
	}

	cmd.Flags().String("severity", "", "Filter by severity")
	cmd.Flags().Bool("unresolved", false, "Show only unresolved alerts")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAlert queries a specific alert
func CmdQueryAlert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert [alert-id]",
		Short: "Query a specific alert by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("Querying alert: %s\n", args[0])
			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryNetworkHealth queries network health metrics
func CmdQueryNetworkHealth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network-health",
		Short: "Query current network health metrics",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Println("Network health metrics - proto messages needed")
			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryValidatorUptime queries validator uptime
func CmdQueryValidatorUptime() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-uptime [validator-address]",
		Short: "Query validator uptime statistics",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("Querying validator uptime: %s\n", args[0])
			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryGasPriceTracking queries gas price tracking
func CmdQueryGasPriceTracking() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gas-price-tracking",
		Short: "Query gas price tracking and trends",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Println("Gas price tracking - proto messages needed")
			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTVLMonitoring queries TVL monitoring
func CmdQueryTVLMonitoring() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tvl-monitoring",
		Short: "Query Total Value Locked (TVL) monitoring",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Println("TVL monitoring - proto messages needed")
			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTransactionMonitor queries transaction monitoring data
func CmdQueryTransactionMonitor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction-monitor [tx-hash]",
		Short: "Query transaction monitoring data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("Querying transaction monitor: %s\n", args[0])
			_ = clientCtx
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAnomalies queries anomaly detections
func CmdQueryAnomalies() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "anomalies",
		Short: "Query detected anomalies",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			anomalyType, _ := cmd.Flags().GetString("type")
			aboveThreshold, _ := cmd.Flags().GetBool("above-threshold")

			fmt.Printf("Querying anomalies - type: %s, above threshold: %v\n", anomalyType, aboveThreshold)
			_ = clientCtx
			return nil
		},
	}

	cmd.Flags().String("type", "", "Filter by anomaly type")
	cmd.Flags().Bool("above-threshold", false, "Show only anomalies above threshold")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySecurityEvents queries security events
func CmdQuerySecurityEvents() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security-events",
		Short: "Query security events",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			severity, _ := cmd.Flags().GetString("severity")

			fmt.Printf("Querying security events - severity: %s\n", severity)
			_ = clientCtx
			return nil
		},
	}

	cmd.Flags().String("severity", "", "Filter by severity")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
