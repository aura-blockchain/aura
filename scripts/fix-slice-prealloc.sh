#!/bin/bash
# Pre-allocate slices in Go files to reduce memory allocations
# Pattern: var x []T followed by loop append -> x := make([]T, 0, 64)

set -e

CHAIN_DIR="/home/hudson/blockchain-projects/aura/chain/x"

# Find and fix patterns in keeper files
fix_file() {
    local file="$1"

    # Use sed to replace 'var x []T' with 'x := make([]T, 0, 64)' when followed by append
    # This is a simplified fix - manual review recommended for edge cases

    # Pattern 1: var records []*types.XXX
    sed -i 's/var \([a-zA-Z]*\) \[\]\*types\.\([a-zA-Z]*\)$/\1 := make([]*types.\2, 0, 64)/g' "$file"

    # Pattern 2: var records []types.XXX
    sed -i 's/var \([a-zA-Z]*\) \[\]types\.\([a-zA-Z]*\)$/\1 := make([]types.\2, 0, 64)/g' "$file"

    # Pattern 3: var items [][]byte
    sed -i 's/var \([a-zA-Z]*\) \[\]\[\]byte$/\1 := make([][]byte, 0, 64)/g' "$file"

    # Pattern 4: var items []string
    sed -i 's/var \([a-zA-Z]*\) \[\]string$/\1 := make([]string, 0, 64)/g' "$file"

    echo "Fixed: $file"
}

# Process all keeper files
echo "Pre-allocating slices in keeper files..."

find "$CHAIN_DIR" -name "*.go" -path "*keeper*" | while read -r file; do
    # Skip test files
    if [[ "$file" == *"_test.go" ]]; then
        continue
    fi

    # Check if file has the pattern
    if grep -q "var .* \[\]" "$file"; then
        fix_file "$file"
    fi
done

echo "Done. Run 'go build ./...' to verify changes."
