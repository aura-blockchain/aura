# PoI Multipliers & Treasury Runway

Purpose: translate RFC-0007’s multiplier rules into concrete treasury consumption checks so we can parameterize the `proof_of_identity` distribution module.

## Assumptions
- Treasury allocation: **200,000,000 AEQ** dedicated to PoI payouts.
- Base reward per verified human: **6 AEQ** (covers verifier fee rebate + participation stipend). Adjust once fee data exists.
- Multipliers: `2×` for first 2M users, `1.5×` for next 3M, `1×` afterwards with **halving every 10M** verified humans.
- Rewards stream continuously but modeled in discrete 10M-user eras for analytics simplicity.

## Runway Check

| Range | Users Covered | Multiplier | Payout/User (AEQ) | Treasury Draw | Treasury Left |
| ----- | ------------- | ---------- | ----------------- | ------------- | ------------- |
| R1 – Early adopters | 0–2M | 2.0× | 12 | 24M | 176M |
| R2 – Growth push | 2–5M | 1.5× | 9 | 27M | 149M |
| R3 – Base rewards | 5–15M | 1.0× | 6 | 60M | 89M |
| R4 – First halving | 15–25M | 0.5× | 3 | 30M | 59M |
| R5 – Second halving | 25–35M | 0.25× | 1.5 | 15M | 44M |
| R6 – Third halving | 35–45M | 0.125× | 0.75 | 7.5M | 36.5M |

Branching past 45M verified users still leaves >36M AEQ (18% of treasury) for future halvings. The CSV version (`poi-multiplier-scenarios.csv`) can seed notebooks that simulate various verification rates and fee offsets.

## Next Steps
1. Validate the **6 AEQ base reward** against expected verifier costs; update both CSV + RFC if it changes.
2. Encode these breakpoints as module params (user-count thresholds + multipliers) so governance can adjust them without code changes.
3. Add Monte Carlo scripts that pull verification growth forecasts to stress-test treasury depletion risk.
