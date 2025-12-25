#!/bin/bash

echo "# Pagination Analysis Report"
echo ""
echo "Analyzing all modules for pagination status..."
echo ""

modules=(
  aiassistant auth bridge compliance confidencescore contractregistry
  cryptography dataregistry dex economics economicsecurity governance
  identity identitychange incidentresponse inclusionroutines monitoring
  networksecurity prevalidation privacy security validatorsecurity
  vcregistry walletsecurity wasm
)

for module in "${modules[@]}"; do
  query_file="chain/x/$module/keeper/query_server.go"

  if [ ! -f "$query_file" ]; then
    continue
  fi

  echo "## Module: $module"

  # Check for list/getall queries
  list_queries=$(grep -E "func \(q|func \(k" "$query_file" | grep -iE "List|GetAll|QueryAll" | wc -l)

  # Check for PageRequest usage
  has_pagination=$(grep -c "PageRequest\|PageResponse" "$query_file" || echo "0")

  # Get actual query method names
  methods=$(grep -E "^func \(.*\) [A-Z]" "$query_file" | sed 's/func (.*) //' | sed 's/(.*$//' | grep -v "^$")

  echo "  List/GetAll queries: $list_queries"
  echo "  Has pagination: $has_pagination"
  echo "  Query methods:"
  echo "$methods" | sed 's/^/    - /'

  if [ $has_pagination -gt 0 ]; then
    echo "  Status: ✓ HAS PAGINATION"
  elif [ $list_queries -gt 0 ]; then
    echo "  Status: ✗ NEEDS PAGINATION"
  else
    echo "  Status: - No list queries"
  fi

  echo ""
done
