// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// SecurityConfig represents the overall security configuration
type SecurityConfig struct {
	QueryRateLimiting QueryRateLimitingConfig `json:"query_rate_limiting"`
}

// QueryRateLimitingConfig configures rate limiting for gRPC queries
type QueryRateLimitingConfig struct {
	Enabled          bool            `json:"enabled"`
	ExpensiveRate    float64         `json:"expensive_rate"`
	ExpensiveBurst   int             `json:"expensive_burst"`
	NormalRate       float64         `json:"normal_rate"`
	NormalBurst      int             `json:"normal_burst"`
	ExpensiveQueries map[string]bool `json:"expensive_queries"`
}

// ToExpensiveQueryConfig converts to ExpensiveQueryConfig for the QueryRateLimiter
func (c *QueryRateLimitingConfig) ToExpensiveQueryConfig() *ExpensiveQueryConfig {
	return &ExpensiveQueryConfig{
		ExpensiveRate:    c.ExpensiveRate,
		ExpensiveBurst:   c.ExpensiveBurst,
		NormalRate:       c.NormalRate,
		NormalBurst:      c.NormalBurst,
		ExpensiveQueries: c.ExpensiveQueries,
	}
}

// LoadSecurityConfig loads security configuration from the config directory
func LoadSecurityConfig(homeDir string) (*SecurityConfig, error) {
	configPath := filepath.Join(homeDir, "config", "security.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config SecurityConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		QueryRateLimiting: QueryRateLimitingConfig{
			Enabled:        true,
			ExpensiveRate:  2.0,
			ExpensiveBurst: 5,
			NormalRate:     10.0,
			NormalBurst:    20,
			ExpensiveQueries: map[string]bool{
				// DEX queries
				"/aura.dex.v1beta1.Query/Orderbook":      true,
				"/aura.dex.v1beta1.Query/AllPools":       true,
				"/aura.dex.v1beta1.Query/UserOrders":     true,
				"/aura.dex.v1beta1.Query/SupportedCoins": true,
				// Privacy queries
				"/aura.privacy.v1beta1.Query/VerifyZKProof": true,
				"/aura.privacy.v1beta1.Query/MixingPools":   true,
				// VCRegistry queries
				"/aura.vcregistry.v1beta1.Query/ResolveDID":          true,
				"/aura.vcregistry.v1beta1.Query/ListUserVCs":         true,
				"/aura.vcregistry.v1beta1.Query/BatchVCStatus":       true,
				"/aura.vcregistry.v1beta1.Query/GetRevocationList":   true,
				"/aura.vcregistry.v1beta1.Query/VerifyPresentation":  true,
				"/aura.vcregistry.v1beta1.Query/GetDIDDocument":      true,
				"/aura.vcregistry.v1beta1.Query/ListDIDsByController": true,
				// Compliance queries
				"/aura.compliance.v1beta1.Query/GetAddressStatus":        true,
				"/aura.compliance.v1beta1.Query/GetComplianceScore":      true,
				"/aura.compliance.v1beta1.Query/ListRestrictedAddresses": true,
			},
		},
	}
}

// ExpensiveQueryConfig configures the query rate limiter
type ExpensiveQueryConfig struct {
	ExpensiveRate    float64
	ExpensiveBurst   int
	NormalRate       float64
	NormalBurst      int
	ExpensiveQueries map[string]bool
}

// DefaultExpensiveQueryConfig returns the default expensive query configuration
func DefaultExpensiveQueryConfig() *ExpensiveQueryConfig {
	cfg := DefaultSecurityConfig()
	return cfg.QueryRateLimiting.ToExpensiveQueryConfig()
}

// QueryRateLimiter limits the rate of gRPC queries per address
type QueryRateLimiter struct {
	config       *ExpensiveQueryConfig
	logger       Logger
	limiters     map[string]*queryLimiterEntry
	mu           sync.RWMutex
	stopChan     chan struct{}
	stats        QueryRateLimiterStats
	statsMu      sync.RWMutex
}

type queryLimiterEntry struct {
	expensiveTokens float64
	normalTokens    float64
	lastUpdate      time.Time
	lastSeen        time.Time
}

// QueryRateLimiterStats tracks rate limiter statistics
type QueryRateLimiterStats struct {
	ExpensiveQueriesBlocked int64
	NormalQueriesBlocked    int64
	TotalQueries            int64
}

