#!/bin/bash
# Participate in Aura governance
# Usage: ./governance-vote.sh <proposal_id> <vote_option>

set -e

REST_API="${AURA_REST:-http://localhost:1317}"
PROPOSAL_ID="${1:-1}"
VOTE="${2:-VOTE_OPTION_YES}"

echo "Governance voting example"
echo "---"

# List active proposals
echo "Active proposals:"
curl -s "$REST_API/cosmos/gov/v1/proposals?proposal_status=PROPOSAL_STATUS_VOTING_PERIOD" | jq '.proposals[] | {id, title: .title, status}'

# Query specific proposal
echo ""
echo "Proposal $PROPOSAL_ID details:"
curl -s "$REST_API/cosmos/gov/v1/proposals/$PROPOSAL_ID" | jq .

# Vote options: VOTE_OPTION_YES, VOTE_OPTION_NO, VOTE_OPTION_ABSTAIN, VOTE_OPTION_NO_WITH_VETO
echo ""
echo "To vote, run:"
echo "  aurad tx gov vote $PROPOSAL_ID $VOTE --from <key> --chain-id aura-testnet-1"
