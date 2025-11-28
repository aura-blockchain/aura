#!/bin/bash
# ============================================================================
# AURA Blockchain - Deployment Security Verification Script
# ============================================================================
# Comprehensive security verification for Docker deployments
# Checks all 14 critical security requirements
# ============================================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0
TOTAL_CHECKS=0

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; ((CHECKS_PASSED++)); ((TOTAL_CHECKS++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; ((CHECKS_FAILED++)); ((TOTAL_CHECKS++)); }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; ((CHECKS_WARNING++)); ((TOTAL_CHECKS++)); }

echo "============================================================================"
echo "AURA Blockchain - Deployment Security Verification"
echo "============================================================================"
echo ""

# ============================================================================
# CHECK 1: Hardcoded Credentials
# ============================================================================
log_info "Checking for hardcoded credentials..."

if grep -r "password.*=" docker-compose.yml 2>/dev/null | grep -qv "PASSWORD_FILE\|password:"; then
    log_fail "Hardcoded passwords found in docker-compose.yml"
else
    log_pass "No hardcoded passwords in docker-compose.yml"
fi

if [ -f docker-compose.secure.yml ]; then
    if grep -q "secrets:" docker-compose.secure.yml && ! grep -q "POSTGRES_PASSWORD: changeme" docker-compose.secure.yml; then
        log_pass "docker-compose.secure.yml uses Docker secrets"
    else
        log_fail "docker-compose.secure.yml not using Docker secrets properly"
    fi
else
    log_warn "docker-compose.secure.yml not found"
fi

# ============================================================================
# CHECK 2: Port Exposure
# ============================================================================
log_info "Checking port exposure configuration..."

if [ -f docker-compose.secure.yml ]; then
    # Check if sensitive ports are bound to localhost
    if grep -E "ports:" docker-compose.secure.yml -A 5 | grep -q "127.0.0.1:"; then
        log_pass "Sensitive ports bound to localhost"
    else
        log_fail "Ports not properly restricted to localhost"
    fi

    # Check for ports exposed to 0.0.0.0
    if grep -E '"[0-9]+:[0-9]+"' docker-compose.secure.yml | grep -qv "127.0.0.1"; then
        log_warn "Some ports may be exposed to 0.0.0.0"
    fi
fi

# ============================================================================
# CHECK 3: Resource Limits
# ============================================================================
log_info "Checking resource limits..."

if [ -f docker-compose.secure.yml ]; then
    if grep -q "limits:" docker-compose.secure.yml && grep -q "cpus:" docker-compose.secure.yml; then
        log_pass "CPU limits configured"
    else
        log_fail "CPU limits not configured"
    fi

    if grep -q "memory:" docker-compose.secure.yml; then
        log_pass "Memory limits configured"
    else
        log_fail "Memory limits not configured"
    fi

    if grep -q "pids:" docker-compose.secure.yml; then
        log_pass "PID limits configured"
    else
        log_warn "PID limits not configured"
    fi
fi

# ============================================================================
# CHECK 4: Volume Mounts
# ============================================================================
log_info "Checking volume mount security..."

if [ -f docker-compose.secure.yml ]; then
    if grep -q "/:/rootfs:ro" docker-compose.secure.yml; then
        log_fail "Dangerous volume mount found: entire host filesystem"
    else
        log_pass "No dangerous host filesystem mounts"
    fi

    if grep -q "read_only: true" docker-compose.secure.yml; then
        log_pass "Read-only volume mounts configured"
    else
        log_warn "Read-only volume mounts not extensively used"
    fi
fi

# ============================================================================
# CHECK 5: Image Pinning
# ============================================================================
log_info "Checking image pinning..."

if [ -f docker-compose.secure.yml ]; then
    if grep "image:" docker-compose.secure.yml | grep -q "@sha256:"; then
        log_pass "Images pinned with SHA256 digests"
    else
        log_fail "Images not pinned with SHA256 digests"
    fi

    if grep "image:" docker-compose.secure.yml | grep -q ":latest"; then
        log_fail "Using mutable 'latest' tag"
    else
        log_pass "Not using mutable 'latest' tag"
    fi
fi

# ============================================================================
# CHECK 6: Security Context
# ============================================================================
log_info "Checking security context configuration..."

if [ -f docker-compose.secure.yml ]; then
    if grep -q "no-new-privileges:true" docker-compose.secure.yml; then
        log_pass "no-new-privileges configured"
    else
        log_fail "no-new-privileges not configured"
    fi

    if grep -q "cap_drop:" docker-compose.secure.yml; then
        log_pass "Capability dropping configured"
    else
        log_fail "Capability dropping not configured"
    fi

    if grep -q "cap_drop:" docker-compose.secure.yml -A 1 | grep -q "ALL"; then
        log_pass "All capabilities dropped by default"
    else
        log_warn "Not dropping all capabilities by default"
    fi

    if grep -q "read_only: true" docker-compose.secure.yml; then
        log_pass "Read-only root filesystem configured"
    else
        log_warn "Read-only root filesystem not configured"
    fi
