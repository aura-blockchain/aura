# AURA Blockchain - Deployment Security

Complete security hardening for production deployment of AURA blockchain infrastructure.

## Overview

This directory contains comprehensive security configurations, scripts, and documentation for deploying AURA blockchain with production-grade security hardening. All 14 critical/high security vulnerabilities have been addressed.

## Quick Start

```bash
# 1. Generate secrets
./deployment-security/scripts/generate-secrets.sh

# 2. Setup TLS certificates
./deployment-security/scripts/tls-setup.sh letsencrypt

# 3. Verify security configuration
./deployment-security/scripts/verify-deployment-security.sh

# 4. Deploy with secure configuration
docker-compose -f docker-compose.secure.yml up -d

# 5. Run security scans
cd chain && make -f Makefile.security security-all
```

## What's Fixed

### ✓ CRITICAL Vulnerabilities (6)

1. **Hardcoded Credentials** - All credentials now use Docker secrets
2. **Public Port Exposure** - Sensitive ports bound to 127.0.0.1
3. **Missing Resource Limits** - CPU, memory, and PID limits configured
4. **Insecure Volume Mounts** - Removed dangerous host filesystem mounts
5. **Missing Image Pinning** - All images pinned with SHA256 digests
6. **No Secrets Management** - Comprehensive secrets management system

### ✓ HIGH Vulnerabilities (8)

7. **No Security Context** - Configured no-new-privileges, cap-drop, read-only
8. **Weak TLS/Certs** - Proper certificate management with rotation
9. **No Restart Policy** - Changed to on-failure (prevents hiding attacks)
10. **Weak Health Checks** - Enhanced to validate security controls
11. **No Network Segmentation** - Implemented 3-tier network architecture
12. **Insecure Prometheus** - Basic auth and localhost binding
13. **Docker Build ARG Injection** - Build argument validation
14. **Missing Makefile Security** - Comprehensive security targets

## Directory Structure

```
deployment-security/
├── README.md                          # This file
├── HARDENING_CHECKLIST.md            # Step-by-step hardening guide
├── SECRETS_GUIDE.md                  # Secrets management documentation
├── scripts/
│   ├── generate-secrets.sh           # Secret generation
│   ├── tls-setup.sh                  # TLS certificate setup
│   ├── rotate-secrets.sh             # Secret rotation
│   └── verify-deployment-security.sh # Security verification
├── nginx/
│   └── ssl-config.conf               # Nginx TLS configuration
└── prometheus/
    └── security-rules.yml            # Security alerting rules
```

## Security Files

### Core Configuration

- **docker-compose.secure.yml** - Hardened Docker Compose configuration
- **docker-compose.secrets.template** - Secrets template and guide
- **Dockerfile.secure** - Security-enhanced Dockerfile with SAST
- **Makefile.security** - Security scanning and audit targets

### Scripts

All scripts are located in `deployment-security/scripts/`:

| Script | Purpose | Usage |
|--------|---------|-------|
| `generate-secrets.sh` | Generate all secrets | `./generate-secrets.sh` |
| `tls-setup.sh` | Setup TLS certificates | `./tls-setup.sh [letsencrypt\|self-signed]` |
| `rotate-secrets.sh` | Rotate secrets safely | `./rotate-secrets.sh` |
| `verify-deployment-security.sh` | Verify security config | `./verify-deployment-security.sh` |

### Configuration Files

| File | Purpose |
|------|---------|
| `nginx/ssl-config.conf` | Strong TLS settings for Nginx |
| `prometheus/security-rules.yml` | Security monitoring alerts |

## Deployment Workflow

### 1. Pre-Deployment Setup

```bash
# Create secrets directory
mkdir -p ./secrets/tls
chmod 700 ./secrets

# Generate secrets
./deployment-security/scripts/generate-secrets.sh

# Setup TLS (production)
./deployment-security/scripts/tls-setup.sh letsencrypt

# Or for development
./deployment-security/scripts/tls-setup.sh self-signed
```

### 2. Security Verification

```bash
# Run security verification
./deployment-security/scripts/verify-deployment-security.sh

# Expected: 90%+ security score
```

### 3. Build Secure Images

```bash
# Build with security validation
docker build \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -f Dockerfile.secure \
  -t aura-node:secure \
  .

# Scan image
docker scan aura-node:secure
trivy image aura-node:secure
```

### 4. Deploy Services

```bash
# Deploy with secure configuration
docker-compose -f docker-compose.secure.yml up -d

# Verify deployment
docker-compose -f docker-compose.secure.yml ps
docker-compose -f docker-compose.secure.yml logs -f
```

### 5. Post-Deployment Verification

```bash
# Check security
./deployment-security/scripts/verify-deployment-security.sh

# Monitor logs
docker-compose -f docker-compose.secure.yml logs -f aura-node

# Run security scans
cd chain && make -f Makefile.security security-all
```

## Security Features

### Network Security

- **3-Tier Network Segmentation**:
  - Frontend: Public-facing (nginx only)
  - Backend: Internal services (aura-node, prometheus, grafana)
  - Data: Databases (postgres, redis)

