#!/bin/bash
# Issue a verifiable credential on Aura blockchain
# Usage: ./issue-credential.sh <issuer> <holder> <credential_type>

set -e

REST_API="${AURA_REST:-http://localhost:1317}"
RPC="${AURA_RPC:-http://localhost:26657}"

ISSUER="${1:-aura1issuer...}"
HOLDER="${2:-aura1holder...}"
CRED_TYPE="${3:-isVerifiedHuman}"

echo "Issuing credential: $CRED_TYPE"
echo "Issuer: $ISSUER"
echo "Holder: $HOLDER"
echo "---"

# Create credential issuance transaction
TX_BODY=$(cat <<EOF
{
  "body": {
    "messages": [{
      "@type": "/aura.vcregistry.v1.MsgIssueCredential",
      "issuer": "$ISSUER",
      "holder": "$HOLDER",
      "credential_type": "$CRED_TYPE",
      "claims": {"verified": true}
    }]
  }
}
EOF
)

echo "Transaction body:"
echo "$TX_BODY" | jq .

# Note: Sign and broadcast with aurad tx sign / broadcast
echo ""
echo "Sign with: aurad tx sign unsigned.json --from <key>"
echo "Broadcast with: aurad tx broadcast signed.json"
