# Aura Verifier Portal

Lightweight, dependency-free web UI for verifiers, compliance responders, and assistant operators. It runs entirely in the browser (static HTML/JS/CSS) and talks to the Aura REST endpoint, so you can host it via GitHub Pages, S3, or behind your SOC VPN.

## Features

- Assistant fleet table sourced from `/aura/aiassistant/v1beta1/assistants` (pagination limit 50), complete with stake, sponsorship balances, locale coverage, and misbehavior counts.
- Wallet confidence lookup via `/aura/confidencescore/v1beta1/user_score/{wallet}`.
- Inline IR completion history (top 5 results from `/aura/confidencescore/v1beta1/user_completions/{wallet}`) so reviewers can spot anomalous assistants or users quickly.
- Local settings persistence (REST endpoint, assistant regex filter, wallet under review).

## Usage

```
cd verifier-portal
npx serve .
# or use any static web server
```

Then open `http://localhost:3000` (or whichever port your server exposes), configure the REST endpoint, and start reviewing traffic.

## Security

- All requests are read-only (ledger queries). Anything requiring signing must still flow through `aurad` or the `wallet-tools` CLI.
- Because the portal is static, you can host it wherever makes sense (CloudFront, hardened Nginx, etc.) and enforce mutual TLS/CSP/pinned origins as required.
