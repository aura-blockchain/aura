// E2E Testnet Validation Tool
// Run against live AURA testnet infrastructure
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	e2e "e2e_testnet"
)

func main() {
	// Command line flags
	var (
		verbose     = flag.Bool("v", false, "Verbose output")
		scenario    = flag.String("scenario", "", "Run specific scenario (empty=all)")
		listOnly    = flag.Bool("list", false, "List available scenarios")
		outputDir   = flag.String("output", "", "Output directory for results (default: ~/testnets/aura-mvp-1/results)")
		jsonOutput  = flag.Bool("json", false, "Output results as JSON")
		mdOutput    = flag.Bool("md", true, "Output results as Markdown")
		timeout     = flag.Duration("timeout", 5*time.Minute, "Test timeout")
		allTests    = flag.Bool("all", false, "Run all validation tests (phases 1-5)")
		scenarioRun = flag.Bool("scenarios", false, "Run E2E scenarios")
		localOnly   = flag.Bool("local", true, "Test only local validators (no SSH to other servers)")
		fullNetwork = flag.Bool("full", false, "Test all validators including remote (requires SSH)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "AURA Testnet E2E Validation Tool\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -all                     # Run all tests (local validators only, no SSH)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -all -full               # Run all tests on ALL validators (requires SSH)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -scenarios               # Run all E2E scenarios\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -scenario network_health # Run specific scenario\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list                    # List available scenarios\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nDefault: Tests only local validators (val1, val2 on this server)\n")
		fmt.Fprintf(os.Stderr, "Results saved to: ~/testnets/aura-mvp-1/results/\n")
	}

	flag.Parse()

	// Create config - use local-only by default (no SSH required)
	var cfg *e2e.TestnetConfig
	if *fullNetwork {
		cfg = e2e.DefaultTestnetConfig()
	} else if *localOnly {
		cfg = e2e.LocalOnlyConfig()
	} else {
		cfg = e2e.LocalOnlyConfig()
	}
	e2e.LoadConfigFromEnv(cfg)

	if *timeout > 0 {
		cfg.Timeout = *timeout
	}

	// Set default output directory if not specified
	if *outputDir == "" {
		*outputDir = e2e.DefaultOutputDir()
	}

	// List scenarios if requested
	if *listOnly {
		sr := e2e.NewScenarioRunner(cfg, *verbose)
		fmt.Println("Available E2E Scenarios:")
		for _, s := range sr.ListScenarios() {
			fmt.Printf("  %-20s %s\n", s.Name, s.Description)
		}
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Run tests
	if *allTests || (!*scenarioRun && *scenario == "") {
		// Run all validation tests (phases 1-5)
		runner := e2e.NewRunner(cfg, *verbose)
		suite := runner.RunAll(ctx)

		// Save results
		timestamp := time.Now().Format("20060102-150405")

		if *jsonOutput {
			jsonPath := filepath.Join(*outputDir, fmt.Sprintf("validation-%s.json", timestamp))
			if err := runner.SaveResults(jsonPath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save JSON: %v\n", err)
			} else {
				fmt.Printf("Results saved to: %s\n", jsonPath)
			}
		}

		if *mdOutput {
			mdPath := filepath.Join(*outputDir, fmt.Sprintf("VALIDATION-%s.md", timestamp))
			if err := runner.SaveMarkdown(mdPath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save Markdown: %v\n", err)
			} else {
				fmt.Printf("Report saved to: %s\n", mdPath)
			}
		}

		if suite.FailedTests > 0 {
			os.Exit(1)
		}
		return
	}

	// Run scenarios
	if *scenarioRun || *scenario != "" {
		sr := e2e.NewScenarioRunner(cfg, *verbose)

		var err error
		if *scenario != "" {
			err = sr.RunScenario(ctx, *scenario)
		} else {
			err = sr.RunAllScenarios(ctx)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Scenario failed: %v\n", err)
			os.Exit(1)
		}
	}
}
