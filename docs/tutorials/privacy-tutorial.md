# Privacy Features Tutorial

## Overview
Shielded transactions and privacy features.

## Create Shielded Account

```bash
aurad tx privacy create-shielded-account \
  --from mykey
```

## Shield Tokens (Public → Private)

```bash
aurad tx privacy shield \
  --amount 1000uaura \
  --from mykey
```

## Query Shielded Balance

```bash
# Requires view key
aurad query privacy shielded-balance \
  --address aura1... \
  --view-key <view-key>
```

## Private Transfer

```bash
aurad tx privacy transfer \
  --amount 500uaura \
  --recipient <shielded-address> \
  --from mykey
```

## Unshield Tokens (Private → Public)

```bash
aurad tx privacy unshield \
  --amount 500uaura \
  --from mykey
```

## View Key Management

### Generate View Key
```bash
aurad keys show mykey --view-key
```

### Share View Key for Auditing
```bash
aurad tx privacy grant-view-access \
  --viewer aura1auditor... \
  --expires 2026-01-01 \
  --from mykey
```

## Mixing Pool (Enhanced Privacy)

```bash
aurad tx privacy join-mixing-pool \
  --pool-id 1 \
  --amount 1000uaura \
  --from mykey
```
