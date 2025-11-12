# ZK Governance

## Purpose
Implement the zero-knowledge voting layer that enforces 1-verified-person-1-vote while maintaining voter anonymity and defending against Sybil attacks.

## Anchors
- `docs/rfcs/0005-governance-zkp.md` outlines the commitment registration, voting proof circuits, nullifier tracking, and emergency fallback.
- `docs/architecture/flows/zk-governance-vote.puml` visualizes the voting lifecycle and state transitions.

## Next steps
1. Choose the circuit framework (Circom, Halo2, gnark) and detail the trusted setup/verification strategy.
2. Implement the on-chain nullifier set, vote submission flow, and dispute window enforcement.
3. Design auditor tooling that can prove the commitment registry is up to date and identify any missing nullifiers.
4. Define fallback governance behavior for verifier downtime and tie security considerations into `docs/ops/incident-template.md`.

