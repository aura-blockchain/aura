// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug    LogLevel = "debug"
	LogLevelInfo     LogLevel = "info"
	LogLevelWarn     LogLevel = "warn"
	LogLevelError    LogLevel = "error"
	LogLevelSecurity LogLevel = "security"
)

// LogEntry represents a structured log entry meeting blockchain community standards
type LogEntry struct {
	Timestamp     string                 `json:"timestamp"`                // ISO 8601 in UTC
	Level         LogLevel               `json:"level"`                    // Log level
	Message       string                 `json:"message,omitempty"`        // Human-readable message
	Type          string                 `json:"type,omitempty"`           // Event type for security events
	Service       string                 `json:"service"`                  // Service name
	Module        string                 `json:"module,omitempty"`         // Module/component name
	CorrelationID string                 `json:"correlation_id,omitempty"` // Request tracing ID
	TraceID       string                 `json:"trace_id,omitempty"`       // Distributed trace ID
	SpanID        string                 `json:"span_id,omitempty"`        // Span ID for tracing
	ChainID       string                 `json:"chain_id,omitempty"`       // Blockchain chain ID
	BlockHeight   int64                  `json:"block_height,omitempty"`   // Current block height
	ErrorCode     string                 `json:"error_code,omitempty"`     // Structured error code
	ErrorMessage  string                 `json:"error_message,omitempty"`  // Error details
	StackTrace    string                 `json:"stack_trace,omitempty"`    // Stack trace for errors
	Data          map[string]interface{} `json:"data,omitempty"`           // Additional contextual data
	Host          string                 `json:"host,omitempty"`           // Hostname
	Environment   string                 `json:"environment,omitempty"`    // Environment (testnet/mainnet)
	Version       string                 `json:"version,omitempty"`        // Application version
}

// Logger interface for security event logging
type Logger interface {
	SecurityEvent(eventType string, data map[string]interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	// Enhanced methods for structured logging
	WithCorrelationID(id string) Logger
	WithModule(module string) Logger
	WithBlockHeight(height int64) Logger
	ErrorWithCode(code string, msg string, err error)
}

// SecurityLogger logs security events to a file with structured JSON format
type SecurityLogger struct {
	logFile       *os.File
	mu            sync.Mutex
	logPath       string
	rotation      bool
	serviceName   string
	chainID       string
	environment   string
	version       string
	host          string
	correlationID string
	module        string
	blockHeight   int64
}

// NewSecurityLogger creates a new security logger with enhanced structured logging
func NewSecurityLogger(homeDir string, enableRotation bool) (*SecurityLogger, error) {
	logDir := filepath.Join(homeDir, "logs")
	if err := os.MkdirAll(logDir, SecureDirPerms); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "security.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, SecureFilePerms)
	if err != nil {
		return nil, fmt.Errorf("failed to open security log: %w", err)
	}

	hostname, _ := os.Hostname()

	return &SecurityLogger{
		logFile:     logFile,
		logPath:     logPath,
		rotation:    enableRotation,
		serviceName: "aurad",
		chainID:     os.Getenv("CHAIN_ID"),
		environment: getEnvironment(),
		version:     os.Getenv("AURAD_VERSION"),
		host:        hostname,
	}, nil
}

// NewSecurityLoggerWithConfig creates a logger with explicit configuration
func NewSecurityLoggerWithConfig(homeDir string, serviceName, chainID, environment, version string) (*SecurityLogger, error) {
	logger, err := NewSecurityLogger(homeDir, true)
	if err != nil {
		return nil, err
	}
	logger.serviceName = serviceName
	logger.chainID = chainID
	logger.environment = environment
	logger.version = version
	return logger, nil
}

// getEnvironment determines the environment from env vars or defaults to "development"
func getEnvironment() string {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if os.Getenv("TESTNET") == "true" {
		return "testnet"
	}
	return "development"
}

// generateCorrelationID creates a unique correlation ID
func generateCorrelationID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// SecurityEvent logs a security event with full structured context
func (sl *SecurityLogger) SecurityEvent(eventType string, data map[string]interface{}) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         LogLevelSecurity,
		Type:          eventType,
		Service:       sl.serviceName,
		Module:        sl.module,
		CorrelationID: sl.correlationID,
		ChainID:       sl.chainID,
		BlockHeight:   sl.blockHeight,
		Environment:   sl.environment,
		Version:       sl.version,
		Host:          sl.host,
		Data:          data,
	}

	// Generate correlation ID if not set
	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal security event: %v\n", err)
		return
	}

	if _, err := sl.logFile.Write(append(jsonData, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write security event: %v\n", err)
	}

	// Also log to stderr for critical events
	if isCriticalEvent(eventType) {
		fmt.Fprintf(os.Stderr, "[SECURITY] %s: %s\n", eventType, string(jsonData))
	}
}

