# Aura Web Wallet (SPA)

This folder contains a web-ready build of the Aura wallet UI generated from the desktop React application. It uses the shared chain configuration in `wallet/config/chain.js` (bech32 prefix `aura`, denom `uaura`, RPC/REST endpoints from `docs/chain-registry/aura.json`).

## Running locally

1. From the repository root: `cd wallet/web`
2. Serve the static assets: `npx serve .` (or any static file server)
3. Open the printed URL in your browser (defaults to `http://localhost:3000`)

The bundle supports send/staking/governance/DEX flows with the same mocked signing paths used in the desktop build. Hardware and WalletConnect integrations are available in the browser extension; this SPA is intended for quick web access to the same UI.
