// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "auth",
		Short:                      "auth transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdCreateRole(),
		GetCmdAssignRole(),
		GetCmdRevokeRole(),
		GetCmdCreateMultisigWallet(),
		GetCmdCreateMultisigProposal(),
		GetCmdSignMultisigProposal(),
		GetCmdExecuteMultisigProposal(),
		GetCmdProposeTimeLockedAction(),
		GetCmdExecuteTimeLockedAction(),
		GetCmdCancelTimeLockedAction(),
		GetCmdActivateEmergencyAdmin(),
		GetCmdDeactivateEmergencyAdmin(),
		GetCmdInitiateValidatorKeyRotation(),
		GetCmdCompleteValidatorKeyRotation(),
		GetCmdCreateSession(),
		GetCmdRevokeSession(),
	)

	return cmd
}

// GetCmdCreateRole returns the command to create a new role
func GetCmdCreateRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-role [name] [permissions] [description]",
		Short: "Create a new role with specified permissions",
		Long: `Create a new role with a name, list of permissions, and description.
Permissions should be comma-separated.

Example:
  aurad tx auth create-role admin "CREATE_ROLE,ASSIGN_ROLE,MANAGE_PARAMS" "Administrator role" --from mykey
  aurad tx auth create-role operator "EXECUTE_TX,VIEW_LOGS" "Operator role" --from mykey`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name := args[0]
			permissions := strings.Split(args[1], ",")
			description := args[2]

			// Trim whitespace from permissions
			for i, perm := range permissions {
				permissions[i] = strings.TrimSpace(perm)
			}

			msg := &v1beta1.MsgCreateRole{
				Creator:     clientCtx.GetFromAddress().String(),
				Name:        name,
				Permissions: permissions,
				Description: description,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAssignRole returns the command to assign a role to an address
func GetCmdAssignRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign-role [address] [role-name]",
		Short: "Assign a role to an address",
		Long: `Assign a role to a specific address with optional expiration.

Example:
  aurad tx auth assign-role aura1abc... admin --from mykey
  aurad tx auth assign-role aura1abc... operator --expires-in 86400 --from mykey`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			roleName := args[1]

			expiresIn, err := cmd.Flags().GetInt64("expires-in")
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgAssignRole{
				Assigner:         clientCtx.GetFromAddress().String(),
				Address:          address,
				RoleName:         roleName,
				ExpiresInSeconds: expiresIn,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int64("expires-in", 0, "Expiration time in seconds (0 for no expiry)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRevokeRole returns the command to revoke a role from an address
func GetCmdRevokeRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-role [address] [role-name]",
		Short: "Revoke a role from an address",
		Long: `Revoke a role assignment from a specific address.

Example:
  aurad tx auth revoke-role aura1abc... admin --from mykey`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]
			roleName := args[1]

			msg := &v1beta1.MsgRevokeRole{
				Revoker:  clientCtx.GetFromAddress().String(),
				Address:  address,
				RoleName: roleName,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateMultisigWallet returns the command to create a multisig wallet
func GetCmdCreateMultisigWallet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-multisig-wallet [signers] [threshold]",
		Short: "Create a multisig wallet",
		Long: `Create a multisig wallet with specified signers and threshold.
Signers should be comma-separated addresses.

Example:
  aurad tx auth create-multisig-wallet "aura1abc...,aura1def...,aura1ghi..." 2 --wallet-type 3-of-5 --from mykey`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signers := strings.Split(args[0], ",")
			for i, signer := range signers {
				signers[i] = strings.TrimSpace(signer)
			}

			threshold, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid threshold: %w", err)
			}

			walletTypeStr, err := cmd.Flags().GetString("wallet-type")
			if err != nil {
				return err
			}

			walletType := parseWalletType(walletTypeStr)

			msg := &v1beta1.MsgCreateMultisigWallet{
				Creator:    clientCtx.GetFromAddress().String(),
				Signers:    signers,
				Threshold:  uint32(threshold),
				WalletType: walletType,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("wallet-type", "custom", "Wallet type (3-of-5, 5-of-7, custom)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateMultisigProposal returns the command to create a multisig proposal
func GetCmdCreateMultisigProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-multisig-proposal [wallet-id] [title] [description] [payload-hex]",
		Short: "Create a multisig proposal",
		Long: `Create a proposal that requires multiple signatures to execute.
Payload should be hex-encoded transaction data.

Example:
  aurad tx auth create-multisig-proposal wallet123 "Transfer funds" "Transfer 100 tokens" abcd1234 --expires-in 86400 --from mykey`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			walletID := args[0]
			title := args[1]
			description := args[2]

			payload, err := hex.DecodeString(args[3])
			if err != nil {
				return fmt.Errorf("invalid payload hex: %w", err)
			}

			expiresIn, err := cmd.Flags().GetInt64("expires-in")
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgCreateMultisigProposal{
				Proposer:         clientCtx.GetFromAddress().String(),
				WalletId:         walletID,
				Title:            title,
				Description:      description,
				Payload:          payload,
				ExpiresInSeconds: expiresIn,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int64("expires-in", 86400, "Expiration time in seconds")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdSignMultisigProposal returns the command to sign a multisig proposal
func GetCmdSignMultisigProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign-multisig-proposal [proposal-id]",
		Short: "Sign a multisig proposal",
		Long: `Sign a pending multisig proposal.

Example:
  aurad tx auth sign-multisig-proposal proposal123 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID := args[0]

			msg := &v1beta1.MsgSignMultisigProposal{
				Signer:     clientCtx.GetFromAddress().String(),
				ProposalId: proposalID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdExecuteMultisigProposal returns the command to execute a multisig proposal
func GetCmdExecuteMultisigProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-multisig-proposal [proposal-id]",
		Short: "Execute an approved multisig proposal",
		Long: `Execute a multisig proposal that has reached the required threshold of signatures.

Example:
  aurad tx auth execute-multisig-proposal proposal123 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID := args[0]

			msg := &v1beta1.MsgExecuteMultisigProposal{
				Executor:   clientCtx.GetFromAddress().String(),
				ProposalId: proposalID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdProposeTimeLockedAction returns the command to propose a time-locked action
func GetCmdProposeTimeLockedAction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose-timelocked-action [action-type] [payload-hex] [delay-seconds]",
		Short: "Propose a time-locked admin action",
		Long: `Propose an admin action that will be executable after a specified delay.
Payload should be hex-encoded action data.

Example:
  aurad tx auth propose-timelocked-action UPDATE_PARAMS abcd1234 86400 --from mykey`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			actionType := args[0]

			payload, err := hex.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("invalid payload hex: %w", err)
			}

			delaySeconds, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid delay: %w", err)
			}

			msg := &v1beta1.MsgProposeTimeLockedAction{
				Proposer:     clientCtx.GetFromAddress().String(),
				ActionType:   actionType,
				Payload:      payload,
				DelaySeconds: delaySeconds,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdExecuteTimeLockedAction returns the command to execute a time-locked action
func GetCmdExecuteTimeLockedAction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-timelocked-action [action-id]",
		Short: "Execute a ready time-locked action",
		Long: `Execute a time-locked action that has passed its delay period.

Example:
  aurad tx auth execute-timelocked-action action123 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			actionID := args[0]

			msg := &v1beta1.MsgExecuteTimeLockedAction{
				Executor: clientCtx.GetFromAddress().String(),
				ActionId: actionID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCancelTimeLockedAction returns the command to cancel a time-locked action
func GetCmdCancelTimeLockedAction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-timelocked-action [action-id]",
		Short: "Cancel a pending time-locked action",
		Long: `Cancel a time-locked action before it becomes executable.

Example:
  aurad tx auth cancel-timelocked-action action123 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			actionID := args[0]

			msg := &v1beta1.MsgCancelTimeLockedAction{
				Canceller: clientCtx.GetFromAddress().String(),
				ActionId:  actionID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdActivateEmergencyAdmin returns the command to activate an emergency admin
func GetCmdActivateEmergencyAdmin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate-emergency-admin [admin-address] [privileges]",
		Short: "Activate an emergency admin with specific privileges",
		Long: `Activate an emergency admin key with specific privileges and optional expiration.
Privileges should be comma-separated.

Example:
  aurad tx auth activate-emergency-admin aura1abc... "PAUSE_SYSTEM,EMERGENCY_WITHDRAWAL" --expires-in 3600 --from mykey`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			adminAddress := args[0]
			privileges := strings.Split(args[1], ",")
			for i, priv := range privileges {
				privileges[i] = strings.TrimSpace(priv)
			}

			expiresIn, err := cmd.Flags().GetInt64("expires-in")
			if err != nil {
				return err
			}

			msg := &v1beta1.MsgActivateEmergencyAdmin{
				Activator:        clientCtx.GetFromAddress().String(),
				AdminAddress:     adminAddress,
				Privileges:       privileges,
				ExpiresInSeconds: expiresIn,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int64("expires-in", 0, "Expiration time in seconds (0 for no expiry)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDeactivateEmergencyAdmin returns the command to deactivate an emergency admin
func GetCmdDeactivateEmergencyAdmin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-emergency-admin [admin-address]",
		Short: "Deactivate an emergency admin",
		Long: `Deactivate an active emergency admin key.

Example:
  aurad tx auth deactivate-emergency-admin aura1abc... --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			adminAddress := args[0]

			msg := &v1beta1.MsgDeactivateEmergencyAdmin{
				Deactivator:  clientCtx.GetFromAddress().String(),
				AdminAddress: adminAddress,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdInitiateValidatorKeyRotation returns the command to initiate validator key rotation
func GetCmdInitiateValidatorKeyRotation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initiate-key-rotation [validator-address] [new-consensus-pubkey]",
		Short: "Initiate validator key rotation",
		Long: `Initiate the rotation of a validator's consensus public key.

Example:
  aurad tx auth initiate-key-rotation auravaloper1abc... '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"..."}' --from mykey`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validatorAddress := args[0]
			newConsensusPubkey := args[1]

			msg := &v1beta1.MsgInitiateValidatorKeyRotation{
				Initiator:          clientCtx.GetFromAddress().String(),
				ValidatorAddress:   validatorAddress,
				NewConsensusPubkey: newConsensusPubkey,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCompleteValidatorKeyRotation returns the command to complete validator key rotation
func GetCmdCompleteValidatorKeyRotation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete-key-rotation [validator-address]",
		Short: "Complete validator key rotation",
		Long: `Complete the rotation of a validator's consensus public key.

Example:
  aurad tx auth complete-key-rotation auravaloper1abc... --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validatorAddress := args[0]

			msg := &v1beta1.MsgCompleteValidatorKeyRotation{
				Completer:        clientCtx.GetFromAddress().String(),
				ValidatorAddress: validatorAddress,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdCreateSession returns the command to create a new session
func GetCmdCreateSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-session [ip-address]",
		Short: "Create a new API session",
		Long: `Create a new API session with optional metadata.

Example:
  aurad tx auth create-session 192.168.1.1 --metadata "device=mobile,app=wallet" --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ipAddress := args[0]

			metadataStr, err := cmd.Flags().GetString("metadata")
			if err != nil {
				return err
			}

			metadata := make(map[string]string)
			if metadataStr != "" {
				pairs := strings.Split(metadataStr, ",")
				for _, pair := range pairs {
					kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
					if len(kv) == 2 {
						metadata[kv[0]] = kv[1]
					}
				}
			}

			msg := &v1beta1.MsgCreateSession{
				UserAddress: clientCtx.GetFromAddress().String(),
				IpAddress:   ipAddress,
				Metadata:    metadata,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("metadata", "", "Session metadata as key=value pairs (comma-separated)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRevokeSession returns the command to revoke a session
func GetCmdRevokeSession() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-session [session-id]",
		Short: "Revoke an active session",
		Long: `Revoke an active API session.

Example:
  aurad tx auth revoke-session session123 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sessionID := args[0]

			msg := &v1beta1.MsgRevokeSession{
				UserAddress: clientCtx.GetFromAddress().String(),
				SessionId:   sessionID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// parseWalletType converts a string to WalletType enum
func parseWalletType(s string) v1beta1.WalletType {
	switch strings.ToLower(s) {
	case "3-of-5":
		return v1beta1.WalletType_WALLET_TYPE_3_OF_5
	case "5-of-7":
		return v1beta1.WalletType_WALLET_TYPE_5_OF_7
	case "custom":
		return v1beta1.WalletType_WALLET_TYPE_CUSTOM
	default:
		return v1beta1.WalletType_WALLET_TYPE_UNSPECIFIED
	}
}
