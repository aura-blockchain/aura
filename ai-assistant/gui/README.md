# Aura Assistant GUI

An Electron-based desktop console that wraps the new AI assistant workflows:

- Broadcasts `tx aiassistant heartbeat` via your local `aurad` binary + OS keyring.
- Surfaces assistant status (`stake`, `sponsorship`, locales, heartbeat age) via REST.
- Issues and redeems sponsorship vouchers by shelling into the `aura-voucher` CLI (so secrets never touch the renderer).
- Pushes metrics to the Prometheus Pushgateway (when configured) so Grafana/ROI dashboards update immediately.

## Requirements

- Node.js ≥ 18
- Pre-built `aurad` binary with keys in the OS keyring.
- `aura-voucher` binary from `ai-assistant/vouchers` (see README there).

## Setup

```bash
cd ai-assistant/gui
npm install
npm start          # development
npm run dist       # produce signed installers via electron-builder
```

When the window opens:

1. Point to your `aurad` and `aura-voucher` binaries using the browse buttons.
2. Set REST/RPC endpoints, assistant address, chain ID, key name, optional Pushgateway, and (new) voucher passphrase.
3. Click “Store passphrase” to persist the voucher key secret in the OS keychain (powered by `keytar`), then refresh status to verify connectivity.

## Security notes

- The renderer never receives mnemonics or private keys. All sensitive work happens in the main process (isolated via `contextBridge`) and delegated to `aurad`/`aura-voucher`.
- Voucher passphrases are written to the OS keychain and injected as the `AURA_VOUCHER_PASSPHRASE` environment variable only for the lifespan of the CLI process.
- Logs are kept in-memory inside the UI to avoid writing sensitive payloads to disk. Copy anything you need manually.
