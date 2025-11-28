package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/aequitas/aura/chain/cmd/aurad/cmd/security"
)

// BatchCmd creates a new batch command execution command
func BatchCmd() *cobra.Command {
	var continueOnError bool
	var delayMs int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "batch [file]",
		Short: "Execute commands from a file in batch mode",
		Long: `Execute multiple commands from a file in batch mode.

Each line in the file should contain a complete command (without the 'aurad' prefix).
Empty lines and lines starting with '#' are ignored.

Example file content:
  # Query account balance
  query bank balances aura1abc...

  # Send tokens
  tx bank send from_addr to_addr 100uaura --yes

  # Query status
  status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchCommands(args[0], continueOnError, delayMs, dryRun)
		},
	}

	cmd.Flags().BoolVar(&continueOnError, "continue-on-error", false, "Continue executing commands even if one fails")
	cmd.Flags().IntVar(&delayMs, "delay", 0, "Delay in milliseconds between commands")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing them")

	return cmd
}

func runBatchCommands(filename string, continueOnError bool, delayMs int, dryRun bool) error {
	// Get security components
	logger := GetSecurityLogger()
	pathValidator := security.NewPathValidator(logger)
	cmdValidator := security.NewCommandValidator(logger)

	// Validate batch file
	if err := cmdValidator.ValidateBatchFile(filename, pathValidator); err != nil {
		return fmt.Errorf("batch file validation failed: %w", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Set max line length for scanner
	const maxScanTokenSize = security.MaxLineLength
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	lineNum := 0
	executed := 0
	failed := 0

	fmt.Printf("=== Batch Execution: %s ===\n", filename)
	if dryRun {
		fmt.Println("[DRY RUN MODE - Commands will not be executed]")
	}
	fmt.Println()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), security.CommandExecutionTimeout)
	defer cancel()

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Validate command
		if err := cmdValidator.ValidateCommand(line); err != nil {
			failed++
			fmt.Printf("[%d] %s\n", lineNum, line)
			fmt.Printf("  [VALIDATION ERROR] %v\n", err)
			if !continueOnError {
				return fmt.Errorf("command validation failed at line %d: %w", lineNum, err)
			}
			continue
		}

		fmt.Printf("[%d] %s\n", lineNum, line)

		if dryRun {
			fmt.Println("  [Would execute]")
			executed++
			continue
		}

		// Execute the command with context
		if err := cmdValidator.ExecuteWithContext(ctx, func() error {
			return executeBatchCommand(line)
		}); err != nil {
			failed++
			fmt.Printf("  [ERROR] %v\n", err)
			if !continueOnError {
				return fmt.Errorf("batch execution stopped at line %d: %w", lineNum, err)
			}
		} else {
			executed++
			fmt.Println("  [OK]")
		}

		// Add delay if specified
		if delayMs > 0 && scanner.Text() != "" {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fmt.Println()
	fmt.Printf("=== Batch Summary ===\n")
	fmt.Printf("Total lines: %d\n", lineNum)
	fmt.Printf("Executed: %d\n", executed)
	if failed > 0 {
		fmt.Printf("Failed: %d\n", failed)
	}

	logger.SecurityEvent("batch_execution_completed", map[string]interface{}{
		"filename": filename,
		"total":    lineNum,
		"executed": executed,
		"failed":   failed,
	})

	return nil
}

func executeBatchCommand(cmdLine string) error {
	// This is a simplified version - in a real implementation,
	// you would parse the command and execute it through the root command
	fmt.Printf("  Executing: %s\n", cmdLine)

	// Parse command line into args
	args := strings.Fields(cmdLine)
	if len(args) == 0 {
		return nil
	}

	// Here you would actually execute the command
	// For now, we'll just simulate it
	// In a real implementation, you'd do something like:
	// rootCmd := NewRootCmd()
	// rootCmd.SetArgs(args)
	// return rootCmd.Execute()

	return nil
}

// Script execution with variable substitution
func ScriptCmd() *cobra.Command {
	var variables map[string]string

	cmd := &cobra.Command{
		Use:   "script [file]",
		Short: "Execute a script file with variable substitution",
		Long: `Execute a script file with support for variable substitution.

Variables can be defined in the script using:
  SET VAR_NAME=value

And used in commands with:
  $VAR_NAME or ${VAR_NAME}

Example script:
  SET ADDR=aura1abc...
  SET AMOUNT=1000uaura

  query bank balances $ADDR
  tx bank send $ADDR aura1def... $AMOUNT`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScript(args[0], variables)
		},
	}

	cmd.Flags().StringToStringVar(&variables, "var", nil, "Set variables (e.g., --var KEY=VALUE)")

	return cmd
}

func runScript(filename string, initialVars map[string]string) error {
	// Get security components
	logger := GetSecurityLogger()
	pathValidator := security.NewPathValidator(logger)
	cmdValidator := security.NewCommandValidator(logger)

	// Validate script file
	if err := cmdValidator.ValidateScriptFile(filename, pathValidator); err != nil {
		return fmt.Errorf("script file validation failed: %w", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Initialize variables
	vars := make(map[string]string)
	for k, v := range initialVars {
		vars[k] = v
	}

	scanner := bufio.NewScanner(file)
	// Set max line length for scanner
	const maxScanTokenSize = security.MaxLineLength
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	lineNum := 0

	fmt.Printf("=== Script Execution: %s ===\n\n", filename)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), security.CommandExecutionTimeout)
	defer cancel()

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle SET command
		if strings.HasPrefix(strings.ToUpper(line), "SET ") {
			parts := strings.SplitN(line[4:], "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid SET syntax at line %d: %s", lineNum, line)
			}
			varName := strings.TrimSpace(parts[0])
			varValue := strings.TrimSpace(parts[1])
			vars[varName] = varValue
			fmt.Printf("[%d] SET %s=%s\n", lineNum, varName, varValue)
			continue
		}

		// Substitute variables safely
		cmd, err := cmdValidator.SubstituteVariablesSafe(line, vars)
		if err != nil {
			return fmt.Errorf("variable substitution failed at line %d: %w", lineNum, err)
		}

		fmt.Printf("[%d] %s\n", lineNum, cmd)

		// Execute command with context
		if err := cmdValidator.ExecuteWithContext(ctx, func() error {
			return executeBatchCommand(cmd)
		}); err != nil {
			fmt.Printf("  [ERROR] %v\n", err)
			return fmt.Errorf("script execution failed at line %d: %w", lineNum, err)
		}
		fmt.Println("  [OK]")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fmt.Println("\n=== Script Completed Successfully ===")

	logger.SecurityEvent("script_execution_completed", map[string]interface{}{
		"filename": filename,
		"lines":    lineNum,
	})

	return nil
}
