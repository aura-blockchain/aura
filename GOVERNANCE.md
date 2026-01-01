# Governance

Aura exposes a custom governance module for proposal submission, deposits, voting, and execution. Parameters (deposit/voting periods, thresholds, and categories) are chain-configured and can be queried on-chain.

## CLI

```bash
# Submit a proposal
aurad tx governance submit-proposal "Title" "Description" text --from <key> --initial-deposit 1000uaura

# Deposit to a proposal
aurad tx governance deposit <proposal-id> 1000uaura --from <key>

# Vote on a proposal
aurad tx governance vote <proposal-id> yes --from <key>

# Query governance params
aurad query governance params
```

## Notes

- Proposal categories and voting behavior are defined in the governance module and genesis config.
- Use `aurad query governance proposal <id>` and `aurad query governance proposals` to inspect active proposals.
