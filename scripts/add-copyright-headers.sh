#!/bin/bash
# Script to add copyright headers to Go files in chain/ directory

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Copyright header to add
HEADER="// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0"

# Counters
TOTAL=0
UPDATED=0
SKIPPED=0

echo "Scanning for Go files in chain/ directory..."

# Find all .go files in chain/ directory
while IFS= read -r -d '' file; do
    TOTAL=$((TOTAL + 1))

    # Check if file already has copyright header
    if head -n 5 "$file" | grep -q "Copyright.*Aequitas Foundation\|SPDX-License-Identifier"; then
        echo -e "${YELLOW}SKIP${NC}: $file (already has header)"
        SKIPPED=$((SKIPPED + 1))
    else
        # Create temporary file with header
        {
            echo "$HEADER"
            echo ""
            cat "$file"
        } > "$file.tmp"

        # Replace original file
        mv "$file.tmp" "$file"

        echo -e "${GREEN}ADD${NC}: $file"
        UPDATED=$((UPDATED + 1))
    fi
done < <(find /home/hudson/blockchain-projects/aura/chain -type f -name "*.go" -print0)

echo ""
echo "============================================"
echo "Summary:"
echo "  Total files scanned: $TOTAL"
echo "  Files updated: $UPDATED"
echo "  Files skipped: $SKIPPED"
echo "============================================"

exit 0
