# RFC-0005: ZK Governance (1-Person-1-Vote)

- **Author(s):** Governance Team
- **Status:** Draft
- **Created:** 2025-11-11
- **Target Release:** Community Testnet

## Summary

Specify the zero-knowledge voting system that restricts governance participation to wallets holding `VC:isVerifiedHuman`, enforces one vote per proposal via nullifiers, and preserves voter anonymity.

## Motivation & Goals

- Prevent Sybil attacks while maintaining privacy.
- Ensure cryptographic verifiability of each vote without linking wallet addresses.
- Provide fallback procedures if circuits or verifiers fail.

## Detailed Design

- **Circuits:**
  - `CommitmentRegistration`: user publishes `hash(secret, nullifier_seed)` on-chain.
  - `VoteProof`: proves possession of secret + valid VC + unused proposal nullifier.
- **On-Chain Flow:**
  1. User registers commitment once.
  2. Proposal enters voting period.
  3. User submits `MsgSubmitVoteProof {proposal_id, proof, nullifier}`.
  4. Contract verifies proof, checks nullifier not seen, records vote weight = 1.
- **Nullifier Set:** Map `proposal_id -> used_nullifiers`.  
- **Verifier Implementation:** Could be Go-native (gnark) or CosmWasm wrapper with HALO2/Circom verification keys stored on-chain.
- **Emergency Mode:** If verifier contract halted, governance falls back to weighted validator vote until fixed (rare, needs DAO approval).

## State Machine

Reference diagram: `docs/architecture/flows/zk-governance-vote.puml` (SVG exported alongside).

| State | Description | Allowed Transitions |
| ----- | ----------- | ------------------- |
| `RegistrationOpen` | DAO accepts new identity commitments; circuit key `vk_commit` active. | `RegistrationOpen` (self), `VotingScheduled` |
| `VotingScheduled` | Proposal created; block heights for opening/closing locked. | `VotingOpen`, `Cancelled` |
| `VotingOpen` | Contracts accept `MsgSubmitZKVote`; per-proposal nullifier set mutable. | `VotingClosed`, `PausedEmergency` |
| `PausedEmergency` | Governance halts ZK verifier (circuit bug); validator-weighted vote optionally runs in parallel. | `VotingOpen` (resume), `VotingClosed` |
| `VotingClosed` | Submission window ended. Proofs rejected; tallies aggregated. | `Finalized`, `Disputed` |
| `Disputed` | Audit detected fault; DAO may rerun vote or migrate commitments. | `VotingScheduled` (rerun), `Finalized` |
| `Finalized` | Results posted, proposal enacted per outcome. | — |

Transition guards:
- `RegistrationOpen -> VotingScheduled`: requires proposal deposit + quorum check.
- `VotingOpen -> VotingClosed`: when block height ≥ `close_height` **and** all queued proofs processed.
- `VotingOpen -> PausedEmergency`: triggered by governance veto or verifier panic.
- `VotingClosed -> Disputed`: if audit evidence submitted within `dispute_window` (default 7 days).

Events emitted at each transition keep the wallet light client and governance dashboards in sync.

## Security / Privacy Considerations

- Trusted setup ceremony requirements, key rotation plan, storage of transcripts.
- Anti-spam: votes incur minimal fee, rate-limited per block.
- Replay protection ensured by proposal-specific nullifier domain separation.

## Validation Plan

- Formal verification/audit of circuits and verifier implementation.
- Testnet with incentivized bug bounty before mainnet activation.

## Parameters & Interfaces (Draft)

- `registration_fee`: default 0 (could require dust to avoid spam).
- `vote_fee`: nominal gas + 0.1 AURA burn to deter mass submissions.
- `dispute_window`: 7 days (configurable via governance).
- gRPC queries:
  - `NullifierUsed(proposal_id, nullifier_hash)`.
  - `CommitmentStatus(addr)` returning registered commitment + revocation flag.

## Open Questions

- Should we support delegated voting (without breaking 1p1v)?
- How to handle VC revocation mid-vote?
- Do we need per-proposal parameter overrides (longer `dispute_window`, higher quorum) for constitutional changes?
