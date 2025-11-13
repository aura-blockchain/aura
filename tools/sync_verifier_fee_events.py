#!/usr/bin/env python3
"""Fetch verifier fee events archives and refresh the monthly aggregates."""

from __future__ import annotations

import argparse
import fnmatch
import shutil
import subprocess
import tempfile
import zipfile
from pathlib import Path
from urllib.request import Request, urlopen


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--source-url",
        type=str,
        required=True,
        help="HTTP(S) URL pointing to a zip archive containing verifier fee CSVs.",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path("data/verifier-fee-events"),
        help="Directory to write the extracted CSV files.",
    )
    parser.add_argument(
        "--pattern",
        type=str,
        default="*.csv",
        help="Glob pattern used to filter entries inside the archive.",
    )
    parser.add_argument(
        "--auth-token",
        type=str,
        default=None,
        help="Optional bearer token for private telemetry buckets.",
    )
    parser.add_argument(
        "--clean",
        action="store_true",
        help="Remove existing CSVs from the output directory before extraction.",
    )
    parser.add_argument(
        "--run-aggregator",
        action="store_true",
        help="Invoke tools/aggregate_verifier_fees.py after refreshing the CSVs.",
    )
    return parser.parse_args()


def download_archive(url: str, headers: dict[str, str]) -> Path:
    req = Request(url, headers=headers)
    with urlopen(req) as response, tempfile.NamedTemporaryFile(delete=False) as tmp:
        shutil.copyfileobj(response, tmp)
        return Path(tmp.name)


def extract_archived_csvs(archive_path: Path, output_dir: Path, pattern: str) -> list[Path]:
    if not zipfile.is_zipfile(archive_path):
        raise ValueError(f"{archive_path} is not a valid zip archive")
    extracted_files: list[Path] = []
    with zipfile.ZipFile(archive_path) as archive:
        for entry in archive.infolist():
            if not fnmatch.fnmatch(entry.filename, pattern):
                continue
            dest = output_dir / Path(entry.filename).name
            dest.parent.mkdir(parents=True, exist_ok=True)
            with archive.open(entry) as src, dest.open("wb") as dst:
                shutil.copyfileobj(src, dst)
            extracted_files.append(dest)
    if not extracted_files:
        raise FileNotFoundError(f"no files matching {pattern} found in {archive_path}")
    return extracted_files


def clean_output_dir(output_dir: Path) -> None:
    for csv_file in output_dir.glob("*.csv"):
        csv_file.unlink()


def run_aggregator(output_dir: Path) -> None:
    subprocess.run(
        ["python", "tools/aggregate_verifier_fees.py", "--input-dir", str(output_dir)],
        check=True,
    )


def main() -> None:
    args = parse_args()
    output_dir = args.output_dir.resolve()
    if args.clean and output_dir.exists():
        clean_output_dir(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)
    headers: dict[str, str] = {}
    if args.auth_token:
        headers["Authorization"] = f"Bearer {args.auth_token}"

    archive_path = download_archive(args.source_url, headers)
    try:
        extracted = extract_archived_csvs(archive_path, output_dir, args.pattern)
        print(f"Extracted {len(extracted)} verifier fee CSV(s) to {output_dir}")
    finally:
        archive_path.unlink(missing_ok=True)

    if args.run_aggregator:
        print("Running verifier fee aggregator...")
        run_aggregator(output_dir)


if __name__ == "__main__":
    main()
