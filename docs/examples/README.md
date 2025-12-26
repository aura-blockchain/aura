# Aura Examples

Runnable examples for interacting with the Aura blockchain.

## Prerequisites

- Running Aura node (RPC: localhost:26657, REST: localhost:1317)
- `curl` and `jq` installed

## Examples

| Script | Description |
|--------|-------------|
| `query-identity.sh` | Query DID document and identity info |
| `issue-credential.sh` | Issue a verifiable credential |
| `governance-vote.sh` | Submit and vote on governance proposals |

## Usage

```bash
chmod +x *.sh
./query-identity.sh aura1abc...
```
