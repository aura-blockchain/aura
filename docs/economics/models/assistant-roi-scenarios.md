# Assistant ROI Scenarios

This note captures high-level ROI estimates for different assistant deployment strategies. Use it as a starting point for spreadsheets/notebooks that incorporate real cost data once pilots launch.

## Core Assumptions
- **Base PoI reward:** 6 AEQ per verified human (per tokenomics RFC); sample scenarios assume assistants earn 40–65% of that depending on tier.
- **AEQ reference price:** $0.50 for planning purposes.
- **Reward frequency:** Each IR completion triggers the assistant reward immediately (no batching).
- **OpEx inputs:** Cover labor, connectivity, moderation, and support tooling per scenario.
- **CapEx inputs:** Hardware kits, kiosk build-outs, or GPU servers amortized linearly.

See `assistant-roi-scenarios.csv` for full figures (IR/day, AEQ/day, USD/day, OpEx, CapEx, breakeven months).

## Scenario Highlights
- **Solo Launch:** Single operator running ~40 IR/day with mobile gear. Breaks even in just over 2 months if OpEx stays ≈$1.2k/month.
- **Regional Team:** 3-assistant pod handling 120 IR/day across locales. Higher reward tier (premium IR mix) drives breakeven ≈1.3 months despite $4.2k OpEx.
- **Verification Kiosk:** Physical storefront doing 200 IR/day but carrying rent/staff. Lower reward tier and high OpEx push breakeven to ~10 months.
- **AI Aggregator:** Cloud orchestrator processing 500 IR/day; lower per-IR share (0.4) but scale offsets costs, hitting breakeven in ~3.3 months.

## Next Steps
1. Replace placeholder AEQ price with live market feeds once token lists.
2. Layer in dynamic PoI multipliers (see `poi-multiplier-scenarios.*`) so early cohorts show larger margins.
3. Build a Jupyter notebook that ingests verifier fee data + assistant operating expenses to stress-test profitability under varying utilization.
