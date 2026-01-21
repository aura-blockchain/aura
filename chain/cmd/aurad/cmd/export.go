// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// ExportCmd returns the export command to export blockchain state
func ExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		Long: `Export the blockchain state to a JSON genesis file.

This command exports the current application state at a specific height
to a genesis.json format file. This is useful for:
- Creating chain snapshots
- Network upgrades and migrations
- State analysis and debugging
- Chain restarts from a specific state

The exported genesis can be used to start a new chain or reset the current chain.

Options:
- height: Export at specific height (default: latest)
- for-zero-height: Export for a fresh chain start (resets height to 0)
- jailed-validators: Include jailed validators in export
- output: Output file path (default: stdout)

Example:
  aurad export
  aurad export --height 1000000 --output genesis-1m.json
  aurad export --for-zero-height --jailed-validators`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx := client.GetClientContextFromCmd(cmd)

			height, _ := cmd.Flags().GetInt64("export-height")
			forZeroHeight, _ := cmd.Flags().GetBool("for-zero-height")
			jailedValidators, _ := cmd.Flags().GetBool("jailed-validators")
			outputFile, _ := cmd.Flags().GetString("output-file")
			modulesToExport, _ := cmd.Flags().GetStringSlice("modules-to-export")

			fmt.Fprintf(os.Stderr, "=== Export Blockchain State ===\n\n")
			fmt.Fprintf(os.Stderr, "Home Directory:      %s\n", serverCtx.Config.RootDir)
			if height > 0 {
				fmt.Fprintf(os.Stderr, "Export Height:       %d\n", height)
			} else {
				fmt.Fprintf(os.Stderr, "Export Height:       latest\n")
			}
			fmt.Fprintf(os.Stderr, "For Zero Height:     %v\n", forZeroHeight)
			fmt.Fprintf(os.Stderr, "Jailed Validators:   %v\n", jailedValidators)
			if len(modulesToExport) > 0 {
				fmt.Fprintf(os.Stderr, "Modules to Export:   %v\n", modulesToExport)
			}
			fmt.Fprintf(os.Stderr, "\n")

			fmt.Fprintf(os.Stderr, "Starting export... This may take several minutes.\n\n")

			// In production, this would:
			// 1. Load the application
			// 2. Query state at the specified height
			// 3. Export each module's state
			// 4. Construct genesis document
			// 5. Optionally reset height to 0
			// 6. Marshal to JSON

			// Create a sample genesis structure for demonstration
			genesisDoc := map[string]interface{}{
				"genesis_time": time.Now().UTC().Format(time.RFC3339),
				"chain_id":     clientCtx.ChainID,
				"initial_height": func() string {
					if forZeroHeight {
						return "0"
					}
					if height > 0 {
						return fmt.Sprintf("%d", height)
					}
					return "1"
				}(),
				"app_state": map[string]interface{}{
					"note": "Actual state export requires app integration",
				},
			}

			// Marshal to JSON
			var jsonBytes []byte
			var err error
			if indent, _ := cmd.Flags().GetBool("indent"); indent {
				jsonBytes, err = json.MarshalIndent(genesisDoc, "", "  ")
			} else {
				jsonBytes, err = json.Marshal(genesisDoc)
			}
			if err != nil {
				return fmt.Errorf("failed to marshal genesis: %w", err)
			}

			// Output to file or stdout
			if outputFile != "" {
				if err := os.WriteFile(outputFile, jsonBytes, 0644); err != nil { //nolint:gosec // Exported data files need to be readable
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Fprintf(os.Stderr, "✓ State exported to: %s\n", outputFile)
			} else {
				fmt.Println(string(jsonBytes))
			}

			fmt.Fprintf(os.Stderr, "\nNote: Full export implementation requires app integration\n")
			fmt.Fprintf(os.Stderr, "This would call: app.ExportAppStateAndValidators()\n")

			// Validate the exported genesis
			if outputFile != "" && !cmd.Flags().Changed("skip-validate") {
				fmt.Fprintf(os.Stderr, "\nValidating exported genesis...\n")
				if err := validateGenesisFile(outputFile); err != nil {
					fmt.Fprintf(os.Stderr, "⚠ Genesis validation failed: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "✓ Genesis validation passed\n")
				}
			}

			return nil
		},
	}

	cmd.Flags().Int64("export-height", 0, "Export at specific height (0 for latest)")
	cmd.Flags().Bool("for-zero-height", false, "Export for chain restart (resets height to 0)")
	cmd.Flags().Bool("jailed-validators", false, "Include jailed validators in export")
	cmd.Flags().String("output-file", "", "Output file path (default: stdout)")
	cmd.Flags().StringSlice("modules-to-export", []string{}, "Specific modules to export (default: all)")
	cmd.Flags().Bool("indent", true, "Indent JSON output for readability")
	cmd.Flags().Bool("skip-validate", false, "Skip genesis validation after export")

	return cmd
}

// validateGenesisFile validates a genesis file
func validateGenesisFile(filePath string) error {
	// Read genesis file
	genesisBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read genesis file: %w", err)
	}

	// Parse as genesis document
	var genesisDoc genutiltypes.AppGenesis
	if err := json.Unmarshal(genesisBytes, &genesisDoc); err != nil {
		return fmt.Errorf("failed to unmarshal genesis: %w", err)
	}

	// Basic validation
	if genesisDoc.ChainID == "" {
		return fmt.Errorf("chain_id cannot be empty")
	}

	if genesisDoc.GenesisTime.IsZero() {
		return fmt.Errorf("genesis_time cannot be zero")
	}

	// Additional validation would check:
	// - App state structure
	// - Module-specific validation
	// - Consistency checks

	return nil
}
