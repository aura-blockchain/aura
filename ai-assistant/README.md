# AI Assistant Network

## Purpose
Define the staking/oracle model, attestation signer, and GUI that allow geographically distributed AI assistants to verify inclusion routines off-chain and attest proofs on-chain.

## Anchors
- `docs/rfcs/0003-ai-assistant-network.md` for stake/slashing, routing, GUI, sponsorship vouchers, and monitoring.
- `docs/architecture/flows/assistant-slashing.svg` / `assistant-slashing.puml` for feedback loops (fraud detection, stake slashing, model updates).

## Next steps
1. Design the on-chain assistant registry (state, messages, heartbeat/slashing hooks) so the chain can route IR assignments by locale.
2. Draft the off-chain signer/GUI flow that hashes IR proofs, attaches assistant metadata, and keeps API keys encrypted on-device.
3. Sketch sponsorship voucher lifecycle and the telemetry needed to recharge assistant compute credits.
4. Coordinate with the economics team to publish the assistant ROI scenarios once verifier fee data is wired in.

