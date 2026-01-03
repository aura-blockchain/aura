# AURA Blockchain

## R2 Artifacts
- **Bucket**: `aura-testnet-artifacts`
- **URL**: https://artifacts.aurablockchain.org
- **Account ID**: `069b2e071fe1c5bea116a29786f2074c`

### Upload Artifacts
```bash
# Env vars pre-configured in ~/.bashrc
wrangler r2 object put aura-testnet-artifacts/genesis.json --file genesis.json --remote
wrangler r2 object put aura-testnet-artifacts/peers.txt --file peers.txt --remote
wrangler r2 object put aura-testnet-artifacts/addrbook.json --file addrbook.json --remote
```

### Delete
```bash
wrangler r2 object delete aura-testnet-artifacts/<path> --remote
```

## Testnet Server
```bash
ssh aura-testnet  # 158.69.119.76
```

## Chain Info
- Chain ID: `aura-testnet-1`
- Binary: `~/.aura/cosmovisor/genesis/bin/aurad`
- Home: `~/.aura`
