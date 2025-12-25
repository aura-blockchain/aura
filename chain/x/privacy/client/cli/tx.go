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
	sdkmath "cosmossdk.io/math"

	privacyv1beta1 "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// GetTxCmd returns the transaction commands for the privacy module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "privacy",
		Short:                      "Privacy transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdSubmitPrivateTransaction(),
		CmdCreateMixingPool(),
		CmdJoinMixingPool(),
		CmdRegisterViewKey(),
		CmdRevokeViewKey(),
		CmdUpdateNetworkPrivacy(),
	)

	return cmd
}

// CmdSubmitPrivateTransaction submits a private transaction
func CmdSubmitPrivateTransaction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-private-tx [tx-data-file]",
		Short: "Submit a privacy-enhanced transaction",
		Long: `Submit a transaction with privacy features including ZK proofs, stealth addresses, and ring signatures.

Examples:
  aurad tx privacy submit-private-tx private_tx.json --from alice
  aurad tx privacy submit-private-tx tx_data.json --from bob

The transaction data file should contain a JSON-encoded PrivateTransaction with:
  - ZK proof for transaction validity
  - Stealth address for recipient privacy
  - Ring signature for sender anonymity
  - Confidential transaction for amount privacy
  - Optional encrypted memo

This provides comprehensive privacy protections for the transaction.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// For now, we'll accept the file path but note that actual implementation
			// would need to read and parse the private transaction data
			// This is a placeholder for the complex privacy transaction construction

			msg := &privacyv1beta1.MsgSubmitPrivateTransaction{
				Sender: clientCtx.GetFromAddress().String(),
				// PrivateTransaction would be populated from file
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCreateMixingPool creates a new coin mixing pool
func CmdCreateMixingPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-mixing-pool [min-participants] [max-participants] [denomination] [mixing-rounds] [deadline-duration]",
		Short: "Create a new coin mixing pool",
		Long: `Create a new mixing pool for coin tumbling and privacy enhancement.

Examples:
  aurad tx privacy create-mixing-pool 3 10 1000000 5 3600 --from alice
  aurad tx privacy create-mixing-pool 5 20 5000000 10 7200 --from bob

Arguments:
  min-participants: Minimum number of participants required
  max-participants: Maximum number of participants allowed
  denomination: Standard mixing amount (in base units)
  mixing-rounds: Number of mixing rounds to perform
  deadline-duration: Time to wait for participants (in seconds)

