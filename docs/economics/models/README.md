# Economic Models

Store spreadsheets, CSVs, or notebooks used to simulate AEQ emissions, validator/APR projections, AI assistant ROI, and PoI reward uptake.

Current assets:
- `emissions-schedule.md/csv` – year-by-year validator + assistant emission targets aligned with RFC-0007.
- `poi-multiplier-scenarios.md/csv` – Proof-of-Identity payout multipliers, treasury draw, remaining runway.
- `assistant-roi-scenarios.md/csv` – operating assumptions, margins, and breakeven estimates for assistant archetypes.
- `validator-apr-scenarios.md/csv` – validator/delegator APR projections across emission eras, including uptime + commission sensitivity.
- `economics-scenarios.ipynb` – lightweight dashboard that loads the above CSVs and recomputes summaries (requires `matplotlib` for the embedded charts).
- `verifier-fee-data.csv` – observed verifier fee volume (monthly) with burn, validator, assistant, and treasury shares used for ROI/APR overlays.

## Data Pipeline

- Raw per-request verifier fee events live in `data/verifier-fee-events/`. Each row represents a signed WalletConnect proof + fee paid by a verifier.
- Run `python tools/aggregate_verifier_fees.py` to regenerate `verifier-fee-data.csv` whenever new raw logs are added. Adjust `--burn-rate` or share flags if governance updates tokenomics parameters.
- The notebook pulls directly from the aggregated CSV, so keeping this pipeline up-to-date ensures ROI/APR modeling uses the latest real fee numbers.
- After updating any CSVs, run `python tools/build_economics_notebook.py` to rebuild the dashboard so the checked-in notebook reflects the latest data and chart structure.

Suggested future work:
- `assistant-roi.ipynb`
- `validator-apr-sensitivity.xlsx`

Each model must include a short assumption note and a one-paragraph conclusion so governance reviewers can interpret the output quickly.
