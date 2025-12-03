#!/bin/bash
set -e

# WASM Gas Benchmarking Script
# This script runs comprehensive gas benchmarks for WASM operations
# and generates a detailed report with recommendations.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHAIN_DIR="$PROJECT_ROOT/chain"
RESULTS_DIR="$PROJECT_ROOT/docs/benchmarks"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_FILE="$RESULTS_DIR/wasm_gas_benchmark_${TIMESTAMP}.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}AURA WASM Gas Benchmark Suite${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Create results directory if it doesn't exist
mkdir -p "$RESULTS_DIR"

# Check if contracts are compiled
echo -e "${YELLOW}Checking for compiled contracts...${NC}"
if [ ! -f "$PROJECT_ROOT/contracts/artifacts/binding_tester.wasm" ]; then
    echo -e "${RED}Error: Contracts not found. Please run 'make optimize-wasm' first.${NC}"
    echo -e "${YELLOW}Attempting to build contracts...${NC}"
    cd "$PROJECT_ROOT/contracts"
    if [ -f "Makefile" ]; then
        make optimize-wasm || {
            echo -e "${RED}Failed to build contracts. Please build manually.${NC}"
            exit 1
        }
    else
        echo -e "${RED}Makefile not found in contracts directory.${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ Contracts found${NC}"
fi

# Navigate to chain directory
cd "$CHAIN_DIR"

echo ""
echo -e "${YELLOW}Running WASM gas benchmarks...${NC}"
echo -e "${YELLOW}This may take 5-10 minutes depending on your system.${NC}"
echo ""

# Run benchmarks and capture output
go test -bench=^BenchmarkWasm -benchmem -benchtime=100x \
    -run=^$ \
    ./x/wasm/keeper \
    -timeout 30m \
    | tee "$RESULTS_FILE"

# Check if benchmarks ran successfully
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Benchmarks completed successfully${NC}"
    echo -e "${GREEN}Results saved to: $RESULTS_FILE${NC}"
    echo ""
else
    echo ""
    echo -e "${RED}✗ Benchmarks failed or were skipped${NC}"
    echo -e "${YELLOW}This may be because:${NC}"
    echo -e "  1. Contracts are not compiled (run 'make optimize-wasm')"
    echo -e "  2. Test dependencies are missing"
    echo -e "  3. WASM keeper is not properly initialized in tests"
    echo ""
    echo -e "${YELLOW}Partial results saved to: $RESULTS_FILE${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Benchmark Analysis${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Parse and analyze results
if [ -f "$RESULTS_FILE" ]; then
    echo -e "${YELLOW}Extracting gas consumption metrics...${NC}"
    echo ""

    # Extract store code benchmarks
    echo -e "${BLUE}Store Code Gas Consumption:${NC}"
    grep -A 1 "BenchmarkWasmStoreCode" "$RESULTS_FILE" | grep "gas/op" || echo "No store code results found"
    echo ""

    # Extract instantiate benchmarks
    echo -e "${BLUE}Instantiate Contract Gas Consumption:${NC}"
    grep -A 1 "BenchmarkWasmInstantiateContract" "$RESULTS_FILE" | grep "gas/op" || echo "No instantiate results found"
    echo ""

    # Extract execute benchmarks
    echo -e "${BLUE}Execute Contract Gas Consumption:${NC}"
    grep -A 1 "BenchmarkWasmExecuteContract" "$RESULTS_FILE" | grep "gas/op" || echo "No execute results found"
    echo ""

    # Extract full lifecycle benchmarks
    echo -e "${BLUE}Full Lifecycle Gas Consumption:${NC}"
    grep -A 1 "BenchmarkWasmFullLifecycle" "$RESULTS_FILE" | grep -E "(total_gas|store_gas|instantiate_gas|execute_gas)" || echo "No lifecycle results found"
    echo ""
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Recommendations${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cat << 'EOF'
Based on typical WASM benchmarks, recommended gas settings are:

1. Store Code:
   - Small contracts (< 200KB): 5,000,000 - 10,000,000 gas
   - Medium contracts (200-500KB): 10,000,000 - 20,000,000 gas
   - Large contracts (> 500KB): 20,000,000 - 50,000,000 gas

2. Instantiate Contract:
   - Simple init (minimal state): 500,000 - 1,000,000 gas
   - Complex init (with state): 1,000,000 - 3,000,000 gas
   - Heavy init (large state): 3,000,000 - 10,000,000 gas

3. Execute Contract:
   - Simple query (read-only): 100,000 - 300,000 gas
   - Simple write (single value): 300,000 - 800,000 gas
   - Complex write (batch): 800,000 - 2,000,000 gas
   - Compute-heavy: 2,000,000 - 10,000,000 gas

4. Admin Operations:
   - Set/Get/Check admin: 50,000 - 100,000 gas

Update your app.toml with these settings based on your actual results.
See docs/benchmarks/WASM_GAS_TUNING.md for detailed tuning guide.
EOF

echo ""
echo -e "${GREEN}Benchmark suite completed!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Review results in: $RESULTS_FILE"
echo -e "  2. Update app.toml gas limits based on results"
echo -e "  3. Run local testnet to validate settings"
echo -e "  4. Document baselines in docs/benchmarks/"
echo ""
