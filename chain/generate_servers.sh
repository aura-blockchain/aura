#!/bin/bash

# List of modules that need gRPC server implementations
MODULES=(
    "dex"
    "governance"
    "incidentresponse"
    "monitoring"
    "prevalidation"
    "privacy"
    "validatorsecurity"
    "walletsecurity"
)

echo "Modules requiring gRPC server implementation:"
for module in "${MODULES[@]}"; do
    echo "  - $module"
    echo "    Proto: proto/aura/$module/v1beta1/"
    echo "    Keeper: x/$module/keeper/"
done
