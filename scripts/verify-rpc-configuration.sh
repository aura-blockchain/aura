#!/bin/bash
# Verify RPC configuration files and infrastructure setup
# This script checks that all configuration files are valid and properly set up

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}Aura RPC Configuration Verification${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""

CHECKS_PASSED=0
CHECKS_FAILED=0

# Helper function
check_file() {
    local file="$1"
    local description="$2"

    echo -n "Checking $description... "

    if [ -f "$file" ]; then
        echo -e "${GREEN}FOUND${NC}"
        echo "  Path: $file"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
        return 0
    else
        echo -e "${RED}NOT FOUND${NC}"
        echo "  Expected: $file"
        CHECKS_FAILED=$((CHECKS_FAILED + 1))
        return 1
    fi
}

check_toml() {
    local file="$1"
    local key="$2"
    local expected_value="$3"
    local description="$4"

    echo -n "  Verifying $description... "

    if [ ! -f "$file" ]; then
        echo -e "${RED}FILE NOT FOUND${NC}"
        CHECKS_FAILED=$((CHECKS_FAILED + 1))
        return 1
    fi

    # Simple grep-based check (works for basic TOML)
    if grep -q "^${key}.*${expected_value}" "$file"; then
        echo -e "${GREEN}OK${NC}"
        CHECKS_PASSED=$((CHECKS_PASSED + 1))
        return 0
    else
        echo -e "${YELLOW}WARNING${NC} (may need manual verification)"
        return 0  # Don't fail on this
    fi
}

echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}1. Configuration Files${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_file "/home/decri/blockchain-projects/aura/networks/testnet/config.toml" "Testnet config.toml"
check_file "/home/decri/blockchain-projects/aura/networks/testnet/app.toml" "Testnet app.toml"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}2. Nginx Configuration${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_file "/home/decri/blockchain-projects/aura/nginx/rpc-proxy.conf" "Nginx RPC proxy config"
check_file "/home/decri/blockchain-projects/aura/nginx/ssl-config.conf" "Nginx SSL config"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}3. SSL/TLS Certificates${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_file "/home/decri/blockchain-projects/aura/nginx/ssl/aura-testnet.crt" "SSL certificate"
check_file "/home/decri/blockchain-projects/aura/nginx/ssl/aura-testnet.key" "SSL private key"

if [ -f "/home/decri/blockchain-projects/aura/nginx/ssl/aura-testnet.crt" ]; then
    echo ""
    echo "Certificate details:"
    openssl x509 -in /home/decri/blockchain-projects/aura/nginx/ssl/aura-testnet.crt -noout -subject -dates 2>/dev/null || true
fi

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}4. Docker Configuration${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_file "/home/decri/blockchain-projects/aura/docker/docker-compose.rpc.yml" "Docker Compose RPC config"
check_file "/home/decri/blockchain-projects/aura/docker/init-rpc-node.sh" "Docker init script"
check_file "/home/decri/blockchain-projects/aura/docker/prometheus-rpc.yml" "Prometheus config"
check_file "/home/decri/blockchain-projects/aura/docker/grafana-datasources.yml" "Grafana datasources config"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}5. Kubernetes Configuration${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_file "/home/decri/blockchain-projects/aura/k8s/rpc-node-deployment.yaml" "K8s RPC deployment manifest"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}6. Scripts${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_file "/home/decri/blockchain-projects/aura/scripts/generate-ssl-certs.sh" "SSL certificate generation script"
check_file "/home/decri/blockchain-projects/aura/scripts/test-rpc-endpoints.sh" "RPC endpoint test script"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}7. Documentation${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_file "/home/decri/blockchain-projects/aura/docs/ops/PUBLIC_RPC_ENDPOINTS.md" "RPC endpoints documentation"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}8. Configuration Values Verification${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

echo "Checking config.toml settings:"
check_toml "/home/decri/blockchain-projects/aura/networks/testnet/config.toml" "laddr" "26657" "RPC port 26657"
check_toml "/home/decri/blockchain-projects/aura/networks/testnet/config.toml" "cors_allowed_origins" "*" "CORS enabled"
check_toml "/home/decri/blockchain-projects/aura/networks/testnet/config.toml" "prometheus" "true" "Prometheus metrics"

echo ""
echo "Checking app.toml settings:"
check_toml "/home/decri/blockchain-projects/aura/networks/testnet/app.toml" "enable = true" "" "API enabled"
check_toml "/home/decri/blockchain-projects/aura/networks/testnet/app.toml" "1317" "" "API port 1317"
check_toml "/home/decri/blockchain-projects/aura/networks/testnet/app.toml" "9090" "" "gRPC port 9090"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}9. Port Availability Check${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

check_port() {
    local port=$1
    local service=$2

    echo -n "Checking if port $port is available for $service... "

    if command -v nc >/dev/null 2>&1; then
        if nc -z localhost $port 2>/dev/null; then
            echo -e "${YELLOW}IN USE${NC} (may need to stop existing service)"
        else
            echo -e "${GREEN}AVAILABLE${NC}"
            CHECKS_PASSED=$((CHECKS_PASSED + 1))
        fi
    elif command -v lsof >/dev/null 2>&1; then
        if lsof -i :$port >/dev/null 2>&1; then
            echo -e "${YELLOW}IN USE${NC} (may need to stop existing service)"
        else
            echo -e "${GREEN}AVAILABLE${NC}"
            CHECKS_PASSED=$((CHECKS_PASSED + 1))
        fi
    else
        echo -e "${YELLOW}SKIP${NC} (nc/lsof not available)"
    fi
}

check_port 26657 "Tendermint RPC"
check_port 1317 "Cosmos API"
check_port 9090 "gRPC"
check_port 9091 "gRPC-Web"
check_port 26660 "Prometheus metrics"
check_port 443 "HTTPS"
check_port 80 "HTTP"

echo ""
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}10. Binary and Dependencies${NC}"
echo -e "${BLUE}==================================================${NC}"
echo ""

