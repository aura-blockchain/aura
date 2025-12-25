// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// CommandAliases provides a mapping of command aliases to their full names
var CommandAliases = map[string]string{
	"vc":        "vcregistry",
	"vcr":       "vcregistry",
	"br":        "bridge",
	"xchain":    "bridge",
	"ir":        "inclusionroutines",
	"gov":       "governance",
	"proposal":  "governance",
	"swap":      "dex",
	"exchange":  "dex",
	"crypto":    "cryptography",
	"crypt":     "cryptography",
	"kyc":       "compliance",
	"comp":      "compliance",
	"cs":        "confidencescore",
	"score":     "confidencescore",
	"ns":        "networksecurity",
	"netsec":    "networksecurity",
	"vs":        "validatorsecurity",
	"valsec":    "validatorsecurity",
	"ws":        "walletsecurity",
	"walletsec": "walletsecurity",
	"econ":      "economicsecurity",
	"econseс":   "economicsecurity",
	"mon":       "monitoring",
	"monitor":   "monitoring",
	"priv":      "privacy",
	"identity":  "identitychange",
	"idchange":  "identitychange",
	"data":      "dataregistry",
	"registry":  "dataregistry",
	"contract":  "contractregistry",
	"wasm":      "wasm",
}

// GetFullCommandName returns the full command name for an alias or the original if not an alias
func GetFullCommandName(alias string) string {
	if fullName, exists := CommandAliases[alias]; exists {
		return fullName
	}
	return alias
}

