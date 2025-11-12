# Verifier Portal & Fee Tracking

## Purpose
Provide the business-facing portal that requests VC proofs, pays verifier fees, and feeds telemetry into the economics modeling stack.

## Anchors
- `docs/ops/runbooks/genesis-wallet-prep.md` and `incident-template.md` describe operations that the portal must support (alerts, verification audits, genesis prep).
- `tools/aggregate_verifier_fees.py` and `docs/economics/models/` demonstrate how fee events feed downstream analysis.

## Next steps
1. Define the verifier workflow for requesting proofs, settling fees, and logging outcomes so assistant rewards and burns stay auditable.
2. Automate telemetry exports from the portal into `data/verifier-fee-events/` and ensure `aggregate_verifier_fees.py` remains in sync with the CSV schema.
3. Build dashboards or CSV exports that duplicate the aggregated metrics in APIs the economics team can ingest.
4. Tie alerting/SLAs to ops runbooks so manipulations and outages fire actionable incidents.