- **Port Binding**:
  - Sensitive ports: 127.0.0.1 only
  - Public ports: 80, 443 (nginx with TLS)

### Container Security

- **Security Options**:
  - `no-new-privileges: true`
  - `apparmor=docker-default`
  - Custom seccomp profiles

- **Capabilities**:
  - Drop ALL by default
  - Add only required capabilities
  - Minimal privilege principle

- **Resource Limits**:
  - CPU limits prevent DoS
  - Memory limits prevent exhaustion
  - PID limits prevent fork bombs

### Secrets Management

- **Docker Secrets**: All credentials in Docker secrets (not env vars)
- **File Permissions**: 600 for secrets, 700 for directories
- **Encryption**: Secrets encrypted at rest and in transit
- **Rotation**: Automated rotation scripts
- **No Version Control**: Secrets never committed to git

### TLS/SSL

- **Modern Protocols**: TLS 1.2+ only
- **Strong Ciphers**: Perfect Forward Secrecy
- **HSTS**: Strict Transport Security
- **Certificate Pinning**: SHA256 digest pinning
- **Auto-Renewal**: Let's Encrypt integration

## Security Scanning

### Code Security

```bash
cd chain

# Run all security scans
make -f Makefile.security security-all

# Individual scans
make -f Makefile.security security-dependencies  # govulncheck
make -f Makefile.security security-sast          # gosec, staticcheck
make -f Makefile.security security-baseline      # Hardcoded secrets check
```

### Container Security

```bash
# Trivy scan
make -f Makefile.security security-container-trivy

# Grype scan
make -f Makefile.security security-container-grype

# Docker image linting
make -f Makefile.security security-container-dockle
```

### Secret Scanning

```bash
# TruffleHog
make -f Makefile.security security-secrets-trufflehog

# Gitleaks
make -f Makefile.security security-secrets-gitleaks
```

### Generate Security Report

```bash
cd chain
make -f Makefile.security security-report

# View report
cat security-reports/SECURITY_REPORT.md
```

## Monitoring & Alerting

### Prometheus Security Rules

Security-focused alerting rules in `prometheus/security-rules.yml`:

- Authentication failures
- Suspicious network activity
- Resource exhaustion
- Certificate expiration
- Byzantine behavior
- Container security issues

### Alert Integration

```yaml
# Add to prometheus.yml
rule_files:
  - /etc/prometheus/rules/security-rules.yml
```

## Maintenance

### Secret Rotation (Every 90 Days)

```bash
# Automated rotation
./deployment-security/scripts/rotate-secrets.sh

# Manual rotation
./deployment-security/scripts/generate-secrets.sh --force
docker-compose -f docker-compose.secure.yml up -d --force-recreate
```

### Certificate Renewal

```bash
# Let's Encrypt auto-renewal
sudo certbot renew --deploy-hook './deployment-security/scripts/tls-setup.sh --copy-certs'

# Manual renewal
./deployment-security/scripts/tls-setup.sh letsencrypt
docker-compose -f docker-compose.secure.yml restart nginx
```

### Security Updates

```bash
# Update base images
docker-compose -f docker-compose.secure.yml pull

# Rebuild with security patches
docker-compose -f docker-compose.secure.yml build --no-cache

# Update Go dependencies
cd chain
go get -u ./...
go mod tidy
make -f Makefile.security security-dependencies
```

## Compliance

### Hardening Standards

- [x] CIS Docker Benchmark
- [x] NIST Cybersecurity Framework
- [x] OWASP Container Security
- [x] PCI DSS Container Requirements

### Audit Checklist

See `HARDENING_CHECKLIST.md` for complete audit checklist.

## Troubleshooting

### Common Issues

**Q: Container won't start after hardening**
```bash
# Check logs
docker-compose -f docker-compose.secure.yml logs <service>

# Verify permissions
ls -la ./secrets/

# Check security options
docker inspect <container> | grep -A 20 SecurityOpt
```

**Q: TLS certificate errors**
```bash
# Verify certificate
openssl x509 -in ./secrets/tls/server.crt -text -noout

# Check nginx config
docker-compose -f docker-compose.secure.yml exec nginx nginx -t

# Regenerate certificate
./deployment-security/scripts/tls-setup.sh letsencrypt
```

**Q: Secret access denied**
```bash
# Fix permissions
chmod 700 ./secrets
chmod 600 ./secrets/*.txt

# Verify Docker secrets
docker secret ls
```

### Debug Mode

```bash
# Run with verbose logging
docker-compose -f docker-compose.secure.yml --verbose up

# Check security verification
./deployment-security/scripts/verify-deployment-security.sh

# View security audit
cd chain && make -f Makefile.security security-audit
```

## Support

For security issues or questions:

- **Security Email**: security@aura.network
- **Documentation**: See `HARDENING_CHECKLIST.md` and `SECRETS_GUIDE.md`
- **Verification**: Run `verify-deployment-security.sh`

## References

- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [OWASP Container Security](https://owasp.org/www-project-docker-top-10/)
- [Mozilla TLS Guidelines](https://wiki.mozilla.org/Security/Server_Side_TLS)

---

**Last Updated**: 2025-01-26
**Security Version**: 1.0.0
**Minimum Security Score**: 90%
