package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	compliancecli "github.com/aequitas/aura/chain/x/compliance/client/cli"
	confidencescorecli "github.com/aequitas/aura/chain/x/confidencescore/client/cli"
	wasmcli "github.com/aequitas/aura/chain/x/wasm/client/cli"
)

// TxCmd returns the transaction command for the Aura daemon
func TxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Create and sign transactions",
		Long: `Create, sign, and broadcast transactions to the blockchain.

Available subcommands allow interacting with different modules.`,
	}

	cmd.AddCommand(
		txBroadcastCmd(),
		txSignCmd(),
		txIdentityChangeCmd(),
		txInclusionRoutinesCmd(),
		txConfidenceScoreCmd(),
		txComplianceCmd(),
		txVCRegistryCmd(),
		txDataRegistryCmd(),
		txGovernanceCmd(),
		txWasmCmd(),
	)

	return cmd
}

// txBroadcastCmd returns the broadcast transaction command
func txBroadcastCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "broadcast [file]",
		Short: "Broadcast a signed transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			fmt.Printf("Broadcasting transaction from file: %s\n", file)
			fmt.Printf("This feature requires full Cosmos SDK integration.\n")
			return nil
		},
	}
}

// txSignCmd returns the sign transaction command
func txSignCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sign [file]",
		Short: "Sign a transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			fmt.Printf("Signing transaction from file: %s\n", file)
			fmt.Printf("This feature requires full Cosmos SDK integration.\n")
			return nil
		},
	}
}

// txIdentityChangeCmd returns the identity change transaction command
func txIdentityChangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identitychange",
		Short: "Identity change module transactions",
		Long:  "Create and broadcast identity change transactions.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "register [did]",
			Short: "Register a new decentralized identifier",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				did := args[0]
				fmt.Printf("Registering DID: %s\n", did)
				fmt.Printf("This will create and broadcast a RegisterDID transaction.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "update [did]",
			Short: "Update a decentralized identifier",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				did := args[0]
				fmt.Printf("Updating DID: %s\n", did)
				fmt.Printf("This will create and broadcast an UpdateDID transaction.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "deactivate [did]",
			Short: "Deactivate a decentralized identifier",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				did := args[0]
				fmt.Printf("Deactivating DID: %s\n", did)
				fmt.Printf("This will create and broadcast a DeactivateDID transaction.\n")
				return nil
			},
		},
	)

	return cmd
}

// txInclusionRoutinesCmd returns the inclusion routines transaction command
func txInclusionRoutinesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inclusionroutines",
		Short: "Inclusion routines module transactions",
		Long:  "Create and broadcast inclusion routines transactions.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "propose [routine-json]",
			Short: "Propose a new inclusion routine",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				routineJSON := args[0]
				fmt.Printf("Proposing inclusion routine: %s\n", routineJSON)
				fmt.Printf("This will create and broadcast a ProposeRoutine transaction.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "vote [routine-id] [vote]",
			Short: "Vote on an inclusion routine proposal",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				routineID := args[0]
				vote := args[1]
				fmt.Printf("Voting on routine %s: %s\n", routineID, vote)
				fmt.Printf("This will create and broadcast a VoteRoutine transaction.\n")
				return nil
			},
		},
	)

	return cmd
}

// txConfidenceScoreCmd returns the confidence score transaction command
func txConfidenceScoreCmd() *cobra.Command {
	// Use the comprehensive CLI commands from the confidencescore module
	return confidencescorecli.GetTxCmd()
}

func txComplianceCmd() *cobra.Command {
	return compliancecli.GetTxCmd()
}

// txVCRegistryCmd returns the VC registry transaction command
func txVCRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vcregistry",
		Short: "VC registry module transactions",
		Long:  "Create and broadcast verifiable credential transactions.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "issue [credential-json]",
			Short: "Issue a new verifiable credential",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				credentialJSON := args[0]
				fmt.Printf("Issuing credential: %s\n", credentialJSON)
				fmt.Printf("This will create and broadcast an IssueCredential transaction.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "revoke [credential-id]",
			Short: "Revoke a verifiable credential",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				credentialID := args[0]
				fmt.Printf("Revoking credential: %s\n", credentialID)
				fmt.Printf("This will create and broadcast a RevokeCredential transaction.\n")
				return nil
			},
		},
	)

	return cmd
}

// txDataRegistryCmd returns the data registry transaction command
func txDataRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataregistry",
		Short: "Data registry module transactions",
		Long:  "Create and broadcast data registry transactions.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "register [data-json]",
			Short: "Register new data",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				dataJSON := args[0]
				fmt.Printf("Registering data: %s\n", dataJSON)
				fmt.Printf("This will create and broadcast a RegisterData transaction.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "update [data-id] [data-json]",
			Short: "Update registered data",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				dataID := args[0]
				dataJSON := args[1]
				fmt.Printf("Updating data %s: %s\n", dataID, dataJSON)
				fmt.Printf("This will create and broadcast an UpdateData transaction.\n")
				return nil
			},
		},
	)

	return cmd
}

// txGovernanceCmd returns the governance transaction command
func txGovernanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "governance",
		Short: "Governance module transactions",
		Long:  "Create and broadcast governance transactions.",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "submit-proposal [proposal-json]",
			Short: "Submit a new governance proposal",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				proposalJSON := args[0]
				fmt.Printf("Submitting proposal: %s\n", proposalJSON)
				fmt.Printf("This will create and broadcast a SubmitProposal transaction.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "vote [proposal-id] [vote]",
			Short: "Vote on a governance proposal",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				proposalID := args[0]
				vote := args[1]
				fmt.Printf("Voting on proposal %s: %s\n", proposalID, vote)
				fmt.Printf("This will create and broadcast a Vote transaction.\n")
				return nil
			},
		},
		&cobra.Command{
			Use:   "deposit [proposal-id] [amount]",
			Short: "Deposit tokens to a proposal",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				proposalID := args[0]
				amount := args[1]
				fmt.Printf("Depositing %s to proposal %s\n", amount, proposalID)
				fmt.Printf("This will create and broadcast a Deposit transaction.\n")
				return nil
			},
		},
	)

	return cmd
}

// txWasmCmd returns the wasm transaction command
func txWasmCmd() *cobra.Command {
	return wasmcli.GetTxCmd()
}
