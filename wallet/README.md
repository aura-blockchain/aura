# Aura Wallet Suite

This directory hosts end-user wallet tooling adapted from the PAW and Crypto reference implementations. Each subfolder is self-contained and Aura-ready:

- `browser-extension/` – Aura Browser Wallet for Chrome/Edge/Firefox that exposes staking controls, token transfers, and DEX order submission directly from Aura REST endpoints (defaults: bech32 `aura`, denom `uaura`, REST `http://localhost:1317`).
- `desktop/` – Aura Desktop Wallet (Electron + React) with mnemonic management, staking, DEX swaps, and secure local storage.
- `mobile/` – Aura Mobile Wallet (React Native) for iOS/Android featuring wallet connect, QR receive, and staking controls.
- `web/` – Aura Web Wallet (SPA) suitable for embedding into marketing sites or hosted portals.

Each project retains its original README/build scripts, now updated for Aura network prefixes, endpoints, and branding.

## Shared Chain Configuration

- Canonical chain-registry metadata: `docs/chain-registry/aura.json`
- Shared wallet constants (bech32 prefix, slip44, denom/decimals, gas prices): `wallet/config/chain.js`
- Integrate by importing `wallet/config/chain.js` and using `CHAIN_CONFIG`/`GAS_PRICE_TIERS` for defaults (RPC/REST, fees, derivation path, bech32).
