# Mobile WalletConnect + Keystone Plan (Concise)

Goal: reach parity with the extension for WalletConnect v2 and hardware-adjacent flows (Keystone QR + Ledger Live handoff).

## Scope
- WalletConnect v2 client: deep link/QR connect, session persistence, request routing to TransactionService signer.
- Keystone QR signer: display WC URI as QR for phone-to-phone? (skip) — instead, render Keystone signing QR for offline signing of built tx; import signature to broadcast.
- Ledger Live: deep-link signing handoff (non-HID).

## Implementation Steps
1) Add `wallet/mobile/src/services/WalletConnectService.js`:
   - wrap `@walletconnect/sign-client` init/connect, persist session in AsyncStorage, emit status updates.
   - request handlers: `cosmos_signDirect`/`cosmos_signAmino` call into TransactionService (software key) or Keystone flow when selected.
2) UI (React Native):
   - Add WC section to Settings/Connect screen: connect button, show URI/deeplink, status, disconnect.
   - Toggle signer source: software (default) vs Keystone QR vs Ledger Live handoff.
3) Keystone QR flow:
   - Build tx bytes via TransactionService; generate QR with Keystone payload; accept scanned signature (hex/base64) and broadcast via PawAPI.
4) Ledger Live:
   - Deep link to Ledger Live with tx payload; poll for signed tx result; broadcast via PawAPI.
5) Tests:
   - Jest mocks for WalletConnect sign-client, AsyncStorage persistence, and Keystone payloads.
   - Integration test: simulate WC session_request and ensure signer path invoked, with rejection on address mismatch/missing key.
6) Docs:
   - Update `wallet/WALLET_TESTING_REPORT.md` with mobile coverage and manual checklist (iOS/Android, WC connect, Keystone QR sign, Ledger Live handoff).