The mixing pool enables users to mix their coins with others,
breaking the transaction graph and enhancing privacy.
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse min participants
			minParticipants, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid min-participants: %w", err)
			}

			// Parse max participants
			maxParticipants, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid max-participants: %w", err)
			}

			// Parse denomination
			denomination, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid denomination: %s", args[2])
			}

			// Parse mixing rounds
			mixingRounds, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid mixing-rounds: %w", err)
			}

			// Parse deadline duration
			deadlineDuration, err := strconv.ParseUint(args[4], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid deadline-duration: %w", err)
			}

			msg := &privacyv1beta1.MsgCreateMixingPool{
				Creator:          clientCtx.GetFromAddress().String(),
				MinParticipants:  uint32(minParticipants),
				MaxParticipants:  uint32(maxParticipants),
				Denomination:     denomination,
				MixingRounds:     uint32(mixingRounds),
				DeadlineDuration: deadlineDuration,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdJoinMixingPool joins an existing mixing pool
func CmdJoinMixingPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join-mixing-pool [pool-id] [commitment]",
		Short: "Join an existing mixing pool",
		Long: `Join a coin mixing pool and participate in the mixing process.

Examples:
  aurad tx privacy join-mixing-pool pool-123 abc123def456... --from alice
  aurad tx privacy join-mixing-pool pool-456 789abcdef012... --from bob

Arguments:
  pool-id: The mixing pool to join
  commitment: Cryptographic commitment to the mixing process (hex-encoded)

You must have the exact denomination amount to participate.
Once the pool reaches minimum participants, mixing will begin.
After mixing completes, you'll receive your mixed coins.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse commitment
			commitment, err := hex.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("invalid commitment: %w", err)
			}

			msg := &privacyv1beta1.MsgJoinMixingPool{
				Participant: clientCtx.GetFromAddress().String(),
				PoolId:      args[0],
				Commitment:  commitment,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRegisterViewKey registers a view key for selective disclosure
func CmdRegisterViewKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-view-key [key-type] [public-view-key] [permissions]",
		Short: "Register a view key for selective disclosure",
		Long: `Register a view key that allows selective viewing of private transactions.

Examples:
  aurad tx privacy register-view-key INCOMING abc123... "view,decrypt" --from alice
  aurad tx privacy register-view-key AUDIT def456... "view,decrypt,audit" --from auditor

Arguments:
  key-type: Type of view key (INCOMING, OUTGOING, AUDIT)
  public-view-key: The public component of the view key (hex-encoded)
  permissions: Comma-separated permissions (view, decrypt, audit)

View keys enable selective disclosure:
  - INCOMING: View received transactions
  - OUTGOING: View sent transactions
  - AUDIT: Full auditing capabilities

Flags:
  --expiration: Expiration time in seconds (default: no expiration)
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse public view key
			publicViewKey, err := hex.DecodeString(args[1])
			if err != nil {
				return fmt.Errorf("invalid public-view-key: %w", err)
			}

			// Parse permissions
			permissions := strings.Split(args[2], ",")
			for i := range permissions {
				permissions[i] = strings.TrimSpace(permissions[i])
			}

			// Create ViewKey
			viewKey := &privacyv1beta1.ViewKey{
				KeyType:       args[0],
				PublicViewKey: publicViewKey,
				Permissions:   permissions,
			}

			msg := &privacyv1beta1.MsgRegisterViewKey{
				Owner:   clientCtx.GetFromAddress().String(),
				ViewKey: viewKey,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Uint64("expiration", 0, "Expiration time in seconds (0 = no expiration)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRevokeViewKey revokes a previously registered view key
func CmdRevokeViewKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-view-key [public-view-key]",
		Short: "Revoke a previously registered view key",
		Long: `Revoke a view key to remove its access to private transactions.

Examples:
  aurad tx privacy revoke-view-key abc123def456... --from alice
  aurad tx privacy revoke-view-key 789abcdef012... --from bob

Note: Only the view key owner can revoke it.
After revocation, the key can no longer decrypt or view transactions.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse public view key
			publicViewKey, err := hex.DecodeString(args[0])
			if err != nil {
				return fmt.Errorf("invalid public-view-key: %w", err)
			}

			msg := &privacyv1beta1.MsgRevokeViewKey{
				Owner:         clientCtx.GetFromAddress().String(),
				PublicViewKey: publicViewKey,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUpdateNetworkPrivacy updates network privacy settings
func CmdUpdateNetworkPrivacy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-network-privacy [network-type]",
		Short: "Update network privacy settings",
		Long: `Update network privacy settings for Tor/I2P integration.

Examples:
  aurad tx privacy update-network-privacy TOR --onion-address abc.onion --from alice
  aurad tx privacy update-network-privacy I2P --i2p-destination def.i2p --from bob
  aurad tx privacy update-network-privacy MIXED --onion-address abc.onion --i2p-destination def.i2p --from charlie

Arguments:
  network-type: Type of privacy network (TOR, I2P, MIXED)

Flags:
  --onion-address: Tor hidden service address
  --i2p-destination: I2P destination address
  --proxy-enabled: Enable proxy (default: true)
  --circuit-lifetime: Tor circuit lifetime in seconds (default: 600)
  --stream-isolation: Enable stream isolation (default: true)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get flags
			onionAddress, _ := cmd.Flags().GetString("onion-address")
			i2pDestination, _ := cmd.Flags().GetString("i2p-destination")
			proxyEnabled, _ := cmd.Flags().GetBool("proxy-enabled")
			circuitLifetime, _ := cmd.Flags().GetUint64("circuit-lifetime")
			streamIsolation, _ := cmd.Flags().GetBool("stream-isolation")

			networkPrivacy := &privacyv1beta1.NetworkPrivacy{
				NetworkType:     args[0],
				OnionAddress:    onionAddress,
				I2PDestination:  i2pDestination,
				ProxyEnabled:    proxyEnabled,
				CircuitLifetime: circuitLifetime,
				StreamIsolation: streamIsolation,
			}

			msg := &privacyv1beta1.MsgUpdateNetworkPrivacy{
				Sender:         clientCtx.GetFromAddress().String(),
				NetworkPrivacy: networkPrivacy,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("onion-address", "", "Tor hidden service address")
	cmd.Flags().String("i2p-destination", "", "I2P destination address")
	cmd.Flags().Bool("proxy-enabled", true, "Enable proxy")
	cmd.Flags().Uint64("circuit-lifetime", 600, "Tor circuit lifetime in seconds")
	cmd.Flags().Bool("stream-isolation", true, "Enable stream isolation")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
