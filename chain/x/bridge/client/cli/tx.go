package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "bridge",
		Short:                      "Bridge transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdLinkAddress(),
		CmdLockTokens(),
		CmdUnlockTokens(),
		CmdMintTokens(),
		CmdBurnTokens(),
		CmdCrossChainSwap(),
		CmdRelayTransfer(),
	)

	return cmd
}

// CmdLinkAddress links AURA/PAW/XAI addresses for shared identity
func CmdLinkAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-address [aura-address] [paw-address] [xai-address]",
		Short: "Link AURA, PAW, and XAI addresses for shared identity verification",
		Long: `Link addresses across AURA, PAW, and XAI chains to enable cross-chain identity verification.

Examples:
  aurad tx bridge link-address aura1abc... paw1def... xai1ghi... --from alice
  aurad tx bridge link-address aura1abc... paw1def... "" --from alice (link only PAW)
  aurad tx bridge link-address aura1abc... "" xai1ghi... --from alice (link only XAI)

This command links your addresses across chains, allowing:
  - Shared verification status across AURA, PAW, and XAI
  - Cross-chain reputation and IR score propagation
  - Seamless cross-chain transfers between linked addresses
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auraAddress := args[0]
			pawAddress := args[1]
			xaiAddress := args[2]

			// Get optional signatures
			pawSigHex, _ := cmd.Flags().GetString("paw-signature")
			xaiSigHex, _ := cmd.Flags().GetString("xai-signature")

			var pawSig, xaiSig []byte
			if pawSigHex != "" {
				pawSig, err = hex.DecodeString(pawSigHex)
				if err != nil {
					return fmt.Errorf("invalid paw signature: %w", err)
				}
			}
			if xaiSigHex != "" {
				xaiSig, err = hex.DecodeString(xaiSigHex)
				if err != nil {
					return fmt.Errorf("invalid xai signature: %w", err)
				}
			}

			msg := &types.MsgLinkAddress{
				AuraAddress:  auraAddress,
				PawAddress:   pawAddress,
				XaiAddress:   xaiAddress,
				PawSignature: pawSig,
				XaiSignature: xaiSig,
				Signer:       clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("paw-signature", "", "Signature from PAW address (hex-encoded)")
	cmd.Flags().String("xai-signature", "", "Signature from XAI address (hex-encoded)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdLockTokens initiates a cross-chain transfer by locking tokens
func CmdLockTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-tokens [target-chain] [recipient] [amount]",
		Short: "Lock tokens on AURA to transfer to PAW or XAI",
		Long: `Lock tokens on AURA to initiate a cross-chain transfer to PAW or XAI.

Examples:
  aurad tx bridge lock-tokens paw paw1def... 1000uaura --from alice
  aurad tx bridge lock-tokens xai xai1ghi... 5000uaura --from alice

Supported target chains: paw, xai

