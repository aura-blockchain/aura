// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 30 * time.Second
)

// ServerManager manages server lifecycle with graceful shutdown
type ServerManager struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	mu         sync.Mutex
	logger     Logger
	running    bool
	stopChan   chan struct{}
}

// NewServerManager creates a new server manager
func NewServerManager(logger Logger) *ServerManager {
	return &ServerManager{
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// RegisterGRPCServer registers a gRPC server
func (sm *ServerManager) RegisterGRPCServer(server *grpc.Server) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.grpcServer = server
}

// RegisterHTTPServer registers an HTTP server
func (sm *ServerManager) RegisterHTTPServer(server *http.Server) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.httpServer = server
}

// Start marks the servers as running
func (sm *ServerManager) Start() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.running = true

	sm.logger.SecurityEvent("servers_started", map[string]interface{}{
		"grpc_registered": sm.grpcServer != nil,
		"http_registered": sm.httpServer != nil,
	})
}

// Shutdown gracefully shuts down all servers
func (sm *ServerManager) Shutdown(ctx context.Context) error {
	sm.mu.Lock()
	if !sm.running {
		sm.mu.Unlock()
		return nil
	}
	sm.running = false
	sm.mu.Unlock()

	sm.logger.Info("Initiating graceful shutdown...")

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	// Shutdown errors channel
	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	// Shutdown gRPC server
	if sm.grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.logger.Info("Shutting down gRPC server...")

			// Create a channel to signal when GracefulStop completes
			done := make(chan struct{})
			go func() {
				sm.grpcServer.GracefulStop()
				close(done)
			}()

			// Wait for graceful stop or timeout
			select {
			case <-done:
				sm.logger.Info("gRPC server stopped gracefully")
			case <-shutdownCtx.Done():
				sm.logger.Warn("gRPC server shutdown timeout, forcing stop")
				sm.grpcServer.Stop()
			}
		}()
	}

	// Shutdown HTTP server
	if sm.httpServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.logger.Info("Shutting down HTTP server...")

			if err := sm.httpServer.Shutdown(shutdownCtx); err != nil {
				errChan <- fmt.Errorf("HTTP server shutdown error: %w", err)
				sm.logger.Error("HTTP server shutdown error: %v", err)
			} else {
				sm.logger.Info("HTTP server stopped gracefully")
			}
		}()
	}

	// Wait for all shutdowns to complete
	wg.Wait()
	close(errChan)

	// Collect any errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	sm.logger.SecurityEvent("servers_shutdown", map[string]interface{}{
		"graceful":    len(errors) == 0,
		"error_count": len(errors),
	})

	if len(errors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors", len(errors))
	}

	close(sm.stopChan)
	return nil
}

// IsRunning returns whether servers are running
func (sm *ServerManager) IsRunning() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.running
}

// Wait waits for the server manager to stop
func (sm *ServerManager) Wait() {
	<-sm.stopChan
}

// AddCleanupHandler adds a cleanup handler to run on shutdown
type CleanupHandler func() error

// ShutdownHandler manages cleanup handlers
type ShutdownHandler struct {
	handlers []CleanupHandler
	mu       sync.Mutex
	logger   Logger
}

// NewShutdownHandler creates a new shutdown handler
func NewShutdownHandler(logger Logger) *ShutdownHandler {
	return &ShutdownHandler{
		handlers: make([]CleanupHandler, 0),
		logger:   logger,
	}
}

// Register registers a cleanup handler
func (sh *ShutdownHandler) Register(handler CleanupHandler) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.handlers = append(sh.handlers, handler)
}

// Execute executes all cleanup handlers
func (sh *ShutdownHandler) Execute() error {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.logger.Info("Executing cleanup handlers...")

	var errors []error
	for i, handler := range sh.handlers {
		if err := handler(); err != nil {
			sh.logger.Error("Cleanup handler %d failed: %v", i, err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup completed with %d errors", len(errors))
	}

	sh.logger.Info("All cleanup handlers executed successfully")
	return nil
}

// HealthChecker performs health checks on the system
type HealthChecker struct {
	checks map[string]HealthCheck
	mu     sync.RWMutex
	logger Logger
}

// HealthCheck is a function that performs a health check
type HealthCheck func() error

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger Logger) *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
		logger: logger,
	}
}

// Register registers a health check
func (hc *HealthChecker) Register(name string, check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// Check performs all health checks
func (hc *HealthChecker) Check() (map[string]bool, error) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	results := make(map[string]bool)
	var failures []string

	for name, check := range hc.checks {
		if err := check(); err != nil {
			results[name] = false
			failures = append(failures, fmt.Sprintf("%s: %v", name, err))
			hc.logger.Warn("Health check failed: %s: %v", name, err)
		} else {
			results[name] = true
		}
	}

	if len(failures) > 0 {
		return results, fmt.Errorf("health checks failed: %v", failures)
	}

	return results, nil
}

// HTTPHealthHandler creates an HTTP handler for health checks
func (hc *HealthChecker) HTTPHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results, err := hc.Check()

		w.Header().Set("Content-Type", "application/json")

		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"unhealthy","checks":%v}`, formatResults(results))
		} else {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"healthy","checks":%v}`, formatResults(results))
		}
	}
}

func formatResults(results map[string]bool) string {
	// Simple JSON formatting
	str := "{"
	first := true
	for name, status := range results {
		if !first {
			str += ","
		}
		str += fmt.Sprintf(`"%s":%t`, name, status)
		first = false
	}
	str += "}"
	return str
}
