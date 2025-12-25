// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoUnsafeUsage ensures that no module in chain/x uses the unsafe package.
// This is a critical security requirement as unsafe pointer operations can:
// - Violate type safety and memory safety
// - Lead to undefined behavior
// - Cause crashes and data corruption
// - Create security vulnerabilities
//
// All type conversions should use proper protobuf marshaling/unmarshaling
// or safe type assertions.
func TestNoUnsafeUsage(t *testing.T) {
	// Navigate to the x directory from the test location
	xDir := filepath.Join("..", "..")

	// Walk through all Go files in chain/x
	err := filepath.Walk(xDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files (this file itself)
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip the internal/tests directory itself
		if strings.Contains(path, "internal/tests") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", path, err)
			return nil
		}

		// Check imports for "unsafe"
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if importPath == "unsafe" {
				t.Errorf("CRITICAL SECURITY VIOLATION: File %s imports 'unsafe' package.\n"+
					"Unsafe pointer operations are strictly forbidden.\n"+
					"Use proper protobuf marshaling (codec.Marshal/Unmarshal) or safe type conversions instead.",
					path)
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk x directory: %v", err)
	}
}

// TestNoUnsafePointerCasts checks for unsafe.Pointer casts in the codebase
// by parsing the AST of all Go files and looking for unsafe.Pointer type assertions.
func TestNoUnsafePointerCasts(t *testing.T) {
	// Navigate to the x directory from the test location
	xDir := filepath.Join("..", "..")

	// Walk through all Go files in chain/x
	err := filepath.Walk(xDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip the internal/tests directory itself
		if strings.Contains(path, "internal/tests") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read %s: %v", path, err)
			return nil
		}

		fileContent := string(content)

		// Check for common unsafe patterns
		unsafePatterns := []string{
			"unsafe.Pointer",
			"(*unsafe.Pointer)",
			"unsafe.Sizeof",
			"unsafe.Alignof",
			"unsafe.Offsetof",
		}

		for _, pattern := range unsafePatterns {
			if strings.Contains(fileContent, pattern) {
				t.Errorf("CRITICAL SECURITY VIOLATION: File %s contains unsafe pattern '%s'.\n"+
					"All unsafe operations are strictly forbidden in production code.",
					path, pattern)
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk x directory: %v", err)
	}
}

// TestProperTypeConversions provides examples of proper, safe type conversion patterns
// that should be used instead of unsafe pointer operations.
func TestProperTypeConversions(t *testing.T) {
	t.Log("✓ APPROVED PATTERN: Direct type alias usage")
	t.Log("  type Params = pb.Params")
	t.Log("  p := DefaultParams()")
	t.Log("  return &p  // Safe: p is already the correct type")
	t.Log("")

	t.Log("✓ APPROVED PATTERN: Protobuf marshaling for complex conversions")
	t.Log("  bz, err := k.cdc.Marshal(sourceObj)")
	t.Log("  if err != nil { return err }")
	t.Log("  err = k.cdc.Unmarshal(bz, &targetObj)")
	t.Log("")

	t.Log("✓ APPROVED PATTERN: Manual field-by-field conversion")
	t.Log("  return &TargetType{")
	t.Log("    Field1: src.Field1,")
	t.Log("    Field2: src.Field2,")
	t.Log("  }")
	t.Log("")

	t.Log("✗ FORBIDDEN PATTERN: Unsafe pointer conversion")
	t.Log("  return (*TargetType)(unsafe.Pointer(src))  // NEVER DO THIS")
	t.Log("")

	// This test always passes - it's documentation
	t.Log("These patterns ensure type safety, prevent memory corruption, and maintain security.")
}
