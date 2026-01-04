# Archive Node Setup

Use an archive node to serve full historical queries without pruning.

## Setup

1. Provision a dedicated host or storage volume for the archive node.
2. Initialize the node with a separate home directory:
   ```bash
   AURA_HOME=$HOME/.aura-archive \
   MONIKER=aura-testnet-archive \
   scripts/join-aura-testnet.sh
   ```
3. Enable archive settings:
   ```bash
   AURA_HOME=$HOME/.aura-archive \
   bash infra/testnet/archive/enable-archive.sh
   ```
4. Install the archive systemd unit:
   ```bash
   sudo cp infra/testnet/systemd/aurad-archive.service /etc/systemd/system/aurad-archive.service
   sudo cp infra/testnet/systemd/aurad-archive.env /etc/aura/aurad-archive.env
   sudo systemctl daemon-reload
   sudo systemctl enable --now aurad-archive
   ```
5. If this node shares a host with the main validator, update ports in
   `config/config.toml` and `config/app.toml` to avoid conflicts.

## Notes

- Keep the archive node RPC/API on a private network or separate domain.
- Reuse `infra/testnet/nginx/testnet-public.conf` if exposing public endpoints.
