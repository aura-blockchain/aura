# AURA CLI Security Quick Reference

## For Developers

### Quick Security Checklist

When working with the AURA CLI, always:

1. ✅ **Validate paths** before file operations
2. ✅ **Validate commands** before execution
3. ✅ **Validate user input** before processing
4. ✅ **Use secure permissions** for files/directories
5. ✅ **Enable TLS** in production
6. ✅ **Log security events** for auditing

---

## Common Security Patterns

### 1. Path Validation

```go
// Always validate paths
logger := security.NewConsoleLogger()
validator := security.NewPathValidator(logger)

validPath, err := validator.ValidateAndCleanHomePath(userProvidedPath)
if err != nil {
    return fmt.Errorf("invalid path: %w", err)
}
// Use validPath safely
```

### 2. Command Validation

```go
// Validate commands before execution
logger := GetSecurityLogger()
cmdValidator := security.NewCommandValidator(logger)

if err := cmdValidator.ValidateCommand(cmdLine); err != nil {
    return fmt.Errorf("invalid command: %w", err)
}
// Execute command safely
```

### 3. Input Validation

```go
// Validate all user inputs
logger := GetSecurityLogger()
inputValidator := security.NewInputValidator(logger)

// Chain ID
if err := inputValidator.ValidateChainID(chainID); err != nil {
    return fmt.Errorf("invalid chain ID: %w", err)
}

// Moniker
if err := inputValidator.ValidateMoniker(moniker); err != nil {
    return fmt.Errorf("invalid moniker: %w", err)
}

// Network address
if err := inputValidator.ValidateNetworkAddress(address); err != nil {
    return fmt.Errorf("invalid address: %w", err)
}
```

### 4. Secure File Creation

```go
// Create directories with secure permissions
logger := GetSecurityLogger()

// For keys/data (0700)
err := security.CreateSecureDirectory(path, logger)

// Create files with secure permissions
data := []byte("sensitive content")
err := security.CreateSecureFile(path, data, logger)
```

### 5. Config Validation

```go
// Validate config keys and values
logger := GetSecurityLogger()
configValidator := security.NewConfigValidator(logger)

if err := configValidator.ValidateKey(key); err != nil {
    return fmt.Errorf("invalid config key: %w", err)
}

if err := configValidator.ValidateValue(key, value); err != nil {
    return fmt.Errorf("invalid config value: %w", err)
}
```

### 6. Security Logging

```go
// Log security events
logger := GetSecurityLogger()

logger.SecurityEvent("event_type", map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
})

// Log messages
logger.Info("Informational message")
logger.Warn("Warning message")
logger.Error("Error message")
```

---

## Security Constants

### File Permissions

```go
import "github.com/aequitas/aura/chain/cmd/aurad/cmd/security"

// Directories
security.SecureDirPerms    // 0700 - Private directories (keys, data)
security.ConfigDirPerms    // 0750 - Config directories
security.PublicDirPerms    // 0755 - Public directories

// Files
security.SecureFilePerms   // 0600 - Private files (keys)
security.ConfigFilePerms   // 0640 - Config files
security.PublicFilePerms   // 0644 - Public files
```

### Limits

```go
// Command validation
security.MaxFileSize      // 10MB - Max batch file size
security.MaxLineCount     // 10,000 - Max lines in batch file
security.MaxLineLength    // 4096 - Max line length
security.CommandExecutionTimeout // 5 minutes

// Path validation
security.MaxPathLength    // 4096 - Max path length

// Rate limiting
security.DefaultRateLimit // 10 req/sec
security.DefaultBurst     // 20 requests
```

---

## Testing Security Features

### Run Full Security Test Suite

```bash
cd /home/decri/blockchain-projects/aura/chain/cmd/aurad/cmd/security
./security_test_suite.sh
```

### Manual Testing

#### Test Path Traversal Protection

```bash
# Should fail
aurad init --home "../../../etc/passwd" test
aurad init --home "/etc" test
aurad init --home "~/.ssh" test
```

#### Test Command Injection Protection

