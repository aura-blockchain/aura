# Troubleshooting

## Node is not catching up

- Check peers and seeds in `~/.aura/config/config.toml`.
- Verify RPC connectivity and firewall rules.
- Confirm system time is correct (NTP enabled).

## RPC or API endpoints are unavailable

- Confirm `aurad` is running and listening on ports 26657 (RPC) and 1317 (API).
- Check `app.toml` and `config.toml` for enabled endpoints.

## State sync fails

- Ensure the RPC endpoint provides `/block` and `/block?height=...`.
- Increase `TRUST_OFFSET` in `scripts/join-aura-testnet.sh` if needed.

## App hash mismatch

- Remove `~/.aura/data` and resync if the chain was reset.
- Confirm the correct genesis file and chain ID.
