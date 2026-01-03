#!/bin/bash
# AURA Validator Encrypted Backup Script
set -e

CHAIN="${1:-aura}"
BUCKET="${2:-aura-testnet-artifacts}"
BACKUP_DIR="$HOME/.validator-backups/${CHAIN}"
PASSPHRASE_FILE="$HOME/.validator-backups/.backup-passphrase"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR"

# Check passphrase exists
if [ ! -f "$PASSPHRASE_FILE" ]; then
    echo "ERROR: Passphrase file not found at $PASSPHRASE_FILE"
    echo "Create one with: openssl rand -base64 32 > $PASSPHRASE_FILE && chmod 600 $PASSPHRASE_FILE"
    exit 1
fi

echo "=== AURA Validator Backup ==="
echo "Timestamp: $TIMESTAMP"

# Get files from remote server
echo "Fetching validator keys..."
ssh aura-testnet 'tar -czf - -C ~/.aura/config priv_validator_key.json node_key.json genesis.json' > /tmp/aura-keys-${TIMESTAMP}.tar.gz

# Encrypt with GPG
echo "Encrypting backup..."
gpg --batch --yes --passphrase-file "$PASSPHRASE_FILE" --symmetric --cipher-algo AES256 \
    -o "${BACKUP_DIR}/validator-keys-${CHAIN}-${TIMESTAMP}.tar.gz.gpg" \
    /tmp/aura-keys-${TIMESTAMP}.tar.gz

# Create checksum
sha256sum "${BACKUP_DIR}/validator-keys-${CHAIN}-${TIMESTAMP}.tar.gz.gpg" > "${BACKUP_DIR}/validator-keys-${CHAIN}-${TIMESTAMP}.tar.gz.gpg.sha256"

# Upload to R2
echo "Uploading to R2..."
source ~/.nvm/nvm.sh && nvm use 20 > /dev/null 2>&1
wrangler r2 object put "${BUCKET}/backups/validator-keys-${CHAIN}-${TIMESTAMP}.tar.gz.gpg" \
    --file "${BACKUP_DIR}/validator-keys-${CHAIN}-${TIMESTAMP}.tar.gz.gpg" --remote

# Cleanup temp files
rm -f /tmp/aura-keys-${TIMESTAMP}.tar.gz

echo "Backup complete: validator-keys-${CHAIN}-${TIMESTAMP}.tar.gz.gpg"
echo "Checksum: $(cat ${BACKUP_DIR}/validator-keys-${CHAIN}-${TIMESTAMP}.tar.gz.gpg.sha256)"