```bash
# Create malicious batch file
cat > /tmp/test.txt << 'EOF'
query status; rm -rf /
EOF

# Should fail
aurad batch /tmp/test.txt
```

#### Test Input Validation

```bash
# Should fail
aurad init --chain-id "evil; rm -rf /" test
aurad init --moniker "test<script>" test
```

#### Test Config Validation

```bash
# Should fail
aurad config set "unknown.key" "value"
aurad config set "chain-id\n[evil]" "malicious"
```

---

## Security Event Types

Monitor these events in `$HOME/.aura/logs/security.log`:

### Critical Events (Immediate Action Required)

- `path_validation_failed` - Path traversal attempt
- `command_injection_detected` - Command injection attempt
- `shell_metacharacter_detected` - Dangerous shell characters
- `suspicious_command_detected` - Suspicious command pattern
- `toml_injection_detected` - Config injection attempt
- `tls_certificate_load_failed` - TLS setup failure

### Warning Events (Review Required)

- `rate_limit_exceeded` - Rate limit violation
- `config_key_not_allowed` - Unauthorized config access
- `suspicious_path_detected` - Suspicious path pattern
- `batch_file_too_large` - Oversized batch file
- `command_execution_timeout` - Command timeout

### Info Events (Monitoring)

- `path_validated` - Successful path validation
- `batch_execution_completed` - Batch completed
- `servers_started` - Servers started
- `servers_shutdown` - Graceful shutdown

---

## Common Mistakes to Avoid

### ❌ DON'T: Use paths without validation

```go
// WRONG - No validation
homeDir := viper.GetString("home")
os.MkdirAll(homeDir, 0755)
```

### ✅ DO: Always validate paths

```go
// CORRECT - Validate first
homeDir := GetHomeDir() // Already validated
os.MkdirAll(homeDir, security.SecureDirPerms)
```

---

### ❌ DON'T: Execute commands without validation

```go
// WRONG - Direct execution
cmd := exec.Command("sh", "-c", userInput)
cmd.Run()
```

### ✅ DO: Validate commands first

```go
// CORRECT - Validate before execution
cmdValidator := security.NewCommandValidator(logger)
if err := cmdValidator.ValidateCommand(userInput); err != nil {
    return err
}
// Now execute safely
```

---

### ❌ DON'T: Use world-readable permissions

```go
// WRONG - World readable
os.MkdirAll(keysDir, 0755)
os.WriteFile(keyFile, data, 0644)
```

### ✅ DO: Use secure permissions

```go
// CORRECT - Secure permissions
os.MkdirAll(keysDir, security.SecureDirPerms) // 0700
os.WriteFile(keyFile, data, security.SecureFilePerms) // 0600
os.Chmod(keyFile, security.SecureFilePerms) // Verify
```

---

### ❌ DON'T: Accept arbitrary config keys

```go
// WRONG - No validation
viper.Set(userKey, userValue)
```

### ✅ DO: Validate config operations

```go
// CORRECT - Validate first
configValidator := security.NewConfigValidator(logger)
if err := configValidator.ValidateValue(userKey, userValue); err != nil {
    return err
}
viper.Set(userKey, userValue)
```

---

## TLS Configuration

### Development (Optional)

```go
// TLS optional in development
tlsConfig := security.NewTLSConfig(homeDir, logger)
tlsCfg, err := tlsConfig.LoadTLSConfig()
if err != nil {
    // Fall back to insecure mode
    logger.Warn("Running without TLS")
}
```

### Production (Required)

```bash
# Generate certificates
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

# Place in TLS directory
mkdir -p ~/.aura/config/tls
mv server.crt server.key ~/.aura/config/tls/

# Enable mutual TLS (optional)
tlsConfig.EnableMutualTLS()
```

---

## Rate Limiting Configuration

### Default Settings

```go
// 10 requests/second, burst of 20
rateLimiter := security.NewDefaultRateLimiter(logger)
```

### Custom Settings

```go
// 100 requests/second, burst of 200
rateLimiter := security.NewRateLimiter(100, 200, logger)
```

