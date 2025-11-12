#!/usr/bin/env python3
"""Rebuild docs/economics/models/economics-scenarios.ipynb from templates."""

from __future__ import annotations

import json
import textwrap
from pathlib import Path


def md(text: str) -> dict:
    return {
        "cell_type": "markdown",
        "metadata": {},
        "source": textwrap.dedent(text).strip("\n").splitlines(True),
    }


def code(text: str) -> dict:
    return {
        "cell_type": "code",
        "execution_count": None,
        "metadata": {},
        "outputs": [],
        "source": textwrap.dedent(text).lstrip("\n").splitlines(True),
    }


cells = [
    md(
        """
        # AURA Economics Scenario Dashboard
        This notebook links emissions, PoI multiplier, assistant ROI, verifier fee, and validator APR scenarios so token economics assumptions can be tweaked in one place.
        """
    ),
    md(
        """
        Update the CSVs under `docs/economics/models/` (use `tools/aggregate_verifier_fees.py` for fee data)
        or override parameters in the cells below, then re-run to regenerate summaries and charts.
        """
    ),
    md(
        """
        ## 1. Load Scenario Data
        Point `DATA_DIR` to a custom folder if experimenting with ad-hoc files.
        """
    ),
    code(
        """
        from __future__ import annotations
        import csv
        from pathlib import Path

        DATA_DIR = Path('docs/economics/models')
        AEQ_PRICE_USD = 0.50  # override per-market

        def read_csv(name: str):
            with (DATA_DIR / name).open() as fh:
                return list(csv.DictReader(fh))

        emissions = read_csv('emissions-schedule.csv')
        poi_ranges = read_csv('poi-multiplier-scenarios.csv')
        assistant_roi = read_csv('assistant-roi-scenarios.csv')
        validator_apr = read_csv('validator-apr-scenarios.csv')
        verifier_fees = read_csv('verifier-fee-data.csv')

        len(emissions), len(poi_ranges), len(assistant_roi), len(validator_apr), len(verifier_fees)
        """
    ),
    md(
        """
        ## 2. Verifier Fee Data Summary
        Real fee data informs both assistant ROI (per-IR bonus) and validator APR (fee component).
        """
    ),
    code(
        """
        def to_float(row, key):
            return float(row[key])

        def to_int(row, key):
            return int(row[key])

        total_requests = sum(to_int(r, 'Verifier_Requests') for r in verifier_fees)
        assistant_fee_total = sum(to_float(r, 'Assistant_Fees_AEQ') for r in verifier_fees)
        validator_fee_total = sum(to_float(r, 'Validator_Fees_AEQ') for r in verifier_fees)
        burn_total = sum(to_float(r, 'Burned_AEQ') for r in verifier_fees)
        treasury_total = sum(to_float(r, 'Treasury_Fees_AEQ') for r in verifier_fees)
        observed_months = len(verifier_fees) or 1

        ASSISTANT_FEE_PER_REQUEST = assistant_fee_total / total_requests if total_requests else 0
        VALIDATOR_FEE_PER_REQUEST = validator_fee_total / total_requests if total_requests else 0
        monthly_validator_fee = validator_fee_total / observed_months
        monthly_assistant_fee = assistant_fee_total / observed_months
        ANNUAL_VALIDATOR_FEE_POOL = monthly_validator_fee * 12
        ANNUAL_ASSISTANT_FEE_POOL = monthly_assistant_fee * 12

        print(f'Total verifier requests: {total_requests:,}')
        print(f'Assistant fees (total): {assistant_fee_total:,.2f} AEQ')
        print(f'Validator fees (total): {validator_fee_total:,.2f} AEQ')
        print(f'Burned (25%): {burn_total:,.2f} AEQ')
        print(f'Treasury share: {treasury_total:,.2f} AEQ')
        print(f'Avg assistant fee / request: {ASSISTANT_FEE_PER_REQUEST:.4f} AEQ')
        print(f'Avg validator fee / request: {VALIDATOR_FEE_PER_REQUEST:.4f} AEQ')
        print(f'Annualized validator fee pool: {ANNUAL_VALIDATOR_FEE_POOL:,.2f} AEQ')
        print(f'Annualized assistant fee pool: {ANNUAL_ASSISTANT_FEE_POOL:,.2f} AEQ')
        """
    ),
    md(
        """
        ## 3. Emissions vs. Validator Rewards
        Quick comparison of annual minting vs. total validator payouts in USD.
        """
    ),
    code(
        """
        def fmt(num: float) -> str:
            return f'{num:,.2f}'

        summary = []
        for row in emissions:
            year = int(row['Year'])
            minted = float(row['Minted_AEQ'])
            validator_share = float(row['Validator_AEQ'])
            assistant_share = float(row['Assistant_AEQ'])
            summary.append({'Year': year, 'Minted_AEQ': minted, 'Validator_AEQ': validator_share, 'Assistant_AEQ': assistant_share, 'Validator_USD': validator_share * AEQ_PRICE_USD})

        print('Year | Minted (M AEQ) | Validator (M AEQ) | Validator (USD)')
        for item in summary:
            print(f"{item['Year']:>4} | {item['Minted_AEQ']/1e6:>10.2f} | {item['Validator_AEQ']/1e6:>15.2f} | ${fmt(item['Validator_USD'])}")
        """
    ),
    md(
        """
        ## 4. Assistant ROI Stress Test (with Fee Boost)
        Fee data adds a per-IR AEQ bonus before OpEx/CapEx analysis.
        """
    ),
    code(
        """
        PRICE_MULTIPLIERS = [0.5, 1.0, 1.5]  # relative to AEQ_PRICE_USD

        def compute_roi(row, price):
            gross_reward_aeq_day = float(row['Gross_AEQ_per_day'])
            fee_aeq_day = float(row['IRs_per_day']) * ASSISTANT_FEE_PER_REQUEST
            gross_usd_day = gross_reward_aeq_day * price
            fee_usd_day = fee_aeq_day * price
            monthly_gross = (gross_usd_day + fee_usd_day) * 30
            opex = float(row['Monthly_OpEx_USD'])
            capex = float(row['Hardware_CapEx_USD'])
            net = monthly_gross - opex
            break_even = capex / net if net > 0 else float('inf')
            return monthly_gross, net, break_even, fee_usd_day * 30

        for multiplier in PRICE_MULTIPLIERS:
            price = AEQ_PRICE_USD * multiplier
            print(f"\n=== AEQ price = ${price:.2f} ===")
            for row in assistant_roi:
                monthly_gross, net, break_even, fee_monthly = compute_roi(row, price)
                name = row['Scenario']
                print(f"{name:>18}: gross=${monthly_gross:8.0f} (fees+${fee_monthly:6.0f}) net=${net:8.0f} breakeven={break_even:5.2f} mo")
        """
    ),
    md(
        """
        ## 5. Validator APR Snapshot (Minted + Fees)
        Annualized validator fees are distributed proportional to stake and uptime, then combined with emissions before commission splits.
        """
    ),
    code(
        """
        for row in validator_apr:
            name = row['Scenario']
            stake = float(row['Validator_Stake_AEQ'])
            total_bonded = float(row['Total_Bonded_AEQ'])
            uptime = float(row['Uptime'])
            commission = float(row['Commission_Pct'])
            self_pct = float(row['SelfBond_Pct'])
            gross_minted = float(row['Gross_AEQ'])
            fee_component = ANNUAL_VALIDATOR_FEE_POOL * (stake / total_bonded) * uptime
            gross = gross_minted + fee_component
            selfbond = stake * self_pct
            delegator_stake = stake - selfbond
            validator_self_rewards = gross * (selfbond / stake)
            commission_rewards = gross * (delegator_stake / stake) * commission
            validator_earn = validator_self_rewards + commission_rewards
            delegator_rewards = gross * (delegator_stake / stake) * (1 - commission)
            validator_apr_pct = (validator_earn / selfbond) * 100 if selfbond else 0
            delegator_apr_pct = (delegator_rewards / delegator_stake) * 100 if delegator_stake else 0
            usd_total = gross * AEQ_PRICE_USD
            usd_fees = fee_component * AEQ_PRICE_USD
            print(f"{name:>18}: Val APR={validator_apr_pct:5.2f}% | Del APR={delegator_apr_pct:5.2f}% | Annual USD=${usd_total:,.0f} (fees ${usd_fees:,.0f})")
        """
    ),
    md(
        """
        ## 6. Fee & APR Visualizations
        Simple matplotlib charts to visualize fee distributions and validator APR changes.
        """
    ),
    code(
        """
        import matplotlib.pyplot as plt

        months = [row['Month'] for row in verifier_fees]
        assistant_vals = [float(row['Assistant_Fees_AEQ']) for row in verifier_fees]
        validator_vals = [float(row['Validator_Fees_AEQ']) for row in verifier_fees]
        treasury_vals = [float(row['Treasury_Fees_AEQ']) for row in verifier_fees]
        burned_vals = [float(row['Burned_AEQ']) for row in verifier_fees]

        fig, ax = plt.subplots(figsize=(8, 4))
        ax.bar(months, burned_vals, label='Burned', color='#f97316')
        ax.bar(months, validator_vals, bottom=burned_vals, label='Validator', color='#2563eb')
        bottom = [b + v for b, v in zip(burned_vals, validator_vals)]
        ax.bar(months, assistant_vals, bottom=bottom, label='Assistant', color='#16a34a')
        bottom = [b + a for b, a in zip(bottom, assistant_vals)]
        ax.bar(months, treasury_vals, bottom=bottom, label='Treasury', color='#9333ea')
        ax.set_ylabel('AEQ')
        ax.set_title('Verifier Fee Distribution per Month')
        ax.legend(loc='upper left')
        plt.show()
        """
    ),
    code(
        """
        import matplotlib.pyplot as plt

        scenarios = [row['Scenario'] for row in validator_apr]
        minted = [float(row['Gross_AEQ']) for row in validator_apr]
        fee_components = [ANNUAL_VALIDATOR_FEE_POOL * (float(row['Validator_Stake_AEQ']) / float(row['Total_Bonded_AEQ'])) * float(row['Uptime']) for row in validator_apr]
        totals = [m + f for m, f in zip(minted, fee_components)]

        x = range(len(scenarios))
        fig, ax = plt.subplots(figsize=(9, 4))
        ax.bar(x, minted, label='Minted Rewards', color='#38bdf8')
        ax.bar(x, fee_components, bottom=minted, label='Fee Component', color='#facc15')
        ax.set_xticks(x)
        ax.set_xticklabels(scenarios, rotation=20, ha='right')
        ax.set_ylabel('AEQ / year')
        ax.set_title('Validator Reward Composition')
        ax.legend()
        plt.tight_layout()
        plt.show()
        """
    ),
    md(
        """
        ## 7. TODO
        - Automate pulling verifier fee event CSVs from the verifier portal telemetry bucket.
        - Add probabilistic modeling (Monte Carlo) for assistant utilization + APR sensitivity.
        - Export results as CSV/JSON for governance dashboards.
        """
    ),
]

notebook = {
    "cells": cells,
    "metadata": {
        "kernelspec": {"display_name": "Python 3", "language": "python", "name": "python3"},
        "language_info": {"name": "python", "version": "3.11"},
    },
    "nbformat": 4,
    "nbformat_minor": 5,
}

OUTPUT_PATH = Path("docs/economics/models/economics-scenarios.ipynb")
OUTPUT_PATH.parent.mkdir(parents=True, exist_ok=True)
with OUTPUT_PATH.open("w", encoding="utf-8") as fh:
    json.dump(notebook, fh, indent=1)

print(f"Updated {OUTPUT_PATH}")
