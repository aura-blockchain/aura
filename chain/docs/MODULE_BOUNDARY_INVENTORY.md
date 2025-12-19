# Module Boundary Inventory (2026-02-14)

One-line keeper dependency + ingress/egress map for module boundary audit.

- aiassistant: deps bank; ingress Msg/Query; egress balance checks + sponsorship spending via bank.
- aura-bindings: deps vcregistry; ingress Msg/Query; egress VC lookups for CosmWasm bindings.
- auth: isolated; ingress Msg/Query; egress KV-only (roles, sessions, audit logs).
- bridge: deps bank/account/vcregistry/staking; ingress Msg/Query; egress bank mints/escrow + staking slashing hooks + VC attest validation.
- compliance: isolated; ingress Msg/Query; egress params-driven KYC/OFAC decisions (no keeper calls).
- confidencescore: deps inclusionroutines via IRRegistry; ingress Msg/Query; egress score lookups used by VC/contractregistry.
- contractregistry: deps compliance/vcregistry/confidencescore (setters); ingress Msg/Query; egress policy gates for wasm uploads/registry queries.
- cryptography: isolated; ingress Msg/Query; egress deterministic crypto ops only.
- dataregistry: deps bank (setter) + IPFS client; ingress Msg/Query; egress bank fee debits + IPFS writes.
- dex: deps bank/account/vcregistry/security; ingress Msg/Query; egress bank transfers/liquidity accounting + VC checks + security pause guard.
- economics: isolated; ingress Msg/Query; egress params reads only.
- economicsecurity: isolated; ingress Msg/Query; egress params/invariant reads only.
- governance: deps staking/bank/security; ingress Msg/Query; egress staking hooks for voting power, bank movements for deposits, security authority enforcement.
- identity: isolated; ingress Msg/Query; egress KV-only identity state.
- identitychange: isolated; ingress Msg/Query; egress params-driven identity change records.
- incidentresponse: isolated; ingress Msg/Query; egress pause state + incident registry in KV.
- inclusionroutines: isolated; ingress Msg/Query; egress prerequisites consumed by confidencescore.
- monitoring: isolated; ingress Msg/Query; egress metrics emission + KV params.
- networksecurity: isolated; ingress Msg/Query; egress rate-limit state + message cache (no external keepers).
- prevalidation: isolated; ingress Msg/Query; egress KV validation metadata (no compliance hook yet).
- privacy: deps auth/bank; ingress Msg/Query; egress bank transfers + memo/ZK helpers with pluggable crypto services.
- security: deps bank/staking/account; ingress Msg/Query; egress pause controls, spending limits, staking interactions.
- validatorsecurity: deps staking/slashing/bank; ingress Msg/Query + staking hooks; egress slashing + bank penalties.
- vcregistry: deps confidencescore; ingress Msg/Query; egress credential proofs consumed by bridge/dex/contractregistry/aura-bindings.
- walletsecurity: isolated; ingress Msg/Query; egress KV-only wallet policy state.
- wasm: deps wasmd keeper + contractregistry; ingress Msg/Query; egress contract ops guarded by registry + wasm runtime.
