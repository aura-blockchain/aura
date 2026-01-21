// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package coverage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// CoverageConfig defines the configuration for coverage requirements
type CoverageConfig struct {
	MinimumCoverage float64 // Minimum coverage percentage required (e.g., 90.0)
	ExcludePaths    []string
	IncludePaths    []string
}

// DefaultCoverageConfig returns the default coverage configuration
func DefaultCoverageConfig() *CoverageConfig {
	return &CoverageConfig{
		MinimumCoverage: 90.0,
		ExcludePaths: []string{
			"*_test.go",
			"*.pb.go",
			"**/testutil/**",
			"**/mock/**",
		},
		IncludePaths: []string{
			"**/keeper/**",
			"**/types/**",
			"**/handler/**",
		},
	}
}

// ModuleCoverage represents coverage statistics for a module
type ModuleCoverage struct {
	ModuleName      string
	TotalStatements int
	CoveredLines    int
	UncoveredLines  int
	CoveragePercent float64
	Files           []FileCoverage
}

// FileCoverage represents coverage for a single file
type FileCoverage struct {
	FilePath        string
	TotalStatements int
	CoveredLines    int
	CoveragePercent float64
}

// CoverageReport holds the overall coverage report
type CoverageReport struct {
	Modules         []ModuleCoverage
	TotalCoverage   float64
	PassesThreshold bool
	Threshold       float64
}

// AnalyzeModuleCoverage analyzes coverage for a specific module
func AnalyzeModuleCoverage(modulePath string, config *CoverageConfig) (*ModuleCoverage, error) {
	module := &ModuleCoverage{
		ModuleName: filepath.Base(modulePath),
		Files:      make([]FileCoverage, 0),
	}

	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files and generated files
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".pb.go") {
			return nil
		}

		// Skip excluded paths
		for _, exclude := range config.ExcludePaths {
			if matched, _ := filepath.Match(exclude, filepath.Base(path)); matched {
				return nil
			}
		}

		// Parse file and count statements
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", path, err)
		}

		stmtCount := countStatements(node)
		module.TotalStatements += stmtCount

		fileCov := FileCoverage{
			FilePath:        path,
			TotalStatements: stmtCount,
		}
		module.Files = append(module.Files, fileCov)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return module, nil
}

// countStatements counts the number of statements in an AST node
func countStatements(node ast.Node) int {
	if node == nil {
		return 0
	}

	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.ExprStmt, *ast.AssignStmt, *ast.ReturnStmt,
			*ast.IfStmt, *ast.ForStmt, *ast.RangeStmt,
			*ast.SwitchStmt, *ast.SelectStmt, *ast.GoStmt,
			*ast.DeferStmt, *ast.SendStmt, *ast.IncDecStmt:
			count++
		}
		return true
	})
	return count
}

// GenerateReport generates a comprehensive coverage report
func GenerateReport(rootPath string, config *CoverageConfig) (*CoverageReport, error) {
	report := &CoverageReport{
		Modules:   make([]ModuleCoverage, 0),
		Threshold: config.MinimumCoverage,
	}

	// Find all modules
	modulesPath := filepath.Join(rootPath, "x")
	entries, err := os.ReadDir(modulesPath)
	if err != nil {
		return nil, fmt.Errorf("error reading modules directory: %w", err)
	}

	totalStmts := 0
	totalCovered := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modulePath := filepath.Join(modulesPath, entry.Name())
		moduleCov, err := AnalyzeModuleCoverage(modulePath, config)
		if err != nil {
			return nil, fmt.Errorf("error analyzing module %s: %w", entry.Name(), err)
		}

		report.Modules = append(report.Modules, *moduleCov)
		totalStmts += moduleCov.TotalStatements
		totalCovered += moduleCov.CoveredLines
	}

	if totalStmts > 0 {
		report.TotalCoverage = (float64(totalCovered) / float64(totalStmts)) * 100
	}

	report.PassesThreshold = report.TotalCoverage >= config.MinimumCoverage

	return report, nil
}

// PrintReport prints a human-readable coverage report
func PrintReport(report *CoverageReport) {
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Println("  AURA BLOCKCHAIN COVERAGE REPORT")
	fmt.Println("=" + strings.Repeat("=", 78))
	fmt.Printf("\nThreshold: %.2f%%\n", report.Threshold)
	fmt.Printf("Total Coverage: %.2f%%\n", report.TotalCoverage)
	fmt.Printf("Status: ")
	if report.PassesThreshold {
		fmt.Println("✓ PASS")
	} else {
		fmt.Println("✗ FAIL")
	}
	fmt.Println()

	fmt.Println("Module Breakdown:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-30s %15s %15s %15s\n", "Module", "Total Stmts", "Covered", "Coverage")
	fmt.Println(strings.Repeat("-", 80))

	for _, module := range report.Modules {
		fmt.Printf("%-30s %15d %15d %14.2f%%\n",
			module.ModuleName,
			module.TotalStatements,
			module.CoveredLines,
			module.CoveragePercent)
	}
	fmt.Println(strings.Repeat("-", 80))
}