echo -n "Checking for aurad binary... "
if [ -f "/home/decri/blockchain-projects/aura/chain/aurad" ]; then
    echo -e "${GREEN}FOUND${NC}"
    /home/decri/blockchain-projects/aura/chain/aurad version 2>/dev/null || echo "  (version check failed)"
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
else
    echo -e "${YELLOW}NOT FOUND${NC}"
    echo "  Build with: cd chain && go build -o aurad ./cmd/aurad"
fi

echo -n "Checking for Docker... "
if command -v docker >/dev/null 2>&1; then
    echo -e "${GREEN}FOUND${NC}"
    docker --version
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
else
    echo -e "${YELLOW}NOT FOUND${NC} (required for Docker deployment)"
fi

echo -n "Checking for Docker Compose... "
if command -v docker-compose >/dev/null 2>&1; then
    echo -e "${GREEN}FOUND${NC}"
    docker-compose --version
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
else
    echo -e "${YELLOW}NOT FOUND${NC} (required for Docker deployment)"
fi

echo -n "Checking for kubectl... "
if command -v kubectl >/dev/null 2>&1; then
    echo -e "${GREEN}FOUND${NC}"
    kubectl version --client 2>/dev/null || true
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
else
    echo -e "${YELLOW}NOT FOUND${NC} (required for Kubernetes deployment)"
fi

echo -n "Checking for grpcurl... "
if command -v grpcurl >/dev/null 2>&1; then
    echo -e "${GREEN}FOUND${NC}"
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
else
    echo -e "${YELLOW}NOT FOUND${NC} (optional, for gRPC testing)"
    echo "  Install: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
fi

echo ""
echo -e "${YELLOW}==================================================${NC}"
echo -e "${YELLOW}Verification Summary${NC}"
echo -e "${YELLOW}==================================================${NC}"
echo ""
echo "Total checks: $((CHECKS_PASSED + CHECKS_FAILED))"
echo -e "${GREEN}Passed: $CHECKS_PASSED${NC}"
echo -e "${RED}Failed: $CHECKS_FAILED${NC}"
echo ""

if [ $CHECKS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All critical checks passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Start RPC node: docker-compose -f docker/docker-compose.rpc.yml up -d"
    echo "  2. Test endpoints: ./scripts/test-rpc-endpoints.sh"
    echo "  3. View logs: docker-compose -f docker/docker-compose.rpc.yml logs -f"
    echo "  4. Access Grafana: http://localhost:3001 (admin/admin)"
    echo ""
    exit 0
else
    echo -e "${YELLOW}⚠ Some checks failed, but configuration may still work.${NC}"
    echo "Review the failed checks above and fix as needed."
    echo ""
    exit 1
fi
