# Hardware Wallet & Connector Upgrade Plan (Concise)

- Goals: Ledger (webhid/node-hid), Trezor Connect, Keystone QR; WalletConnect v2; Keplr/Leap provider compat; shared message builders.

- Packages:
  - Browser: `@ledgerhq/hw-transport-webhid`, `@ledgerhq/hw-app-cosmos`, `trezor-connect`, `@keystonehq/keystone-sdk`, `@walletconnect/sign-client`.
  - Desktop: `@ledgerhq/hw-transport-node-hid`, `@ledgerhq/hw-app-cosmos`, `trezor-connect`.
  - Shared: reuse `wallet/config/chain.js` for slip44/paths/fees.

- Browser extension steps:
  1) Replace `hardware-wallet.js` WebUSB flow with ledgerjs webhid + Cosmos app; add Trezor Connect + Keystone QR signer; support BIP44 account/index picker and on-device address verify.
  2) Add WalletConnect v2 client (QR modal) and Keplr/Leap provider compatibility shim exposing `getKey`, `signAmino/Direct`, `sendTx`.
  3) Tests: ledgerjs emulator harness for address/sign; mocked Trezor Connect responses; QR golden samples for Keystone.

- Desktop steps:
  1) Add ledgerjs node-hid + Cosmos app; add Trezor Connect bridge; optional Keystone QR.
  2) Abstract signer interface shared with browser; reuse message builders.
  3) Tests: ledgerjs emulator + mocked Trezor.

- Mobile steps:
  1) Add Keystone QR signer; add Ledger Live deeplink/handoff for signing (no HID).
  2) WalletConnect v2 deep link; reuse shared message builders.
  3) Tests: mocked Keystone QR payloads + WalletConnect loopback.

- Keplr/Leap plan:
  - Provide chain-info builder (`wallet/browser-extension/src/keplr.js`) based on shared chain config for window.keplr/Leap registration.
  - Add compatibility shim exposing `getKey` + `signAmino`/`signDirect` using existing Tx builders.

- Security/UX:
  - Enforce derivation path (m/44'/118'/account'/0/index), show chain-id/fees/memo review, device binding + attest cache, fallback to software signing with warning.

- Verification:
  - Add integration tests per wallet (ledger emulator, Trezor mock, QR samples), document steps in `WALLET_TESTING_REPORT.md`.
