#!/bin/bash
# ============================================================================
# AURA Testnet Faucet Setup Script
# ============================================================================
# This script creates and funds a faucet wallet for the AURA testnet.
# It must be run after the testnet is initialized and running.
#
# Prerequisites:
#   - Testnet must be running (docker-compose -f docker-compose.testnet.yml up -d)
#   - aurad binary must be built
#
# What this script does:
#   1. Creates a new faucet account with a mnemonic
#   2. Funds the faucet from validator-1's account
#   3. Generates a .env file for the faucet service
#   4. Displays setup instructions
# ============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
CHAIN_ID="aura-local-4"
DENOM="uaura"
FAUCET_FUNDING_AMOUNT="100000000000000"  # 100,000,000 AURA (plenty for testnet)
KEYRING_BACKEND="${AURA_KEYRING_BACKEND:-test}"
BINARY="aurad"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY_PATH="${REPO_ROOT}/chain/${BINARY}"
TESTNET_DIR="${REPO_ROOT}/testnet-data"
VALIDATOR_HOME="${TESTNET_DIR}/validator-1"
FAUCET_ENV_FILE="${REPO_ROOT}/.env.faucet"

echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}AURA Testnet Faucet Setup${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

# ============================================================================
# Step 1: Verify prerequisites
# ============================================================================
echo -e "${YELLOW}[1/6]${NC} Verifying prerequisites..."

# Check if binary exists
if [ ! -f "${BINARY_PATH}" ]; then
    echo -e "${RED}✗ Binary not found at: ${BINARY_PATH}${NC}"
    echo -e "${YELLOW}  Building binary...${NC}"
    cd "${REPO_ROOT}/chain"
    go build -o "${BINARY}" ./cmd/aurad
    if [ ! -f "${BINARY}" ]; then
        echo -e "${RED}✗ Failed to build ${BINARY}${NC}"
        exit 1
    fi
    chmod +x "${BINARY}"
    echo -e "${GREEN}✓ Binary built successfully${NC}"
else
    echo -e "${GREEN}✓ Binary found${NC}"
fi

# Check if testnet is initialized
if [ ! -d "${TESTNET_DIR}" ]; then
    echo -e "${RED}✗ Testnet not initialized${NC}"
    echo -e "${YELLOW}  Run: ./scripts/testnet-init.sh${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Testnet data found${NC}"

# Check if testnet is running
if ! docker ps | grep -q "aura-validator-1"; then
    echo -e "${RED}✗ Testnet not running${NC}"
    echo -e "${YELLOW}  Run: docker-compose -f docker-compose.testnet.yml up -d${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Testnet is running${NC}"

# ============================================================================
# Step 2: Create faucet account
# ============================================================================
echo ""
echo -e "${YELLOW}[2/6]${NC} Creating faucet account..."

FAUCET_KEY_NAME="faucet"
FAUCET_HOME="${TESTNET_DIR}/faucet"
mkdir -p "${FAUCET_HOME}"

# Check if faucet key already exists
if "${BINARY_PATH}" keys show "${FAUCET_KEY_NAME}" \
    --home "${FAUCET_HOME}" \
    --keyring-backend "${KEYRING_BACKEND}" &>/dev/null; then

    echo -e "${YELLOW}⚠ Faucet key already exists${NC}"
    read -p "Do you want to use the existing key? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}  Deleting old faucet key...${NC}"
        echo "password123" | "${BINARY_PATH}" keys delete "${FAUCET_KEY_NAME}" \
            --home "${FAUCET_HOME}" \
            --keyring-backend "${KEYRING_BACKEND}" \
            --yes &>/dev/null || true
    else
        # Retrieve existing key info
        FAUCET_ADDRESS=$(echo "password123" | "${BINARY_PATH}" keys show "${FAUCET_KEY_NAME}" \
            --home "${FAUCET_HOME}" \
            --keyring-backend "${KEYRING_BACKEND}" \
            --address 2>/dev/null)

        # Export mnemonic
        FAUCET_MNEMONIC_FILE="${FAUCET_HOME}/faucet.mnemonic"
        if [ -f "${FAUCET_MNEMONIC_FILE}" ]; then
            FAUCET_MNEMONIC=$(cat "${FAUCET_MNEMONIC_FILE}")
        else
            echo -e "${RED}✗ Mnemonic file not found. Cannot recover mnemonic for existing key.${NC}"
            echo -e "${YELLOW}  You'll need to manually set FAUCET_MNEMONIC in .env.faucet${NC}"
            FAUCET_MNEMONIC="<ENTER_MNEMONIC_MANUALLY>"
        fi

        echo -e "${GREEN}✓ Using existing faucet key${NC}"
        echo -e "${GREEN}  Address: ${FAUCET_ADDRESS}${NC}"

        # Skip to step 3
        SKIP_KEY_CREATION=1
    fi
fi

if [ -z "${SKIP_KEY_CREATION}" ]; then
    # Create new key with mnemonic
    echo "password123" | "${BINARY_PATH}" keys add "${FAUCET_KEY_NAME}" \
        --home "${FAUCET_HOME}" \
        --keyring-backend "${KEYRING_BACKEND}" \
        --output json 2>&1 | tee "${FAUCET_HOME}/faucet-key.json"

    # Extract address and mnemonic
    FAUCET_ADDRESS=$(jq -r '.address' "${FAUCET_HOME}/faucet-key.json")
    FAUCET_MNEMONIC=$(jq -r '.mnemonic' "${FAUCET_HOME}/faucet-key.json")

    # Save mnemonic to file for recovery
    echo "${FAUCET_MNEMONIC}" > "${FAUCET_HOME}/faucet.mnemonic"
    chmod 600 "${FAUCET_HOME}/faucet.mnemonic"

    echo -e "${GREEN}✓ Faucet account created${NC}"
    echo -e "${GREEN}  Address: ${FAUCET_ADDRESS}${NC}"
    echo -e "${CYAN}  Mnemonic saved to: ${FAUCET_HOME}/faucet.mnemonic${NC}"
fi

# ============================================================================
# Step 3: Get validator-1 address
# ============================================================================
echo ""
echo -e "${YELLOW}[3/6]${NC} Getting validator-1 address..."

VALIDATOR_ADDRESS=$(echo "password123" | "${BINARY_PATH}" keys show validator-1 \
    --home "${VALIDATOR_HOME}" \
    --keyring-backend "${KEYRING_BACKEND}" \
    --address 2>/dev/null)

if [ -z "${VALIDATOR_ADDRESS}" ]; then
    echo -e "${RED}✗ Failed to get validator-1 address${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Validator address: ${VALIDATOR_ADDRESS}${NC}"

# ============================================================================
# Step 4: Wait for chain to be ready
# ============================================================================
echo ""
echo -e "${YELLOW}[4/6]${NC} Waiting for chain to be ready..."

MAX_RETRIES=30
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker exec aura-validator-1 aurad status &>/dev/null; then
        echo -e "${GREEN}✓ Chain is ready${NC}"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo -e "  Waiting... (${RETRY_COUNT}/${MAX_RETRIES})"
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo -e "${RED}✗ Chain did not become ready in time${NC}"
    exit 1
fi

# ============================================================================
# Step 5: Fund faucet account from validator-1
# ============================================================================
echo ""
echo -e "${YELLOW}[5/6]${NC} Funding faucet account..."
echo -e "  Amount: ${FAUCET_FUNDING_AMOUNT}${DENOM} (100,000,000 AURA)"

# Execute transfer inside validator container
docker exec aura-validator-1 bash -c "
    echo 'password123' | aurad tx bank send ${VALIDATOR_ADDRESS} ${FAUCET_ADDRESS} ${FAUCET_FUNDING_AMOUNT}${DENOM} \
        --chain-id ${CHAIN_ID} \
        --keyring-backend ${KEYRING_BACKEND} \
        --gas 200000 \
        --gas-prices 0.025${DENOM} \
        --yes \
        --output json
" | tee "${FAUCET_HOME}/funding-tx.json"

# Wait for transaction to be processed
echo -e "  ${CYAN}Waiting for transaction to be processed...${NC}"
sleep 6

# Verify balance
FAUCET_BALANCE=$(docker exec aura-validator-1 aurad query bank balances ${FAUCET_ADDRESS} \
    --chain-id ${CHAIN_ID} \
    --output json | jq -r ".balances[] | select(.denom==\"${DENOM}\") | .amount")

if [ -z "${FAUCET_BALANCE}" ] || [ "${FAUCET_BALANCE}" == "0" ]; then
    echo -e "${RED}✗ Failed to fund faucet account${NC}"
    echo -e "${YELLOW}  Check the transaction details in: ${FAUCET_HOME}/funding-tx.json${NC}"
    exit 1
fi

# Convert to human-readable AURA (divide by 1,000,000)
FAUCET_BALANCE_AURA=$((FAUCET_BALANCE / 1000000))

echo -e "${GREEN}✓ Faucet funded successfully${NC}"
echo -e "${GREEN}  Balance: ${FAUCET_BALANCE}${DENOM} (${FAUCET_BALANCE_AURA} AURA)${NC}"

# ============================================================================
# Step 6: Generate .env file for faucet service
# ============================================================================
echo ""
echo -e "${YELLOW}[6/6]${NC} Generating .env file for faucet service..."

cat > "${FAUCET_ENV_FILE}" <<EOF
# ============================================================================
# AURA Testnet Faucet Configuration
# ============================================================================
# Generated by: scripts/faucet-setup.sh
# Generated at: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
# ============================================================================

# Faucet Wallet Configuration
FAUCET_MNEMONIC="${FAUCET_MNEMONIC}"
FAUCET_ADDRESS="${FAUCET_ADDRESS}"

# Faucet Service Configuration
FAUCET_ENVIRONMENT=development
FAUCET_CORS_ORIGINS=*
FAUCET_LOG_LEVEL=info

# Faucet Distribution Settings
FAUCET_AMOUNT_PER_REQUEST=100000000  # 100 AURA per request
FAUCET_RATE_LIMIT_IP=20              # Max requests per IP per window
FAUCET_RATE_LIMIT_ADDR=3             # Max requests per address per window
FAUCET_RATE_WINDOW=24                # Rate limit window in hours

# Database Configuration
FAUCET_DB_PASSWORD=faucet_secure_password_change_me

# Captcha Configuration (optional for local testnet)
FAUCET_TURNSTILE_SECRET=

# ============================================================================
# Advanced Configuration (usually no need to change)
# ============================================================================
# FAUCET_GAS_LIMIT=200000
# FAUCET_GAS_PRICE=0.025uaura
EOF

chmod 600 "${FAUCET_ENV_FILE}"

echo -e "${GREEN}✓ Configuration saved to: ${FAUCET_ENV_FILE}${NC}"

# ============================================================================
# Summary and Next Steps
# ============================================================================
echo ""
echo -e "${BLUE}============================================================================${NC}"
echo -e "${GREEN}✓ Faucet Setup Complete${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""
echo -e "${CYAN}Faucet Details:${NC}"
echo -e "  Address:  ${FAUCET_ADDRESS}"
echo -e "  Balance:  ${FAUCET_BALANCE_AURA} AURA"
echo -e "  Mnemonic: (saved in ${FAUCET_HOME}/faucet.mnemonic)"
echo ""
echo -e "${CYAN}Configuration:${NC}"
echo -e "  Env file: ${FAUCET_ENV_FILE}"
echo -e "  Per request: 100 AURA"
echo -e "  Rate limit: 3 requests per address per 24h"
echo ""
echo -e "${CYAN}Next Steps:${NC}"
echo -e "  1. Review the configuration:"
echo -e "     ${YELLOW}cat ${FAUCET_ENV_FILE}${NC}"
echo ""
echo -e "  2. Start the faucet service:"
echo -e "     ${YELLOW}docker-compose -f docker-compose.faucet.yml --env-file ${FAUCET_ENV_FILE} up -d${NC}"
echo ""
echo -e "  3. View faucet logs:"
echo -e "     ${YELLOW}docker-compose -f docker-compose.faucet.yml logs -f faucet-backend${NC}"
echo ""
echo -e "  4. Access the faucet:"
echo -e "     Web UI: ${GREEN}http://localhost:8081${NC}"
echo -e "     API:    ${GREEN}http://localhost:8081/api/v1${NC}"
echo -e "     Health: ${GREEN}http://localhost:8081/api/v1/health${NC}"
echo ""
echo -e "  5. Test the faucet:"
echo -e "     ${YELLOW}curl http://localhost:8081/api/v1/faucet/info${NC}"
echo ""
echo -e "${YELLOW}⚠ SECURITY WARNING:${NC}"
echo -e "  The faucet mnemonic is stored in: ${FAUCET_HOME}/faucet.mnemonic"
echo -e "  Keep this file secure! Do not commit it to version control."
echo -e "  This is a TESTNET wallet - never use this mnemonic on mainnet!"
echo ""
echo -e "${BLUE}============================================================================${NC}"
