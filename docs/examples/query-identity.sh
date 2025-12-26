#!/bin/bash
# Query identity information from Aura blockchain
# Usage: ./query-identity.sh <did_or_address>

set -e

REST_API="${AURA_REST:-http://localhost:1317}"
DID="${1:-did:aura:example}"

echo "Querying identity: $DID"
echo "---"

# Query DID document
curl -s "$REST_API/aura/identity/v1/did/$DID" | jq .

# Query identity metadata
curl -s "$REST_API/aura/identity/v1/identity/$DID" | jq .

# Query associated credentials
curl -s "$REST_API/aura/vcregistry/v1/credentials_by_holder/$DID" | jq .