### Apply to HTTP Server

```go
handler := rateLimiter.RateLimitMiddleware(mux)
```

---

## Graceful Shutdown

### Basic Setup

```go
// Create server manager
serverMgr := security.NewServerManager(logger)

// Register servers
serverMgr.RegisterGRPCServer(grpcServer)
serverMgr.RegisterHTTPServer(httpServer)

// Start servers
serverMgr.Start()

// On shutdown signal
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := serverMgr.Shutdown(ctx); err != nil {
    logger.Error("Shutdown failed: %v", err)
}
```

### With Cleanup Handlers

```go
// Add cleanup handlers
shutdownHandler := security.NewShutdownHandler(logger)
shutdownHandler.Register(func() error {
    // Cleanup logic
    return nil
})

// Execute on shutdown
shutdownHandler.Execute()
```

---

## Health Checks

### Setup Health Checker

```go
healthChecker := security.NewHealthChecker(logger)

// Register checks
healthChecker.Register("database", func() error {
    return checkDatabaseConnection()
})

healthChecker.Register("disk_space", func() error {
    return checkDiskSpace()
})

// Add HTTP endpoint
mux.Handle("/health", healthChecker.HTTPHealthHandler())
```

---

## Production Deployment

### Pre-Flight Checklist

```bash
# 1. Generate TLS certificates
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

# 2. Set secure permissions
chmod 700 ~/.aura/keys ~/.aura/data
chmod 750 ~/.aura/config
chmod 600 ~/.aura/keys/*
chmod 640 ~/.aura/config/*

# 3. Verify security features
./security_test_suite.sh

# 4. Check security logs
tail -f ~/.aura/logs/security.log

# 5. Test graceful shutdown
aurad start &
PID=$!
sleep 5
kill -SIGTERM $PID
wait $PID
```

---

## Troubleshooting

### Issue: Path validation failing

```bash
# Check if path is under allowed directories
echo $HOME
ls -la ~/.aura

# Verify no symlinks pointing outside
ls -laR ~/.aura | grep "^l"
```

### Issue: Commands being blocked

```bash
# Check command whitelist
# Allowed: query, tx, status, keys, config, version, help, completion

# Verify no shell metacharacters
# Blocked: ; | & $ ` < > ( ) { } \
```

### Issue: TLS not working

```bash
# Check certificate files
ls -la ~/.aura/config/tls/
# Should contain: server.crt, server.key

# Verify permissions
stat -c "%a" ~/.aura/config/tls/server.key
# Should be 600

# Check certificate validity
openssl x509 -in ~/.aura/config/tls/server.crt -text -noout
```

### Issue: Rate limiting too aggressive

```go
// Increase rate limit
rateLimiter := security.NewRateLimiter(100, 200, logger) // 100 req/s

// Or disable for development
// (Don't do this in production!)
```

---

## Getting Help

### Documentation

- Main README: `/chain/cmd/aurad/cmd/security/README.md`
- Implementation Report: `/CLI_SECURITY_IMPLEMENTATION_REPORT.md`
- This Quick Reference: `/CLI_SECURITY_QUICK_REFERENCE.md`

### Testing

- Security Test Suite: `/chain/cmd/aurad/cmd/security/security_test_suite.sh`

### Monitoring

- Security Logs: `~/.aura/logs/security.log`
- Format: JSON structured logs
- Monitor critical events daily

---

## Security Principles

1. **Defense in Depth**: Multiple layers of security
2. **Fail Securely**: Fail closed, not open
3. **Least Privilege**: Minimal permissions by default
4. **Input Validation**: Validate all user input
5. **Secure Defaults**: Security enabled by default
6. **Audit Logging**: Log all security events
7. **Graceful Degradation**: Handle errors securely

---

## Version

**Security Module Version**: 1.0.0
**Last Updated**: 2025-01-26
**Status**: Production Ready

---

## Contact

For security issues or questions:
- Review documentation in `/chain/cmd/aurad/cmd/security/`
- Check security logs in `~/.aura/logs/security.log`
- Run security test suite to verify protections