// Info logs an info message with structured context
func (sl *SecurityLogger) Info(msg string, args ...interface{}) {
	sl.logStructured(LogLevelInfo, msg, args...)
}

// Warn logs a warning message with structured context
func (sl *SecurityLogger) Warn(msg string, args ...interface{}) {
	sl.logStructured(LogLevelWarn, msg, args...)
}

// Error logs an error message with structured context
func (sl *SecurityLogger) Error(msg string, args ...interface{}) {
	sl.logStructured(LogLevelError, msg, args...)
}

// ErrorWithCode logs an error with a structured error code for easier filtering
func (sl *SecurityLogger) ErrorWithCode(code string, msg string, err error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         LogLevelError,
		Message:       msg,
		Service:       sl.serviceName,
		Module:        sl.module,
		CorrelationID: sl.correlationID,
		ChainID:       sl.chainID,
		BlockHeight:   sl.blockHeight,
		Environment:   sl.environment,
		Version:       sl.version,
		Host:          sl.host,
		ErrorCode:     code,
	}

	if err != nil {
		entry.ErrorMessage = err.Error()
		entry.StackTrace = getStackTrace()
	}

	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}

	jsonData, marshalErr := json.Marshal(entry)
	if marshalErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log: %v\n", marshalErr)
		return
	}

	if _, writeErr := sl.logFile.Write(append(jsonData, '\n')); writeErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", writeErr)
	}
}

// WithCorrelationID returns a new logger with the correlation ID set
func (sl *SecurityLogger) WithCorrelationID(id string) Logger {
	return &SecurityLogger{
		logFile:       sl.logFile,
		logPath:       sl.logPath,
		rotation:      sl.rotation,
		serviceName:   sl.serviceName,
		chainID:       sl.chainID,
		environment:   sl.environment,
		version:       sl.version,
		host:          sl.host,
		correlationID: id,
		module:        sl.module,
		blockHeight:   sl.blockHeight,
	}
}

// WithModule returns a new logger with the module name set
func (sl *SecurityLogger) WithModule(module string) Logger {
	return &SecurityLogger{
		logFile:       sl.logFile,
		logPath:       sl.logPath,
		rotation:      sl.rotation,
		serviceName:   sl.serviceName,
		chainID:       sl.chainID,
		environment:   sl.environment,
		version:       sl.version,
		host:          sl.host,
		correlationID: sl.correlationID,
		module:        module,
		blockHeight:   sl.blockHeight,
	}
}

// WithBlockHeight returns a new logger with the block height set
func (sl *SecurityLogger) WithBlockHeight(height int64) Logger {
	return &SecurityLogger{
		logFile:       sl.logFile,
		logPath:       sl.logPath,
		rotation:      sl.rotation,
		serviceName:   sl.serviceName,
		chainID:       sl.chainID,
		environment:   sl.environment,
		version:       sl.version,
		host:          sl.host,
		correlationID: sl.correlationID,
		module:        sl.module,
		blockHeight:   height,
	}
}

func (sl *SecurityLogger) logStructured(level LogLevel, msg string, args ...interface{}) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	formattedMsg := fmt.Sprintf(msg, args...)
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         level,
		Message:       formattedMsg,
		Service:       sl.serviceName,
		Module:        sl.module,
		CorrelationID: sl.correlationID,
		ChainID:       sl.chainID,
		BlockHeight:   sl.blockHeight,
		Environment:   sl.environment,
		Version:       sl.version,
		Host:          sl.host,
	}

	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log: %v\n", err)
		return
	}

	if _, err := sl.logFile.Write(append(jsonData, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log: %v\n", err)
	}
}

