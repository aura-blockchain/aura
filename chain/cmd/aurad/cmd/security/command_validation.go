// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	// MaxFileSize is the maximum file size for batch/script files (10MB)
	MaxFileSize = 10 * 1024 * 1024

	// MaxLineCount is the maximum number of lines in a batch/script file
	MaxLineCount = 10000

	// MaxLineLength is the maximum length of a single line
	MaxLineLength = 4096

	// CommandExecutionTimeout is the timeout for command execution
	CommandExecutionTimeout = 5 * time.Minute
)

// CommandValidator validates and sanitizes commands
type CommandValidator struct {
	allowedCommands map[string]bool
	logger          Logger
}

// NewCommandValidator creates a new command validator
func NewCommandValidator(logger Logger) *CommandValidator {
	// Whitelist of allowed commands
	allowedCommands := map[string]bool{
		"query":      true,
		"tx":         true,
		"status":     true,
		"keys":       true,
		"config":     true,
		"version":    true,
		"help":       true,
		"completion": true,
	}

	return &CommandValidator{
		allowedCommands: allowedCommands,
		logger:          logger,
	}
}

// ValidateCommand validates a command string
func (cv *CommandValidator) ValidateCommand(cmdLine string) error {
	if cmdLine == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check line length
	if len(cmdLine) > MaxLineLength {
		cv.logger.SecurityEvent("command_validation_failed", map[string]interface{}{
			"reason": "line_too_long",
			"length": len(cmdLine),
		})
		return fmt.Errorf("command exceeds maximum length of %d characters", MaxLineLength)
	}

	// Parse command
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return fmt.Errorf("command cannot be empty")
	}

	baseCmd := parts[0]

	// Check if command is in whitelist
	if !cv.allowedCommands[baseCmd] {
		cv.logger.SecurityEvent("command_validation_failed", map[string]interface{}{
			"reason":  "command_not_whitelisted",
			"command": baseCmd,
		})
		return fmt.Errorf("command not allowed: %s", baseCmd)
	}

	// Check for shell metacharacters that could lead to injection
	if err := cv.checkShellMetacharacters(cmdLine); err != nil {
		return err
	}

	// Check for suspicious patterns
	if err := cv.checkSuspiciousPatterns(cmdLine); err != nil {
		return err
	}

	return nil
}

// checkShellMetacharacters checks for dangerous shell metacharacters
func (cv *CommandValidator) checkShellMetacharacters(cmdLine string) error {
	// Dangerous shell metacharacters
	dangerous := []string{
		";",  // Command separator
		"|",  // Pipe
		"&",  // Background execution
		"$",  // Variable expansion (except in allowed contexts)
		"`",  // Command substitution
		"<",  // Input redirection
		">",  // Output redirection
		"\n", // Newline
		"\r", // Carriage return
		"(",  // Subshell
		")",  // Subshell
		"{",  // Brace expansion
		"}",  // Brace expansion
		"\\", // Escape character
	}

	for _, char := range dangerous {
		if strings.Contains(cmdLine, char) {
			cv.logger.SecurityEvent("shell_metacharacter_detected", map[string]interface{}{
				"command":   sanitizeCommand(cmdLine),
				"character": char,
			})
			return fmt.Errorf("command contains disallowed character: %s", char)
		}
	}

	return nil
}

