// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// HelpCmd creates a new enhanced help command
func HelpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [module]",
		Short: "Get detailed help for a module or command",
		Long: `Get detailed help information for AURA blockchain commands and modules.

Without arguments, displays general help.
With a module name, displays detailed help for that module.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return showGeneralHelp()
			}

			moduleName := GetFullCommandName(args[0])
			PrintModuleHelp(moduleName)
			return nil
		},
	}

	return cmd
}

func showGeneralHelp() error {
	fmt.Print(`
AURA Blockchain - Command Line Interface

USAGE:
  aurad [command] [subcommand] [flags]

CORE COMMANDS:
  init                 Initialize node configuration
  start                Start the blockchain node
  version              Display version information
  status               Show node status
  keys                 Manage cryptographic keys
  query                Query blockchain state
  tx                   Create and sign transactions

INTERACTIVE MODES:
  interactive          Start interactive mode with guided wizards
  batch                Execute batch commands from file
  script               Execute script file

CONFIGURATION:
  config               Manage node configuration
  completion           Generate shell completion scripts

MAIN MODULES:
  vcregistry (vc)      Verifiable Credentials and DIDs
  bridge (br)          Cross-chain operations (PAW/XAI)
  inclusionroutines (ir) Verification routines
  governance (gov)     Proposal and voting system
  dex (swap)           Decentralized exchange
  cryptography (crypto) Cryptographic operations
  compliance (kyc)     KYC and compliance
  confidencescore (cs) Confidence scoring system

GETTING STARTED:
  # Quick start guide
  aurad help quickstart

  # Module-specific help
  aurad help vcregistry
  aurad help bridge
  aurad help inclusionroutines

  # Interactive mode (recommended for beginners)
  aurad interactive

  # Enable auto-completion
  source <(aurad completion bash)

EXAMPLES:
  # Mint a verifiable credential
  aurad tx vcregistry mint-vc did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from alice

  # Using alias
  aurad tx vc mint did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from alice

  # Link cross-chain addresses
  aurad tx bridge link-address aura1... paw1... xai1... --from alice

  # Query your VCs
  aurad query vcregistry vcs-by-holder $(aurad keys show alice -a)

COMMAND ALIASES:
  vc, vcr              → vcregistry
  br, xchain           → bridge
  ir                   → inclusionroutines
  gov, proposal        → governance
  swap, exchange       → dex
  crypto, crypt        → cryptography
  kyc, comp            → compliance
  cs, score            → confidencescore

FLAGS:
  --help               Show help for command
  --home               Directory for config and data
  --chain-id           Chain identifier
  --node               RPC node to connect to
  --output             Output format (text|json|yaml)

For more information about a specific command:
  aurad [command] --help
  aurad help [module]

Documentation: https://docs.aura.network
`)

	return nil
}
