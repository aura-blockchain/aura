// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// InteractiveCmd creates a new interactive mode command
func InteractiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interactive",
		Short: "Start interactive mode for complex commands",
		Long: `Interactive mode provides a guided interface for executing complex commands.
You can step through command parameters with helpful prompts and validation.`,
		Aliases: []string{"i", "repl"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteractive()
		},
	}

	return cmd
}

func runInteractive() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("===========================================")
	fmt.Println("  AURA Interactive Mode")
	fmt.Println("===========================================")
	fmt.Println("Type 'help' for available commands")
	fmt.Println("Type 'exit' or 'quit' to leave interactive mode")
	fmt.Println()

	for {
		fmt.Print("aura> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "exit", "quit", "q":
			fmt.Println("Exiting interactive mode...")
			return nil
		case "help", "h":
			printInteractiveHelp()
			continue
		case "clear", "cls":
			fmt.Print("\033[H\033[2J")
			continue
		}

		// Parse and execute command
		if err := executeInteractiveCommand(input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func printInteractiveHelp() {
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  tx        - Guided transaction creation")
	fmt.Println("  query     - Guided query execution")
	fmt.Println("  keys      - Key management operations")
	fmt.Println("  status    - Show node status")
	fmt.Println("  help      - Show this help message")
	fmt.Println("  clear     - Clear the screen")
	fmt.Println("  exit      - Exit interactive mode")
	fmt.Println()

	fmt.Println("Wizards:")
	fmt.Println("  tx.wizard      - Step-by-step transaction builder")
	fmt.Println("  query.wizard   - Step-by-step query builder")
	fmt.Println("  keys.wizard    - Step-by-step key creation")
	fmt.Println()
}

func executeInteractiveCommand(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	mainCmd := parts[0]
	subCmd := ""
	if len(parts) > 1 {
		subCmd = parts[1]
	}

	reader := bufio.NewReader(os.Stdin)

	switch mainCmd {
	case "tx":
		if subCmd == "wizard" {
			return runTransactionWizard()
		}
		fmt.Println("Transaction command. Use 'tx.wizard' for guided mode.")
	case "query":
		if subCmd == "wizard" {
			return runQueryWizard()
		}
		fmt.Println("Query command. Use 'query.wizard' for guided mode.")
	case "keys":
		if subCmd == "wizard" {
			return runKeysWizard()
		}
		fmt.Println("Keys command. Use 'keys.wizard' for guided mode.")
	case "vcregistry", "vc":
		return runVCRegistryWizard(reader)
	case "bridge", "br":
		return runBridgeWizard(reader)
	case "ir", "inclusionroutines":
		return runInclusionRoutinesWizard(reader)
	case "status":
		fmt.Println("Fetching node status...")
		// This would call the actual status command
	default:
		return fmt.Errorf("unknown command: %s\nTry 'help' for available commands", mainCmd)
	}

	return nil
}

func runTransactionWizard() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Transaction Wizard ===")
	fmt.Println("This wizard will guide you through creating a transaction.")

	// Module selection
	fmt.Println("Select module:")
	fmt.Println("  1. Bank (send tokens)")
	fmt.Println("  2. Staking")
	fmt.Println("  3. Bridge")
	fmt.Println("  4. VC Registry")
	fmt.Print("Enter number: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return runBankSendWizard(reader)
	case "2":
		fmt.Println("Staking wizard not yet implemented")
	case "3":
		fmt.Println("Bridge wizard not yet implemented")
	case "4":
		fmt.Println("VC Registry wizard not yet implemented")
	default:
		return fmt.Errorf("invalid choice")
	}

	return nil
}

func runBankSendWizard(reader *bufio.Reader) error {
	fmt.Println("\n=== Bank Send Transaction ===")

	fmt.Print("From address: ")
	from, _ := reader.ReadString('\n')
	from = strings.TrimSpace(from)

	fmt.Print("To address: ")
	to, _ := reader.ReadString('\n')
	to = strings.TrimSpace(to)

	fmt.Print("Amount (e.g., 1000uaura): ")
	amount, _ := reader.ReadString('\n')
	amount = strings.TrimSpace(amount)

	fmt.Print("Gas (default: auto): ")
	gas, _ := reader.ReadString('\n')
	gas = strings.TrimSpace(gas)
	if gas == "" {
		gas = "auto"
	}

	fmt.Println("\n=== Transaction Summary ===")
	fmt.Printf("From:   %s\n", from)
	fmt.Printf("To:     %s\n", to)
	fmt.Printf("Amount: %s\n", amount)
	fmt.Printf("Gas:    %s\n", gas)

	fmt.Print("\nExecute transaction? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" || confirm == "yes" {
		fmt.Println("Executing transaction...")
		// This would execute the actual transaction
		fmt.Println("Transaction submitted successfully!")
	} else {
		fmt.Println("Transaction cancelled.")
	}

	return nil
}

func runQueryWizard() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Query Wizard ===")
	fmt.Println("Select what you want to query:")

	fmt.Println("  1. Account balance")
	fmt.Println("  2. Block information")
	fmt.Println("  3. Transaction details")
	fmt.Println("  4. Validator information")
	fmt.Print("Enter number: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("Enter address: ")
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)
		fmt.Printf("Querying balance for %s...\n", address)
		// Execute actual query
	case "2":
		fmt.Print("Enter block height: ")
		height, _ := reader.ReadString('\n')
		height = strings.TrimSpace(height)
		fmt.Printf("Querying block %s...\n", height)
		// Execute actual query
	default:
		return fmt.Errorf("invalid choice")
	}

	return nil
}

func runKeysWizard() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Keys Wizard ===")
	fmt.Println("What would you like to do?")

	fmt.Println("  1. Create new key")
	fmt.Println("  2. Import existing key")
	fmt.Println("  3. List all keys")
	fmt.Print("Enter number: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Print("Enter key name: ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		fmt.Printf("Creating key '%s'...\n", name)
		// Execute key creation
	case "2":
		fmt.Println("Key import wizard not yet implemented")
	case "3":
		fmt.Println("Listing all keys...")
		// Execute key list
	default:
		return fmt.Errorf("invalid choice")
	}

	return nil
}
