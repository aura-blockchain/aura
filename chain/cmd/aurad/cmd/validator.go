package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ValidatorCmd returns the validator operations command
func ValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator",
		Short: "Validator operations and management",
		Long: `Comprehensive validator operations for node operators.

This command provides specialized tools for validators including:
- Validator setup and initialization
- Key management and security
- Delegation management
- Commission management
- Performance monitoring
- Uptime tracking
- Slashing protection
- Emergency operations`,
	}

	cmd.AddCommand(
		validatorInfoCmd(),
		validatorSetupCmd(),
		validatorStatusCmd(),
		validatorDelegationsCmd(),
		validatorCommissionCmd(),
		validatorUptimeCmd(),
		validatorSlashingCmd(),
		validatorEditCmd(),
		validatorUnjailCmd(),
		validatorSigningInfoCmd(),
	)

	return cmd
}

// validatorInfoCmd displays validator information
func validatorInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [validator-address]",
		Short: "Display validator information",
		Long: `Display comprehensive information about a validator including:
- Operator address and consensus pubkey
- Commission rates
- Delegation amounts
- Status (bonded/unbonding/unbonded)
- Jailed status
- Signing info

Example:
  aurad validator info auravaloper1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid validator address: %w", err)
			}

			queryClient := types.NewQueryClient(clientCtx)

			// Query validator
			res, err := queryClient.Validator(cmd.Context(), &types.QueryValidatorRequest{
				ValidatorAddr: valAddr.String(),
			})
			if err != nil {
				return fmt.Errorf("failed to query validator: %w", err)
			}

			val := res.Validator

			fmt.Printf("=== Validator Information ===\n\n")
			fmt.Printf("Operator Address:    %s\n", val.OperatorAddress)
			fmt.Printf("Consensus Pubkey:    %s\n", val.ConsensusPubkey)
			fmt.Printf("Jailed:              %v\n", val.Jailed)
			fmt.Printf("Status:              %s\n", val.Status.String())
			fmt.Printf("Tokens:              %s\n", val.Tokens.String())
			fmt.Printf("Delegator Shares:    %s\n", val.DelegatorShares.String())
			fmt.Printf("\nCommission:\n")
			fmt.Printf("  Rate:              %s\n", val.Commission.CommissionRates.Rate.String())
			fmt.Printf("  Max Rate:          %s\n", val.Commission.CommissionRates.MaxRate.String())
			fmt.Printf("  Max Change Rate:   %s\n", val.Commission.CommissionRates.MaxChangeRate.String())
			fmt.Printf("  Update Time:       %s\n", val.Commission.UpdateTime.Format(time.RFC3339))
			fmt.Printf("\nDescription:\n")
			fmt.Printf("  Moniker:           %s\n", val.Description.Moniker)
			fmt.Printf("  Website:           %s\n", val.Description.Website)
			fmt.Printf("  Security Contact:  %s\n", val.Description.SecurityContact)
			fmt.Printf("  Details:           %s\n", val.Description.Details)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// validatorSetupCmd helps setup a new validator
func validatorSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Interactive validator setup wizard",
		Long: `Interactive wizard to help setup a new validator node.

This wizard will guide you through:
1. Generating validator keys
2. Creating validator configuration
3. Preparing the create-validator transaction
4. Pre-flight checks

Note: This does NOT submit the transaction. Review the output carefully
before broadcasting.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidatorSetupWizard()
		},
	}

	return cmd
}

