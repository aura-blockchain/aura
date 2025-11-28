# Aura Bindings (Rust)

This crate provides the Rust-side message/query types required to call Aura’s custom CosmWasm bindings. It mirrors the Go implementation in `chain/x/aura-bindings` and is the canonical source for contract developers.

## Features

- Type-safe definitions for Aura custom messages (`AuraMsg`) and queries (`AuraQuery`).
- `CustomMsg` implementation so contracts can emit `CosmosMsg::Custom(AuraMsg)`.
- Helper types such as `VerifiableCredential`.
- Serialization tests that ensure Rust → JSON matches what the Cosmos SDK runtime expects.

## Installation

Add to your contract `Cargo.toml`:

```toml
[dependencies]
aura-bindings = { path = "../packages/aura-bindings" }
cosmwasm-std = "1.5"
serde = { version = "1.0", features = ["derive"] }
schemars = "0.8"
```

## Usage

```rust
use aura_bindings::{AuraMsg, AuraQuery, GetVCQuery, VerifiableCredential};
use cosmwasm_std::{Deps, DepsMut, MessageInfo, QueryRequest, Response, StdResult};

pub fn execute_register_vc(
    _deps: DepsMut,
    info: MessageInfo,
    vc_base64: String,
) -> StdResult<Response<AuraMsg>> {
    let msg = AuraMsg::VCRegistry {
        register_vc: aura_bindings::msg::RegisterVCMsg {
            address: info.sender.to_string(),
            vc_base64,
        },
    };
    Ok(Response::new().add_message(cosmwasm_std::CosmosMsg::Custom(msg)))
}

pub fn query_vc(deps: Deps, address: String) -> StdResult<VerifiableCredential> {
    let query = AuraQuery::VCRegistry {
        get_vc: GetVCQuery { address },
    };
    let request: QueryRequest<AuraQuery> = QueryRequest::Custom(query);
    deps.querier.query(&request)
}
```

## Current Surface Area

To avoid overpromising, we document only the bindings implemented today.

| Module | Messages | Queries |
| --- | --- | --- |
| VC Registry | `AuraMsg::VCRegistry { register_vc }` | `AuraQuery::VCRegistry { get_vc }` |

See `contracts/binding-tester` for an end-to-end contract using these APIs, and `chain/x/aura-bindings` for the Go-side message/query plugins.

## Roadmap

Planned additions (see [`contracts/PORTFOLIO_PLAN.md`](../PORTFOLIO_PLAN.md)):

- VC registry extensions: revoke/list/status queries, issuer helper utilities.
- Disclosure bindings: request/response flows and policy queries.
- Confidence score queries (user score, IR completion).

Each new binding will ship with:
1. Type definitions (`msg.rs`, `query.rs`).
2. Serialization tests (`src/lib.rs`).
3. Schema updates for the consuming contracts.
4. README updates describing the new surface.

## Testing

Run the crate tests to ensure JSON compatibility:

```bash
cargo test -p aura-bindings
```

The tests assert the serialized JSON matches the Go runtime expectations; add new cases whenever additional messages/queries are introduced.