// Close closes the security logger
func (sl *SecurityLogger) Close() error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if sl.logFile != nil {
		return sl.logFile.Close()
	}
	return nil
}

// isCriticalEvent determines if an event is critical
func isCriticalEvent(eventType string) bool {
	criticalEvents := []string{
		"path_validation_failed",
		"command_injection_detected",
		"shell_metacharacter_detected",
		"suspicious_command_detected",
		"suspicious_path_detected",
		"rate_limit_exceeded",
		"ip_not_whitelisted",
		"toml_injection_detected",
		"tls_certificate_load_failed",
		"ca_certificate_missing",
	}

	for _, critical := range criticalEvents {
		if eventType == critical {
			return true
		}
	}
	return false
}

// ConsoleLogger is a simple logger that writes to stdout with structured JSON
type ConsoleLogger struct {
	correlationID string
	module        string
	blockHeight   int64
}

// NewConsoleLogger creates a new console logger
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// SecurityEvent logs a security event to console
func (cl *ConsoleLogger) SecurityEvent(eventType string, data map[string]interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         LogLevelSecurity,
		Type:          eventType,
		Service:       "aurad",
		Module:        cl.module,
		CorrelationID: cl.correlationID,
		BlockHeight:   cl.blockHeight,
		Data:          data,
	}
	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}
	jsonData, _ := json.Marshal(entry)
	fmt.Printf("%s\n", string(jsonData))
}

// Info logs an info message
func (cl *ConsoleLogger) Info(msg string, args ...interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         LogLevelInfo,
		Message:       fmt.Sprintf(msg, args...),
		Service:       "aurad",
		Module:        cl.module,
		CorrelationID: cl.correlationID,
		BlockHeight:   cl.blockHeight,
	}
	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}
	jsonData, _ := json.Marshal(entry)
	fmt.Printf("%s\n", string(jsonData))
}

// Warn logs a warning message
func (cl *ConsoleLogger) Warn(msg string, args ...interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         LogLevelWarn,
		Message:       fmt.Sprintf(msg, args...),
		Service:       "aurad",
		Module:        cl.module,
		CorrelationID: cl.correlationID,
		BlockHeight:   cl.blockHeight,
	}
	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}
	jsonData, _ := json.Marshal(entry)
	fmt.Printf("%s\n", string(jsonData))
}

// Error logs an error message
func (cl *ConsoleLogger) Error(msg string, args ...interface{}) {
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         LogLevelError,
		Message:       fmt.Sprintf(msg, args...),
		Service:       "aurad",
		Module:        cl.module,
		CorrelationID: cl.correlationID,
		BlockHeight:   cl.blockHeight,
	}
	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}
	jsonData, _ := json.Marshal(entry)
	fmt.Fprintf(os.Stderr, "%s\n", string(jsonData))
}

// ErrorWithCode logs an error with a structured error code
func (cl *ConsoleLogger) ErrorWithCode(code string, msg string, err error) {
	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Level:         LogLevelError,
		Message:       msg,
		Service:       "aurad",
		Module:        cl.module,
		CorrelationID: cl.correlationID,
		BlockHeight:   cl.blockHeight,
		ErrorCode:     code,
	}
	if err != nil {
		entry.ErrorMessage = err.Error()
	}
	if entry.CorrelationID == "" {
		entry.CorrelationID = generateCorrelationID()
	}
	jsonData, _ := json.Marshal(entry)
	fmt.Fprintf(os.Stderr, "%s\n", string(jsonData))
}

// WithCorrelationID returns a new logger with the correlation ID set
func (cl *ConsoleLogger) WithCorrelationID(id string) Logger {
	return &ConsoleLogger{
		correlationID: id,
		module:        cl.module,
		blockHeight:   cl.blockHeight,
	}
}

// WithModule returns a new logger with the module name set
func (cl *ConsoleLogger) WithModule(module string) Logger {
	return &ConsoleLogger{
		correlationID: cl.correlationID,
		module:        module,
		blockHeight:   cl.blockHeight,
	}
}

// WithBlockHeight returns a new logger with the block height set
func (cl *ConsoleLogger) WithBlockHeight(height int64) Logger {
	return &ConsoleLogger{
		correlationID: cl.correlationID,
		module:        cl.module,
		blockHeight:   height,
	}
}
