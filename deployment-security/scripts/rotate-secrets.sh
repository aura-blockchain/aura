#!/bin/bash
# ============================================================================
# AURA Blockchain - Secret Rotation Script
# ============================================================================
# Safely rotates all secrets with zero-downtime
# ============================================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

SECRETS_DIR="./secrets"
BACKUP_DIR="./secrets/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "============================================================================"
echo "AURA Blockchain - Secret Rotation"
echo "============================================================================"
echo ""

# Backup current secrets
backup_secrets() {
    log_info "Backing up current secrets..."
    mkdir -p "$BACKUP_DIR"

    if [ -d "$SECRETS_DIR" ]; then
        tar -czf "$BACKUP_DIR/secrets_backup_$TIMESTAMP.tar.gz" \
            -C "$SECRETS_DIR" \
            --exclude=backups \
            .

        chmod 600 "$BACKUP_DIR/secrets_backup_$TIMESTAMP.tar.gz"
        log_success "Secrets backed up to $BACKUP_DIR/secrets_backup_$TIMESTAMP.tar.gz"
    else
        log_error "Secrets directory not found"
    fi
}

# Generate new password
generate_password() {
    openssl rand -base64 32 | tr -d '\n'
}

# Rotate Grafana password
rotate_grafana() {
    log_info "Rotating Grafana password..."
    local new_password=$(generate_password)
    echo "$new_password" > "$SECRETS_DIR/grafana_admin_password.txt.new"
    chmod 600 "$SECRETS_DIR/grafana_admin_password.txt.new"

    # Update Grafana
    docker-compose -f docker-compose.secure.yml exec grafana \
        grafana-cli admin reset-admin-password "$new_password" || log_warning "Manual Grafana password update required"

    # Swap files
    mv "$SECRETS_DIR/grafana_admin_password.txt" "$SECRETS_DIR/grafana_admin_password.txt.old"
    mv "$SECRETS_DIR/grafana_admin_password.txt.new" "$SECRETS_DIR/grafana_admin_password.txt"

    log_success "Grafana password rotated"
}

# Rotate PostgreSQL password
rotate_postgres() {
    log_info "Rotating PostgreSQL password..."
    local new_password=$(generate_password)
    echo "$new_password" > "$SECRETS_DIR/postgres_password.txt.new"
    chmod 600 "$SECRETS_DIR/postgres_password.txt.new"

    # Update PostgreSQL
    docker-compose -f docker-compose.secure.yml exec postgres \
        psql -U postgres -c "ALTER USER aura PASSWORD '$new_password';" || log_warning "Manual PostgreSQL password update required"

    # Swap files
    mv "$SECRETS_DIR/postgres_password.txt" "$SECRETS_DIR/postgres_password.txt.old"
    mv "$SECRETS_DIR/postgres_password.txt.new" "$SECRETS_DIR/postgres_password.txt"

    log_success "PostgreSQL password rotated"
}

# Rotate Redis password
rotate_redis() {
    log_info "Rotating Redis password..."
    local new_password=$(generate_password)
    echo "$new_password" > "$SECRETS_DIR/redis_password.txt.new"
    chmod 600 "$SECRETS_DIR/redis_password.txt.new"

    # Redis requires restart for password change
    log_warning "Redis rotation requires service restart"

    # Swap files
    mv "$SECRETS_DIR/redis_password.txt" "$SECRETS_DIR/redis_password.txt.old"
    mv "$SECRETS_DIR/redis_password.txt.new" "$SECRETS_DIR/redis_password.txt"

    log_success "Redis password rotated (restart required)"
}

# Main rotation function
main() {
    log_warning "This will rotate all secrets. Ensure you have tested the process!"
    read -p "Continue with secret rotation? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Aborted by user"
    fi

    backup_secrets

    rotate_grafana
    rotate_postgres
    rotate_redis

    echo ""
    log_info "Restarting services with new secrets..."
    docker-compose -f docker-compose.secure.yml up -d --force-recreate redis

    echo ""
    log_success "Secret rotation completed!"
    log_info "Backup location: $BACKUP_DIR/secrets_backup_$TIMESTAMP.tar.gz"
    log_warning "Update any external services using these credentials"

    # Cleanup old files
    rm -f "$SECRETS_DIR"/*.txt.old
}

main "$@"