// validatorStatusCmd displays real-time validator status
func validatorStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [validator-address]",
		Short: "Display real-time validator status",
		Long: `Display real-time status including:
- Current voting power
- Signing status
- Recent block signatures
- Uptime percentage
- Slashing risks

Example:
  aurad validator status auravaloper1abc...
  aurad validator status --self  # Query own validator`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			var valAddr string
			if len(args) > 0 {
				valAddr = args[0]
			} else {
				// Try to get from local validator key
				selfFlag, _ := cmd.Flags().GetBool("self")
				if selfFlag {
					return fmt.Errorf("--self flag requires validator address in config or keyring")
				}
				return fmt.Errorf("validator address required")
			}

			fmt.Printf("=== Validator Status ===\n\n")
			fmt.Printf("Validator:           %s\n", valAddr)
			fmt.Printf("Timestamp:           %s\n\n", time.Now().Format(time.RFC3339))

			// Get node status
			node, err := clientCtx.GetNode()
			if err != nil {
				return fmt.Errorf("failed to get node: %w", err)
			}

			status, err := node.Status(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get node status: %w", err)
			}

			fmt.Printf("Latest Block:        %d\n", status.SyncInfo.LatestBlockHeight)
			fmt.Printf("Catching Up:         %v\n", status.SyncInfo.CatchingUp)
			fmt.Printf("\n")

			// Query validator
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Validator(cmd.Context(), &types.QueryValidatorRequest{
				ValidatorAddr: valAddr,
			})
			if err != nil {
				return fmt.Errorf("failed to query validator: %w", err)
			}

			val := res.Validator
			fmt.Printf("Voting Power:        %s\n", val.Tokens.String())
			fmt.Printf("Status:              %s\n", val.Status.String())
			fmt.Printf("Jailed:              %v\n", val.Jailed)

			return nil
		},
	}

	cmd.Flags().Bool("self", false, "Query own validator (from local keys)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// validatorDelegationsCmd lists validator delegations
func validatorDelegationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [validator-address]",
		Short: "List all delegations to a validator",
		Long: `List all delegations to a validator with optional filtering and sorting.

Example:
  aurad validator delegations auravaloper1abc...
  aurad validator delegations auravaloper1abc... --min-amount 1000`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			valAddr := args[0]
			minAmount, _ := cmd.Flags().GetString("min-amount")

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.ValidatorDelegations(cmd.Context(), &types.QueryValidatorDelegationsRequest{
				ValidatorAddr: valAddr,
			})
			if err != nil {
				return fmt.Errorf("failed to query delegations: %w", err)
			}

			fmt.Printf("=== Validator Delegations ===\n\n")
			fmt.Printf("Validator:           %s\n", valAddr)
			fmt.Printf("Total Delegations:   %d\n\n", len(res.DelegationResponses))

			totalShares := math.LegacyZeroDec()
			for i, del := range res.DelegationResponses {
				// Filter by minimum amount if specified
				if minAmount != "" {
					minAmt, ok := math.NewIntFromString(minAmount)
					if !ok {
						continue
					}
					if del.Balance.Amount.LT(minAmt) {
						continue
					}
				}

				fmt.Printf("[%d] Delegator: %s\n", i+1, del.Delegation.DelegatorAddress)
				fmt.Printf("    Shares:    %s\n", del.Delegation.Shares.String())
				fmt.Printf("    Balance:   %s\n", del.Balance.String())
				totalShares = totalShares.Add(del.Delegation.Shares)
			}

			fmt.Printf("\nTotal Shares:        %s\n", totalShares.String())

			return nil
		},
	}

	cmd.Flags().String("min-amount", "", "Filter delegations by minimum amount")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// validatorCommissionCmd manages validator commission
func validatorCommissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commission [validator-address]",
		Short: "Query validator commission",
		Long: `Query validator commission including accumulated rewards.

Example:
  aurad validator commission auravaloper1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			valAddr := args[0]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Validator(cmd.Context(), &types.QueryValidatorRequest{
				ValidatorAddr: valAddr,
			})
			if err != nil {
				return fmt.Errorf("failed to query validator: %w", err)
			}

			val := res.Validator

			fmt.Printf("=== Validator Commission ===\n\n")
			fmt.Printf("Validator:           %s\n", valAddr)
			fmt.Printf("\nCommission Rates:\n")
			fmt.Printf("  Current Rate:      %s (%.2f%%)\n",
				val.Commission.CommissionRates.Rate.String(),
				val.Commission.CommissionRates.Rate.MustFloat64()*100)
			fmt.Printf("  Max Rate:          %s (%.2f%%)\n",
				val.Commission.CommissionRates.MaxRate.String(),
				val.Commission.CommissionRates.MaxRate.MustFloat64()*100)
			fmt.Printf("  Max Change Rate:   %s (%.2f%%)\n",
				val.Commission.CommissionRates.MaxChangeRate.String(),
				val.Commission.CommissionRates.MaxChangeRate.MustFloat64()*100)
			fmt.Printf("  Last Updated:      %s\n", val.Commission.UpdateTime.Format(time.RFC3339))

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// validatorUptimeCmd shows validator uptime statistics
func validatorUptimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uptime [validator-address]",
		Short: "Show validator uptime statistics",
		Long: `Show validator uptime statistics including:
- Signing percentage
- Missed blocks
- Recent signing history

Example:
  aurad validator uptime auravaloper1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr := args[0]

			fmt.Printf("=== Validator Uptime ===\n\n")
			fmt.Printf("Validator:           %s\n", valAddr)
			fmt.Printf("\nUptime tracking requires slashing module queries\n")
			fmt.Printf("Implementation pending proto message definitions\n")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// validatorSlashingCmd shows slashing information
func validatorSlashingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slashing [validator-address]",
		Short: "Show validator slashing information",
		Long: `Show validator slashing information including:
- Slashing events
- Jail status
- Missed blocks
- Downtime risk

Example:
  aurad validator slashing auravaloper1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr := args[0]

			fmt.Printf("=== Validator Slashing Info ===\n\n")
			fmt.Printf("Validator:           %s\n", valAddr)
			fmt.Printf("\nSlashing information requires slashing module queries\n")
			fmt.Printf("Implementation pending proto message definitions\n")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// validatorEditCmd creates an edit validator transaction
func validatorEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit validator description",
		Long: `Create a transaction to edit validator description.

You can update:
- Moniker (name)
- Website
- Security contact
- Details

