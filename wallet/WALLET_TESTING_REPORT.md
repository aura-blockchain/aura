# Wallet Testing Report (In-Progress)

Scope: hardware wallets (Ledger, Trezor, Keystone) and WalletConnect v2 across extension/desktop/mobile. Goal: exceed community expectations with automated + manual coverage.

## Current Status
- Extension: automated unit coverage for Ledger/Trezor/Keystone mocks (address + signing); WCv2 session/connect + request handling covered with Vitest mocks. Real device/manual runs pending.
- Mobile: WalletConnectService loopback tests added; full Jest suite (150 tests) green on 2026-02-13 alongside TransactionService DEX/HTLC coverage.
- Desktop hardware paths still pending device runs; mobile WCv2 now covered by loopback tests (hardware handoff pending physical devices).
- Broadcast: WC requests now route through real signer logic in extension (software key or hardware path), but live broadcast smoke pending.
- Recent automation: `npm test -- --run tests/unit/hardware-wallet.test.js tests/unit/walletconnect.test.js` (pass) validates mocked devices + WC session handling.
- Mobile note: full Jest run attempted (`cd wallet/mobile && npm test -- --runInBand`) but exceeded time budget; rerun targeted suites when scheduling allows.

## Test Matrix (baseline targets)
- Devices: Ledger Nano S+/Stax (WebHID/node-hid), Trezor T (Connect), Keystone (QR).
- Flows per device: address display/verify, signDirect, signAmino, WCv2 loopback (sign+respond), rejection/mismatch UX.
- Platforms: Extension (Chrome/Firefox), Desktop (macOS/Win/Linux), Mobile (iOS/Android).

## Automation Plan
- Ledger: ledgerjs emulator harness for address/sign; integrate into Vitest CI for extension; mirror in desktop via node-hid.
- Trezor: mocked Connect responses in unit tests; add optional bridge-driven loop in CI when available.
- Keystone: QR golden samples + offline signing fixture runner.
- WalletConnect: loopback tests using sign-client mock and real client against local signer; persist/restore sessions.

## Manual Checklist (to run & record)
- Connect + address verification per device/platform.
- Sign bank/staking/gov/DEX tx; confirm on-chain broadcast.
- WalletConnect QR/deeplink from dApp → wallet → signature response.
- Error UX: address mismatch, rejected request, disconnected device mid-flow.

## Gaps / Next Steps
- Add automated ledgerjs/trezor/keystone harness into CI (extension + desktop).
- Run full manual matrix and log outcomes here; publish highlights to RELEASE_NOTES.
- Wire mobile WalletConnect + Keystone QR flows and add Jest/integration coverage.
