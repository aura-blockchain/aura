# Aequitas / AURA Identity Blockchain

## Overview
Aequitas is a zero-PII Layer-1 built with the Cosmos SDK to serve as a decentralized identity trust anchor for W3C verifiable credentials, AI assistants, and governance backed by proof-of-identity (PoI) rewards. This repository captures the technical narrative (see `Auquitas AURAcoin Blockchain.md`), RFCs, economics models, and operations playbook required to launch the chain, assistant network, and companion apps.

## Repository layout
- `Auquitas AURAcoin Blockchain.md`: the master technical specification for the protocol.
- `PROJECT_STATUS.md` / `progress.md`: living summaries of the project snapshot and short-term focus areas.
- `docs/`: architecture diagrams, RFCs, economics models, and runbooks referenced by implementations.
- `data/`: raw verifier fee event dumps consumed by the economics tooling.
- `tools/`: helper scripts such as `aggregate_verifier_fees.py` and `build_economics_notebook.py`.
- `chain/`, `ai-assistant/`, `wallet/`, `verifier-portal/`, `zkp/`, `infra/`: planned module areas for the chain, assistants, light wallet, verifier UX, zk-governance, and infrastructure/tooling. Each directory begins with a README that ties the modules back to the RFCs and next steps.

## Getting started
1. Read `PROJECT_STATUS.md` and `progress.md` to understand current milestones.
2. Review `docs/rfcs/` (especially RFCs 0002, 0003, 0005, 0006, and 0007) to internalize the target flows, modules, and parameters.
3. Inspect `docs/economics/models/` and run `python tools/aggregate_verifier_fees.py` to regenerate the fee summaries before refreshing `docs/economics/models/economics-scenarios.ipynb` via `python tools/build_economics_notebook.py`.
4. Refer to the per-module README under `chain/`, `ai-assistant/`, etc., to see the immediate implementation targets.

## Tooling reminders
- `python tools/aggregate_verifier_fees.py --input-dir data/verifier-fee-events --output docs/economics/models/verifier-fee-data.csv` recalculates monthly totals with the tokenomics burn/distribution shares.
- `python tools/build_economics_notebook.py` rewrites the scenario notebook from the template cells tracked in the script. Run it after updating any of the CSV inputs.

## Next milestones
- Lock down the RFC wording (`docs/rfcs/0002`, `0003`, `0005`–`0007`) and capture reviewer feedback ahead of coding.
- Begin implementing the chain identity modules (inclusion routines, manager) and the monitoring components that feed them from the assistant/verification stack.
- Wire the CI placeholder jobs to real build, lint, and test commands for Go, AI tooling, and the wallet stack.
- Coordinate genesis preparation and custodian approvals while refreshing the economics dashboard with Monte Carlo simulations tied to fee telemetry.

## References
- RFC overview: `docs/rfcs/README` (add if needed).
- Operations runbooks: `docs/ops/`.
- Keep the new `progress.md` updated as work ticks across modules.
