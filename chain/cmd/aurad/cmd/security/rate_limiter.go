package security

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// DefaultRateLimit is the default rate limit (requests per second)
	DefaultRateLimit = 10

	// DefaultBurst is the default burst size
	DefaultBurst = 20

	// CleanupInterval is how often to clean up old limiters
	CleanupInterval = 5 * time.Minute

	// MaxIdleTime is how long to keep idle limiters
	MaxIdleTime = 10 * time.Minute
)

// RateLimiter manages rate limiting for API endpoints
type RateLimiter struct {
	limiters map[string]*rateLimiterEntry
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	logger   Logger
	stopChan chan struct{}
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rateLimit float64, burst int, logger Logger) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rateLimiterEntry),
		rate:     rate.Limit(rateLimit),
		burst:    burst,
		logger:   logger,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// NewDefaultRateLimiter creates a rate limiter with default settings
func NewDefaultRateLimiter(logger Logger) *RateLimiter {
	return NewRateLimiter(DefaultRateLimit, DefaultBurst, logger)
}

// getLimiter gets or creates a rate limiter for an IP address
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = &rateLimiterEntry{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	entry.lastSeen = time.Now()
	return entry.limiter
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	limiter := rl.getLimiter(ip)
	return limiter.Allow()
}

// Wait waits until the request is allowed or context is cancelled
func (rl *RateLimiter) Wait(ip string) error {
	limiter := rl.getLimiter(ip)
	return limiter.Wait(context.TODO())
}

// cleanupRoutine periodically removes old limiters
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanup removes limiters that haven't been used recently
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	removed := 0

	for ip, entry := range rl.limiters {
		if now.Sub(entry.lastSeen) > MaxIdleTime {
			delete(rl.limiters, ip)
			removed++
		}
	}

	if removed > 0 {
		rl.logger.SecurityEvent("rate_limiter_cleanup", map[string]interface{}{
			"removed_count": removed,
			"active_count":  len(rl.limiters),
		})
	}
}

// Stop stops the rate limiter cleanup routine
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"active_limiters": len(rl.limiters),
		"rate_limit":      float64(rl.rate),
		"burst":           rl.burst,
	}
}

// RateLimitMiddleware creates an HTTP middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract IP address
		ip := extractIP(r)

		// Check rate limit
		if !rl.Allow(ip) {
			rl.logger.SecurityEvent("rate_limit_exceeded", map[string]interface{}{
				"ip":     ip,
				"path":   r.URL.Path,
				"method": r.Method,
			})

			// Return 429 Too Many Requests
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			// Ignore write error - client may have disconnected
			_, _ = w.Write([]byte(`{"error":"rate limit exceeded","retry_after":60}`))
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// extractIP extracts the real IP address from the request
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For header (for requests behind proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if ip := parseXForwardedFor(xff); ip != "" {
			return ip
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// parseXForwardedFor parses the X-Forwarded-For header
func parseXForwardedFor(xff string) string {
	// Take the first IP from the comma-separated list
	if idx := len(xff); idx > 0 {
		for i, c := range xff {
			if c == ',' {
				idx = i
				break
			}
		}
		ip := xff[:idx]
		// Trim spaces
		ip = trimSpace(ip)
		return ip
	}
	return ""
}

// trimSpace removes leading and trailing spaces
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

// SecurityHeadersMiddleware adds security headers to HTTP responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Force HTTPS (only if not localhost)
		if r.Host != "localhost" && !isLocalhost(r.Host) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (formerly Feature-Policy)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// isLocalhost checks if the host is localhost
func isLocalhost(host string) bool {
	// Extract hostname without port
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}

// IPWhitelist creates a middleware that only allows requests from whitelisted IPs
type IPWhitelist struct {
	allowedIPs map[string]bool
	logger     Logger
}

// NewIPWhitelist creates a new IP whitelist
func NewIPWhitelist(allowedIPs []string, logger Logger) *IPWhitelist {
	ipMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		ipMap[ip] = true
	}

	return &IPWhitelist{
		allowedIPs: ipMap,
		logger:     logger,
	}
}

// Middleware creates a middleware that enforces the IP whitelist
func (ipw *IPWhitelist) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		// Check if IP is whitelisted
		if !ipw.allowedIPs[ip] {
			ipw.logger.SecurityEvent("ip_not_whitelisted", map[string]interface{}{
				"ip":     ip,
				"path":   r.URL.Path,
				"method": r.Method,
			})

			w.WriteHeader(http.StatusForbidden)
			// Ignore write error - client may have disconnected
			_, _ = w.Write([]byte(`{"error":"access denied"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestLogger creates a middleware that logs all requests
type RequestLogger struct {
	logger Logger
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(logger Logger) *RequestLogger {
	return &RequestLogger{logger: logger}
}

// Middleware creates a middleware that logs requests
func (rl *RequestLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapper, r)

		// Log request
		duration := time.Since(start)
		rl.logger.SecurityEvent("http_request", map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"ip":          extractIP(r),
			"status_code": wrapper.statusCode,
			"duration_ms": duration.Milliseconds(),
			"user_agent":  r.UserAgent(),
		})
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// TimeoutMiddleware creates a middleware that enforces request timeouts
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, fmt.Sprintf(`{"error":"request timeout after %s"}`, timeout))
	}
}
