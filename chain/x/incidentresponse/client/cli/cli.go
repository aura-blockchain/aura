// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for incident response module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "incidentresponse",
		Short:                      "Incident response transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       nil,
	}

	cmd.AddCommand(
		GetCmdReportIncident(),
		GetCmdUpdateIncidentStatus(),
		GetCmdRequestChainPause(),
		GetCmdResumeChain(),
		GetCmdSetWalletLimits(),
		GetCmdCreatePostMortem(),
		GetCmdCloseIncident(),
		GetCmdTriggerBackup(),
		GetCmdTriggerInsuranceClaim(),
	)

	return cmd
}

// GetQueryCmd returns the query commands for incident response module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "incidentresponse",
		Short:                      "Querying commands for incident response module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       nil,
	}

	cmd.AddCommand(
		GetCmdQueryIncident(),
		GetCmdQueryAllIncidents(),
		GetCmdQueryPauseState(),
		GetCmdQueryWalletLimits(),
		GetCmdQueryParams(),
	)

	return cmd
}

// Transaction commands

func GetCmdReportIncident() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report-incident [title] [description] [severity] [affected-systems]",
		Short: "Report a new security incident",
		Long: `Report a new security incident to the incident response system.

Severity levels: low, medium, high, critical
Affected systems should be comma-separated list.

Example:
  aurad tx incidentresponse report-incident \
    "Database breach" \
    "Unauthorized access detected" \
    critical \
    "validator-db,api-server" \
    --from mykey
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Reporting incident: %s (severity: %s)\n", args[0], args[2])
			return nil
		},
	}
	return cmd
}

func GetCmdUpdateIncidentStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status [incident-id] [status] [notes]",
		Short: "Update incident status",
		Long: `Update the status of an existing incident.

Status values: new, investigating, contained, resolved, post_mortem, closed

Example:
  aurad tx incidentresponse update-status \
    INC-001 \
    investigating \
    "Security team investigating root cause" \
    --from mykey
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Updating incident %s to status: %s\n", args[0], args[1])
			return nil
		},
	}
	return cmd
}

func GetCmdRequestChainPause() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-pause [level] [reason] [incident-id] [duration]",
		Short: "Request emergency chain pause",
		Long: `Request an emergency chain pause. Requires multi-signature authorization.

Pause levels: transactions, modules, full
Duration format: 1h, 30m, 2h30m

Example:
  aurad tx incidentresponse request-pause \
    full \
    "Critical security vulnerability" \
    INC-001 \
    2h \
    --from mykey
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			duration, err := time.ParseDuration(args[3])
			if err != nil {
				return fmt.Errorf("invalid duration: %w", err)
			}
			fmt.Printf("Requesting chain pause (level: %s, duration: %v)\n", args[0], duration)
			return nil
		},
	}
	return cmd
}

func GetCmdResumeChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume [reason]",
		Short: "Resume chain operations after pause",
		Long: `Resume chain operations after an emergency pause.

Example:
  aurad tx incidentresponse resume \
    "Vulnerability patched and verified" \
    --from mykey
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Resuming chain: %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func GetCmdSetWalletLimits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-wallet-limits [address] [max-balance] [max-tx-size] [daily-limit]",
		Short: "Set hot wallet security limits",
		Long: `Configure security limits for a hot wallet.

All amounts should be in base denomination (uaura).

Example:
  aurad tx incidentresponse set-wallet-limits \
    aura1... \
    10000000000 \
    1000000000 \
    5000000000 \
    --from mykey
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Setting wallet limits for %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func GetCmdCreatePostMortem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-postmortem [incident-id] [summary] [root-cause] [impact] [resolution]",
		Short: "Create post-mortem analysis",
		Long: `Create a post-mortem analysis for a resolved incident.

Example:
  aurad tx incidentresponse create-postmortem \
    INC-001 \
    "Database breach incident" \
    "SQL injection vulnerability" \
    "100 users affected, no funds lost" \
    "Patched vulnerability, enhanced monitoring" \
    --from mykey
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Creating post-mortem for incident %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func GetCmdCloseIncident() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close [incident-id]",
		Short: "Close an incident after post-mortem",
		Long: `Close an incident after post-mortem is completed.

Example:
  aurad tx incidentresponse close INC-001 --from mykey
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Closing incident %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func GetCmdTriggerBackup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger-backup [backup-type]",
		Short: "Trigger manual backup operation",
		Long: `Trigger a manual backup operation.

Backup types: state, validator, keys, config, full

Example:
  aurad tx incidentresponse trigger-backup state --from mykey
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Triggering %s backup\n", args[0])
			return nil
		},
	}
	return cmd
}

func GetCmdTriggerInsuranceClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger-insurance-claim [incident-id] [amount]",
		Short: "Trigger insurance claim for incident",
		Long: `Submit an insurance claim for a covered incident.

Requires multi-signature authorization.

Example:
  aurad tx incidentresponse trigger-insurance-claim \
    INC-001 \
    1000000000000 \
    --from mykey
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Triggering insurance claim for incident %s: %s\n", args[0], args[1])
			return nil
		},
	}
	return cmd
}

// Query commands

func GetCmdQueryIncident() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incident [incident-id]",
		Short: "Query incident details",
		Long: `Query detailed information about a specific incident.

Example:
  aurad query incidentresponse incident INC-001
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Querying incident %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func GetCmdQueryAllIncidents() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incidents",
		Short: "Query all incidents",
		Long: `Query all incidents in the system.

Example:
  aurad query incidentresponse incidents
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Querying all incidents...")
			return nil
		},
	}
	return cmd
}

func GetCmdQueryPauseState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause-state",
		Short: "Query chain pause state",
		Long: `Query the current chain pause state.

Example:
  aurad query incidentresponse pause-state
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Querying chain pause state...")
			return nil
		},
	}
	return cmd
}

func GetCmdQueryWalletLimits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet-limits [address]",
		Short: "Query wallet security limits",
		Long: `Query the security limits configured for a hot wallet.

Example:
  aurad query incidentresponse wallet-limits aura1...
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Querying wallet limits for %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query incident response module parameters",
		Long: `Query the current parameters of the incident response module.

Example:
  aurad query incidentresponse params
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Querying incident response parameters...")
			return nil
		},
	}
	return cmd
}
