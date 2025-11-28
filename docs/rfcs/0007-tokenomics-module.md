# RFC-0007: Tokenomics Module & Emissions

- **Author(s):** Economics Team
- **Status:** Draft
- **Created:** 2025-11-11
- **Target Release:** Devnet

## Summary

Define the on-chain `tokenomics` module that mints AURA emissions, enforces fee burn rules, routes rewards to validators/AI assistants/users, and manages vesting schedules (including founder/community wallets).

## Emission Overview

- **Total Supply:** 1,000,000,000 AURA (fixed).
- **Breakdown:**
  - 40% validator/assistant emissions over 10 years (front-loaded: Yr1 12%, Yr2 8%, Yr3-5 4% each, Yr6-10 1.6% each per docs/economics/models/emissions-schedule.md).
  - 20% Proof-of-Identity rewards treasury (200,000,000 AURA).
  - 20% Ecosystem/Foundation (vesting, governance controlled).
  - 20% Core Team/Founders (vesting, includes special community wallets per `docs/economics/founder-wallets.md`).
- **Fee Burn:** 25% of verifier-paid fees burned automatically each block.

## Module Responsibilities

- Track emission schedule per block/epoch and mint to distribution accounts.
- Split validator vs. AI assistant rewards (e.g., 60/40) with locale weighting.
- Manage PoI treasury payouts; apply multipliers (2× first 2M users, 1.5× next 3M, base thereafter, halving every 10M verified users).
- Enforce vesting schedules for founder/team/community wallets and expose query APIs.
- Provide governance hooks to adjust parameters (e.g., reward weights) within predefined bounds.

## Security / Economics Considerations

- Emission changes require governance proposals with cooling-off periods.
- Treasury disbursements logged with proposal references.
- PoI multipliers capped to prevent runaway inflation.

## Validation Plan

- Unit tests for emission curves, fee burn accounting, vesting payouts.
- Economic simulation scripts (docs/economics/models) verifying sustainability.

## Open Questions

- Exact validator vs. assistant split? (placeholder 60/40).
- Should fee burn target be adjustable via governance? 
