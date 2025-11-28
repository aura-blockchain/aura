# Aura Browser Wallet Extension

This Chrome/Edge/Firefox-compatible extension exposes Aura’s REST APIs (wallet management, staking/mining controls, and DEX trading) in a single secure popup.  
It is intentionally focused on API access (no custodial storage) and serves as a light WalletConnect-style pane for Aura wallets.

## Features

- Configure the API host (defaults to `http://localhost:1317`) and wallet address (`aura1...`).
- Trigger staking / validator operations (renamed from “mining”) so you can manage rewards directly from the popup.
- Browse current trade orders and matches.
- Submit new orders (sell or buy) routed through Aura’s `/wallet-trades/orders`.
- Automatically refresh matches and broadcast events on settlement.

## Installation

1. Build the extension by leaving the `wallet/browser-extension/` folder as-is or running `npm run build`.
2. In your Chromium-based or Firefox browser open the extensions page (e.g., `chrome://extensions`).
3. Enable developer mode (if required) and choose “Load unpacked”/“Load Temporary Add-on”.
4. Point it at `wallet/browser-extension/dist` (after `npm run build`) or the raw folder for development.

## API Notes

- “Mining” controls now call Aura staking endpoints (validator start/stop/status). The labels remain for compatibility, but the UX clarifies staking usage.
  -- Replace the default API host (`http://localhost:1317`) using the API Host field in the popup if your node runs elsewhere.
  -- The wallet registers a WalletConnect-style session (`/wallet-trades/register`) and signs each order payload with the session secret before posting; if you operate multiple nodes, configure `AURA_WALLET_TRADE_PEER_SECRET/AURA_WALLET_TRADE_PEERS` so they gossip orders via `/wallet-trades/gossip`.
  -- For enhanced security the extension performs a WalletConnect-style ECDH handshake via `/wallet-trades/wc/handshake` and `/wallet-trades/wc/confirm`, deriving per-session secrets used for signing/encrypted trade payloads.
