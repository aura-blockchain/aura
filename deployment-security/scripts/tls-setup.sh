#!/bin/bash
# ============================================================================
# AURA Blockchain - TLS Certificate Setup Script
# ============================================================================
# Sets up TLS certificates for production or development
# Supports: Let's Encrypt, self-signed, custom CA
# ============================================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
DOMAIN="${DOMAIN:-aura.example.com}"
EMAIL="${EMAIL:-admin@aura.example.com}"
TLS_DIR="./secrets/tls"
CERT_TYPE="${1:-self-signed}"  # self-signed, letsencrypt, custom

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Create TLS directory
setup_tls_directory() {
    log_info "Setting up TLS directory..."
    mkdir -p "${TLS_DIR}"
    chmod 700 "${TLS_DIR}"
}

# Generate self-signed certificate
generate_self_signed() {
    log_info "Generating self-signed certificate for ${DOMAIN}..."
    log_warning "Self-signed certificates should NOT be used in production!"

    # Generate private key
    openssl genrsa -out "${TLS_DIR}/server.key" 4096
    chmod 600 "${TLS_DIR}/server.key"

    # Create OpenSSL config
    cat > "${TLS_DIR}/openssl.cnf" <<EOF
[req]
default_bits = 4096
prompt = no
default_md = sha384
distinguished_name = dn
req_extensions = v3_req

[dn]
C=US
ST=State
L=City
O=AURA Network
OU=Security
CN=${DOMAIN}

[v3_req]
keyUsage = keyEncipherment, dataEncipherment, digitalSignature
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${DOMAIN}
DNS.2 = *.${DOMAIN}
DNS.3 = localhost
IP.1 = 127.0.0.1
EOF

    # Generate certificate
    openssl req -new -x509 -sha384 -days 365 \
        -key "${TLS_DIR}/server.key" \
        -out "${TLS_DIR}/server.crt" \
        -config "${TLS_DIR}/openssl.cnf" \
        -extensions v3_req

    chmod 644 "${TLS_DIR}/server.crt"
    rm "${TLS_DIR}/openssl.cnf"

    log_success "Self-signed certificate generated"
    log_warning "Valid for 365 days - expires: $(date -d "+365 days" +"%Y-%m-%d" 2>/dev/null || date -v+365d +"%Y-%m-%d")"
}

# Setup Let's Encrypt certificate
setup_letsencrypt() {
    log_info "Setting up Let's Encrypt certificate for ${DOMAIN}..."

    if ! command_exists certbot; then
        log_error "certbot not installed. Install: https://certbot.eff.org/"
    fi

    # Check if port 80 is available
    if netstat -tuln | grep -q ":80 "; then
        log_warning "Port 80 is in use. Stop services using port 80 first."
        read -p "Stop services and continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_error "Aborted by user"
        fi
    fi

    # Run certbot
    log_info "Running certbot for ${DOMAIN}..."
    sudo certbot certonly --standalone \
        --non-interactive \
        --agree-tos \
        --email "${EMAIL}" \
        -d "${DOMAIN}" \
        --preferred-challenges http

    # Copy certificates
    sudo cp "/etc/letsencrypt/live/${DOMAIN}/fullchain.pem" "${TLS_DIR}/server.crt"
    sudo cp "/etc/letsencrypt/live/${DOMAIN}/privkey.pem" "${TLS_DIR}/server.key"

    # Fix permissions
    sudo chown $(id -u):$(id -g) "${TLS_DIR}/server.crt" "${TLS_DIR}/server.key"
    chmod 644 "${TLS_DIR}/server.crt"
    chmod 600 "${TLS_DIR}/server.key"

    log_success "Let's Encrypt certificate installed"
    log_info "Certificate will auto-renew via certbot"
    log_info "Set up renewal hook: sudo certbot renew --deploy-hook './deployment-security/scripts/tls-setup.sh --copy-certs'"
}

# Copy existing certificates
copy_custom_certs() {
    log_info "Setting up custom certificates..."

    if [ ! -f "$2" ] || [ ! -f "$3" ]; then
        log_error "Usage: $0 custom <cert-file> <key-file>"
    fi

    local cert_file="$2"
    local key_file="$3"

    # Validate certificate
    if ! openssl x509 -in "${cert_file}" -noout 2>/dev/null; then
        log_error "Invalid certificate file: ${cert_file}"
    fi

    # Validate private key
    if ! openssl rsa -in "${key_file}" -check -noout 2>/dev/null; then
        log_error "Invalid private key file: ${key_file}"
    fi

    # Copy files
    cp "${cert_file}" "${TLS_DIR}/server.crt"
    cp "${key_file}" "${TLS_DIR}/server.key"

    chmod 644 "${TLS_DIR}/server.crt"
    chmod 600 "${TLS_DIR}/server.key"

    log_success "Custom certificates installed"
}

