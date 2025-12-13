package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger interface for security event logging
type Logger interface {
	SecurityEvent(eventType string, data map[string]interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// SecurityLogger logs security events to a file
type SecurityLogger struct {
	logFile  *os.File
	mu       sync.Mutex
	logPath  string
	rotation bool
}

// NewSecurityLogger creates a new security logger
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

	return &SecurityLogger{
		logFile:  logFile,
		logPath:  logPath,
		rotation: enableRotation,
	}, nil
}

// SecurityEvent logs a security event
func (sl *SecurityLogger) SecurityEvent(eventType string, data map[string]interface{}) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	event := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"type":      eventType,
		"level":     "security",
		"data":      data,
	}

	jsonData, err := json.Marshal(event)
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

// Info logs an info message
func (sl *SecurityLogger) Info(msg string, args ...interface{}) {
	sl.log("info", msg, args...)
}

// Warn logs a warning message
func (sl *SecurityLogger) Warn(msg string, args ...interface{}) {
	sl.log("warn", msg, args...)
}

// Error logs an error message
func (sl *SecurityLogger) Error(msg string, args ...interface{}) {
	sl.log("error", msg, args...)
}

func (sl *SecurityLogger) log(level, msg string, args ...interface{}) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	formattedMsg := fmt.Sprintf(msg, args...)
	event := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"message":   formattedMsg,
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log: %v\n", err)
		return
	}

	if _, err := sl.logFile.Write(append(jsonData, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write security log: %v\n", err)
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

// ConsoleLogger is a simple logger that writes to stdout
type ConsoleLogger struct{}

// NewConsoleLogger creates a new console logger
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// SecurityEvent logs a security event to console
func (cl *ConsoleLogger) SecurityEvent(eventType string, data map[string]interface{}) {
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"type":      eventType,
		"data":      data,
	}, "", "  ")
	fmt.Printf("[SECURITY] %s\n", string(jsonData))
}

// Info logs an info message
func (cl *ConsoleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] %s\n", fmt.Sprintf(msg, args...))
}

// Warn logs a warning message
func (cl *ConsoleLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] %s\n", fmt.Sprintf(msg, args...))
}

// Error logs an error message
func (cl *ConsoleLogger) Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] %s\n", fmt.Sprintf(msg, args...))
}
