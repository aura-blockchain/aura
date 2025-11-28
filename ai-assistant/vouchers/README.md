# Aura Assistant Sponsorship Vouchers

This toolchain issues, verifies, and audits sponsorship vouchers that refill AI assistants' compute credits. Vouchers are signed by a sponsor's ed25519 key, logged locally for auditors, and optionally exported to Prometheus via Pushgateway so dashboards and ROI spreadsheets stay in sync.

## CLI overview

```
cd ai-assistant/vouchers
go build ./cmd/aura-voucher
./aura-voucher --help
```

The CLI stores state under `~/.aura/assistant-vouchers` by default:

- `sponsor_key.json` – sponsor ed25519 key (optionally AES-GCM encrypted with a passphrase)
- `issued.json` – JSON array of issued voucher payloads
- `redeemed.json` – JSON array of redemption records

### 1. Generate a sponsor key

```
./aura-voucher keygen --profile europe --passphrase-file ~/secrets/voucher.pass
```

Use `--profile` to separate keys/ledgers per sponsor (e.g., `default`, `europe`, `latam`). The command prints the public key (base64) that assistants can use to verify vouchers. Keep the passphrase offline, or set the `AURA_VOUCHER_PASSPHRASE` environment variable instead of providing a file.

### 2. Issue a voucher

```
./aura-voucher --profile europe issue \
  --amount 5000000 \
  --denom uaura \
  --sponsor aura1sponsor... \
  --expires 2025-05-01T00:00:00Z \
  --notes "Mexico City IR campaign" \
  --passphrase-file ~/secrets/voucher.pass \
  --pushgateway http://localhost:9091
```

The command prints a base64 payload that you can paste into Signal, Matrix, or the new assistant GUI. If `--pushgateway` is provided, the CLI increments `assistant_voucher_issue_total` with labels (`sponsor`, `denom`) so Grafana instantly reflects the issuance.

### 3. Redeem and log a voucher

Assistants paste the payload into the GUI (see below) or run:

```
./aura-voucher redeem \
  --voucher "$(cat voucher.b64)" \
  --assistant aura1assistant... \
  --expect-pubkey <sponsor_pubkey_base64> \
  --tx-hash ABC123... \
  --pushgateway http://localhost:9091
```

This verifies the signature, checks expiry, appends `redeemed.json`, and emits `assistant_voucher_redeem_total` so telemetry dashboards pick it up immediately.

### Inspect vouchers without redeeming

```
./aura-voucher inspect --voucher <payload>
```

Prints the decoded JSON for auditing or manual review.

### REST automation (optional)

Expose a local API for bots and scheduling systems:

```
./aura-voucher serve \
  --listen :8787 \
  --token supersecret \
  --profile europe
```

- `POST /api/issue` with `{ "amount": "5000000", "sponsor": "aura1..." }` returns both JSON and base64 payloads.
- `POST /api/redeem` with `{ "voucher": "<payload>", "assistant": "aura1..." }` validates and logs the redemption.
- Include `Authorization: Bearer supersecret` on every request (omit `--token` only for localhost testing).
- The service respects `AURA_VOUCHER_PASSPHRASE` or `--passphrase-file`, so signing keys stay decrypted in-memory only while requests execute.

## Internals

- **Key storage** – ed25519 keys are stored as base64. When `--passphrase-file` or `AURA_VOUCHER_PASSPHRASE` is provided, the private key is encrypted with scrypt (N=2¹⁵, r=8, p=1) + AES-256-GCM.
- **Payload format** – `SignedVoucher` includes the voucher body, base64 signature, base64 public key, and `issued_at` timestamp. The CLI exposes helper functions in `pkg/voucher` if you want to integrate vouchers into other tooling.
- **Telemetry** – When `--pushgateway` is supplied, the CLI pushes a single counter metric to the configured Pushgateway. Grafana dashboards use those counters in tandem with on-chain heartbeat telemetry to keep ROI models live.

## GUI integration

The Electron GUI under `ai-assistant/gui` shells into `aura-voucher issue/redeem` so non-technical operators never touch the terminal. When you click “Store passphrase,” the GUI saves the sponsor key secret in the OS keychain and injects it via `AURA_VOUCHER_PASSPHRASE` when issuing vouchers. Point the GUI at the CLI binary, paste vouchers, and the interface renders issuance history + remaining credit automatically.