// checkSuspiciousPatterns checks for suspicious command patterns
func (cv *CommandValidator) checkSuspiciousPatterns(cmdLine string) error {
	suspicious := []string{
		"rm ",
		"del ",
		"format ",
		"mkfs",
		"dd ",
		"curl ",
		"wget ",
		"nc ",
		"netcat",
		"telnet",
		"ssh ",
		"scp ",
		"sftp",
	}

	lowerCmd := strings.ToLower(cmdLine)
	for _, pattern := range suspicious {
		if strings.Contains(lowerCmd, pattern) {
			cv.logger.SecurityEvent("suspicious_command_detected", map[string]interface{}{
				"command": sanitizeCommand(cmdLine),
				"pattern": pattern,
			})
			return fmt.Errorf("command contains suspicious pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateBatchFile validates a batch file before execution
func (cv *CommandValidator) ValidateBatchFile(filename string, pv *PathValidator) error {
	// Validate the file path
	if err := pv.ValidateFilePath(filename, MaxFileSize); err != nil {
		return fmt.Errorf("invalid batch file path: %w", err)
	}

	// Open and validate file contents
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open batch file: %w", err)
	}
	defer file.Close()

	// Check file info
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.Size() > MaxFileSize {
		cv.logger.SecurityEvent("batch_file_too_large", map[string]interface{}{
			"filename": filename,
			"size":     info.Size(),
		})
		return fmt.Errorf("batch file size %d exceeds maximum %d", info.Size(), MaxFileSize)
	}

	// Validate file contents
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		if lineNum > MaxLineCount {
			cv.logger.SecurityEvent("batch_file_too_many_lines", map[string]interface{}{
				"filename":   filename,
				"line_count": lineNum,
			})
			return fmt.Errorf("batch file exceeds maximum %d lines", MaxLineCount)
		}

		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Validate each command
		if err := cv.ValidateCommand(line); err != nil {
			return fmt.Errorf("invalid command at line %d: %w", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading batch file: %w", err)
	}

	cv.logger.SecurityEvent("batch_file_validated", map[string]interface{}{
		"filename":   filename,
		"line_count": lineNum,
	})

	return nil
}

// SubstituteVariablesSafe safely substitutes variables in a command
func (cv *CommandValidator) SubstituteVariablesSafe(line string, vars map[string]string) (string, error) {
	result := line

	// Use regex to find variable references
	varPattern := regexp.MustCompile(`\$\{?([A-Za-z0-9_]+)\}?`)

	// Find all matches
	matches := varPattern.FindAllStringSubmatch(line, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		varName := match[1]
		varValue, exists := vars[varName]

		if !exists {
			return "", fmt.Errorf("undefined variable: %s", varName)
		}

		// Validate the variable value before substitution
		if err := cv.validateVariableValue(varValue); err != nil {
			cv.logger.SecurityEvent("invalid_variable_value", map[string]interface{}{
				"variable": varName,
				"error":    err.Error(),
			})
			return "", fmt.Errorf("invalid variable value for %s: %w", varName, err)
		}

		// Replace the variable
		result = strings.Replace(result, match[0], varValue, 1)
	}

	// Validate the final command after substitution
	if err := cv.ValidateCommand(result); err != nil {
		return "", fmt.Errorf("command invalid after variable substitution: %w", err)
	}

	return result, nil
}

// validateVariableValue validates a variable value
func (cv *CommandValidator) validateVariableValue(value string) error {
	// Check length
	if len(value) > MaxLineLength {
		return fmt.Errorf("variable value too long")
	}

	// Check for shell metacharacters
	if err := cv.checkShellMetacharacters(value); err != nil {
		return err
	}

	return nil
}

// ValidateScriptFile validates a script file with variable support
func (cv *CommandValidator) ValidateScriptFile(filename string, pv *PathValidator) error {
	// First validate as a batch file
	if err := cv.ValidateBatchFile(filename, pv); err != nil {
		return err
	}

	// Additional validation for SET commands
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open script file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	varNames := make(map[string]bool)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for SET commands
		if strings.HasPrefix(strings.ToUpper(line), "SET ") {
			// Validate SET syntax
			parts := strings.SplitN(line[4:], "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid SET syntax at line %d", lineNum)
			}

			varName := strings.TrimSpace(parts[0])
			varValue := strings.TrimSpace(parts[1])

			// Validate variable name (alphanumeric and underscore only)
			if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(varName) {
				return fmt.Errorf("invalid variable name at line %d: %s", lineNum, varName)
			}

			// Validate variable value
			if err := cv.validateVariableValue(varValue); err != nil {
				return fmt.Errorf("invalid variable value at line %d: %w", lineNum, err)
			}

			varNames[varName] = true
		}
	}

	cv.logger.SecurityEvent("script_file_validated", map[string]interface{}{
		"filename":       filename,
		"line_count":     lineNum,
		"variable_count": len(varNames),
	})

	return nil
}

// ExecuteWithContext executes a function with timeout and cancellation
func (cv *CommandValidator) ExecuteWithContext(ctx context.Context, fn func() error) error {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, CommandExecutionTimeout)
	defer cancel()

	// Execute in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- fn()
	}()

	// Wait for completion or timeout
	select {
	case err := <-errChan:
		return err
	case <-timeoutCtx.Done():
		cv.logger.SecurityEvent("command_execution_timeout", map[string]interface{}{
			"timeout": CommandExecutionTimeout.String(),
		})
		return fmt.Errorf("command execution timeout after %s", CommandExecutionTimeout)
	}
}

// sanitizeCommand removes sensitive information from commands for logging
func sanitizeCommand(cmd string) string {
	// Remove potential passwords, keys, etc.
	patterns := []string{
		`--password\s+\S+`,
		`--key\s+\S+`,
		`--secret\s+\S+`,
		`--token\s+\S+`,
		`--private-key\s+\S+`,
	}

	result := cmd
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "--[REDACTED]")
	}

	return result
}
