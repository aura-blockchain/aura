# AI Assistant Network

## Purpose
Define the staking/oracle model, attestation signer, and GUI that allow geographically distributed AI assistants to verify inclusion routines off-chain and attest proofs on-chain.

## Anchors
- `docs/rfcs/0003-ai-assistant-network.md` for stake/slashing, routing, GUI, sponsorship vouchers, and monitoring.
- `docs/architecture/flows/assistant-slashing.svg` / `assistant-slashing.puml` for feedback loops (fraud detection, stake slashing, model updates).

## Current status (2025-04-03)

- ✅ `chain/x/aiassistant` module now persists assistant records (stake, sponsorship, locales), exposes Msg/Query services, enforces heartbeats, and emits slashing telemetry.
- ✅ CLI support (`aurad tx aiassistant ...`) allows register/update-locales/heartbeat/report flows using the new protobuf services.
- ✅ Wallet tooling (`wallet-tools`) can now query DEX telemetry, create wallets, and delegate stake, providing the operational glue for assistant operators.
- ✅ Cross-platform GUI (`ai-assistant/gui`) shells into `aurad`/`aura-voucher` for heartbeats, voucher issuance/redemption, and status dashboards so non-technical operators can participate securely.
- ✅ Sponsorship vouchers (`ai-assistant/vouchers`) introduce cryptographically signed credits, encrypted key storage, and Pushgateway metrics so ROI dashboards stay up to date.
- ✅ Assistant telemetry dashboard (`grafana/dashboards/aiassistant-monitoring.json`) + `docs/economics/assistant-telemetry.md` tie heartbeat/voucher metrics straight into the economics workbook.

## Next steps
1. Harden the GUI for packaging (code signing, auto-updaters, OS keyring integration for voucher passphrases).
2. Extend vouchers with multi-sponsor federation (allowing DAO-managed sponsor keys + governance hooks).
3. Add deeper automation hooks (REST control plane for provisioning assistants at scale, not just CLI/Electron callers).
