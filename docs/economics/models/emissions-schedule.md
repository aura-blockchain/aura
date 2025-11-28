# AURA Emissions Schedule Draft

This note translates RFC-0007’s high-level splits into concrete per-year emission targets for the validator + AI assistant allocation (40% of the 1B AURA supply). Values are mirrored in `emissions-schedule.csv` for spreadsheet import.

## Key Inputs
- Fixed supply: **1,000,000,000 AURA**.
- Validator/assistant allocation: **40% (400M AURA)** released over the first 10 years.
- Reward split: **60% validators / 40% assistants** with regional weighting handled in chain code.
- Blocks per year: assume **~31,536,000 seconds** (Cosmos default 1s block target) for per-block rates.

## Yearly Targets

| Year | % of Supply | Minted (AURA) | Validator Share | Assistant Share | Cumulative % |
| ---- | ----------- | ------------ | --------------- | --------------- | ------------ |
| 1 | 12.0% | 120,000,000 | 72,000,000 | 48,000,000 | 12.0% |
| 2 | 8.0% | 80,000,000 | 48,000,000 | 32,000,000 | 20.0% |
| 3 | 4.0% | 40,000,000 | 24,000,000 | 16,000,000 | 24.0% |
| 4 | 4.0% | 40,000,000 | 24,000,000 | 16,000,000 | 28.0% |
| 5 | 4.0% | 40,000,000 | 24,000,000 | 16,000,000 | 32.0% |
| 6 | 1.6% | 16,000,000 | 9,600,000 | 6,400,000 | 33.6% |
| 7 | 1.6% | 16,000,000 | 9,600,000 | 6,400,000 | 35.2% |
| 8 | 1.6% | 16,000,000 | 9,600,000 | 6,400,000 | 36.8% |
| 9 | 1.6% | 16,000,000 | 9,600,000 | 6,400,000 | 38.4% |
| 10 | 1.6% | 16,000,000 | 9,600,000 | 6,400,000 | 40.0% |

Per-block emission for year *y* is simply `minted_y / 31,536,000`. Example: Year 1 target ≈ **3.81 AURA/block** (2.29 AURA validators, 1.52 AURA assistants).

## Notes & Follow-Ups
- RFC text mentioned “Yr6-10 remaining 12%”; using that verbatim would overshoot the 40% cap. This model spreads the *actually remaining 8%* evenly across years 6–10 to stay within the allocation. Update RFC if this interpretation is accepted.
- Tail emission (years 6–10) keeps annual inflation under 2% before transitioning to fee-burn dominance.
- Once module parameters exist, convert this schedule into on-chain constants plus integration tests that calculate per-block minting deltas.
