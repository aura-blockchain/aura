// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
)

// GetTxCmd returns the transaction commands for the validatorsecurity module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "validatorsecurity",
		Short:                      "Validator security transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRegisterValidator(),
		CmdUpdateSecurityInfo(),
		CmdRegisterSentryNode(),
		CmdReportDoubleSign(),
		CmdUnjail(),
		CmdAcknowledgeAlert(),
	)

	return cmd
}

// CmdRegisterValidator registers a validator with security information
func CmdRegisterValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-validator [hot-key] [cold-key] [region] [country-code]",
		Short: "Register a validator with security information",
		Long: `Register a validator with comprehensive security information including:
  - Hot and cold key separation
  - Geographic location for decentralization
  - Backup validators for redundancy

Examples:
  aurad tx validatorsecurity register-validator hot123 cold456 us-west US --from validator --latitude 37.7749 --longitude -122.4194 --backup-validators "val1,val2"
  aurad tx validatorsecurity register-validator hot789 cold012 eu-central DE --from validator --latitude 52.5200 --longitude 13.4050

Flags:
  --latitude: Validator's latitude
  --longitude: Validator's longitude
  --backup-validators: Comma-separated list of backup validator addresses
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hotKey := args[0]
			coldKey := args[1]
			region := args[2]
			countryCode := args[3]

			latitude, _ := cmd.Flags().GetFloat64("latitude")
			longitude, _ := cmd.Flags().GetFloat64("longitude")
			backupValidators, _ := cmd.Flags().GetString("backup-validators")

			var backupAddrs []string
			if backupValidators != "" {
				backupAddrs = strings.Split(backupValidators, ",")
			}

			msg := &v1beta1.MsgRegisterValidator{
				ValidatorAddress:         clientCtx.GetFromAddress().String(),
				HotKey:                   hotKey,
				ColdKey:                  coldKey,
				Region:                   region,
				CountryCode:              countryCode,
				Latitude:                 latitude,
				Longitude:                longitude,
				BackupValidatorAddresses: backupAddrs,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Float64("latitude", 0.0, "Validator's latitude")
	cmd.Flags().Float64("longitude", 0.0, "Validator's longitude")
	cmd.Flags().String("backup-validators", "", "Comma-separated list of backup validator addresses")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateSecurityInfo updates validator security information
func CmdUpdateSecurityInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-security-info [hot-key] [cold-key] [region] [country-code]",
		Short: "Update validator security information",
		Long: `Update existing validator security information.

Examples:
  aurad tx validatorsecurity update-security-info hot999 cold888 us-east US --from validator --latitude 40.7128 --longitude -74.0060
  aurad tx validatorsecurity update-security-info hot777 cold666 ap-south IN --from validator --latitude 28.6139 --longitude 77.2090 --backup-validators "val3,val4"
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hotKey := args[0]
			coldKey := args[1]
			region := args[2]
			countryCode := args[3]

			latitude, _ := cmd.Flags().GetFloat64("latitude")
			longitude, _ := cmd.Flags().GetFloat64("longitude")
			backupValidators, _ := cmd.Flags().GetString("backup-validators")

			var backupAddrs []string
			if backupValidators != "" {
				backupAddrs = strings.Split(backupValidators, ",")
			}

			msg := &v1beta1.MsgUpdateSecurityInfo{
				ValidatorAddress:         clientCtx.GetFromAddress().String(),
				HotKey:                   hotKey,
				ColdKey:                  coldKey,
				Region:                   region,
				CountryCode:              countryCode,
				Latitude:                 latitude,
				Longitude:                longitude,
				BackupValidatorAddresses: backupAddrs,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Float64("latitude", 0.0, "Validator's latitude")
	cmd.Flags().Float64("longitude", 0.0, "Validator's longitude")
	cmd.Flags().String("backup-validators", "", "Comma-separated list of backup validator addresses")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRegisterSentryNode registers a sentry node for a validator
func CmdRegisterSentryNode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-sentry [sentry-address] [ip-address] [port]",
		Short: "Register a sentry node for DDoS protection",
		Long: `Register a sentry node to protect the validator from DDoS attacks.

Sentry nodes act as a protective layer between the validator and the public network.

Examples:
  aurad tx validatorsecurity register-sentry sentry1abc... 203.0.113.5 26656 --from validator
  aurad tx validatorsecurity register-sentry sentry2def... 198.51.100.10 26656 --from validator
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sentryAddress := args[0]
			ipAddress := args[1]
			port, err := strconv.ParseInt(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid port: %w", err)
			}

			msg := &v1beta1.MsgRegisterSentryNode{
				ValidatorAddress: clientCtx.GetFromAddress().String(),
				SentryAddress:    sentryAddress,
				IpAddress:        ipAddress,
				Port:             int32(port),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdReportDoubleSign reports double signing evidence
func CmdReportDoubleSign() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report-double-sign [validator-address] [height] [vote-a-hex] [vote-b-hex]",
		Short: "Report double signing evidence",
		Long: `Report evidence of a validator double signing at a specific height.

Double signing is a critical security violation where a validator signs two different blocks
at the same height, which can lead to network consensus issues.

Examples:
  aurad tx validatorsecurity report-double-sign auravaloper1abc... 12345 0xvoteA... 0xvoteB... --from reporter
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validatorAddress := args[0]
			height, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid height: %w", err)
			}

			voteA, err := hex.DecodeString(strings.TrimPrefix(args[2], "0x"))
			if err != nil {
				return fmt.Errorf("invalid vote A hex: %w", err)
			}

			voteB, err := hex.DecodeString(strings.TrimPrefix(args[3], "0x"))
			if err != nil {
				return fmt.Errorf("invalid vote B hex: %w", err)
			}

			msg := &v1beta1.MsgReportDoubleSign{
				ReporterAddress:  clientCtx.GetFromAddress().String(),
				ValidatorAddress: validatorAddress,
				Height:           height,
				VoteA:            voteA,
				VoteB:            voteB,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUnjail allows a jailed validator to unjail themselves
func CmdUnjail() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail",
		Short: "Unjail a validator that was jailed for downtime",
		Long: `Unjail a validator that was jailed for downtime or other non-severe violations.

Requirements:
  - Validator must have been jailed (not tombstoned)
  - Sufficient time must have passed since jailing
  - Validator must fix the underlying issue

Examples:
  aurad tx validatorsecurity unjail --from validator
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgUnjail{
				ValidatorAddress: sdk.ValAddress(clientCtx.GetFromAddress()).String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAcknowledgeAlert acknowledges a validator security alert
func CmdAcknowledgeAlert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acknowledge-alert [alert-id]",
		Short: "Acknowledge a validator security alert",
		Long: `Acknowledge receipt and review of a security alert.

Security alerts may include:
  - Unusual downtime patterns
  - Geographic proximity warnings
  - Upgrade requirements
  - Performance degradation

Examples:
  aurad tx validatorsecurity acknowledge-alert alert-123 --from validator
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			alertID := args[0]

			msg := &v1beta1.MsgAcknowledgeAlert{
				AlertId:             alertID,
				AcknowledgerAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