// SuggestCommands suggests similar commands when a command is not found
func SuggestCommands(input string, availableCommands []string) []string {
	input = strings.ToLower(input)
	var suggestions []string

	// Check aliases first
	if fullName, exists := CommandAliases[input]; exists {
		suggestions = append(suggestions, fullName)
	}

	// Find similar commands using Levenshtein-like logic
	for _, cmd := range availableCommands {
		cmdLower := strings.ToLower(cmd)

		// Exact prefix match
		if strings.HasPrefix(cmdLower, input) {
			suggestions = append(suggestions, cmd)
			continue
		}

		// Contains match
		if strings.Contains(cmdLower, input) {
			suggestions = append(suggestions, cmd)
			continue
		}

		// Check if input contains part of command
		if strings.Contains(input, cmdLower) {
			suggestions = append(suggestions, cmd)
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, s := range suggestions {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}

	return unique
}

// FormatCommandExample formats a command example with proper indentation
func FormatCommandExample(description, command string) string {
	return fmt.Sprintf("  # %s\n  %s\n", description, command)
}

// PrintModuleHelp prints comprehensive help for a module
func PrintModuleHelp(moduleName string) {
	helpText := map[string]string{
		"vcregistry": `
VC Registry Module - Manage Verifiable Credentials and DIDs

Transaction Commands:
  mint-vc, mint, create-vc         - Mint a new verifiable credential
  revoke-vc, revoke                - Revoke your own verifiable credential
  register-did, reg-did, create-did - Register a new DID document
  update-did                       - Update an existing DID document

Query Commands:
  vc [vc-id]                       - Query a verifiable credential by ID
  did [did]                        - Query a DID document
  dids-by-controller [address]     - Query DIDs controlled by an address
  policy [vc-type]                 - Query VC policy
  policies                         - List all VC policies

Examples:
  aurad tx vcregistry mint-vc did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from alice
  aurad tx vc mint did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from alice
  aurad query vcregistry vc vc-123456
  aurad query vc did did:aura:mainnet:user123

Interactive Mode:
  aurad interactive
  > vcregistry  (or 'vc')
`,
		"bridge": `
Bridge Module - Cross-chain operations with PAW and XAI

Transaction Commands:
  link-address, link, link-addr    - Link AURA/PAW/XAI addresses
  lock-tokens, lock                - Lock tokens for cross-chain transfer
  unlock-tokens, unlock            - Unlock tokens from another chain

Query Commands:
  linked-addresses [address]       - Query linked addresses
  params                           - Query bridge parameters

Examples:
  aurad tx bridge link-address aura1abc... paw1def... xai1ghi... --from alice
  aurad tx br link aura1abc... paw1def... "" --from alice
  aurad query bridge linked-addresses aura1abc...

Interactive Mode:
  aurad interactive
  > bridge  (or 'br')
`,
		"inclusionroutines": `
Inclusion Routines Module - Complete verification routines

Transaction Commands:
  complete                         - Complete an inclusion routine

Query Commands:
  ir [ir-id]                       - Query inclusion routine by ID
  list                             - List all available IRs
  user-irs [address]               - Query user's completed IRs

Examples:
  aurad tx inclusionroutines complete ir-basic-verification --from alice
  aurad tx ir complete ir-basic-verification --from alice
  aurad query ir list
  aurad query ir user-irs aura1abc...

Interactive Mode:
  aurad interactive
  > ir  (or 'inclusionroutines')
`,
		"governance": `
Governance Module - Submit and vote on proposals

Transaction Commands:
  submit-proposal                  - Submit a governance proposal
  deposit                          - Deposit tokens to a proposal
  vote                             - Vote on a proposal
  vote-weighted                    - Vote with weighted options

Query Commands:
  proposal [id]                    - Query proposal by ID
  proposals                        - Query all proposals
  votes [proposal-id]              - Query votes on a proposal

Examples:
  aurad tx governance submit-proposal --from alice
  aurad tx gov vote 1 yes --from alice
  aurad query gov proposal 1

Aliases: gov, proposal
`,
		"dex": `
DEX Module - Decentralized exchange operations

Transaction Commands:
  create-pool                      - Create a liquidity pool
  add-liquidity                    - Add liquidity to a pool
  remove-liquidity                 - Remove liquidity from a pool
  swap                             - Swap tokens

Query Commands:
  pool [pool-id]                   - Query liquidity pool
  pools                            - List all pools
  order [order-id]                 - Query order by ID

Examples:
  aurad tx dex create-pool uaura:upaw --from alice
  aurad tx swap swap 1000uaura upaw --from alice
  aurad query dex pools

Aliases: swap, exchange
`,
	}

	if help, exists := helpText[moduleName]; exists {
		fmt.Println(help)
	} else {
		fmt.Printf("No detailed help available for module: %s\n", moduleName)
		fmt.Println("Try: aurad [module] --help")
	}
}

// AddCommonFlags adds commonly used flags to a command
func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("yes", false, "Skip confirmation prompts")
	cmd.Flags().Bool("dry-run", false, "Simulate transaction without broadcasting")
	cmd.Flags().String("memo", "", "Add memo to transaction")
}

// ValidateCommonArgs validates common argument patterns
func ValidateCommonArgs(args []string, expectedCount int, argNames []string) error {
	if len(args) != expectedCount {
		var names string
		if len(argNames) > 0 {
			names = " [" + strings.Join(argNames, "] [") + "]"
		}
		return fmt.Errorf("expected %d arguments%s, got %d", expectedCount, names, len(args))
	}
	return nil
}

// PrintSuccessMessage prints a formatted success message
func PrintSuccessMessage(operation, details string) {
	fmt.Printf("\n✓ Success: %s\n", operation)
	if details != "" {
		fmt.Printf("  %s\n", details)
	}
	fmt.Println()
}

// PrintWarningMessage prints a formatted warning message
func PrintWarningMessage(warning string) {
	fmt.Printf("\n⚠ Warning: %s\n\n", warning)
}

// ConfirmAction prompts the user to confirm an action
func ConfirmAction(action string) bool {
	fmt.Printf("Are you sure you want to %s? (y/N): ", action)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// GetQuickStartGuide returns a quick start guide for new users
func GetQuickStartGuide() string {
	return `
AURA Blockchain - Quick Start Guide

1. Initialize your node:
   aurad init <moniker> --chain-id aura-mainnet

2. Create or import a key:
   aurad keys add mykey
   aurad keys add mykey --recover (import with mnemonic)

3. Check your balance:
   aurad query bank balances $(aurad keys show mykey -a)

4. Common Operations:

   Mint a VC:
     aurad tx vcregistry mint-vc did:aura:mainnet:user123 VC_TYPE_VERIFIED_HUMAN --from mykey

   Link cross-chain addresses:
     aurad tx bridge link-address <aura-addr> <paw-addr> <xai-addr> --from mykey

   Complete an IR:
     aurad tx inclusionroutines complete ir-basic-verification --from mykey

5. Interactive Mode (Recommended for beginners):
   aurad interactive

6. Get Help:
   aurad --help
   aurad [command] --help
   aurad help [module]

7. Auto-completion:
   source <(aurad completion bash)   # For Bash
   aurad completion zsh > ~/.zsh/completion/_aurad  # For Zsh

Command Aliases:
  vc, vcr          → vcregistry
  br, xchain       → bridge
  ir               → inclusionroutines
  gov, proposal    → governance
  swap, exchange   → dex

For more information, visit: https://docs.aura.network
`
}
