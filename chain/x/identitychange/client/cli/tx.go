// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	identitychangev1beta1 "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

// GetTxCmd returns the transaction commands for the identitychange module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "identitychange",
		Aliases:                    []string{"identity", "idchange"},
		Short:                      "Identity change transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdRequestIdentityChange(),
		CmdSubmitAssistantProof(),
		CmdApplyIdentityChange(),
		CmdRejectIdentityChange(),
		CmdSuspendIdentityChanges(),
	)

	return cmd
}

// CmdRequestIdentityChange creates a new identity change request
func CmdRequestIdentityChange() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request [target-did] [metadata-hash] [ir-id] [proof-hash]",
		Short: "Request an identity change for a DID",
		Long: `Create a new identity change request for a target DID.

Examples:
  aurad tx identitychange request did:aura:abc123 metadata-hash-456 ir-789 proof-xyz --from alice
  aurad tx identitychange request did:aura:def456 hash789 ir-012 proof-abc --from bob

Arguments:
  target-did: The DID to be changed
  metadata-hash: Hash of the change metadata
  ir-id: Associated Inclusion Routine ID
  proof-hash: Hash of the proof data

This initiates an identity change request that requires verification by an assistant.
The request will be in PENDING_VERIFICATION status until processed.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &identitychangev1beta1.MsgRequestIdentityChange{
				Requester:    clientCtx.GetFromAddress().String(),
				TargetDid:    args[0],
				MetadataHash: args[1],
				IrId:         args[2],
				ProofHash:    args[3],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSubmitAssistantProof submits verification proof from an assistant
func CmdSubmitAssistantProof() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proof [request-id] [proof-hash] [confidence-delta] [success]",
		Short: "Submit assistant verification proof for an identity change request",
		Long: `Submit proof and verification result for an identity change request (assistant only).

Examples:
  aurad tx identitychange submit-proof req-123 proof-xyz 10 true --from assistant
  aurad tx identitychange submit-proof req-456 proof-abc -5 false --from assistant

Arguments:
  request-id: The identity change request ID
  proof-hash: Hash of the verification proof
  confidence-delta: Change in confidence score (can be negative)
  success: Whether verification was successful (true/false)

This command is restricted to authorized assistants.
Successful verification moves the request to READY_TO_APPLY status.
Failed verification moves it to REJECTED status.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse confidence delta
			confidenceDelta, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid confidence-delta: %w", err)
			}

			// Parse success
			success, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid success value: %w", err)
			}

			msg := &identitychangev1beta1.MsgSubmitAssistantProof{
				Assistant:       clientCtx.GetFromAddress().String(),
				RequestId:       args[0],
				ProofHash:       args[1],
				ConfidenceDelta: confidenceDelta,
				Success:         success,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdApplyIdentityChange applies an approved identity change
func CmdApplyIdentityChange() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [request-id]",
		Short: "Apply an approved identity change",
		Long: `Apply an identity change that has been verified and is ready to apply.

Examples:
  aurad tx identitychange apply req-123 --from alice
  aurad tx identitychange apply req-456 --from bob

Note: Only the original requester can apply the change.
The request must be in READY_TO_APPLY status.
This updates the identity record with the new confidence score and metadata.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &identitychangev1beta1.MsgApplyIdentityChange{
				Requester: clientCtx.GetFromAddress().String(),
				RequestId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRejectIdentityChange rejects an identity change request
func CmdRejectIdentityChange() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject [request-id] [reason]",
		Short: "Reject an identity change request",
		Long: `Reject an identity change request with a reason.

Examples:
  aurad tx identitychange reject req-123 "Insufficient proof" --from actor
  aurad tx identitychange reject req-456 "Policy violation" --from governance

Note: Can be called by the requester, assistant, or governance.
The request will be moved to REJECTED status.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &identitychangev1beta1.MsgRejectIdentityChange{
				Actor:     clientCtx.GetFromAddress().String(),
				RequestId: args[0],
				Reason:    args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSuspendIdentityChanges suspends all identity changes (governance only)
func CmdSuspendIdentityChanges() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend [reason]",
		Short: "Suspend all identity changes (governance only)",
		Long: `Suspend the entire identity change module for emergency situations.

Examples:
  aurad tx identitychange suspend "Security issue detected" --from authority
  aurad tx identitychange suspend "System maintenance" --from governance

Note: This command is restricted to governance/authority.
When suspended, no new requests can be created or processed.
Existing requests remain in their current state.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &identitychangev1beta1.MsgSuspendIdentityChanges{
				Authority: clientCtx.GetFromAddress().String(),
				Reason:    args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
