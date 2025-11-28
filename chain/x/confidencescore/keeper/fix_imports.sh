#!/bin/bash

# Fix all sdk.PrefixEndBytes to storetypes.PrefixEndBytes
for file in score_decay.go score_delegation.go score_marketplace.go score_verification.go slash.go; do
  # Add import if not present
  if ! grep -q "storetypes \"cosmossdk.io/store/types\"" "$file"; then
    sed -i '/import (/a\        storetypes "cosmossdk.io/store/types"' "$file"
  fi
  # Replace sdk.PrefixEndBytes
  sed -i 's/sdk\.PrefixEndBytes/storetypes.PrefixEndBytes/g' "$file"
done

echo "Fixed imports in all files"
