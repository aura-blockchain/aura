#!/usr/bin/env python3
"""Aggregate raw verifier fee events into monthly totals.

Reads CSV files from the input directory (default: data/verifier-fee-events)
and produces docs/economics/models/verifier-fee-data.csv with monthly totals,
burn amounts, and validator/assistant/treasury shares per the tokenomics spec.
"""

from __future__ import annotations

import argparse
import csv
from collections import defaultdict
from dataclasses import dataclass
from pathlib import Path


@dataclass
class Shares:
    burn_rate: float = 0.25
    validator_share: float = 0.35
    assistant_share: float = 0.45
    treasury_share: float = 0.20

    def validate(self) -> None:
        remainder = 1 - self.burn_rate
        total_dist = self.validator_share + self.assistant_share + self.treasury_share
        if abs(total_dist - 1.0) > 1e-9:
            raise ValueError(
                f"validator+assistant+treasury shares must sum to 1, got {total_dist}"
            )
        if not 0 <= self.burn_rate <= 1:
            raise ValueError("burn_rate must be between 0 and 1")
        if remainder <= 0:
            raise ValueError("burn_rate consumes full fee pool; no remainder to distribute")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--input-dir",
        type=Path,
        default=Path("data/verifier-fee-events"),
        help="Directory containing raw verifier fee event CSVs",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("docs/economics/models/verifier-fee-data.csv"),
        help="Destination CSV for monthly aggregates",
    )
    parser.add_argument("--burn-rate", type=float, default=0.25)
    parser.add_argument("--validator-share", type=float, default=0.35)
    parser.add_argument("--assistant-share", type=float, default=0.45)
    parser.add_argument("--treasury-share", type=float, default=0.20)
    return parser.parse_args()


def month_key(ts: str) -> str:
    # ts format: YYYY-MM-DDTHH:MM:SSZ
    return ts[:7]


def load_events(directory: Path) -> list[dict[str, str]]:
    events: list[dict[str, str]] = []
    for csv_path in sorted(directory.glob("*.csv")):
        with csv_path.open() as fh:
            reader = csv.DictReader(fh)
            for row in reader:
                events.append(row)
    if not events:
        raise FileNotFoundError(f"No CSV files found in {directory}")
    return events


def aggregate(events: list[dict[str, str]], shares: Shares) -> list[dict[str, str]]:
    per_month: dict[str, dict[str, float | int]] = defaultdict(
        lambda: {
            "requests": 0,
            "total_fee": 0.0,
            "burned": 0.0,
            "validator": 0.0,
            "assistant": 0.0,
            "treasury": 0.0,
        }
    )

    dist_multiplier = 1 - shares.burn_rate

    for event in events:
        fee = float(event["fee_aeq"])
        month = month_key(event["timestamp"])
        bucket = per_month[month]
        bucket["requests"] += 1
        bucket["total_fee"] += fee
        burned = fee * shares.burn_rate
        remainder = fee - burned
        bucket["burned"] += burned
        bucket["validator"] += remainder * shares.validator_share
        bucket["assistant"] += remainder * shares.assistant_share
        bucket["treasury"] += remainder * shares.treasury_share

    output_rows: list[dict[str, str]] = []
    for month in sorted(per_month.keys()):
        bucket = per_month[month]
        avg_fee = bucket["total_fee"] / bucket["requests"]
        output_rows.append(
            {
                "Month": month,
                "Verifier_Requests": str(bucket["requests"]),
                "Avg_Fee_AEQ": f"{avg_fee:.4f}",
                "Total_Fees_AEQ": f"{bucket['total_fee']:.4f}",
                "Burned_AEQ": f"{bucket['burned']:.4f}",
                "Validator_Fees_AEQ": f"{bucket['validator']:.4f}",
                "Assistant_Fees_AEQ": f"{bucket['assistant']:.4f}",
                "Treasury_Fees_AEQ": f"{bucket['treasury']:.4f}",
                "Notes": "Aggregated from raw verifier fee events",
            }
        )

    return output_rows


def write_csv(rows: list[dict[str, str]], destination: Path) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    fieldnames = [
        "Month",
        "Verifier_Requests",
        "Avg_Fee_AEQ",
        "Total_Fees_AEQ",
        "Burned_AEQ",
        "Validator_Fees_AEQ",
        "Assistant_Fees_AEQ",
        "Treasury_Fees_AEQ",
        "Notes",
    ]
    with destination.open("w", newline="") as fh:
        writer = csv.DictWriter(fh, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)


def main() -> None:
    args = parse_args()
    shares = Shares(
        burn_rate=args.burn_rate,
        validator_share=args.validator_share,
        assistant_share=args.assistant_share,
        treasury_share=args.treasury_share,
    )
    shares.validate()
    events = load_events(args.input_dir)
    rows = aggregate(events, shares)
    write_csv(rows, args.output)
    print(f"Wrote {len(rows)} monthly rows to {args.output}")


if __name__ == "__main__":
    main()
