#!/bin/bash
# Generate self-signed SSL certificates for Aura Testnet RPC endpoints
# For production, replace with Let's Encrypt or proper CA-signed certificates

set -e

# Configuration
CERT_DIR="${CERT_DIR:-/home/decri/blockchain-projects/aura/nginx/ssl}"
DOMAIN="${DOMAIN:-rpc.testnet.aura.network}"
DAYS="${DAYS:-365}"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating self-signed SSL certificates for Aura Testnet${NC}"
echo "Domain: $DOMAIN"
echo "Certificate directory: $CERT_DIR"
echo "Validity: $DAYS days"
echo ""

# Create directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate private key
echo -e "${GREEN}Generating private key...${NC}"
openssl genrsa -out "$CERT_DIR/aura-testnet.key" 4096

# Set proper permissions on private key
chmod 600 "$CERT_DIR/aura-testnet.key"

# Generate certificate signing request
echo -e "${GREEN}Generating certificate signing request...${NC}"
openssl req -new -key "$CERT_DIR/aura-testnet.key" \
    -out "$CERT_DIR/aura-testnet.csr" \
    -subj "/C=US/ST=State/L=City/O=Aura Blockchain/OU=Testnet/CN=$DOMAIN" \
    -addext "subjectAltName=DNS:$DOMAIN,DNS:*.$DOMAIN,DNS:localhost,IP:127.0.0.1"

# Generate self-signed certificate
echo -e "${GREEN}Generating self-signed certificate...${NC}"
openssl x509 -req -days "$DAYS" \
    -in "$CERT_DIR/aura-testnet.csr" \
    -signkey "$CERT_DIR/aura-testnet.key" \
    -out "$CERT_DIR/aura-testnet.crt" \
    -extensions v3_req \
    -extfile <(cat <<EOF
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = $DOMAIN
DNS.2 = *.$DOMAIN
DNS.3 = localhost
IP.1 = 127.0.0.1
EOF
)

# Set proper permissions
chmod 644 "$CERT_DIR/aura-testnet.crt"

# Display certificate information
echo ""
echo -e "${GREEN}Certificate generated successfully!${NC}"
echo ""
echo "Files created:"
echo "  Private key: $CERT_DIR/aura-testnet.key"
echo "  Certificate: $CERT_DIR/aura-testnet.crt"
echo "  CSR: $CERT_DIR/aura-testnet.csr"
echo ""

# Show certificate details
echo -e "${GREEN}Certificate details:${NC}"
openssl x509 -in "$CERT_DIR/aura-testnet.crt" -text -noout | grep -A 3 "Subject:"
openssl x509 -in "$CERT_DIR/aura-testnet.crt" -text -noout | grep -A 1 "Validity"
openssl x509 -in "$CERT_DIR/aura-testnet.crt" -text -noout | grep -A 3 "Subject Alternative Name"

echo ""
echo -e "${YELLOW}NOTE: This is a self-signed certificate for testnet use only.${NC}"
echo -e "${YELLOW}For production, use Let's Encrypt or a proper CA-signed certificate.${NC}"
echo ""
echo "To verify the certificate:"
echo "  openssl verify -CAfile $CERT_DIR/aura-testnet.crt $CERT_DIR/aura-testnet.crt"
echo ""
echo "To trust this certificate locally (for testing):"
echo "  # Ubuntu/Debian:"
echo "  sudo cp $CERT_DIR/aura-testnet.crt /usr/local/share/ca-certificates/"
echo "  sudo update-ca-certificates"
echo ""
echo "  # macOS:"
echo "  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERT_DIR/aura-testnet.crt"
