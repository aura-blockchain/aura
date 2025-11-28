package main

import (
	"fmt"
	"os"

	"github.com/aequitas/aura/chain/cmd/aurad/cmd"
)

// main is the entry point for the Aura blockchain daemon.
// It initializes the root command and executes the CLI.
func main() {
	// Create root command
	rootCmd := cmd.NewRootCmd()

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing aurad: %v\n", err)
		os.Exit(1)
	}
}
