# Advanced Testnet Patterns

This folder captures lightweight patterns for the lower-priority roadmap
items without overcomplicating the primary testnet deployment.

## Multiple Validators

- Provision at least one additional node host.
- Initialize a new node home (separate `AURA_HOME`).
- Reuse `infra/testnet/systemd/aurad.service` on each node.
- Share the `genesis.json` + `seed-nodes.json` published on the status page.

## Geographic Distribution

- Deploy secondary validators in different regions.
- Publish their `node_id@ip:port` entries in `seed-nodes.json` using
  `EXTRA_SEEDS` or `EXTRA_PEERS` in `publish-seeds.sh`.

## Load Balanced RPC

- Add multiple RPC servers to the `aura_rpc` upstream in
  `infra/testnet/nginx/testnet-public.conf`.
- Keep the public DNS pointing at the nginx load balancer.

## Archive Node

- Follow `infra/testnet/archive/README.md` to run a full-history node.
- Expose it via a dedicated hostname if needed.
