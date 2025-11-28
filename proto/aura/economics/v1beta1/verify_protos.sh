#!/bin/bash
# Verification script for economics module proto files

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Economics Module Proto Verification ==="
echo ""

# Check files exist
echo "Checking files exist..."
files=(
  "economics.proto"
  "genesis.proto"
  "query.proto"
  "tx.proto"
  "README.md"
  "QUICK_REFERENCE.md"
)

for file in "${files[@]}"; do
  if [ -f "$file" ]; then
    echo "  ✓ $file"
  else
    echo "  ✗ $file NOT FOUND"
    exit 1
  fi
done
echo ""

# Count lines
echo "File statistics:"
echo "  economics.proto:     $(wc -l < economics.proto) lines"
echo "  genesis.proto:       $(wc -l < genesis.proto) lines"
echo "  query.proto:         $(wc -l < query.proto) lines"
echo "  tx.proto:            $(wc -l < tx.proto) lines"
echo "  Total proto lines:   $(cat *.proto | wc -l) lines"
echo ""

# Check syntax
echo "Checking proto syntax..."
check_proto() {
  local file=$1
  local errors=0

  # Check for syntax = "proto3"
  if ! grep -q 'syntax = "proto3";' "$file"; then
    echo "  ✗ $file: Missing proto3 syntax declaration"
    errors=$((errors + 1))
  fi

  # Check for package declaration
  if ! grep -q 'package aura.economics.v1beta1;' "$file"; then
    echo "  ✗ $file: Missing or incorrect package declaration"
    errors=$((errors + 1))
  fi

  # Check for go_package option
  if ! grep -q 'option go_package' "$file"; then
    echo "  ✗ $file: Missing go_package option"
    errors=$((errors + 1))
  fi

  if [ $errors -eq 0 ]; then
    echo "  ✓ $file: Syntax checks passed"
  fi

  return $errors
}

total_errors=0
for proto in economics.proto genesis.proto query.proto tx.proto; do
  check_proto "$proto" || total_errors=$((total_errors + $?))
done
echo ""

# Count message types
echo "Message type counts:"
echo "  economics.proto:"
echo "    Messages: $(grep -c '^message ' economics.proto)"
echo "    Enums:    $(grep -c '^enum ' economics.proto)"
echo "  genesis.proto:"
echo "    Messages: $(grep -c '^message ' genesis.proto)"
echo "  query.proto:"
echo "    Messages: $(grep -c '^message ' query.proto)"
echo "    Service:  $(grep -c '^service ' query.proto)"
echo "  tx.proto:"
echo "    Messages: $(grep -c '^message ' tx.proto)"
echo "    Service:  $(grep -c '^service ' tx.proto)"
echo ""

# Check imports
echo "Checking imports..."
check_import() {
  local file=$1
  local import=$2
  if grep -q "import \"$import\"" "$file"; then
    echo "  ✓ $file imports $import"
  else
    echo "  ⚠ $file missing import: $import"
  fi
}

check_import economics.proto "gogoproto/gogo.proto"
check_import economics.proto "google/protobuf/timestamp.proto"
check_import economics.proto "google/protobuf/duration.proto"
check_import economics.proto "cosmos/base/v1beta1/coin.proto"

check_import genesis.proto "aura/economics/v1beta1/economics.proto"

check_import query.proto "google/api/annotations.proto"
check_import query.proto "aura/economics/v1beta1/economics.proto"

check_import tx.proto "aura/economics/v1beta1/economics.proto"
echo ""

# Check for common patterns
echo "Checking for common patterns..."
patterns=(
  "customtype.*cosmossdk.io/math.Int"
  "customtype.*cosmossdk.io/math.LegacyDec"
  "stdtime.*true"
  "stdduration.*true"
  "nullable.*false"
)

for pattern in "${patterns[@]}"; do
  count=$(grep -c "$pattern" *.proto || true)
  echo "  Pattern '$pattern': $count occurrences"
done
echo ""

# Check service definitions
echo "Checking service definitions..."
echo "  Query service RPCs: $(grep -c 'rpc ' query.proto)"
echo "  Msg service RPCs:   $(grep -c 'rpc ' tx.proto)"
echo ""

# Check HTTP annotations
echo "Checking HTTP annotations in query.proto..."
http_count=$(grep -c 'google.api.http' query.proto || true)
rpc_count=$(grep -c 'rpc ' query.proto || true)
if [ "$http_count" -eq "$rpc_count" ]; then
  echo "  ✓ All $rpc_count RPCs have HTTP annotations"
else
  echo "  ⚠ $http_count HTTP annotations for $rpc_count RPCs"
fi
echo ""

# Summary
if [ $total_errors -eq 0 ]; then
  echo "=== Verification Complete: All Checks Passed ==="
  exit 0
else
  echo "=== Verification Complete: $total_errors Errors Found ==="
  exit 1
fi
