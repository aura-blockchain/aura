# AURA Networks

Network configurations for AURA blockchain.

## Networks

| Network | Chain ID | Status | Description |
|---------|----------|--------|-------------|
| aura-mvp-1 | `aura-mvp-1` | Active | MVP testnet |

## Directory Structure

```
networks/
└── {chain-id}/
    ├── genesis.json    # Genesis file
    ├── peers.txt       # Persistent peers (optional)
    └── README.md       # Network-specific docs (optional)
```

## Usage

Copy the genesis file to your node's config directory:

```bash
cp networks/aura-mvp-1/genesis.json ~/.aurad/config/genesis.json
```