fi

# ============================================================================
# CHECK 7: TLS Configuration
# ============================================================================
log_info "Checking TLS configuration..."

if [ -d ./secrets/tls ]; then
    if [ -f ./secrets/tls/server.crt ] && [ -f ./secrets/tls/server.key ]; then
        log_pass "TLS certificates present"

        # Check certificate expiration
        if command -v openssl >/dev/null 2>&1; then
            expiry=$(openssl x509 -in ./secrets/tls/server.crt -noout -enddate 2>/dev/null | cut -d= -f2)
            if [ -n "$expiry" ]; then
                expiry_epoch=$(date -d "$expiry" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$expiry" +%s 2>/dev/null)
                now_epoch=$(date +%s)
                days_until_expiry=$(( ($expiry_epoch - $now_epoch) / 86400 ))

                if [ $days_until_expiry -lt 0 ]; then
                    log_fail "TLS certificate expired!"
                elif [ $days_until_expiry -lt 30 ]; then
                    log_warn "TLS certificate expires in $days_until_expiry days"
                else
                    log_pass "TLS certificate valid for $days_until_expiry days"
                fi
            fi
        fi
    else
        log_fail "TLS certificates not found"
    fi
else
    log_fail "TLS directory not found"
fi

if [ -f ./nginx/ssl-config.conf ]; then
    log_pass "Nginx SSL configuration present"
else
    log_warn "Nginx SSL configuration not found"
fi

# ============================================================================
# CHECK 8: Restart Policy
# ============================================================================
log_info "Checking restart policies..."

if [ -f docker-compose.secure.yml ]; then
    if grep -q "restart: on-failure" docker-compose.secure.yml; then
        log_pass "Secure restart policy configured (on-failure)"
    elif grep -q "restart: unless-stopped\|restart: always" docker-compose.secure.yml; then
        log_warn "Using auto-restart policy (may hide attacks)"
    else
        log_fail "No restart policy configured"
    fi
fi

# ============================================================================
# CHECK 9: Health Checks
# ============================================================================
log_info "Checking health check configuration..."

if [ -f docker-compose.secure.yml ]; then
    healthcheck_count=$(grep -c "healthcheck:" docker-compose.secure.yml || echo 0)
    if [ "$healthcheck_count" -gt 0 ]; then
        log_pass "Health checks configured ($healthcheck_count services)"

        # Check if health checks validate security
        if grep -A 10 "healthcheck:" docker-compose.secure.yml | grep -q "ps\|netstat\|wc -l"; then
            log_pass "Health checks include security validation"
        else
            log_warn "Health checks don't validate security controls"
        fi
    else
        log_fail "No health checks configured"
    fi
fi

# ============================================================================
# CHECK 10: Network Segmentation
# ============================================================================
log_info "Checking network segmentation..."

if [ -f docker-compose.secure.yml ]; then
    network_count=$(grep -c "^  [a-z_]*:$" docker-compose.secure.yml | grep -A 1 "networks:" || echo 0)

    if [ "$network_count" -gt 1 ]; then
        log_pass "Multiple networks configured (segmentation)"

        if grep -q "internal: true" docker-compose.secure.yml; then
            log_pass "Internal-only networks configured"
        else
            log_warn "No internal-only networks"
        fi
    else
        log_fail "Network segmentation not configured"
    fi
fi

# ============================================================================
# CHECK 11: Secrets Management
# ============================================================================
log_info "Checking secrets management..."

if [ -d ./secrets ]; then
    log_pass "Secrets directory exists"

    # Check directory permissions
    secrets_perms=$(stat -c %a ./secrets 2>/dev/null || stat -f %A ./secrets 2>/dev/null)
    if [ "$secrets_perms" == "700" ]; then
        log_pass "Secrets directory has correct permissions (700)"
    else
        log_warn "Secrets directory permissions: $secrets_perms (should be 700)"
    fi

    # Check if secrets are in gitignore
    if [ -f .gitignore ] && grep -q "secrets/" .gitignore; then
        log_pass "Secrets excluded from version control"
    else
        log_fail "Secrets not in .gitignore"
    fi

    # Check for required secret files
    required_secrets=("grafana_admin_password.txt" "postgres_password.txt" "redis_password.txt")
    for secret in "${required_secrets[@]}"; do
        if [ -f "./secrets/$secret" ]; then
            file_perms=$(stat -c %a "./secrets/$secret" 2>/dev/null || stat -f %A "./secrets/$secret" 2>/dev/null)
            if [ "$file_perms" == "600" ]; then
                log_pass "Secret $secret has correct permissions"
            else
                log_warn "Secret $secret permissions: $file_perms (should be 600)"
            fi
        else
            log_warn "Secret file missing: $secret"
        fi
    done
else
    log_fail "Secrets directory not found"
fi

# ============================================================================
# CHECK 12: Dockerfile Security
# ============================================================================
log_info "Checking Dockerfile security..."

if [ -f Dockerfile.secure ]; then
    log_pass "Secure Dockerfile present"

    if grep -q "FROM.*@sha256:" Dockerfile.secure; then
        log_pass "Base images pinned with digest"
    else
        log_fail "Base images not pinned with digest"
    fi

    if grep -q "USER.*aura\|USER [0-9]" Dockerfile.secure; then
        log_pass "Non-root user configured"
    else
        log_fail "Running as root user"
    fi

    if grep -q "govulncheck\|gosec" Dockerfile.secure; then
        log_pass "Security scanning integrated in build"
    else
        log_warn "Security scanning not in Dockerfile"
    fi
else
    log_warn "Dockerfile.secure not found"
fi

# ============================================================================
# CHECK 13: CI/CD Security
# ============================================================================
log_info "Checking CI/CD security configuration..."

if [ -f .github/workflows/ci.yml ]; then
    if grep -q "trivy\|gosec\|govulncheck" .github/workflows/ci.yml; then
        log_pass "Security scanning in CI pipeline"
    else
        log_fail "No security scanning in CI pipeline"
    fi

    if grep -q "exit-code.*1\|exit-code: '1'" .github/workflows/ci.yml; then
        log_pass "CI fails on security vulnerabilities"
    else
        log_warn "CI doesn't enforce security scan failures"
    fi

    if grep -q "uses:.*@[a-f0-9]{40}" .github/workflows/ci.yml; then
        log_pass "GitHub Actions pinned to SHA"
    else
        log_warn "GitHub Actions not pinned to SHA"
    fi
else
    log_warn "CI configuration not found"
fi

# ============================================================================
# CHECK 14: Makefile Security Targets
# ============================================================================
log_info "Checking Makefile security targets..."

if [ -f chain/Makefile.security ]; then
    log_pass "Security Makefile present"

    required_targets=("security-scan" "security-container" "security-secrets" "security-audit")
    for target in "${required_targets[@]}"; do
        if grep -q "^$target:" chain/Makefile.security; then
            log_pass "Target $target present"
        else
            log_fail "Target $target missing"
        fi
    done
else
    log_fail "Makefile.security not found"
fi

# ============================================================================
# Additional Security Checks
# ============================================================================
log_info "Running additional security checks..."

# Check for .env files in repo
if git ls-files | grep -q "\.env$\|\.env\."; then
    log_fail ".env files tracked in git"
else
    log_pass "No .env files in git"
fi

# Check Docker daemon security
if command -v docker >/dev/null 2>&1; then
    if docker info 2>/dev/null | grep -q "Security Options.*apparmor\|Security Options.*seccomp"; then
        log_pass "Docker security features enabled"
    else
        log_warn "Docker security features may not be enabled"
    fi
fi

# ============================================================================
# Summary Report
# ============================================================================
echo ""
echo "============================================================================"
echo "Security Verification Summary"
echo "============================================================================"
echo ""
echo "Total Checks: $TOTAL_CHECKS"
echo -e "${GREEN}Passed: $CHECKS_PASSED${NC}"
echo -e "${YELLOW}Warnings: $CHECKS_WARNING${NC}"
echo -e "${RED}Failed: $CHECKS_FAILED${NC}"
echo ""

# Calculate score
if [ $TOTAL_CHECKS -gt 0 ]; then
    score=$(( (CHECKS_PASSED * 100) / TOTAL_CHECKS ))
    echo "Security Score: $score%"
    echo ""

    if [ $score -ge 90 ]; then
        echo -e "${GREEN}✓ Excellent security posture${NC}"
        exit_code=0
    elif [ $score -ge 75 ]; then
        echo -e "${YELLOW}⚠ Good security, but improvements needed${NC}"
        exit_code=0
    elif [ $score -ge 50 ]; then
        echo -e "${YELLOW}⚠ Moderate security, significant improvements required${NC}"
        exit_code=1
    else
        echo -e "${RED}✗ Poor security posture - immediate action required${NC}"
        exit_code=1
    fi
else
    echo "No checks performed"
    exit_code=1
fi

echo ""
echo "For detailed remediation steps, see:"
echo "  - deployment-security/HARDENING_CHECKLIST.md"
echo "  - deployment-security/SECRETS_GUIDE.md"
echo "============================================================================"

exit $exit_code
