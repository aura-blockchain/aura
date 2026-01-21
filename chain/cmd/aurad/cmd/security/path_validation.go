// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Maximum path length to prevent DoS
	MaxPathLength = 4096
)

// PathValidator validates and sanitizes file paths
type PathValidator struct {
	allowedBasePaths []string
	logger           Logger
}

// NewPathValidator creates a new path validator
func NewPathValidator(logger Logger) *PathValidator {
	userHome, _ := os.UserHomeDir()
	return &PathValidator{
		allowedBasePaths: []string{
			userHome,
			"/tmp",
			"/var/tmp",
		},
		logger: logger,
	}
}

// ValidateAndCleanHomePath validates and cleans the home directory path
// Returns an error if the path is suspicious or invalid
func (pv *PathValidator) ValidateAndCleanHomePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check path length
	if len(path) > MaxPathLength {
		pv.logger.SecurityEvent("path_validation_failed", map[string]interface{}{
			"reason": "path_too_long",
			"length": len(path),
		})
		return "", fmt.Errorf("path exceeds maximum length of %d characters", MaxPathLength)
	}

	// Check for null bytes (path traversal attack vector)
	if strings.Contains(path, "\x00") {
		pv.logger.SecurityEvent("path_validation_failed", map[string]interface{}{
			"reason": "null_byte_detected",
			"path":   sanitizePath(path),
		})
		return "", fmt.Errorf("path contains null bytes")
	}

	// Clean the path (removes .., ., etc.)
	cleanedPath := filepath.Clean(path)

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanedPath)
	if err != nil {
		pv.logger.SecurityEvent("path_validation_failed", map[string]interface{}{
			"reason": "abs_path_failed",
			"error":  err.Error(),
		})
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Resolve symlinks to detect symlink attacks
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If the path doesn't exist yet, that's okay for init commands
		// But we still validate the parent directory
		if !os.IsNotExist(err) {
			pv.logger.SecurityEvent("path_validation_failed", map[string]interface{}{
				"reason": "symlink_resolution_failed",
				"error":  err.Error(),
			})
			return "", fmt.Errorf("failed to resolve symlinks: %w", err)
		}
		// For non-existent paths, use the absolute path
		realPath = absPath
	}

	// Verify the path is within allowed base paths
	allowed := false
	for _, basePath := range pv.allowedBasePaths {
		// Resolve base path symlinks
		realBasePath, err := filepath.EvalSymlinks(basePath)
		if err != nil {
			// If base path doesn't exist, use original
			realBasePath = basePath
		}

		// Check if path is under this base path
		relPath, err := filepath.Rel(realBasePath, realPath)
		if err == nil && !strings.HasPrefix(relPath, "..") {
			allowed = true
			break
		}
	}

	if !allowed {
		pv.logger.SecurityEvent("path_validation_failed", map[string]interface{}{
			"reason":        "path_outside_allowed_bases",
			"path":          sanitizePath(realPath),
			"allowed_bases": pv.allowedBasePaths,
		})
		return "", fmt.Errorf("path must be within user home directory or /tmp")
	}

	// Additional checks for suspicious patterns
	if err := pv.checkSuspiciousPatterns(realPath); err != nil {
		return "", err
	}

	pv.logger.SecurityEvent("path_validated", map[string]interface{}{
		"original_path": sanitizePath(path),
		"resolved_path": sanitizePath(realPath),
	})

	return realPath, nil
}

// checkSuspiciousPatterns checks for suspicious path patterns
func (pv *PathValidator) checkSuspiciousPatterns(path string) error {
	suspicious := []string{
		"/etc/",
		"/root/",
		"/sys/",
		"/proc/",
		"/dev/",
		"/boot/",
		"/.ssh/",
		"/../",
		"/./",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range suspicious {
		if strings.Contains(lowerPath, pattern) {
			pv.logger.SecurityEvent("suspicious_path_detected", map[string]interface{}{
				"path":    sanitizePath(path),
				"pattern": pattern,
			})
			return fmt.Errorf("path contains suspicious pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateFilePath validates a file path for reading/writing
func (pv *PathValidator) ValidateFilePath(path string, maxSize int64) error {
	// First validate the path itself
	validPath, err := pv.ValidateAndCleanHomePath(path)
	if err != nil {
		return err
	}

	// Check if file exists and get info
	info, err := os.Stat(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, which is okay
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if it's actually a file
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	// Check file size
	if info.Size() > maxSize {
		pv.logger.SecurityEvent("file_size_exceeded", map[string]interface{}{
			"path":     sanitizePath(validPath),
			"size":     info.Size(),
			"max_size": maxSize,
		})
		return fmt.Errorf("file size %d exceeds maximum %d", info.Size(), maxSize)
	}

	return nil
}

// sanitizePath removes sensitive information from paths for logging
func sanitizePath(path string) string {
	userHome, _ := os.UserHomeDir()
	if userHome != "" {
		path = strings.Replace(path, userHome, "~", 1)
	}
	return path
}

// IsPathTraversal checks if a path contains traversal patterns
func IsPathTraversal(path string) bool {
	// Check for common path traversal patterns
	patterns := []string{
		"../",
		"..\\",
		"..",
		"%2e%2e",
		"%252e%252e",
		"..%2f",
		"..%5c",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range patterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	return false
}

// SecureJoin safely joins path components, preventing traversal
func SecureJoin(base string, paths ...string) (string, error) {
	result := filepath.Clean(base)

	for _, path := range paths {
		// Check for traversal attempts
		if IsPathTraversal(path) {
			return "", fmt.Errorf("path traversal detected in: %s", path)
		}

		// Clean the path component
		cleanPath := filepath.Clean(path)

		// Join with result
		result = filepath.Join(result, cleanPath)
	}

	// Verify the result is still under the base
	relPath, err := filepath.Rel(base, result)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("joined path escapes base directory")
	}

	return result, nil
}