// NewQueryRateLimiter creates a new query rate limiter
func NewQueryRateLimiter(config *ExpensiveQueryConfig, logger Logger) *QueryRateLimiter {
	qrl := &QueryRateLimiter{
		config:   config,
		logger:   logger,
		limiters: make(map[string]*queryLimiterEntry),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go qrl.cleanupRoutine()

	return qrl
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for rate limiting
func (qrl *QueryRateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		address := qrl.extractAddress(ctx)
		method := info.FullMethod

		if !qrl.allow(address, method) {
			qrl.logger.SecurityEvent("query_rate_limit_exceeded", map[string]interface{}{
				"address":    address,
				"method":     method,
				"is_expensive": qrl.isExpensive(method),
			})
			return nil, grpc.Errorf(429, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor for rate limiting
func (qrl *QueryRateLimiter) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		address := qrl.extractAddress(ctx)
		method := info.FullMethod

		if !qrl.allow(address, method) {
			qrl.logger.SecurityEvent("query_rate_limit_exceeded", map[string]interface{}{
				"address":    address,
				"method":     method,
				"is_expensive": qrl.isExpensive(method),
			})
			return grpc.Errorf(429, "rate limit exceeded")
		}

		return handler(srv, ss)
	}
}

// extractAddress extracts the client address from the gRPC context
func (qrl *QueryRateLimiter) extractAddress(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "unknown"
}

// isExpensive checks if a method is considered expensive
func (qrl *QueryRateLimiter) isExpensive(method string) bool {
	return qrl.config.ExpensiveQueries[method]
}

// allow checks if a request is allowed based on rate limiting
func (qrl *QueryRateLimiter) allow(address, method string) bool {
	qrl.mu.Lock()
	defer qrl.mu.Unlock()

	qrl.statsMu.Lock()
	qrl.stats.TotalQueries++
	qrl.statsMu.Unlock()

	now := time.Now()
	entry, exists := qrl.limiters[address]

	if !exists {
		entry = &queryLimiterEntry{
			expensiveTokens: float64(qrl.config.ExpensiveBurst),
			normalTokens:    float64(qrl.config.NormalBurst),
			lastUpdate:      now,
			lastSeen:        now,
		}
		qrl.limiters[address] = entry
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(entry.lastUpdate).Seconds()
	entry.expensiveTokens = min(float64(qrl.config.ExpensiveBurst), entry.expensiveTokens+elapsed*qrl.config.ExpensiveRate)
	entry.normalTokens = min(float64(qrl.config.NormalBurst), entry.normalTokens+elapsed*qrl.config.NormalRate)
	entry.lastUpdate = now
	entry.lastSeen = now

	// Check rate limit
	if qrl.isExpensive(method) {
		if entry.expensiveTokens < 1 {
			qrl.statsMu.Lock()
			qrl.stats.ExpensiveQueriesBlocked++
			qrl.statsMu.Unlock()
			return false
		}
		entry.expensiveTokens--
	} else {
		if entry.normalTokens < 1 {
			qrl.statsMu.Lock()
			qrl.stats.NormalQueriesBlocked++
			qrl.statsMu.Unlock()
			return false
		}
		entry.normalTokens--
	}

	return true
}

// cleanupRoutine periodically removes old limiters
func (qrl *QueryRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			qrl.cleanup()
		case <-qrl.stopChan:
			return
		}
	}
}

// cleanup removes limiters that haven't been used in 15 minutes
func (qrl *QueryRateLimiter) cleanup() {
	qrl.mu.Lock()
	defer qrl.mu.Unlock()

	now := time.Now()
	removed := 0

	for address, entry := range qrl.limiters {
		if now.Sub(entry.lastSeen) > 15*time.Minute {
			delete(qrl.limiters, address)
			removed++
		}
	}

	if removed > 0 {
		qrl.logger.SecurityEvent("query_rate_limiter_cleanup", map[string]interface{}{
			"removed_count": removed,
			"active_count":  len(qrl.limiters),
		})
	}
}

// GetStats returns current statistics
func (qrl *QueryRateLimiter) GetStats() QueryRateLimiterStats {
	qrl.statsMu.RLock()
	defer qrl.statsMu.RUnlock()
	return qrl.stats
}

// Stop stops the query rate limiter
func (qrl *QueryRateLimiter) Stop() {
	close(qrl.stopChan)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
