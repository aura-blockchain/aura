# AURA Blockchain - Secrets Management Guide

Comprehensive guide for managing secrets in AURA blockchain deployment.

## Table of Contents

1. [Overview](#overview)
2. [Secret Types](#secret-types)
3. [Generation](#generation)
4. [Storage](#storage)
5. [Rotation](#rotation)
6. [Access Control](#access-control)
7. [Backup & Recovery](#backup--recovery)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Overview

Proper secrets management is critical for blockchain security. This guide covers all aspects of secret handling from generation to rotation.

### What are Secrets?

Secrets are sensitive credentials that must be protected:
- Database passwords
- API keys
- Private keys
- TLS certificates
- Authentication tokens
- Encryption keys

### Why Docker Secrets?

Docker Secrets provide:
- ✓ Encrypted at rest and in transit
- ✓ Never stored in images or containers
- ✓ Mounted as in-memory filesystem
- ✓ Automatic cleanup on container removal
- ✓ Access control via service assignments

## Secret Types

### Application Secrets

| Secret | Usage | Rotation Frequency |
|--------|-------|-------------------|
| `grafana_admin_password.txt` | Grafana admin login | 90 days |
| `postgres_password.txt` | PostgreSQL database | 90 days |
| `redis_password.txt` | Redis cache | 90 days |
| `prometheus_basic_auth.txt` | Prometheus HTTP auth | 90 days |

### Infrastructure Secrets

| Secret | Usage | Rotation Frequency |
|--------|-------|-------------------|
| `tls/server.crt` | TLS certificate | 365 days (Let's Encrypt: 90) |
| `tls/server.key` | TLS private key | 365 days (Let's Encrypt: 90) |
| `tls/dhparam.pem` | DH parameters | 365 days |

## Generation

### Quick Start

```bash
# Generate all secrets automatically
./deployment-security/scripts/generate-secrets.sh
```

### Manual Generation

#### Strong Passwords

```bash
# Method 1: OpenSSL (recommended)
openssl rand -base64 32 > ./secrets/password.txt

# Method 2: /dev/urandom
head -c 32 /dev/urandom | base64 > ./secrets/password.txt

# Method 3: pwgen (if installed)
pwgen -s 32 1 > ./secrets/password.txt
```

#### TLS Certificates

```bash
# Production: Let's Encrypt
./deployment-security/scripts/tls-setup.sh letsencrypt

# Development: Self-signed
./deployment-security/scripts/tls-setup.sh self-signed

# Custom CA
./deployment-security/scripts/tls-setup.sh custom /path/to/cert.crt /path/to/key.key
```

#### Basic Authentication

```bash
# For Prometheus/Nginx
htpasswd -nB admin > ./secrets/basic_auth.txt

# Or with Python
python3 -c "import bcrypt; print('admin:' + bcrypt.hashpw(b'PASSWORD', bcrypt.gensalt()).decode())" > ./secrets/basic_auth.txt
```

### Password Requirements

**Minimum Standards:**
- Length: 32 characters
- Entropy: 256 bits
- Complexity: Alphanumeric + symbols
- No dictionary words
- Unique per service

**Generation Best Practices:**
```bash
# Good: Cryptographically secure
openssl rand -base64 32

# Bad: Predictable
echo "password123"

# Bad: Low entropy
date | md5sum
```

## Storage

### Directory Structure

```
./secrets/
├── README.txt                 # Documentation (no secrets!)
├── grafana_admin_password.txt # Grafana password
├── postgres_password.txt      # PostgreSQL password
├── redis_password.txt         # Redis password
├── prometheus_basic_auth.txt  # Prometheus auth
├── prometheus_password.txt    # Prometheus password (reference)
├── tls/
│   ├── server.crt            # TLS certificate
│   ├── server.key            # TLS private key
│   └── dhparam.pem           # DH parameters
└── backups/                   # Encrypted backups
    └── secrets_backup_*.tar.gz
```

### File Permissions

```bash
# Set correct permissions
chmod 700 ./secrets           # Directory: owner only
chmod 700 ./secrets/tls       # TLS directory: owner only
chmod 600 ./secrets/*.txt     # Files: owner read/write only
chmod 600 ./secrets/tls/*     # TLS files: owner read/write only
```

### Git Ignore

Ensure `.gitignore` contains:

```gitignore
# Secrets
secrets/
*.secret
*.key
*.pem
*.crt
!secrets/README.txt

# Environment files
.env
.env.*
!.env.example
```

### Verification

```bash
# Check permissions
ls -la ./secrets/
ls -la ./secrets/tls/

# Verify not in git
git status --ignored

# Check for secrets in git history
git log -p | grep -i "password\|secret\|key"
```

## Rotation

### When to Rotate

**Scheduled Rotation:**
- Every 90 days (application secrets)
- Every 365 days (TLS certificates)
- Every 180 days (API keys)

**Emergency Rotation:**
- Suspected compromise
- Employee departure
- Failed audit
- Security incident

### Rotation Process

#### Automated Rotation

```bash
# Rotate all secrets
./deployment-security/scripts/rotate-secrets.sh

# This script:
# 1. Backs up current secrets
# 2. Generates new secrets
# 3. Updates services
# 4. Restarts containers
# 5. Verifies new secrets work
```

#### Manual Rotation

##### 1. Backup Current Secrets

```bash
# Create backup
tar -czf secrets_backup_$(date +%Y%m%d).tar.gz \
    -C ./secrets \
    --exclude=backups \
    .

# Encrypt backup
openssl enc -aes-256-cbc \
    -in secrets_backup_$(date +%Y%m%d).tar.gz \
    -out secrets_backup_$(date +%Y%m%d).tar.gz.enc

# Secure backup
chmod 600 secrets_backup_$(date +%Y%m%d).tar.gz.enc
```

##### 2. Generate New Secret

```bash
# Generate new password
new_password=$(openssl rand -base64 32)
echo "$new_password" > ./secrets/postgres_password.txt.new
chmod 600 ./secrets/postgres_password.txt.new
```

##### 3. Update Service

```bash
# Update PostgreSQL
docker-compose -f docker-compose.secure.yml exec postgres \
    psql -U postgres -c "ALTER USER aura PASSWORD '$new_password';"
```

##### 4. Swap Secrets

```bash
# Atomic swap
mv ./secrets/postgres_password.txt ./secrets/postgres_password.txt.old
mv ./secrets/postgres_password.txt.new ./secrets/postgres_password.txt
```

##### 5. Restart & Verify

```bash
# Restart service
docker-compose -f docker-compose.secure.yml restart postgres

# Verify connection
docker-compose -f docker-compose.secure.yml exec postgres \
    psql -U aura -d aura_indexer -c "\conninfo"
```

##### 6. Cleanup

```bash
# Remove old secret after verification
shred -vfz -n 3 ./secrets/postgres_password.txt.old
```

### TLS Certificate Rotation

#### Let's Encrypt Auto-Renewal

```bash
# Setup auto-renewal
sudo certbot renew --deploy-hook \
    './deployment-security/scripts/tls-setup.sh --copy-certs'

# Test renewal
sudo certbot renew --dry-run
```

#### Manual Certificate Renewal

```bash
# Renew certificate
./deployment-security/scripts/tls-setup.sh letsencrypt

# Reload Nginx
docker-compose -f docker-compose.secure.yml exec nginx nginx -s reload

# Verify
openssl s_client -connect localhost:443 -servername aura.example.com
```

## Access Control

### Docker Secrets Access

```yaml
# In docker-compose.secure.yml
services:
  postgres:
    secrets:
      - postgres_password  # Service can access this secret
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
```

### File System Access

```bash
# Restrict to specific user
sudo chown aura:aura ./secrets
sudo chmod 700 ./secrets

# Use ACLs for fine-grained control
setfacl -m u:aura:rwx ./secrets
setfacl -m u:deployer:rx ./secrets
setfacl -d -m u:aura:rwx ./secrets
```

### Role-Based Access

| Role | Access Level | Permissions |
|------|-------------|-------------|
| DevSecOps Lead | Full | Read, write, rotate all secrets |
| Deployment Engineer | Limited | Read secrets, deploy services |
| Developer | None | No direct access |
| Auditor | Read-only | View secret metadata (not values) |

### Audit Logging

```bash
# Enable audit logging
sudo auditctl -w ./secrets -p wa -k secrets_access

# View access logs
sudo ausearch -k secrets_access

# Monitor in real-time
sudo tail -f /var/log/audit/audit.log | grep secrets_access
```

## Backup & Recovery

### Backup Strategy

#### Automated Backups

```bash
# Create backup script: /etc/cron.daily/backup-secrets
#!/bin/bash
cd /opt/aura
tar -czf /backup/secrets_$(date +%Y%m%d).tar.gz -C ./secrets .
openssl enc -aes-256-cbc -in /backup/secrets_$(date +%Y%m%d).tar.gz \
    -out /backup/secrets_$(date +%Y%m%d).tar.gz.enc \
    -pass file:/root/.secrets_backup_key
rm /backup/secrets_$(date +%Y%m%d).tar.gz
find /backup -name "secrets_*.tar.gz.enc" -mtime +30 -delete
```

#### Manual Backup

```bash
# Backup secrets
./deployment-security/scripts/generate-secrets.sh --backup-only

# Or manually
tar -czf secrets_backup.tar.gz -C ./secrets --exclude=backups .
openssl enc -aes-256-cbc -in secrets_backup.tar.gz \
    -out secrets_backup.tar.gz.enc
```

### Recovery Process

#### 1. Retrieve Backup

```bash
# Decrypt backup
openssl enc -aes-256-cbc -d \
    -in secrets_backup.tar.gz.enc \
    -out secrets_backup.tar.gz
```

#### 2. Verify Backup

```bash
# List contents
tar -tzf secrets_backup.tar.gz

# Verify integrity
tar -xzf secrets_backup.tar.gz -C /tmp/verify
```

#### 3. Restore Secrets

```bash
# Stop services
docker-compose -f docker-compose.secure.yml down

# Restore
tar -xzf secrets_backup.tar.gz -C ./secrets

# Fix permissions
chmod 700 ./secrets
chmod 600 ./secrets/*.txt
chmod 700 ./secrets/tls
chmod 600 ./secrets/tls/*

# Start services
docker-compose -f docker-compose.secure.yml up -d
```

#### 4. Verify Recovery

```bash
# Test services
docker-compose -f docker-compose.secure.yml ps
docker-compose -f docker-compose.secure.yml logs

# Verify secret access
./deployment-security/scripts/verify-deployment-security.sh
```

### Off-Site Backup

```bash
# S3 (encrypted)
aws s3 cp secrets_backup.tar.gz.enc \
    s3://aura-backups/secrets/$(date +%Y%m%d)/ \
    --server-side-encryption AES256

# Azure Blob Storage
az storage blob upload \
    --account-name aurabackups \
    --container-name secrets \
    --name secrets_$(date +%Y%m%d).tar.gz.enc \
    --file secrets_backup.tar.gz.enc

# Google Cloud Storage
gsutil cp secrets_backup.tar.gz.enc \
    gs://aura-backups/secrets/$(date +%Y%m%d)/
```

## Best Practices

### DO's

✓ **Use strong, random passwords**
```bash
openssl rand -base64 32  # Good
```

✓ **Store secrets in Docker secrets**
```yaml
secrets:
  db_password:
    file: ./secrets/postgres_password.txt
```

✓ **Rotate secrets regularly**
```bash
# Every 90 days
./deployment-security/scripts/rotate-secrets.sh
```

✓ **Encrypt backups**
```bash
openssl enc -aes-256-cbc -in backup.tar.gz -out backup.tar.gz.enc
```

✓ **Use separate secrets per environment**
```
production/secrets/
staging/secrets/
development/secrets/
```

✓ **Audit secret access**
```bash
sudo auditctl -w ./secrets -p wa -k secrets_access
```

### DON'Ts

✗ **Never commit secrets to git**
```bash
# Bad!
git add ./secrets/*.txt
```

✗ **Never use environment variables for secrets**
```yaml
# Bad!
environment:
  - DATABASE_PASSWORD=mypassword123
```

✗ **Never hardcode secrets**
```go
// Bad!
password := "mypassword123"
```

✗ **Never reuse secrets across services**
```bash
# Bad!
cp postgres_password.txt redis_password.txt
```

✗ **Never use weak passwords**
```bash
# Bad!
echo "password123" > secret.txt
```

✗ **Never skip rotation**
```bash
# Bad - last rotation: 2023-01-01
# Current date: 2025-01-26
```

## Troubleshooting

### Common Issues

#### Secret Not Found

**Problem:** Container can't access secret

```bash
# Check secret exists
ls -la ./secrets/postgres_password.txt

# Check docker secret
docker secret ls

# Verify service has access
docker-compose -f docker-compose.secure.yml config | grep -A 5 secrets

# Check container mount
docker exec -it aura-postgres ls -la /run/secrets/
```

#### Permission Denied

**Problem:** Can't read secret file

```bash
# Fix ownership
sudo chown aura:aura ./secrets/*.txt

# Fix permissions
chmod 600 ./secrets/*.txt

# Verify
ls -la ./secrets/
```

#### Invalid Password

**Problem:** Service rejects password

```bash
# Check for trailing whitespace
cat -A ./secrets/postgres_password.txt

# Regenerate if corrupted
openssl rand -base64 32 | tr -d '\n' > ./secrets/postgres_password.txt
```

#### Certificate Errors

**Problem:** TLS handshake fails

```bash
# Verify certificate
openssl x509 -in ./secrets/tls/server.crt -text -noout

# Check expiration
openssl x509 -in ./secrets/tls/server.crt -noout -dates

# Verify key matches
openssl x509 -in ./secrets/tls/server.crt -noout -modulus | openssl md5
openssl rsa -in ./secrets/tls/server.key -noout -modulus | openssl md5
```

### Emergency Recovery

#### Compromised Secret

```bash
# 1. Immediate rotation
./deployment-security/scripts/rotate-secrets.sh

# 2. Review access logs
sudo ausearch -k secrets_access -ts recent

# 3. Check for unauthorized access
docker-compose -f docker-compose.secure.yml logs | grep -i "authentication failed"

# 4. Incident response
# See HARDENING_CHECKLIST.md - Incident Response section
```

#### Lost Secrets

```bash
# 1. Restore from backup
./deployment-security/scripts/restore-secrets.sh

# 2. If no backup, regenerate
./deployment-security/scripts/generate-secrets.sh --force

# 3. Update all services
docker-compose -f docker-compose.secure.yml up -d --force-recreate
```

## Security Checklist

- [ ] All secrets generated with cryptographically secure methods
- [ ] File permissions: 600 for files, 700 for directories
- [ ] Secrets excluded from version control (.gitignore)
- [ ] Docker secrets used instead of environment variables
- [ ] Secrets rotated every 90 days
- [ ] Backups encrypted and stored off-site
- [ ] Access logs enabled and monitored
- [ ] Emergency rotation procedure documented and tested
- [ ] Separate secrets per environment
- [ ] Audit trail maintained for all secret operations

## Additional Resources

- [Docker Secrets Documentation](https://docs.docker.com/engine/swarm/secrets/)
- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [NIST SP 800-57: Key Management](https://csrc.nist.gov/publications/detail/sp/800-57-part-1/rev-5/final)
- [HashiCorp Vault](https://www.vaultproject.io/)
- [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)

---

**Document Version**: 1.0.0
**Last Updated**: 2025-01-26
**Next Review**: 2025-04-26
