#!/bin/bash
# ============================================================================
# AURA Blockchain - Secure Secret Generation Script
# ============================================================================
# Generates cryptographically secure secrets for all services
# ============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SECRETS_DIR="./secrets"
TLS_DIR="${SECRETS_DIR}/tls"
DOMAIN="${DOMAIN:-aura.example.com}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Generate secure random password
generate_password() {
    local length="${1:-32}"
    openssl rand -base64 "${length}" | tr -d '\n'
}

# Main function
main() {
    log_info "Starting secure secret generation for AURA Blockchain"

    # Create secrets directory with restrictive permissions
    log_info "Creating secrets directory structure..."
    mkdir -p "${SECRETS_DIR}" "${TLS_DIR}"
    chmod 700 "${SECRETS_DIR}" "${TLS_DIR}"

    # Generate Grafana admin password
    log_info "Generating Grafana admin password..."
    if [ ! -f "${SECRETS_DIR}/grafana_admin_password.txt" ]; then
        generate_password 32 > "${SECRETS_DIR}/grafana_admin_password.txt"
        chmod 600 "${SECRETS_DIR}/grafana_admin_password.txt"
        log_success "Grafana admin password generated"
    else
        log_warning "Grafana admin password already exists, skipping"
    fi

    # Generate PostgreSQL password
    log_info "Generating PostgreSQL password..."
    if [ ! -f "${SECRETS_DIR}/postgres_password.txt" ]; then
        generate_password 32 > "${SECRETS_DIR}/postgres_password.txt"
        chmod 600 "${SECRETS_DIR}/postgres_password.txt"
        log_success "PostgreSQL password generated"
    else
        log_warning "PostgreSQL password already exists, skipping"
    fi

    # Generate Redis password
    log_info "Generating Redis password..."
    if [ ! -f "${SECRETS_DIR}/redis_password.txt" ]; then
        generate_password 32 > "${SECRETS_DIR}/redis_password.txt"
        chmod 600 "${SECRETS_DIR}/redis_password.txt"
        log_success "Redis password generated"
    else
        log_warning "Redis password already exists, skipping"
    fi

    # Generate Prometheus basic auth
    log_info "Generating Prometheus basic auth..."
    if [ ! -f "${SECRETS_DIR}/prometheus_basic_auth.txt" ]; then
        if command_exists htpasswd; then
            local prom_password=$(generate_password 32)
            echo -n "admin:$(openssl passwd -apr1 ${prom_password})" > "${SECRETS_DIR}/prometheus_basic_auth.txt"
            echo "${prom_password}" > "${SECRETS_DIR}/prometheus_password.txt"
            chmod 600 "${SECRETS_DIR}/prometheus_basic_auth.txt"
            chmod 600 "${SECRETS_DIR}/prometheus_password.txt"
            log_success "Prometheus basic auth generated"
        else
            log_warning "htpasswd not found, creating placeholder"
            generate_password 32 > "${SECRETS_DIR}/prometheus_password.txt"
            echo "admin:REPLACE_WITH_HTPASSWD_HASH" > "${SECRETS_DIR}/prometheus_basic_auth.txt"
            chmod 600 "${SECRETS_DIR}/prometheus_password.txt"
            chmod 600 "${SECRETS_DIR}/prometheus_basic_auth.txt"
            log_warning "Run: htpasswd -nB admin > ${SECRETS_DIR}/prometheus_basic_auth.txt"
        fi
    else
        log_warning "Prometheus basic auth already exists, skipping"
    fi

    # Generate TLS certificates
    log_info "Generating TLS certificates..."
    if [ ! -f "${TLS_DIR}/server.crt" ] || [ ! -f "${TLS_DIR}/server.key" ]; then
        log_info "Creating self-signed certificate for ${DOMAIN}..."
        log_warning "For production, use Let's Encrypt or your CA!"

        openssl req -x509 -newkey rsa:4096 -nodes \
            -keyout "${TLS_DIR}/server.key" \
            -out "${TLS_DIR}/server.crt" \
            -days 365 \
            -subj "/C=US/ST=State/L=City/O=AURA/OU=Security/CN=${DOMAIN}" \
            -addext "subjectAltName=DNS:${DOMAIN},DNS:*.${DOMAIN},DNS:localhost"

        chmod 600 "${TLS_DIR}/server.key"
        chmod 644 "${TLS_DIR}/server.crt"
        log_success "Self-signed TLS certificate generated"
        log_warning "Valid for 365 days - set up renewal before expiration!"
    else
        log_warning "TLS certificates already exist, skipping"
    fi

    # Create secrets summary
    log_info "Creating secrets summary..."
    cat > "${SECRETS_DIR}/README.txt" <<EOF
AURA Blockchain - Secrets Directory
====================================

Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

Files in this directory:
- grafana_admin_password.txt: Grafana admin password
- postgres_password.txt: PostgreSQL database password
- redis_password.txt: Redis cache password
- prometheus_basic_auth.txt: Prometheus basic auth credentials
- prometheus_password.txt: Prometheus password (plaintext for reference)
- tls/server.crt: TLS certificate
- tls/server.key: TLS private key

SECURITY WARNINGS:
==================
1. NEVER commit these files to version control
2. Restrict access to authorized personnel only
3. Rotate secrets every 90 days
4. Use external secret management in production (Vault, AWS Secrets Manager)
5. Backup secrets securely with encryption
6. Monitor access to secrets directory
7. Use proper TLS certificates in production (Let's Encrypt)

To rotate secrets:
./deployment-security/scripts/rotate-secrets.sh

For production deployment:
./deployment-security/scripts/tls-setup.sh --production
EOF
    chmod 600 "${SECRETS_DIR}/README.txt"

    # Display summary
    echo ""
    log_success "All secrets generated successfully!"
    echo ""
    log_info "Secrets location: ${SECRETS_DIR}"
    log_info "Files created:"
    ls -la "${SECRETS_DIR}"
    ls -la "${TLS_DIR}"
    echo ""
    log_warning "Security Checklist:"
    echo "  [ ] Verify .gitignore excludes ./secrets/"
    echo "  [ ] Set file permissions: chmod 600 ./secrets/*.txt"
    echo "  [ ] Document secret storage location"
    echo "  [ ] Set up secret rotation schedule"
    echo "  [ ] Configure backup encryption"
    echo "  [ ] For production: Replace self-signed cert with proper TLS"
    echo ""
    log_info "To view a secret: cat ${SECRETS_DIR}/grafana_admin_password.txt"
    log_info "To rotate secrets: ./deployment-security/scripts/rotate-secrets.sh"
    echo ""
}

# Run main function
main "$@"
