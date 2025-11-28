package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version is the application version
	// This will be set during build using ldflags
	Version = "development"

	// Commit is the git commit hash
	// This will be set during build using ldflags
	Commit = "unknown"

	// BuildDate is the build date
	// This will be set during build using ldflags
	BuildDate = "unknown"
)

// VersionCmd returns the version command for the Aura daemon
func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the application version information",
		Long: `Print the application version, git commit, build date, and Go version.

Use the --long flag to display additional details including module information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printVersion(cmd)
		},
	}

	cmd.Flags().Bool("long", false, "print long version information")

	return cmd
}

// printVersion prints the version information
func printVersion(cmd *cobra.Command) error {
	long, err := cmd.Flags().GetBool("long")
	if err != nil {
		return fmt.Errorf("failed to get long flag: %w", err)
	}

	// Print basic version information
	if long {
		fmt.Printf("Aura Blockchain Daemon (aurad)\n\n")
	}

	fmt.Printf("Version:      %s\n", Version)
	fmt.Printf("Git Commit:   %s\n", Commit)
	fmt.Printf("Build Date:   %s\n", BuildDate)

	if long {
		fmt.Printf("Go Version:   %s\n", runtime.Version())
		fmt.Printf("OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("\n")
		fmt.Printf("Core Modules:\n")
		fmt.Printf("  - auth             (authentication and authorization)\n")
		fmt.Printf("  - identitychange   (identity management)\n")
		fmt.Printf("  - inclusionroutines (inclusion routines)\n")
		fmt.Printf("  - confidencescore  (confidence scoring)\n")
		fmt.Printf("  - vcregistry       (verifiable credentials)\n")
		fmt.Printf("  - dataregistry     (data registry)\n")
		fmt.Printf("  - governance       (governance)\n")
		fmt.Printf("  - compliance       (compliance)\n")
		fmt.Printf("  - cryptography     (cryptographic operations)\n")
		fmt.Printf("  - monitoring       (network monitoring)\n")
		fmt.Printf("  - networksecurity  (network security)\n")
		fmt.Printf("  - walletsecurity   (wallet security)\n")
		fmt.Printf("  - validatorsecurity (validator security)\n")
		fmt.Printf("  - economicsecurity (economic security)\n")
		fmt.Printf("  - privacy          (privacy features)\n")
		fmt.Printf("\n")
		fmt.Printf("Optional Modules:\n")
		fmt.Printf("  - dex              (decentralized exchange)\n")
		fmt.Printf("  - bridge           (cross-chain bridge)\n")
	}

	return nil
}