Example:
  aurad validator edit --from mykey --moniker "My Validator" --website "https://example.com"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			valAddr := clientCtx.GetFromAddress()

			moniker, _ := cmd.Flags().GetString("moniker")
			website, _ := cmd.Flags().GetString("website")
			identity, _ := cmd.Flags().GetString("identity")
			securityContact, _ := cmd.Flags().GetString("security-contact")
			details, _ := cmd.Flags().GetString("details")

			description := types.Description{
				Moniker:         moniker,
				Website:         website,
				Identity:        identity,
				SecurityContact: securityContact,
				Details:         details,
			}

			msg := types.NewMsgEditValidator(
				sdk.ValAddress(valAddr).String(),
				description,
				nil, // commission rate
				nil, // min self delegation
			)

			// ValidateBasic check is handled by msg server in Cosmos SDK 0.53+
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("moniker", "", "Validator name")
	cmd.Flags().String("website", "", "Validator website")
	cmd.Flags().String("identity", "", "Validator identity (keybase)")
	cmd.Flags().String("security-contact", "", "Security contact email")
	cmd.Flags().String("details", "", "Validator details")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// validatorUnjailCmd creates an unjail transaction
func validatorUnjailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail",
		Short: "Unjail a validator",
		Long: `Create a transaction to unjail a validator after being jailed for downtime.

Requirements:
- Validator must be jailed
- Must wait for the minimum jail period
- Must have sufficient signing percentage

Example:
  aurad validator unjail --from mykey`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Unjail command requires slashing module integration\n")
			fmt.Printf("Implementation pending proto message definitions\n")
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// validatorSigningInfoCmd shows signing information
func validatorSigningInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-info [validator-consensus-address]",
		Short: "Show validator signing information",
		Long: `Show validator signing information from the slashing module.

Example:
  aurad validator signing-info auravalcons1abc...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			consAddr := args[0]

			fmt.Printf("=== Validator Signing Info ===\n\n")
			fmt.Printf("Consensus Address:   %s\n", consAddr)
			fmt.Printf("\nSigning info requires slashing module queries\n")
			fmt.Printf("Implementation pending proto message definitions\n")

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// runValidatorSetupWizard runs the interactive validator setup wizard
func runValidatorSetupWizard() error {
	fmt.Printf("===========================================\n")
	fmt.Printf("  Validator Setup Wizard\n")
	fmt.Printf("===========================================\n\n")

	fmt.Printf("This wizard will help you setup a validator node.\n\n")

	// Check for validator key
	validatorKeyFile := filepath.Join(homeDir, "config", "priv_validator_key.json")

	fmt.Printf("Step 1: Validator Key\n")
	fmt.Printf("---------------------\n")
	if fileExists(validatorKeyFile) {
		fmt.Printf("✓ Validator key exists: %s\n", validatorKeyFile)

		// Read and display pubkey
		data, err := os.ReadFile(validatorKeyFile)
		if err != nil {
			return fmt.Errorf("failed to read validator key: %w", err)
		}

		var keyData map[string]interface{}
		if err := json.Unmarshal(data, &keyData); err != nil {
			return fmt.Errorf("failed to parse validator key: %w", err)
		}

		if pubkey, ok := keyData["pub_key"].(map[string]interface{}); ok {
			if value, ok := pubkey["value"].(string); ok {
				fmt.Printf("  Public Key: %s\n", value)
			}
		}
	} else {
		fmt.Printf("✗ No validator key found\n")
		fmt.Printf("  Run: aurad init <moniker>\n")
		return fmt.Errorf("validator key required")
	}

	fmt.Printf("\nStep 2: Node Key\n")
	fmt.Printf("---------------------\n")
	nodeKeyFile := filepath.Join(homeDir, "config", "node_key.json")
	if fileExists(nodeKeyFile) {
		fmt.Printf("✓ Node key exists: %s\n", nodeKeyFile)
	} else {
		fmt.Printf("✗ No node key found\n")
		return fmt.Errorf("node key required")
	}

	fmt.Printf("\nStep 3: Operator Key\n")
	fmt.Printf("---------------------\n")
	fmt.Printf("You need a key in the keyring to be the validator operator.\n")
	fmt.Printf("Run: aurad keys list\n")
	fmt.Printf("Or create new: aurad keys add <name>\n")

	fmt.Printf("\nStep 4: Create Validator Transaction\n")
	fmt.Printf("-------------------------------------\n")
	fmt.Printf("Once you have:\n")
	fmt.Printf("  1. Synced your node\n")
	fmt.Printf("  2. Funded your operator account\n")
	fmt.Printf("  3. Chosen your commission rates\n\n")
	fmt.Printf("Create your validator with:\n")
	fmt.Printf("  aurad tx staking create-validator \\\n")
	fmt.Printf("    --amount=1000000uaura \\\n")
	fmt.Printf("    --pubkey=$(aurad cometbft show-validator) \\\n")
	fmt.Printf("    --moniker=\"My Validator\" \\\n")
	fmt.Printf("    --commission-rate=\"0.10\" \\\n")
	fmt.Printf("    --commission-max-rate=\"0.20\" \\\n")
	fmt.Printf("    --commission-max-change-rate=\"0.01\" \\\n")
	fmt.Printf("    --min-self-delegation=\"1\" \\\n")
	fmt.Printf("    --from=<operator-key>\n")

	fmt.Printf("\n===========================================\n")
	fmt.Printf("  Setup wizard complete!\n")
	fmt.Printf("===========================================\n")

	return nil
}
