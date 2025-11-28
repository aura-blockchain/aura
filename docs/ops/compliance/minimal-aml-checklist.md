# Minimal AML Checklist

Baseline controls expected by crypto-native verifiers and community auditors.

## Verifier Onboarding
- Collect business name, jurisdiction, contact, and public wallet address.
- Run one-time sanctions/PEP screening via approved API; store reference ID + hash on-chain in proposal metadata.
- Require self-attestation of compliance with local regulations.

## Transaction Monitoring
- Throttle verifier API keys showing:
  - >N failed proof requests in M minutes.
  - Sudden spikes in queries (>X× baseline).
- Flag PoI reward claims exceeding limits per locale/IP cluster.
- Emit structured events (`aml.alert`) for downstream analytics.

## Escalation Runbook
1. Auto-flag event created.
2. Freeze associated verifier key or PoI wallet pending review.
3. Notify DAO security council with incident summary.
4. Review evidence (logs, hashes) → vote to reinstate or slash/ban.
5. Document outcome in `docs/ops/runbooks/incidents/<id>.md`.

## Data Handling
- Store only hashes/IDs of screenings or alerts; no raw PII.
- Retain logs for 90 days (encrypted), then wipe unless tied to active investigation.

## Periodic Tasks
- Quarterly sanity checks of sanctions provider.
- Annual review of thresholds/heuristics via governance proposal.