The tokens will be locked on AURA and wrapped tokens will be minted on the target chain.
Transfer ID will be returned for tracking.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			targetChain := args[0]
			recipient := args[1]
			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			msg := &types.MsgLockTokens{
				Sender:      clientCtx.GetFromAddress().String(),
				TargetChain: targetChain,
				Recipient:   recipient,
				Amount:      amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdUnlockTokens unlocks tokens after burn proof from target chain
func CmdUnlockTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock-tokens [source-chain] [burn-tx-hash] [amount] [denom]",
		Short: "Unlock tokens on AURA after burn on target chain",
		Long: `Unlock tokens on AURA after providing burn proof from PAW or XAI.

Examples:
  aurad tx bridge unlock-tokens paw 0xabc123... 1000 uaura --from alice --validator-signatures sig1,sig2,sig3
  aurad tx bridge unlock-tokens xai 0xdef456... 5000 uaura --from alice --validator-signatures sig1,sig2

Requires validator signatures to prove tokens were burned on the source chain.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sourceChain := args[0]
			burnTxHash := args[1]
			amount, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[2])
			}
			denom := args[3]

			// Parse validator signatures
			signaturesHex, err := cmd.Flags().GetStringSlice("validator-signatures")
			if err != nil {
				return err
			}

			var validatorSignatures [][]byte
			for _, sigHex := range signaturesHex {
				sig, err := hex.DecodeString(sigHex)
				if err != nil {
					return fmt.Errorf("invalid signature: %w", err)
				}
				validatorSignatures = append(validatorSignatures, sig)
			}

			msg := &types.MsgUnlockTokens{
				Sender:              clientCtx.GetFromAddress().String(),
				SourceChain:         sourceChain,
				BurnTxHash:          burnTxHash,
				Amount:              amount,
				Denom:               denom,
				ValidatorSignatures: validatorSignatures,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringSlice("validator-signatures", []string{}, "Validator signatures (hex-encoded, comma-separated)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdMintTokens mints wrapped tokens on AURA (validator-only)
func CmdMintTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint-tokens [source-chain] [source-tx-hash] [recipient] [amount] [denom]",
		Short: "Mint wrapped tokens on AURA from PAW or XAI (validator-only)",
		Long: `Mint wrapped tokens on AURA after lock proof from PAW or XAI.

Examples:
  aurad tx bridge mint-tokens paw 0xabc123... aura1def... 1000 paw.token --from validator --validator-signature sig1
  aurad tx bridge mint-tokens xai 0xdef456... aura1ghi... 5000 xai.coin --from validator --validator-signature sig2

This command is restricted to authorized validators.
Requires validator signature over (source_chain, source_tx_hash, recipient, amount, denom).
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sourceChain := args[0]
			sourceTxHash := args[1]
			recipient := args[2]
			amount, ok := sdkmath.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[3])
			}
			denom := args[4]

			// Get validator signature
			sigHex, err := cmd.Flags().GetString("validator-signature")
			if err != nil {
				return err
			}
			validatorSig, err := hex.DecodeString(sigHex)
			if err != nil {
				return fmt.Errorf("invalid validator signature: %w", err)
			}

			msg := &types.MsgMintTokens{
				Validator:          clientCtx.GetFromAddress().String(),
				SourceChain:        sourceChain,
				SourceTxHash:       sourceTxHash,
				Recipient:          recipient,
				Amount:             amount,
				Denom:              denom,
				ValidatorSignature: validatorSig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("validator-signature", "", "Validator signature (hex-encoded)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdBurnTokens burns wrapped tokens to unlock on source chain
func CmdBurnTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-tokens [target-chain] [recipient] [amount]",
		Short: "Burn wrapped tokens on AURA to unlock on source chain",
		Long: `Burn wrapped tokens on AURA to unlock original tokens on PAW or XAI.

Examples:
  aurad tx bridge burn-tokens paw paw1def... 1000paw.token --from alice
  aurad tx bridge burn-tokens xai xai1ghi... 5000xai.coin --from alice

The wrapped tokens will be burned on AURA and original tokens will be unlocked on the target chain.
Transfer ID will be returned for tracking.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			targetChain := args[0]
			recipient := args[1]
			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			msg := &types.MsgBurnTokens{
				Sender:      clientCtx.GetFromAddress().String(),
				TargetChain: targetChain,
				Recipient:   recipient,
				Amount:      amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCrossChainSwap initiates a cross-chain swap
func CmdCrossChainSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cross-chain-swap [source-chain] [input-amount] [target-chain] [target-denom] [min-target-amount]",
		Short: "Initiate a cross-chain swap between AURA, PAW, and XAI",
		Long: `Swap tokens across chains using the bridge and DEXs.

Examples:
  aurad tx bridge cross-chain-swap aura 1000uaura paw paw.token 950 --from alice --max-slippage 500
  aurad tx bridge cross-chain-swap paw 1000paw.token xai xai.coin 900 --from alice --recipient xai1ghi... --max-slippage 1000

Options:
  --max-slippage: Maximum slippage in basis points (500 = 5%, 1000 = 10%)
  --recipient: Recipient address on target chain (defaults to sender)

The swap may route through multiple chains (e.g., AURA -> Osmosis -> PAW).
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sourceChain := args[0]
			inputCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid input amount: %w", err)
			}
			targetChain := args[2]
			targetDenom := args[3]
			minTargetAmount, ok := sdkmath.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid min target amount: %s", args[4])
			}

			// Get optional flags
			recipient, _ := cmd.Flags().GetString("recipient")
			if recipient == "" {
				recipient = clientCtx.GetFromAddress().String()
			}
			maxSlippageBps, _ := cmd.Flags().GetUint64("max-slippage")

			msg := &types.MsgCrossChainSwap{
				Sender:          clientCtx.GetFromAddress().String(),
				SourceChain:     sourceChain,
				InputCoin:       inputCoin,
				TargetChain:     targetChain,
				TargetDenom:     targetDenom,
				MinTargetAmount: minTargetAmount,
				Recipient:       recipient,
				MaxSlippageBps:  maxSlippageBps,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("recipient", "", "Recipient address on target chain (defaults to sender)")
	cmd.Flags().Uint64("max-slippage", 500, "Maximum slippage in basis points (500 = 5%)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRelayTransfer relays a cross-chain transfer (relayer-only)
func CmdRelayTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay-transfer [transfer-id] [target-tx-hash] [status]",
		Short: "Relay a cross-chain transfer (relayer-only)",
		Long: `Update the status of a cross-chain transfer after relaying to target chain.

Examples:
  aurad tx bridge relay-transfer transfer-123 0xabc123... COMPLETED --from relayer
  aurad tx bridge relay-transfer transfer-456 0xdef456... RELAYED --from relayer

This command is restricted to authorized relayers.
Status options: PENDING, CONFIRMED, RELAYED, COMPLETED, FAILED, REFUNDED
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			transferID := args[0]
			targetTxHash := args[1]
			status := args[2]

			msg := &types.MsgRelayTransfer{
				Relayer:      clientCtx.GetFromAddress().String(),
				TransferId:   transferID,
				TargetTxHash: targetTxHash,
				Status:       status,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
