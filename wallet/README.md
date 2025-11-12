# Wallet & Light Client

## Purpose
Deliver the non-custodial wallet experience that consumes the identity manager, inclusion routine metadata, and verifier attestations while staying lightweight for mobile use.

## Anchors
- `docs/rfcs/0006-wallet-light-client.md` describes the expected flows, state machines, and security boundaries for the wallet.
- `docs/architecture/flows/ir-completion.puml` plus the verifier proof diagrams explain how the wallet interacts with assistants and the chain.

## Next steps
1. Draft the initial wallet architecture (sync layer, IR queueing, credential cache) and how it validates VC issuance events.
2. Define UI/UX requirements for onboarding, proof submission, and governance voting eligibility.
3. Determine which attestation/IR metadata needs to be cached locally versus fetched via gRPC/REST.
4. Plan light-client sync strategy (e.g., block headers, IBC relays) for the mobile bridge to the Cosmos chain.