# Verify certificate
verify_certificate() {
    log_info "Verifying certificate..."

    if [ ! -f "${TLS_DIR}/server.crt" ] || [ ! -f "${TLS_DIR}/server.key" ]; then
        log_error "Certificate files not found"
    fi

    # Check certificate validity
    local expiry_date=$(openssl x509 -in "${TLS_DIR}/server.crt" -noout -enddate | cut -d= -f2)
    log_info "Certificate expires: ${expiry_date}"

    # Check if certificate matches private key
    local cert_modulus=$(openssl x509 -in "${TLS_DIR}/server.crt" -noout -modulus | openssl md5)
    local key_modulus=$(openssl rsa -in "${TLS_DIR}/server.key" -noout -modulus | openssl md5)

    if [ "${cert_modulus}" != "${key_modulus}" ]; then
        log_error "Certificate and private key do not match!"
    fi

    # Display certificate details
    log_info "Certificate details:"
    openssl x509 -in "${TLS_DIR}/server.crt" -noout -subject -issuer -dates -ext subjectAltName

    log_success "Certificate verification passed"
}

# Generate DH parameters for stronger encryption
generate_dhparams() {
    log_info "Generating DH parameters (this may take a while)..."

    if [ ! -f "${TLS_DIR}/dhparam.pem" ]; then
        openssl dhparam -out "${TLS_DIR}/dhparam.pem" 4096
        chmod 644 "${TLS_DIR}/dhparam.pem"
        log_success "DH parameters generated"
    else
        log_warning "DH parameters already exist, skipping"
    fi
}

# Create nginx SSL configuration
create_nginx_ssl_config() {
    log_info "Creating nginx SSL configuration..."

    mkdir -p ./nginx

    cat > ./nginx/ssl-config.conf <<'EOF'
# AURA Blockchain - Nginx SSL Configuration
# Strong TLS security settings based on Mozilla SSL Configuration Generator

# SSL Protocol and Ciphers (Modern configuration)
ssl_protocols TLSv1.3 TLSv1.2;
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';
ssl_prefer_server_ciphers on;

# SSL Certificates
ssl_certificate /run/secrets/tls_cert;
ssl_certificate_key /run/secrets/tls_key;

# DH Parameters for Perfect Forward Secrecy
ssl_dhparam /etc/nginx/ssl/dhparam.pem;

# SSL Session Settings
ssl_session_timeout 1d;
ssl_session_cache shared:SSL:50m;
ssl_session_tickets off;

# OCSP Stapling
ssl_stapling on;
ssl_stapling_verify on;
resolver 8.8.8.8 8.8.4.4 valid=300s;
resolver_timeout 5s;

# Security Headers
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:; frame-ancestors 'self';" always;

# Rate Limiting
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=login_limit:10m rate=1r/s;

# SSL Verification Depth
ssl_verify_depth 2;

# Buffer Sizes (prevent large header attacks)
client_body_buffer_size 1K;
client_header_buffer_size 1k;
client_max_body_size 1m;
large_client_header_buffers 2 1k;
EOF

    chmod 644 ./nginx/ssl-config.conf
    log_success "Nginx SSL configuration created"
}

# Main function
main() {
    log_info "AURA Blockchain TLS Setup"
    echo ""

    setup_tls_directory

    case "${CERT_TYPE}" in
        self-signed|--self-signed)
            generate_self_signed
            ;;
        letsencrypt|--letsencrypt|production|--production)
            setup_letsencrypt
            ;;
        custom|--custom)
            copy_custom_certs "$@"
            ;;
        --copy-certs)
            # For certbot renewal hook
            setup_letsencrypt
            ;;
        --verify)
            verify_certificate
            exit 0
            ;;
        --help|-h)
            echo "Usage: $0 [OPTION]"
            echo ""
            echo "Options:"
            echo "  self-signed         Generate self-signed certificate (default)"
            echo "  letsencrypt         Setup Let's Encrypt certificate"
            echo "  production          Alias for letsencrypt"
            echo "  custom <cert> <key> Install custom certificate"
            echo "  --verify            Verify existing certificate"
            echo "  --help              Show this help"
            echo ""
            echo "Environment variables:"
            echo "  DOMAIN              Domain name (default: aura.example.com)"
            echo "  EMAIL               Email for Let's Encrypt (default: admin@aura.example.com)"
            echo ""
            exit 0
            ;;
        *)
            log_error "Unknown option: ${CERT_TYPE}. Use --help for usage."
            ;;
    esac

    verify_certificate
    generate_dhparams
    create_nginx_ssl_config

    echo ""
    log_success "TLS setup completed successfully!"
    echo ""
    log_info "Next steps:"
    echo "  1. Review certificate: openssl x509 -in ${TLS_DIR}/server.crt -text -noout"
    echo "  2. Test nginx config: docker-compose -f docker-compose.secure.yml config"
    echo "  3. Start services: docker-compose -f docker-compose.secure.yml up -d"
    echo "  4. Verify TLS: openssl s_client -connect ${DOMAIN}:443"
    echo ""

    if [ "${CERT_TYPE}" == "self-signed" ] || [ "${CERT_TYPE}" == "--self-signed" ]; then
        log_warning "Self-signed certificate in use!"
        log_warning "For production, run: $0 letsencrypt"
    fi
}

main "$@"
