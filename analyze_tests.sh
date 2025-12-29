#!/bin/bash

echo "Module                  Tests MsgSrv Query Genesis Fuzz Invariants"
echo "-------------------------------------------------------------------"

for module in chain/x/*/; do
    name=$(basename "$module")
    test_count=$(find "$module" -name "*_test.go" 2>/dev/null | wc -l)
    msg_server_tests=$(find "$module" -name "*msg_server*_test.go" 2>/dev/null | wc -l)
    query_tests=$(find "$module" -name "*query*_test.go" 2>/dev/null | wc -l)
    genesis_tests=$(find "$module" -name "*genesis*_test.go" 2>/dev/null | wc -l)
    fuzz_tests=$(find "$module" -name "*fuzz*_test.go" 2>/dev/null | wc -l)
    invariant_tests=$(find "$module" -name "*invariant*_test.go" 2>/dev/null | wc -l)

    printf "%-23s %3d    %3d   %3d     %3d    %3d       %3d\n" "$name" "$test_count" "$msg_server_tests" "$query_tests" "$genesis_tests" "$fuzz_tests" "$invariant_tests"
done
