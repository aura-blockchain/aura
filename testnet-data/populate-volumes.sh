#!/bin/bash
# Populate Docker volumes with initialized testnet data

set -e

TESTNET_DIR="$(dirname "$0")"
VALIDATORS=("validator-1" "validator-2" "validator-3" "validator-4")

echo "Populating Docker volumes..."

for VALIDATOR in "${VALIDATORS[@]}"; do
    VOLUME_NAME="aura_${VALIDATOR}-data"

    # Create a temporary container to copy data
    echo "  Copying data for ${VALIDATOR}..."

    docker run --rm \
        -v "${VOLUME_NAME}:/home/aura/.aura" \
        -v "${TESTNET_DIR}/${VALIDATOR}:/source:ro" \
        alpine sh -c "cp -r /source/config /home/aura/.aura/ && \
                      cp -r /source/data /home/aura/.aura/ && \
                      cp -r /source/keyring-test /home/aura/.aura/ 2>/dev/null || true && \
                      chown -R 1000:1000 /home/aura/.aura"

    echo "  ✓ ${VALIDATOR} volume populated"
done

echo "All volumes populated successfully!"
