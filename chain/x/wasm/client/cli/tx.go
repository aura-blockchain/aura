// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

const (
	flagInstantiateByEverybody = "instantiate-everybody"
	flagInstantiateNobody      = "instantiate-nobody"
	flagInstantiateOnlyAddress = "instantiate-only-address"
	flagLabel                  = "label"
	flagAdmin                  = "admin"
	flagNoAdmin                = "no-admin"
	flagAmount                 = "amount"
	flagSource                 = "source"
	flagBuilder                = "builder"
	flagCodeHash               = "code-hash"
)

// GetTxCmd returns the transaction commands for the wasm module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Wasm transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdStoreCode(),
		GetCmdInstantiateContract(),
		GetCmdExecuteContract(),
		GetCmdMigrateContract(),
		GetCmdUpdateAdmin(),
		GetCmdClearAdmin(),
		GetCmdAuthorizeUploader(),
		GetCmdRevokeUploader(),
		GetCmdPauseContract(),
		GetCmdUnpauseContract(),
		GetCmdUpdateParams(),
	)

	return txCmd
}

// GetCmdStoreCode returns the command to upload a wasm contract code
func GetCmdStoreCode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store [wasm-file]",
		Short: "Upload a wasm binary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read wasm file
			wasmCode, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("error reading wasm file: %w", err)
			}

			// Validate file size
			if len(wasmCode) == 0 {
				return fmt.Errorf("wasm file cannot be empty")
			}

			msg := &types.MsgStoreCode{
				Sender:       clientCtx.GetFromAddress().String(),
				WASMByteCode: wasmCode,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdInstantiateContract returns the command to instantiate a wasm contract
func GetCmdInstantiateContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instantiate [code-id] [init-msg] --label [label] --admin [admin-address] --amount [coins]",
		Short: "Instantiate a wasm contract",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse code ID
			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid code id: %w", err)
			}

			// Parse init message
			var initMsg json.RawMessage
			if err := json.Unmarshal([]byte(args[1]), &initMsg); err != nil {
				return fmt.Errorf("invalid init message JSON: %w", err)
			}

			// Get flags
			label, err := cmd.Flags().GetString(flagLabel)
			if err != nil {
				return err
			}
			if label == "" {
				return fmt.Errorf("label is required")
			}

			admin, err := cmd.Flags().GetString(flagAdmin)
			if err != nil {
				return err
			}

			noAdmin, err := cmd.Flags().GetBool(flagNoAdmin)
			if err != nil {
				return err
			}

			if noAdmin {
				admin = ""
			}

			amountStr, err := cmd.Flags().GetString(flagAmount)
			if err != nil {
				return err
			}

			var funds sdk.Coins
			if amountStr != "" {
				funds, err = sdk.ParseCoinsNormalized(amountStr)
				if err != nil {
					return fmt.Errorf("invalid amount: %w", err)
				}
			}

			msg := &types.MsgInstantiateContract{
				Sender: clientCtx.GetFromAddress().String(),
				Admin:  admin,
				CodeID: codeID,
				Label:  label,
				Msg:    types.RawContractMessage(initMsg),
				Funds:  funds,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagLabel, "", "A human-readable label for this contract")
	cmd.Flags().String(flagAdmin, "", "Address of an admin")
	cmd.Flags().Bool(flagNoAdmin, false, "You must set this explicitly if you don't want an admin")
	cmd.Flags().String(flagAmount, "", "Coins to send to the contract during instantiation")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdExecuteContract returns the command to execute a wasm contract
func GetCmdExecuteContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute [contract-address] [execute-msg] --amount [coins]",
		Short: "Execute a command on a wasm contract",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			// Parse execute message
			var execMsg json.RawMessage
			if err := json.Unmarshal([]byte(args[1]), &execMsg); err != nil {
				return fmt.Errorf("invalid execute message JSON: %w", err)
			}

			amountStr, err := cmd.Flags().GetString(flagAmount)
			if err != nil {
				return err
			}

			var funds sdk.Coins
			if amountStr != "" {
				funds, err = sdk.ParseCoinsNormalized(amountStr)
				if err != nil {
					return fmt.Errorf("invalid amount: %w", err)
				}
			}

			msg := &types.MsgExecuteContract{
				Sender:   clientCtx.GetFromAddress().String(),
				Contract: contractAddr,
				Msg:      types.RawContractMessage(execMsg),
				Funds:    funds,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagAmount, "", "Coins to send to the contract")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdMigrateContract returns the command to migrate a wasm contract
func GetCmdMigrateContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [contract-address] [new-code-id] [migrate-msg]",
		Short: "Migrate a wasm contract to a new code version",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			newCodeID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid new code id: %w", err)
			}

			var migrateMsg json.RawMessage
			if err := json.Unmarshal([]byte(args[2]), &migrateMsg); err != nil {
				return fmt.Errorf("invalid migrate message JSON: %w", err)
			}

			msg := &types.MsgMigrateContract{
				Sender:   clientCtx.GetFromAddress().String(),
				Contract: contractAddr,
				CodeID:   newCodeID,
				Msg:      types.RawContractMessage(migrateMsg),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUpdateAdmin returns the command to update the admin of a wasm contract
func GetCmdUpdateAdmin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-admin [contract-address] [new-admin-address]",
		Short: "Set a new admin for a contract",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			newAdmin := args[1]
			if _, err := sdk.AccAddressFromBech32(newAdmin); err != nil {
				return fmt.Errorf("invalid new admin address: %w", err)
			}

			msg := &types.MsgUpdateAdmin{
				Sender:   clientCtx.GetFromAddress().String(),
				Contract: contractAddr,
				NewAdmin: newAdmin,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdClearAdmin returns the command to clear the admin of a wasm contract
func GetCmdClearAdmin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear-admin [contract-address]",
		Short: "Clear admin for a contract to prevent further migrations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			msg := &types.MsgClearAdmin{
				Sender:   clientCtx.GetFromAddress().String(),
				Contract: contractAddr,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAuthorizeUploader returns the command to authorize a contract uploader
func GetCmdAuthorizeUploader() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorize-uploader [address]",
		Short: "Authorize an address to upload contracts (governance only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			uploaderAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(uploaderAddr); err != nil {
				return fmt.Errorf("invalid uploader address: %w", err)
			}

			msg := &types.MsgAuthorizeUploader{
				Authority: clientCtx.GetFromAddress().String(),
				Uploader:  uploaderAddr,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRevokeUploader returns the command to revoke a contract uploader
func GetCmdRevokeUploader() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-uploader [address]",
		Short: "Revoke an address's permission to upload contracts (governance only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			uploaderAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(uploaderAddr); err != nil {
				return fmt.Errorf("invalid uploader address: %w", err)
			}

			msg := &types.MsgRevokeUploader{
				Authority: clientCtx.GetFromAddress().String(),
				Uploader:  uploaderAddr,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdPauseContract returns the command to pause a contract
func GetCmdPauseContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause-contract [contract-address]",
		Short: "Pause a contract to prevent execution (governance only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			msg := &types.MsgPauseContract{
				Authority: clientCtx.GetFromAddress().String(),
				Contract:  contractAddr,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUnpauseContract returns the command to unpause a contract
func GetCmdUnpauseContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unpause-contract [contract-address]",
		Short: "Unpause a contract to allow execution (governance only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contractAddr := args[0]
			if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
				return fmt.Errorf("invalid contract address: %w", err)
			}

			msg := &types.MsgUnpauseContract{
				Authority: clientCtx.GetFromAddress().String(),
				Contract:  contractAddr,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdUpdateParams returns the command to update module parameters
func GetCmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [params-json]",
		Short: "Update module parameters (governance only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var params types.Params
			if err := json.Unmarshal([]byte(args[0]), &params); err != nil {
				return fmt.Errorf("invalid params JSON: %w", err)
			}

			msg := &types.MsgUpdateParams{
				Authority: clientCtx.GetFromAddress().String(),
				Params:    params,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
